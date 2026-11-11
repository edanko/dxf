package table

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// complexLTypeDXF carries one dash element, one shape element, one gap, and
// one text element in that order.
const complexLTypeDXF = `  0
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
SHAPETEXT
 70
0
  3
Shape and text pattern
 72
65
 73
4
 40
1.5
 49
0.5
 49
-0.25
 74
2
 75
1
 46
1.0
 50
90
 44
0.0
 45
0.0
 49
-0.25
 49
0.25
 74
0
 75
-2
 46
1.0
  9
ABC
 50
15
 44
0.1
 45
-0.1
  0
ENDTAB
  0
ENDSEC
`

func loadComplexLType(t *testing.T) *LineType {
	t.Helper()
	tags, err := lldxf.LoadTags(strings.NewReader(complexLTypeDXF))
	if err != nil {
		t.Fatal(err)
	}
	ts := ParseTables(tags)
	lt, err := ts[0].Lookup("SHAPETEXT")
	if err != nil {
		t.Fatal(err)
	}
	return lt.(*LineType)
}

func TestComplexLinetypeLoadPreservesElements(t *testing.T) {
	lt := loadComplexLType(t)
	els := lt.Elements()
	if len(els) != 4 {
		t.Fatalf("got %d elements, want 4: %#v", len(els), els)
	}
	if els[0].DashLength != 0.5 || els[0].ShapeNum != -1 {
		t.Errorf("dash element = %#v, want 0.5/-1", els[0])
	}
	if els[1].DashLength != -0.25 || els[1].SegLen != 2 || els[1].ShapeNum != 1 ||
		els[1].XScale != 1 || els[1].Rotation != 90 {
		t.Errorf("shape element = %#v", els[1])
	}
	if els[2].DashLength != -0.25 || els[2].ShapeNum != -1 {
		t.Errorf("gap element = %#v, want -0.25/-1", els[2])
	}
	if els[3].DashLength != 0.25 || els[3].ShapeNum != -2 || els[3].TextValue != "ABC" ||
		els[3].Rotation != 15 || els[3].XOffset != 0.1 || els[3].YOffset != -0.1 {
		t.Errorf("text element = %#v", els[3])
	}
}

func TestComplexLinetypeRoundTrip(t *testing.T) {
	lt := loadComplexLType(t)
	f := format.NewASCII()
	lt.Format(f)
	out := f.Output()
	for _, want := range []string{"\n74\n2\n", "\n75\n1\n", "\n9\nABC\n", "\n50\n15.000000\n", "\n44\n0.100000\n", "\n45\n-0.100000\n", "\n49\n0.500000\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("formatted linetype missing %q:\n%s", want, out)
		}
	}

	// Re-load the formatted record: every element and parameter must survive.
	wrapped := "  0\nSECTION\n  2\nTABLES\n  0\nTABLE\n  2\nLTYPE\n  5\nB\n 70\n1\n" + out + "  0\nENDTAB\n  0\nENDSEC\n"
	tags, err := lldxf.LoadTags(strings.NewReader(wrapped))
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := ParseTables(tags)[0].Lookup("SHAPETEXT")
	if err != nil {
		t.Fatal(err)
	}
	lt2 := reloaded.(*LineType)
	got, want := lt2.Elements(), lt.Elements()
	if len(got) != len(want) {
		t.Fatalf("reloaded %d elements, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("element %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestSimpleLinetypeStillDashes(t *testing.T) {
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
22
330
B
100
AcDbSymbolTableRecord
100
AcDbLinetypeTableRecord
  2
DASHED
 70
0
  3
Dashed
 72
65
 73
2
 40
0.75
 49
0.5
 49
-0.25
  0
ENDTAB
  0
ENDSEC
`
	tags, err := lldxf.LoadTags(strings.NewReader(dxf))
	if err != nil {
		t.Fatal(err)
	}
	entry, err := ParseTables(tags)[0].Lookup("DASHED")
	if err != nil {
		t.Fatal(err)
	}
	lt := entry.(*LineType)
	if len(lt.Elements()) != 0 {
		t.Errorf("simple linetype gained elements: %#v", lt.Elements())
	}
}
