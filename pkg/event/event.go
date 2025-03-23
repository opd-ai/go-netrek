// pkg/event/event.go
package event

import (
	"sync"
)

// Type represents the type of event
type Type string

// Common event types
const (
	ShipCreated      Type = "ship_created"
	ShipDestroyed    Type = "ship_destroyed"
	PlanetCaptured   Type = "planet_captured"
	ProjectileFired  Type = "projectile_fired"
	EntityCollision  Type = "entity_collision"
	PlayerJoined     Type = "player_joined"
	PlayerLeft       Type = "player_left"
	GameStarted      Type = "game_started"
	GameEnded        Type = "game_ended"
	TeamScoreChanged Type = "team_score_changed"
)

// Event is the base interface for all events
type Event interface {
	GetType() Type
	GetSource() interface{}
}

// BaseEvent provides common functionality for all events
type BaseEvent struct {
	EventType Type
	Source    interface{}
}

// GetType returns the event type
func (e *BaseEvent) GetType() Type {
	return e.EventType
}

// GetSource returns the event source
func (e *BaseEvent) GetSource() interface{} {
	return e.Source
}

// Handler is a function that handles events
type Handler func(Event)

// Bus manages event subscriptions and dispatching
type Bus struct {
	handlers map[Type][]Handler
	mu       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *Bus {
	return &Bus{
		handlers: make(map[Type][]Handler),
	}
}

// Subscribe registers a handler for a specific event type
func (b *Bus) Subscribe(eventType Type, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Unsubscribe removes a handler for a specific event type
func (b *Bus) Unsubscribe(eventType Type, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers, ok := b.handlers[eventType]
	if !ok {
		return
	}

	// Find and remove the handler
	for i, h := range handlers {
		if &h == &handler {
			b.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish sends an event to all subscribed handlers
func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	handlers, ok := b.handlers[event.GetType()]
	b.mu.RUnlock()

	if !ok {
		return
	}

	// Call each handler
	for _, handler := range handlers {
		handler(event)
	}
}

// Specific event implementations

// ShipEvent contains information about ship-related events
type ShipEvent struct {
	BaseEvent
	ShipID uint64
	TeamID int
}

// NewShipEvent creates a new ship event
func NewShipEvent(eventType Type, source interface{}, shipID uint64, teamID int) *ShipEvent {
	return &ShipEvent{
		BaseEvent: BaseEvent{
			EventType: eventType,
			Source:    source,
		},
		ShipID: shipID,
		TeamID: teamID,
	}
}

// PlanetEvent contains information about planet-related events
type PlanetEvent struct {
	BaseEvent
	PlanetID  uint64
	TeamID    int
	OldTeamID int
}

// NewPlanetEvent creates a new planet event
func NewPlanetEvent(eventType Type, source interface{}, planetID uint64, teamID, oldTeamID int) *PlanetEvent {
	return &PlanetEvent{
		BaseEvent: BaseEvent{
			EventType: eventType,
			Source:    source,
		},
		PlanetID:  planetID,
		TeamID:    teamID,
		OldTeamID: oldTeamID,
	}
}

// CollisionEvent contains information about entity collisions
type CollisionEvent struct {
	BaseEvent
	EntityA uint64
	EntityB uint64
}

// NewCollisionEvent creates a new collision event
func NewCollisionEvent(source interface{}, entityA, entityB uint64) *CollisionEvent {
	return &CollisionEvent{
		BaseEvent: BaseEvent{
			EventType: EntityCollision,
			Source:    source,
		},
		EntityA: entityA,
		EntityB: entityB,
	}
}
