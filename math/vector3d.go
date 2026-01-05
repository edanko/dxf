package math

import (
	"fmt"
	"math"
)

// Mathematical constants
const (
	// Pi is the ratio of a circle's circumference to its diameter
	Pi = math.Pi
)

// Vec3 represents a 3D vector
type Vec3 [3]float64

// NewVec3 creates a new 3D vector
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

// NewVec3FromVec2 creates a 3D vector from a 2D vector (z=0)
func NewVec3FromVec2(v Vec2) Vec3 {
	return Vec3{v[0], v[1], 0}
}

// NewVec3FromSlice creates a 3D vector from a slice
func NewVec3FromSlice(slice []float64) Vec3 {
	if len(slice) >= 3 {
		return Vec3{slice[0], slice[1], slice[2]}
	}
	if len(slice) == 2 {
		return Vec3{slice[0], slice[1], 0}
	}
	return Vec3{0, 0, 0}
}

// X returns X component
func (v Vec3) X() float64 { return v[0] }

// Y returns Y component
func (v Vec3) Y() float64 { return v[1] }

// Z returns Z component
func (v Vec3) Z() float64 { return v[2] }

// XY returns XY as Vec2
func (v Vec3) XY() Vec2 { return Vec2{v[0], v[1]} }

// XZ returns XZ as Vec2
func (v Vec3) XZ() Vec2 { return Vec2{v[0], v[2]} }

// YZ returns YZ as Vec2
func (v Vec3) YZ() Vec2 { return Vec2{v[1], v[2]} }

// Add adds another vector
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v[0] + other[0], v[1] + other[1], v[2] + other[2]}
}

// Sub subtracts another vector
func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v[0] - other[0], v[1] - other[1], v[2] - other[2]}
}

// Mul multiplies by a scalar
func (v Vec3) Mul(scalar float64) Vec3 {
	return Vec3{v[0] * scalar, v[1] * scalar, v[2] * scalar}
}

// Div divides by a scalar
func (v Vec3) Div(scalar float64) Vec3 {
	return Vec3{v[0] / scalar, v[1] / scalar, v[2] / scalar}
}

// Dot returns the dot product
func (v Vec3) Dot(other Vec3) float64 {
	return v[0]*other[0] + v[1]*other[1] + v[2]*other[2]
}

// Cross returns the cross product
func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		v[1]*other[2] - v[2]*other[1],
		v[2]*other[0] - v[0]*other[2],
		v[0]*other[1] - v[1]*other[0],
	}
}

// Length returns the vector length
func (v Vec3) Length() float64 {
	return math.Sqrt(v.Dot(v))
}

// LengthSquared returns the squared length
func (v Vec3) LengthSquared() float64 {
	return v.Dot(v)
}

// Normalized returns a normalized vector
func (v Vec3) Normalized() Vec3 {
	length := v.Length()
	if length == 0 {
		return Vec3{0, 0, 0}
	}
	return v.Div(length)
}

// Normalize normalizes the vector in place
func (v *Vec3) Normalize() {
	*v = v.Normalized()
}

// Distance returns distance to another vector
func (v Vec3) Distance(other Vec3) float64 {
	return v.Sub(other).Length()
}

// DistanceSquared returns squared distance to another vector
func (v Vec3) DistanceSquared(other Vec3) float64 {
	return v.Sub(other).LengthSquared()
}

// Lerp linearly interpolates between two vectors
func (v Vec3) Lerp(other Vec3, t float64) Vec3 {
	return v.Add(other.Sub(v).Mul(t))
}

// Angle returns the angle between two vectors in radians
func (v Vec3) Angle(other Vec3) float64 {
	lengths := v.Length() * other.Length()
	if lengths == 0 {
		return 0
	}
	dot := v.Dot(other) / lengths
	// Clamp to [-1, 1] to avoid domain errors due to floating point precision
	if dot > 1 {
		dot = 1
	} else if dot < -1 {
		dot = -1
	}
	return math.Acos(dot)
}

// Project projects the vector onto another vector
func (v Vec3) Project(other Vec3) Vec3 {
	if other.LengthSquared() == 0 {
		return Vec3{0, 0, 0}
	}
	return other.Mul(v.Dot(other) / other.LengthSquared())
}

// Reject returns the rejection component (perpendicular to other)
func (v Vec3) Reject(other Vec3) Vec3 {
	return v.Sub(v.Project(other))
}

// Reflect reflects the vector across a normal
func (v Vec3) Reflect(normal Vec3) Vec3 {
	return v.Sub(normal.Mul(2 * v.Dot(normal)))
}

// IsEqual checks if two vectors are approximately equal
func (v Vec3) IsEqual(other Vec3, epsilon float64) bool {
	return math.Abs(v[0]-other[0]) < epsilon &&
		math.Abs(v[1]-other[1]) < epsilon &&
		math.Abs(v[2]-other[2]) < epsilon
}

// IsZero checks if the vector is approximately zero
func (v Vec3) IsZero(epsilon float64) bool {
	return math.Abs(v[0]) < epsilon &&
		math.Abs(v[1]) < epsilon &&
		math.Abs(v[2]) < epsilon
}

// ToSlice converts vector to slice
func (v Vec3) ToSlice() []float64 {
	return []float64{v[0], v[1], v[2]}
}

// Add3 adds another 3D vector to this vector
func (v Vec3) Add3(other Vec3) Vec3 {
	return Vec3{v[0] + other[0], v[1] + other[1], v[2] + other[2]}
}

// String returns string representation
func (v Vec3) String() string {
	return fmt.Sprintf("Vec3(%.6f, %.6f, %.6f)", v[0], v[1], v[2])
}

// Vec4 represents a 4D vector (for homogeneous coordinates)
type Vec4 [4]float64

// NewVec4 creates a new 4D vector
func NewVec4(x, y, z, w float64) Vec4 {
	return Vec4{x, y, z, w}
}

// NewVec4FromVec3 creates a 4D vector from a 3D vector (w=1)
func NewVec4FromVec3(v Vec3) Vec4 {
	return Vec4{v[0], v[1], v[2], 1}
}

// X returns X component
func (v Vec4) X() float64 { return v[0] }

// Y returns Y component
func (v Vec4) Y() float64 { return v[1] }

// Z returns Z component
func (v Vec4) Z() float64 { return v[2] }

// W returns W component
func (v Vec4) W() float64 { return v[3] }

// XYZ returns XYZ as Vec3
func (v Vec4) XYZ() Vec3 {
	if v[3] != 0 {
		return Vec3{v[0] / v[3], v[1] / v[3], v[2] / v[3]}
	}
	return Vec3{v[0], v[1], v[2]}
}

// Add adds another vector
func (v Vec4) Add(other Vec4) Vec4 {
	return Vec4{v[0] + other[0], v[1] + other[1], v[2] + other[2], v[3] + other[3]}
}

// Sub subtracts another vector
func (v Vec4) Sub(other Vec4) Vec4 {
	return Vec4{v[0] - other[0], v[1] - other[1], v[2] - other[2], v[3] - other[3]}
}

// Mul multiplies by a scalar
func (v Vec4) Mul(scalar float64) Vec4 {
	return Vec4{v[0] * scalar, v[1] * scalar, v[2] * scalar, v[3] * scalar}
}

// String returns string representation
func (v Vec4) String() string {
	return fmt.Sprintf("Vec4(%.6f, %.6f, %.6f, %.6f)", v[0], v[1], v[2], v[3])
}
