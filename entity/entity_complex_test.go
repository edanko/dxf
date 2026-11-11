package entity

import (
	"math"
	"testing"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/table"
)

func TestPolylineBasic(t *testing.T) {
	p := NewPolyline()

	if p == nil {
		t.Fatal("NewPolyline should not return nil")
	}

	if p.Flag != 8 {
		t.Errorf("Default Flag should be 8, got %d", p.Flag)
	}

	if p.Vertices == nil {
		t.Error("Vertices should be initialized")
	}
}

func TestPolylineAddVertex(t *testing.T) {
	p := NewPolyline()

	v := p.AddVertex(0, 0, 0)
	if v == nil {
		t.Fatal("AddVertex should return a vertex")
	}

	if len(p.Vertices) != 1 {
		t.Errorf("Expected 1 vertex, got %d", len(p.Vertices))
	}

	p.AddVertex(1, 1, 1)
	p.AddVertex(2, 2, 2)

	if len(p.Vertices) != 3 {
		t.Errorf("Expected 3 vertices, got %d", len(p.Vertices))
	}
}

func TestPolylineVertexCoords(t *testing.T) {
	p := NewPolyline()
	p.AddVertex(5, 10, 15)
	p.AddVertex(20, 30, 40)

	if p.Vertices[0].Coord[0] != 5 || p.Vertices[0].Coord[1] != 10 || p.Vertices[0].Coord[2] != 15 {
		t.Errorf("Vertex 0 coords incorrect: [%f, %f, %f]", p.Vertices[0].Coord[0], p.Vertices[0].Coord[1], p.Vertices[0].Coord[2])
	}

	if p.Vertices[1].Coord[0] != 20 || p.Vertices[1].Coord[1] != 30 || p.Vertices[1].Coord[2] != 40 {
		t.Errorf("Vertex 1 coords incorrect: [%f, %f, %f]", p.Vertices[1].Coord[0], p.Vertices[1].Coord[1], p.Vertices[1].Coord[2])
	}
}

func TestPolylineClose(t *testing.T) {
	p := NewPolyline()

	if p.Flag&1 != 0 {
		t.Error("Polyline should not be closed initially")
	}

	p.Close()

	if p.Flag&1 == 0 {
		t.Error("Polyline should be closed after Close()")
	}
}

func TestPolylineBBox(t *testing.T) {
	p := NewPolyline()
	p.AddVertex(0, 0, 0)
	p.AddVertex(10, 10, 10)
	p.AddVertex(5, 5, 5)

	mins, maxs := p.BBox()

	if mins[0] != 0 || mins[1] != 0 || mins[2] != 0 {
		t.Errorf("Mins should be [0, 0, 0], got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 10 || maxs[1] != 10 || maxs[2] != 10 {
		t.Errorf("Maxs should be [10, 10, 10], got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestPolylineBBoxEmpty(t *testing.T) {
	p := NewPolyline()

	mins, maxs := p.BBox()

	if mins == nil || maxs == nil {
		t.Error("BBox should return non-nil for empty polyline")
	}
}

func TestPolylineIsEntity(t *testing.T) {
	p := NewPolyline()

	if !p.IsEntity() {
		t.Error("Polyline should implement IsEntity() returning true")
	}
}

func TestPolylineCopy(t *testing.T) {
	p := NewPolyline()
	p.AddVertex(1, 2, 3)
	p.AddVertex(4, 5, 6)
	p.Close()

	copied := p.Copy().(*Polyline)

	if len(copied.Vertices) != 2 {
		t.Errorf("Expected 2 vertices, got %d", len(copied.Vertices))
	}

	if copied.Vertices[0].Coord[0] != 1 {
		t.Error("Vertex not copied correctly")
	}
}

func TestPolylineFormat(t *testing.T) {
	p := NewPolyline()
	p.SetLayer(table.NewLayer("0", 0, nil))
	p.AddVertex(1, 2, 3)

	f := format.NewASCII()

	p.Format(f)

	if f == nil {
		t.Error("Format should work with formatter")
	}
}

func TestSplineBasic(t *testing.T) {
	s := NewSpline(3)

	if s == nil {
		t.Fatal("NewSpline should not return nil")
	}

	if s.Degree != 3 {
		t.Errorf("Default degree should be 3, got %d", s.Degree)
	}

	if s.Controls == nil {
		t.Error("Controls should be initialized")
	}

	if s.Knots == nil {
		t.Error("Knots should be initialized")
	}

	if s.Weights == nil {
		t.Error("Weights should be initialized")
	}
}

func TestSplineWithDegree(t *testing.T) {
	s := NewSpline(2)

	if s.Degree != 2 {
		t.Errorf("Degree should be 2, got %d", s.Degree)
	}
}

func TestSplineAddControlPoint(t *testing.T) {
	s := NewSpline(3)
	s.AddControlPoint(0, 0, 0)
	s.AddControlPoint(1, 1, 0)
	s.AddControlPoint(2, 0, 0)

	if len(s.Controls) != 3 {
		t.Errorf("Expected 3 controls, got %d", len(s.Controls))
	}
}

func TestSplineAddKnot(t *testing.T) {
	s := NewSpline(3)
	s.AddKnot(0)
	s.AddKnot(0)
	s.AddKnot(0)
	s.AddKnot(1)
	s.AddKnot(1)
	s.AddKnot(1)

	if len(s.Knots) != 6 {
		t.Errorf("Expected 6 knots, got %d", len(s.Knots))
	}
}

func TestSplineBBox(t *testing.T) {
	s := NewSpline(3)
	s.AddControlPoint(0, 0, 0)
	s.AddControlPoint(10, 10, 0)
	s.AddControlPoint(20, 0, 0)
	s.AddControlPoint(30, 10, 0)

	mins, maxs := s.BBox()

	if mins[0] < 0 || mins[1] < 0 || mins[2] < 0 {
		t.Errorf("Mins should be within controls: [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] > 30 || maxs[1] > 10 || maxs[2] > 0 {
		t.Errorf("Maxs should be within controls: [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestSplineIsEntity(t *testing.T) {
	s := NewSpline(3)

	if !s.IsEntity() {
		t.Error("Spline should implement IsEntity() returning true")
	}
}

func TestSplineCopy(t *testing.T) {
	s := NewSpline(3)
	s.AddControlPoint(0, 0, 0)
	s.AddControlPoint(10, 10, 0)
	s.AddKnot(0)
	s.AddKnot(0)
	s.AddKnot(0)
	s.AddKnot(1)
	s.AddKnot(1)
	s.AddKnot(1)

	copied := s.Copy().(*Spline)

	if len(copied.Controls) != 2 {
		t.Error("Controls not copied")
	}

	if len(copied.Knots) != 6 {
		t.Error("Knots not copied")
	}
}

func TestSplineFormat(t *testing.T) {
	s := NewSpline(3)
	s.AddControlPoint(0, 0, 0)
	s.AddControlPoint(1, 0, 0)
	s.SetLayer(table.NewLayer("0", 0, nil))

	f := format.NewASCII()

	s.Format(f)

	if f == nil {
		t.Error("Format should work with formatter")
	}
}

func TestSplineClosed(t *testing.T) {
	s := NewSpline(3)

	if s.IsClosed() {
		t.Error("Spline should not be closed initially")
	}

	s.SetClosed(true)

	if !s.IsClosed() {
		t.Error("Spline should be closed after SetClosed(true)")
	}
}

func TestEllipseBasic(t *testing.T) {
	e := NewEllipse()

	if e == nil {
		t.Fatal("NewEllipse should not return nil")
	}

	if e.Center == nil || len(e.Center) != 3 {
		t.Error("Center should be initialized with 3 elements")
	}

	if e.Ratio <= 0 || e.Ratio > 1 {
		t.Errorf("Ratio should be between 0 and 1, got %f", e.Ratio)
	}
}

func TestEllipseCenter(t *testing.T) {
	e := NewEllipse()
	e.Center = []float64{100, 200, 0}

	if e.Center[0] != 100 || e.Center[1] != 200 || e.Center[2] != 0 {
		t.Errorf("Center should be [100, 200, 0], got [%f, %f, %f]", e.Center[0], e.Center[1], e.Center[2])
	}
}

func TestEllipseBBox(t *testing.T) {
	e := NewEllipse()
	e.Center = []float64{10, 10, 0}
	e.Major = 5.0
	e.Ratio = 0.5

	mins, maxs := e.BBox()

	if mins == nil || maxs == nil {
		t.Error("BBox should return non-nil")
	}

	if mins[0] > 10 || maxs[0] < 10 {
		t.Error("BBox should span across center in X")
	}
}

func TestEllipseIsEntity(t *testing.T) {
	e := NewEllipse()

	if !e.IsEntity() {
		t.Error("Ellipse should implement IsEntity() returning true")
	}
}

func TestEllipseCopy(t *testing.T) {
	e := NewEllipse()
	e.Center = []float64{100, 200, 0}
	e.Major = 50.0
	e.Ratio = 0.5
	e.Start = 0
	e.End = 180

	copied := e.Copy().(*Ellipse)

	if copied.Center[0] != 100 {
		t.Error("Center not copied")
	}

	if copied.Major != 50.0 {
		t.Error("Major not copied")
	}
}

func TestMTextBasic(t *testing.T) {
	m := NewMText()

	if m == nil {
		t.Fatal("NewMText should not return nil")
	}

	if m.Coord1 == nil || len(m.Coord1) != 3 {
		t.Error("Coord1 should be initialized")
	}
}

func TestMTextSetText(t *testing.T) {
	m := NewMText()
	m.SetText("Hello, World!")

	if m.Value != "Hello, World!" {
		t.Errorf("Value should be 'Hello, World!', got '%s'", m.Value)
	}
}

func TestMTextBBox(t *testing.T) {
	m := NewMText()
	m.SetText("Test")
	m.Coord1 = []float64{0, 0, 0}

	mins, maxs := m.BBox()

	if mins == nil || maxs == nil {
		t.Error("BBox should return non-nil")
	}
}

func TestMTextIsEntity(t *testing.T) {
	m := NewMText()

	if !m.IsEntity() {
		t.Error("MText should implement IsEntity() returning true")
	}
}

func TestMTextSetCoord(t *testing.T) {
	m := NewMText()
	m.SetCoord([]float64{100, 200, 0})

	if m.Coord1[0] != 100 || m.Coord1[1] != 200 || m.Coord1[2] != 0 {
		t.Error("SetCoord did not set correctly")
	}
}

func TestMTextSetRotation(t *testing.T) {
	m := NewMText()
	m.SetRotation(45.0)

	if m.Rotation != 45.0 {
		t.Error("Rotation not set correctly")
	}
}

func TestMTextSetHeight(t *testing.T) {
	m := NewMText()
	m.SetHeight(5.0)

	if m.Height != 5.0 {
		t.Error("Height not set correctly")
	}
}

func TestMTextSetAttachmentPoint(t *testing.T) {
	m := NewMText()
	m.SetAttachmentPoint(MTextMiddleCenter)

	if m.AttachmentPoint != MTextMiddleCenter {
		t.Error("AttachmentPoint not set correctly")
	}
}

func TestTextBasic(t *testing.T) {
	txt := NewText()

	if txt == nil {
		t.Fatal("NewText should not return nil")
	}

	if txt.Height != 1.0 {
		t.Error("Height should be initialized to 1.0")
	}
}

func TestTextValue(t *testing.T) {
	txt := NewText()
	txt.Value = "Hello"

	if txt.Value != "Hello" {
		t.Errorf("Value should be 'Hello', got '%s'", txt.Value)
	}
}

func TestTextCoord(t *testing.T) {
	txt := NewText()
	txt.Value = "Test"

	mins, maxs := txt.BBox()

	if mins == nil || maxs == nil {
		t.Error("BBox should return non-nil")
	}
}

func TestTextIsEntity(t *testing.T) {
	txt := NewText()

	if !txt.IsEntity() {
		t.Error("Text should implement IsEntity() returning true")
	}
}

func TestTextCopy(t *testing.T) {
	txt := NewText()
	txt.Value = "Test"
	txt.Height = 2.5

	copied := txt.Copy().(*Text)

	if copied.Value != "Test" {
		t.Error("Value not copied")
	}

	if copied.Height != 2.5 {
		t.Error("Height not copied")
	}
}

func TestSolidBasic(t *testing.T) {
	s := NewSolid()

	if s == nil {
		t.Fatal("NewSolid should not return nil")
	}

	if s.FirstPoint == nil || s.SecondPoint == nil || s.ThirdPoint == nil || s.FourthPoint == nil {
		t.Error("All points should be initialized")
	}
}

func TestSolidPoints(t *testing.T) {
	s := NewSolid()
	s.FirstPoint = []float64{0, 0, 0}
	s.SecondPoint = []float64{10, 0, 0}
	s.ThirdPoint = []float64{10, 10, 0}
	s.FourthPoint = []float64{0, 10, 0}

	if s.FirstPoint[0] != 0 {
		t.Error("FirstPoint not set correctly")
	}

	if s.FourthPoint[0] != 0 {
		t.Error("FourthPoint not set correctly")
	}
}

func TestSolidBBox(t *testing.T) {
	s := NewSolid()
	s.FirstPoint = []float64{0, 0, 0}
	s.SecondPoint = []float64{10, 0, 0}
	s.ThirdPoint = []float64{10, 10, 0}
	s.FourthPoint = []float64{0, 10, 0}

	mins, maxs := s.BBox()

	if mins[0] != 0 || mins[1] != 0 || mins[2] != 0 {
		t.Errorf("Mins should be [0, 0, 0], got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 10 || maxs[1] != 10 || maxs[2] != 0 {
		t.Errorf("Maxs should be [10, 10, 0], got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestSolidIsEntity(t *testing.T) {
	s := NewSolid()

	if !s.IsEntity() {
		t.Error("Solid should implement IsEntity() returning true")
	}
}

func TestSolidCopy(t *testing.T) {
	s := NewSolid()
	s.FirstPoint = []float64{0, 0, 0}
	s.SecondPoint = []float64{10, 0, 0}
	s.ThirdPoint = []float64{10, 10, 0}
	s.FourthPoint = []float64{0, 10, 0}

	copied := s.Copy().(*Solid)

	if copied.FirstPoint[0] != 0 {
		t.Error("FirstPoint not copied")
	}

	if copied.FourthPoint[0] != 0 {
		t.Error("FourthPoint not copied")
	}
}

func Test3DFaceBasic(t *testing.T) {
	f := New3DFace()

	if f == nil {
		t.Fatal("New3DFace should not return nil")
	}

	if len(f.Points) != 4 {
		t.Errorf("Expected 4 points, got %d", len(f.Points))
	}
}

func Test3DFacePoints(t *testing.T) {
	f := New3DFace()
	f.Points[0] = []float64{0, 0, 0}
	f.Points[1] = []float64{1, 0, 0}
	f.Points[2] = []float64{1, 1, 0}
	f.Points[3] = []float64{0, 1, 0}

	if f.Points[0][0] != 0 {
		t.Error("Points[0] not set correctly")
	}

	if f.Points[3][0] != 0 {
		t.Error("Points[3] not set correctly")
	}
}

func Test3DFaceBBox(t *testing.T) {
	f := New3DFace()
	f.Points[0] = []float64{0, 0, 0}
	f.Points[1] = []float64{10, 0, 0}
	f.Points[2] = []float64{10, 10, 0}
	f.Points[3] = []float64{0, 10, 0}

	mins, maxs := f.BBox()

	if mins[0] != 0 || mins[1] != 0 || mins[2] != 0 {
		t.Errorf("Mins should be [0, 0, 0], got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 10 || maxs[1] != 10 || maxs[2] != 0 {
		t.Errorf("Maxs should be [10, 10, 0], got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func Test3DFaceIsEntity(t *testing.T) {
	f := New3DFace()

	if !f.IsEntity() {
		t.Error("3DFace should implement IsEntity() returning true")
	}
}

func Test3DFaceCopy(t *testing.T) {
	f := New3DFace()
	f.Points[0] = []float64{0, 0, 0}
	f.Points[1] = []float64{10, 0, 0}
	f.Points[2] = []float64{10, 10, 0}
	f.Points[3] = []float64{0, 10, 0}

	copied := f.Copy().(*ThreeDFace)

	if copied.Points[0][0] != 0 {
		t.Error("Points[0] not copied")
	}

	if copied.Points[3][0] != 0 {
		t.Error("Points[3] not copied")
	}
}

func TestInsertBasic(t *testing.T) {
	i := NewInsert()

	if i == nil {
		t.Fatal("NewInsert should not return nil")
	}

	point := i.InsertPoint()
	if point.X() != 0 || point.Y() != 0 || point.Z() != 0 {
		t.Error("InsertPoint should be at origin by default")
	}
}

func TestInsertSetPoint(t *testing.T) {
	i := NewInsert()
	i.SetInsertPoint(dxfmath.NewVec3(10, 20, 0))
	i.SetRotation(45.0)

	copied := i.Copy().(*Insert)

	point := copied.InsertPoint()
	if point.X() != 10 {
		t.Error("Point not copied")
	}

	if math.Abs(copied.Rotation()-45.0) > 1e-10 {
		t.Error("Rotation not copied")
	}
}

func TestVertexBasic(t *testing.T) {
	v := NewVertex(1, 2, 3)

	if v == nil {
		t.Fatal("NewVertex should not return nil")
	}

	if v.Coord[0] != 1 || v.Coord[1] != 2 || v.Coord[2] != 3 {
		t.Errorf("Coord should be [1, 2, 3], got [%f, %f, %f]", v.Coord[0], v.Coord[1], v.Coord[2])
	}
}

func TestVertexBBox(t *testing.T) {
	v := NewVertex(5, 10, 15)

	mins, maxs := v.BBox()

	if mins[0] != 5 || mins[1] != 10 || mins[2] != 15 {
		t.Errorf("Mins should be [5, 10, 15], got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 5 || maxs[1] != 10 || maxs[2] != 15 {
		t.Errorf("Maxs should be [5, 10, 15], got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestVertexIsEntity(t *testing.T) {
	v := NewVertex(1, 2, 3)

	if !v.IsEntity() {
		t.Error("Vertex should implement IsEntity() returning true")
	}
}

func TestVertexCopy(t *testing.T) {
	v := NewVertex(5, 10, 15)

	copied := v.Copy().(*Vertex)

	if copied.Coord[0] != 5 || copied.Coord[1] != 10 || copied.Coord[2] != 15 {
		t.Error("Coord not copied correctly")
	}
}

func TestSplineClosedPeriodic(t *testing.T) {
	s := NewSpline(3)
	s.SetClosed(true)
	s.SetPeriodic(true)

	if !s.IsClosed() {
		t.Error("Spline should be closed")
	}

	if !s.IsPeriodic() {
		t.Error("Spline should be periodic")
	}
}

func TestSplineRationalLinearPlanar(t *testing.T) {
	s := NewSpline(3)
	s.SetRational(true)
	s.SetLinear(true)
	s.SetPlanar(true)

	if !s.IsRational() {
		t.Error("Spline should be rational")
	}

	if !s.IsLinear() {
		t.Error("Spline should be linear")
	}

	if !s.IsPlanar() {
		t.Error("Spline should be planar")
	}
}

func TestMTextLineSpacing(t *testing.T) {
	m := NewMText()
	m.SetLineSpacing(MTextExact, 1.5)

	if m.LineSpacingStyle != MTextExact {
		t.Error("LineSpacingStyle not set correctly")
	}

	if m.LineSpacingFactor != 1.5 {
		t.Error("LineSpacingFactor not set correctly")
	}
}

func TestMTextAdditionalText(t *testing.T) {
	m := NewMText()
	m.AddAdditionalText("Additional text")

	if len(m.AdditionalText) != 1 {
		t.Error("AdditionalText not added correctly")
	}
}

func TestInsertGrid(t *testing.T) {
	i := NewInsert()
	i.SetGrid(3, 4, 10.0, 20.0)

	if i.columnCount != 3 {
		t.Error("columnCount not set correctly")
	}

	if i.rowCount != 4 {
		t.Error("rowCount not set correctly")
	}

	if i.columnSpacing != 10.0 {
		t.Error("columnSpacing not set correctly")
	}

	if i.rowSpacing != 20.0 {
		t.Error("rowSpacing not set correctly")
	}

	if !i.isMultiple {
		t.Error("isMultiple should be true")
	}
}

func TestSolidThickness(t *testing.T) {
	s := NewSolid()
	s.Thickness = 5.0

	if s.Thickness != 5.0 {
		t.Error("Thickness not set correctly")
	}
}
