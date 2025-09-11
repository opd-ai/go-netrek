// cmd/client/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/EngoEngine/engo"
	"github.com/sirupsen/logrus"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/event"
	"github.com/opd-ai/go-netrek/pkg/network"
	engorender "github.com/opd-ai/go-netrek/pkg/render/engo"
)

// getMainCallerInfo returns the calling function name for logging context
func getMainCallerInfo() string {
	if pc, _, _, ok := runtime.Caller(1); ok {
		return runtime.FuncForPC(pc).Name()
	}
	return "unknown"
}

func main() {
	// Initialize logger with caller reporting
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetReportCaller(true)
	ctx := context.Background()

	caller := getMainCallerInfo()
	logger.WithField("caller", caller).WithField("function", "main").Info("Starting netrek client")

	logger.WithField("caller", caller).WithField("function", "main").Info("Parsing command line arguments")
	args := parseCommandLineArguments(logger, ctx)

	logger.WithField("caller", caller).WithField("function", "main").Info("Loading game configuration")
	gameConfig := loadGameConfiguration(logger, ctx, args.configPath)

	logger.WithField("caller", caller).WithField("function", "main").Info("Resolving server address")
	serverAddr := resolveServerAddress(logger, ctx, args.serverAddr, gameConfig)

	logger.WithField("caller", caller).WithField("function", "main").Info("Creating event bus")
	eventBus := event.NewEventBus()

	logger.WithField("caller", caller).WithField("function", "main").Info("Initializing game client")
	client := initializeGameClient(eventBus, serverAddr, args.playerName, args.teamID)

	logger.WithField("caller", caller).WithField("function", "main").Info("Setting up event subscriptions")
	setupEventSubscriptions(eventBus)

	logger.WithField("caller", caller).WithField("function", "main").Info("Starting selected renderer")
	startSelectedRenderer(args.renderer, client, eventBus, args.teamID, args.width, args.height, args.fullscreen)

	logger.WithField("caller", caller).WithField("function", "main").Info("Client shutdown completed")
}

// clientArgs holds parsed command line arguments for the client application.
type clientArgs struct {
	configPath string
	serverAddr string
	playerName string
	teamID     int
	renderer   string
	fullscreen bool
	width      int
	height     int
}

// parseCommandLineArguments parses and returns command line arguments for the client.
func parseCommandLineArguments(logger *logrus.Logger, ctx context.Context) *clientArgs {
	caller := getMainCallerInfo()
	logger.WithField("caller", caller).WithField("function", "parseCommandLineArguments").Info("Parsing command line flags")

	args := &clientArgs{}
	flag.StringVar(&args.configPath, "config", "config.json", "Path to configuration file")
	flag.StringVar(&args.serverAddr, "server", "", "Server address (overrides config)")
	flag.StringVar(&args.playerName, "name", "Player", "Player name")
	flag.IntVar(&args.teamID, "team", 0, "Team ID")
	flag.StringVar(&args.renderer, "renderer", "terminal", "Renderer type: 'terminal' or 'engo'")
	flag.BoolVar(&args.fullscreen, "fullscreen", false, "Run in fullscreen mode (Engo only)")
	flag.IntVar(&args.width, "width", 1024, "Window width (Engo only)")
	flag.IntVar(&args.height, "height", 768, "Window height (Engo only)")
	flag.Parse()

	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":    "parseCommandLineArguments",
		"config_path": args.configPath,
		"server_addr": args.serverAddr,
		"player_name": args.playerName,
		"team_id":     args.teamID,
		"renderer":    args.renderer,
		"fullscreen":  args.fullscreen,
		"width":       args.width,
		"height":      args.height,
	}).Info("Command line arguments parsed successfully")

	return args
}

// loadGameConfiguration loads the game configuration from file or returns default configuration.
func loadGameConfiguration(logger *logrus.Logger, ctx context.Context, configPath string) *config.GameConfig {
	caller := getMainCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":    "loadGameConfiguration",
		"config_path": configPath,
	}).Info("Loading game configuration")

	var gameConfig *config.GameConfig

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.WithField("caller", caller).WithFields(logrus.Fields{
			"function":    "loadGameConfiguration",
			"config_path": configPath,
		}).Warn("Configuration file not found, using default configuration")
		gameConfig = config.DefaultConfig()
	} else {
		var err error
		gameConfig, err = config.LoadConfig(configPath)
		if err != nil {
			logger.WithField("caller", caller).WithFields(logrus.Fields{
				"function":    "loadGameConfiguration",
				"config_path": configPath,
				"error":       err.Error(),
			}).Fatal("Failed to load configuration")
		}
		logger.WithField("caller", caller).WithFields(logrus.Fields{
			"function":    "loadGameConfiguration",
			"config_path": configPath,
		}).Info("Configuration loaded successfully from file")
	}

	// Apply environment variable overrides
	logger.WithField("caller", caller).WithField("function", "loadGameConfiguration").Info("Applying environment variable overrides")
	if err := config.ApplyEnvironmentOverrides(gameConfig); err != nil {
		logger.WithField("caller", caller).WithFields(logrus.Fields{
			"function": "loadGameConfiguration",
			"error":    err.Error(),
		}).Fatal("Failed to apply environment configuration")
	}

	logger.WithField("caller", caller).WithField("function", "loadGameConfiguration").Info("Game configuration loaded and configured successfully")
	return gameConfig
}

// resolveServerAddress determines the final server address to use for connection.
func resolveServerAddress(logger *logrus.Logger, ctx context.Context, cmdLineAddr string, gameConfig *config.GameConfig) string {
	caller := getMainCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":      "resolveServerAddress",
		"cmd_line_addr": cmdLineAddr,
	}).Info("Resolving server address")

	// Command line argument takes highest priority
	if cmdLineAddr != "" {
		logger.WithField("caller", caller).WithFields(logrus.Fields{
			"function": "resolveServerAddress",
			"source":   "command_line",
			"address":  cmdLineAddr,
		}).Info("Using command line server address")
		return cmdLineAddr
	}

	// Check environment variable
	envAddr := os.Getenv("NETREK_SERVER_ADDR")
	if envAddr != "" {
		port := os.Getenv("NETREK_SERVER_PORT")
		if port == "" {
			port = "4566" // Default port
		}
		return fmt.Sprintf("%s:%s", envAddr, port)
	}

	// Fall back to config file
	if gameConfig.NetworkConfig.ServerAddress != "" {
		return gameConfig.NetworkConfig.ServerAddress
	}

	// Final fallback to localhost
	return "localhost:4566"
}

// initializeGameClient creates and connects a new game client to the server.
func initializeGameClient(eventBus *event.Bus, serverAddr, playerName string, teamID int) *network.GameClient {
	client := network.NewGameClient(eventBus)

	log.Printf("Connecting to server at %s", serverAddr)
	if err := client.Connect(serverAddr, playerName, teamID); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	log.Printf("Connected to server")

	return client
}

// setupEventSubscriptions configures all event handlers for client-server communication.
func setupEventSubscriptions(eventBus *event.Bus) {
	eventBus.Subscribe(network.ChatMessageReceived, func(e event.Event) {
		if chatEvent, ok := e.(*network.ChatEvent); ok {
			fmt.Printf("[%s]: %s\n", chatEvent.SenderName, chatEvent.Message)
		}
	})

	eventBus.Subscribe(network.ClientDisconnected, func(e event.Event) {
		log.Printf("Disconnected from server")
	})

	eventBus.Subscribe(network.ClientReconnected, func(e event.Event) {
		log.Printf("Reconnected to server")
	})

	eventBus.Subscribe(network.ClientReconnectFailed, func(e event.Event) {
		log.Printf("Failed to reconnect to server")
		os.Exit(1)
	})
}

// startSelectedRenderer launches the appropriate renderer based on user selection.
func startSelectedRenderer(renderer string, client *network.GameClient, eventBus *event.Bus, teamID, width, height int, fullscreen bool) {
	switch renderer {
	case "engo":
		startEngoRenderer(client, eventBus, uint64(teamID), width, height, fullscreen)
	case "terminal":
		fallthrough
	default:
		startTerminalRenderer(client, eventBus)
	}
}

// startEngoRenderer starts the Engo GUI client
func startEngoRenderer(client *network.GameClient, eventBus *event.Bus, playerID uint64, width, height int, fullscreen bool) {
	// Create the game scene
	scene := engorender.NewGameScene(client, eventBus, playerID)

	// Configure Engo options
	opts := engo.RunOptions{
		Title:      "Go Netrek",
		Width:      width,
		Height:     height,
		Fullscreen: fullscreen,
		VSync:      true,
	}

	// Run Engo with the game scene
	engo.Run(opts, scene)
}

// startTerminalRenderer starts the terminal-based client (existing implementation)
func startTerminalRenderer(client *network.GameClient, eventBus *event.Bus) {
	// Handle game state updates
	go func() {
		for gameState := range client.GetGameStateChannel() {
			// Process game state
			// In a real client, this would update the game view
			log.Printf("Received game state update: tick=%d ships=%d planets=%d",
				gameState.Tick, len(gameState.Ships), len(gameState.Planets))
		}
	}()

	// Example input simulation (in a real client, this would be based on user input)
	go func() {
		for {
			// Send input every 100ms
			time.Sleep(100 * time.Millisecond)

			// Example input: thrust, turn right, fire weapon 0
			client.SendInput(true, false, true, 0, false, false, 0, 0)
		}
	}()

	// Example chat message
	go func() {
		time.Sleep(3 * time.Second)
		log.Printf("Sending chat message")
		client.SendChatMessage("Hello, Netrek!")
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Disconnecting from server...")
	client.Disconnect()
}
