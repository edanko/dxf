package geometry

import (
	"math"
	"testing"
)

func TestArbitraryAxis(t *testing.T) {
	t.Run("ZAxisUp", func(t *testing.T) {
		// Test with Z axis up (0, 0, 1)
		ax, ay, err := ArbitraryAxis([]float64{0, 0, 1})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if ax == nil || ay == nil {
			t.Error("ax and ay should not be nil")
		}

		// ax should be [1, 0, 0] when d is [0, 0, 1]
		if len(ax) != 3 || len(ay) != 3 {
			t.Error("ax and ay should have length 3")
		}
	})

	t.Run("XAxis", func(t *testing.T) {
		// Test with X axis (1, 0, 0)
		ax, ay, err := ArbitraryAxis([]float64{1, 0, 0})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if ax == nil || ay == nil {
			t.Error("ax and ay should not be nil")
		}
	})

	t.Run("YAxis", func(t *testing.T) {
		// Test with Y axis (0, 1, 0)
		ax, ay, err := ArbitraryAxis([]float64{0, 1, 0})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if ax == nil || ay == nil {
			t.Error("ax and ay should not be nil")
		}
	})

	t.Run("Diagonal", func(t *testing.T) {
		// Test with diagonal vector
		ax, ay, err := ArbitraryAxis([]float64{1, 1, 1})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if ax == nil || ay == nil {
			t.Error("ax and ay should not be nil")
		}
	})
}

func TestArbitraryAxisError(t *testing.T) {
	t.Run("EmptySlice", func(t *testing.T) {
		_, _, err := ArbitraryAxis([]float64{})
		if err == nil {
			t.Error("Expected error for empty slice")
		}
	})

	t.Run("ShortSlice", func(t *testing.T) {
		_, _, err := ArbitraryAxis([]float64{1, 2})
		if err == nil {
			t.Error("Expected error for slice with less than 3 elements")
		}
	})

	t.Run("TwoElements", func(t *testing.T) {
		_, _, err := ArbitraryAxis([]float64{1, 2})
		if err == nil {
			t.Error("Expected error for slice with 2 elements")
		}
	})
}

func TestArbitraryAxisThreshold(t *testing.T) {
	// Test with vectors near the threshold
	// Threshold is 1/64 = 0.015625

	// Small X and Y components (should trigger first branch)
	ax, ay, err := ArbitraryAxis([]float64{0.01, 0.01, 1.0})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ax == nil || ay == nil {
		t.Error("ax and ay should not be nil")
	}
}

func TestArbitraryAxisVectors(t *testing.T) {
	// Test that ax and ay are orthogonal
	ax, ay, err := ArbitraryAxis([]float64{0, 0, 1})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Dot product of ax and ay should be 0
	dot := ax[0]*ay[0] + ax[1]*ay[1] + ax[2]*ay[2]
	if math.Abs(dot) > 1e-10 {
		t.Errorf("ax and ay should be orthogonal, dot product = %f", dot)
	}
}

func TestArbitraryAxisUnitVectors(t *testing.T) {
	// Test that result vectors have reasonable magnitude
	ax, ay, err := ArbitraryAxis([]float64{0, 0, 1})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check magnitude of ax (should be close to 1)
	axMag := math.Sqrt(ax[0]*ax[0] + ax[1]*ax[1] + ax[2]*ax[2])
	if math.Abs(axMag-1.0) > 1e-10 {
		t.Errorf("ax should have magnitude close to 1, got %f", axMag)
	}

	// Check magnitude of ay (should be close to 1)
	ayMag := math.Sqrt(ay[0]*ay[0] + ay[1]*ay[1] + ay[2]*ay[2])
	if math.Abs(ayMag-1.0) > 1e-10 {
		t.Errorf("ay should have magnitude close to 1, got %f", ayMag)
	}
}

func TestArbitraryAxisConsistency(t *testing.T) {
	// Test that the same input gives the same output
	ax1, ay1, err1 := ArbitraryAxis([]float64{0, 0, 1})
	if err1 != nil {
		t.Fatalf("Unexpected error: %v", err1)
	}

	ax2, ay2, err2 := ArbitraryAxis([]float64{0, 0, 1})
	if err2 != nil {
		t.Fatalf("Unexpected error: %v", err2)
	}

	for i := range 3 {
		if ax1[i] != ax2[i] {
			t.Errorf("ax[%d] differs: %f vs %f", i, ax1[i], ax2[i])
		}
		if ay1[i] != ay2[i] {
			t.Errorf("ay[%d] differs: %f vs %f", i, ay1[i], ay2[i])
		}
	}
}

func TestArbitraryAxisDifferentInputs(t *testing.T) {
	// Different inputs should give different outputs
	ax1, ay1, _ := ArbitraryAxis([]float64{0, 0, 1})
	ax2, ay2, _ := ArbitraryAxis([]float64{1, 0, 0})

	// At least some elements should be different
	different := false
	for i := range 3 {
		if ax1[i] != ax2[i] || ay1[i] != ay2[i] {
			different = true
			break
		}
	}

	if !different {
		t.Error("Different inputs should give different outputs")
	}
}
