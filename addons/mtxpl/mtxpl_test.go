package mtxpl_test

import (
	"testing"

	"github.com/edanko/dxf/addons/mtxpl"
	"github.com/edanko/dxf/entity"
)

func TestLinesAndExplode(t *testing.T) {
	m := entity.NewMText()
	m.Coord1 = []float64{0, 10, 0}
	m.Height = 2.5
	m.Width = 100
	m.Value = "AA\\PBB"

	lines := mtxpl.Lines(m)
	if len(lines) != 2 {
		t.Fatalf("lines: %d (%v)", len(lines), lines)
	}
	if lines[0].Text != "AA" || lines[1].Text != "BB" {
		t.Fatalf("line texts: %v", lines)
	}
	if lines[0].Y <= lines[1].Y {
		t.Fatalf("lines must advance downwards: %v", lines)
	}

	exploded := mtxpl.Explode(m)
	if len(exploded) != 2 {
		t.Fatalf("exploded: %d", len(exploded))
	}
	for i, e := range exploded {
		if e.Value != lines[i].Text {
			t.Fatalf("exploded %d value: %q", i, e.Value)
		}
		if e.Height != 2.5 {
			t.Fatalf("exploded %d height: %v", i, e.Height)
		}
	}
}
