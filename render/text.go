package render

import (
	"math"

	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
	"github.com/edanko/dxf/text2path"
)

// multiLineText lays out MText content as glyph paths. Lines break at
// \P codes and wrap at the given column width; supported formatting
// codes are stripped. Text is placed top-down from (x, y). TrueType
// outlines are preferred; the built-in vector font is the fallback,
// recorded in nt when used. The font file name is returned, empty for
// the fallback.
func multiLineText(text string, height, width, x, y, rotation float64, nt *sceneNotes) (*path.Path, string) {
	if text == "" || height <= 0 {
		return nil, ""
	}
	lines := text2path.LayoutMText(text, height, width)
	if len(lines) == 0 {
		return nil, ""
	}
	ttp := text2path.NewTextToPath(height / 120.0)
	out := path.NewPath()
	linePitch := 1.6714 * height
	fontName := ""
	gridUsed := false
	for i, line := range lines {
		if line == "" {
			continue
		}
		var glyphs *path.Path
		if ps, name, err := text2path.TextToPathsTTF(line, height); err == nil && len(ps) > 0 && ps[0] != nil {
			glyphs, fontName = ps[0], name
		} else {
			glyphs = ttp.Convert(line)
			gridUsed = true
		}
		if glyphs == nil {
			continue
		}
		m := dxfmath.Translation(x, y-float64(i)*linePitch, 0).
			Mul(dxfmath.RotationZ(rotation * math.Pi / 180))
		linePath := glyphs.Transform(&m)
		appendPathAll(out, linePath)
	}
	if gridUsed {
		fontName = ""
		if nt != nil {
			nt.gridText = true
		}
	}
	return out, fontName
}

// appendPathAll appends every command of src to dst.
func appendPathAll(dst, src *path.Path) {
	for _, cmd := range src.Commands() {
		switch c := cmd.(type) {
		case *path.PathMoveTo:
			dst.MoveTo(c.End.X(), c.End.Y())
		case *path.PathLineTo:
			dst.LineTo(c.End.X(), c.End.Y())
		case *path.PathCurve3To:
			dst.Curve3To(c.End.X(), c.End.Y(), c.Ctrl.X(), c.Ctrl.Y())
		case *path.PathCurve4To:
			dst.Curve4To(c.End.X(), c.End.Y(), c.Ctrl1.X(), c.Ctrl1.Y(), c.Ctrl2.X(), c.Ctrl2.Y())
		}
	}
}

// StripMTextCodes removes common MText formatting codes, keeping
// \P paragraph breaks as newlines.
func StripMTextCodes(text string) string {
	return text2path.StripCodes(text)
}
