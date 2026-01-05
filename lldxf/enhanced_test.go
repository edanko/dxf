package lldxf

import (
	"testing"
)

func TestValidator(t *testing.T) {
	// Test basic validation functionality
	validator := NewValidator(DefaultValidationOptions())

	// Test with minimal valid structure
	minimalTags := Tags{
		NewDXFTag(0, "SECTION"),
		NewDXFTag(2, "HEADER"),
		NewDXFTag(9, "$ACADVER"),
		NewDXFTag(1, "AC1015"),
		NewDXFTag(0, "ENDSEC"),
	}

	info, err := validator.Validate(minimalTags)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if info.Version != "AC1015" {
		t.Errorf("Expected version AC1015, got %s", info.Version)
	}

	// Test with more complete structure to increase coverage
	completeTags := Tags{
		NewDXFTag(0, "SECTION"),
		NewDXFTag(2, "HEADER"),
		NewDXFTag(9, "$ACADVER"),
		NewDXFTag(1, "AC1015"),
		NewDXFTag(0, "ENDSEC"),

		NewDXFTag(0, "SECTION"),
		NewDXFTag(2, "ENTITIES"),
		NewDXFTag(0, "LINE"),
		NewDXFTag(10, 0.0),
		NewDXFTag(20, 0.0),
		NewDXFTag(30, 0.0),
		NewDXFTag(11, 0.0),
		NewDXFTag(21, 0.0),
		NewDXFTag(31, 0.0),
		NewDXFTag(0, "ENDSEC"),
	}

	info2, err2 := validator.Validate(completeTags)
	if err2 != nil {
		t.Fatalf("Validation failed: %v", err2)
	}

	// Just check that validation completes without errors
	if info2.Version != "AC1015" {
		t.Errorf("Expected version AC1015, got %s", info2.Version)
	}
}
