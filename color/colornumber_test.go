package color

import (
	"testing"
)

func TestColorNumberConstants(t *testing.T) {
	if Red != 1 {
		t.Errorf("Red should be 1, got %d", Red)
	}
	if Yellow != 2 {
		t.Errorf("Yellow should be 2, got %d", Yellow)
	}
	if Green != 3 {
		t.Errorf("Green should be 3, got %d", Green)
	}
	if Cyan != 4 {
		t.Errorf("Cyan should be 4, got %d", Cyan)
	}
	if Blue != 5 {
		t.Errorf("Blue should be 5, got %d", Blue)
	}
	if Magenta != 6 {
		t.Errorf("Magenta should be 6, got %d", Magenta)
	}
	if White != 7 {
		t.Errorf("White should be 7, got %d", White)
	}
	if ByLayer != 255 {
		t.Errorf("ByLayer should be 255, got %d", ByLayer)
	}
}

func TestColorGreyConstants(t *testing.T) {
	if Grey128 != 8 {
		t.Errorf("Grey128 should be 8, got %d", Grey128)
	}
	if Grey192 != 9 {
		t.Errorf("Grey192 should be 9, got %d", Grey192)
	}
	if Grey51 != 250 {
		t.Errorf("Grey51 should be 250, got %d", Grey51)
	}
	if Grey255 != 255 {
		t.Errorf("Grey255 should be 255, got %d", Grey255)
	}
}

func TestColorNumberType(t *testing.T) {
	var cn ColorNumber = 10
	if cn != 10 {
		t.Errorf("ColorNumber should hold value 10, got %d", cn)
	}
}

func TestColorIndexRed(t *testing.T) {
	index := ColorIndex([]int{255, 0, 0})
	if index != Red {
		t.Errorf("Red RGB [255,0,0] should map to color %d, got %d", Red, index)
	}
}

func TestColorIndexGreen(t *testing.T) {
	index := ColorIndex([]int{0, 255, 0})
	if index != Green {
		t.Errorf("Green RGB [0,255,0] should map to color %d, got %d", Green, index)
	}
}

func TestColorIndexBlue(t *testing.T) {
	index := ColorIndex([]int{0, 0, 255})
	if index != Blue {
		t.Errorf("Blue RGB [0,0,255] should map to color %d, got %d", Blue, index)
	}
}

func TestColorIndexWhite(t *testing.T) {
	index := ColorIndex([]int{255, 255, 255})
	if index != White {
		t.Errorf("White RGB [255,255,255] should map to color %d, got %d", White, index)
	}
}

func TestColorIndexBlack(t *testing.T) {
	index := ColorIndex([]int{0, 0, 0})
	if index != 0 {
		t.Errorf("Black RGB [0,0,0] should map to color 0, got %d", index)
	}
}

func TestColorIndexYellow(t *testing.T) {
	index := ColorIndex([]int{255, 255, 0})
	if index != Yellow {
		t.Errorf("Yellow RGB [255,255,0] should map to color %d, got %d", Yellow, index)
	}
}

func TestColorIndexCyan(t *testing.T) {
	index := ColorIndex([]int{0, 255, 255})
	if index != Cyan {
		t.Errorf("Cyan RGB [0,255,255] should map to color %d, got %d", Cyan, index)
	}
}

func TestColorIndexMagenta(t *testing.T) {
	index := ColorIndex([]int{255, 0, 255})
	if index != Magenta {
		t.Errorf("Magenta RGB [255,0,255] should map to color %d, got %d", Magenta, index)
	}
}

func TestIndexColor(t *testing.T) {
	rgb := IndexColor(1)
	if len(rgb) != 3 {
		t.Errorf("IndexColor should return 3 values, got %d", len(rgb))
	}
	if rgb[0] != 255 || rgb[1] != 0 || rgb[2] != 0 {
		t.Errorf("IndexColor(1) should be [255,0,0], got [%d,%d,%d]", rgb[0], rgb[1], rgb[2])
	}

	rgb = IndexColor(5)
	if rgb[0] != 0 || rgb[1] != 0 || rgb[2] != 255 {
		t.Errorf("IndexColor(5) should be [0,0,255], got [%d,%d,%d]", rgb[0], rgb[1], rgb[2])
	}

	rgb = IndexColor(7)
	if rgb[0] != 255 || rgb[1] != 255 || rgb[2] != 255 {
		t.Errorf("IndexColor(7) should be [255,255,255], got [%d,%d,%d]", rgb[0], rgb[1], rgb[2])
	}
}

func TestIndexColorBounds(t *testing.T) {
	rgb := IndexColor(0)
	if rgb[0] != 0 || rgb[1] != 0 || rgb[2] != 0 {
		t.Errorf("IndexColor(0) should be [0,0,0], got [%d,%d,%d]", rgb[0], rgb[1], rgb[2])
	}

	rgb = IndexColor(255)
	if rgb[0] != 255 || rgb[1] != 255 || rgb[2] != 255 {
		t.Errorf("IndexColor(255) should be [255,255,255], got [%d,%d,%d]", rgb[0], rgb[1], rgb[2])
	}
}

func TestColorIndexSimilarColors(t *testing.T) {
	index1 := ColorIndex([]int{200, 0, 0})
	index2 := ColorIndex([]int{204, 0, 0})
	if index1 == index2 {
		t.Logf("Similar colors map to same index: %d", index1)
	}
}

func TestColorRGBLength(t *testing.T) {
	if len(ColorRGB) != 256 {
		t.Errorf("ColorRGB should have 256 entries, got %d", len(ColorRGB))
	}
}
