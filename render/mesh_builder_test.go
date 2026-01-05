package render

import (
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func TestMeshBuilder(t *testing.T) {
	mb := NewMeshBuilder()

	// Test adding vertices
	v1 := dxfmath.NewVec3(0, 0, 0)
	v2 := dxfmath.NewVec3(1, 0, 0)
	v3 := dxfmath.NewVec3(0, 1, 0)

	i1 := mb.AddVertex(v1)
	i2 := mb.AddVertex(v2)
	i3 := mb.AddVertex(v3)

	if i1 != 0 || i2 != 1 || i3 != 2 {
		t.Errorf("Expected vertex indices 0,1,2, got %d,%d,%d", i1, i2, i3)
	}

	// Test adding a face
	err := mb.AddFace([]int{i1, i2, i3})
	if err != nil {
		t.Errorf("Error adding face: %v", err)
	}

	// Test mesh validation
	err = mb.Validate()
	if err != nil {
		t.Errorf("Mesh validation failed: %v", err)
	}

	// Test counts
	if mb.GetVertexCount() != 3 {
		t.Errorf("Expected 3 vertices, got %d", mb.GetVertexCount())
	}

	if mb.GetFaceCount() != 1 {
		t.Errorf("Expected 1 face, got %d", mb.GetFaceCount())
	}

	// Test face normal calculation
	normal, err := mb.CalculateFaceNormal(0)
	if err != nil {
		t.Errorf("Error calculating face normal: %v", err)
	}

	// For triangle (0,0,0), (1,0,0), (0,1,0), normal should be (0,0,1)
	expectedNormal := dxfmath.NewVec3(0, 0, 1)

	if !normal.IsEqual(expectedNormal, 1e-10) {
		t.Errorf("Expected normal %v, got %v", expectedNormal, normal)
	}

	// Test bounding box
	min, max := mb.CalculateBoundingBox()
	expectedMin := dxfmath.NewVec3(0, 0, 0)
	expectedMax := dxfmath.NewVec3(1, 1, 0)

	if !min.IsEqual(expectedMin, 1e-10) || !max.IsEqual(expectedMax, 1e-10) {
		t.Errorf("Expected bounding box min %v, max %v, got min %v, max %v",
			expectedMin, expectedMax, min, max)
	}
}

func TestMeshBuilderArea(t *testing.T) {
	mb := NewMeshBuilder()

	// Create a unit triangle
	v1 := dxfmath.NewVec3(0, 0, 0)
	v2 := dxfmath.NewVec3(1, 0, 0)
	v3 := dxfmath.NewVec3(0, 1, 0)

	i1 := mb.AddVertex(v1)
	i2 := mb.AddVertex(v2)
	i3 := mb.AddVertex(v3)

	mb.AddFace([]int{i1, i2, i3})

	// Test face area
	area, err := mb.CalculateFaceArea(0)
	if err != nil {
		t.Errorf("Error calculating face area: %v", err)
	}

	expectedArea := 0.5 // 0.5 * |cross((1,0,0), (0,1,0))| = 0.5 * 1
	if area != expectedArea {
		t.Errorf("Expected area %f, got %f", expectedArea, area)
	}

	// Test surface area
	surfaceArea := mb.CalculateSurfaceArea()
	if surfaceArea != expectedArea {
		t.Errorf("Expected surface area %f, got %f", expectedArea, surfaceArea)
	}
}

func TestMeshBuilderQuad(t *testing.T) {
	mb := NewMeshBuilder()

	// Create a quad (two triangles)
	v1 := dxfmath.NewVec3(0, 0, 0)
	v2 := dxfmath.NewVec3(1, 0, 0)
	v3 := dxfmath.NewVec3(1, 1, 0)
	v4 := dxfmath.NewVec3(0, 1, 0)

	i1 := mb.AddVertex(v1)
	i2 := mb.AddVertex(v2)
	i3 := mb.AddVertex(v3)
	i4 := mb.AddVertex(v4)

	// Add as quad
	err := mb.AddQuad(i1, i2, i3, i4)
	if err != nil {
		t.Errorf("Error adding quad: %v", err)
	}

	if mb.GetFaceCount() != 1 {
		t.Errorf("Expected 1 face, got %d", mb.GetFaceCount())
	}

	// Test face area (quad area = 1)
	area, err := mb.CalculateFaceArea(0)
	if err != nil {
		t.Errorf("Error calculating quad area: %v", err)
	}

	expectedArea := 1.0
	if area != expectedArea {
		t.Errorf("Expected quad area %f, got %f", expectedArea, area)
	}
}

func TestMeshBuilderEdgeStatistics(t *testing.T) {
	mb := NewMeshBuilder()

	// Create a triangle
	v1 := dxfmath.NewVec3(0, 0, 0)
	v2 := dxfmath.NewVec3(1, 0, 0)
	v3 := dxfmath.NewVec3(0, 1, 0)

	i1 := mb.AddVertex(v1)
	i2 := mb.AddVertex(v2)
	i3 := mb.AddVertex(v3)

	mb.AddTriangle(i1, i2, i3)

	stats := mb.GetEdgeStatistics()

	// Triangle has 3 edges
	if len(stats) != 3 {
		t.Errorf("Expected 3 edge statistics, got %d", len(stats))
	}

	// All edges should be border edges (used by only one face)
	borderCount := 0
	for _, stat := range stats {
		if stat.IsBorder {
			borderCount++
		}
	}

	if borderCount != 3 {
		t.Errorf("Expected 3 border edges, got %d", borderCount)
	}
}

func TestMeshBuilderTransform(t *testing.T) {
	mb := NewMeshBuilder()

	// Create a triangle
	v1 := dxfmath.NewVec3(0, 0, 0)
	v2 := dxfmath.NewVec3(1, 0, 0)
	v3 := dxfmath.NewVec3(0, 1, 0)

	i1 := mb.AddVertex(v1)
	i2 := mb.AddVertex(v2)
	i3 := mb.AddVertex(v3)

	mb.AddTriangle(i1, i2, i3)

	// Test translation
	offset := dxfmath.NewVec3(10, 20, 30)
	mb.Translate(offset)

	// Check first vertex
	v, err := mb.GetVertex(0)
	if err != nil {
		t.Errorf("Error getting vertex: %v", err)
	}

	expected := dxfmath.NewVec3(10, 20, 30)
	if !v.IsEqual(expected, 1e-10) {
		t.Errorf("Expected vertex %v, got %v", expected, v)
	}

	// Test scaling
	mb.UniformScale(2.0)
	v, err = mb.GetVertex(0)
	if err != nil {
		t.Errorf("Error getting vertex after scaling: %v", err)
	}

	expected = dxfmath.NewVec3(20, 40, 60)
	if !v.IsEqual(expected, 1e-10) {
		t.Errorf("Expected scaled vertex %v, got %v", expected, v)
	}
}

func TestMeshBuilderClone(t *testing.T) {
	mb := NewMeshBuilder()

	// Create a simple mesh
	v1 := dxfmath.NewVec3(0, 0, 0)
	v2 := dxfmath.NewVec3(1, 0, 0)
	v3 := dxfmath.NewVec3(0, 1, 0)

	i1 := mb.AddVertex(v1)
	i2 := mb.AddVertex(v2)
	i3 := mb.AddVertex(v3)

	mb.AddTriangle(i1, i2, i3)

	// Clone mesh
	clone := mb.Clone()

	// Verify they're equal
	if clone.GetVertexCount() != mb.GetVertexCount() {
		t.Errorf("Clone vertex count mismatch: %d vs %d", clone.GetVertexCount(), mb.GetVertexCount())
	}

	if clone.GetFaceCount() != mb.GetFaceCount() {
		t.Errorf("Clone face count mismatch: %d vs %d", clone.GetFaceCount(), mb.GetFaceCount())
	}

	// Modify original and verify clone is unchanged
	mb.AddVertex(dxfmath.NewVec3(5, 5, 5))

	if clone.GetVertexCount() == mb.GetVertexCount() {
		t.Errorf("Clone should be independent of original")
	}
}
