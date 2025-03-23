# Event Package

The event package provides a robust event system for Go Netrek, enabling communication between different game components through a publish/subscribe pattern.

## Core Components

### Event Types

Pre-defined event types for common game occurrences:

```go
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
```

### Event Bus

The central component for event management:

```go
bus := event.NewEventBus()
```

Features:
- Thread-safe event handling
- Multiple handlers per event type
- Dynamic subscription management

## Event Types

### Base Event
All events implement the Event interface:

```go
type Event interface {
    GetType() Type
    GetSource() interface{}
}
```

### Specific Events

1. **Ship Events**
```go
shipEvent := NewShipEvent(
    ShipCreated,
    source,
    shipID,
    teamID,
)
```

2. **Planet Events**
```go
planetEvent := NewPlanetEvent(
    PlanetCaptured,
    source,
    planetID,
    newTeamID,
    oldTeamID,
)
```

3. **Collision Events**
```go
collisionEvent := NewCollisionEvent(
    source,
    entityAID,
    entityBID,
)
```

## Usage Examples

### Subscribe to Events

```go
// Handle ship creation
bus.Subscribe(event.ShipCreated, func(e event.Event) {
    shipEvent := e.(*event.ShipEvent)
    // Handle new ship
})

// Handle planet captures
bus.Subscribe(event.PlanetCaptured, func(e event.Event) {
    planetEvent := e.(*event.PlanetEvent)
    // Handle planet capture
})
```

### Publish Events

```go
// Publish ship destroyed event
event := NewShipEvent(
    event.ShipDestroyed,
    gameEngine,
    shipID,
    teamID,
)
bus.Publish(event)
```

### Unsubscribe from Events

```go
handler := func(e event.Event) {
    // Event handling
}
bus.Subscribe(event.PlayerJoined, handler)
// Later...
bus.Unsubscribe(event.PlayerJoined, handler)
```

## Best Practices

1. **Type Safety**
   - Always check event types when casting
   - Use specific event structs for different event types

2. **Resource Management**
   - Unsubscribe handlers when no longer needed
   - Avoid memory leaks by cleaning up subscriptions

3. **Event Design**
   - Keep events focused and specific
   - Include all necessary context in event data
   - Avoid circular event dependencies

4. **Error Handling**
   - Handle panics in event handlers
   - Validate event data before publishing

## Thread Safety

The event system is thread-safe and can be used across multiple goroutines:

```go
// Safe to use from multiple goroutines
go func() {
    bus.Publish(event)
}()
```

## Example Game Loop Integration

```go
func gameLoop(bus *event.Bus) {
    // Subscribe to necessary events
    bus.Subscribe(event.EntityCollision, handleCollision)
    bus.Subscribe(event.ShipDestroyed, updateScore)
    
    for {
        // Game update logic
        if collision detected {
            bus.Publish(NewCollisionEvent(...))
        }
        // More game logic
    }
}
```

## Performance Considerations

- Events are processed synchronously
- Handlers should be lightweight
- Heavy processing should be offloaded to separate goroutines
- Consider buffering events for batch processing if needed

## Future Improvements

- Event prioritization
- Async event processing
- Event filtering
- Event replay capabilities
- Event logging and debugging tools