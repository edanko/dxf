package csg

import (
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func TestMeshUnion(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(1, 1, 1), 2)

	result := MeshUnion(mesh1, mesh2)

	if len(result.Vertices) < 3 {
		t.Errorf("MeshUnion should have at least 3 vertices, got %d", len(result.Vertices))
	}
	if len(result.Faces) < 1 {
		t.Errorf("MeshUnion should have at least 1 face, got %d", len(result.Faces))
	}
}

func TestMeshDifference(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(1, 1, 1), 2)

	result := MeshDifference(mesh1, mesh2)

	if result.Vertices == nil {
		t.Error("MeshDifference should return non-nil vertices")
	}
	if result.Faces == nil {
		t.Error("MeshDifference should return non-nil faces")
	}
}

func TestMeshXOR(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(0.25, 0.25, 0.25), 1)

	result := MeshXOR(mesh1, mesh2)

	if len(result.Vertices) == 0 || len(result.Faces) == 0 {
		t.Fatalf("MeshXOR empty: %d verts %d faces", len(result.Vertices), len(result.Faces))
	}
	if vol := Volume(result); vol < 6.9 || vol > 7.1 {
		t.Errorf("MeshXOR volume = %v, want ~7", vol)
	}
}

func TestMeshIntersect(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(0.25, 0.25, 0.25), 1)

	result := MeshIntersect(mesh1, mesh2)

	if len(result.Vertices) == 0 || len(result.Faces) == 0 {
		t.Fatalf("MeshIntersect empty: %d verts %d faces", len(result.Vertices), len(result.Faces))
	}
	if vol := Volume(result); vol < 0.9 || vol > 1.1 {
		t.Errorf("MeshIntersect volume = %v, want ~1", vol)
	}
}

func TestCSGMeshUnion(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(1, 1, 1), 2)

	csg := MeshToCSG(mesh1)
	result := csg.MeshUnion(mesh2)

	if len(result.Vertices) < 3 {
		t.Errorf("CSG.MeshUnion should have at least 3 vertices, got %d", len(result.Vertices))
	}
}

func TestCSGMeshDifference(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(1, 1, 1), 2)

	csg := MeshToCSG(mesh1)
	result := csg.MeshDifference(mesh2)

	if result.Vertices == nil {
		t.Error("CSG.MeshDifference should return non-nil vertices")
	}
}

func TestCSGMeshXOR(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(0.25, 0.25, 0.25), 1)

	c := MeshToCSG(mesh1)
	result := c.MeshXOR(mesh2)

	if len(result.Vertices) == 0 {
		t.Fatal("CSG.MeshXOR should return non-nil vertices")
	}
	if vol := Volume(result); vol < 6.9 || vol > 7.1 {
		t.Errorf("CSG.MeshXOR volume = %v, want ~7", vol)
	}
}

func TestCSGMeshIntersect(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(0.25, 0.25, 0.25), 1)

	c := MeshToCSG(mesh1)
	result := c.MeshIntersect(mesh2)

	if len(result.Vertices) == 0 {
		t.Fatal("CSG.MeshIntersect should return non-nil vertices")
	}
	if vol := Volume(result); vol < 0.9 || vol > 1.1 {
		t.Errorf("CSG.MeshIntersect volume = %v, want ~1", vol)
	}
}

func TestMeshUnionNonOverlapping(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(10, 10, 10), 2)

	result := MeshUnion(mesh1, mesh2)

	// Two disjoint closed boxes keep both of their surfaces.
	if len(result.Vertices) != 16 {
		t.Errorf("Non-overlapping union should have 16 vertices, got %d", len(result.Vertices))
	}
	if len(result.Faces) != 24 {
		t.Errorf("Non-overlapping union should have 24 faces, got %d", len(result.Faces))
	}
	if vol := Volume(result); vol < 15.9 || vol > 16.1 {
		t.Errorf("Non-overlapping union volume = %v, want 16", vol)
	}
}

func TestMeshDifferenceNoOverlap(t *testing.T) {
	mesh1 := Box(dxfmath.NewVec3(0, 0, 0), 2)
	mesh2 := Box(dxfmath.NewVec3(10, 10, 10), 2)

	result := MeshDifference(mesh1, mesh2)

	// With no overlap the difference is just mesh1.
	if len(result.Vertices) != 8 {
		t.Errorf("Non-overlapping difference should have 8 vertices, got %d", len(result.Vertices))
	}
	if vol := Volume(result); vol < 7.9 || vol > 8.1 {
		t.Errorf("Non-overlapping difference volume = %v, want 8", vol)
	}
}
