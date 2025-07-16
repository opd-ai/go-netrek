package physics

import (
	"math"
	"testing"
)

func TestUpdateMovement_BasicMovement(t *testing.T) {
	state := &MovementState{
		Position: Vector2D{X: 0, Y: 0},
		Velocity: Vector2D{X: 0, Y: 0},
		Heading:  0,
		Mass:     100,
		Thrust:   200,
		MaxSpeed: 500,
	}

	// Apply thrust forward for 1 second
	UpdateMovement(state, 1.0, 1.0, 0.0)

	// Should have moved forward and gained velocity
	if state.Position.X <= 0 {
		t.Errorf("Expected positive X position, got %f", state.Position.X)
	}
	if state.Velocity.X <= 0 {
		t.Errorf("Expected positive X velocity, got %f", state.Velocity.X)
	}
	if state.Heading != 0 {
		t.Errorf("Expected heading to remain 0, got %f", state.Heading)
	}
}

func TestUpdateMovement_Rotation(t *testing.T) {
	tests := []struct {
		name        string
		turnInput   float64
		deltaTime   float64
		expectedDir float64 // Expected direction change
	}{
		{
			name:        "Turn right",
			turnInput:   1.0,
			deltaTime:   1.0,
			expectedDir: 1.0,
		},
		{
			name:        "Turn left",
			turnInput:   -1.0,
			deltaTime:   1.0,
			expectedDir: -1.0,
		},
		{
			name:        "No turn",
			turnInput:   0.0,
			deltaTime:   1.0,
			expectedDir: 0.0,
		},
		{
			name:        "Half turn right",
			turnInput:   1.0,
			deltaTime:   0.5,
			expectedDir: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &MovementState{
				Position: Vector2D{X: 0, Y: 0},
				Velocity: Vector2D{X: 0, Y: 0},
				Heading:  0,
				Mass:     100,
				Thrust:   200,
				MaxSpeed: 500,
			}

			UpdateMovement(state, tt.deltaTime, 0.0, tt.turnInput)

			expectedHeading := tt.expectedDir
			if math.Abs(state.Heading-expectedHeading) > 1e-9 {
				t.Errorf("Expected heading %f, got %f", expectedHeading, state.Heading)
			}
		})
	}
}

func TestUpdateMovement_ThrustAndMovement(t *testing.T) {
	tests := []struct {
		name         string
		thrustInput  float64
		deltaTime    float64
		initialVel   Vector2D
		expectingVel bool // Whether we expect velocity change
	}{
		{
			name:         "Full thrust",
			thrustInput:  1.0,
			deltaTime:    1.0,
			initialVel:   Vector2D{X: 0, Y: 0},
			expectingVel: true,
		},
		{
			name:         "No thrust",
			thrustInput:  0.0,
			deltaTime:    1.0,
			initialVel:   Vector2D{X: 0, Y: 0},
			expectingVel: false,
		},
		{
			name:         "Half thrust",
			thrustInput:  0.5,
			deltaTime:    1.0,
			initialVel:   Vector2D{X: 0, Y: 0},
			expectingVel: true,
		},
		{
			name:         "Reverse thrust",
			thrustInput:  -1.0,
			deltaTime:    1.0,
			initialVel:   Vector2D{X: 0, Y: 0},
			expectingVel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &MovementState{
				Position: Vector2D{X: 0, Y: 0},
				Velocity: tt.initialVel,
				Heading:  0, // Facing right (positive X)
				Mass:     100,
				Thrust:   200,
				MaxSpeed: 500,
			}

			initialVelLength := state.Velocity.Length()
			UpdateMovement(state, tt.deltaTime, tt.thrustInput, 0.0)

			velChanged := math.Abs(state.Velocity.Length()-initialVelLength) > 1e-9
			if tt.expectingVel && !velChanged {
				t.Errorf("Expected velocity to change, but it didn't")
			}
			if !tt.expectingVel && velChanged {
				t.Errorf("Expected velocity to remain the same, but it changed")
			}
		})
	}
}

func TestUpdateMovement_SpeedLimit(t *testing.T) {
	state := &MovementState{
		Position: Vector2D{X: 0, Y: 0},
		Velocity: Vector2D{X: 0, Y: 0},
		Heading:  0,
		Mass:     100,
		Thrust:   1000, // High thrust
		MaxSpeed: 100,  // Low max speed
	}

	// Apply thrust for many iterations to build up speed
	for i := 0; i < 100; i++ {
		UpdateMovement(state, 0.1, 1.0, 0.0)
	}

	// Velocity should be capped at MaxSpeed
	speed := state.Velocity.Length()
	if speed > state.MaxSpeed+1e-9 {
		t.Errorf("Speed %f exceeds MaxSpeed %f", speed, state.MaxSpeed)
	}

	// Should be very close to max speed
	if speed < state.MaxSpeed-1 {
		t.Errorf("Speed %f is unexpectedly low compared to MaxSpeed %f", speed, state.MaxSpeed)
	}
}

func TestUpdateMovement_PositionUpdate(t *testing.T) {
	state := &MovementState{
		Position: Vector2D{X: 100, Y: 200},
		Velocity: Vector2D{X: 50, Y: -25}, // Moving right and up
		Heading:  0,
		Mass:     100,
		Thrust:   0, // No thrust, just coast
		MaxSpeed: 500,
	}

	deltaTime := 2.0
	expectedNewPos := Vector2D{
		X: state.Position.X + state.Velocity.X*deltaTime,
		Y: state.Position.Y + state.Velocity.Y*deltaTime,
	}

	UpdateMovement(state, deltaTime, 0.0, 0.0)

	if math.Abs(state.Position.X-expectedNewPos.X) > 1e-9 {
		t.Errorf("Expected X position %f, got %f", expectedNewPos.X, state.Position.X)
	}
	if math.Abs(state.Position.Y-expectedNewPos.Y) > 1e-9 {
		t.Errorf("Expected Y position %f, got %f", expectedNewPos.Y, state.Position.Y)
	}
}

func TestUpdateMovement_CombinedOperations(t *testing.T) {
	state := &MovementState{
		Position: Vector2D{X: 0, Y: 0},
		Velocity: Vector2D{X: 0, Y: 0},
		Heading:  0,
		Mass:     100,
		Thrust:   200,
		MaxSpeed: 500,
	}

	// Turn 90 degrees and apply thrust
	UpdateMovement(state, 1.0, 1.0, math.Pi/2)

	// Should be facing up (positive Y direction) and have velocity
	if math.Abs(state.Heading-math.Pi/2) > 1e-9 {
		t.Errorf("Expected heading %f, got %f", math.Pi/2, state.Heading)
	}

	// Velocity should be primarily in Y direction due to the thrust after rotation
	if state.Velocity.Length() <= 0 {
		t.Errorf("Expected non-zero velocity")
	}

	// Position should have changed
	if state.Position.X == 0 && state.Position.Y == 0 {
		t.Errorf("Expected position to change")
	}
}

func TestUpdateMovement_ZeroDeltaTime(t *testing.T) {
	state := &MovementState{
		Position: Vector2D{X: 100, Y: 200},
		Velocity: Vector2D{X: 50, Y: 25},
		Heading:  1.5,
		Mass:     100,
		Thrust:   200,
		MaxSpeed: 500,
	}

	originalState := *state

	// Update with zero delta time - nothing should change
	UpdateMovement(state, 0.0, 1.0, 1.0)

	if state.Position != originalState.Position {
		t.Errorf("Position changed with zero delta time")
	}
	if state.Velocity != originalState.Velocity {
		t.Errorf("Velocity changed with zero delta time")
	}
	if state.Heading != originalState.Heading {
		t.Errorf("Heading changed with zero delta time")
	}
}
