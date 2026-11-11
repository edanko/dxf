package object

import (
	"fmt"
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// parseOne parses a single OBJECTS-section record from a DXF tag stream.
func parseOne(t *testing.T, dxf string) Object {
	t.Helper()
	tags, err := lldxf.LoadTags(strings.NewReader(dxf))
	if err != nil {
		t.Fatalf("load tags: %v", err)
	}
	if len(tags) == 0 {
		t.Fatal("no tags parsed")
	}
	return ParseObject(tags)
}

// formatObject renders an object to ASCII.
func formatObject(t *testing.T, o Object) string {
	t.Helper()
	f := format.NewASCII()
	o.Format(f)
	return f.Output()
}

// reparse renders then parses an object again.
func reparse(t *testing.T, o Object) Object {
	t.Helper()
	out := formatObject(t, o)
	tags, err := lldxf.LoadTags(strings.NewReader(out))
	if err != nil {
		t.Fatalf("reload tags: %v", err)
	}
	return ParseObject(tags)
}

func TestStandardObjectTypedModels(t *testing.T) {
	cases := []struct {
		typeName string
		body     string
		wantType string
	}{
		{"LAYOUT", "  1\nLayout1\n100\nAcDbLayout\n 70\n1\n330\nA0\n", "*object.Layout"},
		{"PLOTSETTINGS", "100\nAcDbPlotSettings\n  1\nSetup1\n", "*object.PlotSettings"},
		{"MLINESTYLE", "100\nAcDbMlineStyle\n  2\nStandard\n  3\nDesc\n 70\n1\n", "*object.MLineStyle"},
		{"MLEADERSTYLE", "100\nAcDbMLeaderStyle\n  3\nStandard\n179\n2\n", "*object.MLeaderStyle"},
		{"MATERIAL", "100\nAcDbMaterial\n  1\nMat\n  2\nDesc\n", "*object.Material"},
		{"VISUALSTYLE", "100\nAcDbVisualStyle\n  2\nConceptual\n", "*object.VisualStyle"},
		{"TABLESTYLE", "100\nAcDbTableStyle\n  3\nStandard\n", "*object.TableStyle"},
		{"SORTENTSTABLE", "100\nAcDbSortentsTable\n330\nBLK\n", "*object.SortEntsTable"},
		{"IDBUFFER", "100\nAcDbIdBuffer\n330\nE1\n", "*object.IDBuffer"},
		{"FIELD", "100\nAcDbField\n", "*object.Field"},
		{"FIELDLIST", "100\nAcDbIdSet\n330\nE1\n100\nAcDbFieldList\n", "*object.FieldList"},
		{"GEODATA", "100\nAcDbGeoData\n", "*object.GeoData"},
		{"SUN", "100\nAcDbSun\n", "*object.Sun"},
		{"SPATIAL_FILTER", "100\nAcDbFilter\n", "*object.SpatialFilter"},
		{"RASTERVARIABLES", "100\nAcDbRasterVariables\n", "*object.RasterVariables"},
		{"WIPEOUTVARIABLES", "100\nAcDbWipeoutVariables\n", "*object.WipeoutVariables"},
		{"IMAGEDEF", "100\nAcDbRasterImageDef\n  1\nimg.png\n 10\n640\n 20\n480\n", "*object.ImageDef"},
		{"IMAGEDEF_REACTOR", "100\nAcDbRasterImageDefReactor\n330\nIMG\n", "*object.ImageDefReactor"},
		{"UNDERLAYDEFINITION", "100\nAcDbUnderlayDefinition\n  1\nplan.pdf\n  2\npage1\n", "*object.UnderlayDefinition"},
		{"PDFDEFINITION", "100\nAcDbUnderlayDefinition\n  1\nplan.pdf\n  2\npage1\n", "*object.PdfDefinition"},
	}

	for _, tc := range cases {
		t.Run(tc.typeName, func(t *testing.T) {
			dxf := fmt.Sprintf("  0\n%s\n  5\nAB\n330\nOWNER\n%s", tc.typeName, tc.body)
			o := parseOne(t, dxf)
			if got := fmt.Sprintf("%T", o); got != tc.wantType {
				t.Fatalf("concrete type = %s, want %s", got, tc.wantType)
			}
			if got := TypeName(o); got != tc.typeName {
				t.Fatalf("TypeName = %q, want %q", got, tc.typeName)
			}
			if o.Handle() != "AB" {
				t.Fatalf("handle = %q, want AB", o.Handle())
			}
			so, ok := o.(interface{ Owner() string })
			if !ok {
				t.Fatalf("%s does not expose Owner()", tc.wantType)
			}
			if so.Owner() != "OWNER" {
				t.Fatalf("owner = %q, want OWNER", so.Owner())
			}

			out1 := formatObject(t, o)
			out2 := formatObject(t, reparse(t, o))
			if out1 != out2 {
				t.Fatalf("round trip unstable:\n--- first ---\n%s--- second ---\n%s", out1, out2)
			}
			if !strings.Contains(out1, tc.typeName) || !strings.Contains(out1, "AB") {
				t.Fatalf("formatted record lost identity:\n%s", out1)
			}
		})
	}
}

func TestLayoutNameAndBlockRecord(t *testing.T) {
	dxf := "  0\nLAYOUT\n  5\n1A\n330\nDICT\n100\nAcDbPlotSettings\n  1\nPageSetup\n100\nAcDbLayout\n  1\nMyLayout\n 70\n1\n 71\n0\n330\nBLK0\n331\nVP0\n"
	o := parseOne(t, dxf)
	lay, ok := o.(*Layout)
	if !ok {
		t.Fatalf("got %T, want *Layout", o)
	}
	if lay.Name() != "MyLayout" {
		t.Errorf("name = %q, want MyLayout", lay.Name())
	}
	if lay.BlockRecordHandle() != "BLK0" {
		t.Errorf("block record = %q, want BLK0", lay.BlockRecordHandle())
	}
	if lay.Owner() != "DICT" {
		t.Errorf("owner = %q, want DICT", lay.Owner())
	}
}

func TestImageDefPathAndSize(t *testing.T) {
	dxf := "  0\nIMAGEDEF\n  5\n2A\n330\nDICT\n100\nAcDbRasterImageDef\n 90\n0\n  1\n/tmp/a.png\n 10\n640\n 20\n480\n 11\n0.1\n 21\n0.2\n280\n1\n"
	o := parseOne(t, dxf)
	img, ok := o.(*ImageDef)
	if !ok {
		t.Fatalf("got %T, want *ImageDef", o)
	}
	if img.FilePath() != "/tmp/a.png" {
		t.Errorf("path = %q", img.FilePath())
	}
	w, h := img.ImageSize()
	if w != 640 || h != 480 {
		t.Errorf("image size = %v x %v, want 640 x 480", w, h)
	}
	px, py := img.PixelSize()
	if px != 0.1 || py != 0.2 {
		t.Errorf("pixel size = %v x %v, want 0.1 x 0.2", px, py)
	}
}

func TestSortEntsTableOrder(t *testing.T) {
	dxf := "  0\nSORTENTSTABLE\n  5\n3A\n330\nDICT\n100\nAcDbSortentsTable\n330\nBLK\n331\nE1\n  5\n10\n331\nE2\n  5\n20\n"
	o := parseOne(t, dxf)
	set, ok := o.(*SortEntsTable)
	if !ok {
		t.Fatalf("got %T, want *SortEntsTable", o)
	}
	if set.BlockRecordHandle() != "BLK" {
		t.Errorf("block record = %q, want BLK", set.BlockRecordHandle())
	}
	want := map[string]string{"E1": "10", "E2": "20"}
	got := set.SortOrder()
	for k, v := range want {
		if got[k] != v {
			t.Errorf("sort[%s] = %q, want %q", k, got[k], v)
		}
	}
	// Inner code-5 sort handles must survive the round trip.
	out := formatObject(t, o)
	if !strings.Contains(out, "331\nE1\n5\n10") {
		t.Errorf("sort pair E1/10 lost:\n%s", out)
	}
}

func TestIDBufferEntityHandles(t *testing.T) {
	dxf := "  0\nIDBUFFER\n  5\n4A\n330\nDICT\n100\nAcDbIdBuffer\n330\nE1\n330\nE2\n"
	o := parseOne(t, dxf)
	buf, ok := o.(*IDBuffer)
	if !ok {
		t.Fatalf("got %T, want *IDBuffer", o)
	}
	hs := buf.EntityHandles()
	if len(hs) != 2 || hs[0] != "E1" || hs[1] != "E2" {
		t.Errorf("entity handles = %v, want [E1 E2]", hs)
	}
}

func TestObjectReactorsAndExtDict(t *testing.T) {
	dxf := "  0\nMATERIAL\n  5\n5A\n330\nDICT\n102\n{ACAD_REACTORS\n330\nR1\n330\nR2\n102\n}\n100\nAcDbMaterial\n  1\nMat\n  2\nDesc\n350\nXD0\n"
	o := parseOne(t, dxf)
	mat, ok := o.(*Material)
	if !ok {
		t.Fatalf("got %T, want *Material", o)
	}
	if mat.Name() != "Mat" || mat.Description() != "Desc" {
		t.Errorf("name/description = %q/%q", mat.Name(), mat.Description())
	}
	rs := mat.ReactorHandles()
	if len(rs) != 2 || rs[0] != "R1" || rs[1] != "R2" {
		t.Errorf("reactors = %v, want [R1 R2]", rs)
	}
	if mat.ExtDictHandle() != "XD0" {
		t.Errorf("ext dict = %q, want XD0", mat.ExtDictHandle())
	}
	out := formatObject(t, mat)
	for _, want := range []string{"{ACAD_REACTORS", "330\nR1", "330\nR2", "350\nXD0"} {
		if !strings.Contains(out, want) {
			t.Errorf("formatted record lost %q:\n%s", want, out)
		}
	}
}
