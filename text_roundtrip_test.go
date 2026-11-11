package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestTextRoundTripText(t *testing.T) {
	body := "0\nTEXT\n5\nT1\n100\nAcDbEntity\n8\n0\n10\n1.0\n20\n2.0\n30\n0.0\n40\n" + f(2.5) + "\n1\nHello Wörld greek αβγ\n50\n" + f(45.0) + "\n7\nSTANDARD\n41\n" + f(0.8) + "\n51\n" + f(12.0) + "\n72\n" + f(1) + "\n11\n3.0\n21\n4.0\n73\n" + f(0) + "\n7\nSTANDARD\n"
	e, out := loadOne(t, body)
	tx, ok := e.(*entity.Text)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if tx.Value != "Hello Wörld greek αβγ" {
		t.Errorf("value = %q", tx.Value)
	}
	for _, want := range []string{"40\n2.5", "1\nHello Wörld", "αβγ", "50\n45"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestTextRoundTripMText(t *testing.T) {
	body := "0\nMTEXT\n5\nM1\n100\nAcDbEntity\n8\n0\n10\n1.0\n20\n2.0\n30\n0.0\n40\n" + f(3.0) + "\n41\n" + f(10.0) + "\n71\n" + f(5) + "\n72\n" + f(1) + "\n73\n" + f(3) + "\n1\nFirst line\\Psecond{\\H2.0x}line ünïcode \U0001F600\n44\n" + f(1.5) + "\n"
	e, out := loadOne(t, body)
	mt, ok := e.(*entity.MText)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if !strings.Contains(mt.Value, "First line") || !strings.Contains(mt.Value, "ünïcode") {
		t.Errorf("value = %q", mt.Value)
	}
	for _, want := range []string{"40\n3", "1\nFirst line", "ünïcode"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestTextRoundTripAttribAttdef(t *testing.T) {
	body := "0\nATTDEF\n5\nD1\n100\nAcDbEntity\n8\n0\n10\n1.0\n20\n2.0\n30\n0.0\n40\n" + f(2.0) + "\n1\nDefaultValue\n2\nTAG1\n3\nPrompt me\n70\n" + f(0) + "\n"
	e, out := loadOne(t, body)
	ad, ok := e.(*entity.Attdef)
	if !ok {
		t.Fatalf("type %T", e)
	}
	if ad.GetTag() != "TAG1" {
		t.Errorf("tag = %q", ad.GetTag())
	}
	for _, want := range []string{"2\nTAG1", "3\nPrompt me"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestTextRoundTripTolerance(t *testing.T) {
	body := "0\nTOLERANCE\n5\nTL1\n100\nAcDbEntity\n8\n0\n10\n1.0\n20\n2.0\n30\n0.0\n40\n" + f(0.5) + "\n1\ngt 0.01 gt 0.02\n71\n" + f(1) + "\n72\n" + f(2) + "\n73\n" + f(3) + "\n74\n" + f(4) + "\n75\n" + f(5) + "\n7\nSTANDARD\n"
	e, out := loadOne(t, body)
	if _, ok := e.(*entity.Tolerance); !ok {
		t.Fatalf("type %T", e)
	}
	for _, want := range []string{"gt 0.01"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}
