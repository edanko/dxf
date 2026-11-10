package dxf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/edanko/dxf/comments"
	"github.com/edanko/dxf/drawing"
)

func writeString(t *testing.T, d *drawing.Drawing) string {
	t.Helper()
	var out bytes.Buffer
	if _, err := d.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// TestUnknownEntityCodeRetained verifies that an unmodelled group code on a
// typed entity survives load, save, and reload.
func TestUnknownEntityCodeRetained(t *testing.T) {
	src := "0\nSECTION\n2\nENTITIES\n0\nLINE\n5\nE1\n100\nAcDbEntity\n8\n0\n100\nAcDbLine\n10\n0\n20\n0\n30\n0\n11\n1\n21\n1\n31\n0\n500\nMYSTERY\n0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	out := writeString(t, d)
	if !strings.Contains(out, "500\nMYSTERY\n") {
		t.Fatalf("unknown entity code dropped:\n%s", out)
	}
	again, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	if out2 := writeString(t, again); !strings.Contains(out2, "500\nMYSTERY\n") {
		t.Fatalf("unknown entity code lost on reload:\n%s", out2)
	}
}

// TestUnknownTableAndObjectCodesRetained verifies retention for a typed table
// entry, a block header, and a typed object.
func TestUnknownTableAndObjectCodesRetained(t *testing.T) {
	src := "0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nLAYER\n70\n1\n0\nLAYER\n5\nL1\n100\nAcDbSymbolTableRecord\n100\nAcDbLayerTableRecord\n2\nCustom\n70\n0\n62\n7\n347\nABC\n0\nENDTAB\n0\nENDSEC\n" +
		"0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n5\nB1\n8\n0\n2\nBlock1\n70\n0\n10\n0\n20\n0\n30\n0\n350\nBLOCKTAG\n0\nENDBLK\n5\nB2\n8\n0\n0\nENDSEC\n" +
		"0\nSECTION\n2\nOBJECTS\n0\nDICTIONARY\n5\nD1\n100\nAcDbDictionary\n281\n1\n3\nKEY\n350\nVALUE\n347\nKEEP\n0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	out := writeString(t, d)
	for _, want := range []string{"347\nABC\n", "350\nBLOCKTAG\n", "347\nKEEP\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("retained tag %q missing:\n%s", want, out)
		}
	}
}

// TestCommentsKeepInterEntityPosition verifies that a comment between two
// entities is written between them after a round trip and that stripping
// removes it from the writer as well.
func TestCommentsKeepInterEntityPosition(t *testing.T) {
	src := "0\nSECTION\n2\nENTITIES\n" +
		"0\nPOINT\n5\nA1\n10\n1\n20\n2\n30\n3\n" +
		"999\nbetween points\n" +
		"0\nPOINT\n5\nA2\n10\n4\n20\n5\n30\n6\n" +
		"0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := d.PositionedComments(); len(got) != 1 || got[0].Index != 1 || got[0].Text != "between points" {
		t.Fatalf("positioned comments = %+v", got)
	}
	out := writeString(t, d)
	first := strings.Index(out, "5\nA1\n")
	comment := strings.Index(out, "between points")
	second := strings.Index(out, "5\nA2\n")
	if first < 0 || comment < 0 || second < 0 || !(first < comment && comment < second) {
		t.Fatalf("comment moved (first=%d comment=%d second=%d):\n%s", first, comment, second, out)
	}

	comments.StripComments(d)
	out = writeString(t, d)
	if strings.Contains(out, "between points") {
		t.Fatalf("comment survived strip:\n%s", out)
	}
}
