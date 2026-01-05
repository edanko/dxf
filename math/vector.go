package math

import (
	"fmt"
	"math"
)

// Vec2 represents a 2D vector
type Vec2 [2]float64

// NewVec2 creates a new 2D vector
func NewVec2(x, y float64) Vec2 {
	return Vec2{x, y}
}

// X returns the X component
func (v Vec2) X() float64 { return v[0] }

// Y returns the Y component
func (v Vec2) Y() float64 { return v[1] }

// Add adds another vector
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{v[0] + other[0], v[1] + other[1]}
}

// Sub subtracts another vector
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{v[0] - other[0], v[1] - other[1]}
}

// Mul multiplies by a scalar
func (v Vec2) Mul(scalar float64) Vec2 {
	return Vec2{v[0] * scalar, v[1] * scalar}
}

// Div divides by a scalar
func (v Vec2) Div(scalar float64) Vec2 {
	return Vec2{v[0] / scalar, v[1] / scalar}
}

// Dot returns the dot product
func (v Vec2) Dot(other Vec2) float64 {
	return v[0]*other[0] + v[1]*other[1]
}

// Length returns the vector length
func (v Vec2) Length() float64 {
	return math.Sqrt(v.Dot(v))
}

// LengthSquared returns the squared length
func (v Vec2) LengthSquared() float64 {
	return v.Dot(v)
}

// Normalized returns a normalized vector
func (v Vec2) Normalized() Vec2 {
	length := v.Length()
	if length == 0 {
		return Vec2{0, 0}
	}
	return v.Div(length)
}

// Normalize normalizes the vector in place
func (v *Vec2) Normalize() {
	*v = v.Normalized()
}

// Distance returns distance to another vector
func (v Vec2) Distance(other Vec2) float64 {
	return v.Sub(other).Length()
}

// DistanceSquared returns squared distance to another vector
func (v Vec2) DistanceSquared(other Vec2) float64 {
	return v.Sub(other).LengthSquared()
}

// Lerp linearly interpolates between two vectors
func (v Vec2) Lerp(other Vec2, t float64) Vec2 {
	return v.Add(other.Sub(v).Mul(t))
}

// Rotate rotates the vector by angle (radians)
func (v Vec2) Rotate(angle float64) Vec2 {
	cos, sin := math.Cos(angle), math.Sin(angle)
	return Vec2{
		v[0]*cos - v[1]*sin,
		v[0]*sin + v[1]*cos,
	}
}

// Angle returns the angle of the vector
func (v Vec2) Angle() float64 {
	return math.Atan2(v[1], v[0])
}

// AngleTo returns the angle to another vector
func (v Vec2) AngleTo(other Vec2) float64 {
	return math.Atan2(other[1]-v[1], other[0]-v[0])
}

// Perpendicular returns a perpendicular vector (rotated 90 degrees)
func (v Vec2) Perpendicular() Vec2 {
	return Vec2{-v[1], v[0]}
}

// Reflect reflects the vector across a normal
func (v Vec2) Reflect(normal Vec2) Vec2 {
	return v.Sub(normal.Mul(2 * v.Dot(normal)))
}

// Project projects the vector onto another vector
func (v Vec2) Project(other Vec2) Vec2 {
	if other.LengthSquared() == 0 {
		return Vec2{0, 0}
	}
	return other.Mul(v.Dot(other) / other.LengthSquared())
}

// IsEqual checks if two vectors are approximately equal
func (v Vec2) IsEqual(other Vec2, epsilon float64) bool {
	return math.Abs(v[0]-other[0]) < epsilon && math.Abs(v[1]-other[1]) < epsilon
}

// IsZero checks if the vector is approximately zero
func (v Vec2) IsZero(epsilon float64) bool {
	return math.Abs(v[0]) < epsilon && math.Abs(v[1]) < epsilon
}

// String returns string representation
func (v Vec2) String() string {
	return fmt.Sprintf("Vec2(%.6f, %.6f)", v[0], v[1])
}
