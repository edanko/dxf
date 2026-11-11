package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func TestAnalyzeJSON(t *testing.T) {
	bin := buildCLI(t)
	out, _, code := runCLI(t, bin, "analyze", "--format", "json", "../../testdata/Minimal_DXF_AC1009.dxf")
	if code != 0 {
		t.Fatalf("exit %d out=%s", code, out)
	}
	var report struct {
		Format  string `json:"format"`
		Version string `json:"version"`
		Tags    struct {
			Total    int            `json:"total"`
			Sections map[string]int `json:"sections"`
		} `json:"tags"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if report.Format == "" || report.Version == "" {
		t.Fatalf("missing format/version: %s", out)
	}
	if report.Tags.Total == 0 || len(report.Tags.Sections) == 0 {
		t.Fatalf("missing tag statistics: %s", out)
	}
}

func TestDiffTagsReportsEntityLayerChange(t *testing.T) {
	first := writeTestLineDocument(t, "0")
	second := writeTestLineDocument(t, "HIDDEN")

	bin := buildCLI(t)
	out, _, code := runCLI(t, bin, "difftags", "--format", "json", first, second)
	if code != 1 {
		t.Fatalf("expected exit 1 for differences, got %d out=%s", code, out)
	}
	var report struct {
		Changed   []tagDifference `json:"changed"`
		Different bool            `json:"different"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if !report.Different {
		t.Fatal("report should mark files as different")
	}
	found := false
	for _, diff := range report.Changed {
		if strings.Contains(diff.Path, "ENTITIES/LINE") && strings.HasSuffix(diff.Path, "/8") {
			if diff.Old != "0" || diff.New != "HIDDEN" {
				t.Fatalf("unexpected layer values: %+v", diff)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("entity layer change missing: %s", out)
	}
}

func TestDebugReportsRecovery(t *testing.T) {
	bin := buildCLI(t)

	out, _, code := runCLI(t, bin, "debug", "--format", "json", "../../testdata/Minimal_DXF_AC1009.dxf")
	if code != 0 {
		t.Fatalf("clean debug exit %d out=%s", code, out)
	}
	var clean debugReport
	if err := json.Unmarshal([]byte(out), &clean); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if clean.Version == "" || clean.Format == "" || clean.Encoding == "" || len(clean.Sections) == 0 {
		t.Fatalf("missing diagnostics: %s", out)
	}

	data, err := os.ReadFile("../../testdata/acad_table_simple.dxf")
	if err != nil {
		t.Fatal(err)
	}
	truncated := filepath.Join(t.TempDir(), "truncated.dxf")
	if err := os.WriteFile(truncated, data[:len(data)/2], 0644); err != nil {
		t.Fatal(err)
	}
	out, _, code = runCLI(t, bin, "debug", "--format", "json", truncated)
	if code != 0 {
		t.Fatalf("truncated debug should exit 0, got %d out=%s", code, out)
	}
	var damaged debugReport
	if err := json.Unmarshal([]byte(out), &damaged); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(damaged.RecoveryIssues) == 0 {
		t.Fatalf("expected recovery issue for truncated file: %s", out)
	}
}

func writeTestLineDocument(t *testing.T, layer string) string {
	t.Helper()
	doc, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 0}
	doc.SetEntityLayer(line, layer)
	if err := doc.AddEntity(line); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "line-"+layer+".dxf")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
