package entity

import (
	"fmt"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// MLeader context types
const (
	MLeaderContextTypeNone            = 0
	MLeaderContextTypeLeft            = 1
	MLeaderContextTypeRight           = 2
	MLeaderContextTypeAngle           = 3
	MLeaderContextTypeFirstLineAngle  = 4
	MLeaderContextTypeSecondLineAngle = 5
	MLeaderContextTypeDiameter        = 6
	MLeaderContextTypeSquare          = 7
)

// MLeader attachment types
const (
	MLeaderAttachmentTopLeft      = 1
	MLeaderAttachmentTopCenter    = 2
	MLeaderAttachmentTopRight     = 3
	MLeaderAttachmentMiddleLeft   = 4
	MLeaderAttachmentMiddleCenter = 5
	MLeaderAttachmentMiddleRight  = 6
	MLeaderAttachmentBottomLeft   = 7
	MLeaderAttachmentBottomCenter = 8
	MLeaderAttachmentBottomRight  = 9
)

// MLeader line types
const (
	MLeaderLineTypeStraight = 0
	MLeaderLineTypeSpline   = 1
	MLeaderLineTypeHidden   = 2
	MLeaderLineTypeHeart    = 3
)

// MLeader arrowhead types
const (
	MLeaderArrowheadNone        = 0
	MLeaderArrowheadClosed      = 1
	MLeaderArrowheadDot         = 2
	MLeaderArrowheadArch        = 3
	MLeaderArrowheadTick        = 4
	MLeaderArrowheadOpen        = 5
	MLeaderArrowheadOpen90      = 6
	MLeaderArrowheadOpen180     = 7
	MLeaderArrowheadOrigin      = 8
	MLeaderArrowheadDotSmall    = 9
	MLeaderArrowheadDotBlank    = 10
	MLeaderArrowheadSmall       = 11
	MLeaderArrowheadBox         = 12
	MLeaderArrowheadBoxBlank    = 13
	MLeaderArrowheadClosedBlank = 14
	MLeaderArrowheadTriangle    = 15
	MLeaderArrowheadBoxFilled   = 16
	MLeaderArrowheadDiamond     = 17
	MLeaderArrowheadOblique     = 18
)

// Leader line vertex with bulge support
type LeaderVertex struct {
	Point []float64 // 3D point coordinates
	Bulge float64   // Bulge for arc segments (0=straight line)
}

// Leader line data (enhanced)
type LeaderLine struct {
	Type         int               // Line type (straight, spline, etc.)
	Vertices     []LeaderVertex    // Leader line vertices with bulge support
	Color        color.ColorNumber // Line color number
	Weight       float64           // Line weight
	Visible      bool              // Line visibility
	HasBreak     bool              // Has break in line
	BreakPoints  [][]float64       // Break point locations
	HasDogleg    bool              // Has dogleg
	DoglegLength float64           // Dogleg length
	DoglegAngle  float64           // Dogleg angle
}

// Leader context data
type LeaderContext struct {
	Type         int     // Context type
	Data         float64 // Context data
	DefaultValue float64 // Default value
}

// Connection box for attachment points
type ConnectionBox struct {
	Left, Top, Right, Bottom float64
	HasLeftUnderline         bool
	HasRightUnderline        bool
	HasTopOverline           bool
	HasBottomUnderline       bool
}

// MText data for enhanced content
type MTextData struct {
	Content              string            // Text content
	Style                string            // Text style name
	Color                color.ColorNumber // Text color number
	CharHeight           float64           // Character height
	Rotation             float64           // Text rotation angle
	Width                float64           // Text width factor
	Alignment            int               // Text alignment
	LineSpacingFactor    float64           // Line spacing factor
	LineSpacingStyle     int               // Line spacing style
	AttachmentPoint      int               // Attachment point
	BackgroundColor      color.ColorNumber // Background color number
	BackgroundColorScale float64           // Background color scale factor
	UseBackgroundColor   bool              // Use background color
	FillColor            color.ColorNumber // Fill color number
	FillColorScale       float64           // Fill color scale factor
	UseFillColor         bool              // Use fill color
	ColumnType           int               // Column type
	UseAutoHeight        bool              // Use automatic height
	ColumnWidth          float64           // Column width
	GutterWidth          float64           // Gutter width
	FlowDirection        int               // Text flow direction
	ConnectionBox        ConnectionBox     // Connection box for attachment
}

// Block data for block content
type BlockData struct {
	Name          string        // Block name
	Transform     math.Matrix44 // Transformation matrix
	Scale         float64       // Block scale
	Rotation      float64       // Block rotation
	XDirection    math.Vec3     // X direction vector
	YDirection    math.Vec3     // Y direction vector
	HasAttribs    bool          // Has attributes
	AttribHandles []string      // Attribute handles
}

// Leader data for individual leader lines
type LeaderData struct {
	LeaderLines     []LeaderLine // Individual leader lines
	HasBreak        bool         // Has break in leader
	BreakPoints     [][]float64  // Break point locations
	HasDogleg       bool         // Has dogleg
	DoglegLength    float64      // Dogleg length
	DoglegAngle     float64      // Dogleg angle
	Index           int          // Leader index
	RootPoint       []float64    // Root point for this leader
	LastLeaderPoint []float64    // Last leader point before text
}

// MLeader context data (enhanced)
type MLeaderContext struct {
	Leaders         []LeaderData // Leader data collection
	Scale           float64      // Context scale
	BasePoint       math.Vec3    // Base point
	CharHeight      float64      // Character height
	ArrowHeadSize   float64      // Arrow head size
	LandingGapSize  float64      // Landing gap size
	LeftAttachment  int          // Left attachment type
	RightAttachment int          // Right attachment type
	TextAlignType   int          // Text alignment type
	AttachmentType  int          // Overall attachment type
	Mtext           *MTextData   // MTEXT content data
	Block           *BlockData   // Block content data
	PlaneOrigin     math.Vec3    // Plane origin
	PlaneXAxis      math.Vec3    // Plane X-axis
	PlaneYAxis      math.Vec3    // Plane Y-axis
	PlaneNormal     math.Vec3    // Plane normal vector
}

// MLeader text content (legacy compatibility)
type MLeaderText struct {
	Content              string            // Text content
	Style                string            // Text style name
	Color                color.ColorNumber // Text color number
	Height               float64           // Text height
	Rotation             float64           // Text rotation angle
	Width                float64           // Text width factor
	Alignment            int               // Text alignment
	LineSpacingFactor    float64           // Line spacing factor
	LineSpacingStyle     int               // Line spacing style
	AttachmentPoint      int               // Attachment point
	BackgroundColor      color.ColorNumber // Background color number
	BackgroundColorScale float64           // Background color scale factor
	UseBackgroundColor   bool              // Use background color
	FillColor            color.ColorNumber // Fill color number
	FillColorScale       float64           // Fill color scale factor
	UseFillColor         bool              // Use fill color
	ColumnType           int               // Column type
	UseAutoHeight        bool              // Use automatic height
	ColumnWidth          float64           // Column width
	GutterWidth          float64           // Gutter width
	FlowDirection        int               // Text flow direction
}

// MLeader represents MULTILEADER Entity (enhanced)
type MLeader struct {
	*entity

	// Basic properties
	Version         int     // MLEADER version number
	StyleHandle     string  // Style handle string
	AnnotationScale float64 // Annotation scale
	ContextType     int     // Context type
	BitFlags        int     // Bit flags

	// Leader lines
	Lines              []LeaderLine // Leader line segments
	HasLanding         bool         // Has landing line
	LandingDistance    float64      // Landing distance
	LandingGap         float64      // Landing gap
	HookLineDirection  float64      // Hook line direction
	HookLineLength     float64      // Hook line length
	ArrowheadDirection float64      // Arrowhead direction

	// Arrowhead data
	ArrowheadEnabled bool    // Arrowhead enabled
	ArrowheadSize    float64 // Arrowhead size
	ArrowheadType    int     // Arrowhead type

	// Text content
	ContentType     int         // Content type (0=none, 1=mtext, 2=block, 3=tolerance)
	Text            MLeaderText // Text data (legacy)
	MText           *MTextData  // Enhanced MTEXT data
	Block           *BlockData  // Enhanced block data
	BlockName       string      // Block name for block content (legacy)
	ToleranceType   int         // Tolerance type for tolerance content
	ToleranceValue1 float64     // First tolerance value
	ToleranceValue2 float64     // Second tolerance value

	// Attachment data
	LeaderDirection int  // Leader direction (horizontal, vertical)
	LeaderLineType  int  // Leader line type
	HasLeaderLine   bool // Has leader line
	HasDogleg       bool // Has dogleg

	// Scale and positioning
	Scale                float64 // Overall scale factor
	HorizontalAttachment int     // Horizontal attachment point
	VerticalAttachment   int     // Vertical attachment point

	// Color and visibility
	Color          color.ColorNumber // Main color number
	TextColor      color.ColorNumber // Text color number
	ArrowheadColor color.ColorNumber // Arrowhead color number

	// Contextual data
	Context            LeaderContext   // Context data (legacy)
	EnhancedContext    *MLeaderContext // Enhanced context data
	DefaultTextContent string          // Default text content

	// Enhanced features
	HasStyleOverrides bool // Has style overrides
	OverrideFlags     int  // Override flag bitmask
}

// NewMLeader creates a new MULTILEADER entity
func NewMLeader() *MLeader {
	return &MLeader{
		entity:             NewEntity(MLEADER),
		Version:            2, // Latest MLEADER version
		StyleHandle:        "",
		AnnotationScale:    1.0,
		ContextType:        MLeaderContextTypeNone,
		BitFlags:           0,
		Lines:              make([]LeaderLine, 0),
		HasLanding:         false,
		LandingDistance:    0.0,
		LandingGap:         0.0,
		HookLineDirection:  0.0,
		HookLineLength:     0.0,
		ArrowheadDirection: 0.0,
		ArrowheadEnabled:   true,
		ArrowheadSize:      1.0,
		ArrowheadType:      MLeaderArrowheadClosed,
		ContentType:        1, // MTEXT by default
		Text: MLeaderText{
			Content:              "",
			Style:                "Standard",
			Color:                0,
			Height:               1.0,
			Rotation:             0.0,
			Width:                1.0,
			Alignment:            MLeaderAttachmentMiddleCenter,
			LineSpacingFactor:    1.0,
			LineSpacingStyle:     1, // At least
			AttachmentPoint:      MLeaderAttachmentMiddleCenter,
			BackgroundColor:      0,
			BackgroundColorScale: 1.0,
			UseBackgroundColor:   false,
			FillColor:            0,
			FillColorScale:       1.0,
			UseFillColor:         false,
			ColumnType:           0,
			UseAutoHeight:        true,
			ColumnWidth:          0.0,
			GutterWidth:          0.0,
			FlowDirection:        0,
		},
		BlockName:            "",
		ToleranceType:        0,
		ToleranceValue1:      0.0,
		ToleranceValue2:      0.0,
		LeaderDirection:      0, // Horizontal
		LeaderLineType:       MLeaderLineTypeStraight,
		HasLeaderLine:        true,
		HasDogleg:            false,
		Scale:                1.0,
		HorizontalAttachment: MLeaderAttachmentMiddleCenter,
		VerticalAttachment:   MLeaderAttachmentMiddleCenter,
		Color:                0,
		TextColor:            0,
		ArrowheadColor:       0,
		Context: LeaderContext{
			Type:         MLeaderContextTypeNone,
			Data:         0.0,
			DefaultValue: 0.0,
		},
		DefaultTextContent: "",
	}
}

// IsEntity is for Entity interface
func (m *MLeader) IsEntity() bool {
	return true
}

// SetLeaderLine adds a leader line segment
func (m *MLeader) SetLeaderLine(points [][]float64, lineType int, lineColor color.ColorNumber, weight float64) {
	vertices := make([]LeaderVertex, len(points))
	for i, point := range points {
		vertices[i] = LeaderVertex{
			Point: point,
			Bulge: 0.0, // Default to straight line
		}
	}
	line := LeaderLine{
		Type:      lineType,
		Vertices:  vertices,
		Color:     lineColor,
		Weight:    weight,
		Visible:   true,
		HasBreak:  false,
		HasDogleg: false,
	}
	m.Lines = append(m.Lines, line)
	m.HasLeaderLine = true
}

// SetLanding sets landing properties
func (m *MLeader) SetLanding(distance, gap float64, hasLanding bool) {
	m.HasLanding = hasLanding
	m.LandingDistance = distance
	m.LandingGap = gap
}

// SetArrowhead sets arrowhead properties
func (m *MLeader) SetArrowhead(arrowheadType int, size float64, arrowheadColor color.ColorNumber) {
	m.ArrowheadType = arrowheadType
	m.ArrowheadSize = size
	m.ArrowheadColor = arrowheadColor
	m.ArrowheadEnabled = true
}

// SetTextContent sets text content properties
func (m *MLeader) SetTextContent(content string, style string, height float64, textColor color.ColorNumber) {
	m.ContentType = 1 // MTEXT
	m.Text.Content = content
	m.Text.Style = style
	m.Text.Height = height
	m.Text.Color = textColor
}

// SetBlockContent sets block content
func (m *MLeader) SetBlockContent(blockName string) {
	m.ContentType = 2 // BLOCK
	m.BlockName = blockName
}

// SetToleranceContent sets tolerance content
func (m *MLeader) SetToleranceContent(toleranceType int, value1, value2 float64) {
	m.ContentType = 3 // TOLERANCE
	m.ToleranceType = toleranceType
	m.ToleranceValue1 = value1
	m.ToleranceValue2 = value2
}

// SetAttachment sets attachment points
func (m *MLeader) SetAttachment(horizontal, vertical int) {
	m.HorizontalAttachment = horizontal
	m.VerticalAttachment = vertical
	m.Text.AttachmentPoint = horizontal
}

// SetContext sets context data
func (m *MLeader) SetContext(contextType int, data float64) {
	m.Context = LeaderContext{
		Type:         contextType,
		Data:         data,
		DefaultValue: 0.0,
	}
	m.ContextType = contextType
}

// SetScale sets scale factors
func (m *MLeader) SetScale(scale, annotationScale float64) {
	m.Scale = scale
	m.AnnotationScale = annotationScale
}

// SetDogleg enables dogleg leader
func (m *MLeader) SetDogleg(hasDogleg bool, hookDirection, hookLength float64) {
	m.HasDogleg = hasDogleg
	m.HookLineDirection = hookDirection
	m.HookLineLength = hookLength
}

// SetColor sets the main color
func (m *MLeader) SetColor(mainColor color.ColorNumber) {
	m.Color = mainColor
}

// SetTextColor sets the text color
func (m *MLeader) SetTextColor(textColor color.ColorNumber) {
	m.Text.Color = textColor
}

// SetStyleHandle sets the style handle
func (m *MLeader) SetStyleHandle(handle string) {
	m.StyleHandle = handle
}

// SetBitFlags sets various option flags
func (m *MLeader) SetBitFlags(flags int) {
	m.BitFlags = flags
}

// Format writes data to formatter
func (m *MLeader) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbMLeader")

	// Basic properties
	f.WriteInt(170, m.Version)
	if m.StyleHandle != "" {
		f.WriteString(3, m.StyleHandle)
	}
	f.WriteFloat(140, m.AnnotationScale)
	f.WriteInt(171, m.ContextType)
	f.WriteInt(172, m.BitFlags)
	f.WriteFloat(141, m.Scale)

	// Arrowhead properties
	if m.ArrowheadEnabled {
		f.WriteInt(290, 1) // Arrowhead enabled
	} else {
		f.WriteInt(290, 0)
	}
	f.WriteInt(340, m.ArrowheadType)
	f.WriteFloat(141, m.ArrowheadSize)
	if m.ArrowheadColor != 0 {
		f.WriteInt(91, int(m.ArrowheadColor))
	}

	// Leader line properties
	if m.HasLeaderLine {
		f.WriteInt(291, 1) // Leader line enabled
	} else {
		f.WriteInt(291, 0)
	}
	f.WriteInt(341, m.LeaderLineType)
	f.WriteFloat(92, m.ArrowheadDirection)

	// Landing properties
	if m.HasLanding {
		f.WriteInt(292, 1) // Landing enabled
	} else {
		f.WriteInt(292, 0)
	}
	f.WriteFloat(141, m.LandingDistance)
	f.WriteFloat(142, m.LandingGap)

	// Hook line properties
	if m.HasDogleg {
		f.WriteInt(293, 1) // Dogleg enabled
	} else {
		f.WriteInt(293, 0)
	}
	f.WriteInt(11, int(m.HookLineDirection))
	f.WriteFloat(40, m.HookLineLength)

	// Content type
	f.WriteInt(172, m.ContentType)

	// Write content based on type
	switch m.ContentType {
	case 1: // MTEXT
		m.writeMTextContent(f)
	case 2: // BLOCK
		m.writeBlockContent(f)
	case 3: // TOLERANCE
		m.writeToleranceContent(f)
	default:
		// No content
	}

	// Attachment points
	f.WriteInt(173, m.HorizontalAttachment)
	f.WriteInt(174, m.VerticalAttachment)

	// Colors
	if m.Color != 0 {
		f.WriteInt(92, int(m.Color))
	}
	if m.Text.Color != 0 {
		f.WriteInt(94, int(m.Text.Color))
	}
	if m.ArrowheadColor != 0 {
		f.WriteInt(91, int(m.ArrowheadColor))
	}
}

// writeMTextContent writes MTEXT content data
func (m *MLeader) writeMTextContent(f format.Formatter) {
	f.WriteString(304, m.Text.Content)
	if m.Text.Style != "" {
		f.WriteString(3, m.Text.Style)
	}
	if m.Text.Color != 0 {
		f.WriteInt(90, int(m.Text.Color))
	}
	if m.Text.Height != 1.0 {
		f.WriteFloat(40, m.Text.Height)
	}
	if m.Text.Rotation != 0.0 {
		f.WriteFloat(50, m.Text.Rotation)
	}
	if m.Text.Width != 1.0 {
		f.WriteFloat(41, m.Text.Width)
	}
	if m.Text.Alignment != MLeaderAttachmentMiddleCenter {
		f.WriteInt(71, m.Text.Alignment)
	}
	if m.Text.LineSpacingFactor != 1.0 {
		f.WriteFloat(42, m.Text.LineSpacingFactor)
	}
	if m.Text.LineSpacingStyle != 1 {
		f.WriteInt(43, m.Text.LineSpacingStyle)
	}
	if m.Text.AttachmentPoint != MLeaderAttachmentMiddleCenter {
		f.WriteInt(71, m.Text.AttachmentPoint)
	}
	if m.Text.BackgroundColor != 0 || m.Text.UseBackgroundColor {
		f.WriteInt(92, int(m.Text.BackgroundColor))
	}
	if m.Text.BackgroundColorScale != 1.0 {
		f.WriteFloat(141, m.Text.BackgroundColorScale)
	}
	if m.Text.UseBackgroundColor {
		f.WriteInt(291, 1)
	} else {
		f.WriteInt(291, 0)
	}
	if m.Text.FillColor != 0 || m.Text.UseFillColor {
		f.WriteInt(93, int(m.Text.FillColor))
	}
	if m.Text.FillColorScale != 1.0 {
		f.WriteFloat(142, m.Text.FillColorScale)
	}
	if m.Text.UseFillColor {
		f.WriteInt(292, 1)
	} else {
		f.WriteInt(292, 0)
	}
	if m.Text.ColumnType != 0 {
		f.WriteInt(174, m.Text.ColumnType)
	}
	if m.Text.UseAutoHeight {
		f.WriteInt(274, 1)
	} else {
		f.WriteInt(274, 0)
	}
	if m.Text.ColumnWidth != 0.0 {
		f.WriteFloat(175, m.Text.ColumnWidth)
	}
	if m.Text.GutterWidth != 0.0 {
		f.WriteFloat(176, m.Text.GutterWidth)
	}
	if m.Text.FlowDirection != 0 {
		f.WriteInt(177, m.Text.FlowDirection)
	}
}

// writeBlockContent writes block content data
func (m *MLeader) writeBlockContent(f format.Formatter) {
	if m.BlockName != "" {
		f.WriteString(3, m.BlockName)
	}
}

// writeToleranceContent writes tolerance content data
func (m *MLeader) writeToleranceContent(f format.Formatter) {
	f.WriteInt(175, m.ToleranceType)
	f.WriteFloat(176, m.ToleranceValue1)
	f.WriteFloat(177, m.ToleranceValue2)
}

// AddLeaderPoint adds a point to the first leader line
func (m *MLeader) AddLeaderPoint(x, y, z float64) {
	if len(m.Lines) == 0 {
		// Create first line if none exists
		m.Lines = append(m.Lines, LeaderLine{
			Type:      MLeaderLineTypeStraight,
			Vertices:  make([]LeaderVertex, 0),
			Color:     0,
			Weight:    0,
			Visible:   true,
			HasBreak:  false,
			HasDogleg: false,
		})
	}
	m.Lines[0].Vertices = append(m.Lines[0].Vertices, LeaderVertex{
		Point: []float64{x, y, z},
		Bulge: 0.0,
	})
}

// GetLeaderPoints returns all points from the first leader line
func (m *MLeader) GetLeaderPoints() [][]float64 {
	if len(m.Lines) > 0 {
		points := make([][]float64, len(m.Lines[0].Vertices))
		for i, vertex := range m.Lines[0].Vertices {
			points[i] = vertex.Point
		}
		return points
	}
	return make([][]float64, 0)
}

// BBox returns bounding box
func (m *MLeader) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)

	// Initialize with large/small values
	for i := 0; i < 3; i++ {
		mins[i] = 1e100
		maxs[i] = -1e100
	}

	// Calculate bbox from all leader line points
	for _, line := range m.Lines {
		for _, vertex := range line.Vertices {
			point := vertex.Point
			for i := 0; i < 3 && i < len(point); i++ {
				if point[i] < mins[i] {
					mins[i] = point[i]
				}
				if point[i] > maxs[i] {
					maxs[i] = point[i]
				}
			}
		}
	}

	return mins, maxs
}

// SetDefaultTextContent sets the default text content
func (m *MLeader) SetDefaultTextContent(content string) {
	m.DefaultTextContent = content
}

// SetMTextContent sets enhanced MTEXT content
func (m *MLeader) SetMTextContent(content string, style string, height float64, textColor color.ColorNumber, options ...func(*MTextData)) {
	m.ContentType = 1 // MTEXT
	m.MText = &MTextData{
		Content:              content,
		Style:                style,
		Color:                textColor,
		CharHeight:           height,
		Rotation:             0.0,
		Width:                1.0,
		Alignment:            MLeaderAttachmentMiddleCenter,
		LineSpacingFactor:    1.0,
		LineSpacingStyle:     1, // At least
		AttachmentPoint:      MLeaderAttachmentMiddleCenter,
		BackgroundColor:      0,
		BackgroundColorScale: 1.0,
		UseBackgroundColor:   false,
		FillColor:            0,
		FillColorScale:       1.0,
		UseFillColor:         false,
		ColumnType:           0,
		UseAutoHeight:        true,
		ColumnWidth:          0.0,
		GutterWidth:          0.0,
		FlowDirection:        0,
		ConnectionBox:        ConnectionBox{},
	}

	// Apply optional configuration
	for _, option := range options {
		option(m.MText)
	}

	// Update legacy text data for compatibility
	m.Text.Content = content
	m.Text.Style = style
	m.Text.Color = textColor
	m.Text.Height = height
}

// SetEnhancedBlockContent sets enhanced block content
func (m *MLeader) SetEnhancedBlockContent(blockName string, transform math.Matrix44, options ...func(*BlockData)) {
	m.ContentType = 2 // BLOCK
	m.Block = &BlockData{
		Name:          blockName,
		Transform:     transform,
		Scale:         1.0,
		Rotation:      0.0,
		XDirection:    math.NewVec3(1, 0, 0),
		YDirection:    math.NewVec3(0, 1, 0),
		HasAttribs:    false,
		AttribHandles: make([]string, 0),
	}

	// Apply optional configuration
	for _, option := range options {
		option(m.Block)
	}

	// Update legacy block name for compatibility
	m.BlockName = blockName
}

// SetEnhancedContext sets the enhanced context data
func (m *MLeader) SetEnhancedContext(context *MLeaderContext) {
	m.EnhancedContext = context

	// Update basic properties from context
	if context.Scale != 0.0 {
		m.Scale = context.Scale
	}
	if context.CharHeight != 0.0 {
		m.Text.Height = context.CharHeight
	}
	if context.ArrowHeadSize != 0.0 {
		m.ArrowheadSize = context.ArrowHeadSize
	}
	if context.LandingGapSize != 0.0 {
		m.LandingGap = context.LandingGapSize
	}
}

// CalculateConnectionBox calculates the connection box for MTEXT content
func (m *MLeader) CalculateConnectionBox() ConnectionBox {
	if m.MText == nil {
		return ConnectionBox{}
	}

	// Simplified connection box calculation
	// In a full implementation, this would calculate actual text extents
	textWidth := float64(len(m.MText.Content)) * m.MText.CharHeight * 0.6
	textHeight := m.MText.CharHeight

	return ConnectionBox{
		Left:   -textWidth / 2,
		Top:    textHeight / 2,
		Right:  textWidth / 2,
		Bottom: -textHeight / 2,
	}
}

// AddLeaderVertex adds a vertex with bulge to first leader line
func (m *MLeader) AddLeaderVertex(x, y, z float64, bulge float64) {
	if len(m.Lines) == 0 {
		// Create first line if none exists
		m.Lines = append(m.Lines, LeaderLine{
			Type:      MLeaderLineTypeStraight,
			Vertices:  make([]LeaderVertex, 0),
			Color:     0,
			Weight:    0,
			Visible:   true,
			HasBreak:  false,
			HasDogleg: false,
		})
	}
	m.Lines[0].Vertices = append(m.Lines[0].Vertices, LeaderVertex{
		Point: []float64{x, y, z},
		Bulge: bulge,
	})
}

// SetLeaderLineType sets the leader line type for all lines
func (m *MLeader) SetLeaderLineType(lineType int) {
	m.LeaderLineType = lineType
	for i := range m.Lines {
		m.Lines[i].Type = lineType
	}
}

// SetStyleOverride enables style override flags
func (m *MLeader) SetStyleOverride(overrideFlags int) {
	m.HasStyleOverrides = true
	m.OverrideFlags = overrideFlags
}

// String returns string representation
func (m *MLeader) String() string {
	var contentType string
	switch m.ContentType {
	case 1:
		contentType = "MTEXT"
	case 2:
		contentType = "BLOCK"
	case 3:
		contentType = "TOLERANCE"
	default:
		contentType = "NONE"
	}

	var enhancedInfo string
	if m.EnhancedContext != nil {
		enhancedInfo = fmt.Sprintf(", EnhancedContext")
	}
	if m.MText != nil {
		enhancedInfo += fmt.Sprintf(", MText(%s)", m.MText.Content[:min(10, len(m.MText.Content))])
	}
	if m.Block != nil {
		enhancedInfo += fmt.Sprintf(", Block(%s)", m.Block.Name)
	}

	return fmt.Sprintf("MLeader{Type: %d, Content: %s, Lines: %d%s}",
		m.ContextType, contentType, len(m.Lines), enhancedInfo)
}

// Helper function for string min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
