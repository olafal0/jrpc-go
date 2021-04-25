package main

import (
	"example"
	"example/handlers"
	"log"
	"net/http"

	// Import to have in go.mod to run local copy of generator directly (with go run)
	_ "golang.org/x/tools/go/packages"
)

func main() {
	svc := &example.Service{}
	mux := handlers.Handler(svc)
	log.Fatal(http.ListenAndServe(":7744", mux))
}
