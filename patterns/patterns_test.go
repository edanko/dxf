package patterns

import (
	"math"
	"testing"
)

func TestPatternCategoryConstants(t *testing.T) {
	if PatternCategoryANSI != "ANSI" {
		t.Errorf("Expected 'ANSI', got %s", PatternCategoryANSI)
	}
	if PatternCategoryISO != "ISO" {
		t.Errorf("Expected 'ISO', got %s", PatternCategoryISO)
	}
	if PatternCategoryACAD != "ACAD" {
		t.Errorf("Expected 'ACAD', got %s", PatternCategoryACAD)
	}
	if PatternCategoryArch != "ARCHITECTURAL" {
		t.Errorf("Expected 'ARCHITECTURAL', got %s", PatternCategoryArch)
	}
	if PatternCategoryGeo != "GEOMETRIC" {
		t.Errorf("Expected 'GEOMETRIC', got %s", PatternCategoryGeo)
	}
	if PatternCategoryMaterial != "MATERIAL" {
		t.Errorf("Expected 'MATERIAL', got %s", PatternCategoryMaterial)
	}
	if PatternCategoryCustom != "CUSTOM" {
		t.Errorf("Expected 'CUSTOM', got %s", PatternCategoryCustom)
	}
}

func TestGetPatternDefinition(t *testing.T) {
	pattern, exists := GetPatternDefinition("ANSI31")
	if !exists {
		t.Error("ANSI31 pattern should exist")
	}
	if pattern.Name != "ANSI31" {
		t.Errorf("Expected name 'ANSI31', got %s", pattern.Name)
	}
	if pattern.Category != PatternCategoryANSI {
		t.Errorf("Expected category ANSI, got %s", pattern.Category)
	}
	if len(pattern.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(pattern.Lines))
	}

	_, exists = GetPatternDefinition("NONEXISTENT")
	if exists {
		t.Error("Non-existent pattern should not exist")
	}
}

func TestGetPatternsByCategory(t *testing.T) {
	patterns := GetPatternsByCategory(PatternCategoryANSI)
	if len(patterns) == 0 {
		t.Error("Should have ANSI patterns")
	}

	for name, pattern := range patterns {
		if pattern.Category != PatternCategoryANSI {
			t.Errorf("Pattern %s should have ANSI category", name)
		}
	}
}

func TestGetAllPatterns(t *testing.T) {
	patterns := GetAllPatterns()
	if len(patterns) == 0 {
		t.Error("Should have predefined patterns")
	}
	if len(patterns) != len(PredefinedPatterns) {
		t.Errorf("Expected %d patterns, got %d", len(PredefinedPatterns), len(patterns))
	}
}

func TestPatternScales(t *testing.T) {
	scale, exists := GetScaleByName("Metric")
	if !exists {
		t.Error("Metric scale should exist")
	}
	if scale.Factor != 1.0 {
		t.Errorf("Expected factor 1.0, got %f", scale.Factor)
	}

	scale, exists = GetScaleByName("Imperial")
	if !exists {
		t.Error("Imperial scale should exist")
	}
	if scale.Factor != 25.4 {
		t.Errorf("Expected factor 25.4, got %f", scale.Factor)
	}

	_, exists = GetScaleByName("Nonexistent")
	if exists {
		t.Error("Non-existent scale should not exist")
	}
}

func TestGetAllScales(t *testing.T) {
	scales := GetAllScales()
	if len(scales) != len(PatternScales) {
		t.Errorf("Expected %d scales, got %d", len(PatternScales), len(scales))
	}
}

func TestPatternAngles(t *testing.T) {
	angle, exists := GetAngleByName("45°")
	if !exists {
		t.Error("45° angle should exist")
	}
	if angle.Degrees != 45.0 {
		t.Errorf("Expected 45.0 degrees, got %f", angle.Degrees)
	}
	if angle.Radians != math.Pi/4 {
		t.Errorf("Expected %f radians, got %f", math.Pi/4, angle.Radians)
	}

	angle, exists = GetAngleByName("90°")
	if !exists {
		t.Error("90° angle should exist")
	}
	if angle.Radians != math.Pi/2 {
		t.Errorf("Expected %f radians, got %f", math.Pi/2, angle.Radians)
	}
}

func TestGetAllAngles(t *testing.T) {
	angles := GetAllAngles()
	if len(angles) != 4 {
		t.Errorf("Expected 4 angles, got %d", len(angles))
	}
}

func TestPatternTypes(t *testing.T) {
	ptype, exists := GetPatternTypeByName("Predefined")
	if !exists {
		t.Error("Predefined type should exist")
	}
	if ptype.Category != PatternCategoryACAD {
		t.Errorf("Expected ACAD category, got %s", ptype.Category)
	}

	_, exists = GetPatternTypeByName("Nonexistent")
	if exists {
		t.Error("Non-existent type should not exist")
	}
}

func TestGetAllPatternTypes(t *testing.T) {
	types := GetAllPatternTypes()
	if len(types) != 3 {
		t.Errorf("Expected 3 types, got %d", len(types))
	}
}

func TestMeasurementSystems(t *testing.T) {
	system, exists := GetMeasurementSystemByName("Metric")
	if !exists {
		t.Error("Metric system should exist")
	}
	if system.ScaleFactor != 1.0 {
		t.Errorf("Expected scale factor 1.0, got %f", system.ScaleFactor)
	}
	if system.Unit != "mm" {
		t.Errorf("Expected unit 'mm', got %s", system.Unit)
	}

	system, exists = GetMeasurementSystemByName("Imperial")
	if !exists {
		t.Error("Imperial system should exist")
	}
	if system.ScaleFactor != 25.4 {
		t.Errorf("Expected scale factor 25.4, got %f", system.ScaleFactor)
	}

	_, exists = GetMeasurementSystemByName("Nonexistent")
	if exists {
		t.Error("Non-existent system should not exist")
	}
}

func TestGetAllMeasurementSystems(t *testing.T) {
	systems := GetAllMeasurementSystems()
	if len(systems) != 4 {
		t.Errorf("Expected 4 systems, got %d", len(systems))
	}
}

func TestPatternDefinition(t *testing.T) {
	pattern := PatternDefinition{
		Name:        "TEST",
		Description: "Test pattern",
		Category:    PatternCategoryCustom,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0.1, 0}, DashLengthItems: []float64{0.5, -0.25}},
		},
	}

	if pattern.Name != "TEST" {
		t.Errorf("Expected name 'TEST', got %s", pattern.Name)
	}
	if len(pattern.Lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(pattern.Lines))
	}
}

func TestPatternLine(t *testing.T) {
	line := PatternLine{
		Angle:           45.0,
		BasePoint:       []float64{1, 2},
		Offset:          []float64{0.1, 0.2},
		DashLengthItems: []float64{0.5, -0.25, 0.25},
	}

	if line.Angle != 45.0 {
		t.Errorf("Expected angle 45.0, got %f", line.Angle)
	}
	if line.BasePoint[0] != 1 || line.BasePoint[1] != 2 {
		t.Errorf("Expected BasePoint [1,2], got [%f,%f]", line.BasePoint[0], line.BasePoint[1])
	}
	if len(line.DashLengthItems) != 3 {
		t.Errorf("Expected 3 dash items, got %d", len(line.DashLengthItems))
	}
}

func TestPatternScalesLength(t *testing.T) {
	if len(PatternScales) != 4 {
		t.Errorf("Expected 4 standard scales, got %d", len(PatternScales))
	}
}

func TestPatternAnglesLength(t *testing.T) {
	if len(PatternAngles) != 4 {
		t.Errorf("Expected 4 standard angles, got %d", len(PatternAngles))
	}
}

func TestPredefinedPatternsCount(t *testing.T) {
	if len(PredefinedPatterns) < 8 {
		t.Errorf("Expected at least 8 ANSI patterns, got %d", len(PredefinedPatterns))
	}
}

func TestISOPatterns(t *testing.T) {
	names := []string{
		"ACAD_ISO02W100", "ACAD_ISO03W100", "ACAD_ISO04W100", "ACAD_ISO05W100",
		"ACAD_ISO06W100", "ACAD_ISO07W100", "ACAD_ISO08W100", "ACAD_ISO09W100",
		"ACAD_ISO10W100", "ACAD_ISO11W100", "ACAD_ISO12W100", "ACAD_ISO13W100",
		"ACAD_ISO14W100", "ACAD_ISO15W100",
	}
	for _, name := range names {
		def, ok := GetPatternDefinition(name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		if def.Category != PatternCategoryISO {
			t.Fatalf("%s category = %s", name, def.Category)
		}
		if len(def.Lines) == 0 {
			t.Fatalf("%s has no lines", name)
		}
	}
	iso := GetPatternsByCategory(PatternCategoryISO)
	if len(iso) < len(names) {
		t.Fatalf("ISO patterns = %d, want at least %d", len(iso), len(names))
	}
}
