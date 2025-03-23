# Config Package

The config package handles configuration management for Go Netrek, providing flexible game settings through JSON configuration files.

## Usage

```go
import "github.com/opd-ai/go-netrek/pkg/config"

// Load configuration from file
config, err := config.LoadConfig("config.json")

// Create default configuration
defaultConfig := config.DefaultConfig()

// Save configuration to file 
err := config.SaveConfig(config, "config.json")
```

## Configuration File Format

The configuration uses JSON format. Here's a minimal example:

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

## Configuration Options

### Core Settings
- `worldSize`: Size of the game world (float64)
- `maxPlayers`: Maximum number of concurrent players

### Teams Configuration
Each team has the following settings:
- `name`: Team name (string)
- `color`: Team color in hex format (e.g. "#0000FF")
- `maxShips`: Maximum ships allowed per team
- `startingShip`: Default ship class for new players

### Planets Configuration 
Each planet has:
- `name`: Planet name
- `x`, `y`: Planet coordinates
- `type`: Planet type (0=Agricultural, 1=Industrial, 2=Military, 3=Homeworld)
- `homeWorld`: Boolean indicating if this is a team's homeworld
- `teamID`: ID of owning team (-1 for neutral)
- `initialArmies`: Starting number of armies

### Physics Settings
- `gravity`: Global gravity strength 
- `friction`: Movement friction coefficient
- `collisionDamage`: Base damage from collisions

### Network Settings
- `updateRate`: Server update frequency in Hz
- `ticksPerState`: Number of ticks between full state updates
- `usePartialState`: Enable partial state updates
- `serverPort`: Port for game server
- `serverAddress`: Server address

### Game Rules
- `winCondition`: Victory condition ("conquest" or "score")
- `timeLimit`: Match time limit in seconds
- `maxScore`: Score needed for victory
- `respawnDelay`: Delay before ship respawn
- `friendlyFire`: Allow friendly fire damage
- `startingArmies`: Initial armies per player

## Environment Variables

The following environment variables can override config file settings:

- `NETREK_SERVER`: Override server address
- `NETREK_PORT`: Override server port
- `NETREK_MAX_PLAYERS`: Override maximum players

## Default Configuration

The package provides a default configuration via `DefaultConfig()` which can be used as a starting point. See the source code for the default values.

## Error Handling

Configuration loading errors are returned with context:

```go
config, err := config.LoadConfig("config.json")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}
```