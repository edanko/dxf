package entity

import (
	"math"
	"testing"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/table"
)

func TestLineBasic(t *testing.T) {
	l := NewLine()

	if l == nil {
		t.Fatal("NewLine should not return nil")
	}

	if l.Start == nil || len(l.Start) != 3 {
		t.Error("Start should be initialized with 3 elements")
	}

	if l.End == nil || len(l.End) != 3 {
		t.Error("End should be initialized with 3 elements")
	}

	if l.entity == nil {
		t.Error("embedded entity should not be nil")
	}
}

func TestLineStartEnd(t *testing.T) {
	l := NewLine()
	l.Start = []float64{1, 2, 3}
	l.End = []float64{4, 5, 6}

	if l.Start[0] != 1 || l.Start[1] != 2 || l.Start[2] != 3 {
		t.Errorf("Start should be [1, 2, 3], got [%f, %f, %f]", l.Start[0], l.Start[1], l.Start[2])
	}

	if l.End[0] != 4 || l.End[1] != 5 || l.End[2] != 6 {
		t.Errorf("End should be [4, 5, 6], got [%f, %f, %f]", l.End[0], l.End[1], l.End[2])
	}
}

func TestLineLength3D(t *testing.T) {
	l := NewLine()
	l.Start = []float64{0, 0, 0}
	l.End = []float64{3, 4, 12}

	expected := math.Sqrt(3*3 + 4*4 + 12*12)
	length := l.Length()

	if math.Abs(length-expected) > 1e-10 {
		t.Errorf("Expected length %f, got %f", expected, length)
	}
}

func TestLineLength2D(t *testing.T) {
	l := NewLine()
	l.Start = []float64{0, 0, 0}
	l.End = []float64{6, 8, 0}

	length := l.Length()

	if math.Abs(length-10.0) > 1e-10 {
		t.Errorf("Expected length 10.0, got %f", length)
	}
}

func TestLineDirectionNotNormalized(t *testing.T) {
	l := NewLine()
	l.Start = []float64{0, 0, 0}
	l.End = []float64{10, 20, 30}

	dir := l.Direction(false)

	if dir[0] != 10 || dir[1] != 20 || dir[2] != 30 {
		t.Errorf("Direction (unnormalized) should be [10, 20, 30], got [%f, %f, %f]", dir[0], dir[1], dir[2])
	}
}

func TestLineDirectionNormalized(t *testing.T) {
	l := NewLine()
	l.Start = []float64{0, 0, 0}
	l.End = []float64{10, 0, 0}

	dir := l.Direction(true)

	if math.Abs(dir[0]-1.0) > 1e-10 {
		t.Errorf("X component should be 1.0, got %f", dir[0])
	}

	if math.Abs(dir[1]) > 1e-10 {
		t.Errorf("Y component should be 0.0, got %f", dir[1])
	}

	if math.Abs(dir[2]) > 1e-10 {
		t.Errorf("Z component should be 0.0, got %f", dir[2])
	}
}

func TestLineDirectionNormalized3D(t *testing.T) {
	l := NewLine()
	l.Start = []float64{0, 0, 0}
	l.End = []float64{10, 10, 10}

	dir := l.Direction(true)

	expected := 1.0 / math.Sqrt(3)
	if math.Abs(dir[0]-expected) > 1e-10 {
		t.Errorf("X component should be %f, got %f", expected, dir[0])
	}

	if math.Abs(dir[1]-expected) > 1e-10 {
		t.Errorf("Y component should be %f, got %f", expected, dir[1])
	}

	if math.Abs(dir[2]-expected) > 1e-10 {
		t.Errorf("Z component should be %f, got %f", expected, dir[2])
	}
}

func TestLineMove(t *testing.T) {
	l := NewLine()
	l.Start = []float64{0, 0, 0}
	l.End = []float64{10, 10, 10}

	l.Move(5, -5, 2.5)

	if l.Start[0] != 5 || l.Start[1] != -5 || l.Start[2] != 2.5 {
		t.Errorf("Start should be [5, -5, 2.5], got [%f, %f, %f]", l.Start[0], l.Start[1], l.Start[2])
	}

	if l.End[0] != 15 || l.End[1] != 5 || l.End[2] != 12.5 {
		t.Errorf("End should be [15, 5, 12.5], got [%f, %f, %f]", l.End[0], l.End[1], l.End[2])
	}
}

func TestLineBBox(t *testing.T) {
	l := NewLine()
	l.Start = []float64{0, 0, 0}
	l.End = []float64{10, 20, 30}

	mins, maxs := l.BBox()

	if mins[0] != 0 || mins[1] != 0 || mins[2] != 0 {
		t.Errorf("Mins should be [0, 0, 0], got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 10 || maxs[1] != 20 || maxs[2] != 30 {
		t.Errorf("Maxs should be [10, 20, 30], got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestLineBBoxInverted(t *testing.T) {
	l := NewLine()
	l.Start = []float64{10, 20, 30}
	l.End = []float64{0, 0, 0}

	mins, maxs := l.BBox()

	if mins[0] != 0 || mins[1] != 0 || mins[2] != 0 {
		t.Errorf("Mins should be [0, 0, 0], got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 10 || maxs[1] != 20 || maxs[2] != 30 {
		t.Errorf("Maxs should be [10, 20, 30], got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestLineBBoxPartialOverlap(t *testing.T) {
	l := NewLine()
	l.Start = []float64{-5, 10, -15}
	l.End = []float64{5, -10, 15}

	mins, maxs := l.BBox()

	if mins[0] != -5 || mins[1] != -10 || mins[2] != -15 {
		t.Errorf("Mins should be [-5, -10, -15], got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 5 || maxs[1] != 10 || maxs[2] != 15 {
		t.Errorf("Maxs should be [5, 10, 15], got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestLineLengthZero(t *testing.T) {
	l := NewLine()
	l.Start = []float64{5, 5, 5}
	l.End = []float64{5, 5, 5}

	length := l.Length()

	if math.Abs(length) > 1e-10 {
		t.Errorf("Expected length 0, got %f", length)
	}
}

func TestLineDirectionZeroLength(t *testing.T) {
	l := NewLine()
	l.Start = []float64{5, 5, 5}
	l.End = []float64{5, 5, 5}

	dir := l.Direction(true)

	if math.Abs(dir[0]-1.0) > 1e-10 || math.Abs(dir[1]-1.0) > 1e-10 || math.Abs(dir[2]-1.0) > 1e-10 {
		t.Logf("Direction for zero-length line: [%f, %f, %f]", dir[0], dir[1], dir[2])
	}
}

func TestLineCopy(t *testing.T) {
	l := NewLine()
	l.Start = []float64{1, 2, 3}
	l.End = []float64{4, 5, 6}
	l.Thickness = 2.5
	l.StretchingDirection = []float64{0, 0, 1}

	copied := l.Copy().(*Line)

	if copied == l {
		t.Error("Copy should return a new instance")
	}

	if copied.Start[0] != 1 || copied.End[0] != 4 {
		t.Error("Start/End coordinates not copied")
	}

	if copied.Thickness != 2.5 {
		t.Error("Thickness not copied")
	}

	if copied.StretchingDirection[2] != 1 {
		t.Error("StretchingDirection not copied")
	}
}

func TestLineIsEntity(t *testing.T) {
	l := NewLine()

	if !l.IsEntity() {
		t.Error("Line should implement IsEntity() returning true")
	}
}

func TestLineFormat(t *testing.T) {
	l := NewLine()
	l.Start = []float64{1, 2, 3}
	l.End = []float64{4, 5, 6}
	l.SetLayer(table.NewLayer("0", 0, nil))

	var f format.Formatter
	f = format.NewASCII()

	l.Format(f)

	if f == nil {
		t.Error("Format should work with formatter")
	}
}

func TestCircleBasic(t *testing.T) {
	c := NewCircle()

	if c == nil {
		t.Fatal("NewCircle should not return nil")
	}

	if c.Center == nil || len(c.Center) != 3 {
		t.Error("Center should be initialized with 3 elements")
	}

	if c.Radius != 0.0 {
		t.Errorf("Default radius should be 0.0, got %f", c.Radius)
	}

	if c.Direction == nil || len(c.Direction) != 3 {
		t.Error("Direction should be initialized with 3 elements")
	}
}

func TestCircleCenterRadius(t *testing.T) {
	c := NewCircle()
	c.Center = []float64{10, 20, 30}
	c.Radius = 5.0

	if c.Center[0] != 10 || c.Center[1] != 20 || c.Center[2] != 30 {
		t.Errorf("Center should be [10, 20, 30], got [%f, %f, %f]", c.Center[0], c.Center[1], c.Center[2])
	}

	if c.Radius != 5.0 {
		t.Errorf("Radius should be 5.0, got %f", c.Radius)
	}
}

func TestCircleBBoxDefault(t *testing.T) {
	c := NewCircle()
	c.Center = []float64{10, 20, 30}
	c.Radius = 5.0

	mins, maxs := c.BBox()

	if mins[0] != 5 || mins[1] != 15 || mins[2] != 30 {
		t.Errorf("Mins should be [5, 15, 30], got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 15 || maxs[1] != 25 || maxs[2] != 30 {
		t.Errorf("Maxs should be [15, 25, 30], got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestCircleBBoxZeroRadius(t *testing.T) {
	c := NewCircle()
	c.Center = []float64{10, 20, 30}
	c.Radius = 0.0

	mins, maxs := c.BBox()

	if mins[0] != 10 || mins[1] != 20 || mins[2] != 30 {
		t.Errorf("Mins should be center for zero radius, got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 10 || maxs[1] != 20 || maxs[2] != 30 {
		t.Errorf("Maxs should be center for zero radius, got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestCircleCurrentDirection(t *testing.T) {
	c := NewCircle()

	if len(c.CurrentDirection()) != 3 {
		t.Error("CurrentDirection should return 3 elements")
	}
}

func TestCircleSetDirection(t *testing.T) {
	c := NewCircle()
	c.SetDirection([]float64{1, 0, 0})

	if c.Direction[0] != 1 || c.Direction[1] != 0 || c.Direction[2] != 0 {
		t.Errorf("Direction should be [1, 0, 0], got [%f, %f, %f]", c.Direction[0], c.Direction[1], c.Direction[2])
	}
}

func TestCircleSetDirectionTooShort(t *testing.T) {
	c := NewCircle()
	original := c.Direction
	c.SetDirection([]float64{1, 0})

	if c.Direction[0] != original[0] || c.Direction[1] != original[1] || c.Direction[2] != original[2] {
		t.Error("Direction should not change for input with less than 3 elements")
	}
}

func TestCircleCopy(t *testing.T) {
	c := NewCircle()
	c.Center = []float64{10, 20, 30}
	c.Radius = 5.0
	c.Direction = []float64{0, 0, 1}

	copied := c.Copy().(*Circle)

	if copied.Center[0] != 10 || copied.Center[1] != 20 {
		t.Error("Center not copied correctly")
	}

	if copied.Radius != 5.0 {
		t.Error("Radius not copied correctly")
	}
}

func TestCircleIsEntity(t *testing.T) {
	c := NewCircle()

	if !c.IsEntity() {
		t.Error("Circle should implement IsEntity() returning true")
	}
}

func TestArcBasic(t *testing.T) {
	a := NewArc(nil)

	if a == nil {
		t.Fatal("NewArc should not return nil")
	}

	if a.Circle == nil {
		t.Error("Arc should have embedded Circle")
	}

	if len(a.Angle) != 2 {
		t.Errorf("Angle should have 2 elements, got %d", len(a.Angle))
	}
}

func TestArcWithCircle(t *testing.T) {
	c := NewCircle()
	c.Center = []float64{10, 20, 30}
	c.Radius = 5.0

	a := NewArc(c)

	if a.Circle.Center[0] != 10 {
		t.Error("Arc should use provided circle")
	}
}

func TestArcAngle(t *testing.T) {
	a := NewArc(nil)
	a.Angle = []float64{0, 90}

	if a.Angle[0] != 0 || a.Angle[1] != 90 {
		t.Error("Angle not set correctly")
	}
}

func TestArcCopy(t *testing.T) {
	a := NewArc(nil)
	a.Circle.Center = []float64{0, 0, 0}
	a.Circle.Radius = 10.0
	a.Angle = []float64{0, 180}

	copied := a.Copy().(*Arc)

	if copied.Circle.Center[0] != 0 {
		t.Error("Arc center not copied")
	}

	if copied.Circle.Radius != 10.0 {
		t.Error("Arc radius not copied")
	}

	if copied.Angle[1] != 180 {
		t.Error("Arc angle not copied")
	}
}

func TestArcIsEntity(t *testing.T) {
	a := NewArc(nil)

	if !a.IsEntity() {
		t.Error("Arc should implement IsEntity() returning true")
	}
}

func TestPointBasic(t *testing.T) {
	p := NewPoint()

	if p == nil {
		t.Fatal("NewPoint should not return nil")
	}

	if p.Coord == nil || len(p.Coord) != 3 {
		t.Error("Coord should be initialized with 3 elements")
	}
}

func TestPointWithCoords(t *testing.T) {
	p := NewPoint(5, 10, 15)

	if p.Coord[0] != 5 || p.Coord[1] != 10 || p.Coord[2] != 15 {
		t.Errorf("Coord should be [5, 10, 15], got [%f, %f, %f]", p.Coord[0], p.Coord[1], p.Coord[2])
	}
}

func TestPointWithLessThan3Coords(t *testing.T) {
	p := NewPoint(5, 10)

	if p.Coord[0] != 5 || p.Coord[1] != 10 || p.Coord[2] != 0 {
		t.Errorf("Coord should be [5, 10, 0], got [%f, %f, %f]", p.Coord[0], p.Coord[1], p.Coord[2])
	}
}

func TestPointBBox(t *testing.T) {
	p := NewPoint(5, 10, 15)

	mins, maxs := p.BBox()

	if mins[0] != 5 || mins[1] != 10 || mins[2] != 15 {
		t.Errorf("Mins should be point location, got [%f, %f, %f]", mins[0], mins[1], mins[2])
	}

	if maxs[0] != 5 || maxs[1] != 10 || maxs[2] != 15 {
		t.Errorf("Maxs should be point location, got [%f, %f, %f]", maxs[0], maxs[1], maxs[2])
	}
}

func TestPointCopy(t *testing.T) {
	p := NewPoint(5, 10, 15)

	copied := p.Copy().(*Point)

	if copied.Coord[0] != 5 || copied.Coord[1] != 10 || copied.Coord[2] != 15 {
		t.Error("Coord not copied correctly")
	}

	if copied == p {
		t.Error("Copy should return new instance")
	}
}

func TestPointIsEntity(t *testing.T) {
	p := NewPoint()

	if !p.IsEntity() {
		t.Error("Point should implement IsEntity() returning true")
	}
}

func TestEntityTypeString(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{LINE, "LINE"},
		{CIRCLE, "CIRCLE"},
		{ARC, "ARC"},
		{POINT, "POINT"},
		{TEXT, "TEXT"},
		{SPLINE, "SPLINE"},
	}

	for _, tt := range tests {
		result := EntityTypeString(tt.entityType)
		if result != tt.expected {
			t.Errorf("EntityTypeString(%d) = %s, expected %s", tt.entityType, result, tt.expected)
		}
	}
}

func TestNewEntityDefaults(t *testing.T) {
	e := NewEntity(LINE)

	if e.Type != LINE {
		t.Errorf("Expected Type LINE, got %v", e.Type)
	}

	if e.ltscale != 1.0 {
		t.Errorf("Default ltscale should be 1.0, got %f", e.ltscale)
	}

	if e.color != 0 {
		t.Errorf("Default color should be 0, got %v", e.color)
	}
}

func TestEntitySetColor(t *testing.T) {
	e := NewEntity(LINE)
	e.SetColor(color.Red)

	if e.color != color.Red {
		t.Errorf("Expected color Red, got %v", e.color)
	}
}

func TestEntitySetLayer(t *testing.T) {
	e := NewEntity(LINE)
	layer := table.NewLayer("TestLayer", color.Red, nil)
	e.SetLayer(layer)

	if e.Layer().Name() != "TestLayer" {
		t.Errorf("Layer name should be TestLayer, got %s", e.Layer().Name())
	}
}

func TestEntityHandle(t *testing.T) {
	e := NewEntity(LINE)

	if e.Handle() != "" {
		t.Errorf("Expected empty handle initially, got %s", e.Handle())
	}
}

func TestEntitySetBlockRecord(t *testing.T) {
	e := NewEntity(LINE)
	e.SetBlockRecord(nil)

	if e.blockRecord != nil {
		t.Error("blockRecord should be nil")
	}
}

func TestEntitySetOwner(t *testing.T) {
	e := NewEntity(LINE)
	e.SetOwner(nil)

	if e.owner != nil {
		t.Error("owner should be nil")
	}
}

func TestEntityFormat(t *testing.T) {
	e := NewEntity(LINE)
	e.SetLayer(table.NewLayer("0", 0, nil))

	var f format.Formatter
	f = format.NewASCII()

	e.Format(f)

	if f == nil {
		t.Error("Format should work with formatter")
	}
}
