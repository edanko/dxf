package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestMPolygonRoundTrip(t *testing.T) {
	body := "0\nMPOLYGON\n5\nP1\n100\nAcDbEntity\n8\n0\n100\nAcDbMPolygon\n70\n" + f(3) + "\n92\n" + f(1) + "\n72\n" + f(1) + "\n93\n" + f(4) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(4.0) + "\n10\n" + f(0.0) + "\n20\n" + f(4.0) + "\n78\n" + f(0) + "\n63\n" + f(256) + "\n11\n" + f(0.0) + "\n21\n" + f(0.0) + "\n"
	e, out := loadOne(t, body)
	h, ok := e.(*entity.Hatch)
	if !ok {
		t.Fatalf("type %T", e)
	}
	paths := h.BoundaryPaths()
	if len(paths) != 1 {
		t.Fatalf("paths = %d (out:\n%s)", len(paths), out)
	}
	pp, ok := paths[0].(*entity.PolylinePath)
	if !ok {
		t.Fatalf("path type %T", paths[0])
	}
	if !pp.IsClosed() {
		t.Errorf("closed = false")
	}
	if pp.PathFlags() != 3 {
		t.Errorf("path flags = %d", pp.PathFlags())
	}
	if len(pp.Vertices()) != 4 {
		t.Errorf("vertices = %d", len(pp.Vertices()))
	}
	for _, want := range []string{"MPOLYGON", "93\n4", "72\n1"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestMPolygonMultiplePaths(t *testing.T) {
	body := "0\nMPOLYGON\n5\nP2\n100\nAcDbEntity\n8\n0\n100\nAcDbMPolygon\n70\n" + f(3) + "\n92\n" + f(1) + "\n72\n" + f(0) + "\n93\n" + f(4) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(4.0) + "\n10\n" + f(0.0) + "\n20\n" + f(4.0) + "\n70\n" + f(3) + "\n92\n" + f(0) + "\n72\n" + f(0) + "\n93\n" + f(3) + "\n10\n" + f(1.0) + "\n20\n" + f(1.0) + "\n10\n" + f(2.0) + "\n20\n" + f(1.0) + "\n10\n" + f(1.5) + "\n20\n" + f(2.0) + "\n78\n" + f(0) + "\n63\n" + f(256) + "\n11\n" + f(0.0) + "\n21\n" + f(0.0) + "\n"
	e, out := loadOne(t, body)
	h, ok := e.(*entity.Hatch)
	if !ok {
		t.Fatalf("type %T", e)
	}
	paths := h.BoundaryPaths()
	if len(paths) != 2 {
		t.Fatalf("paths = %d (out:\n%s)", len(paths), out)
	}
	p0 := paths[0].(*entity.PolylinePath)
	p1 := paths[1].(*entity.PolylinePath)
	if !p0.IsClosed() || p1.IsClosed() {
		t.Errorf("closed flags = %v %v", p0.IsClosed(), p1.IsClosed())
	}
	if len(p0.Vertices()) != 4 || len(p1.Vertices()) != 3 {
		t.Errorf("vertex counts = %d %d", len(p0.Vertices()), len(p1.Vertices()))
	}
}
