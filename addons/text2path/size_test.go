package text2path_test

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/text2path"
	"github.com/edanko/dxf/entity"
	t2p "github.com/edanko/dxf/text2path"
)

func TestTextSizeReportsMeasuredExtents(t *testing.T) {
	tx := entity.NewText()
	tx.Value = "Measured text"
	tx.Height = 10
	tx.WidthFactor = 1.5
	tx.Coord1 = []float64{5, 7, 0}
	tx.Rotation = 30
	size := text2path.Size(tx)
	if size.Width <= 0 || size.Height <= 0 || size.Ascent <= 0 || size.Descent < 0 || size.Advance <= 0 {
		t.Fatalf("unexpected extents: %+v", size)
	}
	if size.Width < size.Advance {
		t.Fatalf("width factor not applied: width=%v advance=%v", size.Width, size.Advance)
	}
	if size.MaxX <= size.MinX || size.MaxY <= size.MinY {
		t.Fatalf("empty rotated bounds: %+v", size)
	}
	if size.Fallback && size.FallbackReason == "" {
		t.Fatal("fallback size must report a reason")
	}
	if !size.Fallback && size.Font == "" {
		t.Fatal("font-backed size must report the font name")
	}
}

func TestMTextSizeWrapsAndAddsParagraphLines(t *testing.T) {
	m := entity.NewMText()
	m.Value = "alpha beta gamma delta epsilon zeta"
	m.Height = 10
	m.Width = 40
	m.Coord1 = []float64{0, 0, 0}
	narrow := text2path.MTextSize(m)
	if narrow.Lines < 3 {
		t.Fatalf("expected wrapped lines, got %+v", narrow)
	}
	for _, line := range t2p.LayoutMText(m.Value, m.Height, m.Width) {
		if w := t2p.MeasureAdvance(line, m.Height); w > m.Width+1e-9 && strings.Contains(line, " ") {
			t.Fatalf("line %q width %v exceeds column %v", line, w, m.Width)
		}
	}

	wide := entity.NewMText()
	wide.Value = "alpha beta gamma delta epsilon zeta"
	wide.Height = 10
	wide.Width = 0
	wide.Coord1 = []float64{0, 0, 0}
	if got := text2path.MTextSize(wide); got.Lines != 1 {
		t.Fatalf("unconstrained text lines %d, want 1", got.Lines)
	}

	m.Value = "first\\Psecond\\Pthird"
	m.Width = 0
	paragraphs := text2path.MTextSize(m)
	if paragraphs.Lines != 3 {
		t.Fatalf("paragraph lines %d, want 3", paragraphs.Lines)
	}
	if paragraphs.Height <= narrow.Height/3 {
		t.Fatalf("paragraph height not accumulated: %+v", paragraphs)
	}
}

func TestMTextSizeColumnWidth(t *testing.T) {
	m := entity.NewMText()
	m.Value = "one two three four five six"
	m.Height = 10
	m.Width = 200
	m.Columns = entity.NewMTextColumns()
	m.Columns.Count = 2
	m.Columns.Width = 35
	m.Columns.TotalWidth = 80
	size := text2path.MTextSize(m)
	if size.Lines < 3 {
		t.Fatalf("column width did not narrow layout: %+v", size)
	}
	if size.Width < m.Columns.TotalWidth {
		t.Fatalf("total column width not reported: %+v", size)
	}
}
