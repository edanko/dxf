package path

import (
	"testing"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func TestMakePathWithLine(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}

	path, err := MakePath(line)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
	if len(path.Commands()) < 2 {
		t.Errorf("Expected at least 2 commands (MoveTo, LineTo), got %d", len(path.Commands()))
	}
}

func TestMakePathWithLwPolyline(t *testing.T) {
	lwp := entity.NewLwPolyline(4)
	lwp.Closed = true
	lwp.Vertices = [][]float64{
		{0, 0},
		{100, 0},
		{100, 100},
		{0, 100},
	}

	path, err := MakePath(lwp)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
	if !path.IsClosed() {
		t.Error("Closed LWPolyline should produce closed path")
	}
}

func TestMakePathWithPolyline(t *testing.T) {
	pline := entity.NewPolyline()
	pline.AddVertex(0, 0, 0)
	pline.AddVertex(100, 0, 0)
	pline.AddVertex(100, 100, 0)

	path, err := MakePath(pline)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}

func TestMakePathWithCircle(t *testing.T) {
	circle := entity.NewCircle()
	circle.Center = []float64{50, 50, 0}
	circle.Radius = 25

	path, err := MakePath(circle)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}

func TestMakePathWithArc(t *testing.T) {
	circle := entity.NewCircle()
	circle.Center = []float64{0, 0, 0}
	circle.Radius = 100
	arc := entity.NewArc(circle)
	arc.Angle = []float64{0, 90}

	path, err := MakePath(arc)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}

func TestMakePathWithEllipse(t *testing.T) {
	ellipse := entity.NewEllipse()
	ellipse.Center = []float64{50, 50, 0}
	ellipse.Major = 100
	ellipse.Ratio = 0.5
	ellipse.Start = 0
	ellipse.End = 360

	path, err := MakePath(ellipse)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}

func TestMakePathUnsupported(t *testing.T) {
	point := entity.NewPoint()
	point.Coord = []float64{50, 50, 0}

	_, err := MakePath(point)
	if err == nil {
		t.Error("MakePath should fail for unsupported entity type")
	}
}

func TestWithSegmentsOption(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}

	path, err := MakePath(line, WithSegments(32))
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}

func TestWithDistanceOption(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}

	path, err := MakePath(line, WithDistance(0.05))
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}

func TestWithExtrusionOption(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}

	path, err := MakePath(line, WithExtrusion(dxfmath.NewVec3(0, 0, 1)))
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}

func TestWithPrecisionOption(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}

	path, err := MakePath(line, WithPrecision(1e-8))
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()
	if opts.Segments != 16 {
		t.Errorf("Expected default Segments 16, got %d", opts.Segments)
	}
	if opts.Distance != 0.1 {
		t.Errorf("Expected default Distance 0.1, got %f", opts.Distance)
	}
	if opts.Precision != 1e-6 {
		t.Errorf("Expected default Precision 1e-6, got %f", opts.Precision)
	}
}

func TestLwPolylineEmpty(t *testing.T) {
	lwp := entity.NewLwPolyline(0)
	lwp.Closed = false
	lwp.Vertices = [][]float64{}

	path, err := MakePath(lwp)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
	if len(path.Commands()) != 0 {
		t.Errorf("Empty LWPolyline should produce empty path, got %d commands", len(path.Commands()))
	}
}

func TestPolylineEmpty(t *testing.T) {
	pline := entity.NewPolyline()

	path, err := MakePath(pline)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
	if len(path.Commands()) != 0 {
		t.Errorf("Empty Polyline should produce empty path, got %d commands", len(path.Commands()))
	}
}

func TestMultipleOptions(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}

	path, err := MakePath(line,
		WithSegments(32),
		WithDistance(0.05),
		WithExtrusion(dxfmath.NewVec3(0, 0, 1)),
		WithPrecision(1e-8),
	)
	if err != nil {
		t.Fatalf("MakePath failed: %v", err)
	}
	if path == nil {
		t.Fatal("Path should not be nil")
	}
}
