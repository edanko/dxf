package entity

import (
	"errors"
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

// ErrNonUniformScale is returned by circular types (CIRCLE, ARC) when a
// transform contains non-uniform scaling: the result is an ellipse, which the
// type cannot represent. The documented policy is to report the error rather
// than silently convert; callers may convert to ELLIPSE themselves.
var ErrNonUniformScale = errors.New("entity: transform contains non-uniform scale")

// TransformPoint applies an affine matrix to a point stored as [x y z].
func TransformPoint(m dxfmath.Matrix44, p []float64) []float64 {
	v := m.MulVec(toVec3(p))
	return []float64{v.X(), v.Y(), v.Z()}
}

// TransformDirection applies the linear part of a matrix to a direction.
func TransformDirection(m dxfmath.Matrix44, d []float64) []float64 {
	v := m.MulVec4(dxfmath.NewVec4FromVec3(toVec3(d)))
	_ = v
	return []float64{v.X(), v.Y(), v.Z()}
}

// vec3norm normalizes a direction vector returned by TransformDirection.
func vec3norm(d []float64) dxfmath.Vec3 {
	return dxfmath.NewVec3(d[0], d[1], d[2]).Normalized()
}

// ScaleFactors extracts the signed axis scale factors from the matrix linear
// part (lengths of the basis columns).
// TransformRotationDeg returns the Z-rotation component of a matrix in
// degrees, for entities that store rotation as an angle.
func TransformRotationDeg(m dxfmath.Matrix44) float64 {
	dx := TransformDirection(m, []float64{1, 0, 0})
	return math.Atan2(dx[1], dx[0]) * 180 / math.Pi
}

func ScaleFactors(m dxfmath.Matrix44) (sx, sy, sz float64) {
	sx = dxfmath.NewVec3(m.Get(0, 0), m.Get(1, 0), m.Get(2, 0)).Length()
	sy = dxfmath.NewVec3(m.Get(0, 1), m.Get(1, 1), m.Get(2, 1)).Length()
	sz = dxfmath.NewVec3(m.Get(0, 2), m.Get(1, 2), m.Get(2, 2)).Length()
	return sx, sy, sz
}

// IsUniformScale reports whether all axis scale factors are equal (within eps).
func IsUniformScale(m dxfmath.Matrix44, eps float64) bool {
	sx, sy, sz := ScaleFactors(m)
	if eps <= 0 {
		eps = 1e-9
	}
	d1 := sx - sy
	d2 := sy - sz
	return d1 <= eps && d1 >= -eps && d2 <= eps && d2 >= -eps
}

func toVec3(p []float64) dxfmath.Vec3 {
	v := dxfmath.NewVec3(0, 0, 0)
	if len(p) > 0 {
		v[0] = p[0]
	}
	if len(p) > 1 {
		v[1] = p[1]
	}
	if len(p) > 2 {
		v[2] = p[2]
	}
	return v
}
