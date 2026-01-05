package entity

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// VisualStyleType represents different visual style types
type VisualStyleType int

const (
	VisualStyleTypeFlat             VisualStyleType = 0
	VisualStyleTypeFlatWithEdges    VisualStyleType = 1
	VisualStyleTypeGouraud          VisualStyleType = 2
	VisualStyleTypeGouraudWithEdges VisualStyleType = 3
	VisualStyleType2dWireframe      VisualStyleType = 4
	VisualStyleTypeWireframe        VisualStyleType = 5
	VisualStyleTypeHidden           VisualStyleType = 6
	VisualStyleTypeBasic            VisualStyleType = 7
	VisualStyleTypeRealistic        VisualStyleType = 8
	VisualStyleTypeConceptual       VisualStyleType = 9
	VisualStyleTypeDim              VisualStyleType = 11
	VisualStyleTypeBrighten         VisualStyleType = 12
	VisualStyleTypeThicken          VisualStyleType = 13
	VisualStyleTypeLinePattern      VisualStyleType = 14
	VisualStyleTypeColorChange      VisualStyleType = 15
	VisualStyleTypeFacePattern      VisualStyleType = 16
	VisualStyleTypeShadedWithEdges  VisualStyleType = 17
	VisualStyleTypeShaded           VisualStyleType = 18
)

// String returns string representation of visual style type
func (vst VisualStyleType) String() string {
	switch vst {
	case VisualStyleTypeFlat:
		return "Flat"
	case VisualStyleTypeFlatWithEdges:
		return "FlatWithEdges"
	case VisualStyleTypeGouraud:
		return "Gouraud"
	case VisualStyleTypeGouraudWithEdges:
		return "GouraudWithEdges"
	case VisualStyleType2dWireframe:
		return "2dWireframe"
	case VisualStyleTypeWireframe:
		return "Wireframe"
	case VisualStyleTypeHidden:
		return "Hidden"
	case VisualStyleTypeBasic:
		return "Basic"
	case VisualStyleTypeRealistic:
		return "Realistic"
	case VisualStyleTypeConceptual:
		return "Conceptual"
	case VisualStyleTypeDim:
		return "Dim"
	case VisualStyleTypeBrighten:
		return "Brighten"
	case VisualStyleTypeThicken:
		return "Thicken"
	case VisualStyleTypeLinePattern:
		return "LinePattern"
	case VisualStyleTypeColorChange:
		return "ColorChange"
	case VisualStyleTypeFacePattern:
		return "FacePattern"
	case VisualStyleTypeShadedWithEdges:
		return "ShadedWithEdges"
	case VisualStyleTypeShaded:
		return "Shaded"
	default:
		return "Unknown"
	}
}

// FaceLightingModel represents face lighting models
type FaceLightingModel int

const (
	FaceLightingInvisible FaceLightingModel = 0
	FaceLightingVisible   FaceLightingModel = 1
	FaceLightingPhong     FaceLightingModel = 2
	FaceLightingGooch     FaceLightingModel = 3
)

// String returns string representation of face lighting model
func (flm FaceLightingModel) String() string {
	switch flm {
	case FaceLightingInvisible:
		return "Invisible"
	case FaceLightingVisible:
		return "Visible"
	case FaceLightingPhong:
		return "Phong"
	case FaceLightingGooch:
		return "Gooch"
	default:
		return "Unknown"
	}
}

// FaceLightingQuality represents face lighting quality levels
type FaceLightingQuality int

const (
	FaceLightingNoLighting FaceLightingQuality = 0
	FaceLightingPerFace    FaceLightingQuality = 1
	FaceLightingPerVertex  FaceLightingQuality = 2
	FaceLightingUnknown    FaceLightingQuality = 3
)

// String returns string representation of face lighting quality
func (flq FaceLightingQuality) String() string {
	switch flq {
	case FaceLightingNoLighting:
		return "NoLighting"
	case FaceLightingPerFace:
		return "PerFace"
	case FaceLightingPerVertex:
		return "PerVertex"
	case FaceLightingUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// EdgeStyleModel represents edge style models
type EdgeStyleModel int

const (
	EdgeStyleNoEdges    EdgeStyleModel = 0
	EdgeStyleIsolines   EdgeStyleModel = 1
	EdgeStyleFacetEdges EdgeStyleModel = 2
)

// String returns string representation of edge style model
func (esm EdgeStyleModel) String() string {
	switch esm {
	case EdgeStyleNoEdges:
		return "NoEdges"
	case EdgeStyleIsolines:
		return "Isolines"
	case EdgeStyleFacetEdges:
		return "FacetEdges"
	default:
		return "Unknown"
	}
}

// FaceColorMode represents face color modes
type FaceColorMode int

const (
	FaceColorNoColor         FaceColorMode = 0
	FaceColorObjectColor     FaceColorMode = 1
	FaceColorBackgroundColor FaceColorMode = 2
	FaceColorMonoColor       FaceColorMode = 3
	FaceColorCustomColor     FaceColorMode = 4
)

// String returns string representation of face color mode
func (fcm FaceColorMode) String() string {
	switch fcm {
	case FaceColorNoColor:
		return "NoColor"
	case FaceColorObjectColor:
		return "ObjectColor"
	case FaceColorBackgroundColor:
		return "BackgroundColor"
	case FaceColorMonoColor:
		return "MonoColor"
	case FaceColorCustomColor:
		return "CustomColor"
	default:
		return "Unknown"
	}
}

// EdgeColorMode represents edge color modes
type EdgeColorMode int

const (
	EdgeColorEntityColor EdgeColorMode = 0
	EdgeColorACIColor    EdgeColorMode = 1
	EdgeColorMonoColor   EdgeColorMode = 2
)

// String returns string representation of edge color mode
func (ecm EdgeColorMode) String() string {
	switch ecm {
	case EdgeColorEntityColor:
		return "EntityColor"
	case EdgeColorACIColor:
		return "ACIColor"
	case EdgeColorMonoColor:
		return "MonoColor"
	default:
		return "Unknown"
	}
}

// EdgeModifiers represents edge modifier flags
type EdgeModifiers int

const (
	EdgeModifierNone     EdgeModifiers = 0
	EdgeModifierOpacity  EdgeModifiers = 1
	EdgeModifierSpecular EdgeModifiers = 2
)

// String returns string representation of edge modifiers
func (em EdgeModifiers) String() string {
	switch em {
	case EdgeModifierNone:
		return "None"
	case EdgeModifierOpacity:
		return "Opacity"
	case EdgeModifierSpecular:
		return "Specular"
	default:
		return "Unknown"
	}
}

// VisualStyle represents a DXF VISUALSTYLE entity for advanced rendering
type VisualStyle struct {
	*entity
	handle string

	// Core properties
	Description string          `dxf:"2"`  // Style description
	StyleType   VisualStyleType `dxf:"70"` // Visual style type

	// Face lighting properties
	FaceLightingModel   FaceLightingModel   `dxf:"71"` // Face lighting model
	FaceLightingQuality FaceLightingQuality `dxf:"72"` // Face lighting quality
	FaceColorMode       FaceColorMode       `dxf:"73"` // Face color mode

	// Edge properties
	EdgeStyleModel           EdgeStyleModel    `dxf:"74"`  // Edge style model
	EdgeStyle                int               `dxf:"91"`  // Edge style
	EdgeColorMode            EdgeColorMode     `dxf:"66"`  // Edge color mode
	EdgeIntersectionColor    color.ColorNumber `dxf:"64"`  // Edge intersection color
	EdgeObscuredColor        color.ColorNumber `dxf:"65"`  // Edge obscured color
	EdgeLinetype             string            `dxf:"75"`  // Edge linetype
	EdgeObliqueType          int               `dxf:"175"` // Edge oblique type
	EdgeCreaseAngle          float64           `dxf:"42"`  // Edge crease angle
	EdgeObliqueSize          float64           `dxf:"43"`  // Edge oblique size
	EdgeModifiers            EdgeModifiers     `dxf:"92"`  // Edge modifier flags
	EdgeWidth                float64           `dxf:"40"`  // Edge width
	EdgeOverhang             float64           `dxf:"44"`  // Edge overhang
	EdgeIntersectionLinetype int               `dxf:"77"`  // Edge intersection linetype
	EdgeJitter               float64           `dxf:"78"`  // Edge jitter
	EdgeSilhouetteWidth      float64           `dxf:"79"`  // Edge silhouette width
	EdgeSilhouetteColor      color.ColorNumber `dxf:"67"`  // Edge silhouette color
	EdgeIsolineCount         int               `dxf:"171"` // Edge isoline count
	EdgeHidePrecision        bool              `dxf:"290"` // Hide precision flag
	EdgeStyleApply           bool              `dxf:"174"` // Apply edge style

	// Display properties
	Brightness float64           `dxf:"44"`  // Overall brightness
	ShadowType int               `dxf:"173"` // Shadow type
	Color1     color.ColorNumber `dxf:"62"`  // Color 1
	Color2     color.ColorNumber `dxf:"63"`  // Color 2

	// Face modifier properties
	FaceModifier      int `dxf:"90"` // Face modifier flags
	FaceOpacityLevel  int `dxf:"43"` // Face opacity level
	FaceSpecularLevel int `dxf:"41"` // Face specular level

	// Style settings
	InternalUseOnly bool `dxf:"291"` // Internal use only flag
}

// NewVisualStyle creates a new VISUALSTYLE entity with default values
func NewVisualStyle() *VisualStyle {
	hg := handle.NewHandleGenerator()
	vs := &VisualStyle{
		entity:              NewEntity(VISUALSTYLE),
		handle:              hg.Next(),
		Description:         "",
		StyleType:           VisualStyleTypeRealistic,
		FaceLightingModel:   FaceLightingVisible,
		FaceLightingQuality: FaceLightingPerVertex,
		FaceColorMode:       FaceColorNoColor,

		// Edge properties - disabled by default
		EdgeStyleModel:           EdgeStyleNoEdges,
		EdgeStyle:                0,
		EdgeColorMode:            EdgeColorEntityColor,
		EdgeIntersectionColor:    color.ByLayer,
		EdgeObscuredColor:        color.ByLayer,
		EdgeLinetype:             "CONTINUOUS",
		EdgeObliqueType:          0,
		EdgeCreaseAngle:          0,
		EdgeObliqueSize:          0,
		EdgeModifiers:            EdgeModifierNone,
		EdgeWidth:                1.0,
		EdgeOverhang:             0,
		EdgeIntersectionLinetype: 0,
		EdgeJitter:               0,
		EdgeSilhouetteWidth:      0,
		EdgeSilhouetteColor:      color.ByLayer,
		EdgeIsolineCount:         0,
		EdgeHidePrecision:        false,
		EdgeStyleApply:           true,

		// Display properties
		Brightness: 50.0,
		ShadowType: 0,
		Color1:     color.White,
		Color2:     color.White,

		// Face modifier properties - disabled by default
		FaceModifier:      0,
		FaceOpacityLevel:  50,
		FaceSpecularLevel: 50,

		// Style settings
		InternalUseOnly: false,
	}
	return vs
}

// NewFlatVisualStyle creates a flat visual style
func NewFlatVisualStyle(description string) *VisualStyle {
	vs := NewVisualStyle()
	vs.Description = description
	vs.StyleType = VisualStyleTypeFlat
	return vs
}

// NewWireframeVisualStyle creates a wireframe visual style
func NewWireframeVisualStyle(description string) *VisualStyle {
	vs := NewVisualStyle()
	vs.Description = description
	vs.StyleType = VisualStyleTypeWireframe
	return vs
}

// NewHiddenVisualStyle creates a hidden visual style
func NewHiddenVisualStyle(description string) *VisualStyle {
	vs := NewVisualStyle()
	vs.Description = description
	vs.StyleType = VisualStyleTypeHidden
	return vs
}

// NewRealisticVisualStyle creates a realistic visual style
func NewRealisticVisualStyle(description string) *VisualStyle {
	vs := NewVisualStyle()
	vs.Description = description
	vs.StyleType = VisualStyleTypeRealistic
	return vs
}

// NewConceptualVisualStyle creates a conceptual visual style
func NewConceptualVisualStyle(description string) *VisualStyle {
	vs := NewVisualStyle()
	vs.Description = description
	vs.StyleType = VisualStyleTypeConceptual
	return vs
}

// IsOn returns true if visual style is enabled
func (vs *VisualStyle) IsOn() bool {
	return vs.StyleType != VisualStyleTypeHidden
}

// TurnOn enables the visual style
func (vs *VisualStyle) TurnOn() {
	if vs.StyleType == VisualStyleTypeHidden {
		vs.StyleType = VisualStyleTypeBasic // Enable basic visualization
	}
}

// TurnOff disables the visual style
func (vs *VisualStyle) TurnOff() {
	vs.StyleType = VisualStyleTypeHidden
}

// SetDescription sets the visual style description
func (vs *VisualStyle) SetDescription(description string) {
	vs.Description = description
}

// SetBrightness sets the overall brightness (0-100)
func (vs *VisualStyle) SetBrightness(brightness float64) {
	vs.Brightness = math.Max(0, math.Min(100, brightness))
}

// SetShadowType sets the shadow type
func (vs *VisualStyle) SetShadowType(shadowType int) {
	vs.ShadowType = shadowType
}

// SetColors sets face colors
func (vs *VisualStyle) SetColors(color1, color2 color.ColorNumber) {
	vs.Color1 = color1
	vs.Color2 = color2
}

// SetFaceLighting configures face lighting
func (vs *VisualStyle) SetFaceLighting(model FaceLightingModel, quality FaceLightingQuality) {
	vs.FaceLightingModel = model
	vs.FaceLightingQuality = quality
}

// SetEdgeStyle configures edge rendering
func (vs *VisualStyle) SetEdgeStyle(model EdgeStyleModel, style int, colorMode EdgeColorMode) {
	vs.EdgeStyleModel = model
	vs.EdgeStyle = style
	vs.EdgeColorMode = colorMode
}

// SetEdgeWidth sets edge width
func (vs *VisualStyle) SetEdgeWidth(width float64) {
	vs.EdgeWidth = math.Max(0, width)
}

// SetEdgeOblique sets edge oblique properties
func (vs *VisualStyle) SetEdgeOblique(obliqueType int, size float64) {
	vs.EdgeObliqueType = obliqueType
	vs.EdgeObliqueSize = math.Max(0, size)
}

// SetInternalUseOnly sets internal use flag
func (vs *VisualStyle) SetInternalUseOnly(internalOnly bool) {
	vs.InternalUseOnly = internalOnly
}

// Validate checks if visual style entity has valid properties
func (vs *VisualStyle) Validate() []string {
	errors := make([]string, 0)

	// Check style type
	if vs.StyleType < VisualStyleTypeFlat || vs.StyleType > VisualStyleTypeConceptual {
		errors = append(errors, fmt.Sprintf("Invalid visual style type: %d", vs.StyleType))
	}

	// Check brightness
	if vs.Brightness < 0 || vs.Brightness > 100 {
		errors = append(errors, fmt.Sprintf("Brightness must be between 0 and 100, got: %f", vs.Brightness))
	}

	// Check face lighting
	if vs.FaceLightingModel < FaceLightingInvisible || vs.FaceLightingModel > FaceLightingGooch {
		errors = append(errors, fmt.Sprintf("Invalid face lighting model: %d", vs.FaceLightingModel))
	}

	// Check edge properties
	if vs.EdgeWidth < 0 {
		errors = append(errors, fmt.Sprintf("Edge width must be non-negative, got: %f", vs.EdgeWidth))
	}

	// Check color values
	if vs.Color1 < 0 || vs.Color1 > 255 {
		errors = append(errors, fmt.Sprintf("Color1 must be between 0 and 255, got: %d", vs.Color1))
	}

	if vs.Color2 < 0 || vs.Color2 > 255 {
		errors = append(errors, fmt.Sprintf("Color2 must be between 0 and 255, got: %d", vs.Color2))
	}

	return errors
}

// IsEntity is for Entity interface.
func (vs *VisualStyle) IsEntity() bool {
	return true
}

// Handle returns the visual style's handle
func (vs *VisualStyle) Handle() string {
	return vs.handle
}

// SetHandle sets the visual style's handle
func (vs *VisualStyle) SetHandle(h *handle.HandleGenerator) {
	vs.handle = h.Next()
}

// SetBlockRecord sets block record for visual style
func (vs *VisualStyle) SetBlockRecord(h handle.Handler) {
	vs.entity.SetBlockRecord(h)
}

// Format writes data to formatter.
func (vs *VisualStyle) Format(f format.Formatter) {
	vs.entity.Format(f)
	f.WriteString(100, "AcDbVisualStyle")

	// Write core properties
	f.WriteString(2, vs.Description)
	f.WriteInt(70, int(vs.StyleType))

	// Write face lighting properties
	f.WriteInt(71, int(vs.FaceLightingModel))
	f.WriteInt(72, int(vs.FaceLightingQuality))
	f.WriteInt(73, int(vs.FaceColorMode))

	// Write edge properties
	f.WriteInt(74, int(vs.EdgeStyleModel))
	f.WriteInt(91, vs.EdgeStyle)
	f.WriteInt(66, int(vs.EdgeColorMode))
	f.WriteInt(64, int(vs.EdgeIntersectionColor))
	f.WriteInt(65, int(vs.EdgeObscuredColor))
	f.WriteString(75, vs.EdgeLinetype)
	f.WriteInt(175, vs.EdgeObliqueType)
	f.WriteFloat(42, vs.EdgeCreaseAngle)
	f.WriteFloat(43, vs.EdgeObliqueSize)
	f.WriteInt(92, int(vs.EdgeModifiers))
	f.WriteFloat(40, vs.EdgeWidth)
	f.WriteFloat(44, vs.EdgeOverhang)
	f.WriteInt(77, vs.EdgeIntersectionLinetype)
	f.WriteFloat(78, vs.EdgeJitter)
	f.WriteFloat(79, vs.EdgeSilhouetteWidth)
	f.WriteInt(67, int(vs.EdgeSilhouetteColor))
	f.WriteInt(171, vs.EdgeIsolineCount)
	f.WriteBool(290, vs.EdgeHidePrecision)
	f.WriteBool(174, vs.EdgeStyleApply)

	// Write display properties
	f.WriteFloat(44, vs.Brightness)
	f.WriteInt(173, vs.ShadowType)
	f.WriteInt(62, int(vs.Color1))
	f.WriteInt(63, int(vs.Color2))

	// Write face modifier properties
	f.WriteInt(90, vs.FaceModifier)
	f.WriteInt(43, vs.FaceOpacityLevel)
	f.WriteInt(41, vs.FaceSpecularLevel)

	// Write style settings
	f.WriteBool(291, vs.InternalUseOnly)
}

// Clone creates a copy of the visual style entity
func (vs *VisualStyle) Clone() Entity {
	hg := handle.NewHandleGenerator()
	clone := &VisualStyle{
		entity: NewEntity(VISUALSTYLE),
		handle: hg.Next(),

		// Copy visual style properties
		Description:         vs.Description,
		StyleType:           vs.StyleType,
		FaceLightingModel:   vs.FaceLightingModel,
		FaceLightingQuality: vs.FaceLightingQuality,
		FaceColorMode:       vs.FaceColorMode,

		// Copy edge properties
		EdgeStyleModel:           vs.EdgeStyleModel,
		EdgeStyle:                vs.EdgeStyle,
		EdgeColorMode:            vs.EdgeColorMode,
		EdgeIntersectionColor:    vs.EdgeIntersectionColor,
		EdgeObscuredColor:        vs.EdgeObscuredColor,
		EdgeLinetype:             vs.EdgeLinetype,
		EdgeObliqueType:          vs.EdgeObliqueType,
		EdgeCreaseAngle:          vs.EdgeCreaseAngle,
		EdgeObliqueSize:          vs.EdgeObliqueSize,
		EdgeModifiers:            vs.EdgeModifiers,
		EdgeWidth:                vs.EdgeWidth,
		EdgeOverhang:             vs.EdgeOverhang,
		EdgeIntersectionLinetype: vs.EdgeIntersectionLinetype,
		EdgeJitter:               vs.EdgeJitter,
		EdgeSilhouetteWidth:      vs.EdgeSilhouetteWidth,
		EdgeSilhouetteColor:      vs.EdgeSilhouetteColor,
		EdgeIsolineCount:         vs.EdgeIsolineCount,
		EdgeHidePrecision:        vs.EdgeHidePrecision,
		EdgeStyleApply:           vs.EdgeStyleApply,

		// Copy display properties
		Brightness: vs.Brightness,
		ShadowType: vs.ShadowType,
		Color1:     vs.Color1,
		Color2:     vs.Color2,

		// Copy face modifier properties
		FaceModifier:      vs.FaceModifier,
		FaceOpacityLevel:  vs.FaceOpacityLevel,
		FaceSpecularLevel: vs.FaceSpecularLevel,

		// Copy style settings
		InternalUseOnly: vs.InternalUseOnly,
	}

	// Copy base entity fields
	clone.entity.Type = vs.entity.Type
	clone.entity.handle = vs.entity.handle
	clone.entity.blockRecord = vs.entity.blockRecord
	clone.entity.owner = vs.entity.owner
	clone.entity.layer = vs.entity.layer
	clone.entity.ltscale = vs.entity.ltscale
	clone.entity.color = vs.entity.color

	return clone
}

// String returns string representation of the visual style
func (vs *VisualStyle) String() string {
	return fmt.Sprintf("VISUALSTYLE(%s): %s lighting", vs.Description, vs.StyleType.String())
}

// BBox returns bounding box for visual style (affects entire drawing)
func (vs *VisualStyle) BBox() ([]float64, []float64) {
	// Visual styles affect the entire drawing, so return infinite bounds
	return []float64{-1e10, -1e10, -1e10}, []float64{1e10, 1e10, 1e10}
}
