package auth

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

type ctxAuthUserInfo struct{} // stores authenticated user info (e.g. github profile)

type AuthenticatedUser interface {
	DBKey() string
	DisplayName() string
	ProfileURL() string
}

type Authenticator interface {
	Authenticate(rpcCtx context.Context) (AuthenticatedUser, error)
}

func AuthenticatingInterceptor(a Authenticator) grpc_auth.AuthFunc {
	return func(rpcCtx context.Context) (context.Context, error) {
		user, err := a.Authenticate(rpcCtx)
		if err != nil {
			return rpcCtx, err
		}
		ctx := context.WithValue(rpcCtx, ctxAuthUserInfo{}, user)
		return ctx, nil
	}
}

// AuthInfoFromContext extracts authenticated user info from the ctx.
func AuthInfoFromContext(ctx context.Context) AuthenticatedUser {
	v := ctx.Value(ctxAuthUserInfo{})
	if v == nil {
		return nil
	}
	return v.(AuthenticatedUser)
}

type MockAuthenticator struct {
	F func(context.Context) (AuthenticatedUser, error)
}

func (m MockAuthenticator) Authenticate(ctx context.Context) (AuthenticatedUser, error) {
	return m.F(ctx)
}
