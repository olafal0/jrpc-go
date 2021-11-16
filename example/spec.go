package example

import (
	"context"
	"example/somelib"
)

//go:generate go run ../generate.go -genpath handlers -genpkg handlers -receiver Service

type Service struct{}

type User struct {
	ID       string
	Username string
}

func (s *Service) CreateUser(ctx context.Context, username string) (*User, error) {
	return &User{
		ID:       "random-uuid",
		Username: username,
	}, nil
}

func (s *Service) GetUser(ctx context.Context, userID somelib.UID) (*User, error) {
	return &User{
		ID:       "random-uuid",
		Username: "some user id",
	}, nil
}
