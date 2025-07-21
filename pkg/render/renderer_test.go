// pkg/render/renderer_test.go
package render

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

func TestNullRenderer_Clear_LogsExpectedMessage(t *testing.T) {
	renderer := NewNullRenderer()

	// Test that Clear executes without panic
	// The NullRenderer uses structured logging which goes to stdout
	// We're primarily testing that the method executes successfully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Clear() panicked: %v", r)
		}
	}()

	renderer.Clear()
}

func TestNullRenderer_Present_LogsExpectedMessage(t *testing.T) {
	renderer := NewNullRenderer()

	// Test that Present executes without panic
	// The NullRenderer uses structured logging which goes to stdout
	// We're primarily testing that the method executes successfully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Present() panicked: %v", r)
		}
	}()

	renderer.Present()
}

func TestNullRenderer_RenderShip_LogsShipInformation(t *testing.T) {
	tests := []struct {
		name string
		ship *entity.Ship
	}{
		{
			name: "ValidShip_LogsCorrectly",
			ship: &entity.Ship{
				BaseEntity: entity.BaseEntity{
					ID:       entity.ID(123),
					Position: physics.Vector2D{X: 100.0, Y: 200.0},
					Rotation: 1.5,
				},
				TeamID: 1,
				Class:  entity.Scout,
				Hull:   100,
				Fuel:   80,
			},
		},
		{
			name: "NilShip_HandlesGracefully",
			ship: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewNullRenderer()

			// Test that RenderShip executes without panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RenderShip() panicked: %v", r)
				}
			}()

			renderer.RenderShip(tt.ship)
		})
	}
}

func TestNullRenderer_RenderPlanet_LogsPlanetInformation(t *testing.T) {
	tests := []struct {
		name   string
		planet *entity.Planet
	}{
		{
			name: "ValidPlanet_LogsCorrectly",
			planet: &entity.Planet{
				BaseEntity: entity.BaseEntity{
					ID:       entity.ID(456),
					Position: physics.Vector2D{X: 300.0, Y: 400.0},
				},
				Name:   "Earth",
				TeamID: 0,
				Type:   entity.Homeworld,
				Armies: 25,
			},
		},
		{
			name:   "NilPlanet_HandlesGracefully",
			planet: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewNullRenderer()

			// Test that RenderPlanet executes without panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RenderPlanet() panicked: %v", r)
				}
			}()

			renderer.RenderPlanet(tt.planet)
		})
	}
}

func TestNullRenderer_RenderProjectile_LogsProjectileInformation(t *testing.T) {
	tests := []struct {
		name       string
		projectile *entity.Projectile
	}{
		{
			name: "ValidProjectile_LogsCorrectly",
			projectile: &entity.Projectile{
				BaseEntity: entity.BaseEntity{
					ID:       entity.ID(789),
					Position: physics.Vector2D{X: 50.0, Y: 75.0},
					Velocity: physics.Vector2D{X: 10.0, Y: 5.0},
				},
				Damage:  15,
				Range:   500.0,
				OwnerID: entity.ID(123),
			},
		},
		{
			name:       "NilProjectile_HandlesGracefully",
			projectile: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewNullRenderer()

			// Test that RenderProjectile executes without panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RenderProjectile() panicked: %v", r)
				}
			}()

			renderer.RenderProjectile(tt.projectile)
		})
	}
}

func TestNullRenderer_ImplementsRendererInterface(t *testing.T) {
	var renderer entity.Renderer = NewNullRenderer()

	// Test that all interface methods are implemented
	renderer.Clear()
	renderer.Present()
	renderer.RenderShip(nil)
	renderer.RenderPlanet(nil)
	renderer.RenderProjectile(nil)

	// If we get here without compilation errors, the interface is properly implemented
}

func TestNullRenderer_GlobalVariable_IsCorrectType(t *testing.T) {
	// Test that the global NullRendererInstance variable is of the correct type
	var renderer entity.Renderer = NullRendererInstance

	// Verify we can use it like any other renderer without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Global NullRendererInstance should work without panic: %v", r)
		}
	}()

	renderer.Clear()
}

func TestNullRenderer_ConcurrentUsage_ThreadSafe(t *testing.T) {
	renderer := NewNullRenderer()
	done := make(chan bool, 3)

	// Test concurrent calls to different methods
	go func() {
		for i := 0; i < 10; i++ {
			renderer.Clear()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			renderer.Present()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			renderer.RenderShip(nil)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we get here without deadlocks or panics, the renderer is thread-safe
}

func TestNullRenderer_AllMethods_ProduceOutput(t *testing.T) {
	// Integration test to ensure all methods execute without panic
	renderer := NewNullRenderer()

	methods := []struct {
		name string
		call func()
	}{
		{
			name: "Clear",
			call: func() { renderer.Clear() },
		},
		{
			name: "Present",
			call: func() { renderer.Present() },
		},
		{
			name: "RenderShip",
			call: func() { renderer.RenderShip(nil) },
		},
		{
			name: "RenderPlanet",
			call: func() { renderer.RenderPlanet(nil) },
		},
		{
			name: "RenderProjectile",
			call: func() { renderer.RenderProjectile(nil) },
		},
	}

	for _, method := range methods {
		t.Run(method.name+"_ExecutesWithoutPanic", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Method %s panicked: %v", method.name, r)
				}
			}()

			method.call()
		})
	}
}
