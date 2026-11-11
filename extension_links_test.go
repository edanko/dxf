package dxf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/object"
)

func TestEntityExtensionDictionaryAndReactors(t *testing.T) {
	body := "0\nLINE\n5\nL1\n" +
		"102\n{ACAD_REACTORS\n330\nR1\n102\n}\n" +
		"102\n{ACAD_XDICTIONARY\n360\nX1\n102\n}\n" +
		"330\nO1\n100\nAcDbEntity\n8\n0\n100\nAcDbLine\n" +
		"10\n0.0\n20\n0.0\n30\n0.0\n11\n1.0\n21\n1.0\n31\n1.0\n"
	e, out := loadOne(t, body)

	reactorReader, ok := e.(entity.ReactorReader)
	if !ok {
		t.Fatalf("type %T has no reactor reader", e)
	}
	if got := reactorReader.ReactorHandles(); len(got) != 1 || got[0] != "R1" {
		t.Errorf("reactors = %v", got)
	}
	xdictReader, ok := e.(entity.ExtDictReader)
	if !ok {
		t.Fatalf("type %T has no extension dictionary reader", e)
	}
	if got := xdictReader.ExtDictHandle(); got != "X1" {
		t.Errorf("xdict = %q", got)
	}
	owner, ok := e.(entity.OwnerGetter)
	if !ok || owner.Owner() == nil || owner.Owner().Handle() != "O1" {
		t.Errorf("owner not preserved")
	}
	for _, want := range []string{
		"102\n{ACAD_REACTORS", "330\nR1", "102\n}",
		"102\n{ACAD_XDICTIONARY", "360\nX1", "330\nO1",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if got := e.(entity.ReactorReader).ReactorHandles(); len(got) != 1 || got[0] != "R1" {
		t.Errorf("reloaded reactors = %v", got)
	}
}

func TestCollectionExtensionDictionary(t *testing.T) {
	src := "0\nSECTION\n2\nOBJECTS\n" +
		"0\nMLEADERSTYLECOLLECTION\n5\nC1\n" +
		"102\n{ACAD_XDICTIONARY\n360\nX1\n102\n}\n" +
		"330\n0\n100\nAcDbDictionary\n281\n1\n3\nStandard\n350\nM1\n" +
		"0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	o, ok := d.GetObject("C1")
	if !ok {
		t.Fatal("collection object not loaded")
	}
	c, ok := o.(*object.MLeaderStyleCollection)
	if !ok {
		t.Fatalf("typed object = %T", o)
	}
	if got := c.ExtDictHandle(); got != "X1" {
		t.Errorf("xdict = %q", got)
	}
	if got := c.MemberHandles(); len(got) != 1 || got[0] != "M1" {
		t.Errorf("members = %v", got)
	}

	var out bytes.Buffer
	if _, err := d.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"0\nMLEADERSTYLECOLLECTION", "360\nX1", "3\nStandard", "350\nM1"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in\n%s", want, out.String())
		}
	}
}

func TestObjectExtensionDictionaryAndReactors(t *testing.T) {
	src := "0\nSECTION\n2\nOBJECTS\n" +
		"0\nLAYOUT\n5\nA1\n" +
		"102\n{ACAD_REACTORS\n330\nR1\n102\n}\n" +
		"102\n{ACAD_XDICTIONARY\n360\nX1\n102\n}\n" +
		"330\n0\n100\nAcDbLayout\n1\nLayout1\n70\n1\n" +
		"0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	o, ok := d.GetObject("A1")
	if !ok {
		t.Fatal("LAYOUT object not loaded")
	}
	reactorReader, ok := o.(object.ReactorReader)
	if !ok {
		t.Fatalf("type %T has no reactor reader", o)
	}
	if got := reactorReader.ReactorHandles(); len(got) != 1 || got[0] != "R1" {
		t.Errorf("reactors = %v", got)
	}
	xdictReader, ok := o.(object.ExtDictReader)
	if !ok {
		t.Fatalf("type %T has no extension dictionary reader", o)
	}
	if got := xdictReader.ExtDictHandle(); got != "X1" {
		t.Errorf("xdict = %q", got)
	}
	if _, ok := o.(*object.Layout); !ok {
		t.Errorf("typed object = %T", o)
	}

	var out bytes.Buffer
	if _, err := d.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"0\nLAYOUT", "102\n{ACAD_REACTORS", "330\nR1", "102\n{ACAD_XDICTIONARY", "360\nX1"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in\n%s", want, out.String())
		}
	}
}
