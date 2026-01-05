package math

import (
	"math"
)

// Construct2D provides 2D geometric construction utilities
type Construct2D struct{}

// NewConstruct2D creates a new 2D construction helper
func NewConstruct2D() *Construct2D {
	return &Construct2D{}
}

// CircleFrom3Points creates a circle from three points
func (c *Construct2D) CircleFrom3Points(p1, p2, p3 Vec2) (center Vec2, radius float64, ok bool) {
	// Check if points are collinear
	area := (p2.X()-p1.X())*(p3.Y()-p1.Y()) - (p2.Y()-p1.Y())*(p3.X()-p1.X())
	if math.Abs(area) < 1e-10 {
		return Vec2{}, 0, false
	}

	// Calculate perpendicular bisectors
	mid1 := p1.Lerp(p2, 0.5)
	mid2 := p2.Lerp(p3, 0.5)

	dx1 := p2.X() - p1.X()
	dy1 := p2.Y() - p1.Y()
	dx2 := p3.X() - p2.X()
	dy2 := p3.Y() - p2.Y()

	// Solve for intersection of perpendicular bisectors
	perp1 := Vec2{-dy1, dx1}
	perp2 := Vec2{-dy2, dx2}

	center = c.LineIntersection(mid1, mid1.Add(perp1), mid2, mid2.Add(perp2))
	radius = center.Distance(p1)

	return center, radius, true
}

// LineIntersection finds intersection of two lines defined by two points each
func (c *Construct2D) LineIntersection(p1, p2, p3, p4 Vec2) Vec2 {
	x1, y1 := p1.X(), p1.Y()
	x2, y2 := p2.X(), p2.Y()
	x3, y3 := p3.X(), p3.Y()
	x4, y4 := p4.X(), p4.Y()

	den := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
	if math.Abs(den) < 1e-10 {
		return Vec2{} // Parallel lines
	}

	t := ((x1-x3)*(y3-y4) - (y1-y3)*(x3-x4)) / den
	return Vec2{
		x1 + t*(x2-x1),
		y1 + t*(y2-y1),
	}
}

// PointOnLine finds the closest point on a line segment to a given point
func (c *Construct2D) PointOnLine(point, lineStart, lineEnd Vec2) Vec2 {
	lineVec := lineEnd.Sub(lineStart)
	t := point.Sub(lineStart).Dot(lineVec) / lineVec.LengthSquared()
	t = math.Max(0, math.Min(1, t)) // Clamp to segment
	return lineStart.Add(lineVec.Mul(t))
}

// PointToLineDistance calculates distance from point to line segment
func (c *Construct2D) PointToLineDistance(point, lineStart, lineEnd Vec2) float64 {
	closest := c.PointOnLine(point, lineStart, lineEnd)
	return point.Distance(closest)
}

// IsPointInPolygon checks if a point is inside a polygon using ray casting
func (c *Construct2D) IsPointInPolygon(point Vec2, polygon []Vec2) bool {
	if len(polygon) < 3 {
		return false
	}

	x, y := point.X(), point.Y()
	n := len(polygon)
	inside := false

	for i, j := 0, n-1; i < n; j, i = i+1, i+1 {
		xi, yi := polygon[i].X(), polygon[i].Y()
		xj, yj := polygon[j].X(), polygon[j].Y()

		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
	}

	return inside
}

// BoundingBox calculates bounding box of points
func (c *Construct2D) BoundingBox(points []Vec2) (min, max Vec2) {
	if len(points) == 0 {
		return Vec2{}, Vec2{}
	}

	min = points[0]
	max = points[0]

	for _, p := range points[1:] {
		if p.X() < min.X() {
			min[0] = p[0]
		}
		if p.Y() < min.Y() {
			min[1] = p[1]
		}
		if p.X() > max.X() {
			max[0] = p[0]
		}
		if p.Y() > max.Y() {
			max[1] = p[1]
		}
	}

	return min, max
}

// PolygonArea calculates signed area of a polygon (positive = counter-clockwise)
func (c *Construct2D) PolygonArea(polygon []Vec2) float64 {
	n := len(polygon)
	if n < 3 {
		return 0
	}

	area := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += polygon[i].X()*polygon[j].Y() - polygon[j].X()*polygon[i].Y()
	}

	return area / 2.0
}

// PolygonCentroid calculates centroid of a polygon
func (c *Construct2D) PolygonCentroid(polygon []Vec2) Vec2 {
	n := len(polygon)
	if n == 0 {
		return Vec2{}
	}
	if n == 1 {
		return polygon[0]
	}

	area := c.PolygonArea(polygon)
	if math.Abs(area) < 1e-10 {
		// Degenerate polygon, use average of vertices
		centroid := Vec2{}
		for _, p := range polygon {
			centroid = centroid.Add(p)
		}
		return centroid.Div(float64(n))
	}

	cx, cy := 0.0, 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		factor := polygon[i].X()*polygon[j].Y() - polygon[j].X()*polygon[i].Y()
		cx += (polygon[i].X() + polygon[j].X()) * factor
		cy += (polygon[i].Y() + polygon[j].Y()) * factor
	}

	area6 := area * 6.0
	return Vec2{cx / area6, cy / area6}
}

// ConvexHull computes convex hull using Graham scan algorithm
func (c *Construct2D) ConvexHull(points []Vec2) []Vec2 {
	if len(points) <= 1 {
		return points
	}

	// Find the point with lowest Y (and X if tie)
	start := 0
	for i, p := range points {
		if p.Y() < points[start].Y() || (p.Y() == points[start].Y() && p.X() < points[start].X()) {
			start = i
		}
	}

	// Sort points by polar angle relative to start
	sorted := make([]Vec2, len(points)-1)
	copyIdx := 0
	for i, p := range points {
		if i != start {
			sorted[copyIdx] = p
			copyIdx++
		}
	}

	// Simple insertion sort by angle
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0; j-- {
			angle1 := c.polarAngle(points[start], sorted[j])
			angle2 := c.polarAngle(points[start], sorted[j-1])
			if angle1 < angle2 {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			} else {
				break
			}
		}
	}

	// Build hull
	hull := []Vec2{points[start]}
	for _, p := range sorted {
		for len(hull) > 1 && c.ccw(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}

	return hull
}

// polarAngle calculates polar angle between two points
func (c *Construct2D) polarAngle(from, to Vec2) float64 {
	return math.Atan2(to.Y()-from.Y(), to.X()-from.X())
}

// ccw checks if three points make a counter-clockwise turn
func (c *Construct2D) ccw(a, b, pt Vec2) float64 {
	return (b.X()-a.X())*(pt.Y()-a.Y()) - (b.Y()-a.Y())*(pt.X()-a.X())
}

// DouglasPeucker simplifies a polyline using Douglas-Peucker algorithm
func (c *Construct2D) DouglasPeucker(points []Vec2, epsilon float64) []Vec2 {
	if len(points) <= 2 {
		return points
	}

	// Find the point with maximum distance
	maxDist := 0.0
	maxIndex := 0

	for i := 1; i < len(points)-1; i++ {
		dist := c.PointToLineDistance(points[i], points[0], points[len(points)-1])
		if dist > maxDist {
			maxDist = dist
			maxIndex = i
		}
	}

	// If max distance is greater than epsilon, recursively simplify
	if maxDist > epsilon {
		left := c.DouglasPeucker(points[:maxIndex+1], epsilon)
		right := c.DouglasPeucker(points[maxIndex:], epsilon)

		// Combine results (avoiding duplicate middle point)
		result := append(left[:len(left)-1], right...)
		return result
	}

	// Otherwise, return the first and last points
	return []Vec2{points[0], points[len(points)-1]}
}
