package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestPolylineRoundTrip3D(t *testing.T) {
	body := "0\nPOLYLINE\n5\nPL1\n100\nAcDbEntity\n8\n0\n66\n" + f(1) + "\n70\n" + f(32) + "\n10\n0.0\n20\n0.0\n30\n" + f(5.0) + "\n210\n0.0\n220\n0.0\n230\n-1.0\n" +
		"0\nVERTEX\n5\nV1\n100\nAcDbVertex\n10\n0.0\n20\n0.0\n30\n0.0\n42\n" + f(0.5) + "\n" +
		"0\nVERTEX\n5\nV2\n100\nAcDbVertex\n10\n1.0\n20\n1.0\n30\n1.0\n40\n" + f(0.25) + "\n41\n" + f(0.75) + "\n" +
		"0\nSEQEND\n5\nSE1\n"
	e, out := loadOne(t, body)
	p, ok := e.(*entity.Polyline)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if len(p.Vertices) != 2 {
		t.Fatalf("vertices = %d", len(p.Vertices))
	}
	if p.Vertices[0].Bulge != 0.5 {
		t.Errorf("vertex bulge %v", p.Vertices[0].Bulge)
	}
	if p.EndHandle() != "SE1" {
		t.Errorf("end handle = %q", p.EndHandle())
	}
	for _, want := range []string{"66\n1", "70\n32", "VERTEX", "42\n0.5", "40\n0.25", "41\n0.75", "SEQEND"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	// Subentity access + copy
	var copier entity.Copier = p
	cp := copier.Copy().(*entity.Polyline)
	if len(cp.Vertices) != 2 {
		t.Errorf("copy vertices = %d", len(cp.Vertices))
	}
	cp.Vertices[0].Coord[0] = 99
	if p.Vertices[0].Coord[0] == 99 {
		t.Error("copy shares vertex geometry")
	}
}

func TestPolylineSubentitiesInterface(t *testing.T) {
	body := "0\nPOLYLINE\n5\nPL2\n100\nAcDbEntity\n8\n0\n66\n" + f(1) + "\n70\n" + f(0) + "\n" +
		"0\nVERTEX\n5\nV1\n100\nAcDbVertex\n10\n0.0\n20\n0.0\n30\n0.0\n" +
		"0\nVERTEX\n5\nV2\n100\nAcDbVertex\n10\n1.0\n20\n1.0\n30\n0.0\n" +
		"0\nSEQEND\n5\nSE2\n"
	e, _ := loadOne(t, body)
	owner, ok := e.(entity.SubentityOwner)
	if !ok {
		t.Fatal("Polyline must implement SubentityOwner")
	}
	subs := owner.Subentities()
	if len(subs) != 2 {
		t.Fatalf("subentities = %d", len(subs))
	}
	for _, s := range subs {
		if s.DXFType() != "VERTEX" {
			t.Errorf("subentity type %s", s.DXFType())
		}
	}
}
