package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	mathx "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
	"github.com/edanko/dxf/text2path"
)

func rgba(r, g, b, a uint8) color.RGBA { return color.RGBA{R: r, G: g, B: b, A: a} }

type stubBackend struct{ began bool }

func (s *stubBackend) Begin(width, height float64) error { s.began = true; return nil }
func (s *stubBackend) SetStyle(st Style) error           { return nil }
func (s *stubBackend) DrawPath(p *path.Path) error       { return nil }
func (s *stubBackend) Clip(p *path.Path) error           { return nil }
func (s *stubBackend) DrawImage(im image.Image, x, y, w, h, rotDeg, opacity float64) error {
	return nil
}
func (s *stubBackend) FillGradient(b *path.Path, g *GradientFill) error { return nil }
func (s *stubBackend) End() error                                       { return nil }

func TestBackendRegistry(t *testing.T) {
	for _, name := range []string{"svg", "pdf", "png"} {
		var buf bytes.Buffer
		r, err := NewBackend(name, &buf, 800, 600)
		if err != nil {
			t.Fatalf("NewBackend(%q): %v", name, err)
		}
		if r == nil {
			t.Fatalf("NewBackend(%q): nil renderer", name)
		}
	}
	if _, err := NewBackend("nope", &bytes.Buffer{}, 10, 10); err == nil {
		t.Fatal("unknown backend must error")
	}
	Register("unit-test-backend", func(w io.Writer, width, height float64) (Renderer, error) {
		return &stubBackend{}, nil
	})
	found := false
	for _, n := range BackendNames() {
		if n == "unit-test-backend" {
			found = true
		}
	}
	if !found {
		t.Fatalf("BackendNames: %v", BackendNames())
	}
	var buf bytes.Buffer
	r, err := NewBackend("unit-test-backend", &buf, 10, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(*stubBackend); !ok {
		t.Fatal("dispatch returned wrong type")
	}
}

func rectPath(x0, y0, x1, y1 float64) *path.Path {
	p := path.NewPath()
	p.MoveTo(x0, y0)
	p.LineTo(x1, y0)
	p.LineTo(x1, y1)
	p.LineTo(x0, y1)
	p.LineTo(x0, y0)
	return p
}

func linePath(x0, y0, x1, y1 float64) *path.Path {
	p := path.NewPath()
	p.MoveTo(x0, y0)
	p.LineTo(x1, y1)
	return p
}

func newSVG(t *testing.T, buf *bytes.Buffer, w, h float64) Renderer {
	t.Helper()
	r, err := NewSVGRenderer(buf, w, h)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Begin(w, h); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestSVGPathGeometryAndStyle(t *testing.T) {
	var buf bytes.Buffer
	r := newSVG(t, &buf, 800, 600)
	st := Style{
		Stroke:   rgba(255, 0, 0, 255),
		Fill:     rgba(0, 0, 255, 128),
		Width:    4,
		Dash:     []float64{10, 5},
		FillRule: "evenodd",
	}
	if err := r.SetStyle(st); err != nil {
		t.Fatal(err)
	}
	p := path.NewPath()
	p.MoveTo(0, 0)
	p.LineTo(100, 0)
	p.Curve3To(120, 20, 110, 10)
	p.Curve4To(140, 0, 125, -10, 135, 10)
	if err := r.DrawPath(p); err != nil {
		t.Fatal(err)
	}
	if err := r.Clip(rectPath(10, 10, 50, 50)); err != nil {
		t.Fatal(err)
	}
	q := path.NewPath()
	q.MoveTo(0, 0)
	q.LineTo(10, 10)
	q.LineTo(0, 10)
	q.LineTo(0, 0)
	if err := r.DrawPath(q); err != nil {
		t.Fatal(err)
	}
	if err := r.Clip(nil); err != nil {
		t.Fatal(err)
	}
	if err := r.End(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"<svg", "width=\"800\"", "viewBox",
		"stroke=\"rgb(255,0,0)\"", "stroke-width=\"4\"",
		"stroke-dasharray=\"10,5\"",
		"fill=\"rgb(0,0,255)\"", "fill-opacity=\"0.502\"",
		"fill-rule=\"evenodd\"",
		"<clipPath", "clip-path=\"url(#clip1)\"",
		"M0,0", "L100,0", "Q110,10,120,20", "C125,-10,135,10,140,0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("svg output missing %q\n%s", want, out)
		}
	}
	if strings.Count(out, `stroke="rgb(255,0,0)"`) != 2 {
		t.Errorf("want 2 stroked paths, got %d", strings.Count(out, `stroke="rgb(255,0,0)"`))
	}
}

func TestSVGFlattensArcs(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddCircle(0, 0, 0, 10); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	r := newSVG(t, &buf, 800, 600)
	for _, sp := range sc.Paths {
		if err := r.SetStyle(sp.Style); err != nil {
			t.Fatal(err)
		}
		if err := r.DrawPath(sp.Path); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.End(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if n := strings.Count(out, "L"); n < 8 {
		t.Fatalf("circle must render as a flattened polygon (>=8 L segments), got %d:\n%s", n, out)
	}
	if strings.Contains(out, " A ") {
		t.Fatal("arcs must be flattened, not emitted as elliptical arcs")
	}
}

func TestPDFStructure(t *testing.T) {
	var buf bytes.Buffer
	r, err := NewPDFRenderer(&buf, 300, 200)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Begin(300, 200); err != nil {
		t.Fatal(err)
	}
	if err := r.SetStyle(Style{Stroke: rgba(0, 0, 0, 255), Width: 2}); err != nil {
		t.Fatal(err)
	}
	if err := r.Clip(rectPath(10, 10, 200, 150)); err != nil {
		t.Fatal(err)
	}
	if err := r.DrawPath(linePath(0, 100, 300, 100)); err != nil {
		t.Fatal(err)
	}
	if err := r.Clip(nil); err != nil {
		t.Fatal(err)
	}
	if err := r.SetStyle(Style{Fill: rgba(255, 0, 0, 100)}); err != nil {
		t.Fatal(err)
	}
	if err := r.DrawPath(rectPath(50, 50, 100, 100)); err != nil {
		t.Fatal(err)
	}
	if err := r.End(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "%PDF-1.4") {
		t.Fatalf("bad header: %.20s", out)
	}
	if !strings.Contains(out, "startxref") || !strings.HasSuffix(strings.TrimSpace(out), "%%EOF") {
		t.Fatal("missing startxref/EOF")
	}
	si := strings.LastIndex(out, "startxref")
	var pos int
	if _, err := fmt.Sscanf(strings.TrimSpace(out[si+len("startxref"):]), "%d", &pos); err != nil {
		t.Fatal(err)
	}
	xref := out[pos:]
	if !strings.HasPrefix(xref, "xref\n") {
		t.Fatalf("xref not at startxref position:\n%.40s", xref)
	}
	lines := strings.Split(xref, "\n")
	var count int
	if _, err := fmt.Sscanf(lines[1], "0 %d", &count); err != nil {
		t.Fatal(err)
	}
	if len(lines) < 2+count {
		t.Fatalf("xref truncated: %d lines", len(lines))
	}
	for i := 1; i < count; i++ {
		var off, gen int
		var typ string
		if _, err := fmt.Sscanf(lines[2+i], "%d %d %s", &off, &gen, &typ); err != nil {
			t.Fatalf("entry %d: %v", i, err)
		}
		if typ != "n" {
			t.Fatalf("entry %d type %q, want n", i, typ)
		}
		want := fmt.Sprintf("%d 0 obj", i)
		if !strings.HasPrefix(out[off:], want) {
			t.Fatalf("offset %d does not point at %q", off, want)
		}
	}
	content := contentBetween(out, "stream\n", "\nendstream")
	for _, want := range []string{" m\n", " l\n", "S\n", "f\n", "W n"} {
		if !strings.Contains(content, want) {
			t.Errorf("pdf content missing %q", want)
		}
	}
	if !strings.Contains(out, "/ExtGState") {
		t.Fatal("pdf missing /ExtGState resources")
	}
}

func contentBetween(out, start, end string) string {
	_, after, ok := strings.Cut(out, start)
	if !ok {
		return ""
	}
	rest := after
	before0, _, ok0 := strings.Cut(rest, end)
	if !ok0 {
		return ""
	}
	return before0
}

func decodePNG(t *testing.T, buf *bytes.Buffer) (int, int, func(x, y int) (uint8, uint8, uint8)) {
	t.Helper()
	img, err := png.Decode(buf)
	if err != nil {
		t.Fatalf("png decode: %v", err)
	}
	b := img.Bounds()
	return b.Dx(), b.Dy(), func(x, y int) (uint8, uint8, uint8) {
		r, g, bb, _ := img.At(x, y).RGBA()
		return uint8(r >> 8), uint8(g >> 8), uint8(bb >> 8)
	}
}

func renderScene(r Renderer, sc *StyledScene) error {
	if err := r.Begin(800, 600); err != nil {
		return err
	}
	for _, sp := range sc.Paths {
		if err := r.SetStyle(sp.Style); err != nil {
			return err
		}
		if sp.FillHatch != nil && sp.FillHatch.Gradient != nil {
			if err := r.FillGradient(sp.Path, sp.FillHatch.Gradient); err != nil {
				return err
			}
			continue
		}
		if err := r.DrawPath(sp.Path); err != nil {
			return err
		}
	}
	return r.End()
}

func TestPNGRenderedViaRegistry(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 100, 0, 100, 100, 0, d.WithLayer("W"), drawing.WithColor(1)); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	r, err := NewBackend("png", &buf, 800, 600)
	if err != nil {
		t.Fatal(err)
	}
	if err := renderScene(r, sc); err != nil {
		t.Fatal(err)
	}
	w, h, at := decodePNG(t, &buf)
	if w != 800 || h != 600 {
		t.Fatalf("png dims %dx%d", w, h)
	}
	rr, gg, bb := at(400, 300)
	if !(rr > 150 && gg < 200 && gg < rr-30 && bb < rr-30) {
		t.Fatalf("center pixel (%d,%d,%d) not red line", rr, gg, bb)
	}
}

func TestRasterStyleColorWidthDash(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 100, 0, 200, 100, 0, drawing.WithColor(1)); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	sc.Paths[0].Style.Width = 10
	sc.Paths[0].Style.Dash = []float64{10, -5}
	var buf bytes.Buffer
	r := NewRasterRenderer(&buf, 800, 600, WithAntiAliasing(false))
	if err := renderScene(r, sc); err != nil {
		t.Fatal(err)
	}
	_, _, at := decodePNG(t, &buf)
	s := sc.Scale
	tx := (800-(sc.ExtMaxX-sc.ExtMinX)*s)/2 - sc.ExtMinX*s
	ty := (600+(sc.ExtMaxY-sc.ExtMinY)*s)/2 + sc.ExtMaxY*s
	start := int(math.Round(tx))
	y := int(math.Round(ty - s*100))
	// dash 10 on, 5 off (canvas units, no AA): on at +5, gap at +12
	r0, g0, b0 := at(start+5, y)
	if !(r0 > 150 && g0 < 120 && b0 < 120) {
		t.Fatalf("dash ON pixel (%d,%d)=(%d,%d,%d) not red", start+5, y, r0, g0, b0)
	}
	r1, g1, b1 := at(start+12, y)
	if !(r1 > 200 && g1 > 200 && b1 > 200) {
		t.Fatalf("dash gap pixel (%d,%d)=(%d,%d,%d) not background", start+12, y, r1, g1, b1)
	}
	red := 0
	for dy := -8; dy <= 8; dy++ {
		rr, gg, bb := at(start+5, y+dy)
		if rr > 150 && gg < 150 && bb < 150 {
			red++
		}
	}
	if red < 6 {
		t.Fatalf("fat dash column red coverage %d rows, want >= 6", red)
	}
}

func TestRasterClip(t *testing.T) {
	var buf bytes.Buffer
	r := NewRasterRenderer(&buf, 100, 100)
	if err := r.Begin(100, 100); err != nil {
		t.Fatal(err)
	}
	if err := r.Clip(rectPath(40, 40, 60, 60)); err != nil {
		t.Fatal(err)
	}
	if err := r.SetStyle(Style{Stroke: rgba(0, 0, 0, 255), Width: 4}); err != nil {
		t.Fatal(err)
	}
	if err := r.DrawPath(linePath(0, 50, 100, 50)); err != nil {
		t.Fatal(err)
	}
	if err := r.End(); err != nil {
		t.Fatal(err)
	}
	_, _, at := decodePNG(t, &buf)
	if rr, gg, bb := at(50, 50); !(rr < 100 && gg < 100 && bb < 100) {
		t.Fatalf("inside clip pixel (%d,%d,%d) not black", rr, gg, bb)
	}
	if rr, gg, bb := at(20, 50); !(rr > 200 && gg > 200 && bb > 200) {
		t.Fatalf("outside clip pixel (%d,%d,%d) not background", rr, gg, bb)
	}
}

func TestBuildSceneStyledEntities(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 0, 0, d.WithLayer("LRED"), drawing.WithColor(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddCircle(0, 0, 0, 5); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddText(0, 0, 0, 10, "HELLO"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 20, 0, 10, 20, 0, WithInvisibleV(true)); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Paths) < 3 {
		t.Fatalf("paths %d (line, circle, text glyphs expected; invisible excluded)", len(sc.Paths))
	}
	red := false
	for _, sp := range sc.Paths {
		if sp.Style.Stroke.R == 255 && sp.Style.Stroke.G == 0 && sp.Style.Stroke.B == 0 {
			red = true
		}
		if sp.Style.Stroke.A == 0 && sp.Style.Fill.A == 0 {
			t.Fatal("drawn path has no paint")
		}
	}
	if !red {
		t.Fatal("ACI 1 did not resolve to red")
	}
	if _, err := d.Model().AddLine(0, 30, 0, 10, 30, 0, d.WithLayer("FROZEN")); err != nil {
		t.Fatal(err)
	}
	if l, ok := d.Layers["FROZEN"]; ok {
		l.SetFlag(1)
	}
	sc2, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc2.Paths) != len(sc.Paths) {
		t.Fatalf("frozen layer geometry included: %d vs %d", len(sc2.Paths), len(sc.Paths))
	}
	if _, err := d.Model().AddLine(0, 40, 0, 10, 40, 0, d.WithLayer("TC"), drawing.WithColor(256)); err != nil {
		t.Fatal(err)
	}
	if l, ok := d.Layers["TC"]; ok {
		l.SetTrueColor(0x00FF00)
	}
	sc3, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	green := false
	for _, sp := range sc3.Paths {
		if sp.Style.Stroke.G == 255 && sp.Style.Stroke.R == 0 {
			green = true
		}
	}
	if !green {
		t.Fatal("layer TrueColor not resolved")
	}
}

func TestSceneHatchFill(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	he, err := d.Model().AddHatch(d.WithLayer("HATCH"))
	if err != nil {
		t.Fatal(err)
	}
	he.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true,
		Points: [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}}
	he.SetSolidFill(true)
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	var solid bool
	for _, sp := range sc.Paths {
		if sp.Style.Fill.A > 0 {
			solid = true
		}
	}
	if !solid {
		t.Fatal("solid hatch must fill")
	}
	d2, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	he2, err := d2.Model().AddHatch()
	if err != nil {
		t.Fatal(err)
	}
	he2.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true,
		Points: [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}}
	he2.SetPredefinedPattern("ANSI31", 0, 1, 0, false)
	sc2, err := BuildScene(d2, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	var pat bool
	for _, sp := range sc2.Paths {
		if sp.FillHatch != nil && sp.FillHatch.Pattern != nil &&
			sp.FillHatch.Pattern.Name == "ANSI31" && len(sp.FillHatch.Pattern.Lines) > 0 {
			pat = true
		}
	}
	if !pat {
		t.Fatal("ANSI31 hatch fill definition missing")
	}
}

func TestTextOutlines(t *testing.T) {
	p := TextToPath("A", 10, 0, 0, 0)
	if p == nil || !p.HasLines() {
		t.Fatal("text produced no glyph strokes")
	}
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddMText(0, 0, 0, 20, 10, "AAAA\nBBBB"); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	var multi bool
	for _, sp := range sc.Paths {
		if sp.Path.HasSubPaths() {
			multi = true
		}
	}
	if !multi {
		t.Fatal("two-line MText must produce subpaths")
	}
}

func TestPathDataAndFlatten(t *testing.T) {
	p := path.NewPath()
	p.MoveTo(0, 0)
	p.LineTo(5, 0)
	p.Curve3To(10, 5, 7, 2)
	p.Curve4To(15, 0, 12, 3, 13, -3)
	d := pathData(p)
	if !strings.HasPrefix(d, "M0,0L5,0Q") || !strings.Contains(d, "C12,3,13,-3,15,0") {
		t.Fatalf("path data: %q", d)
	}
	poly := flattenPath(p, 1e-4)
	if len(poly) != 1 {
		t.Fatalf("subpaths %d", len(poly))
	}
	if len(poly[0]) < 10 {
		t.Fatalf("flattened points %d", len(poly[0]))
	}
	c := path.NewPath()
	c.MoveTo(10, 0)
	c.Curve4To(-10, 0, 10, 12.5, -10, 12.5)
	if !c.HasCurves() {
		t.Fatal("setup: curves expected")
	}
	if len(flattenPath(c, 1e-4)[0]) < 12 {
		t.Fatal("curve flattening too coarse")
	}
}

func TestConformanceAllBackends(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 100, 0, 0, drawing.WithColor(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddCircle(50, 50, 0, 20, drawing.WithColor(3)); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddText(0, 60, 0, 10, "AB"); err != nil {
		t.Fatal(err)
	}
	he3, err := d.Model().AddHatch()
	if err != nil {
		t.Fatal(err)
	}
	he3.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true,
		Points: [][]float64{{80, 0}, {100, 0}, {100, 20}, {80, 20}}}}
	he3.SetSolidFill(true)
	he4, err := d.Model().AddHatch()
	if err != nil {
		t.Fatal(err)
	}
	he4.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true,
		Points: [][]float64{{130, 0}, {150, 0}, {150, 10}, {130, 10}}}}
	he4.HasGradient = true
	he4.Gradient = &entity.HatchGradient{Color1: 1, Color2: 5, Angle: 0}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range BackendNames() {
		if name == "unit-test-backend" {
			continue
		}
		var buf bytes.Buffer
		r, err := NewBackend(name, &buf, 800, 600)
		if err != nil {
			t.Fatal(err)
		}
		if err := renderScene(r, sc); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		out := buf.String()
		if len(out) == 0 {
			t.Fatalf("%s: empty output", name)
		}
		switch name {
		case "svg":
			for _, want := range []string{"<path", "stroke=", "rgb(255,0,0)", `d="M`, "linearGradient"} {
				if !strings.Contains(out, want) {
					t.Errorf("svg missing %q", want)
				}
			}
		case "pdf":
			for _, want := range []string{"%PDF", "startxref", "%%EOF", "/ShadingType 2", "/Sh1 sh"} {
				if !strings.Contains(out, want) {
					t.Errorf("pdf missing %q", want)
				}
			}
		case "png":
			w, h, at := decodePNG(t, &buf)
			if w != 800 || h != 600 {
				t.Fatalf("png %dx%d", w, h)
			}
			found := false
			for y := 0; y < h && !found; y++ {
				for x := 0; x < w && !found; x++ {
					rr, gg, bb := at(x, y)
					if rr > 150 && gg < 200 && gg < rr-30 && bb < rr-30 {
						found = true
					}
				}
			}
			if !found {
				t.Fatal("png: no red line pixel")
			}
			blue := false
			for y := 0; y < h && !blue; y++ {
				for x := 0; x < w && !blue; x++ {
					_, gg, bb := at(x, y)
					if bb > 150 && gg < 150 && bb > gg {
						blue = true
					}
				}
			}
			if !blue {
				t.Fatal("png: no blue gradient pixel")
			}
		}
	}
}

var _ image.Image = (*image.RGBA)(nil)

func WithInvisibleV(on bool) drawing.EntityOption {
	return func(e entity.Entity) error {
		if setter, ok := e.(interface{ SetInvisible(bool) }); ok {
			setter.SetInvisible(on)
			return nil
		}
		return fmt.Errorf("no SetInvisible on %T", e)
	}
}

func TestFillGradientBackends(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	he, err := d.Model().AddHatch()
	if err != nil {
		t.Fatal(err)
	}
	he.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true,
		Points: [][]float64{{100, 0}, {120, 0}, {120, 10}, {100, 10}}}}
	he.HasGradient = true
	he.Gradient = &entity.HatchGradient{Color1: 1, Color2: 5, Angle: 0}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	// extents (100,0)-(120,10): scale 38, device x 20..780, y 110..490
	var svgBuf, pdfBuf, pngBuf bytes.Buffer
	for _, name := range []string{"svg", "pdf", "png"} {
		r, err := NewBackend(name, mapWriter(name, &svgBuf, &pdfBuf, &pngBuf), 800, 600)
		if err != nil {
			t.Fatal(err)
		}
		if err := renderScene(r, sc); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(svgBuf.String(), "linearGradient") ||
		!strings.Contains(svgBuf.String(), "url(#grad") {
		t.Errorf("svg gradient missing:\n%s", svgBuf.String())
	}
	if !strings.Contains(pdfBuf.String(), "/ShadingType 2") ||
		!strings.Contains(pdfBuf.String(), "/Sh1 sh") {
		t.Errorf("pdf gradient shading missing:\n%s", pdfBuf.String())
	}
	w, h, at := decodePNG(t, &pngBuf)
	if w != 800 || h != 600 {
		t.Fatalf("png %dx%d", w, h)
	}
	rr, gg, bb := at(25, 300)
	if !(rr > 150 && bb < 120 && rr > gg) {
		t.Errorf("gradient start pixel (%d,%d,%d), want red", rr, gg, bb)
	}
	rr, gg, bb = at(775, 300)
	if !(bb > 150 && rr < 120 && bb > gg) {
		t.Errorf("gradient end pixel (%d,%d,%d), want blue", rr, gg, bb)
	}
}

func mapWriter(name string, svg, pdf, pngB *bytes.Buffer) *bytes.Buffer {
	switch name {
	case "svg":
		return svg
	case "pdf":
		return pdf
	}
	return pngB
}

func TestViewportScene(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	vp, err := d.Model().AddViewport()
	if err != nil {
		t.Fatal(err)
	}
	vp.Status = 1
	vp.ViewHeight = 10
	vp.ViewCenterPoint = mathx.Vec2{0, 0}
	if _, err := d.Model().AddLine(0, -1, 0, 0, 1, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(1000, 0, 0, 1001, 0, 0); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Scale != 60 {
		t.Errorf("viewport scale %v, want 60", sc.Scale)
	}
	if sc.ExtMinX > -6 || sc.ExtMaxX < 6 {
		t.Errorf("viewport x extents [%v, %v], want ~[-6.67, 6.67]", sc.ExtMinX, sc.ExtMaxX)
	}
	// near line maps to canvas center x=400
	var foundNear bool
	for _, sp := range sc.Paths {
		lo, hi := sp.Path.BoundingBox()
		if lo.X() == 400 && hi.X() == 400 && lo.Y() >= 230 && hi.Y() <= 370 {
			foundNear = true
		}
	}
	if !foundNear {
		t.Error("viewport: near line not centered on canvas")
	}
}

func TestSceneExtentsOverride(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 10, 0); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 100, 100, &SceneOptions{Min: []float64{0, 0}, Max: []float64{10, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if sc.Scale != 9.5 {
		t.Errorf("explicit extents scale %v, want 9.5", sc.Scale)
	}
	if sc.ExtMinX != 0 || sc.ExtMaxX != 10 {
		t.Errorf("extents [%v..%v], want [0..10]", sc.ExtMinX, sc.ExtMaxX)
	}
}

func TestTTFTextPath(t *testing.T) {
	ps, name, err := text2path.TextToPathsTTF("AB", 10)
	if err != nil || name == "" || len(ps) == 0 {
		t.Skipf("no system TrueType font available: %v (%q)", err, name)
	}
	p := ps[0]
	lo, hi := p.BoundingBox()
	if hi.X()-lo.X() < 5 {
		t.Errorf("glyph bbox width %v, want > 5", hi.X()-lo.X())
	}
	if lo.Y() < -0.4*10 || hi.Y() > 1.3*10 {
		t.Errorf("glyph bbox y [%v, %v] far from baseline band", lo.Y(), hi.Y())
	}
	if len(p.Vertices()) < 40 {
		t.Errorf("contour vertices %d, want many", len(p.Vertices()))
	}
	// scene integration: filled text style
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddText(0, 0, 0, 10, "Hello"); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	filled := false
	for _, sp := range sc.Paths {
		if sp.Style.Fill.A > 0 && sp.Style.FillRule == "evenodd" && len(sp.Path.Vertices()) > 100 {
			filled = true
		}
	}
	if !filled {
		t.Error("TTF text scene path not filled with evenodd")
	}
}

func TestHatchPatternExpanded(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	he, err := d.Model().AddHatch(d.WithLayer("HP"))
	if err != nil {
		t.Fatal(err)
	}
	he.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true,
		Points: [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}}
	he.SetPredefinedPattern("ANSI31", 0, 1, 1, false)
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	boundary, segments := false, 0
	for _, sp := range sc.Paths {
		if sp.FillHatch != nil {
			if sp.Style.Fill.A == 0 && sp.Style.Stroke.A > 0 {
				boundary = true
			}
			continue
		}
		if sp.Style.Fill.A == 0 && sp.Style.Stroke.A > 0 {
			segments++
		}
	}
	if !boundary {
		t.Error("pattern hatch boundary not stroked unfilled")
	}
	if segments == 0 {
		t.Error("pattern hatch segments missing from scene")
	}
}

func TestISOHatchPatternExpanded(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	he, err := d.Model().AddHatch(d.WithLayer("HP"))
	if err != nil {
		t.Fatal(err)
	}
	he.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true,
		Points: [][]float64{{0, 0}, {20, 0}, {20, 20}, {0, 20}}}}
	he.SetPredefinedPattern("ACAD_ISO04W100", 0, 1, 1, false)
	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	boundary, segments := false, 0
	for _, sp := range sc.Paths {
		if sp.FillHatch != nil {
			if sp.FillHatch.Pattern != nil && sp.FillHatch.Pattern.Name != "ACAD_ISO04W100" {
				t.Fatalf("pattern name = %q", sp.FillHatch.Pattern.Name)
			}
			if sp.Style.Fill.A == 0 && sp.Style.Stroke.A > 0 {
				boundary = true
			}
			continue
		}
		if sp.Style.Fill.A == 0 && sp.Style.Stroke.A > 0 {
			segments++
		}
	}
	if !boundary {
		t.Error("ISO hatch boundary not stroked unfilled")
	}
	if segments == 0 {
		t.Error("ISO hatch segments missing from scene")
	}
}
