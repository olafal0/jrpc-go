package example

import (
	"context"
	"encoding/json"
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

// Marshal returns the JSON encoding of v. Defining this method allows custom
// marshaling in jrpc-generated code.
func (s *Service) Marshal(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// Unmarshal also allows custom unmarshaling in jrpc-generated code.
func (s *Service) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
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
