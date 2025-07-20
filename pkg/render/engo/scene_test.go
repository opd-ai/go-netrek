// pkg/render/engo/scene_test.go
package engo

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/event"
	"github.com/opd-ai/go-netrek/pkg/network"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// TestNewGameScene tests the creation of a new game scene
func TestNewGameScene(t *testing.T) {
	client := network.NewGameClient(event.NewEventBus())
	eventBus := event.NewEventBus()
	playerID := uint64(123)

	scene := NewGameScene(client, eventBus, playerID)

	if scene == nil {
		t.Fatal("NewGameScene() returned nil")
	}

	if scene.client != client {
		t.Errorf("Expected client to be set correctly")
	}

	if scene.eventBus != eventBus {
		t.Errorf("Expected eventBus to be set correctly")
	}

	if scene.playerID != playerID {
		t.Errorf("Expected playerID to be %d, got %d", playerID, scene.playerID)
	}

	if scene.world == nil {
		t.Errorf("Expected world to be initialized")
	}
}

// TestGameScene_Type tests the Type method
func TestGameScene_Type(t *testing.T) {
	client := network.NewGameClient(event.NewEventBus())
	eventBus := event.NewEventBus()
	scene := NewGameScene(client, eventBus, 123)

	expectedType := "GameScene"
	actualType := scene.Type()

	if actualType != expectedType {
		t.Errorf("Expected Type() to return %q, got %q", expectedType, actualType)
	}
}

// TestGameScene_ConvertShipStateToEntity tests ship state conversion
func TestGameScene_ConvertShipStateToEntity(t *testing.T) {
	tests := []struct {
		name      string
		shipState engine.ShipState
	}{
		{
			name: "basic_ship_conversion",
			shipState: engine.ShipState{
				ID:       entity.ID(1),
				Position: physics.Vector2D{X: 100, Y: 200},
				Rotation: 45.0,
				Velocity: physics.Vector2D{X: 10, Y: 20},
				Hull:     100,
				Shields:  50,
				Fuel:     80,
				Armies:   5,
				TeamID:   1,
				Class:    entity.Scout,
			},
		},
		{
			name: "destroyer_ship_conversion",
			shipState: engine.ShipState{
				ID:       entity.ID(2),
				Position: physics.Vector2D{X: 0, Y: 0},
				Rotation: 0.0,
				Velocity: physics.Vector2D{X: 0, Y: 0},
				Hull:     200,
				Shields:  100,
				Fuel:     150,
				Armies:   10,
				TeamID:   2,
				Class:    entity.Destroyer,
			},
		},
	}

	client := network.NewGameClient(event.NewEventBus())
	eventBus := event.NewEventBus()
	scene := NewGameScene(client, eventBus, 123)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ship := scene.convertShipStateToEntity(tt.shipState)

			if ship == nil {
				t.Fatal("convertShipStateToEntity() returned nil")
			}

			// Verify BaseEntity fields
			if ship.ID != tt.shipState.ID {
				t.Errorf("Expected ID %v, got %v", tt.shipState.ID, ship.ID)
			}

			if ship.Position != tt.shipState.Position {
				t.Errorf("Expected Position %v, got %v", tt.shipState.Position, ship.Position)
			}

			if ship.Rotation != tt.shipState.Rotation {
				t.Errorf("Expected Rotation %v, got %v", tt.shipState.Rotation, ship.Rotation)
			}

			if ship.Velocity != tt.shipState.Velocity {
				t.Errorf("Expected Velocity %v, got %v", tt.shipState.Velocity, ship.Velocity)
			}

			if !ship.Active {
				t.Errorf("Expected Active to be true")
			}

			// Verify Ship-specific fields
			if ship.Class != tt.shipState.Class {
				t.Errorf("Expected Class %v, got %v", tt.shipState.Class, ship.Class)
			}

			if ship.TeamID != tt.shipState.TeamID {
				t.Errorf("Expected TeamID %v, got %v", tt.shipState.TeamID, ship.TeamID)
			}

			if ship.Hull != tt.shipState.Hull {
				t.Errorf("Expected Hull %v, got %v", tt.shipState.Hull, ship.Hull)
			}

			if ship.Shields != tt.shipState.Shields {
				t.Errorf("Expected Shields %v, got %v", tt.shipState.Shields, ship.Shields)
			}

			if ship.Fuel != tt.shipState.Fuel {
				t.Errorf("Expected Fuel %v, got %v", tt.shipState.Fuel, ship.Fuel)
			}

			if ship.Armies != tt.shipState.Armies {
				t.Errorf("Expected Armies %v, got %v", tt.shipState.Armies, ship.Armies)
			}
		})
	}
}

// TestGameScene_ConvertPlanetStateToEntity tests planet state conversion
func TestGameScene_ConvertPlanetStateToEntity(t *testing.T) {
	tests := []struct {
		name        string
		planetState engine.PlanetState
	}{
		{
			name: "earth_planet",
			planetState: engine.PlanetState{
				ID:       entity.ID(10),
				Name:     "Earth",
				Position: physics.Vector2D{X: 500, Y: 300},
				TeamID:   1,
				Armies:   20,
			},
		},
		{
			name: "neutral_planet",
			planetState: engine.PlanetState{
				ID:       entity.ID(11),
				Name:     "Mars",
				Position: physics.Vector2D{X: -100, Y: -200},
				TeamID:   0, // Neutral
				Armies:   15,
			},
		},
	}

	client := network.NewGameClient(event.NewEventBus())
	eventBus := event.NewEventBus()
	scene := NewGameScene(client, eventBus, 123)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planet := scene.convertPlanetStateToEntity(tt.planetState)

			if planet == nil {
				t.Fatal("convertPlanetStateToEntity() returned nil")
			}

			// Verify BaseEntity fields
			if planet.ID != tt.planetState.ID {
				t.Errorf("Expected ID %v, got %v", tt.planetState.ID, planet.ID)
			}

			if planet.Position != tt.planetState.Position {
				t.Errorf("Expected Position %v, got %v", tt.planetState.Position, planet.Position)
			}

			if !planet.Active {
				t.Errorf("Expected Active to be true")
			}

			// Verify Planet-specific fields
			if planet.Name != tt.planetState.Name {
				t.Errorf("Expected Name %v, got %v", tt.planetState.Name, planet.Name)
			}

			if planet.TeamID != tt.planetState.TeamID {
				t.Errorf("Expected TeamID %v, got %v", tt.planetState.TeamID, planet.TeamID)
			}

			if planet.Armies != tt.planetState.Armies {
				t.Errorf("Expected Armies %v, got %v", tt.planetState.Armies, planet.Armies)
			}
		})
	}
}

// TestGameScene_ConvertProjectileStateToEntity tests projectile state conversion
func TestGameScene_ConvertProjectileStateToEntity(t *testing.T) {
	tests := []struct {
		name            string
		projectileState engine.ProjectileState
	}{
		{
			name: "torpedo_projectile",
			projectileState: engine.ProjectileState{
				ID:       entity.ID(100),
				Position: physics.Vector2D{X: 50, Y: 75},
				Velocity: physics.Vector2D{X: 100, Y: 0},
				Type:     "torpedo",
				TeamID:   1,
			},
		},
		{
			name: "phaser_projectile",
			projectileState: engine.ProjectileState{
				ID:       entity.ID(101),
				Position: physics.Vector2D{X: -25, Y: 30},
				Velocity: physics.Vector2D{X: 0, Y: -50},
				Type:     "phaser",
				TeamID:   2,
			},
		},
	}

	client := network.NewGameClient(event.NewEventBus())
	eventBus := event.NewEventBus()
	scene := NewGameScene(client, eventBus, 123)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectile := scene.convertProjectileStateToEntity(tt.projectileState)

			if projectile == nil {
				t.Fatal("convertProjectileStateToEntity() returned nil")
			}

			// Verify BaseEntity fields
			if projectile.ID != tt.projectileState.ID {
				t.Errorf("Expected ID %v, got %v", tt.projectileState.ID, projectile.ID)
			}

			if projectile.Position != tt.projectileState.Position {
				t.Errorf("Expected Position %v, got %v", tt.projectileState.Position, projectile.Position)
			}

			if projectile.Velocity != tt.projectileState.Velocity {
				t.Errorf("Expected Velocity %v, got %v", tt.projectileState.Velocity, projectile.Velocity)
			}

			if !projectile.Active {
				t.Errorf("Expected Active to be true")
			}

			// Verify Projectile-specific fields
			if projectile.Type != tt.projectileState.Type {
				t.Errorf("Expected Type %v, got %v", tt.projectileState.Type, projectile.Type)
			}

			if projectile.TeamID != tt.projectileState.TeamID {
				t.Errorf("Expected TeamID %v, got %v", tt.projectileState.TeamID, projectile.TeamID)
			}
		})
	}
}

// TestGameScene_IsPlayerShip tests player ship identification
func TestGameScene_IsPlayerShip(t *testing.T) {
	tests := []struct {
		name     string
		playerID uint64
		shipID   uint64
		expected bool
	}{
		{
			name:     "is_player_ship",
			playerID: 123,
			shipID:   123,
			expected: true,
		},
		{
			name:     "is_not_player_ship",
			playerID: 123,
			shipID:   456,
			expected: false,
		},
		{
			name:     "zero_ship_id",
			playerID: 0,
			shipID:   0,
			expected: true,
		},
	}

	client := network.NewGameClient(event.NewEventBus())
	eventBus := event.NewEventBus()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scene := NewGameScene(client, eventBus, tt.playerID)
			result := scene.isPlayerShip(tt.shipID)

			if result != tt.expected {
				t.Errorf("Expected isPlayerShip(%d) to return %t, got %t", tt.shipID, tt.expected, result)
			}
		})
	}
}

// TestGameScene_Preload tests the Preload method
func TestGameScene_Preload(t *testing.T) {
	client := network.NewGameClient(event.NewEventBus())
	eventBus := event.NewEventBus()
	scene := NewGameScene(client, eventBus, 123)

	// Preload should not panic or error
	scene.Preload()

	// Since Preload is currently a no-op, just verify it doesn't crash
	// In a real implementation, this would verify asset loading
}

// TestGameScene_Exit tests the Exit method
func TestGameScene_Exit(t *testing.T) {
	client := network.NewGameClient(event.NewEventBus())
	eventBus := event.NewEventBus()
	scene := NewGameScene(client, eventBus, 123)

	// Exit should not panic or error
	scene.Exit()

	// Since Exit is currently a no-op, just verify it doesn't crash
	// In a real implementation, this would verify cleanup
}
