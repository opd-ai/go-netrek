// pkg/event/event_test.go
package event

import (
	"sync"
	"testing"
	"time"
)

// TestNewEventBus tests the creation of a new event bus
func TestNewEventBus_Creation_ReturnsInitializedBus(t *testing.T) {
	bus := NewEventBus()

	if bus == nil {
		t.Fatal("NewEventBus() returned nil")
	}

	if bus.handlers == nil {
		t.Error("handlers map not initialized")
	}

	if bus.nextID != 1 {
		t.Errorf("expected nextID to be 1, got %d", bus.nextID)
	}
}

// TestBaseEvent tests the BaseEvent functionality
func TestBaseEvent_GetType_ReturnsCorrectType(t *testing.T) {
	tests := []struct {
		name      string
		eventType Type
		source    interface{}
	}{
		{
			name:      "ShipCreated event",
			eventType: ShipCreated,
			source:    "test_source",
		},
		{
			name:      "PlanetCaptured event",
			eventType: PlanetCaptured,
			source:    123,
		},
		{
			name:      "Empty source",
			eventType: GameStarted,
			source:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &BaseEvent{
				EventType: tt.eventType,
				Source:    tt.source,
			}

			if event.GetType() != tt.eventType {
				t.Errorf("GetType() = %v, want %v", event.GetType(), tt.eventType)
			}

			if event.GetSource() != tt.source {
				t.Errorf("GetSource() = %v, want %v", event.GetSource(), tt.source)
			}
		})
	}
}

// TestBusSubscribe tests event subscription functionality
func TestBusSubscribe_SingleHandler_ReturnsValidSubscription(t *testing.T) {
	bus := NewEventBus()

	handler := func(e Event) {
		// Handler for testing subscription
	}

	sub := bus.Subscribe(ShipCreated, handler)

	if sub == nil {
		t.Fatal("Subscribe() returned nil subscription")
	}

	if sub.ID == 0 {
		t.Error("subscription ID should not be 0")
	}

	if sub.Cancel == nil {
		t.Error("subscription Cancel function should not be nil")
	}

	// Verify handler was registered
	bus.mu.RLock()
	handlers := bus.handlers[ShipCreated]
	bus.mu.RUnlock()

	if len(handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(handlers))
	}
}

// TestBusSubscribe_MultipleHandlers tests multiple subscriptions
func TestBusSubscribe_MultipleHandlers_AllRegistered(t *testing.T) {
	bus := NewEventBus()
	var callCount int

	handler1 := func(e Event) { callCount++ }
	handler2 := func(e Event) { callCount++ }
	handler3 := func(e Event) { callCount++ }

	sub1 := bus.Subscribe(ShipCreated, handler1)
	sub2 := bus.Subscribe(ShipCreated, handler2)
	_ = bus.Subscribe(PlanetCaptured, handler3)

	// Check unique IDs
	if sub1.ID == sub2.ID {
		t.Error("subscriptions should have unique IDs")
	}

	// Check handlers count
	bus.mu.RLock()
	shipHandlers := bus.handlers[ShipCreated]
	planetHandlers := bus.handlers[PlanetCaptured]
	bus.mu.RUnlock()

	if len(shipHandlers) != 2 {
		t.Errorf("expected 2 handlers for ShipCreated, got %d", len(shipHandlers))
	}

	if len(planetHandlers) != 1 {
		t.Errorf("expected 1 handler for PlanetCaptured, got %d", len(planetHandlers))
	}
}

// TestBusPublish tests event publishing functionality
func TestBusPublish_WithSubscribers_CallsAllHandlers(t *testing.T) {
	bus := NewEventBus()
	var callCount int
	var receivedEvents []Event

	handler1 := func(e Event) {
		callCount++
		receivedEvents = append(receivedEvents, e)
	}

	handler2 := func(e Event) {
		callCount++
		receivedEvents = append(receivedEvents, e)
	}

	bus.Subscribe(ShipCreated, handler1)
	bus.Subscribe(ShipCreated, handler2)

	event := &BaseEvent{
		EventType: ShipCreated,
		Source:    "test",
	}

	bus.Publish(event)

	if callCount != 2 {
		t.Errorf("expected 2 handler calls, got %d", callCount)
	}

	if len(receivedEvents) != 2 {
		t.Errorf("expected 2 received events, got %d", len(receivedEvents))
	}

	for _, e := range receivedEvents {
		if e.GetType() != ShipCreated {
			t.Errorf("expected event type %v, got %v", ShipCreated, e.GetType())
		}
	}
}

// TestBusPublish_NoSubscribers tests publishing without subscribers
func TestBusPublish_NoSubscribers_NoError(t *testing.T) {
	bus := NewEventBus()

	event := &BaseEvent{
		EventType: ShipCreated,
		Source:    "test",
	}

	// Should not panic or error
	bus.Publish(event)
}

// TestBusPublish_WrongEventType tests publishing to non-subscribed event type
func TestBusPublish_WrongEventType_HandlersNotCalled(t *testing.T) {
	bus := NewEventBus()
	handlerCalled := false

	handler := func(e Event) {
		handlerCalled = true
	}

	bus.Subscribe(ShipCreated, handler)

	event := &BaseEvent{
		EventType: PlanetCaptured,
		Source:    "test",
	}

	bus.Publish(event)

	if handlerCalled {
		t.Error("handler should not have been called for different event type")
	}
}

// TestSubscriptionCancel tests canceling subscriptions
func TestSubscriptionCancel_ValidSubscription_RemovesHandler(t *testing.T) {
	bus := NewEventBus()
	handlerCalled := false

	handler := func(e Event) {
		handlerCalled = true
	}

	sub := bus.Subscribe(ShipCreated, handler)

	// Verify handler is registered
	bus.mu.RLock()
	handlersBefore := len(bus.handlers[ShipCreated])
	bus.mu.RUnlock()

	if handlersBefore != 1 {
		t.Errorf("expected 1 handler before cancel, got %d", handlersBefore)
	}

	// Cancel subscription
	sub.Cancel()

	// Verify handler is removed
	bus.mu.RLock()
	handlersAfter := len(bus.handlers[ShipCreated])
	bus.mu.RUnlock()

	if handlersAfter != 0 {
		t.Errorf("expected 0 handlers after cancel, got %d", handlersAfter)
	}

	// Verify handler is not called after cancellation
	event := &BaseEvent{
		EventType: ShipCreated,
		Source:    "test",
	}

	bus.Publish(event)

	if handlerCalled {
		t.Error("handler should not be called after cancellation")
	}
}

// TestConcurrentAccess tests thread safety
func TestBusSubscribe_ConcurrentAccess_ThreadSafe(t *testing.T) {
	bus := NewEventBus()
	var wg sync.WaitGroup
	handlerCount := 0
	var mu sync.Mutex

	handler := func(e Event) {
		mu.Lock()
		handlerCount++
		mu.Unlock()
	}

	// Start multiple goroutines to subscribe concurrently
	numGoroutines := 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			bus.Subscribe(ShipCreated, handler)
		}()
	}

	wg.Wait()

	// Verify all subscriptions were registered
	bus.mu.RLock()
	handlers := bus.handlers[ShipCreated]
	bus.mu.RUnlock()

	if len(handlers) != numGoroutines {
		t.Errorf("expected %d handlers, got %d", numGoroutines, len(handlers))
	}

	// Test concurrent publishing
	event := &BaseEvent{
		EventType: ShipCreated,
		Source:    "test",
	}

	// Publish concurrently
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			bus.Publish(event)
		}()
	}

	wg.Wait()

	// Give handlers time to execute
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	expectedCalls := numGoroutines * 3
	if handlerCount != expectedCalls {
		t.Errorf("expected %d handler calls, got %d", expectedCalls, handlerCount)
	}
	mu.Unlock()
}

// TestNewShipEvent tests ship event creation
func TestNewShipEvent_ValidParameters_ReturnsCorrectEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType Type
		source    interface{}
		shipID    uint64
		teamID    int
	}{
		{
			name:      "Ship created event",
			eventType: ShipCreated,
			source:    "game_engine",
			shipID:    12345,
			teamID:    1,
		},
		{
			name:      "Ship destroyed event",
			eventType: ShipDestroyed,
			source:    nil,
			shipID:    67890,
			teamID:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewShipEvent(tt.eventType, tt.source, tt.shipID, tt.teamID)

			if event == nil {
				t.Fatal("NewShipEvent() returned nil")
			}

			if event.GetType() != tt.eventType {
				t.Errorf("GetType() = %v, want %v", event.GetType(), tt.eventType)
			}

			if event.GetSource() != tt.source {
				t.Errorf("GetSource() = %v, want %v", event.GetSource(), tt.source)
			}

			if event.ShipID != tt.shipID {
				t.Errorf("ShipID = %v, want %v", event.ShipID, tt.shipID)
			}

			if event.TeamID != tt.teamID {
				t.Errorf("TeamID = %v, want %v", event.TeamID, tt.teamID)
			}
		})
	}
}

// TestNewPlanetEvent tests planet event creation
func TestNewPlanetEvent_ValidParameters_ReturnsCorrectEvent(t *testing.T) {
	eventType := PlanetCaptured
	source := "combat_system"
	planetID := uint64(555)
	teamID := 3
	oldTeamID := 1

	event := NewPlanetEvent(eventType, source, planetID, teamID, oldTeamID)

	if event == nil {
		t.Fatal("NewPlanetEvent() returned nil")
	}

	if event.GetType() != eventType {
		t.Errorf("GetType() = %v, want %v", event.GetType(), eventType)
	}

	if event.GetSource() != source {
		t.Errorf("GetSource() = %v, want %v", event.GetSource(), source)
	}

	if event.PlanetID != planetID {
		t.Errorf("PlanetID = %v, want %v", event.PlanetID, planetID)
	}

	if event.TeamID != teamID {
		t.Errorf("TeamID = %v, want %v", event.TeamID, teamID)
	}

	if event.OldTeamID != oldTeamID {
		t.Errorf("OldTeamID = %v, want %v", event.OldTeamID, oldTeamID)
	}
}

// TestNewCollisionEvent tests collision event creation
func TestNewCollisionEvent_ValidParameters_ReturnsCorrectEvent(t *testing.T) {
	source := "physics_engine"
	entityA := uint64(100)
	entityB := uint64(200)

	event := NewCollisionEvent(source, entityA, entityB)

	if event == nil {
		t.Fatal("NewCollisionEvent() returned nil")
	}

	if event.GetType() != EntityCollision {
		t.Errorf("GetType() = %v, want %v", event.GetType(), EntityCollision)
	}

	if event.GetSource() != source {
		t.Errorf("GetSource() = %v, want %v", event.GetSource(), source)
	}

	if event.EntityA != entityA {
		t.Errorf("EntityA = %v, want %v", event.EntityA, entityA)
	}

	if event.EntityB != entityB {
		t.Errorf("EntityB = %v, want %v", event.EntityB, entityB)
	}
}

// TestEventTypes tests that all event type constants are properly defined
func TestEventTypes_Constants_AllDefined(t *testing.T) {
	expectedTypes := []Type{
		ShipCreated,
		ShipDestroyed,
		PlanetCaptured,
		ProjectileFired,
		EntityCollision,
		PlayerJoined,
		PlayerLeft,
		GameStarted,
		GameEnded,
		TeamScoreChanged,
	}

	for _, eventType := range expectedTypes {
		if string(eventType) == "" {
			t.Errorf("event type %v is empty", eventType)
		}
	}
}

// TestCancelMultipleSubscriptions tests canceling multiple subscriptions
func TestCancelMultipleSubscriptions_DifferentTypes_OnlyTargetRemoved(t *testing.T) {
	bus := NewEventBus()

	handler1Called := false
	handler2Called := false
	handler3Called := false

	handler1 := func(e Event) { handler1Called = true }
	handler2 := func(e Event) { handler2Called = true }
	handler3 := func(e Event) { handler3Called = true }

	sub1 := bus.Subscribe(ShipCreated, handler1)
	_ = bus.Subscribe(ShipCreated, handler2)
	_ = bus.Subscribe(PlanetCaptured, handler3)

	// Cancel only the first subscription
	sub1.Cancel()

	// Publish ShipCreated event
	shipEvent := &BaseEvent{EventType: ShipCreated, Source: "test"}
	bus.Publish(shipEvent)

	// Publish PlanetCaptured event
	planetEvent := &BaseEvent{EventType: PlanetCaptured, Source: "test"}
	bus.Publish(planetEvent)

	if handler1Called {
		t.Error("handler1 should not be called after cancellation")
	}

	if !handler2Called {
		t.Error("handler2 should be called")
	}

	if !handler3Called {
		t.Error("handler3 should be called")
	}
}
