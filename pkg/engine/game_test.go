// Package engine provides unit tests for game.go
package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/event"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

func defaultConfig() *config.GameConfig {
	return &config.GameConfig{
		WorldSize: 1000,
		Teams:     []config.TeamConfig{{Name: "Red", Color: "#f00"}, {Name: "Blue", Color: "#00f"}},
		Planets:   []config.PlanetConfig{{Name: "Earth", X: 0, Y: 0, Type: entity.Homeworld, HomeWorld: true, TeamID: 0, InitialArmies: 10}},
		GameRules: config.GameRules{TimeLimit: 60, WinCondition: "score"},
	}
}

func TestNewGame_InitializesState(t *testing.T) {
	cfg := defaultConfig()
	game := NewGame(cfg)
	if game == nil {
		t.Fatal("NewGame returned nil")
	}
	if len(game.Teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(game.Teams))
	}
	if len(game.Planets) != 1 {
		t.Errorf("expected 1 planet, got %d", len(game.Planets))
	}
	if game.SpatialIndex == nil {
		t.Error("SpatialIndex not initialized")
	}
}

func TestGame_StartStop_Transitions(t *testing.T) {
	game := NewGame(defaultConfig())
	game.Start()
	if !game.Running || game.Status != GameStatusActive {
		t.Error("Game did not start correctly")
	}
	game.Stop()
	if game.Running {
		t.Error("Game did not stop correctly")
	}
}

func TestGame_AddPlayer_TableDriven(t *testing.T) {
	game := NewGame(defaultConfig())
	tests := []struct {
		name    string
		teamID  int
		wantErr bool
	}{
		{"Alice", 0, false},
		{"Bob", 1, false},
		{"Eve", 99, true}, // invalid team
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, err := game.AddPlayer(tc.name, tc.teamID)
			if tc.wantErr && err == nil {
				t.Errorf("expected error for teamID %d", tc.teamID)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.wantErr && id == 0 {
				t.Error("expected valid player ID")
			}
		})
	}
}

func TestGame_RemovePlayer_ErrorCases(t *testing.T) {
	game := NewGame(defaultConfig())
	id, _ := game.AddPlayer("Test", 0)
	if err := game.RemovePlayer(id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := game.RemovePlayer(9999); err == nil {
		t.Error("expected error for non-existent player")
	}
}

func TestGame_RespawnShip_ErrorCases(t *testing.T) {
	game := NewGame(defaultConfig())
	id, _ := game.AddPlayer("Test", 0)
	if err := game.RespawnShip(id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := game.RespawnShip(9999); err == nil {
		t.Error("expected error for non-existent player")
	}
}

func TestGame_BeamArmies_Scenarios(t *testing.T) {
	game := NewGame(defaultConfig())
	pid, _ := game.AddPlayer("Test", 0)
	var planet *entity.Planet
	for _, p := range game.Planets {
		planet = p
		break
	}
	ship := game.Ships[game.Teams[0].Players[pid].ShipID]
	ship.Position = planet.Position
	ship.Armies = 5
	planet.TeamID = 0
	planet.Armies = 10

	// Success: beam down
	trans, err := game.BeamArmies(ship.ID, planet.ID, "down", 3)
	if err != nil || trans != 3 {
		t.Errorf("beam down failed: %v, %d", err, trans)
	}
	// Error: invalid direction
	_, err = game.BeamArmies(ship.ID, planet.ID, "sideways", 1)
	if err == nil {
		t.Error("expected error for invalid direction")
	}
}

func TestGame_FireWeapon_ErrorCases(t *testing.T) {
	game := NewGame(defaultConfig())
	pid, _ := game.AddPlayer("Test", 0)
	ship := game.Ships[game.Teams[0].Players[pid].ShipID]
	ship.Active = false
	if err := game.FireWeapon(ship.ID, 0); err == nil {
		t.Error("expected error for inactive ship")
	}
	if err := game.FireWeapon(9999, 0); err == nil {
		t.Error("expected error for non-existent ship")
	}
}

func TestGame_GetGameState_ReflectsEntities(t *testing.T) {
	game := NewGame(defaultConfig())
	_, _ = game.AddPlayer("Test", 0)
	state := game.GetGameState()
	if len(state.Ships) == 0 {
		t.Error("expected at least one ship in game state")
	}
	if len(state.Planets) == 0 {
		t.Error("expected at least one planet in game state")
	}
}

func TestGame_Update_TickAndEntityLifecycle(t *testing.T) {
	game := NewGame(defaultConfig())
	pid, _ := game.AddPlayer("Test", 0)
	ship := game.Ships[game.Teams[0].Players[pid].ShipID]
	ship.Active = true
	oldTick := game.CurrentTick
	game.Update()
	if game.CurrentTick != oldTick+1 {
		t.Errorf("expected tick to increment, got %d -> %d", oldTick, game.CurrentTick)
	}
}

func TestGame_Update_EndsOnTimeLimit(t *testing.T) {
	cfg := defaultConfig()
	cfg.GameRules.TimeLimit = 1 // 1 second
	game := NewGame(cfg)
	game.Start()
	game.StartTime = game.StartTime.Add(-2 * time.Second) // simulate time passage
	game.Update()
	if game.Status != GameStatusEnded {
		t.Error("game should end when time limit is reached")
	}
}

func TestGame_registerEventHandlers_ShipDestroyedEvent(t *testing.T) {
	game := NewGame(defaultConfig())
	game.registerEventHandlers()
	team := game.Teams[0]
	team.ShipCount = 0 // simulate all ships destroyed
	evt := event.NewShipEvent(event.ShipDestroyed, game, 1, 0)
	game.EventBus.Publish(evt)
	if game.Status != GameStatusEnded {
		t.Error("game status should be ended after ship destruction event")
	}
}

func TestGame_wrapEntityPosition_WrapsCorrectly(t *testing.T) {
	game := NewGame(defaultConfig())
	pid, _ := game.AddPlayer("Test", 0)
	ship := game.Ships[game.Teams[0].Players[pid].ShipID]
	world := game.Config.WorldSize / 2
	ship.Position = physics.Vector2D{X: world + 10, Y: world + 10}
	game.wrapEntityPosition(ship)
	if ship.Position.X > world || ship.Position.Y > world {
		t.Error("ship position not wrapped correctly")
	}
}

func TestGame_cleanupInactiveEntities_RemovesProjectiles(t *testing.T) {
	game := NewGame(defaultConfig())
	proj := &entity.Projectile{BaseEntity: entity.BaseEntity{ID: 42, Active: false}}
	game.Projectiles[42] = proj
	game.cleanupInactiveEntities()
	if _, ok := game.Projectiles[42]; ok {
		t.Error("inactive projectile not removed")
	}
}

func TestGame_endGame_SetsStatusAndWinner(t *testing.T) {
	game := NewGame(defaultConfig())
	game.Teams[0].Score = 100
	game.endGame()
	if game.Status != GameStatusEnded {
		t.Error("game status not set to ended")
	}
	if game.WinningTeam != 0 {
		t.Errorf("expected winning team 0, got %d", game.WinningTeam)
	}
}
