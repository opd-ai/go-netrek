// pkg/physics/vector.go
package physics

import "math"

// Vector2D represents a 2D vector with x and y components
type Vector2D struct {
	X float64
	Y float64
}

// Add returns the sum of two vectors
func (v Vector2D) Add(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

// Sub returns the difference between two vectors
func (v Vector2D) Sub(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
}

// Scale multiplies the vector by a scalar value
func (v Vector2D) Scale(factor float64) Vector2D {
	return Vector2D{
		X: v.X * factor,
		Y: v.Y * factor,
	}
}

// Length returns the magnitude of the vector
func (v Vector2D) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Normalize returns a unit vector in the same direction
func (v Vector2D) Normalize() Vector2D {
	length := v.Length()
	if length == 0 {
		return Vector2D{}
	}
	return Vector2D{
		X: v.X / length,
		Y: v.Y / length,
	}
}

// Distance returns the distance between two vectors
func (v Vector2D) Distance(other Vector2D) float64 {
	return v.Sub(other).Length()
}

// Angle returns the angle of the vector in radians
func (v Vector2D) Angle() float64 {
	return math.Atan2(v.Y, v.X)
}

// FromAngle creates a vector from an angle and magnitude
func FromAngle(angle float64, magnitude float64) Vector2D {
	return Vector2D{
		X: magnitude * math.Cos(angle),
		Y: magnitude * math.Sin(angle),
	}
}

// Dot returns the dot product of two vectors
func (v Vector2D) Dot(other Vector2D) float64 {
	return v.X*other.X + v.Y*other.Y
}

// Rotate rotates the vector by angle (in radians)
func (v Vector2D) Rotate(angle float64) Vector2D {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Vector2D{
		X: v.X*cos - v.Y*sin,
		Y: v.X*sin + v.Y*cos,
	}
}

// LengthSquared returns magnitude squared (optimization for comparisons)
func (v Vector2D) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y
}
