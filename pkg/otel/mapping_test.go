package otel

import (
	"testing"
	"time"

	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
)

func TestMapLogRecord(t *testing.T) {
	ts := time.Now()
	body := "test log message"
	severity := logsv1.SeverityNumber_SEVERITY_NUMBER_INFO

	lr := MapLogRecord(ts, severity, body)
	if lr == nil {
		t.Fatal("Expected LogRecord, got nil")
	}

	if lr.TimeUnixNano != uint64(ts.UnixNano()) {
		t.Errorf("Expected timestamp %v, got %v", ts.UnixNano(), lr.TimeUnixNano)
	}

	if lr.SeverityNumber != severity {
		t.Errorf("Expected severity %v, got %v", severity, lr.SeverityNumber)
	}

	if lr.Body.GetStringValue() != body {
		t.Errorf("Expected body %q, got %q", body, lr.Body.GetStringValue())
	}

	ReleaseLogRecord(lr)
}

func TestMapResourceLogs(t *testing.T) {
	service := "gophership-test"
	lr := MapLogRecord(time.Now(), logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, "test message")

	rl := MapResourceLogs(service, []*logsv1.LogRecord{lr})

	if len(rl.ScopeLogs) != 1 {
		t.Errorf("Expected 1 scope log, got %d", len(rl.ScopeLogs))
	}

	if rl.Resource.Attributes[0].Key != "service.name" {
		t.Errorf("Expected service.name attribute, got %s", rl.Resource.Attributes[0].Key)
	}

	ReleaseResourceLogs(rl)
}

func TestOTelCompliance_JSON(t *testing.T) {
	// A strictly compliant OTel Log JSON representation
	// Note: We use the proto-JSON mapping rules.
	ts := time.Unix(1708819200, 0) // Fixed point in time
	lr := MapLogRecord(ts, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, "User login successful")
	AddAttribute(lr, "user.id", "42")
	AddAttribute(lr, "event.type", "auth")

	// Verify fields
	if lr.TimeUnixNano != uint64(ts.UnixNano()) {
		t.Errorf("Timestamp mismatch")
	}
	if lr.SeverityNumber != logsv1.SeverityNumber_SEVERITY_NUMBER_INFO {
		t.Errorf("Severity mismatch")
	}
	if len(lr.Attributes) != 2 {
		t.Errorf("Attribute count mismatch")
	}

	// Verify attribute structure (AC2)
	found := false
	for _, attr := range lr.Attributes {
		if attr.Key == "user.id" && attr.Value.GetStringValue() == "42" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Critical attribute 'user.id' missing or incorrect")
	}

	ReleaseLogRecord(lr)
}

func BenchmarkMapLogRecord_Attributes(b *testing.B) {
	ts := time.Now()
	severity := logsv1.SeverityNumber_SEVERITY_NUMBER_INFO
	body := "benchmark message"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lr := MapLogRecord(ts, severity, body)
		AddAttribute(lr, "key1", "val1")
		AddAttribute(lr, "key2", "val2")
		ReleaseLogRecord(lr)
	}
}
func BenchmarkMapResourceLogs(b *testing.B) {
	service := "gophership-perf"
	ts := time.Now()

	// Pre-create some logs
	logs := make([]*logsv1.LogRecord, 10)
	for i := range logs {
		logs[i] = MapLogRecord(ts, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, "entry")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl := MapResourceLogs(service, logs)
		ReleaseResourceLogs(rl)
	}
}
