package audit

import (
	"testing"
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
	for i := 0; i < b.N; i++ {
		_ = NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Benchmark error")
	}
}

// BenchmarkErrorString benchmarks error string formatting performance
func BenchmarkErrorString(b *testing.B) {
	err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", "123", "Benchmark error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.String()
	}
}
