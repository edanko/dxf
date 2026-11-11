package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func TestMLineRoundTrip(t *testing.T) {
	body := "0\nMLINE\n5\nM1\n100\nAcDbEntity\n8\n0\n100\nAcDbMLine\n2\nSTANDARD\n70\n" + f(1) + "\n71\n" + f(2) + "\n72\n" + f(0) + "\n40\n" + f(4.0) + "\n73\n" + f(2) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n41\n" + f(0.5) + "\n42\n" + f(0.25) + "\n73\n" + f(1) + "\n43\n" + f(1.0) + "\n44\n" + f(0.0) + "\n45\n" + f(0.0) + "\n10\n" + f(10.0) + "\n20\n" + f(5.0) + "\n30\n" + f(0.0) + "\n73\n" + f(1) + "\n43\n" + f(0.0) + "\n44\n" + f(1.0) + "\n45\n" + f(0.0) + "\n74\n" + f(1) + "\n"
	e, out := loadOne(t, body)
	ml, ok := e.(*entity.MLine)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if ml.Scale != 4.0 {
		t.Errorf("scale = %v", ml.Scale)
	}
	if len(ml.Vertices) != 2 {
		t.Fatalf("vertices = %d (out:\n%s)", len(ml.Vertices), out)
	}
	v0 := ml.Vertices[0]
	if v0.Bulge != 0.25 || v0.EndWidth != 0.5 {
		t.Errorf("v0 bulge=%v endw=%v", v0.Bulge, v0.EndWidth)
	}
	if !ml.HasFill {
		t.Errorf("HasFill = false")
	}
	for _, want := range []string{"STANDARD", "74\n1", "42\n0.25"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestMLineTransform(t *testing.T) {
	body := "0\nMLINE\n5\nM1\n100\nAcDbEntity\n8\n0\n100\nAcDbMLine\n2\nSTANDARD\n70\n" + f(0) + "\n71\n" + f(1) + "\n72\n" + f(0) + "\n40\n" + f(1.0) + "\n73\n" + f(1) + "\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(0.0) + "\n74\n" + f(0) + "\n"
	e, _ := loadOne(t, body)
	ml := e.(*entity.MLine)
	m := dxfmath.Translation(5, 6, 0)
	if err := ml.Transform(m); err != nil {
		t.Fatal(err)
	}
	got := ml.Vertices[0].Point
	if got[0] != 6.0 || got[1] != 8.0 {
		t.Errorf("translated point = %v", got)
	}
}
