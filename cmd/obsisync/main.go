package main

import (
	"flag"
	"log/slog"
	"obsisync/internal/config"
	"obsisync/internal/server"
	"obsisync/internal/startup"
	"os"
)

func main() {
	startup.PrintBanner()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	}))
	slog.SetDefault(logger)

	var configPath string
	flag.StringVar(&configPath, "config-path", "", "path to config file")
	flag.Parse()

	configPath, err := config.ResolvePath(configPath)
	if err != nil {
		slog.Error("failed resolve path", "error", err)
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed load config", "error", err)
		os.Exit(1)
	}

	startup.PrintInfo(*cfg)

	sv := server.New(cfg)

	if err := sv.ListenAndServe(); err != nil {
		slog.Error("failed start server", "error", err)
		os.Exit(1)
	}
}
