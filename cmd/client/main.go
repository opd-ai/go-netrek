// cmd/client/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EngoEngine/engo"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/event"
	"github.com/opd-ai/go-netrek/pkg/network"
	engorender "github.com/opd-ai/go-netrek/pkg/render/engo"
)

func main() {
	args := parseCommandLineArguments()
	gameConfig := loadGameConfiguration(args.configPath)
	serverAddr := resolveServerAddress(args.serverAddr, gameConfig)

	eventBus := event.NewEventBus()
	client := initializeGameClient(eventBus, serverAddr, args.playerName, args.teamID)
	setupEventSubscriptions(eventBus)
	startSelectedRenderer(args.renderer, client, eventBus, args.teamID, args.width, args.height, args.fullscreen)
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
func parseCommandLineArguments() *clientArgs {
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
	return args
}

// loadGameConfiguration loads the game configuration from file or returns default configuration.
func loadGameConfiguration(configPath string) *config.GameConfig {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Configuration file not found, using default configuration")
		return config.DefaultConfig()
	}

	gameConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	return gameConfig
}

// resolveServerAddress determines the final server address to use for connection.
func resolveServerAddress(cmdLineAddr string, gameConfig *config.GameConfig) string {
	if cmdLineAddr == "" {
		return gameConfig.NetworkConfig.ServerAddress
	}
	return cmdLineAddr
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
