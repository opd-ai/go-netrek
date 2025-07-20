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
	configPath := flag.String("config", "config.json", "Path to configuration file")
	serverAddr := flag.String("server", "", "Server address (overrides config)")
	playerName := flag.String("name", "Player", "Player name")
	teamID := flag.Int("team", 0, "Team ID")
	renderer := flag.String("renderer", "terminal", "Renderer type: 'terminal' or 'engo'")
	fullscreen := flag.Bool("fullscreen", false, "Run in fullscreen mode (Engo only)")
	width := flag.Int("width", 1024, "Window width (Engo only)")
	height := flag.Int("height", 768, "Window height (Engo only)")
	flag.Parse()

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

	// Use server address from command line if provided
	if *serverAddr == "" {
		*serverAddr = gameConfig.NetworkConfig.ServerAddress
	}

	// Create event bus
	eventBus := event.NewEventBus()

	// Create client
	client := network.NewGameClient(eventBus)

	// Connect to server
	log.Printf("Connecting to server at %s", *serverAddr)
	if err := client.Connect(*serverAddr, *playerName, *teamID); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	log.Printf("Connected to server")

	// Subscribe to events
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

	// Choose renderer based on command line flag
	switch *renderer {
	case "engo":
		// Start Engo GUI renderer
		startEngoRenderer(client, eventBus, uint64(*teamID), *width, *height, *fullscreen)
	case "terminal":
		fallthrough
	default:
		// Start terminal renderer (existing implementation)
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
