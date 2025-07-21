// cmd/server/main.go
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/health"
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

	// Setup health checks
	healthChecker := health.NewHealthChecker()

	// Add game engine health check
	healthChecker.AddCheck(health.NewGameEngineHealthCheck(
		func() bool { return server.GetGameRunning() },
	))

	// Add network health check
	healthChecker.AddCheck(health.NewNetworkHealthCheck(
		func() string { return server.GetListenerAddress() },
	))

	// Add memory health check (limit: 500MB)
	healthChecker.AddCheck(health.NewMemoryHealthCheck(500, func() int64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return int64(m.Alloc / 1024 / 1024)
	}))

	// Start health check HTTP server
	healthPort := "8080" // Default health check port
	if envPort := os.Getenv("NETREK_HEALTH_PORT"); envPort != "" {
		if _, err := strconv.Atoi(envPort); err == nil {
			healthPort = envPort
		}
	}

	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", healthChecker.LivenessHandler)
	healthMux.HandleFunc("/ready", healthChecker.ReadinessHandler)

	healthServer := &http.Server{
		Addr:         ":" + healthPort,
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Start health check server in background
	go func() {
		logger.Info(ctx, "Starting health check server",
			"port", healthPort,
		)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "Health check server failed", err)
		}
	}()

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
	
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Shutdown health check server
	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "Health check server shutdown failed", err)
	}
	
	// Stop game server
	server.Stop()
}
