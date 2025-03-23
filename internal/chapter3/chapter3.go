package chapter3

import "net/http"

type Entry struct {
	name   string
	userID int
	salary int
}

// revive:disable

func Get(w http.ResponseWriter, r *http.Request) {
}
