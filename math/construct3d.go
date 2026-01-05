package math

import (
	"math"
)

// Construct3D provides 3D geometric construction utilities
type Construct3D struct{}

// NewConstruct3D creates a new 3D construction helper
func NewConstruct3D() *Construct3D {
	return &Construct3D{}
}

// Plane represents a plane in 3D space
type Plane struct {
	Normal Vec3
	Point  Vec3
}

// NewPlane creates a plane from a point and normal
func NewPlane(point, normal Vec3) Plane {
	return Plane{
		Normal: normal.Normalized(),
		Point:  point,
	}
}

// NewPlaneFrom3Points creates a plane from three points
func NewPlaneFrom3Points(p1, p2, p3 Vec3) Plane {
	normal := p2.Sub(p1).Cross(p3.Sub(p1))
	if normal.LengthSquared() < 1e-10 {
		return Plane{Vec3{}, p1} // Degenerate plane
	}
	return NewPlane(p1, normal)
}

// DistanceToPoint calculates distance from plane to a point
func (p Plane) DistanceToPoint(point Vec3) float64 {
	return math.Abs(p.Normal.Dot(point.Sub(p.Point)))
}

// ProjectPoint projects a point onto plane
func (p Plane) ProjectPoint(point Vec3) Vec3 {
	return point.Sub(p.Normal.Mul(p.Normal.Dot(point.Sub(p.Point))))
}

// CircleFrom3Points creates a circle from three points in 3D
func (c *Construct3D) CircleFrom3Points(p1, p2, p3 Vec3) (center Vec3, radius, radiusSquared float64, ok bool) {
	// Create plane from three points
	plane := NewPlaneFrom3Points(p1, p2, p3)

	// Check if points are collinear
	if plane.Normal.LengthSquared() < 1e-10 {
		return Vec3{}, 0, 0, false
	}

	// Transform to 2D by projecting onto plane
	c2d := NewConstruct2D()
	proj1 := plane.ProjectPoint(p1).XY()
	proj2 := plane.ProjectPoint(p2).XY()
	proj3 := plane.ProjectPoint(p3).XY()

	// Find 2D circle center
	center2D, radius2D, ok2D := c2d.CircleFrom3Points(proj1, proj2, proj3)
	if !ok2D {
		return Vec3{}, 0, 0, false
	}

	// Transform back to 3D
	center = plane.Point.Add(NewVec3FromVec2(center2D))
	radius = radius2D
	radiusSquared = radius * radius

	return center, radius, radiusSquared, true
}

// LineLineIntersection finds closest points between two 3D lines
func (c *Construct3D) LineLineIntersection(p1, d1, p2, d2 Vec3) (point1, point2 Vec3, parallel bool) {
	// Vector between points on two lines
	w := p1.Sub(p2)

	// Check if lines are parallel
	a := d1.Dot(d1)
	b := d1.Dot(d2)
	cc := d2.Dot(d2)
	d := d1.Dot(w)
	e := d2.Dot(w)

	den := a*cc - b*b
	if math.Abs(den) < 1e-10 {
		parallel = true
		return p1, p2, true
	}

	// Find parameters for closest points
	s := (b*e - cc*d) / den
	t := (a*e - b*d) / den

	point1 = p1.Add(d1.Mul(s))
	point2 = p2.Add(d2.Mul(t))

	return point1, point2, false
}

// PointToLineDistance calculates distance from point to line in 3D
func (c *Construct3D) PointToLineDistance(point, linePoint, lineDir Vec3) float64 {
	// Project point-line vector onto line direction
	pointToLine := point.Sub(linePoint)
	projLength := pointToLine.Dot(lineDir) / lineDir.LengthSquared()
	closest := linePoint.Add(lineDir.Mul(projLength))

	return point.Distance(closest)
}

// PointToSegmentDistance calculates distance from point to line segment in 3D
func (c *Construct3D) PointToSegmentDistance(point, segStart, segEnd Vec3) float64 {
	segVec := segEnd.Sub(segStart)
	pointToStart := point.Sub(segStart)

	// Parameter along segment
	t := pointToStart.Dot(segVec) / segVec.LengthSquared()
	t = math.Max(0, math.Min(1, t)) // Clamp to segment

	closest := segStart.Add(segVec.Mul(t))
	return point.Distance(closest)
}

// TriangleArea calculates area of a triangle in 3D
func (c *Construct3D) TriangleArea(a, b, triC Vec3) float64 {
	return b.Sub(a).Cross(triC.Sub(a)).Length() / 2.0
}

// TriangleNormal calculates normal of a triangle (normalized)
func (c *Construct3D) TriangleNormal(a, b, triC Vec3) Vec3 {
	return b.Sub(a).Cross(triC.Sub(a)).Normalized()
}

// BoundingBox calculates bounding box of 3D points
func (c *Construct3D) BoundingBox(points []Vec3) (min, max Vec3) {
	if len(points) == 0 {
		return Vec3{}, Vec3{}
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
		if p.Z() < min.Z() {
			min[2] = p[2]
		}
		if p.X() > max.X() {
			max[0] = p[0]
		}
		if p.Y() > max.Y() {
			max[1] = p[1]
		}
		if p.Z() > max.Z() {
			max[2] = p[2]
		}
	}

	return min, max
}

// IsPointInSphere checks if a point is inside a sphere
func (c *Construct3D) IsPointInSphere(point, center Vec3, radius float64) bool {
	return point.DistanceSquared(center) <= radius*radius
}

// IsPointInBox checks if a point is inside an axis-aligned bounding box
func (c *Construct3D) IsPointInBox(point, min, max Vec3) bool {
	return point.X() >= min.X() && point.X() <= max.X() &&
		point.Y() >= min.Y() && point.Y() <= max.Y() &&
		point.Z() >= min.Z() && point.Z() <= max.Z()
}

// SphereFrom4Points creates a sphere from four non-coplanar points
func (c *Construct3D) SphereFrom4Points(p1, p2, p3, p4 Vec3) (center Vec3, radius float64, ok bool) {
	// Check if points are coplanar
	plane1 := NewPlaneFrom3Points(p1, p2, p3)
	if math.Abs(plane1.DistanceToPoint(p4)) < 1e-10 {
		return Vec3{}, 0, false // Coplanar points
	}

	// Create matrices for sphere equation solver
	// This is a simplified approach - production code would use more robust methods

	// Edge directions
	d1 := p2.Sub(p1)
	d2 := p3.Sub(p1)
	d3 := p4.Sub(p1)

	// Create system of linear equations to find sphere center
	// This is a simplified approach
	a := 2.0 * d1.X()
	b := 2.0 * d1.Y()
	cc := 2.0 * d1.Z()

	d := 2.0 * d2.X()
	e := 2.0 * d2.Y()
	f := 2.0 * d2.Z()

	g := 2.0 * d3.X()
	h := 2.0 * d3.Y()
	i := 2.0 * d3.Z()

	j := p2.LengthSquared() - p1.LengthSquared()
	k := p3.LengthSquared() - p1.LengthSquared()
	l := p4.LengthSquared() - p1.LengthSquared()

	// Solve 3x3 system using Cramer's rule
	det := a*(e*i-f*h) - b*(d*i-f*g) + cc*(d*h-e*g)

	if math.Abs(det) < 1e-10 {
		return Vec3{}, 0, false
	}

	invDet := 1.0 / det

	centerX := invDet * (j*(e*i-f*h) - b*(k*i-f*l) + cc*(k*h-e*l))
	centerY := invDet * (a*(k*i-f*l) - j*(d*i-f*g) + cc*(d*l-k*g))
	centerZ := invDet * (a*(e*l-k*h) - b*(d*l-k*g) + j*(d*h-e*g))

	center = Vec3{centerX, centerY, centerZ}
	radius = center.Distance(p1)

	return center, radius, true
}

// BarycentricCoords calculates barycentric coordinates of point in triangle
func (c *Construct3D) BarycentricCoords(point, a, b, triC Vec3) (u, v, w float64) {
	v0 := triC.Sub(a)
	v1 := b.Sub(a)
	v2 := point.Sub(a)

	d00 := v0.Dot(v0)
	d01 := v0.Dot(v1)
	d11 := v1.Dot(v1)
	d20 := v2.Dot(v0)
	d21 := v2.Dot(v1)

	den := d00*d11 - d01*d01
	if math.Abs(den) < 1e-10 {
		return 0, 0, 0 // Degenerate triangle
	}

	v = (d11*d20 - d01*d21) / den
	w = (d00*d21 - d01*d20) / den
	u = 1.0 - v - w

	return u, v, w
}

// RayIntersectsTriangle checks if a ray intersects a triangle
func (c *Construct3D) RayIntersectsTriangle(rayOrigin, rayDir, v0, v1, v2 Vec3) (intersection Vec3, intersected bool, t float64) {
	// Moller-Trumbore algorithm
	edge1 := v1.Sub(v0)
	edge2 := v2.Sub(v0)
	h := rayDir.Cross(edge2)
	a := edge1.Dot(h)

	if a > -1e-10 && a < 1e-10 {
		return Vec3{}, false, 0 // Ray is parallel to triangle
	}

	f := 1.0 / a
	s := rayOrigin.Sub(v0)
	u := f * s.Dot(h)

	if u < 0.0 || u > 1.0 {
		return Vec3{}, false, 0
	}

	q := s.Cross(edge1)
	v := f * rayDir.Dot(q)

	if v < 0.0 || u+v > 1.0 {
		return Vec3{}, false, 0
	}

	t = f * edge2.Dot(q)
	if t > 1e-10 {
		intersection = rayOrigin.Add(rayDir.Mul(t))
		return intersection, true, t
	}

	return Vec3{}, false, 0
}
