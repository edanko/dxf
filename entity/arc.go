package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

type Arc struct {
	*Circle
	Angle []float64 // 50, 51 (Degree)
}

// IsEntity is for Entity interface.
func (a *Arc) IsEntity() bool {
	return true
}

// NewArc creates a new Arc.
func NewArc(c *Circle) *Arc {
	if c == nil {
		c = NewCircle()
	}
	a := &Arc{
		Circle: c,
		Angle:  []float64{0.0, 180.0},
	}
	a.SetEntityType(ARC)
	return a
}

// Format writes data to formatter.
func (a *Arc) Format(f format.Formatter) {
	a.Circle.Format(f)
	f.WriteString(100, "AcDbArc")
	for i := range 2 {
		f.WriteFloat(50+i, a.Angle[i])
	}
}

// Transform applies an affine matrix to the arc. The center transforms as a
// point and the radius as a circle; angles shift by the rotation applied in
// the XY plane (2D rotations only). Non-uniform scale returns
// ErrNonUniformScale per the CIRCLE policy.
func (a *Arc) Transform(m dxfmath.Matrix44) error {
	if !IsUniformScale(m, 1e-9) {
		return ErrNonUniformScale
	}
	sx, _, _ := ScaleFactors(m)
	a.Center = TransformPoint(m, a.Center)
	a.Radius = abs64(a.Radius * sx)
	rot := math.Atan2(m.Get(1, 0), m.Get(0, 0))
	if rot != 0 {
		deg := rot * 180 / math.Pi
		a.Angle[0] += deg
		a.Angle[1] += deg
	}
	if len(a.Direction) >= 3 {
		d := TransformDirection(m, a.Direction)
		n := dxfmath.NewVec3(d[0], d[1], d[2]).Normalized()
		a.Direction = []float64{n.X(), n.Y(), n.Z()}
	}
	return nil
}

// Copy creates a deep copy of the Arc entity
func (a *Arc) Copy() Entity {
	arc := NewArc(a.Circle.Copy().(*Circle))
	arc.entity = a.entity.clone()
	if len(a.Angle) >= 2 {
		arc.Angle = []float64{a.Angle[0], a.Angle[1]}
	}
	return arc
}
