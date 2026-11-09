package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func near(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-9
}

func TestShapeRoundTrip(t *testing.T) {
	body := "0\nSHAPE\n5\nSH1\n100\nAcDbEntity\n8\n0\n100\nAcDbShape\n39\n" + f(2.0) + "\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(0.0) + "\n40\n" + f(5.0) + "\n2\nSTANDARDBOX\n41\n" + f(1.5) + "\n50\n" + f(90.0) + "\n51\n" + f(15.0) + "\n"
	e, out := loadOne(t, body)
	sh, ok := e.(*entity.Shape)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	pt := sh.GetInsertPoint()
	if pt.X() != 1.0 || pt.Y() != 2.0 {
		t.Errorf("insert = %v", []float64{pt.X(), pt.Y()})
	}
	if sh.GetShapeName() != "STANDARDBOX" || sh.GetSize() != 5.0 {
		t.Errorf("name=%q size=%v", sh.GetShapeName(), sh.GetSize())
	}
	if sh.GetRotation() != 90.0 || sh.GetXScale() != 1.5 {
		t.Errorf("rot=%v xscale=%v", sh.GetRotation(), sh.GetXScale())
	}
	for _, want := range []string{"SHAPE", "2\nSTANDARDBOX", "41\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestXLineRayRoundTrip(t *testing.T) {
	for _, c := range []struct{ name, sub string }{
		{"XLINE", "AcDbXline"},
		{"RAY", "AcDbRay"},
	} {
		body := "0\n" + c.name + "\n5\nX1\n100\nAcDbEntity\n8\n0\n100\n" + c.sub + "\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(3.0) + "\n11\n" + f(0.5) + "\n21\n" + f(0.5) + "\n31\n" + f(0.7071067811865476) + "\n"
		e, out := loadOne(t, body)
		x, ok := e.(*entity.XLine)
		if !ok {
			ray, isRay := e.(*entity.Ray)
			if isRay && c.name == "RAY" {
				x = ray.XLine
			} else {
				t.Fatalf("%s: type %T", c.name, e)
			}
		}
		st := x.Start()
		if st.X() != 1.0 || st.Y() != 2.0 || st.Z() != 3.0 {
			t.Errorf("%s: start = %v", c.name, []float64{st.X(), st.Y(), st.Z()})
		}
		uv := x.UnitVector()
		if !near(uv.X(), 0.5) || !near(uv.Y(), 0.5) {
			t.Errorf("%s: unit vector = %v", c.name, []float64{uv.X(), uv.Y()})
		}
		if !strings.Contains(out, "0\n"+c.name) {
			t.Errorf("%s: record name not re-emitted in\n%s", c.name, out)
		}
	}
}

func TestWipeoutRoundTrip(t *testing.T) {
	body := "0\nWIPEOUT\n5\nW1\n100\nAcDbEntity\n8\n0\n100\nAcDbWipeout\n90\n" + f(0) + "\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(0.0) + "\n11\n" + f(4.0) + "\n21\n" + f(0.0) + "\n31\n" + f(0.0) + "\n12\n" + f(0.0) + "\n22\n" + f(3.0) + "\n32\n" + f(0.0) + "\n13\n" + f(640.0) + "\n23\n" + f(480.0) + "\n70\n" + f(1) + "\n71\n" + f(0) + "\n"
	e, out := loadOne(t, body)
	w, ok := e.(*entity.Wipeout)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	ip := w.InsertPoint()
	if ip.X() != 1.0 || ip.Y() != 2.0 {
		t.Errorf("insert = %v", []float64{ip.X(), ip.Y()})
	}
	uvv := w.UVector()
	if uvv.X() != 4.0 {
		t.Errorf("u vector = %v", []float64{uvv.X(), uvv.Y()})
	}
	sz := w.ImageSize()
	if sz.X() != 640.0 || sz.Y() != 480.0 {
		t.Errorf("image size = %v", []float64{sz.X(), sz.Y()})
	}
	if w.Flags() != 1 {
		t.Errorf("flags = %d", w.Flags())
	}
	for _, want := range []string{"WIPEOUT", "13\n640"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestRetainedRecordSurvival(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"OLE2FRAME", "0\nOLE2FRAME\n5\nO1\n100\nAcDbEntity\n8\n0\n102\n{ACAD_REACTORS\n330\nABC\n102\n}\n100\nAcDbOLE2Frame\n90\n" + f(4) + "\n310\nDEADBEEF\n" + "71\n" + f(1) + "\n"},
		{"ACAD_PROXY_ENTITY", "0\nACAD_PROXY_ENTITY\n5\nPX1\n100\nAcDbEntity\n8\n0\n100\nAcDbProxyEntity\n90\n" + f(50614) + "\n91\n" + f(3) + "\n92\n" + f(10) + "\n93\n" + f(2) + "\n310\nAABBCCDDEEFF\n340\nXYZ\n71\n" + f(1) + "\n72\n" + f(1) + "\n"},
		{"THIRD_PARTY_APP", "0\nTHIRD_PARTY_APP\n5\nTP1\n100\nAcDbEntity\n8\n0\n90\n" + f(7) + "\n1\ncustom data\n310\nCAFE\n"},
		{"ZZZ_UNKNOWN", "0\nZZZ_UNKNOWN\n5\nUZ1\n100\nAcDbEntity\n8\n0\n70\n" + f(3) + "\n10\n" + f(1.0) + "\n"},
	}
	for _, c := range cases {
		e, out := loadOne(t, c.body)
		r, ok := e.(*entity.RetainedRecordType)
		if !ok {
			t.Fatalf("%s: loaded %T", c.name, e)
		}
		if r.DXFType() != c.name {
			t.Errorf("%s: DXFType = %q", c.name, r.DXFType())
		}
		for _, want := range []string{"0\n" + c.name, "5\nO1", "5\nPX1", "5\nTP1", "5\nUZ1"} {
			if !strings.Contains(c.body, want) {
				continue // not part of this fixture
			}
			if !strings.Contains(out, want) {
				t.Errorf("%s: missing %q in\n%s", c.name, want, out)
			}
		}
		for _, want := range []string{"90", "310"} {
			if strings.Contains(c.body, "\n"+want+"\n") && !strings.Contains(out, "\n"+want+"\n") {
				t.Errorf("%s: code %s not retained in\n%s", c.name, want, out)
			}
		}
	}
}

func TestAttdefAttribRoundedRoundTrip(t *testing.T) {
	// Insertion point at y=2: the loader must not zero x/z on each component.
	body := "0\nATTDEF\n5\nA1\n100\nAcDbEntity\n8\n0\n100\nAcDbAttributeDefinition\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n" + f(3.0) + "\n40\n" + f(2.5) + "\n1\nDEFVAL\n2\nTAG1\n3\nPROMPT1\n7\nSTANDARD\n50\n" + f(45.0) + "\n"
	e, out := loadOne(t, body)
	a, ok := e.(*entity.Attdef)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	ip := a.GetInsertPoint()
	if len(ip) < 3 || ip[0] != 1.0 || ip[1] != 2.0 || ip[2] != 3.0 {
		t.Errorf("insert point = %v", ip)
	}
	for _, want := range []string{"1\nDEFVAL", "2\nTAG1", "3\nPROMPT1"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}

	body2 := "0\nATTRIB\n5\nA2\n100\nAcDbEntity\n8\n0\n100\nAcDbAttribute\n10\n" + f(4.0) + "\n20\n" + f(5.0) + "\n30\n" + f(6.0) + "\n40\n" + f(2.5) + "\n1\nTEXTVAL\n2\nTAG2\n70\n" + f(0) + "\n"
	e2, out2 := loadOne(t, body2)
	at, ok := e2.(*entity.Attrib)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e2, out2)
	}
	ip2 := at.GetInsertPoint()
	if len(ip2) < 3 || ip2[0] != 4.0 || ip2[1] != 5.0 || ip2[2] != 6.0 {
		t.Errorf("attrib insert point = %v", ip2)
	}
}

func TestLightSunRetainedRoundTrip(t *testing.T) {
	cases := []struct{ name, body string }{
		{"LIGHT", "0\nLIGHT\n5\nL1\n100\nAcDbEntity\n8\n0\n100\nAcDbLight\n90\n" + f(50) + "\n1\nMyPointLight\n70\n" + f(2) + "\n40\n" + f(1.5) + "\n"},
		{"SUN", "0\nSUN\n5\nSU1\n100\nAcDbEntity\n8\n0\n100\nAcDbSun\n90\n" + f(1) + "\n63\n" + f(255) + "\n421\n" + f(16777215) + "\n"},
	}
	for _, c := range cases {
		e, out := loadOne(t, c.body)
		r, ok := e.(*entity.RetainedRecordType)
		if !ok {
			t.Fatalf("%s: loaded %T", c.name, e)
		}
		if r.DXFType() != c.name {
			t.Errorf("%s: DXFType = %q", c.name, r.DXFType())
		}
		for _, want := range []string{"0\n" + c.name, "5\nL1", "5\nSU1", "100\nAcDbLight", "100\nAcDbSun"} {
			if strings.Contains(c.body, want) && !strings.Contains(out, want) {
				t.Errorf("%s: missing %q in\n%s", c.name, want, out)
			}
		}
	}
}

func TestFieldRetainedRoundTrip(t *testing.T) {
	body := "0\nFIELD\n5\nF1\n100\nAcDbField\n1\nACAD_EVALUATOR\n2\n\\V3.3.P\n90\n" + f(0) + "\n"
	e, out := loadOne(t, body)
	r, ok := e.(*entity.RetainedRecordType)
	if !ok {
		t.Fatalf("loaded %T", e)
	}
	if r.DXFType() != "FIELD" {
		t.Errorf("DXFType = %q", r.DXFType())
	}
	for _, want := range []string{"0\nFIELD", "1\nACAD_EVALUATOR", "2\n\\V3.3.P"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestAcadTableRetainedRoundTrip(t *testing.T) {
	body := "0\nACAD_TABLE\n5\nTB1\n100\nAcDbEntity\n8\n0\n100\nAcDbBlockReference\n100\nAcDbTable\n70\n" + f(3) + "\n71\n" + f(2) + "\n62\n" + f(7) + "\n"
	e, out := loadOne(t, body)
	r, ok := e.(*entity.RetainedRecordType)
	if !ok {
		t.Fatalf("loaded %T", e)
	}
	if r.DXFType() != "ACAD_TABLE" {
		t.Errorf("DXFType = %q", r.DXFType())
	}
	for _, want := range []string{"0\nACAD_TABLE", "100\nAcDbTable", "70\n3"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}
