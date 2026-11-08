package entity

import (
	"math"
	"testing"
)

func TestCircleBBox(t *testing.T) {
	t.Run("Default extrusion (Z axis)", func(t *testing.T) {
		circle := NewCircle()
		circle.Center = []float64{10, 20, 30}
		circle.Radius = 5.0
		// Direction defaults to [0, 0, 1]

		mins, maxs := circle.BBox()

		if mins[0] != 5 || mins[1] != 15 || mins[2] != 30 {
			t.Errorf("Unexpected mins: %v", mins)
		}
		if maxs[0] != 15 || maxs[1] != 25 || maxs[2] != 30 {
			t.Errorf("Unexpected maxs: %v", maxs)
		}
	})

	t.Run("Custom extrusion direction", func(t *testing.T) {
		circle := NewCircle()
		circle.Center = []float64{0, 0, 0}
		circle.Radius = 10.0
		// Extrusion along X axis
		circle.Direction = []float64{1, 0, 0}

		mins, maxs := circle.BBox()

		// Circle should be in the YZ plane, so X extent should be 0
		// (the circle has no thickness in the direction of the extrusion)
		xExtent := maxs[0] - mins[0]
		if xExtent > 0.1 {
			t.Errorf("X extent should be ~0 for X-axis extrusion, got %f", xExtent)
		}

		// Y and Z should extend to roughly +/-10
		if mins[1] > -10 || maxs[1] < 10 {
			t.Errorf("Y extent should be +/-10 for X-axis extrusion, got mins[1]=%f, maxs[1]=%f", mins[1], maxs[1])
		}
		if mins[2] > -10 || maxs[2] < 10 {
			t.Errorf("Z extent should be +/-10 for X-axis extrusion, got mins[2]=%f, maxs[2]=%f", mins[2], maxs[2])
		}
	})

	t.Run("Diagonal extrusion direction", func(t *testing.T) {
		circle := NewCircle()
		circle.Center = []float64{0, 0, 0}
		circle.Radius = 10.0
		// Extrusion at 45 degrees in XY plane
		circle.Direction = []float64{1, 1, 0}
		circle.Direction = normalize(circle.Direction)

		mins, maxs := circle.BBox()

		// For diagonal extrusion in XY plane, Z extent should be the full diameter
		// because the circle is perpendicular to Z
		zExtent := maxs[2] - mins[2]
		if zExtent < 19.9 || zExtent > 20.1 {
			t.Errorf("Z extent should be ~20 for XY-plane extrusion, got %.2f", zExtent)
		}

		// X and Y extents should be ~14.14 = 10*sqrt(2)
		xExtent := maxs[0] - mins[0]
		yExtent := maxs[1] - mins[1]
		expected := 10.0 * math.Sqrt(2)
		if math.Abs(xExtent-expected) > 1.0 {
			t.Errorf("X extent should be ~%.2f, got %.2f", expected, xExtent)
		}
		if math.Abs(yExtent-expected) > 1.0 {
			t.Errorf("Y extent should be ~%.2f, got %.2f", expected, yExtent)
		}
	})

	t.Run("Diagonal extrusion (1,1,1)", func(t *testing.T) {
		circle := NewCircle()
		circle.Center = []float64{0, 0, 0}
		circle.Radius = 10.0
		circle.Direction = []float64{1, 1, 1}
		circle.Direction = normalize(circle.Direction)

		mins, maxs := circle.BBox()

		// For (1,1,1) direction, the circle is equally distributed
		// The actual extent depends on how the sampling works
		// For a circle in a plane, the bounding box is determined by sampling
		zExtent := maxs[2] - mins[2]
		// Just verify it's between 15 and 17
		if zExtent < 15 || zExtent > 17 {
			t.Errorf("Z extent should be ~15-17 for (1,1,1) extrusion, got %.2f", zExtent)
		}
	})

	t.Run("Circle at origin with default extrusion", func(t *testing.T) {
		circle := NewCircle()
		circle.Center = []float64{0, 0, 0}
		circle.Radius = 1.0

		mins, maxs := circle.BBox()

		if mins[0] != -1 || mins[1] != -1 || mins[2] != 0 {
			t.Errorf("Unexpected mins: %v", mins)
		}
		if maxs[0] != 1 || maxs[1] != 1 || maxs[2] != 0 {
			t.Errorf("Unexpected maxs: %v", maxs)
		}
	})

	t.Run("Nil center returns zeros", func(t *testing.T) {
		circle := NewCircle()
		circle.Center = nil

		mins, maxs := circle.BBox()

		for i := 0; i < 3; i++ {
			if mins[i] != 0 || maxs[i] != 0 {
				t.Errorf("Expected zeros for nil center, got mins=%v, maxs=%v", mins, maxs)
			}
		}
	})

	t.Run("Empty direction uses default", func(t *testing.T) {
		circle := NewCircle()
		circle.Center = []float64{5, 5, 5}
		circle.Radius = 2.0
		circle.Direction = nil

		mins, maxs := circle.BBox()

		if mins[0] != 3 || mins[1] != 3 || mins[2] != 5 {
			t.Errorf("Unexpected mins with nil direction: %v", mins)
		}
		if maxs[0] != 7 || maxs[1] != 7 || maxs[2] != 5 {
			t.Errorf("Unexpected maxs with nil direction: %v", maxs)
		}
	})
}

func normalize(v []float64) []float64 {
	if len(v) < 3 {
		return v
	}
	len := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if len == 0 {
		return v
	}
	return []float64{v[0] / len, v[1] / len, v[2] / len}
}
