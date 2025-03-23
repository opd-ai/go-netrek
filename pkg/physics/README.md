# Physics Package

The physics package provides core physics functionality for the Go Netrek game, including vector math, movement mechanics, and collision detection.

## Components

### Vector Math ([vector.go](vector.go))
Implements 2D vector operations via the `Vector2D` struct:

```go
type Vector2D struct {
    X float64
    Y float64
}
```

Key operations:
- Vector addition/subtraction
- Scalar multiplication 
- Normalization
- Distance calculation
- Rotation
- Dot product

### Ship Movement ([ship.go](ship.go))
Handles ship movement physics via the `MovementState` struct:

```go
type MovementState struct {
    Position Vector2D
    Velocity Vector2D
    Heading  float64
    Mass     float64
    Thrust   float64 
    MaxSpeed float64
}
```

The `UpdateMovement()` function applies thrust and turning forces while respecting speed limits.

### Collision Detection ([collision.go](collision.go))
Provides collision detection between entities:

```go
type Circle struct {
    Center Vector2D
    Radius float64
}
```

Features:
- Circle-circle collision detection
- Collision response calculations 
- Spatial partitioning via QuadTree for efficient collision queries

## Usage Examples

Basic vector operations:
```go
v1 := Vector2D{X: 1, Y: 2}
v2 := Vector2D{X: 2, Y: 3}
sum := v1.Add(v2)
dist := v1.Distance(v2)
```

Ship movement:
```go
state := &MovementState{
    Position: Vector2D{X: 0, Y: 0},
    Heading: 0,
    MaxSpeed: 100,
}
UpdateMovement(state, deltaTime, thrust, turn)
```

Collision detection:
```go
c1 := Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5}
c2 := Circle{Center: Vector2D{X: 3, Y: 4}, Radius: 5}
if c1.Collides(c2) {
    // Handle collision
}
```

Using the QuadTree for spatial partitioning:
```go
qt := NewQuadTree(Rect{
    Center: Vector2D{X: 0, Y: 0},
    Width: 1000,
    Height: 1000,
}, 10)

// Insert objects
qt.Insert(obj.Position, obj)

// Query for potential collisions
nearby := qt.Query(searchArea)
```