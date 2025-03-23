# Entity Package

The entity package provides core game object implementations for Go Netrek, including ships, planets, weapons, and projectiles.

## Components

### Base Entity ([entity.go](entity.go))
Provides the foundation for all game objects via the `Entity` interface and `BaseEntity` struct:

```go
type Entity interface {
    GetID() ID
    GetPosition() physics.Vector2D
    GetCollider() physics.Circle
    Update(deltaTime float64)
    Render()
}
```

### Ships ([ship.go](ship.go))
Implements player-controlled vessels with the `Ship` struct:

```go
type Ship struct {
    BaseEntity
    Class     ShipClass
    Stats     ShipStats
    TeamID    int
    Hull      int
    Shields   int
    Weapons   []Weapon
    // ... additional fields
}
```

Key features:
- Multiple ship classes (Scout, Destroyer, etc.)
- Weapon systems
- Shield and hull damage management
- Movement and physics integration
- Army transport capabilities

### Planets ([planet.go](planet.go))
Implements planetary bodies with the `Planet` struct:

```go
type Planet struct {
    BaseEntity
    Name       string
    Type       PlanetType
    TeamID     int
    Armies     int
    Resources  int
    Production int
    // ... additional fields
}
```

Features:
- Different planet types (Agricultural, Industrial, etc.)
- Army production and management
- Team ownership and conquest mechanics
- Resource management

### Weapons ([weapon.go](weapon.go))
Implements weapon systems with interfaces and concrete types:

```go
type Weapon interface {
    GetName() string
    GetCooldown() time.Duration
    GetFuelCost() int
    CreateProjectile(...) *Projectile
}
```

Available weapons:
- Torpedoes (longer range, higher damage)
- Phasers (faster firing, lower damage)

## Usage Examples

Creating a new ship:
```go
position := physics.Vector2D{X: 0, Y: 0}
ship := NewShip(GenerateID(), Scout, teamID, position)
```

Firing weapons:
```go
if projectile := ship.FireWeapon(0); projectile != nil {
    // Handle projectile creation
}
```

Planet management:
```go
planet := NewPlanet(GenerateID(), "Earth", position, Homeworld)
damage := planet.Bomb(20)
armies, captured := planet.BeamDownArmies(teamID, 5)
```

## Entity Relationships

```
BaseEntity
    |
    |---> Ship
    |     |---> Weapons
    |           |---> Projectiles
    |
    |---> Planet
    |
    |---> Projectile
```

## Common Operations

### Movement and Physics
All entities inherit basic physics functionality from `BaseEntity`:

```go
func (e *BaseEntity) Update(deltaTime float64) {
    e.Position = e.Position.Add(e.Velocity.Scale(deltaTime))
    e.Collider.Center = e.Position
}
```

### Collision Detection
Entities use circular colliders for collision detection:

```go
func (e *BaseEntity) GetCollider() physics.Circle {
    return physics.Circle{
        Center: e.Position,
        Radius: e.Collider.Radius,
    }
}
```

### ID Generation
Unique IDs are generated for all entities:

```go
id := GenerateID()
```

## Best Practices

1. Always check entity Active status before processing
2. Use appropriate collision detection for different entity types
3. Manage weapon cooldowns and resource costs
4. Handle team ownership appropriately for planets
5. Clean up inactive projectiles regularly

## Future Improvements

- Additional ship classes
- More weapon varieties
- Enhanced planet types and resources
- Improved damage models
- Advanced movement mechanics