package export

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/table"
)

func TestSVGExportWithLine(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if result == "" {
		t.Error("Export should return non-empty SVG for drawing with entities")
	}

	if !strings.Contains(result, "<svg") {
		t.Error("Export should contain <svg tag")
	}

	if !strings.Contains(result, "<line") {
		t.Error("Export should contain <line tag for line entity")
	}

	if !strings.Contains(result, "</svg>") {
		t.Error("Export should contain </svg> closing tag")
	}
}

func TestSVGExportWithCircle(t *testing.T) {
	d, _ := drawing.New()
	circle := entity.NewCircle()
	circle.Center = []float64{50, 50, 0}
	circle.Radius = 25
	circle.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(circle)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<circle") {
		t.Error("Export should contain <circle tag for circle entity")
	}
}

func TestSVGExportWithArc(t *testing.T) {
	d, _ := drawing.New()
	arc := entity.NewArc(nil)
	arc.Center = []float64{50, 50, 0}
	arc.Radius = 25
	arc.Angle = []float64{0, 90}
	arc.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(arc)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<path") {
		t.Error("Export should contain <path tag for arc entity")
	}
}

func TestSVGExportWithPoint(t *testing.T) {
	d, _ := drawing.New()
	point := entity.NewPoint(100, 200, 0)
	point.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(point)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<circle") {
		t.Error("Export should contain <circle tag for point entity")
	}
}

func TestSVGExportWithMultipleEntities(t *testing.T) {
	d, _ := drawing.New()

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	circle := entity.NewCircle()
	circle.Center = []float64{50, 50, 0}
	circle.Radius = 25
	circle.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(circle)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<line") {
		t.Error("Export should contain <line tag")
	}

	if !strings.Contains(result, "<circle") {
		t.Error("Export should contain <circle tag")
	}
}

func TestSVGExportBBoxCalculation(t *testing.T) {
	d, _ := drawing.New()

	line1 := entity.NewLine()
	line1.Start = []float64{10, 10, 0}
	line1.End = []float64{50, 50, 0}
	line1.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line1)

	line2 := entity.NewLine()
	line2.Start = []float64{100, 100, 0}
	line2.End = []float64{150, 150, 0}
	line2.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line2)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, `viewBox`) {
		t.Error("Export should contain viewBox attribute")
	}

	if !strings.Contains(result, `width=`) {
		t.Error("Export should contain width attribute")
	}

	if !strings.Contains(result, `height=`) {
		t.Error("Export should contain height attribute")
	}
}

func TestSVGExportBackground(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, `fill="white"`) {
		t.Error("Export should have white background fill")
	}
}

func TestSVGExportStrokeColor(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, `stroke="black"`) {
		t.Error("Export should use black stroke color")
	}
}

func TestSVGExportStrokeWidth(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, `stroke-width="1"`) {
		t.Error("Export should have stroke-width of 1")
	}
}

func TestSVGFormatCoord(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	tests := []struct {
		input     float64
		precision int
		expected  string
	}{
		{1.23456, 2, "1.23"},
		{1.23456, 4, "1.2346"},
		{1.2, 4, "1.2000"},
		{0.0, 4, "0.0000"},
		{-5.5, 1, "-5.5"},
	}

	for _, tt := range tests {
		exporter.SetPrecision(tt.precision)
		result := exporter.formatCoord(tt.input)
		if result != tt.expected {
			t.Errorf("formatCoord(%f, %d) = %s, expected %s", tt.input, tt.precision, result, tt.expected)
		}
	}
}

func TestSVGExportWithLwPolyline(t *testing.T) {
	d, _ := drawing.New()
	lw := entity.NewLwPolyline(3)
	lw.Vertices[0] = []float64{0, 0}
	lw.Vertices[1] = []float64{100, 0}
	lw.Vertices[2] = []float64{100, 100}
	lw.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(lw)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<polygon") {
		t.Error("Export should contain <polygon tag for LwPolyline")
	}
}

func TestSVGExportWithClosedLwPolyline(t *testing.T) {
	d, _ := drawing.New()
	lw := entity.NewLwPolyline(3)
	lw.Vertices[0] = []float64{0, 0}
	lw.Vertices[1] = []float64{100, 0}
	lw.Vertices[2] = []float64{100, 100}
	lw.Closed = true
	lw.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(lw)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<polygon") {
		t.Error("Export should contain <polygon tag for closed LwPolyline")
	}
}

func TestSVGExportWithPolyline(t *testing.T) {
	d, _ := drawing.New()
	pl := entity.NewPolyline()
	pl.AddVertex(0, 0, 0)
	pl.AddVertex(100, 0, 0)
	pl.AddVertex(100, 100, 0)
	pl.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(pl)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<polyline") {
		t.Error("Export should contain <polyline tag for Polyline")
	}
}

func TestSVGCoordTransform(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{10, 20, 0}
	line.End = []float64{30, 40, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<line") {
		t.Error("Should produce line element")
	}
}

func TestSVGExportMinimumSize(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{50, 50, 0}
	line.End = []float64{51, 51, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, `width="`) {
		t.Error("Should have width attribute")
	}

	if !strings.Contains(result, `height="`) {
		t.Error("Should have height attribute")
	}
}

func TestSVGExportPadding(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{100, 100, 0}
	line.End = []float64{200, 200, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, `viewBox`) {
		t.Error("Should have viewBox for coordinate system")
	}
}

func TestSVGGetColor(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))

	color := exporter.getColor(line)
	if color != "black" {
		t.Errorf("Expected 'black' color for entity with layer, got '%s'", color)
	}
}

func TestSVGArcAngleConversion(t *testing.T) {
	d, _ := drawing.New()
	arc := entity.NewArc(nil)
	arc.Center = []float64{50, 50, 0}
	arc.Radius = 25
	arc.Angle = []float64{45, 135}
	arc.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(arc)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<path") {
		t.Error("Should produce path element for arc")
	}

	if !strings.Contains(result, "A") {
		t.Error("SVG path should contain arc command (A)")
	}
}

func TestSVGArcLargeArcFlag(t *testing.T) {
	d, _ := drawing.New()
	arc := entity.NewArc(nil)
	arc.Center = []float64{50, 50, 0}
	arc.Radius = 25
	arc.Angle = []float64{0, 270}
	arc.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(arc)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<path") {
		t.Error("Should produce path element for large arc")
	}
}

func TestSVGLineWithNegativeCoordinates(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{-50, -50, 0}
	line.End = []float64{50, 50, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, "<line") {
		t.Error("Should handle negative coordinates")
	}
}

func TestSVGExportEmptyDrawing(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if result != "" {
		t.Error("Export should return empty string for empty drawing")
	}
}

func TestSVGExportNamespace(t *testing.T) {
	d, _ := drawing.New()
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{100, 100, 0}
	line.SetLayer(table.NewLayer("0", 0, nil))
	d.AddEntity(line)

	exporter := NewSVGExporter(d)
	result := exporter.Export()

	if !strings.Contains(result, `xmlns="http://www.w3.org/2000/svg"`) {
		t.Error("Export should include SVG namespace")
	}
}
