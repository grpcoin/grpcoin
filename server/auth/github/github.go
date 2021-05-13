// Copyright 2021 Ahmet Alp Balkan
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/grpcoin/grpcoin/auth/github"
	"github.com/grpcoin/grpcoin/server/auth"
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

	// TODO make use of the redis cache for the GH API call responses
	u, err := github.VerifyUser(v)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("token denied: %s", err))
	}
	return u, nil
}
