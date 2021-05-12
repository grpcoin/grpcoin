package main

import (
	"context"
	"fmt"
	"time"

	firestore "cloud.google.com/go/firestore"
	"github.com/ahmetb/grpcoin/server/auth"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	fsUserCol = "users"
)

type User struct {
	ID          string
	DisplayName string
	ProfileURL  string
	CreatedAt   time.Time
}

type userDB struct {
	fs *firestore.Client
}

func (u *userDB) create(ctx context.Context, au auth.AuthenticatedUser) error {
	newUser := User{ID: au.DBKey(),
		DisplayName: au.DisplayName(),
		ProfileURL:  au.ProfileURL(),
		CreatedAt:   time.Now(),
	}
	_, err := u.fs.Collection(fsUserCol).Doc(au.DBKey()).Create(ctx, newUser)
	return err
}

func (u *userDB) get(ctx context.Context, au auth.AuthenticatedUser) (User, bool, error) {
	doc, err := u.fs.Collection(fsUserCol).Doc(au.DBKey()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return User{}, false, nil
		}
		return User{}, false, err
	}
	var uv User
	if err := doc.DataTo(&uv); err != nil {
		return User{}, false, fmt.Errorf("failed to unpack user record %q: %w", au.DBKey(), err)
	}
	return uv, true, nil
}

func (u *userDB) ensureAccountExists(ctx context.Context, au auth.AuthenticatedUser) (User, error) {
	// TODO could use lots of caching here
	err := u.create(ctx, au)
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return User{}, err
	}
	user, _, err := u.get(ctx, au)
	return user, err

}

// ensureAccountExistsInterceptor creates an account for the authenticated
// client (or retrieves it) and augments the ctx with the user's db record.
func (u *userDB) ensureAccountExistsInterceptor() grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		authenticatedUser := auth.AuthInfoFromContext(ctx)
		if authenticatedUser == nil {
			return ctx, status.Error(codes.Internal, "req ctx did not have user info")
		}
		v, ok := authenticatedUser.(auth.AuthenticatedUser)
		if !ok {
			return ctx, status.Errorf(codes.Internal, "unknown authed user type %T", authenticatedUser)
		}
		uv, err := u.ensureAccountExists(ctx, v)
		if err != nil {
			return ctx, err
		}
		return context.WithValue(ctx, auth.CtxUserRecord, uv), nil
	}
}
