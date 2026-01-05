package entity

// Dimension types (from group code 70)
const (
	DimensionTypeRotated     = 0 // Linear dimension (rotated)
	DimensionTypeAligned     = 1 // Aligned linear dimension
	DimensionTypeAngular     = 2 // Angular dimension (2-point)
	DimensionTypeDiameter    = 3 // Diameter dimension
	DimensionTypeRadius      = 4 // Radius dimension
	DimensionTypeAngular3P   = 5 // Angular dimension (3-point)
	DimensionTypeOrdinate    = 6 // Ordinate dimension
	DimensionTypeArc         = 8 // Arc length dimension (DXF R2018+)
	DimensionTypeLargeRadial = 7 // Large radial dimension
)

// Measurement units (from group code 71)
const (
	DimensionUnitsDecimal = 0 // Decimal units
	DimensionUnitsDegrees = 1 // Degrees
	DimensionUnitsDMS     = 2 // Degrees/Minutes/Seconds
	DimensionUnitsGrad    = 3 // Gradians
	DimensionUnitsRadians = 4 // Radians
)

// Text movement flags (from group code 71)
const (
	DimensionTextMoveDefault   = 0 // Default movement
	DimensionTextMoveAddLeader = 1 // Add leader
	DimensionTextMoveNoText    = 2 // No text
)

// Tolerance display (from group code 71)
const (
	DimensionToleranceNone   = 0 // No tolerance
	DimensionToleranceBasic  = 1 // Basic tolerance
	DimensionToleranceLimits = 2 // Limits
)

// Text positioning flags (from group code 71)
const (
	DimensionTextPosAbove  = 0 // Text above dimension line
	DimensionTextPosCenter = 1 // Text in center
	DimensionTextPosBelow  = 2 // Text below dimension line
)

// Arrow and extension flags (from group code 71)
const (
	DimensionArrowDefault  = 0 // Default arrow placement
	DimensionArrowUser     = 1 // User-defined
	DimensionArrowSuppress = 2 // Suppress first arrow
)

// Text alignment attachment points (from group code 71)
const (
	DimensionAttachmentTopLeft      = 1
	DimensionAttachmentTopCenter    = 2
	DimensionAttachmentTopRight     = 3
	DimensionAttachmentMiddleLeft   = 4
	DimensionAttachmentMiddleCenter = 5
	DimensionAttachmentMiddleRight  = 6
	DimensionAttachmentBottomLeft   = 7
	DimensionAttachmentBottomCenter = 8
	DimensionAttachmentBottomRight  = 9
)

// Line spacing styles (from group code 72)
const (
	DimensionSpacingAtLeast = 1 // Use at least spacing
	DimensionSpacingExact   = 2 // Use exact spacing
)
