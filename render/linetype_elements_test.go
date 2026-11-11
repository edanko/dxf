package render

import (
	"math"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
	"github.com/edanko/dxf/table"
	"github.com/edanko/dxf/text2path"
)

func TestComplexLinetypeShapePlacedAlongPath(t *testing.T) {
	font := syntheticShapeFont(t)
	elems := []table.LineTypeElement{
		{DashLength: 10},
		{DashLength: 0, SegLen: 2, ShapeNum: 1, XScale: 2},
		{DashLength: -15},
	}
	p := straightPath(dxfmath.NewVec2(0, 0), dxfmath.NewVec2(20, 0))
	nt := &sceneNotes{}
	opts := &SceneOptions{ShapeFonts: map[string]*text2path.SHXFont{"ltypeshp.shx": font}}
	out := placeLineTypeElements(pointsOf(p), elems, 1, 25, "SHAPETEXT", opts, Style{}, map[string]bool{}, nt)
	if len(out) != 1 {
		t.Fatalf("shape paths = %d, want 1", len(out))
	}
	min, max := out[0].Path.BoundingBox()
	assertNear(t, "shape min x", min.X(), 10)
	assertNear(t, "shape min y", min.Y(), 0)
	assertNear(t, "shape max x", max.X(), 12)
	assertNear(t, "shape max y", max.Y(), 1)
	if len(nt.fallbacks) != 0 {
		t.Fatalf("unexpected fallbacks: %v", nt.fallbacks)
	}
}

func TestComplexLinetypeShapeFollowsTangent(t *testing.T) {
	font := syntheticShapeFont(t)
	elems := []table.LineTypeElement{
		{DashLength: 10},
		{DashLength: 0, SegLen: 2, ShapeNum: 1, XScale: 2},
		{DashLength: -15},
	}
	p := straightPath(dxfmath.NewVec2(0, 0), dxfmath.NewVec2(0, 20))
	opts := &SceneOptions{ShapeFonts: map[string]*text2path.SHXFont{"ltypeshp.shx": font}}
	out := placeLineTypeElements(pointsOf(p), elems, 1, 25, "SHAPETEXT", opts, Style{}, map[string]bool{}, &sceneNotes{})
	if len(out) != 1 {
		t.Fatalf("shape paths = %d, want 1", len(out))
	}
	min, max := out[0].Path.BoundingBox()
	assertNear(t, "rotated shape min x", min.X(), -1)
	assertNear(t, "rotated shape max x", max.X(), 0)
	assertNear(t, "rotated shape min y", min.Y(), 10)
	assertNear(t, "rotated shape max y", max.Y(), 12)
}

func TestComplexLinetypeShapeRotationAndOffset(t *testing.T) {
	font := syntheticShapeFont(t)
	elems := []table.LineTypeElement{
		{DashLength: 4},
		{DashLength: 0, SegLen: 2, ShapeNum: 1, XScale: 2, Rotation: 90, YOffset: 1},
		{DashLength: -20},
	}
	p := straightPath(dxfmath.NewVec2(0, 0), dxfmath.NewVec2(20, 0))
	opts := &SceneOptions{ShapeFonts: map[string]*text2path.SHXFont{"ltypeshp.shx": font}}
	out := placeLineTypeElements(pointsOf(p), elems, 1, 24, "SHAPETEXT", opts, Style{}, map[string]bool{}, &sceneNotes{})
	if len(out) != 1 {
		t.Fatalf("shape paths = %d, want 1", len(out))
	}
	min, max := out[0].Path.BoundingBox()
	// Rotated 90 degrees around (4,1): local (0,0),(2,0),(2,1) become
	// (4,1),(4,3),(3,3).
	assertNear(t, "offset shape min x", min.X(), 3)
	assertNear(t, "offset shape max x", max.X(), 4)
	assertNear(t, "offset shape min y", min.Y(), 1)
	assertNear(t, "offset shape max y", max.Y(), 3)
}

func TestComplexLinetypeTextElementPlaced(t *testing.T) {
	elems := []table.LineTypeElement{
		{DashLength: 5},
		{DashLength: 0, TextValue: "AB", XScale: 2},
		{DashLength: -20},
	}
	p := straightPath(dxfmath.NewVec2(0, 0), dxfmath.NewVec2(20, 0))
	out := placeLineTypeElements(pointsOf(p), elems, 1, 25, "SHAPETEXT", &SceneOptions{}, Style{}, map[string]bool{}, &sceneNotes{})
	if len(out) != 1 {
		t.Fatalf("text paths = %d, want 1", len(out))
	}
	min, max := out[0].Path.BoundingBox()
	if min.X() < 4.5 || min.X() > 6 {
		t.Fatalf("text min x = %v, want near 5", min.X())
	}
	if max.Y()-min.Y() < 0.5 {
		t.Fatalf("text height = %v, want a visible glyph", max.Y()-min.Y())
	}
}

func TestComplexLinetypeMissingShapeKeepsText(t *testing.T) {
	elems := []table.LineTypeElement{
		{DashLength: 5},
		{DashLength: 0, SegLen: 2, ShapeNum: 1, XScale: 2},
		{DashLength: 5},
		{DashLength: 0, TextValue: "AB", XScale: 2},
		{DashLength: -20},
	}
	p := straightPath(dxfmath.NewVec2(0, 0), dxfmath.NewVec2(30, 0))
	nt := &sceneNotes{}
	out := placeLineTypeElements(pointsOf(p), elems, 1, 30, "SHAPETEXT", &SceneOptions{}, Style{}, map[string]bool{}, nt)
	if len(out) != 1 {
		t.Fatalf("paths = %d, want only the text element", len(out))
	}
	if len(nt.fallbacks) != 1 || nt.fallbacks[0] != "missing shape 1 in linetype SHAPETEXT" {
		t.Fatalf("fallbacks = %v", nt.fallbacks)
	}
}

func TestComplexLinetypeSceneRendersShapeAndText(t *testing.T) {
	d := complexLTypeDrawing(t, true)
	font := syntheticShapeFont(t)
	opts := &SceneOptions{
		Min:        []float64{0, 0},
		Max:        []float64{20, 5},
		ShapeFonts: map[string]*text2path.SHXFont{"ltypeshp.shx": font},
	}
	sc, err := BuildScene(d, 200, 200, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Fallbacks) != 0 {
		t.Fatalf("unexpected fallbacks: %v", sc.Fallbacks)
	}
	// The line itself plus the shape and text element paths.
	if len(sc.Paths) != 3 {
		t.Fatalf("paths = %d, want 3", len(sc.Paths))
	}
}

func TestComplexLinetypeSceneMissingShapeFallbackStable(t *testing.T) {
	d := complexLTypeDrawing(t, true)
	build := func() *StyledScene {
		sc, err := BuildScene(d, 200, 200, &SceneOptions{
			Min: []float64{0, 0},
			Max: []float64{20, 5},
		})
		if err != nil {
			t.Fatal(err)
		}
		return sc
	}
	first := build()
	second := build()
	want := "missing shape 1 in linetype SHAPETEXT"
	if len(first.Fallbacks) != 1 || first.Fallbacks[0] != want {
		t.Fatalf("fallbacks = %v, want [%q]", first.Fallbacks, want)
	}
	if strings.Join(first.Fallbacks, "|") != strings.Join(second.Fallbacks, "|") {
		t.Fatalf("fallback report not stable: %v vs %v", first.Fallbacks, second.Fallbacks)
	}
	// The line and the text element survive without the shape font.
	if len(first.Paths) != 2 {
		t.Fatalf("paths = %d, want line and text", len(first.Paths))
	}
}

// complexLTypeDrawing builds a drawing with a line using the complex linetype
// SHAPETEXT: a dash, a shape element, a dash and a text element.
func complexLTypeDrawing(t *testing.T, withText bool) *drawing.Drawing {
	t.Helper()
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	lt, err := d.AddLineType("SHAPETEXT", "")
	if err != nil {
		t.Fatal(err)
	}
	lt.AddElement(table.LineTypeElement{DashLength: 5})
	lt.AddElement(table.LineTypeElement{DashLength: 0, SegLen: 2, ShapeNum: 1, XScale: 2})
	lt.AddElement(table.LineTypeElement{DashLength: 5})
	if withText {
		lt.AddElement(table.LineTypeElement{DashLength: 0, TextValue: "AB", XScale: 2})
	}
	lt.AddElement(table.LineTypeElement{DashLength: -15})
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{20, 0, 0}
	line.SetLinetype("SHAPETEXT")
	if err := d.AddEntity(line); err != nil {
		t.Fatal(err)
	}
	return d
}

func straightPath(a, b dxfmath.Vec2) *path.Path {
	p := path.NewPath()
	p.MoveTo(a.X(), a.Y())
	p.LineTo(b.X(), b.Y())
	return p
}

func pointsOf(p *path.Path) []dxfmath.Vec2 {
	return p.Vertices()
}

func assertNear(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

// syntheticShapeFont builds a shapes-format SHX font with glyph 1 drawing a
// right triangle 10 units wide and 5 units high.
func syntheticShapeFont(t *testing.T) *text2path.SHXFont {
	t.Helper()
	return &text2path.SHXFont{
		Name:  "ltypeshp",
		Above: 10,
		Glyphs: map[int]*text2path.SHXGlyph{
			// Pen down, right 10, up 5, back to the origin.
			1: {Number: 1, Name: "FLAG", Codes: []int{1, 160, 84, 169, 0}},
		},
	}
}
