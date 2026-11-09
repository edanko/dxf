package table

import (
	"errors"
	"strings"
	"testing"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

func newTestLayerTables() Tables {
	lt := NewLineType("Continuous", "Solid line")
	l := NewLayer("0", color.ColorNumber(7), lt)
	st := NewStyle("Standard")
	ts := New([]*LineType{lt}, l, st)
	return ts
}

func TestTableLookupAddRemoveRename(t *testing.T) {
	ts := newTestLayerTables()
	layers := ts[2]

	if _, err := layers.Lookup("0"); err != nil {
		t.Fatalf("lookup 0: %v", err)
	}
	// case-insensitive lookup
	if _, err := layers.Lookup("standard"); err == nil {
		t.Logf("style table lookup is per-table; ok")
	}

	w := NewLayer("Walls", color.ColorNumber(3), NewLineType("Continuous", ""))
	if err := layers.AddChecked(w); err != nil {
		t.Fatalf("AddChecked: %v", err)
	}
	got, err := layers.Lookup("walls")
	if err != nil {
		t.Fatalf("lookup Walls: %v", err)
	}
	if got.Name() != "Walls" {
		t.Fatalf("got %q", got.Name())
	}
	// duplicate rejected
	dup := NewLayer("WALLS", color.ColorNumber(1), nil)
	if err := layers.AddChecked(dup); !errors.Is(err, ErrDuplicateEntry) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	// rename
	if err := layers.Rename("Walls", "Partitions"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if _, err := layers.Lookup("Partitions"); err != nil {
		t.Fatalf("lookup renamed: %v", err)
	}
	if err := layers.Rename("Partitions", "0"); !errors.Is(err, ErrDuplicateEntry) {
		t.Fatalf("expected duplicate on rename collision, got %v", err)
	}
	// remove
	if err := layers.Remove("Partitions"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := layers.Lookup("Partitions"); !errors.Is(err, ErrEntryNotFound) && err == nil {
		t.Fatalf("expected not found after remove")
	}
	if err := layers.Remove("Nope"); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("expected not-found error, got %v", err)
	}
	if layers.Len() != 1 {
		t.Fatalf("Len = %d, want 1", layers.Len())
	}
}

func TestTableHandlesOwnersUnique(t *testing.T) {
	ts := newTestLayerTables()
	hg := handle.NewHandleGenerator()
	ts.SetHandle(hg)

	seen := map[string]bool{}
	for _, tbl := range ts {
		if seen[tbl.Handle()] {
			t.Fatalf("duplicate table handle %s", tbl.Handle())
		}
		seen[tbl.Handle()] = true
		for _, e := range tbl.Entries() {
			if seen[e.Handle()] {
				t.Fatalf("duplicate entry handle %s", e.Handle())
			}
			seen[e.Handle()] = true
			ownerSetter, ok := e.(interface{ Owner() handle.Handler })
			if !ok {
				continue
			}
			if ownerSetter.Owner() == nil || ownerSetter.Owner().Handle() != tbl.Handle() {
				t.Fatalf("entry %s owner is not table %s", e.Name(), tbl.Name())
			}
		}
	}
}

func TestLayerTrueColorRoundTrip(t *testing.T) {
	dxf := `  0
SECTION
  2
TABLES
  0
TABLE
  2
LAYER
  5
A
100
AcDbSymbolTable
 70
1
  0
LAYER
  5
11
330
A
100
AcDbSymbolTableRecord
100
AcDbLayerTableRecord
  2
TrueColor
 70
0
 62
-1
420
16711680
  6
Continuous
370
-3
  0
ENDTAB
  0
ENDSEC
`
	tags, err := lldxf.LoadTags(strings.NewReader(dxf))
	if err != nil {
		t.Fatal(err)
	}
	ts := ParseTables(tags)
	l, err := ts[0].Lookup("TrueColor")
	if err != nil {
		t.Fatal(err)
	}
	lay, ok := l.(*Layer)
	if !ok {
		t.Fatalf("layer with truecolor became %T; typed model expected", l)
	}
	if lay.TrueColor() != 16711680 {
		t.Errorf("TrueColor = %d, want 16711680", lay.TrueColor())
	}
	if lay.Color != -1 {
		t.Errorf("Color(62) = %d, want -1 (true-color marker)", lay.Color)
	}
	f := format.NewASCII()
	lay.Format(f)
	out := f.Output()
	for _, want := range []string{"420", "16711680", "62", "-1"} {
		if !strings.Contains(out, want) {
			t.Errorf("re-emitted record lost %q:\n%s", want, out)
		}
	}
}

func TestLTypeOwnerRoundTrip(t *testing.T) {
	dxf := `  0
SECTION
  2
TABLES
  0
TABLE
  2
LTYPE
  5
B
100
AcDbSymbolTable
 70
1
  0
LTYPE
  5
21
330
B
100
AcDbSymbolTableRecord
100
AcDbLinetypeTableRecord
  2
Continuous
 70
0
  3
Solid line
 72
65
 73
0
 40
0
  0
ENDTAB
  0
ENDSEC
`
	tags, err := lldxf.LoadTags(strings.NewReader(dxf))
	if err != nil {
		t.Fatal(err)
	}
	ts := ParseTables(tags)
	lt, err := ts[0].Lookup("Continuous")
	if err != nil {
		t.Fatal(err)
	}
	f := format.NewASCII()
	lt.Format(f)
	out := f.Output()
	if !strings.Contains(out, "330\nB\n") {
		t.Errorf("owner handle 330/B missing from:\n%s", out)
	}
}

func TestParseTablesEntryCounts(t *testing.T) {
	dxf := `  0
SECTION
  2
TABLES
  0
TABLE
  2
LAYER
  5
A
330
F
100
AcDbSymbolTable
 70
3
  0
LAYER
  5
11
330
A
100
AcDbSymbolTableRecord
100
AcDbLayerTableRecord
  2
0
 70
0
 62
7
  6
Continuous
370
-3
  0
LAYER
  5
12
330
A
100
AcDbSymbolTableRecord
100
AcDbLayerTableRecord
  2
Defpoints
 70
0
 62
7
  6
Continuous
290
0
370
-3
440
50
  0
LAYER
  5
13
330
A
100
AcDbSymbolTableRecord
100
AcDbLayerTableRecord
  2
Walls
 70
4
 62
6
  6
DASHED
  0
ENDTAB
  0
TABLE
  2
LTYPE
  5
B
330
F
100
AcDbSymbolTable
 70
2
  0
LTYPE
  5
21
330
B
100
AcDbSymbolTableRecord
100
AcDbLinetypeTableRecord
  2
Continuous
 70
0
  3
Solid line
 72
65
 73
0
 40
0
  0
LTYPE
  5
22
330
B
100
AcDbSymbolTableRecord
100
AcDbLinetypeTableRecord
  2
HIDDEN
 70
0
  3
Hidden __ __ . __ __ .
 72
65
 73
2
 40
0.035
49
0.025
74
0
49
-0.01
74
0
  0
ENDTAB
  0
ENDSEC
`
	tags, err := lldxf.LoadTags(strings.NewReader(dxf))
	if err != nil {
		t.Fatalf("tagger: %v", err)
	}
	ts := ParseTables(tags)
	if len(ts) != 2 {
		t.Fatalf("tables = %d, want 2", len(ts))
	}
	var layers, ltypes *Table
	for _, tbl := range ts {
		switch tbl.Name() {
		case "LAYER":
			layers = tbl
		case "LTYPE":
			ltypes = tbl
		}
	}
	if layers == nil || ltypes == nil {
		t.Fatalf("missing tables: %v", ts)
	}
	if layers.Len() != 3 || ltypes.Len() != 2 {
		t.Fatalf("entry counts: layers=%d ltypes=%d", layers.Len(), ltypes.Len())
	}
	defp, err := layers.Lookup("Defpoints")
	if err != nil {
		t.Fatal(err)
	}
	lp, ok := defp.(*Layer)
	if !ok {
		t.Fatalf("Defpoints not typed: %T", defp)
	}
	if lp.Plottable() {
		t.Errorf("Defpoints should be non-plottable")
	}
	if lp.Transparency() != 50 {
		t.Errorf("transparency = %d, want 50", lp.Transparency())
	}
	walls, _ := layers.Lookup("Walls")
	wl := walls.(*Layer)
	if wl.Flags()&4 == 0 {
		t.Errorf("Walls should be locked (70=4), flags=%d", wl.Flags())
	}
	hid, _ := ltypes.Lookup("HIDDEN")
	hlt := hid.(*LineType)
	if len(hlt.lengths) != 2 {
		t.Fatalf("HIDDEN pattern lengths = %v, want 2 dashes", hlt.lengths)
	}
	if got := hlt.TotalLength(); got > 0.036 || got < 0.034 {
		t.Errorf("TotalLength = %v, want ~0.035", got)
	}
}
