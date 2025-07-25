# Go Netrek

A modern implementation of the classic multiplayer game Netrek in Go.

## Overview

Go Netrek is a reinterpretation of the classic Netrek game, implementing its core gameplay elements in a flexible, modular Go architecture. The game preserves the team-based strategic space combat that made the original Netrek popular while leveraging modern software design principles.

## Core Features

- Team-based space combat for up to 16 players
- Ship-to-ship combat with various weapon systems
- Planet conquest mechanics
- Real-time multiplayer over TCP/IP
- Configurable game rules and galaxy maps
- Extensible, event-driven architecture

## Architecture

The project is structured using a clean, modular architecture:

- **Engine**: Core game state and logic management
- **Entity**: Game object definitions (ships, planets, weapons, etc.)
- **Physics**: Movement, collision detection, and spatial partitioning
- **Network**: Client-server communication
- **Event**: Message passing system for game events
- **Config**: Game configuration management

## Getting Started

### Requirements

- Go 1.18 or higher

### Running a Server

```bash
go run cmd/server/main.go --config=config.json
```

Or to create a default configuration:

```bash
go run cmd/server/main.go --default --config=config.json
```

### Running a Client

```bash
go run cmd/client/main.go --server=localhost:4566 --name=Player1 --team=0
```

## Development

### Building from Source

```bash
# Build server
go build -o netrek-server cmd/server/main.go

# Build client
go build -o netrek-client cmd/client/main.go
```

### Testing

```bash
go test ./...
```

## Examples

The `examples/` directory contains example implementations:

- `simple_server`: A basic server setup with a small galaxy
- `ai_client`: AI-controlled client bots with different behaviors

## Configuration

Game configuration is stored in JSON format. See `pkg/config/config.go` for the structure.

Example configuration:

```json
{
  "worldSize": 10000,
  "maxPlayers": 16,
  "teams": [
    {
      "name": "Federation",
      "color": "#0000FF",
      "maxShips": 8,
      "startingShip": "Scout"
    },
    {
      "name": "Romulans",
      "color": "#00FF00",
      "maxShips": 8,
      "startingShip": "Scout"
    }
  ],
  "planets": [
    {
      "name": "Earth",
      "x": -4000,
      "y": 0,
      "type": 3,
      "homeWorld": true,
      "teamID": 0,
      "initialArmies": 30
    },
    {
      "name": "Romulus",
      "x": 4000,
      "y": 0,
      "type": 3,
      "homeWorld": true,
      "teamID": 1,
      "initialArmies": 30
    }
  ],
  "physics": {
    "gravity": 0,
    "friction": 0.1,
    "collisionDamage": 20
  },
  "network": {
    "updateRate": 20,
    "ticksPerState": 3,
    "usePartialState": true,
    "serverPort": 4566,
    "serverAddress": "localhost:4566"
  },
  "gameRules": {
    "winCondition": "conquest",
    "timeLimit": 1800,
    "maxScore": 100,
    "respawnDelay": 5,
    "friendlyFire": false,
    "startingArmies": 0
  }
}
```

## Extending the Game

The modular architecture makes it easy to extend the game:

- Add new ship classes in `pkg/entity/ship.go`
- Implement new weapons in `pkg/entity/weapon.go`
- Create custom game rules in `pkg/engine/game.go`
- Design new network protocols in `pkg/network/`

## License

[MIT License](LICENSE)