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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/grpcoin/grpcoin/server/auth"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GitHubUser struct {
	ID       uint64 // github user ID
	Username string // github username (can be changed by the user on GitHub)
}

func (g GitHubUser) DBKey() string       { return fmt.Sprintf("github_%v", g.ID) }
func (g GitHubUser) DisplayName() string { return g.Username }
func (g GitHubUser) ProfileURL() string  { return "https://github.com/" + g.Username }

func VerifyUser(token string) (GitHubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GitHubUser{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var v struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&v)
		return GitHubUser{}, fmt.Errorf("github: failed to authenticate (%d): %s", resp.StatusCode, v.Message)
	}
	var user struct {
		Login string `json:"login"`
		Id    uint64 `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&user)
	return GitHubUser{ID: user.Id, Username: user.Login}, err
}

// GitHubAuthenticator authentices with GitHub personal access token in the
// Authorization header (bearer token format).
type GitHubAuthenticator struct {
	T     trace.Tracer
	Cache *redis.Client
}

func (a *GitHubAuthenticator) Authenticate(ctx context.Context) (auth.AuthenticatedUser, error) {
	m, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot parse grpc metadata")
	}
	vs, ok := m["authorization"]
	if !ok || len(vs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "you need to provide a GitHub personal access token "+
			"from https://github.com/settings/tokens and by setting it on the grpc metadata (header) named 'authorization' "+
			"e.g. authorization: Bearer TOKEN")
	}
	v := vs[0]
	if !strings.HasPrefix(v, "Bearer ") {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("token %q must begin with \"Bearer \".", v))
	}
	v = strings.TrimPrefix(v, "bearer ")
	v = strings.TrimPrefix(v, "Bearer ")

	// TODO make use of the redis cache for the GH API call responses
	_, s := a.T.Start(ctx, "token cache read")
	u, ok, err := a.tokenCached(ctx, v)
	if ok {
		return u, nil
	} else if err != nil {
		ctxzap.Extract(ctx).Warn("redis read fail", zap.Error(err))
	}
	s.End()

	_, s = a.T.Start(ctx, "github auth")
	u, err = VerifyUser(v)
	if err != nil {
		s.End()
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("token denied: %s", err))
	}
	s.End()

	_, s = a.T.Start(ctx, "cache gh token")
	defer s.End()
	err = a.cacheToken(ctx, v, u)
	if err != nil {
		ctxzap.Extract(ctx).Warn("redis set fail", zap.Error(err))
	}
	return u, nil
}

func (g *GitHubAuthenticator) tokenCached(ctx context.Context, tok string) (GitHubUser, bool, error) {
	b, err := g.Cache.Get(ctx, tokenCacheHash(tok)).Bytes()
	if err != nil || b == nil { // added b==nil because of local dummyCache always returning success&empty
		if err == redis.Nil {
			return GitHubUser{}, false, nil
		}
		return GitHubUser{}, false, err
	}
	var v GitHubUser
	err = json.Unmarshal(b, &v)
	return v, true, nil
}

func (g *GitHubAuthenticator) cacheToken(ctx context.Context, tok string, v GitHubUser) error {
	b, _ := json.Marshal(v)
	return g.Cache.Set(ctx, tokenCacheHash(tok), b, 0).Err()
}

func tokenCacheHash(tok string) string {
	h := sha256.New()
	h.Write([]byte(tok))
	return fmt.Sprintf("ghtoken_v1_%x", h.Sum(nil))
}
