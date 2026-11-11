package entity

import (
	"fmt"
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

// singularDeterminantEps is the determinant threshold below which a
// transformation matrix is treated as singular.
const singularDeterminantEps = 1e-12

// Transform applies the affine matrix m to every defining point and
// direction of the solid: the geometric center, all eight cached bounding
// box corners, and the extrusion direction are transformed, and the ACIS
// local-to-world placement is composed with m (new placement = m ∘ old)
// while the inverse placement is re-derived. The retained ACIS payload
// (text SAT lines or binary SAB data) is preserved verbatim: transforming
// the placement is how a 3DSOLID is moved without rewriting its opaque
// modeler data.
//
// A matrix whose linear part is singular (zero determinant) is rejected
// with an error naming the singular matrix, and the solid is left
// unchanged.
func (s *Solid3d) Transform(m dxfmath.Matrix44) error {
	if s == nil {
		return fmt.Errorf("solid transform: nil solid")
	}
	det := m.Determinant()
	if math.Abs(det) < singularDeterminantEps {
		return fmt.Errorf("solid transform: singular matrix (determinant %g)", det)
	}

	// Direction: the extrusion vector is transformed by the linear part of
	// the matrix only and renormalized.
	if e := s.entity.Extrusion(); len(e) == 3 {
		v := transformLinear(m, dxfmath.NewVec3(e[0], e[1], e[2]))
		if l := v.Length(); l > 0 {
			v = v.Mul(1 / l)
		}
		s.entity.SetExtrusion([]float64{v.X(), v.Y(), v.Z()})
	}

	// Point: the geometric center.
	s.center = m.MulVec(s.center)

	// Points: the eight bounding box corners define a new axis-aligned box.
	if s.boundingBox != nil {
		min, max := s.boundingBox.Min, s.boundingBox.Max
		first := true
		var newMin, newMax dxfmath.Vec3
		for xi := 0; xi < 2; xi++ {
			for yi := 0; yi < 2; yi++ {
				for zi := 0; zi < 2; zi++ {
					p := dxfmath.NewVec3(corner(xi, min.X(), max.X()), corner(yi, min.Y(), max.Y()), corner(zi, min.Z(), max.Z()))
					tp := m.MulVec(p)
					if first {
						newMin, newMax = tp, tp
						first = false
						continue
					}
					newMin = dxfmath.NewVec3(
						math.Min(newMin.X(), tp.X()),
						math.Min(newMin.Y(), tp.Y()),
						math.Min(newMin.Z(), tp.Z()),
					)
					newMax = dxfmath.NewVec3(
						math.Max(newMax.X(), tp.X()),
						math.Max(newMax.Y(), tp.Y()),
						math.Max(newMax.Z(), tp.Z()),
					)
				}
			}
		}
		s.boundingBox = NewACISBoundingBox(newMin, newMax)
	}

	// Placement: compose the entity transform with the ACIS placement so
	// the payload coordinates (ACIS local space) are untouched.
	base := s.Placement
	if base == nil {
		base = NewACASTransform()
	}
	baseM := dxfmath.NewMatrix44FromSlice(base.Matrix[:])
	composed := m.Mul(baseM)
	s.Placement = &ACASTransform{Matrix: composed}

	if inv, ok := composed.Inverse(); ok {
		s.InverseTransform = &ACASTransform{Matrix: inv}
	} else {
		// The composed matrix cannot be singular after the determinant
		// check above, but keep the field consistent if the inverse
		// computation ever fails numerically.
		s.InverseTransform = nil
	}
	return nil
}

// corner returns one of the two box bounds per axis index.
func corner(i int, lo, hi float64) float64 {
	if i == 0 {
		return lo
	}
	return hi
}

// transformLinear applies only the linear (3x3) part of an affine matrix.
func transformLinear(m dxfmath.Matrix44, v dxfmath.Vec3) dxfmath.Vec3 {
	return dxfmath.NewVec3(
		m.Get(0, 0)*v.X()+m.Get(0, 1)*v.Y()+m.Get(0, 2)*v.Z(),
		m.Get(1, 0)*v.X()+m.Get(1, 1)*v.Y()+m.Get(1, 2)*v.Z(),
		m.Get(2, 0)*v.X()+m.Get(2, 1)*v.Y()+m.Get(2, 2)*v.Z(),
	)
}

// Inverse returns the inverse of the 4x4 ACIS transform. It reports false
// when the transform is singular.
func (t *ACASTransform) Inverse() (*ACASTransform, bool) {
	if t == nil {
		return nil, false
	}
	m := dxfmath.NewMatrix44FromSlice(t.Matrix[:])
	inv, ok := m.Inverse()
	if !ok {
		return nil, false
	}
	return &ACASTransform{Matrix: inv}, true
}
