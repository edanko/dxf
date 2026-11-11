package entity

import (
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

// TestCopyIsolation covers mutable families whose payloads are easy to share
// accidentally: retained tags/XDATA on the base record, ACIS text and binary
// payloads, nested sub-bodies, and MPolygon boundary state.
func TestCopyIsolation(t *testing.T) {
	t.Run("base retained tags", func(t *testing.T) {
		line := NewLine()
		line.SetRawTags([]TagPair{{Code: 500, Value: "SOURCE"}})
		line.SetColorRGB(0x112233)
		copied := line.Copy().(*Line)
		copied.SetRawTags([]TagPair{{Code: 500, Value: "COPY"}})
		copied.SetColorRGB(0x445566)
		if got := line.RawTags(); len(got) != 1 || got[0].Value != "SOURCE" {
			t.Fatalf("original raw tags changed: %+v", got)
		}
		if line.ColorRGB() != 0x112233 {
			t.Fatalf("original true color changed: %06X", line.ColorRGB())
		}
	})

	t.Run("ACIS body payload", func(t *testing.T) {
		body := NewBody()
		body.SetSATPayload([]string{"original"}, []int{1})
		body.SATData = []byte{1, 2, 3}
		region := NewRegion()
		region.SetSATPayload([]string{"region-original"}, []int{1})
		body.Regions = []*Region{region}

		copied := body.Copy().(*Body)
		copied.SATData[0] = 9
		copied.satLines[0] = "copy"
		copied.Regions[0].satLines[0] = "region-copy"

		if body.SATData[0] != 1 {
			t.Fatalf("binary ACIS payload shared: %v", body.SATData)
		}
		if body.satLines[0] != "original" {
			t.Fatalf("text ACIS payload shared: %v", body.satLines)
		}
		if body.Regions[0] == copied.Regions[0] {
			t.Fatal("nested region pointer shared")
		}
		if body.Regions[0].satLines[0] != "region-original" {
			t.Fatalf("nested region payload shared: %v", body.Regions[0].satLines)
		}
	})

	t.Run("MPolygon state", func(t *testing.T) {
		polygon := NewMPolygon()
		path := polygon.paths.AddPath(1, true)
		path.AddVertex(1, 2, 0)
		polygon.SetPatternFill("ANSI31", 0, 2)
		polygon.SetPatternScale(2)
		polygon.pattern = &Pattern{Lines: []PatternLine{{Angle: 45, BaseX: 1, BaseY: 2, OffsetX: 3, OffsetY: 4, DashItems: []float64{1.5, 0.5}}}}
		polygon.seeds = []dxfmath.Vec2{dxfmath.NewVec2(3, 4)}

		copied := polygon.Copy().(*MPolygonEntity)
		copied.paths.GetPath(0).Vertices[0].X = 99
		copied.SetPatternScale(7)
		copied.pattern.Lines[0].DashItems[0] = 12.5
		copied.seeds[0] = dxfmath.NewVec2(9, 9)

		if polygon.paths.GetPath(0).Vertices[0].X != 1 {
			t.Fatal("boundary path vertices shared")
		}
		if polygon.paths == copied.paths {
			t.Fatal("boundary paths pointer shared")
		}
		if polygon.pattern == copied.pattern {
			t.Fatal("fill pattern pointer shared")
		}
		if polygon.pattern.Lines[0].DashItems[0] != 1.5 {
			t.Fatalf("pattern dash items shared: %v", polygon.pattern.Lines[0].DashItems)
		}
		if polygon.seeds[0].X() != 3 {
			t.Fatalf("hatch seeds shared: %v", polygon.seeds)
		}
	})
}
