package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

type Polygon struct {
	*entity
	Vertices []math.Vec3
	Flags    int
}

func NewPolygon() *Polygon {
	return &Polygon{
		entity:   NewEntity(POLYGON),
		Vertices: make([]math.Vec3, 0),
		Flags:    0,
	}
}

func (p *Polygon) IsEntity() bool {
	return true
}

func (p *Polygon) Format(f format.Formatter) {
	p.entity.Format(f)
	f.WriteString(100, "AcDbPolygon")
	f.WriteInt(90, len(p.Vertices))
	for i, v := range p.Vertices {
		f.WriteFloat(10+i*10, v.X())
		f.WriteFloat(20+i*10, v.Y())
		f.WriteFloat(30+i*10, v.Z())
	}
	f.WriteInt(70, p.Flags)
}

func (p *Polygon) BBox() ([]float64, []float64) {
	if len(p.Vertices) == 0 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}
	min := []float64{1e10, 1e10, 1e10}
	max := []float64{-1e10, -1e10, -1e10}
	for _, v := range p.Vertices {
		if v.X() < min[0] {
			min[0] = v.X()
		}
		if v.Y() < min[1] {
			min[1] = v.Y()
		}
		if v.Z() < min[2] {
			min[2] = v.Z()
		}
		if v.X() > max[0] {
			max[0] = v.X()
		}
		if v.Y() > max[1] {
			max[1] = v.Y()
		}
		if v.Z() > max[2] {
			max[2] = v.Z()
		}
	}
	return min, max
}

func (p *Polygon) Copy() Entity {
	pg := NewPolygon()
	pg.entity = p.entity
	pg.Vertices = make([]math.Vec3, len(p.Vertices))
	for i, v := range p.Vertices {
		pg.Vertices[i] = v
	}
	pg.Flags = p.Flags
	return pg
}

func (p *Polygon) Validate() error {
	return nil
}

func (p *Polygon) AddVertex(v math.Vec3) {
	p.Vertices = append(p.Vertices, v)
}

func (p *Polygon) ClearVertices() {
	p.Vertices = make([]math.Vec3, 0)
}

func (p *Polygon) SetVertices(vertices []math.Vec3) {
	p.Vertices = make([]math.Vec3, len(vertices))
	for i, v := range vertices {
		p.Vertices[i] = v
	}
}
