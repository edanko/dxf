package render

import (
	"image/color"
	"math"
	"strings"

	dxfcolor "github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
	"github.com/edanko/dxf/patterns"
	"github.com/edanko/dxf/table"
	"github.com/edanko/dxf/text2path"
)

// SceneOptions controls scene construction.
type SceneOptions struct {
	// Layout selects a layout by name; empty means model space.
	Layout string
	// Min and Max give explicit model-space extents (x, y). When nil the
	// drawing header extents (or the scene bounding box) are used.
	Min, Max []float64
	// Background is the canvas color used by raster output and to resolve
	// ACI color 7 (white/black). Zero value means white.
	Background color.RGBA
	// Tessellation is the maximum curve deviation in model units.
	Tessellation float64
}

// StyledScene is a canvas-ready drawing: paths with resolved styles
// under a fitted model-to-canvas transform.
type StyledScene struct {
	Paths      []StyledPath
	Background color.RGBA
	Scale      float64
	ExtMinX    float64
	ExtMinY    float64
	ExtMaxX    float64
	ExtMaxY    float64
	// Fallbacks lists degradations applied during scene construction,
	// e.g. text rendered with the built-in vector font.
	Fallbacks []string
}

// BuildScene lowers every visible entity of the drawing (or of the
// named layout) to styled canvas-space paths fitted into a
// width x height canvas with a top-left origin.
func BuildScene(d *drawing.Drawing, width, height float64, opts *SceneOptions) (*StyledScene, error) {
	if opts == nil {
		opts = &SceneOptions{}
	}
	bg := opts.Background
	if bg.A == 0 {
		bg = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	dev := opts.Tessellation
	if dev <= 0 {
		dev = 1e-4
	}

	var ents entity.Entities
	switch {
	case opts.Layout == "" || strings.EqualFold(opts.Layout, "Model"):
		ents = d.Entities()
	default:
		ents = d.LayoutEntities(opts.Layout)
	}

	var paths []StyledPath
	nt := &sceneNotes{}
	for _, e := range ents {
		st, ok := resolveStyle(e, bg)
		if !ok {
			continue
		}
		paths = append(paths, entityPaths(d, e, st, dev, bg, nt)...)
	}

	// Fit transform: an active viewport in the entity list projects the
	// model space through its view window; otherwise the extents are
	// fitted into the canvas with a 5% margin.
	minX, minY, maxX, maxY, ok := sceneExtents(d, opts, paths)
	scale := 1.0
	tx, ty := 0.0, 0.0
	explicitExtents := len(opts.Min) >= 2 && len(opts.Max) >= 2
	if vp := pickViewport(ents); vp != nil && vp.ViewHeight > 0 &&
		width > 0 && height > 0 && !explicitExtents {
		scale = height / vp.ViewHeight
		vcx, vcy := vp.ViewCenterPoint[0], vp.ViewCenterPoint[1]
		halfW, halfH := width/(2*scale), height/(2*scale)
		minX, maxX = vcx-halfW, vcx+halfW
		minY, maxY = vcy-halfH, vcy+halfH
		tx, ty = width/2-vcx*scale, height/2+vcy*scale
		ok = true
	} else if ok && width > 0 && height > 0 {
		w, h := maxX-minX, maxY-minY
		if w <= 0 {
			w = 1e-9
		}
		if h <= 0 {
			h = 1e-9
		}
		scale = 0.95 * math.Min(width/w, height/h)
		tx = (width-w*scale)/2 - minX*scale
		ty = (height-h*scale)/2 + maxY*scale
	}
	fit := dxfmath.Matrix44{}
	fit[0] = scale
	fit[3] = tx
	fit[5] = -scale
	fit[7] = ty
	fit[10] = 1
	fit[15] = 1

	for i := range paths {
		paths[i].Path = paths[i].Path.Transform(&fit)
		for j := range paths[i].Style.Dash {
			paths[i].Style.Dash[j] *= scale
		}
		if hf := paths[i].FillHatch; hf != nil {
			hf.Scale *= scale
			hf.Angle = -hf.Angle
			if hf.Gradient != nil {
				hf.Gradient.Rotation = -hf.Gradient.Rotation
			}
		}
	}
	// Expand hatch pattern fills into a boundary stroke plus clipped
	// pattern segments.
	var expanded []StyledPath
	for _, sp := range paths {
		if sp.FillHatch != nil && sp.FillHatch.Pattern != nil {
			hf := sp.FillHatch
			segs := HatchFillSegments(sp.Path, hf, sp.Style.Width, dev*math.Max(scale, 1e-9))
			bnd := sp.Style
			bnd.Fill = color.RGBA{}
			expanded = append(expanded, StyledPath{Path: sp.Path, Style: bnd, FillHatch: hf})
			if segs != nil && segs.HasLines() {
				st := sp.Style
				st.Fill = color.RGBA{}
				expanded = append(expanded, StyledPath{Path: segs, Style: st})
			}
			continue
		}
		expanded = append(expanded, sp)
	}
	paths = expanded

	sc := &StyledScene{
		Paths: paths, Background: bg, Scale: scale,
		ExtMinX: minX, ExtMinY: minY, ExtMaxX: maxX, ExtMaxY: maxY,
	}
	if nt.gridText {
		sc.Fallbacks = append(sc.Fallbacks,
			"TEXT/MTEXT: built-in vector font used (no system TrueType font available)")
	}
	return sc, nil
}

func sceneExtents(d *drawing.Drawing, opts *SceneOptions, paths []StyledPath) (float64, float64, float64, float64, bool) {
	if len(opts.Min) >= 2 && len(opts.Max) >= 2 {
		return opts.Min[0], opts.Min[1], opts.Max[0], opts.Max[1], true
	}
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, sp := range paths {
		lo, hi := sp.Path.BoundingBox()
		if lo.X() < minX {
			minX = lo.X()
		}
		if lo.Y() < minY {
			minY = lo.Y()
		}
		if hi.X() > maxX {
			maxX = hi.X()
		}
		if hi.Y() > maxY {
			maxY = hi.Y()
		}
	}
	if len(paths) == 0 {
		return 0, 0, 0, 0, false
	}
	if mn, mx, ok := d.Extents(); ok && (mx.X() > mn.X() || mx.Y() > mn.Y()) {
		// Union with header extents: header may lag entities added after
		// the last extents update (hatches, exploded inserts, ...).
		if mn.X() < minX {
			minX = mn.X()
		}
		if mn.Y() < minY {
			minY = mn.Y()
		}
		if mx.X() > maxX {
			maxX = mx.X()
		}
		if mx.Y() > maxY {
			maxY = mx.Y()
		}
	}
	return minX, minY, maxX, maxY, true
}

// entityPaths converts one entity into styled model-space paths.
func entityPaths(d *drawing.Drawing, e entity.Entity, st Style, dev float64, bg color.RGBA, nt *sceneNotes) []StyledPath {
	switch t := e.(type) {
	case *entity.Text:
		p, fontName := runText(t.Value, t.Height, t.Coord1[0], t.Coord1[1], t.Rotation)
		if p == nil {
			return nil
		}
		nt.recordText(fontName)
		stt := st
		if fontName != "" {
			stt = textStyle(st)
		}
		return []StyledPath{{Path: p, Style: stt}}
	case *entity.MText:
		p, fontName := multiLineText(t.Value, t.Height, t.Width, t.Coord1[0], t.Coord1[1], t.Rotation, nt)
		if p == nil {
			return nil
		}
		stt := st
		if fontName != "" {
			stt = textStyle(st)
		}
		return []StyledPath{{Path: p, Style: stt}}
	case *entity.Insert:
		kids, err := t.Explode()
		if err != nil {
			return nil
		}
		var out []StyledPath
		for _, k := range kids {
			if tk, isText := k.(*entity.Text); isText {
				if p, fontName := runText(tk.Value, tk.Height, tk.Coord1[0], tk.Coord1[1], tk.Rotation); p != nil {
					nt.recordText(fontName)
					stt := st
					if fontName != "" {
						stt = textStyle(st)
					}
					out = append(out, StyledPath{Path: p, Style: stt})
				}
				continue
			}
			p, err := path.MakePath(k, path.WithDistance(dev))
			if err == nil {
				out = append(out, StyledPath{Path: p, Style: st})
			}
		}
		return out
	case *entity.Hatch:
		p, err := path.MakePath(t, path.WithDistance(dev))
		if err != nil {
			return nil
		}
		if t.IsSolid() {
			st.Fill = st.Stroke
			return []StyledPath{{Path: p, Style: st}}
		}
		hf := hatchFill(t, 1, 0)
		return []StyledPath{{Path: p, Style: st, FillHatch: hf}}
	case *entity.HatchEnhanced:
		p := hatchBoundaryPath(t)
		if p == nil {
			return nil
		}
		switch {
		case t.IsSolid:
			st.Fill = st.Stroke
			return []StyledPath{{Path: p, Style: st}}
		case t.HasGradient && t.Gradient != nil:
			hf := gradientFillFromEnhanced(t)
			return []StyledPath{{Path: p, Style: st, FillHatch: hf}}
		case t.PatternName != "":
			hf := hatchFill(&entity.Hatch{}, t.PatternScale, t.PatternAngle)
			if hf != nil && hf.Pattern != nil {
				hf.Pattern.Name = t.PatternName
				if def, ok := patterns.GetPatternDefinition(t.PatternName); ok {
					hf.Pattern.Name = def.Name
					hf.Pattern.Lines = nil
					for _, l := range def.Lines {
						hf.Pattern.Lines = append(hf.Pattern.Lines, PatternLine{
							Angle:  l.Angle,
							Base:   v2f(l.BasePoint),
							Offset: v2f(l.Offset),
							Dashes: l.DashLengthItems,
						})
					}
				}
			}
			return []StyledPath{{Path: p, Style: st, FillHatch: hf}}
		default:
			return []StyledPath{{Path: p, Style: st}}
		}
	case *entity.Solid, *entity.Trace, *entity.ThreeDFace:
		p, err := path.MakePath(t, path.WithDistance(dev))
		if err != nil {
			return nil
		}
		st.Fill = st.Stroke
		return []StyledPath{{Path: p, Style: st}}
	}
	p, err := path.MakePath(e, path.WithDistance(dev))
	if err != nil {
		return nil
	}
	return []StyledPath{{Path: p, Style: st}}
}

// hatchFill maps a non-solid hatch entity to pattern filling. Known pattern
// definitions win; otherwise the library's default hatch lines are used.
func hatchFill(t *entity.Hatch, scale, angle float64) *HatchFill {
	if def, ok := patterns.GetPatternDefinition(t.PatternName()); ok {
		pf := &PatternFill{Name: def.Name}
		for _, l := range def.Lines {
			pf.Lines = append(pf.Lines, PatternLine{
				Angle:  l.Angle,
				Base:   v2f(l.BasePoint),
				Offset: v2f(l.Offset),
				Dashes: l.DashLengthItems,
			})
		}
		return &HatchFill{Pattern: pf, Scale: scale, Angle: angle}
	}
	if lines := entity.DefaultPatternLines(); len(lines) > 0 {
		pf := &PatternFill{Name: t.PatternName()}
		for _, l := range lines {
			pf.Lines = append(pf.Lines, PatternLine{
				Angle:  l.Angle,
				Base:   pv2(l.BasePoint),
				Offset: pv2(l.Offset),
				Dashes: l.DashLengthItems,
			})
		}
		return &HatchFill{Pattern: pf, Scale: scale, Angle: angle}
	}
	return nil
}

func pv2(v dxfmath.Vec2) [2]float64 { return [2]float64{v[0], v[1]} }

func v2f(v []float64) [2]float64 {
	if len(v) >= 2 {
		return [2]float64{v[0], v[1]}
	}
	return [2]float64{}
}

// resolveStyle derives the drawing style of an entity, resolving ByLayer
// color/lineweight/linetype/transparency, true color, and ACI 7 against
// the background. ok is false when the entity must not be drawn.
func resolveStyle(e entity.Entity, bg color.RGBA) (Style, bool) {
	if inv, ok := e.(interface{ Invisible() bool }); ok && inv.Invisible() {
		return Style{}, false
	}
	layer := e.Layer()
	if layer != nil && layer.Flags()&5 != 0 { // frozen or off
		return Style{}, false
	}
	attrs := e.GetAttributes()
	aci := 7
	if v, ok := attrs["color"].(int); ok {
		aci = v
	}
	ltscale := 1.0
	if v, ok := attrs["ltscale"].(float64); ok && v > 0 {
		ltscale = v
	}

	// Stroke color: true color (code 420) wins, then ACI with ByLayer.
	var rgb [3]uint8
	if tc, ok := e.(interface{ ColorRGB() int }); ok && tc.ColorRGB() > 0 {
		v := tc.ColorRGB()
		rgb = [3]uint8{uint8(v >> 16), uint8(v >> 8), uint8(v)}
	} else {
		if aci == 256 { // ByLayer
			if layer != nil {
				if layer.TrueColor() > 0 {
					v := layer.TrueColor()
					rgb = [3]uint8{uint8(v >> 16), uint8(v >> 8), uint8(v)}
				} else {
					aci = int(layer.Color)
				}
			} else {
				aci = 7
			}
		}
		if aci == 0 { // ByBlock fallback
			aci = 7
		}
		if aci == 7 {
			rgb = aci7OnBackground(bg)
		} else if rgb == (rgbTriple{}) && aci >= 1 && aci <= 255 {
			c := dxfcolor.IndexColor(uint8(aci))
			rgb = [3]uint8{c[0], c[1], c[2]}
		}
	}
	_ = rgb

	alpha := uint8(255)
	if tr, ok := e.(interface{ Transparency() float64 }); ok && tr.Transparency() > 0 {
		alpha = uint8((1 - tr.Transparency()) * 255)
	} else if layer != nil && layer.Transparency() >= 0 {
		alpha = uint8((1 - float64(layer.Transparency())/100) * 255)
	}

	lw := 25 // default 0.25 mm
	if v, ok := e.(interface{ LineWeight() int }); ok && v.LineWeight() >= 0 {
		lw = v.LineWeight()
	} else if layer != nil && layer.LineWidth() >= 0 {
		lw = layer.LineWidth()
	}
	width := float64(lw) / 100.0 / 25.4 * 96.0

	stroke := color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: alpha}
	st := Style{Stroke: stroke, Width: width, Dash: dashPattern(e, layer, ltscale)}
	return st, true
}

type rgbTriple = [3]uint8

// aci7OnBackground resolves ACI white/black to a canvas color.
func aci7OnBackground(bg color.RGBA) [3]uint8 {
	lum := 0.299*float64(bg.R) + 0.587*float64(bg.G) + 0.114*float64(bg.B)
	if lum > 128 {
		return [3]uint8{0, 0, 0}
	}
	return [3]uint8{255, 255, 255}
}

// dashPattern resolves the linetype of an entity to dash lengths in
// model units (scaled later by the fit transform).
func dashPattern(e entity.Entity, layer *table.Layer, ltscale float64) []float64 {
	name := ""
	if lt, ok := e.(interface{ Linetype() string }); ok {
		name = lt.Linetype()
	}
	var lt *table.LineType
	if layer != nil {
		lt = layer.LineType
	}
	if name == "ByLayer" || name == "" {
		// use layer linetype (already in lt)
	} else if name == "ByBlock" || name == "Continuous" {
		lt = nil
	}
	if lt == nil || lt.TotalLength() <= 0 {
		return nil
	}
	var dash []float64
	for _, el := range lt.Elements() {
		d := el.DashLength
		if d == 0 {
			d = -1e-6 // dot renders as a tiny gap-adjacent point
		}
		dash = append(dash, d*ltscale)
	}
	return dash
}

// TextToPath lays out a single-line TEXT string as glyph paths placed at
// (x, y) with the given height and rotation (degrees).
func TextToPath(text string, height, x, y, rotation float64) *path.Path {
	if text == "" || height <= 0 {
		return nil
	}
	glyphs := text2path.NewTextToPath(height / 120.0).Convert(text)
	if glyphs == nil {
		return nil
	}
	m := dxfmath.Translation(x, y, 0).Mul(dxfmath.RotationZ(rotation * math.Pi / 180))
	return glyphs.Transform(&m)
}

// MTextToPath lays out an MText entity into glyph paths.
func MTextToPath(t *entity.MText) *path.Path {
	text := t.Value
	if text == "" {
		return nil
	}
	x, y := 0.0, 0.0
	if len(t.Coord1) >= 2 {
		x, y = t.Coord1[0], t.Coord1[1]
	}
	p, _ := multiLineText(text, t.Height, t.Width, x, y, t.Rotation, nil)
	return p
}

// vec2From converts the first two components of a float slice into a Vec2.
func vec2From(v []float64) dxfmath.Vec2 {
	if len(v) >= 2 {
		return dxfmath.Vec2{v[0], v[1]}
	}
	return dxfmath.Vec2{}
}

// hatchBoundaryPath builds the boundary path of an enhanced hatch from its
// polyline boundary paths (edge paths are approximated by their vertices).
func hatchBoundaryPath(h *entity.HatchEnhanced) *path.Path {
	out := path.NewPath()
	any := false
	for _, bp := range h.BoundaryPaths {
		if len(bp.Points) == 0 {
			continue
		}
		out.MoveTo(bp.Points[0][0], bp.Points[0][1])
		for _, pt := range bp.Points[1:] {
			out.LineTo(pt[0], pt[1])
		}
		if bp.IsClosed && len(bp.Points) > 2 {
			out.LineTo(bp.Points[0][0], bp.Points[0][1])
		}
		any = true
	}
	if !any {
		return nil
	}
	return out
}

// gradientFillFromEnhanced maps an enhanced hatch gradient to a linear
// two-color fill. Spherical/cylindrical variants become centered fills;
// focus/tint parameters are not modeled.
func gradientFillFromEnhanced(h *entity.HatchEnhanced) *HatchFill {
	g := h.Gradient
	if g == nil {
		return &HatchFill{Scale: 1}
	}
	return &HatchFill{
		Scale: 1,
		Angle: 0,
		Gradient: &GradientFill{
			Color1:   gradientColorRGB(int(g.Color1)),
			Color2:   gradientColorRGB(int(g.Color2)),
			Rotation: g.Angle,
			Centered: g.Centered,
		},
	}
}

// gradientColorRGB converts a DXF color value (truecolor flag bit 30 set,
// otherwise ACI) to normalized RGB. ACI 7 resolves to black on canvas.
func gradientColorRGB(v int) [3]float64 {
	if v&(1<<30) != 0 {
		return [3]float64{float64(v>>16&0xff) / 255, float64(v>>8&0xff) / 255, float64(v&0xff) / 255}
	}
	aci := v & 0x3FFFFFFF
	if aci == 0 || aci == 7 || aci > 255 {
		return [3]float64{0, 0, 0}
	}
	rgb := dxfcolor.IndexColor(uint8(aci))
	return [3]float64{float64(rgb[0]) / 255, float64(rgb[1]) / 255, float64(rgb[2]) / 255}
}

// sceneNotes records scene-level degradations applied while lowering entities.
type sceneNotes struct{ gridText bool }

// recordText notes text rendered with the built-in vector font
// (fontName empty) as a fallback.
func (n *sceneNotes) recordText(fontName string) {
	if fontName == "" {
		n.gridText = true
	}
}

// textStyle renders TrueType text as filled glyph outlines in the
// entity color, with nested contours (holes) cancelled by even-odd fill.
func textStyle(st Style) Style {
	out := st
	out.Fill = st.Stroke
	out.Stroke = color.RGBA{}
	out.FillRule = "evenodd"
	return out
}

// runText lays out single-line TEXT, preferring system TrueType outlines
// and falling back to the built-in vector font. The font file name is
// returned; it is empty when the fallback font was used.
func runText(text string, height, x, y, rotation float64) (*path.Path, string) {
	if text == "" || height <= 0 {
		return nil, ""
	}
	if ps, name, err := text2path.TextToPathsTTF(text, height); err == nil && len(ps) > 0 && ps[0] != nil {
		m := dxfmath.Translation(x, y, 0).Mul(dxfmath.RotationZ(rotation * math.Pi / 180))
		return ps[0].Transform(&m), name
	}
	return TextToPath(text, height, x, y, rotation), ""
}

// pickViewport returns the first active VIEWPORT in the entity list.
func pickViewport(ents entity.Entities) *entity.Viewport {
	for _, e := range ents {
		if vp, ok := e.(*entity.Viewport); ok && vp.Status > 0 {
			return vp
		}
	}
	return nil
}
