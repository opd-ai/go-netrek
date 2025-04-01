// pkg/entity/planet.go
package entity

import (
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// PlanetType defines the type of planet
type PlanetType int

const (
	Agricultural PlanetType = iota
	Industrial
	Military
	Homeworld
)

// Planet represents a planet in the Netrek galaxy
type Planet struct {
	BaseEntity
	Name        string
	Type        PlanetType
	TeamID      int // -1 for neutral
	Armies      int
	Resources   int
	Production  int
	Temperature int  // Affects bombing effectiveness
	Atmosphere  bool // Affects bombing
	MaxArmies   int
}

// NewPlanet creates a new planet
func NewPlanet(id ID, name string, position physics.Vector2D, planetType PlanetType) *Planet {
	planet := &Planet{
		BaseEntity: BaseEntity{
			ID:       id,
			Position: position,
			Collider: physics.Circle{
				Center: position,
				Radius: 50, // Planet size
			},
			Active: true,
		},
		Name:        name,
		Type:        planetType,
		TeamID:      -1, // Start neutral
		Armies:      10, // Default armies
		Resources:   100,
		Production:  5,
		Temperature: 50,
		Atmosphere:  true,
		MaxArmies:   100,
	}

	// Adjust properties based on planet type
	switch planetType {
	case Agricultural:
		planet.Production = 8
		planet.MaxArmies = 80
	case Industrial:
		planet.Resources = 200
		planet.Production = 10
	case Military:
		planet.Armies = 20
		planet.MaxArmies = 120
	case Homeworld:
		planet.Armies = 30
		planet.MaxArmies = 200
		planet.Production = 12
	}

	return planet
}

// Update handles the planet's state update for a single game tick
func (p *Planet) Update(deltaTime float64) {
	// Planets don't move, but they produce armies
	if p.TeamID >= 0 { // If owned by a team
		productionRate := float64(p.Production) * deltaTime / 10.0 // Armies per second
		newArmies := int(productionRate)

		if newArmies > 0 && p.Armies < p.MaxArmies {
			p.Armies += newArmies
			if p.Armies > p.MaxArmies {
				p.Armies = p.MaxArmies
			}
		}
	}
}

// Bomb reduces the number of armies on the planet
func (p *Planet) Bomb(damage int) int {
	// Can't bomb a planet with your own team ID
	if p.Armies <= 0 {
		return 0
	}

	// Adjust damage based on planet properties
	if p.Atmosphere {
		damage = int(float64(damage) * 0.7) // Atmosphere reduces bombing effectiveness
	}

	// Apply temperature factor
	tempFactor := 1.0 - (float64(p.Temperature)/100.0)*0.5
	damage = int(float64(damage) * tempFactor)

	// Ensure at least 1 damage if bombing
	if damage <= 0 {
		damage = 1
	}

	// Apply damage to armies
	if damage > p.Armies {
		damage = p.Armies
	}

	p.Armies -= damage

	// If all armies are destroyed and planet was owned, make it neutral
	if p.Armies == 0 && p.TeamID >= 0 {
		p.TeamID = -1
	}

	return damage
}

// BeamDownArmies transfers armies from a ship to the planet
func (p *Planet) BeamDownArmies(shipTeamID, amount int) (transferred int, captured bool) {
	// Can't beam to enemy planets without conquering
	if p.TeamID >= 0 && p.TeamID != shipTeamID {
		return 0, false
	}

	// If planet is neutral or friendly
	if p.TeamID == -1 || p.TeamID == shipTeamID {
		// Limit transfer to available space
		spaceAvailable := p.MaxArmies - p.Armies
		if amount > spaceAvailable {
			amount = spaceAvailable
		}

		p.Armies += amount

		// If planet was neutral and now has armies, it's captured
		if p.TeamID == -1 && p.Armies > 0 {
			p.TeamID = shipTeamID
			captured = true
		}

		return amount, captured
	}

	return 0, false
}

// BeamUpArmies transfers armies from the planet to a ship
func (p *Planet) BeamUpArmies(shipTeamID, maxAmount int) int {
	// Can only beam up from friendly planets
	if p.TeamID != shipTeamID || p.Armies <= 0 {
		return 0
	}

	// Determine how many armies to transfer
	amount := maxAmount
	if amount > p.Armies {
		amount = p.Armies
	}

	p.Armies -= amount
	return amount
}
