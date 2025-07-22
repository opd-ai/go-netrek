// examples/simple_server/main.go
// Simple server example demonstrating basic Netrek server setup
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/network"
)

func main() {
	// Parse command line arguments
	port, maxClients := parseCommandLineFlags()

	log.Println("Starting Simple Netrek Server...")

	// Initialize and configure the game
	game, gameConfig := initializeGame(port, maxClients)

	// Start the server
	server := startGameServer(game, gameConfig, port)

	// Start background game loop
	startGameLoop(game)

	// Display connection instructions
	displayConnectionInstructions(port)

	// Handle graceful shutdown
	handleGracefulShutdown(server, game)
}

// parseCommandLineFlags parses and returns command line arguments for server configuration.
func parseCommandLineFlags() (string, int) {
	port := flag.String("port", "4566", "Server port")
	maxClients := flag.Int("max-clients", 8, "Maximum number of clients")
	flag.Parse()
	return *port, *maxClients
}

// createSimpleGameConfig creates a basic game configuration suitable for testing
func createSimpleGameConfig(port string, maxClients int) *config.GameConfig {
	return &config.GameConfig{
		WorldSize:  8000,
		MaxPlayers: maxClients,
		Teams: []config.TeamConfig{
			{
				Name:         "Federation",
				Color:        "#0066FF",
				MaxShips:     4,
				StartingShip: "Scout",
			},
			{
				Name:         "Klingons",
				Color:        "#FF6600",
				MaxShips:     4,
				StartingShip: "Scout",
			},
		},
		Planets: []config.PlanetConfig{
			{
				Name:          "Earth",
				X:             -2000,
				Y:             0,
				Type:          entity.Homeworld,
				HomeWorld:     true,
				TeamID:        0,
				InitialArmies: 25,
			},
			{
				Name:          "Qo'noS",
				X:             2000,
				Y:             0,
				Type:          entity.Homeworld,
				HomeWorld:     true,
				TeamID:        1,
				InitialArmies: 25,
			},
			{
				Name:          "Neutral Station",
				X:             0,
				Y:             1500,
				Type:          entity.Industrial,
				HomeWorld:     false,
				TeamID:        -1, // Neutral
				InitialArmies: 15,
			},
			{
				Name:          "Mining Colony",
				X:             0,
				Y:             -1500,
				Type:          entity.Agricultural,
				HomeWorld:     false,
				TeamID:        -1, // Neutral
				InitialArmies: 10,
			},
		},
		PhysicsConfig: config.PhysicsConfig{
			Gravity:         0,
			Friction:        0.05,
			CollisionDamage: 15,
		},
		NetworkConfig: config.NetworkConfig{
			UpdateRate:      20,
			TicksPerState:   2,
			UsePartialState: true,
			ServerPort:      4566,
			ServerAddress:   "localhost:" + port,
		},
		GameRules: config.GameRules{
			WinCondition:   "conquest",
			TimeLimit:      900, // 15 minutes
			MaxScore:       50,
			RespawnDelay:   3,
			FriendlyFire:   false,
			StartingArmies: 0,
		},
	}
}

// initializeGame creates and configures the game engine with the specified parameters.
func initializeGame(port string, maxClients int) (*engine.Game, *config.GameConfig) {
	gameConfig := createSimpleGameConfig(port, maxClients)
	game := engine.NewGame(gameConfig)
	log.Printf("Game created with %d teams and %d planets", len(gameConfig.Teams), len(gameConfig.Planets))
	return game, gameConfig
}

// startGameServer creates and starts the network server for the game.
func startGameServer(game *engine.Game, gameConfig *config.GameConfig, port string) *network.GameServer {
	server := network.NewGameServer(game, gameConfig.MaxPlayers)
	serverAddr := "localhost:" + port

	log.Printf("Starting server on %s", serverAddr)
	if err := server.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	return server
}

// startGameLoop initiates the background game update loop.
func startGameLoop(game *engine.Game) {
	go func() {
		log.Println("Starting game loop...")
		for game.Running {
			game.Update()
			time.Sleep(time.Second / 60) // 60 FPS
		}
	}()
}

// displayConnectionInstructions shows users how to connect clients to the server.
func displayConnectionInstructions(port string) {
	log.Printf("Server is ready! Connect with:")
	log.Printf("  go run examples/ai_client/main.go -server=localhost:%s -name=Player1 -team=0", port)
	log.Printf("  go run examples/ai_client/main.go -server=localhost:%s -name=Player2 -team=1", port)
}

// handleGracefulShutdown manages clean server shutdown on interruption signals.
func handleGracefulShutdown(server *network.GameServer, game *engine.Game) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down server...")
	server.Stop()
	game.Stop()
	log.Println("Server stopped.")
}
