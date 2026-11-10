package render

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"

	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
)

// RasterRenderer renders styled paths into an RGBA image encoded as PNG.
type RasterRenderer struct {
	w       io.Writer
	width   int
	height  int
	aa      bool
	s       int // supersample factor
	bg      color.RGBA
	sup     *image.RGBA
	style   Style
	clips   [][]dxfmath.Vec2 // device-space convex clip polygons
	started bool
}

// RasterOption configures the raster backend.
type RasterOption func(*RasterRenderer)

// WithBackground sets the canvas background color.
func WithBackground(c color.RGBA) RasterOption {
	return func(r *RasterRenderer) { r.bg = c }
}

// WithAntiAliasing enables (true) or disables (false) supersampled
// anti-aliasing. Enabled by default.
func WithAntiAliasing(on bool) RasterOption {
	return func(r *RasterRenderer) { r.aa = on }
}

// NewRasterRenderer creates a raster backend with a pixel canvas.
func NewRasterRenderer(w io.Writer, width, height int, opts ...RasterOption) *RasterRenderer {
	r := &RasterRenderer{w: w, width: width, height: height, aa: true,
		bg: color.RGBA{R: 255, G: 255, B: 255, A: 255}}
	for _, o := range opts {
		o(r)
	}
	return r
}

// NewPNGRenderer adapts the raster backend to the backend registry
// (canvas size in units equals pixels). It is registered as "png".
func NewPNGRenderer(w io.Writer, width, height float64) (Renderer, error) {
	return NewRasterRenderer(w, int(math.Round(width)), int(math.Round(height))), nil
}

func init() { Register("png", NewPNGRenderer) }

// Begin allocates the canvas and paints the background.
func (r *RasterRenderer) Begin(width, height float64) error {
	r.width = int(math.Round(width))
	r.height = int(math.Round(height))
	if r.width <= 0 || r.height <= 0 {
		return nil // deferred Begin error: End reports it
	}
	r.s = 1
	if r.aa {
		r.s = 4
	}
	r.sup = image.NewRGBA(image.Rect(0, 0, r.width*r.s, r.height*r.s))
	for y := 0; y < r.sup.Rect.Dy(); y++ {
		for x := 0; x < r.sup.Rect.Dx(); x++ {
			r.sup.SetRGBA(x, y, r.bg)
		}
	}
	r.clips = nil
	r.started = true
	return nil
}

// SetStyle selects the style for subsequent paths.
func (r *RasterRenderer) SetStyle(style Style) error {
	r.style = style
	return nil
}

// Clip restricts drawing to the interior of the flattened path polygons;
// clip polygons must be convex. nil resets the clip.
func (r *RasterRenderer) Clip(p *path.Path) error {
	if p == nil {
		r.clips = nil
		return nil
	}
	s := float64(r.s)
	for _, poly := range flattenPath(p, 0.5/float64(r.s)) {
		dev := make([]dxfmath.Vec2, 0, len(poly))
		for _, v := range poly {
			dev = append(dev, dxfmath.Vec2{v.X() * s, v.Y() * s})
		}
		if len(dev) > 2 {
			r.clips = append(r.clips, dev)
		}
	}
	return nil
}

// DrawPath fills and strokes the flattened path with coverage sampling.
func (r *RasterRenderer) DrawPath(p *path.Path) error {
	if !r.started {
		return nil
	}
	if p == nil {
		return nil
	}
	s := float64(r.s)
	polys := flattenPath(p, 0.5/s)
	dev := make([][]dxfmath.Vec2, 0, len(polys))
	fillLoops := make([][]dxfmath.Vec2, 0, len(polys))
	for _, poly := range polys {
		loop := make([]dxfmath.Vec2, 0, len(poly))
		for _, v := range poly {
			loop = append(loop, dxfmath.Vec2{v.X() * s, v.Y() * s})
		}
		if len(loop) < 2 {
			continue
		}
		dev = append(dev, loop)
		if len(loop) >= 3 {
			fillLoops = append(fillLoops, loop)
		}
	}
	if r.style.Fill.A > 0 && len(fillLoops) > 0 {
		r.fillPolys(fillLoops, r.style.Fill, r.style.FillRule == "evenodd")
	}
	if r.style.Stroke.A > 0 {
		r.strokePolys(dev, r.style)
	}
	return nil
}

// FillGradient fills the boundary with a linear gradient interpolated
// along the gradient axis across the boundary bounding box.
func (r *RasterRenderer) FillGradient(bnd *path.Path, g *GradientFill) error {
	if !r.started || bnd == nil || g == nil {
		return nil
	}
	s := float64(r.s)
	polys := flattenPath(bnd, 0.5/s)
	dev := make([][]dxfmath.Vec2, 0, len(polys))
	for _, poly := range polys {
		loop := make([]dxfmath.Vec2, 0, len(poly))
		for _, v := range poly {
			loop = append(loop, dxfmath.Vec2{v.X() * s, v.Y() * s})
		}
		if len(loop) >= 3 {
			dev = append(dev, loop)
		}
	}
	if len(dev) == 0 {
		return nil
	}
	lo, hi := bnd.BoundingBox()
	ang := g.Rotation * math.Pi / 180
	ux, uy := math.Cos(ang), math.Sin(ang)
	hw := (math.Abs(ux)*(hi.X()-lo.X()) + math.Abs(uy)*(hi.Y()-lo.Y())) / 2
	if hw <= 0 {
		hw = 1
	}
	cx := (lo.X() + hi.X()) / 2 * s
	cy := (lo.Y() + hi.Y()) / 2 * s
	r.fillPolysGrad(dev, cx, cy, ux, uy, hw*s, gradientRGBA(g.Color1), gradientRGBA(g.Color2))
	return nil
}

// fillPolysGrad scans polygons with even-odd parity, filling each span
// with the gradient color interpolated at every pixel.
func (r *RasterRenderer) fillPolysGrad(polys [][]dxfmath.Vec2, cx, cy, ux, uy, hw float64, c1, c2 color.RGBA) {
	minY, maxY := 1<<30, -(1 << 30)
	for _, poly := range polys {
		for i := range poly {
			for _, y := range []float64{poly[i].Y(), poly[(i+1)%len(poly)].Y()} {
				yi := int(math.Floor(y))
				if yi < minY {
					minY = yi
				}
				if yi+1 > maxY {
					maxY = yi + 1
				}
			}
		}
	}
	if minY < 0 {
		minY = 0
	}
	if maxY > r.sup.Rect.Dy() {
		maxY = r.sup.Rect.Dy()
	}
	for y := minY; y < maxY; y++ {
		scan := float64(y) + 0.5
		var xs []float64
		for _, poly := range polys {
			for i := range poly {
				a, b := poly[i], poly[(i+1)%len(poly)]
				ay, by := a.Y(), b.Y()
				if (ay <= scan && by > scan) || (by <= scan && ay > scan) {
					t := (scan - ay) / (by - ay)
					xs = append(xs, a.X()+(b.X()-a.X())*t)
				}
			}
		}
		if len(xs) == 0 {
			continue
		}
		sortFloats(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			x0, x1 := xs[i], xs[i+1]
			ix0 := int(math.Ceil(x0 - 0.5))
			ix1 := int(math.Floor(x1-0.5)) + 1
			if ix0 < 0 {
				ix0 = 0
			}
			if ix1 > r.sup.Rect.Dx() {
				ix1 = r.sup.Rect.Dx()
			}
			for x := ix0; x < ix1; x++ {
				px := float64(x) + 0.5
				if !r.pointInClips(px, scan) {
					continue
				}
				t := 0.5 + ((px-cx)*ux+(scan-cy)*uy)/hw
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				r.blend(x, y, lerpRGBA(c1, c2, t))
			}
		}
	}
}

// lerpRGBA interpolates two opaque colors.
func lerpRGBA(a, b color.RGBA, t float64) color.RGBA {
	f := func(x, y uint8) uint8 {
		v := float64(x) + (float64(y)-float64(x))*t
		if v < 0.5 {
			v = 0
		}
		if v > 254.5 {
			return 255
		}
		return uint8(v + 0.5)
	}
	return color.RGBA{f(a.R, b.R), f(a.G, b.G), f(a.B, b.B), 255}
}

// End downsamples the supersampled canvas and writes the PNG.
func (r *RasterRenderer) End() error {
	if !r.started {
		return fmt.Errorf("render: raster backend: Begin was not called")
	}
	r.started = false
	img := image.NewRGBA(image.Rect(0, 0, r.width, r.height))
	s := r.s
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			var rs, gs, bs, as uint32
			for dy := 0; dy < s; dy++ {
				for dx := 0; dx < s; dx++ {
					c := r.sup.RGBAAt(x*s+dx, y*s+dy)
					rs += uint32(c.R)
					gs += uint32(c.G)
					bs += uint32(c.B)
					as += uint32(c.A)
				}
			}
			n := uint32(s * s)
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(rs / n), G: uint8(gs / n), B: uint8(bs / n), A: uint8(as / n)})
		}
	}
	return png.Encode(r.w, img)
}

// fillPolys scans all polygons with one even-odd (or winding) parity pass.
func (r *RasterRenderer) fillPolys(polys [][]dxfmath.Vec2, c color.RGBA, evenOdd bool) {
	minY, maxY := 1<<30, -(1 << 30)
	edges := 0
	for _, poly := range polys {
		for i := range poly {
			a, b := poly[i], poly[(i+1)%len(poly)]
			edges++
			for _, y := range []float64{a.Y(), b.Y()} {
				yi := int(math.Floor(y))
				if yi < minY {
					minY = yi
				}
				if yi+1 > maxY {
					maxY = yi + 1
				}
			}
		}
	}
	if minY < 0 {
		minY = 0
	}
	if maxY > r.sup.Rect.Dy() {
		maxY = r.sup.Rect.Dy()
	}
	for y := minY; y < maxY; y++ {
		scan := float64(y) + 0.5
		var xs []float64
		var winds []float64
		for _, poly := range polys {
			for i := range poly {
				a, b := poly[i], poly[(i+1)%len(poly)]
				ay, by := a.Y(), b.Y()
				if (ay <= scan && by > scan) || (by <= scan && ay > scan) {
					t := (scan - ay) / (by - ay)
					x := a.X() + (b.X()-a.X())*t
					xs = append(xs, x)
					if by > ay {
						winds = append(winds, x)
					} else {
						winds = append(winds, -x)
					}
				}
			}
		}
		if len(xs) == 0 {
			continue
		}
		if evenOdd {
			sortFloats(xs)
			for i := 0; i+1 < len(xs); i += 2 {
				r.fillSpan(y, xs[i], xs[i+1], c)
			}
		} else {
			sortFloatsAbs(winds)
			w := 0.0
			prev := 0.0
			for j, v := range winds {
				sign := 1.0
				if v < 0 {
					sign = -1.0
				}
				w += sign
				x := math.Abs(v)
				if w == 0 {
					r.fillSpan(y, math.Abs(prev), x, c)
				}
				_ = j
				prev = v
			}
		}
	}
}

// fillSpan paints the [x0,x1) span on row y, blending source over
// and honoring the clip polygons.
func (r *RasterRenderer) fillSpan(y int, x0, x1 float64, c color.RGBA) {
	ix0 := int(math.Ceil(x0 - 0.5))
	ix1 := int(math.Floor(x1-0.5)) + 1
	if ix0 < 0 {
		ix0 = 0
	}
	if ix1 > r.sup.Rect.Dx() {
		ix1 = r.sup.Rect.Dx()
	}
	scan := float64(y) + 0.5
	for x := ix0; x < ix1; x++ {
		cx := float64(x) + 0.5
		if !r.pointInClips(cx, scan) {
			continue
		}
		r.blend(x, y, c)
	}
}

// strokePolys draws each polyline segment as a width quad, splitting
// runs by the dash pattern.
func (r *RasterRenderer) strokePolys(polys [][]dxfmath.Vec2, st Style) {
	w := st.Width * float64(r.s)
	if w < 1 {
		w = 1
	}
	for _, poly := range polys {
		segments := dashSegments(poly, st.Dash)
		for _, seg := range segments {
			quad := segmentQuad(seg[0], seg[1], w)
			for _, clip := range r.clips {
				quad = dxfmath.ClipPolygonToConvex(quad, clip)
				if len(quad) == 0 {
					break
				}
			}
			if len(quad) >= 3 {
				r.fillPolys([][]dxfmath.Vec2{quad}, st.Stroke, true)
			}
		}
	}
}

// pointInClips tests the device point against all convex clip polygons.
func (r *RasterRenderer) pointInClips(x, y float64) bool {
	p := dxfmath.Vec2{x, y}
	for _, clip := range r.clips {
		if !pointInConvex(p, clip) {
			return false
		}
	}
	return true
}

// blend composites src over the destination pixel.
func (r *RasterRenderer) blend(x, y int, src color.RGBA) {
	a := float64(src.A) / 255
	if a == 0 {
		return
	}
	dst := r.sup.RGBAAt(x, y)
	out := color.RGBA{
		R: uint8(float64(src.R)*a + float64(dst.R)*(1-a)),
		G: uint8(float64(src.G)*a + float64(dst.G)*(1-a)),
		B: uint8(float64(src.B)*a + float64(dst.B)*(1-a)),
		A: 255,
	}
	if out.R == dst.R && out.G == dst.G && out.B == dst.B {
		return
	}
	r.sup.SetRGBA(x, y, out)
}

// segmentQuad expands a segment to a rectangle of the given width.
func segmentQuad(a, b dxfmath.Vec2, width float64) []dxfmath.Vec2 {
	dx, dy := b.X()-a.X(), b.Y()-a.Y()
	l := math.Hypot(dx, dy)
	if l == 0 {
		h := width / 2
		return []dxfmath.Vec2{{a.X() - h, a.Y() - h}, {a.X() + h, a.Y() - h},
			{a.X() + h, a.Y() + h}, {a.X() - h, a.Y() + h}}
	}
	nx, ny := -dy/l*width/2, dx/l*width/2
	return []dxfmath.Vec2{
		{a.X() + nx, a.Y() + ny}, {b.X() + nx, b.Y() + ny},
		{b.X() - nx, b.Y() - ny}, {a.X() - nx, a.Y() - ny},
	}
}

// dashSegments cuts a closed polyline into drawn segments per dash cycle.
func dashSegments(poly []dxfmath.Vec2, dash []float64) [][2]dxfmath.Vec2 {
	if len(poly) < 2 {
		return nil
	}
	if len(dash) == 0 {
		var out [][2]dxfmath.Vec2
		for i := 0; i+1 < len(poly); i++ {
			out = append(out, [2]dxfmath.Vec2{poly[i], poly[i+1]})
		}
		return out
	}
	var out [][2]dxfmath.Vec2
	rem := math.Abs(dash[0])
	drawing := dash[0] >= 0
	di := 0
	for i := 0; i+1 < len(poly); i++ {
		a, b := poly[i], poly[i+1]
		segLen := math.Hypot(b.X()-a.X(), b.Y()-a.Y())
		if segLen == 0 {
			continue
		}
		ux, uy := (b.X()-a.X())/segLen, (b.Y()-a.Y())/segLen
		pos := 0.0
		for pos < segLen {
			take := math.Min(rem, segLen-pos)
			if drawing {
				p0 := dxfmath.Vec2{a.X() + ux*pos, a.Y() + uy*pos}
				p1 := dxfmath.Vec2{a.X() + ux*(pos+take), a.Y() + uy*(pos+take)}
				out = append(out, [2]dxfmath.Vec2{p0, p1})
			}
			rem -= take
			pos += take
			if rem <= 1e-9 {
				di = (di + 1) % len(dash)
				drawing = dash[di] >= 0
				rem = math.Abs(dash[di])
				if rem == 0 {
					rem = 1e-9 // dot
				}
			}
		}
	}
	return out
}

// pointInConvex tests p against a convex polygon (consistent winding).
func pointInConvex(p dxfmath.Vec2, poly []dxfmath.Vec2) bool {
	pos, neg := 0, 0
	for i := range poly {
		a, b := poly[i], poly[(i+1)%len(poly)]
		cross := (b.X()-a.X())*(p.Y()-a.Y()) - (b.Y()-a.Y())*(p.X()-a.X())
		if cross > 1e-9 {
			pos++
		} else if cross < -1e-9 {
			neg++
		}
	}
	return pos == 0 || neg == 0
}

func sortFloats(xs []float64) {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
}

func sortFloatsAbs(xs []float64) {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && math.Abs(xs[j]) < math.Abs(xs[j-1]); j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
}

// DrawImage blits a raster image into the supersampled canvas with
// inverse-affine sampling, respecting the current clip and opacity.
func (r *RasterRenderer) DrawImage(im image.Image, x, y, w, h, rotDeg, opacity float64) error {
	if !r.started || im == nil || w <= 0 || h <= 0 || opacity <= 0 {
		return nil
	}
	if opacity > 1 {
		opacity = 1
	}
	sb := im.Bounds()
	sw, sh := float64(sb.Dx()), float64(sb.Dy())
	if sw <= 0 || sh <= 0 {
		return nil
	}
	rad := rotDeg * math.Pi / 180
	cos, sin := math.Cos(rad), math.Sin(rad)
	cx, cy := x+w/2, y+h/2
	x0, x1 := int(math.Floor(x*float64(r.s))), int(math.Ceil((x+w)*float64(r.s)))
	y0, y1 := int(math.Floor(y*float64(r.s))), int(math.Ceil((y+h)*float64(r.s)))
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if px < 0 || py < 0 || px >= r.sup.Bounds().Max.X || py >= r.sup.Bounds().Max.Y {
				continue
			}
			if !r.pointInClips(float64(px), float64(py)) {
				continue
			}
			dx := (float64(px)+0.5)/float64(r.s) - cx
			dy := (float64(py)+0.5)/float64(r.s) - cy
			rx := dx*cos + dy*sin
			ry := -dx*sin + dy*cos
			ix := sb.Min.X + int(math.Floor(rx/w*sw+sw/2))
			iy := sb.Min.Y + int(math.Floor(ry/h*sh+sh/2))
			if ix < sb.Min.X || iy < sb.Min.Y || ix >= sb.Max.X || iy >= sb.Max.Y {
				continue
			}
			c := im.At(ix, iy)
			rr, gg, bb, aa := c.RGBA()
			if aa == 0 {
				continue
			}
			src := color.RGBA{uint8(rr >> 8), uint8(gg >> 8), uint8(bb >> 8), uint8(float64(uint8(aa>>8)) * opacity)}
			r.blend(px, py, src)
		}
	}
	return nil
}
