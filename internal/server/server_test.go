package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"obsisync/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// read-only paths
var tmpDir string
var notesPath string
var workPath string
var personalPath string

func createVault(vault, filename string) (string, error) {
	var err error

	vaultDir := filepath.Join(tmpDir, vault)
	if err = os.Mkdir(vaultDir, 0o755); err != nil {
		return "", err
	}

	path := filepath.Join(vaultDir, filename)
	if err = os.WriteFile(path, []byte("test_data"), 0o600); err != nil {
		return "", err
	}

	return vaultDir, nil
}

func createConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		Server: config.Server{Host: "127.0.0.1", Port: 8080},
		Auth:   config.Auth{Username: "user", Password: "password"},
		Vaults: make(map[string]string),
	}
}

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = os.MkdirTemp("", "obsisync-tests-*")
	if err != nil {
		fmt.Printf("failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	notesPath, err = createVault("notes", "note.md")
	if err != nil {
		fmt.Printf("notes vault is not created: %s\n", err)
		os.Exit(1)
	}

	workPath, err = createVault("work", "work.md")
	if err != nil {
		fmt.Printf("work vault is not created: %s\n", err)
		os.Exit(1)
	}

	personalPath, err = createVault("personal", "personal.md")
	if err != nil {
		fmt.Printf("personal vault is not created: %s\n", err)
		os.Exit(1)
	}

	code := m.Run()

	os.RemoveAll(tmpDir)

	os.Exit(code)
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Server: config.Server{Host: "127.0.0.1", Port: 8080},
	}

	srv := New(&cfg)

	if srv == nil {
		t.Errorf("New() returned nil")
	}

	if srv.Addr != "127.0.0.1:8080" {
		t.Errorf("Addr = %q, want %q", srv.Addr, "127.0.0.1:8080")
	}

	if srv.Handler == nil {
		t.Fatalf("Handler = nil, want not nil")
	}
}

func TestServerResponse(t *testing.T) {
	t.Parallel()

	cfg := createConfig(t)
	cfg.Vaults["/notes"] = notesPath

	srv := New(cfg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PROPFIND", "/notes/", nil)

	req.SetBasicAuth("user", "password")
	srv.Handler.ServeHTTP(rec, req)

	t.Run("returns 207 Multi-Status", func(t *testing.T) {
		if rec.Code != http.StatusMultiStatus {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusMultiStatus)
		}
	})

	t.Run("response body contains vault contents", func(t *testing.T) {
		if !strings.Contains(rec.Body.String(), "note.md") {
			t.Fatalf("body does not contain 'note.md', got: %s", rec.Body.String())
		}
	})
}

func TestServerAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		login    string
		pass     string
		expected int
	}{
		{
			name:     "valid credentials return 207",
			login:    "user",
			pass:     "password",
			expected: http.StatusMultiStatus,
		},
		{
			name:     "invalid password returns 401",
			login:    "user",
			pass:     "pass",
			expected: http.StatusUnauthorized,
		},
		{
			name:     "empty credentials return 401",
			login:    "",
			pass:     "",
			expected: http.StatusUnauthorized,
		},
	}
	cfg := createConfig(t)
	cfg.Vaults["/notes"] = notesPath

	srv := New(cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("PROPFIND", "/notes/", nil)

			req.SetBasicAuth(tt.login, tt.pass)
			srv.Handler.ServeHTTP(rec, req)

			if rec.Code != tt.expected {
				t.Errorf("status = %d, want %d", rec.Code, tt.expected)
			}
		})
	}

}

func TestServerWithoutAuth(t *testing.T) {
	t.Parallel()

	cfg := createConfig(t)
	cfg.Auth = config.Auth{Username: "", Password: ""}
	cfg.Vaults["/notes"] = notesPath

	srv := New(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PROPFIND", "/notes/", nil)

	srv.Handler.ServeHTTP(rec, req)

	t.Run("returns 401 Unauthorized", func(t *testing.T) {
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("WWW-Authenticate header contains Basic", func(t *testing.T) {
		auth := rec.Header().Get("WWW-Authenticate")
		if !strings.Contains(auth, "Basic") {
			t.Fatalf("WWW-Authenticate = %q, want it to contain 'Basic'", auth)
		}
	})
}

func TestIsolationVaults(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Server: config.Server{Host: "127.0.0.1", Port: 8080},
		Auth:   config.Auth{Username: "user", Password: "password"},
		Vaults: map[string]string{
			"/work":     workPath,
			"/personal": personalPath,
		},
	}

	srv := New(&cfg)

	t.Run("work vault contains work.md", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PROPFIND", "/work/", nil)

		req.SetBasicAuth("user", "password")
		srv.Handler.ServeHTTP(rec, req)

		if !strings.Contains(rec.Body.String(), "work.md") {
			t.Fatalf("body does not contain 'work.md', got: %s", rec.Body.String())
		}
	})

	t.Run("personal vault contains personal.md", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PROPFIND", "/personal/", nil)

		req.SetBasicAuth("user", "password")
		srv.Handler.ServeHTTP(rec, req)

		if !strings.Contains(rec.Body.String(), "personal.md") {
			t.Fatalf("body does not contain 'personal.md', got: %s", rec.Body.String())
		}
	})
}

func TestNewWebDav(t *testing.T) {
	t.Parallel()

	h := newWebDav("/notes", notesPath)

	if h.Prefix != "/notes" {
		t.Errorf("Prefix = %q, want %q", h.Prefix, "/notes")
	}

	if h.FileSystem == nil {
		t.Fatal("FileSystem = nil, want non-nil")
	}

	if h.LockSystem == nil {
		t.Fatal("LockSystem = nil, want non-nil")
	}
}
