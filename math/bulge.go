package math

import (
	"fmt"
	"math"
)

// Bulge represents a bulge value used in LWPOLYLINE arc segments
// A bulge of 0 represents a straight line segment
// The arc is defined by: tan(angle/4) where angle is the included angle
type Bulge float64

// Constants for bulge calculations
const (
	// Angle tolerance for bulge calculations
	BulgeAngleTolerance = 1e-12
	// Maximum arc segments for bulge approximation
	DefaultArcSegments = 16
)

// NewBulgeFromAngle creates a bulge from an angle in radians
func NewBulgeFromAngle(angle float64) Bulge {
	return Bulge(math.Tan(angle / 4.0))
}

// NewBulgeFromRadiusAndChord creates a bulge from radius and chord length
func NewBulgeFromRadiusAndChord(radius, chord float64) Bulge {
	if radius == 0 || chord == 0 {
		return 0
	}

	// Calculate central angle using chord length formula
	sinHalfAngle := chord / (2.0 * radius)

	// Check if chord is too long
	if math.Abs(sinHalfAngle) > 1.0 {
		return 0 // Return straight line if impossible
	}

	halfAngle := math.Asin(sinHalfAngle)
	return NewBulgeFromAngle(2.0 * halfAngle)
}

// Angle returns the included angle in radians
func (b Bulge) Angle() float64 {
	return 4.0 * math.Atan(float64(b))
}

// IsArc returns true if bulge represents an arc (non-zero)
func (b Bulge) IsArc() bool {
	return math.Abs(float64(b)) > BulgeAngleTolerance
}

// IsClockwise returns true if the arc goes clockwise
func (b Bulge) IsClockwise() bool {
	return float64(b) < 0
}

// Radius returns the radius of the arc defined by bulge and chord
func (b Bulge) Radius(chord float64) float64 {
	if !b.IsArc() || chord == 0 {
		return math.Inf(1)
	}

	angle := b.Angle()
	sinHalfAngle := math.Sin(angle / 2.0)
	if math.Abs(sinHalfAngle) < BulgeAngleTolerance {
		return math.Inf(1)
	}

	return math.Abs(chord) / (2.0 * sinHalfAngle)
}

// Sagitta returns the sagitta (height) of the arc
func (b Bulge) Sagitta(chord float64) float64 {
	if !b.IsArc() {
		return 0
	}

	radius := b.Radius(chord)
	if math.IsInf(radius, 0) {
		return 0
	}

	angle := b.Angle()
	return radius * (1.0 - math.Cos(angle/2.0))
}

// Center returns the center point of the arc defined by start, end, and bulge
func (b Bulge) Center(start, end Vec2) Vec2 {
	if !b.IsArc() {
		return start.Add(end).Mul(0.5) // Return midpoint for straight line
	}

	midpoint := start.Add(end).Mul(0.5)
	radius := b.Radius(start.Distance(end))

	if math.IsInf(radius, 0) {
		return midpoint
	}

	// Direction from midpoint to center
	dir := end.Sub(start).Perpendicular().Normalized()
	if !b.IsClockwise() {
		dir = dir.Mul(-1)
	}

	// Distance from midpoint to center
	chordSq := start.Distance(end) * start.Distance(end)
	dist := math.Sqrt(radius*radius - chordSq/4.0)

	return midpoint.Add(dir.Mul(dist))
}

// ApproximateArc approximates the arc with line segments
func (b Bulge) ApproximateArc(start, end Vec2, segments int) []Vec2 {
	if segments < 2 {
		segments = DefaultArcSegments
	}

	if !b.IsArc() {
		return []Vec2{start, end}
	}

	center := b.Center(start, end)
	radius := b.Radius(start.Distance(end))

	if math.IsInf(radius, 0) {
		return []Vec2{start, end}
	}

	// Calculate start and end angles
	startAngle := math.Atan2(start.Y()-center.Y(), start.X()-center.X())
	endAngle := math.Atan2(end.Y()-center.Y(), end.X()-center.X())

	// Adjust for direction
	if b.IsClockwise() {
		if endAngle > startAngle {
			endAngle -= 2.0 * math.Pi
		}
	} else {
		if endAngle < startAngle {
			endAngle += 2.0 * math.Pi
		}
	}

	// Generate points
	points := make([]Vec2, segments+1)
	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		angle := startAngle + t*(endAngle-startAngle)

		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		points[i] = Vec2{x, y}
	}

	return points
}

// Length returns the actual length of the arc segment
func (b Bulge) Length(chord float64) float64 {
	if !b.IsArc() {
		return chord
	}

	radius := b.Radius(chord)
	angle := b.Angle()

	if math.IsInf(radius, 0) {
		return chord
	}

	return math.Abs(radius * angle)
}

// ApproximateLength calculates length using chord and sagitta
func (b Bulge) ApproximateLength(start, end Vec2) float64 {
	chord := start.Distance(end)
	if !b.IsArc() {
		return chord
	}

	sagitta := b.Sagitta(chord)
	if sagitta == 0 {
		return chord
	}

	// Use parabolic approximation: length ≈ chord * (1 + 8/3 * (sagitta/chord)²)
	// This is more accurate than simple chord approximation
	ratio := sagitta / chord
	return chord * (1.0 + (8.0/3.0)*ratio*ratio)
}

// BoundingBox returns the bounding box of the arc segment
func (b Bulge) BoundingBox(start, end Vec2) (min, max Vec2) {
	if !b.IsArc() {
		// Simple line segment bounding box
		return Vec2{
				math.Min(start.X(), end.X()),
				math.Min(start.Y(), end.Y()),
			}, Vec2{
				math.Max(start.X(), end.X()),
				math.Max(start.Y(), end.Y()),
			}
	}

	center := b.Center(start, end)
	radius := b.Radius(start.Distance(end))

	if math.IsInf(radius, 0) {
		// Fallback to line segment
		return b.BoundingBox(start, end)
	}

	// Bounding box of full circle
	circleMin := Vec2{
		center.X() - radius,
		center.Y() - radius,
	}
	circleMax := Vec2{
		center.X() + radius,
		center.Y() + radius,
	}

	// Determine if arc endpoints extend beyond circle bounds
	// For simplicity, return the circle bounding box (conservative)
	return circleMin, circleMax
}

// Reverse reverses the direction of the bulge (start/end swap)
func (b Bulge) Reverse() Bulge {
	return -b
}

// Scale scales the bulge by a factor (maintains arc angle proportion)
func (b Bulge) Scale(factor float64) Bulge {
	return Bulge(float64(b) * factor)
}

// String returns string representation
func (b Bulge) String() string {
	return fmt.Sprintf("Bulge(%.6f)", b)
}

// BulgeSegment represents a polyline segment with bulge information
type BulgeSegment struct {
	Start Vec2
	End   Vec2
	Bulge Bulge
}

// Length returns the length of the bulge segment
func (bs BulgeSegment) Length() float64 {
	chord := bs.Start.Distance(bs.End)
	return bs.Bulge.Length(chord)
}

// Midpoint returns the midpoint of the segment
func (bs BulgeSegment) Midpoint() Vec2 {
	return bs.Start.Add(bs.End).Mul(0.5)
}

// BoundingBox returns the bounding box of the bulge segment
func (bs BulgeSegment) BoundingBox() (min, max Vec2) {
	return bs.Bulge.BoundingBox(bs.Start, bs.End)
}

// Approximate returns the segment approximated by line segments
func (bs BulgeSegment) Approximate(segments int) []Vec2 {
	return bs.Bulge.ApproximateArc(bs.Start, bs.End, segments)
}

// String returns string representation
func (bs BulgeSegment) String() string {
	return fmt.Sprintf("BulgeSegment{Start: %v, End: %v, Bulge: %v}", bs.Start, bs.End, bs.Bulge)
}
