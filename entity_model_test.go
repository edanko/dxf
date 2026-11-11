package dxf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

// Registry integration: root loadEntity dispatches to registered loaders and
// retained records; raw-tag retention keeps unhandled codes.
func TestRegistryDispatchIntegration(t *testing.T) {
	for _, typ := range []string{"LINE", "CIRCLE", "ARC", "POINT", "ELLIPSE", "SPLINE", "TEXT", "MTEXT"} {
		if !entity.IsRegistered(typ) {
			t.Errorf("type %s not registered by root init", typ)
		}
	}
	src := "0\nSECTION\n2\nENTITIES\n0\nLINE\n5\nA2\n100\nAcDbEntity\n8\n0\n10\n1.0\n20\n2.0\n30\n0.0\n11\n4.0\n21\n6.0\n31\n0.0\n999\nkeepme\n0\nENDSEC\n0\nEOF\n"
	first, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	l, ok := first.Entities()[0].(*entity.Line)
	if !ok {
		t.Fatalf("type %T", first.Entities()[0])
	}
	if l.Start[0] != 1.0 || l.End[1] != 6.0 {
		t.Errorf("geometry not loaded: %v %v", l.Start, l.End)
	}
	var out bytes.Buffer
	if _, err := first.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "999\nkeepme") {
		t.Error("raw tag 999 not retained")
	}
}

func TestRetainedRecordIntegration(t *testing.T) {
	src := "0\nSECTION\n2\nENTITIES\n0\nACAD_PROXY_ENTITY\n5\nB1\n100\nAcDbEntity\n8\n0\n310\nDEADBEEF\n0\nENDSEC\n0\nEOF\n"
	first, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := first.Entities()[0].DXFType(); got != "ACAD_PROXY_ENTITY" {
		t.Errorf("type = %q", got)
	}
	var out bytes.Buffer
	first.WriteTo(&out)
	second, err := FromStringData(out.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Entities()) != 1 {
		t.Fatalf("entities = %d", len(second.Entities()))
	}
	e := second.Entities()[0]
	if e.DXFType() != "ACAD_PROXY_ENTITY" || e.Handle() != "B1" {
		t.Errorf("retention lost type/handle: %s %s", e.DXFType(), e.Handle())
	}
	var out2 bytes.Buffer
	second.WriteTo(&out2)
	if !strings.Contains(out2.String(), "310\nDEADBEEF") {
		t.Error("proxy payload tag 310 lost")
	}
}
