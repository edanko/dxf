package export

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
)

func TestNewSVGExporter(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	if exporter == nil {
		t.Error("NewSVGExporter should return non-nil")
	}
	if exporter.precision != 4 {
		t.Errorf("Expected precision 4, got %d", exporter.precision)
	}
	if exporter.drawing != d {
		t.Error("Drawing should be set")
	}
}

func TestSVGExporterSetPrecision(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	// Valid precision
	exporter.SetPrecision(6)
	if exporter.precision != 6 {
		t.Errorf("Expected precision 6, got %d", exporter.precision)
	}

	// Invalid precision (too low)
	exporter.SetPrecision(0)
	if exporter.precision != 6 {
		t.Errorf("Precision should not change for 0, got %d", exporter.precision)
	}

	// Invalid precision (too high)
	exporter.SetPrecision(15)
	if exporter.precision != 6 {
		t.Errorf("Precision should not change for 15, got %d", exporter.precision)
	}

	// Boundary values
	exporter.SetPrecision(1)
	if exporter.precision != 1 {
		t.Errorf("Expected precision 1, got %d", exporter.precision)
	}

	exporter.SetPrecision(10)
	if exporter.precision != 10 {
		t.Errorf("Expected precision 10, got %d", exporter.precision)
	}
}

func TestSVGExporterExportEmpty(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	result := exporter.Export()
	if result != "" {
		t.Errorf("Expected empty string for empty drawing, got %q", result)
	}
}

func TestSVGExporterExportWithPrecision(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)
	exporter.SetPrecision(6)

	result := exporter.Export()
	if result != "" {
		t.Errorf("Expected empty string for empty drawing, got %q", result)
	}
}

func TestSVGExporterStructure(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	result := exporter.Export()

	// Empty drawing should return empty
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestSVGExporterType(t *testing.T) {
	d, _ := drawing.New()
	var exporter any = NewSVGExporter(d)

	if _, ok := exporter.(*SVGExporter); !ok {
		t.Error("NewSVGExporter should return *SVGExporter")
	}
}

func TestSVGExporterNilDrawing(t *testing.T) {
	exporter := NewSVGExporter(nil)
	result := exporter.Export()
	if result != "" {
		t.Errorf("Expected empty string for nil drawing, got '%s'", result)
	}
}

func TestSVGExporterMultipleExports(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	result1 := exporter.Export()
	result2 := exporter.Export()

	if result1 != result2 {
		t.Error("Multiple exports should return same result")
	}
}

func TestSVGExporterWithCustomPrecision(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	for p := 1; p <= 10; p++ {
		exporter.SetPrecision(p)
		if exporter.precision != p {
			t.Errorf("Expected precision %d, got %d", p, exporter.precision)
		}
	}
}

func TestSVGExporterInvalidPrecisionValues(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	// Test various invalid precision values
	invalidPrecisions := []int{-1, 0, 11, 100, -100}

	for _, p := range invalidPrecisions {
		original := exporter.precision
		exporter.SetPrecision(p)
		if p > 0 && p <= 10 {
			// Should change
		} else {
			// Should not change
			if exporter.precision != original {
				t.Errorf("Precision should not change for invalid value %d", p)
			}
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSVGExporterExportMethodExists(t *testing.T) {
	d, _ := drawing.New()
	exporter := NewSVGExporter(d)

	// Test that Export method exists and returns string
	result := exporter.Export()
	if result == "" {
		// Empty drawing is OK
		t.Log("Export returned empty string for empty drawing (expected)")
	} else {
		// Non-empty result should be valid
		if !strings.HasPrefix(result, "<svg") && result != "" {
			t.Error("Export should return SVG or empty string")
		}
	}
}
