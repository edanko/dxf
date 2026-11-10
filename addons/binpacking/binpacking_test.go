package binpacking_test

import (
	"testing"

	"github.com/edanko/dxf/addons/binpacking"
)

func TestPackRectanglesNoOverlapWithinBounds(t *testing.T) {
	rects := []binpacking.Rectangle{
		{Width: 2, Height: 2},
		{Width: 2, Height: 2},
		{Width: 3, Height: 1},
	}
	packed, err := binpacking.PackRectangles(rects, 5, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(packed) != 3 {
		t.Fatalf("packed: %d", len(packed))
	}
	for i := range packed {
		p := packed[i]
		if p.Pos.X < 0 || p.Pos.Y < 0 || p.Pos.X+p.Rect.Width > 5 || p.Pos.Y+p.Rect.Height > 5 {
			t.Fatalf("item %d out of bounds: %+v", i, p)
		}
		for j := i + 1; j < len(packed); j++ {
			if binpacking.Overlaps(p, packed[j]) {
				t.Fatalf("items %d and %d overlap", i, j)
			}
		}
	}
}

func TestPackRectanglesErrors(t *testing.T) {
	if _, err := binpacking.PackRectangles([]binpacking.Rectangle{{Width: 6, Height: 1}}, 5, 5); err == nil {
		t.Fatal("oversized item must error")
	}
	if _, err := binpacking.PackRectangles([]binpacking.Rectangle{{Width: 4, Height: 4}, {Width: 4, Height: 4}}, 5, 5); err == nil {
		t.Fatal("overflow must error")
	}
}
