package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestLeaderRoundTrip(t *testing.T) {
	body := "0\nLEADER\n5\nL1\n100\nAcDbEntity\n8\n0\n100\nAcDbLeader\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(5.0) + "\n30\n" + f(0.0) + "\n71\n" + f(1) + "\n72\n" + f(1) + "\n40\n" + f(2.5) + "\n140\n" + f(3.5) + "\n3\nSTANDARD\n"
	e, out := loadOne(t, body)
	ld, ok := e.(*entity.Leader)
	if !ok {
		t.Fatalf("type %T", e)
	}
	pts := ld.GetPoints()
	if len(pts) != 2 {
		t.Errorf("points = %d (out:\n%s)", len(pts), out)
	}
	for _, want := range []string{"100\nAcDbLeader", "71\n1", "40\n2.5", "140\n3.5"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestMLeaderRoundTrip(t *testing.T) {
	body := "0\nMULTILEADER\n5\nML1\n100\nAcDbEntity\n8\n0\n100\nAcDbMLeader\n170\n" + f(2) + "\n340\nMLSTYLE1\n140\n" + f(1.5) + "\n171\n" + f(0) + "\n172\n" + f(0) + "\n41\n" + f(1.0) + "\n290\n" + f(1) + "\n173\n" + f(1) + "\n40\n" + f(2.0) + "\n271\n" + f(1) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n10\n" + f(4.0) + "\n20\n" + f(6.0) + "\n30\n" + f(0.0) + "\n291\n" + f(1) + "\n44\n" + f(3.0) + "\n45\n" + f(1.0) + "\n292\n" + f(0) + "\n175\n" + f(1) + "\n304\nCallout text two lines\n176\n" + f(5) + "\n177\n" + f(3) + "\n"
	e, out := loadOne(t, body)
	ml, ok := e.(*entity.MLeader)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if ml.StyleHandle != "MLSTYLE1" {
		t.Errorf("style handle = %q", ml.StyleHandle)
	}
	if len(ml.Lines) != 1 || len(ml.Lines[0].Vertices) != 2 {
		t.Errorf("lines = %+v", ml.Lines)
	}
	if !strings.Contains(ml.Text.Content, "Callout text") {
		t.Errorf("content = %q", ml.Text.Content)
	}
	for _, want := range []string{"340\nMLSTYLE1", "271\n1", "304\nCallout text", "44\n3"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}
