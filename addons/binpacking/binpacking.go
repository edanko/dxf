// Package binpacking packs rectangles into a bin using shelf packing.
package binpacking

import "fmt"

// Rectangle is an item to pack.
type Rectangle struct {
	Width  float64
	Height float64
}

// Position is the lower-left corner of a packed item.
type Position struct {
	X float64
	Y float64
}

// Packed pairs an item with its placement.
type Packed struct {
	Rect Rectangle
	Pos  Position
}

// PackRectangles shelves the rectangles into a bin of binW x binH: items are
// placed left to right; when an item no longer fits the current shelf a new
// shelf is started above it. It returns an error when an item exceeds the bin
// or the bin is full.
func PackRectangles(rects []Rectangle, binW, binH float64) ([]Packed, error) {
	if binW <= 0 || binH <= 0 {
		return nil, fmt.Errorf("binpacking: bin size must be positive")
	}
	var out []Packed
	x, y, shelfH := 0.0, 0.0, 0.0
	for i, r := range rects {
		if r.Width <= 0 || r.Height <= 0 {
			return nil, fmt.Errorf("binpacking: rectangle %d has non-positive size", i)
		}
		if r.Width > binW || r.Height > binH {
			return nil, fmt.Errorf("binpacking: rectangle %d (%vx%v) exceeds the bin", i, r.Width, r.Height)
		}
		if x+r.Width > binW {
			x = 0
			y += shelfH
			shelfH = 0
		}
		if y+r.Height > binH {
			return nil, fmt.Errorf("binpacking: rectangle %d does not fit (bin full)", i)
		}
		out = append(out, Packed{Rect: r, Pos: Position{X: x, Y: y}})
		x += r.Width
		if r.Height > shelfH {
			shelfH = r.Height
		}
	}
	return out, nil
}

// Overlaps reports whether two packed items overlap.
func Overlaps(a, b Packed) bool {
	return a.Pos.X < b.Pos.X+b.Rect.Width && b.Pos.X < a.Pos.X+a.Rect.Width &&
		a.Pos.Y < b.Pos.Y+b.Rect.Height && b.Pos.Y < a.Pos.Y+a.Rect.Height
}
