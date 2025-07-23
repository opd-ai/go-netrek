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
	s.updateRotation(deltaTime)
	s.updateAcceleration(deltaTime)
	s.applyDrag(deltaTime)
	s.BaseEntity.Update(deltaTime)
	s.regenerateShields(deltaTime)
}

// updateRotation processes ship rotation based on turn input
func (s *Ship) updateRotation(deltaTime float64) {
	if s.TurningCW {
		s.Rotation += s.Stats.TurnRate * deltaTime
	}
	if s.TurningCCW {
		s.Rotation -= s.Stats.TurnRate * deltaTime
	}
}

// updateAcceleration handles thrust input, acceleration calculation, speed limiting, and fuel consumption
func (s *Ship) updateAcceleration(deltaTime float64) {
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
}

// applyDrag reduces ship velocity over time due to space friction
func (s *Ship) applyDrag(deltaTime float64) {
	drag := 0.1 // Drag coefficient
	s.Velocity = s.Velocity.Scale(1.0 - drag*deltaTime)
}

// regenerateShields increases shield strength over time up to maximum capacity
func (s *Ship) regenerateShields(deltaTime float64) {
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

var shipTypeStats map[string]ShipStats

// SetShipTypeStats allows the config loader to inject custom ship stats
func SetShipTypeStats(stats map[string]ShipStats) {
	shipTypeStats = stats
}

// getShipStats returns the base statistics for a ship class, using config if available
func getShipStats(class ShipClass) ShipStats {
	name := class.String()
	if shipTypeStats != nil {
		if s, ok := shipTypeStats[name]; ok {
			return s
		}
	}
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
	case Cruiser:
		return ShipStats{
			MaxHull:      200,
			MaxShields:   200,
			MaxFuel:      1400,
			Acceleration: 120,
			TurnRate:     2.0,
			MaxSpeed:     220,
			WeaponSlots:  4,
			MaxArmies:    8,
		}
	case Battleship:
		return ShipStats{
			MaxHull:      300,
			MaxShields:   250,
			MaxFuel:      1600,
			Acceleration: 80,
			TurnRate:     1.5,
			MaxSpeed:     180,
			WeaponSlots:  5,
			MaxArmies:    12,
		}
	case Assault:
		return ShipStats{
			MaxHull:      180,
			MaxShields:   120,
			MaxFuel:      1300,
			Acceleration: 140,
			TurnRate:     2.2,
			MaxSpeed:     240,
			WeaponSlots:  3,
			MaxArmies:    15,
		}
	default:
		// Fallback to Scout stats for unknown classes
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
	}
}

// String returns the string name for a ShipClass
func (c ShipClass) String() string {
	switch c {
	case Scout:
		return "Scout"
	case Destroyer:
		return "Destroyer"
	case Cruiser:
		return "Cruiser"
	case Battleship:
		return "Battleship"
	case Assault:
		return "Assault"
	default:
		return "Unknown"
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
