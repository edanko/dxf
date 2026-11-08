package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/math"
)

func TestSVGBackend(t *testing.T) {
	var buf bytes.Buffer
	svg := NewSVGBackend(&buf)

	// Test basic rendering
	ctx := NewRenderContext()
	svg.WriteSVGHeader()

	// Test path creation
	err := svg.BeginPath(ctx)
	if err != nil {
		t.Fatalf("BeginPath failed: %v", err)
	}

	// Test path operations
	err = svg.MoveTo(10, 10)
	if err != nil {
		t.Fatalf("MoveTo failed: %v", err)
	}

	err = svg.LineTo(100, 10)
	if err != nil {
		t.Fatalf("LineTo failed: %v", err)
	}

	err = svg.SetStrokeColor(color.Blue)
	if err != nil {
		t.Fatalf("SetStrokeColor failed: %v", err)
	}

	err = svg.SetLineWidth(2.0)
	if err != nil {
		t.Fatalf("SetLineWidth failed: %v", err)
	}

	err = svg.EndPath(ctx)
	if err != nil {
		t.Fatalf("EndPath failed: %v", err)
	}

	svg.WriteSVGFooter()

	svg.Flush()
	result := buf.String()

	// Verify expected SVG content
	expectedContent := "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"0.000000\" height=\"0.000000\" viewBox=\"0 0 0.000000 0.000000\">\n<path d=\"M10.000,10.000L100.000,10.000\" stroke=\"rgb(0,0,255)\" stroke-width=\"2.000000\" />\n</svg>"

	// Normalize result (remove extra whitespace)
	result = strings.ReplaceAll(result, "\n", "")
	result = strings.ReplaceAll(result, "\t", "")
	expectedContent = strings.ReplaceAll(expectedContent, "\n", "")
	expectedContent = strings.ReplaceAll(expectedContent, "\t", "")

	if result != expectedContent {
		t.Errorf("SVG content mismatch\nExpected: %s\nGot: %s", expectedContent, result)
	}

	// Test transform
	matrix := math.Identity()
	err = svg.SetTransform(&matrix)
	if err != nil {
		t.Fatalf("SetTransform failed: %v", err)
	}
}

func TestRenderContext(t *testing.T) {
	ctx := NewRenderContext()

	// Test default values
	if ctx.LineWidth != 1.0 {
		t.Errorf("Expected LineWidth 1.0, got %f", ctx.LineWidth)
	}

	if ctx.Scale != 1.0 {
		t.Errorf("Expected Scale 1.0, got %f", ctx.Scale)
	}

	if ctx.BackgroundColor != color.White {
		t.Errorf("Expected White background, got %v", ctx.BackgroundColor)
	}

	// Test identity matrix
	identity := math.Identity()
	if !ctx.Transform.IsEqual(identity, 1e-10) {
		t.Error("Expected identity transform")
	}
}

func TestColorNumberToRGB(t *testing.T) {
	tests := []struct {
		input    color.ColorNumber
		expected struct{ R, G, B int }
	}{
		{color.Red, struct{ R, G, B int }{255, 0, 0}},
		{color.Green, struct{ R, G, B int }{0, 255, 0}},
		{color.Blue, struct{ R, G, B int }{0, 0, 255}},
		{color.White, struct{ R, G, B int }{255, 255, 255}},
		{color.ByLayer, struct{ R, G, B int }{0, 0, 0}}, // Default case
	}

	for _, test := range tests {
		result := colorNumberToRGB(test.input)
		if result != test.expected {
			t.Errorf("For color %v, expected RGB %v, got %v", test.input, test.expected, result)
		}
	}
}

func TestArrowTypeConstants(t *testing.T) {
	if ArrowNone != 0 {
		t.Errorf("ArrowNone should be 0, got %d", ArrowNone)
	}
	if ArrowClosedBlank != 1 {
		t.Errorf("ArrowClosedBlank should be 1, got %d", ArrowClosedBlank)
	}
	if ArrowClosedFilled != 2 {
		t.Errorf("ArrowClosedFilled should be 2, got %d", ArrowClosedFilled)
	}
	if ArrowOpen != 3 {
		t.Errorf("ArrowOpen should be 3, got %d", ArrowOpen)
	}
}

func TestNewArrow(t *testing.T) {
	arrow := NewArrow(ArrowClosedFilled, 10.0, 45.0)

	if arrow.ArrowType != ArrowClosedFilled {
		t.Errorf("Expected ArrowType %d, got %d", ArrowClosedFilled, arrow.ArrowType)
	}
	if arrow.Size != 10.0 {
		t.Errorf("Expected Size 10.0, got %f", arrow.Size)
	}
	if arrow.Angle != 45.0 {
		t.Errorf("Expected Angle 45.0, got %f", arrow.Angle)
	}
	if arrow.Vertices == nil {
		t.Error("Vertices should not be nil")
	}
}

func TestNewArrowNone(t *testing.T) {
	arrow := NewArrow(ArrowNone, 10.0, 0)

	if arrow.Vertices != nil {
		t.Error("ArrowNone should have nil vertices")
	}
}

func TestArrowVec3Operations(t *testing.T) {
	v1 := NewVec3(1, 2, 3)
	v2 := NewVec3(4, 5, 6)

	sum := v1.Add(v2)
	if sum.X != 5 || sum.Y != 7 || sum.Z != 9 {
		t.Errorf("Add failed: expected (5,7,9), got (%f,%f,%f)", sum.X, sum.Y, sum.Z)
	}

	scaled := v1.Mul(2)
	if scaled.X != 2 || scaled.Y != 4 || scaled.Z != 6 {
		t.Errorf("Mul failed: expected (2,4,6), got (%f,%f,%f)", scaled.X, scaled.Y, scaled.Z)
	}
}

func TestArrowRotate(t *testing.T) {
	arrow := NewArrow(ArrowClosedFilled, 1.0, 0)

	// Rotate 180 degrees around origin
	arrow.Rotate(math.Pi)

	// After 180 degree rotation, point at (1,0) should be at (-1,0)
	if len(arrow.Vertices) >= 2 {
		// Vertices should have been transformed
		if arrow.Vertices[0].X == 1 && arrow.Vertices[0].Y == 0 {
			t.Error("Rotate should have transformed vertices")
		}
	}
}

func TestArrowTranslate(t *testing.T) {
	arrow := NewArrow(ArrowClosedFilled, 1.0, 0)
	originalX := arrow.Vertices[0].X

	arrow.Translate(NewVec3(10, 20, 0))

	if arrow.Vertices[0].X != originalX+10 {
		t.Errorf("Translate X failed: expected %f, got %f", originalX+10, arrow.Vertices[0].X)
	}
	if arrow.Vertices[0].Y != 20 {
		t.Errorf("Translate Y failed: expected 20, got %f", arrow.Vertices[0].Y)
	}
}

func TestGetArrowTypeName(t *testing.T) {
	tests := []struct {
		arrowType ArrowType
		expected  string
	}{
		{ArrowNone, "NONE"},
		{ArrowClosedBlank, "CLOSED_BLANK"},
		{ArrowClosedFilled, "CLOSED_FILLED"},
		{ArrowOpen, "OPEN"},
		{ArrowOpen30, "OPEN_30"},
		{ArrowOpen90, "OPEN_90"},
		{ArrowCircle, "CIRCLE"},
		{ArrowDot, "DOT"},
		{ArrowArchTick, "ARCH_TICK"},
		{ArrowOblique, "OBLIQUE"},
		{ArrowBox, "UNKNOWN"},
		{ArrowBoxFilled, "UNKNOWN"},
		{ArrowDiamond, "UNKNOWN"},
		{ArrowDotSmall, "UNKNOWN"},
	}

	for _, test := range tests {
		result := GetArrowTypeName(test.arrowType)
		if result != test.expected {
			t.Errorf("GetArrowTypeName(%d) = %s, expected %s", test.arrowType, result, test.expected)
		}
	}
}
