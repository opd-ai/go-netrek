package physics

// MovementState tracks ship physics
type MovementState struct {
	Position Vector2D
	Velocity Vector2D
	Heading  float64 // radians
	Mass     float64
	Thrust   float64
	MaxSpeed float64
}

func UpdateMovement(state *MovementState, deltaTime float64, thrustInput float64, turnInput float64) {
	// Apply rotation
	state.Heading += turnInput * deltaTime

	// Calculate thrust vector
	thrustVector := FromAngle(state.Heading, thrustInput*state.Thrust)

	// Update velocity
	state.Velocity = state.Velocity.Add(thrustVector.Scale(deltaTime))

	// Limit speed
	if state.Velocity.Length() > state.MaxSpeed {
		state.Velocity = state.Velocity.Normalize().Scale(state.MaxSpeed)
	}

	// Update position
	state.Position = state.Position.Add(state.Velocity.Scale(deltaTime))
}
