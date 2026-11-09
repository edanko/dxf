package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestInsertAttribsRoundTrip(t *testing.T) {
	body := "0\nINSERT\n5\nI1\n100\nAcDbEntity\n8\n0\n2\nMYBLOCK\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n0.0\n41\n" + f(2.0) + "\n42\n" + f(3.0) + "\n43\n" + f(1.0) + "\n50\n" + f(90.0) + "\n70\n" + f(0) + "\n66\n" + f(1) + "\n" +
		"0\nATTRIB\n5\nA1\n100\nAcDbText\n8\n0\n10\n" + f(1.0) + "\n20\n" + f(2.0) + "\n30\n0.0\n40\n" + f(2.5) + "\n72\n" + f(0) + "\n74\n" + f(0) + "\n1\nValueOne\n2\nTAG1\n3\nPrompt one\n" +
		"0\nATTRIB\n5\nA2\n100\nAcDbText\n8\n0\n10\n" + f(5.0) + "\n20\n" + f(6.0) + "\n30\n0.0\n40\n" + f(2.5) + "\n72\n" + f(0) + "\n74\n" + f(0) + "\n1\nValueTwo\n2\nTAG2\n" +
		"0\nSEQEND\n5\nSE1\n"
	e, out := loadOne(t, body)
	ins, ok := e.(*entity.Insert)
	if !ok {
		t.Fatalf("type %T", e)
	}
	attrs := ins.GetAllAttributes()
	if len(attrs) != 2 {
		t.Fatalf("attributes = %d (out:\n%s)", len(attrs), out)
	}
	if attrs[0].Tag() != "TAG1" || attrs[0].Text() != "ValueOne" {
		t.Errorf("attrib0 tag=%q text=%q", attrs[0].Tag(), attrs[0].Text())
	}
	if attrs[0].Prompt() != "Prompt one" {
		t.Errorf("prompt = %q", attrs[0].Prompt())
	}
	if attrs[1].EndHandle() != "SE1" {
		t.Errorf("seqend handle = %q", attrs[1].EndHandle())
	}
	for _, want := range []string{"2\nMYBLOCK", "66\n1", "ATTRIB", "2\nTAG1", "1\nValueOne", "SEQEND"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestInsertNoAttribsNoSeqend(t *testing.T) {
	body := "0\nINSERT\n5\nI2\n100\nAcDbEntity\n8\n0\n2\nPLAIN\n10\n0.0\n20\n0.0\n30\n0.0\n"
	_, out := loadOne(t, body)
	if strings.Contains(out, "66") || strings.Contains(out, "SEQEND") {
		t.Errorf("plain insert emitted attribute structure:\n%s", out)
	}
}
