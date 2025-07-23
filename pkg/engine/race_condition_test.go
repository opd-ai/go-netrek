// pkg/engine/race_condition_test.go
package engine

import (
	"sync"
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/entity"
)

// TestGameRaceCondition tests for race conditions when accessing entity maps
// This test demonstrates the bug by running concurrent operations that access entity maps
func TestGameRaceCondition(t *testing.T) {
	// Create a game
	cfg := config.DefaultConfig()
	game := NewGame(cfg)
	game.Start()

	// Add initial players to have some entities
	playerIDs := make([]entity.ID, 5)
	for i := 0; i < 5; i++ {
		id, err := game.AddPlayer("TestPlayer", 0)
		if err != nil {
			t.Fatalf("Failed to add player: %v", err)
		}
		playerIDs[i] = id
	}

	// Test for race conditions by running concurrent operations
	var wg sync.WaitGroup
	done := make(chan bool, 1)

	// Start continuous game updates (reading from maps)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				game.Update() // This iterates over entity maps
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Concurrently add and remove players (writing to maps)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			// Add player
			id, err := game.AddPlayer("TempPlayer", 0)
			if err != nil {
				t.Errorf("Failed to add player: %v", err)
				return
			}

			// Immediately remove player
			err = game.RemovePlayer(id)
			if err != nil {
				t.Errorf("Failed to remove player: %v", err)
				return
			}

			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Add projectiles by firing weapons (writing to maps)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			// Find a ship to fire from
			var shipID entity.ID
			game.EntityLock.RLock() // We need to protect this read
			for _, player := range game.Teams[0].Players {
				shipID = player.ShipID
				break
			}
			game.EntityLock.RUnlock()

			if shipID != 0 {
				// Fire weapon - this adds projectiles to the map
				game.FireWeapon(shipID, 0)
			}
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// Let the race condition test run for a short time
	time.Sleep(100 * time.Millisecond)
	done <- true

	wg.Wait()
	game.Stop()

	// If we get here without panicking, the race condition might not have been triggered
	// But the race detector should catch it if run with `go test -race`
	t.Log("Race condition test completed - run with 'go test -race' to detect data races")
}

// TestGameConcurrentMapAccess specifically tests concurrent map access patterns
func TestGameConcurrentMapAccess(t *testing.T) {
	cfg := config.DefaultConfig()
	game := NewGame(cfg)

	// Add some initial entities
	playerID, _ := game.AddPlayer("TestPlayer", 0)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Simulate the problematic pattern: Update() reading while AddPlayer() writing
	wg.Add(2)

	// Goroutine 1: Continuously call Update (proper public API)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Use the public Update method which properly locks
			game.Update()
			time.Sleep(1 * time.Microsecond)
		}
	}()

	// Goroutine 2: Continuously modify entity maps
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Add player (modifies g.Ships map)
			id, err := game.AddPlayer("TempPlayer", 0)
			if err != nil {
				errors <- err
				return
			}

			// Remove player (modifies g.Ships map)
			err = game.RemovePlayer(id)
			if err != nil {
				errors <- err
				return
			}
			time.Sleep(1 * time.Microsecond)
		}
	}()

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}

	// Clean up
	game.RemovePlayer(playerID)
}
