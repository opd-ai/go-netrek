// pkg/physics/vector_test.go
package physics

import (
	"math"
	"testing"
)

func TestVector2D_Add(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector2D
		v2       Vector2D
		expected Vector2D
	}{
		{
			name:     "positive_vectors",
			v1:       Vector2D{X: 3, Y: 4},
			v2:       Vector2D{X: 1, Y: 2},
			expected: Vector2D{X: 4, Y: 6},
		},
		{
			name:     "negative_vectors",
			v1:       Vector2D{X: -3, Y: -4},
			v2:       Vector2D{X: -1, Y: -2},
			expected: Vector2D{X: -4, Y: -6},
		},
		{
			name:     "mixed_signs",
			v1:       Vector2D{X: 5, Y: -3},
			v2:       Vector2D{X: -2, Y: 7},
			expected: Vector2D{X: 3, Y: 4},
		},
		{
			name:     "zero_vector",
			v1:       Vector2D{X: 0, Y: 0},
			v2:       Vector2D{X: 5, Y: -3},
			expected: Vector2D{X: 5, Y: -3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Add(tt.v2)
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("Add() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Sub(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector2D
		v2       Vector2D
		expected Vector2D
	}{
		{
			name:     "positive_result",
			v1:       Vector2D{X: 5, Y: 7},
			v2:       Vector2D{X: 2, Y: 3},
			expected: Vector2D{X: 3, Y: 4},
		},
		{
			name:     "negative_result",
			v1:       Vector2D{X: 2, Y: 3},
			v2:       Vector2D{X: 5, Y: 7},
			expected: Vector2D{X: -3, Y: -4},
		},
		{
			name:     "same_vectors",
			v1:       Vector2D{X: 4, Y: 6},
			v2:       Vector2D{X: 4, Y: 6},
			expected: Vector2D{X: 0, Y: 0},
		},
		{
			name:     "subtract_zero",
			v1:       Vector2D{X: 4, Y: 6},
			v2:       Vector2D{X: 0, Y: 0},
			expected: Vector2D{X: 4, Y: 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Sub(tt.v2)
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("Sub() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Scale(t *testing.T) {
	tests := []struct {
		name     string
		vector   Vector2D
		factor   float64
		expected Vector2D
	}{
		{
			name:     "positive_scale",
			vector:   Vector2D{X: 3, Y: 4},
			factor:   2,
			expected: Vector2D{X: 6, Y: 8},
		},
		{
			name:     "negative_scale",
			vector:   Vector2D{X: 3, Y: 4},
			factor:   -2,
			expected: Vector2D{X: -6, Y: -8},
		},
		{
			name:     "zero_scale",
			vector:   Vector2D{X: 3, Y: 4},
			factor:   0,
			expected: Vector2D{X: 0, Y: 0},
		},
		{
			name:     "fractional_scale",
			vector:   Vector2D{X: 4, Y: 8},
			factor:   0.5,
			expected: Vector2D{X: 2, Y: 4},
		},
		{
			name:     "identity_scale",
			vector:   Vector2D{X: 3, Y: 4},
			factor:   1,
			expected: Vector2D{X: 3, Y: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.vector.Scale(tt.factor)
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("Scale() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Length(t *testing.T) {
	tests := []struct {
		name     string
		vector   Vector2D
		expected float64
	}{
		{
			name:     "unit_vector_x",
			vector:   Vector2D{X: 1, Y: 0},
			expected: 1,
		},
		{
			name:     "unit_vector_y",
			vector:   Vector2D{X: 0, Y: 1},
			expected: 1,
		},
		{
			name:     "zero_vector",
			vector:   Vector2D{X: 0, Y: 0},
			expected: 0,
		},
		{
			name:     "pythagorean_triple",
			vector:   Vector2D{X: 3, Y: 4},
			expected: 5,
		},
		{
			name:     "negative_components",
			vector:   Vector2D{X: -3, Y: -4},
			expected: 5,
		},
		{
			name:     "mixed_signs",
			vector:   Vector2D{X: -3, Y: 4},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.vector.Length()
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Length() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_LengthSquared(t *testing.T) {
	tests := []struct {
		name     string
		vector   Vector2D
		expected float64
	}{
		{
			name:     "unit_vector",
			vector:   Vector2D{X: 1, Y: 0},
			expected: 1,
		},
		{
			name:     "zero_vector",
			vector:   Vector2D{X: 0, Y: 0},
			expected: 0,
		},
		{
			name:     "pythagorean_triple",
			vector:   Vector2D{X: 3, Y: 4},
			expected: 25,
		},
		{
			name:     "negative_components",
			vector:   Vector2D{X: -3, Y: -4},
			expected: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.vector.LengthSquared()
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("LengthSquared() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Normalize(t *testing.T) {
	t.Run("unit_vector_unchanged", func(t *testing.T) {
		vector := Vector2D{X: 1, Y: 0}
		result := vector.Normalize()
		expected := Vector2D{X: 1, Y: 0}

		if math.Abs(result.X-expected.X) > 1e-9 || math.Abs(result.Y-expected.Y) > 1e-9 {
			t.Errorf("Normalize() = %v, expected %v", result, expected)
		}
	})

	t.Run("zero_vector", func(t *testing.T) {
		vector := Vector2D{X: 0, Y: 0}
		result := vector.Normalize()
		expected := Vector2D{X: 1, Y: 0} // Default unit vector for zero-length input

		if result.X != expected.X || result.Y != expected.Y {
			t.Errorf("Normalize() = %v, expected %v", result, expected)
		}

		// Verify the result is actually a unit vector
		length := result.Length()
		if math.Abs(length-1) > 1e-9 {
			t.Errorf("Normalized zero vector length = %v, expected 1", length)
		}
	})

	t.Run("regular_vector", func(t *testing.T) {
		vector := Vector2D{X: 3, Y: 4}
		result := vector.Normalize()

		// Check that the result has unit length
		length := result.Length()
		if math.Abs(length-1) > 1e-9 {
			t.Errorf("Normalized vector length = %v, expected 1", length)
		}

		// Check direction is preserved
		expectedX := 3.0 / 5.0
		expectedY := 4.0 / 5.0
		if math.Abs(result.X-expectedX) > 1e-9 || math.Abs(result.Y-expectedY) > 1e-9 {
			t.Errorf("Normalize() = %v, expected approximately (%v, %v)", result, expectedX, expectedY)
		}
	})

	t.Run("negative_vector", func(t *testing.T) {
		vector := Vector2D{X: -6, Y: -8}
		result := vector.Normalize()

		// Check that the result has unit length
		length := result.Length()
		if math.Abs(length-1) > 1e-9 {
			t.Errorf("Normalized vector length = %v, expected 1", length)
		}
	})
}

func TestVector2D_NormalizeZeroVector_ReturnsUnitVector(t *testing.T) {
	// Test that normalizing a zero vector returns a default unit vector
	// instead of returning zero vector which would create mathematical inconsistencies
	zeroVector := Vector2D{X: 0, Y: 0}
	normalized := zeroVector.Normalize()

	// Should return default unit vector (1, 0)
	expected := Vector2D{X: 1, Y: 0}
	if normalized.X != expected.X || normalized.Y != expected.Y {
		t.Errorf("Normalize() on zero vector = %v, expected %v", normalized, expected)
	}

	// Verify it's actually a unit vector
	length := normalized.Length()
	if math.Abs(length-1) > 1e-9 {
		t.Errorf("Normalized zero vector length = %v, expected 1", length)
	}

	// Regression test: ensure we don't return zero vector which breaks physics
	if normalized.X == 0 && normalized.Y == 0 {
		t.Error("Normalize() on zero vector must not return zero vector - this creates mathematical inconsistencies in physics calculations")
	}
}

func TestVector2D_Distance(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector2D
		v2       Vector2D
		expected float64
	}{
		{
			name:     "same_point",
			v1:       Vector2D{X: 3, Y: 4},
			v2:       Vector2D{X: 3, Y: 4},
			expected: 0,
		},
		{
			name:     "unit_distance_x",
			v1:       Vector2D{X: 0, Y: 0},
			v2:       Vector2D{X: 1, Y: 0},
			expected: 1,
		},
		{
			name:     "unit_distance_y",
			v1:       Vector2D{X: 0, Y: 0},
			v2:       Vector2D{X: 0, Y: 1},
			expected: 1,
		},
		{
			name:     "pythagorean_distance",
			v1:       Vector2D{X: 0, Y: 0},
			v2:       Vector2D{X: 3, Y: 4},
			expected: 5,
		},
		{
			name:     "negative_coordinates",
			v1:       Vector2D{X: -1, Y: -1},
			v2:       Vector2D{X: 2, Y: 3},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Distance(tt.v2)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Distance() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Angle(t *testing.T) {
	tests := []struct {
		name     string
		vector   Vector2D
		expected float64
	}{
		{
			name:     "positive_x_axis",
			vector:   Vector2D{X: 1, Y: 0},
			expected: 0,
		},
		{
			name:     "positive_y_axis",
			vector:   Vector2D{X: 0, Y: 1},
			expected: math.Pi / 2,
		},
		{
			name:     "negative_x_axis",
			vector:   Vector2D{X: -1, Y: 0},
			expected: math.Pi,
		},
		{
			name:     "negative_y_axis",
			vector:   Vector2D{X: 0, Y: -1},
			expected: -math.Pi / 2,
		},
		{
			name:     "45_degrees",
			vector:   Vector2D{X: 1, Y: 1},
			expected: math.Pi / 4,
		},
		{
			name:     "135_degrees",
			vector:   Vector2D{X: -1, Y: 1},
			expected: 3 * math.Pi / 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.vector.Angle()
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Angle() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFromAngle(t *testing.T) {
	tests := []struct {
		name      string
		angle     float64
		magnitude float64
		expectedX float64
		expectedY float64
	}{
		{
			name:      "zero_angle_unit_magnitude",
			angle:     0,
			magnitude: 1,
			expectedX: 1,
			expectedY: 0,
		},
		{
			name:      "90_degrees_unit_magnitude",
			angle:     math.Pi / 2,
			magnitude: 1,
			expectedX: 0,
			expectedY: 1,
		},
		{
			name:      "180_degrees_unit_magnitude",
			angle:     math.Pi,
			magnitude: 1,
			expectedX: -1,
			expectedY: 0,
		},
		{
			name:      "270_degrees_unit_magnitude",
			angle:     -math.Pi / 2,
			magnitude: 1,
			expectedX: 0,
			expectedY: -1,
		},
		{
			name:      "45_degrees_magnitude_2",
			angle:     math.Pi / 4,
			magnitude: 2,
			expectedX: math.Sqrt(2),
			expectedY: math.Sqrt(2),
		},
		{
			name:      "zero_magnitude",
			angle:     math.Pi / 4,
			magnitude: 0,
			expectedX: 0,
			expectedY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromAngle(tt.angle, tt.magnitude)
			if math.Abs(result.X-tt.expectedX) > 1e-9 || math.Abs(result.Y-tt.expectedY) > 1e-9 {
				t.Errorf("FromAngle() = %v, expected (%v, %v)", result, tt.expectedX, tt.expectedY)
			}
		})
	}
}

func TestVector2D_Dot(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector2D
		v2       Vector2D
		expected float64
	}{
		{
			name:     "orthogonal_vectors",
			v1:       Vector2D{X: 1, Y: 0},
			v2:       Vector2D{X: 0, Y: 1},
			expected: 0,
		},
		{
			name:     "parallel_vectors",
			v1:       Vector2D{X: 1, Y: 0},
			v2:       Vector2D{X: 2, Y: 0},
			expected: 2,
		},
		{
			name:     "antiparallel_vectors",
			v1:       Vector2D{X: 1, Y: 0},
			v2:       Vector2D{X: -2, Y: 0},
			expected: -2,
		},
		{
			name:     "general_vectors",
			v1:       Vector2D{X: 3, Y: 4},
			v2:       Vector2D{X: 1, Y: 2},
			expected: 11, // 3*1 + 4*2 = 11
		},
		{
			name:     "zero_vectors",
			v1:       Vector2D{X: 0, Y: 0},
			v2:       Vector2D{X: 5, Y: 3},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Dot(tt.v2)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Dot() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Rotate(t *testing.T) {
	tests := []struct {
		name      string
		vector    Vector2D
		angle     float64
		expectedX float64
		expectedY float64
	}{
		{
			name:      "no_rotation",
			vector:    Vector2D{X: 1, Y: 0},
			angle:     0,
			expectedX: 1,
			expectedY: 0,
		},
		{
			name:      "90_degree_rotation",
			vector:    Vector2D{X: 1, Y: 0},
			angle:     math.Pi / 2,
			expectedX: 0,
			expectedY: 1,
		},
		{
			name:      "180_degree_rotation",
			vector:    Vector2D{X: 1, Y: 0},
			angle:     math.Pi,
			expectedX: -1,
			expectedY: 0,
		},
		{
			name:      "270_degree_rotation",
			vector:    Vector2D{X: 1, Y: 0},
			angle:     -math.Pi / 2,
			expectedX: 0,
			expectedY: -1,
		},
		{
			name:      "45_degree_rotation",
			vector:    Vector2D{X: 1, Y: 0},
			angle:     math.Pi / 4,
			expectedX: math.Cos(math.Pi / 4),
			expectedY: math.Sin(math.Pi / 4),
		},
		{
			name:      "rotate_arbitrary_vector",
			vector:    Vector2D{X: 2, Y: 3},
			angle:     math.Pi / 2,
			expectedX: -3,
			expectedY: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.vector.Rotate(tt.angle)
			if math.Abs(result.X-tt.expectedX) > 1e-9 || math.Abs(result.Y-tt.expectedY) > 1e-9 {
				t.Errorf("Rotate() = %v, expected (%v, %v)", result, tt.expectedX, tt.expectedY)
			}
		})
	}
}

// Benchmark tests for performance verification
func BenchmarkVector2D_Add(b *testing.B) {
	v1 := Vector2D{X: 3, Y: 4}
	v2 := Vector2D{X: 1, Y: 2}

	for i := 0; i < b.N; i++ {
		_ = v1.Add(v2)
	}
}

func BenchmarkVector2D_Length(b *testing.B) {
	v := Vector2D{X: 3, Y: 4}

	for i := 0; i < b.N; i++ {
		_ = v.Length()
	}
}

func BenchmarkVector2D_Normalize(b *testing.B) {
	v := Vector2D{X: 3, Y: 4}

	for i := 0; i < b.N; i++ {
		_ = v.Normalize()
	}
}

func BenchmarkVector2D_Rotate(b *testing.B) {
	v := Vector2D{X: 3, Y: 4}
	angle := math.Pi / 4

	for i := 0; i < b.N; i++ {
		_ = v.Rotate(angle)
	}
}
