package dxf

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

// per-type round trips: build source DXF, load, save, reload, compare typed
// fields via accessors and emitted codes.

func loadOne(t *testing.T, body string) (entity.Entity, string) {
	t.Helper()
	src := "0\nSECTION\n2\nENTITIES\n" + body + "0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities()) != 1 {
		t.Fatalf("entities = %d", len(d.Entities()))
	}
	var out bytes.Buffer
	if _, err := d.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	second, err := FromStringData(out.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Entities()) != 1 {
		t.Fatalf("reload entities = %d", len(second.Entities()))
	}
	return second.Entities()[0], out.String()
}

func f(v float64) string { return fmt.Sprintf("%v", v) }

func TestGeomRoundTripPoint(t *testing.T) {
	body := "0\nPOINT\n5\nP1\n100\nAcDbEntity\n8\n0\n10\n1.5\n20\n2.5\n30\n3.5\n39\n" + f(4.0) + "\n210\n0.0\n220\n0.0\n230\n-1.0\n"
	e, out := loadOne(t, body)
	p, ok := e.(*entity.Point)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if p.Coord[0] != 1.5 || p.Coord[2] != 3.5 {
		t.Errorf("coord %v", p.Coord)
	}
	for _, want := range []string{"10\n1.5", "20\n2.5", "30\n3.5"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestGeomRoundTripCircle(t *testing.T) {
	body := "0\nCIRCLE\n5\nC1\n100\nAcDbEntity\n8\n0\n10\n1.0\n20\n2.0\n30\n0.0\n40\n" + f(7.25) + "\n39\n" + f(2.0) + "\n210\n0.0\n220\n0.0\n230\n-1.0\n"
	e, out := loadOne(t, body)
	c, ok := e.(*entity.Circle)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if c.Radius != 7.25 {
		t.Errorf("radius %v", c.Radius)
	}
	for _, want := range []string{"40\n7.25", "210\n0.0", "230\n-1.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestGeomRoundTripArc(t *testing.T) {
	body := "0\nARC\n5\nA1\n100\nAcDbEntity\n8\n0\n10\n1.0\n20\n2.0\n30\n0.0\n40\n" + f(5.0) + "\n50\n" + f(30.0) + "\n51\n" + f(270.0) + "\n210\n0.0\n220\n0.0\n230\n-1.0\n"
	e, out := loadOne(t, body)
	a, ok := e.(*entity.Arc)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if a.Angle[0] != 30.0 || a.Angle[1] != 270.0 {
		t.Errorf("angles %v", a.Angle)
	}
	for _, want := range []string{"50\n30", "51\n270"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestGeomRoundTripEllipse(t *testing.T) {
	body := "0\nELLIPSE\n5\nE1\n100\nAcDbEntity\n8\n0\n10\n1.0\n20\n2.0\n30\n0.0\n40\n" + f(6.5) + "\n41\n" + f(0.5) + "\n50\n" + f(0.25) + "\n51\n" + f(3.0) + "\n210\n0.0\n220\n0.0\n230\n-1.0\n"
	e, out := loadOne(t, body)
	el, ok := e.(*entity.Ellipse)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if el.Major != 6.5 || el.Minor != 0.5 || el.Start != 0.25 || el.End != 3.0 {
		t.Errorf("params %v %v %v %v", el.Major, el.Minor, el.Start, el.End)
	}
	for _, want := range []string{"40\n6.5", "41\n0.5", "50\n0.25", "51\n3"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestGeomRoundTripSpline(t *testing.T) {
	body := "0\nSPLINE\n5\nS1\n100\nAcDbEntity\n8\n0\n70\n" + f(32) + "\n71\n" + f(3) + "\n72\n" + f(0) + "\n73\n" + f(0) + "\n74\n" + f(0) + "\n42\n" + f(1e-10) + "\n43\n" + f(1e-10) + "\n40\n0.0\n40\n0.0\n40\n0.0\n40\n0.0\n10\n0.0\n20\n0.0\n10\n1.0\n20\n1.0\n11\n1.0\n21\n1.0\n210\n0.0\n220\n0.0\n230\n-1.0\n"
	e, out := loadOne(t, body)
	sp, ok := e.(*entity.Spline)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if sp.Degree != 3 {
		t.Errorf("degree %v", sp.Degree)
	}
	for _, want := range []string{"71\n3", "10\n0.0", "10\n1.0", "11\n1.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestGeomRoundTripLwPolyline(t *testing.T) {
	body := "0\nLWPOLYLINE\n5\nL1\n100\nAcDbEntity\n8\n0\n90\n" + f(3) + "\n70\n" + f(1) + "\n43\n" + f(0.5) + "\n10\n0.0\n20\n0.0\n10\n1.0\n20\n1.0\n42\n" + f(0.25) + "\n10\n2.0\n20\n0.0\n40\n" + f(0.1) + "\n41\n" + f(0.2) + "\n"
	e, out := loadOne(t, body)
	lw, ok := e.(*entity.LwPolyline)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if len(lw.Vertices) != 3 {
		t.Fatalf("vertices = %d", len(lw.Vertices))
	}
	if len(lw.Bulges) < 2 || lw.Bulges[1] != 0.25 {
		t.Errorf("bulges %v", lw.Bulges)
	}
	for _, want := range []string{"90\n3", "70\n1", "43\n0.5", "42\n0.25"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}
