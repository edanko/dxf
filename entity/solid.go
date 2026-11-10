package entity

import (
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

type Solid struct {
	*entity
	FirstPoint          []float64 //10, 20, 30
	SecondPoint         []float64 //11, 21, 31
	ThirdPoint          []float64 //12 ,22, 32
	FourthPoint         []float64 // 13,23, 33
	Thickness           float64   //0, 0, 1
	StretchingDirection []float64 // 210,220, 230
}

func NewSolid() *Solid {
	return &Solid{
		entity:              NewEntity(SOLID),
		FirstPoint:          []float64{0.0, 0.0, 0.0},
		SecondPoint:         []float64{0.0, 0.0, 0.0},
		ThirdPoint:          []float64{0.0, 0.0, 0.0},
		FourthPoint:         []float64{0.0, 0.0, 0.0},
		Thickness:           0.0,
		StretchingDirection: []float64{0.0, 0.0, 1.0},
	}
}

func (s *Solid) IsEntity() bool {
	return true
}

func (s *Solid) Format(f format.Formatter) {
	s.entity.Format(f)
	f.WriteString(100, "AcDbTrace")
	for i := range 3 {
		f.WriteFloat((i+1)*10, s.FirstPoint[i])
	}
	for i := range 3 {
		f.WriteFloat((i+1)*10+1, s.SecondPoint[i])
	}
	for i := range 3 {
		f.WriteFloat((i+1)*10+2, s.ThirdPoint[i])
	}
	for i := range 3 {
		f.WriteFloat((i+1)*10+3, s.FourthPoint[i])
	}
}

func (s *Solid) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	points := [][]float64{s.FirstPoint, s.SecondPoint, s.ThirdPoint, s.FourthPoint}
	for _, p := range points {
		for i := range 3 {
			if p[i] < mins[i] {
				mins[i] = p[i]
			}
			if p[i] > maxs[i] {
				maxs[i] = p[i]
			}
		}
	}
	return mins, maxs
}

// Transform applies an affine matrix to the solid corner points.
func (s *Solid) Transform(m dxfmath.Matrix44) error {
	s.FirstPoint = TransformPoint(m, s.FirstPoint)
	s.SecondPoint = TransformPoint(m, s.SecondPoint)
	s.ThirdPoint = TransformPoint(m, s.ThirdPoint)
	s.FourthPoint = TransformPoint(m, s.FourthPoint)
	return nil
}

// Copy creates a deep copy of the Solid entity
func (s *Solid) Copy() Entity {
	solid := NewSolid()
	solid.entity = s.entity
	solid.FirstPoint = []float64{s.FirstPoint[0], s.FirstPoint[1], s.FirstPoint[2]}
	solid.SecondPoint = []float64{s.SecondPoint[0], s.SecondPoint[1], s.SecondPoint[2]}
	solid.ThirdPoint = []float64{s.ThirdPoint[0], s.ThirdPoint[1], s.ThirdPoint[2]}
	solid.FourthPoint = []float64{s.FourthPoint[0], s.FourthPoint[1], s.FourthPoint[2]}
	solid.Thickness = s.Thickness
	solid.StretchingDirection = []float64{s.StretchingDirection[0], s.StretchingDirection[1], s.StretchingDirection[2]}
	return solid
}
