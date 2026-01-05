package entity

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// Dimension style reference
type DimensionStyle struct {
	Name            string            // Style name
	TextStyle       string            // Text style name
	ArrowBlock1     string            // Arrow block name 1
	ArrowBlock2     string            // Arrow block name 2
	DimLineColor    color.ColorNumber // Dimension line color
	ExtLineColor    color.ColorNumber // Extension line color
	TextColor       color.ColorNumber // Text color
	ArrowSize       float64           // Arrow size
	Scale           float64           // Overall scale
	DecimalPlaces   int               // Decimal places
	ToleranceType   int               // Tolerance type
	ToleranceValue  float64           // Tolerance value
	ToleranceColor  color.ColorNumber // Tolerance color
	TextHeight      float64           // Text height
	TextGap         float64           // Text gap
	ExtensionLength float64           // Extension length
}

// Arc dimension data (for arc length and angular dimensions)
type DimensionArc struct {
	StartAngle float64   // Start angle
	EndAngle   float64   // End angle
	Radius     float64   // Arc radius
	IsCCW      bool      // Counter-clockwise direction
	ArcCenter  math.Vec3 // Arc center point
}

// Ordinate dimension data
type DimensionOrdinate struct {
	Origin          math.Vec3 // Origin point for ordinate
	FeatureType     int       // Feature type (X/Y)
	FeatureLocation math.Vec3 // Feature location
	LeaderLength    float64   // Leader length
}

// Dimension represents a DXF DIMENSION entity (enhanced)
type Dimension struct {
	*entity

	// Basic dimension properties
	dimensionType       int
	definitionPoint     math.Vec3
	textMidpoint        math.Vec3
	text                string
	style               string
	attachmentPoint     int
	lineSpacingStyle    int
	lineSpacingFactor   float64
	actualMeasurement   float64
	obliqueAngle        float64
	obliqueRotation     float64
	horizontalDirection float64
	extrusion           math.Vec3

	// Enhanced dimension properties
	measurementUnits int     // Measurement units
	textMovement     int     // Text movement flags
	toleranceDisplay int     // Tolerance display
	textPosition     int     // Text positioning
	arrowFlags       int     // Arrow and extension flags
	userText         string  // User-defined text
	overrideText     string  // Override text
	toleranceUpper   float64 // Upper tolerance
	toleranceLower   float64 // Lower tolerance
	toleranceHeight  float64 // Tolerance text height
	limitsUpper      float64 // Upper limit
	limitsLower      float64 // Lower limit
	alternateUnits   string  // Alternate units string
	alternateValue   float64 // Alternate measurement value

	// Style and resource references (simplified)
	dimStyle          *DimensionStyle // Dimension style reference
	arrowBlockHandle1 string          // Arrow block handle 1
	arrowBlockHandle2 string          // Arrow block handle 2

	// Extended geometry
	extensionPoint  math.Vec3 // Extension point
	extensionLength float64   // Extension length
	jogPoint        math.Vec3 // Jog point (for special dimensions)
	leaderPoint     math.Vec3 // Leader point (outside dimension)

	// Advanced geometry for different dimension types
	arcData      *DimensionArc      // Arc data for angular/arc length
	ordinateData *DimensionOrdinate // Ordinate data
	basePoint    math.Vec3          // Base point for linear measurements
	featurePoint math.Vec3          // Feature point for ordinate/linear

	// Text and formatting
	textFormat           string    // Text format string
	textWidth            float64   // Text width
	textAlignment        int       // Text alignment override
	textRotationOverride float64   // Text rotation override
	userLocation         math.Vec3 // User-defined text location

	// Dimension-specific properties
	hasLeader         bool // Has leader line
	hasExtension      bool // Has extension line
	suppressExtension bool // Suppress extension line
	suppressArrow1    bool // Suppress first arrow
	suppressArrow2    bool // Suppress second arrow

	// Construction data
	constructionPlane  math.Vec3 // Construction plane normal
	constructionPoint1 math.Vec3 // First construction point
	constructionPoint2 math.Vec3 // Second construction point
	constructionPoint3 math.Vec3 // Third construction point

	// Legacy geometry for compatibility
	angle float64
}

// NewDimension creates a new dimension entity
func NewDimension() *Dimension {
	return &Dimension{
		entity:              NewEntity(DIMENSION),
		dimensionType:       DimensionTypeRotated,
		definitionPoint:     math.NewVec3(0, 0, 0),
		textMidpoint:        math.NewVec3(0, 0, 0),
		text:                "",
		style:               "Standard",
		attachmentPoint:     DimensionAttachmentMiddleCenter,
		lineSpacingStyle:    DimensionSpacingAtLeast,
		lineSpacingFactor:   1.0,
		actualMeasurement:   0,
		obliqueAngle:        0,
		obliqueRotation:     0,
		horizontalDirection: 0,
		extrusion:           math.NewVec3(0, 0, 1),

		// Enhanced dimension properties
		measurementUnits: DimensionUnitsDecimal,
		textMovement:     DimensionTextMoveDefault,
		toleranceDisplay: DimensionToleranceNone,
		textPosition:     DimensionTextPosAbove,
		arrowFlags:       DimensionArrowDefault,
		userText:         "",
		overrideText:     "",
		toleranceUpper:   0.0,
		toleranceLower:   0.0,
		toleranceHeight:  0.0,
		limitsUpper:      0.0,
		limitsLower:      0.0,
		alternateUnits:   "",
		alternateValue:   0.0,

		// Style and resource references
		dimStyle:          nil,
		arrowBlockHandle1: "",
		arrowBlockHandle2: "",

		// Extended geometry
		extensionPoint:  math.NewVec3(0, 0, 0),
		extensionLength: 0.0,
		jogPoint:        math.NewVec3(0, 0, 0),
		leaderPoint:     math.NewVec3(0, 0, 0),

		// Advanced geometry for different dimension types
		arcData:      nil,
		ordinateData: nil,
		basePoint:    math.NewVec3(0, 0, 0),
		featurePoint: math.NewVec3(0, 0, 0),

		// Text and formatting
		textFormat:           "",
		textWidth:            0.0,
		textAlignment:        0,
		textRotationOverride: 0.0,
		userLocation:         math.NewVec3(0, 0, 0),

		// Dimension-specific properties
		hasLeader:         false,
		hasExtension:      false,
		suppressExtension: false,
		suppressArrow1:    false,
		suppressArrow2:    false,

		// Construction data
		constructionPlane:  math.NewVec3(0, 0, 1),
		constructionPoint1: math.NewVec3(0, 0, 0),
		constructionPoint2: math.NewVec3(0, 0, 0),
		constructionPoint3: math.NewVec3(0, 0, 0),

		// Legacy geometry for compatibility
		angle: 0,
	}
}

// IsEntity is for Entity interface.
func (d *Dimension) IsEntity() bool {
	return true
}

// GetDimensionType returns the dimension type
func (d *Dimension) GetDimensionType() int {
	return d.dimensionType
}

// Additional setter methods needed for the example

// SetDimensionType sets the dimension type
func (d *Dimension) SetDimensionType(dimType int) {
	d.dimensionType = dimType
}

// SetDefinitionPoint sets the definition point (10,20,30)
func (d *Dimension) SetDefinitionPoint(point []float64) {
	if len(point) >= 3 {
		d.definitionPoint = math.NewVec3(point[0], point[1], point[2])
	} else if len(point) >= 2 {
		d.definitionPoint = math.NewVec3(point[0], point[1], 0)
	}
}

// SetText sets the dimension text
func (d *Dimension) SetText(text string) {
	d.text = text
}

// SetStyle sets the dimension style name
func (d *Dimension) SetStyle(style string) {
	d.style = style
}

// SetAttachmentPoint sets the text attachment point
func (d *Dimension) SetAttachmentPoint(point int) {
	d.attachmentPoint = point
}

// SetLineSpacingStyle sets the line spacing style
func (d *Dimension) SetLineSpacingStyle(style int) {
	d.lineSpacingStyle = style
}

// SetLineSpacingFactor sets the line spacing factor
func (d *Dimension) SetLineSpacingFactor(factor float64) {
	d.lineSpacingFactor = factor
}

// SetObliqueAngle sets the oblique angle
func (d *Dimension) SetObliqueAngle(angle float64) {
	d.obliqueAngle = angle
}

// SetHorizontalDirection sets the horizontal direction
func (d *Dimension) SetHorizontalDirection(direction float64) {
	d.horizontalDirection = direction
}

// SetExtrusion sets the extrusion direction vector
func (d *Dimension) SetExtrusion(extrusion []float64) {
	if len(extrusion) >= 3 {
		d.extrusion = math.NewVec3(extrusion[0], extrusion[1], extrusion[2])
	} else if len(extrusion) >= 2 {
		d.extrusion = math.NewVec3(extrusion[0], extrusion[1], 0)
	}
}

// SetBasePoint sets the base point for linear measurements
func (d *Dimension) SetBasePoint(point []float64) {
	if len(point) >= 3 {
		d.basePoint = math.NewVec3(point[0], point[1], point[2])
	} else if len(point) >= 2 {
		d.basePoint = math.NewVec3(point[0], point[1], 0)
	}
}

// SetFeaturePoint sets the feature point for measurements
func (d *Dimension) SetFeaturePoint(point []float64) {
	if len(point) >= 3 {
		d.featurePoint = math.NewVec3(point[0], point[1], point[2])
	} else if len(point) >= 2 {
		d.featurePoint = math.NewVec3(point[0], point[1], 0)
	}
}

// SetExtensionPoint sets the extension point for linear dimensions
func (d *Dimension) SetExtensionPoint(point []float64) {
	if len(point) >= 3 {
		d.extensionPoint = math.NewVec3(point[0], point[1], point[2])
	} else if len(point) >= 2 {
		d.extensionPoint = math.NewVec3(point[0], point[1], 0)
	}
}

// SetExtensionLength sets the extension length
func (d *Dimension) SetExtensionLength(length float64) {
	d.extensionLength = length
}

// SetLeaderLength sets the leader length for radial dimensions
func (d *Dimension) SetLeaderLength(length float64) {
	// For radial dimensions, store leader length
	if d.dimensionType == DimensionTypeRadius || d.dimensionType == DimensionTypeDiameter {
		d.extensionLength = length
	}
}

// SetRotation sets the rotation angle for linear dimensions
func (d *Dimension) SetRotation(angle float64) {
	// For rotated dimensions, store the angle
	d.angle = angle
}

// SetToleranceDisplay sets the tolerance display type
func (d *Dimension) SetToleranceDisplay(display int) {
	d.toleranceDisplay = display
}

// SetToleranceUpper sets the upper tolerance value
func (d *Dimension) SetToleranceUpper(upper float64) {
	d.toleranceUpper = upper
}

// SetToleranceLower sets the lower tolerance value
func (d *Dimension) SetToleranceLower(lower float64) {
	d.toleranceLower = lower
}

// SetToleranceHeight sets the tolerance text height
func (d *Dimension) SetToleranceHeight(height float64) {
	d.toleranceHeight = height
}

// SetLimitsUpper sets the upper limit value
func (d *Dimension) SetLimitsUpper(upper float64) {
	d.limitsUpper = upper
}

// SetLimitsLower sets the lower limit value
func (d *Dimension) SetLimitsLower(lower float64) {
	d.limitsLower = lower
}

// Format writes data to formatter.
func (d *Dimension) Format(f format.Formatter) {
	d.entity.Format(f)
	f.WriteString(100, "AcDbDimension")

	// Write definition point (10,20,30)
	f.WriteFloat(10, d.definitionPoint.X())
	f.WriteFloat(20, d.definitionPoint.Y())
	f.WriteFloat(30, d.definitionPoint.Z())

	// Write text midpoint (11,21,31)
	f.WriteFloat(11, d.textMidpoint.X())
	f.WriteFloat(21, d.textMidpoint.Y())
	f.WriteFloat(31, d.textMidpoint.Z())

	// Write dimension type and other properties
	f.WriteInt(70, d.dimensionType)
	if d.text != "" {
		f.WriteString(1, d.text)
	}
	if d.style != "" {
		f.WriteString(3, d.style)
	}
	f.WriteInt(71, d.attachmentPoint)
	f.WriteInt(72, d.lineSpacingStyle)
	f.WriteFloat(41, d.lineSpacingFactor)
	f.WriteFloat(42, d.actualMeasurement)
	f.WriteFloat(52, d.obliqueAngle)
	f.WriteFloat(53, d.obliqueRotation)
	f.WriteFloat(51, d.horizontalDirection)

	// Write extrusion direction (210,220,230)
	f.WriteFloat(210, d.extrusion.X())
	f.WriteFloat(220, d.extrusion.Y())
	f.WriteFloat(230, d.extrusion.Z())

	// Write extension points for linear dimensions
	if d.IsLinear() {
		// First extension line point (13,23,33)
		if !d.suppressExtension {
			f.WriteFloat(13, d.extensionPoint.X())
			f.WriteFloat(23, d.extensionPoint.Y())
			f.WriteFloat(33, d.extensionPoint.Z())
		}

		// Second extension line point (14,24,34) - using same point for simplicity
		if !d.suppressExtension {
			f.WriteFloat(14, d.extensionPoint.X())
			f.WriteFloat(24, d.extensionPoint.Y())
			f.WriteFloat(34, d.extensionPoint.Z())
		}
	}

	// Write enhanced properties for different dimension types
	if d.measurementUnits != DimensionUnitsDecimal {
		f.WriteInt(71, d.measurementUnits)
	}
	if d.textMovement != DimensionTextMoveDefault {
		f.WriteInt(71, d.textMovement)
	}
	if d.toleranceDisplay != DimensionToleranceNone {
		f.WriteInt(71, d.toleranceDisplay)
	}
	if d.textPosition != DimensionTextPosAbove {
		f.WriteInt(71, d.textPosition)
	}
	if d.arrowFlags != DimensionArrowDefault {
		f.WriteInt(71, d.arrowFlags)
	}

	// Write tolerance and limits
	if d.toleranceDisplay != DimensionToleranceNone {
		if d.toleranceUpper != 0.0 || d.toleranceLower != 0.0 {
			f.WriteFloat(41, d.toleranceUpper)
			f.WriteFloat(42, d.toleranceLower)
		}
		if d.toleranceHeight != 0.0 {
			f.WriteFloat(140, d.toleranceHeight)
		}
	}
	if d.limitsUpper != 0.0 || d.limitsLower != 0.0 {
		f.WriteFloat(142, d.limitsUpper)
		f.WriteFloat(143, d.limitsLower)
	}

	// Write alternate units
	if d.alternateUnits != "" {
		f.WriteString(1, d.alternateUnits)
		f.WriteFloat(2, d.alternateValue)
	}

	// Write user text
	if d.userText != "" {
		f.WriteString(1, d.userText)
	}
	if d.overrideText != "" {
		f.WriteString(1, d.overrideText)
	}

	// Write arc dimension data
	if d.IsArcLength() && d.arcData != nil {
		f.WriteFloat(50, d.arcData.StartAngle)
		f.WriteFloat(51, d.arcData.EndAngle)
		f.WriteFloat(40, d.arcData.Radius)
		if d.arcData.IsCCW {
			f.WriteInt(71, 1) // Counter-clockwise flag
		}
		if !d.arcData.ArcCenter.IsEqual(math.NewVec3(0, 0, 0), 1e-9) {
			f.WriteFloat(10, d.arcData.ArcCenter.X())
			f.WriteFloat(20, d.arcData.ArcCenter.Y())
			f.WriteFloat(30, d.arcData.ArcCenter.Z())
		}
	}

	// Write ordinate dimension data
	if d.IsOrdinate() && d.ordinateData != nil {
		f.WriteInt(70, d.ordinateData.FeatureType) // X/Y ordinate type
		// Write origin point
		f.WriteFloat(10, d.ordinateData.Origin.X())
		f.WriteFloat(20, d.ordinateData.Origin.Y())
		f.WriteFloat(30, d.ordinateData.Origin.Z())
		// Write feature location
		f.WriteFloat(11, d.ordinateData.FeatureLocation.X())
		f.WriteFloat(21, d.ordinateData.FeatureLocation.Y())
		f.WriteFloat(31, d.ordinateData.FeatureLocation.Z())
		if d.ordinateData.LeaderLength > 0.0 {
			f.WriteFloat(40, d.ordinateData.LeaderLength)
		}
	}

	// Write construction data
	if !d.constructionPlane.IsEqual(math.NewVec3(0, 0, 1), 1e-9) {
		f.WriteFloat(11, d.constructionPlane.X())
		f.WriteFloat(21, d.constructionPlane.Y())
		f.WriteFloat(31, d.constructionPlane.Z())
	}
	if !d.constructionPoint1.IsEqual(math.NewVec3(0, 0, 0), 1e-9) {
		f.WriteFloat(13, d.constructionPoint1.X())
		f.WriteFloat(23, d.constructionPoint1.Y())
		f.WriteFloat(33, d.constructionPoint1.Z())
	}
	if !d.constructionPoint2.IsEqual(math.NewVec3(0, 0, 0), 1e-9) {
		f.WriteFloat(14, d.constructionPoint2.X())
		f.WriteFloat(24, d.constructionPoint2.Y())
		f.WriteFloat(34, d.constructionPoint2.Z())
	}
	if !d.constructionPoint3.IsEqual(math.NewVec3(0, 0, 0), 1e-9) {
		f.WriteFloat(15, d.constructionPoint3.X())
		f.WriteFloat(25, d.constructionPoint3.Y())
		f.WriteFloat(35, d.constructionPoint3.Z())
	}

	// Write jog point for special dimensions
	if !d.jogPoint.IsEqual(math.NewVec3(0, 0, 0), 1e-9) {
		f.WriteFloat(15, d.jogPoint.X())
		f.WriteFloat(25, d.jogPoint.Y())
		f.WriteFloat(35, d.jogPoint.Z())
	}

	// Write leader point (outside dimension)
	if d.hasLeader && !d.leaderPoint.IsEqual(math.NewVec3(0, 0, 0), 1e-9) {
		f.WriteFloat(10, d.leaderPoint.X())
		f.WriteFloat(20, d.leaderPoint.Y())
		f.WriteFloat(30, d.leaderPoint.Z())
	}

	// Write extension length
	if d.extensionLength != 0.0 {
		f.WriteFloat(42, d.extensionLength)
	}
}

// BBox returns bounding box
func (d *Dimension) BBox() ([]float64, []float64) {
	// Simple bounding box using definition and extension points
	mins := []float64{d.definitionPoint.X(), d.definitionPoint.Y(), d.definitionPoint.Z()}
	maxs := []float64{d.definitionPoint.X(), d.definitionPoint.Y(), d.definitionPoint.Z()}

	// Include text midpoint in bounding box
	if d.textMidpoint.X() < mins[0] {
		mins[0] = d.textMidpoint.X()
	}
	if d.textMidpoint.Y() < mins[1] {
		mins[1] = d.textMidpoint.Y()
	}
	if d.textMidpoint.Z() < mins[2] {
		mins[2] = d.textMidpoint.Z()
	}

	if d.textMidpoint.X() > maxs[0] {
		maxs[0] = d.textMidpoint.X()
	}
	if d.textMidpoint.Y() > maxs[1] {
		maxs[1] = d.textMidpoint.Y()
	}
	if d.textMidpoint.Z() > maxs[2] {
		maxs[2] = d.textMidpoint.Z()
	}

	return mins, maxs
}

// IsLinear returns true for linear dimension types
func (d *Dimension) IsLinear() bool {
	return d.dimensionType == DimensionTypeRotated ||
		d.dimensionType == DimensionTypeAligned ||
		d.dimensionType == DimensionTypeDiameter ||
		d.dimensionType == DimensionTypeRadius ||
		d.dimensionType == DimensionTypeLargeRadial
}

// IsAngular returns true for angular dimension types
func (d *Dimension) IsAngular() bool {
	return d.dimensionType == DimensionTypeAngular ||
		d.dimensionType == DimensionTypeAngular3P
}

// IsArcLength returns true for arc length dimension
func (d *Dimension) IsArcLength() bool {
	return d.dimensionType == DimensionTypeArc
}

// IsOrdinate returns true for ordinate dimension
func (d *Dimension) IsOrdinate() bool {
	return d.dimensionType == DimensionTypeOrdinate
}

// SetArcData sets arc dimension properties
func (d *Dimension) SetArcData(startAngle, endAngle, radius float64, isCCW bool) {
	d.arcData = &DimensionArc{
		StartAngle: startAngle,
		EndAngle:   endAngle,
		Radius:     radius,
		IsCCW:      isCCW,
	}
	d.dimensionType = DimensionTypeArc
}

// SetOrdinateData sets ordinate dimension properties
func (d *Dimension) SetOrdinateData(origin math.Vec3, featureType int, featureLocation math.Vec3, leaderLength float64) {
	d.ordinateData = &DimensionOrdinate{
		Origin:          origin,
		FeatureType:     featureType,
		FeatureLocation: featureLocation,
		LeaderLength:    leaderLength,
	}
	d.dimensionType = DimensionTypeOrdinate
}

// SetTolerance sets tolerance values and display
func (d *Dimension) SetTolerance(display int, upper, lower float64, toleranceColor color.ColorNumber) {
	d.toleranceDisplay = display
	d.toleranceUpper = upper
	d.toleranceLower = lower
	if toleranceColor != 0 {
		d.toleranceHeight = 2.5 // Default tolerance height
	}
}

// SetLimits sets limit values
func (d *Dimension) SetLimits(upper, lower float64) {
	d.limitsUpper = upper
	d.limitsLower = lower
}

// SetAlternateUnits sets alternate unit display
func (d *Dimension) SetAlternateUnits(units string, value float64) {
	d.alternateUnits = units
	d.alternateValue = value
}

// SetArrowConfig sets arrow and extension flags
func (d *Dimension) SetArrowConfig(suppressArrow1, suppressArrow2, suppressExt bool) {
	d.suppressArrow1 = suppressArrow1
	d.suppressArrow2 = suppressArrow2
	d.suppressExtension = suppressExt
}

// SetTextPosition sets text positioning and movement
func (d *Dimension) SetTextPosition(position int, movement int, userLocation math.Vec3) {
	d.textPosition = position
	d.textMovement = movement
	d.userLocation = userLocation
}

// SetConstructionData sets construction geometry for complex dimensions
func (d *Dimension) SetConstructionData(plane math.Vec3, p1, p2, p3 math.Vec3) {
	d.constructionPlane = plane
	d.constructionPoint1 = p1
	d.constructionPoint2 = p2
	d.constructionPoint3 = p3
}

// SetStyleFromStruct sets dimension style reference
func (d *Dimension) SetStyleFromStruct(style *DimensionStyle) {
	d.dimStyle = style
	if style != nil {
		d.style = style.Name
	}
}

// SetExtensionGeometry sets extension line properties
func (d *Dimension) SetExtensionGeometry(extensionLength float64, extensionPoint math.Vec3) {
	d.extensionLength = extensionLength
	d.extensionPoint = extensionPoint
}

// SetUserText sets custom text values
func (d *Dimension) SetUserText(userText, overrideText string) {
	d.userText = userText
	d.overrideText = overrideText
}

// SetLeader enables leader line
func (d *Dimension) SetLeader(hasLeader bool, leaderPoint math.Vec3) {
	d.hasLeader = hasLeader
	d.leaderPoint = leaderPoint
}

// SetExtension enables extension line
func (d *Dimension) SetExtension(hasExtension bool) {
	d.hasExtension = hasExtension
}

// SetUserLocation sets user-defined text location
func (d *Dimension) SetUserLocation(location math.Vec3) {
	d.userLocation = location
}
