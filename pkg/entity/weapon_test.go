// pkg/entity/weapon_test.go
package entity

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/physics"
)

func TestBaseWeapon_GetName(t *testing.T) {
	tests := []struct {
		name     string
		weapon   BaseWeapon
		expected string
	}{
		{
			name: "Torpedo weapon name",
			weapon: BaseWeapon{
				Name: "Torpedo",
			},
			expected: "Torpedo",
		},
		{
			name: "Phaser weapon name",
			weapon: BaseWeapon{
				Name: "Phaser",
			},
			expected: "Phaser",
		},
		{
			name: "Empty weapon name",
			weapon: BaseWeapon{
				Name: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.weapon.GetName()
			if result != tt.expected {
				t.Errorf("GetName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBaseWeapon_GetCooldown(t *testing.T) {
	tests := []struct {
		name     string
		weapon   BaseWeapon
		expected time.Duration
	}{
		{
			name: "Standard cooldown",
			weapon: BaseWeapon{
				Cooldown: 500 * time.Millisecond,
			},
			expected: 500 * time.Millisecond,
		},
		{
			name: "Zero cooldown",
			weapon: BaseWeapon{
				Cooldown: 0,
			},
			expected: 0,
		},
		{
			name: "Long cooldown",
			weapon: BaseWeapon{
				Cooldown: 2 * time.Second,
			},
			expected: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.weapon.GetCooldown()
			if result != tt.expected {
				t.Errorf("GetCooldown() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBaseWeapon_GetFuelCost(t *testing.T) {
	tests := []struct {
		name     string
		weapon   BaseWeapon
		expected int
	}{
		{
			name: "Standard fuel cost",
			weapon: BaseWeapon{
				FuelCost: 10,
			},
			expected: 10,
		},
		{
			name: "Zero fuel cost",
			weapon: BaseWeapon{
				FuelCost: 0,
			},
			expected: 0,
		},
		{
			name: "High fuel cost",
			weapon: BaseWeapon{
				FuelCost: 50,
			},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.weapon.GetFuelCost()
			if result != tt.expected {
				t.Errorf("GetFuelCost() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestNewTorpedo(t *testing.T) {
	ownerID := ID(42)
	torpedo := NewTorpedo(ownerID)

	if torpedo == nil {
		t.Fatal("NewTorpedo() returned nil")
	}

	// Test torpedo properties
	if torpedo.GetName() != "Torpedo" {
		t.Errorf("Torpedo name = %q, want %q", torpedo.GetName(), "Torpedo")
	}

	if torpedo.GetCooldown() != 500*time.Millisecond {
		t.Errorf("Torpedo cooldown = %v, want %v", torpedo.GetCooldown(), 500*time.Millisecond)
	}

	if torpedo.GetFuelCost() != 10 {
		t.Errorf("Torpedo fuel cost = %d, want %d", torpedo.GetFuelCost(), 10)
	}

	if torpedo.BaseWeapon.Damage != 40 {
		t.Errorf("Torpedo damage = %d, want %d", torpedo.BaseWeapon.Damage, 40)
	}

	if torpedo.BaseWeapon.Speed != 500 {
		t.Errorf("Torpedo speed = %f, want %f", torpedo.BaseWeapon.Speed, 500.0)
	}

	if torpedo.Range != 2000 {
		t.Errorf("Torpedo range = %f, want %f", torpedo.Range, 2000.0)
	}

	if torpedo.BaseWeapon.OwnerID != ownerID {
		t.Errorf("Torpedo owner ID = %d, want %d", torpedo.BaseWeapon.OwnerID, ownerID)
	}
}

func TestNewPhaser(t *testing.T) {
	ownerID := ID(123)
	phaser := NewPhaser(ownerID)

	if phaser == nil {
		t.Fatal("NewPhaser() returned nil")
	}

	// Test phaser properties
	if phaser.GetName() != "Phaser" {
		t.Errorf("Phaser name = %q, want %q", phaser.GetName(), "Phaser")
	}

	if phaser.GetCooldown() != 200*time.Millisecond {
		t.Errorf("Phaser cooldown = %v, want %v", phaser.GetCooldown(), 200*time.Millisecond)
	}

	if phaser.GetFuelCost() != 5 {
		t.Errorf("Phaser fuel cost = %d, want %d", phaser.GetFuelCost(), 5)
	}

	if phaser.BaseWeapon.Damage != 20 {
		t.Errorf("Phaser damage = %d, want %d", phaser.BaseWeapon.Damage, 20)
	}

	if phaser.BaseWeapon.Speed != 1000 {
		t.Errorf("Phaser speed = %f, want %f", phaser.BaseWeapon.Speed, 1000.0)
	}

	if phaser.Range != 800 {
		t.Errorf("Phaser range = %f, want %f", phaser.Range, 800.0)
	}

	if phaser.BaseWeapon.OwnerID != ownerID {
		t.Errorf("Phaser owner ID = %d, want %d", phaser.BaseWeapon.OwnerID, ownerID)
	}
}

func TestTorpedo_CreateProjectile(t *testing.T) {
	ownerID := ID(42)
	torpedo := NewTorpedo(ownerID)
	position := physics.Vector2D{X: 100, Y: 200}
	angle := math.Pi / 4 // 45 degrees
	teamID := 1

	projectile := torpedo.CreateProjectile(ownerID, position, angle, teamID)

	if projectile == nil {
		t.Fatal("CreateProjectile() returned nil")
	}

	// Test projectile properties
	if projectile.Type != "Torpedo" {
		t.Errorf("Projectile type = %q, want %q", projectile.Type, "Torpedo")
	}

	if projectile.OwnerID != ownerID {
		t.Errorf("Projectile owner ID = %d, want %d", projectile.OwnerID, ownerID)
	}

	if projectile.TeamID != teamID {
		t.Errorf("Projectile team ID = %d, want %d", projectile.TeamID, teamID)
	}

	if projectile.Damage != 40 {
		t.Errorf("Projectile damage = %d, want %d", projectile.Damage, 40)
	}

	if projectile.Range != 2000 {
		t.Errorf("Projectile range = %f, want %f", projectile.Range, 2000.0)
	}

	if projectile.Position != position {
		t.Errorf("Projectile position = %v, want %v", projectile.Position, position)
	}

	if projectile.Rotation != angle {
		t.Errorf("Projectile rotation = %f, want %f", projectile.Rotation, angle)
	}

	if !projectile.Active {
		t.Error("Projectile should be active")
	}

	// Test velocity is calculated correctly from angle and speed
	expectedVelocity := physics.FromAngle(angle, 500)
	if math.Abs(projectile.Velocity.X-expectedVelocity.X) > 0.001 ||
		math.Abs(projectile.Velocity.Y-expectedVelocity.Y) > 0.001 {
		t.Errorf("Projectile velocity = %v, want %v", projectile.Velocity, expectedVelocity)
	}

	// Test collider properties
	if projectile.Collider.Radius != 5 {
		t.Errorf("Projectile collider radius = %f, want %f", projectile.Collider.Radius, 5.0)
	}

	if projectile.Collider.Center != position {
		t.Errorf("Projectile collider center = %v, want %v", projectile.Collider.Center, position)
	}
}

func TestPhaser_CreateProjectile(t *testing.T) {
	ownerID := ID(123)
	phaser := NewPhaser(ownerID)
	position := physics.Vector2D{X: 300, Y: 400}
	angle := math.Pi / 2 // 90 degrees
	teamID := 2

	projectile := phaser.CreateProjectile(ownerID, position, angle, teamID)

	if projectile == nil {
		t.Fatal("CreateProjectile() returned nil")
	}

	// Test projectile properties
	if projectile.Type != "Phaser" {
		t.Errorf("Projectile type = %q, want %q", projectile.Type, "Phaser")
	}

	if projectile.OwnerID != ownerID {
		t.Errorf("Projectile owner ID = %d, want %d", projectile.OwnerID, ownerID)
	}

	if projectile.TeamID != teamID {
		t.Errorf("Projectile team ID = %d, want %d", projectile.TeamID, teamID)
	}

	if projectile.Damage != 20 {
		t.Errorf("Projectile damage = %d, want %d", projectile.Damage, 20)
	}

	if projectile.Range != 800 {
		t.Errorf("Projectile range = %f, want %f", projectile.Range, 800.0)
	}

	// Test velocity is calculated correctly from angle and speed
	expectedVelocity := physics.FromAngle(angle, 1000)
	if math.Abs(projectile.Velocity.X-expectedVelocity.X) > 0.001 ||
		math.Abs(projectile.Velocity.Y-expectedVelocity.Y) > 0.001 {
		t.Errorf("Projectile velocity = %v, want %v", projectile.Velocity, expectedVelocity)
	}

	// Test collider properties
	if projectile.Collider.Radius != 3 {
		t.Errorf("Projectile collider radius = %f, want %f", projectile.Collider.Radius, 3.0)
	}
}

func TestProjectile_Update(t *testing.T) {
	// Create a projectile for testing
	projectile := &Projectile{
		BaseEntity: BaseEntity{
			ID:       ID(1),
			Position: physics.Vector2D{X: 0, Y: 0},
			Velocity: physics.Vector2D{X: 100, Y: 0}, // Moving right at 100 units/sec
			Active:   true,
		},
		Type:             "TestProjectile",
		Range:            500,
		DistanceTraveled: 0,
	}

	// Test normal update
	deltaTime := 0.1 // 100ms
	projectile.Update(deltaTime)

	// Should have moved 10 units to the right
	expectedPos := physics.Vector2D{X: 10, Y: 0}
	if math.Abs(projectile.Position.X-expectedPos.X) > 0.001 ||
		math.Abs(projectile.Position.Y-expectedPos.Y) > 0.001 {
		t.Errorf("Position after update = %v, want %v", projectile.Position, expectedPos)
	}

	// Distance traveled should be 10
	if math.Abs(projectile.DistanceTraveled-10) > 0.001 {
		t.Errorf("Distance traveled = %f, want %f", projectile.DistanceTraveled, 10.0)
	}

	// Should still be active
	if !projectile.Active {
		t.Error("Projectile should still be active")
	}
}

func TestProjectile_Update_ExceedsRange(t *testing.T) {
	// Create a projectile that will exceed its range
	projectile := &Projectile{
		BaseEntity: BaseEntity{
			ID:       ID(1),
			Position: physics.Vector2D{X: 0, Y: 0},
			Velocity: physics.Vector2D{X: 1000, Y: 0}, // Fast movement
			Active:   true,
		},
		Type:             "TestProjectile",
		Range:            100, // Short range
		DistanceTraveled: 90,  // Already close to limit
	}

	// Update with enough time to exceed range
	deltaTime := 0.1 // This will move 100 units, total distance = 190
	projectile.Update(deltaTime)

	// Should be deactivated
	if projectile.Active {
		t.Error("Projectile should be deactivated after exceeding range")
	}

	// Distance should exceed range
	if projectile.DistanceTraveled <= projectile.Range {
		t.Errorf("Distance traveled (%f) should exceed range (%f)",
			projectile.DistanceTraveled, projectile.Range)
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	// Test that GenerateID produces unique IDs
	const numIDs = 1000
	ids := make(map[ID]bool, numIDs)

	for i := 0; i < numIDs; i++ {
		id := GenerateID()
		if ids[id] {
			t.Errorf("GenerateID() produced duplicate ID: %d", id)
		}
		ids[id] = true
	}

	if len(ids) != numIDs {
		t.Errorf("Expected %d unique IDs, got %d", numIDs, len(ids))
	}
}

func TestGenerateID_ThreadSafety(t *testing.T) {
	// Test that GenerateID is thread-safe
	const numGoroutines = 10
	const idsPerGoroutine = 100
	const totalIDs = numGoroutines * idsPerGoroutine

	ids := make(chan ID, totalIDs)
	var wg sync.WaitGroup

	// Start multiple goroutines generating IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				ids <- GenerateID()
			}
		}()
	}

	wg.Wait()
	close(ids)

	// Collect all IDs and check for uniqueness
	idSet := make(map[ID]bool, totalIDs)
	count := 0
	for id := range ids {
		if idSet[id] {
			t.Errorf("GenerateID() produced duplicate ID in concurrent test: %d", id)
		}
		idSet[id] = true
		count++
	}

	if count != totalIDs {
		t.Errorf("Expected %d total IDs, got %d", totalIDs, count)
	}

	if len(idSet) != totalIDs {
		t.Errorf("Expected %d unique IDs, got %d", totalIDs, len(idSet))
	}
}

func TestWeapon_Interface(t *testing.T) {
	// Test that Torpedo implements Weapon interface
	torpedo := NewTorpedo(ID(1))
	var _ Weapon = torpedo // Compile-time interface check

	// Test that Phaser implements Weapon interface
	phaser := NewPhaser(ID(2))
	var _ Weapon = phaser // Compile-time interface check

	// Test interface methods work
	if torpedo.GetName() == "" {
		t.Error("Torpedo GetName() should return non-empty string")
	}

	if phaser.GetName() == "" {
		t.Error("Phaser GetName() should return non-empty string")
	}

	// Test CreateProjectile through interface
	pos := physics.Vector2D{X: 0, Y: 0}
	proj1 := torpedo.CreateProjectile(ID(1), pos, 0, 0)
	if proj1 == nil {
		t.Error("Torpedo CreateProjectile() should return non-nil projectile")
	}

	proj2 := phaser.CreateProjectile(ID(2), pos, 0, 0)
	if proj2 == nil {
		t.Error("Phaser CreateProjectile() should return non-nil projectile")
	}
}
