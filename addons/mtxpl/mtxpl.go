// Package mtxpl lays out MTEXT content and explodes it into TEXT entities.
package mtxpl

import (
	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/text2path"
)

func init() { _ = addons.Register("mtxpl", capability{}) }

type capability struct{}

func (capability) Name() string        { return "mtxpl" }
func (capability) Description() string { return "MTEXT line layout helpers and TEXT explosion" }

// LinePitch is the vertical distance between MTEXT lines relative to the
// text height.
const LinePitch = 1.6714

// MTextLine is one laid-out line of an MTEXT entity.
type MTextLine struct {
	Text   string
	X, Y   float64
	Height float64
}

// Lines lays out the text of m top-down from its insertion point.
func Lines(m *entity.MText) []MTextLine {
	if m == nil || m.Height <= 0 || len(m.Coord1) < 2 {
		return nil
	}
	text := text2path.StripCodes(m.Value)
	lines := text2path.LayoutMText(text, m.Height, m.Width)
	out := make([]MTextLine, 0, len(lines))
	for i, ln := range lines {
		out = append(out, MTextLine{
			Text:   ln,
			X:      m.Coord1[0],
			Y:      m.Coord1[1] - float64(i)*LinePitch*m.Height,
			Height: m.Height,
		})
	}
	return out
}

// Explode converts m into positioned TEXT entities, one per laid-out line.
func Explode(m *entity.MText) []*entity.Text {
	lines := Lines(m)
	out := make([]*entity.Text, 0, len(lines))
	z := 0.0
	if len(m.Coord1) > 2 {
		z = m.Coord1[2]
	}
	for _, ln := range lines {
		t := entity.NewText()
		t.Coord1 = []float64{ln.X, ln.Y, z}
		t.Height = ln.Height
		t.Value = ln.Text
		t.Rotation = m.Rotation
		if l := m.Layer(); l != nil {
			t.SetLayer(l)
		}
		out = append(out, t)
	}
	return out
}
