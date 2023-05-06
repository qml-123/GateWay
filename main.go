package main

import (
	"log"

	"github.com/qml-123/GateWay/http"
)

const (
	configPath = "config/services.json"
)

func main() {
	server := http.NewServer(configPath)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
