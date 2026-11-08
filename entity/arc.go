package entity

import (
	"github.com/edanko/dxf/format"
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
	for i := 0; i < 2; i++ {
		f.WriteFloat(50+i, a.Angle[i])
	}
}

// Copy creates a deep copy of the Arc entity
func (a *Arc) Copy() Entity {
	arc := NewArc(a.Circle.Copy().(*Circle))
	arc.entity = a.entity
	if len(a.Angle) >= 2 {
		arc.Angle = []float64{a.Angle[0], a.Angle[1]}
	}
	return arc
}
