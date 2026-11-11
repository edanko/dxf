package render

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/patterns"
)

func TestParsedPatternExpandsToClippedSegments(t *testing.T) {
	def := patterns.PatternDefinition{
		Name:        "CUSTOMHATCH",
		Description: "horizontal test lines",
		Category:    patterns.PatternCategoryCustom,
		Lines: []patterns.PatternLine{{
			Angle:  0,
			Offset: []float64{0, 2},
		}},
	}
	if err := patterns.RegisterPattern(def); err != nil {
		t.Fatal(err)
	}
	defer patterns.UnregisterPattern(def.Name)

	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	h, err := d.Model().AddHatch()
	if err != nil {
		t.Fatal(err)
	}
	h.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true, Points: [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}}
	h.SetPredefinedPattern(def.Name, 0, 1, 0, false)

	sc, err := BuildScene(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	segments := 0
	named := false
	for _, sp := range sc.Paths {
		if sp.FillHatch != nil {
			if sp.FillHatch.Pattern != nil && sp.FillHatch.Pattern.Name == def.Name {
				named = true
			}
			continue
		}
		if sp.Style.Stroke.A > 0 {
			segments++
		}
	}
	if !named {
		t.Fatal("parsed pattern name not preserved")
	}
	if segments == 0 {
		t.Fatal("parsed pattern produced no clipped segments")
	}
}

func TestUnknownHatchPatternKeepsBoundaryAndReports(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	h, err := d.Model().AddHatch()
	if err != nil {
		t.Fatal(err)
	}
	h.BoundaryPaths = []entity.HatchBoundaryPath{{Type: 0, Flags: 1, IsClosed: true, Points: [][]float64{{0, 0}, {5, 0}, {5, 5}, {0, 5}}}}
	h.SetPredefinedPattern("MISSINGPAT", 0, 1, 0, false)

	sc, err := BuildScene(d, 400, 400, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range sc.Fallbacks {
		if strings.Contains(f, h.Handle()) && strings.Contains(f, "MISSINGPAT") {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing pattern fallback not reported: %v", sc.Fallbacks)
	}
	boundary := false
	for _, sp := range sc.Paths {
		if sp.FillHatch != nil && sp.FillHatch.Pattern != nil && sp.FillHatch.Pattern.Name == "MISSINGPAT" {
			boundary = true
		}
	}
	if !boundary {
		t.Fatal("unknown pattern boundary was dropped")
	}
}
