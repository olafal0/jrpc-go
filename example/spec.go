package example

import "context"

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
