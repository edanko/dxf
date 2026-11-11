package text2path

import (
	"testing"

	"github.com/edanko/dxf/path"
)

func TestNewTextToPath(t *testing.T) {
	tt := NewTextToPath(1.0)
	if tt == nil {
		t.Error("NewTextToPath should return non-nil")
	}
	if tt.Scale != 1.0 {
		t.Errorf("Expected scale 1.0, got %f", tt.Scale)
	}
}

func TestNewTextToPathWithScale(t *testing.T) {
	tt := NewTextToPath(2.5)
	if tt.Scale != 2.5 {
		t.Errorf("Expected scale 2.5, got %f", tt.Scale)
	}
}

func TestTextToPathConvertEmpty(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("")

	if result == nil {
		t.Error("Convert should return non-nil path for empty string")
	}
}

func TestTextToPathConvertWhitespace(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("   ")

	if result == nil {
		t.Error("Convert should return non-nil path for whitespace")
	}
}

func TestTextToPathConvertNewline(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("\n")

	if result == nil {
		t.Error("Convert should return non-nil path for newline")
	}
}

func TestTextToPathConvertSimple(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("A")

	if result == nil {
		t.Error("Convert should return non-nil path for single character")
	}
}

func TestTextToPathConvertMultipleChars(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("ABC")

	if result == nil {
		t.Error("Convert should return non-nil path for multiple characters")
	}
}

func TestTextToPathConvertMixed(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("Hello World")

	if result == nil {
		t.Error("Convert should return non-nil path for mixed text")
	}
}

func TestTextToPathConvertWithNewlines(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("Line1\nLine2")

	if result == nil {
		t.Error("Convert should return non-nil path for text with newlines")
	}
}

func TestTextToPathConvertWithTabs(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("Tab\tHere")

	if result == nil {
		t.Error("Convert should return non-nil path for text with tabs")
	}
}

func TestTextToPathConvertLowercase(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("abc")

	if result == nil {
		t.Error("Convert should return non-nil path for lowercase")
	}
}

func TestTextToPathConvertNumbers(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("12345")

	if result == nil {
		t.Error("Convert should return non-nil path for numbers")
	}
}

func TestTextToPathConvertSymbols(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("!@#$%")

	if result == nil {
		t.Error("Convert should return non-nil path for symbols")
	}
}

func TestTextToPathConvertToDxfPath(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.ConvertToDxfPath()

	if result == "" {
		t.Error("ConvertToDxfPath should return non-empty string")
	}
	if !contains(result, "1.00") {
		t.Error("ConvertToDxfPath should contain scale value")
	}
}

func TestTextToPathGetTextMetricsEmpty(t *testing.T) {
	tt := NewTextToPath(1.0)
	width, height := tt.GetTextMetrics("")

	if width != 0 || height != 0 {
		t.Errorf("Expected 0,0 for empty text, got %f,%f", width, height)
	}
}

func TestTextToPathGetTextMetricsSingleChar(t *testing.T) {
	tt := NewTextToPath(1.0)
	width, height := tt.GetTextMetrics("A")

	if width != 100.0 {
		t.Errorf("Expected width 100.0, got %f", width)
	}
	if height != 120.0 {
		t.Errorf("Expected height 120.0, got %f", height)
	}
}

func TestTextToPathGetTextMetricsMultipleChars(t *testing.T) {
	tt := NewTextToPath(1.0)
	width, height := tt.GetTextMetrics("ABC")

	if width != 300.0 {
		t.Errorf("Expected width 300.0, got %f", width)
	}
	if height != 120.0 {
		t.Errorf("Expected height 120.0, got %f", height)
	}
}

func TestTextToPathGetTextMetricsWithSpaces(t *testing.T) {
	tt := NewTextToPath(1.0)
	width, _ := tt.GetTextMetrics("A B")

	// Spaces should not count toward width
	if width != 200.0 {
		t.Errorf("Expected width 200.0, got %f", width)
	}
}

func TestTextToPathGetTextMetricsWithScale(t *testing.T) {
	tt := NewTextToPath(2.0)
	width, height := tt.GetTextMetrics("AB")

	if width != 400.0 {
		t.Errorf("Expected width 400.0, got %f", width)
	}
	if height != 240.0 {
		t.Errorf("Expected height 240.0, got %f", height)
	}
}

func TestTextToPathGetTextMetricsNewline(t *testing.T) {
	tt := NewTextToPath(1.0)
	width, _ := tt.GetTextMetrics("A\nB")

	// Newlines should not count toward width
	if width != 200.0 {
		t.Errorf("Expected width 200.0, got %f", width)
	}
}

func TestTextToPathCreateCharPath(t *testing.T) {
	tt := NewTextToPath(1.0)

	charPath := tt.createCharPath('A', 0, 0)
	if charPath == nil {
		t.Error("createCharPath should return non-nil path")
	}
}

func TestTextToPathCreateCharPathUnknown(t *testing.T) {
	tt := NewTextToPath(1.0)

	// Test unknown character (will use default box)
	charPath := tt.createCharPath('?', 0, 0)
	if charPath == nil {
		t.Error("createCharPath should return non-nil path for unknown char")
	}
}

func TestTextToPathPathType(t *testing.T) {
	tt := NewTextToPath(1.0)
	result := tt.Convert("Test")

	if _, ok := any(result).(*path.Path); !ok {
		t.Error("Convert should return *path.Path type")
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
