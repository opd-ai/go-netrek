package render

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// TestNewTerminalRenderer tests the creation of a new terminal renderer
func TestNewTerminalRenderer_CreatesValidRenderer_WithCorrectDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		scale  float64
	}{
		{
			name:   "small renderer",
			width:  10,
			height: 5,
			scale:  1.0,
		},
		{
			name:   "medium renderer",
			width:  80,
			height: 24,
			scale:  10.0,
		},
		{
			name:   "large renderer",
			width:  120,
			height: 40,
			scale:  5.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTerminalRenderer(tt.width, tt.height, tt.scale)

			if renderer == nil {
				t.Fatal("NewTerminalRenderer returned nil")
			}

			if renderer.width != tt.width {
				t.Errorf("expected width %d, got %d", tt.width, renderer.width)
			}

			if renderer.height != tt.height {
				t.Errorf("expected height %d, got %d", tt.height, renderer.height)
			}

			if renderer.scale != tt.scale {
				t.Errorf("expected scale %f, got %f", tt.scale, renderer.scale)
			}

			// Check buffer dimensions
			if len(renderer.buffer) != tt.height {
				t.Errorf("expected buffer height %d, got %d", tt.height, len(renderer.buffer))
			}

			for i, row := range renderer.buffer {
				if len(row) != tt.width {
					t.Errorf("row %d: expected width %d, got %d", i, tt.width, len(row))
				}
			}

			// Check center position is initialized to origin
			expectedCenter := physics.Vector2D{X: 0, Y: 0}
			if renderer.centerPos.X != expectedCenter.X || renderer.centerPos.Y != expectedCenter.Y {
				t.Errorf("expected center %v, got %v", expectedCenter, renderer.centerPos)
			}
		})
	}
}

// TestSetCenter tests setting the center position
func TestSetCenter_UpdatesCenterPosition_Correctly(t *testing.T) {
	renderer := NewTerminalRenderer(80, 24, 1.0)

	tests := []struct {
		name     string
		position physics.Vector2D
	}{
		{
			name:     "origin",
			position: physics.Vector2D{X: 0, Y: 0},
		},
		{
			name:     "positive coordinates",
			position: physics.Vector2D{X: 100.5, Y: 200.75},
		},
		{
			name:     "negative coordinates",
			position: physics.Vector2D{X: -50.25, Y: -75.5},
		},
		{
			name:     "mixed coordinates",
			position: physics.Vector2D{X: -25.0, Y: 30.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer.SetCenter(tt.position)

			if renderer.centerPos.X != tt.position.X {
				t.Errorf("expected center X %f, got %f", tt.position.X, renderer.centerPos.X)
			}

			if renderer.centerPos.Y != tt.position.Y {
				t.Errorf("expected center Y %f, got %f", tt.position.Y, renderer.centerPos.Y)
			}
		})
	}
}

// TestWorldToScreen tests coordinate conversion from world to screen space
func TestWorldToScreen_ConvertsCoordinates_Correctly(t *testing.T) {
	renderer := NewTerminalRenderer(80, 24, 10.0) // 80x24 screen, scale 10

	tests := []struct {
		name      string
		centerPos physics.Vector2D
		worldPos  physics.Vector2D
		expectedX int
		expectedY int
	}{
		{
			name:      "center at origin, world at origin",
			centerPos: physics.Vector2D{X: 0, Y: 0},
			worldPos:  physics.Vector2D{X: 0, Y: 0},
			expectedX: 40, // width/2
			expectedY: 12, // height/2
		},
		{
			name:      "center at origin, world offset",
			centerPos: physics.Vector2D{X: 0, Y: 0},
			worldPos:  physics.Vector2D{X: 100, Y: 50},
			expectedX: 50, // 40 + 100/10
			expectedY: 17, // 12 + 50/10
		},
		{
			name:      "center offset, world at origin",
			centerPos: physics.Vector2D{X: 50, Y: 25},
			worldPos:  physics.Vector2D{X: 0, Y: 0},
			expectedX: 35, // 40 + (0-50)/10
			expectedY: 9,  // 12 + (0-25)/10
		},
		{
			name:      "both center and world offset",
			centerPos: physics.Vector2D{X: 100, Y: 50},
			worldPos:  physics.Vector2D{X: 200, Y: 150},
			expectedX: 50, // 40 + (200-100)/10
			expectedY: 22, // 12 + (150-50)/10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer.SetCenter(tt.centerPos)
			x, y := renderer.worldToScreen(tt.worldPos)

			if x != tt.expectedX {
				t.Errorf("expected screen X %d, got %d", tt.expectedX, x)
			}

			if y != tt.expectedY {
				t.Errorf("expected screen Y %d, got %d", tt.expectedY, y)
			}
		})
	}
}

// TestClear tests clearing the buffer
func TestClear_ClearsBuffer_WithSpaces(t *testing.T) {
	renderer := NewTerminalRenderer(10, 5, 1.0)

	// Fill buffer with some characters
	for y := 0; y < renderer.height; y++ {
		for x := 0; x < renderer.width; x++ {
			renderer.buffer[y][x] = 'X'
		}
	}

	// Clear the buffer
	renderer.Clear()

	// Verify all positions are spaces
	for y := 0; y < renderer.height; y++ {
		for x := 0; x < renderer.width; x++ {
			if renderer.buffer[y][x] != ' ' {
				t.Errorf("position (%d, %d) expected space, got %c", x, y, renderer.buffer[y][x])
			}
		}
	}
}

// TestRenderShip tests ship rendering
func TestRenderShip_RendersShip_AtCorrectPosition(t *testing.T) {
	renderer := NewTerminalRenderer(20, 10, 1.0)
	renderer.Clear()

	tests := []struct {
		name         string
		ship         *entity.Ship
		expectedChar rune
		expectRender bool
	}{
		{
			name: "ship at center",
			ship: &entity.Ship{
				BaseEntity: entity.BaseEntity{
					Position: physics.Vector2D{X: 0, Y: 0},
				},
				Class: entity.Scout,
			},
			expectedChar: 'S',
			expectRender: true,
		},
		{
			name: "ship out of bounds",
			ship: &entity.Ship{
				BaseEntity: entity.BaseEntity{
					Position: physics.Vector2D{X: 1000, Y: 1000},
				},
				Class: entity.Destroyer,
			},
			expectedChar: 'D',
			expectRender: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer.Clear()
			renderer.RenderShip(tt.ship)

			if tt.expectRender {
				// Find the rendered character
				found := false
				x, y := renderer.worldToScreen(tt.ship.Position)
				if x >= 0 && x < renderer.width && y >= 0 && y < renderer.height {
					if renderer.buffer[y][x] == tt.expectedChar {
						found = true
					}
				}

				if !found {
					t.Errorf("expected to find character %c at screen position, but didn't", tt.expectedChar)
				}
			} else {
				// Verify nothing was rendered (all spaces)
				for y := 0; y < renderer.height; y++ {
					for x := 0; x < renderer.width; x++ {
						if renderer.buffer[y][x] != ' ' {
							t.Errorf("expected no rendering, but found %c at (%d, %d)", renderer.buffer[y][x], x, y)
						}
					}
				}
			}
		})
	}
}

// TestRenderPlanet tests planet rendering
func TestRenderPlanet_RendersPlanet_AtCorrectPosition(t *testing.T) {
	renderer := NewTerminalRenderer(20, 10, 1.0)
	renderer.Clear()

	tests := []struct {
		name         string
		planet       *entity.Planet
		expectRender bool
	}{
		{
			name: "planet at center",
			planet: &entity.Planet{
				BaseEntity: entity.BaseEntity{
					Position: physics.Vector2D{X: 0, Y: 0},
				},
			},
			expectRender: true,
		},
		{
			name: "planet out of bounds",
			planet: &entity.Planet{
				BaseEntity: entity.BaseEntity{
					Position: physics.Vector2D{X: 1000, Y: 1000},
				},
			},
			expectRender: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer.Clear()
			renderer.RenderPlanet(tt.planet)

			if tt.expectRender {
				// Find the rendered character
				found := false
				x, y := renderer.worldToScreen(tt.planet.Position)
				if x >= 0 && x < renderer.width && y >= 0 && y < renderer.height {
					if renderer.buffer[y][x] == 'O' {
						found = true
					}
				}

				if !found {
					t.Error("expected to find character 'O' at screen position, but didn't")
				}
			} else {
				// Verify nothing was rendered (all spaces)
				for y := 0; y < renderer.height; y++ {
					for x := 0; x < renderer.width; x++ {
						if renderer.buffer[y][x] != ' ' {
							t.Errorf("expected no rendering, but found %c at (%d, %d)", renderer.buffer[y][x], x, y)
						}
					}
				}
			}
		})
	}
}

// TestRenderProjectile tests projectile rendering
func TestRenderProjectile_RendersProjectile_AtCorrectPosition(t *testing.T) {
	renderer := NewTerminalRenderer(20, 10, 1.0)
	renderer.Clear()

	tests := []struct {
		name         string
		projectile   *entity.Projectile
		expectRender bool
	}{
		{
			name: "projectile at center",
			projectile: &entity.Projectile{
				BaseEntity: entity.BaseEntity{
					Position: physics.Vector2D{X: 0, Y: 0},
				},
			},
			expectRender: true,
		},
		{
			name: "projectile out of bounds",
			projectile: &entity.Projectile{
				BaseEntity: entity.BaseEntity{
					Position: physics.Vector2D{X: 1000, Y: 1000},
				},
			},
			expectRender: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer.Clear()
			renderer.RenderProjectile(tt.projectile)

			if tt.expectRender {
				// Find the rendered character
				found := false
				x, y := renderer.worldToScreen(tt.projectile.Position)
				if x >= 0 && x < renderer.width && y >= 0 && y < renderer.height {
					if renderer.buffer[y][x] == '.' {
						found = true
					}
				}

				if !found {
					t.Error("expected to find character '.' at screen position, but didn't")
				}
			} else {
				// Verify nothing was rendered (all spaces)
				for y := 0; y < renderer.height; y++ {
					for x := 0; x < renderer.width; x++ {
						if renderer.buffer[y][x] != ' ' {
							t.Errorf("expected no rendering, but found %c at (%d, %d)", renderer.buffer[y][x], x, y)
						}
					}
				}
			}
		})
	}
}

// TestPresent tests the present method doesn't panic
func TestPresent_ExecutesWithoutError_ForVariousSizes(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small", 5, 3},
		{"medium", 20, 10},
		{"large", 80, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTerminalRenderer(tt.width, tt.height, 1.0)
			renderer.Clear()

			// This test mainly ensures Present doesn't panic
			// In a real test environment, we might capture stdout to verify output
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Present() panicked: %v", r)
				}
			}()

			renderer.Present()
		})
	}
}

// TestIntegration tests rendering multiple entities together
func TestIntegration_RendersMultipleEntities_Correctly(t *testing.T) {
	renderer := NewTerminalRenderer(20, 10, 2.0)
	renderer.Clear()

	// Create test entities at different positions
	ship := &entity.Ship{
		BaseEntity: entity.BaseEntity{
			Position: physics.Vector2D{X: 0, Y: 0},
		},
		Class: entity.Scout,
	}

	planet := &entity.Planet{
		BaseEntity: entity.BaseEntity{
			Position: physics.Vector2D{X: 4, Y: 2},
		},
	}

	projectile := &entity.Projectile{
		BaseEntity: entity.BaseEntity{
			Position: physics.Vector2D{X: -2, Y: -1},
		},
	}

	// Render all entities
	renderer.RenderShip(ship)
	renderer.RenderPlanet(planet)
	renderer.RenderProjectile(projectile)

	// Verify each entity is rendered at the correct position
	shipX, shipY := renderer.worldToScreen(ship.Position)
	if renderer.buffer[shipY][shipX] != 'S' {
		t.Errorf("ship not rendered at expected position (%d, %d)", shipX, shipY)
	}

	planetX, planetY := renderer.worldToScreen(planet.Position)
	if renderer.buffer[planetY][planetX] != 'O' {
		t.Errorf("planet not rendered at expected position (%d, %d)", planetX, planetY)
	}

	projX, projY := renderer.worldToScreen(projectile.Position)
	if renderer.buffer[projY][projX] != '.' {
		t.Errorf("projectile not rendered at expected position (%d, %d)", projX, projY)
	}
}
