package entity

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/table"
)

// MTextAttachmentPoint represents MTEXT attachment point
type MTextAttachmentPoint int

const (
	MTextTopLeft MTextAttachmentPoint = iota + 1
	MTextTopCenter
	MTextTopRight
	MTextMiddleLeft
	MTextMiddleCenter
	MTextMiddleRight
	MTextBottomLeft
	MTextBottomCenter
	MTextBottomRight
)

// MTextFlowDirection represents text flow direction
type MTextFlowDirection int

const (
	MTextLeftToRight MTextFlowDirection = iota + 1
	MTextTopToBottom
	MTextByStyle
)

// MTextLineSpacingStyle represents line spacing style
type MTextLineSpacingStyle int

const (
	MTextAtLeast MTextLineSpacingStyle = iota + 1
	MTextExact
)

// MTextBackgroundFillType represents background fill type
type MTextBackgroundFillType int

const (
	MTextBackgroundOff MTextBackgroundFillType = iota
	MTextBackgroundColor
	MTextBackgroundDrawingColor
	MTextBackgroundUseColor
	MTextBackgroundFrame
)

// MTextColumnType represents column type
type MTextColumnType int

const (
	MTextColumnNone MTextColumnType = iota
	MTextColumnStatic
	MTextColumnDynamic
)

// MTextColumns represents MTEXT column configuration
type MTextColumns struct {
	Type               MTextColumnType
	Count              int
	AutoHeight         bool
	ReversedColumnFlow bool
	DefinedHeight      float64
	Width              float64
	GutterWidth        float64
	TotalWidth         float64
	TotalHeight        float64
	Heights            []float64 // Heights for each column when auto_height is false
	LinkedColumns      []*MText  // Linked MTEXT entities for DXF versions < R2018
}

// NewMTextColumns creates a new MTextColumns with default values
func NewMTextColumns() *MTextColumns {
	return &MTextColumns{
		Type:               MTextColumnStatic,
		Count:              1,
		AutoHeight:         false,
		ReversedColumnFlow: false,
		DefinedHeight:      0.0,
		Width:              0.0,
		GutterWidth:        0.0,
		TotalWidth:         0.0,
		TotalHeight:        0.0,
		Heights:            make([]float64, 0),
		LinkedColumns:      make([]*MText, 0),
	}
}

// MText represents enhanced MTEXT Entity with full Python ezdxf compatibility
type MText struct {
	*entity

	// Basic properties
	Coord1        []float64 // 10, 20, 30 - Insertion point
	Coord2        []float64 // 11, 21, 31 - Direction vector (when rotation is present)
	Height        float64   // 40 - Nominal text height
	Width         float64   // 41 - Reference column width
	DefinedHeight float64   // 46 - Defined height (R2018+)
	Rotation      float64   // 50 - Text rotation in degrees
	WidthFactor   float64   // 41 - Width factor
	ObliqueAngle  float64   // 51 - Oblique angle in degrees

	// Text content and styling
	Value          string       // 1 - Primary text content
	AdditionalText []string     // 3 - Additional text content (multiple tags allowed)
	Style          *table.Style // 7 - Text style name

	// Attachment and alignment
	AttachmentPoint MTextAttachmentPoint // 71 - Attachment point
	FlowDirection   MTextFlowDirection   // 72 - Flow direction

	// Line spacing
	LineSpacingStyle  MTextLineSpacingStyle // 73 - Line spacing style
	LineSpacingFactor float64               // 44 - Line spacing factor (0.25-4.0)

	// Background fill
	BgFillColor      MTextBackgroundFillType // 90 - Background fill type
	BgFillColorValue int                     // 63 - Background fill color as ACI
	BgFillColorTrue  int                     // 421 - Background fill true color
	BgFillColorAlpha int                     // 431 - Background fill color with transparency
	BgFillScale      float64                 // 45 - Box fill scale

	// Computed dimensions (read-only)
	RectWidth  float64 // 42 - Horizontal width of characters (read-only)
	RectHeight float64 // 43 - Vertical height of MTEXT (read-only)

	// 3D properties
	Extrusion []float64 // 210, 220, 230 - Extrusion direction

	// Enhanced features
	Columns *MTextColumns // Column configuration for R2018+

	// Legacy flags for compatibility
	GenFlag        int // 71 - Legacy generation flag
	HorizontalFlag int // 72 - Legacy horizontal flag
	VerticalFlag   int // 73 - Legacy vertical flag
}

// MTextTextFragment represents a formatted text fragment
type MTextTextFragment struct {
	Text        string
	Font        string
	Height      float64
	Color       int
	TrueColor   int // RGB color
	Bold        bool
	Italic      bool
	Underline   bool
	Strikeout   bool
	Overline    bool
	Alignment   MTextTextAlignment
	Tracking    float64 // Character spacing
	WidthFactor float64
	Stacking    *MTextStacking // Fraction/stacking information
}

// MTextTextAlignment represents text fragment alignment
type MTextTextAlignment int

const (
	MTextAlignDefault MTextTextAlignment = iota
	MTextAlignBottom
	MTextAlignMiddle
	MTextAlignTop
)

// MTextStacking represents text stacking/fraction information
type MTextStacking struct {
	TopText    string
	BottomText string
	Type       MTextStackingType
}

// MTextStackingType represents stacking type
type MTextStackingType int

const (
	MTextStackingNone    MTextStackingType = iota
	MTextStackingOver                      // A over B (no line)
	MTextStackingSlanted                   // A over B (with diagonal line)
	MTextStackingLinear                    // A/B (with horizontal line)
)

// MTextParagraphProperties represents paragraph formatting
type MTextParagraphProperties struct {
	Indent      float64                 // First line indent
	LeftIndent  float64                 // Left paragraph indent
	RightIndent float64                 // Right paragraph indent
	Alignment   MTextParagraphAlignment // Paragraph alignment
	TabStops    []MTextTabStop          // Tab stop positions
}

// MTextParagraphAlignment represents paragraph alignment
type MTextParagraphAlignment int

const (
	MTextParagraphLeft MTextParagraphAlignment = iota
	MTextParagraphRight
	MTextParagraphCenter
	MTextParagraphJustified
	MTextParagraphDistributed
)

// MTextTabStop represents a tab stop position
type MTextTabStop struct {
	Position float64
	Type     MTextTabType
}

// MTextTabType represents tab type
type MTextTabType int

const (
	MTextTabLeft MTextTabType = iota
	MTextTabCenter
	MTextTabRight
)

// MTextFont represents font information
type MTextFont struct {
	Name     string
	Bold     bool
	Italic   bool
	CodePage int
	Pitch    float64
}

// MTextFormattingState represents current formatting state during parsing
type MTextFormattingState struct {
	Font        *MTextFont
	Height      float64
	Color       int
	TrueColor   int
	Bold        bool
	Italic      bool
	Underline   bool
	Overline    bool
	Strikeout   bool
	Alignment   MTextTextAlignment
	Tracking    float64
	WidthFactor float64
	Paragraph   *MTextParagraphProperties
}

// IsEntity is for Entity interface
func (t *MText) IsEntity() bool {
	return true
}

// NewMText creates a new enhanced MText with default values
func NewMText() *MText {
	mtext := &MText{
		entity:            NewEntity(MTEXT),
		Coord1:            []float64{0.0, 0.0, 0.0},
		Coord2:            []float64{1.0, 0.0, 0.0}, // Default text direction
		Height:            2.5,                      // Default text height
		Width:             0.0,
		DefinedHeight:     0.0,
		Rotation:          0.0,
		WidthFactor:       1.0,
		ObliqueAngle:      0.0,
		Value:             "",
		AdditionalText:    make([]string, 0),
		Style:             nil, // Will use "Standard" if nil
		AttachmentPoint:   MTextTopLeft,
		FlowDirection:     MTextLeftToRight,
		LineSpacingStyle:  MTextAtLeast,
		LineSpacingFactor: 1.0,
		BgFillColor:       MTextBackgroundOff,
		BgFillColorValue:  0,
		BgFillColorTrue:   0,
		BgFillColorAlpha:  0,
		BgFillScale:       1.5,
		RectWidth:         0.0,
		RectHeight:        0.0,
		Extrusion:         []float64{0.0, 0.0, 1.0}, // Default extrusion
		Columns:           NewMTextColumns(),
		GenFlag:           0,
		HorizontalFlag:    0,
		VerticalFlag:      0,
	}
	return mtext
}

// SetText sets primary text content
func (m *MText) SetText(text string) {
	m.Value = text
}

// GetText returns complete text content
func (m *MText) GetText() string {
	result := m.Value
	for _, text := range m.AdditionalText {
		result += text
	}
	return result
}

// AddAdditionalText adds additional text content
func (m *MText) AddAdditionalText(text string) {
	m.AdditionalText = append(m.AdditionalText, text)
}

// SetAttachmentPoint sets attachment point
func (m *MText) SetAttachmentPoint(point MTextAttachmentPoint) {
	m.AttachmentPoint = point
	// Update legacy flags for compatibility
	m.updateLegacyFlags()
}

// SetRotation sets text rotation in degrees
func (m *MText) SetRotation(rotation float64) {
	m.Rotation = rotation
}

// SetWidth sets the reference column width
func (m *MText) SetWidth(width float64) {
	m.Width = width
}

// SetLineSpacing sets line spacing parameters
func (m *MText) SetLineSpacing(style MTextLineSpacingStyle, factor float64) {
	m.LineSpacingStyle = style
	if factor >= 0.25 && factor <= 4.0 {
		m.LineSpacingFactor = factor
	}
}

// SetBackgroundFill sets background fill parameters
func (m *MText) SetBackgroundFill(fillType MTextBackgroundFillType, color int, trueColor int, alpha int, scale float64) {
	m.BgFillColor = fillType
	m.BgFillColorValue = color
	m.BgFillColorTrue = trueColor
	m.BgFillColorAlpha = alpha
	if scale > 0 {
		m.BgFillScale = scale
	}
}

// SetColumns sets column configuration
func (m *MText) SetColumns(columns *MTextColumns) {
	m.Columns = columns
}

// GetStyleName returns style name
func (m *MText) GetStyleName() string {
	if m.Style != nil {
		return m.Style.Name()
	}
	return "Standard"
}

// SetCoord sets insertion point coordinates
func (m *MText) SetCoord(coord []float64) {
	if len(coord) >= 3 && len(m.Coord1) >= 3 {
		m.Coord1[0] = coord[0]
		m.Coord1[1] = coord[1]
		m.Coord1[2] = coord[2]
	}
}

// SetHeight sets text height
func (m *MText) SetHeight(height float64) {
	m.Height = height
}

// SetObliqueAngle sets oblique angle
func (m *MText) SetObliqueAngle(angle float64) {
	m.ObliqueAngle = angle
}

// SetExtrusion sets extrusion vector
func (m *MText) SetExtrusion(extrusion []float64) {
	if len(extrusion) >= 3 && len(m.Extrusion) >= 3 {
		m.Extrusion[0] = extrusion[0]
		m.Extrusion[1] = extrusion[1]
		m.Extrusion[2] = extrusion[2]
	}
}

// SetLayer assigns the entity to a layer
func (m *MText) SetLayer(layer *table.Layer) {
	m.entity.layer = layer
}

// updateLegacyFlags updates legacy flag system for compatibility
func (m *MText) updateLegacyFlags() {
	// Convert attachment point to legacy flags
	switch m.AttachmentPoint {
	case MTextTopLeft:
		m.HorizontalFlag = 0
		m.VerticalFlag = 0
	case MTextTopCenter:
		m.HorizontalFlag = 1
		m.VerticalFlag = 0
	case MTextTopRight:
		m.HorizontalFlag = 2
		m.VerticalFlag = 0
	case MTextMiddleLeft:
		m.HorizontalFlag = 0
		m.VerticalFlag = 2
	case MTextMiddleCenter:
		m.HorizontalFlag = 1
		m.VerticalFlag = 2
	case MTextMiddleRight:
		m.HorizontalFlag = 2
		m.VerticalFlag = 2
	case MTextBottomLeft:
		m.HorizontalFlag = 0
		m.VerticalFlag = 1
	case MTextBottomCenter:
		m.HorizontalFlag = 1
		m.VerticalFlag = 1
	case MTextBottomRight:
		m.HorizontalFlag = 2
		m.VerticalFlag = 1
	}
}

// Format writes data to formatter with enhanced MTEXT support
func (m *MText) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbMText")

	// Insertion point
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, m.Coord1[i])
	}

	// Text height
	f.WriteFloat(40, m.Height)

	// Reference width
	if m.Width != 0 {
		f.WriteFloat(41, m.Width)
	}

	// Defined height (R2018+)
	if m.DefinedHeight != 0 {
		f.WriteFloat(46, m.DefinedHeight)
	}

	// Text rotation
	if m.Rotation != 0 {
		f.WriteFloat(50, m.Rotation)
	}

	// Width factor
	if m.WidthFactor != 1.0 {
		f.WriteFloat(41, m.WidthFactor)
	}

	// Oblique angle
	if m.ObliqueAngle != 0 {
		f.WriteFloat(51, m.ObliqueAngle)
	}

	// Text content
	if m.Value != "" {
		f.WriteString(1, m.Value)
	}

	// Additional text content
	for _, text := range m.AdditionalText {
		f.WriteString(3, text)
	}

	// Style name
	f.WriteString(7, m.GetStyleName())

	// Enhanced attachment point
	if m.AttachmentPoint != MTextTopLeft {
		f.WriteInt(71, int(m.AttachmentPoint))
	}

	// Legacy flag support for compatibility
	if m.GenFlag != 0 {
		f.WriteInt(71, m.GenFlag)
	}
	if m.HorizontalFlag != 0 {
		f.WriteInt(72, m.HorizontalFlag)
	}
	if m.VerticalFlag != 0 {
		f.WriteInt(73, m.VerticalFlag)
	}

	// Flow direction
	if m.FlowDirection != MTextLeftToRight {
		f.WriteInt(72, int(m.FlowDirection))
	}

	// Line spacing style
	if m.LineSpacingStyle != MTextAtLeast {
		f.WriteInt(73, int(m.LineSpacingStyle))
	}

	// Line spacing factor
	if m.LineSpacingFactor != 1.0 {
		f.WriteFloat(44, m.LineSpacingFactor)
	}

	// Background fill (R2018+)
	if m.BgFillColor != MTextBackgroundOff {
		f.WriteInt(90, int(m.BgFillColor))

		// Background fill scale
		f.WriteFloat(45, m.BgFillScale)

		// Background color values
		if m.BgFillColorValue != 0 {
			f.WriteInt(63, m.BgFillColorValue)
		}
		if m.BgFillColorTrue != 0 {
			f.WriteInt(421, m.BgFillColorTrue)
		}
		if m.BgFillColorAlpha != 0 {
			f.WriteInt(431, m.BgFillColorAlpha)
		}
	}

	// Rect width (read-only, but include if present)
	if m.RectWidth != 0 {
		f.WriteFloat(42, m.RectWidth)
	}

	// Rect height (read-only, but include if present)
	if m.RectHeight != 0 {
		f.WriteFloat(43, m.RectHeight)
	}

	// Extrusion direction
	if len(m.Extrusion) == 3 {
		for i := 0; i < 3; i++ {
			f.WriteFloat(210+i, m.Extrusion[i])
		}
	}

	f.WriteString(100, "AcDbMText")
}

// String outputs data using default formatter
func (t *MText) String() string {
	f := format.NewASCII()
	return t.FormatString(f)
}

// FormatString outputs data using given formatter
func (t *MText) FormatString(f format.Formatter) string {
	t.Format(f)
	return f.Output()
}

// ParseFormattingCode parses a single MTEXT formatting code with enhanced support
func (m *MText) ParseFormattingCode(text string) (*MTextFormattingCode, int) {
	if len(text) < 2 || text[0] != '\\' {
		return nil, 0
	}

	code := &MTextFormattingCode{}
	consumed := 1

	switch text[1] {
	// Single character commands
	case 'L':
		// Start underline
		code.Type = "underline_start"
		consumed = 2
	case 'l':
		// Stop underline
		code.Type = "underline_end"
		consumed = 2
	case 'O':
		// Start overline
		code.Type = "overline_start"
		consumed = 2
	case 'o':
		// Stop overline
		code.Type = "overline_end"
		consumed = 2
	case 'K':
		// Start strike-through
		code.Type = "strikeout_start"
		consumed = 2
	case 'k':
		// Stop strike-through
		code.Type = "strikeout_end"
		consumed = 2
	case 'P':
		// New paragraph
		code.Type = "paragraph"
		consumed = 2
	case 'N':
		// New column
		code.Type = "column_break"
		consumed = 2
	case '~':
		// Non-breaking space
		code.Type = "nbsp"
		consumed = 2
	case '{':
		// Group start
		code.Type = "group_start"
		consumed = 2
	case '}':
		// Group end
		code.Type = "group_end"
		consumed = 2
	case '\\':
		// Escaped backslash
		code.Type = "char"
		code.Value = '\\'
		consumed = 2

	// Multi-character commands (terminated by ;)
	case 'A', 'C', 'c', 'F', 'f', 'H', 'Q', 'S', 'T', 'W', 'p':
		return m.parseLongCommand(text)
	default:
		// Unknown command, skip the backslash
		return &MTextFormattingCode{Type: "unknown", Value: string(text[1])}, 2
	}

	return code, consumed
}

// MTextFormattingCode represents a parsed formatting code
type MTextFormattingCode struct {
	Type     string
	Value    interface{}
	Consumed int
	IsValid  bool
}

// CaretNotation represents special caret-encoded characters
type CaretNotation int

const (
	CaretTab      CaretNotation = iota // ^I
	CaretLineFeed CaretNotation = iota // ^J
	CaretReturn   CaretNotation = iota // ^M
)

// SpecialCharEncoding represents special character encodings
type SpecialCharEncoding int

const (
	SpecialCharDiameter  SpecialCharEncoding = iota // %%c = Ø
	SpecialCharDegree    SpecialCharEncoding = iota // %%d = °
	SpecialCharPlusMinus SpecialCharEncoding = iota // %%p = ±
)

// decodeCaretNotation decodes caret-encoded special characters
func (m *MText) decodeCaretNotation(text string) string {
	result := text
	// Decode caret notation
	replacePairs := []struct {
		from string
		to   string
	}{
		{"^I", "\t"}, // Tab
		{"^J", "\n"}, // Line feed
		{"^M", "\r"}, // Carriage return
	}

	for _, pair := range replacePairs {
		result = strings.ReplaceAll(result, pair.from, pair.to)
	}
	return result
}

// encodeCaretNotation encodes special characters using caret notation
func (m *MText) encodeCaretNotation(text string) string {
	result := text
	replacePairs := []struct {
		from string
		to   string
	}{
		{"\t", "^I"}, // Tab
		{"\n", "^J"}, // Line feed
		{"\r", "^M"}, // Carriage return
	}

	for _, pair := range replacePairs {
		result = strings.ReplaceAll(result, pair.from, pair.to)
	}
	return result
}

// decodeSpecialCharacters decodes %% special character encodings
func (m *MText) decodeSpecialCharacters(text string) string {
	result := text
	replacePairs := []struct {
		from string
		to   string
	}{
		{"%%c", "Ø"}, // Diameter symbol
		{"%%d", "°"}, // Degree symbol
		{"%%p", "±"}, // Plus-minus symbol
	}

	for _, pair := range replacePairs {
		result = strings.ReplaceAll(result, pair.from, pair.to)
	}
	return result
}

// encodeSpecialCharacters encodes special characters using %% notation
func (m *MText) encodeSpecialCharacters(text string) string {
	result := text
	replacePairs := []struct {
		from string
		to   string
	}{
		{"Ø", "%%c"}, // Diameter symbol
		{"°", "%%d"}, // Degree symbol
		{"±", "%%p"}, // Plus-minus symbol
	}

	for _, pair := range replacePairs {
		result = strings.Replace(result, pair.from, pair.to, -1)
	}
	return result
}

// extractPlainText removes all formatting codes to return plain text
func (m *MText) extractPlainText(text string, fast bool) string {
	if fast {
		// Fast mode - simple backslash removal
		result := strings.Builder{}
		i := 0
		for i < len(text) {
			if text[i] == '\\' && i+1 < len(text) {
				// Skip formatting code and its argument
				i += 2
				// Skip to next semicolon for multi-character codes
				if i < len(text) && strings.ContainsRune("ACFfHcQSTWp", rune(text[i])) {
					for i < len(text) && text[i] != ';' {
						i++
					}
					if i < len(text) {
						i++ // Skip semicolon
					}
				}
			} else {
				result.WriteByte(text[i])
				i++
			}
		}
		return m.decodeCaretNotation(m.decodeSpecialCharacters(result.String()))
	}

	// Accurate mode - full parsing
	fragments := m.ParseTextContent()
	result := strings.Builder{}
	for _, fragment := range fragments {
		result.WriteString(fragment.Text)
	}
	return result.String()
}

// estimateTextExtents estimates text dimensions with basic font metrics
func (m *MText) estimateTextExtents(text string, font *MTextFont, width float64) (float64, float64) {
	if font == nil {
		font = &MTextFont{Name: "Arial"}
	}

	// Basic font metrics estimation
	charWidth := m.Height * 0.6 // Average character width multiplier
	if font.Pitch > 0 {
		charWidth = font.Pitch
	}

	// Count actual characters (excluding formatting)
	plainText := m.extractPlainText(text, false)
	charCount := float64(len(plainText))

	textWidth := charCount * charWidth
	if width > 0 {
		textWidth = width // Use specified width if available
	}

	// Estimate lines needed
	lines := 1.0
	if textWidth > width && width > 0 {
		lines = math.Ceil(textWidth / width)
	}

	textHeight := m.Height * lines
	return textWidth, textHeight
}

// wrapText wraps text to specified width using AutoCAD-compatible word breaking
func (m *MText) wrapText(text string, width float64, font *MTextFont) []string {
	if width <= 0 {
		return []string{text}
	}

	plainText := m.extractPlainText(text, false)
	words := strings.Fields(plainText)

	lines := []string{""}
	currentLine := strings.Builder{}
	currentWidth := 0.0

	charWidth := m.Height * 0.6 // Average character width
	if font != nil && font.Pitch > 0 {
		charWidth = font.Pitch
	}

	for _, word := range words {
		wordWidth := float64(len(word)) * charWidth

		if currentWidth+wordWidth <= width {
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
				currentWidth += charWidth
			}
			currentLine.WriteString(word)
			currentWidth += wordWidth
		} else {
			// Word doesn't fit, start new line
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
			currentWidth = wordWidth
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// parseLongCommand handles multi-character formatting commands
func (m *MText) parseLongCommand(text string) (*MTextFormattingCode, int) {
	if len(text) < 3 {
		return nil, 1
	}

	code := &MTextFormattingCode{}
	startIndex := 2

	// Find terminating semicolon
	semicolonIndex := -1
	for i := startIndex; i < len(text); i++ {
		if text[i] == ';' {
			semicolonIndex = i
			break
		}
	}

	if semicolonIndex == -1 {
		return nil, 1
	}

	command := text[1]
	args := text[startIndex:semicolonIndex]
	consumed := semicolonIndex + 1

	switch command {
	case 'A':
		// Alignment relative to current line
		if align, err := strconv.Atoi(args); err == nil {
			code.Type = "align_line"
			switch align {
			case 0:
				code.Value = MTextAlignBottom
			case 1:
				code.Value = MTextAlignMiddle
			case 2:
				code.Value = MTextAlignTop
			default:
				code.Value = MTextAlignDefault
			}
		}

	case 'C':
		// Color change by ACI
		if color, err := strconv.Atoi(args); err == nil {
			code.Type = "color_aci"
			code.Value = color
		}

	case 'c':
		// Color change by RGB
		if color, err := strconv.Atoi(args); err == nil {
			code.Type = "color_rgb"
			code.Value = color
		}

	case 'F', 'f':
		// Font selection
		code.Type = "font"
		code.Value = m.parseFontString(args)

	case 'H':
		// Text height
		if strings.HasSuffix(args, "x") {
			// Relative height factor
			if factor, err := strconv.ParseFloat(args[:len(args)-1], 64); err == nil {
				code.Type = "height_factor"
				code.Value = factor
			}
		} else {
			// Absolute height
			if height, err := strconv.ParseFloat(args, 64); err == nil {
				code.Type = "height_absolute"
				code.Value = height
			}
		}

	case 'Q':
		// Oblique angle
		if angle, err := strconv.ParseFloat(args, 64); err == nil {
			code.Type = "oblique_angle"
			code.Value = angle
		}

	case 'S':
		// Stacking/fractions
		code.Type = "stacking"
		code.Value = m.parseStackingText(args)

	case 'T':
		// Tracking (character spacing)
		if strings.HasSuffix(args, "x") {
			// Relative tracking factor
			if factor, err := strconv.ParseFloat(args[:len(args)-1], 64); err == nil {
				code.Type = "tracking_factor"
				code.Value = factor
			}
		} else {
			// Absolute tracking
			if tracking, err := strconv.ParseFloat(args, 64); err == nil {
				code.Type = "tracking_absolute"
				code.Value = tracking
			}
		}

	case 'W':
		// Width factor
		if strings.HasSuffix(args, "x") {
			// Relative width factor
			if factor, err := strconv.ParseFloat(args[:len(args)-1], 64); err == nil {
				code.Type = "width_factor_relative"
				code.Value = factor
			}
		} else {
			// Absolute width factor
			if factor, err := strconv.ParseFloat(args, 64); err == nil {
				code.Type = "width_factor_absolute"
				code.Value = factor
			}
		}

	case 'p':
		// Paragraph properties
		code.Type = "paragraph_properties"
		code.Value = m.parseParagraphProperties(args)
	}

	return code, consumed
}

// parseFontString parses font specification string
func (m *MText) parseFontString(fontStr string) *MTextFont {
	font := &MTextFont{
		Name:   "Arial", // Default
		Bold:   false,
		Italic: false,
	}

	// Parse font name and properties
	parts := strings.Split(fontStr, "|")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		switch part[0] {
		case 'b':
			// Bold setting
			if len(part) > 1 {
				font.Bold = part[1] != '0'
			}
		case 'i':
			// Italic setting
			if len(part) > 1 {
				font.Italic = part[1] != '0'
			}
		case 'c':
			// Code page
			if len(part) > 1 {
				if codepage, err := strconv.Atoi(part[1:]); err == nil {
					font.CodePage = codepage
				}
			}
		case 'p':
			// Pitch
			if len(part) > 1 {
				if pitch, err := strconv.ParseFloat(part[1:], 64); err == nil {
					font.Pitch = pitch
				}
			}
		default:
			// Font name (first part without property prefix)
			if font.Name == "Arial" {
				font.Name = part
			}
		}
	}

	return font
}

// parseStackingText parses stacking/fraction text
func (m *MText) parseStackingText(stackStr string) *MTextStacking {
	stacking := &MTextStacking{}

	// Different stacking formats
	if strings.Contains(stackStr, "^") {
		// A over B (no line)
		parts := strings.SplitN(stackStr, "^", 2)
		if len(parts) == 2 {
			stacking.TopText = strings.TrimSpace(parts[0])
			stacking.BottomText = strings.TrimSpace(parts[1])
			stacking.Type = MTextStackingOver
		}
	} else if strings.Contains(stackStr, "/") {
		// A over B (with line)
		parts := strings.SplitN(stackStr, "/", 2)
		if len(parts) == 2 {
			stacking.TopText = strings.TrimSpace(parts[0])
			stacking.BottomText = strings.TrimSpace(parts[1])
			stacking.Type = MTextStackingSlanted
		}
	} else if strings.Contains(stackStr, "#") {
		// Diagonal fraction
		parts := strings.SplitN(stackStr, "#", 2)
		if len(parts) == 2 {
			stacking.TopText = strings.TrimSpace(parts[0])
			stacking.BottomText = strings.TrimSpace(parts[1])
			stacking.Type = MTextStackingLinear
		}
	}

	return stacking
}

// parseParagraphProperties parses paragraph properties
func (m *MText) parseParagraphProperties(propStr string) *MTextParagraphProperties {
	props := &MTextParagraphProperties{
		Alignment: MTextParagraphLeft,
		TabStops:  make([]MTextTabStop, 0),
	}

	// Parse properties like: pi5,l10,r15,qc
	parts := strings.Split(propStr, ",")
	for _, part := range parts {
		if len(part) < 2 {
			continue
		}

		prefix := part[0]
		value := part[1:]

		switch prefix {
		case 'i':
			// First line indent
			if indent, err := strconv.ParseFloat(value, 64); err == nil {
				props.Indent = indent
			}
		case 'l':
			// Left indent
			if indent, err := strconv.ParseFloat(value, 64); err == nil {
				props.LeftIndent = indent
			}
		case 'r':
			// Right indent
			if indent, err := strconv.ParseFloat(value, 64); err == nil {
				props.RightIndent = indent
			}
		case 'q':
			// Alignment
			switch value {
			case "l":
				props.Alignment = MTextParagraphLeft
			case "r":
				props.Alignment = MTextParagraphRight
			case "c":
				props.Alignment = MTextParagraphCenter
			case "j":
				props.Alignment = MTextParagraphJustified
			case "d":
				props.Alignment = MTextParagraphDistributed
			}
		case 't':
			// Tab stops
			props.TabStops = m.parseTabStops(value)
		}
	}

	return props
}

// parseTabStops parses tab stop specifications
func (m *MText) parseTabStops(tabStr string) []MTextTabStop {
	tabStops := make([]MTextTabStop, 0)

	// Parse tab positions like: 5,c8,r12,16,c20
	tabs := strings.Split(tabStr, ",")
	for _, tab := range tabs {
		if len(tab) == 0 {
			continue
		}

		tabStop := MTextTabStop{Type: MTextTabLeft}

		if tab[0] == 'c' {
			tabStop.Type = MTextTabCenter
			tab = tab[1:]
		} else if tab[0] == 'r' {
			tabStop.Type = MTextTabRight
			tab = tab[1:]
		}

		if pos, err := strconv.ParseFloat(tab, 64); err == nil {
			tabStop.Position = pos
			tabStops = append(tabStops, tabStop)
		}
	}

	return tabStops
}

// HasFormattingCodes checks if text contains MTEXT formatting codes
func (m *MText) HasFormattingCodes(text string) bool {
	// Each inline formatting code starts with a backslash "\"
	// Remove common codes and check if any backslashes remain
	cleaned := text
	// Remove paragraph breaks (\P)
	for strings.Contains(cleaned, "\\P") {
		cleaned = strings.Replace(cleaned, "\\P", "", -1)
	}
	// Remove non-breaking spaces (\~)
	for strings.Contains(cleaned, "\\~") {
		cleaned = strings.Replace(cleaned, "\\~", "", -1)
	}

	return strings.Contains(cleaned, "\\")
}

// BBox calculates bounding box of MTEXT entity
func (m *MText) BBox() ([]float64, []float64) {
	if len(m.Coord1) < 3 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}

	minX := m.Coord1[0]
	minY := m.Coord1[1]
	minZ := m.Coord1[2]

	// Estimate text dimensions based on width and height
	textWidth := m.Width
	if textWidth == 0 {
		// Estimate from text content
		textWidth = float64(len(m.Value)) * m.Height * 0.6 // Rough estimate
	}

	maxX := minX + textWidth
	maxY := minY + m.Height
	maxZ := minZ

	// Adjust for attachment point
	switch m.AttachmentPoint {
	case MTextTopCenter, MTextMiddleCenter, MTextBottomCenter:
		centerX := minX + textWidth/2
		maxX = centerX + textWidth/2
		minX = centerX - textWidth/2
	case MTextTopRight, MTextMiddleRight, MTextBottomRight:
		maxX = minX + textWidth
	case MTextTopLeft, MTextMiddleLeft, MTextBottomLeft:
		maxX = minX + textWidth
	}

	// Adjust for vertical alignment
	switch m.AttachmentPoint {
	case MTextMiddleLeft, MTextMiddleCenter, MTextMiddleRight:
		centerY := minY + m.Height/2
		maxY = centerY + m.Height/2
		minY = centerY - m.Height/2
	case MTextBottomLeft, MTextBottomCenter, MTextBottomRight:
		maxY = minY + m.Height
	case MTextTopLeft, MTextTopCenter, MTextTopRight:
		maxY = minY
		minY = minY - m.Height
	}

	// Consider columns if present
	if m.Columns != nil && m.Columns.Count > 1 && m.Columns.TotalWidth > 0 {
		maxX = minX + m.Columns.TotalWidth
		if m.Columns.TotalHeight > 0 {
			maxY = minY + m.Columns.TotalHeight
		}
	}

	return []float64{minX, minY, minZ}, []float64{maxX, maxY, maxZ}
}

// Helper functions for legacy compatibility

func (t *MText) togglegenflag(val int) {
	if t.GenFlag&val != 0 {
		t.GenFlag &= ^val
	} else {
		t.GenFlag |= val
	}
}

// Anchor sets anchor point flags (legacy compatibility)
func (t *MText) Anchor(pos int) {
	switch pos {
	case 0: // LEFT_BASE
		t.SetAttachmentPoint(MTextTopLeft)
	case 1: // CENTER_BASE
		t.SetAttachmentPoint(MTextTopCenter)
	case 2: // RIGHT_BASE
		t.SetAttachmentPoint(MTextTopRight)
	case 3: // LEFT_BOTTOM
		t.SetAttachmentPoint(MTextBottomLeft)
	case 4: // CENTER_BOTTOM
		t.SetAttachmentPoint(MTextBottomCenter)
	case 5: // RIGHT_BOTTOM
		t.SetAttachmentPoint(MTextBottomRight)
	case 6: // LEFT_CENTER
		t.SetAttachmentPoint(MTextMiddleLeft)
	case 7: // CENTER_CENTER
		t.SetAttachmentPoint(MTextMiddleCenter)
	case 8: // RIGHT_CENTER
		t.SetAttachmentPoint(MTextMiddleRight)
	case 9: // LEFT_TOP
		t.SetAttachmentPoint(MTextTopLeft)
	case 10: // CENTER_TOP
		t.SetAttachmentPoint(MTextTopCenter)
	case 11: // RIGHT_TOP
		t.SetAttachmentPoint(MTextTopRight)
	}
}

// ParseTextContent parses MTEXT content with formatting codes into fragments
func (m *MText) ParseTextContent() []MTextTextFragment {
	fragments := make([]MTextTextFragment, 0)
	text := m.GetText()

	state := &MTextFormattingState{
		Height:      m.Height,
		Color:       1, // Default white/ByLayer
		TrueColor:   0,
		Alignment:   MTextAlignDefault,
		Tracking:    0.0,
		WidthFactor: 1.0,
	}

	var currentText strings.Builder
	i := 0

	for i < len(text) {
		if text[i] == '\\' && i+1 < len(text) {
			// Add current text as fragment if not empty
			if currentText.Len() > 0 {
				fragments = append(fragments, m.createFragment(currentText.String(), state))
				currentText.Reset()
			}

			// Parse formatting code
			code, consumed := m.ParseFormattingCode(text[i:])
			if code != nil {
				m.applyFormattingCode(state, code)
			}
			i += consumed
		} else {
			currentText.WriteByte(text[i])
			i++
		}
	}

	// Add remaining text
	if currentText.Len() > 0 {
		fragments = append(fragments, m.createFragment(currentText.String(), state))
	}

	return fragments
}

// createFragment creates a text fragment from current state
func (m *MText) createFragment(text string, state *MTextFormattingState) MTextTextFragment {
	fragment := MTextTextFragment{
		Text:        text,
		Height:      state.Height,
		Color:       state.Color,
		TrueColor:   state.TrueColor,
		Bold:        state.Bold,
		Italic:      state.Italic,
		Underline:   state.Underline,
		Strikeout:   state.Strikeout,
		Overline:    state.Overline,
		Alignment:   state.Alignment,
		Tracking:    state.Tracking,
		WidthFactor: state.WidthFactor,
	}

	if state.Font != nil {
		fragment.Font = state.Font.Name
	}

	return fragment
}

// applyFormattingCode applies formatting code to current state
func (m *MText) applyFormattingCode(state *MTextFormattingState, code *MTextFormattingCode) {
	switch code.Type {
	case "underline_start":
		state.Underline = true
	case "underline_end":
		state.Underline = false
	case "overline_start":
		state.Overline = true
	case "overline_end":
		state.Overline = false
	case "strikeout_start":
		state.Strikeout = true
	case "strikeout_end":
		state.Strikeout = false
	case "color_aci":
		if color, ok := code.Value.(int); ok {
			state.Color = color
			state.TrueColor = 0
		}
	case "color_rgb":
		if color, ok := code.Value.(int); ok {
			state.TrueColor = color
		}
	case "font":
		if font, ok := code.Value.(*MTextFont); ok {
			state.Font = font
			state.Bold = font.Bold
			state.Italic = font.Italic
		}
	case "height_absolute":
		if height, ok := code.Value.(float64); ok {
			state.Height = height
		}
	case "height_factor":
		if factor, ok := code.Value.(float64); ok {
			state.Height = m.Height * factor
		}
	case "oblique_angle":
		// This would affect text rendering, not stored in fragment
		// Could be stored if needed for advanced rendering
	case "align_line":
		if align, ok := code.Value.(MTextTextAlignment); ok {
			state.Alignment = align
		}
	case "tracking_absolute":
		if tracking, ok := code.Value.(float64); ok {
			state.Tracking = tracking
		}
	case "tracking_factor":
		if factor, ok := code.Value.(float64); ok {
			state.Tracking = m.Height * factor * 0.1 // Rough conversion
		}
	case "width_factor_absolute":
		if factor, ok := code.Value.(float64); ok {
			state.WidthFactor = factor
		}
	case "width_factor_relative":
		if factor, ok := code.Value.(float64); ok {
			state.WidthFactor = factor
		}
	case "paragraph_properties":
		if props, ok := code.Value.(*MTextParagraphProperties); ok {
			state.Paragraph = props
		}
	}
}

// ToFormattedString converts MTEXT to properly formatted DXF string
func (m *MText) ToFormattedString() string {
	// Start with the main text content
	result := m.Value

	// Add additional text content
	for _, text := range m.AdditionalText {
		result += text
	}

	// Escape any backslashes that aren't formatting codes
	// This is a simple implementation - more sophisticated escaping may be needed
	var escaped strings.Builder
	for i, char := range result {
		if char == '\\' && i+1 < len([]rune(result)) {
			nextChar := []rune(result)[i+1]
			// Check if next character starts a valid formatting code
			if !m.isValidFormattingCode(nextChar) {
				escaped.WriteRune('\\')
				escaped.WriteRune('\\')
			} else {
				escaped.WriteRune(char)
			}
		} else {
			escaped.WriteRune(char)
		}
	}

	return escaped.String()
}

// isValidFormattingCode checks if character starts a valid formatting code
func (m *MText) isValidFormattingCode(char rune) bool {
	validCodes := "LCRJXTP{}~\\AOlKfFcSFHTWp"
	return strings.ContainsRune(validCodes, char)
}

// MTextEditor provides utility methods for building formatted MTEXT
type MTextEditor struct {
	builder strings.Builder
	state   *MTextFormattingState
}

// NewMTextEditor creates a new MTextEditor
func NewMTextEditor() *MTextEditor {
	return &MTextEditor{
		builder: strings.Builder{},
		state: &MTextFormattingState{
			Height:      2.5,
			Color:       1,
			Alignment:   MTextAlignDefault,
			Tracking:    0.0,
			WidthFactor: 1.0,
		},
	}
}

// AddText adds plain text
func (e *MTextEditor) AddText(text string) *MTextEditor {
	e.builder.WriteString(text)
	return e
}

// AddParagraphBreak adds a paragraph break
func (e *MTextEditor) AddParagraphBreak() *MTextEditor {
	e.builder.WriteString("\\P")
	return e
}

// AddColumnBreak adds a column break
func (e *MTextEditor) AddColumnBreak() *MTextEditor {
	e.builder.WriteString("\\N")
	return e
}

// AddNonBreakingSpace adds a non-breaking space
func (e *MTextEditor) AddNonBreakingSpace() *MTextEditor {
	e.builder.WriteString("\\~")
	return e
}

// SetColor sets text color by ACI
func (e *MTextEditor) SetColor(aci int) *MTextEditor {
	e.builder.WriteString(fmt.Sprintf("\\C%d;", aci))
	e.state.Color = aci
	return e
}

// SetTrueColor sets text color by RGB
func (e *MTextEditor) SetTrueColor(rgb int) *MTextEditor {
	e.builder.WriteString(fmt.Sprintf("\\c%d;", rgb))
	e.state.TrueColor = rgb
	return e
}

// SetHeight sets absolute text height
func (e *MTextEditor) SetHeight(height float64) *MTextEditor {
	e.builder.WriteString(fmt.Sprintf("\\H%.2f;", height))
	e.state.Height = height
	return e
}

// SetHeightFactor sets relative text height factor
func (e *MTextEditor) SetHeightFactor(factor float64) *MTextEditor {
	e.builder.WriteString(fmt.Sprintf("\\H%.2fx;", factor))
	return e
}

// SetFont sets font with formatting
func (e *MTextEditor) SetFont(name string, bold, italic bool) *MTextEditor {
	fontSpec := name
	if bold {
		fontSpec += "|b1"
	}
	if italic {
		fontSpec += "|i1"
	}
	e.builder.WriteString(fmt.Sprintf("\\F%s;", fontSpec))
	return e
}

// SetWidthFactor sets width factor
func (e *MTextEditor) SetWidthFactor(factor float64) *MTextEditor {
	e.builder.WriteString(fmt.Sprintf("\\W%.2f;", factor))
	e.state.WidthFactor = factor
	return e
}

// SetTracking sets character spacing
func (e *MTextEditor) SetTracking(tracking float64) *MTextEditor {
	e.builder.WriteString(fmt.Sprintf("\\T%.2f;", tracking))
	e.state.Tracking = tracking
	return e
}

// SetObliqueAngle sets oblique angle
func (e *MTextEditor) SetObliqueAngle(angle float64) *MTextEditor {
	e.builder.WriteString(fmt.Sprintf("\\Q%.2f;", angle))
	return e
}

// SetAlignment sets line alignment
func (e *MTextEditor) SetAlignment(align MTextTextAlignment) *MTextEditor {
	var alignCode int
	switch align {
	case MTextAlignBottom:
		alignCode = 0
	case MTextAlignMiddle:
		alignCode = 1
	case MTextAlignTop:
		alignCode = 2
	default:
		alignCode = 0
	}
	e.builder.WriteString(fmt.Sprintf("\\A%d;", alignCode))
	e.state.Alignment = align
	return e
}

// StartUnderline starts underlining text
func (e *MTextEditor) StartUnderline() *MTextEditor {
	e.builder.WriteString("\\L")
	e.state.Underline = true
	return e
}

// StopUnderline stops underlining text
func (e *MTextEditor) StopUnderline() *MTextEditor {
	e.builder.WriteString("\\l")
	e.state.Underline = false
	return e
}

// StartOverline starts overlining text
func (e *MTextEditor) StartOverline() *MTextEditor {
	e.builder.WriteString("\\O")
	e.state.Overline = true
	return e
}

// StopOverline stops overlining text
func (e *MTextEditor) StopOverline() *MTextEditor {
	e.builder.WriteString("\\o")
	e.state.Overline = false
	return e
}

// StartStrikeout starts strike-through text
func (e *MTextEditor) StartStrikeout() *MTextEditor {
	e.builder.WriteString("\\K")
	e.state.Strikeout = true
	return e
}

// StopStrikeout stops strike-through text
func (e *MTextEditor) StopStrikeout() *MTextEditor {
	e.builder.WriteString("\\k")
	e.state.Strikeout = false
	return e
}

// AddStackedText adds stacked/fraction text
func (e *MTextEditor) AddStackedText(top, bottom string, stackType MTextStackingType) *MTextEditor {
	var stackStr string
	switch stackType {
	case MTextStackingOver:
		stackStr = fmt.Sprintf("\\S%s^%s;", top, bottom)
	case MTextStackingSlanted:
		stackStr = fmt.Sprintf("\\S%s/%s;", top, bottom)
	case MTextStackingLinear:
		stackStr = fmt.Sprintf("\\S%s#%s;", top, bottom)
	default:
		stackStr = fmt.Sprintf("%s/%s", top, bottom)
	}
	e.builder.WriteString(stackStr)
	return e
}

// SetParagraphProperties sets paragraph formatting
func (e *MTextEditor) SetParagraphProperties(props *MTextParagraphProperties) *MTextEditor {
	var propStr strings.Builder

	if props.Indent != 0 {
		propStr.WriteString(fmt.Sprintf("i%.2f", props.Indent))
	}
	if props.LeftIndent != 0 {
		if propStr.Len() > 0 {
			propStr.WriteString(",")
		}
		propStr.WriteString(fmt.Sprintf("l%.2f", props.LeftIndent))
	}
	if props.RightIndent != 0 {
		if propStr.Len() > 0 {
			propStr.WriteString(",")
		}
		propStr.WriteString(fmt.Sprintf("r%.2f", props.RightIndent))
	}

	// Alignment
	if props.Alignment != MTextParagraphLeft {
		if propStr.Len() > 0 {
			propStr.WriteString(",")
		}
		switch props.Alignment {
		case MTextParagraphRight:
			propStr.WriteString("qr")
		case MTextParagraphCenter:
			propStr.WriteString("qc")
		case MTextParagraphJustified:
			propStr.WriteString("qj")
		case MTextParagraphDistributed:
			propStr.WriteString("qd")
		}
	}

	// Tab stops
	if len(props.TabStops) > 0 {
		if propStr.Len() > 0 {
			propStr.WriteString(",")
		}
		propStr.WriteString("t")
		for i, tab := range props.TabStops {
			if i > 0 {
				propStr.WriteString(",")
			}
			switch tab.Type {
			case MTextTabCenter:
				propStr.WriteString("c")
			case MTextTabRight:
				propStr.WriteString("r")
			}
			propStr.WriteString(fmt.Sprintf("%.2f", tab.Position))
		}
	}

	if propStr.Len() > 0 {
		e.builder.WriteString(fmt.Sprintf("\\p%s;", propStr.String()))
	}

	return e
}

// StartGroup starts a formatting group
func (e *MTextEditor) StartGroup() *MTextEditor {
	e.builder.WriteString("{")
	return e
}

// EndGroup ends a formatting group
func (e *MTextEditor) EndGroup() *MTextEditor {
	e.builder.WriteString("}")
	return e
}

// AddTab adds a tab character
func (e *MTextEditor) AddTab() *MTextEditor {
	e.builder.WriteString("\\T")
	return e
}

// Clear clears the current content
func (e *MTextEditor) Clear() *MTextEditor {
	e.builder.Reset()
	return e
}

// String returns the formatted MTEXT string
func (e *MTextEditor) String() string {
	return e.builder.String()
}

// ApplyToMText applies the formatted text to an MTEXT entity
func (e *MTextEditor) ApplyToMText(mtext *MText) *MText {
	mtext.SetText(e.String())
	return mtext
}

// Convenience methods for common formatting patterns

// BoldText adds bold text
func (e *MTextEditor) BoldText(text string) *MTextEditor {
	return e.StartGroup().SetFont("", true, false).AddText(text).EndGroup()
}

// ItalicText adds italic text
func (e *MTextEditor) ItalicText(text string) *MTextEditor {
	return e.StartGroup().SetFont("", false, true).AddText(text).EndGroup()
}

// BoldItalicText adds bold italic text
func (e *MTextEditor) BoldItalicText(text string) *MTextEditor {
	return e.StartGroup().SetFont("", true, true).AddText(text).EndGroup()
}

// UnderlineText adds underlined text
func (e *MTextEditor) UnderlineText(text string) *MTextEditor {
	return e.StartGroup().StartUnderline().AddText(text).StopUnderline().EndGroup()
}

// ColorText adds text with specific color
func (e *MTextEditor) ColorText(text string, aci int) *MTextEditor {
	return e.StartGroup().SetColor(aci).AddText(text).EndGroup()
}

// HeightText adds text with specific height
func (e *MTextEditor) HeightText(text string, height float64) *MTextEditor {
	return e.StartGroup().SetHeight(height).AddText(text).EndGroup()
}

// FontText adds text with specific font
func (e *MTextEditor) FontText(text, fontName string, bold, italic bool) *MTextEditor {
	return e.StartGroup().SetFont(fontName, bold, italic).AddText(text).EndGroup()
}
