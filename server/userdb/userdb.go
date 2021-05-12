package userdb

import (
	"context"
	"fmt"
	"time"

	firestore "cloud.google.com/go/firestore"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpcoin/grpcoin/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ctxUserRecordKey struct{}

const (
	fsUserCol = "users"
)

type User struct {
	ID          string
	DisplayName string
	ProfileURL  string
	CreatedAt   time.Time
}

type UserDB struct {
	DB *firestore.Client
}

func (u *UserDB) Create(ctx context.Context, au auth.AuthenticatedUser) error {
	newUser := User{ID: au.DBKey(),
		DisplayName: au.DisplayName(),
		ProfileURL:  au.ProfileURL(),
		CreatedAt:   time.Now(),
	}
	_, err := u.DB.Collection(fsUserCol).Doc(au.DBKey()).Create(ctx, newUser)
	return err
}

func (u *UserDB) Get(ctx context.Context, au auth.AuthenticatedUser) (User, bool, error) {
	doc, err := u.DB.Collection(fsUserCol).Doc(au.DBKey()).Get(ctx)
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

func (u *UserDB) EnsureAccountExists(ctx context.Context, au auth.AuthenticatedUser) (User, error) {
	// TODO could use lots of caching here
	err := u.Create(ctx, au)
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return User{}, err
	}
	user, _, err := u.Get(ctx, au)
	return user, err

}

// ensureAccountExistsInterceptor creates an account for the authenticated
// client (or retrieves it) and augments the ctx with the user's db record.
func (u *UserDB) EnsureAccountExistsInterceptor() grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		authenticatedUser := auth.AuthInfoFromContext(ctx)
		if authenticatedUser == nil {
			return ctx, status.Error(codes.Internal, "req ctx did not have user info")
		}
		v, ok := authenticatedUser.(auth.AuthenticatedUser)
		if !ok {
			return ctx, status.Errorf(codes.Internal, "unknown authed user type %T", authenticatedUser)
		}
		uv, err := u.EnsureAccountExists(ctx, v)
		if err != nil {
			return ctx, err
		}
		return context.WithValue(ctx, ctxUserRecordKey{}, uv), nil
	}
}

func UserRecordFromContext(ctx context.Context) (User, bool) {
	v := ctx.Value(ctxUserRecordKey{})
	vv, ok := v.(User)
	return vv, ok
}
