// pkg/engine/win_condition_test.go
package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// TestWinCondition_ConquestWin verifies that conquest-based win conditions work correctly
func TestWinCondition_ConquestWin(t *testing.T) {
	// Create a game with conquest win condition
	cfg := config.DefaultConfig()
	cfg.GameRules.WinCondition = "conquest"
	cfg.GameRules.TimeLimit = 0 // Disable time limit to focus on conquest

	game := NewGame(cfg)
	game.Start()

	// Note: Game automatically creates 2 default planets (Earth=team 0, Romulus=team 1)
	// For conquest, one team needs to control ALL planets

	// Team 1 conquers all planets (including the default ones)
	for _, planet := range game.Planets {
		planet.TeamID = 1 // Team 1 takes control of all planets
	}

	// Game should end when team 1 controls all planets
	initialStatus := game.Status

	// Simulate game updates - game should end during this loop
	for i := 0; i < 10; i++ {
		game.Update()
		if game.Status == GameStatusEnded {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Verify game ended correctly
	if game.Status != GameStatusEnded {
		t.Errorf("Expected game to end due to conquest (Status = %v), but Status = %v", GameStatusEnded, game.Status)
	}

	if game.Running != false {
		t.Errorf("Expected game.Running to be false after conquest win, but got %v", game.Running)
	}

	t.Logf("Conquest win condition working correctly")
	t.Logf("Initial status: %v, Final status: %v", initialStatus, game.Status)
	t.Logf("Total planets: %d, all controlled by team 1", len(game.Planets))
}

// TestWinCondition_ScoreWin verifies that score-based win conditions work correctly
func TestWinCondition_ScoreWin(t *testing.T) {
	// Create a game with score win condition
	cfg := config.DefaultConfig()
	cfg.GameRules.WinCondition = "score"
	cfg.GameRules.MaxScore = 100
	cfg.GameRules.TimeLimit = 0 // Disable time limit

	game := NewGame(cfg)
	game.Start()

	// Create teams
	team1 := &Team{ID: 1, Name: "Team 1", PlanetCount: 0, Score: 50}
	team2 := &Team{ID: 2, Name: "Team 2", PlanetCount: 0, Score: 30}
	game.Teams[1] = team1
	game.Teams[2] = team2

	// Team 1 reaches max score
	team1.Score = 150 // Exceeds MaxScore of 100

	// Game should end when team 1 reaches max score
	initialStatus := game.Status

	// Simulate game updates - game should end during this loop
	for i := 0; i < 10; i++ {
		game.Update()
		if game.Status == GameStatusEnded {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Verify game ended correctly
	if game.Status != GameStatusEnded {
		t.Errorf("Expected game to end due to score limit (Status = %v), but Status = %v", GameStatusEnded, game.Status)
	}

	if game.Running != false {
		t.Errorf("Expected game.Running to be false after score win, but got %v", game.Running)
	}

	t.Logf("Score win condition working correctly")
	t.Logf("Initial status: %v, Final status: %v", initialStatus, game.Status)
	t.Logf("Team 1 score: %d (max: %d), Team 2 score: %d", team1.Score, cfg.GameRules.MaxScore, team2.Score)
}

// TestWinCondition_NoWinCondition verifies games continue when no win conditions are met
func TestWinCondition_NoWinCondition(t *testing.T) {
	// Create a game with conquest win condition
	cfg := config.DefaultConfig()
	cfg.GameRules.WinCondition = "conquest"
	cfg.GameRules.TimeLimit = 0 // Disable time limit

	game := NewGame(cfg)
	game.Start()

	// Create teams with no clear winner
	team1 := &Team{ID: 1, Name: "Team 1", PlanetCount: 1, Score: 0}
	team2 := &Team{ID: 2, Name: "Team 2", PlanetCount: 1, Score: 0}
	game.Teams[1] = team1
	game.Teams[2] = team2

	// Create planets with split ownership (no winner)
	planet1 := entity.NewPlanet(1, "Planet 1", physics.Vector2D{X: 100, Y: 100}, entity.Homeworld)
	planet2 := entity.NewPlanet(2, "Planet 2", physics.Vector2D{X: 200, Y: 200}, entity.Homeworld)
	planet1.TeamID = 1
	planet2.TeamID = 2
	game.Planets[1] = planet1
	game.Planets[2] = planet2

	// Simulate several game updates
	for i := 0; i < 5; i++ {
		game.Update()
		time.Sleep(1 * time.Millisecond)
	}

	// Game should still be running since no win conditions are met
	if game.Status != GameStatusActive {
		t.Errorf("Expected game to still be active, but Status = %v", game.Status)
	}

	if game.Running != true {
		t.Errorf("Expected game.Running to be true, but got %v", game.Running)
	}

	t.Logf("Game correctly continues when no win conditions are met")
	t.Logf("Team 1 planets: %d, Team 2 planets: %d", team1.PlanetCount, team2.PlanetCount)
}
