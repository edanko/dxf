package object

import (
	"strings"
	"testing"
)

func TestMLeaderStyleCollectionMembers(t *testing.T) {
	dxf := "  0\nDICTIONARY\n  5\n124\n330\nC\n100\nAcDbDictionary\n281\n1\n" +
		"  3\nStandard\n350\n125\n  3\nAnnotative\n350\n132\n"
	d, ok := parseOne(t, dxf).(*Dictionary)
	if !ok {
		t.Fatal("not a dictionary")
	}
	collection, ok := WrapCollection(d, "MLEADERSTYLECOLLECTION").(*MLeaderStyleCollection)
	if !ok {
		t.Fatal("not an MLeaderStyleCollection")
	}
	if got := TypeName(collection); got != "MLEADERSTYLECOLLECTION" {
		t.Fatalf("TypeName = %q", got)
	}
	if collection.Handle() != "124" || collection.Owner != "C" {
		t.Fatalf("handle/owner = %q/%q", collection.Handle(), collection.Owner)
	}
	if got := collection.MemberHandles(); len(got) != 2 || got[0] != "125" || got[1] != "132" {
		t.Fatalf("member handles = %v", got)
	}
	if got, ok := collection.MemberHandle("Annotative"); !ok || got != "132" {
		t.Fatalf("Annotative handle = %q, %v", got, ok)
	}
	members := collection.ResolveMembers(func(handle string) (Object, bool) {
		if handle == "125" {
			return NewRawObject("MLEADERSTYLE", nil), true
		}
		return nil, false
	})
	if len(members) != 1 || members["Standard"] == nil {
		t.Fatalf("resolved members = %v", members)
	}
	out := formatObject(t, collection)
	for _, want := range []string{"DICTIONARY", "124", "330\nC", "Standard", "125", "Annotative", "132"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestDirectCollectionRecord(t *testing.T) {
	dxf := "  0\nMLINESTYLECOLLECTION\n  5\n17\n330\nC\n100\nAcDbDictionary\n281\n1\n" +
		"  3\nStandard\n350\n18\n"
	o := parseOne(t, dxf)
	collection, ok := o.(*MLineStyleCollection)
	if !ok {
		t.Fatalf("type = %T", o)
	}
	if got := TypeName(collection); got != "MLINESTYLECOLLECTION" {
		t.Fatalf("TypeName = %q", got)
	}
	if got := collection.MemberHandles(); len(got) != 1 || got[0] != "18" {
		t.Fatalf("member handles = %v", got)
	}
	out := formatObject(t, collection)
	if !strings.Contains(out, "0\nMLINESTYLECOLLECTION") || !strings.Contains(out, "350\n18") {
		t.Fatalf("collection record not preserved:\n%s", out)
	}
}

func TestLayerFilterTyped(t *testing.T) {
	dxf := "  0\nLAYER_FILTER\n  5\nLF\n330\nOWNER\n100\nAcDbFilter\n" +
		"100\nAcDbLayerFilter\n330\nL1\n330\nL2\n  1\nWalls\n  1\nDoors\n" +
		"  3\nname==\"Walls\"\n999\nunknown\n"
	o := parseOne(t, dxf)
	filter, ok := o.(*LayerFilter)
	if !ok {
		t.Fatalf("type = %T", o)
	}
	if got := filter.LayerHandles(); len(got) != 2 || got[0] != "L1" || got[1] != "L2" {
		t.Fatalf("layer handles = %v", got)
	}
	if got := filter.LayerNames(); len(got) != 2 || got[0] != "Walls" || got[1] != "Doors" {
		t.Fatalf("layer names = %v", got)
	}
	if got := filter.FilterExpressions(); len(got) != 1 || got[0] != `name=="Walls"` {
		t.Fatalf("filter expressions = %v", got)
	}
	out := formatObject(t, filter)
	if !strings.Contains(out, "999\nunknown") {
		t.Fatalf("unknown filter tag lost:\n%s", out)
	}
	reparsed, ok := reparse(t, filter).(*LayerFilter)
	if !ok {
		t.Fatal("reparsed object is not a LayerFilter")
	}
	if got := reparsed.LayerHandles(); len(got) != 2 {
		t.Fatalf("reparsed layer handles = %v", got)
	}
}
