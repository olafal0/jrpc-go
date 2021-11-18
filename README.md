# jrpc-go

`jrpc-go` lets you create a generated JSON API server by just specifying handler methods. For example, you can write:

```go
package service

//go:generate jrpc-go -genpkg handlers -receiver Service

type Service struct{}

type User struct {
  ID       string
  Username string
}

func (s *Service) CreateUser(ctx context.Context, username string) (*User, error) {
  // <Insert a user into a database here>
  return &User{
    ID:       "random-uuid",
    Username: username,
  }, nil
}
```

...and run `go generate`. Any exported functions that you define in the package that you run `jrpc-go` for will be wrapped in HTTP request handlers.

The generated code (which you can see in `example/handlers`) will:

- Expose a function that creates a new `http.ServeMux` with handlers for each of your exported methods (e.g. `/CreateUser`)
- Create input structures for each method (e.g. `struct{Username string}`)
- Automatically unmarshal JSON from requests and marshal response structures
- Write an error in the response if your handler errors

You can use the generated code like this:

```go
func main() {
  svc := &example.Service{}
  mux := handlers.HTTPHandler(svc)
  log.Fatal(http.ListenAndServe(":7744", mux))
}
```
