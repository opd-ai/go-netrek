# Go-Netrek Examples

This directory contains example implementations that demonstrate how to use the go-netrek framework.

## Examples

### 1. Simple Server (`simple_server/`)

A basic Netrek server setup that creates a small galaxy with two teams and four planets.

**Usage:**
```bash
# Start the server (default port 4566)
go run examples/simple_server/main.go

# Start with custom settings
go run examples/simple_server/main.go -port=4567 -max-clients=16
```

**Configuration:**
- 2 teams: Federation (blue) vs Klingons (orange)
- 4 planets: 2 home worlds + 2 neutral planets
- World size: 8000 units
- Game time limit: 15 minutes

### 2. AI Client (`ai_client/`)

AI-controlled client bots with different behavioral patterns.

**Usage:**
```bash
# Explorer bot (default behavior)
go run examples/ai_client/main.go -server=localhost:4566 -name=Explorer -team=0

# Aggressive bot
go run examples/ai_client/main.go -server=localhost:4566 -name=Warrior -team=1 -behavior=aggressor

# Defender bot  
go run examples/ai_client/main.go -server=localhost:4566 -name=Guardian -team=0 -behavior=defender

# Bomber bot
go run examples/ai_client/main.go -server=localhost:4566 -name=Bomber -team=1 -behavior=bomber
```

**AI Behaviors:**
- **Explorer**: Flies around exploring, sends friendly messages
- **Aggressor**: Seeks out and attacks enemy ships
- **Defender**: Stays near home planets and defends them
- **Bomber**: Attacks enemy planets with armies

## Quick Start Demo

1. **Start the server:**
```bash
go run examples/simple_server/main.go
```

2. **Connect AI clients (in separate terminals):**
```bash
# Team 0 bots
go run examples/ai_client/main.go -name=Federation_Explorer -team=0 -behavior=explorer
go run examples/ai_client/main.go -name=Federation_Defender -team=0 -behavior=defender

# Team 1 bots  
go run examples/ai_client/main.go -name=Klingon_Warrior -team=1 -behavior=aggressor
go run examples/ai_client/main.go -name=Klingon_Bomber -team=1 -behavior=bomber
```

3. **Watch the action in the server logs!**

The AI bots will automatically:
- Spawn ships and start moving
- Engage in combat when they encounter enemies
- Send chat messages based on their personality
- Attempt to capture neutral planets
- Defend their team's territory

## Extending the Examples

### Adding New AI Behaviors

To add a new AI behavior:

1. Add a new behavior constant in `ai_client/main.go`:
```go
const (
    BehaviorExplorer AIBehavior = iota
    BehaviorAggressor
    BehaviorDefender  
    BehaviorBomber
    BehaviorYourNewBehavior  // Add here
)
```

2. Implement the behavior function:
```go
func (ai *AIClient) executeYourNewBehavior(myShip engine.ShipState) {
    // Your AI logic here
    thrust := true
    turnLeft := false
    turnRight := false
    fireWeapon := -1
    
    // ... behavior implementation ...
    
    ai.client.SendInput(thrust, turnLeft, turnRight, fireWeapon, false, false, 0, 0)
}
```

3. Add it to the behavior switch statement in `executeBehavior()`.

### Customizing the Server

The simple server demonstrates basic configuration. You can customize:

- **Teams**: Add more teams, change colors and names
- **Galaxy**: Add more planets, change positions and types  
- **Physics**: Adjust gravity, friction, collision damage
- **Game Rules**: Modify win conditions, time limits, respawn delays
- **Network**: Change update rates, enable/disable partial states

See `pkg/config/config.go` for all available configuration options.

## Architecture Learning

These examples demonstrate key go-netrek concepts:

- **Event-driven design**: AI clients subscribe to game events
- **Network protocol**: Client-server communication via TCP
- **Game state management**: Centralized state on server, synchronized to clients
- **Entity system**: Ships, planets, and projectiles as game entities
- **Physics integration**: Movement, collision detection, spatial queries
- **Configuration system**: JSON-based game configuration

Study the code to understand how these systems work together to create a multiplayer space combat game!
