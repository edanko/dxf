package standards

import (
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func TestLayerNameRuleReportsTableViolation(t *testing.T) {
	doc := newTestDrawing(t)
	layer, err := doc.AddLayer("BAD*NAME", 7, nil)
	if err != nil {
		t.Fatalf("AddLayer: %v", err)
	}
	layer.SetHandleValue("10")

	report := DefaultRuleSet().Run(doc)
	if !report.HasViolations() {
		t.Fatal("expected a violation")
	}
	violation := report.Violations[0]
	if violation.Code != LayerNameRuleCode {
		t.Fatalf("code = %q", violation.Code)
	}
	if violation.Severity != SeverityError {
		t.Fatalf("severity = %q", violation.Severity)
	}
	if violation.Path != "TABLES/LAYER/BAD*NAME" {
		t.Fatalf("path = %q", violation.Path)
	}
	if violation.Handle != "10" {
		t.Fatalf("handle = %q", violation.Handle)
	}
}

func TestLayerNameRuleReportsEntityViolation(t *testing.T) {
	doc := newTestDrawing(t)
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 0}
	if err := doc.AddEntity(line); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}
	doc.SetEntityLayer(line, "BAD*NAME")

	report := DefaultRuleSet().Run(doc)
	if !report.HasViolations() {
		t.Fatal("expected a violation")
	}
	violation := report.Violations[0]
	wantPath := "ENTITIES/LINE[" + line.Handle() + "]/8"
	if violation.Path != wantPath {
		t.Fatalf("path = %q, want %q", violation.Path, wantPath)
	}
	if violation.Handle != line.Handle() {
		t.Fatalf("handle = %q, want %q", violation.Handle, line.Handle())
	}
}

func TestRuleSetSeverityOverrideAndDisable(t *testing.T) {
	doc := newTestDrawing(t)
	if _, err := doc.AddLayer("BAD*NAME", 7, nil); err != nil {
		t.Fatalf("AddLayer: %v", err)
	}

	rules := NewRuleSet(NewLayerNameRule("^[A-Z0-9]+$", SeverityError))
	if err := rules.SetSeverity(LayerNameRuleCode, SeverityWarning); err != nil {
		t.Fatalf("SetSeverity: %v", err)
	}
	report := rules.Run(doc)
	if len(report.Violations) != 1 {
		t.Fatalf("violations = %d", len(report.Violations))
	}
	if report.Violations[0].Severity != SeverityWarning {
		t.Fatalf("severity = %q", report.Violations[0].Severity)
	}
	if report.Count(SeverityWarning) != 1 {
		t.Fatalf("warning count = %d", report.Count(SeverityWarning))
	}

	if err := rules.Disable(LayerNameRuleCode); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if report := rules.Run(doc); report.HasViolations() {
		t.Fatal("disabled rule still reported a violation")
	}
	if err := rules.Enable(LayerNameRuleCode); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if report := rules.Run(doc); !report.HasViolations() {
		t.Fatal("re-enabled rule did not report a violation")
	}
}

func TestRuleSetRejectsUnknownRuleAndSeverity(t *testing.T) {
	rules := DefaultRuleSet()
	if err := rules.SetSeverity("UNKNOWN", SeverityError); err == nil {
		t.Fatal("expected unknown rule error")
	}
	if err := rules.SetSeverity(LayerNameRuleCode, Severity("bogus")); err == nil {
		t.Fatal("expected invalid severity error")
	}
	if err := rules.Disable("UNKNOWN"); err == nil {
		t.Fatal("expected unknown rule error")
	}
}

func TestDefaultRuleSetAcceptsValidLayer(t *testing.T) {
	doc := newTestDrawing(t)
	if report := DefaultRuleSet().Run(doc); report.HasViolations() {
		t.Fatalf("unexpected violations: %+v", report.Violations)
	}
}

func newTestDrawing(t *testing.T) *drawing.Drawing {
	t.Helper()
	doc, err := drawing.New()
	if err != nil {
		t.Fatalf("drawing.New: %v", err)
	}
	return doc
}
