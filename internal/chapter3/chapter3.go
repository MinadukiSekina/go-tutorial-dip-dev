package chapter3

import (
	"context"
	"net/http"

	"github.com/dip-dev/go-tutorial/internal/helper/networking"
)

type User struct {
	ID   int
	Name string
	Age  int
}

type Entry struct {
	Name   string
	UserID int
	Salary int
}

const targetURL = "http://mock-api"

func Get(w http.ResponseWriter, r *http.Request) {
}