package render

import (
	"fmt"
	"image/color"
	"io"
	"math"
	"strings"

	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"

	"bytes"
	"encoding/base64"
	"image"
	"image/png"
)

// SVGRenderer renders styled paths into an SVG document.
type SVGRenderer struct {
	w        io.Writer
	width    float64
	height   float64
	style    Style
	defs     []string
	body     strings.Builder
	clipRef  string
	nextClip int
	nextGrad int
	started  bool
}

// NewSVGRenderer creates an SVG backend. It is registered as "svg".
func NewSVGRenderer(w io.Writer, width, height float64) (Renderer, error) {
	return &SVGRenderer{w: w, width: width, height: height}, nil
}

func init() { Register("svg", NewSVGRenderer) }

// Begin starts the canvas.
func (r *SVGRenderer) Begin(width, height float64) error {
	r.width, r.height = width, height
	r.started = true
	return nil
}

// SetStyle selects the style for subsequent paths.
func (r *SVGRenderer) SetStyle(style Style) error {
	r.style = style
	return nil
}

// Clip restricts drawing to the interior of p; nil resets the clip.
func (r *SVGRenderer) Clip(p *path.Path) error {
	if p == nil {
		r.clipRef = ""
		return nil
	}
	r.nextClip++
	id := fmt.Sprintf("clip%d", r.nextClip)
	r.defs = append(r.defs,
		fmt.Sprintf("<clipPath id=%q><path d=%q/></clipPath>", id, pathData(p)))
	r.clipRef = fmt.Sprintf("url(#%s)", id)
	return nil
}

// FillGradient fills the boundary with a linear gradient along the
// gradient axis across the boundary bounding box.
func (r *SVGRenderer) FillGradient(bnd *path.Path, g *GradientFill) error {
	if !r.started || bnd == nil || g == nil {
		return nil
	}
	lo, hi := bnd.BoundingBox()
	cx, cy := (lo.X()+hi.X())/2, (lo.Y()+hi.Y())/2
	ang := g.Rotation * math.Pi / 180
	ux, uy := math.Cos(ang), math.Sin(ang)
	hw := (math.Abs(ux)*(hi.X()-lo.X()) + math.Abs(uy)*(hi.Y()-lo.Y())) / 2
	if hw <= 0 {
		hw = 1
	}
	r.nextGrad++
	id := fmt.Sprintf("grad%d", r.nextGrad)
	c1 := svgColor(gradientRGBA(g.Color1))
	c2 := svgColor(gradientRGBA(g.Color2))
	r.defs = append(r.defs, fmt.Sprintf(
		`<linearGradient id=%s gradientUnits="userSpaceOnUse" x1=%s y1=%s x2=%s y2=%s>`+
			`<stop offset="0" stop-color=%s/><stop offset="1" stop-color=%s/></linearGradient>`,
		fmt.Sprintf("%q", id), num(cx-ux*hw), num(cy-uy*hw), num(cx+ux*hw), num(cy+uy*hw),
		fmt.Sprintf("%q", c1), fmt.Sprintf("%q", c2)))
	var b strings.Builder
	b.WriteString(`<path d="`)
	b.WriteString(pathData(bnd))
	b.WriteString(`" fill="url(#`)
	b.WriteString(id)
	b.WriteString(`)"`)
	if r.style.Stroke.A > 0 {
		fmt.Fprintf(&b, ` stroke="%s"`, svgColor(r.style.Stroke))
		writeOpacity(&b, "stroke-opacity", r.style.Stroke.A)
		fmt.Fprintf(&b, ` stroke-width="%s"`, num(r.style.Width))
	}
	if r.clipRef != "" {
		fmt.Fprintf(&b, ` clip-path="%s"`, r.clipRef)
	}
	b.WriteString("/>")
	r.body.WriteString(b.String())
	return nil
}

// gradientRGBA maps normalized gradient components to 8-bit color.
func gradientRGBA(c [3]float64) color.RGBA {
	f := func(v float64) uint8 {
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA{f(c[0]), f(c[1]), f(c[2]), 255}
}

// DrawPath emits one styled path element.
func (r *SVGRenderer) DrawPath(p *path.Path) error {
	if p == nil {
		return nil
	}
	var b strings.Builder
	b.WriteString(`<path d="`)
	b.WriteString(pathData(p))
	b.WriteString(`"`)
	if r.style.Fill.A > 0 {
		fmt.Fprintf(&b, ` fill="%s"`, svgColor(r.style.Fill))
		writeOpacity(&b, "fill-opacity", r.style.Fill.A)
		if r.style.FillRule == "evenodd" {
			b.WriteString(` fill-rule="evenodd"`)
		}
	} else {
		b.WriteString(` fill="none"`)
	}
	if r.style.Stroke.A > 0 {
		fmt.Fprintf(&b, ` stroke="%s"`, svgColor(r.style.Stroke))
		writeOpacity(&b, "stroke-opacity", r.style.Stroke.A)
		fmt.Fprintf(&b, ` stroke-width="%s"`, num(r.style.Width))
		if len(r.style.Dash) > 0 {
			dash := make([]string, 0, len(r.style.Dash))
			for _, d := range r.style.Dash {
				dash = append(dash, num(d))
			}
			fmt.Fprintf(&b, ` stroke-dasharray="%s"`, strings.Join(dash, ","))
		}
		if r.style.LineCap != "" && r.style.LineCap != "butt" {
			fmt.Fprintf(&b, ` stroke-linecap=%q`, r.style.LineCap)
		}
		if r.style.LineJoin != "" && r.style.LineJoin != "miter" {
			fmt.Fprintf(&b, ` stroke-linejoin=%q`, r.style.LineJoin)
		}
	}
	if r.clipRef != "" {
		fmt.Fprintf(&b, ` clip-path="%s"`, r.clipRef)
	}
	b.WriteString("/>")
	r.body.WriteString(b.String())
	return nil
}

// End writes the final SVG document.
func (r *SVGRenderer) End() error {
	if !r.started {
		return fmt.Errorf("render: SVG backend: Begin was not called")
	}
	var out strings.Builder
	out.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintf(&out, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s">`,
		num(r.width), num(r.height), num(r.width), num(r.height))
	if len(r.defs) > 0 {
		out.WriteString("<defs>")
		for _, d := range r.defs {
			out.WriteString(d)
		}
		out.WriteString("</defs>")
	}
	out.WriteString(r.body.String())
	out.WriteString("</svg>")
	r.started = false
	_, err := r.w.Write([]byte(out.String()))
	return err
}

func writeOpacity(b *strings.Builder, name string, a uint8) {
	if a < 255 {
		fmt.Fprintf(b, ` %s="%.3f"`, name, float64(a)/255)
	}
}

func svgColor(c color.RGBA) string {
	return fmt.Sprintf("rgb(%d,%d,%d)", c.R, c.G, c.B)
}

// num formats a canvas coordinate with stable precision.
func num(v float64) string {
	s := fmt.Sprintf("%.4f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "-0" {
		s = "0"
	}
	return s
}

// pathData serializes a path into SVG path data with quadratic (Q) and
// cubic (C) curves preserved.
func pathData(p *path.Path) string {
	var b strings.Builder
	for _, cmd := range p.Commands() {
		switch c := cmd.(type) {
		case *path.PathMoveTo:
			fmt.Fprintf(&b, "M%s,%s", num(c.End.X()), num(c.End.Y()))
		case *path.PathLineTo:
			fmt.Fprintf(&b, "L%s,%s", num(c.End.X()), num(c.End.Y()))
		case *path.PathCurve3To:
			fmt.Fprintf(&b, "Q%s,%s,%s,%s",
				num(c.Ctrl.X()), num(c.Ctrl.Y()), num(c.End.X()), num(c.End.Y()))
		case *path.PathCurve4To:
			fmt.Fprintf(&b, "C%s,%s,%s,%s,%s,%s",
				num(c.Ctrl1.X()), num(c.Ctrl1.Y()),
				num(c.Ctrl2.X()), num(c.Ctrl2.Y()),
				num(c.End.X()), num(c.End.Y()))
		}
	}
	return b.String()
}

// flattenPath returns polygonal sub-paths for a path, approximating curves
// with the given maximum deviation. Each sub-slice is a closed vertex loop.
func flattenPath(p *path.Path, deviation float64) [][]dxfmath.Vec2 {
	if deviation <= 0 {
		deviation = 0.5
	}
	var polys [][]dxfmath.Vec2
	var cur []dxfmath.Vec2
	var last dxfmath.Vec2
	flush := func() {
		if len(cur) > 0 {
			polys = append(polys, cur)
		}
		cur = nil
	}
	for _, cmd := range p.Commands() {
		switch c := cmd.(type) {
		case *path.PathMoveTo:
			flush()
			last = c.End
			cur = append(cur, c.End)
		case *path.PathLineTo:
			cur = append(cur, c.End)
			last = c.End
		case *path.PathCurve3To:
			bz, err := dxfmath.NewBezier3P(
				to3(last), to3(c.Ctrl), to3(c.End))
			if err != nil {
				cur = append(cur, c.End)
			} else {
				for _, pt := range bz.Flattening(deviation, 4) {
					cur = append(cur, to2(pt))
				}
			}
			last = c.End
		case *path.PathCurve4To:
			bz, err := dxfmath.NewBezier4P(
				to3(last), to3(c.Ctrl1), to3(c.Ctrl2), to3(c.End))
			if err != nil {
				cur = append(cur, c.End)
			} else {
				for _, pt := range bz.Flattening(deviation, 4) {
					cur = append(cur, to2(pt))
				}
			}
			last = c.End
		}
	}
	flush()
	return polys
}

func to2(v dxfmath.Vec3) dxfmath.Vec2 { return dxfmath.Vec2{v.X(), v.Y()} }
func to3(v dxfmath.Vec2) dxfmath.Vec3 {
	return dxfmath.Vec3{v.X(), v.Y(), 0}
}

// DrawImage embeds a raster image as a base64 PNG, honoring the current clip.
func (r *SVGRenderer) DrawImage(im image.Image, x, y, w, h, rotDeg, opacity float64) error {
	if r == nil || !r.started || im == nil || w <= 0 || h <= 0 || opacity <= 0 {
		return nil
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, im); err != nil {
		return err
	}
	cx, cy := x+w/2, y+h/2
	var b strings.Builder
	b.WriteString("<image ")
	b.WriteString(fmt.Sprintf("x=%q y=%q width=%q height=%q ", num(x), num(y), num(w), num(h)))
	if rotDeg != 0 {
		fmt.Fprintf(&b, "transform=%q ", "rotate("+num(-rotDeg)+" "+num(cx)+" "+num(cy)+")")
	}
	if opacity < 1 {
		b.WriteString(fmt.Sprintf("opacity=%q ", num(opacity)))
	}
	if r.clipRef != "" {
		fmt.Fprintf(&b, "clip-path=%q ", r.clipRef)
	}
	fmt.Fprintf(&b, "href=%q", "data:image/png;base64,"+base64.StdEncoding.EncodeToString(buf.Bytes()))
	b.WriteString("/>")
	r.body.WriteString(b.String())
	return nil
}
