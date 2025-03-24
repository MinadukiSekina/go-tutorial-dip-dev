package test

import (
	"net/http"
)

type Handler struct {
	Path    string
	Handler http.HandlerFunc
}

func Route(handlers ...Handler) *http.ServeMux {
	m := http.NewServeMux()

	for _, h := range handlers {
		m.HandleFunc(h.Path, h.Handler)
	}

	return m
}