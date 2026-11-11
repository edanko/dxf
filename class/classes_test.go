package class

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
)

func TestNewClasses(t *testing.T) {
	classes := New()
	if classes == nil {
		t.Error("New() should return non-nil")
	}
	if cap(classes) != 0 {
		t.Errorf("Expected capacity 0, got %d", cap(classes))
	}
}

func TestClassesAdd(t *testing.T) {
	classes := New()
	class := &Class{}
	classes = append(classes, class)

	if len(classes) != 1 {
		t.Errorf("Expected 1 class, got %d", len(classes))
	}
}

func TestClassFormat(t *testing.T) {
	class := &Class{}
	f := format.NewASCII()
	class.Format(f)
	result := f.Output()

	if !strings.Contains(result, "CLASS") {
		t.Error("Format should contain 'CLASS'")
	}
}

func TestClassesFormat(t *testing.T) {
	classes := New()
	classes = append(classes, &Class{})
	classes = append(classes, &Class{})

	f := format.NewASCII()
	classes.Format(f)
	result := f.Output()

	if !strings.Contains(result, "SECTION") {
		t.Error("Format should contain 'SECTION'")
	}
	if !strings.Contains(result, "CLASSES") {
		t.Error("Format should contain 'CLASSES'")
	}
	if !strings.Contains(result, "ENDSEC") {
		t.Error("Format should contain 'ENDSEC'")
	}
	if strings.Count(result, "CLASS") != 3 {
		t.Errorf("Format should contain 3 CLASS entries (1 for section + 2 classes), got %d", strings.Count(result, "CLASS"))
	}
}

func TestClassesSetHandle(t *testing.T) {
	classes := New()
	classes.SetHandle(nil)
}

func TestClassType(t *testing.T) {
	class := &Class{}
	if class == nil {
		t.Error("Class should be non-nil")
	}
}

func TestClassesMultiple(t *testing.T) {
	classes := New()

	for range 5 {
		classes = append(classes, &Class{})
	}

	if len(classes) != 5 {
		t.Errorf("Expected 5 classes, got %d", len(classes))
	}
}
