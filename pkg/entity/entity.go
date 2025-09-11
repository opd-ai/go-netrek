// pkg/entity/entity.go
package entity

import (
	"runtime"

	"github.com/opd-ai/go-netrek/pkg/physics"
	"github.com/sirupsen/logrus"
)

// ID is a unique identifier for an entity
type ID uint64

// getCallerInfo returns the calling function name for logging context
func getCallerInfo() string {
	if pc, _, _, ok := runtime.Caller(2); ok {
		return runtime.FuncForPC(pc).Name()
	}
	return "unknown"
}

// logger is the package-level logger for entity operations
var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetReportCaller(true)
}

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
	caller := getCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":  "GetID",
		"entity_id": e.ID,
	}).Debug("Retrieving entity ID")
	return e.ID
}

// GetPosition returns the entity's position
func (e *BaseEntity) GetPosition() physics.Vector2D {
	caller := getCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":   "GetPosition",
		"entity_id":  e.ID,
		"position_x": e.Position.X,
		"position_y": e.Position.Y,
	}).Debug("Retrieving entity position")
	return e.Position
}

// GetCollider returns the entity's collision shape
func (e *BaseEntity) GetCollider() physics.Circle {
	caller := getCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":        "GetCollider",
		"entity_id":       e.ID,
		"collider_radius": e.Collider.Radius,
		"center_x":        e.Position.X,
		"center_y":        e.Position.Y,
	}).Debug("Retrieving entity collider")

	collider := physics.Circle{
		Center: e.Position,
		Radius: e.Collider.Radius,
	}

	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":     "GetCollider",
		"entity_id":    e.ID,
		"result_valid": collider.Radius > 0,
	}).Debug("Collider created successfully")

	return collider
}

// Update updates the entity's position based on velocity
func (e *BaseEntity) Update(deltaTime float64) {
	caller := getCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":      "Update",
		"entity_id":     e.ID,
		"delta_time":    deltaTime,
		"initial_pos_x": e.Position.X,
		"initial_pos_y": e.Position.Y,
		"velocity_x":    e.Velocity.X,
		"velocity_y":    e.Velocity.Y,
		"entity_active": e.Active,
	}).Debug("Starting entity update")

	if !e.Active {
		logger.WithField("caller", caller).WithFields(logrus.Fields{
			"function":  "Update",
			"entity_id": e.ID,
		}).Debug("Entity is inactive, skipping update")
		return
	}

	// Calculate velocity-based movement
	velocityDelta := e.Velocity.Scale(deltaTime)
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":         "Update",
		"entity_id":        e.ID,
		"velocity_delta_x": velocityDelta.X,
		"velocity_delta_y": velocityDelta.Y,
	}).Debug("Calculated velocity delta")

	// Update position
	e.Position = e.Position.Add(velocityDelta)
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":  "Update",
		"entity_id": e.ID,
		"new_pos_x": e.Position.X,
		"new_pos_y": e.Position.Y,
	}).Debug("Updated entity position")

	// Update collider position
	e.Collider.Center = e.Position
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":  "Update",
		"entity_id": e.ID,
	}).Debug("Updated collider center to match position")

	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":  "Update",
		"entity_id": e.ID,
	}).Debug("Entity update completed successfully")
}

// Implement the missing method in BaseEntity
func (e *BaseEntity) Render(r Renderer) {
	caller := getCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":  "Render",
		"entity_id": e.ID,
		"renderer":  "BaseEntity",
	}).Debug("Base entity render called (no-op)")
	// Base implementation does nothing, derived types will implement
}

// Then implement the Entity.Render() method in each entity type:
func (s *Ship) Render(r Renderer) {
	caller := getCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function": "Render",
		"ship_id":  s.GetID(),
		"renderer": "Ship",
	}).Debug("Rendering ship entity")
	r.RenderShip(s)
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function": "Render",
		"ship_id":  s.GetID(),
	}).Debug("Ship rendering completed")
}

func (p *Planet) Render(r Renderer) {
	caller := getCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":    "Render",
		"planet_id":   p.GetID(),
		"renderer":    "Planet",
		"planet_name": p.Name,
	}).Debug("Rendering planet entity")
	r.RenderPlanet(p)
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":  "Render",
		"planet_id": p.GetID(),
	}).Debug("Planet rendering completed")
}

func (p *Projectile) Render(r Renderer) {
	caller := getCallerInfo()
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":      "Render",
		"projectile_id": p.GetID(),
		"renderer":      "Projectile",
	}).Debug("Rendering projectile entity")
	r.RenderProjectile(p)
	logger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":      "Render",
		"projectile_id": p.GetID(),
	}).Debug("Projectile rendering completed")
}
