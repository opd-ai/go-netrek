// pkg/event/event.go
package event

import (
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
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

// getEventCallerInfo returns the calling function name for event logging
func getEventCallerInfo() string {
	if pc, _, _, ok := runtime.Caller(2); ok {
		return runtime.FuncForPC(pc).Name()
	}
	return "unknown"
}

// eventLogger is the package-level logger for event operations
var eventLogger *logrus.Logger

func init() {
	eventLogger = logrus.New()
	eventLogger.SetFormatter(&logrus.JSONFormatter{})
	eventLogger.SetReportCaller(true)
}

// Subscription represents a subscription token
type Subscription struct {
	ID     uint64
	Cancel func()
}

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

// handlerWrapper wraps a handler with an ID for unsubscription
type handlerWrapper struct {
	id      uint64
	handler Handler
}

// Bus manages event subscriptions and dispatching
type Bus struct {
	handlers map[Type][]handlerWrapper
	nextID   uint64
	mu       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *Bus {
	caller := getEventCallerInfo()
	eventLogger.WithField("caller", caller).WithField("function", "NewEventBus").Info("Creating new event bus")

	bus := &Bus{
		handlers: make(map[Type][]handlerWrapper),
		nextID:   1,
	}

	eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
		"function": "NewEventBus",
		"next_id":  bus.nextID,
	}).Debug("Event bus initialized successfully")

	return bus
}

// Subscribe registers a handler for a specific event type and returns a subscription token
func (b *Bus) Subscribe(eventType Type, handler Handler) *Subscription {
	caller := getEventCallerInfo()
	eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":   "Subscribe",
		"event_type": string(eventType),
	}).Info("Subscribing to event type")

	eventLogger.WithField("caller", caller).WithField("function", "Subscribe").Debug("Acquiring event bus lock")
	b.mu.Lock()
	defer func() {
		b.mu.Unlock()
		eventLogger.WithField("caller", caller).WithField("function", "Subscribe").Debug("Released event bus lock")
	}()

	id := b.nextID
	b.nextID++

	eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":        "Subscribe",
		"event_type":      string(eventType),
		"subscription_id": id,
		"next_id":         b.nextID,
	}).Debug("Generated subscription ID")

	wrapper := handlerWrapper{
		id:      id,
		handler: handler,
	}

	b.handlers[eventType] = append(b.handlers[eventType], wrapper)

	currentHandlerCount := len(b.handlers[eventType])
	eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":      "Subscribe",
		"event_type":    string(eventType),
		"handler_count": currentHandlerCount,
	}).Debug("Handler added to event type")

	sub := &Subscription{
		ID: id,
		Cancel: func() {
			eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
				"function":        "Subscribe.Cancel",
				"event_type":      string(eventType),
				"subscription_id": id,
			}).Debug("Cancelling subscription")
			b.unsubscribeByID(eventType, id)
		},
	}

	eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":        "Subscribe",
		"event_type":      string(eventType),
		"subscription_id": id,
	}).Info("Successfully subscribed to event type")

	return sub
}

// unsubscribeByID removes a handler by its ID
func (b *Bus) unsubscribeByID(eventType Type, id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers, ok := b.handlers[eventType]
	if !ok {
		return
	}

	// Find and remove the handler with matching ID
	for i, wrapper := range handlers {
		if wrapper.id == id {
			b.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Unsubscribe removes a handler for a specific event type (deprecated - use Subscription.Cancel instead)
func (b *Bus) Unsubscribe(eventType Type, handler Handler) {
	// This method is kept for backward compatibility but cannot work reliably
	// Users should use the Subscription.Cancel() method returned from Subscribe
}

// Publish sends an event to all subscribed handlers
func (b *Bus) Publish(event Event) {
	caller := getEventCallerInfo()
	eventType := event.GetType()

	eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":   "Publish",
		"event_type": string(eventType),
	}).Info("Publishing event")

	eventLogger.WithField("caller", caller).WithField("function", "Publish").Debug("Acquiring read lock for event handlers")
	b.mu.RLock()
	handlers, ok := b.handlers[eventType]
	handlerCount := len(handlers)
	b.mu.RUnlock()
	eventLogger.WithField("caller", caller).WithField("function", "Publish").Debug("Released read lock for event handlers")

	if !ok {
		eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
			"function":   "Publish",
			"event_type": string(eventType),
		}).Debug("No handlers registered for event type")
		return
	}

	eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":      "Publish",
		"event_type":    string(eventType),
		"handler_count": handlerCount,
	}).Debug("Found handlers for event type")

	// Call each handler
	for i, wrapper := range handlers {
		eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
			"function":      "Publish",
			"event_type":    string(eventType),
			"handler_index": i,
			"handler_id":    wrapper.id,
		}).Debug("Calling event handler")

		wrapper.handler(event)

		eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
			"function":      "Publish",
			"event_type":    string(eventType),
			"handler_index": i,
			"handler_id":    wrapper.id,
		}).Debug("Event handler completed")
	}

	eventLogger.WithField("caller", caller).WithFields(logrus.Fields{
		"function":        "Publish",
		"event_type":      string(eventType),
		"handlers_called": handlerCount,
	}).Info("Event published to all handlers successfully")
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
