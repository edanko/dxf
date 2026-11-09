package math

import (
	"errors"
	"math"
)

// ErrSingularMatrix is returned when a matrix cannot be inverted or
// decomposed because it is singular (determinant effectively zero).
var ErrSingularMatrix = errors.New("math: singular matrix")

// Decomposition holds the result of decomposing an affine 4x4 matrix into
// translation, rotation, scale, and shear components.
//
// The decomposition uses the convention M = T * R * Scale * Shear, with
// row-major storage and column-vector multiplication (the same convention as
// Matrix44).
type Decomposition struct {
	Translation Vec3
	Rotation    Matrix44
	Scale       Vec3
	Shear       Vec3 // Shear.X = XY, Shear.Y = XZ, Shear.Z = YZ
}

// Decompose splits an affine matrix into translation, rotation, scale, and
// shear. It returns ErrSingularMatrix when the linear part is singular.
func (m Matrix44) Decompose() (Decomposition, error) {
	var d Decomposition

	col0 := Vec3{m[0], m[4], m[8]}
	col1 := Vec3{m[1], m[5], m[9]}
	col2 := Vec3{m[2], m[6], m[10]}
	d.Translation = Vec3{m[3], m[7], m[11]}

	det := col0.Dot(col1.Cross(col2))
	if math.Abs(det) <= DefaultTolerance {
		return d, ErrSingularMatrix
	}

	sx := col0.Length()
	if sx <= DefaultTolerance {
		return d, ErrSingularMatrix
	}
	e0 := col0.Div(sx)

	xy := col1.Dot(e0)
	col1p := col1.Sub(e0.Mul(xy))
	sy := col1p.Length()
	if sy <= DefaultTolerance {
		return d, ErrSingularMatrix
	}
	e1 := col1p.Div(sy)

	xz := col2.Dot(e0)
	col2p := col2.Sub(e0.Mul(xz))
	yz := col2p.Dot(e1)
	col2p = col2p.Sub(e1.Mul(yz))
	sz := col2p.Length()
	if sz <= DefaultTolerance {
		return d, ErrSingularMatrix
	}
	e2 := col2p.Div(sz)

	d.Scale = Vec3{sx, sy, sz}
	d.Shear = Vec3{xy / sx, xz / sx, yz / sy}

	// Preserve reflections in the scale sign rather than the rotation.
	if det < 0 {
		d.Scale[2] = -d.Scale[2]
		e2 = e2.Mul(-1)
	}

	d.Rotation = Matrix44{
		e0[0], e1[0], e2[0], 0,
		e0[1], e1[1], e2[1], 0,
		e0[2], e1[2], e2[2], 0,
		0, 0, 0, 1,
	}
	return d, nil
}

// InverseChecked is the error-returning form of Inverse.
func (m Matrix44) InverseChecked() (Matrix44, error) {
	inv, ok := m.Inverse()
	if !ok {
		return NewMatrix44(), ErrSingularMatrix
	}
	return inv, nil
}
