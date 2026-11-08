package csg

import (
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func createTriangleMesh(v1, v2, v3 dxfmath.Vec3) dxfmath.Mesh3D {
	builder := dxfmath.NewMesh3DBuilder()
	builder.AddVertex(v1)
	builder.AddVertex(v2)
	builder.AddVertex(v3)
	builder.AddTriangle(0, 1, 2)
	return builder.Build()
}

func TestMeshUnion(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(0, 1, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(0.5, 0, 0),
		dxfmath.NewVec3(1.5, 0, 0),
		dxfmath.NewVec3(0.5, 1, 0),
	)

	result := MeshUnion(mesh1, mesh2)

	if len(result.Vertices) < 3 {
		t.Errorf("MeshUnion should have at least 3 vertices, got %d", len(result.Vertices))
	}
	if len(result.Faces) < 1 {
		t.Errorf("MeshUnion should have at least 1 face, got %d", len(result.Faces))
	}
}

func TestMeshDifference(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(2, 0, 0),
		dxfmath.NewVec3(0, 2, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(3, 0, 0),
		dxfmath.NewVec3(1, 2, 0),
	)

	result := MeshDifference(mesh1, mesh2)

	if result.Vertices == nil {
		t.Error("MeshDifference should return non-nil vertices")
	}
	if result.Faces == nil {
		t.Error("MeshDifference should return non-nil faces")
	}
}

func TestMeshXOR(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(0, 1, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(0.5, 0, 0),
		dxfmath.NewVec3(1.5, 0, 0),
		dxfmath.NewVec3(0.5, 1, 0),
	)

	result := MeshXOR(mesh1, mesh2)

	if result.Vertices == nil {
		t.Error("MeshXOR should return non-nil vertices")
	}
	if result.Faces == nil {
		t.Error("MeshXOR should return non-nil faces")
	}
}

func TestMeshIntersect(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(2, 0, 0),
		dxfmath.NewVec3(0, 2, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(3, 0, 0),
		dxfmath.NewVec3(1, 2, 0),
	)

	result := MeshIntersect(mesh1, mesh2)

	if result.Vertices == nil {
		t.Error("MeshIntersect should return non-nil vertices")
	}
	if result.Faces == nil {
		t.Error("MeshIntersect should return non-nil faces")
	}
}

func TestCSGMeshUnion(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(0, 1, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(0.5, 0, 0),
		dxfmath.NewVec3(1.5, 0, 0),
		dxfmath.NewVec3(0.5, 1, 0),
	)

	csg := MeshToCSG(mesh1)
	result := csg.MeshUnion(mesh2)

	if len(result.Vertices) < 3 {
		t.Errorf("CSG.MeshUnion should have at least 3 vertices, got %d", len(result.Vertices))
	}
}

func TestCSGMeshDifference(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(2, 0, 0),
		dxfmath.NewVec3(0, 2, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(3, 0, 0),
		dxfmath.NewVec3(1, 2, 0),
	)

	csg := MeshToCSG(mesh1)
	result := csg.MeshDifference(mesh2)

	if result.Vertices == nil {
		t.Error("CSG.MeshDifference should return non-nil vertices")
	}
}

func TestCSGMeshXOR(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(0, 1, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(0.5, 0, 0),
		dxfmath.NewVec3(1.5, 0, 0),
		dxfmath.NewVec3(0.5, 1, 0),
	)

	csg := MeshToCSG(mesh1)
	result := csg.MeshXOR(mesh2)

	if result.Vertices == nil {
		t.Error("CSG.MeshXOR should return non-nil vertices")
	}
}

func TestCSGMeshIntersect(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(2, 0, 0),
		dxfmath.NewVec3(0, 2, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(3, 0, 0),
		dxfmath.NewVec3(1, 2, 0),
	)

	csg := MeshToCSG(mesh1)
	result := csg.MeshIntersect(mesh2)

	if result.Vertices == nil {
		t.Error("CSG.MeshIntersect should return non-nil vertices")
	}
}

func TestMeshUnionNonOverlapping(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(0, 1, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(10, 10, 0),
		dxfmath.NewVec3(11, 10, 0),
		dxfmath.NewVec3(10, 11, 0),
	)

	result := MeshUnion(mesh1, mesh2)

	if len(result.Vertices) != 6 {
		t.Errorf("Non-overlapping union should have 6 vertices, got %d", len(result.Vertices))
	}
	if len(result.Faces) != 2 {
		t.Errorf("Non-overlapping union should have 2 faces, got %d", len(result.Faces))
	}
}

func TestMeshDifferenceNoOverlap(t *testing.T) {
	mesh1 := createTriangleMesh(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(0, 1, 0),
	)

	mesh2 := createTriangleMesh(
		dxfmath.NewVec3(10, 10, 0),
		dxfmath.NewVec3(11, 10, 0),
		dxfmath.NewVec3(10, 11, 0),
	)

	result := MeshDifference(mesh1, mesh2)

	if len(result.Vertices) != 3 {
		t.Errorf("Non-overlapping difference should have 3 vertices, got %d", len(result.Vertices))
	}
}
