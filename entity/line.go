package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// Line represents LINE Entity.
type Line struct {
	*entity
	Start               []float64 // 10, 20, 30
	End                 []float64 // 11, 21, 31
	Thickness           float64   //0, 0, 1
	StretchingDirection []float64 // 210,220, 230
}

// IsEntity is for Entity interface.
func (l *Line) IsEntity() bool {
	return true
}

// NewLine creates a new Line.
func NewLine() *Line {
	l := &Line{
		entity: NewEntity(LINE),
		Start:  []float64{0.0, 0.0, 0.0},
		End:    []float64{0.0, 0.0, 0.0},
	}
	return l
}

// Format writes data to formatter.
func (l *Line) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLine")
	if l.thickness != 0 {
		f.WriteFloat(39, l.thickness)
	}
	for i := range 3 {
		f.WriteFloat((i+1)*10, l.Start[i])
	}
	for i := range 3 {
		f.WriteFloat((i+1)*10+1, l.End[i])
	}
	l.entity.FormatBaseExtrusion(f)
}

func (l *Line) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for i := range 3 {
		if l.Start[i] <= l.End[i] {
			mins[i] = l.Start[i]
			maxs[i] = l.End[i]
		} else {
			mins[i] = l.End[i]
			maxs[i] = l.Start[i]
		}
	}
	return mins, maxs
}

func (l *Line) Length() float64 {
	sum := 0.0
	for i := range 3 {
		sum += math.Pow(l.End[i]-l.Start[i], 2.0)
	}
	return math.Sqrt(sum)
}

func (l *Line) Direction(normalize bool) []float64 {
	vec := make([]float64, 3)
	var length float64
	if normalize {
		length = l.Length()
		if length == 0.0 {
			length = 1.0
		}
	}
	for i := range 3 {
		vec[i] = l.End[i] - l.Start[i]
		if normalize {
			vec[i] /= length
		}
	}
	return vec
}

func (l *Line) Move(x, y, z float64) {
	for i, val := range []float64{x, y, z} {
		l.Start[i] += val
		l.End[i] += val
	}
}

// Transform applies an affine matrix to the line endpoints.
func (l *Line) Transform(m dxfmath.Matrix44) error {
	l.Start = TransformPoint(m, l.Start)
	l.End = TransformPoint(m, l.End)
	if len(l.StretchingDirection) >= 3 {
		l.StretchingDirection = TransformDirection(m, l.StretchingDirection)
	}
	return nil
}

// Copy creates a deep copy of the Line entity
func (l *Line) Copy() Entity {
	line := NewLine()
	line.entity = l.entity.clone()
	if len(l.Start) >= 3 {
		line.Start = []float64{l.Start[0], l.Start[1], l.Start[2]}
	}
	if len(l.End) >= 3 {
		line.End = []float64{l.End[0], l.End[1], l.End[2]}
	}
	line.Thickness = l.Thickness
	if len(l.StretchingDirection) >= 3 {
		line.StretchingDirection = []float64{l.StretchingDirection[0], l.StretchingDirection[1], l.StretchingDirection[2]}
	}
	return line
}
