package patterns

import (
	"math"
)

// PatternCategory represents pattern categories
type PatternCategory string

const (
	PatternCategoryANSI     PatternCategory = "ANSI"
	PatternCategoryISO      PatternCategory = "ISO"
	PatternCategoryACAD     PatternCategory = "ACAD"
	PatternCategoryArch     PatternCategory = "ARCHITECTURAL"
	PatternCategoryGeo      PatternCategory = "GEOMETRIC"
	PatternCategoryMaterial PatternCategory = "MATERIAL"
	PatternCategoryCustom   PatternCategory = "CUSTOM"
)

// PatternDefinition represents a predefined hatch pattern
type PatternDefinition struct {
	Name        string
	Description string
	Category    PatternCategory
	Lines       []PatternLine
}

// PatternLine represents a line in a hatch pattern
type PatternLine struct {
	Angle           float64
	BasePoint       []float64
	Offset          []float64
	DashLengthItems []float64
}

// ANSI 31 Standard Patterns (Most Critical for Engineering)
var PredefinedPatterns = map[string]PatternDefinition{
	"ANSI31": {
		Name:        "ANSI31",
		Description: "Iron, Brick, Stone masonry",
		Category:    PatternCategoryANSI,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{0.125, -0.125, 0.125, -0.125, -0.125, -0.125}},
			{Angle: 135.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, -0.125, -0.125, -0.125, -0.125}},
		},
	},
	"ANSI32": {
		Name:        "ANSI32",
		Description: "Steel",
		Category:    PatternCategoryANSI,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0.125, 0}, DashLengthItems: []float64{0.375, -0.125, 0.125, 0.125, -0.125, -0.125}},
			{Angle: 135.0, BasePoint: []float64{0, 0}, Offset: []float64{0.125, 0}, DashLengthItems: []float64{0.375, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
		},
	},
	"ANSI33": {
		Name:        "ANSI33",
		Description: "Brass, Bronze or Copper",
		Category:    PatternCategoryANSI,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
			{Angle: 135.0, BasePoint: []float64{0, 0}, Offset: []float64{0.125, 0}, DashLengthItems: []float64{0.125, -0.125, 0.125, 0.125, 0.125, 0.125}},
		},
	},
	"ANSI34": {
		Name:        "ANSI34",
		Description: "Plastic, Rubber",
		Category:    PatternCategoryANSI,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{0.375, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125, -0.125}},
			{Angle: 135.0, BasePoint: []float64{0, 0}, Offset: []float64{0.125, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
		},
	},
	"ANSI35": {
		Name:        "ANSI35",
		Description: "Firebrick and Refractory material",
		Category:    PatternCategoryANSI,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, -0.125, -0.125, -0.125, -0.125}},
			{Angle: 135.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, -0.125, -0.125, -0.125, -0.125}},
		},
	},
	"ANSI36": {
		Name:        "ANSI36",
		Description: "Concrete",
		Category:    PatternCategoryANSI,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{0.375, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
			{Angle: 135.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
		},
	},
	"ANSI37": {
		Name:        "ANSI37",
		Description: "Earth",
		Category:    PatternCategoryANSI,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{0.375, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
			{Angle: 135.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
		},
	},
	"ANSI38": {
		Name:        "ANSI38",
		Description: "Translucent Material",
		Category:    PatternCategoryANSI,
		Lines: []PatternLine{
			{Angle: 45.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
			{Angle: 135.0, BasePoint: []float64{0, 0}, Offset: []float64{0, 0}, DashLengthItems: []float64{-0.125, -0.125, 0.125, 0.125, -0.125, -0.125, -0.125}},
		},
	},
}

// GetPatternDefinition retrieves a predefined pattern by name
func GetPatternDefinition(name string) (PatternDefinition, bool) {
	pattern, exists := PredefinedPatterns[name]
	return pattern, exists
}

// GetPatternsByCategory returns all patterns in a category
func GetPatternsByCategory(category PatternCategory) map[string]PatternDefinition {
	result := make(map[string]PatternDefinition)
	for name, pattern := range PredefinedPatterns {
		if pattern.Category == category {
			result[name] = pattern
		}
	}
	return result
}

// GetAllPatterns returns all predefined patterns
func GetAllPatterns() map[string]PatternDefinition {
	result := make(map[string]PatternDefinition)
	for name, pattern := range PredefinedPatterns {
		result[name] = pattern
	}
	return result
}

// PatternScale represents pattern scaling for different measurement systems
type PatternScale struct {
	Name        string
	Description string
	Factor      float64
}

// Standard measurement scales
var PatternScales = []PatternScale{
	{"Imperial", "Imperial (1/25.4 scale factor)", 25.4},
	{"Metric", "Metric (1.0 scale factor)", 1.0},
	{"Architectural", "Architectural (1/8 scale factor)", 8.0},
	{"Engineering", "Engineering (1/16 scale factor)", 16.0},
}

// GetScaleByName retrieves scale by name
func GetScaleByName(name string) (PatternScale, bool) {
	for _, scale := range PatternScales {
		if scale.Name == name {
			return scale, true
		}
	}
	return PatternScale{}, false
}

// GetAllScales returns all available scales
func GetAllScales() []PatternScale {
	return PatternScales
}

// PatternAngle represents pattern rotation angles
type PatternAngle struct {
	Name        string
	Description string
	Degrees     float64
	Radians     float64
}

// Standard pattern angles
var PatternAngles = []PatternAngle{
	{"0°", "No rotation", 0.0, 0.0},
	{"45°", "45° diagonal", 45.0, math.Pi / 4},
	{"90°", "90° vertical", 90.0, math.Pi / 2},
	{"135°", "135° diagonal", 135.0, 3 * math.Pi / 4},
}

// GetAngleByName retrieves angle by name
func GetAngleByName(name string) (PatternAngle, bool) {
	for _, angle := range PatternAngles {
		if angle.Name == name {
			return angle, true
		}
	}
	return PatternAngle{}, false
}

// GetAllAngles returns all available angles
func GetAllAngles() []PatternAngle {
	return PatternAngles
}

// PatternType represents pattern types
type PatternType struct {
	Name        string
	Description string
	Category    PatternCategory
}

var PatternTypes = []PatternType{
	{"Predefined", "Predefined patterns from standard libraries", PatternCategoryACAD},
	{"UserDefined", "Custom user-defined patterns", PatternCategoryCustom},
	{"Custom", "User-defined hatch patterns", PatternCategoryCustom},
}

// GetPatternTypeByName retrieves pattern type by name
func GetPatternTypeByName(name string) (PatternType, bool) {
	for _, ptype := range PatternTypes {
		if ptype.Name == name {
			return ptype, true
		}
	}
	return PatternType{}, false
}

// GetAllPatternTypes returns all pattern types
func GetAllPatternTypes() []PatternType {
	return PatternTypes
}

// MeasurementSystem represents measurement systems
type MeasurementSystem struct {
	Name        string
	Description string
	ScaleFactor float64
	Unit        string
}

var MeasurementSystems = []MeasurementSystem{
	{"Imperial", "Imperial system (feet and inches)", 25.4, "inches"},
	{"Metric", "Metric system (mm and cm)", 1.0, "mm"},
	{"Architectural", "Architectural units", 8.0, "feet"},
	{"Engineering", "Engineering units", 16.0, "feet"},
}

// GetMeasurementSystemByName retrieves measurement system by name
func GetMeasurementSystemByName(name string) (MeasurementSystem, bool) {
	for _, system := range MeasurementSystems {
		if system.Name == name {
			return system, true
		}
	}
	return MeasurementSystem{}, false
}

// GetAllMeasurementSystems returns all measurement systems
func GetAllMeasurementSystems() []MeasurementSystem {
	return MeasurementSystems
}
