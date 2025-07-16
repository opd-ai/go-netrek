// pkg/entity/entity_test.go
package entity

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/physics"
)

func TestBaseEntity_GetID(t *testing.T) {
	tests := []struct {
		name     string
		entityID ID
		expected ID
	}{
		{
			name:     "zero_id",
			entityID: 0,
			expected: 0,
		},
		{
			name:     "positive_id",
			entityID: 42,
			expected: 42,
		},
		{
			name:     "large_id",
			entityID: 18446744073709551615, // max uint64
			expected: 18446744073709551615,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := &BaseEntity{
				ID: tt.entityID,
			}

			result := entity.GetID()
			if result != tt.expected {
				t.Errorf("GetID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBaseEntity_GetPosition(t *testing.T) {
	tests := []struct {
		name     string
		position physics.Vector2D
		expected physics.Vector2D
	}{
		{
			name:     "zero_position",
			position: physics.Vector2D{X: 0, Y: 0},
			expected: physics.Vector2D{X: 0, Y: 0},
		},
		{
			name:     "positive_coordinates",
			position: physics.Vector2D{X: 100, Y: 200},
			expected: physics.Vector2D{X: 100, Y: 200},
		},
		{
			name:     "negative_coordinates",
			position: physics.Vector2D{X: -50, Y: -75},
			expected: physics.Vector2D{X: -50, Y: -75},
		},
		{
			name:     "mixed_coordinates",
			position: physics.Vector2D{X: 123.45, Y: -67.89},
			expected: physics.Vector2D{X: 123.45, Y: -67.89},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := &BaseEntity{
				Position: tt.position,
			}

			result := entity.GetPosition()
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("GetPosition() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBaseEntity_GetCollider(t *testing.T) {
	tests := []struct {
		name           string
		position       physics.Vector2D
		colliderRadius float64
		expectedCenter physics.Vector2D
		expectedRadius float64
	}{
		{
			name:           "basic_collider",
			position:       physics.Vector2D{X: 10, Y: 20},
			colliderRadius: 5.0,
			expectedCenter: physics.Vector2D{X: 10, Y: 20},
			expectedRadius: 5.0,
		},
		{
			name:           "zero_radius_collider",
			position:       physics.Vector2D{X: 0, Y: 0},
			colliderRadius: 0.0,
			expectedCenter: physics.Vector2D{X: 0, Y: 0},
			expectedRadius: 0.0,
		},
		{
			name:           "large_radius_collider",
			position:       physics.Vector2D{X: -100, Y: 200},
			colliderRadius: 999.99,
			expectedCenter: physics.Vector2D{X: -100, Y: 200},
			expectedRadius: 999.99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := &BaseEntity{
				Position: tt.position,
				Collider: physics.Circle{
					Center: physics.Vector2D{}, // This should be overridden
					Radius: tt.colliderRadius,
				},
			}

			result := entity.GetCollider()

			// Check that the center matches the entity position (not the collider's stored center)
			if result.Center.X != tt.expectedCenter.X || result.Center.Y != tt.expectedCenter.Y {
				t.Errorf("GetCollider().Center = %v, want %v", result.Center, tt.expectedCenter)
			}

			// Check that the radius matches the collider radius
			if result.Radius != tt.expectedRadius {
				t.Errorf("GetCollider().Radius = %v, want %v", result.Radius, tt.expectedRadius)
			}
		})
	}
}

func TestBaseEntity_Update(t *testing.T) {
	tests := []struct {
		name                   string
		initialPosition        physics.Vector2D
		velocity               physics.Vector2D
		deltaTime              float64
		expectedPosition       physics.Vector2D
		expectedColliderCenter physics.Vector2D
	}{
		{
			name:                   "no_movement_zero_velocity",
			initialPosition:        physics.Vector2D{X: 10, Y: 20},
			velocity:               physics.Vector2D{X: 0, Y: 0},
			deltaTime:              1.0,
			expectedPosition:       physics.Vector2D{X: 10, Y: 20},
			expectedColliderCenter: physics.Vector2D{X: 10, Y: 20},
		},
		{
			name:                   "positive_velocity_movement",
			initialPosition:        physics.Vector2D{X: 0, Y: 0},
			velocity:               physics.Vector2D{X: 50, Y: 100},
			deltaTime:              1.0,
			expectedPosition:       physics.Vector2D{X: 50, Y: 100},
			expectedColliderCenter: physics.Vector2D{X: 50, Y: 100},
		},
		{
			name:                   "negative_velocity_movement",
			initialPosition:        physics.Vector2D{X: 100, Y: 200},
			velocity:               physics.Vector2D{X: -25, Y: -50},
			deltaTime:              2.0,
			expectedPosition:       physics.Vector2D{X: 50, Y: 100},
			expectedColliderCenter: physics.Vector2D{X: 50, Y: 100},
		},
		{
			name:                   "fractional_delta_time",
			initialPosition:        physics.Vector2D{X: 0, Y: 0},
			velocity:               physics.Vector2D{X: 100, Y: 200},
			deltaTime:              0.5,
			expectedPosition:       physics.Vector2D{X: 50, Y: 100},
			expectedColliderCenter: physics.Vector2D{X: 50, Y: 100},
		},
		{
			name:                   "zero_delta_time",
			initialPosition:        physics.Vector2D{X: 42, Y: 84},
			velocity:               physics.Vector2D{X: 1000, Y: 2000},
			deltaTime:              0.0,
			expectedPosition:       physics.Vector2D{X: 42, Y: 84},
			expectedColliderCenter: physics.Vector2D{X: 42, Y: 84},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := &BaseEntity{
				Position: tt.initialPosition,
				Velocity: tt.velocity,
				Collider: physics.Circle{
					Center: tt.initialPosition, // Start with initial position
					Radius: 10.0,
				},
			}

			entity.Update(tt.deltaTime)

			// Check position update
			if entity.Position.X != tt.expectedPosition.X || entity.Position.Y != tt.expectedPosition.Y {
				t.Errorf("Update() position = %v, want %v", entity.Position, tt.expectedPosition)
			}

			// Check that collider center is updated to match position
			if entity.Collider.Center.X != tt.expectedColliderCenter.X || entity.Collider.Center.Y != tt.expectedColliderCenter.Y {
				t.Errorf("Update() collider center = %v, want %v", entity.Collider.Center, tt.expectedColliderCenter)
			}
		})
	}
}

func TestBaseEntity_Render(t *testing.T) {
	// Test that the base Render method can be called without panic
	// This is a basic test since the base implementation does nothing
	entity := &BaseEntity{
		ID:       1,
		Position: physics.Vector2D{X: 10, Y: 20},
	}

	// Create a mock renderer for testing
	mockRenderer := &MockRenderer{}

	// This should not panic and should do nothing
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("BaseEntity.Render() panicked: %v", r)
		}
	}()

	entity.Render(mockRenderer)
}

func TestBaseEntity_IntegrationTest(t *testing.T) {
	// Integration test that combines multiple operations
	entity := &BaseEntity{
		ID:       123,
		Position: physics.Vector2D{X: 0, Y: 0},
		Velocity: physics.Vector2D{X: 10, Y: 20},
		Rotation: 0.5,
		Collider: physics.Circle{
			Center: physics.Vector2D{X: 0, Y: 0},
			Radius: 15.0,
		},
		Active: true,
	}

	// Verify initial state
	if entity.GetID() != 123 {
		t.Errorf("Initial ID = %v, want %v", entity.GetID(), 123)
	}

	initialPos := entity.GetPosition()
	if initialPos.X != 0 || initialPos.Y != 0 {
		t.Errorf("Initial position = %v, want {0, 0}", initialPos)
	}

	// Update entity
	entity.Update(2.0)

	// Verify updated state
	newPos := entity.GetPosition()
	expectedPos := physics.Vector2D{X: 20, Y: 40}
	if newPos.X != expectedPos.X || newPos.Y != expectedPos.Y {
		t.Errorf("Updated position = %v, want %v", newPos, expectedPos)
	}

	// Verify collider follows position
	collider := entity.GetCollider()
	if collider.Center.X != expectedPos.X || collider.Center.Y != expectedPos.Y {
		t.Errorf("Collider center = %v, want %v", collider.Center, expectedPos)
	}

	if collider.Radius != 15.0 {
		t.Errorf("Collider radius = %v, want %v", collider.Radius, 15.0)
	}
}

// MockRenderer implements the Renderer interface for testing
type MockRenderer struct {
	RenderShipCalled       bool
	RenderPlanetCalled     bool
	RenderProjectileCalled bool
	ClearCalled            bool
	PresentCalled          bool
}

func (m *MockRenderer) RenderShip(ship *Ship) {
	m.RenderShipCalled = true
}

func (m *MockRenderer) RenderPlanet(planet *Planet) {
	m.RenderPlanetCalled = true
}

func (m *MockRenderer) RenderProjectile(projectile *Projectile) {
	m.RenderProjectileCalled = true
}

func (m *MockRenderer) Clear() {
	m.ClearCalled = true
}

func (m *MockRenderer) Present() {
	m.PresentCalled = true
}
