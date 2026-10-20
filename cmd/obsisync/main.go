package main

import (
	"flag"
	"fmt"
	"log"
	"obsisync/internal/config"
	"obsisync/internal/server"
)

func main() {
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

	fmt.Println("Server started!")
	sv := server.New(cfg)
	err = sv.ListenAndServe()
	if err != nil {
		log.Fatalf("start server: %v", err)
	}

}
