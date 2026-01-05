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
