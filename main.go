package main

import (
	"log"
	"net/http"

	"github.com/dip-dev/go-tutorial/internal/chapter1"
	"github.com/dip-dev/go-tutorial/internal/chapter2"
)

func main() {
	mux := http.NewServeMux()

	// EchoAPI
	mux.HandleFunc("/echo", chapter1.GetEcho)

	// FIXME: ハンドラ追加時はこちらにコードを追加してください
	mux.HandleFunc("/users", chapter2.Get)
	mux.HandleFunc("/users", chapter2.Create)

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		log.Fatalf("failed to launch service: %+v", err)
	}
}
