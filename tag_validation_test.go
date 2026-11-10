package dxf_test

import (
	"strings"
	"testing"

	dxf "github.com/edanko/dxf"
	"github.com/edanko/dxf/audit"
)

// TestTagValidationReportsCodeValueHandle checks pre-bind validation reports
// the offending group code, value, and record handle for malformed records.
func TestTagValidationReportsCodeValueHandle(t *testing.T) {
	src := "0\nSECTION\n2\nENTITIES\n" +
		"0\nPOINT\n5\nA\n100\nAcDbEntity\n8\n0\n62\n999\n10\n0\n20\n0\n30\n0\n" +
		"0\nPOINT\n5\nA\n100\nAcDbEntity\n8\n0\n10\n1\n20\n1\n30\n0\n" +
		"0\nENDSEC\n0\nEOF\n"
	d, err := dxf.FromReaderWithOptions(strings.NewReader(src), &dxf.LoadOptions{Recover: true})
	if err != nil {
		t.Fatal(err)
	}
	rep := d.RecoveryReport()
	if rep == nil {
		t.Fatal("no recovery report")
	}
	joined := strings.Join(rep.Issues, "\n")
	for _, want := range []string{
		`code 62 value "999" handle "A"`,
		`code 5 value "A" handle "A"`,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("report missing %q:\n%s", want, joined)
		}
	}
}

// TestRepairInsertsMissingBlockTerminator verifies the repair pass inserts the
// missing ENDBLK, records the action, and re-validates clean.
func TestRepairInsertsMissingBlockTerminator(t *testing.T) {
	src := "0\nSECTION\n2\nBLOCKS\n" +
		"0\nBLOCK\n5\nB1\n330\nE\n100\nAcDbBlockBegin\n70\n0\n8\n0\n2\nB1\n10\n0\n20\n0\n30\n0\n" +
		"0\nENDSEC\n" +
		"0\nSECTION\n2\nENTITIES\n0\nENDSEC\n0\nEOF\n"
	reader := strings.NewReader(src)
	d, err := dxf.FromReaderWithOptions(reader, &dxf.LoadOptions{Recover: true})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(d.RecoveryReport().Issues, "\n")
	if !strings.Contains(joined, "block missing ENDBLK") {
		t.Fatalf("repair action not recorded:\n%s", joined)
	}
	found := false
	for _, b := range d.Blocks() {
		if b.Name == "B1" {
			found = true
		}
	}
	if !found {
		t.Fatal("repaired block B1 missing")
	}
	if issues := audit.Run(d, nil); issues.HasErrors() {
		t.Fatalf("repaired file does not audit clean: %s", issues)
	}
}
