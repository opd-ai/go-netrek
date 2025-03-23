// cmd/server/main.go
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/network"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	createDefault := flag.Bool("default", false, "Create default configuration file")
	flag.Parse()

	// Create default configuration file if requested
	if *createDefault {
		defaultConfig := config.DefaultConfig()
		if err := config.SaveConfig(defaultConfig, *configPath); err != nil {
			log.Fatalf("Failed to create default configuration: %v", err)
		}
		log.Printf("Created default configuration file: %s", *configPath)
		return
	}

	// Load configuration
	var gameConfig *config.GameConfig

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Printf("Configuration file not found, using default configuration")
		gameConfig = config.DefaultConfig()
	} else {
		gameConfig, err = config.LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	// Create game
	game := engine.NewGame(gameConfig)

	// Create server
	server := network.NewGameServer(game, gameConfig.MaxPlayers)

	// Start server
	serverAddr := gameConfig.NetworkConfig.ServerAddress
	if serverAddr == "" {
		serverAddr = "localhost:4566"
	}

	log.Printf("Starting server on %s", serverAddr)
	if err := server.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down server...")
	server.Stop()
}
