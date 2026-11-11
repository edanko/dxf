package audit

import (
	"testing"

	"github.com/edanko/dxf/table"
)

// TestAuditError tests audit error functionality
func TestAuditError(t *testing.T) {
	t.Run("ErrorCreation", func(t *testing.T) {
		err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Test error message")

		if err.Code != ErrorUndefinedLayer {
			t.Errorf("Expected error code %d, got %d", ErrorUndefinedLayer, err.Code)
		}

		if err.Severity != SeverityError {
			t.Errorf("Expected severity %d, got %d", SeverityError, err.Severity)
		}

		if err.Entity != "ENTITY" {
			t.Errorf("Expected entity 'ENTITY', got '%s'", err.Entity)
		}

		if err.Handle != "123" {
			t.Errorf("Expected handle '123', got '%s'", err.Handle)
		}

		if err.Message != "Test error message" {
			t.Errorf("Expected message 'Test error message', got '%s'", err.Message)
		}
	})

	t.Run("ErrorFixing", func(t *testing.T) {
		err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Test error")

		if !err.CanFix {
			t.Error("ErrorUndefinedLayer should be fixable")
		}

		err.MarkFixed("Fixed by creating missing layer")

		if !err.Fixed {
			t.Error("Error should be marked as fixed")
		}

		if err.RecoveryLog != "Fixed by creating missing layer" {
			t.Errorf("Expected recovery log 'Fixed by creating missing layer', got '%s'", err.RecoveryLog)
		}
	})

	t.Run("ErrorString", func(t *testing.T) {
		err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Test error")
		err.Context = "Layer context"

		str := err.String()

		// Check that key components are in string
		if !contains(str, "ERROR") {
			t.Error("Error string should contain severity")
		}

		if !contains(str, "Reference to undefined layer") {
			t.Error("Error string should contain error description")
		}

		if !contains(str, "ENTITY") {
			t.Error("Error string should contain entity type")
		}

		if !contains(str, "123") {
			t.Error("Error string should contain handle")
		}
	})
}

// TestErrorCategories tests error categorization
func TestErrorCategories(t *testing.T) {
	testCases := []struct {
		code     ErrorCode
		expected ErrorCategory
	}{
		{ErrorMissingSection, CategoryStructure},
		{ErrorUndefinedLayer, CategoryReference},
		{ErrorInvalidColor, CategoryProperty},
		{ErrorInvalidPoint, CategoryGeometry},
	}

	for _, tc := range testCases {
		t.Run(tc.code.String(), func(t *testing.T) {
			category := tc.code.GetCategory()
			if category != tc.expected {
				t.Errorf("Expected category %s for error %s, got %s",
					tc.expected, tc.code.String(), category)
			}
		})
	}
}

// TestAuditLogger tests audit logger functionality
func TestAuditLogger(t *testing.T) {
	t.Run("LoggerCreation", func(t *testing.T) {
		logger := NewAuditLogger(false)

		if logger.Verbose {
			t.Error("Verbose should be false by default")
		}

		verboseLogger := NewAuditLogger(true)

		if !verboseLogger.Verbose {
			t.Error("Verbose should be true when set")
		}
	})
}

// TestRecoverySystem tests recovery functionality
func TestRecoverySystem(t *testing.T) {
	t.Run("RecoveryOptions", func(t *testing.T) {
		options := DefaultRecoveryOptions()

		if options.Mode != RecoveryConservative {
			t.Errorf("Expected default mode %d, got %d", RecoveryConservative, options.Mode)
		}

		if !options.DetectEncoding {
			t.Error("DetectEncoding should be true by default")
		}

		if !options.FixHandles {
			t.Error("FixHandles should be true by default")
		}
	})

	t.Run("RecovererCreation", func(t *testing.T) {
		recoverer := NewRecoverer(nil)

		if recoverer == nil {
			t.Error("Recoverer should not be nil")
		}

		if recoverer.Options == nil {
			t.Error("Options should not be nil")
		}
	})
}

// TestAuditConfig tests audit configuration
func TestAuditConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultAuditConfig()

		if !config.AutoFix {
			t.Error("AutoFix should be true by default")
		}

		if config.StrictMode {
			t.Error("StrictMode should be false by default")
		}

		if !config.DetectCycles {
			t.Error("DetectCycles should be true by default")
		}

		if config.MaxErrors != 1000 {
			t.Errorf("Expected MaxErrors 1000, got %d", config.MaxErrors)
		}
	})
}

// TestComprehensiveAuditor tests comprehensive auditor functionality
func TestComprehensiveAuditor(t *testing.T) {
	t.Run("AuditorCreation", func(t *testing.T) {
		config := DefaultAuditConfig()
		auditor := NewComprehensiveAuditor(config)

		if auditor == nil {
			t.Error("Auditor should not be nil")
		}

		if auditor.config == nil {
			t.Error("Config should not be nil")
		}

		if auditor.logger == nil {
			t.Error("Logger should not be nil")
		}
	})
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkErrorCreation benchmarks error creation performance
func BenchmarkErrorCreation(b *testing.B) {
	for b.Loop() {
		_ = NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Benchmark error")
	}
}

// BenchmarkErrorString benchmarks error string formatting performance
func BenchmarkErrorString(b *testing.B) {
	err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Benchmark error")


	for b.Loop() {
		_ = err.String()
	}
}

func TestIsValidLayerName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid simple name", "LAYER1", true},
		{"Valid with underscore", "my_layer", true},
		{"Valid with numbers", "layer123", true},
		{"Empty string", "", false},
		{"Too long", string(make([]byte, 256)), false},
		{"Contains <", "layer<test>", false},
		{"Contains >", "layer>test", false},
		{"Contains /", "layer/test", false},
		{"Contains \\", "layer\\test", false},
		{"Contains \"", "layer\"test", false},
		{"Contains :", "layer:test", false},
		{"Contains ;", "layer;test", false},
		{"Contains ?", "layer?test", false},
		{"Contains *", "layer*test", false},
		{"Contains |", "layer|test", false},
		{"Contains ,", "layer,test", false},
		{"Contains =", "layer=test", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidLayerName(tc.input)
			if result != tc.expected {
				t.Errorf("IsValidLayerName(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIsValidStyleName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid style name", "mystyle", true},
		{"Empty string", "", false},
		{"Too long", string(make([]byte, 256)), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidStyleName(tc.input)
			if result != tc.expected {
				t.Errorf("IsValidStyleName(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestValidateEntityHandle(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid hex handle", "1F", true},
		{"Valid longer handle", "ABC123", true},
		{"Valid single digit", "1", true},
		{"Empty string", "", false},
		{"Contains G", "1G", false},
		{"Contains Z", "1Z", false},
		{"Contains lowercase", "1f", false},
		{"Contains special", "1@", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateEntityHandle(tc.input)
			if result != tc.expected {
				t.Errorf("ValidateEntityHandle(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestValidationResult(t *testing.T) {
	t.Run("Initial state", func(t *testing.T) {
		result := &ValidationResult{}
		if result.Valid {
			t.Error("New ValidationResult should be invalid by default")
		}
		if len(result.Errors) != 0 {
			t.Error("New ValidationResult should have no errors")
		}
		if len(result.Warnings) != 0 {
			t.Error("New ValidationResult should have no warnings")
		}
	})

	t.Run("AddError", func(t *testing.T) {
		result := &ValidationResult{Valid: true}
		result.AddError("error1")
		if result.Valid {
			t.Error("Result should be invalid after adding error")
		}
		if len(result.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(result.Errors))
		}
		if result.Errors[0] != "error1" {
			t.Errorf("Expected error 'error1', got '%s'", result.Errors[0])
		}
	})

	t.Run("AddWarning", func(t *testing.T) {
		result := &ValidationResult{Valid: true}
		result.AddWarning("warning1")
		if !result.Valid {
			t.Error("Result should still be valid after adding warning")
		}
		if len(result.Warnings) != 1 {
			t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
		}
	})
}

func TestEntityValidator(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		validator := NewEntityValidator()
		if validator == nil {
			t.Error("NewEntityValidator should not return nil")
		}
		if validator.validLayers == nil {
			t.Error("validLayers map should be initialized")
		}
		if validator.validStyles == nil {
			t.Error("validStyles map should be initialized")
		}
		if validator.validLinetypes == nil {
			t.Error("validLinetypes map should be initialized")
		}
	})

	t.Run("SetValidLayers", func(t *testing.T) {
		validator := NewEntityValidator()
		layers := map[string]*table.Layer{
			"layer1": nil,
			"layer2": nil,
		}
		validator.SetValidLayers(layers)

		if !validator.validLayers["layer1"] {
			t.Error("layer1 should be in validLayers")
		}
		if !validator.validLayers["layer2"] {
			t.Error("layer2 should be in validLayers")
		}
	})

	t.Run("Validate nil entity", func(t *testing.T) {
		validator := NewEntityValidator()
		result := validator.ValidateEntity(nil)
		if result.Valid {
			t.Error("Result should be invalid for nil entity")
		}
		if len(result.Errors) != 1 || result.Errors[0] != "entity is nil" {
			t.Errorf("Expected error 'entity is nil', got '%v'", result.Errors)
		}
	})
}

func TestDrawingStats(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		stats := &DrawingStats{
			TotalEntities:   100,
			InvalidEntities: 5,
			TotalErrors:     10,
			TotalWarnings:   3,
		}
		str := stats.String()
		if !contains(str, "100") {
			t.Error("String should contain total count")
		}
		if !contains(str, "5") {
			t.Error("String should contain invalid count")
		}
		if !contains(str, "10") {
			t.Error("String should contain error count")
		}
		if !contains(str, "3") {
			t.Error("String should contain warning count")
		}
	})
}

func TestAuditErrorSeverity(t *testing.T) {
	testCases := []struct {
		severity ErrorSeverity
		expected string
	}{
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			err := NewError(ErrorUndefinedLayer, tc.severity, "ENTITY", "123", "test")
			if err.Severity != tc.severity {
				t.Errorf("Expected severity %d, got %d", tc.severity, err.Severity)
			}
		})
	}
}

func TestAuditErrorCodes(t *testing.T) {
	t.Run("Error code categories", func(t *testing.T) {
		codes := []ErrorCode{
			ErrorMissingSection,
			ErrorMissingHeader,
			ErrorUndefinedLayer,
			ErrorInvalidColor,
			ErrorInvalidPoint,
			ErrorCircularReference,
		}

		for _, code := range codes {
			t.Run(code.String(), func(t *testing.T) {
				if code.GetCategory() == "" {
					t.Errorf("Code %d should have a category", code)
				}
			})
		}
	})
}

func TestAuditErrorWithLocation(t *testing.T) {
	err := NewErrorWithLocation(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Test error", 100, 50)

	if err.Line != 100 {
		t.Errorf("Expected Line 100, got %d", err.Line)
	}
	if err.Column != 50 {
		t.Errorf("Expected Column 50, got %d", err.Column)
	}
}

func TestAuditErrorWithContext(t *testing.T) {
	err := NewErrorWithContext(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Test error", "Layer context")

	if err.Context != "Layer context" {
		t.Errorf("Expected Context 'Layer context', got '%s'", err.Context)
	}
}

func TestAuditLoggerVerbose(t *testing.T) {
	logger := NewAuditLogger(true)
	if !logger.Verbose {
		t.Error("Logger should be verbose")
	}

	logger.LogMessage("Test message")
}

func TestAuditLoggerLogRecovery(t *testing.T) {
	logger := NewAuditLogger(false)
	logger.LogRecovery("Fixed", "Fixed layer error")
}

func TestRecoveryOptions(t *testing.T) {
	t.Run("Conservative mode", func(t *testing.T) {
		options := &RecoveryOptions{Mode: RecoveryConservative}
		if options.Mode != RecoveryConservative {
			t.Error("Should be conservative mode")
		}
	})

	t.Run("Aggressive mode", func(t *testing.T) {
		options := &RecoveryOptions{Mode: RecoveryAggressive}
		if options.Mode != RecoveryAggressive {
			t.Error("Should be aggressive mode")
		}
	})

	t.Run("Disabled mode", func(t *testing.T) {
		options := &RecoveryOptions{Mode: RecoveryNone}
		if options.Mode != RecoveryNone {
			t.Error("Should be disabled mode")
		}
	})
}

func TestRecovererCreation(t *testing.T) {
	recoverer := NewRecoverer(&RecoveryOptions{
		Mode:           RecoveryConservative,
		DetectEncoding: true,
		FixHandles:     true,
	})

	if recoverer.Options.Mode != RecoveryConservative {
		t.Error("Recoverer should have correct mode")
	}
	if !recoverer.Options.DetectEncoding {
		t.Error("DetectEncoding should be true")
	}
}

func TestAuditConfigExtended(t *testing.T) {
	t.Run("Custom config", func(t *testing.T) {
		config := &AuditConfig{
			AutoFix:      false,
			StrictMode:   true,
			DetectCycles: false,
			MaxErrors:    500,
		}

		if config.AutoFix {
			t.Error("AutoFix should be false")
		}
		if !config.StrictMode {
			t.Error("StrictMode should be true")
		}
		if config.DetectCycles {
			t.Error("DetectCycles should be false")
		}
		if config.MaxErrors != 500 {
			t.Errorf("Expected MaxErrors 500, got %d", config.MaxErrors)
		}
	})
}

func TestValidationResultMultiple(t *testing.T) {
	result := &ValidationResult{Valid: true}

	result.AddError("Error 1")
	result.AddWarning("Warning 1")
	result.AddError("Error 2")

	if result.Valid {
		t.Error("Result should be invalid with errors")
	}
	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestEntityValidatorValidateEntity(t *testing.T) {
	validator := NewEntityValidator()

	result := validator.ValidateEntity(nil)
	if result.Valid {
		t.Error("Result should be invalid for nil entity")
	}
}

func TestIsValidLayerNameEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid normal name", "Layer_123", true},
		{"Valid single char", "A", true},
		{"Valid long name", string(make([]byte, 255)), true},
		{"Invalid empty", "", false},
		{"Invalid too long", string(make([]byte, 256)), false},
		{"Invalid with <", "Layer<test>", false},
		{"Invalid with >", "Layer>test", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidLayerName(tc.input)
			if result != tc.expected {
				t.Errorf("IsValidLayerName(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestValidateEntityHandleExtended(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid single hex", "A", true},
		{"Valid multi hex", "ABCDEF", true},
		{"Valid mixed", "A1B2C3", true},
		{"Empty", "", false},
		{"Invalid G", "G", false},
		{"Invalid Z", "Z", false},
		{"Invalid lowercase", "abc", false},
		{"Invalid special", "@", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateEntityHandle(tc.input)
			if result != tc.expected {
				t.Errorf("ValidateEntityHandle(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestErrorCategoryStrings(t *testing.T) {
	if string(CategoryStructure) != "STRUCTURE" {
		t.Error("CategoryStructure should be STRUCTURE")
	}
	if string(CategoryReference) != "REFERENCE" {
		t.Error("CategoryReference should be REFERENCE")
	}
	if string(CategoryProperty) != "PROPERTY" {
		t.Error("CategoryProperty should be PROPERTY")
	}
	if string(CategoryGeometry) != "GEOMETRY" {
		t.Error("CategoryGeometry should be GEOMETRY")
	}
	if string(CategoryData) != "DATA" {
		t.Error("CategoryData should be DATA")
	}
	if string(CategoryRecovery) != "RECOVERY" {
		t.Error("CategoryRecovery should be RECOVERY")
	}
}

func TestDrawingStatsExtended(t *testing.T) {
	stats := &DrawingStats{
		TotalEntities:   100,
		EntityCounts:    make(map[string]int),
		LayerCounts:     make(map[string]int),
	}

	stats.EntityCounts["POINT"] = 50
	stats.EntityCounts["LINE"] = 50
	stats.LayerCounts["0"] = 100

	if stats.TotalEntities != 100 {
		t.Errorf("Expected TotalEntities 100, got %d", stats.TotalEntities)
	}
	if stats.EntityCounts["POINT"] != 50 {
		t.Errorf("Expected 50 POINT entities, got %d", stats.EntityCounts["POINT"])
	}
	if stats.LayerCounts["0"] != 100 {
		t.Errorf("Expected 100 entities on layer 0, got %d", stats.LayerCounts["0"])
	}
}

func TestRecoveryResultString(t *testing.T) {
	result := &RecoveryResult{
		Success:         true,
		OriginalSize:    1000,
		RecoveredSize:   800,
		ErrorsFixed:     5,
		ErrorsRemaining: 2,
		Warnings:        []string{"Warning 1", "Warning 2"},
		Encoding:        "utf-8",
		Actions:         []string{"Action 1", "Action 2"},
	}

	str := result.String()
	if str == "" {
		t.Error("String should not be empty")
	}
	if !contains(str, "DXF Recovery Result") {
		t.Error("String should contain 'DXF Recovery Result'")
	}
	if !contains(str, "Success: true") {
		t.Error("String should contain 'Success: true'")
	}
	if !contains(str, "utf-8") {
		t.Error("String should contain encoding")
	}
	if !contains(str, "Warnings: 2") {
		t.Error("String should contain warning count")
	}
}

func TestRecoveryResultEmpty(t *testing.T) {
	result := &RecoveryResult{
		Success:       true,
		OriginalSize:  100,
		RecoveredSize: 100,
	}

	str := result.String()
	if str == "" {
		t.Error("String should not be empty")
	}
}

func TestRecovererNew(t *testing.T) {
	recoverer := NewRecoverer(nil)
	if recoverer == nil {
		t.Error("NewRecoverer should not return nil")
	}
	if recoverer.Options == nil {
		t.Error("Options should not be nil")
	}
	if recoverer.Logger == nil {
		t.Error("Logger should not be nil")
	}
}

func TestRecovererWithOptions(t *testing.T) {
	options := &RecoveryOptions{
		Mode:            RecoveryAggressive,
		DetectEncoding:  false,
		FixHandles:      false,
		FixReferences:   true,
		RemoveCorrupted: false,
		MaxErrors:       500,
	}

	recoverer := NewRecoverer(options)

	if recoverer.Options.Mode != RecoveryAggressive {
		t.Error("Mode should be RecoveryAggressive")
	}
	if recoverer.Options.DetectEncoding {
		t.Error("DetectEncoding should be false")
	}
	if recoverer.Options.MaxErrors != 500 {
		t.Errorf("Expected MaxErrors 500, got %d", recoverer.Options.MaxErrors)
	}
}

func TestAuditLoggerCreation(t *testing.T) {
	logger := NewAuditLogger(true)
	if logger == nil {
		t.Error("NewAuditLogger should not return nil")
	}
	if !logger.Verbose {
		t.Error("Logger should be verbose when set to true")
	}

	logger2 := NewAuditLogger(false)
	if logger2.Verbose {
		t.Error("Logger should not be verbose when set to false")
	}
}

func TestAuditLoggerLogMessage(t *testing.T) {
	logger := NewAuditLogger(true)
	logger.LogMessage("Test message")
}

func TestAuditConfigNonDefault(t *testing.T) {
	config := &AuditConfig{
		AutoFix:      false,
		StrictMode:   true,
		DetectCycles: false,
		MaxErrors:    100,
	}

	if config.AutoFix {
		t.Error("AutoFix should be false")
	}
	if !config.StrictMode {
		t.Error("StrictMode should be true")
	}
	if config.DetectCycles {
		t.Error("DetectCycles should be false")
	}
	if config.MaxErrors != 100 {
		t.Errorf("Expected MaxErrors 100, got %d", config.MaxErrors)
	}
}

func TestValidationResultWithWarnings(t *testing.T) {
	result := &ValidationResult{Valid: true}

	result.AddWarning("Warning 1")
	result.AddWarning("Warning 2")

	if !result.Valid {
		t.Error("Result should still be valid with only warnings (AddWarning doesn't invalidate)")
	}
	if len(result.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(result.Warnings))
	}
}

func TestEntityValidatorWithLayers(t *testing.T) {
	validator := NewEntityValidator()
	layers := map[string]*table.Layer{
		"Layer1": nil,
		"Layer2": nil,
	}
	validator.SetValidLayers(layers)

	if !validator.validLayers["Layer1"] {
		t.Error("Layer1 should be valid")
	}
	if !validator.validLayers["Layer2"] {
		t.Error("Layer2 should be valid")
	}
	if validator.validLayers["Layer3"] {
		t.Error("Layer3 should not be valid")
	}
}

func TestEntityValidatorWithStyles(t *testing.T) {
	validator := NewEntityValidator()
	styles := map[string]*table.Style{
		"Style1": nil,
	}
	validator.SetValidStyles(styles)

	if !validator.validStyles["Style1"] {
		t.Error("Style1 should be valid")
	}
}

func TestErrorCodeGetCategoryEdgeCases(t *testing.T) {
	testCases := []ErrorCode{
		ErrorMissingSection,
		ErrorUndefinedLayer,
		ErrorInvalidColor,
		ErrorInvalidPoint,
	}

	for _, code := range testCases {
		category := code.GetCategory()
		if category == "" {
			t.Errorf("Error code %d should have a category", code)
		}
	}
}

func TestErrorSeverityString(t *testing.T) {
	if SeverityInfo.String() != "INFO" {
		t.Errorf("SeverityInfo.String() = %s, expected INFO", SeverityInfo.String())
	}
	if SeverityWarning.String() != "WARNING" {
		t.Errorf("SeverityWarning.String() = %s, expected WARNING", SeverityWarning.String())
	}
	if SeverityError.String() != "ERROR" {
		t.Errorf("SeverityError.String() = %s, expected ERROR", SeverityError.String())
	}
	if SeverityCritical.String() != "CRITICAL" {
		t.Errorf("SeverityCritical.String() = %s, expected CRITICAL", SeverityCritical.String())
	}
}
