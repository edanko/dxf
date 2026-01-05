package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// MESH entity flags
const (
	MeshFlagNoSmooth    = 0 // No vertex smoothing
	MeshFlagQuadSmooth  = 1 // Quadratic smoothing
	MeshFlagCubicSmooth = 2 // Cubic smoothing
	MeshFlagNoColor     = 0 // No vertex colors
	MeshFlagHasColor    = 4 // Vertex colors present
)

// MESH represents a DXF MESH entity (polygon mesh)
type Mesh struct {
	*entity

	// Basic properties
	version           int // 71 - Mesh version (default=2)
	blendCrease       int // 72 - Smoothing type (0=none, 1=quad, 2=cubic)
	subdivisionLevels int // 91 - Subdivision levels (default=0)

	// Vertex data
	vertices []dxfmath.Vec3 // Vertex positions

	// Face data
	faces [][]int // Face vertex indices (triangles or polygons)

	// Optional edge data (for wireframe display)
	edges [][]int // Edge vertex indices (optional)

	// Optional vertex colors
	vertexColors []int // Vertex colors (optional)

	// Optional edge crease values
	creaseValues []float64 // Edge crease weights (optional)

	// Optional property overrides
	color int    // Vertex color override (62)
	layer string // Layer override (8)
}

// NewMesh creates a new MESH entity
func NewMesh() *Mesh {
	return &Mesh{
		entity:            NewEntity(MESH),
		version:           2,
		blendCrease:       0,
		subdivisionLevels: 0,
		vertices:          make([]dxfmath.Vec3, 0),
		faces:             make([][]int, 0),
		edges:             make([][]int, 0),
		vertexColors:      make([]int, 0),
		creaseValues:      make([]float64, 0),
		color:             0,
		layer:             "",
	}
}

// IsEntity returns true for MESH entities
func (m *Mesh) IsEntity() bool {
	return true
}

// Format writes MESH entity data to DXF format
func (m *Mesh) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbSubDMesh")

	// Write mesh version
	if m.version != 2 {
		f.WriteInt(71, m.version)
	}

	// Write blend crease settings
	if m.blendCrease != 0 {
		f.WriteInt(72, m.blendCrease)
	}

	// Write subdivision levels
	if m.subdivisionLevels != 0 {
		f.WriteInt(91, m.subdivisionLevels)
	}

	// Write vertex count and position data
	if len(m.vertices) > 0 {
		f.WriteInt(92, len(m.vertices))

		// Write vertex positions (10,20,30)
		for _, vertex := range m.vertices {
			f.WriteFloat(10, vertex.X())
			f.WriteFloat(20, vertex.Y())
			f.WriteFloat(30, vertex.Z())
		}
	}

	// Write face data
	if len(m.faces) > 0 {
		f.WriteInt(93, len(m.faces))

		// Write face vertex indices (90)
		for _, face := range m.faces {
			if len(face) >= 3 && len(face) <= 4 {
				// Write face with group code 90
				for _, vertexIndex := range face {
					f.WriteInt(90, vertexIndex)
				}
			}
		}
	}

	// Write edge data (wireframe)
	if len(m.edges) > 0 {
		f.WriteInt(94, len(m.edges))

		// Write edge vertex indices (90)
		for _, edge := range m.edges {
			if len(edge) >= 2 {
				// Write edge with group code 90
				for _, vertexIndex := range edge {
					f.WriteInt(90, vertexIndex)
				}
			}
		}
	}

	// Write vertex colors
	if len(m.vertexColors) > 0 {
		// Enable color flag
		if m.color == 0 {
			m.color = 62 // Enable vertex colors
		}

		// Write vertex colors (90, 92)
		for i, colorIndex := range m.vertexColors {
			if i < len(m.vertices) {
				f.WriteInt(90, i)          // Vertex index
				f.WriteInt(92, colorIndex) // Color index
			}
		}
	}

	// Write edge crease values
	if len(m.creaseValues) > 0 {
		// Write crease values (95, 140)
		for _, creaseValue := range m.creaseValues {
			f.WriteFloat(95, creaseValue)
		}
	}

	// Write layer override
	if m.layer != "" {
		f.WriteString(8, m.layer)
	}

	// Write color override
	if m.color != 0 {
		f.WriteInt(62, m.color)
	}
}

// SetVertices sets the mesh vertices
func (m *Mesh) SetVertices(vertices []dxfmath.Vec3) {
	m.vertices = make([]dxfmath.Vec3, len(vertices))
	for i, vertex := range vertices {
		m.vertices[i] = vertex
	}
}

// AddVertex adds a single vertex to the mesh
func (m *Mesh) AddVertex(vertex dxfmath.Vec3) {
	m.vertices = append(m.vertices, vertex)
}

// GetVertex returns a vertex at the given index
func (m *Mesh) GetVertex(index int) (dxfmath.Vec3, bool) {
	if index >= 0 && index < len(m.vertices) {
		return m.vertices[index], true
	}
	return dxfmath.Vec3{}, false
}

// VertexCount returns the number of vertices
func (m *Mesh) VertexCount() int {
	return len(m.vertices)
}

// SetFaces sets the mesh faces (triangles or polygons)
func (m *Mesh) SetFaces(faces [][]int) {
	m.faces = make([][]int, len(faces))
	for i, face := range faces {
		m.faces[i] = make([]int, len(face))
		for j, val := range face {
			m.faces[i][j] = val
		}
	}
}

// AddFace adds a single face (triangle or polygon)
func (m *Mesh) AddFace(face []int) {
	m.faces = append(m.faces, face)
}

// GetFace returns a face at the given index
func (m *Mesh) GetFace(index int) ([]int, bool) {
	if index >= 0 && index < len(m.faces) {
		return m.faces[index], true
	}
	return []int{}, false
}

// FaceCount returns the number of faces
func (m *Mesh) FaceCount() int {
	return len(m.faces)
}

// SetVertexColors sets vertex colors
func (m *Mesh) SetVertexColors(colors []int) {
	m.vertexColors = make([]int, len(colors))
	for i, color := range colors {
		m.vertexColors[i] = color
	}
}

// SetEdges sets mesh edges (wireframe)
func (m *Mesh) SetEdges(edges [][]int) {
	m.edges = make([][]int, len(edges))
	for i, edge := range edges {
		m.edges[i] = make([]int, len(edge))
		for j, val := range edge {
			m.edges[i][j] = val
		}
	}
}

// AddEdge adds a single edge
func (m *Mesh) AddEdge(edge []int) {
	m.edges = append(m.edges, edge)
}

// EdgeCount returns the number of edges
func (m *Mesh) EdgeCount() int {
	return len(m.edges)
}

// SetCreaseValues sets edge crease values
func (m *Mesh) SetCreaseValues(values []float64) {
	m.creaseValues = make([]float64, len(values))
	for i, val := range values {
		m.creaseValues[i] = val
	}
}

// Clear removes all geometry from the mesh
func (m *Mesh) Clear() {
	m.vertices = make([]dxfmath.Vec3, 0)
	m.faces = make([][]int, 0)
	m.edges = make([][]int, 0)
	m.vertexColors = make([]int, 0)
	m.creaseValues = make([]float64, 0)
}

// IsEmpty returns true if the mesh has no geometry
func (m *Mesh) IsEmpty() bool {
	return len(m.vertices) == 0 && len(m.faces) == 0
}

// BBox calculates bounding box of the mesh
func (m *Mesh) BBox() ([]float64, []float64) {
	if m.IsEmpty() {
		return []float64{}, []float64{}
	}

	minX, minY, minZ := m.vertices[0].X(), m.vertices[0].Y(), m.vertices[0].Z()
	maxX, maxY, maxZ := m.vertices[0].X(), m.vertices[0].Y(), m.vertices[0].Z()

	for _, vertex := range m.vertices[1:] {
		if vertex.X() < minX {
			minX = vertex.X()
		}
		if vertex.Y() < minY {
			minY = vertex.Y()
		}
		if vertex.Z() < minZ {
			minZ = vertex.Z()
		}
		if vertex.X() > maxX {
			maxX = vertex.X()
		}
		if vertex.Y() > maxY {
			maxY = vertex.Y()
		}
		if vertex.Z() > maxZ {
			maxZ = vertex.Z()
		}
	}

	return []float64{minX, minY, minZ}, []float64{maxX, maxY, maxZ}
}

// CreateCube creates a simple cube mesh
func CreateCube(size float64) *Mesh {
	mesh := NewMesh()
	halfSize := size / 2.0

	// Define cube vertices
	vertices := []dxfmath.Vec3{
		// Front face (+Z)
		dxfmath.NewVec3(-halfSize, -halfSize, halfSize), // 0
		dxfmath.NewVec3(halfSize, -halfSize, halfSize),  // 1
		dxfmath.NewVec3(halfSize, halfSize, halfSize),   // 2
		dxfmath.NewVec3(-halfSize, halfSize, halfSize),  // 3

		// Back face (-Z)
		dxfmath.NewVec3(-halfSize, -halfSize, -halfSize), // 4
		dxfmath.NewVec3(halfSize, -halfSize, -halfSize),  // 5
		dxfmath.NewVec3(halfSize, halfSize, -halfSize),   // 6
		dxfmath.NewVec3(-halfSize, halfSize, -halfSize),  // 7
		dxfmath.NewVec3(-halfSize, halfSize, -halfSize),  // 8
	}

	// Define cube faces (2 triangles per face, 12 faces total)
	faces := [][]int{
		// Front face (+Z)
		{0, 1, 2}, {0, 2, 3}, {0, 3, 1},
		// Right face (+X)
		{1, 5, 6}, {1, 6, 7}, {1, 7, 5},
		// Back face (-Z)
		{4, 7, 6}, {4, 6, 7}, {4, 7, 5},
		// Left face (-X)
		{0, 3, 7}, {0, 7, 4}, {0, 4, 3},
		// Top face (+Y)
		{3, 2, 7}, {3, 7, 2}, {3, 7, 2},
		// Bottom face (-Y)
		{4, 7, 6}, {4, 7, 0}, {4, 7, 6},
	}

	mesh.SetVertices(vertices)
	mesh.SetFaces(faces)
	return mesh
}

// CreatePyramid creates a simple 4-sided pyramid mesh
func CreatePyramid(baseSize, height float64) *Mesh {
	mesh := NewMesh()
	halfBase := baseSize / 2.0

	// Define pyramid vertices
	vertices := []dxfmath.Vec3{
		// Base vertices
		dxfmath.NewVec3(-halfBase, -halfBase, 0), // 0
		dxfmath.NewVec3(halfBase, -halfBase, 0),  // 1
		dxfmath.NewVec3(halfBase, halfBase, 0),   // 2
		dxfmath.NewVec3(halfBase, halfBase, 0),   // 3
		// Apex vertex
		dxfmath.NewVec3(0, 0, height), // 4
	}

	// Define pyramid faces (4 triangular faces)
	faces := [][]int{
		{0, 1, 4}, // Front
		{1, 2, 4}, // Right
		{2, 3, 4}, // Back
		{3, 0, 4}, // Left
	}

	mesh.SetVertices(vertices)
	mesh.SetFaces(faces)
	return mesh
}

// CreateSphere creates a simple sphere mesh using subdivision
func CreateSphere(radius float64, segments int) *Mesh {
	mesh := NewMesh()

	// Generate sphere vertices
	vertices := make([]dxfmath.Vec3, 0)
	faces := make([][]int, 0)

	// Simple sphere implementation using latitude/longitude
	for lat := 0; lat <= segments; lat++ {
		theta := float64(lat) * 3.14159 / float64(segments)
		for lon := 0; lon <= segments; lon++ {
			phi := float64(lon) * 2.0 * 3.14159 / float64(segments)

			x := radius * math.Sin(theta) * math.Sin(phi)
			y := radius * math.Cos(theta) * math.Sin(phi)
			z := radius * math.Cos(theta)

			vertices = append(vertices, dxfmath.NewVec3(x, y, z))
		}
	}

	// Generate triangular faces
	for lat := 0; lat < segments; lat++ {
		for lon := 0; lon < segments; lon++ {
			// Current row
			current := lat * (segments + 1)
			next := current + (segments + 1)

			// Next row
			nextNext := next + (segments + 1)

			// Triangle 1
			faces = append(faces, []int{current, next, nextNext})
			// Triangle 2
			if lon < segments-1 {
				faces = append(faces, []int{next, nextNext, (nextNext + 1)})
			}
		}
	}

	mesh.SetVertices(vertices)
	mesh.SetFaces(faces)
	return mesh
}
