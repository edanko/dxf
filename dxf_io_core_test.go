package dxf

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/object"
)

type documentInventory struct {
	Entities []string
	Tables   []string
	Blocks   []string
	Objects  []string
}

func inventoryOf(d *drawing.Drawing) documentInventory {
	inv := documentInventory{}
	for _, e := range d.Entities() {
		inv.Entities = append(inv.Entities, e.DXFType()+"@"+e.Handle())
	}
	for _, t := range d.Tables() {
		inv.Tables = append(inv.Tables, t.Name())
		for _, entry := range t.Entries() {
			inv.Tables = append(inv.Tables, t.Name()+":"+entry.Name())
		}
	}
	for _, b := range d.Blocks() {
		inv.Blocks = append(inv.Blocks, b.Name)
	}
	for _, o := range d.Objects() {
		inv.Objects = append(inv.Objects, object.TypeName(o)+"@"+o.Handle())
	}
	sort.Strings(inv.Entities)
	sort.Strings(inv.Tables)
	sort.Strings(inv.Blocks)
	sort.Strings(inv.Objects)
	return inv
}

func roundTrip(t *testing.T, data []byte) (*drawing.Drawing, *drawing.Drawing) {
	t.Helper()
	first, err := FromStringData(string(data))
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	var out bytes.Buffer
	if _, err := first.WriteTo(&out); err != nil {
		t.Fatalf("save: %v", err)
	}
	second, err := FromStringData(out.String())
	if err != nil {
		t.Fatalf("second load: %v\n%s", err, out.String())
	}
	return first, second
}

func equalStrings(t *testing.T, label string, a, b []string) {
	t.Helper()
	if strings.Join(a, "\n") != strings.Join(b, "\n") {
		t.Fatalf("%s differs after round trip:\nfirst (%d):\n%s\nsecond (%d):\n%s",
			label, len(a), strings.Join(a, "\n"), len(b), strings.Join(b, "\n"))
	}
}

func TestLoadBindsAllDocumentSections(t *testing.T) {
	data, err := os.ReadFile("testdata/columns_R2018.dxf")
	if err != nil {
		t.Fatal(err)
	}
	d, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Header().Version; got != "AC1032" {
		t.Errorf("header version = %q, want AC1032", got)
	}
	if len(d.Layers) < 1 {
		t.Errorf("layers = %d, want at least 1", len(d.Layers))
	}
	if len(d.Blocks()) <= 3 {
		t.Errorf("blocks = %d, want loaded block definitions", len(d.Blocks()))
	}
	if len(d.Objects()) <= 3 {
		t.Errorf("objects = %d, want loaded objects", len(d.Objects()))
	}
	if len(d.Tables()) == 0 {
		t.Fatal("no tables bound")
	}
	if len(d.Entities()) == 0 {
		t.Fatal("no entities bound")
	}
}

func TestDocumentRoundTripInventory(t *testing.T) {
	data, err := os.ReadFile("testdata/columns_R2018.dxf")
	if err != nil {
		t.Fatal(err)
	}
	first, second := roundTrip(t, data)
	a, b := inventoryOf(first), inventoryOf(second)
	equalStrings(t, "entities", a.Entities, b.Entities)
	equalStrings(t, "tables", a.Tables, b.Tables)
	equalStrings(t, "blocks", a.Blocks, b.Blocks)
	equalStrings(t, "objects", a.Objects, b.Objects)
}

func TestUnknownEntityRoundTrip(t *testing.T) {
	src := `0
SECTION
2
ENTITIES
0
FUTURE_ENTITY
5
A1
100
AcDbEntity
8
0
123
456
0
ENDSEC
0
EOF
`
	first, second := roundTrip(t, []byte(src))
	if len(first.Entities()) != 1 {
		t.Fatalf("entities = %d, want 1", len(first.Entities()))
	}
	if got := first.Entities()[0].DXFType(); got != "FUTURE_ENTITY" {
		t.Fatalf("type = %q", got)
	}
	if len(second.Entities()) != 1 || second.Entities()[0].DXFType() != "FUTURE_ENTITY" {
		t.Fatalf("unknown entity did not survive: %+v", second.Entities())
	}
	if second.Entities()[0].Handle() != "A1" {
		t.Errorf("handle = %q, want A1", second.Entities()[0].Handle())
	}
}

func TestUnknownGroupCodeOnKnownEntityRoundTrip(t *testing.T) {
	src := `0
SECTION
2
ENTITIES
0
LINE
5
A2
100
AcDbEntity
8
0
100
AcDbLine
10
0
20
0
30
0
11
1
21
1
31
0
1000
hello
0
ENDSEC
0
EOF
`
	first, second := roundTrip(t, []byte(src))
	if len(first.Entities()) != 1 || first.Entities()[0].DXFType() != "LINE" {
		t.Fatalf("expected one LINE, got %+v", first.Entities())
	}
	var out bytes.Buffer
	if _, err := second.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1000\nhello\n") {
		t.Fatalf("unknown group code lost:\n%s", out.String())
	}
}

func TestCommentsArePreservedUnlessStripped(t *testing.T) {
	src := `0
SECTION
2
ENTITIES
999
keep me
0
POINT
5
A3
10
1
20
2
30
3
0
ENDSEC
0
EOF
`
	first, second := roundTrip(t, []byte(src))
	if len(first.Comments()["ENTITIES"]) != 1 {
		t.Fatalf("comments = %+v", first.Comments())
	}
	var out bytes.Buffer
	if _, err := second.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "keep me") {
		t.Fatalf("comment lost:\n%s", out.String())
	}

	stripped, err := FromReaderWithOptions(strings.NewReader(src), &LoadOptions{StripComments: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(stripped.Comments()["ENTITIES"]) != 0 {
		t.Fatalf("comments not stripped: %+v", stripped.Comments())
	}
}

func TestOpaqueSectionsAreRetained(t *testing.T) {
	src := `0
SECTION
2
THUMBNAILIMAGE
90
1
310
DEADBEEF
0
ENDSEC
0
SECTION
2
ENTITIES
0
ENDSEC
0
EOF
`
	first, second := roundTrip(t, []byte(src))
	if len(first.OpaqueSections()["THUMBNAILIMAGE"]) == 0 {
		t.Fatal("thumbnail not retained")
	}
	var out bytes.Buffer
	if _, err := second.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "THUMBNAILIMAGE") || !strings.Contains(out.String(), "DEADBEEF") {
		t.Fatalf("thumbnail lost:\n%s", out.String())
	}
}

func TestRecoveryLoadsTruncatedDocument(t *testing.T) {
	src := `0
SECTION
2
ENTITIES
0
POINT
5
A4
10
1
20
2
30
3
0
LINE
5
A5
10
0
`
	d, err := FromReaderWithOptions(strings.NewReader(src), &LoadOptions{Recover: true})
	if err != nil {
		t.Fatalf("recovery load failed: %v", err)
	}
	if d.RecoveryReport() == nil || len(d.RecoveryReport().Issues) == 0 {
		t.Fatal("expected a non-empty recovery report")
	}
	if len(d.Entities()) == 0 {
		t.Fatal("recovery dropped all complete entities")
	}
}

func TestRecoveryMissingEndsec(t *testing.T) {
	src := `0
SECTION
2
ENTITIES
0
POINT
5
A6
10
1
20
2
30
3
`
	d, err := FromReaderWithOptions(strings.NewReader(src), &LoadOptions{Recover: true})
	if err != nil {
		t.Fatalf("recovery load failed: %v", err)
	}
	if len(d.Entities()) != 1 {
		t.Fatalf("entities = %d, want 1", len(d.Entities()))
	}
	if d.RecoveryReport() == nil || len(d.RecoveryReport().Issues) == 0 {
		t.Fatal("missing ENDSEC was not reported")
	}
}

func TestWriterSectionOrderAndUniqueHandles(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Point(1, 2, 3); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if _, err := d.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out.String(), "\n")
	order := []string{"HEADER", "CLASSES", "TABLES", "BLOCKS", "ENTITIES", "OBJECTS"}
	last := -1
	section := ""
	handles := make(map[string]bool)
	for i := 0; i+1 < len(lines); i++ {
		code := strings.TrimSpace(lines[i])
		value := strings.TrimSpace(lines[i+1])
		if code == "2" && i > 0 && strings.TrimSpace(lines[i-1]) == "SECTION" {
			section = value
			if idx := indexOf(order, value); idx >= 0 {
				if idx < last {
					t.Fatalf("section %s out of order", value)
				}
				last = idx
			}
			continue
		}
		if (code == "5" || code == "105") && section != "" && section != "HEADER" {
			if handles[value] {
				t.Fatalf("duplicate handle %s", value)
			}
			handles[value] = true
		}
	}
	if last != len(order)-1 {
		t.Fatalf("sections not all present, last index %d", last)
	}
}

func indexOf(values []string, value string) int {
	for i, v := range values {
		if v == value {
			return i
		}
	}
	return -1
}

func TestLoadOptionsPlumbing(t *testing.T) {
	src := "0\nSECTION\n2\nENTITIES\n0\nPOINT\n10\n1\n20\n2\n30\n3\n0\nENDSEC\n0\nEOF\n"
	opts := &LoadOptions{Recover: true, StripComments: true, PreserveUnknown: true}
	d, err := FromReaderWithOptions(strings.NewReader(src), opts)
	if err != nil {
		t.Fatal(err)
	}
	if d.RecoveryReport() == nil {
		t.Fatal("options were not applied")
	}
}

func TestHeaderVariablesRoundTrip(t *testing.T) {
	src := `0
SECTION
2
HEADER
9
$ACADVER
1
AC1027
9
$CUSTOMVAR
70
42
0
ENDSEC
0
SECTION
2
ENTITIES
0
ENDSEC
0
EOF
`
	first, second := roundTrip(t, []byte(src))
	if first.Header().Version != "AC1027" {
		t.Fatalf("version = %q", first.Header().Version)
	}
	if v, ok := first.Header().Variable("$CUSTOMVAR"); !ok || len(v.Tags) == 0 {
		t.Fatalf("custom header variable lost: %+v", first.Header().Extra)
	}
	var out bytes.Buffer
	if _, err := second.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "$CUSTOMVAR") {
		t.Fatalf("custom header variable not written:\n%s", out.String())
	}
}

func TestClassesRoundTrip(t *testing.T) {
	src := `0
SECTION
2
CLASSES
0
CLASS
1
MY_CLASS
2
MyCppClass
3
MYAPP
90
1
91
0
280
0
281
0
0
ENDSEC
0
SECTION
2
ENTITIES
0
ENDSEC
0
EOF
`
	first, second := roundTrip(t, []byte(src))
	if len(first.Classes()) != 1 || first.Classes()[0].DXFName != "MY_CLASS" {
		t.Fatalf("classes = %+v", first.Classes())
	}
	if len(second.Classes()) != 1 || second.Classes()[0].CppName != "MyCppClass" {
		t.Fatalf("classes after round trip = %+v", second.Classes())
	}
}

func TestObjectsRetrievableByHandle(t *testing.T) {
	src := `0
SECTION
2
OBJECTS
0
DICTIONARY
5
D1
100
AcDbDictionary
280
1
3
MYKEY
350
E1
0
XRECORD
5
X1
100
AcDbXRecord
280
1
1
payload
0
GROUP
5
G1
100
AcDbGroup
300
my group
71
1
340
E1
0
ACDBPLACEHOLDER
5
P1
0
ENDSEC
0
SECTION
2
ENTITIES
0
ENDSEC
0
EOF
`
	first, second := roundTrip(t, []byte(src))
	for _, d := range []*drawing.Drawing{first, second} {
		dict, ok := d.GetObject("D1")
		if !ok {
			t.Fatalf("dictionary not retrievable")
		}
		if _, ok := dict.(*object.Dictionary); !ok {
			t.Fatalf("dictionary type = %T", dict)
		}
		if _, ok := d.GetObject("X1"); !ok {
			t.Fatal("xrecord not retrievable")
		}
		if _, ok := d.GetObject("G1"); !ok {
			t.Fatal("group not retrievable")
		}
		if _, ok := d.GetObject("P1"); !ok {
			t.Fatal("placeholder not retrievable")
		}
	}
}
