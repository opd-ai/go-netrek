// pkg/render/renderer_test.go
package render

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// captureLog captures log output for testing
func captureLog(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	f()
	return buf.String()
}

func TestNullRenderer_Clear_LogsExpectedMessage(t *testing.T) {
	renderer := &NullRenderer{}

	output := captureLog(func() {
		renderer.Clear()
	})

	if !strings.Contains(output, "Clear called") {
		t.Errorf("Expected log to contain 'Clear called', got: %s", output)
	}
}

func TestNullRenderer_Present_LogsExpectedMessage(t *testing.T) {
	renderer := &NullRenderer{}

	output := captureLog(func() {
		renderer.Present()
	})

	if !strings.Contains(output, "Present called") {
		t.Errorf("Expected log to contain 'Present called', got: %s", output)
	}
}

func TestNullRenderer_RenderShip_LogsShipInformation(t *testing.T) {
	tests := []struct {
		name     string
		ship     *entity.Ship
		expected string
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
			expected: "RenderShip called for ship:",
		},
		{
			name:     "NilShip_HandlesGracefully",
			ship:     nil,
			expected: "RenderShip called for ship:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := &NullRenderer{}

			output := captureLog(func() {
				renderer.RenderShip(tt.ship)
			})

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected log to contain '%s', got: %s", tt.expected, output)
			}
		})
	}
}

func TestNullRenderer_RenderPlanet_LogsPlanetInformation(t *testing.T) {
	tests := []struct {
		name     string
		planet   *entity.Planet
		expected string
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
			expected: "RenderPlanet called for planet:",
		},
		{
			name:     "NilPlanet_HandlesGracefully",
			planet:   nil,
			expected: "RenderPlanet called for planet:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := &NullRenderer{}

			output := captureLog(func() {
				renderer.RenderPlanet(tt.planet)
			})

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected log to contain '%s', got: %s", tt.expected, output)
			}
		})
	}
}

func TestNullRenderer_RenderProjectile_LogsProjectileInformation(t *testing.T) {
	tests := []struct {
		name       string
		projectile *entity.Projectile
		expected   string
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
			expected: "RenderProjectile called for projectile:",
		},
		{
			name:       "NilProjectile_HandlesGracefully",
			projectile: nil,
			expected:   "RenderProjectile called for projectile:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := &NullRenderer{}

			output := captureLog(func() {
				renderer.RenderProjectile(tt.projectile)
			})

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected log to contain '%s', got: %s", tt.expected, output)
			}
		})
	}
}

func TestNullRenderer_ImplementsRendererInterface(t *testing.T) {
	var renderer entity.Renderer = &NullRenderer{}

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

	// Verify we can use it like any other renderer
	output := captureLog(func() {
		renderer.Clear()
	})

	if !strings.Contains(output, "Clear called") {
		t.Errorf("Global NullRendererInstance should work like NullRenderer instance")
	}
}

func TestNullRenderer_ConcurrentUsage_ThreadSafe(t *testing.T) {
	renderer := &NullRenderer{}
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
	// Integration test to ensure all methods produce some log output
	renderer := &NullRenderer{}

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
		t.Run(method.name+"_ProducesOutput", func(t *testing.T) {
			output := captureLog(method.call)

			if strings.TrimSpace(output) == "" {
				t.Errorf("Method %s should produce log output, but got empty string", method.name)
			}
		})
	}
}
