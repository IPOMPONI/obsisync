package server

import (
	"fmt"
	"net/http"
	"obsisync/internal/config"
	"obsisync/internal/middleware"
	"time"

	"golang.org/x/net/webdav"
)

func New(cfg *config.Config) *http.Server {
	mux := http.NewServeMux()

	for prefix, path := range cfg.Vaults {
		wdh := newWebDav(prefix, path)
		mux.Handle(prefix+"/", wdh)
	}

	handler := middleware.Chain(
		mux,
		func(next http.Handler) http.Handler {
			return middleware.AuthMiddleware(cfg.Auth, next)
		},
		middleware.LoggingMiddleware,
	)

	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func newWebDav(prefix, pathDir string) *webdav.Handler {
	return &webdav.Handler{
		Prefix:     prefix,
		FileSystem: webdav.Dir(pathDir),
		LockSystem: webdav.NewMemLS(),
	}
}
