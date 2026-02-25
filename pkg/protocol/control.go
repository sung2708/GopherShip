package protocol

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SomaticZone represents the current pressure state.
type SomaticZone int32

const (
	SomaticZone_ZONE_UNSPECIFIED SomaticZone = 0
	SomaticZone_ZONE_GREEN       SomaticZone = 1
	SomaticZone_ZONE_YELLOW      SomaticZone = 2
	SomaticZone_ZONE_RED         SomaticZone = 3
)

func (z SomaticZone) String() string {
	switch z {
	case SomaticZone_ZONE_GREEN:
		return "GREEN"
	case SomaticZone_ZONE_YELLOW:
		return "YELLOW"
	case SomaticZone_ZONE_RED:
		return "RED"
	default:
		return "UNSPECIFIED"
	}
}

// PingResponse represents the response to a Ping heartbeat.
type PingResponse struct {
	Version       string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	UptimeSeconds int64  `protobuf:"varint,2,opt,name=uptime_seconds,json=uptimeSeconds,proto3" json:"uptime_seconds,omitempty"`
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
}

func (x *PingResponse) String() string {
	return x.Version
}

func (*PingResponse) ProtoMessage() {}

// StatusResponse represents the engine's internal health telemetry.
type StatusResponse struct {
	Zone             SomaticZone `protobuf:"varint,1,opt,name=zone,proto3,enum=gophership.protocol.v1.SomaticZone" json:"zone,omitempty"`
	PressureScore    uint32      `protobuf:"varint,2,opt,name=pressure_score,json=pressureScore,proto3" json:"pressure_score,omitempty"`
	MemoryUsageBytes uint64      `protobuf:"varint,3,opt,name=memory_usage_bytes,json=memoryUsageBytes,proto3" json:"memory_usage_bytes,omitempty"`
	HeapObjects      uint64      `protobuf:"varint,4,opt,name=heap_objects,json=heapObjects,proto3" json:"heap_objects,omitempty"`
	GoroutineCount   uint32      `protobuf:"varint,5,opt,name=goroutine_count,json=goroutineCount,proto3" json:"goroutine_count,omitempty"`
}

func (x *StatusResponse) Reset() {
	*x = StatusResponse{}
}

func (x *StatusResponse) String() string {
	return x.Zone.String()
}

func (*StatusResponse) ProtoMessage() {}

// WatchStatusRequest specifies the refresh interval for telemetry streams.
type WatchStatusRequest struct {
	RefreshIntervalMs uint32 `protobuf:"varint,1,opt,name=refresh_interval_ms,json=refreshIntervalMs,proto3" json:"refresh_interval_ms,omitempty"`
}

func (x *WatchStatusRequest) Reset() {
	*x = WatchStatusRequest{}
}

func (x *WatchStatusRequest) String() string {
	return "WatchStatusRequest"
}

func (*WatchStatusRequest) ProtoMessage() {}

// OverrideSomaticZoneRequest specifies the target somatic zone for a manual override.
type OverrideSomaticZoneRequest struct {
	Zone SomaticZone `protobuf:"varint,1,opt,name=zone,proto3" json:"zone,omitempty"`
}

func (x *OverrideSomaticZoneRequest) Reset() {
	*x = OverrideSomaticZoneRequest{}
}

func (x *OverrideSomaticZoneRequest) String() string {
	return "OverrideSomaticZoneRequest"
}

func (*OverrideSomaticZoneRequest) ProtoMessage() {}

// ControlServiceClient is the client API for ControlService.
type ControlServiceClient interface {
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*PingResponse, error)
	GetSomaticStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*StatusResponse, error)
	WatchSomaticStatus(ctx context.Context, in *WatchStatusRequest, opts ...grpc.CallOption) (ControlService_WatchSomaticStatusClient, error)
	OverrideSomaticZone(ctx context.Context, in *OverrideSomaticZoneRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type ControlService_WatchSomaticStatusClient interface {
	Recv() (*StatusResponse, error)
	grpc.ClientStream
}

type controlServiceWatchSomaticStatusClient struct {
	grpc.ClientStream
}

func (x *controlServiceWatchSomaticStatusClient) Recv() (*StatusResponse, error) {
	m := new(StatusResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

type controlServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewControlServiceClient(cc grpc.ClientConnInterface) ControlServiceClient {
	return &controlServiceClient{cc}
}

func (c *controlServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/gophership.protocol.v1.ControlService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) GetSomaticStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, "/gophership.protocol.v1.ControlService/GetSomaticStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) WatchSomaticStatus(ctx context.Context, in *WatchStatusRequest, opts ...grpc.CallOption) (ControlService_WatchSomaticStatusClient, error) {
	stream, err := c.cc.NewStream(ctx, &ControlService_ServiceDesc.Streams[0], "/gophership.protocol.v1.ControlService/WatchSomaticStatus", opts...)
	if err != nil {
		return nil, err
	}
	x := &controlServiceWatchSomaticStatusClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

func (c *controlServiceClient) OverrideSomaticZone(ctx context.Context, in *OverrideSomaticZoneRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/gophership.protocol.v1.ControlService/OverrideSomaticZone", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ControlServiceServer is the server API for ControlService.
type ControlServiceServer interface {
	Ping(context.Context, *emptypb.Empty) (*PingResponse, error)
	GetSomaticStatus(context.Context, *emptypb.Empty) (*StatusResponse, error)
	WatchSomaticStatus(*WatchStatusRequest, ControlService_WatchSomaticStatusServer) error
	OverrideSomaticZone(context.Context, *OverrideSomaticZoneRequest) (*emptypb.Empty, error)
}

type ControlService_WatchSomaticStatusServer interface {
	Send(*StatusResponse) error
	grpc.ServerStream
}

type controlServiceWatchSomaticStatusServer struct {
	grpc.ServerStream
}

func (x *controlServiceWatchSomaticStatusServer) Send(m *StatusResponse) error {
	return x.ServerStream.SendMsg(m)
}

// RegisterControlServiceServer registers the service with a gRPC server.
func RegisterControlServiceServer(s grpc.ServiceRegistrar, srv ControlServiceServer) {
	s.RegisterService(&ControlService_ServiceDesc, srv)
}

var ControlService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gophership.protocol.v1.ControlService",
	HandlerType: (*ControlServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _ControlService_Ping_Handler,
		},
		{
			MethodName: "GetSomaticStatus",
			Handler:    _ControlService_GetSomaticStatus_Handler,
		},
		{
			MethodName: "OverrideSomaticZone",
			Handler:    _ControlService_OverrideSomaticZone_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "WatchSomaticStatus",
			Handler:       _ControlService_WatchSomaticStatus_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/protocol/control.proto",
}

func _ControlService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gophership.protocol.v1.ControlService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_GetSomaticStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).GetSomaticStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gophership.protocol.v1.ControlService/GetSomaticStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).GetSomaticStatus(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_WatchSomaticStatus_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(WatchStatusRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ControlServiceServer).WatchSomaticStatus(m, &controlServiceWatchSomaticStatusServer{stream})
}

func _ControlService_OverrideSomaticZone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OverrideSomaticZoneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).OverrideSomaticZone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gophership.protocol.v1.ControlService/OverrideSomaticZone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).OverrideSomaticZone(ctx, req.(*OverrideSomaticZoneRequest))
	}
	return interceptor(ctx, in, info, handler)
}
