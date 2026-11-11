package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
)

// TestEntityConformanceMatrix fails when a registered entity type has no
// conformance row or no golden fixture, then verifies that every row loads as
// its typed model and survives a round trip.
func TestEntityConformanceMatrix(t *testing.T) {
	if missing := entity.MissingEntityConformance(); len(missing) > 0 {
		t.Fatalf("registered types without conformance row or fixture: %v", missing)
	}
	for _, name := range entity.RegisteredTypes() {
		row, ok := entity.ConformanceFor(name)
		if !ok {
			t.Errorf("%s: no conformance row", name)
			continue
		}
		src := "0\nSECTION\n2\nENTITIES\n" + conformanceFixture(name) + "0\nENDSEC\n0\nEOF\n"
		d, err := FromStringData(src)
		if err != nil {
			t.Errorf("%s (%s): parse: %v", name, row.Reference, err)
			continue
		}
		es := d.Entities()
		if len(es) != 1 {
			t.Errorf("%s: loaded %d entities", name, len(es))
			continue
		}
		if _, retained := es[0].(*entity.RetainedRecordType); retained {
			t.Errorf("%s: fell back to whole-record retention", name)
			continue
		}
		out := writeString(t, d)
		again, err := FromStringData(out)
		if err != nil {
			t.Errorf("%s: reload: %v", name, err)
			continue
		}
		if len(again.Entities()) != 1 {
			t.Errorf("%s: reloaded %d entities", name, len(again.Entities()))
		}
	}
}

// conformanceFixture returns a deterministic golden record for a type. Key
// types use hand-written records exercising their documented codes; the rest
// use the minimal valid record the registry accepts.
func conformanceFixture(name string) string {
	switch name {
	case "LINE":
		return "0\nLINE\n5\nC1\n100\nAcDbEntity\n8\n0\n100\nAcDbLine\n10\n1\n20\n2\n30\n3\n11\n4\n21\n5\n31\n6\n"
	case "SPLINE":
		return "0\nSPLINE\n5\nC1\n100\nAcDbEntity\n8\n0\n100\nAcDbSpline\n70\n3\n71\n3\n72\n8\n73\n4\n74\n0\n" +
			"40\n0\n40\n0\n40\n0\n40\n1\n40\n1\n40\n1\n40\n1\n" +
			"10\n0\n20\n0\n30\n0\n10\n1\n21\n1\n30\n1\n10\n2\n21\n2\n30\n2\n10\n3\n21\n3\n30\n3\n"
	case "HATCH":
		return "0\nHATCH\n5\nC1\n100\nAcDbEntity\n8\n0\n100\nAcDbHatch\n10\n0\n20\n0\n30\n0\n" +
			"210\n0\n220\n0\n230\n1\n2\nSOLID\n70\n1\n71\n0\n91\n1\n92\n1\n" +
			"93\n4\n10\n0\n20\n0\n10\n10\n20\n0\n10\n10\n20\n10\n10\n0\n20\n10\n" +
			"97\n0\n75\n1\n76\n1\n98\n0\n"
	case "ACAD_TABLE":
		return "0\nACAD_TABLE\n5\nC1\n100\nAcDbEntity\n8\n0\n100\nAcDbBlockReference\n2\n*T1\n10\n0\n20\n0\n30\n0\n" +
			"100\nAcDbTable\n90\n4\n91\n1\n92\n1\n140\n10\n141\n8\n"
	}
	return "0\n" + name + "\n5\nC1\n100\nAcDbEntity\n8\n0\n"
}

// TestVersionDowngradeDropsGatedField verifies that a field unavailable in the
// save target is omitted and the conversion is reported.
func TestVersionDowngradeDropsGatedField(t *testing.T) {
	src := "0\nSECTION\n2\nENTITIES\n0\nLINE\n5\nE1\n100\nAcDbEntity\n8\n0\n420\n16711680\n10\n0\n20\n0\n30\n0\n11\n1\n21\n1\n31\n0\n0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	d.SetTargetVersion(format.AC1009)
	out := writeString(t, d)
	if strings.Contains(out, "\n420\n") {
		t.Fatalf("true color survived R12 downgrade:\n%s", out)
	}
	report := d.Conversion()
	if report == nil || !containsNote(report.Dropped, "true color") {
		t.Fatalf("conversion report = %+v", report)
	}
}

func containsNote(notes []string, want string) bool {
	for _, note := range notes {
		if strings.Contains(note, want) {
			return true
		}
	}
	return false
}
