package main

import (
	"context"
	"net"
	"testing"

	"github.com/ahmetb/grpcoin/api/grpcoin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTestAuth(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	grpcoin.RegisterAccountServer(srv, newAccountService(accountServiceOpts{}))
	go srv.Serve(l)
	defer srv.Stop()

	cc, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	client := grpcoin.NewAccountClient(cc)

	_, err = client.TestAuth(context.Background(), &grpcoin.TestAuthRequest{})
	if err == nil {
		t.Fatal("expected err without any creds")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatal("not a grpc status!")
	}
	if s.Code() != codes.Unauthenticated {
		t.Fatalf("got code: %v; expected Unauthenticated", s.Code())
	}

	md := metadata.New(map[string]string{"authorization": "bad format"})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err = client.TestAuth(ctx, &grpcoin.TestAuthRequest{})
	if err == nil {
		t.Fatal("expected err with bad format")
	}
	s, ok = status.FromError(err)
	if !ok {
		t.Fatal("not a grpc status!")
	}
	if s.Code() != codes.InvalidArgument {
		t.Fatalf("got code: %v; expected InvalidArgument", s.Code())
	}

	md = metadata.New(map[string]string{"authorization": "Bearer 123"})
	ctx = metadata.NewOutgoingContext(context.Background(), md)
	_, err = client.TestAuth(ctx, &grpcoin.TestAuthRequest{})
	if err == nil {
		t.Fatal("expected err with bad creds")
	}
	s, ok = status.FromError(err)
	if !ok {
		t.Fatal("not a grpc status!")
	}
	if s.Code() != codes.PermissionDenied {
		t.Fatalf("got code: %v; expected PermissionDenied: %v", s.Code(), err)
	}
}
