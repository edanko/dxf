package render

import (
	"testing"

	"github.com/edanko/dxf/addons/acadctb"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/internal/blockadapter"
	dxfmath "github.com/edanko/dxf/math"
)

func TestDimensionRendersReferencedBlock(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	b, err := d.NewBlock("*D1", "dimension geometry")
	if err != nil {
		t.Fatal(err)
	}
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 0, 0}
	b.AddEntity(blockadapter.Adapt(line))
	txt := entity.NewText()
	txt.Coord1 = []float64{4, 1, 0}
	txt.Height = 1
	txt.Value = "10.00"
	b.AddEntity(blockadapter.Adapt(txt))

	dim := entity.NewDimension()
	dim.SetBlockName("*D1")
	dim.SetExtensionPoints([]float64{0, 0, 0}, []float64{10, 0, 0})
	dim.SetDefinitionPoint([]float64{0, 0, 0})
	dim.SetTextMidpoint([]float64{4, 1, 0})
	if err := d.AddEntity(dim); err != nil {
		t.Fatal(err)
	}

	sc, err := BuildScene(d, 400, 200, &SceneOptions{Min: []float64{0, 0}, Max: []float64{10, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Paths) < 2 {
		t.Fatalf("dimension block paths = %d, want line and text", len(sc.Paths))
	}
}

func TestDimensionFallbackUsesMeasurementText(t *testing.T) {
	dim := entity.NewDimension()
	dim.SetExtensionPoints([]float64{0, 0, 0}, []float64{10, 0, 0})
	dim.SetDefinitionPoint([]float64{0, 5, 0})
	dim.SetTextMidpoint([]float64{4, 6, 0})
	if got := dim.MeasurementText(); got != "10.00" {
		t.Fatalf("MeasurementText = %q, want 10.00", got)
	}
	parts, err := dim.Explode()
	if err != nil {
		t.Fatal(err)
	}
	texts := 0
	for _, p := range parts {
		if _, ok := p.(*entity.Text); ok {
			texts++
		}
	}
	if texts != 1 {
		t.Fatalf("exploded texts = %d, want 1 measured text", texts)
	}
}

func TestMeshRendersClosedFaces(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	mesh := entity.NewMesh()
	for _, p := range [][3]float64{{0, 0, 0}, {2, 0, 0}, {2, 2, 0}, {0, 2, 0}} {
		mesh.AddVertex(dxfmath.NewVec3(p[0], p[1], p[2]))
	}
	mesh.AddFace([]int{0, 1, 2, 3})
	mesh.AddFace([]int{0, 1, 9}) // invalid, must be skipped
	if err := d.AddEntity(mesh); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 200, 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Paths) != 1 {
		t.Fatalf("mesh paths = %d, want 1 valid face", len(sc.Paths))
	}
	if !sc.Paths[0].Path.IsClosed() {
		t.Fatal("mesh face path is not closed")
	}
}

func TestPolyfaceRendersClosedFaces(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	pl := entity.NewPolyline()
	pl.AddPolyfaceVertex(0, 0, 0)
	pl.AddPolyfaceVertex(3, 0, 0)
	pl.AddPolyfaceVertex(3, 3, 0)
	pl.AddFace(1, 2, 3)
	if err := d.AddEntity(pl); err != nil {
		t.Fatal(err)
	}
	sc, err := BuildScene(d, 200, 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Paths) != 1 {
		t.Fatalf("polyface paths = %d, want 1 face", len(sc.Paths))
	}
	if !sc.Paths[0].Path.IsClosed() {
		t.Fatal("polyface path is not closed")
	}
}

func TestPlotStyleOverridesRendering(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 0, 0}
	if err := d.AddEntity(line); err != nil {
		t.Fatal(err)
	}
	table := &acadctb.PlotStyleTable{
		Name:    "test.ctb",
		ByColor: true,
		Styles: []acadctb.PlotStyle{{
			Name:         "7",
			Index:        6, // ACI 7
			Color:        1,
			Lineweight:   50,
			Screening:    100,
			ScreeningSet: true,
		}},
	}
	sc, err := BuildScene(d, 200, 200, &SceneOptions{PlotStyles: table})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Paths) != 1 {
		t.Fatalf("paths = %d", len(sc.Paths))
	}
	stroke := sc.Paths[0].Style.Stroke
	if stroke.R <= stroke.G || stroke.R <= stroke.B {
		t.Fatalf("plot color not applied: %+v", stroke)
	}
	wantWidth := 50.0 / 100.0 / 25.4 * 96.0
	if diff := sc.Paths[0].Style.Width - wantWidth; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("plot lineweight = %v, want %v", sc.Paths[0].Style.Width, wantWidth)
	}
}

func TestNamedPlotStyleOverridesRendering(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 0, 0}
	if err := d.AddEntity(line); err != nil {
		t.Fatal(err)
	}
	table := &acadctb.PlotStyleTable{
		Name: "test.stb",
		Styles: []acadctb.PlotStyle{{
			Name:         "Black",
			Index:        -1,
			Color:        1,
			Lineweight:   50,
			Screening:    100,
			ScreeningSet: true,
		}},
	}
	sc, err := BuildScene(d, 200, 200, &SceneOptions{PlotStyles: table, PlotStyleName: "Black"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Paths) != 1 {
		t.Fatalf("paths = %d", len(sc.Paths))
	}
	stroke := sc.Paths[0].Style.Stroke
	if stroke.R <= stroke.G || stroke.R <= stroke.B {
		t.Fatalf("named plot color not applied: %+v", stroke)
	}
}

func TestArrowHeadGeometryStyles(t *testing.T) {
	tests := []struct {
		kind   ArrowType
		name   string
		closed bool
		filled bool
		min    int
	}{
		{ArrowClosedBlank, "CLOSED_BLANK", true, false, 3},
		{ArrowClosedFilled, "CLOSED_FILLED", true, true, 3},
		{ArrowOpen, "OPEN", false, false, 3},
		{ArrowOpen30, "OPEN_30", false, false, 3},
		{ArrowOpen90, "OPEN_90", false, false, 3},
		{ArrowCircle, "CIRCLE", true, false, 8},
		{ArrowDot, "DOT", true, true, 8},
		{ArrowArchTick, "ARCH_TICK", false, false, 2},
		{ArrowOblique, "OBLIQUE", false, false, 2},
		{ArrowBox, "BOX", true, false, 4},
		{ArrowBoxFilled, "BOX_FILLED", true, true, 4},
		{ArrowDiamond, "DIAMOND", true, false, 4},
		{ArrowDotSmall, "DOT_SMALL", true, true, 8},
	}
	for _, tc := range tests {
		a := NewArrow(tc.kind, 2, 0)
		if len(a.Vertices) < tc.min {
			t.Fatalf("%s vertices = %d, want >= %d", tc.name, len(a.Vertices), tc.min)
		}
		if a.Closed != tc.closed || a.Filled != tc.filled {
			t.Fatalf("%s closed/filled = %v/%v, want %v/%v", tc.name, a.Closed, a.Filled, tc.closed, tc.filled)
		}
		if got := GetArrowTypeName(tc.kind); got != tc.name {
			t.Fatalf("GetArrowTypeName(%d) = %q, want %q", tc.kind, got, tc.name)
		}
	}
}
