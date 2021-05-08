package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmetb/grpcoin/api/grpcoin"
	"github.com/ahmetb/grpcoin/auth/github"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type accountServiceOpts struct {
	redisIP string
}

type accountService struct {
	redis *redis.Client
	grpcoin.UnimplementedAccountServer
}

func newAccountService(opts accountServiceOpts) grpcoin.AccountServer {
	var r *redis.Client
	if opts.redisIP != "" {
		r = redis.NewClient(&redis.Options{
			Addr: opts.redisIP + ":6379",
		})
	} else {
		r, _ = redismock.NewClientMock()
	}
	return &accountService{redis: r}
}

func (s *accountService) createAccount(ctx context.Context) {
	panic("unimplemented")
}

func (s *accountService) TestAuth(ctx context.Context, req *grpcoin.TestAuthRequest) (*grpcoin.TestAuthResponse, error) {
	m, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot parse grpc meta")
	}
	vs, ok := m["authorization"]
	if !ok || len(vs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "you need to provide a GitHub personal access token "+
			"by setting the grpc metadata (header) named 'authorization'")
	}
	v := vs[0]
	if !strings.HasPrefix(v, "Bearer ") {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("token %q must begin with \"Bearer \".", v))
	}

	_, err := github.VerifyUser(v)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &grpcoin.TestAuthResponse{}, nil
}
