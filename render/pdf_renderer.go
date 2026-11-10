package render

import (
	"fmt"
	"image/color"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
)

// PDFRenderer renders styled paths into a single-page PDF document.
// Canvas units map one-to-one to PDF points; the page coordinate system
// is flipped so canvas coordinates keep a top-left origin.
type PDFRenderer struct {
	w       io.Writer
	width   float64
	height  float64
	style   Style
	content strings.Builder
	depth   int // q/Q nesting for clips
	gstates map[[2]uint8]int
	gsNames []string
	started bool
	clipped bool
}

// NewPDFRenderer creates a PDF backend. It is registered as "pdf".
func NewPDFRenderer(w io.Writer, width, height float64) (Renderer, error) {
	return &PDFRenderer{w: w, width: width, height: height}, nil
}

func init() { Register("pdf", NewPDFRenderer) }

// Begin starts the canvas.
func (r *PDFRenderer) Begin(width, height float64) error {
	r.width, r.height = width, height
	r.content.Reset()
	r.depth = 0
	r.gsNames = nil
	r.gstates = map[[2]uint8]int{}
	r.content.WriteString("q 1 0 0 -1 0 ")
	r.content.WriteString(pdfNum(height))
	r.content.WriteString(" cm\n")
	r.started = true
	return nil
}

// SetStyle selects the style for subsequent paths.
func (r *PDFRenderer) SetStyle(style Style) error {
	r.style = style
	return nil
}

// Clip restricts drawing to the interior of p; nil resets the clip.
func (r *PDFRenderer) Clip(p *path.Path) error {
	if p == nil {
		for r.depth > 0 {
			r.content.WriteString("Q\n")
			r.depth--
		}
		return nil
	}
	r.content.WriteString("q\n")
	r.depth++
	r.content.WriteString(pathOps(p))
	r.content.WriteString(" W n\n")
	return nil
}

// DrawPath fills and strokes the path per the current style.
func (r *PDFRenderer) DrawPath(p *path.Path) error {
	if p == nil {
		return nil
	}
	ops := pathOps(p)
	st, fi := r.style.Stroke, r.style.Fill
	if fi.A > 0 {
		r.useGState(st.A, fi.A)
		r.content.WriteString(pdfColor(fi) + " rg\n")
		r.content.WriteString(ops)
		if r.style.FillRule == "evenodd" {
			r.content.WriteString("f*\n")
		} else {
			r.content.WriteString("f\n")
		}
	}
	if st.A > 0 {
		r.useGState(st.A, fi.A)
		r.content.WriteString(pdfColor(st) + " RG\n")
		r.content.WriteString(pdfNum(r.style.Width) + " w\n")
		if len(r.style.Dash) > 0 {
			r.content.WriteString(pdfDash(r.style.Dash))
		} else {
			r.content.WriteString("[] 0 d\n")
		}
		r.content.WriteString(pdfCapJoin(r.style) + "\n")
		r.content.WriteString(ops)
		r.content.WriteString("S\n")
	}
	return nil
}

// FillGradient fills the boundary with the gradient's first color.
// Full PDF shading patterns are out of scope (minimal gradient support).
func (r *PDFRenderer) FillGradient(bnd *path.Path, g *GradientFill) error {
	if !r.started || bnd == nil || g == nil {
		return nil
	}
	c1 := gradientRGBA(g.Color1)
	r.useGState(255, 255)
	r.content.WriteString(pdfColor(c1) + " rg\n")
	r.content.WriteString(pathOps(bnd))
	r.content.WriteString("f\n")
	return nil
}

// useGState records the (stroke, fill) alpha pair and emits the gs operator.
func (r *PDFRenderer) useGState(strokeA, fillA uint8) {
	if strokeA == 255 && fillA == 255 {
		r.content.WriteString("/GS0 gs\n")
		return
	}
	key := [2]uint8{strokeA, fillA}
	if _, ok := r.gstates[key]; !ok {
		r.gstates[key] = len(r.gsNames) + 1
		r.gsNames = append(r.gsNames,
			fmt.Sprintf("/GS%d %.3f %.3f", len(r.gsNames)+1,
				float64(strokeA)/255, float64(fillA)/255))
	}
	fmt.Fprintf(&r.content, "/GS%d gs\n", r.gstates[key])
}

// End writes the complete PDF file.
func (r *PDFRenderer) End() error {
	if !r.started {
		return fmt.Errorf("render: PDF backend: Begin was not called")
	}
	for r.depth > 0 {
		r.content.WriteString("Q\n")
		r.depth--
	}
	r.content.WriteString("Q\n")
	body := r.content.String()

	// Object layout: 1 catalog, 2 pages, 3 page, 4 content stream,
	// 5.. ExtGState dictionaries (GS0 opaque plus recorded alphas).
	objs := make([]string, 0, 5+len(r.gsNames))
	objs = append(objs,
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %s %s] /Contents 4 0 R /Resources << /ExtGState %s >> >>",
			pdfNum(r.width), pdfNum(r.height), r.extGStateDict()),
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(body), body),
		"<< /Type /ExtGState /CA 1.0 /ca 1.0 >>", // GS0
	)
	for _, spec := range r.gsNames {
		parts := strings.Fields(spec) // /GSn CA ca
		objs = append(objs, fmt.Sprintf("<< /Type /ExtGState /CA %s /ca %s >>", parts[1], parts[2]))
	}

	var out strings.Builder
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objs)+1)
	for i, obj := range objs {
		offsets[i+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(objs)+1)
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objs); i++ {
		fmt.Fprintf(&out, "%010d 00000 n \n", offsets[i])
	}
	out.WriteString("trailer\n")
	fmt.Fprintf(&out, "<< /Size %d /Root 1 0 R >>\n", len(objs)+1)
	out.WriteString("startxref\n")
	fmt.Fprintf(&out, "%d\n%%%%EOF\n", xref)
	r.started = false
	_, err := r.w.Write([]byte(out.String()))
	return err
}

func (r *PDFRenderer) extGStateDict() string {
	entries := []string{"/GS0 5 0 R"}
	for i := range r.gsNames {
		entries = append(entries, fmt.Sprintf("/GS%d %d 0 R", i+1, 6+i))
	}
	sort.Strings(entries)
	return "<< " + strings.Join(entries, " ") + " >>"
}

// pathOps serializes a path to PDF path construction operators.
// Quadratic curves are promoted to cubic ones.
func pathOps(p *path.Path) string {
	var b strings.Builder
	var last, first dxfmath.Vec2
	for _, cmd := range p.Commands() {
		switch c := cmd.(type) {
		case *path.PathMoveTo:
			fmt.Fprintf(&b, "%s %s m\n", pdfNum(c.End.X()), pdfNum(c.End.Y()))
			last, first = c.End, c.End
		case *path.PathLineTo:
			fmt.Fprintf(&b, "%s %s l\n", pdfNum(c.End.X()), pdfNum(c.End.Y()))
			last = c.End
		case *path.PathCurve3To:
			c1 := lerp2(last, c.Ctrl, 2.0/3)
			c2 := lerp2(c.End, c.Ctrl, 2.0/3)
			fmt.Fprintf(&b, "%s %s %s %s %s %s c\n",
				pdfNum(c1.X()), pdfNum(c1.Y()), pdfNum(c2.X()), pdfNum(c2.Y()),
				pdfNum(c.End.X()), pdfNum(c.End.Y()))
			last = c.End
		case *path.PathCurve4To:
			fmt.Fprintf(&b, "%s %s %s %s %s %s c\n",
				pdfNum(c.Ctrl1.X()), pdfNum(c.Ctrl1.Y()),
				pdfNum(c.Ctrl2.X()), pdfNum(c.Ctrl2.Y()),
				pdfNum(c.End.X()), pdfNum(c.End.Y()))
			last = c.End
		}
	}
	// PDF subpaths auto-close for fills from the current point to subpath
	// start; the scene closes subpaths explicitly with a line back.
	_ = first
	return b.String()
}

func lerp2(a, b dxfmath.Vec2, t float64) dxfmath.Vec2 {
	return dxfmath.Vec2{a.X() + (b.X()-a.X())*t, a.Y() + (b.Y()-a.Y())*t}
}

func pdfColor(c color.RGBA) string {
	return fmt.Sprintf("%.3f %.3f %.3f", float64(c.R)/255, float64(c.G)/255, float64(c.B)/255)
}

func pdfDash(dash []float64) string {
	parts := make([]string, 0, len(dash))
	for _, d := range dash {
		parts = append(parts, pdfNum(d))
	}
	return "[" + strings.Join(parts, " ") + "] 0 d\n"
}

func pdfCapJoin(s Style) string {
	cap := 0
	switch s.LineCap {
	case "round":
		cap = 1
	case "square":
		cap = 2
	}
	join := 0
	switch s.LineJoin {
	case "round":
		join = 1
	case "bevel":
		join = 2
	}
	return fmt.Sprintf("%d J %d j", cap, join)
}

func pdfNum(v float64) string {
	if math.Abs(v) < 1e-9 {
		v = 0
	}
	return strconv.FormatFloat(v, 'f', 4, 64)
}
