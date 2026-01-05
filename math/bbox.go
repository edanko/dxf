package math

import (
	"fmt"
	"math"
)

// BoundingBox represents a 2D axis-aligned bounding box
type BoundingBox struct {
	Min Vec2
	Max Vec2
}

// NewBoundingBox creates a bounding box from two corners
func NewBoundingBox(min, max Vec2) BoundingBox {
	return BoundingBox{
		Min: Vec2{math.Min(min.X(), max.X()), math.Min(min.Y(), max.Y())},
		Max: Vec2{math.Max(min.X(), max.X()), math.Max(min.Y(), max.Y())},
	}
}

// NewBoundingBoxFromPoints creates a bounding box from a set of points
func NewBoundingBoxFromPoints(points []Vec2) BoundingBox {
	if len(points) == 0 {
		return BoundingBox{}
	}

	min := points[0]
	max := points[0]

	for _, point := range points[1:] {
		min = Vec2{math.Min(min.X(), point.X()), math.Min(min.Y(), point.Y())}
		max = Vec2{math.Max(max.X(), point.X()), math.Max(max.Y(), point.Y())}
	}

	return BoundingBox{Min: min, Max: max}
}

// NewBoundingBoxFrom3DPoints creates a 2D bounding box from 3D points (ignoring Z)
func NewBoundingBoxFrom3DPoints(points []Vec3) BoundingBox {
	if len(points) == 0 {
		return BoundingBox{}
	}

	min := points[0].XY()
	max := points[0].XY()

	for _, point := range points[1:] {
		xy := point.XY()
		min = Vec2{math.Min(min.X(), xy.X()), math.Min(min.Y(), xy.Y())}
		max = Vec2{math.Max(max.X(), xy.X()), math.Max(max.Y(), xy.Y())}
	}

	return BoundingBox{Min: min, Max: max}
}

// IsEmpty returns true if bounding box has zero area
func (bb BoundingBox) IsEmpty() bool {
	return bb.Min.X() >= bb.Max.X() || bb.Min.Y() >= bb.Max.Y()
}

// Width returns width of bounding box
func (bb BoundingBox) Width() float64 {
	if bb.IsEmpty() {
		return 0
	}
	return bb.Max.X() - bb.Min.X()
}

// Height returns height of bounding box
func (bb BoundingBox) Height() float64 {
	if bb.IsEmpty() {
		return 0
	}
	return bb.Max.Y() - bb.Min.Y()
}

// Size returns width and height as Vec2
func (bb BoundingBox) Size() Vec2 {
	return Vec2{bb.Width(), bb.Height()}
}

// Area returns area of bounding box
func (bb BoundingBox) Area() float64 {
	return bb.Width() * bb.Height()
}

// Center returns center point of bounding box
func (bb BoundingBox) Center() Vec2 {
	if bb.IsEmpty() {
		return Vec2{}
	}
	return bb.Min.Add(bb.Max).Mul(0.5)
}

// ContainsPoint checks if point is inside bounding box
func (bb BoundingBox) ContainsPoint(point Vec2) bool {
	if bb.IsEmpty() {
		return false
	}

	return point.X() >= bb.Min.X() && point.X() <= bb.Max.X() &&
		point.Y() >= bb.Min.Y() && point.Y() <= bb.Max.Y()
}

// ContainsBoundingBox checks if another bounding box is fully contained
func (bb BoundingBox) ContainsBoundingBox(other BoundingBox) bool {
	if bb.IsEmpty() || other.IsEmpty() {
		return false
	}

	return bb.ContainsPoint(other.Min) && bb.ContainsPoint(other.Max)
}

// IntersectsBoundingBox checks if two bounding boxes intersect
func (bb BoundingBox) IntersectsBoundingBox(other BoundingBox) bool {
	if bb.IsEmpty() || other.IsEmpty() {
		return false
	}

	return !(bb.Max.X() < other.Min.X() || bb.Min.X() > other.Max.X() ||
		bb.Max.Y() < other.Min.Y() || bb.Min.Y() > other.Max.Y())
}

// Union returns the smallest bounding box containing both bounding boxes
func (bb BoundingBox) Union(other BoundingBox) BoundingBox {
	if bb.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return bb
	}

	return BoundingBox{
		Min: Vec2{
			math.Min(bb.Min.X(), other.Min.X()),
			math.Min(bb.Min.Y(), other.Min.Y()),
		},
		Max: Vec2{
			math.Max(bb.Max.X(), other.Max.X()),
			math.Max(bb.Max.Y(), other.Max.Y()),
		},
	}
}

// Intersection returns the intersection of two bounding boxes
func (bb BoundingBox) Intersection(other BoundingBox) BoundingBox {
	if !bb.IntersectsBoundingBox(other) {
		return BoundingBox{} // Empty intersection
	}

	return BoundingBox{
		Min: Vec2{
			math.Max(bb.Min.X(), other.Min.X()),
			math.Max(bb.Min.Y(), other.Min.Y()),
		},
		Max: Vec2{
			math.Min(bb.Max.X(), other.Max.X()),
			math.Min(bb.Max.Y(), other.Max.Y()),
		},
	}
}

// Expand expands bounding box by a margin in all directions
func (bb BoundingBox) Expand(margin float64) BoundingBox {
	if bb.IsEmpty() {
		return bb
	}

	return BoundingBox{
		Min: bb.Min.Sub(Vec2{margin, margin}),
		Max: bb.Max.Add(Vec2{margin, margin}),
	}
}

// ExpandToPoint expands bounding box to include a point
func (bb BoundingBox) ExpandToPoint(point Vec2) BoundingBox {
	if bb.IsEmpty() {
		return BoundingBox{Min: point, Max: point}
	}

	return BoundingBox{
		Min: Vec2{
			math.Min(bb.Min.X(), point.X()),
			math.Min(bb.Min.Y(), point.Y()),
		},
		Max: Vec2{
			math.Max(bb.Max.X(), point.X()),
			math.Max(bb.Max.Y(), point.Y()),
		},
	}
}

// ClosestPoint returns the closest point on the bounding box to a given point
func (bb BoundingBox) ClosestPoint(point Vec2) Vec2 {
	if bb.IsEmpty() {
		return point
	}

	closest := point

	// Clamp X coordinate
	if point.X() < bb.Min.X() {
		closest = Vec2{bb.Min.X(), closest.Y()}
	} else if point.X() > bb.Max.X() {
		closest = Vec2{bb.Max.X(), closest.Y()}
	}

	// Clamp Y coordinate
	if point.Y() < bb.Min.Y() {
		closest = Vec2{closest.X(), bb.Min.Y()}
	} else if point.Y() > bb.Max.Y() {
		closest = Vec2{closest.X(), bb.Max.Y()}
	}

	return closest
}

// DistanceToPoint returns the minimum distance from bounding box to a point
func (bb BoundingBox) DistanceToPoint(point Vec2) float64 {
	if bb.ContainsPoint(point) {
		return 0
	}

	closest := bb.ClosestPoint(point)
	return point.Distance(closest)
}

// Translate moves bounding box by a vector
func (bb BoundingBox) Translate(vec Vec2) BoundingBox {
	if bb.IsEmpty() {
		return bb
	}

	return BoundingBox{
		Min: bb.Min.Add(vec),
		Max: bb.Max.Add(vec),
	}
}

// Scale scales bounding box around its center
func (bb BoundingBox) Scale(scale Vec2) BoundingBox {
	if bb.IsEmpty() {
		return bb
	}

	center := bb.Center()
	halfSize := bb.Size().Mul(0.5)

	// Apply scale to each dimension separately
	scaledHalfX := halfSize.X() * scale.X()
	scaledHalfY := halfSize.Y() * scale.Y()

	return BoundingBox{
		Min: Vec2{center.X() - scaledHalfX, center.Y() - scaledHalfY},
		Max: Vec2{center.X() + scaledHalfX, center.Y() + scaledHalfY},
	}
}

// Corners returns the four corners of the bounding box
func (bb BoundingBox) Corners() []Vec2 {
	if bb.IsEmpty() {
		return []Vec2{}
	}

	return []Vec2{
		bb.Min,
		Vec2{bb.Max.X(), bb.Min.Y()},
		bb.Max,
		Vec2{bb.Min.X(), bb.Max.Y()},
	}
}

// Perimeter returns the perimeter of the bounding box
func (bb BoundingBox) Perimeter() float64 {
	return 2.0 * (bb.Width() + bb.Height())
}

// AspectRatio returns the width/height aspect ratio
func (bb BoundingBox) AspectRatio() float64 {
	if bb.Height() == 0 {
		return 0
	}
	return bb.Width() / bb.Height()
}

// IsSquare returns true if width equals height (within tolerance)
func (bb BoundingBox) IsSquare(tolerance float64) bool {
	if bb.IsEmpty() {
		return false
	}

	return math.Abs(bb.Width()-bb.Height()) <= tolerance
}

// LongestSide returns the longer dimension
func (bb BoundingBox) LongestSide() float64 {
	return math.Max(bb.Width(), bb.Height())
}

// ShortestSide returns the shorter dimension
func (bb BoundingBox) ShortestSide() float64 {
	if bb.IsEmpty() {
		return 0
	}
	return math.Min(bb.Width(), bb.Height())
}

// String returns string representation
func (bb BoundingBox) String() string {
	return fmt.Sprintf("BoundingBox{Min: %v, Max: %v}", bb.Min, bb.Max)
}

// BoundingBox3D represents a 3D axis-aligned bounding box
type BoundingBox3D struct {
	Min Vec3
	Max Vec3
}

// NewBoundingBox3D creates a 3D bounding box from two corners
func NewBoundingBox3D(min, max Vec3) BoundingBox3D {
	return BoundingBox3D{
		Min: Vec3{
			math.Min(min.X(), max.X()),
			math.Min(min.Y(), max.Y()),
			math.Min(min.Z(), max.Z()),
		},
		Max: Vec3{
			math.Max(min.X(), max.X()),
			math.Max(min.Y(), max.Y()),
			math.Max(min.Z(), max.Z()),
		},
	}
}

// NewBoundingBox3DFromPoints creates a 3D bounding box from points
func NewBoundingBox3DFromPoints(points []Vec3) BoundingBox3D {
	if len(points) == 0 {
		return BoundingBox3D{}
	}

	min := points[0]
	max := points[0]

	for _, point := range points[1:] {
		min = Vec3{
			math.Min(min.X(), point.X()),
			math.Min(min.Y(), point.Y()),
			math.Min(min.Z(), point.Z()),
		}
		max = Vec3{
			math.Max(max.X(), point.X()),
			math.Max(max.Y(), point.Y()),
			math.Max(max.Z(), point.Z()),
		}
	}

	return BoundingBox3D{Min: min, Max: max}
}

// Width returns width of 3D bounding box
func (bb BoundingBox3D) Width() float64 {
	return bb.Max.X() - bb.Min.X()
}

// Height returns height of 3D bounding box
func (bb BoundingBox3D) Height() float64 {
	return bb.Max.Y() - bb.Min.Y()
}

// Depth returns depth of 3D bounding box
func (bb BoundingBox3D) Depth() float64 {
	return bb.Max.Z() - bb.Min.Z()
}

// Size returns width, height, depth as Vec3
func (bb BoundingBox3D) Size() Vec3 {
	return Vec3{bb.Width(), bb.Height(), bb.Depth()}
}

// Center returns center point of 3D bounding box
func (bb BoundingBox3D) Center() Vec3 {
	return bb.Min.Add(bb.Max).Mul(0.5)
}

// Volume returns volume of 3D bounding box
func (bb BoundingBox3D) Volume() float64 {
	return bb.Width() * bb.Height() * bb.Depth()
}

// ContainsPoint3D checks if 3D point is inside bounding box
func (bb BoundingBox3D) ContainsPoint3D(point Vec3) bool {
	return point.X() >= bb.Min.X() && point.X() <= bb.Max.X() &&
		point.Y() >= bb.Min.Y() && point.Y() <= bb.Max.Y() &&
		point.Z() >= bb.Min.Z() && point.Z() <= bb.Max.Z()
}

// String returns string representation of 3D bounding box
func (bb BoundingBox3D) String() string {
	return fmt.Sprintf("BoundingBox3D{Min: %v, Max: %v}", bb.Min, bb.Max)
}
