// examples/ai_client/main.go
// AI-controlled client bot demonstrating different behaviors
package main

import (
	"flag"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/event"
	"github.com/opd-ai/go-netrek/pkg/network"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// AIBehavior defines different AI behaviors
type AIBehavior int

const (
	BehaviorExplorer  AIBehavior = iota // Flies around exploring
	BehaviorAggressor                   // Seeks and attacks enemies
	BehaviorDefender                    // Defends home planets
	BehaviorBomber                      // Attacks enemy planets
)

// AIClient represents an AI-controlled game client
type AIClient struct {
	client         *network.GameClient
	eventBus       *event.Bus
	behavior       AIBehavior
	playerName     string
	teamID         int
	lastGameState  *engine.GameState
	myShipID       entity.ID
	targetPlanetID entity.ID
	targetShipID   entity.ID
	lastActionTime time.Time
	actionCooldown time.Duration
	random         *rand.Rand
}

func main() {
	// Command line flags
	serverAddr := flag.String("server", "localhost:4566", "Server address")
	playerName := flag.String("name", "AIBot", "Player name")
	teamID := flag.Int("team", 0, "Team ID (0 or 1)")
	behaviorStr := flag.String("behavior", "explorer", "AI behavior (explorer, aggressor, defender, bomber)")
	flag.Parse()

	// Parse behavior
	var behavior AIBehavior
	switch *behaviorStr {
	case "explorer":
		behavior = BehaviorExplorer
	case "aggressor":
		behavior = BehaviorAggressor
	case "defender":
		behavior = BehaviorDefender
	case "bomber":
		behavior = BehaviorBomber
	default:
		log.Fatalf("Unknown behavior: %s", *behaviorStr)
	}

	log.Printf("Starting AI Client: %s (Team %d, Behavior: %s)", *playerName, *teamID, *behaviorStr)

	// Create AI client
	aiClient := NewAIClient(*playerName, *teamID, behavior)

	// Connect to server
	if err := aiClient.Connect(*serverAddr); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	log.Printf("Connected to server at %s", *serverAddr)
	log.Printf("AI behavior: %s", getBehaviorDescription(behavior))

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down AI client...")
	aiClient.Disconnect()
}

// NewAIClient creates a new AI client
func NewAIClient(playerName string, teamID int, behavior AIBehavior) *AIClient {
	eventBus := event.NewEventBus()

	ai := &AIClient{
		client:         network.NewGameClient(eventBus),
		eventBus:       eventBus,
		behavior:       behavior,
		playerName:     playerName,
		teamID:         teamID,
		actionCooldown: time.Millisecond * 200, // Act every 200ms
		random:         rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(teamID))),
	}

	// Subscribe to events
	ai.setupEventHandlers()

	return ai
}

// Connect connects to the game server
func (ai *AIClient) Connect(serverAddr string) error {
	if err := ai.client.Connect(serverAddr, ai.playerName, ai.teamID); err != nil {
		return err
	}

	// Start AI behavior loop
	go ai.behaviorLoop()

	return nil
}

// Disconnect disconnects from the server
func (ai *AIClient) Disconnect() {
	ai.client.Disconnect()
}

// setupEventHandlers sets up event handling
func (ai *AIClient) setupEventHandlers() {
	ai.eventBus.Subscribe(network.ChatMessageReceived, func(e event.Event) {
		if chatEvent, ok := e.(*network.ChatEvent); ok {
			log.Printf("[Chat] %s: %s", chatEvent.SenderName, chatEvent.Message)
		}
	})

	ai.eventBus.Subscribe(network.ClientDisconnected, func(e event.Event) {
		log.Printf("Disconnected from server")
		os.Exit(0)
	})
}

// behaviorLoop runs the main AI behavior loop
func (ai *AIClient) behaviorLoop() {
	// Game state processing
	go func() {
		for gameState := range ai.client.GetGameStateChannel() {
			ai.lastGameState = gameState
			ai.updateMyShipID(gameState)
		}
	}()

	// Main behavior loop
	for {
		time.Sleep(ai.actionCooldown)

		if ai.lastGameState == nil {
			continue
		}

		// Execute behavior based on type
		ai.executeBehavior()
		ai.lastActionTime = time.Now()
	}
}

// updateMyShipID finds and updates the AI's ship ID
func (ai *AIClient) updateMyShipID(gameState *engine.GameState) {
	// Find our ship by looking for ships belonging to our team
	for shipID, shipState := range gameState.Ships {
		if shipState.TeamID == ai.teamID {
			// For simplicity, assume first ship found is ours
			// In a real implementation, we'd track player-to-ship mapping
			ai.myShipID = shipID
			break
		}
	}
}

// executeBehavior executes the AI's current behavior
func (ai *AIClient) executeBehavior() {
	if ai.myShipID == 0 {
		return // Don't have a ship yet
	}

	myShip, exists := ai.lastGameState.Ships[ai.myShipID]
	if !exists {
		return // Ship doesn't exist
	}

	switch ai.behavior {
	case BehaviorExplorer:
		ai.executeExplorerBehavior(myShip)
	case BehaviorAggressor:
		ai.executeAggressorBehavior(myShip)
	case BehaviorDefender:
		ai.executeDefenderBehavior(myShip)
	case BehaviorBomber:
		ai.executeBomberBehavior(myShip)
	}
}

// executeExplorerBehavior makes the AI explore the galaxy
func (ai *AIClient) executeExplorerBehavior(myShip engine.ShipState) {
	// Simple exploration: move in random directions, occasionally change course
	thrust := true
	turnLeft := false
	turnRight := false

	// Randomly change direction
	if ai.random.Float64() < 0.1 { // 10% chance to turn
		if ai.random.Float64() < 0.5 {
			turnLeft = true
		} else {
			turnRight = true
		}
	}

	// Send input
	ai.client.SendInput(thrust, turnLeft, turnRight, -1, false, false, 0, 0)

	// Occasionally send a friendly message
	if ai.random.Float64() < 0.005 { // 0.5% chance
		messages := []string{
			"Exploring the galaxy!",
			"Beautiful stars out here",
			"Anyone seen any interesting planets?",
			"The universe is vast",
		}
		msg := messages[ai.random.IntN(len(messages))]
		ai.client.SendChatMessage(msg)
	}
}

// executeAggressorBehavior makes the AI seek and attack enemies
func (ai *AIClient) executeAggressorBehavior(myShip engine.ShipState) {
	// Find the nearest enemy ship
	nearestEnemy := ai.findNearestEnemyShip(myShip.Position)

	thrust := true
	turnLeft := false
	turnRight := false
	fireWeapon := -1

	if nearestEnemy != nil {
		// Calculate angle to enemy
		targetVector := nearestEnemy.Position.Sub(myShip.Position)
		targetAngle := targetVector.Angle()

		// Calculate angle difference
		angleDiff := targetAngle - myShip.Rotation

		// Normalize angle difference to [-π, π]
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}
		for angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		// Turn towards enemy
		if math.Abs(angleDiff) > 0.1 {
			if angleDiff > 0 {
				turnLeft = true
			} else {
				turnRight = true
			}
		}

		// Fire if pointing roughly at enemy
		distance := targetVector.Length()
		if math.Abs(angleDiff) < 0.3 && distance < 1000 {
			fireWeapon = 0 // Fire first weapon
		}

		// Taunt enemies occasionally
		if ai.random.Float64() < 0.01 {
			ai.client.SendChatMessage("Prepare for battle!")
		}
	}

	ai.client.SendInput(thrust, turnLeft, turnRight, fireWeapon, false, false, 0, 0)
}

// executeDefenderBehavior makes the AI defend home planets
func (ai *AIClient) executeDefenderBehavior(myShip engine.ShipState) {
	homePlanet := ai.findHomePlanet()

	thrust, turnLeft, turnRight := ai.initializeDefenseInputs()

	if homePlanet != nil {
		distance := ai.calculateDistanceToHome(myShip.Position, homePlanet.Position)

		if distance > 500 {
			turnLeft, turnRight = ai.calculateNavigationToHome(myShip, homePlanet)
		} else {
			turnLeft, turnRight = ai.executePatrolBehavior()
		}
	}

	ai.client.SendInput(thrust, turnLeft, turnRight, -1, false, false, 0, 0)
}

// initializeDefenseInputs sets up the default input state for defensive behavior.
func (ai *AIClient) initializeDefenseInputs() (thrust bool, turnLeft bool, turnRight bool) {
	return true, false, false
}

// calculateDistanceToHome computes the distance between the ship and home planet.
func (ai *AIClient) calculateDistanceToHome(shipPosition, homePosition physics.Vector2D) float64 {
	homeVector := homePosition.Sub(shipPosition)
	return homeVector.Length()
}

// calculateNavigationToHome determines turn direction to navigate toward the home planet.
func (ai *AIClient) calculateNavigationToHome(myShip engine.ShipState, homePlanet *engine.PlanetState) (turnLeft bool, turnRight bool) {
	homeVector := homePlanet.Position.Sub(myShip.Position)
	targetAngle := homeVector.Angle()
	angleDiff := ai.normalizeAngleDifference(targetAngle - myShip.Rotation)

	return ai.determineTurnDirection(angleDiff)
}

// normalizeAngleDifference normalizes angle difference to the range [-π, π].
func (ai *AIClient) normalizeAngleDifference(angleDiff float64) float64 {
	for angleDiff > math.Pi {
		angleDiff -= 2 * math.Pi
	}
	for angleDiff < -math.Pi {
		angleDiff += 2 * math.Pi
	}
	return angleDiff
}

// determineTurnDirection decides whether to turn left or right based on angle difference.
func (ai *AIClient) determineTurnDirection(angleDiff float64) (turnLeft bool, turnRight bool) {
	if math.Abs(angleDiff) > 0.1 {
		if angleDiff > 0 {
			return true, false
		}
		return false, true
	}
	return false, false
}

// executePatrolBehavior implements random patrol movement around the home planet.
func (ai *AIClient) executePatrolBehavior() (turnLeft bool, turnRight bool) {
	if ai.random.Float64() < 0.2 {
		if ai.random.Float64() < 0.5 {
			return true, false
		}
		return false, true
	}
	return false, false
}

// executeBomberBehavior makes the AI attack enemy planets
func (ai *AIClient) executeBomberBehavior(myShip engine.ShipState) {
	// Find enemy planets to bomb
	target := ai.findEnemyPlanet(myShip.Position)

	thrust := true
	turnLeft := false
	turnRight := false
	beamDown := false

	if target != nil {
		// Move towards target planet
		targetVector := target.Position.Sub(myShip.Position)
		distance := targetVector.Length()
		targetAngle := targetVector.Angle()

		angleDiff := targetAngle - myShip.Rotation
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}
		for angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		if math.Abs(angleDiff) > 0.1 {
			if angleDiff > 0 {
				turnLeft = true
			} else {
				turnRight = true
			}
		}

		// If close to planet, beam down armies (bomb)
		if distance < 100 && myShip.Armies > 0 {
			beamDown = true
		}
	}

	ai.client.SendInput(thrust, turnLeft, turnRight, -1, beamDown, false, 1, 0)
}

// Helper functions for AI decision making

func (ai *AIClient) findNearestEnemyShip(myPos physics.Vector2D) *engine.ShipState {
	var nearest *engine.ShipState
	var nearestDistance float64 = math.Inf(1)

	for _, ship := range ai.lastGameState.Ships {
		if ship.TeamID != ai.teamID {
			distance := ship.Position.Distance(myPos)
			if distance < nearestDistance {
				nearest = &ship
				nearestDistance = distance
			}
		}
	}

	return nearest
}

func (ai *AIClient) findHomePlanet() *engine.PlanetState {
	// Look for planets owned by our team - use planet names to identify home planets
	for _, planet := range ai.lastGameState.Planets {
		if planet.TeamID == ai.teamID {
			// Common home planet names
			if planet.Name == "Earth" || planet.Name == "Qo'noS" ||
				planet.Name == "Romulus" || planet.Name == "Cardassia" {
				return &planet
			}
		}
	}

	// If no home planet found by name, return any planet owned by our team
	for _, planet := range ai.lastGameState.Planets {
		if planet.TeamID == ai.teamID {
			return &planet
		}
	}

	return nil
}

func (ai *AIClient) findEnemyPlanet(myPos physics.Vector2D) *engine.PlanetState {
	var target *engine.PlanetState
	var nearestDistance float64 = math.Inf(1)

	for _, planet := range ai.lastGameState.Planets {
		if planet.TeamID != ai.teamID && planet.TeamID >= 0 {
			distance := planet.Position.Distance(myPos)
			if distance < nearestDistance {
				target = &planet
				nearestDistance = distance
			}
		}
	}

	return target
}

func getBehaviorDescription(behavior AIBehavior) string {
	switch behavior {
	case BehaviorExplorer:
		return "Explores the galaxy peacefully"
	case BehaviorAggressor:
		return "Seeks out and attacks enemy ships"
	case BehaviorDefender:
		return "Defends team's home planets"
	case BehaviorBomber:
		return "Attacks enemy planets"
	default:
		return "Unknown"
	}
}
