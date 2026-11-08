package header

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/insunit"
)

func TestNewHeader(t *testing.T) {
	header := New()
	if header == nil {
		t.Error("New() should return non-nil")
	}
	if header.Version != "AC1015" {
		t.Errorf("Expected version 'AC1015', got %s", header.Version)
	}
	if len(header.InsBase) != 3 {
		t.Errorf("Expected InsBase length 3, got %d", len(header.InsBase))
	}
	if len(header.ExtMin) != 3 {
		t.Errorf("Expected ExtMin length 3, got %d", len(header.ExtMin))
	}
	if len(header.ExtMax) != 3 {
		t.Errorf("Expected ExtMax length 3, got %d", len(header.ExtMax))
	}
	if header.LtScale != 1.0 {
		t.Errorf("Expected LtScale 1.0, got %f", header.LtScale)
	}
}

func TestHeaderVersion(t *testing.T) {
	header := New()
	if header.Version != "AC1015" {
		t.Errorf("Expected version 'AC1015', got %s", header.Version)
	}

	header.Version = "AC1024"
	if header.Version != "AC1024" {
		t.Errorf("Expected version 'AC1024', got %s", header.Version)
	}
}

func TestHeaderInsBase(t *testing.T) {
	header := New()
	if header.InsBase[0] != 0 || header.InsBase[1] != 0 || header.InsBase[2] != 0 {
		t.Errorf("Expected InsBase [0,0,0], got [%f,%f,%f]", header.InsBase[0], header.InsBase[1], header.InsBase[2])
	}

	header.InsBase[0] = 10
	if header.InsBase[0] != 10 {
		t.Errorf("Expected InsBase[0] = 10, got %f", header.InsBase[0])
	}
}

func TestHeaderInsUnit(t *testing.T) {
	header := New()
	if header.InsUnit != insunit.Unitless {
		t.Errorf("Expected InsUnit Unitless, got %v", header.InsUnit)
	}

	header.InsUnit = insunit.Millimeters
	if header.InsUnit != insunit.Millimeters {
		t.Errorf("Expected InsUnit Millimeters, got %v", header.InsUnit)
	}
}

func TestHeaderInsLUnit(t *testing.T) {
	header := New()
	if header.InsLUnit != insunit.Decimal {
		t.Errorf("Expected InsLUnit Decimal, got %v", header.InsLUnit)
	}

	header.InsLUnit = insunit.Engineering
	if header.InsLUnit != insunit.Engineering {
		t.Errorf("Expected InsLUnit Engineering, got %v", header.InsLUnit)
	}
}

func TestHeaderExtMin(t *testing.T) {
	header := New()
	if header.ExtMin[0] != 0 || header.ExtMin[1] != 0 || header.ExtMin[2] != 0 {
		t.Errorf("Expected ExtMin [0,0,0], got [%f,%f,%f]", header.ExtMin[0], header.ExtMin[1], header.ExtMin[2])
	}
}

func TestHeaderExtMax(t *testing.T) {
	header := New()
	if header.ExtMax[0] != 0 || header.ExtMax[1] != 0 || header.ExtMax[2] != 0 {
		t.Errorf("Expected ExtMax [0,0,0], got [%f,%f,%f]", header.ExtMax[0], header.ExtMax[1], header.ExtMax[2])
	}
}

func TestHeaderLtScale(t *testing.T) {
	header := New()
	if header.LtScale != 1.0 {
		t.Errorf("Expected LtScale 1.0, got %f", header.LtScale)
	}

	header.LtScale = 2.0
	if header.LtScale != 2.0 {
		t.Errorf("Expected LtScale 2.0, got %f", header.LtScale)
	}
}

func TestHeaderSetHandle(t *testing.T) {
	header := New()
	header.SetHandle(handle.NewHandleGenerator())

	if header.handseed == "" {
		t.Error("handseed should be set after SetHandle()")
	}
}

func TestHeaderFormat(t *testing.T) {
	header := New()
	header.SetHandle(handle.NewHandleGenerator())

	f := format.NewASCII()
	header.Format(f)
	result := f.Output()

	if !strings.Contains(result, "SECTION") {
		t.Error("Format should contain 'SECTION'")
	}
	if !strings.Contains(result, "HEADER") {
		t.Error("Format should contain 'HEADER'")
	}
	if !strings.Contains(result, "ENDSEC") {
		t.Error("Format should contain 'ENDSEC'")
	}
	if !strings.Contains(result, "$ACADVER") {
		t.Error("Format should contain '$ACADVER'")
	}
}
