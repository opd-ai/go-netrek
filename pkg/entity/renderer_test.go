package entity

import (
	"fmt"
	"testing"

	"github.com/opd-ai/go-netrek/pkg/physics"
)

// AdvancedMockRenderer is an enhanced test implementation of the Renderer interface
// that tracks detailed call information
type AdvancedMockRenderer struct {
	// Track method calls with parameters for verification
	RenderShipCalls       []AdvancedRenderShipCall
	RenderPlanetCalls     []AdvancedRenderPlanetCall
	RenderProjectileCalls []AdvancedRenderProjectileCall
	ClearCallCount        int
	PresentCallCount      int
}

// AdvancedRenderShipCall captures the parameters of a RenderShip call
type AdvancedRenderShipCall struct {
	Ship *Ship
}

// AdvancedRenderPlanetCall captures the parameters of a RenderPlanet call
type AdvancedRenderPlanetCall struct {
	Planet *Planet
}

// AdvancedRenderProjectileCall captures the parameters of a RenderProjectile call
type AdvancedRenderProjectileCall struct {
	Projectile *Projectile
}

// NewAdvancedMockRenderer creates a new advanced mock renderer
func NewAdvancedMockRenderer() *AdvancedMockRenderer {
	return &AdvancedMockRenderer{
		RenderShipCalls:       make([]AdvancedRenderShipCall, 0),
		RenderPlanetCalls:     make([]AdvancedRenderPlanetCall, 0),
		RenderProjectileCalls: make([]AdvancedRenderProjectileCall, 0),
	}
}

// RenderShip implements the Renderer interface
func (m *AdvancedMockRenderer) RenderShip(ship *Ship) {
	m.RenderShipCalls = append(m.RenderShipCalls, AdvancedRenderShipCall{Ship: ship})
}

// RenderPlanet implements the Renderer interface
func (m *AdvancedMockRenderer) RenderPlanet(planet *Planet) {
	m.RenderPlanetCalls = append(m.RenderPlanetCalls, AdvancedRenderPlanetCall{Planet: planet})
}

// RenderProjectile implements the Renderer interface
func (m *AdvancedMockRenderer) RenderProjectile(projectile *Projectile) {
	m.RenderProjectileCalls = append(m.RenderProjectileCalls, AdvancedRenderProjectileCall{Projectile: projectile})
}

// Clear implements the Renderer interface
func (m *AdvancedMockRenderer) Clear() {
	m.ClearCallCount++
}

// Present implements the Renderer interface
func (m *AdvancedMockRenderer) Present() {
	m.PresentCallCount++
}

// Reset clears all recorded calls
func (m *AdvancedMockRenderer) Reset() {
	m.RenderShipCalls = make([]AdvancedRenderShipCall, 0)
	m.RenderPlanetCalls = make([]AdvancedRenderPlanetCall, 0)
	m.RenderProjectileCalls = make([]AdvancedRenderProjectileCall, 0)
	m.ClearCallCount = 0
	m.PresentCallCount = 0
}

// TestRenderer_InterfaceCompliance verifies that both mock renderers implement Renderer
func TestRenderer_InterfaceCompliance(t *testing.T) {
	t.Run("MockRenderer implements Renderer", func(t *testing.T) {
		var _ Renderer = (*MockRenderer)(nil)
	})

	t.Run("AdvancedMockRenderer implements Renderer", func(t *testing.T) {
		var _ Renderer = (*AdvancedMockRenderer)(nil)
	})
}

// TestRenderer_RenderShip tests the RenderShip method behavior
func TestRenderer_RenderShip(t *testing.T) {
	tests := []struct {
		name        string
		ship        *Ship
		expectPanic bool
	}{
		{
			name:        "Valid ship with scout class",
			ship:        NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0}),
			expectPanic: false,
		},
		{
			name:        "Nil ship should not panic",
			ship:        nil,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewAdvancedMockRenderer()

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but none occurred")
					}
				}()
			}

			renderer.RenderShip(tt.ship)

			if len(renderer.RenderShipCalls) != 1 {
				t.Errorf("Expected 1 RenderShip call, got %d", len(renderer.RenderShipCalls))
			}

			if renderer.RenderShipCalls[0].Ship != tt.ship {
				t.Errorf("Expected ship %v, got %v", tt.ship, renderer.RenderShipCalls[0].Ship)
			}
		})
	}
}

// TestRenderer_RenderPlanet tests the RenderPlanet method behavior
func TestRenderer_RenderPlanet(t *testing.T) {
	tests := []struct {
		name        string
		planet      *Planet
		expectPanic bool
	}{
		{
			name:        "Valid planet",
			planet:      NewPlanet(ID(1), "Test Planet", physics.Vector2D{X: 100, Y: 100}, Homeworld),
			expectPanic: false,
		},
		{
			name:        "Nil planet should not panic",
			planet:      nil,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewAdvancedMockRenderer()

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but none occurred")
					}
				}()
			}

			renderer.RenderPlanet(tt.planet)

			if len(renderer.RenderPlanetCalls) != 1 {
				t.Errorf("Expected 1 RenderPlanet call, got %d", len(renderer.RenderPlanetCalls))
			}

			if renderer.RenderPlanetCalls[0].Planet != tt.planet {
				t.Errorf("Expected planet %v, got %v", tt.planet, renderer.RenderPlanetCalls[0].Planet)
			}
		})
	}
}

// TestRenderer_RenderProjectile tests the RenderProjectile method behavior
func TestRenderer_RenderProjectile(t *testing.T) {
	tests := []struct {
		name        string
		projectile  *Projectile
		expectPanic bool
	}{
		{
			name: "Valid projectile",
			projectile: &Projectile{
				BaseEntity: BaseEntity{
					ID:       ID(1),
					Position: physics.Vector2D{X: 50, Y: 50},
				},
				Type:    "torpedo",
				OwnerID: ID(42),
				TeamID:  1,
				Damage:  100,
			},
			expectPanic: false,
		},
		{
			name:        "Nil projectile should not panic",
			projectile:  nil,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewAdvancedMockRenderer()

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but none occurred")
					}
				}()
			}

			renderer.RenderProjectile(tt.projectile)

			if len(renderer.RenderProjectileCalls) != 1 {
				t.Errorf("Expected 1 RenderProjectile call, got %d", len(renderer.RenderProjectileCalls))
			}

			if renderer.RenderProjectileCalls[0].Projectile != tt.projectile {
				t.Errorf("Expected projectile %v, got %v", tt.projectile, renderer.RenderProjectileCalls[0].Projectile)
			}
		})
	}
}

// TestRenderer_Clear tests the Clear method behavior
func TestRenderer_Clear(t *testing.T) {
	t.Run("Single clear call", func(t *testing.T) {
		renderer := NewAdvancedMockRenderer()

		renderer.Clear()

		if renderer.ClearCallCount != 1 {
			t.Errorf("Expected 1 Clear call, got %d", renderer.ClearCallCount)
		}
	})

	t.Run("Multiple clear calls", func(t *testing.T) {
		renderer := NewAdvancedMockRenderer()

		renderer.Clear()
		renderer.Clear()
		renderer.Clear()

		if renderer.ClearCallCount != 3 {
			t.Errorf("Expected 3 Clear calls, got %d", renderer.ClearCallCount)
		}
	})

	t.Run("Clear is idempotent", func(t *testing.T) {
		renderer := NewAdvancedMockRenderer()

		// Add some entities
		renderer.RenderShip(NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0}))
		renderer.RenderPlanet(NewPlanet(ID(2), "Test", physics.Vector2D{X: 0, Y: 0}, Homeworld))

		// Clear multiple times
		renderer.Clear()
		renderer.Clear()

		// Both clears should be recorded
		if renderer.ClearCallCount != 2 {
			t.Errorf("Expected 2 Clear calls, got %d", renderer.ClearCallCount)
		}

		// Previous render calls should still be recorded
		if len(renderer.RenderShipCalls) != 1 {
			t.Errorf("Expected 1 RenderShip call to remain, got %d", len(renderer.RenderShipCalls))
		}
		if len(renderer.RenderPlanetCalls) != 1 {
			t.Errorf("Expected 1 RenderPlanet call to remain, got %d", len(renderer.RenderPlanetCalls))
		}
	})
}

// TestRenderer_Present tests the Present method behavior
func TestRenderer_Present(t *testing.T) {
	t.Run("Single present call", func(t *testing.T) {
		renderer := NewAdvancedMockRenderer()

		renderer.Present()

		if renderer.PresentCallCount != 1 {
			t.Errorf("Expected 1 Present call, got %d", renderer.PresentCallCount)
		}
	})

	t.Run("Multiple present calls", func(t *testing.T) {
		renderer := NewAdvancedMockRenderer()

		renderer.Present()
		renderer.Present()

		if renderer.PresentCallCount != 2 {
			t.Errorf("Expected 2 Present calls, got %d", renderer.PresentCallCount)
		}
	})
}

// TestRenderer_CompleteRenderingSequence tests a typical rendering frame sequence
func TestRenderer_CompleteRenderingSequence(t *testing.T) {
	renderer := NewAdvancedMockRenderer()

	// Create test entities
	ship := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})
	planet := NewPlanet(ID(2), "Test Planet", physics.Vector2D{X: 100, Y: 100}, Homeworld)
	projectile := &Projectile{
		BaseEntity: BaseEntity{
			ID:       ID(3),
			Position: physics.Vector2D{X: 50, Y: 50},
		},
		Type:    "torpedo",
		OwnerID: ID(1),
		TeamID:  0,
		Damage:  100,
	}

	// Simulate a complete rendering frame
	renderer.Clear()
	renderer.RenderShip(ship)
	renderer.RenderPlanet(planet)
	renderer.RenderProjectile(projectile)
	renderer.Present()

	// Verify the sequence
	if renderer.ClearCallCount != 1 {
		t.Errorf("Expected 1 Clear call, got %d", renderer.ClearCallCount)
	}
	if len(renderer.RenderShipCalls) != 1 {
		t.Errorf("Expected 1 RenderShip call, got %d", len(renderer.RenderShipCalls))
	}
	if len(renderer.RenderPlanetCalls) != 1 {
		t.Errorf("Expected 1 RenderPlanet call, got %d", len(renderer.RenderPlanetCalls))
	}
	if len(renderer.RenderProjectileCalls) != 1 {
		t.Errorf("Expected 1 RenderProjectile call, got %d", len(renderer.RenderProjectileCalls))
	}
	if renderer.PresentCallCount != 1 {
		t.Errorf("Expected 1 Present call, got %d", renderer.PresentCallCount)
	}

	// Verify entity parameters
	if renderer.RenderShipCalls[0].Ship != ship {
		t.Errorf("Expected ship %v, got %v", ship, renderer.RenderShipCalls[0].Ship)
	}
	if renderer.RenderPlanetCalls[0].Planet != planet {
		t.Errorf("Expected planet %v, got %v", planet, renderer.RenderPlanetCalls[0].Planet)
	}
	if renderer.RenderProjectileCalls[0].Projectile != projectile {
		t.Errorf("Expected projectile %v, got %v", projectile, renderer.RenderProjectileCalls[0].Projectile)
	}
}

// TestRenderer_MultipleEntitiesOfSameType tests rendering multiple entities of the same type
func TestRenderer_MultipleEntitiesOfSameType(t *testing.T) {
	renderer := NewAdvancedMockRenderer()

	// Create multiple ships
	ships := make([]*Ship, 5)
	for i := range ships {
		ships[i] = NewShip(ID(i+1), Scout, 0, physics.Vector2D{X: float64(i * 10), Y: 0})
	}

	// Render all ships
	for _, ship := range ships {
		renderer.RenderShip(ship)
	}

	// Verify all ships were rendered
	if len(renderer.RenderShipCalls) != 5 {
		t.Errorf("Expected 5 RenderShip calls, got %d", len(renderer.RenderShipCalls))
	}

	// Verify ship order is maintained
	for i, call := range renderer.RenderShipCalls {
		if call.Ship != ships[i] {
			t.Errorf("Expected ship %d to be %v, got %v", i, ships[i], call.Ship)
		}
	}
}

// TestRenderer_RenderingOrder tests that rendering calls are recorded in order
func TestRenderer_RenderingOrder(t *testing.T) {
	renderer := NewAdvancedMockRenderer()

	// Create entities
	ship1 := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})
	planet1 := NewPlanet(ID(2), "Planet 1", physics.Vector2D{X: 100, Y: 100}, Homeworld)
	ship2 := NewShip(ID(3), Destroyer, 1, physics.Vector2D{X: 200, Y: 200})
	projectile1 := &Projectile{
		BaseEntity: BaseEntity{
			ID:       ID(4),
			Position: physics.Vector2D{X: 50, Y: 50},
		},
		Type:    "torpedo",
		OwnerID: ID(1),
		TeamID:  0,
		Damage:  100,
	}

	// Render in specific order
	renderer.RenderShip(ship1)
	renderer.RenderPlanet(planet1)
	renderer.RenderShip(ship2)
	renderer.RenderProjectile(projectile1)

	// Verify order within each type
	if len(renderer.RenderShipCalls) != 2 {
		t.Errorf("Expected 2 RenderShip calls, got %d", len(renderer.RenderShipCalls))
	}
	if renderer.RenderShipCalls[0].Ship != ship1 {
		t.Errorf("Expected first ship to be %v, got %v", ship1, renderer.RenderShipCalls[0].Ship)
	}
	if renderer.RenderShipCalls[1].Ship != ship2 {
		t.Errorf("Expected second ship to be %v, got %v", ship2, renderer.RenderShipCalls[1].Ship)
	}

	if len(renderer.RenderPlanetCalls) != 1 {
		t.Errorf("Expected 1 RenderPlanet call, got %d", len(renderer.RenderPlanetCalls))
	}
	if renderer.RenderPlanetCalls[0].Planet != planet1 {
		t.Errorf("Expected planet to be %v, got %v", planet1, renderer.RenderPlanetCalls[0].Planet)
	}

	if len(renderer.RenderProjectileCalls) != 1 {
		t.Errorf("Expected 1 RenderProjectile call, got %d", len(renderer.RenderProjectileCalls))
	}
	if renderer.RenderProjectileCalls[0].Projectile != projectile1 {
		t.Errorf("Expected projectile to be %v, got %v", projectile1, renderer.RenderProjectileCalls[0].Projectile)
	}
}

// TestAdvancedMockRenderer_Reset tests the Reset functionality
func TestAdvancedMockRenderer_Reset(t *testing.T) {
	renderer := NewAdvancedMockRenderer()

	// Add some calls
	renderer.RenderShip(NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0}))
	renderer.RenderPlanet(NewPlanet(ID(2), "Test", physics.Vector2D{X: 0, Y: 0}, Homeworld))
	renderer.RenderProjectile(&Projectile{
		BaseEntity: BaseEntity{ID: ID(3)},
		Type:       "torpedo",
	})
	renderer.Clear()
	renderer.Present()

	// Verify calls were recorded
	if len(renderer.RenderShipCalls) != 1 {
		t.Errorf("Expected 1 RenderShip call before reset, got %d", len(renderer.RenderShipCalls))
	}
	if len(renderer.RenderPlanetCalls) != 1 {
		t.Errorf("Expected 1 RenderPlanet call before reset, got %d", len(renderer.RenderPlanetCalls))
	}
	if len(renderer.RenderProjectileCalls) != 1 {
		t.Errorf("Expected 1 RenderProjectile call before reset, got %d", len(renderer.RenderProjectileCalls))
	}
	if renderer.ClearCallCount != 1 {
		t.Errorf("Expected 1 Clear call before reset, got %d", renderer.ClearCallCount)
	}
	if renderer.PresentCallCount != 1 {
		t.Errorf("Expected 1 Present call before reset, got %d", renderer.PresentCallCount)
	}

	// Reset
	renderer.Reset()

	// Verify all calls were cleared
	if len(renderer.RenderShipCalls) != 0 {
		t.Errorf("Expected 0 RenderShip calls after reset, got %d", len(renderer.RenderShipCalls))
	}
	if len(renderer.RenderPlanetCalls) != 0 {
		t.Errorf("Expected 0 RenderPlanet calls after reset, got %d", len(renderer.RenderPlanetCalls))
	}
	if len(renderer.RenderProjectileCalls) != 0 {
		t.Errorf("Expected 0 RenderProjectile calls after reset, got %d", len(renderer.RenderProjectileCalls))
	}
	if renderer.ClearCallCount != 0 {
		t.Errorf("Expected 0 Clear calls after reset, got %d", renderer.ClearCallCount)
	}
	if renderer.PresentCallCount != 0 {
		t.Errorf("Expected 0 Present calls after reset, got %d", renderer.PresentCallCount)
	}
}

// TestRenderer_PolymorphicBehavior tests that the interface can be used polymorphically
func TestRenderer_PolymorphicBehavior(t *testing.T) {
	renderers := []Renderer{
		&MockRenderer{},
		NewAdvancedMockRenderer(),
	}

	for i, renderer := range renderers {
		t.Run(fmt.Sprintf("Renderer_%d", i), func(t *testing.T) {
			// Create test entities
			ship := NewShip(ID(1), Scout, 0, physics.Vector2D{X: 0, Y: 0})
			planet := NewPlanet(ID(2), "Test Planet", physics.Vector2D{X: 100, Y: 100}, Homeworld)
			projectile := &Projectile{
				BaseEntity: BaseEntity{
					ID:       ID(3),
					Position: physics.Vector2D{X: 50, Y: 50},
				},
				Type:    "torpedo",
				OwnerID: ID(1),
				TeamID:  0,
				Damage:  100,
			}

			// Test that all methods can be called without error
			renderer.Clear()
			renderer.RenderShip(ship)
			renderer.RenderPlanet(planet)
			renderer.RenderProjectile(projectile)
			renderer.Present()

			// All methods should complete without panicking
		})
	}
}
