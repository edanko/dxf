package math

import (
	"math"
	"sort"
)

// DelaunayTriangulation implements Delaunay triangulation for 2D point sets
type DelaunayTriangulation struct {
	points    []Point2D
	triangles []*Triangle
	edges     map[[2]int]bool
}

// Triangle represents a triangle in Delaunay triangulation
type Triangle struct {
	Vertices     [3]int // Indices into points array
	Circumcenter Point2D
	Circumradius float64
}

// NewDelaunayTriangulation creates a new Delaunay triangulation
func NewDelaunayTriangulation(points []Point2D) *DelaunayTriangulation {
	if len(points) < 3 {
		return &DelaunayTriangulation{
			points: points,
		}
	}

	dt := &DelaunayTriangulation{
		points:    points,
		triangles: make([]*Triangle, 0),
		edges:     make(map[[2]int]bool),
	}

	dt.triangulate()
	return dt
}

// triangulate performs the Delaunay triangulation using Bowyer-Watson algorithm
func (dt *DelaunayTriangulation) triangulate() {
	// Create super-triangle that contains all points
	superTriangle := dt.createSuperTriangle()
	dt.triangles = append(dt.triangles, superTriangle)

	// Incrementally insert points
	for i, point := range dt.points {
		dt.insertPoint(i, point)
	}

	// Remove triangles that share vertices with super-triangle
	dt.removeSuperTriangle(superTriangle)
}

// createSuperTriangle creates a large triangle that contains all points
func (dt *DelaunayTriangulation) createSuperTriangle() *Triangle {
	if len(dt.points) == 0 {
		return nil
	}

	// Find bounding box
	minX, minY := dt.points[0].X, dt.points[0].Y
	maxX, maxY := dt.points[0].X, dt.points[0].Y

	for _, p := range dt.points {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	// Expand bounding box
	dx := maxX - minX
	dy := maxY - minY
	maxDelta := math.Max(dx, dy)
	midX := (minX + maxX) / 2.0
	midY := (minY + maxY) / 2.0

	// Create super-triangle vertices
	superPoints := []Point2D{
		{X: midX - 20*maxDelta, Y: midY - maxDelta},
		{X: midX, Y: midY + 20*maxDelta},
		{X: midX + 20*maxDelta, Y: midY - maxDelta},
	}

	// Add super-triangle points to the point list
	originalPointCount := len(dt.points)
	dt.points = append(dt.points, superPoints...)

	return &Triangle{
		Vertices:     [3]int{originalPointCount, originalPointCount + 1, originalPointCount + 2},
		Circumcenter: Point2D{X: midX, Y: midY},
		Circumradius: 20 * maxDelta,
	}
}

// insertPoint inserts a point into the triangulation
func (dt *DelaunayTriangulation) insertPoint(pointIndex int, point Point2D) {
	var badTriangles []*Triangle

	// Find triangles whose circumcircles contain the point
	for _, tri := range dt.triangles {
		if dt.pointInCircumcircle(point, tri) {
			badTriangles = append(badTriangles, tri)
		}
	}

	// Find boundary of polygonal hole
	var boundary [][2]int
	for _, tri := range badTriangles {
		edges := [][2]int{
			{tri.Vertices[0], tri.Vertices[1]},
			{tri.Vertices[1], tri.Vertices[2]},
			{tri.Vertices[2], tri.Vertices[0]},
		}

		for _, edge := range edges {
			edgeKey := dt.edgeKey(edge[0], edge[1])
			if dt.edges[edgeKey] {
				// Edge is shared, remove it
				delete(dt.edges, edgeKey)
			} else {
				// Edge is unique boundary edge
				dt.edges[edgeKey] = true
				boundary = append(boundary, edge)
			}
		}
	}

	// Remove bad triangles
	var newTriangles []*Triangle
	for _, tri := range dt.triangles {
		isBad := false
		for _, badTri := range badTriangles {
			if dt.trianglesEqual(tri, badTri) {
				isBad = true
				break
			}
		}
		if !isBad {
			newTriangles = append(newTriangles, tri)
		}
	}
	dt.triangles = newTriangles

	// Create new triangles from boundary edges and new point
	for _, edge := range boundary {
		newTri := dt.createTriangle(edge[0], edge[1], pointIndex)
		dt.triangles = append(dt.triangles, newTri)
	}

	// Clear edges for next iteration
	dt.edges = make(map[[2]int]bool)
}

// removeSuperTriangle removes triangles that use super-triangle vertices
func (dt *DelaunayTriangulation) removeSuperTriangle(superTriangle *Triangle) {
	var newTriangles []*Triangle
	originalPointCount := len(dt.points) - 3

	for _, tri := range dt.triangles {
		allValid := true
		for _, vertex := range tri.Vertices {
			if vertex >= originalPointCount {
				allValid = false
				break
			}
		}
		if allValid {
			newTriangles = append(newTriangles, tri)
		}
	}
	dt.triangles = newTriangles
}

// createTriangle creates a new triangle and calculates its circumcircle
func (dt *DelaunayTriangulation) createTriangle(i, j, k int) *Triangle {
	p1 := dt.points[i]
	p2 := dt.points[j]
	p3 := dt.points[k]

	// Calculate circumcenter using perpendicular bisectors
	ax := p1.X
	ay := p1.Y
	bx := p2.X
	by := p2.Y
	cx := p3.X
	cy := p3.Y

	d := 2.0 * (ax*(by-cy) + bx*(cy-ay) + cx*(ay-by))
	if math.Abs(d) < 1e-12 {
		// Points are collinear, return degenerate triangle
		return &Triangle{
			Vertices:     [3]int{i, j, k},
			Circumcenter: Point2D{X: 0, Y: 0},
			Circumradius: math.MaxFloat64,
		}
	}

	ux := ((ax*ax+ay*ay)*(by-cy) + (bx*bx+by*by)*(cy-ay) + (cx*cx+cy*cy)*(ay-by)) / d
	uy := ((ax*ax+ay*ay)*(cx-bx) + (bx*bx+by*by)*(ax-cx) + (cx*cx+cy*cy)*(bx-ax)) / d

	circumcenter := Point2D{X: ux, Y: uy}
	circumradius := math.Sqrt((ax-ux)*(ax-ux) + (ay-uy)*(ay-uy))

	return &Triangle{
		Vertices:     [3]int{i, j, k},
		Circumcenter: circumcenter,
		Circumradius: circumradius,
	}
}

// pointInCircumcircle checks if a point is inside a triangle's circumcircle
func (dt *DelaunayTriangulation) pointInCircumcircle(point Point2D, triangle *Triangle) bool {
	dx := point.X - triangle.Circumcenter.X
	dy := point.Y - triangle.Circumcenter.Y
	distanceSquared := dx*dx + dy*dy
	return distanceSquared < triangle.Circumradius*triangle.Circumradius-1e-12
}

// edgeKey creates a canonical key for an edge
func (dt *DelaunayTriangulation) edgeKey(i, j int) [2]int {
	if i < j {
		return [2]int{i, j}
	}
	return [2]int{j, i}
}

// trianglesEqual checks if two triangles are equal
func (dt *DelaunayTriangulation) trianglesEqual(a, b *Triangle) bool {
	if len(a.Vertices) != len(b.Vertices) {
		return false
	}

	aVerts := make([]int, len(a.Vertices))
	copy(aVerts, a.Vertices[:])
	sort.Ints(aVerts)

	bVerts := make([]int, len(b.Vertices))
	copy(bVerts, b.Vertices[:])
	sort.Ints(bVerts)

	for i := range aVerts {
		if aVerts[i] != bVerts[i] {
			return false
		}
	}
	return true
}

// GetTriangles returns the triangulation result
func (dt *DelaunayTriangulation) GetTriangles() []*Triangle {
	return dt.triangles
}

// GetTriangleIndices returns triangle vertex indices
func (dt *DelaunayTriangulation) GetTriangleIndices() [][3]int {
	indices := make([][3]int, len(dt.triangles))
	for i, tri := range dt.triangles {
		indices[i] = tri.Vertices
	}
	return indices
}

// GetTrianglePoints returns triangle vertices as points
func (dt *DelaunayTriangulation) GetTrianglePoints() [][]Point2D {
	triangles := make([][]Point2D, len(dt.triangles))
	for i, tri := range dt.triangles {
		triangles[i] = []Point2D{
			dt.points[tri.Vertices[0]],
			dt.points[tri.Vertices[1]],
			dt.points[tri.Vertices[2]],
		}
	}
	return triangles
}

// Validate checks if the triangulation is valid
func (dt *DelaunayTriangulation) Validate() bool {
	// Check that all triangles have valid circumcircles
	for _, tri := range dt.triangles {
		if tri.Circumradius <= 0 || math.IsNaN(tri.Circumradius) || math.IsInf(tri.Circumradius, 0) {
			return false
		}
	}

	// Check Delaunay condition: no point should be inside any triangle's circumcircle
	for i, tri := range dt.triangles {
		for j, point := range dt.points {
			if i != j {
				if dt.pointInCircumcircle(point, tri) {
					// Check if point is one of the triangle's vertices
					isVertex := false
					for _, vertex := range tri.Vertices {
						if vertex == j {
							isVertex = true
							break
						}
					}
					if !isVertex {
						return false
					}
				}
			}
		}
	}

	return true
}
