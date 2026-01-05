package entity

import (
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

type SpatialFilter struct {
	*entity
	Name           string
	ClippingPoints []dxfmath.Vec2
	BoundaryPoints []dxfmath.Vec2
	IsInverted     bool
}

func NewSpatialFilter() *SpatialFilter {
	return &SpatialFilter{
		entity:         NewEntity(SPATIALFILTER),
		Name:           "",
		ClippingPoints: make([]dxfmath.Vec2, 0),
		BoundaryPoints: make([]dxfmath.Vec2, 0),
		IsInverted:     false,
	}
}

func (s *SpatialFilter) IsEntity() bool {
	return true
}

func (s *SpatialFilter) Format(f format.Formatter) {
	s.entity.Format(f)
	f.WriteString(100, "AcDbSpatialFilter")
	f.WriteString(2, s.Name)
	f.WriteInt(90, len(s.ClippingPoints))
	for _, pt := range s.ClippingPoints {
		f.WriteFloat(10, pt.X())
		f.WriteFloat(20, pt.Y())
	}
	f.WriteInt(91, len(s.BoundaryPoints))
	for _, pt := range s.BoundaryPoints {
		f.WriteFloat(11, pt.X())
		f.WriteFloat(21, pt.Y())
	}
	if s.IsInverted {
		f.WriteInt(70, 1)
	}
}

func (s *SpatialFilter) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (s *SpatialFilter) Copy() Entity {
	sf := NewSpatialFilter()
	sf.entity = s.entity
	sf.Name = s.Name
	sf.ClippingPoints = make([]dxfmath.Vec2, len(s.ClippingPoints))
	for i, pt := range s.ClippingPoints {
		sf.ClippingPoints[i] = pt
	}
	sf.BoundaryPoints = make([]dxfmath.Vec2, len(s.BoundaryPoints))
	for i, pt := range s.BoundaryPoints {
		sf.BoundaryPoints[i] = pt
	}
	sf.IsInverted = s.IsInverted
	return sf
}

func (s *SpatialFilter) Validate() error {
	return nil
}

func (s *SpatialFilter) AddClippingPoint(pt dxfmath.Vec2) {
	s.ClippingPoints = append(s.ClippingPoints, pt)
}

func (s *SpatialFilter) AddBoundaryPoint(pt dxfmath.Vec2) {
	s.BoundaryPoints = append(s.BoundaryPoints, pt)
}

func (s *SpatialFilter) ClearClippingPoints() {
	s.ClippingPoints = make([]dxfmath.Vec2, 0)
}

func (s *SpatialFilter) ClearBoundaryPoints() {
	s.BoundaryPoints = make([]dxfmath.Vec2, 0)
}
