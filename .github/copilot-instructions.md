# go-netrek Project Copilot Instructions

## Project Overview

I am a Go programming assistant specializing in the go-netrek project - a modern implementation of the classic multiplayer space strategy game Netrek. This project aims to deliver:

- A clean, idiomatic Go implementation of the Netrek game
- Multiplayer capabilities with client-server architecture
- Team-based strategic space combat gameplay
- Planet conquest and resource management systems
- Extensible, event-driven game architecture

## Project Structure

The codebase follows a modular structure with clear separation of concerns:

```
go-netrek/
├── cmd/
│   ├── client/     # Client application
│   └── server/     # Server application
├── pkg/
│   ├── entity/     # Game objects (ships, planets, weapons)
│   ├── physics/    # Vector math and collision detection
│   ├── event/      # Event system for game state changes
│   ├── engine/     # Core game logic and state management
│   ├── network/    # Client-server communication
│   ├── config/     # Game configuration handling
│   └── render/     # Game rendering interfaces
```

## Core Implementation Focus

When working with this codebase, prioritize these core components:

1. **Game Engine** - Core game loop, state management, and entity tracking
2. **Entity System** - Ships, planets, weapons, and their interactions
3. **Physics System** - Movement, collision detection, and spatial partitioning
4. **Networking** - Client-server communication for multiplayer gameplay
5. **Event System** - Message passing for decoupled game components

## Code Style Guidelines

All implementations should adhere to these principles:

1. **Idiomatic Go** - Follow Go conventions and best practices
2. **Memory Management** - Leverage Go's garbage collection effectively
3. **Concurrency** - Use goroutines and channels appropriately for concurrent game systems
4. **Error Handling** - Implement proper Go-style error handling and propagation
5. **API Design** - Maintain a clean, consistent API following Go conventions

## Networking Best Practices

When declaring network variables, always use interface types:
 - never use net.UDPAddr, net.IPAddr, or net.TCPAddr. Use net.Addr only instead.
 - never use net.UDPConn, use net.PacketConn instead
 - never use net.TCPConn, use net.Conn instead
 - never use net.UDPListener net.TCPListener, use net.Listener instead

This approach enhances testability and flexibility when working with different network implementations or mocks.

## Implementation Pattern

The typical usage pattern for go-netrek server follows this structure:

```go
// Create game configuration
gameConfig := config.DefaultConfig()

// Create game engine
game := engine.NewGame(gameConfig)

// Create and start server
server := network.NewGameServer(game, gameConfig.MaxPlayers)
server.Start(serverAddr)

// Start game loop
for game.Running {
    game.Update()
    time.Sleep(game.TimeStep)
}
```

For client-side implementation:

```go
// Create event bus for communication
eventBus := event.NewEventBus()

// Create and connect client
client := network.NewGameClient(eventBus)
client.Connect(serverAddr, playerName, teamID)

// Handle game state updates
for gameState := range client.GetGameStateChannel() {
    // Update local game state
    // Render game elements
}

// Send player input
client.SendInput(thrust, turnLeft, turnRight, fireWeapon, beamDown, beamUp, beamAmount, targetID)
```

## Game Physics Implementation

The physics package should implement necessary operations including:
- Vector math for movement and positioning
- Collision detection between entities
- Spatial partitioning via QuadTree for efficient collision queries
- Movement mechanics for different ship classes

## Implementation Considerations

When implementing or reviewing code:

1. **Game Balance** - Consider gameplay balance in ship and weapon statistics
2. **Performance** - Optimize for smooth gameplay with many entities
3. **Networking Efficiency** - Minimize bandwidth usage with partial state updates
4. **Cross-Platform** - Ensure the code works across different operating systems
5. **Testing** - Write comprehensive unit tests for game logic components

## Config Management

Follow established patterns for configuration management:
- Use struct-based configuration options
- Provide sensible defaults for game settings
- Allow overriding through command line arguments
- Support JSON-based configuration files

## Entity System

The entity system is a core aspect of go-netrek. When implementing game entities:

1. Use composition over inheritance
2. Implement the Entity interface for all game objects
3. Use appropriate collision shapes for different entity types
4. Maintain proper entity lifecycle management
5. Consider performance implications of entity interactions

## Event System Usage

The event system facilitates communication between game components:

```go
// Subscribe to events
eventBus.Subscribe(event.ShipDestroyed, func(e event.Event) {
    if shipEvent, ok := e.(*event.ShipEvent); ok {
        // Handle ship destruction
        updateScore(shipEvent.TeamID)
    }
})

// Publish events
eventBus.Publish(event.NewShipEvent(
    event.ShipDestroyed,
    source,
    shipID,
    teamID,
))
```

## Contributing Guidelines

When contributing to the project:

1. Ensure code follows the project's style and patterns
2. Provide comprehensive documentation for public APIs
3. Include unit tests for new functionality
4. Consider performance implications for the game loop
5. Maintain compatibility with the existing game mechanics