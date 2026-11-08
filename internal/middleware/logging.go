package middleware

import (
	"log/slog"
	"net/http"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
		}

		switch {
		case recorder.status >= 500:
			slog.Error("request failed", attrs...)
		case recorder.status >= 400:
			slog.Warn("request error", attrs...)
		default:
			slog.Info("request completed", attrs...)
		}
	})
}
