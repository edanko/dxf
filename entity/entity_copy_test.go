package entity

import (
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func TestEntityCopy(t *testing.T) {
	t.Run("Copy Line", func(t *testing.T) {
		line := NewLine()
		line.Start = []float64{1, 2, 3}
		line.End = []float64{4, 5, 6}
		line.SetColor(1)

		copied := line.Copy().(*Line)

		if copied == line {
			t.Error("Copy should create a new instance")
		}

		if copied.Start[0] != 1 || copied.Start[1] != 2 || copied.Start[2] != 3 {
			t.Error("Start coordinates not copied correctly")
		}

		if copied.End[0] != 4 || copied.End[1] != 5 || copied.End[2] != 6 {
			t.Error("End coordinates not copied correctly")
		}
	})

	t.Run("Copy Circle", func(t *testing.T) {
		circle := NewCircle()
		circle.Center = []float64{10, 20, 30}
		circle.Radius = 5.0

		copied := circle.Copy().(*Circle)

		if copied.Center[0] != 10 || copied.Center[1] != 20 || copied.Center[2] != 30 {
			t.Error("Center not copied correctly")
		}

		if copied.Radius != 5.0 {
			t.Error("Radius not copied correctly")
		}
	})

	t.Run("Copy Arc", func(t *testing.T) {
		arc := NewArc(nil)
		arc.Circle.Center = []float64{0, 0, 0}
		arc.Circle.Radius = 10.0
		arc.Angle = []float64{0, 90}

		copied := arc.Copy().(*Arc)

		if copied.Circle.Center[0] != 0 {
			t.Error("Arc center not copied")
		}

		if copied.Angle[0] != 0 || copied.Angle[1] != 90 {
			t.Error("Arc angles not copied")
		}
	})

	t.Run("Copy Polyline with Vertices", func(t *testing.T) {
		polyline := NewPolyline()
		polyline.AddVertex(0, 0, 0)
		polyline.AddVertex(1, 0, 0)
		polyline.AddVertex(1, 1, 0)

		copied := polyline.Copy().(*Polyline)

		if len(copied.Vertices) != 3 {
			t.Error("Vertex count not copied correctly")
		}

		if copied.Vertices[0].Coord[0] != 0 {
			t.Error("Vertex 0 not copied")
		}
	})

	t.Run("Copy Insert with Attributes", func(t *testing.T) {
		insert := NewInsert()
		insert.SetBlockName("test_block")
		insert.SetInsertPoint(dxfmath.NewVec3(10, 20, 30))
		insert.SetRotation(45.0)

		attr := NewAttrib()
		attr.tag = "TAG1"
		attr.textValue = "Value1"
		insert.AddAttribute(attr)

		copied := insert.Copy().(*Insert)

		if copied.BlockName() != "test_block" {
			t.Error("Block name not copied")
		}

		if len(copied.GetAllAttributes()) != 1 {
			t.Error("Attributes not copied")
		}

		if copied.GetAllAttributes()[0].tag != "TAG1" {
			t.Error("Attribute tag not copied")
		}
	})

	t.Run("Copy Ellipse", func(t *testing.T) {
		ellipse := NewEllipse()
		ellipse.Center = []float64{100, 200, 0}
		ellipse.Major = 50.0
		ellipse.Minor = 30.0
		ellipse.Start = 0
		ellipse.End = 180

		copied := ellipse.Copy().(*Ellipse)

		if copied.Center[0] != 100 {
			t.Error("Ellipse center not copied")
		}

		if copied.Major != 50.0 || copied.Minor != 30.0 {
			t.Error("Ellipse radii not copied")
		}
	})

	t.Run("Copy Spline", func(t *testing.T) {
		spline := NewSpline(3)
		spline.Controls = [][]float64{
			{0, 0, 0},
			{1, 1, 0},
			{2, 0, 0},
		}
		spline.Knots = []float64{0, 0, 0, 0, 1, 1, 1, 1}

		copied := spline.Copy().(*Spline)

		if len(copied.Controls) != 3 {
			t.Error("Spline controls not copied")
		}

		if len(copied.Knots) != 8 {
			t.Error("Spline knots not copied")
		}
	})

	t.Run("Copy 3DFace", func(t *testing.T) {
		face := New3DFace()
		face.Points[0] = []float64{0, 0, 0}
		face.Points[1] = []float64{1, 0, 0}
		face.Points[2] = []float64{1, 1, 0}
		face.Points[3] = []float64{0, 1, 0}

		copied := face.Copy().(*ThreeDFace)

		if copied.Points[0][0] != 0 {
			t.Error("3DFace points not copied")
		}
	})

	t.Run("Copy Text", func(t *testing.T) {
		text := NewText()
		text.Value = "Hello"
		text.Height = 2.5
		text.Coord1 = []float64{100, 200, 0}

		copied := text.Copy().(*Text)

		if copied.Value != "Hello" {
			t.Error("Text value not copied")
		}

		if copied.Height != 2.5 {
			t.Error("Text height not copied")
		}
	})

	t.Run("Copy Point", func(t *testing.T) {
		point := NewPoint(5, 10, 15)

		copied := point.Copy().(*Point)

		if copied.Coord[0] != 5 || copied.Coord[1] != 10 || copied.Coord[2] != 15 {
			t.Error("Point coordinates not copied")
		}
	})

	t.Run("Copy LwPolyline", func(t *testing.T) {
		lwpolyline := NewLwPolyline(3)
		lwpolyline.Vertices[0] = []float64{0, 0}
		lwpolyline.Vertices[1] = []float64{1, 0}
		lwpolyline.Vertices[2] = []float64{1, 1}
		lwpolyline.Bulges[0] = 0.5

		copied := lwpolyline.Copy().(*LwPolyline)

		if len(copied.Vertices) != 3 {
			t.Error("LwPolyline vertices not copied")
		}

		if copied.Bulges[0] != 0.5 {
			t.Error("LwPolyline bulges not copied")
		}
	})

	t.Run("Copy Solid", func(t *testing.T) {
		solid := NewSolid()
		solid.FirstPoint = []float64{0, 0, 0}
		solid.SecondPoint = []float64{1, 0, 0}
		solid.ThirdPoint = []float64{1, 1, 0}
		solid.FourthPoint = []float64{0, 1, 0}

		copied := solid.Copy().(*Solid)

		if copied.FirstPoint[0] != 0 {
			t.Error("Solid first point not copied")
		}
	})

	t.Run("Copy Trace", func(t *testing.T) {
		trace := NewTrace([]float64{0, 0, 1, 1, 2, 2, 3, 3}, 0, 1.0)

		copied := trace.Copy().(*Trace)

		if copied.Points[0] != 0 {
			t.Error("Trace points not copied")
		}
	})

	t.Run("Copy Attrib", func(t *testing.T) {
		attrib := NewAttrib()
		attrib.tag = "TAG"
		attrib.textValue = "VALUE"
		attrib.insertPoint = []float64{10, 20, 30}

		copied := attrib.Copy().(*Attrib)

		if copied.tag != "TAG" {
			t.Error("Attrib tag not copied")
		}

		if copied.textValue != "VALUE" {
			t.Error("Attrib value not copied")
		}
	})

	t.Run("Copy Vertex", func(t *testing.T) {
		vertex := NewVertex(5, 10, 15)

		copied := vertex.Copy().(*Vertex)

		if copied.Coord[0] != 5 || copied.Coord[1] != 10 || copied.Coord[2] != 15 {
			t.Error("Vertex coordinates not copied")
		}
	})
}
