package entity

import (
	"fmt"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"

	dxfmath "github.com/edanko/dxf/math"
)

// POLYLINE flag bits (group code 70).
const (
	PolylineFlagClosed           = 1
	PolylineFlagCurveFit         = 2
	PolylineFlagSplineFit        = 4
	PolylineFlag3DPolyline       = 8
	PolylineFlag3DPolygonMesh    = 16
	PolylineFlagMeshClosedN      = 32
	PolylineFlagPolyface         = 64
	PolylineFlagGenerateLinetype = 128
)

// Polyline represents a POLYLINE entity. The same record models plain
// polylines, 3D polygon meshes (POLYMESH) and polyface meshes; the Flag field
// selects the variant and MeshVertexCount/MeshFaceCount carry the 71/72 counts.
type Polyline struct {
	*entity
	Flag      int
	size      int
	Vertices  []*Vertex
	endhandle string

	// Mesh counts. For POLYMESH these are M/N vertex counts (71/72); for
	// POLYFACE they are vertex/face counts. SmoothDensity codes are 73/74 and
	// SmoothSurfaceType is 75.
	MeshVertexCount   int
	MeshFaceCount     int
	MSmoothDensity    int
	NSmoothDensity    int
	SmoothSurfaceType int
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
	switch {
	case p.IsPolyface():
		f.WriteString(100, "AcDbPolyFaceMesh")
	case p.IsPolymesh():
		f.WriteString(100, "AcDbPolygonMesh")
	default:
		f.WriteString(100, "AcDb3dPolyline")
	}
	f.WriteInt(66, 1)
	f.WriteString(10, "0.0")
	f.WriteString(20, "0.0")
	f.WriteString(30, "0.0")
	f.WriteInt(70, p.Flag)
	if p.MeshVertexCount != 0 {
		f.WriteInt(71, p.MeshVertexCount)
	}
	if p.MeshFaceCount != 0 {
		f.WriteInt(72, p.MeshFaceCount)
	}
	if p.MSmoothDensity != 0 {
		f.WriteInt(73, p.MSmoothDensity)
	}
	if p.NSmoothDensity != 0 {
		f.WriteInt(74, p.NSmoothDensity)
	}
	if p.SmoothSurfaceType != 0 {
		f.WriteInt(75, p.SmoothSurfaceType)
	}
	for _, v := range p.Vertices {
		v.Format(f)
	}
	f.WriteString(0, "SEQEND")
	if p.endhandle != "" {
		f.WriteString(5, p.endhandle)
	}
	f.WriteString(100, "AcDbEntity")
	if layer := p.Layer(); layer != nil {
		f.WriteString(8, layer.Name())
	} else if name := p.LayerName(); name != "" {
		f.WriteString(8, name)
	}
}

// IsPolyface reports whether the polyline is a POLYFACE mesh.
func (p *Polyline) IsPolyface() bool { return p.Flag&PolylineFlagPolyface != 0 }

// IsPolymesh reports whether the polyline is a 3D polygon mesh (POLYMESH).
func (p *Polyline) IsPolymesh() bool {
	return p.Flag&PolylineFlag3DPolygonMesh != 0 && p.Flag&PolylineFlagPolyface == 0
}

// FaceRecords returns the POLYFACE face-record vertices owned by the polyline.
func (p *Polyline) FaceRecords() []*Vertex {
	records := make([]*Vertex, 0)
	for _, v := range p.Vertices {
		if v.IsFaceRecord() {
			records = append(records, v)
		}
	}
	return records
}

// MeshVertices returns the polyface/polymesh vertex records (face records are
// excluded).
func (p *Polyline) MeshVertices() []*Vertex {
	vertices := make([]*Vertex, 0, len(p.Vertices))
	for _, v := range p.Vertices {
		if !v.IsFaceRecord() {
			vertices = append(vertices, v)
		}
	}
	return vertices
}

// Close closes Polyline.
func (p *Polyline) Close() {
	p.Flag |= 1
}

// AddVertex adds a new vertex to Polyline.
func (p *Polyline) AddVertex(x, y, z float64) *Vertex {
	return p.addVertex(NewVertex(x, y, z))
}

// AddPolyfaceVertex adds a polyface mesh vertex.
func (p *Polyline) AddPolyfaceVertex(x, y, z float64) *Vertex {
	return p.addVertex(NewPolyfaceVertex(x, y, z))
}

// AddFace adds a polyface face record with 1-based vertex indices.
func (p *Polyline) AddFace(indices ...int) *Vertex {
	return p.addVertex(NewFaceRecord(indices...))
}

func (p *Polyline) addVertex(v *Vertex) *Vertex {
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
	if p.endhandle == "" {
		p.endhandle = hg.Next()
	}
}

// SetEndHandleString sets the SEQEND handle loaded from a file.
func (p *Polyline) SetEndHandleString(h string) { p.endhandle = h }

// EndHandle returns the SEQEND handle.
func (p *Polyline) EndHandle() string { return p.endhandle }

// Subentities returns the vertex records owned by this polyline.
// Explode decomposes the polyline into LINE segments between consecutive
// vertices. Bulged segments are emitted as straight chords; a closed
// polyline adds the closing segment.
func (p *Polyline) Explode() ([]Entity, error) {
	if len(p.Vertices) < 2 {
		return nil, fmt.Errorf("polyline: fewer than two vertices")
	}
	var out []Entity
	for i := 0; i+1 < len(p.Vertices); i++ {
		out = append(out, p.lineSegment(p.Vertices[i], p.Vertices[i+1]))
	}
	if p.Flag&1 != 0 && len(p.Vertices) > 2 {
		out = append(out, p.lineSegment(p.Vertices[len(p.Vertices)-1], p.Vertices[0]))
	}
	return out, nil
}

func (p *Polyline) lineSegment(a, b *Vertex) Entity {
	la := NewLine()
	la.Start = append([]float64(nil), a.Coord...)
	la.End = append([]float64(nil), b.Coord...)
	la.SetLayer(p.Layer())
	return la
}

func (p *Polyline) Subentities() []Entity {
	subs := make([]Entity, 0, len(p.Vertices))
	for _, v := range p.Vertices {
		subs = append(subs, v)
	}
	return subs
}

func (p *Polyline) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for _, v := range p.Vertices {
		for i := range 3 {
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

// Transform applies an affine matrix to all polyline vertices.
func (p *Polyline) Transform(m dxfmath.Matrix44) error {
	for _, v := range p.Vertices {
		v.Coord = TransformPoint(m, v.Coord)
	}
	return nil
}

// Copy creates a deep copy of the Polyline entity
func (p *Polyline) Copy() Entity {
	polyline := NewPolyline()
	polyline.entity = p.entity.clone()
	polyline.Flag = p.Flag
	polyline.size = p.size
	polyline.endhandle = p.endhandle
	polyline.Vertices = make([]*Vertex, len(p.Vertices))
	for i, v := range p.Vertices {
		polyline.Vertices[i] = v.Copy().(*Vertex)
	}
	return polyline
}
