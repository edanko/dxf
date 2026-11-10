package entity

import (
	"github.com/edanko/dxf/format"
)

// Vertex represents VERTEX Entity. A vertex belongs to a POLYLINE owner;
// code 330 (owner) is carried on the shared base.
type Vertex struct {
	*entity
	Flag       int       // 70
	Coord      []float64 // 10, 20, 30
	Bulge      float64   // 42
	StartWidth float64   // 40
	EndWidth   float64   // 41
	Extra      int       // 71 (mesh vertex cache)
	EdgeFlags  int       // 72 (mesh vertex edge flags)
}

// IsEntity is for Entity interface.
func (v *Vertex) IsEntity() bool {
	return true
}

// NewVertex creates a new Vertex.
func NewVertex(x, y, z float64) *Vertex {
	v := &Vertex{
		entity: NewEntity(VERTEX),
		Flag:   32,
		Coord:  []float64{x, y, z},
	}
	return v
}

// Format writes data to formatter.
func (v *Vertex) Format(f format.Formatter) {
	v.entity.Format(f)
	f.WriteString(100, "AcDbVertex")
	f.WriteString(100, "AcDb3dPolylineVertex")
	for i := range 3 {
		f.WriteFloat((i+1)*10, v.Coord[i])
	}
	f.WriteInt(70, v.Flag)
	if v.Bulge != 0 {
		f.WriteFloat(42, v.Bulge)
	}
	if v.StartWidth != 0 {
		f.WriteFloat(40, v.StartWidth)
	}
	if v.EndWidth != 0 {
		f.WriteFloat(41, v.EndWidth)
	}
	if v.Extra != 0 {
		f.WriteInt(71, v.Extra)
	}
	if v.EdgeFlags != 0 {
		f.WriteInt(72, v.EdgeFlags)
	}
}

func (v *Vertex) BBox() ([]float64, []float64) {
	mins := []float64{v.Coord[0], v.Coord[1], v.Coord[2]}
	maxs := []float64{v.Coord[0], v.Coord[1], v.Coord[2]}
	return mins, maxs
}

// Copy creates a deep copy of the Vertex entity
func (v *Vertex) Copy() Entity {
	vertex := NewVertex(v.Coord[0], v.Coord[1], v.Coord[2])
	vertex.entity = v.entity
	vertex.Flag = v.Flag
	vertex.Bulge = v.Bulge
	vertex.StartWidth = v.StartWidth
	vertex.EndWidth = v.EndWidth
	vertex.Extra = v.Extra
	vertex.EdgeFlags = v.EdgeFlags
	return vertex
}
