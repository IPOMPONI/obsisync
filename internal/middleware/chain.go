package middleware

import (
	"net/http"
)

func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, mdw := range middlewares {
		handler = mdw(handler)
	}

	return handler
}
