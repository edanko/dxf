package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/edanko/dxf/path"
)

// gradientBoundary is a 100x100 square path used as the fill boundary.
func gradientBoundary() *path.Path {
	p := &path.Path{}
	p.MoveTo(0, 0)
	p.LineTo(100, 0)
	p.LineTo(100, 100)
	p.LineTo(0, 100)
	p.Close()
	return p
}

func renderGradientPDF(t *testing.T, g *GradientFill) (string, []string) {
	t.Helper()
	var buf bytes.Buffer
	r, err := NewPDFRenderer(&buf, 100, 100)
	if err != nil {
		t.Fatalf("NewPDFRenderer: %v", err)
	}
	if err := r.Begin(100, 100); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := r.FillGradient(gradientBoundary(), g); err != nil {
		t.Fatalf("FillGradient: %v", err)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	dr, ok := r.(DiagnosticRenderer)
	if !ok {
		t.Fatalf("PDF renderer does not implement DiagnosticRenderer")
	}
	return buf.String(), dr.Diagnostics()
}

func TestPDFGradientAxialStopsInOrder(t *testing.T) {
	g := &GradientFill{
		Type: "linear",
		Stops: []GradientStop{
			{Offset: 0, Color: [3]float64{1, 0, 0}},
			{Offset: 0.5, Color: [3]float64{0, 1, 0}},
			{Offset: 1, Color: [3]float64{0, 0, 1}},
		},
	}
	pdf, diags := renderGradientPDF(t, g)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	for _, want := range []string{
		"/ShadingType 2",
		"/Coords [0.0000 50.0000 100.0000 50.0000]",
		"/FunctionType 3",
		"/Bounds [0.5000]",
		"/Encode [0 1 0 1]",
		"/C0 [1.0000 0.0000 0.0000] /C1 [0.0000 1.0000 0.0000]",
		"/C0 [0.0000 1.0000 0.0000] /C1 [0.0000 0.0000 1.0000]",
		"/Sh1 sh",
	} {
		if !strings.Contains(pdf, want) {
			t.Fatalf("PDF missing %q:\n%s", want, pdf)
		}
	}
	// Stops must appear in ascending offset order: red->green precedes
	// green->blue.
	first := strings.Index(pdf, "/C0 [1.0000 0.0000 0.0000]")
	second := strings.Index(pdf, "/C0 [0.0000 1.0000 0.0000]")
	if first < 0 || second < 0 || first > second {
		t.Fatalf("gradient stops out of order (first=%d second=%d)", first, second)
	}
}

func TestPDFGradientRadialCenterRadius(t *testing.T) {
	g := &GradientFill{
		Type:   "radial",
		Center: [2]float64{25, 30},
		Radius: 10,
		Stops: []GradientStop{
			{Offset: 0, Color: [3]float64{0.2, 0.4, 0.6}},
			{Offset: 1, Color: [3]float64{0.9, 0.8, 0.7}},
		},
	}
	pdf, diags := renderGradientPDF(t, g)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	for _, want := range []string{
		"/ShadingType 3",
		"/Coords [25.0000 30.0000 0.0000 25.0000 30.0000 10.0000]",
		"/FunctionType 2",
		"/Extend [true true]",
		"/Sh1 sh",
	} {
		if !strings.Contains(pdf, want) {
			t.Fatalf("PDF missing %q:\n%s", want, pdf)
		}
	}
}

func TestPDFGradientUnsupportedColorSpaceFallsBack(t *testing.T) {
	g := &GradientFill{
		ColorSpace: "CMYK",
		Stops: []GradientStop{
			{Offset: 0, Color: [3]float64{0.1, 0.2, 0.3}},
			{Offset: 1, Color: [3]float64{0.9, 0.5, 0.1}},
		},
	}
	pdf, diags := renderGradientPDF(t, g)
	if len(diags) != 1 {
		t.Fatalf("diagnostics = %v, want exactly one", diags)
	}
	if !strings.Contains(diags[0], `unsupported color space "CMYK"`) ||
		!strings.Contains(diags[0], "solid fill") {
		t.Fatalf("diagnostic does not name the color space/fallback: %q", diags[0])
	}
	if strings.Contains(pdf, "/ShadingType") || strings.Contains(pdf, "/Sh1 sh") {
		t.Fatalf("fallback PDF still references a shading:\n%s", pdf)
	}
	if !strings.HasPrefix(pdf, "%PDF-1.4") || !strings.HasSuffix(pdf, "%%EOF\n") {
		t.Fatalf("fallback output is not a complete PDF")
	}
	// First stop color used as a solid fill.
	wantColor := pdfColor(gradientRGBA([3]float64{0.1, 0.2, 0.3}))
	if !strings.Contains(pdf, wantColor+" rg") || !strings.Contains(pdf, "f\n") {
		t.Fatalf("fallback PDF missing solid fill with first stop:\n%s", pdf)
	}
	// Deterministic across runs.
	pdf2, diags2 := renderGradientPDF(t, g)
	if pdf != pdf2 || strings.Join(diags, "\n") != strings.Join(diags2, "\n") {
		t.Fatalf("fallback rendering is not deterministic")
	}
}

func TestPDFGradientInvalidStopsFallBack(t *testing.T) {
	g := &GradientFill{
		Stops: []GradientStop{
			{Offset: 0.5, Color: [3]float64{0.3, 0.3, 0.3}},
			{Offset: 0.5, Color: [3]float64{0.7, 0.7, 0.7}},
		},
	}
	pdf, diags := renderGradientPDF(t, g)
	if len(diags) != 1 || !strings.Contains(diags[0], "invalid color stops") {
		t.Fatalf("diagnostics = %v, want invalid-stops entry", diags)
	}
	if strings.Contains(pdf, "/ShadingType") {
		t.Fatalf("invalid gradient emitted a shading:\n%s", pdf)
	}
	if !strings.HasPrefix(pdf, "%PDF-1.4") || !strings.HasSuffix(pdf, "%%EOF\n") {
		t.Fatalf("fallback output is not a complete PDF")
	}
}
