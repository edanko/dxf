package entity

import (
	"fmt"
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/xdata"
)

// Compile-time checks: capability interfaces are satisfied by the types that
// already implement them.
var (
	_ Copier  = (*Line)(nil)
	_ Copier  = (*Insert)(nil)
	_ Copier  = (*Polyline)(nil)
	_ Bounder = (*Line)(nil)
	_ Bounder = (*Circle)(nil)
	_ Bounder = (*Arc)(nil)
	_ Bounder = (*LwPolyline)(nil)
)

func tagsOf(pairs ...TagPair) lldxf.Tags {
	tags := make(lldxf.Tags, 0, len(pairs))
	for _, p := range pairs {
		tags = append(tags, lldxf.NewDXFTag(lldxf.GroupCode(p.Code), p.Value))
	}
	return tags
}

func ascii(t *testing.T, e Entity) string {
	t.Helper()
	f := format.NewASCII()
	e.Format(f)
	return f.Output()
}

// registerCoreLoaders registers a minimal LINE loader for registry tests; the
// application root package registers the full production set.
func registerCoreLoaders() {
	Register("LINE", func(tags lldxf.Tags) (Entity, error) {
		e := NewEntity(LINE)
		l := &Line{entity: e, Start: []float64{0, 0, 0}, End: []float64{0, 0, 0}}
		for _, tag := range tags {
			v := fmt.Sprintf("%v", tag.Value())
			switch int(tag.Code()) {
			case 5:
				e.SetHandleString(v)
			case 10:
				l.Start[0] = parseFloat(v)
			case 20:
				l.Start[1] = parseFloat(v)
			case 30:
				l.Start[2] = parseFloat(v)
			case 11:
				l.End[0] = parseFloat(v)
			case 21:
				l.End[1] = parseFloat(v)
			case 31:
				l.End[2] = parseFloat(v)
			}
		}
		return l, nil
	})
	SetDefaultLoader(NewRetainedLoader)
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%g", &f)
	return f
}

func TestRegistryLoadLineRoundTrip(t *testing.T) {
	registerCoreLoaders()
	src := []TagPair{
		{0, "LINE"}, {5, "A2"}, {100, "AcDbEntity"}, {8, "0"},
		{10, "1.0"}, {20, "2.0"}, {30, "0.0"},
		{11, "4.0"}, {21, "6.0"}, {31, "0.0"},
		{999, "keepme"},
	}
	e, err := LoadEntity("line", tagsOf(src...))
	if err != nil {
		t.Fatal(err)
	}
	// Raw-tag retention (binder behavior) is exercised by root-package tests;
	// the registry loader returns the typed entity with its typed fields.
	line, ok := e.(*Line)
	if !ok {
		t.Fatalf("type = %T, want *Line", e)
	}
	if line.Handle() != "A2" {
		t.Errorf("handle = %q", line.Handle())
	}
	out := ascii(t, e)
	for _, want := range []string{"LINE", "10\n1.0", "11\n4.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestRegistryRetainedRecord(t *testing.T) {
	src := []TagPair{
		{0, "FUTURE_ENTITY"}, {5, "A1"}, {100, "AcDbEntity"}, {8, "0"}, {123, "456"},
	}
	SetDefaultLoader(NewRetainedLoader)
	registerCoreLoaders()
	e, err := LoadEntity("FUTURE_ENTITY", tagsOf(src...))
	if err != nil {
		t.Fatal(err)
	}
	if e.DXFType() != "FUTURE_ENTITY" {
		t.Errorf("DXFType = %q", e.DXFType())
	}
	if e.Handle() != "A1" {
		t.Errorf("handle = %q, want A1", e.Handle())
	}
	out := ascii(t, e)
	for _, want := range []string{"0\nFUTURE_ENTITY", "5\nA1", "123\n456"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	rec, ok := e.(*RetainedRecordType)
	if !ok {
		t.Fatalf("type = %T, want *RetainedRecordType", e)
	}
	if len(rec.rawTags) == 0 {
		t.Error("retained record lost tags")
	}
}

func TestBaseCommonAttributes(t *testing.T) {
	l := NewLine()
	l.SetLinetype("DASHED")
	l.SetColorRGB(0x123456)
	l.SetLineWeight(25)
	l.SetInvisible(true)
	l.SetTransparency(0.5)
	l.SetExtrusion([]float64{0, 0, -1})
	out := ascii(t, l)
	for _, want := range []string{"6\nDASHED", "420\n1193046", "370\n25", "60\n1", "440\n50", "210\n0.0", "230\n-1.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if l.Linetype() != "DASHED" || l.ColorRGB() != 0x123456 || l.LineWeight() != 25 || !l.Invisible() || l.Transparency() != 0.5 {
		t.Error("accessor round trip failed")
	}
}

func TestBaseXDataFormat(t *testing.T) {
	l := NewLine()
	if err := l.AddXDataTag("ACAD", 1000, "hello"); err != nil {
		t.Fatal(err)
	}
	if err := l.AddXDataTag("ACAD", 1040, 3.5); err != nil {
		t.Fatal(err)
	}
	out := ascii(t, l)
	for _, want := range []string{"1001\nACAD", "1000\nhello", "1040\n3.5"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if !l.HasXData("ACAD") {
		t.Error("HasXData false")
	}
	data, ok := l.GetXData("ACAD")
	if !ok || len(data.Tags) < 2 {
		t.Errorf("GetXData = %+v", data)
	}
	l.ClearXData()
	if l.GetXDataObject().Len() != 0 {
		t.Error("ClearXData failed")
	}
	_ = xdata.NewXData()
}

func TestRegisteredTypesIncludeCore(t *testing.T) {
	registerCoreLoaders()
	for _, name := range []string{"LINE"} {
		if !IsRegistered(name) {
			t.Errorf("type %s not registered", name)
		}
	}
	if IsRegistered("NO_SUCH_TYPE") {
		t.Error("unexpected registration")
	}
}

func TestCopierInterface(t *testing.T) {
	l := NewLine()
	l.SetHandleString("AA")
	c := l.Copy()
	if c.Handle() != "AA" {
		t.Errorf("copy handle = %q", c.Handle())
	}
	cl, ok := c.(*Line)
	if !ok {
		t.Fatalf("copy type %T", c)
	}
	cl.Start[0] = 42
	if l.Start[0] == 42 {
		t.Error("copy shares geometry")
	}
}

func TestBounderInterface(t *testing.T) {
	l := NewLine()
	l.Start = []float64{1, 2, 3}
	l.End = []float64{4, 5, 6}
	mins, maxs := l.BBox()
	if mins[0] != 1 || maxs[2] != 6 {
		t.Errorf("bbox %v %v", mins, maxs)
	}
	_ = math.NewVec3(0, 0, 0)
}
