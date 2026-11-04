package main

import (
	"flag"
	"log"
	"obsisync/internal/config"
	"obsisync/internal/server"
	"obsisync/internal/startup"
)

func main() {
	startup.PrintBanner()

	var configPath string
	flag.StringVar(&configPath, "config-path", "", "path to config file")
	flag.Parse()

	configPath, err := config.ResolvePath(configPath)
	if err != nil {
		log.Fatalf("resolve path: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	startup.PrintInfo(*cfg)

	sv := server.New(cfg)
	err = sv.ListenAndServe()
	if err != nil {
		log.Fatalf("start server: %v", err)
	}
}
