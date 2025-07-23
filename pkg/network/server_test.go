package network

import (
	"encoding/json"
	"testing"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

func TestNewGameServer_ConfiguresPartialStateFromConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.NetworkConfig.UsePartialState = false
	cfg.NetworkConfig.TicksPerState = 7
	cfg.NetworkConfig.UpdateRate = 10
	game := engine.NewGame(cfg)
	server := NewGameServer(game, 8)
	if server.partialState != false {
		t.Errorf("expected partialState false, got %v", server.partialState)
	}
	if server.ticksPerState != 7 {
		t.Errorf("expected ticksPerState 7, got %d", server.ticksPerState)
	}
	if server.updateRate != (1e9 / 10) {
		t.Errorf("expected updateRate 1e8ns, got %v", server.updateRate)
	}
}

func TestGameServer_HandlePlayerInput(t *testing.T) {
	cfg := config.DefaultConfig()
	game := engine.NewGame(cfg)

	// Create a test ship
	shipID := entity.ID(1)
	playerID := entity.ID(100)
	teamID := 1
	position := physics.Vector2D{X: 100, Y: 100}

	ship := entity.NewShip(shipID, entity.Scout, teamID, position)
	ship.PlayerID = playerID

	// Add ship to game
	game.EntityLock.Lock()
	game.Ships[shipID] = ship

	// Create player and team
	if game.Teams == nil {
		game.Teams = make(map[int]*engine.Team)
	}
	if game.Teams[teamID] == nil {
		game.Teams[teamID] = &engine.Team{
			ID:      teamID,
			Players: make(map[entity.ID]*engine.Player),
		}
	}
	game.Teams[teamID].Players[playerID] = &engine.Player{
		ID:     playerID,
		ShipID: shipID,
		TeamID: teamID,
	}
	game.EntityLock.Unlock()

	// Create server
	server := NewGameServer(game, 10)

	// Create a mock client
	client := &Client{
		ID:       entity.ID(1),
		PlayerID: playerID,
		TeamID:   teamID,
	}

	// Test input data
	inputData := PlayerInputData{
		Thrust:     true,
		TurnLeft:   false,
		TurnRight:  true,
		FireWeapon: -1, // Don't fire
		BeamDown:   false,
		BeamUp:     false,
		BeamAmount: 0,
		TargetID:   0,
	}

	// Marshal input to JSON
	jsonData, err := json.Marshal(inputData)
	if err != nil {
		t.Fatalf("Failed to marshal input data: %v", err)
	}

	// Record initial ship state
	initialThrusting := ship.Thrusting
	initialTurningCW := ship.TurningCW
	initialTurningCCW := ship.TurningCCW

	// Test handlePlayerInput method
	server.handlePlayerInput(client, jsonData)

	// Check if ship state was updated according to input
	if ship.Thrusting != inputData.Thrust {
		t.Errorf("Ship thrusting not updated: expected %v, got %v", inputData.Thrust, ship.Thrusting)
	}
	if ship.TurningCW != inputData.TurnRight {
		t.Errorf("Ship turning CW not updated: expected %v, got %v", inputData.TurnRight, ship.TurningCW)
	}
	if ship.TurningCCW != inputData.TurnLeft {
		t.Errorf("Ship turning CCW not updated: expected %v, got %v", inputData.TurnLeft, ship.TurningCCW)
	}

	t.Logf("Input processed successfully - Initial: Thrust=%v, TurnCW=%v, TurnCCW=%v",
		initialThrusting, initialTurningCW, initialTurningCCW)
	t.Logf("After input: Thrust=%v, TurnCW=%v, TurnCCW=%v",
		ship.Thrusting, ship.TurningCW, ship.TurningCCW)
}
