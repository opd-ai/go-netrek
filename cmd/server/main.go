// cmd/server/main.go
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/logging"
	"github.com/opd-ai/go-netrek/pkg/network"
)

func main() {
	logger := logging.NewLogger()
	ctx := context.Background()

	configPath := flag.String("config", "config.json", "Path to configuration file")
	createDefault := flag.Bool("default", false, "Create default configuration file")
	flag.Parse()

	// Create default configuration file if requested
	if *createDefault {
		defaultConfig := config.DefaultConfig()
		if err := config.SaveConfig(defaultConfig, *configPath); err != nil {
			logger.Error(ctx, "Failed to create default configuration", err,
				"config_path", *configPath,
			)
			os.Exit(1)
		}
		logger.Info(ctx, "Created default configuration file",
			"config_path", *configPath,
		)
		return
	}

	// Load configuration
	var gameConfig *config.GameConfig

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		logger.Info(ctx, "Configuration file not found, using default configuration",
			"config_path", *configPath,
		)
		gameConfig = config.DefaultConfig()
	} else {
		gameConfig, err = config.LoadConfig(*configPath)
		if err != nil {
			logger.Error(ctx, "Failed to load configuration", err,
				"config_path", *configPath,
			)
			os.Exit(1)
		}
	}

	// Apply environment variable overrides
	if err := config.ApplyEnvironmentOverrides(gameConfig); err != nil {
		logger.Error(ctx, "Failed to apply environment configuration", err)
		os.Exit(1)
	}

	// Create game
	game := engine.NewGame(gameConfig)

	// Create server
	server := network.NewGameServer(game, gameConfig.MaxPlayers)

	// Start server
	serverAddr := gameConfig.NetworkConfig.ServerAddress
	if serverAddr == "" {
		logger.Error(ctx, "Server address not configured", nil,
			"message", "Set NETREK_SERVER_ADDR and NETREK_SERVER_PORT environment variables or provide in config file",
		)
		os.Exit(1)
	}

	logger.Info(ctx, "Starting server",
		"address", serverAddr,
		"max_players", gameConfig.MaxPlayers,
	)
	if err := server.Start(serverAddr); err != nil {
		logger.Error(ctx, "Failed to start server", err,
			"address", serverAddr,
		)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info(ctx, "Shutting down server")
	server.Stop()
}
