package render

import (
	"math"
	"testing"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/text2path"
)

func TestTextAlignmentUsesMeasuredAdvance(t *testing.T) {
	tx := entity.NewText()
	tx.Value = "center me"
	tx.Height = 10
	tx.Anchor(entity.CENTER_TOP)
	dx, dy := textAnchorOffset(tx)
	wantX := -text2path.MeasureAdvance(tx.Value, tx.Height) / 2
	if math.Abs(dx-wantX) > 1e-9 {
		t.Fatalf("center anchor dx=%v want %v", dx, wantX)
	}
	if math.Abs(dy+tx.Height) > 1e-9 {
		t.Fatalf("top anchor dy=%v want %v", dy, -tx.Height)
	}

	tx.Anchor(entity.RIGHT_BASE)
	dx, _ = textAnchorOffset(tx)
	if math.Abs(dx+text2path.MeasureAdvance(tx.Value, tx.Height)) > 1e-9 {
		t.Fatalf("right anchor dx=%v", dx)
	}
}

func TestMTextAlignmentUsesMeasuredWidth(t *testing.T) {
	m := entity.NewMText()
	m.Value = "alpha beta gamma delta epsilon zeta"
	m.Height = 10
	m.Width = 40
	m.AttachmentPoint = entity.MTextTopCenter
	dx, dy := mtextAnchorOffset(m)
	lines := text2path.LayoutMText(m.Value, m.Height, m.Width)
	maxWidth := 0.0
	for _, line := range lines {
		if w := text2path.MeasureAdvance(line, m.Height); w > maxWidth {
			maxWidth = w
		}
	}
	if math.Abs(dx+maxWidth/2) > 1e-9 {
		t.Fatalf("center MTEXT dx=%v want %v", dx, -maxWidth/2)
	}
	if dy != 0 {
		t.Fatalf("top MTEXT dy=%v want 0", dy)
	}

	p, _ := multiLineText(m.Value, m.Height, m.Width, 0, 0, 0, nil)
	if p == nil {
		t.Fatal("no MTEXT paths")
	}
	min, max := p.BoundingBox()
	if max.Y()-min.Y() <= m.Height*1.5 {
		t.Fatalf("wrapped MTEXT extents too small: %v..%v", min.Y(), max.Y())
	}
}
