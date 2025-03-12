package test

import (
	"errors"
	"net/http"
)

type ErrorResponseWriter struct {
	header     http.Header
	statusCode int
}

func (e *ErrorResponseWriter) Header() http.Header {
	if e.header == nil {
		e.header = http.Header{}
	}
	return e.header
}

func (e *ErrorResponseWriter) Write(_ []byte) (int, error) {
	// 意図的にエラーを返す
	return 0, errors.New("intentional write error")
}

func (e *ErrorResponseWriter) WriteHeader(statusCode int) {
	e.statusCode = statusCode
}

func (e *ErrorResponseWriter) Code() int {
	return e.statusCode
}
