package entity

import (
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/physics"
)

// TestShipClass_String tests the String method of ShipClass
func TestShipClass_String(t *testing.T) {
	tests := []struct {
		name     string
		class    ShipClass
		expected string
	}{
		{"Scout", Scout, "Scout"},
		{"Destroyer", Destroyer, "Destroyer"},
		{"Cruiser", Cruiser, "Cruiser"},
		{"Battleship", Battleship, "Battleship"},
		{"Assault", Assault, "Assault"},
		{"Unknown", ShipClass(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.class.String()
			if result != tt.expected {
				t.Errorf("ShipClass.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestShipClassFromString tests the ShipClassFromString function
func TestShipClassFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ShipClass
	}{
		{"Scout", "Scout", Scout},
		{"Destroyer", "Destroyer", Destroyer},
		{"Cruiser", "Cruiser", Cruiser},
		{"Battleship", "Battleship", Battleship},
		{"Assault", "Assault", Assault},
		{"Unknown", "Unknown", Scout},
		{"EmptyString", "", Scout},
		{"InvalidClass", "InvalidClass", Scout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShipClassFromString(tt.input)
			if result != tt.expected {
				t.Errorf("ShipClassFromString(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNewShip tests the NewShip constructor function
func TestNewShip(t *testing.T) {
	id := ID(1)
	class := Scout
	teamID := 0
	position := physics.Vector2D{X: 100, Y: 200}

	ship := NewShip(id, class, teamID, position)

	// Test basic properties
	if ship.ID != id {
		t.Errorf("Expected ship ID %v, got %v", id, ship.ID)
	}
	if ship.Class != class {
		t.Errorf("Expected ship class %v, got %v", class, ship.Class)
	}
	if ship.TeamID != teamID {
		t.Errorf("Expected team ID %v, got %v", teamID, ship.TeamID)
	}
	if ship.Position != position {
		t.Errorf("Expected position %v, got %v", position, ship.Position)
	}

	// Test that ship starts with full resources
	if ship.Hull != ship.Stats.MaxHull {
		t.Errorf("Expected hull %v, got %v", ship.Stats.MaxHull, ship.Hull)
	}
	if ship.Shields != ship.Stats.MaxShields {
		t.Errorf("Expected shields %v, got %v", ship.Stats.MaxShields, ship.Shields)
	}
	if ship.Fuel != ship.Stats.MaxFuel {
		t.Errorf("Expected fuel %v, got %v", ship.Stats.MaxFuel, ship.Fuel)
	}

	// Test that ship is active
	if !ship.Active {
		t.Error("Expected ship to be active")
	}

	// Test that weapons are initialized
	if len(ship.Weapons) == 0 {
		t.Error("Expected ship to have weapons")
	}

	// Test that LastFired map is initialized
	if ship.LastFired == nil {
		t.Error("Expected LastFired map to be initialized")
	}
}

// TestNewShip_DifferentShipClasses tests NewShip with different ship classes
func TestNewShip_DifferentShipClasses(t *testing.T) {
	position := physics.Vector2D{X: 0, Y: 0}

	classes := []ShipClass{Scout, Destroyer, Cruiser, Battleship, Assault}

	for _, class := range classes {
		t.Run(class.String(), func(t *testing.T) {
			ship := NewShip(ID(1), class, 0, position)

			if ship.Class != class {
				t.Errorf("Expected class %v, got %v", class, ship.Class)
			}

			// Verify ship has appropriate stats
			stats := getShipStats(class)
			if ship.Stats != stats {
				t.Errorf("Expected stats %+v, got %+v", stats, ship.Stats)
			}
		})
	}
}

// TestShip_TakeDamage tests the TakeDamage method
func TestShip_TakeDamage(t *testing.T) {
	ship := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})
	initialHull := ship.Hull
	initialShields := ship.Shields

	tests := []struct {
		name            string
		damage          int
		expectedDead    bool
		expectedShields int
		expectedHull    int
		setup           func(*Ship)
	}{
		{
			name:            "DamageToShields",
			damage:          50,
			expectedDead:    false,
			expectedShields: initialShields - 50,
			expectedHull:    initialHull,
			setup:           func(s *Ship) {},
		},
		{
			name:            "DamageExceedsShields",
			damage:          initialShields + 20,
			expectedDead:    false,
			expectedShields: 0,
			expectedHull:    initialHull - 20,
			setup:           func(s *Ship) { s.Shields = initialShields; s.Hull = initialHull },
		},
		{
			name:            "DamageKillsShip",
			damage:          initialShields + initialHull,
			expectedDead:    true,
			expectedShields: 0,
			expectedHull:    0,
			setup:           func(s *Ship) { s.Shields = initialShields; s.Hull = initialHull },
		},
		{
			name:            "DamageToHullOnly",
			damage:          30,
			expectedDead:    false,
			expectedShields: 0,
			expectedHull:    initialHull - 30,
			setup:           func(s *Ship) { s.Shields = 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset ship state
			ship.Hull = initialHull
			ship.Shields = initialShields
			tt.setup(ship)

			isDead := ship.TakeDamage(tt.damage)

			if isDead != tt.expectedDead {
				t.Errorf("Expected isDead %v, got %v", tt.expectedDead, isDead)
			}
			if ship.Shields != tt.expectedShields {
				t.Errorf("Expected shields %v, got %v", tt.expectedShields, ship.Shields)
			}
			if ship.Hull != tt.expectedHull {
				t.Errorf("Expected hull %v, got %v", tt.expectedHull, ship.Hull)
			}
		})
	}
}

// TestShip_Update tests the Update method
func TestShip_Update(t *testing.T) {
	ship := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})
	deltaTime := 0.1
	initialFuel := ship.Fuel

	t.Run("TurningClockwise", func(t *testing.T) {
		ship.Rotation = 0
		ship.TurningCW = true
		ship.TurningCCW = false

		ship.Update(deltaTime)

		expectedRotation := ship.Stats.TurnRate * deltaTime
		if ship.Rotation != expectedRotation {
			t.Errorf("Expected rotation %v, got %v", expectedRotation, ship.Rotation)
		}
	})

	t.Run("TurningCounterClockwise", func(t *testing.T) {
		ship.Rotation = 0
		ship.TurningCW = false
		ship.TurningCCW = true

		ship.Update(deltaTime)

		expectedRotation := -ship.Stats.TurnRate * deltaTime
		if ship.Rotation != expectedRotation {
			t.Errorf("Expected rotation %v, got %v", expectedRotation, ship.Rotation)
		}
	})

	t.Run("Thrusting", func(t *testing.T) {
		ship.Rotation = 0
		ship.Velocity = physics.Vector2D{X: 0, Y: 0}
		ship.Fuel = initialFuel
		ship.Thrusting = true
		ship.TurningCW = false
		ship.TurningCCW = false

		initialVelocity := ship.Velocity

		ship.Update(deltaTime)

		// Check that velocity changed
		if ship.Velocity == initialVelocity {
			t.Error("Expected velocity to change when thrusting")
		}

		// Check that fuel was consumed
		if ship.Fuel >= initialFuel {
			t.Error("Expected fuel to be consumed when thrusting")
		}
	})

	t.Run("ThrustingWithoutFuel", func(t *testing.T) {
		ship.Fuel = 0
		ship.Thrusting = true
		initialVelocity := ship.Velocity

		ship.Update(deltaTime)

		// Velocity should not increase without fuel
		// Note: velocity might change due to drag, but not due to thrust
		if ship.Velocity.Length() > initialVelocity.Length() {
			t.Error("Expected velocity not to increase without fuel")
		}
	})

	t.Run("ShieldRegeneration", func(t *testing.T) {
		ship.Shields = ship.Stats.MaxShields - 10
		initialShields := ship.Shields
		deltaTime := 1.0 // Use 1 second for meaningful regeneration

		ship.Update(deltaTime)

		if ship.Shields <= initialShields {
			t.Error("Expected shields to regenerate")
		}

		// Shields should not exceed maximum
		if ship.Shields > ship.Stats.MaxShields {
			t.Errorf("Expected shields not to exceed max %v, got %v", ship.Stats.MaxShields, ship.Shields)
		}
	})
}

// TestShip_FireWeapon tests the FireWeapon method
func TestShip_FireWeapon(t *testing.T) {
	ship := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})

	t.Run("ValidWeaponIndex", func(t *testing.T) {
		if len(ship.Weapons) == 0 {
			t.Skip("No weapons available to test")
		}

		// Reset fuel for this test
		ship.Fuel = 1000
		initialFuel := ship.Fuel
		projectile := ship.FireWeapon(0)

		if projectile == nil {
			t.Error("Expected projectile to be created")
		}

		// Check that fuel was consumed
		expectedFuelCost := ship.Weapons[0].GetFuelCost()
		if ship.Fuel != initialFuel-expectedFuelCost {
			t.Errorf("Expected fuel %v, got %v", initialFuel-expectedFuelCost, ship.Fuel)
		}

		// Check that cooldown was set
		weaponName := ship.Weapons[0].GetName()
		if _, exists := ship.LastFired[weaponName]; !exists {
			t.Error("Expected weapon cooldown to be set")
		}
	})

	t.Run("InvalidWeaponIndex", func(t *testing.T) {
		projectile := ship.FireWeapon(-1)
		if projectile != nil {
			t.Error("Expected nil projectile for invalid weapon index")
		}

		projectile = ship.FireWeapon(999)
		if projectile != nil {
			t.Error("Expected nil projectile for invalid weapon index")
		}
	})

	t.Run("InsufficientFuel", func(t *testing.T) {
		if len(ship.Weapons) == 0 {
			t.Skip("No weapons available to test")
		}

		ship.Fuel = 0
		projectile := ship.FireWeapon(0)

		if projectile != nil {
			t.Error("Expected nil projectile when fuel is insufficient")
		}
	})

	t.Run("WeaponOnCooldown", func(t *testing.T) {
		if len(ship.Weapons) == 0 {
			t.Skip("No weapons available to test")
		}

		// Reset fuel and cooldowns for this test
		ship.Fuel = 1000
		ship.LastFired = make(map[string]time.Time)

		// Debug information
		t.Logf("Ship fuel: %d", ship.Fuel)
		t.Logf("Weapon count: %d", len(ship.Weapons))
		if len(ship.Weapons) > 0 {
			t.Logf("Weapon fuel cost: %d", ship.Weapons[0].GetFuelCost())
		}

		// First shot should work
		projectile1 := ship.FireWeapon(0)
		if projectile1 == nil {
			t.Error("Expected first shot to work")
		}

		// Immediate second shot should fail due to cooldown
		projectile2 := ship.FireWeapon(0)
		if projectile2 != nil {
			t.Error("Expected second shot to fail due to cooldown")
		}
	})
}

// TestShip_RepairTick tests the RepairTick method
func TestShip_RepairTick(t *testing.T) {
	ship := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})
	deltaTime := 1.0 // 1 second

	t.Run("RepairModeOff", func(t *testing.T) {
		ship.RepairMode = false
		ship.Hull = ship.Stats.MaxHull - 50
		initialHull := ship.Hull
		initialFuel := ship.Fuel

		ship.RepairTick(deltaTime)

		// Hull and fuel should not change when repair mode is off
		if ship.Hull != initialHull {
			t.Errorf("Expected hull to remain %v, got %v", initialHull, ship.Hull)
		}
		if ship.Fuel != initialFuel {
			t.Errorf("Expected fuel to remain %v, got %v", initialFuel, ship.Fuel)
		}
	})

	t.Run("RepairModeOn", func(t *testing.T) {
		ship.RepairMode = true
		ship.Hull = ship.Stats.MaxHull - 50
		ship.Fuel = 100
		initialHull := ship.Hull
		initialFuel := ship.Fuel

		ship.RepairTick(deltaTime)

		// Hull should increase
		if ship.Hull <= initialHull {
			t.Error("Expected hull to increase during repair")
		}

		// Fuel should decrease
		if ship.Fuel >= initialFuel {
			t.Error("Expected fuel to decrease during repair")
		}
	})

	t.Run("RepairWithInsufficientFuel", func(t *testing.T) {
		ship.RepairMode = true
		ship.Hull = ship.Stats.MaxHull - 50
		ship.Fuel = 1 // Very low fuel

		ship.RepairTick(deltaTime)

		// Repair mode should turn off when fuel runs out
		if ship.RepairMode {
			t.Error("Expected repair mode to turn off when fuel runs out")
		}
		if ship.Fuel != 0 {
			t.Errorf("Expected fuel to be 0, got %v", ship.Fuel)
		}
	})

	t.Run("RepairDoesNotExceedMaxHull", func(t *testing.T) {
		ship.RepairMode = true
		ship.Hull = ship.Stats.MaxHull - 1 // Almost full
		ship.Fuel = 100

		ship.RepairTick(deltaTime)

		if ship.Hull > ship.Stats.MaxHull {
			t.Errorf("Expected hull not to exceed max %v, got %v", ship.Stats.MaxHull, ship.Hull)
		}
	})
}

// TestSetShipTypeStats tests the SetShipTypeStats function
func TestSetShipTypeStats(t *testing.T) {
	// Save original stats to restore later
	originalStats := shipTypeStats

	customStats := map[string]ShipStats{
		"Scout": {
			MaxHull:      200, // Different from default
			MaxShields:   200,
			MaxFuel:      2000,
			Acceleration: 400,
			TurnRate:     6.0,
			MaxSpeed:     600,
			WeaponSlots:  4,
			MaxArmies:    4,
		},
	}

	SetShipTypeStats(customStats)

	// Test that custom stats are used
	stats := getShipStats(Scout)
	expected := customStats["Scout"]
	if stats != expected {
		t.Errorf("Expected custom stats %+v, got %+v", expected, stats)
	}

	// Restore original stats
	shipTypeStats = originalStats
}

// TestGetShipStats tests the getShipStats function
func TestGetShipStats(t *testing.T) {
	tests := []struct {
		name  string
		class ShipClass
	}{
		{"Scout", Scout},
		{"Destroyer", Destroyer},
		{"UnknownClass", ShipClass(999)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := getShipStats(tt.class)

			// All stats should be positive
			if stats.MaxHull <= 0 {
				t.Errorf("Expected positive MaxHull, got %v", stats.MaxHull)
			}
			if stats.MaxShields <= 0 {
				t.Errorf("Expected positive MaxShields, got %v", stats.MaxShields)
			}
			if stats.MaxFuel <= 0 {
				t.Errorf("Expected positive MaxFuel, got %v", stats.MaxFuel)
			}
			if stats.Acceleration <= 0 {
				t.Errorf("Expected positive Acceleration, got %v", stats.Acceleration)
			}
			if stats.TurnRate <= 0 {
				t.Errorf("Expected positive TurnRate, got %v", stats.TurnRate)
			}
			if stats.MaxSpeed <= 0 {
				t.Errorf("Expected positive MaxSpeed, got %v", stats.MaxSpeed)
			}
			if stats.WeaponSlots <= 0 {
				t.Errorf("Expected positive WeaponSlots, got %v", stats.WeaponSlots)
			}
		})
	}
}

// TestShipClassDifferentiation verifies that each ship class has distinct stats
func TestShipClassDifferentiation(t *testing.T) {
	position := physics.Vector2D{X: 0, Y: 0}

	// Create ships of each class
	scout := NewShip(1, Scout, 1, position)
	destroyer := NewShip(2, Destroyer, 1, position)
	cruiser := NewShip(3, Cruiser, 1, position)
	battleship := NewShip(4, Battleship, 1, position)
	assault := NewShip(5, Assault, 1, position)

	ships := []*Ship{scout, destroyer, cruiser, battleship, assault}
	names := []string{"Scout", "Destroyer", "Cruiser", "Battleship", "Assault"}

	// Verify all ship classes have different stats (no two should be identical)
	for i := 0; i < len(ships); i++ {
		for j := i + 1; j < len(ships); j++ {
			ship1, ship2 := ships[i], ships[j]
			name1, name2 := names[i], names[j]

			// At least one stat should be different between any two ship classes
			if ship1.Stats.MaxHull == ship2.Stats.MaxHull &&
				ship1.Stats.MaxShields == ship2.Stats.MaxShields &&
				ship1.Stats.Acceleration == ship2.Stats.Acceleration &&
				ship1.Stats.MaxSpeed == ship2.Stats.MaxSpeed &&
				ship1.Stats.WeaponSlots == ship2.Stats.WeaponSlots &&
				ship1.Stats.MaxArmies == ship2.Stats.MaxArmies {
				t.Errorf("%s and %s have identical stats - ship classes should be differentiated", name1, name2)
			}
		}
	}

	// Verify expected ship progression: Scout -> Destroyer -> Cruiser -> Battleship
	// Hull should generally increase with ship size
	if scout.Stats.MaxHull >= destroyer.Stats.MaxHull {
		t.Error("Scout hull should be less than Destroyer hull")
	}
	if destroyer.Stats.MaxHull >= cruiser.Stats.MaxHull {
		t.Error("Destroyer hull should be less than Cruiser hull")
	}
	if cruiser.Stats.MaxHull >= battleship.Stats.MaxHull {
		t.Error("Cruiser hull should be less than Battleship hull")
	}

	// Speed should generally decrease with ship size (larger ships are slower)
	if scout.Stats.MaxSpeed <= destroyer.Stats.MaxSpeed {
		t.Error("Scout should be faster than Destroyer")
	}
	if destroyer.Stats.MaxSpeed <= battleship.Stats.MaxSpeed {
		t.Error("Destroyer should be faster than Battleship")
	}
}

// Benchmark tests for performance verification
func BenchmarkNewShip(b *testing.B) {
	position := physics.Vector2D{X: 100, Y: 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewShip(ID(i), Scout, 0, position)
	}
}

func BenchmarkShip_Update(b *testing.B) {
	ship := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})
	deltaTime := 0.016 // ~60 FPS

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ship.Update(deltaTime)
	}
}

func BenchmarkShip_TakeDamage(b *testing.B) {
	ship := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ship.TakeDamage(10)
		// Reset health to avoid dying
		if ship.Hull <= 0 {
			ship.Hull = ship.Stats.MaxHull
			ship.Shields = ship.Stats.MaxShields
		}
	}
}
