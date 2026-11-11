package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestHatchRoundTrip(t *testing.T) {
	body := "0\nHATCH\n5\nH1\n100\nAcDbEntity\n8\n0\n100\nAcDbHatch\n70\n" + f(1) + "\n71\n" + f(0) + "\n75\n" + f(0) + "\n76\n" + f(1) + "\n52\n" + f(0) + "\n41\n" + f(1.0) + "\n77\n" + f(0) + "\n78\n" + f(1) + "\n53\n" + f(45.0) + "\n43\n" + f(0.0) + "\n44\n" + f(0.0) + "\n45\n" + f(0.125) + "\n46\n" + f(0.0) + "\n79\n" + f(2) + "\n49\n" + f(0.125) + "\n49\n" + f(-0.125) + "\n2\nANSI31\n88\n" + f(0) + "\n70\n" + f(0) + "\n91\n" + f(1) + "\n92\n" + f(2) + "\n72\n" + f(1) + "\n73\n" + f(1) + "\n44\n" + f(0.0) + "\n97\n" + f(0) + "\n93\n" + f(4) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n42\n" + f(0.5) + "\n10\n" + f(10.0) + "\n20\n" + f(0.0) + "\n42\n" + f(0.0) + "\n10\n" + f(10.0) + "\n20\n" + f(10.0) + "\n42\n" + f(0.0) + "\n10\n" + f(0.0) + "\n20\n" + f(10.0) + "\n42\n" + f(0.0) + "\n97\n" + f(0) + "\n75\n" + f(0) + "\n76\n" + f(1) + "\n98\n" + f(1) + "\n30\n" + f(0.0) + "\n"
	e, out := loadOne(t, body)
	h, ok := e.(*entity.Hatch)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if h.PatternName() != "ANSI31" {
		t.Errorf("pattern = %q", h.PatternName())
	}
	paths := h.BoundaryPaths()
	if len(paths) != 1 {
		t.Fatalf("paths = %d (out:\n%s)", len(paths), out)
	}
	pp, ok := paths[0].(*entity.PolylinePath)
	if !ok {
		t.Fatalf("path type %T", paths[0])
	}
	if len(pp.Vertices()) != 4 {
		t.Errorf("vertices = %d", len(pp.Vertices()))
	}
	if pp.Vertices()[0].Bulge != 0.5 {
		t.Errorf("vertex0 bulge = %v", pp.Vertices()[0].Bulge)
	}
	if !pp.IsClosed() || !pp.HasBulge() {
		t.Errorf("flags closed=%v bulge=%v", pp.IsClosed(), pp.HasBulge())
	}
	for _, want := range []string{"2\nANSI31", "92\n2", "72\n1", "73\n1", "42\n0.5"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}
