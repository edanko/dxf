package entity

import (
	"github.com/edanko/dxf/math"
)

// Capability interfaces for optional entity operations. Callers type-assert:
// a type that does not implement an interface simply does not support the
// operation; there are no no-op implementations hiding failures.

// Transformer applies an affine transform to an entity's geometry in place.
type Transformer interface {
	Transform(m math.Matrix44) error
}

// Copier produces a deep copy of an entity.
type Copier interface {
	Copy() Entity
}

// Bounder computes an axis-aligned bounding box (min, max corners).
type Bounder interface {
	BBox() ([]float64, []float64)
}

// Exploder decomposes an entity into constituent entities (INSERT -> expanded
// content, POLYLINE -> LINE segments, DIMENSION -> geometry, ...).
type Exploder interface {
	Explode() ([]Entity, error)
}

// SubentityOwner exposes child records of composite entities (POLYLINE
// vertices, INSERT attribs, MLEADER leader lines, dimensions).
type SubentityOwner interface {
	Subentities() []Entity
}

// Validator reports entity problems for audit; an empty slice means valid.
type Validator interface {
	Validate() []string
}
