package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestSolid3dRoundTrip(t *testing.T) {
	body := "0\n3DSOLID\n5\nS1\n100\nAcDbEntity\n8\n0\n100\nAcDbModelerGeometry\n70\n" + f(0) + "\n100\nAcDb3dSolid\n70\n" + f(22300) + "\n91\n" + f(544) + "\n90\n" + f(2) + "\n10\nacis/entity/box 11.0 0.0 0.0\n10\nend-of-acis\n40\n" + f(12.5) + "\n41\n" + f(30.0) + "\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(3.0) + "\n"
	e, out := loadOne(t, body)
	s, ok := e.(*entity.Solid3d)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	if s.Version != 22300 {
		t.Errorf("version = %d", s.Version)
	}
	lines := s.SATLines()
	if len(lines) != 2 || lines[0] != "acis/entity/box 11.0 0.0 0.0" {
		t.Fatalf("sat lines = %q", lines)
	}
	for _, want := range []string{"3DSOLID", "90\n2", "10\nacis/entity/box 11.0 0.0 0.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestRegionRoundTrip(t *testing.T) {
	body := "0\nREGION\n5\nR1\n100\nAcDbEntity\n8\n0\n100\nAcDbModelerGeometry\n70\n" + f(0) + "\n100\nAcDbRegion\n90\n" + f(1) + "\n10\nacis/entity/region 1.0 0.0 0.0\n"
	e, out := loadOne(t, body)
	r, ok := e.(*entity.Region)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	if len(r.SATLines()) != 1 {
		t.Fatalf("sat lines = %q", r.SATLines())
	}
	if !strings.Contains(out, "10\nacis/entity/region 1.0 0.0 0.0") {
		t.Errorf("payload not re-emitted in\n%s", out)
	}
}

func TestBodyRoundTrip(t *testing.T) {
	body := "0\nBODY\n5\nB1\n100\nAcDbEntity\n8\n0\n100\nAcDbModelerGeometry\n70\n" + f(20800) + "\n71\n" + f(1) + "\n90\n" + f(1) + "\n10\nacis/entity/body 0.0 0.0 0.0\n"
	e, out := loadOne(t, body)
	b, ok := e.(*entity.Body)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	if b.Version != 20800 || b.Flags != 1 {
		t.Errorf("version=%d flags=%d", b.Version, b.Flags)
	}
	if len(b.SATLines()) != 1 {
		t.Fatalf("sat lines = %q", b.SATLines())
	}
	if !strings.Contains(out, "10\nacis/entity/body 0.0 0.0 0.0") {
		t.Errorf("payload not re-emitted in\n%s", out)
	}
}
