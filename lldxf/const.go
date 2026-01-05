package lldxf

import "fmt"

// DXF section names
const (
	SectionHeader    = "HEADER"
	SectionClasses   = "CLASSES"
	SectionTables    = "TABLES"
	SectionBlocks    = "BLOCKS"
	SectionEntities  = "ENTITIES"
	SectionObjects   = "OBJECTS"
	SectionThumbnail = "THUMBNAILIMAGE"
	SectionAcDsData  = "ACDSDATA"
)

// Common group codes
const (
	GroupCodeEntityName     = 0    // Entity type
	GroupCodeHandle         = 5    // Handle
	GroupCodeText           = 1    // Text value
	GroupCodeXCoord         = 10   // Primary X coordinate
	GroupCodeYCoord         = 20   // Primary Y coordinate
	GroupCodeZCoord         = 30   // Primary Z coordinate
	GroupCodeSecondaryX     = 11   // Secondary X coordinate
	GroupCodeSecondaryY     = 21   // Secondary Y coordinate
	GroupCodeSecondaryZ     = 31   // Secondary Z coordinate
	GroupCodeTertiaryX      = 12   // Tertiary X coordinate
	GroupCodeTertiaryY      = 22   // Tertiary Y coordinate
	GroupCodeTertiaryZ      = 32   // Tertiary Z coordinate
	GroupCodeColor          = 62   // Color number
	GroupCodeLayerName      = 8    // Layer name
	GroupCodeLinetypeName   = 6    // Linetype name
	GroupCodeTextStyle      = 7    // Text style name
	GroupCodeSubclassMarker = 100  // Subclass data marker
	GroupCodeAppID          = 102  // Application-defined data
	GroupCodeExtendedData   = 1001 // Extended data application name
	GroupCodeBinaryData     = 310  // Binary data
)

// Entity type names
const (
	EntityTypeLine            = "LINE"
	EntityTypePoint           = "POINT"
	EntityTypeCircle          = "CIRCLE"
	EntityTypeArc             = "ARC"
	EntityTypeEllipse         = "ELLIPSE"
	EntityTypePolyline        = "POLYLINE"
	EntityTypeLWPolyline      = "LWPOLYLINE"
	EntityTypeVertex          = "VERTEX"
	EntityTypeSeqEnd          = "SEQEND"
	EntityTypeText            = "TEXT"
	EntityTypeMText           = "MTEXT"
	EntityTypeInsert          = "INSERT"
	EntityTypeBlock           = "BLOCK"
	EntityTypeEndBlk          = "ENDBLK"
	EntityTypeSolid           = "SOLID"
	EntityTypeTrace           = "TRACE"
	EntityType3DFace          = "3DFACE"
	EntityTypeSpline          = "SPLINE"
	EntityTypeHatch           = "HATCH"
	EntityTypeDimension       = "DIMENSION"
	EntityTypeLeader          = "LEADER"
	EntityTypeMLine           = "MLINE"
	EntityTypeXLine           = "XLINE"
	EntityTypeRay             = "RAY"
	EntityTypeMesh            = "MESH"
	EntityTypeBody            = "BODY"
	EntityTypeRegion          = "REGION"
	EntityTypeSolid3D         = "3DSOLID"
	EntityTypeSurface         = "SURFACE"
	EntityTypeExtrudedSurface = "EXTRUDEDSURFACE"
	EntityTypeLoftedSurface   = "LOFTEDSURFACE"
	EntityTypeRevolvedSurface = "REVOLVEDSURFACE"
	EntityTypeSweptSurface    = "SWEPTSURFACE"
	EntityTypePlaneSurface    = "PLANESURFACE"
)

// Table type names
const (
	TableTypeVPort       = "VPORT"
	TableTypeLType       = "LTYPE"
	TableTypeLayer       = "LAYER"
	TableTypeStyle       = "STYLE"
	TableTypeView        = "VIEW"
	TableTypeUCS         = "UCS"
	TableTypeAppID       = "APPID"
	TableTypeDimStyle    = "DIMSTYLE"
	TableTypeBlockRecord = "BLOCK_RECORD"
)

// Standard colors
const (
	ColorByLayer   = 256
	ColorByBlock   = 0
	ColorRed       = 1
	ColorYellow    = 2
	ColorGreen     = 3
	ColorCyan      = 4
	ColorBlue      = 5
	ColorMagenta   = 6
	ColorWhite     = 7
	ColorDarkGray  = 8
	ColorLightGray = 9
)

// DXF versions
const (
	VersionR12   = "AC1009"
	VersionR14   = "AC1014"
	VersionR2000 = "AC1015"
	VersionR2004 = "AC1018"
	VersionR2007 = "AC1021"
	VersionR2010 = "AC1024"
	VersionR2013 = "AC1027"
	VersionR2018 = "AC1032"
)

// Error definitions
var (
	ErrInvalidGroupCode   = fmt.Errorf("invalid DXF group code")
	ErrInvalidValue       = fmt.Errorf("invalid DXF tag value")
	ErrMissingRequiredTag = fmt.Errorf("missing required DXF tag")
	ErrInvalidEntity      = fmt.Errorf("invalid DXF entity")
	ErrInvalidSection     = fmt.Errorf("invalid DXF section")
	ErrParseError         = fmt.Errorf("DXF parse error")
	ErrWriteError         = fmt.Errorf("DXF write error")
)

// Standard linetype names
const (
	LtContinuous = "Continuous"
	LtHidden     = "Hidden"
	LtCenter     = "Center"
	LtPhantom    = "Phantom"
	LtDot        = "Dot"
	LtDash       = "Dash"
	LtDashDot    = "DashDot"
	LtDashDotDot = "DashDotDot"
	LtDivide     = "Divide"
	LtCenterLine = "CenterLine"
	LtBorder     = "Border"
	LtZigZag     = "Zigzag"
)

// Text alignment flags
const (
	TextAlignmentLeft         = 0
	TextAlignmentCenter       = 1
	TextAlignmentRight        = 2
	TextAlignmentAligned      = 3
	TextAlignmentMiddle       = 4
	TextAlignmentFit          = 5
	TextAlignmentTopLeft      = 10
	TextAlignmentTopCenter    = 11
	TextAlignmentTopRight     = 12
	TextAlignmentMiddleLeft   = 13
	TextAlignmentMiddleCenter = 14
	TextAlignmentMiddleRight  = 15
	TextAlignmentBottomLeft   = 16
	TextAlignmentBottomCenter = 17
	TextAlignmentBottomRight  = 18
)

// MText attachment points
const (
	MTextAttachmentTopLeft      = 1
	MTextAttachmentTopCenter    = 2
	MTextAttachmentTopRight     = 3
	MTextAttachmentMiddleLeft   = 4
	MTextAttachmentMiddleCenter = 5
	MTextAttachmentMiddleRight  = 6
	MTextAttachmentBottomLeft   = 7
	MTextAttachmentBottomCenter = 8
	MTextAttachmentBottomRight  = 9
)

// Dimension types
const (
	DimTypeRotated       = 0
	DimTypeAligned       = 1
	DimTypeAngular       = 2
	DimTypeDiameter      = 3
	DimTypeRadius        = 4
	DimTypeAngular3Point = 5
	DimTypeOrdinate      = 6
)

// Hatch pattern types
const (
	HatchPatternUserDefined = 0
	HatchPatternPreDefined  = 1
	HatchPatternCustom      = 2
)

// Spline flags
const (
	SplineClosed   = 1
	SplinePeriodic = 2
	SplineRational = 4
	SplinePlanar   = 8
	SplineLinear   = 16
)

// Polyline flags
const (
	PolylineClosed      = 1
	PolylineCurveFit    = 2
	PolylineSplineFit   = 4
	Polyline3DPolyline  = 8
	Polyline3DMesh      = 16
	PolylineMeshClosedM = 32
	PolylineMeshClosedN = 64
	PolylinePolyface    = 128
	PolylineGenerated   = 256
)

// Layer flags
const (
	LayerFrozen    = 1
	LayerFrozenNew = 1 // Alias for compatibility
	LayerLocked    = 4
	LayerPlot      = 64 // 0 = do not plot
)
