package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// Polyline represents POLYLINE Entity.
type Polyline struct {
	*entity
	Flag      int
	size      int
	Vertices  []*Vertex
	endhandle string
}

// IsEntity is for Entity interface.
func (p *Polyline) IsEntity() bool {
	return true
}

// NewPolyline creates a new Polyline.
func NewPolyline() *Polyline {
	vs := make([]*Vertex, 0)
	p := &Polyline{
		entity:   NewEntity(POLYLINE),
		Flag:     8,
		size:     0,
		Vertices: vs,
	}
	return p
}

// Format writes data to formatter.
func (p *Polyline) Format(f format.Formatter) {
	p.entity.Format(f)
	f.WriteString(100, "AcDb3dPolyline")
	f.WriteInt(66, 1)
	f.WriteString(10, "0.0")
	f.WriteString(20, "0.0")
	f.WriteString(30, "0.0")
	f.WriteInt(70, p.Flag)
	for _, v := range p.Vertices {
		v.Format(f)
	}
	f.WriteString(0, "SEQEND")
	f.WriteString(5, p.endhandle)
	f.WriteString(100, "AcDbEntity")
	f.WriteString(8, p.Layer().Name())
}

// Close closes Polyline.
func (p *Polyline) Close() {
	p.Flag |= 1
}

// AddVertex adds a new vertex to Polyline.
func (p *Polyline) AddVertex(x, y, z float64) *Vertex {
	v := NewVertex(x, y, z)
	p.Vertices = append(p.Vertices, v)
	p.size++
	v.SetLayer(p.Layer())
	v.SetOwner(p)
	return v
}

// SetHandle sets handles to itself and its vertices.
func (p *Polyline) SetHandle(hg *handle.HandleGenerator) {
	p.entity.SetHandle(hg)
	for _, v := range p.Vertices {
		v.SetHandle(hg)
	}
	p.endhandle = hg.Next()
}

func (p *Polyline) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for _, v := range p.Vertices {
		for i := 0; i < 3; i++ {
			if v.Coord[i] < mins[i] {
				mins[i] = v.Coord[i]
			}
			if v.Coord[i] > maxs[i] {
				maxs[i] = v.Coord[i]
			}
		}
	}
	return mins, maxs
}
