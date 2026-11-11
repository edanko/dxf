package text2path

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/edanko/dxf/entity"
	t2p "github.com/edanko/dxf/text2path"
)

// LinePitch is the MTEXT line pitch relative to the text height.
const LinePitch = 1.6714

// TextSize reports the measured extents of a TEXT or MTEXT entity in drawing
// units. MinX/MinY/MaxX/MaxY describe the axis-aligned box after applying the
// entity insertion point, alignment, width factor, and rotation.
type TextSize struct {
	Width   float64
	Height  float64
	Ascent  float64
	Descent float64
	Advance float64
	MinX    float64
	MinY    float64
	MaxX    float64
	MaxY    float64
	Lines   int
	Font    string
	// Fallback is true when no system font could be measured and the
	// built-in vector-font metrics were used instead.
	Fallback       bool
	FallbackReason string
}

// Size measures a single-line TEXT entity.
func Size(t *entity.Text) TextSize {
	if t == nil {
		return TextSize{}
	}
	widthFactor := t.WidthFactor
	if widthFactor <= 0 {
		widthFactor = 1
	}
	m, font, fallback, reason := measureLine(t.Value, t.Height)
	m.Width *= widthFactor
	m.Advance *= widthFactor
	out := TextSize{
		Width:          m.Width,
		Height:         m.Height,
		Ascent:         m.Ascent,
		Descent:        m.Descent,
		Advance:        m.Advance,
		Lines:          1,
		Font:           font,
		Fallback:       fallback,
		FallbackReason: reason,
	}
	if t.Value == "" || t.Height <= 0 {
		return out
	}

	// Local text box before alignment: baseline at y=0.
	minX, maxX := 0.0, m.Width
	minY, maxY := -m.Descent, m.Ascent
	switch t.HorizontalFlag % 3 {
	case 1:
		minX, maxX = -m.Width/2, m.Width/2
	case 2:
		minX, maxX = -m.Width, 0
	}
	switch t.VerticalFlag {
	case 2: // middle
		shift := -(minY + maxY) / 2
		minY += shift
		maxY += shift
	case 3: // top
		shift := -maxY
		minY += shift
		maxY += shift
	}

	ox, oy := 0.0, 0.0
	if len(t.Coord1) >= 2 {
		ox, oy = t.Coord1[0], t.Coord1[1]
	}
	out.MinX, out.MinY, out.MaxX, out.MaxY = rotatedBounds(
		ox, oy, minX, minY, maxX, maxY, t.Rotation)
	return out
}

// MTextSize measures an MTEXT entity using measured word widths for wrapping
// and the configured line spacing and column width.
func MTextSize(m *entity.MText) TextSize {
	if m == nil {
		return TextSize{}
	}
	text := m.GetText()
	if text == "" || m.Height <= 0 {
		return TextSize{}
	}
	width := m.Width
	if m.Columns != nil && m.Columns.Count > 1 && m.Columns.Width > 0 {
		width = m.Columns.Width
	}
	lines := t2p.LayoutMText(text, m.Height, width)
	if len(lines) == 0 {
		return TextSize{}
	}

	mw, font, fallback, reason := measureLine(strings.Join(lines, " "), m.Height)
	maxWidth := 0.0
	for _, line := range lines {
		w, _, _, _ := measureLine(line, m.Height)
		if w.Width > maxWidth {
			maxWidth = w.Width
		}
	}
	spacing := m.LineSpacingFactor
	if spacing <= 0 {
		spacing = 1
	}
	pitch := LinePitch * m.Height * spacing
	height := pitch * float64(len(lines))
	out := TextSize{
		Width:          maxWidth,
		Height:         height,
		Ascent:         mw.Ascent,
		Descent:        mw.Descent,
		Advance:        maxWidth,
		Lines:          len(lines),
		Font:           font,
		Fallback:       fallback,
		FallbackReason: reason,
	}
	if m.Columns != nil && m.Columns.Count > 1 && m.Columns.TotalWidth > 0 {
		out.Width = math.Max(out.Width, m.Columns.TotalWidth)
	}

	minX, maxX := 0.0, out.Width
	minY, maxY := -height, 0.0
	switch m.AttachmentPoint {
	case entity.MTextTopCenter, entity.MTextMiddleCenter, entity.MTextBottomCenter:
		minX, maxX = -out.Width/2, out.Width/2
	case entity.MTextTopRight, entity.MTextMiddleRight, entity.MTextBottomRight:
		minX, maxX = -out.Width, 0
	}
	switch m.AttachmentPoint {
	case entity.MTextMiddleLeft, entity.MTextMiddleCenter, entity.MTextMiddleRight:
		minY, maxY = -height/2, height/2
	case entity.MTextBottomLeft, entity.MTextBottomCenter, entity.MTextBottomRight:
		minY, maxY = 0, height
	}

	ox, oy := 0.0, 0.0
	if len(m.Coord1) >= 2 {
		ox, oy = m.Coord1[0], m.Coord1[1]
	}
	out.MinX, out.MinY, out.MaxX, out.MaxY = rotatedBounds(
		ox, oy, minX, minY, maxX, maxY, m.Rotation)
	return out
}

func measureLine(text string, height float64) (t2p.TextMetrics, string, bool, string) {
	if text == "" || height <= 0 {
		return t2p.TextMetrics{}, "", false, ""
	}
	if m, font, err := t2p.MeasureTTF(text, height); err == nil {
		if m.Advance == 0 {
			m.Advance = m.Width
		}
		return m, font, false, ""
	} else {
		return vectorMetrics(text, height), "", true, err.Error()
	}
}

func vectorMetrics(text string, height float64) t2p.TextMetrics {
	advance := float64(utf8.RuneCountInString(text)) * height * 5 / 6
	return t2p.TextMetrics{
		Width:   advance,
		Height:  height,
		Ascent:  height * 0.8,
		Descent: height * 0.2,
		Advance: advance,
	}
}

func rotatedBounds(ox, oy, minX, minY, maxX, maxY, rotation float64) (float64, float64, float64, float64) {
	rad := rotation * math.Pi / 180
	cosR, sinR := math.Cos(rad), math.Sin(rad)
	corners := [][2]float64{{minX, minY}, {maxX, minY}, {maxX, maxY}, {minX, maxY}}
	first := true
	var outMinX, outMinY, outMaxX, outMaxY float64
	for _, c := range corners {
		x := ox + c[0]*cosR - c[1]*sinR
		y := oy + c[0]*sinR + c[1]*cosR
		if first {
			outMinX, outMaxX = x, x
			outMinY, outMaxY = y, y
			first = false
			continue
		}
		outMinX = math.Min(outMinX, x)
		outMaxX = math.Max(outMaxX, x)
		outMinY = math.Min(outMinY, y)
		outMaxY = math.Max(outMaxY, y)
	}
	return outMinX, outMinY, outMaxX, outMaxY
}
