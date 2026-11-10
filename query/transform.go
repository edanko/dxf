// Transform operations for entity sets and selections.
package query

import (
	"math"

	dxfmath "github.com/edanko/dxf/math"

	"github.com/edanko/dxf/entity"
)

// Transformer mutates entity geometry by a matrix.
type Transformer interface {
	Transform(m dxfmath.Matrix44) error
}

// Apply transforms every entity that supports transformation, returning
// per-handle errors for unsupported entities.
func Apply(m dxfmath.Matrix44, ents ...entity.Entity) []error {
	var errs []error
	for _, e := range ents {
		if u, ok := e.(interface{ Unwrap() entity.Entity }); ok {
			e = u.Unwrap()
		}
		t, ok := e.(Transformer)
		if !ok {
			errs = append(errs, newTransformError(e, "entity does not support transformation"))
			continue
		}
		if err := t.Transform(m); err != nil {
			errs = append(errs, newTransformError(e, err.Error()))
		}
	}
	return errs
}

// transformError names the entity a transform failed on.
type transformError struct {
	handle string
	typ    string
	reason string
}

func newTransformError(e entity.Entity, reason string) error {
	return &transformError{handle: e.Handle(), typ: e.DXFType(), reason: reason}
}

func (e *transformError) Error() string {
	return "transform " + e.typ + " [" + e.handle + "]: " + e.reason
}

// Handle returns the handle of the entity the transform failed on.
func (e *transformError) Handle() string { return e.handle }

// TranslateMatrix builds a translation matrix.
func TranslateMatrix(dx, dy, dz float64) dxfmath.Matrix44 {
	return dxfmath.Translation(dx, dy, dz)
}

// RotationMatrix builds a rotation about the Z axis through (cx, cy) by
// angleDeg degrees.
func RotationMatrix(angleDeg, cx, cy float64) dxfmath.Matrix44 {
	rad := angleDeg * math.Pi / 180
	return dxfmath.Translation(cx, cy, 0).
		Mul(dxfmath.RotationZ(rad)).
		Mul(dxfmath.Translation(-cx, -cy, 0))
}

// ScaleMatrix builds a uniform scale about (cx, cy).
func ScaleMatrix(factor, cx, cy float64) dxfmath.Matrix44 {
	return dxfmath.Translation(cx, cy, 0).
		Mul(dxfmath.Scale(factor, factor, factor)).
		Mul(dxfmath.Translation(-cx, -cy, 0))
}

// NonUniformScaleMatrix builds an axis scale about (cx, cy).
func NonUniformScaleMatrix(sx, sy, cx, cy float64) dxfmath.Matrix44 {
	return dxfmath.Translation(cx, cy, 0).
		Mul(dxfmath.Scale(sx, sy, 1)).
		Mul(dxfmath.Translation(-cx, -cy, 0))
}

// Translate moves entities by a vector.
func Translate(dx, dy, dz float64, ents ...entity.Entity) []error {
	return Apply(TranslateMatrix(dx, dy, dz), ents...)
}

// Rotate rotates entities about (cx, cy) by angleDeg degrees.
func Rotate(angleDeg, cx, cy float64, ents ...entity.Entity) []error {
	return Apply(RotationMatrix(angleDeg, cx, cy), ents...)
}

// Scale scales entities uniformly about (cx, cy).
func Scale(factor, cx, cy float64, ents ...entity.Entity) []error {
	return Apply(ScaleMatrix(factor, cx, cy), ents...)
}

// ApplyM transforms the entities behind a selection resolved by src.
func (s *Selection) ApplyM(m dxfmath.Matrix44, src EntitySource) []error {
	return Apply(m, s.Entities(src)...)
}

// Translate moves the selected entities.
func (s *Selection) Translate(dx, dy, dz float64, src EntitySource) []error {
	return Translate(dx, dy, dz, s.Entities(src)...)
}

// Rotate rotates the selected entities about (cx, cy) by angleDeg degrees.
func (s *Selection) Rotate(angleDeg, cx, cy float64, src EntitySource) []error {
	return Rotate(angleDeg, cx, cy, s.Entities(src)...)
}

// Scale scales the selected entities uniformly about (cx, cy).
func (s *Selection) Scale(factor, cx, cy float64, src EntitySource) []error {
	return Scale(factor, cx, cy, s.Entities(src)...)
}
