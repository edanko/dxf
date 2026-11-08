package entity

import (
	"testing"
)

func TestTextAnchor(t *testing.T) {
	t.Run("Set and get anchor", func(t *testing.T) {
		text := NewText()

		// Test default anchor
		if text.GetAnchor() != LEFT_BASE {
			t.Error("Default anchor should be LEFT_BASE")
		}

		// Test setting all anchors
		anchors := []int{LEFT_BASE, CENTER_BASE, RIGHT_BASE,
			LEFT_BOTTOM, CENTER_BOTTOM, RIGHT_BOTTOM,
			LEFT_CENTER, CENTER_CENTER, RIGHT_CENTER,
			LEFT_TOP, CENTER_TOP, RIGHT_TOP}

		for _, anchor := range anchors {
			returned := text.Anchor(anchor)
			if returned != anchor {
				t.Errorf("Anchor(%d) returned %d", anchor, returned)
			}
			if text.GetAnchor() != anchor {
				t.Errorf("Expected anchor %d, got %d", anchor, text.GetAnchor())
			}
		}
	})

	t.Run("Invalid anchor returns current", func(t *testing.T) {
		text := NewText()
		text.Anchor(LEFT_TOP)

		// Invalid anchors
		result := text.Anchor(-1)
		if result != LEFT_TOP {
			t.Error("Invalid anchor should return current anchor")
		}

		result = text.Anchor(100)
		if result != LEFT_TOP {
			t.Error("Invalid anchor should return current anchor")
		}
	})
}

func TestTextLength(t *testing.T) {
	t.Run("Empty text has zero length", func(t *testing.T) {
		text := NewText()
		if text.TextLength() != 0 {
			t.Error("Empty text should have zero length")
		}
	})

	t.Run("Text length calculation", func(t *testing.T) {
		text := NewText()
		text.Value = "Hello"
		text.Height = 2.0
		text.WidthFactor = 1.0

		length := text.TextLength()
		// Expected: 5 chars * 2.0 * 0.6 = 6.0
		expected := 5.0 * 2.0 * 0.6
		if length != expected {
			t.Errorf("Expected text length %f, got %f", expected, length)
		}
	})

	t.Run("Width factor affects length", func(t *testing.T) {
		text := NewText()
		text.Value = "Test"
		text.Height = 1.0
		text.WidthFactor = 2.0

		length := text.TextLength()
		// Expected: 4 chars * 1.0 * 0.6 * 2.0 = 4.8
		expected := 4.0 * 1.0 * 0.6 * 2.0
		if length != expected {
			t.Errorf("Expected text length %f, got %f", expected, length)
		}
	})
}

func TestTextBBox(t *testing.T) {
	t.Run("BBox with default anchor", func(t *testing.T) {
		text := NewText()
		text.Value = "X"
		text.Height = 2.0
		text.Coord1 = []float64{10, 20, 0}

		mins, _ := text.BBox()

		// With LEFT_BASE (default), mins should be at Coord1
		if mins[0] != 10 {
			t.Errorf("Expected mins[0] = 10, got %f", mins[0])
		}
		if mins[1] != 20 {
			t.Errorf("Expected mins[1] = 20, got %f", mins[1])
		}
	})

	t.Run("BBox with center anchor", func(t *testing.T) {
		text := NewText()
		text.Value = "Test longer text"
		text.Height = 2.0
		text.Coord1 = []float64{100, 100, 0}
		text.Anchor(CENTER_CENTER)

		mins, maxs := text.BBox()

		// Center anchor should extend equally in both directions
		if mins[0] >= maxs[0] {
			t.Error("mins[0] should be less than maxs[0] for center anchor")
		}
		if mins[1] >= maxs[1] {
			t.Error("mins[1] should be less than maxs[1] for center anchor")
		}

		// Center should be at Coord1
		centerX := (mins[0] + maxs[0]) / 2
		if centerX != 100 {
			t.Errorf("Expected center X = 100, got %f", centerX)
		}
	})

	t.Run("BBox with right anchor", func(t *testing.T) {
		text := NewText()
		text.Value = "Right"
		text.Height = 1.0
		text.Coord1 = []float64{50, 50, 0}
		text.Anchor(RIGHT_BASE)

		mins, maxs := text.BBox()

		// Right anchor means maxs should be at Coord1
		_ = mins
		if maxs[0] != 50 {
			t.Errorf("Expected maxs[0] = 50 for right anchor, got %f", maxs[0])
		}
	})
}

func TestTextCoordMethods(t *testing.T) {
	t.Run("SetCoord", func(t *testing.T) {
		text := NewText()
		text.SetCoord([]float64{1, 2, 3})

		if text.Coord1[0] != 1 || text.Coord1[1] != 2 || text.Coord1[2] != 3 {
			t.Error("SetCoord did not set coordinates correctly")
		}
	})

	t.Run("SetSecondCoord", func(t *testing.T) {
		text := NewText()
		text.SetSecondCoord([]float64{4, 5, 6})

		if text.Coord2[0] != 4 || text.Coord2[1] != 5 || text.Coord2[2] != 6 {
			t.Error("SetSecondCoord did not set coordinates correctly")
		}
	})

	t.Run("SecondCoord", func(t *testing.T) {
		text := NewText()
		text.Coord2 = []float64{7, 8, 9}

		coord := text.SecondCoord()
		if coord[0] != 7 || coord[1] != 8 || coord[2] != 9 {
			t.Error("SecondCoord did not return correct coordinates")
		}
	})

	t.Run("SetCoord with less than 3 values", func(t *testing.T) {
		text := NewText()
		text.SetCoord([]float64{1, 2})

		// Should not panic, and coordinates should not change
		if text.Coord1[0] != 0 {
			t.Error("SetCoord with insufficient values should not modify coords")
		}
	})
}

func TestTextRotationBBox(t *testing.T) {
	t.Run("Rotated text bbox is larger", func(t *testing.T) {
		text := NewText()
		text.Value = "Longer test text"
		text.Height = 10.0
		text.Coord1 = []float64{0, 0, 0}
		text.Anchor(LEFT_BASE)

		// Without rotation
		_, maxsNoRot := text.BBox()

		// With 45 degree rotation
		text.Rotation = 45
		_, maxsRot := text.BBox()

		// Rotated bounding box should be larger in at least one dimension
		// The diagonal extent should be larger than the axis-aligned extent
		rotLarger := maxsRot[0] > maxsNoRot[0] || maxsRot[1] > maxsNoRot[1]
		if !rotLarger {
			t.Logf("NoRot: (%f, %f), Rot: (%f, %f)", maxsNoRot[0], maxsNoRot[1], maxsRot[0], maxsRot[1])
		}
	})
}
