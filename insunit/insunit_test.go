package insunit

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
)

func TestUnitConstants(t *testing.T) {
	if Unitless != 0 {
		t.Errorf("Unitless should be 0, got %d", Unitless)
	}
	if Inches != 1 {
		t.Errorf("Inches should be 1, got %d", Inches)
	}
	if Feet != 2 {
		t.Errorf("Feet should be 2, got %d", Feet)
	}
	if Miles != 3 {
		t.Errorf("Miles should be 3, got %d", Miles)
	}
	if Millimeters != 4 {
		t.Errorf("Millimeters should be 4, got %d", Millimeters)
	}
	if Centimeters != 5 {
		t.Errorf("Centimeters should be 5, got %d", Centimeters)
	}
	if Meters != 6 {
		t.Errorf("Meters should be 6, got %d", Meters)
	}
	if Kilometers != 7 {
		t.Errorf("Kilometers should be 7, got %d", Kilometers)
	}
	if Microinches != 8 {
		t.Errorf("Microinches should be 8, got %d", Microinches)
	}
	if Mils != 9 {
		t.Errorf("Mils should be 9, got %d", Mils)
	}
	if Yards != 10 {
		t.Errorf("Yards should be 10, got %d", Yards)
	}
	if Angstroms != 11 {
		t.Errorf("Angstroms should be 11, got %d", Angstroms)
	}
	if Nanometers != 12 {
		t.Errorf("Nanometers should be 12, got %d", Nanometers)
	}
	if Microns != 13 {
		t.Errorf("Microns should be 13, got %d", Microns)
	}
	if Decimeters != 14 {
		t.Errorf("Decimeters should be 14, got %d", Decimeters)
	}
	if Decameters != 15 {
		t.Errorf("Decameters should be 15, got %d", Decameters)
	}
	if Hectometers != 16 {
		t.Errorf("Hectometers should be 16, got %d", Hectometers)
	}
	if Gigameters != 17 {
		t.Errorf("Gigameters should be 17, got %d", Gigameters)
	}
	if Astronomical != 18 {
		t.Errorf("Astronomical should be 18, got %d", Astronomical)
	}
	if LightYears != 19 {
		t.Errorf("LightYears should be 19, got %d", LightYears)
	}
	if Parsecs != 20 {
		t.Errorf("Parsecs should be 20, got %d", Parsecs)
	}
}

func TestUnitType(t *testing.T) {
	var u Unit = 5
	if u != 5 {
		t.Errorf("Unit should hold value 5, got %d", u)
	}
}

func TestUnitFormat(t *testing.T) {
	f := format.NewASCII()
	Inches.Format(f)
	result := f.Output()
	if !strings.Contains(result, "1") {
		t.Errorf("Format output for Inches should contain '1', got %q", result)
	}
}

func TestTypeConstants(t *testing.T) {
	if Scientific != -1 {
		t.Errorf("Scientific should be -1, got %d", Scientific)
	}
	if Decimal != 0 {
		t.Errorf("Decimal should be 0, got %d", Decimal)
	}
	if Engineering != 1 {
		t.Errorf("Engineering should be 1, got %d", Engineering)
	}
	if Architectural != 2 {
		t.Errorf("Architectural should be 2, got %d", Architectural)
	}
	if Fractional != 3 {
		t.Errorf("Fractional should be 3, got %d", Fractional)
	}
	if WindowsDesktop != 4 {
		t.Errorf("WindowsDesktop should be 4, got %d", WindowsDesktop)
	}
}

func TestTypeFormat(t *testing.T) {
	f := format.NewASCII()
	Decimal.Format(f)
	result := f.Output()
	if !strings.Contains(result, "2") {
		t.Errorf("Format output for Decimal should contain '2', got %q", result)
	}

	f = format.NewASCII()
	Architectural.Format(f)
	result = f.Output()
	if !strings.Contains(result, "4") {
		t.Errorf("Format output for Architectural should contain '4', got %q", result)
	}
}
