package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/object"
)

func TestObjectsAsEntitiesTyped(t *testing.T) {
	types := []string{
		"LAYOUT", "PLOTSETTINGS", "RASTERVARIABLES", "WIPEOUTVARIABLES",
		"IDBUFFER", "FIELDLIST", "SPATIAL_FILTER", "GEODATA",
	}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			body := "0\n" + typ + "\n5\nA1\n330\n0\n1\nname\n"
			e, out := loadOne(t, body)
			oe, ok := e.(*ObjectEntity)
			if !ok {
				t.Fatalf("type %T (out:\n%s)", e, out)
			}
			if oe.DXFType() != typ {
				t.Fatalf("DXFType = %q, want %q", oe.DXFType(), typ)
			}
			if !strings.Contains(out, "0\n"+typ+"\n") {
				t.Errorf("record not preserved in\n%s", out)
			}
		})
	}
}

func TestLayoutObjectInEntitiesTyped(t *testing.T) {
	body := "0\nLAYOUT\n5\nL1\n330\n0\n100\nAcDbLayout\n1\nLayout1\n70\n1\n71\n1\n999\nlayout-extra\n"
	e, out := loadOne(t, body)

	oe, ok := e.(*ObjectEntity)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	if oe.DXFType() != "LAYOUT" || oe.Handle() != "L1" {
		t.Fatalf("type=%q handle=%q", oe.DXFType(), oe.Handle())
	}
	layout, ok := oe.TypedObject().(*object.Layout)
	if !ok {
		t.Fatalf("typed object %T", oe.TypedObject())
	}
	if layout.Name() != "Layout1" || layout.Flags() != 1 {
		t.Errorf("name=%q flags=%d", layout.Name(), layout.Flags())
	}
	if !strings.Contains(out, "0\nLAYOUT\n") || !strings.Contains(out, "999\nlayout-extra") {
		t.Errorf("layout record not preserved in\n%s", out)
	}
	// The record must remain inside ENTITIES; it must not be moved to OBJECTS.
	entitiesAt := strings.Index(out, "2\nENTITIES")
	layoutAt := strings.Index(out, "0\nLAYOUT\n")
	objectsAt := strings.Index(out, "2\nOBJECTS")
	if entitiesAt < 0 || layoutAt < entitiesAt {
		t.Fatalf("LAYOUT not in ENTITIES section\n%s", out)
	}
	if objectsAt >= 0 && layoutAt > objectsAt {
		t.Fatalf("LAYOUT moved to OBJECTS section\n%s", out)
	}
}
