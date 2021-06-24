// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package grpcoin

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TickerInfoClient is the client API for TickerInfo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TickerInfoClient interface {
	// Watch returns real-time quotes of the ticker.
	// The only supported tickers are "BTC", "ETH", "BNB", "DOGE", "DOT".
	//
	// This stream terminates after 15 minutes, so expect being
	// abruptly disconnected and need to reconnect.
	//
	// No authentication required.
	Watch(ctx context.Context, in *QuoteTicker, opts ...grpc.CallOption) (TickerInfo_WatchClient, error)
}

type tickerInfoClient struct {
	cc grpc.ClientConnInterface
}

func NewTickerInfoClient(cc grpc.ClientConnInterface) TickerInfoClient {
	return &tickerInfoClient{cc}
}

func (c *tickerInfoClient) Watch(ctx context.Context, in *QuoteTicker, opts ...grpc.CallOption) (TickerInfo_WatchClient, error) {
	stream, err := c.cc.NewStream(ctx, &TickerInfo_ServiceDesc.Streams[0], "/grpcoin.TickerInfo/Watch", opts...)
	if err != nil {
		return nil, err
	}
	x := &tickerInfoWatchClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type TickerInfo_WatchClient interface {
	Recv() (*Quote, error)
	grpc.ClientStream
}

type tickerInfoWatchClient struct {
	grpc.ClientStream
}

func (x *tickerInfoWatchClient) Recv() (*Quote, error) {
	m := new(Quote)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TickerInfoServer is the server API for TickerInfo service.
// All implementations must embed UnimplementedTickerInfoServer
// for forward compatibility
type TickerInfoServer interface {
	// Watch returns real-time quotes of the ticker.
	// The only supported tickers are "BTC", "ETH", "BNB", "DOGE", "DOT".
	//
	// This stream terminates after 15 minutes, so expect being
	// abruptly disconnected and need to reconnect.
	//
	// No authentication required.
	Watch(*QuoteTicker, TickerInfo_WatchServer) error
	mustEmbedUnimplementedTickerInfoServer()
}

// UnimplementedTickerInfoServer must be embedded to have forward compatible implementations.
type UnimplementedTickerInfoServer struct {
}

func (UnimplementedTickerInfoServer) Watch(*QuoteTicker, TickerInfo_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
func (UnimplementedTickerInfoServer) mustEmbedUnimplementedTickerInfoServer() {}

// UnsafeTickerInfoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TickerInfoServer will
// result in compilation errors.
type UnsafeTickerInfoServer interface {
	mustEmbedUnimplementedTickerInfoServer()
}

func RegisterTickerInfoServer(s grpc.ServiceRegistrar, srv TickerInfoServer) {
	s.RegisterService(&TickerInfo_ServiceDesc, srv)
}

func _TickerInfo_Watch_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(QuoteTicker)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TickerInfoServer).Watch(m, &tickerInfoWatchServer{stream})
}

type TickerInfo_WatchServer interface {
	Send(*Quote) error
	grpc.ServerStream
}

type tickerInfoWatchServer struct {
	grpc.ServerStream
}

func (x *tickerInfoWatchServer) Send(m *Quote) error {
	return x.ServerStream.SendMsg(m)
}

// TickerInfo_ServiceDesc is the grpc.ServiceDesc for TickerInfo service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TickerInfo_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpcoin.TickerInfo",
	HandlerType: (*TickerInfoServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Watch",
			Handler:       _TickerInfo_Watch_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "grpcoin.proto",
}

// AccountClient is the client API for Account service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AccountClient interface {
	// Tests if your token works.
	//
	// Send a header (gRPC metadata) named "Authorization"
	// with value "Bearer XXX" where XXX is a GitHub Personal Access token
	// from https://github.com/settings/tokens (no permissions needed).
	TestAuth(ctx context.Context, in *TestAuthRequest, opts ...grpc.CallOption) (*TestAuthResponse, error)
}

type accountClient struct {
	cc grpc.ClientConnInterface
}

func NewAccountClient(cc grpc.ClientConnInterface) AccountClient {
	return &accountClient{cc}
}

func (c *accountClient) TestAuth(ctx context.Context, in *TestAuthRequest, opts ...grpc.CallOption) (*TestAuthResponse, error) {
	out := new(TestAuthResponse)
	err := c.cc.Invoke(ctx, "/grpcoin.Account/TestAuth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccountServer is the server API for Account service.
// All implementations must embed UnimplementedAccountServer
// for forward compatibility
type AccountServer interface {
	// Tests if your token works.
	//
	// Send a header (gRPC metadata) named "Authorization"
	// with value "Bearer XXX" where XXX is a GitHub Personal Access token
	// from https://github.com/settings/tokens (no permissions needed).
	TestAuth(context.Context, *TestAuthRequest) (*TestAuthResponse, error)
	mustEmbedUnimplementedAccountServer()
}

// UnimplementedAccountServer must be embedded to have forward compatible implementations.
type UnimplementedAccountServer struct {
}

func (UnimplementedAccountServer) TestAuth(context.Context, *TestAuthRequest) (*TestAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TestAuth not implemented")
}
func (UnimplementedAccountServer) mustEmbedUnimplementedAccountServer() {}

// UnsafeAccountServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AccountServer will
// result in compilation errors.
type UnsafeAccountServer interface {
	mustEmbedUnimplementedAccountServer()
}

func RegisterAccountServer(s grpc.ServiceRegistrar, srv AccountServer) {
	s.RegisterService(&Account_ServiceDesc, srv)
}

func _Account_TestAuth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestAuthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountServer).TestAuth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcoin.Account/TestAuth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountServer).TestAuth(ctx, req.(*TestAuthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Account_ServiceDesc is the grpc.ServiceDesc for Account service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Account_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpcoin.Account",
	HandlerType: (*AccountServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "TestAuth",
			Handler:    _Account_TestAuth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpcoin.proto",
}

// PaperTradeClient is the client API for PaperTrade service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PaperTradeClient interface {
	// Returns authenticated user's portfolio.
	Portfolio(ctx context.Context, in *PortfolioRequest, opts ...grpc.CallOption) (*PortfolioResponse, error)
	// Executes a trade in authenticated user's portfolio.
	// All trades are executed immediately with the real-time market
	// price provided on TickerInfo.Watch endpoint.
	Trade(ctx context.Context, in *TradeRequest, opts ...grpc.CallOption) (*TradeResponse, error)
}

type paperTradeClient struct {
	cc grpc.ClientConnInterface
}

func NewPaperTradeClient(cc grpc.ClientConnInterface) PaperTradeClient {
	return &paperTradeClient{cc}
}

func (c *paperTradeClient) Portfolio(ctx context.Context, in *PortfolioRequest, opts ...grpc.CallOption) (*PortfolioResponse, error) {
	out := new(PortfolioResponse)
	err := c.cc.Invoke(ctx, "/grpcoin.PaperTrade/Portfolio", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paperTradeClient) Trade(ctx context.Context, in *TradeRequest, opts ...grpc.CallOption) (*TradeResponse, error) {
	out := new(TradeResponse)
	err := c.cc.Invoke(ctx, "/grpcoin.PaperTrade/Trade", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PaperTradeServer is the server API for PaperTrade service.
// All implementations must embed UnimplementedPaperTradeServer
// for forward compatibility
type PaperTradeServer interface {
	// Returns authenticated user's portfolio.
	Portfolio(context.Context, *PortfolioRequest) (*PortfolioResponse, error)
	// Executes a trade in authenticated user's portfolio.
	// All trades are executed immediately with the real-time market
	// price provided on TickerInfo.Watch endpoint.
	Trade(context.Context, *TradeRequest) (*TradeResponse, error)
	mustEmbedUnimplementedPaperTradeServer()
}

// UnimplementedPaperTradeServer must be embedded to have forward compatible implementations.
type UnimplementedPaperTradeServer struct {
}

func (UnimplementedPaperTradeServer) Portfolio(context.Context, *PortfolioRequest) (*PortfolioResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Portfolio not implemented")
}
func (UnimplementedPaperTradeServer) Trade(context.Context, *TradeRequest) (*TradeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Trade not implemented")
}
func (UnimplementedPaperTradeServer) mustEmbedUnimplementedPaperTradeServer() {}

// UnsafePaperTradeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PaperTradeServer will
// result in compilation errors.
type UnsafePaperTradeServer interface {
	mustEmbedUnimplementedPaperTradeServer()
}

func RegisterPaperTradeServer(s grpc.ServiceRegistrar, srv PaperTradeServer) {
	s.RegisterService(&PaperTrade_ServiceDesc, srv)
}

func _PaperTrade_Portfolio_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PortfolioRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaperTradeServer).Portfolio(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcoin.PaperTrade/Portfolio",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaperTradeServer).Portfolio(ctx, req.(*PortfolioRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaperTrade_Trade_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TradeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaperTradeServer).Trade(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcoin.PaperTrade/Trade",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaperTradeServer).Trade(ctx, req.(*TradeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PaperTrade_ServiceDesc is the grpc.ServiceDesc for PaperTrade service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PaperTrade_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpcoin.PaperTrade",
	HandlerType: (*PaperTradeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Portfolio",
			Handler:    _PaperTrade_Portfolio_Handler,
		},
		{
			MethodName: "Trade",
			Handler:    _PaperTrade_Trade_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpcoin.proto",
}
