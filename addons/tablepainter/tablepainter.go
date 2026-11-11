// Package tablepainter renders tables as line grids with text cells.
package tablepainter

import (
	"fmt"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
)

type capability struct{}

func (capability) Name() string        { return "tablepainter" }
func (capability) Description() string { return "table rendering as line grids with text cells" }

func init() { _ = addons.Register("tablepainter", capability{}) }

// PaintTable draws a grid of cols x rows cells at (x, y) with the given cell
// size and fills non-empty cell strings with TEXT entities.
func PaintTable(d *drawing.Drawing, x, y, colW, rowH float64, cells [][]string) error {
	if d == nil {
		return fmt.Errorf("tablepainter: nil drawing")
	}
	if colW <= 0 || rowH <= 0 {
		return fmt.Errorf("tablepainter: cell size must be positive")
	}
	if len(cells) == 0 {
		return fmt.Errorf("tablepainter: no rows")
	}
	cols := len(cells[0])
	ms := d.Model()
	for c := 0; c <= cols; c++ {
		if _, err := ms.AddLine(x+float64(c)*colW, y, 0, x+float64(c)*colW, y+float64(len(cells))*rowH, 0); err != nil {
			return err
		}
	}
	for r := 0; r <= len(cells); r++ {
		if _, err := ms.AddLine(x, y+float64(r)*rowH, 0, x+float64(cols)*colW, y+float64(r)*rowH, 0); err != nil {
			return err
		}
	}
	for r, row := range cells {
		for c, text := range row {
			if text == "" {
				continue
			}
			if _, err := ms.AddText(x+float64(c)*colW+0.1*colW, y+float64(r)*rowH+0.2*rowH, 0, 0.5*rowH, text); err != nil {
				return err
			}
		}
	}
	return nil
}
