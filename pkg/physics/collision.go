// pkg/physics/collision.go
package physics

// Circle represents a circular collision shape
type Circle struct {
	Center Vector2D
	Radius float64
}

// Collides checks if two circles are colliding
func (c Circle) Collides(other Circle) bool {
	return c.Center.Distance(other.Center) < c.Radius+other.Radius
}

// CollisionResult contains information about a collision
type CollisionResult struct {
	Collided     bool
	Normal       Vector2D
	Penetration  float64
	ContactPoint Vector2D
}

// CheckCollision performs detailed collision detection between two circles
func CheckCollision(a, b Circle) CollisionResult {
	// Vector from A to B
	normal := b.Center.Sub(a.Center)
	distance := normal.Length()

	// No collision
	if distance > a.Radius+b.Radius {
		return CollisionResult{Collided: false}
	}

	// Get penetration depth
	penetration := a.Radius + b.Radius - distance

	// Calculate collision normal and contact point
	normal = normal.Normalize()
	contactPoint := a.Center.Add(normal.Scale(a.Radius))

	return CollisionResult{
		Collided:     true,
		Normal:       normal,
		Penetration:  penetration,
		ContactPoint: contactPoint,
	}
}

// QuadTree for spatial partitioning
type QuadTree struct {
	Boundary  Rect
	Capacity  int
	Points    []Vector2D
	Objects   []interface{}
	Divided   bool
	NorthWest *QuadTree
	NorthEast *QuadTree
	SouthWest *QuadTree
	SouthEast *QuadTree
}

// Rect represents a rectangular area
type Rect struct {
	Center Vector2D
	Width  float64
	Height float64
}

func (r Rect) Contains(point Vector2D) bool {
	return point.X >= r.Center.X-r.Width/2 &&
		point.X < r.Center.X+r.Width/2 &&
		point.Y >= r.Center.Y-r.Height/2 &&
		point.Y < r.Center.Y+r.Height/2
}

// NewQuadTree creates a new quad tree with the given boundary and capacity
func NewQuadTree(boundary Rect, capacity int) *QuadTree {
	return &QuadTree{
		Boundary: boundary,
		Capacity: capacity,
		Points:   make([]Vector2D, 0, capacity),
		Objects:  make([]interface{}, 0, capacity),
		Divided:  false,
	}
}

// pkg/physics/collision.go
func (qt *QuadTree) Insert(point Vector2D, object interface{}) bool {
	if !qt.Boundary.Contains(point) {
		return false
	}

	if len(qt.Points) < qt.Capacity && !qt.Divided {
		qt.Points = append(qt.Points, point)
		qt.Objects = append(qt.Objects, object)
		return true
	}

	if !qt.Divided {
		qt.Subdivide()
	}

	return qt.NorthWest.Insert(point, object) ||
		qt.NorthEast.Insert(point, object) ||
		qt.SouthWest.Insert(point, object) ||
		qt.SouthEast.Insert(point, object)
}

// Subdivide splits the quadtree into four quadrants
func (qt *QuadTree) Subdivide() {
	x := qt.Boundary.Center.X
	y := qt.Boundary.Center.Y
	w := qt.Boundary.Width / 2
	h := qt.Boundary.Height / 2

	nw := Rect{Center: Vector2D{X: x - w/2, Y: y + h/2}, Width: w, Height: h}
	ne := Rect{Center: Vector2D{X: x + w/2, Y: y + h/2}, Width: w, Height: h}
	sw := Rect{Center: Vector2D{X: x - w/2, Y: y - h/2}, Width: w, Height: h}
	se := Rect{Center: Vector2D{X: x + w/2, Y: y - h/2}, Width: w, Height: h}

	qt.NorthWest = NewQuadTree(nw, qt.Capacity)
	qt.NorthEast = NewQuadTree(ne, qt.Capacity)
	qt.SouthWest = NewQuadTree(sw, qt.Capacity)
	qt.SouthEast = NewQuadTree(se, qt.Capacity)
	qt.Divided = true
}

// Query returns all objects that could be colliding with the given shape
func (qt *QuadTree) Query(area Rect) []interface{} {
	found := make([]interface{}, 0)

	// If area doesn't intersect boundary, return empty
	if !qt.intersects(area) {
		return found
	}

	// Check objects in this quad
	for i, point := range qt.Points {
		if area.Contains(point) {
			found = append(found, qt.Objects[i])
		}
	}

	// If not divided, we're done
	if !qt.Divided {
		return found
	}

	// Check children
	found = append(found, qt.NorthWest.Query(area)...)
	found = append(found, qt.NorthEast.Query(area)...)
	found = append(found, qt.SouthWest.Query(area)...)
	found = append(found, qt.SouthEast.Query(area)...)

	return found
}

func (qt *QuadTree) intersects(area Rect) bool {
	return !(area.Center.X-area.Width/2 > qt.Boundary.Center.X+qt.Boundary.Width/2 ||
		area.Center.X+area.Width/2 < qt.Boundary.Center.X-qt.Boundary.Width/2 ||
		area.Center.Y-area.Height/2 > qt.Boundary.Center.Y+qt.Boundary.Height/2 ||
		area.Center.Y+area.Height/2 < qt.Boundary.Center.Y-qt.Boundary.Height/2)
}
