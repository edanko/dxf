package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditFileComprehensiveLoadsRealFile(t *testing.T) {
	config := DefaultAuditConfig()
	config.GenerateReport = false
	ca := NewComprehensiveAuditor(config)

	const path = "../testdata/columns_R2018.dxf"
	doc, err := ca.loadDrawing(path)
	if err != nil {
		t.Fatalf("loadDrawing: %v", err)
	}
	if len(doc.Entities()) == 0 {
		t.Fatal("clean file loaded no entities")
	}

	result, err := ca.AuditFile(path)
	if err != nil {
		t.Fatalf("AuditFile: %v", err)
	}
	if !result.Success {
		t.Fatal("clean file audit did not succeed")
	}
}

func TestAuditFileComprehensiveMissingFileErrors(t *testing.T) {
	config := DefaultAuditConfig()
	config.GenerateReport = false
	ca := NewComprehensiveAuditor(config)

	const path = "../testdata/definitely-missing-file.dxf"
	_, err := ca.AuditFile(path)
	if err == nil {
		t.Fatal("AuditFile accepted a missing file")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("error does not identify the missing path: %v", err)
	}
}

func TestAuditFileComprehensiveMalformedFileReportsLoadError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "malformed.dxf")
	if err := os.WriteFile(path, []byte("this is not a DXF document\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	config := DefaultAuditConfig()
	config.GenerateReport = false
	ca := NewComprehensiveAuditor(config)

	if _, err := ca.AuditFile(path); err == nil {
		t.Fatal("AuditFile accepted a malformed file")
	}
}

func TestAuditFileComprehensiveDefaultConfig(t *testing.T) {
	const path = "../testdata/columns_R2018.dxf"
	result, err := AuditFileComprehensive(path)
	if err != nil {
		t.Fatalf("AuditFileComprehensive: %v", err)
	}
	if !result.Success {
		t.Fatal("clean file audit did not succeed")
	}
	if result.ReportFile != "" {
		t.Cleanup(func() { _ = os.Remove(result.ReportFile) })
	}
}
