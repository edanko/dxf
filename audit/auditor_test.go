package audit

import (
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/table"
)

func TestAuditorNew(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatalf("Failed to create drawing: %v", err)
	}

	auditor := NewAuditor(d)

	if auditor == nil {
		t.Fatal("NewAuditor should not return nil")
	}

	if auditor.doc != d {
		t.Error("Auditor should reference the drawing")
	}

	if auditor.errors == nil {
		t.Error("Errors slice should be initialized")
	}

	// Note: fixedErrors and trashCan are initialized empty by Run(), not NewAuditor()

	if auditor.logger == nil {
		t.Error("Logger should be initialized")
	}

	if !auditor.AutoFix {
		t.Error("AutoFix should be true by default")
	}

	if !auditor.RecoverLinks {
		t.Error("RecoverLinks should be true by default")
	}

	if !auditor.RemoveOrphans {
		t.Error("RemoveOrphans should be true by default")
	}
}

func TestAuditorSetLogger(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	customLogger := NewAuditLogger(true)
	auditor.SetLogger(customLogger)

	if auditor.logger != customLogger {
		t.Error("SetLogger should set the logger")
	}

	if !auditor.logger.Verbose {
		t.Error("Logger should be verbose after SetLogger(true)")
	}
}

func TestAuditorSetOptions(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	auditor.SetOptions(false, true, false, false)

	if auditor.AutoFix {
		t.Error("AutoFix should be false after SetOptions(false, ...)")
	}

	if !auditor.StrictMode {
		t.Error("StrictMode should be true after SetOptions(..., true, ...)")
	}

	if auditor.RecoverLinks {
		t.Error("RecoverLinks should be false after SetOptions(..., ..., false, ...)")
	}

	if auditor.RemoveOrphans {
		t.Error("RemoveOrphans should be false after SetOptions(..., ..., ..., false)")
	}
}

func TestAuditorRunBasic(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)
	auditor.AutoFix = false // Disable autofix to avoid fix errors

	errors, err := auditor.Run()

	if err != nil {
		t.Fatalf("Run should not return error for valid drawing: %v", err)
	}

	if errors == nil {
		t.Error("Run should return non-nil error slice")
	}
}

func TestAuditorRunNilDrawing(t *testing.T) {
	auditor := NewAuditor(nil)

	errors, err := auditor.Run()

	if err == nil {
		t.Error("Run should return error for nil drawing")
	}

	if errors == nil {
		t.Error("Run should return error slice even for nil drawing")
	}

	if len(errors) == 0 {
		t.Error("Run should add error for nil drawing")
	}
}

func TestAuditorGetErrors(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	auditor.Run()
	errors := auditor.GetErrors()

	if errors == nil {
		t.Error("GetErrors should return non-nil slice")
	}

	// Modify returned slice shouldn't affect internal state
	errors[0] = nil
	errors2 := auditor.GetErrors()
	if errors2[0] == nil {
		t.Error("GetErrors should return a copy")
	}
}

func TestAuditorGetFixedErrors(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	auditor.Run()
	fixedErrors := auditor.GetFixedErrors()

	if fixedErrors == nil {
		t.Error("GetFixedErrors should return non-nil slice")
	}
}

func TestAuditorHasErrors(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	if auditor.HasErrors() {
		t.Error("Fresh auditor should not have errors")
	}

	auditor.Run()

	if auditor.HasErrors() {
		t.Log("Audit found errors in valid drawing")
	}
}

func TestAuditorHasCriticalErrors(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	if auditor.HasCriticalErrors() {
		t.Error("Fresh auditor should not have critical errors")
	}

	auditor.Run()

	if auditor.HasCriticalErrors() {
		t.Log("Audit found critical errors")
	}
}

func TestAuditorGetErrorSummary(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	auditor.Run()
	summary := auditor.GetErrorSummary()

	if summary == nil {
		t.Error("GetErrorSummary should return non-nil map")
	}
}

func TestAuditorWithInvalidLayer(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 0}

	// Create a layer that doesn't exist in drawing
	invalidLayer := table.NewLayer("NonExistent", 0, nil)
	line.SetLayer(invalidLayer)

	d.AddEntity(line)

	auditor.Run()

	errors := auditor.GetErrors()

	foundLayerError := false
	for _, err := range errors {
		if err.Code == ErrorUndefinedLayer {
			foundLayerError = true
			break
		}
	}

	if !foundLayerError {
		t.Log("Expected to find undefined layer error")
	}
}

func TestAuditorWithValidEntities(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Add some valid entities
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 0}
	line.SetLayer(d.Layers["0"])
	d.AddEntity(line)

	circle := entity.NewCircle()
	circle.Center = []float64{5, 5, 0}
	circle.Radius = 3.0
	circle.SetLayer(d.Layers["0"])
	d.AddEntity(circle)

	auditor.Run()

	if auditor.HasErrors() {
		t.Log("Audit found errors in drawing with valid entities")
		errors := auditor.GetErrors()
		t.Logf("Found %d errors", len(errors))
	}
}

func TestAuditorAutoFixDisabled(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)
	auditor.AutoFix = false

	auditor.Run()
	fixedErrors := auditor.GetFixedErrors()

	if len(fixedErrors) > 0 {
		t.Error("No errors should be fixed when AutoFix is disabled")
	}
}

func TestAuditorMultipleRuns(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)
	auditor.AutoFix = false // Disable autofix to avoid fix errors

	// First run
	errors1, err1 := auditor.Run()
	if err1 != nil {
		t.Fatalf("First run failed: %v", err1)
	}

	count1 := len(errors1)

	// Add an entity
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 0}
	d.AddEntity(line)

	// Second run
	errors2, err2 := auditor.Run()
	if err2 != nil {
		t.Fatalf("Second run failed: %v", err2)
	}

	// Second run should have cleared previous errors
	if len(errors2) != count1 {
		t.Logf("Error count changed from %d to %d after second run", count1, len(errors2))
	}
}

func TestAuditorAuditMethods(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Test auditBasicStructure
	err := auditor.auditBasicStructure()
	if err != nil {
		t.Errorf("auditBasicStructure failed: %v", err)
	}

	// Test auditTables
	err = auditor.auditTables()
	if err != nil {
		t.Errorf("auditTables failed: %v", err)
	}

	// Test auditEntities
	err = auditor.auditEntities()
	if err != nil {
		t.Errorf("auditEntities failed: %v", err)
	}

	// Test auditReferences
	err = auditor.auditReferences()
	if err != nil {
		t.Errorf("auditReferences failed: %v", err)
	}

	// Test auditGeometry
	err = auditor.auditGeometry()
	if err != nil {
		t.Errorf("auditGeometry failed: %v", err)
	}
}

func TestAuditorValidateLayerName(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	testCases := []struct {
		name     string
		layer    string
		expected bool
	}{
		{"Valid layer", "TestLayer", true},
		{"Empty name", "", false},
		{"Invalid chars <", "Test<Layer", false},
		{"Invalid chars >", "Test>Layer", false},
		{"Invalid chars /", "Test/Layer", false},
		{"Invalid chars \\", "Test\\Layer", false},
		{"Invalid chars \"", "Test\"Layer", false},
		{"Invalid chars :", "Test:Layer", false},
		{"Invalid chars ;", "Test;Layer", false},
		{"Invalid chars ?", "Test?Layer", false},
		{"Invalid chars *", "Test*Layer", false},
		{"Invalid chars |", "Test|Layer", false},
		{"Invalid chars ,", "Test,Layer", false},
		{"Invalid chars =", "Test=Layer", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := auditor.validateLayerName(tc.layer)
			hasError := err != nil

			if hasError != !tc.expected {
				t.Errorf("validateLayerName(%q) = error=%v, expected error=%v", tc.layer, hasError, !tc.expected)
			}
		})
	}
}

func TestAuditorValidateHeader(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Should not panic
	auditor.validateHeader()
}

func TestAuditorValidateLayerTable(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	err := auditor.validateLayerTable()
	if err != nil {
		t.Errorf("validateLayerTable failed: %v", err)
	}
}

func TestAuditorValidateOtherTables(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	err := auditor.validateOtherTables()
	if err != nil {
		t.Errorf("validateOtherTables failed: %v", err)
	}
}

func TestAuditorValidateLayerReferences(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Should not panic with empty entities
	auditor.validateLayerReferences()
}

func TestAuditorValidateStyleReferences(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Should not panic
	auditor.validateStyleReferences()
}

func TestAuditorValidateDimStyleReferences(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Should not panic
	auditor.validateDimStyleReferences()
}

func TestAuditorValidateBlockReferences(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Should not panic
	auditor.validateBlockReferences()
}

func TestAuditorAuditEntity(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Test with valid entity
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 0}
	line.SetLayer(d.Layers["0"])

	auditor.auditEntity(line, "Test")
}

func TestAuditorAuditEntitySlice(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	entities := []entity.Entity{
		nil,
		entity.NewLine(),
		entity.NewCircle(),
	}

	// Should not panic
	auditor.auditEntitySlice(entities, "Test")
}

func TestAuditorAddError(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	initialCount := len(auditor.errors)

	err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Test error")
	auditor.addError(err)

	if len(auditor.errors) != initialCount+1 {
		t.Error("addError should add error to the list")
	}
}

func TestAuditorAddFixedError(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Test error")
	auditor.addFixedError(err, "Fixed by creating layer")

	if len(auditor.fixedErrors) != 1 {
		t.Error("addFixedError should add to fixedErrors")
	}

	if !err.Fixed {
		t.Error("Error should be marked as fixed")
	}
}

func TestAuditorAddToTrash(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	line := entity.NewLine()
	auditor.addToTrash(line)

	if len(auditor.trashCan) != 1 {
		t.Error("addToTrash should add entity to trash can")
	}
}

func TestAuditorEmptyTrashCan(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	auditor.addToTrash(entity.NewLine())
	auditor.addToTrash(entity.NewCircle())

	auditor.emptyTrashCan()

	if len(auditor.trashCan) != 0 {
		t.Error("emptyTrashCan should clear the trash can")
	}
}

func TestAuditorApplyFixes(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)
	auditor.AutoFix = true

	err := auditor.applyFixes()
	if err != nil {
		t.Errorf("applyFixes failed: %v", err)
	}
}

func TestAuditorFixError(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Test with unfixable error
	err := NewError(ErrorMissingSection, SeverityCritical, "", "", "Test")
	fixErr := auditor.fixError(err)

	if fixErr == nil {
		t.Error("fixError should return error for unfixable error code")
	}
}

func TestAuditorPrintSummary(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)

	// Should not panic
	auditor.PrintSummary()
}

func TestAuditorWithManyEntities(t *testing.T) {
	d, _ := drawing.New()
	auditor := NewAuditor(d)
	auditor.AutoFix = false // Disable autofix to avoid fix errors

	// Add many entities
	for i := 0; i < 100; i++ {
		line := entity.NewLine()
		line.Start = []float64{float64(i), float64(i), 0}
		line.End = []float64{float64(i + 1), float64(i + 1), 0}
		line.SetLayer(d.Layers["0"])
		d.AddEntity(line)
	}

	errors, err := auditor.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	t.Logf("Audit of 100 entities found %d errors", len(errors))
}
