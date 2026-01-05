package math

import (
	"math"
)

// Mesh represents a 3D mesh with faces and normals
type Mesh struct {
	Vertices []Vec3
	Faces    [][]int // Indices into Vertices array
	Normals  []Vec3  // Per-vertex normals
	IsClosed bool    // Whether mesh is closed
}

// MeshBuilder provides methods for constructing and manipulating 3D meshes
type MeshBuilder struct {
	mesh Mesh
}

// NewMeshBuilder creates a new mesh builder
func NewMeshBuilder() *MeshBuilder {
	return &MeshBuilder{
		mesh: Mesh{
			Vertices: make([]Vec3, 0),
			Faces:    make([][]int, 0),
			Normals:  make([]Vec3, 0),
			IsClosed: false,
		},
	}
}

// AddVertex adds a vertex to the mesh
func (mb *MeshBuilder) AddVertex(vertex Vec3) int {
	mb.mesh.Vertices = append(mb.mesh.Vertices, vertex)
	return len(mb.mesh.Vertices) - 1
}

// AddTriangle adds a triangular face to the mesh
func (mb *MeshBuilder) AddTriangle(v1, v2, v3 int) {
	face := []int{v1, v2, v3}
	mb.mesh.Faces = append(mb.mesh.Faces, face)
}

// AddQuad adds a quadrilateral face to the mesh
func (mb *MeshBuilder) AddQuad(v1, v2, v3, v4 int) {
	face := []int{v1, v2, v3, v4}
	mb.mesh.Faces = append(mb.mesh.Faces, face)
}

// Build finalizes the mesh and returns the result
func (mb *MeshBuilder) Build() Mesh {
	mb.calculateNormals()
	return mb.mesh
}

// calculateNormals computes per-vertex normals for the mesh
func (mb *MeshBuilder) calculateNormals() {
	if len(mb.mesh.Vertices) == 0 {
		return
	}

	// Initialize normals to zero
	mb.mesh.Normals = make([]Vec3, len(mb.mesh.Vertices))
	for i := range mb.mesh.Normals {
		mb.mesh.Normals[i] = NewVec3(0, 0, 0)
	}

	// Calculate face normals and accumulate
	for _, face := range mb.mesh.Faces {
		if len(face) < 3 {
			continue
		}

		// Get face vertices
		v0 := mb.mesh.Vertices[face[0]]
		v1 := mb.mesh.Vertices[face[1]]
		v2 := mb.mesh.Vertices[face[2]]

		// Calculate face normal using cross product
		edge1 := v1.Sub(v0)
		edge2 := v2.Sub(v0)
		faceNormal := edge1.Cross(edge2).Normalized()

		// Accumulate normals for each vertex
		mb.mesh.Normals[face[0]] = mb.mesh.Normals[face[0]].Add(faceNormal)
		mb.mesh.Normals[face[1]] = mb.mesh.Normals[face[1]].Add(faceNormal)
		mb.mesh.Normals[face[2]] = mb.mesh.Normals[face[2]].Add(faceNormal)

		// Handle quad faces
		if len(face) == 4 {
			v3 := mb.mesh.Vertices[face[3]]

			// Second triangle of quad
			edge1 := v2.Sub(v3)
			edge2 := v0.Sub(v3)
			faceNormal2 := edge1.Cross(edge2).Normalized()

			mb.mesh.Normals[face[0]] = mb.mesh.Normals[face[0]].Add(faceNormal2)
			mb.mesh.Normals[face[1]] = mb.mesh.Normals[face[1]].Add(faceNormal2)
			mb.mesh.Normals[face[2]] = mb.mesh.Normals[face[2]].Add(faceNormal2)
			mb.mesh.Normals[face[3]] = mb.mesh.Normals[face[3]].Add(faceNormal2)
		}
	}

	// Normalize accumulated normals
	for i := range mb.mesh.Normals {
		if !mb.mesh.Normals[i].IsZero(1e-6) {
			mb.mesh.Normals[i] = mb.mesh.Normals[i].Normalized()
		}
	}
}

// SetClosed marks the mesh as closed
func (mb *MeshBuilder) SetClosed(closed bool) {
	mb.mesh.IsClosed = closed
}

// CalculateVolume calculates volume of a closed mesh using divergence theorem
func (m *Mesh) CalculateVolume() float64 {
	if !m.IsClosed || len(m.Faces) == 0 {
		return 0.0
	}

	volume := 0.0

	for _, face := range m.Faces {
		if len(face) < 3 {
			continue
		}

		// Get face vertices
		v0 := m.Vertices[face[0]]
		v1 := m.Vertices[face[1]]
		v2 := m.Vertices[face[2]]

		// Calculate tetrahedron volume with origin
		// V = (1/6) * |(v1-v0) · ((v2-v0) × (v1-v0))|
		edge1 := v1.Sub(v0)
		edge2 := v2.Sub(v0)
		crossProduct := edge1.Cross(edge2)
		dotProduct := edge1.Dot(crossProduct)

		faceVolume := math.Abs(dotProduct) / 6.0
		volume += faceVolume
	}

	return volume
}

// IsPointInside checks if a point is inside the mesh
func (m *Mesh) IsPointInside(point Vec3) bool {
	if !m.IsClosed {
		return false
	}

	// Simplified point-in-mesh test using ray casting
	intersections := 0

	for _, face := range m.Faces {
		if len(face) < 3 {
			continue
		}

		// Get face vertices
		v0 := m.Vertices[face[0]]

		// Check if point is on same side of all face edges
		signs := make([]int, 3)
		for i := 0; i < 3; i++ {
			var next Vec3
			if i == 2 {
				next = v0
			} else {
				next = m.Vertices[face[i+1]]
			}

			edge := next.Sub(m.Vertices[face[i]])
			normal := edge.Cross(point.Sub(m.Vertices[face[i]]))
			signs[i] = sign(normal.Z())
		}

		// Point is inside if all signs are the same (or zero)
		allSame := true
		for i := 1; i < 3; i++ {
			if signs[i] == 0 || (signs[i] != signs[0] && signs[0] != 0) {
				allSame = false
				break
			}
		}

		if allSame {
			intersections++
		}
	}

	// Point is inside if intersections count is odd
	return intersections%2 == 1
}

// Subdivide subdivides the mesh by adding midpoints
func (m *Mesh) Subdivide(iterations int) Mesh {
	if iterations < 1 {
		return *m
	}

	result := NewMeshBuilder()

	// Simple subdivision - split each triangle into 4 smaller triangles
	for _, face := range m.Faces {
		if len(face) < 3 {
			continue
		}

		v0 := m.Vertices[face[0]]
		v1 := m.Vertices[face[1]]
		v2 := m.Vertices[face[2]]

		// Calculate midpoints
		mid01 := v0.Add(v1).Mul(0.5)
		mid12 := v1.Add(v2).Mul(0.5)
		mid02 := v0.Add(v2).Mul(0.5)
		mid012 := mid01.Add(mid12).Mul(0.5)

		// Add original vertex
		i0 := result.AddVertex(v0)
		i1 := result.AddVertex(v1)
		i2 := result.AddVertex(v2)

		// Add new vertices
		i01 := result.AddVertex(mid01)
		_ = result.AddVertex(mid12) // i12 - unused but needed for symmetry
		i02 := result.AddVertex(mid02)
		i012 := result.AddVertex(mid012)

		// Create 4 new triangles
		result.AddTriangle(i0, i01, i02)
		result.AddTriangle(i01, i1, i012)
		result.AddTriangle(i012, i02, i2)
		result.AddTriangle(i0, i02, i012)
	}

	if m.IsClosed {
		result.SetClosed(true)
	}

	return result.Build()
}

// Smooth applies Laplacian smoothing to the mesh
func (m *Mesh) Smooth(iterations int, lambda float64) Mesh {
	if iterations < 1 {
		return *m
	}

	result := Mesh{
		Vertices: make([]Vec3, len(m.Vertices)),
		Faces:    make([][]int, len(m.Faces)),
		Normals:  make([]Vec3, len(m.Vertices)),
		IsClosed: m.IsClosed,
	}

	// Copy original vertices and faces
	copy(result.Vertices, m.Vertices)
	copy(result.Faces, m.Faces)
	copy(result.Normals, m.Normals)

	// Apply Laplacian smoothing
	for iter := 0; iter < iterations; iter++ {
		newVertices := make([]Vec3, len(result.Vertices))
		copy(newVertices, result.Vertices)

		for i, vertex := range result.Vertices {
			// Find adjacent vertices (simplified - assumes all vertices are connected)
			var neighbors []Vec3
			var neighborCount int

			for _, face := range result.Faces {
				if len(face) < 3 {
					continue
				}

				// Check if vertex is part of this face
				for _, faceVertex := range face {
					if faceVertex == i {
						// Add other vertices of this face as neighbors
						for _, otherVertex := range face {
							if otherVertex != i {
								// Check if this neighbor is already added
								found := false
								for _, existing := range neighbors {
									if existing.IsEqual(result.Vertices[otherVertex], 1e-6) {
										found = true
										break
									}
								}
								if !found {
									neighbors = append(neighbors, result.Vertices[otherVertex])
									neighborCount++
								}
							}
						}
					}
				}
			}

			// Apply Laplacian smoothing: v' = v + λ * (1/n) * Σ(vi - v)
			if neighborCount > 0 {
				var neighborSum Vec3
				for _, neighbor := range neighbors {
					neighborSum = neighborSum.Add(neighbor)
				}

				averageNeighbor := neighborSum.Div(float64(neighborCount))
				displacement := averageNeighbor.Sub(vertex).Mul(lambda)
				newVertices[i] = vertex.Add(displacement)
			} else {
				newVertices[i] = vertex
			}
		}

		result.Vertices = newVertices
	}

	// Recalculate normals
	meshBuilder := NewMeshBuilder()
	for _, vertex := range result.Vertices {
		meshBuilder.AddVertex(vertex)
	}
	for _, face := range result.Faces {
		if len(face) >= 3 {
			meshBuilder.AddTriangle(face[0], face[1], face[2])
		}
	}
	result = meshBuilder.Build()

	return result
}

// Helper functions
func sign(value float64) int {
	if value > 0 {
		return 1
	} else if value < 0 {
		return -1
	}
	return 0
}
