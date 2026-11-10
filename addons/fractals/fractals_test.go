package fractals

import (
	"testing"

	"github.com/edanko/dxf/drawing"
	dxfmath "github.com/edanko/dxf/math"
)

func TestMengerSpongeCubeAndFaceCounts(t *testing.T) {
	sponge, err := NewMengerSponge(dxfmath.Vec3{0, 0, 0}, 3, 1, MengerOriginal)
	if err != nil {
		t.Fatal(err)
	}
	if got := sponge.CubeCount(); got != 20 {
		t.Fatalf("cube count = %d, want 20", got)
	}
	if got := sponge.FaceCountPerCube(); got != 6 {
		t.Fatalf("faces per cube = %d, want 6", got)
	}
	if got := sponge.TotalCubeFaceCount(); got != 120 {
		t.Fatalf("total cube faces = %d, want 120", got)
	}
	if got := len(sponge.Cubes()); got != 20 {
		t.Fatalf("generated cubes = %d, want 20", got)
	}
	meshes := sponge.Meshes()
	if len(meshes) != 20 {
		t.Fatalf("meshes = %d, want 20", len(meshes))
	}
	for _, mesh := range meshes {
		if mesh.FaceCount() != 6 || mesh.VertexCount() != 8 {
			t.Fatalf("cube mesh has %d faces and %d vertices", mesh.FaceCount(), mesh.VertexCount())
		}
	}

	level2, err := NewMengerSponge(dxfmath.Vec3{}, 3, 2, MengerOriginal)
	if err != nil {
		t.Fatal(err)
	}
	if got := level2.CubeCount(); got != 400 {
		t.Fatalf("level 2 cube count = %d, want 400", got)
	}
	if got := len(level2.Cubes()); got != 400 {
		t.Fatalf("level 2 generated cubes = %d, want 400", got)
	}
}

func TestMengerSpongeMergedMeshIsManifold(t *testing.T) {
	sponge, err := NewMengerSponge(dxfmath.Vec3{}, 3, 1, MengerOriginal)
	if err != nil {
		t.Fatal(err)
	}
	mesh := sponge.Mesh()
	if mesh.FaceCount() == 0 || mesh.VertexCount() == 0 {
		t.Fatal("merged mesh is empty")
	}
	if mesh.FaceCount() >= sponge.TotalCubeFaceCount() {
		t.Fatalf("merged mesh kept %d faces; interior faces should be removed from %d", mesh.FaceCount(), sponge.TotalCubeFaceCount())
	}
}

func TestMengerSpongeVariantsAndValidation(t *testing.T) {
	jerusalem, err := NewMengerSponge(dxfmath.Vec3{}, 5, 1, JerusalemCube)
	if err != nil {
		t.Fatal(err)
	}
	if got := jerusalem.CubeCount(); got != 76 {
		t.Fatalf("Jerusalem cube count = %d, want 76", got)
	}
	if _, err := NewMengerSponge(dxfmath.Vec3{}, 1, 0, MengerOriginal); err == nil {
		t.Fatal("expected level validation error")
	}
	if _, err := NewMengerSponge(dxfmath.Vec3{}, 1, 1, MengerKind(9)); err == nil {
		t.Fatal("expected kind validation error")
	}
	if _, err := NewMengerSponge(dxfmath.Vec3{}, 0, 1, MengerOriginal); err == nil {
		t.Fatal("expected length validation error")
	}
}

func TestMengerSpongeAddToDrawing(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	sponge, err := NewMengerSponge(dxfmath.Vec3{}, 3, 1, MengerOriginal)
	if err != nil {
		t.Fatal(err)
	}
	if err := sponge.AddTo(doc); err != nil {
		t.Fatal(err)
	}
	if got := len(doc.Entities()); got != 1 {
		t.Fatalf("drawing entities = %d, want 1", got)
	}

	doc2, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if err := sponge.AddCubesTo(doc2); err != nil {
		t.Fatal(err)
	}
	if got := len(doc2.Entities()); got != 20 {
		t.Fatalf("drawing entities = %d, want 20", got)
	}
	if err := sponge.AddTo(nil); err == nil {
		t.Fatal("expected nil drawing error")
	}
}

func TestSierpinskiPyramidCountsAndMeshes(t *testing.T) {
	pyramid, err := NewSierpinskiPyramid(dxfmath.Vec3{}, 1, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if got := pyramid.PyramidCount(); got != 5 {
		t.Fatalf("pyramid count = %d, want 5", got)
	}
	if got := pyramid.FaceCountPerPyramid(); got != 5 {
		t.Fatalf("faces per pyramid = %d, want 5", got)
	}
	if got := pyramid.TotalFaceCount(); got != 25 {
		t.Fatalf("total faces = %d, want 25", got)
	}
	meshes := pyramid.Meshes()
	if len(meshes) != 5 {
		t.Fatalf("meshes = %d, want 5", len(meshes))
	}
	for _, mesh := range meshes {
		if mesh.FaceCount() != 5 || mesh.VertexCount() != 5 {
			t.Fatalf("pyramid mesh has %d faces and %d vertices", mesh.FaceCount(), mesh.VertexCount())
		}
	}
	if got := pyramid.Mesh().FaceCount(); got != 25 {
		t.Fatalf("merged pyramid faces = %d, want 25", got)
	}

	level2, err := NewSierpinskiPyramid(dxfmath.Vec3{}, 1, 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	if got := level2.PyramidCount(); got != 25 {
		t.Fatalf("level 2 pyramid count = %d, want 25", got)
	}
	if got := len(level2.Pyramids()); got != 25 {
		t.Fatalf("level 2 generated pyramids = %d, want 25", got)
	}
}

func TestSierpinskiTriangleBaseAndValidation(t *testing.T) {
	pyramid, err := NewSierpinskiPyramid(dxfmath.Vec3{}, 2, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got := pyramid.PyramidCount(); got != 4 {
		t.Fatalf("pyramid count = %d, want 4", got)
	}
	if got := pyramid.FaceCountPerPyramid(); got != 4 {
		t.Fatalf("faces per pyramid = %d, want 4", got)
	}
	if len(pyramid.PyramidVertices(pyramid.Pyramids()[0])) != 4 {
		t.Fatal("triangle pyramid should have 4 vertices")
	}
	if _, err := NewSierpinskiPyramid(dxfmath.Vec3{}, 1, 1, 5); err == nil {
		t.Fatal("expected sides validation error")
	}
	if _, err := NewSierpinskiPyramid(dxfmath.Vec3{}, 1, 0, 4); err == nil {
		t.Fatal("expected level validation error")
	}
	if _, err := NewSierpinskiPyramid(dxfmath.Vec3{}, 0, 1, 4); err == nil {
		t.Fatal("expected length validation error")
	}
}

func TestSierpinskiPyramidAddToDrawing(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	pyramid, err := NewSierpinskiPyramid(dxfmath.Vec3{}, 1, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := pyramid.AddTo(doc); err != nil {
		t.Fatal(err)
	}
	if got := len(doc.Entities()); got != 1 {
		t.Fatalf("drawing entities = %d, want 1", got)
	}
	doc2, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if err := pyramid.AddPyramidsTo(doc2); err != nil {
		t.Fatal(err)
	}
	if got := len(doc2.Entities()); got != 5 {
		t.Fatalf("drawing entities = %d, want 5", got)
	}
	if err := pyramid.AddTo(nil); err == nil {
		t.Fatal("expected nil drawing error")
	}
}

func TestCapabilityRegistered(t *testing.T) {
	capability := Capability{}
	if capability.Name() != "fractals" {
		t.Fatalf("capability name = %q", capability.Name())
	}
}
