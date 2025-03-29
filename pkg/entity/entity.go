// pkg/entity/entity.go
package entity

import (
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// ID is a unique identifier for an entity
type ID uint64

// Entity is the base interface for all game objects
type Entity interface {
	GetID() ID
	GetPosition() physics.Vector2D
	GetCollider() physics.Circle
	Update(deltaTime float64)
	Render(r Renderer) // Interface for rendering
}

// BaseEntity contains common functionality for all entities
type BaseEntity struct {
	ID       ID
	Position physics.Vector2D
	Velocity physics.Vector2D
	Rotation float64
	Collider physics.Circle
	Active   bool
}

// GetID returns the entity's unique identifier
func (e *BaseEntity) GetID() ID {
	return e.ID
}

// GetPosition returns the entity's position
func (e *BaseEntity) GetPosition() physics.Vector2D {
	return e.Position
}

// GetCollider returns the entity's collision shape
func (e *BaseEntity) GetCollider() physics.Circle {
	return physics.Circle{
		Center: e.Position,
		Radius: e.Collider.Radius,
	}
}

// Update updates the entity's position based on velocity
func (e *BaseEntity) Update(deltaTime float64) {
	e.Position = e.Position.Add(e.Velocity.Scale(deltaTime))
	// Update collider position
	e.Collider.Center = e.Position
}

// Implement the missing method in BaseEntity
func (e *BaseEntity) Render(r Renderer) {
    // Base implementation does nothing, derived types will implement
}

// Then implement the Entity.Render() method in each entity type:
func (s *Ship) Render(r Renderer) {
	r.RenderShip(s)
}

func (p *Planet) Render(r Renderer) {
	r.RenderPlanet(p)
}

func (p *Projectile) Render(r Renderer) {
	r.RenderProjectile(p)
}
