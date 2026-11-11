package render

import (
	"image/color"
	"math"
	"strings"

	"github.com/edanko/dxf/addons/acadctb"
	dxfcolor "github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
	"github.com/edanko/dxf/patterns"
	"github.com/edanko/dxf/table"
	"github.com/edanko/dxf/text2path"

	"fmt"
	"github.com/edanko/dxf/object"
	"github.com/edanko/dxf/xref"
	"image"
	_ "image/jpeg"
	"os"
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
	// BasePath is the directory used to resolve image and underlay file
	// references; empty means the process working directory.
	BasePath string
	// ImageSearchPaths are additional directories searched for image files.
	ImageSearchPaths []string
	// PlotStyles applies plot style overrides during scene construction. A
	// color-dependent (CTB) table selects the style by entity ACI color.
	PlotStyles *acadctb.PlotStyleTable
	// PlotStyleName selects a named style from a named (STB) table.
	PlotStyleName string
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
	// Rasters are images drawn on top of the vector paths (IMAGE entities).
	Rasters []StyledImage
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
		st, ok := resolveStyle(e, bg, opts)
		if !ok {
			continue
		}
		paths = append(paths, entityPaths(d, e, st, dev, bg, opts, nt)...)
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
		if paths[i].Clip != nil {
			paths[i].Clip = paths[i].Clip.Transform(&fit)
		}
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
	for i := range paths {
		paths[i].Style = applyOpacity(paths[i].Style)
	}
	for i := range nt.rasters {
		ri := &nt.rasters[i]
		mx, my, mw, mh := ri.X, ri.Y, ri.W, ri.H
		// The fit transform flips Y: the canvas top-left corner is the
		// image's model top-left corner.
		q := fit.MulVec(dxfmath.NewVec3(mx, my+mh, 0))
		ri.X, ri.Y = q[0], q[1]
		ri.W = mw * scale
		ri.H = mh * scale
		ri.Rotation = -ri.Rotation
		if ri.Clip != nil {
			ri.Clip = ri.Clip.Transform(&fit)
		}
	}

	sc := &StyledScene{
		Paths: paths, Background: bg, Scale: scale,
		ExtMinX: minX, ExtMinY: minY, ExtMaxX: maxX, ExtMaxY: maxY,
	}
	if nt.gridText {
		sc.Fallbacks = append(sc.Fallbacks,
			"TEXT/MTEXT: built-in vector font used (no system TrueType font available)")
	}
	sc.Fallbacks = append(sc.Fallbacks, nt.fallbacks...)
	sc.Rasters = nt.rasters
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
func entityPaths(d *drawing.Drawing, e entity.Entity, st Style, dev float64, bg color.RGBA, opts *SceneOptions, nt *sceneNotes) []StyledPath {
	switch t := e.(type) {
	case *entity.Text:
		dx, dy := textAnchorOffset(t)
		p, fontName := runText(t.Value, t.Height, t.Coord1[0]+dx, t.Coord1[1]+dy, t.Rotation)
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
		dx, dy := mtextAnchorOffset(t)
		p, fontName := multiLineText(t.Value, t.Height, t.Width, t.Coord1[0]+dx, t.Coord1[1]+dy, t.Rotation, nt)
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
		if t.IsXClipped() {
			if cp := xclipClipPath(t); cp != nil {
				for i := range out {
					out[i].Clip = cp
				}
			}
		}
		return out
	case *entity.Dimension:
		return dimensionEntityPaths(d, t, st, dev, bg, opts, nt)
	case *entity.Image:
		return imageEntityPaths(d, t, st, opts, nt)
	case *entity.Underlay:
		return underlayEntityPaths(t, st, nt)
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
		if t.PatternName() != "" && (hf == nil || hf.Pattern == nil) {
			nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("HATCH %s: pattern %q not found; boundary retained", t.Handle(), t.PatternName()))
			hf = &HatchFill{Pattern: &PatternFill{Name: t.PatternName()}, Scale: 1}
		}
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
			hf := &HatchFill{Scale: t.PatternScale, Angle: t.PatternAngle}
			if def, ok := patterns.LookupPattern(t.PatternName); ok {
				hf.Pattern = patternFillFromDefinition(def)
			} else {
				hf.Pattern = &PatternFill{Name: t.PatternName}
				nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("HATCH %s: pattern %q not found; boundary retained", t.Handle(), t.PatternName))
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
	case *entity.Mesh:
		return faceEntityPaths(meshFaces(t.Vertices(), t.Faces()), st)
	case *entity.Polyline:
		if t.IsPolyface() {
			return faceEntityPaths(polyfaceFaces(t), st)
		}
		if t.IsPolymesh() {
			return faceEntityPaths(polymeshFaces(t), st)
		}
	}
	p, err := path.MakePath(e, path.WithDistance(dev))
	if err != nil {
		return nil
	}
	return []StyledPath{{Path: p, Style: st}}
}

// dimensionEntityPaths renders a DIMENSION from its anonymous geometry block
// when available, falling back to synthesized extension lines, dimension line
// and measured text for synthetic dimensions. Dimension block entities are
// stored in world coordinates, so no insertion transform is applied.
func dimensionEntityPaths(d *drawing.Drawing, t *entity.Dimension, st Style, dev float64, bg color.RGBA, opts *SceneOptions, nt *sceneNotes) []StyledPath {
	if name := t.BlockName(); name != "" {
		if b, err := d.GetBlock(name); err == nil {
			var out []StyledPath
			for _, be := range b.Entities() {
				child, ok := be.(entity.Entity)
				if !ok {
					continue
				}
				if u, ok := child.(entity.Unwrapper); ok {
					child = u.Unwrap()
				}
				out = append(out, entityPaths(d, child, st, dev, bg, opts, nt)...)
			}
			if len(out) > 0 {
				return out
			}
		}
	}

	parts, err := t.Explode()
	if err != nil {
		return nil
	}
	var out []StyledPath
	for _, part := range parts {
		if tx, ok := part.(*entity.Text); ok {
			p, fontName := runText(tx.Value, tx.Height, tx.Coord1[0], tx.Coord1[1], tx.Rotation)
			if p == nil {
				continue
			}
			nt.recordText(fontName)
			stt := st
			if fontName != "" {
				stt = textStyle(st)
			}
			out = append(out, StyledPath{Path: p, Style: stt})
			continue
		}
		p, err := path.MakePath(part, path.WithDistance(dev))
		if err == nil {
			out = append(out, StyledPath{Path: p, Style: st})
		}
	}
	return out
}

// meshFaces extracts valid MESH faces. MESH face indices are zero-based.
func meshFaces(vertices []dxfmath.Vec3, faces [][]int) [][]dxfmath.Vec3 {
	out := make([][]dxfmath.Vec3, 0, len(faces))
	for _, face := range faces {
		if len(face) < 3 {
			continue
		}
		poly := make([]dxfmath.Vec3, 0, len(face))
		valid := true
		for _, idx := range face {
			if idx < 0 || idx >= len(vertices) {
				valid = false
				break
			}
			poly = append(poly, vertices[idx])
		}
		if valid && !degenerateFace(poly) {
			out = append(out, poly)
		}
	}
	return out
}

// polyfaceFaces extracts POLYFACE faces. Face-record indices are one-based;
// negative values mark an invisible edge but still name a vertex.
func polyfaceFaces(p *entity.Polyline) [][]dxfmath.Vec3 {
	vertices := p.MeshVertices()
	out := make([][]dxfmath.Vec3, 0, len(p.FaceRecords()))
	for _, rec := range p.FaceRecords() {
		indices := rec.FaceVertexIndices()
		if len(indices) < 3 {
			continue
		}
		poly := make([]dxfmath.Vec3, 0, len(indices))
		valid := true
		for _, idx := range indices {
			if idx < 0 {
				idx = -idx
			}
			if idx <= 0 || idx > len(vertices) {
				valid = false
				break
			}
			c := vertices[idx-1].Coord
			poly = append(poly, dxfmath.NewVec3(coordAt(c, 0), coordAt(c, 1), coordAt(c, 2)))
		}
		if valid && !degenerateFace(poly) {
			out = append(out, poly)
		}
	}
	return out
}

// polymeshFaces subdivides a POLYMESH row/column grid into quad faces.
func polymeshFaces(p *entity.Polyline) [][]dxfmath.Vec3 {
	vertices := p.MeshVertices()
	m, n := p.MeshVertexCount, p.MeshFaceCount
	if m < 2 || n < 2 || m*n > len(vertices) {
		return nil
	}
	out := make([][]dxfmath.Vec3, 0, (m-1)*(n-1))
	for row := 0; row < n-1; row++ {
		for col := 0; col < m-1; col++ {
			a := vertices[row*m+col].Coord
			b := vertices[row*m+col+1].Coord
			c := vertices[(row+1)*m+col+1].Coord
			d := vertices[(row+1)*m+col].Coord
			face := []dxfmath.Vec3{
				dxfmath.NewVec3(coordAt(a, 0), coordAt(a, 1), coordAt(a, 2)),
				dxfmath.NewVec3(coordAt(b, 0), coordAt(b, 1), coordAt(b, 2)),
				dxfmath.NewVec3(coordAt(c, 0), coordAt(c, 1), coordAt(c, 2)),
				dxfmath.NewVec3(coordAt(d, 0), coordAt(d, 1), coordAt(d, 2)),
			}
			if !degenerateFace(face) {
				out = append(out, face)
			}
		}
	}
	return out
}

func coordAt(c []float64, i int) float64 {
	if i >= 0 && i < len(c) {
		return c[i]
	}
	return 0
}

func degenerateFace(face []dxfmath.Vec3) bool {
	if len(face) < 3 {
		return true
	}
	a := face[0]
	area := 0.0
	for i := 1; i+1 < len(face); i++ {
		b, c := face[i], face[i+1]
		area += (b.X()-a.X())*(c.Y()-a.Y()) - (c.X()-a.X())*(b.Y()-a.Y())
	}
	return math.Abs(area) < 1e-12
}

// faceEntityPaths emits each face as a closed path. Hidden or degenerate
// faces are skipped; a fill on the incoming style requests filled faces,
// otherwise faces render as wireframe.
func faceEntityPaths(faces [][]dxfmath.Vec3, st Style) []StyledPath {
	out := make([]StyledPath, 0, len(faces))
	for _, face := range faces {
		p := path.NewPath()
		p.MoveTo(face[0].X(), face[0].Y())
		for _, v := range face[1:] {
			p.LineTo(v.X(), v.Y())
		}
		p.Close()
		out = append(out, StyledPath{Path: p, Style: st})
	}
	return out
}

// hatchFill maps a non-solid hatch entity to pattern filling. Known pattern
// definitions win. A named but unknown pattern reports no fill so the caller
// can preserve the boundary and report the missing definition; unnamed
// hatches fall back to the library default lines.
func hatchFill(t *entity.Hatch, scale, angle float64) *HatchFill {
	name := t.PatternName()
	if name != "" {
		def, ok := patterns.LookupPattern(name)
		if !ok {
			return nil
		}
		return &HatchFill{Pattern: patternFillFromDefinition(def), Scale: scale, Angle: angle}
	}
	if lines := entity.DefaultPatternLines(); len(lines) > 0 {
		pf := &PatternFill{Name: name}
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

func patternFillFromDefinition(def patterns.PatternDefinition) *PatternFill {
	pf := &PatternFill{Name: def.Name}
	for _, l := range def.Lines {
		pf.Lines = append(pf.Lines, PatternLine{
			Angle:  l.Angle,
			Base:   v2f(l.BasePoint),
			Offset: v2f(l.Offset),
			Dashes: append([]float64(nil), l.DashLengthItems...),
		})
	}
	return pf
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
func resolveStyle(e entity.Entity, bg color.RGBA, opts *SceneOptions) (Style, bool) {
	if inv, ok := e.(entity.InvisibleReader); ok && inv.Invisible() {
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
	if tc, ok := e.(entity.ColorRGBReader); ok && tc.ColorRGB() > 0 {
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
	alpha := uint8(255)
	if tr, ok := e.(entity.TransparencyReader); ok && tr.Transparency() > 0 {
		alpha = uint8((1 - tr.Transparency()) * 255)
	} else if layer != nil && layer.Transparency() >= 0 {
		alpha = uint8((1 - float64(layer.Transparency())/100) * 255)
	}

	lw := 25 // default 0.25 mm
	if v, ok := e.(entity.LineWeightReader); ok && v.LineWeight() >= 0 {
		lw = v.LineWeight()
	} else if layer != nil && layer.LineWidth() >= 0 {
		lw = layer.LineWidth()
	}
	width := float64(lw) / 100.0 / 25.4 * 96.0
	if opts != nil && opts.PlotStyles != nil {
		var overrides acadctb.PlotOverrides
		var ok bool
		if opts.PlotStyles.ByColor {
			overrides, ok = opts.PlotStyles.ColorOverrides(aci)
		} else if opts.PlotStyleName != "" {
			overrides, ok = opts.PlotStyles.Overrides(opts.PlotStyleName)
		}
		if ok {
			if overrides.HasTrueColor {
				rgb = [3]uint8{
					uint8(overrides.TrueColor >> 16),
					uint8(overrides.TrueColor >> 8),
					uint8(overrides.TrueColor),
				}
			} else if overrides.HasColor {
				if overrides.Color == 7 {
					rgb = aci7OnBackground(bg)
				} else {
					c := dxfcolor.IndexColor(uint8(overrides.Color))
					rgb = [3]uint8{c[0], c[1], c[2]}
				}
			}
			if overrides.HasScreening {
				s := float64(overrides.Screening)
				if s < 0 {
					s = 0
				}
				if s > 100 {
					s = 100
				}
				for i := range rgb {
					rgb[i] = uint8(255 - (255-float64(rgb[i]))*s/100)
				}
			}
			if overrides.HasLineweight && overrides.Lineweight > 0 {
				width = float64(overrides.Lineweight) / 100.0 / 25.4 * 96.0
			}
		}
	}

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
	if lt, ok := e.(entity.LinetypeReader); ok {
		name = lt.Linetype()
	}
	var lt *table.LineType
	if layer != nil {
		lt = layer.LineType
	}
	switch name {
	case "ByLayer", "":
		// use layer linetype (already in lt)
	case "ByBlock", "Continuous":
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
	dx, dy := mtextAnchorOffset(t)
	p, _ := multiLineText(text, t.Height, t.Width, x+dx, y+dy, t.Rotation, nil)
	return p
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
type sceneNotes struct {
	gridText  bool
	rasters   []StyledImage
	fallbacks []string
}

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

// textAnchorOffset returns the measured offset from a TEXT insertion point to
// the baseline origin used by the glyph layout.
func textAnchorOffset(t *entity.Text) (float64, float64) {
	if t == nil {
		return 0, 0
	}
	wf := t.WidthFactor
	if wf <= 0 {
		wf = 1
	}
	w := text2path.MeasureAdvance(t.Value, t.Height) * wf
	dx := 0.0
	switch t.HorizontalFlag % 3 {
	case 1:
		dx = -w / 2
	case 2:
		dx = -w
	}
	dy := 0.0
	switch t.VerticalFlag {
	case 2:
		dy = -t.Height / 2
	case 3:
		dy = -t.Height
	}
	return dx, dy
}

// mtextAnchorOffset returns the measured offset from an MTEXT insertion point
// to the top-left origin used by multiLineText.
func mtextAnchorOffset(t *entity.MText) (float64, float64) {
	if t == nil {
		return 0, 0
	}
	lines := text2path.LayoutMText(t.Value, t.Height, t.Width)
	width := 0.0
	for _, line := range lines {
		if w := text2path.MeasureAdvance(line, t.Height); w > width {
			width = w
		}
	}
	height := float64(len(lines)) * text2pathLinePitch(t)
	dx := 0.0
	switch t.AttachmentPoint {
	case entity.MTextTopCenter, entity.MTextMiddleCenter, entity.MTextBottomCenter:
		dx = -width / 2
	case entity.MTextTopRight, entity.MTextMiddleRight, entity.MTextBottomRight:
		dx = -width
	}
	dy := 0.0
	switch t.AttachmentPoint {
	case entity.MTextMiddleLeft, entity.MTextMiddleCenter, entity.MTextMiddleRight:
		dy = height / 2
	case entity.MTextBottomLeft, entity.MTextBottomCenter, entity.MTextBottomRight:
		dy = height
	}
	return dx, dy
}

func text2pathLinePitch(t *entity.MText) float64 {
	spacing := t.LineSpacingFactor
	if spacing <= 0 {
		spacing = 1
	}
	return 1.6714 * t.Height * spacing
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

// imageEntityPaths renders an IMAGE entity: a frame quad plus a raster when
// the referenced image file resolves.
func imageEntityPaths(d *drawing.Drawing, t *entity.Image, st Style, opts *SceneOptions, nt *sceneNotes) []StyledPath {
	org := t.InsertPoint()
	u := t.UVector()
	v := t.VVector()
	size := t.ImageSize()
	c1x := org.X() + u.X()*size.X()
	c1y := org.Y() + u.Y()*size.X()
	c2x := c1x + v.X()*size.Y()
	c2y := c1y + v.Y()*size.Y()
	c3x := org.X() + v.X()*size.Y()
	c3y := org.Y() + v.Y()*size.Y()
	quad := path.NewPath()
	quad.MoveTo(org.X(), org.Y())
	quad.LineTo(c1x, c1y)
	quad.LineTo(c2x, c2y)
	quad.LineTo(c3x, c3y)
	quad.Close()
	out := []StyledPath{{Path: quad, Style: Style{Stroke: st.Stroke, Width: st.Width}}}

	defPath := ""
	if def, ok := d.GetObject(t.ImageDefHandle()); ok {
		if id, ok := def.(*object.ImageDef); ok {
			defPath = id.FilePath()
		}
	}
	if defPath == "" {
		nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("IMAGE %s: image definition %s not found", t.Handle(), t.ImageDefHandle()))
		return out
	}
	basePath, search := "", []string(nil)
	if opts != nil {
		basePath, search = opts.BasePath, opts.ImageSearchPaths
	}
	resolved, found := xref.ResolveExternalPath(defPath, basePath, search)
	if !found {
		nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("IMAGE %s: image file not found: %s", t.Handle(), defPath))
		return out
	}
	f, err := os.Open(resolved)
	if err != nil {
		nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("IMAGE %s: %v", t.Handle(), err))
		return out
	}
	defer f.Close()
	im, _, err := image.Decode(f)
	if err != nil {
		nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("IMAGE %s: decode %s: %v", t.Handle(), resolved, err))
		return out
	}
	rot := math.Atan2(u.Y(), u.X()) * 180 / math.Pi
	nt.rasters = append(nt.rasters, StyledImage{
		Image: im, X: org.X(), Y: org.Y(),
		W: size.X(), H: size.Y(),
		Rotation: rot, Opacity: 1 - t.Fade()/100,
		Clip: imageClipPath(t, org, u, v, size),
	})
	return out
}

// underlayEntityPaths renders the frame of an underlay reference; the
// underlay content itself is not vectorized.
func underlayEntityPaths(t *entity.Underlay, st Style, nt *sceneNotes) []StyledPath {
	if len(t.InsertPoint) < 2 {
		return nil
	}
	w, h := t.ScaleX, t.ScaleY
	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}
	rad := t.Rotation * math.Pi / 180
	cos, sin := math.Cos(rad), math.Sin(rad)
	rot := func(x, y float64) (float64, float64) {
		return t.InsertPoint[0] + x*cos - y*sin, t.InsertPoint[1] + x*sin + y*cos
	}
	x0, y0 := rot(0, 0)
	x1, y1 := rot(w, 0)
	x2, y2 := rot(w, h)
	x3, y3 := rot(0, h)
	quad := path.NewPath()
	quad.MoveTo(x0, y0)
	quad.LineTo(x1, y1)
	quad.LineTo(x2, y2)
	quad.LineTo(x3, y3)
	quad.Close()
	nt.fallbacks = append(nt.fallbacks, fmt.Sprintf("UNDERLAY %s: content not vectorized (frame only)", t.Handle()))
	return []StyledPath{{Path: quad, Style: Style{Stroke: st.Stroke, Width: st.Width, Opacity: 1 - float64(t.Fade)/100}}}
}

// xclipClipPath builds the clip boundary of an INSERT.
func xclipClipPath(t *entity.Insert) *path.Path {
	pts := t.XClipBoundaryPath()
	if len(pts) < 2 {
		return nil
	}
	p := path.NewPath()
	p.MoveTo(pts[0][0], pts[0][1])
	for _, pt := range pts[1:] {
		p.LineTo(pt[0], pt[1])
	}
	p.Close()
	return p
}

// imageClipPath maps the pixel-space clipping boundary of an IMAGE entity
// into model space.
func imageClipPath(t *entity.Image, org, u, v dxfmath.Vec3, size dxfmath.Vec2) *path.Path {
	if t.Clipping() == 0 || size.X() == 0 || size.Y() == 0 {
		return nil
	}
	pts := t.BoundaryPoints()
	if len(pts) == 0 {
		return nil
	}
	var corners [][2]float64
	if t.ClippingBoundaryType() == entity.ImageClippingRectangular && len(pts) >= 2 {
		x0, y0, x1, y1 := pts[0].X, pts[0].Y, pts[1].X, pts[1].Y
		if x0 > x1 {
			x0, x1 = x1, x0
		}
		if y0 > y1 {
			y0, y1 = y1, y0
		}
		corners = [][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}}
	} else {
		for _, p := range pts {
			corners = append(corners, [2]float64{p.X, p.Y})
		}
	}
	if len(corners) < 2 {
		return nil
	}
	mapPt := func(px, py float64) (float64, float64) {
		return org.X() + u.X()*(px/size.X()) + v.X()*(py/size.Y()),
			org.Y() + u.Y()*(px/size.X()) + v.Y()*(py/size.Y())
	}
	p := path.NewPath()
	x, y := mapPt(corners[0][0], corners[0][1])
	p.MoveTo(x, y)
	for _, c := range corners[1:] {
		x, y = mapPt(c[0], c[1])
		p.LineTo(x, y)
	}
	p.Close()
	return p
}
