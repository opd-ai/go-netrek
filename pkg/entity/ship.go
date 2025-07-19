// pkg/entity/ship.go
package entity

import (
	"time"

	"github.com/opd-ai/go-netrek/pkg/physics"
)

// ShipClass defines the type of ship and its capabilities
type ShipClass int

const (
	Scout ShipClass = iota
	Destroyer
	Cruiser
	Battleship
	Assault
)

// ShipStats contains the base statistics for a ship class
type ShipStats struct {
	MaxHull      int
	MaxShields   int
	MaxFuel      int
	Acceleration float64
	TurnRate     float64
	MaxSpeed     float64
	WeaponSlots  int
	MaxArmies    int
}

// Ship represents a player's spaceship in the Netrek game
type Ship struct {
	BaseEntity
	Class          ShipClass
	Stats          ShipStats
	TeamID         int
	PlayerID       ID
	Hull           int
	Shields        int
	Fuel           int
	Weapons        []Weapon
	Armies         int
	Cloaked        bool
	LastFired      map[string]time.Time
	Thrusting      bool
	TurningCW      bool
	TurningCCW     bool
	RepairMode     bool
	Damaged        bool
	Warping        bool
	LastDamageTime time.Time
	LastRepairTime time.Time
}

// NewShip creates a new ship with the specified class and team
func NewShip(id ID, class ShipClass, teamID int, position physics.Vector2D) *Ship {
	stats := getShipStats(class)

	ship := &Ship{
		BaseEntity: BaseEntity{
			ID:       id,
			Position: position,
			Rotation: 0,
			Collider: physics.Circle{
				Center: position,
				Radius: 20, // Adjust based on ship class
			},
			Active: true,
		},
		Class:     class,
		Stats:     stats,
		TeamID:    teamID,
		Hull:      stats.MaxHull,
		Shields:   stats.MaxShields,
		Fuel:      stats.MaxFuel,
		Weapons:   make([]Weapon, 0, stats.WeaponSlots),
		LastFired: make(map[string]time.Time),
	}

	// Add default weapons
	ship.Weapons = append(ship.Weapons, NewTorpedo(id))
	ship.Weapons = append(ship.Weapons, NewPhaser(id))

	return ship
}

// Update handles the ship's state update for a single game tick
func (s *Ship) Update(deltaTime float64) {
	// Handle rotation
	if s.TurningCW {
		s.Rotation += s.Stats.TurnRate * deltaTime
	}
	if s.TurningCCW {
		s.Rotation -= s.Stats.TurnRate * deltaTime
	}

	// Handle acceleration
	if s.Thrusting && s.Fuel > 0 {
		// Calculate acceleration vector based on ship heading
		accelVector := physics.FromAngle(s.Rotation, s.Stats.Acceleration)
		s.Velocity = s.Velocity.Add(accelVector.Scale(deltaTime))

		// Cap speed at max speed
		if s.Velocity.Length() > s.Stats.MaxSpeed {
			s.Velocity = s.Velocity.Normalize().Scale(s.Stats.MaxSpeed)
		}

		// Consume fuel
		s.Fuel--
	}

	// Apply drag
	drag := 0.1 // Drag coefficient
	s.Velocity = s.Velocity.Scale(1.0 - drag*deltaTime)

	// Update position
	s.BaseEntity.Update(deltaTime)

	// Update weapons cooldowns
	// Regenerate shields
	if s.Shields < s.Stats.MaxShields {
		s.Shields += int(deltaTime * 5) // Adjust shield regen rate
		if s.Shields > s.Stats.MaxShields {
			s.Shields = s.Stats.MaxShields
		}
	}
}

// FireWeapon attempts to fire the specified weapon
func (s *Ship) FireWeapon(weaponIndex int) *Projectile {
	if weaponIndex < 0 || weaponIndex >= len(s.Weapons) {
		return nil
	}

	weapon := s.Weapons[weaponIndex]
	now := time.Now()

	// Check cooldown
	lastFired, exists := s.LastFired[weapon.GetName()]
	if exists && now.Sub(lastFired) < weapon.GetCooldown() {
		return nil // Weapon still on cooldown
	}

	// Check fuel/energy requirements
	if s.Fuel < weapon.GetFuelCost() {
		return nil // Not enough fuel
	}

	// Create projectile
	projectile := weapon.CreateProjectile(s.ID, s.Position, s.Rotation, s.TeamID)

	// Update cooldown and consume resources
	s.LastFired[weapon.GetName()] = now
	s.Fuel -= weapon.GetFuelCost()

	return projectile
}

// TakeDamage applies damage to the ship, first to shields then to hull
func (s *Ship) TakeDamage(amount int) bool {
	// Apply to shields first
	if s.Shields > 0 {
		if s.Shields >= amount {
			s.Shields -= amount
			amount = 0
		} else {
			amount -= s.Shields
			s.Shields = 0
		}
	}

	// Apply remaining damage to hull
	if amount > 0 {
		s.Hull -= amount
	}

	// Check if ship is destroyed
	return s.Hull <= 0
}

func (s *Ship) RepairTick(deltaTime float64) {
	if !s.RepairMode {
		return
	}

	repairRate := float64(s.Stats.MaxHull) * 0.1 * deltaTime // 10% per second

	// Repair hull
	if s.Hull < s.Stats.MaxHull {
		s.Hull += int(repairRate)
		if s.Hull > s.Stats.MaxHull {
			s.Hull = s.Stats.MaxHull
		}
	}

	// Consume fuel while repairing
	s.Fuel -= int(deltaTime * 5) // 5 fuel per second
	if s.Fuel < 0 {
		s.Fuel = 0
		s.RepairMode = false
	}
}

// getShipStats returns the base statistics for a ship class
func getShipStats(class ShipClass) ShipStats {
	switch class {
	case Scout:
		return ShipStats{
			MaxHull:      100,
			MaxShields:   100,
			MaxFuel:      1000,
			Acceleration: 200,
			TurnRate:     3.0,
			MaxSpeed:     300,
			WeaponSlots:  2,
			MaxArmies:    2,
		}
	case Destroyer:
		return ShipStats{
			MaxHull:      150,
			MaxShields:   150,
			MaxFuel:      1200,
			Acceleration: 150,
			TurnRate:     2.5,
			MaxSpeed:     250,
			WeaponSlots:  3,
			MaxArmies:    5,
		}
	// Other ship classes would be defined here
	default:
		return ShipStats{
			MaxHull:      120,
			MaxShields:   120,
			MaxFuel:      1000,
			Acceleration: 180,
			TurnRate:     2.8,
			MaxSpeed:     280,
			WeaponSlots:  2,
			MaxArmies:    3,
		}
	}
}

// ShipClassFromString converts a string to a ShipClass enum value.
func ShipClassFromString(s string) ShipClass {
	switch s {
	case "Scout":
		return Scout
	case "Destroyer":
		return Destroyer
	case "Cruiser":
		return Cruiser
	case "Battleship":
		return Battleship
	case "Assault":
		return Assault
	default:
		return Scout // fallback to Scout if unknown
	}
}
