package render

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/text2path"
)

// visualTolerance is the accepted difference between a rendered image and
// its golden fixture: MaxChannelDelta is the per-channel 8-bit deviation
// treated as equal, and MaxMismatchRatio is the largest fraction of pixels
// allowed to exceed it.
type visualTolerance struct {
	MaxChannelDelta  int
	MaxMismatchRatio float64
}

// referenceTolerance records the tolerance supported on the reference
// platform. It is documented in docs/parity/README.md.
var referenceTolerance = visualTolerance{MaxChannelDelta: 8, MaxMismatchRatio: 0.005}

const (
	visualWidth  = 400
	visualHeight = 300
)

var updateVisualGoldens = flag.Bool("update-visual-goldens", false,
	"rewrite the visual conformance golden files")

// visualCase is one curated visual conformance fixture.
type visualCase struct {
	Name  string
	Scene func(t *testing.T) *StyledScene
}

func visualCases() []visualCase {
	return []visualCase{
		{Name: "boundary_edges", Scene: boundaryEdgesScene},
		{Name: "embedded_spline", Scene: embeddedSplineScene},
		{Name: "complex_linetype", Scene: complexLinetypeScene},
		{Name: "pdf_gradient", Scene: gradientScene},
	}
}

// TestVisualConformance renders every curated fixture through the raster and
// SVG backends and compares the result with the checked-in goldens. A
// mismatch reports the fixture, the backend, and the mismatching region.
func TestVisualConformance(t *testing.T) {
	for _, tc := range visualCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			sc := tc.Scene(t)
			if len(sc.Fallbacks) != 0 {
				t.Fatalf("fixture %q: unexpected fallbacks: %v", tc.Name, sc.Fallbacks)
			}
			dir := filepath.Join("testdata", "visual")
			svgPath := filepath.Join(dir, tc.Name+".svg")
			pngPath := filepath.Join(dir, tc.Name+".png")
			if *updateVisualGoldens {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				var svgBuf bytes.Buffer
				if err := drawGoldenScene(t, "svg", &svgBuf, sc); err != nil {
					t.Fatal(err)
				}
				var pngBuf bytes.Buffer
				if err := drawGoldenScene(t, "png", &pngBuf, sc); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(svgPath, svgBuf.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(pngPath, pngBuf.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			wantSVG, err := os.ReadFile(svgPath)
			if err != nil {
				t.Fatalf("fixture %q: read golden: %v (run go test ./render -run TestVisualConformance -update-visual-goldens)", tc.Name, err)
			}
			var svgBuf bytes.Buffer
			if err := drawGoldenScene(t, "svg", &svgBuf, sc); err != nil {
				t.Fatal(err)
			}
			if got := svgBuf.String(); got != string(wantSVG) {
				t.Errorf("fixture %q backend %q: SVG differs from %s: %s",
					tc.Name, "svg", svgPath, firstLineDifference(string(wantSVG), got))
			}

			wantPNG, err := os.ReadFile(pngPath)
			if err != nil {
				t.Fatalf("fixture %q: read golden: %v (run go test ./render -run TestVisualConformance -update-visual-goldens)", tc.Name, err)
			}
			wantImg, err := png.Decode(bytes.NewReader(wantPNG))
			if err != nil {
				t.Fatalf("fixture %q: decode golden PNG: %v", tc.Name, err)
			}
			var pngBuf bytes.Buffer
			if err := drawGoldenScene(t, "png", &pngBuf, sc); err != nil {
				t.Fatal(err)
			}
			gotImg, err := png.Decode(bytes.NewReader(pngBuf.Bytes()))
			if err != nil {
				t.Fatalf("fixture %q: decode rendered PNG: %v", tc.Name, err)
			}
			diff, err := diffRaster(tc.Name, "png", gotImg, wantImg, referenceTolerance)
			if err != nil {
				t.Fatalf("%v (mismatch %d pixels, max channel delta %d, region %s)",
					err, diff.Mismatched, diff.MaxDelta, diff.Region)
			}
		})
	}
}

func drawGoldenScene(t *testing.T, backend string, buf *bytes.Buffer, sc *StyledScene) error {
	t.Helper()
	r, err := NewBackend(backend, buf, visualWidth, visualHeight)
	if err != nil {
		return err
	}
	if err := r.Begin(visualWidth, visualHeight); err != nil {
		return err
	}
	for _, sp := range sc.Paths {
		if err := r.SetStyle(sp.Style); err != nil {
			return err
		}
		if sp.Clip != nil {
			if err := r.Clip(sp.Clip); err != nil {
				return err
			}
		}
		if sp.FillHatch != nil && sp.FillHatch.Gradient != nil {
			if err := r.FillGradient(sp.Path, sp.FillHatch.Gradient); err != nil {
				return err
			}
		} else if err := r.DrawPath(sp.Path); err != nil {
			return err
		}
		if sp.Clip != nil {
			if err := r.Clip(nil); err != nil {
				return err
			}
		}
	}
	return r.End()
}

// rasterDiff reports how far a rendered image deviates from its golden.
type rasterDiff struct {
	Mismatched       int
	MaxDelta         int
	MinX, MinY       int
	MaxX, MaxY       int
	Region           string
	MismatchRatio    float64
	ChannelTolerance int
}

// diffRaster compares got against want using tol. It returns a
// fixture/backend/region failure when the mismatch ratio exceeds the
// tolerance or the images differ in size.
func diffRaster(fixture, backend string, got, want image.Image, tol visualTolerance) (rasterDiff, error) {
	gb, wb := got.Bounds(), want.Bounds()
	if gb.Dx() != wb.Dx() || gb.Dy() != wb.Dy() {
		return rasterDiff{}, fmt.Errorf(
			"visual conformance: fixture %q backend %q: size %dx%d != golden %dx%d",
			fixture, backend, gb.Dx(), gb.Dy(), wb.Dx(), wb.Dy())
	}
	d := rasterDiff{
		MinX: 1 << 30, MinY: 1 << 30,
		MaxX: -1, MaxY: -1,
		ChannelTolerance: tol.MaxChannelDelta,
	}
	total := gb.Dx() * gb.Dy()
	for y := 0; y < gb.Dy(); y++ {
		for x := 0; x < gb.Dx(); x++ {
			gr, gg, gb2, ga := got.At(gb.Min.X+x, gb.Min.Y+y).RGBA()
			wr, wg, wb2, wa := want.At(wb.Min.X+x, wb.Min.Y+y).RGBA()
			deltas := []int{
				absInt(int(gr>>8) - int(wr>>8)),
				absInt(int(gg>>8) - int(wg>>8)),
				absInt(int(gb2>>8) - int(wb2>>8)),
				absInt(int(ga>>8) - int(wa>>8)),
			}
			maxDelta := 0
			for _, v := range deltas {
				if v > maxDelta {
					maxDelta = v
				}
			}
			if maxDelta <= tol.MaxChannelDelta {
				continue
			}
			d.Mismatched++
			if maxDelta > d.MaxDelta {
				d.MaxDelta = maxDelta
			}
			if x < d.MinX {
				d.MinX = x
			}
			if x > d.MaxX {
				d.MaxX = x
			}
			if y < d.MinY {
				d.MinY = y
			}
			if y > d.MaxY {
				d.MaxY = y
			}
		}
	}
	if d.Mismatched == 0 {
		return d, nil
	}
	d.MismatchRatio = float64(d.Mismatched) / float64(total)
	d.Region = fmt.Sprintf("x=%d..%d y=%d..%d", d.MinX, d.MaxX, d.MinY, d.MaxY)
	if d.MismatchRatio > tol.MaxMismatchRatio {
		return d, fmt.Errorf(
			"visual conformance: fixture %q backend %q: %d/%d pixels (%.4f) exceed per-channel tolerance %d (max delta %d) in region %s",
			fixture, backend, d.Mismatched, total, d.MismatchRatio, tol.MaxChannelDelta, d.MaxDelta, d.Region)
	}
	return d, nil
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// firstLineDifference returns context for the first differing line.
func firstLineDifference(want, got string) string {
	wl := strings.Split(want, "\n")
	gl := strings.Split(got, "\n")
	for i := 0; i < len(wl) && i < len(gl); i++ {
		if wl[i] != gl[i] {
			return fmt.Sprintf("line %d: want %q, got %q", i+1, wl[i], gl[i])
		}
	}
	return fmt.Sprintf("line count %d vs %d", len(wl), len(gl))
}

// TestVisualConformanceDetectsDeliberateChange proves that a deliberate
// pixel change fails with the fixture, backend and region named.
func TestVisualConformanceDetectsDeliberateChange(t *testing.T) {
	want := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			want.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	got := image.NewRGBA(want.Bounds())
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			got.SetRGBA(x, y, want.RGBAAt(x, y))
		}
	}
	for y := 8; y <= 12; y++ {
		for x := 12; x <= 16; x++ {
			got.SetRGBA(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	diff, err := diffRaster("deliberate", "png", got, want, referenceTolerance)
	if err == nil {
		t.Fatalf("deliberate pixel change was not detected (diff %+v)", diff)
	}
	for _, wantPart := range []string{`fixture "deliberate"`, `backend "png"`, "x=12..16", "y=8..12"} {
		if !strings.Contains(err.Error(), wantPart) {
			t.Fatalf("failure %q does not name %q", err, wantPart)
		}
	}
}

// visualScene builds a scene with a fixed viewport and background.
func visualScene(t *testing.T, d *drawing.Drawing, minX, minY, maxX, maxY float64) *StyledScene {
	t.Helper()
	sc, err := BuildScene(d, visualWidth, visualHeight, &SceneOptions{
		Min:          []float64{minX, minY},
		Max:          []float64{maxX, maxY},
		Background:   color.RGBA{255, 255, 255, 255},
		Tessellation: 0.05,
	})
	if err != nil {
		t.Fatalf("BuildScene: %v", err)
	}
	if len(sc.Paths) == 0 {
		t.Fatalf("fixture scene has no paths")
	}
	return sc
}

func boundaryEdgesScene(t *testing.T) *StyledScene {
	t.Helper()
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	h := entity.NewHatch()
	h.SetSolidFill(true)

	outer := entity.NewPolylinePath(0, true, false)
	for _, p := range [][2]float64{{0, 0}, {30, 0}, {30, 20}, {0, 20}} {
		outer.AddVertex(entity.NewVertex3D(p[0], p[1], 0, 0))
	}
	h.AddBoundaryPath(outer)

	hole := entity.NewEdgePath(0)
	hole.AddEdge(entity.NewLineEdge(dxfmath.NewVec2(5, 5), dxfmath.NewVec2(15, 5)))
	hole.AddEdge(entity.NewArcEdge(dxfmath.NewVec2(15, 10), 5, -90, 90, true))
	hole.AddEdge(entity.NewLineEdge(dxfmath.NewVec2(15, 15), dxfmath.NewVec2(5, 15)))
	hole.AddEdge(entity.NewArcEdge(dxfmath.NewVec2(5, 10), 5, 90, 270, true))
	h.AddBoundaryPath(hole)

	if err := d.AddEntity(h); err != nil {
		t.Fatal(err)
	}
	return visualScene(t, d, -2, -2, 32, 22)
}

func embeddedSplineScene(t *testing.T) *StyledScene {
	t.Helper()
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	e := entity.NewEllipse()
	e.Center = []float64{10, 10, 0}
	e.Major = 1
	e.Minor = 0.5
	e.Start = 0
	e.End = 360
	sp := entity.NewSpline(3)
	sp.Knots = []float64{0, 0, 0, 0, 0.5, 1, 1, 1, 1}
	sp.Controls = [][]float64{{2, 4, 0}, {6, 18, 0}, {14, 2, 0}, {18, 12, 0}, {22, 6, 0}}
	e.Spline = sp
	if err := d.AddEntity(e); err != nil {
		t.Fatal(err)
	}
	return visualScene(t, d, 0, 0, 24, 20)
}

func complexLinetypeScene(t *testing.T) *StyledScene {
	t.Helper()
	d := complexLTypeDrawing(t, false)
	sc, err := BuildScene(d, visualWidth, visualHeight, &SceneOptions{
		Min:          []float64{0, -2},
		Max:          []float64{22, 4},
		Background:   color.RGBA{255, 255, 255, 255},
		Tessellation: 0.05,
		ShapeFonts:   map[string]*text2path.SHXFont{"ltypeshp.shx": syntheticShapeFont(t)},
	})
	if err != nil {
		t.Fatalf("BuildScene: %v", err)
	}
	return sc
}

func gradientScene(t *testing.T) *StyledScene {
	t.Helper()
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	he, err := d.Model().AddHatch()
	if err != nil {
		t.Fatal(err)
	}
	he.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true,
		Points: [][]float64{{2, 2}, {28, 2}, {28, 18}, {2, 18}}}}
	he.HasGradient = true
	he.Gradient = &entity.HatchGradient{Color1: 1, Color2: 5, Angle: 30}
	return visualScene(t, d, 0, 0, 30, 20)
}
