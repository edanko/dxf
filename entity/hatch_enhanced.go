package entity

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// Hatch gradient types (from Python ezdxf)
const (
	HatchGradientTypeNone             = 0
	HatchGradientTypeLinear           = 1
	HatchGradientTypeCylinder         = 2
	HatchGradientTypeInvCylinder      = 3
	HatchGradientTypeSpherical        = 4
	HatchGradientTypeInvSpherical     = 5
	HatchGradientTypeHemisperical     = 6
	HatchGradientTypeInvHemispherical = 7
	HatchGradientTypeCurved           = 8
	HatchGradientTypeInvCurved        = 9
)

// HatchGradient represents gradient fill (enhanced from Python ezdxf)
type HatchGradient struct {
	Type     int               // Gradient type
	Color1   color.ColorNumber // Start color
	Color2   color.ColorNumber // End color
	Angle    float64           // Gradient angle
	Defocus  float64           // Focus distance
	Center   []float64         // Gradient center point [x, y] (for radial/cylinder)
	Tint     float64           // Tint value (0-1, 0=full color1, 1=full color2)
	Centered bool              // Whether gradient is centered
	HasTint  bool              // Whether tint is enabled
	HasFocus bool              // Whether defocus is enabled
}

// HatchEnhanced represents enhanced HATCH entity
type HatchEnhanced struct {
	*entity

	// Basic properties
	PatternType   int  // Pattern type (predefined, user-defined, custom)
	Style         int  // Hatch style (normal, outer, ignore)
	IsSolid       bool // Solid fill flag
	IsAssociative bool // Associative hatch flag

	// Pattern parameters
	PatternName    string  // Predefined pattern name
	PatternAngle   float64 // Pattern angle
	PatternScale   float64 // Pattern scale
	PatternDouble  bool    // Double hatch pattern
	PatternSpacing float64 // Pattern spacing

	// Advanced pattern support
	PatternLines   []HatchPatternLine // Custom pattern lines
	HasCustomLines bool               // Whether custom pattern lines are defined

	// Gradient parameters
	Gradient    *HatchGradient // Gradient fill data
	HasGradient bool           // Has gradient fill

	// Boundary data
	BoundaryPaths    []HatchBoundaryPath // Boundary paths
	HasEdgePaths     bool                // Whether edge paths are used
	HasPolylinePaths bool                // Whether polyline paths are used

	// Advanced options
	NumberOfSeed    int               // Number of seed lines (for some patterns)
	HatchColor      color.ColorNumber // Hatch color
	FillMode        int               // Fill mode
	BackgroundColor color.ColorNumber // Background color
	Definition      string            // Custom hatch definition

	// Transformation support
	Elevation    []float64 // Elevation point [x, y, z]
	Extrusion    []float64 // Extrusion direction [x, y, z]
	HasElevation bool      // Whether elevation is specified
	HasExtrusion bool      // Whether extrusion is specified
}

// Hatch boundary path types (from Python ezdxf)
const (
	HatchBoundaryTypePolyline = 0 // Polyline path with bulges
	HatchBoundaryTypeEdge     = 1 // Edge path (line, arc, ellipse, spline)
)

// Hatch boundary path flags (from Python ezdxf)
const (
	HatchBoundaryFlagExternal  = 1  // Main boundary
	HatchBoundaryFlagOutermost = 16 // Primary island boundary
	HatchBoundaryFlagDefault   = 0  // Secondary island boundaries
	HatchBoundaryFlagDerived   = 4  // Computed boundaries
	HatchBoundaryFlagTextbox   = 8  // Text containment boundaries
)

// HatchBoundaryPath represents a boundary path for hatch
type HatchBoundaryPath struct {
	Type         int         // Path type (polyline, edge, etc.)
	Flags        int         // Path flags (external, outermost, etc.)
	Points       [][]float64 // Path points
	Bulges       []float64   // Bulge values (for arc segments)
	IsClosed     bool        // Path closed flag
	EdgeTypes    []int       // For edge paths: type of each edge
	HasEdgeTypes bool        // Whether edge types are specified
}

// NewHatchEnhanced creates a new enhanced HATCH entity
func NewHatchEnhanced() *HatchEnhanced {
	return &HatchEnhanced{
		entity:          NewEntity(HATCH),
		PatternType:     HatchPatternTypePredefined,
		Style:           HatchStyleNormal,
		IsSolid:         false,
		IsAssociative:   false,
		PatternName:     "",
		PatternAngle:    0.0,
		PatternScale:    1.0,
		PatternDouble:   false,
		PatternSpacing:  1.0,
		Gradient:        nil,
		HasGradient:     false,
		BoundaryPaths:   make([]HatchBoundaryPath, 0),
		NumberOfSeed:    0,
		HatchColor:      0,
		FillMode:        0,
		BackgroundColor: 0,
		Definition:      "",
	}
}

// IsEntity is for Entity interface
func (h *HatchEnhanced) IsEntity() bool {
	return true
}

// SetPredefinedPattern sets predefined pattern
func (h *HatchEnhanced) SetPredefinedPattern(name string, angle, scale, spacing float64, isDouble bool) {
	h.PatternType = HatchPatternTypePredefined
	h.PatternName = name
	h.PatternAngle = angle
	h.PatternScale = scale
	h.PatternSpacing = spacing
	h.PatternDouble = isDouble
}

// SetGradientLinear sets linear gradient
func (h *HatchEnhanced) SetGradientLinear(color1, color2 color.ColorNumber, angle float64) {
	h.Gradient = &HatchGradient{
		Type:     HatchGradientTypeLinear,
		Color1:   color1,
		Color2:   color2,
		Angle:    angle,
		Centered: false,
		HasTint:  false,
		HasFocus: false,
	}
	h.HasGradient = true
}

// SetGradientSpherical sets spherical gradient
func (h *HatchEnhanced) SetGradientSpherical(color1, color2 color.ColorNumber, center []float64) {
	h.Gradient = &HatchGradient{
		Type:     HatchGradientTypeSpherical,
		Color1:   color1,
		Color2:   color2,
		Center:   center,
		Centered: true,
		HasTint:  false,
		HasFocus: false,
	}
	h.HasGradient = true
}

// SetGradientHemispherical sets hemispherical gradient
func (h *HatchEnhanced) SetGradientHemispherical(color1, color2 color.ColorNumber, center []float64) {
	h.Gradient = &HatchGradient{
		Type:     HatchGradientTypeHemisperical,
		Color1:   color1,
		Color2:   color2,
		Center:   center,
		Centered: true,
		HasTint:  false,
		HasFocus: false,
	}
	h.HasGradient = true
}

// SetGradientCylindrical sets cylindrical gradient
func (h *HatchEnhanced) SetGradientCylindrical(color1, color2 color.ColorNumber, center []float64) {
	h.Gradient = &HatchGradient{
		Type:     HatchGradientTypeCylinder,
		Color1:   color1,
		Color2:   color2,
		Center:   center,
		Centered: false,
		HasTint:  false,
		HasFocus: false,
	}
	h.HasGradient = true
}

// SetGradientWithTint sets gradient with tint control
func (h *HatchEnhanced) SetGradientWithTint(gradientType int, color1, color2 color.ColorNumber, angle float64, center []float64, tint float64) {
	h.Gradient = &HatchGradient{
		Type:     gradientType,
		Color1:   color1,
		Color2:   color2,
		Angle:    angle,
		Center:   center,
		Tint:     tint,
		Centered: false,
		HasTint:  true,
		HasFocus: false,
	}
	h.HasGradient = true
}

// SetGradientWithFocus sets gradient with focus control
func (h *HatchEnhanced) SetGradientWithFocus(gradientType int, color1, color2 color.ColorNumber, angle float64, center []float64, focus float64) {
	h.Gradient = &HatchGradient{
		Type:     gradientType,
		Color1:   color1,
		Color2:   color2,
		Angle:    angle,
		Center:   center,
		Defocus:  focus,
		Centered: false,
		HasTint:  false,
		HasFocus: true,
	}
	h.HasGradient = true
}

// AddPolylineBoundary adds a polyline boundary path
func (h *HatchEnhanced) AddPolylineBoundary(points [][]float64, isClosed bool) {
	path := HatchBoundaryPath{
		Type:         HatchBoundaryTypePolyline,
		Flags:        HatchBoundaryFlagExternal,
		Points:       points,
		IsClosed:     isClosed,
		Bulges:       make([]float64, 0),
		HasEdgeTypes: false,
	}
	h.BoundaryPaths = append(h.BoundaryPaths, path)
}

// AddPolylineBoundaryWithBulges adds a polyline boundary path with bulges
func (h *HatchEnhanced) AddPolylineBoundaryWithBulges(points [][]float64, isClosed bool, bulges []float64) {
	path := HatchBoundaryPath{
		Type:         HatchBoundaryTypePolyline,
		Flags:        HatchBoundaryFlagExternal,
		Points:       points,
		IsClosed:     isClosed,
		Bulges:       bulges,
		HasEdgeTypes: false,
	}
	h.BoundaryPaths = append(h.BoundaryPaths, path)
}

// AddEdgeBoundary adds an edge boundary path
func (h *HatchEnhanced) AddEdgeBoundary(edgeTypes []int) {
	path := HatchBoundaryPath{
		Type:         HatchBoundaryTypeEdge,
		Flags:        HatchBoundaryFlagExternal,
		Points:       make([][]float64, 0),
		IsClosed:     false,
		Bulges:       make([]float64, 0),
		EdgeTypes:    edgeTypes,
		HasEdgeTypes: true,
	}
	h.BoundaryPaths = append(h.BoundaryPaths, path)
}

// AddIslandBoundary adds an island boundary (nested boundary)
func (h *HatchEnhanced) AddIslandBoundary(points [][]float64, isClosed bool) {
	path := HatchBoundaryPath{
		Type:         HatchBoundaryTypePolyline,
		Flags:        HatchBoundaryFlagOutermost,
		Points:       points,
		IsClosed:     isClosed,
		Bulges:       make([]float64, 0),
		HasEdgeTypes: false,
	}
	h.BoundaryPaths = append(h.BoundaryPaths, path)
}

// SetHatchColor sets the hatch color
func (h *HatchEnhanced) SetHatchColor(hatchColor color.ColorNumber) {
	h.HatchColor = hatchColor
}

// SetBackgroundColor sets background color
func (h *HatchEnhanced) SetBackgroundColor(bgColor color.ColorNumber) {
	h.BackgroundColor = bgColor
}

// SetSolidFill enables solid fill
func (h *HatchEnhanced) SetSolidFill(isSolid bool) {
	h.IsSolid = isSolid
}

// SetAssociative sets associative flag
func (h *HatchEnhanced) SetAssociative(isAssociative bool) {
	h.IsAssociative = isAssociative
}

// SetFillMode sets the fill mode
func (h *HatchEnhanced) SetFillMode(fillMode int) {
	h.FillMode = fillMode
}

// Format writes data to formatter
func (h *HatchEnhanced) Format(f format.Formatter) {
	h.entity.Format(f)
	f.WriteString(100, "AcDbHatch")

	// Basic properties
	f.WriteInt(91, h.PatternType)
	f.WriteInt(75, h.Style)

	// Pattern data
	if h.PatternType == HatchPatternTypePredefined && h.PatternName != "" {
		f.WriteString(2, h.PatternName)
	}
	if h.PatternType != HatchPatternTypePredefined {
		f.WriteString(52, h.Definition)
	}
	f.WriteInt(70, int(h.HatchColor))
	f.WriteInt(77, int(h.BackgroundColor))

	// Style flags
	if h.IsSolid {
		f.WriteInt(1, 1) // Solid fill
	} else {
		f.WriteInt(1, 0)
	}
	if h.IsAssociative {
		f.WriteInt(1, 1) // Associative
	} else {
		f.WriteInt(1, 0)
	}
	if h.Style == HatchStyleIgnore {
		f.WriteInt(1, 1) // Ignore nesting style
	} else {
		f.WriteInt(1, 0)
	}

	// Pattern parameters
	f.WriteString(52, h.PatternName)
	if h.PatternAngle != 0.0 {
		f.WriteFloat(53, h.PatternAngle)
	}
	if h.PatternScale != 1.0 {
		f.WriteFloat(41, h.PatternScale)
	}
	if h.PatternDouble {
		f.WriteInt(71, 1) // Double pattern
	} else {
		f.WriteInt(71, 0)
	}
	if h.PatternSpacing != 1.0 {
		f.WriteFloat(53, h.PatternSpacing)
	}
	f.WriteInt(78, h.NumberOfSeed)

	// Gradient data
	if h.HasGradient && h.Gradient != nil {
		f.WriteInt(98, h.Gradient.Type)
		f.WriteInt(63, int(h.Gradient.Color1))
		f.WriteInt(64, int(h.Gradient.Color2))
		if h.Gradient.Type == HatchGradientTypeLinear {
			if h.Gradient.Angle != 0.0 {
				f.WriteFloat(450, h.Gradient.Angle)
			}
		} else if h.Gradient.Type == HatchGradientTypeCylinder {
			// Write center point for cylindrical gradient
			if len(h.Gradient.Center) >= 2 {
				f.WriteFloat(110, h.Gradient.Center[0]) // X
				f.WriteFloat(120, h.Gradient.Center[1]) // Y
				if len(h.Gradient.Center) >= 3 {
					f.WriteFloat(130, h.Gradient.Center[2]) // Z
				}
			}
			f.WriteFloat(460, h.Gradient.Defocus) // Focus distance
		}
	}

	// Boundary paths data
	f.WriteInt(92, len(h.BoundaryPaths))

	// Write each boundary path
	for i, path := range h.BoundaryPaths {
		f.WriteInt(93, i+1) // Boundary path number
		f.WriteInt(72, path.Type)
		f.WriteInt(73, 0) // Closed flag

		// Write path points
		for _, point := range path.Points {
			for k := 0; k < len(point) && k < 3; k++ {
				f.WriteFloat(10+k*10, point[k])
			}
		}

		// Write bulges for arc segments
		for _, bulge := range path.Bulges {
			f.WriteFloat(42, bulge) // Bulge value
		}
	}

	// Fill mode
	if h.FillMode != 0 {
		f.WriteInt(76, h.FillMode)
	}
}

// BBox returns bounding box for Entity interface
func (h *HatchEnhanced) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	// Simplified bbox - could be improved by calculating actual bounds from boundary paths
	return mins, maxs
}

// String returns string representation
func (h *HatchEnhanced) String() string {
	var gradientType string
	if h.Gradient != nil {
		switch h.Gradient.Type {
		case 1:
			gradientType = "Linear"
		case 2:
			gradientType = "Cylindrical"
		default:
			gradientType = "None"
		}
	}

	return fmt.Sprintf("HatchEnhanced{Pattern: %s, Gradient: %s, Paths: %d}",
		h.PatternName, gradientType, len(h.BoundaryPaths))
}

// === Advanced Utility Methods for Python ezdxf Compatibility ===

// AddRectangleBoundary adds a rectangular boundary
func (h *HatchEnhanced) AddRectangleBoundary(minX, minY, maxX, maxY float64) {
	points := [][]float64{
		{minX, minY, 0},
		{maxX, minY, 0},
		{maxX, maxY, 0},
		{minX, maxY, 0},
	}
	h.AddPolylineBoundary(points, true)
}

// AddCircularBoundary adds a circular boundary using polyline approximation
func (h *HatchEnhanced) AddCircularBoundary(center []float64, radius float64, segments int) {
	if segments < 8 {
		segments = 8
	}

	points := make([][]float64, segments)
	for i := 0; i < segments; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(segments)
		x := center[0] + radius*math.Cos(angle)
		y := center[1] + radius*math.Sin(angle)
		z := 0.0
		if len(center) > 2 {
			z = center[2]
		}
		points[i] = []float64{x, y, z}
	}

	h.AddPolylineBoundary(points, true)
}

// SetUserDefinedPattern sets user-defined pattern with custom pattern lines
func (h *HatchEnhanced) SetUserDefinedPattern(patternLines []HatchPatternLine) {
	h.PatternType = HatchPatternTypeUserDefined
	h.PatternLines = patternLines
	h.HasCustomLines = true
	h.PatternName = ""
}

// SetElevation sets the hatch elevation point
func (h *HatchEnhanced) SetElevation(elevation []float64) {
	h.Elevation = elevation
	h.HasElevation = true
}

// SetExtrusion sets the hatch extrusion direction
func (h *HatchEnhanced) SetExtrusion(extrusion []float64) {
	h.Extrusion = extrusion
	h.HasExtrusion = true
}

// Transform applies transformation matrix to all boundary paths
func (h *HatchEnhanced) Transform(matrix dxfmath.Matrix44) {
	for i := range h.BoundaryPaths {
		path := &h.BoundaryPaths[i]
		for j := range path.Points {
			vec := dxfmath.NewVec3(path.Points[j][0], path.Points[j][1], path.Points[j][2])
			transformed := matrix.TransformVector(vec)
			path.Points[j] = []float64{transformed.X(), transformed.Y(), transformed.Z()}
		}
	}
}

// GetBoundaryCount returns number of boundary paths
func (h *HatchEnhanced) GetBoundaryCount() int {
	return len(h.BoundaryPaths)
}

// HasBoundaries returns true if hatch has any boundary paths
func (h *HatchEnhanced) HasBoundaries() bool {
	return len(h.BoundaryPaths) > 0
}

// IsGradientType returns true if hatch uses specified gradient type
func (h *HatchEnhanced) IsGradientType(gradientType int) bool {
	return h.Gradient != nil && h.Gradient.Type == gradientType
}

// IsPatternType returns true if hatch uses specified pattern type
func (h *HatchEnhanced) IsPatternType(patternType int) bool {
	return h.PatternType == patternType
}
