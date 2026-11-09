package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func TestMeshRoundTrip(t *testing.T) {
	body := "0\nMESH\n5\nM1\n100\nAcDbEntity\n8\n0\n100\nAcDbSubDMesh\n71\n" + f(2) + "\n72\n" + f(1) + "\n91\n" + f(3) + "\n92\n" + f(4) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n10\n" + f(1.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n10\n" + f(1.0) + "\n20\n" + f(1.0) + "\n30\n" + f(0.0) + "\n10\n" + f(0.0) + "\n20\n" + f(1.0) + "\n30\n" + f(0.0) + "\n93\n" + f(1) + "\n90\n" + f(4) + "\n90\n" + f(0) + "\n90\n" + f(1) + "\n90\n" + f(2) + "\n90\n" + f(3) + "\n"
	e, out := loadOne(t, body)
	m, ok := e.(*entity.Mesh)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	if m.SubdivisionLevels() != 3 || m.BlendCrease() != 1 {
		t.Errorf("levels=%d blend=%d", m.SubdivisionLevels(), m.BlendCrease())
	}
	if len(m.Vertices()) != 4 {
		t.Fatalf("vertices = %d", len(m.Vertices()))
	}
	v2 := m.Vertices()[2]
	if v2.X() != 1.0 || v2.Y() != 1.0 {
		t.Errorf("vertex2 = %v", []float64{v2.X(), v2.Y()})
	}
	faces := m.Faces()
	if len(faces) != 1 || len(faces[0]) != 4 || faces[0][3] != 3 {
		t.Errorf("faces = %v", faces)
	}
	for _, want := range []string{"MESH", "92\n4", "93\n1"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestHelixRoundTrip(t *testing.T) {
	body := "0\nHELIX\n5\nH1\n100\nAcDbEntity\n8\n0\n100\nAcDbHelix\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(3.0) + "\n11\n" + f(0.0) + "\n21\n" + f(0.0) + "\n31\n" + f(1.0) + "\n40\n" + f(5.0) + "\n90\n" + f(3) + "\n41\n" + f(4.0) + "\n290\n" + f(0) + "\n"
	e, out := loadOne(t, body)
	h, ok := e.(*entity.Helix)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	sp := h.StartPoint()
	if sp.X() != 1.0 || sp.Y() != 2.0 || sp.Z() != 3.0 {
		t.Errorf("start = %v", []float64{sp.X(), sp.Y(), sp.Z()})
	}
	if h.Radius() != 5.0 || h.Turns() != 3 || h.TurnHeight() != 4.0 {
		t.Errorf("radius=%v turns=%v height=%v", h.Radius(), h.Turns(), h.TurnHeight())
	}
	if h.Handedness() != 0 {
		t.Errorf("handedness = %d", h.Handedness())
	}
	for _, want := range []string{"HELIX", "40\n5.0000000000000000", "90\n3"} {
		if !strings.Contains(out, want) && !strings.Contains(out, strings.ReplaceAll(want, "5.0000000000000000", "5")) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestExtrudedSurfaceRoundTrip(t *testing.T) {
	body := "0\nEXTRUDED_SURFACE\n5\nE1\n100\nAcDbEntity\n8\n0\n100\nAcDbExtrudedSurface\n70\n" + f(22300) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(1.0) + "\n42\n" + f(0.5) + "\n45\n" + f(0.25) + "\n72\n" + f(1) + "\n"
	e, out := loadOne(t, body)
	es, ok := e.(*entity.ExtrudedSurface)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	if es.Version != 22300 {
		t.Errorf("version = %d", es.Version)
	}
	sv := es.SweepVector
	if sv.Z() != 1.0 {
		t.Errorf("sweep vector = %v", []float64{sv.X(), sv.Y(), sv.Z()})
	}
	if !es.Solid {
		t.Errorf("solid flag not parsed")
	}
	if es.DraftAngle != 0.5 || es.TwistAngle != 0.25 {
		t.Errorf("draft=%v twist=%v", es.DraftAngle, es.TwistAngle)
	}
	if !strings.Contains(out, "EXTRUDED_SURFACE") {
		t.Errorf("record name not emitted in\n%s", out)
	}
	_ = dxfmath.NewVec3(0, 0, 0)
}

func TestPlaneSurfaceRoundTrip(t *testing.T) {
	body := "0\nPLANESURFACE\n5\nP1\n100\nAcDbEntity\n8\n0\n100\nAcDbPlaneSurface\n70\n" + f(22300) + "\n"
	e, out := loadOne(t, body)
	ps, ok := e.(*entity.PlaneSurface)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	if ps.Version != 22300 {
		t.Errorf("version = %d", ps.Version)
	}
	if !strings.Contains(out, "PLANESURFACE") {
		t.Errorf("record name not emitted in\n%s", out)
	}
}

func TestLoftedRevolvedSweptRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"LOFTED_SURFACE", "0\nLOFTED_SURFACE\n5\nL1\n100\nAcDbEntity\n8\n0\n100\nAcDbLoftedSurface\n70\n" + f(22300) + "\n42\n" + f(0.5) + "\n"},
		{"REVOLVED_SURFACE", "0\nREVOLVED_SURFACE\n5\nR1\n100\nAcDbEntity\n8\n0\n100\nAcDbRevolvedSurface\n70\n" + f(22300) + "\n42\n" + f(1.5) + "\n"},
		{"SWEPT_SURFACE", "0\nSWEPT_SURFACE\n5\nS1\n100\nAcDbEntity\n8\n0\n100\nAcDbSweptSurface\n70\n" + f(22300) + "\n42\n" + f(0.1) + "\n"},
	}
	for _, c := range cases {
		e, out := loadOne(t, c.body)
		if !strings.Contains(out, c.name) {
			t.Errorf("%s: record name not re-emitted in\n%s", c.name, out)
		}
		switch v := e.(type) {
		case *entity.LoftedSurface:
			if v.Version != 22300 || v.DraftAngle != 0.5 {
				t.Errorf("%s: version=%v draft=%v", c.name, v.Version, v.DraftAngle)
			}
		case *entity.RevolvedSurface:
			if v.Version != 22300 {
				t.Errorf("%s: version=%v", c.name, v.Version)
			}
		case *entity.SweptSurface:
			if v.Version != 22300 || v.DraftAngle != 0.1 {
				t.Errorf("%s: version=%v draft=%v", c.name, v.Version, v.DraftAngle)
			}
		default:
			t.Errorf("%s: loaded %T", c.name, e)
		}
	}
}
