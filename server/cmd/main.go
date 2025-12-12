package main

import (
	"log"
	"os"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"github.com/Mahaveer86619/bookture/server/pkg/web"
)

func main() {
	config.LoadConfig()

	srv := web.NewServer()
	if err := srv.Run(); err != nil {
		log.Printf("Server failed to start: %v", err)
		os.Exit(1)
	}
}
