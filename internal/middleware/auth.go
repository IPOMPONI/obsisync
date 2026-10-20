package middleware

import (
	"crypto/subtle"
	"net/http"
	"obsisync/internal/config"
)

func AuthMiddleware(auth config.Auth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			authSetErrHeader(w)
			return
		}

		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(auth.Username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(auth.Password)) == 1

		if !usernameMatch || !passwordMatch {
			authSetErrHeader(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authSetErrHeader(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="obsisync"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
