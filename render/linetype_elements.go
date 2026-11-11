package render

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
	"github.com/edanko/dxf/table"
	"github.com/edanko/dxf/text2path"
)

// linetypeElementPaths places the shape and text elements of a complex
// linetype along the entity paths. Dash and gap elements are already handled
// by the style dash array; only the visible glyph geometry is generated here.
func linetypeElementPaths(d *drawing.Drawing, e entity.Entity, paths []StyledPath, dev float64, opts *SceneOptions, nt *sceneNotes) []StyledPath {
	switch e.(type) {
	case *entity.Insert, *entity.Dimension:
		// Composite entities place geometry through their children.
		return nil
	}
	lt := resolveEntityLineType(d, e)
	if lt == nil {
		return nil
	}
	elems := lt.Elements()
	if len(elems) == 0 {
		return nil
	}
	ltscale := entityLineTypeScale(e)
	total := lt.TotalLength() * ltscale
	if total <= 0 {
		return nil
	}
	seen := make(map[string]bool)
	var out []StyledPath
	for _, sp := range paths {
		if sp.Path == nil {
			continue
		}
		approx := sp.Path.Approximate(dev, 4)
		for _, run := range polylineRuns(approx) {
			out = append(out, placeLineTypeElements(run, elems, ltscale, total, lt.Name(), opts, sp.Style, seen, nt)...)
		}
	}
	return out
}

// resolveEntityLineType resolves the effective linetype of an entity, looking
// up named linetypes in the drawing. ByLayer falls back to the entity layer and
// ByBlock/Continuous resolve to no complex elements.
func resolveEntityLineType(d *drawing.Drawing, e entity.Entity) *table.LineType {
	name := ""
	if lt, ok := e.(entity.LinetypeReader); ok {
		name = lt.Linetype()
	}
	switch name {
	case "ByBlock", "Continuous":
		return nil
	case "ByLayer", "":
		if layer := e.Layer(); layer != nil {
			return layer.LineType
		}
		return nil
	}
	if d == nil {
		return nil
	}
	lt, err := d.LineType(name)
	if err != nil {
		return nil
	}
	return lt
}

func entityLineTypeScale(e entity.Entity) float64 {
	if v, ok := e.GetAttributes()["ltscale"].(float64); ok && v > 0 {
		return v
	}
	return 1
}

// polylineRuns flattens a path into runs of connected points. Curves are
// expected to be approximated already.
func polylineRuns(p *path.Path) [][]dxfmath.Vec2 {
	if p == nil {
		return nil
	}
	cmds := p.Commands()
	verts := p.Vertices()
	var runs [][]dxfmath.Vec2
	var cur []dxfmath.Vec2
	i := 0
	for _, cmd := range cmds {
		switch cmd.Type() {
		case path.CommandMoveTo:
			if len(cur) > 1 {
				runs = append(runs, cur)
			}
			if i < len(verts) {
				cur = []dxfmath.Vec2{verts[i]}
				i++
			}
		case path.CommandLineTo:
			if i < len(verts) {
				cur = append(cur, verts[i])
				i++
			}
		case path.CommandCurve3To:
			i += 2
		case path.CommandCurve4To:
			i += 3
		}
	}
	if len(cur) > 1 {
		runs = append(runs, cur)
	}
	return runs
}

func placeLineTypeElements(run []dxfmath.Vec2, elems []table.LineTypeElement, ltscale, total float64, ltName string, opts *SceneOptions, st Style, seen map[string]bool, nt *sceneNotes) []StyledPath {
	if len(run) < 2 || total <= 0 {
		return nil
	}
	lengths := make([]float64, len(run))
	runLength := 0.0
	for i := 1; i < len(run); i++ {
		lengths[i] = run[i].Distance(run[i-1])
		runLength += lengths[i]
	}
	if runLength <= 0 {
		return nil
	}
	elemStyle := st
	elemStyle.Dash = nil
	elemStyle.Fill = color.RGBA{}

	var out []StyledPath
	for cycle := 0.0; cycle < runLength; cycle += total {
		pos := 0.0
		for _, el := range elems {
			span := math.Abs(el.DashLength) * ltscale
			at := cycle + pos + el.XOffset*ltscale
			if at > runLength {
				break
			}
			switch {
			case el.TextValue != "":
				height := elementHeight(el.XScale, ltscale)
				point, tangent := pointAtDistance(run, lengths, at)
				rotation := math.Atan2(tangent.Y(), tangent.X())*180/math.Pi + el.Rotation
				anchor := elementAnchor(point, tangent, el.YOffset*ltscale)
				if p := TextToPath(el.TextValue, height, anchor.X(), anchor.Y(), rotation); p != nil {
					out = append(out, StyledPath{Path: p, Style: elemStyle})
				}
			case el.ShapeNum > 0:
				height := elementHeight(el.XScale, ltscale)
				point, tangent := pointAtDistance(run, lengths, at)
				rotation := math.Atan2(tangent.Y(), tangent.X())*180/math.Pi + el.Rotation
				anchor := elementAnchor(point, tangent, el.YOffset*ltscale)
				font, ok := shapeFontFor(opts, el.ShapeNum)
				if !ok {
					key := fmt.Sprintf("%s/%d", ltName, el.ShapeNum)
					if !seen[key] {
						seen[key] = true
						nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("missing shape %d in linetype %s", el.ShapeNum, ltName))
					}
					break
				}
				shape := font.ShapePath(el.ShapeNum, height)
				if shape == nil {
					key := fmt.Sprintf("%s/%d", ltName, el.ShapeNum)
					if !seen[key] {
						seen[key] = true
						nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("missing shape %d in linetype %s", el.ShapeNum, ltName))
					}
					break
				}
				m := dxfmath.Translation(anchor.X(), anchor.Y(), 0).Mul(dxfmath.RotationZ(rotation * math.Pi / 180))
				out = append(out, StyledPath{Path: shape.Transform(&m), Style: elemStyle})
			}
			pos += span
			if cycle+pos >= runLength {
				break
			}
		}
	}
	return out
}

func elementHeight(scale, ltscale float64) float64 {
	if scale <= 0 {
		scale = 1
	}
	return scale * ltscale
}

func elementAnchor(point, tangent dxfmath.Vec2, yOffset float64) dxfmath.Vec2 {
	return point.Add(tangent.Perpendicular().Mul(yOffset))
}

// pointAtDistance returns the point and unit tangent at the given arc length
// along the run.
func pointAtDistance(run []dxfmath.Vec2, lengths []float64, distance float64) (dxfmath.Vec2, dxfmath.Vec2) {
	if distance <= 0 {
		return run[0], run[1].Sub(run[0]).Normalized()
	}
	covered := 0.0
	for i := 1; i < len(run); i++ {
		length := lengths[i]
		if distance <= covered+length && length > 0 {
			t := (distance - covered) / length
			point := run[i-1].Lerp(run[i], t)
			tangent := run[i].Sub(run[i-1]).Normalized()
			return point, tangent
		}
		covered += length
	}
	last := run[len(run)-1]
	return last, run[len(run)-1].Sub(run[len(run)-2]).Normalized()
}

// shapeFontFor selects a shape font deterministically: ltypeshp is preferred,
// then fonts with the glyph in lexicographic name order.
func shapeFontFor(opts *SceneOptions, number int) (*text2path.SHXFont, bool) {
	if opts == nil || len(opts.ShapeFonts) == 0 {
		return nil, false
	}
	names := make([]string, 0, len(opts.ShapeFonts))
	for name := range opts.ShapeFonts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		base := strings.ToLower(strings.TrimSuffix(name, ".shx"))
		if strings.Contains(base, "ltypeshp") {
			if font := opts.ShapeFonts[name]; font != nil {
				if _, ok := font.Glyphs[number]; ok {
					return font, true
				}
			}
		}
	}
	for _, name := range names {
		font := opts.ShapeFonts[name]
		if font == nil {
			continue
		}
		if _, ok := font.Glyphs[number]; ok {
			return font, true
		}
	}
	return nil, false
}
