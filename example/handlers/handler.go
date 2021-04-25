// Code generated by jrpc. DO NOT EDIT.

// Package handlers is an auto-generated package providing HTTP handler
// functions that wrap handlers in the example package.
package handlers

import (
	"encoding/json"
	"example"
	"io/ioutil"
	"net/http"
)

type createUserInput struct {
	Username string `json:"username"`
}

func createUserHandler(recv *example.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST requests supported", http.StatusMethodNotAllowed)
			return
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		input := createUserInput{}
		err = json.Unmarshal(bodyBytes, &input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := recv.CreateUser(
			r.Context(),
			input.Username,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resBytes, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resBytes)
	}
}

func Handler(recv *example.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/CreateUser", createUserHandler(recv))
	return mux
}