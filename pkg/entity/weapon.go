// pkg/entity/weapon.go
package entity

import (
	"time"

	"github.com/opd-ai/go-netrek/pkg/physics"
)

// Weapon interface defines the methods all weapons must implement
type Weapon interface {
	GetName() string
	GetCooldown() time.Duration
	GetFuelCost() int
	CreateProjectile(ownerID ID, position physics.Vector2D, angle float64, teamID int) *Projectile
}

// BaseWeapon contains common functionality for all weapons
type BaseWeapon struct {
	Name     string
	Cooldown time.Duration
	FuelCost int
	Damage   int
	Speed    float64
	OwnerID  ID
}

// GetName returns the weapon's name
func (w *BaseWeapon) GetName() string {
	return w.Name
}

// GetCooldown returns the weapon's cooldown time
func (w *BaseWeapon) GetCooldown() time.Duration {
	return w.Cooldown
}

// GetFuelCost returns the fuel cost to fire the weapon
func (w *BaseWeapon) GetFuelCost() int {
	return w.FuelCost
}

// Torpedo weapon implementation
type Torpedo struct {
	BaseWeapon
	Range float64
}

// NewTorpedo creates a new torpedo weapon
func NewTorpedo(ownerID ID) *Torpedo {
	return &Torpedo{
		BaseWeapon: BaseWeapon{
			Name:     "Torpedo",
			Cooldown: 500 * time.Millisecond,
			FuelCost: 10,
			Damage:   40,
			Speed:    500,
			OwnerID:  ownerID,
		},
		Range: 2000,
	}
}

// CreateProjectile creates a torpedo projectile
func (t *Torpedo) CreateProjectile(ownerID ID, position physics.Vector2D, angle float64, teamID int) *Projectile {
	return &Projectile{
		BaseEntity: BaseEntity{
			ID:       GenerateID(),
			Position: position,
			Velocity: physics.FromAngle(angle, t.Speed),
			Rotation: angle,
			Collider: physics.Circle{
				Center: position,
				Radius: 5,
			},
			Active: true,
		},
		Type:             "Torpedo",
		OwnerID:          ownerID,
		TeamID:           teamID,
		Damage:           t.Damage,
		Range:            t.Range,
		DistanceTraveled: 0,
	}
}

// Phaser weapon implementation
type Phaser struct {
	BaseWeapon
	Range float64
}

// NewPhaser creates a new phaser weapon
func NewPhaser(ownerID ID) *Phaser {
	return &Phaser{
		BaseWeapon: BaseWeapon{
			Name:     "Phaser",
			Cooldown: 200 * time.Millisecond,
			FuelCost: 5,
			Damage:   20,
			Speed:    1000,
			OwnerID:  ownerID,
		},
		Range: 800,
	}
}

// CreateProjectile creates a phaser projectile
func (p *Phaser) CreateProjectile(ownerID ID, position physics.Vector2D, angle float64, teamID int) *Projectile {
	return &Projectile{
		BaseEntity: BaseEntity{
			ID:       GenerateID(),
			Position: position,
			Velocity: physics.FromAngle(angle, p.Speed),
			Rotation: angle,
			Collider: physics.Circle{
				Center: position,
				Radius: 3,
			},
			Active: true,
		},
		Type:             "Phaser",
		OwnerID:          ownerID,
		TeamID:           teamID,
		Damage:           p.Damage,
		Range:            p.Range,
		DistanceTraveled: 0,
	}
}

// Projectile represents a weapon projectile in the game
type Projectile struct {
	BaseEntity
	Type             string
	OwnerID          ID
	TeamID           int
	Damage           int
	Range            float64
	DistanceTraveled float64
}

// Update updates the projectile's position and checks if it has exceeded its range
func (p *Projectile) Update(deltaTime float64) {
	oldPos := p.Position
	p.BaseEntity.Update(deltaTime)

	// Calculate distance traveled in this frame
	distThisFrame := oldPos.Distance(p.Position)
	p.DistanceTraveled += distThisFrame

	// Deactivate if exceeded range
	if p.DistanceTraveled >= p.Range {
		p.Active = false
	}
}

// GenerateID generates a unique ID for entities
// In a real implementation, this would use a more robust approach
var nextID ID = 1

func GenerateID() ID {
	id := nextID
	nextID++
	return id
}
