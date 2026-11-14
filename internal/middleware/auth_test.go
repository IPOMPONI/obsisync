package middleware

import (
	"net/http"
	"net/http/httptest"
	"obsisync/internal/config"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	auth := config.Auth{Username: "test_user", Password: "test_pass"}

	tests := []struct {
		name       string
		username   string
		password   string
		authHeader string
		sendAuth   bool
		wantStatus int
		wantNext   bool
	}{
		{
			name:       "valid credentials",
			username:   "test_user",
			password:   "test_pass",
			sendAuth:   true,
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
		{
			name:       "missing Authorization",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong username",
			username:   "admin",
			password:   "test_pass",
			sendAuth:   true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong password",
			username:   "test_user",
			password:   "admin_pass",
			sendAuth:   true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong username and password",
			username:   "admin",
			password:   "admin_pass",
			sendAuth:   true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-Basic scheme",
			authHeader: "Bearer some-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed Basic",
			authHeader: "Basic 1@!-af=",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				nextCalled = true
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("PROPFIND", "/notes/", nil)

			switch {
			case tt.authHeader != "":
				req.Header.Set("Authorization", tt.authHeader)
			case tt.sendAuth:
				req.SetBasicAuth(tt.username, tt.password)
			}

			AuthMiddleware(auth, next).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if nextCalled != tt.wantNext {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNext)
			}
		})
	}
}

func TestAuthSetErrHeader(t *testing.T) {
	t.Parallel()

	auth := config.Auth{Username: "test_user", Password: "test_pass"}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PROPFIND", "/notes/", nil)

	AuthMiddleware(auth, next).ServeHTTP(rec, req)

	wantAuth := `Basic realm="obsisync"`
	if got := rec.Header().Get("WWW-Authenticate"); got != wantAuth {
		t.Errorf("WWW-Authenticate = %q, want %q", got, wantAuth)
	}

	wantBody := "Unauthorized\n"
	if got := rec.Body.String(); got != wantBody {
		t.Errorf("body = %q, want %q", got, wantBody)
	}
}

func TestAuthSetErrHeaderEmpty(t *testing.T) {
	t.Parallel()

	auth := config.Auth{Username: "test_user", Password: "test_pass"}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PROPFIND", "/notes/", nil)
	req.SetBasicAuth(auth.Username, auth.Password)

	AuthMiddleware(auth, next).ServeHTTP(rec, req)

	if got := rec.Header().Get("WWW-Authenticate"); got != "" {
		t.Errorf("WWW-Authenticate = %q, want empty on success", got)
	}
}

func TestAuthMiddlewarePassResponse(t *testing.T) {
	t.Parallel()

	auth := config.Auth{Username: "test_user", Password: "test_pass"}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Vault", "notes")
		w.WriteHeader(http.StatusMultiStatus)
		w.Write([]byte("body content"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PROPFIND", "/notes/", nil)
	req.SetBasicAuth(auth.Username, auth.Password)

	AuthMiddleware(auth, next).ServeHTTP(rec, req)

	if rec.Code != http.StatusMultiStatus {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMultiStatus)
	}

	wantHeader := "notes"
	if got := rec.Header().Get("X-Vault"); got != wantHeader {
		t.Errorf("X-Vault = %q, want %q", got, wantHeader)
	}

	wantBody := "body content"
	if got := rec.Body.String(); got != wantBody {
		t.Errorf("body = %q, want %q", got, wantBody)
	}
}
