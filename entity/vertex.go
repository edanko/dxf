package entity

import (
	"github.com/edanko/dxf/format"
)

// VERTEX flag bits (group code 70).
const (
	VertexFlagExtraCreated       = 1   // Extra vertex created by curve-fitting
	VertexFlagCurveFitTangent    = 2   // Curve-fit tangent defined
	VertexFlagSplineCreated      = 8   // Spline vertex created by spline-fitting
	VertexFlagSplineFrameControl = 16  // Spline frame control point
	VertexFlag3DPolyline         = 32  // 3D polyline vertex
	VertexFlag3DPolygonMesh      = 64  // 3D polygon mesh vertex
	VertexFlag3DPolyfaceMesh     = 128 // Polyface mesh face record
)

// Vertex represents a VERTEX entity. A vertex belongs to a POLYLINE owner;
// code 330 (owner) is carried on the shared base.
//
// For POLYFACE meshes the same record type carries two variants:
//
//   - mesh vertices have VertexFlag3DPolygonMesh set and a real Coord;
//   - face records have VertexFlag3DPolyfaceMesh set, a zero Coord and up to
//     four 1-based vertex indices in FaceIndices (71..74). A negative index
//     marks the start of an invisible edge.
//
// POLYMESH (3D polygon mesh) vertices use VertexFlag3DPolygonMesh without the
// face-record bit. Use IsFaceRecord and IsPolyfaceVertex to distinguish them.
type Vertex struct {
	*entity
	Flag             int       // 70
	Coord            []float64 // 10, 20, 30
	Bulge            float64   // 42
	StartWidth       float64   // 40
	EndWidth         float64   // 41
	Tangent          float64   // 50, curve-fit tangent direction in degrees
	FaceIndices      [4]int    // 71, 72, 73, 74 (polyface face record indices)
	VertexIdentifier int       // 91
}

// IsEntity is for Entity interface.
func (v *Vertex) IsEntity() bool {
	return true
}

// NewVertex creates a new Vertex.
func NewVertex(x, y, z float64) *Vertex {
	v := &Vertex{
		entity: NewEntity(VERTEX),
		Flag:   VertexFlag3DPolyline,
		Coord:  []float64{x, y, z},
	}
	return v
}

// NewPolyfaceVertex creates a polyface mesh vertex record.
func NewPolyfaceVertex(x, y, z float64) *Vertex {
	v := NewVertex(x, y, z)
	v.Flag = VertexFlag3DPolygonMesh
	return v
}

// NewPolymeshVertex creates a 3D polygon mesh vertex record. POLYMESH and
// POLYFACE mesh vertices share the same flag value.
func NewPolymeshVertex(x, y, z float64) *Vertex {
	return NewPolyfaceVertex(x, y, z)
}

// NewFaceRecord creates a polyface face record. indices are 1-based vertex
// indices; a negative value marks the start of an invisible edge.
func NewFaceRecord(indices ...int) *Vertex {
	v := NewVertex(0, 0, 0)
	v.Flag = VertexFlag3DPolygonMesh | VertexFlag3DPolyfaceMesh
	for i := 0; i < len(indices) && i < len(v.FaceIndices); i++ {
		v.FaceIndices[i] = indices[i]
	}
	return v
}

// IsPolyfaceVertex reports whether the record is a POLYFACE mesh vertex.
func (v *Vertex) IsPolyfaceVertex() bool {
	return v.Flag&VertexFlag3DPolygonMesh != 0 && v.Flag&VertexFlag3DPolyfaceMesh == 0
}

// IsPolymeshVertex reports whether the record is a POLYMESH (3D polygon mesh)
// vertex. POLYMESH and POLYFACE vertices share the same wire flag; the owning
// POLYLINE flag distinguishes the structures.
func (v *Vertex) IsPolymeshVertex() bool {
	return v.Flag&VertexFlag3DPolygonMesh != 0 && v.Flag&VertexFlag3DPolyfaceMesh == 0
}

// IsFaceRecord reports whether the record is a POLYFACE face record.
func (v *Vertex) IsFaceRecord() bool {
	return v.Flag&VertexFlag3DPolyfaceMesh != 0
}

// FaceVertexIndices returns the non-zero 1-based indices stored in a face
// record. Negative values indicate the start of an invisible edge.
func (v *Vertex) FaceVertexIndices() []int {
	indices := make([]int, 0, len(v.FaceIndices))
	for _, index := range v.FaceIndices {
		if index != 0 {
			indices = append(indices, index)
		}
	}
	return indices
}

// SetFaceIndices replaces the face record indices (71..74).
func (v *Vertex) SetFaceIndices(indices ...int) {
	v.FaceIndices = [4]int{}
	for i := 0; i < len(indices) && i < len(v.FaceIndices); i++ {
		v.FaceIndices[i] = indices[i]
	}
}

// Format writes data to formatter.
func (v *Vertex) Format(f format.Formatter) {
	v.entity.Format(f)
	f.WriteString(100, "AcDbVertex")
	if v.IsFaceRecord() {
		f.WriteString(100, "AcDbFaceRecord")
	} else if v.Flag&VertexFlag3DPolygonMesh != 0 {
		f.WriteString(100, "AcDbPolygonMeshVertex")
	} else {
		f.WriteString(100, "AcDb3dPolylineVertex")
	}
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
	if v.Tangent != 0 {
		f.WriteFloat(50, v.Tangent)
	}
	for i, index := range v.FaceIndices {
		if index != 0 {
			f.WriteInt(71+i, index)
		}
	}
	if v.VertexIdentifier != 0 {
		f.WriteInt(91, v.VertexIdentifier)
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
	vertex.Tangent = v.Tangent
	vertex.FaceIndices = v.FaceIndices
	vertex.VertexIdentifier = v.VertexIdentifier
	return vertex
}
