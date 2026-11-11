package tablepainter_test

import (
	"testing"

	"github.com/edanko/dxf/addons/tablepainter"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func TestPaintTable(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	cells := [][]string{{"A", "B"}, {"C", "D"}}
	if err := tablepainter.PaintTable(d, 0, 0, 10, 5, cells); err != nil {
		t.Fatal(err)
	}
	lines, texts := 0, 0
	for _, e := range d.Entities() {
		switch e.(type) {
		case *entity.Line:
			lines++
		case *entity.Text:
			texts++
		}
	}
	if lines != 6 {
		t.Fatalf("grid lines: %d (want 6)", lines)
	}
	if texts != 4 {
		t.Fatalf("cell texts: %d (want 4)", texts)
	}
}
