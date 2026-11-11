package render

import (
	"math"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func TestNewFormGenerator(t *testing.T) {
	fg := NewFormGenerator()
	if fg == nil {
		t.Fatal("NewFormGenerator should return non-nil")
	}
	if fg.scale != 1.0 {
		t.Errorf("Expected default scale 1.0, got %f", fg.scale)
	}
	if fg.segments != 32 {
		t.Errorf("Expected default segments 32, got %d", fg.segments)
	}
}

func TestFormGeneratorSetScale(t *testing.T) {
	fg := NewFormGenerator()
	fg.SetScale(2.5)
	if fg.scale != 2.5 {
		t.Errorf("Expected scale 2.5, got %f", fg.scale)
	}
}

func TestFormGeneratorSetSegments(t *testing.T) {
	fg := NewFormGenerator()
	fg.SetSegments(64)
	if fg.segments != 64 {
		t.Errorf("Expected segments 64, got %d", fg.segments)
	}
}

func TestFormGeneratorCircle(t *testing.T) {
	fg := NewFormGenerator()
	points := fg.Circle(10.0)

	if len(points) != 32 {
		t.Errorf("Expected 32 points, got %d", len(points))
	}

	for i, p := range points {
		dist := math.Sqrt(p.X()*p.X() + p.Y()*p.Y())
		if math.Abs(dist-10.0) > 0.001 {
			t.Errorf("Point %d: expected distance 10, got %f", i, dist)
		}
	}
}

func TestFormGeneratorCircleCustomSegments(t *testing.T) {
	fg := NewFormGenerator()
	fg.SetSegments(8)
	points := fg.Circle(5.0)

	if len(points) != 8 {
		t.Errorf("Expected 8 points, got %d", len(points))
	}
}

func TestFormGeneratorEllipse(t *testing.T) {
	fg := NewFormGenerator()
	points := fg.Ellipse(10.0, 5.0)

	if len(points) != 32 {
		t.Errorf("Expected 32 points, got %d", len(points))
	}
}

func TestFormGeneratorSquare(t *testing.T) {
	fg := NewFormGenerator()
	points := fg.Square(10.0)

	if len(points) != 4 {
		t.Errorf("Expected 4 points, got %d", len(points))
	}
}

func TestFormGeneratorSquareClosed(t *testing.T) {
	fg := NewFormGenerator()
	points := fg.Square(10.0, true)

	if len(points) != 5 {
		t.Errorf("Expected 5 points for closed square, got %d", len(points))
	}

	if points[4] != points[0] {
		t.Errorf("Last point should equal first point for closed square")
	}
}

func TestFormGeneratorBox(t *testing.T) {
	fg := NewFormGenerator()
	vertices := fg.Box(10, 20, false)

	if len(vertices) != 4 {
		t.Errorf("Expected 4 vertices for box, got %d", len(vertices))
	}
}

func TestFormGeneratorRectangle(t *testing.T) {
	fg := NewFormGenerator()
	points := fg.Rectangle(10.0, 20.0)

	if len(points) != 4 {
		t.Errorf("Expected 4 points, got %d", len(points))
	}

	width := points[1].X() - points[0].X()
	height := points[2].Y() - points[1].Y()

	if math.Abs(width-10.0) > 0.001 {
		t.Errorf("Expected width 10, got %f", width)
	}
	if math.Abs(height-20.0) > 0.001 {
		t.Errorf("Expected height 20, got %f", height)
	}
}

func TestFormGeneratorTriangle(t *testing.T) {
	fg := NewFormGenerator()
	points := fg.Triangle(10.0)

	if len(points) != 3 {
		t.Errorf("Expected 3 points, got %d", len(points))
	}
}

func TestFormGeneratorArrow(t *testing.T) {
	fg := NewFormGenerator()
	points := fg.Arrow(10.0, 5.0)

	if len(points) < 3 {
		t.Errorf("Arrow should have at least 3 points, got %d", len(points))
	}
}

func TestFormGeneratorHelix(t *testing.T) {
	fg := NewFormGenerator()
	fg.SetSegments(64)
	mesh := fg.Helix(5.0, 20.0, 10.0, 2)

	if len(mesh) == 0 {
		t.Error("Helix should have vertices")
	}
}

func TestFormGeneratorTranslate(t *testing.T) {
	fg := NewFormGenerator()
	points := []dxfmath.Vec3{
		dxfmath.NewVec3(1, 1, 1),
	}

	translated := fg.Translate(points, dxfmath.NewVec3(10, 20, 30))

	if len(translated) != 1 {
		t.Errorf("Expected 1 translated point, got %d", len(translated))
	}
	if translated[0].X() != 11 {
		t.Errorf("Expected X=11 after translate, got %f", translated[0].X())
	}
}

func TestFormGeneratorScale(t *testing.T) {
	fg := NewFormGenerator()
	points := []dxfmath.Vec3{
		dxfmath.NewVec3(1, 1, 1),
	}

	scaled := fg.Scale(points, 2.0)

	if len(scaled) != 1 {
		t.Errorf("Expected 1 scaled point, got %d", len(scaled))
	}
	if scaled[0].X() != 2 {
		t.Errorf("Expected X=2 after scale, got %f", scaled[0].X())
	}
}

func TestFormGeneratorClosePolygon(t *testing.T) {
	fg := NewFormGenerator()
	points := []dxfmath.Vec3{
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(10, 0, 0),
		dxfmath.NewVec3(10, 10, 0),
	}

	mesh := fg.ClosePolygon(points)

	if len(mesh) != 4 {
		t.Errorf("Expected 4 vertices for closed polygon, got %d", len(mesh))
	}
}

func TestNewMeshBuilder(t *testing.T) {
	mb := NewMeshBuilder()
	if mb == nil {
		t.Fatal("NewMeshBuilder should return non-nil")
	}
	if mb.GetVertexCount() != 0 {
		t.Errorf("Expected 0 vertices, got %d", mb.GetVertexCount())
	}
	if mb.GetFaceCount() != 0 {
		t.Errorf("Expected 0 faces, got %d", mb.GetFaceCount())
	}
}

func TestMeshBuilderAddVertex(t *testing.T) {
	mb := NewMeshBuilder()
	idx := mb.AddVertex(dxfmath.NewVec3(1, 2, 3))

	if idx != 0 {
		t.Errorf("Expected first vertex index 0, got %d", idx)
	}
	if mb.GetVertexCount() != 1 {
		t.Errorf("Expected 1 vertex, got %d", mb.GetVertexCount())
	}
}

func TestMeshBuilderAddFace(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(0, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(1, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(0, 1, 0))

	err := mb.AddFace([]int{0, 1, 2})
	if err != nil {
		t.Errorf("AddFace failed: %v", err)
	}
	if mb.GetFaceCount() != 1 {
		t.Errorf("Expected 1 face, got %d", mb.GetFaceCount())
	}
}

func TestMeshBuilderAddFaceInvalidIndices(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(0, 0, 0))

	err := mb.AddFace([]int{0, 1, 2})
	if err == nil {
		t.Error("AddFace should fail with invalid vertex indices")
	}
}

func TestMeshBuilderAddFaceTooFewVertices(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(0, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(1, 0, 0))

	err := mb.AddFace([]int{0, 1})
	if err == nil {
		t.Error("AddFace should fail with fewer than 3 vertices")
	}
}

func TestMeshBuilderAddTriangle(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(0, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(1, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(0, 1, 0))

	err := mb.AddTriangle(0, 1, 2)
	if err != nil {
		t.Errorf("AddTriangle failed: %v", err)
	}
	if mb.GetFaceCount() != 1 {
		t.Errorf("Expected 1 face, got %d", mb.GetFaceCount())
	}
}

func TestMeshBuilderAddQuad(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(0, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(1, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(1, 1, 0))
	mb.AddVertex(dxfmath.NewVec3(0, 1, 0))

	err := mb.AddQuad(0, 1, 2, 3)
	if err != nil {
		t.Errorf("AddQuad failed: %v", err)
	}
	if mb.GetFaceCount() != 1 {
		t.Errorf("Expected 1 face, got %d", mb.GetFaceCount())
	}
}

func TestMeshBuilderGetVertex(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(1, 2, 3))
	mb.AddVertex(dxfmath.NewVec3(4, 5, 6))

	v, err := mb.GetVertex(1)
	if err != nil {
		t.Errorf("GetVertex failed: %v", err)
	}
	if v.X() != 4 || v.Y() != 5 || v.Z() != 6 {
		t.Errorf("Expected (4, 5, 6), got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestMeshBuilderGetVertexInvalidIndex(t *testing.T) {
	mb := NewMeshBuilder()

	_, err := mb.GetVertex(0)
	if err == nil {
		t.Error("GetVertex should fail with invalid index")
	}
}

func TestMeshBuilderGetFace(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(0, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(1, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(0, 1, 0))
	mb.AddFace([]int{0, 1, 2})

	face, err := mb.GetFace(0)
	if err != nil {
		t.Errorf("GetFace failed: %v", err)
	}
	if len(face) != 3 {
		t.Errorf("Expected 3 indices, got %d", len(face))
	}
}

func TestMeshBuilderGetFaceInvalidIndex(t *testing.T) {
	mb := NewMeshBuilder()

	_, err := mb.GetFace(0)
	if err == nil {
		t.Error("GetFace should fail with invalid index")
	}
}

func TestMeshBuilderGetVertices(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(1, 2, 3))
	mb.AddVertex(dxfmath.NewVec3(4, 5, 6))

	vertices := mb.GetVertices()
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices, got %d", len(vertices))
	}
}

func TestMeshBuilderGetFaces(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(0, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(1, 0, 0))
	mb.AddVertex(dxfmath.NewVec3(0, 1, 0))
	mb.AddFace([]int{0, 1, 2})

	faces := mb.GetFaces()
	if len(faces) != 1 {
		t.Errorf("Expected 1 face, got %d", len(faces))
	}
}

func TestMeshBuilderClear(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(1, 2, 3))
	mb.AddFace([]int{0, 0, 0})

	mb.Clear()

	if mb.GetVertexCount() != 0 {
		t.Errorf("Expected 0 vertices after clear, got %d", mb.GetVertexCount())
	}
	if mb.GetFaceCount() != 0 {
		t.Errorf("Expected 0 faces after clear, got %d", mb.GetFaceCount())
	}
}

func TestMeshBuilderTranslate(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(1, 1, 1))

	mb.Translate(dxfmath.NewVec3(10, 20, 30))

	v, _ := mb.GetVertex(0)
	if v.X() != 11 {
		t.Errorf("Expected X=11 after translate, got %f", v.X())
	}
}

func TestMeshBuilderScale(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(1, 1, 1))

	mb.Scale(2.0, 2.0, 2.0)

	v, _ := mb.GetVertex(0)
	if v.X() != 2 {
		t.Errorf("Expected X=2 after scale, got %f", v.X())
	}
}

func TestMeshBuilderUniformScale(t *testing.T) {
	mb := NewMeshBuilder()
	mb.AddVertex(dxfmath.NewVec3(1, 2, 3))

	mb.UniformScale(0.5)

	v, _ := mb.GetVertex(0)
	if v.X() != 0.5 {
		t.Errorf("Expected X=0.5 after uniform scale, got %f", v.X())
	}
}
