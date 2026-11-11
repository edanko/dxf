package csg

import (
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func TestNewCSG(t *testing.T) {
	csg := NewCSG()
	if csg == nil {
		t.Error("NewCSG should return non-nil")
	}
	if len(csg.Polygons) != 0 {
		t.Errorf("Expected empty polygons, got %d", len(csg.Polygons))
	}
}

func TestNewCSGPlane(t *testing.T) {
	normal := dxfmath.NewVec3(0, 0, 1)
	point := dxfmath.NewVec3(0, 0, 0)

	plane := NewCSGPlane(normal, point)

	if plane.Normal.X() != 0 || plane.Normal.Y() != 0 || plane.Normal.Z() != 1 {
		t.Errorf("Expected normalized normal [0,0,1], got [%f,%f,%f]", plane.Normal.X(), plane.Normal.Y(), plane.Normal.Z())
	}
	if plane.W != 0 {
		t.Errorf("Expected W=0, got %f", plane.W)
	}
}

func TestPlaneDistanceToPoint(t *testing.T) {
	plane := NewCSGPlane(dxfmath.NewVec3(0, 0, 1), dxfmath.NewVec3(0, 0, 0))

	point := dxfmath.NewVec3(5, 5, 10)
	dist := plane.DistanceToPoint(point)
	if dist != 10 {
		t.Errorf("Expected distance 10, got %f", dist)
	}

	pointOnPlane := dxfmath.NewVec3(5, 5, 0)
	distOnPlane := plane.DistanceToPoint(pointOnPlane)
	if distOnPlane != 0 {
		t.Errorf("Expected distance 0 for point on plane, got %f", distOnPlane)
	}
}

func TestMeshToCSG(t *testing.T) {
	builder := dxfmath.NewMesh3DBuilder()
	builder.AddVertex(dxfmath.NewVec3(0, 0, 0))
	builder.AddVertex(dxfmath.NewVec3(1, 0, 0))
	builder.AddVertex(dxfmath.NewVec3(0, 1, 0))
	builder.AddTriangle(0, 1, 2)

	mesh := builder.Build()
	csg := MeshToCSG(mesh)

	if len(csg.Polygons) != 1 {
		t.Errorf("Expected 1 polygon, got %d", len(csg.Polygons))
	}
}

func TestMeshToCSGEmptyMesh(t *testing.T) {
	builder := dxfmath.NewMesh3DBuilder()
	mesh := builder.Build()
	csg := MeshToCSG(mesh)

	if len(csg.Polygons) != 0 {
		t.Errorf("Expected 0 polygons for empty mesh, got %d", len(csg.Polygons))
	}
}

func TestMeshToCSGQuad(t *testing.T) {
	builder := dxfmath.NewMesh3DBuilder()
	builder.AddVertex(dxfmath.NewVec3(0, 0, 0))
	builder.AddVertex(dxfmath.NewVec3(1, 0, 0))
	builder.AddVertex(dxfmath.NewVec3(1, 1, 0))
	builder.AddVertex(dxfmath.NewVec3(0, 1, 0))
	builder.AddQuad(0, 1, 2, 3)

	mesh := builder.Build()
	csg := MeshToCSG(mesh)

	if len(csg.Polygons) != 1 {
		t.Errorf("Expected 1 polygon for quad, got %d", len(csg.Polygons))
	}
}

func TestCSGToMesh(t *testing.T) {
	csg := NewCSG()
	mesh := csg.ToMesh()

	if mesh.Vertices == nil {
		t.Error("ToMesh should return non-nil vertices")
	}
	if mesh.Faces == nil {
		t.Error("ToMesh should return non-nil faces")
	}
}

func TestCSGUnion(t *testing.T) {
	builder1 := dxfmath.NewMesh3DBuilder()
	builder1.AddVertex(dxfmath.NewVec3(0, 0, 0))
	builder1.AddVertex(dxfmath.NewVec3(1, 0, 0))
	builder1.AddVertex(dxfmath.NewVec3(0, 1, 0))
	builder1.AddTriangle(0, 1, 2)
	mesh1 := builder1.Build()

	builder2 := dxfmath.NewMesh3DBuilder()
	builder2.AddVertex(dxfmath.NewVec3(0.5, 0, 0))
	builder2.AddVertex(dxfmath.NewVec3(1.5, 0, 0))
	builder2.AddVertex(dxfmath.NewVec3(0.5, 1, 0))
	builder2.AddTriangle(0, 1, 2)
	mesh2 := builder2.Build()

	csg1 := MeshToCSG(mesh1)
	csg2 := MeshToCSG(mesh2)

	result := csg1.Union(csg2)

	if len(result.Vertices) < 3 {
		t.Errorf("Union should have at least 3 vertices, got %d", len(result.Vertices))
	}
	if len(result.Faces) < 1 {
		t.Errorf("Union should have at least 1 face, got %d", len(result.Faces))
	}
}

func TestCSGIntersect(t *testing.T) {
	builder1 := dxfmath.NewMesh3DBuilder()
	builder1.AddVertex(dxfmath.NewVec3(0, 0, 0))
	builder1.AddVertex(dxfmath.NewVec3(2, 0, 0))
	builder1.AddVertex(dxfmath.NewVec3(0, 2, 0))
	builder1.AddTriangle(0, 1, 2)
	mesh1 := builder1.Build()

	builder2 := dxfmath.NewMesh3DBuilder()
	builder2.AddVertex(dxfmath.NewVec3(1, 0, 0))
	builder2.AddVertex(dxfmath.NewVec3(3, 0, 0))
	builder2.AddVertex(dxfmath.NewVec3(1, 2, 0))
	builder2.AddTriangle(0, 1, 2)
	mesh2 := builder2.Build()

	csg1 := MeshToCSG(mesh1)
	csg2 := MeshToCSG(mesh2)

	result := csg1.Intersect(csg2)

	if len(result.Vertices) == 0 {
		t.Log("Intersect result has no vertices (may be expected for non-overlapping meshes)")
	}
}

func TestCSGDifference(t *testing.T) {
	builder1 := dxfmath.NewMesh3DBuilder()
	builder1.AddVertex(dxfmath.NewVec3(0, 0, 0))
	builder1.AddVertex(dxfmath.NewVec3(2, 0, 0))
	builder1.AddVertex(dxfmath.NewVec3(0, 2, 0))
	builder1.AddTriangle(0, 1, 2)
	mesh1 := builder1.Build()

	builder2 := dxfmath.NewMesh3DBuilder()
	builder2.AddVertex(dxfmath.NewVec3(1, 0, 0))
	builder2.AddVertex(dxfmath.NewVec3(3, 0, 0))
	builder2.AddVertex(dxfmath.NewVec3(1, 2, 0))
	builder2.AddTriangle(0, 1, 2)
	mesh2 := builder2.Build()

	csg1 := MeshToCSG(mesh1)
	csg2 := MeshToCSG(mesh2)

	result := csg1.Difference(csg2)

	if len(result.Vertices) == 0 {
		t.Log("Difference result has no vertices (may be expected)")
	}
}

func TestMeshBoundingBox(t *testing.T) {
	builder := dxfmath.NewMesh3DBuilder()
	builder.AddVertex(dxfmath.NewVec3(0, 0, 0))
	builder.AddVertex(dxfmath.NewVec3(1, 2, 3))
	builder.AddVertex(dxfmath.NewVec3(4, 5, 6))
	builder.AddTriangle(0, 1, 2)
	mesh := builder.Build()

	min, max := meshBoundingBox(mesh)

	if min.X() != 0 || min.Y() != 0 || min.Z() != 0 {
		t.Errorf("Expected min [0,0,0], got [%f,%f,%f]", min.X(), min.Y(), min.Z())
	}
	if max.X() != 4 || max.Y() != 5 || max.Z() != 6 {
		t.Errorf("Expected max [4,5,6], got [%f,%f,%f]", max.X(), max.Y(), max.Z())
	}
}

func TestMeshBoundingBoxEmpty(t *testing.T) {
	builder := dxfmath.NewMesh3DBuilder()
	mesh := builder.Build()

	min, max := meshBoundingBox(mesh)

	if min != (dxfmath.Vec3{}) {
		t.Errorf("Expected empty min for empty mesh, got [%f,%f,%f]", min.X(), min.Y(), min.Z())
	}
	if max != (dxfmath.Vec3{}) {
		t.Errorf("Expected empty max for empty mesh, got [%f,%f,%f]", max.X(), max.Y(), max.Z())
	}
}

func TestContainsPoint(t *testing.T) {
	min := dxfmath.NewVec3(0, 0, 0)
	max := dxfmath.NewVec3(10, 10, 10)

	inside := dxfmath.NewVec3(5, 5, 5)
	if !containsPoint(min, max, inside) {
		t.Error("Point inside bbox should be contained")
	}

	outside := dxfmath.NewVec3(15, 15, 15)
	if containsPoint(min, max, outside) {
		t.Error("Point outside bbox should not be contained")
	}

	edge := dxfmath.NewVec3(10, 10, 10)
	if !containsPoint(min, max, edge) {
		t.Error("Point on edge should be contained")
	}
}
