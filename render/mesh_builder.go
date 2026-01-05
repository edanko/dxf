package render

import (
	"fmt"

	dxfmath "github.com/edanko/dxf/math"
)

// MeshBuilder represents a 3D mesh builder for creating and manipulating 3D geometry
type MeshBuilder struct {
	vertices []dxfmath.Vec3
	faces    [][]int
	edges    map[[2]int]int
}

// NewMeshBuilder creates a new mesh builder
func NewMeshBuilder() *MeshBuilder {
	return &MeshBuilder{
		vertices: make([]dxfmath.Vec3, 0),
		faces:    make([][]int, 0),
		edges:    make(map[[2]int]int),
	}
}

// AddVertex adds a vertex to mesh and returns its index
func (mb *MeshBuilder) AddVertex(v dxfmath.Vec3) int {
	mb.vertices = append(mb.vertices, v)
	return len(mb.vertices) - 1
}

// AddFace adds a face to mesh using vertex indices
func (mb *MeshBuilder) AddFace(vertexIndices []int) error {
	if len(vertexIndices) < 3 {
		return fmt.Errorf("face must have at least 3 vertices")
	}

	// Validate vertex indices
	for _, idx := range vertexIndices {
		if idx < 0 || idx >= len(mb.vertices) {
			return fmt.Errorf("vertex index %d out of range [0, %d)", idx, len(mb.vertices))
		}
	}

	// Add face (make a copy to avoid modification)
	face := make([]int, len(vertexIndices))
	copy(face, vertexIndices)
	mb.faces = append(mb.faces, face)

	// Update edge statistics
	for i := 0; i < len(vertexIndices); i++ {
		v1 := vertexIndices[i]
		v2 := vertexIndices[(i+1)%len(vertexIndices)]

		// Create edge key with smaller index first
		if v1 > v2 {
			v1, v2 = v2, v1
		}
		edge := [2]int{v1, v2}
		mb.edges[edge]++
	}

	return nil
}

// AddTriangle adds a triangular face
func (mb *MeshBuilder) AddTriangle(v1, v2, v3 int) error {
	return mb.AddFace([]int{v1, v2, v3})
}

// AddQuad adds a quadrilateral face
func (mb *MeshBuilder) AddQuad(v1, v2, v3, v4 int) error {
	return mb.AddFace([]int{v1, v2, v3, v4})
}

// GetVertexCount returns number of vertices in mesh
func (mb *MeshBuilder) GetVertexCount() int {
	return len(mb.vertices)
}

// GetFaceCount returns number of faces in mesh
func (mb *MeshBuilder) GetFaceCount() int {
	return len(mb.faces)
}

// GetVertex returns a vertex by index
func (mb *MeshBuilder) GetVertex(index int) (dxfmath.Vec3, error) {
	if index < 0 || index >= len(mb.vertices) {
		return dxfmath.Vec3{}, fmt.Errorf("vertex index %d out of range", index)
	}
	return mb.vertices[index], nil
}

// GetFace returns a face by index
func (mb *MeshBuilder) GetFace(index int) ([]int, error) {
	if index < 0 || index >= len(mb.faces) {
		return nil, fmt.Errorf("face index %d out of range", index)
	}
	face := make([]int, len(mb.faces[index]))
	copy(face, mb.faces[index])
	return face, nil
}

// GetVertices returns all vertices
func (mb *MeshBuilder) GetVertices() []dxfmath.Vec3 {
	vertices := make([]dxfmath.Vec3, len(mb.vertices))
	copy(vertices, mb.vertices)
	return vertices
}

// GetFaces returns all faces
func (mb *MeshBuilder) GetFaces() [][]int {
	faces := make([][]int, len(mb.faces))
	for i, face := range mb.faces {
		faces[i] = make([]int, len(face))
		copy(faces[i], face)
	}
	return faces
}

// CalculateFaceNormal calculates normal vector for a face
func (mb *MeshBuilder) CalculateFaceNormal(faceIndex int) (dxfmath.Vec3, error) {
	face, err := mb.GetFace(faceIndex)
	if err != nil {
		return dxfmath.Vec3{}, err
	}

	if len(face) < 3 {
		return dxfmath.Vec3{}, fmt.Errorf("face must have at least 3 vertices")
	}

	// Get first three vertices
	v0 := mb.vertices[face[0]]
	v1 := mb.vertices[face[1]]
	v2 := mb.vertices[face[2]]

	// Calculate two edge vectors
	edge1 := v1.Sub(v0)
	edge2 := v2.Sub(v0)

	// Cross product gives normal
	normal := edge1.Cross(edge2)

	// Normalize
	length := normal.Length()
	if length > 0 {
		return normal.Mul(1.0 / length), nil
	}

	return dxfmath.Vec3{}, fmt.Errorf("degenerate face")
}

// CalculateBoundingBox calculates the bounding box of mesh
func (mb *MeshBuilder) CalculateBoundingBox() (min, max dxfmath.Vec3) {
	if len(mb.vertices) == 0 {
		return dxfmath.Vec3{}, dxfmath.Vec3{}
	}

	min = mb.vertices[0]
	max = mb.vertices[0]

	for _, v := range mb.vertices[1:] {
		if v.X() < min.X() {
			min = dxfmath.NewVec3(v.X(), min.Y(), min.Z())
		}
		if v.Y() < min.Y() {
			min = dxfmath.NewVec3(min.X(), v.Y(), min.Z())
		}
		if v.Z() < min.Z() {
			min = dxfmath.NewVec3(min.X(), min.Y(), v.Z())
		}
		if v.X() > max.X() {
			max = dxfmath.NewVec3(v.X(), max.Y(), max.Z())
		}
		if v.Y() > max.Y() {
			max = dxfmath.NewVec3(max.X(), v.Y(), max.Z())
		}
		if v.Z() > max.Z() {
			max = dxfmath.NewVec3(max.X(), max.Y(), v.Z())
		}
	}

	return min, max
}

// CalculateSurfaceArea calculates total surface area of mesh
func (mb *MeshBuilder) CalculateSurfaceArea() float64 {
	var totalArea float64

	for i := range mb.faces {
		area, err := mb.CalculateFaceArea(i)
		if err == nil {
			totalArea += area
		}
	}

	return totalArea
}

// CalculateFaceArea calculates area of a triangular face
func (mb *MeshBuilder) CalculateFaceArea(faceIndex int) (float64, error) {
	face, err := mb.GetFace(faceIndex)
	if err != nil {
		return 0, err
	}

	if len(face) < 3 {
		return 0, fmt.Errorf("face must have at least 3 vertices")
	}

	// For triangular faces, use cross product method
	if len(face) == 3 {
		v0 := mb.vertices[face[0]]
		v1 := mb.vertices[face[1]]
		v2 := mb.vertices[face[2]]

		edge1 := v1.Sub(v0)
		edge2 := v2.Sub(v0)

		// Area = 0.5 * |edge1 × edge2|
		cross := edge1.Cross(edge2)
		return 0.5 * cross.Length(), nil
	}

	// For non-triangular faces, triangulate
	// This is a simple fan triangulation from the first vertex
	var area float64
	v0 := mb.vertices[face[0]]

	for i := 1; i < len(face)-1; i++ {
		v1 := mb.vertices[face[i]]
		v2 := mb.vertices[face[i+1]]

		edge1 := v1.Sub(v0)
		edge2 := v2.Sub(v0)

		cross := edge1.Cross(edge2)
		area += 0.5 * cross.Length()
	}

	return area, nil
}

// Validate checks if mesh is valid
func (mb *MeshBuilder) Validate() error {
	if len(mb.vertices) == 0 {
		return fmt.Errorf("mesh has no vertices")
	}

	if len(mb.faces) == 0 {
		return fmt.Errorf("mesh has no faces")
	}

	// Check all face indices
	for faceIdx, face := range mb.faces {
		if len(face) < 3 {
			return fmt.Errorf("face %d has less than 3 vertices", faceIdx)
		}

		for vertexIdx, v := range face {
			if v < 0 || v >= len(mb.vertices) {
				return fmt.Errorf("face %d, vertex %d: index %d out of range", faceIdx, vertexIdx, v)
			}
		}
	}

	return nil
}

// Clear removes all vertices and faces from mesh
func (mb *MeshBuilder) Clear() {
	mb.vertices = mb.vertices[:0]
	mb.faces = mb.faces[:0]
	for k := range mb.edges {
		delete(mb.edges, k)
	}
}

// Clone creates a copy of the mesh builder
func (mb *MeshBuilder) Clone() *MeshBuilder {
	clone := NewMeshBuilder()

	// Copy vertices
	clone.vertices = make([]dxfmath.Vec3, len(mb.vertices))
	copy(clone.vertices, mb.vertices)

	// Copy faces
	clone.faces = make([][]int, len(mb.faces))
	for i, face := range mb.faces {
		clone.faces[i] = make([]int, len(face))
		copy(clone.faces[i], face)
	}

	// Copy edges
	clone.edges = make(map[[2]int]int)
	for k, v := range mb.edges {
		clone.edges[k] = v
	}

	return clone
}

// EdgeStatistics provides information about edge usage in the mesh
type EdgeStatistics struct {
	Edge     [2]int
	Count    int
	IsBorder bool
}

// GetEdgeStatistics returns statistics about edges in the mesh
func (mb *MeshBuilder) GetEdgeStatistics() []EdgeStatistics {
	stats := make([]EdgeStatistics, 0, len(mb.edges))

	for edge, count := range mb.edges {
		stats = append(stats, EdgeStatistics{
			Edge:     edge,
			Count:    count,
			IsBorder: count == 1, // Border edges are used by only one face
		})
	}

	return stats
}

// Transform applies a transformation matrix to all vertices
func (mb *MeshBuilder) Transform(matrix *dxfmath.Matrix44) error {
	for i := range mb.vertices {
		mb.vertices[i] = matrix.TransformVector(mb.vertices[i])
	}
	return nil
}

// Translate moves all vertices by the given offset
func (mb *MeshBuilder) Translate(offset dxfmath.Vec3) {
	for i := range mb.vertices {
		mb.vertices[i] = mb.vertices[i].Add(offset)
	}
}

// Scale scales all vertices by the given factors
func (mb *MeshBuilder) Scale(scaleX, scaleY, scaleZ float64) {
	for i := range mb.vertices {
		v := &mb.vertices[i]
		x, y, z := v.X(), v.Y(), v.Z()
		*v = dxfmath.NewVec3(x*scaleX, y*scaleY, z*scaleZ)
	}
}

// UniformScale scales all vertices by the same factor
func (mb *MeshBuilder) UniformScale(scale float64) {
	mb.Scale(scale, scale, scale)
}

// Merge combines another mesh builder into this one
func (mb *MeshBuilder) Merge(other *MeshBuilder) error {
	if other == nil {
		return nil
	}

	vertexOffset := len(mb.vertices)

	// Add vertices
	for _, v := range other.vertices {
		mb.AddVertex(v)
	}

	// Add faces with adjusted vertex indices
	for _, face := range other.faces {
		adjustedFace := make([]int, len(face))
		for i, v := range face {
			adjustedFace[i] = v + vertexOffset
		}
		mb.AddFace(adjustedFace)
	}

	return nil
}
