package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmetb/grpcoin/auth/github"
	"github.com/ahmetb/grpcoin/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GitHubAuthenticator authentices with GitHub personal access token in the
// Authorization header (bearer token format).
type GitHubAuthenticator struct{}

func (a *GitHubAuthenticator) Authenticate(ctx context.Context) (auth.AuthenticatedUser, error) {
	m, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot parse grpc metadata")
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
	v = strings.TrimPrefix(v, "bearer ")
	v = strings.TrimPrefix(v, "Bearer ")

	// TODO make use of the redis cache for the GH API call resp
	u, err := github.VerifyUser(v)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("token denied: %s", err))
	}
	return u, nil
}
