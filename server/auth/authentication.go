package auth

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

var (
	ctxAuthUserInfo = struct{}{} // stores authenticated user info (e.g. github profile)
	CtxUserRecord   = struct{}{} // stores user firestore record
)

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
		ctx := context.WithValue(rpcCtx, ctxAuthUserInfo, user)
		return ctx, nil
	}
}

// AuthInfoFromContext extracts authenticated user info from the ctx.
func AuthInfoFromContext(ctx context.Context) AuthenticatedUser {
	v := ctx.Value(ctxAuthUserInfo)
	if v == nil {
		return nil
	}
	return v.(AuthenticatedUser)
}
