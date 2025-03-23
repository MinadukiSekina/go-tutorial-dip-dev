package test

import (
	"net/http"
)

func Route() *http.ServeMux {
	m := http.NewServeMux()

	m.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/users", http.StatusFound)
	})

	return m
}