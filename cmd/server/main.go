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
	"github.com/opd-ai/go-netrek/pkg/resource"
)

func main() {
	logger := logging.NewLogger()
	ctx := context.Background()

	// Parse command line flags and handle default config creation
	configPath := parseCommandLineFlags(logger, ctx)

	// Load and configure the game
	gameConfig := loadGameConfiguration(logger, ctx, configPath)

	// Initialize core game components
	game, server := initializeGameComponents(gameConfig)

	// Setup health monitoring
	healthServer := setupHealthMonitoring(logger, ctx, server, game)

	// Start the game server
	startGameServer(logger, ctx, server, gameConfig)

	// Handle graceful shutdown
	handleGracefulShutdown(logger, ctx, healthServer, server, game)
}

// parseCommandLineFlags parses command line arguments and handles default config creation if requested.
func parseCommandLineFlags(logger *logging.Logger, ctx context.Context) string {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	createDefault := flag.Bool("default", false, "Create default configuration file")
	galaxyTemplate := flag.String("template", "", "Galaxy map template to use (classic_netrek, small_galaxy, balanced_4team)")
	listTemplates := flag.Bool("list-templates", false, "List available galaxy map templates")
	flag.Parse()
	// Handle listing templates
	if *listTemplates {
		logger.Info(ctx, "Available galaxy map templates:")
		templates := config.ListGalaxyTemplates()
		for name, description := range templates {
			logger.Info(ctx, "Template available", "name", name, "description", description)
		}
		os.Exit(0)
	}

	// Handle default config creation with optional template
	if *createDefault {
		var gameConfig *config.GameConfig

		if *galaxyTemplate != "" {
			logger.Info(ctx, "Creating default configuration with galaxy template", "template", *galaxyTemplate)
			var err error
			gameConfig, err = config.LoadConfigWithTemplate("", *galaxyTemplate)
			if err != nil {
				logger.Error(ctx, "Failed to create config with template", err, "template", *galaxyTemplate)
				os.Exit(1)
			}
		} else {
			logger.Info(ctx, "Creating default configuration file")
			gameConfig = config.DefaultConfig()
		}

		if err := config.SaveConfig(gameConfig, *configPath); err != nil {
			logger.Error(ctx, "Failed to create default configuration", err)
			os.Exit(1)
		}
		logger.Info(ctx, "Default configuration created", "path", *configPath)
		os.Exit(0)
	}

	return *configPath
}

// loadGameConfiguration loads the game configuration from file or uses defaults.
func loadGameConfiguration(logger *logging.Logger, ctx context.Context, configPath string) *config.GameConfig {
	var gameConfig *config.GameConfig

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info(ctx, "Configuration file not found, using default configuration",
			"config_path", configPath,
		)
		gameConfig = config.DefaultConfig()
	} else {
		var err error
		gameConfig, err = config.LoadConfig(configPath)
		if err != nil {
			logger.Error(ctx, "Failed to load configuration", err,
				"config_path", configPath,
			)
			os.Exit(1)
		}
	}

	// Apply environment variable overrides
	if err := config.ApplyEnvironmentOverrides(gameConfig); err != nil {
		logger.Error(ctx, "Failed to apply environment configuration", err)
		os.Exit(1)
	}

	return gameConfig
}

// initializeGameComponents creates the core game engine and server components.
func initializeGameComponents(gameConfig *config.GameConfig) (*engine.Game, *network.GameServer) {
	game := engine.NewGame(gameConfig)

	// Initialize resource management
	if err := game.InitializeResourceManager(); err != nil {
		// Log warning but continue - resource management is optional for basic operation
		logger := logging.NewLogger()
		logger.Warn(context.Background(), "Failed to initialize resource manager",
			"error", err)
	}

	server := network.NewGameServer(game, gameConfig.MaxPlayers)
	return game, server
}

// setupHealthMonitoring configures and starts the health check HTTP server.
func setupHealthMonitoring(logger *logging.Logger, ctx context.Context, server *network.GameServer, game *engine.Game) *http.Server {
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

	// Add resource health check if resource manager is available
	if game != nil && game.ResourceManager != nil {
		resourceCheck := resource.NewResourceHealthCheck(game.ResourceManager)
		healthChecker.AddCheck(resourceCheck)
	}

	healthPort := determineHealthPort()
	healthServer := createHealthServer(healthPort, healthChecker)

	// Start health check server in background
	go func() {
		logger.Info(ctx, "Starting health check server",
			"port", healthPort,
		)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "Health check server failed", err)
		}
	}()

	return healthServer
}

// determineHealthPort gets the health check port from environment or uses default.
func determineHealthPort() string {
	healthPort := "8080" // Default health check port
	if envPort := os.Getenv("NETREK_HEALTH_PORT"); envPort != "" {
		if _, err := strconv.Atoi(envPort); err == nil {
			healthPort = envPort
		}
	}
	return healthPort
}

// createHealthServer creates and configures the HTTP server for health checks.
func createHealthServer(healthPort string, healthChecker *health.HealthChecker) *http.Server {
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", healthChecker.LivenessHandler)
	healthMux.HandleFunc("/ready", healthChecker.ReadinessHandler)

	return &http.Server{
		Addr:         ":" + healthPort,
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

// startGameServer validates configuration and starts the game server.
func startGameServer(logger *logging.Logger, ctx context.Context, server *network.GameServer, gameConfig *config.GameConfig) {
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
}

// handleGracefulShutdown waits for shutdown signals and gracefully stops all services.
func handleGracefulShutdown(logger *logging.Logger, ctx context.Context, healthServer *http.Server, server *network.GameServer, game *engine.Game) {
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

	// Shutdown resource manager
	if game != nil && game.ResourceManager != nil {
		if err := game.ResourceManager.Shutdown(shutdownCtx); err != nil {
			logger.Error(ctx, "Resource manager shutdown failed", err)
		}
	}

	// Stop game server
	server.Stop()
}
