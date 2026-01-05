package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

type MeshVertexCache struct {
	*entity
	Vertices    []math.Vec3
	Normals     []math.Vec3
	TexCoords   []math.Vec2
	FaceIndices [][]int
}

func NewMeshVertexCache() *MeshVertexCache {
	return &MeshVertexCache{
		entity:      NewEntity(MESHVERTEXCACHE),
		Vertices:    make([]math.Vec3, 0),
		Normals:     make([]math.Vec3, 0),
		TexCoords:   make([]math.Vec2, 0),
		FaceIndices: make([][]int, 0),
	}
}

func (m *MeshVertexCache) IsEntity() bool {
	return true
}

func (m *MeshVertexCache) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbMeshVertexCache")
	f.WriteInt(90, len(m.Vertices))
	for _, v := range m.Vertices {
		f.WriteFloat(10, v.X())
		f.WriteFloat(20, v.Y())
		f.WriteFloat(30, v.Z())
	}
	if len(m.Normals) > 0 {
		f.WriteInt(91, len(m.Normals))
		for _, n := range m.Normals {
			f.WriteFloat(11, n.X())
			f.WriteFloat(21, n.Y())
			f.WriteFloat(31, n.Z())
		}
	}
	if len(m.TexCoords) > 0 {
		f.WriteInt(92, len(m.TexCoords))
		for _, t := range m.TexCoords {
			f.WriteFloat(12, t.X())
			f.WriteFloat(22, t.Y())
		}
	}
	if len(m.FaceIndices) > 0 {
		f.WriteInt(93, len(m.FaceIndices))
		for _, face := range m.FaceIndices {
			f.WriteInt(94, len(face))
			for _, idx := range face {
				f.WriteInt(95, idx)
			}
		}
	}
}

func (m *MeshVertexCache) BBox() ([]float64, []float64) {
	if len(m.Vertices) == 0 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}
	min := []float64{1e10, 1e10, 1e10}
	max := []float64{-1e10, -1e10, -1e10}
	for _, v := range m.Vertices {
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

func (m *MeshVertexCache) Copy() Entity {
	mvc := NewMeshVertexCache()
	mvc.entity = m.entity
	mvc.Vertices = make([]math.Vec3, len(m.Vertices))
	for i, v := range m.Vertices {
		mvc.Vertices[i] = v
	}
	mvc.Normals = make([]math.Vec3, len(m.Normals))
	for i, n := range m.Normals {
		mvc.Normals[i] = n
	}
	mvc.TexCoords = make([]math.Vec2, len(m.TexCoords))
	for i, t := range m.TexCoords {
		mvc.TexCoords[i] = t
	}
	mvc.FaceIndices = make([][]int, len(m.FaceIndices))
	for i, face := range m.FaceIndices {
		mvc.FaceIndices[i] = make([]int, len(face))
		for j, idx := range face {
			mvc.FaceIndices[i][j] = idx
		}
	}
	return mvc
}

func (m *MeshVertexCache) Validate() error {
	return nil
}

func (m *MeshVertexCache) AddVertex(v math.Vec3) {
	m.Vertices = append(m.Vertices, v)
}

func (m *MeshVertexCache) AddNormal(n math.Vec3) {
	m.Normals = append(m.Normals, n)
}

func (m *MeshVertexCache) AddTexCoord(t math.Vec2) {
	m.TexCoords = append(m.TexCoords, t)
}

func (m *MeshVertexCache) AddFace(indices []int) {
	face := make([]int, len(indices))
	for i, idx := range indices {
		face[i] = idx
	}
	m.FaceIndices = append(m.FaceIndices, face)
}

func (m *MeshVertexCache) Clear() {
	m.Vertices = make([]math.Vec3, 0)
	m.Normals = make([]math.Vec3, 0)
	m.TexCoords = make([]math.Vec2, 0)
	m.FaceIndices = make([][]int, 0)
}
