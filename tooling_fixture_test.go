package dxf

import (
	"os"
	"testing"

	textsize "github.com/edanko/dxf/addons/text2path"
	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/edges"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/lldxf"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/patterns"
	"github.com/edanko/dxf/query"
)

// TestToolingFixture exercises the tooling scope end to end on one file:
// primitive decomposition, text sizing, pattern parsing, edge topology and
// binding-free raw section loading.
func TestToolingFixture(t *testing.T) {
	path := "testdata/tooling_fixture.dxf"
	d, err := FromFile(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	lookup := func(name string) *block.Block {
		b, err := d.GetBlock(name)
		if err != nil {
			return nil
		}
		return b
	}
	dec := query.RecursiveDecompose(d.Entities(), lookup)
	if len(dec.Diagnostics) != 0 {
		t.Fatalf("decomposition diagnostics: %+v", dec.Diagnostics)
	}
	if len(dec.Cycles) != 0 {
		t.Fatalf("unexpected decomposition cycles: %v", dec.Cycles)
	}
	var nestedLine, nestedText, matrixText, hatchPath bool
	for _, p := range dec.Primitives {
		switch p.Kind {
		case query.PrimitiveLine:
			if len(p.Vertices) == 2 && p.Vertices[0].X() == 10 && p.Vertices[0].Y() == 5 &&
				p.Vertices[1].X() == 20 && p.Vertices[1].Y() == 5 {
				nestedLine = true
			}
		case query.PrimitiveText:
			for _, v := range p.Vertices {
				if v.X() >= 9 && v.X() <= 11 && v.Y() >= 6 && v.Y() <= 9 {
					nestedText = true
				}
				if v.X() >= 29 && v.X() <= 40 && v.Y() >= 4 && v.Y() <= 8 {
					matrixText = true
				}
			}
		case query.PrimitivePolyline:
			if p.Path != nil && len(p.Vertices) == 4 {
				inHatch := true
				for _, v := range p.Vertices {
					if v.X() < 20 || v.X() > 30 || v.Y() < 0 || v.Y() > 10 {
						inHatch = false
						break
					}
				}
				if inHatch {
					hatchPath = true
				}
			}
		}
	}
	if !nestedLine || !nestedText || !matrixText || !hatchPath {
		for i, p := range dec.Primitives {
			vs := 0
			if p.Path != nil {
				vs = len(p.Path.Vertices())
			}
			t.Logf("primitive %d: %s vertices=%d pathVertices=%d", i, p.Kind, len(p.Vertices), vs)
		}
		t.Fatalf("decomposition missing primitives: line=%v nestedText=%v text=%v hatch=%v (%d primitives)",
			nestedLine, nestedText, matrixText, hatchPath, len(dec.Primitives))
	}

	var modelText *entity.Text
	var outline *entity.LwPolyline
	for _, e := range d.Entities() {
		switch value := e.(type) {
		case *entity.Text:
			if value.Value == "MEASURE" {
				modelText = value
			}
		case *entity.LwPolyline:
			outline = value
		}
	}
	if modelText == nil || outline == nil {
		t.Fatalf("fixture text=%v outline=%v", modelText != nil, outline != nil)
	}
	size := textsize.Size(modelText)
	if size.Width <= 0 || size.Height <= 0 || size.Ascent <= 0 || size.Descent < 0 {
		t.Fatalf("text size not measured: %+v", size)
	}

	pat := patterns.ParsePAT([]byte("*TOOLINGDOTS, fixture dots\n0, 0,0, 2,0, 0.2,-0.2\n"), "tooling.pat")
	if len(pat.Patterns) != 1 || len(pat.Errors) != 0 {
		t.Fatalf("pat parse: patterns=%d errors=%+v", len(pat.Patterns), pat.Errors)
	}
	pat.Register()
	if got, ok := patterns.LookupPattern("toolingdots"); !ok || got.Name != "TOOLINGDOTS" {
		t.Fatalf("pattern lookup: %+v ok=%v", got, ok)
	}

	var segs []edges.Edge
	points := outline.Vertices
	for i := range points {
		if i+1 >= len(points) && !outline.Closed {
			break
		}
		next := (i + 1) % len(points)
		segs = append(segs, edges.NewEdge(
			dxfmath.NewVec2(points[i][0], points[i][1]),
			dxfmath.NewVec2(points[next][0], points[next][1]),
		))
	}
	topo := edges.BuildGraph(segs, 1e-6).Analyze()
	if len(topo.Loops) != 1 || len(topo.Loops[0].Vertices) != 4 {
		t.Fatalf("edge topology: loops=%d %+v", len(topo.Loops), topo.Loops)
	}
	if pl := topo.Loops[0].ToLwPolyline(); !pl.Closed || len(pl.Vertices) != 4 {
		t.Fatalf("loop polyline: closed=%v vertices=%d", pl.Closed, len(pl.Vertices))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := lldxf.ParseRawSections(data)
	if err != nil {
		t.Fatalf("raw sections: %v", err)
	}
	for _, name := range []string{"HEADER", "TABLES", "BLOCKS", "ENTITIES", "OBJECTS"} {
		if tags, ok := raw.Tags(name); !ok || len(tags) == 0 {
			t.Fatalf("raw section %s missing (present=%v tags=%d)", name, ok, len(tags))
		}
	}
}
