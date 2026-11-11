package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestDimensionRoundTrip(t *testing.T) {
	body := "0\nDIMENSION\n5\nD1\n100\nAcDbEntity\n8\n0\n100\nAcDbDimension\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(0.0) + "\n11\n" + f(3.0) + "\n21\n" + f(4.0) + "\n31\n" + f(0.0) + "\n70\n" + f(1) + "\n1\nMyDimText\n3\nStandard\n71\n" + f(5) + "\n72\n" + f(1) + "\n41\n" + f(1.0) + "\n42\n" + f(12.5) + "\n100\nAcDbAlignedDimension\n50\n" + f(30.0) + "\n100\nAcDbRotatedDimension\n13\n" + f(0.0) + "\n23\n" + f(0.0) + "\n33\n" + f(0.0) + "\n14\n" + f(10.0) + "\n24\n" + f(5.0) + "\n34\n" + f(0.0) + "\n"
	e, out := loadOne(t, body)
	dim, ok := e.(*entity.Dimension)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if dim.ActualMeasurement() != 12.5 {
		t.Errorf("measurement = %v", dim.ActualMeasurement())
	}
	for _, want := range []string{"MyDimText", "42\n12.5", "13\n0", "14\n10"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestArcDimensionVariant(t *testing.T) {
	body := "0\nARC_DIMENSION\n5\nD2\n100\nAcDbEntity\n8\n0\n100\nAcDbDimension\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(0.0) + "\n11\n" + f(3.0) + "\n21\n" + f(4.0) + "\n31\n" + f(0.0) + "\n70\n" + f(1) + "\n3\nStandard\n71\n" + f(5) + "\n72\n" + f(1) + "\n41\n" + f(1.0) + "\n42\n" + f(7.5) + "\n100\nAcDbArcDimension\n50\n" + f(10.0) + "\n51\n" + f(80.0) + "\n"
	e, out := loadOne(t, body)
	dim, ok := e.(*entity.Dimension)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if !dim.IsArcLength() {
		t.Errorf("not arc dimension: type=%v", dim.DXFType())
	}
	if !strings.Contains(out, "ARC_DIMENSION") {
		t.Errorf("variant name lost in\n%s", out)
	}
}

func TestLargeRadialDimensionVariant(t *testing.T) {
	body := "0\nLARGE_RADIAL_DIMENSION\n5\nD3\n100\nAcDbEntity\n8\n0\n100\nAcDbDimension\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(0.0) + "\n11\n" + f(3.0) + "\n21\n" + f(4.0) + "\n31\n" + f(0.0) + "\n70\n" + f(1) + "\n3\nStandard\n71\n" + f(5) + "\n72\n" + f(1) + "\n41\n" + f(1.0) + "\n42\n" + f(9.0) + "\n100\nAcDbLargeRadialDimension\n15\n" + f(2.0) + "\n25\n" + f(3.0) + "\n35\n" + f(0.0) + "\n16\n" + f(4.0) + "\n26\n" + f(5.0) + "\n36\n" + f(0.0) + "\n"
	e, out := loadOne(t, body)
	dim, ok := e.(*entity.Dimension)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if dim.DXFType() != "LARGE_RADIAL_DIMENSION" {
		t.Errorf("dxf type = %q", dim.DXFType())
	}
	if !strings.Contains(out, "LARGE_RADIAL_DIMENSION") {
		t.Errorf("variant name lost in\n%s", out)
	}
}
