// pkg/physics/collision_test.go
package physics

import (
	"testing"
)

func TestCircle_Collides(t *testing.T) {
	tests := []struct {
		name     string
		circle1  Circle
		circle2  Circle
		expected bool
	}{
		{
			name:     "circles_touching",
			circle1:  Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5},
			circle2:  Circle{Center: Vector2D{X: 10, Y: 0}, Radius: 5},
			expected: false, // Distance equals sum of radii, collision logic uses <
		},
		{
			name:     "circles_overlapping",
			circle1:  Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5},
			circle2:  Circle{Center: Vector2D{X: 5, Y: 0}, Radius: 5},
			expected: true,
		},
		{
			name:     "circles_not_touching",
			circle1:  Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5},
			circle2:  Circle{Center: Vector2D{X: 15, Y: 0}, Radius: 5},
			expected: false,
		},
		{
			name:     "circles_same_position",
			circle1:  Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 3},
			circle2:  Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 2},
			expected: true,
		},
		{
			name:     "circles_diagonal_collision",
			circle1:  Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5},
			circle2:  Circle{Center: Vector2D{X: 3, Y: 4}, Radius: 3},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.circle1.Collides(tt.circle2)
			if result != tt.expected {
				t.Errorf("Circle.Collides() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCheckCollision(t *testing.T) {
	t.Run("no_collision", func(t *testing.T) {
		circle1 := Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5}
		circle2 := Circle{Center: Vector2D{X: 15, Y: 0}, Radius: 5}

		result := CheckCollision(circle1, circle2)

		if result.Collided {
			t.Error("Expected no collision, but got collision")
		}
	})

	t.Run("collision_with_penetration", func(t *testing.T) {
		circle1 := Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5}
		circle2 := Circle{Center: Vector2D{X: 8, Y: 0}, Radius: 5}

		result := CheckCollision(circle1, circle2)

		if !result.Collided {
			t.Error("Expected collision, but got no collision")
		}

		expectedPenetration := 2.0 // 5 + 5 - 8 = 2
		if result.Penetration != expectedPenetration {
			t.Errorf("Expected penetration %v, got %v", expectedPenetration, result.Penetration)
		}

		// Check normal vector is pointing from circle1 to circle2
		expectedNormal := Vector2D{X: 1, Y: 0}
		if result.Normal.X != expectedNormal.X || result.Normal.Y != expectedNormal.Y {
			t.Errorf("Expected normal %v, got %v", expectedNormal, result.Normal)
		}

		// Check contact point
		expectedContact := Vector2D{X: 5, Y: 0}
		if result.ContactPoint.X != expectedContact.X || result.ContactPoint.Y != expectedContact.Y {
			t.Errorf("Expected contact point %v, got %v", expectedContact, result.ContactPoint)
		}
	})

	t.Run("collision_diagonal", func(t *testing.T) {
		circle1 := Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 3}
		circle2 := Circle{Center: Vector2D{X: 3, Y: 4}, Radius: 3}

		result := CheckCollision(circle1, circle2)

		if !result.Collided {
			t.Error("Expected collision, but got no collision")
		}

		// Distance should be 5 (3-4-5 triangle), penetration should be 6 - 5 = 1
		expectedPenetration := 1.0
		if result.Penetration != expectedPenetration {
			t.Errorf("Expected penetration %v, got %v", expectedPenetration, result.Penetration)
		}
	})
}

func TestRect_Contains(t *testing.T) {
	rect := Rect{
		Center: Vector2D{X: 10, Y: 10},
		Width:  20,
		Height: 20,
	}

	tests := []struct {
		name     string
		point    Vector2D
		expected bool
	}{
		{
			name:     "point_inside_center",
			point:    Vector2D{X: 10, Y: 10},
			expected: true,
		},
		{
			name:     "point_inside_corner",
			point:    Vector2D{X: 5, Y: 5},
			expected: true,
		},
		{
			name:     "point_on_edge",
			point:    Vector2D{X: 0, Y: 10},
			expected: true,
		},
		{
			name:     "point_outside",
			point:    Vector2D{X: 25, Y: 25},
			expected: false,
		},
		{
			name:     "point_outside_negative",
			point:    Vector2D{X: -5, Y: 10},
			expected: false,
		},
		{
			name:     "point_on_boundary_edge",
			point:    Vector2D{X: 20, Y: 10},
			expected: false, // < not <=
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rect.Contains(tt.point)
			if result != tt.expected {
				t.Errorf("Rect.Contains(%v) = %v, expected %v", tt.point, result, tt.expected)
			}
		})
	}
}

func TestNewQuadTree(t *testing.T) {
	boundary := Rect{Center: Vector2D{X: 0, Y: 0}, Width: 100, Height: 100}
	capacity := 4

	qt := NewQuadTree(boundary, capacity)

	if qt.Boundary != boundary {
		t.Errorf("Expected boundary %v, got %v", boundary, qt.Boundary)
	}
	if qt.Capacity != capacity {
		t.Errorf("Expected capacity %d, got %d", capacity, qt.Capacity)
	}
	if qt.Divided {
		t.Error("New QuadTree should not be divided")
	}
	if len(qt.Points) != 0 {
		t.Errorf("Expected 0 points, got %d", len(qt.Points))
	}
	if len(qt.Objects) != 0 {
		t.Errorf("Expected 0 objects, got %d", len(qt.Objects))
	}
}

func TestQuadTree_Insert(t *testing.T) {
	boundary := Rect{Center: Vector2D{X: 0, Y: 0}, Width: 100, Height: 100}
	qt := NewQuadTree(boundary, 2)

	t.Run("insert_within_boundary", func(t *testing.T) {
		point := Vector2D{X: 10, Y: 10}
		object := "test_object"

		success := qt.Insert(point, object)

		if !success {
			t.Error("Insert should succeed for point within boundary")
		}
		if len(qt.Points) != 1 {
			t.Errorf("Expected 1 point, got %d", len(qt.Points))
		}
		if qt.Points[0] != point {
			t.Errorf("Expected point %v, got %v", point, qt.Points[0])
		}
		if qt.Objects[0] != object {
			t.Errorf("Expected object %v, got %v", object, qt.Objects[0])
		}
	})

	t.Run("insert_outside_boundary", func(t *testing.T) {
		point := Vector2D{X: 100, Y: 100}
		object := "outside_object"

		success := qt.Insert(point, object)

		if success {
			t.Error("Insert should fail for point outside boundary")
		}
	})

	t.Run("insert_causes_subdivision", func(t *testing.T) {
		// Insert enough points to exceed capacity
		points := []Vector2D{
			{X: -10, Y: -10},
			{X: 10, Y: 10},
		}

		for i, point := range points {
			qt.Insert(point, i)
		}

		if !qt.Divided {
			t.Error("QuadTree should be divided after exceeding capacity")
		}
		if qt.NorthWest == nil || qt.NorthEast == nil || qt.SouthWest == nil || qt.SouthEast == nil {
			t.Error("All quadrants should be created after subdivision")
		}
	})
}

func TestQuadTree_Subdivide(t *testing.T) {
	boundary := Rect{Center: Vector2D{X: 0, Y: 0}, Width: 100, Height: 100}
	qt := NewQuadTree(boundary, 4)

	qt.Subdivide()

	if !qt.Divided {
		t.Error("QuadTree should be marked as divided")
	}

	// Check that all quadrants are created
	if qt.NorthWest == nil {
		t.Error("NorthWest quadrant should be created")
	}
	if qt.NorthEast == nil {
		t.Error("NorthEast quadrant should be created")
	}
	if qt.SouthWest == nil {
		t.Error("SouthWest quadrant should be created")
	}
	if qt.SouthEast == nil {
		t.Error("SouthEast quadrant should be created")
	}

	// Check quadrant boundaries
	expectedNW := Rect{Center: Vector2D{X: -25, Y: 25}, Width: 50, Height: 50}
	if qt.NorthWest.Boundary != expectedNW {
		t.Errorf("NorthWest boundary expected %v, got %v", expectedNW, qt.NorthWest.Boundary)
	}

	expectedNE := Rect{Center: Vector2D{X: 25, Y: 25}, Width: 50, Height: 50}
	if qt.NorthEast.Boundary != expectedNE {
		t.Errorf("NorthEast boundary expected %v, got %v", expectedNE, qt.NorthEast.Boundary)
	}

	expectedSW := Rect{Center: Vector2D{X: -25, Y: -25}, Width: 50, Height: 50}
	if qt.SouthWest.Boundary != expectedSW {
		t.Errorf("SouthWest boundary expected %v, got %v", expectedSW, qt.SouthWest.Boundary)
	}

	expectedSE := Rect{Center: Vector2D{X: 25, Y: -25}, Width: 50, Height: 50}
	if qt.SouthEast.Boundary != expectedSE {
		t.Errorf("SouthEast boundary expected %v, got %v", expectedSE, qt.SouthEast.Boundary)
	}
}

func TestQuadTree_Query(t *testing.T) {
	boundary := Rect{Center: Vector2D{X: 0, Y: 0}, Width: 100, Height: 100}
	qt := NewQuadTree(boundary, 2)

	// Insert test objects
	points := []Vector2D{
		{X: -20, Y: -20}, // SouthWest
		{X: 20, Y: 20},   // NorthEast
		{X: -20, Y: 20},  // NorthWest
		{X: 20, Y: -20},  // SouthEast
	}

	objects := []string{"SW", "NE", "NW", "SE"}

	for i, point := range points {
		qt.Insert(point, objects[i])
	}

	t.Run("query_all", func(t *testing.T) {
		queryArea := Rect{Center: Vector2D{X: 0, Y: 0}, Width: 100, Height: 100}
		results := qt.Query(queryArea)

		if len(results) != 4 {
			t.Errorf("Expected 4 results, got %d", len(results))
		}
	})

	t.Run("query_northeast_quadrant", func(t *testing.T) {
		queryArea := Rect{Center: Vector2D{X: 25, Y: 25}, Width: 50, Height: 50}
		results := qt.Query(queryArea)

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0] != "NE" {
			t.Errorf("Expected 'NE', got %v", results[0])
		}
	})

	t.Run("query_outside_boundary", func(t *testing.T) {
		queryArea := Rect{Center: Vector2D{X: 200, Y: 200}, Width: 50, Height: 50}
		results := qt.Query(queryArea)

		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})
}

func TestQuadTree_intersects(t *testing.T) {
	boundary := Rect{Center: Vector2D{X: 0, Y: 0}, Width: 100, Height: 100}
	qt := NewQuadTree(boundary, 4)

	tests := []struct {
		name     string
		area     Rect
		expected bool
	}{
		{
			name:     "area_inside",
			area:     Rect{Center: Vector2D{X: 0, Y: 0}, Width: 50, Height: 50},
			expected: true,
		},
		{
			name:     "area_overlapping",
			area:     Rect{Center: Vector2D{X: 40, Y: 40}, Width: 50, Height: 50},
			expected: true,
		},
		{
			name:     "area_outside",
			area:     Rect{Center: Vector2D{X: 100, Y: 100}, Width: 50, Height: 50},
			expected: false,
		},
		{
			name:     "area_touching_edge",
			area:     Rect{Center: Vector2D{X: 75, Y: 0}, Width: 50, Height: 50},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qt.intersects(tt.area)
			if result != tt.expected {
				t.Errorf("intersects() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkCircle_Collides(b *testing.B) {
	circle1 := Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5}
	circle2 := Circle{Center: Vector2D{X: 8, Y: 0}, Radius: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		circle1.Collides(circle2)
	}
}

func BenchmarkCheckCollision(b *testing.B) {
	circle1 := Circle{Center: Vector2D{X: 0, Y: 0}, Radius: 5}
	circle2 := Circle{Center: Vector2D{X: 8, Y: 0}, Radius: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckCollision(circle1, circle2)
	}
}

func BenchmarkQuadTree_Insert(b *testing.B) {
	boundary := Rect{Center: Vector2D{X: 0, Y: 0}, Width: 1000, Height: 1000}
	qt := NewQuadTree(boundary, 4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		point := Vector2D{X: float64(i % 500), Y: float64((i / 500) % 500)}
		qt.Insert(point, i)
	}
}
