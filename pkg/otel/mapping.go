package otel

import (
	"sync"
	"time"

	v1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
)

// [NFR.P1] Zero-Allocation Pools for OTel structures.
var (
	logRecordPool = sync.Pool{
		New: func() any { return &logsv1.LogRecord{} },
	}
	scopeLogsPool = sync.Pool{
		New: func() any { return &logsv1.ScopeLogs{} },
	}
	resourceLogsPool = sync.Pool{
		New: func() any { return &logsv1.ResourceLogs{} },
	}
	stringValuePool = sync.Pool{
		New: func() any { return &v1.AnyValue_StringValue{} },
	}
	keyValuePool = sync.Pool{
		New: func() any { return &v1.KeyValue{} },
	}
	anyValuePool = sync.Pool{
		New: func() any { return &v1.AnyValue{} },
	}
	resourcev1Pool = sync.Pool{
		New: func() any { return &resourcev1.Resource{} },
	}
)

// MapLogRecord creates a compliant LogRecord from raw inputs using pooled objects.
func MapLogRecord(timestamp time.Time, severity logsv1.SeverityNumber, body string) *logsv1.LogRecord {
	lr := logRecordPool.Get().(*logsv1.LogRecord)

	// Reset fields for reuse
	lr.TimeUnixNano = uint64(timestamp.UnixNano())
	lr.ObservedTimeUnixNano = uint64(time.Now().UnixNano())
	lr.SeverityNumber = severity
	lr.SeverityText = severity.String()

	if lr.Body == nil {
		lr.Body = &v1.AnyValue{}
	}

	// Reuse or allocate string value wrapper
	sv, ok := lr.Body.Value.(*v1.AnyValue_StringValue)
	if !ok {
		sv = stringValuePool.Get().(*v1.AnyValue_StringValue)
		lr.Body.Value = sv
	}
	sv.StringValue = body

	return lr
}

// ReleaseLogRecord returns a record to the pool.
func ReleaseLogRecord(lr *logsv1.LogRecord) {
	if lr == nil {
		return
	}
	for _, kv := range lr.Attributes {
		releaseKeyValue(kv)
	}
	lr.Attributes = lr.Attributes[:0]

	logRecordPool.Put(lr)
}

// AddAttribute adds a key-value pair to the LogRecord in a pooled fashion.
func AddAttribute(lr *logsv1.LogRecord, key, value string) {
	kv := keyValuePool.Get().(*v1.KeyValue)
	kv.Key = key

	if kv.Value == nil {
		kv.Value = anyValuePool.Get().(*v1.AnyValue)
	}

	sv, ok := kv.Value.Value.(*v1.AnyValue_StringValue)
	if !ok {
		sv = stringValuePool.Get().(*v1.AnyValue_StringValue)
		kv.Value.Value = sv
	}
	sv.StringValue = value

	lr.Attributes = append(lr.Attributes, kv)
}

// MapResourceLogs encapsulates a set of logs into the OTel Resource/Scope structure (AC1).
func MapResourceLogs(serviceName string, logs []*logsv1.LogRecord) *logsv1.ResourceLogs {
	rl := resourceLogsPool.Get().(*logsv1.ResourceLogs)

	if rl.Resource == nil {
		rl.Resource = resourcev1Pool.Get().(*resourcev1.Resource)
	}

	// Pooled service name attribute (AC2, AC4)
	var attr *v1.KeyValue
	if cap(rl.Resource.Attributes) > 0 {
		rl.Resource.Attributes = rl.Resource.Attributes[:1]
		if rl.Resource.Attributes[0] == nil {
			rl.Resource.Attributes[0] = keyValuePool.Get().(*v1.KeyValue)
		}
		attr = rl.Resource.Attributes[0]
	} else {
		attr = keyValuePool.Get().(*v1.KeyValue)
		rl.Resource.Attributes = append(rl.Resource.Attributes, attr)
	}

	attr.Key = "service.name"
	if attr.Value == nil {
		attr.Value = anyValuePool.Get().(*v1.AnyValue)
	}

	sv, ok := attr.Value.Value.(*v1.AnyValue_StringValue)
	if !ok {
		sv = stringValuePool.Get().(*v1.AnyValue_StringValue)
		attr.Value.Value = sv
	}
	sv.StringValue = serviceName

	var sl *logsv1.ScopeLogs
	if cap(rl.ScopeLogs) > 0 {
		rl.ScopeLogs = rl.ScopeLogs[:1]
		if rl.ScopeLogs[0] == nil {
			rl.ScopeLogs[0] = scopeLogsPool.Get().(*logsv1.ScopeLogs)
		}
		sl = rl.ScopeLogs[0]
	} else {
		sl = scopeLogsPool.Get().(*logsv1.ScopeLogs)
		rl.ScopeLogs = append(rl.ScopeLogs, sl)
	}
	sl.LogRecords = logs

	return rl
}

// ReleaseResourceLogs clears and returns the structure and its children to their pools.
func ReleaseResourceLogs(rl *logsv1.ResourceLogs) {
	if rl == nil {
		return
	}

	if rl.Resource != nil {
		for _, kv := range rl.Resource.Attributes {
			releaseKeyValue(kv)
		}
		rl.Resource.Attributes = rl.Resource.Attributes[:0]
		resourcev1Pool.Put(rl.Resource)
		rl.Resource = nil
	}

	for _, sl := range rl.ScopeLogs {
		for _, lr := range sl.LogRecords {
			ReleaseLogRecord(lr)
		}
		sl.LogRecords = sl.LogRecords[:0]
		scopeLogsPool.Put(sl)
	}
	rl.ScopeLogs = rl.ScopeLogs[:0]
	resourceLogsPool.Put(rl)
}

func releaseKeyValue(kv *v1.KeyValue) {
	if kv == nil {
		return
	}
	if kv.Value != nil {
		if sv, ok := kv.Value.Value.(*v1.AnyValue_StringValue); ok {
			stringValuePool.Put(sv)
			kv.Value.Value = nil
		}
		anyValuePool.Put(kv.Value)
		kv.Value = nil
	}
	keyValuePool.Put(kv)
}
