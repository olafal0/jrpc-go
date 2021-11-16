package example

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

//go:generate go run ../generate.go -genpkg handlers -receiver Service

type Service struct{}

func NewService() *Service {
	return &Service{}
}

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

func (s *Service) GetUser(ctx context.Context, req *http.Request) (*User, error) {
	return &User{
		ID:       uuid.NewString(),
		Username: "some user id",
	}, nil
}
