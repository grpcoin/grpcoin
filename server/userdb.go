package main

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	fsUserCol = "users"
)

type AuthenticatedUser interface {
	DBKey() string
	DisplayName() string
	ProfileURL() string
}

type User struct {
	ID          string
	DisplayName string
	ProfileURL  string
}

type userDB struct {
	fs *firestore.Client
}

func (u *userDB) create(ctx context.Context, au AuthenticatedUser) error {
	newUser := User{ID: au.DBKey(),
		DisplayName: au.DisplayName(),
		ProfileURL:  au.ProfileURL(),
	}
	_, err := u.fs.Collection(fsUserCol).Doc(au.DBKey()).Create(ctx, newUser)
	return err
}

func (u *userDB) get(ctx context.Context, au AuthenticatedUser) (User, bool, error) {
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
