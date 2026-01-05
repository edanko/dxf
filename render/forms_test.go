package render

import (
	"math"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func TestFormGenerator(t *testing.T) {
	fg := NewFormGenerator()

	// Test circle
	t.Run("Circle", func(t *testing.T) {
		circle := fg.Circle(5.0)
		if len(circle) != 32 {
			t.Errorf("Expected 32 points for circle, got %d", len(circle))
		}
	})

	// Test square
	t.Run("Square", func(t *testing.T) {
		square := fg.Square(10.0)
		if len(square) != 4 {
			t.Errorf("Expected 4 points for square, got %d", len(square))
		}

		// Test closed square
		closedSquare := fg.Square(10.0, true)
		if len(closedSquare) != 5 {
			t.Errorf("Expected 5 points for closed square, got %d", len(closedSquare))
		}
	})

	// Test triangle
	t.Run("Triangle", func(t *testing.T) {
		triangle := fg.Triangle(10.0)
		if len(triangle) != 3 {
			t.Errorf("Expected 3 points for triangle, got %d", len(triangle))
		}

		// Test closed triangle
		closedTriangle := fg.Triangle(10.0, true)
		if len(closedTriangle) != 4 {
			t.Errorf("Expected 4 points for closed triangle, got %d", len(closedTriangle))
		}
	})

	// Test star
	t.Run("Star", func(t *testing.T) {
		star := fg.Star(5.0, 2.5, 5)
		if len(star) != 12 {
			t.Errorf("Expected 12 points for star, got %d", len(star))
		}
	})

	// Test ngon
	t.Run("NGon", func(t *testing.T) {
		ngon := fg.NGon(6, 5.0)
		if len(ngon) != 6 {
			t.Errorf("Expected 6 points for hexagon, got %d", len(ngon))
		}
	})

	// Test arrow
	t.Run("Arrow", func(t *testing.T) {
		arrow := fg.Arrow(10.0, 3.0)
		if len(arrow) != 5 {
			t.Errorf("Expected 5 points for arrow, got %d", len(arrow))
		}

		// Test closed arrow
		closedArrow := fg.Arrow(10.0, 3.0, true)
		if len(closedArrow) != 6 {
			t.Errorf("Expected 6 points for closed arrow, got %d", len(closedArrow))
		}
	})

	// Test helix
	t.Run("Helix", func(t *testing.T) {
		helix := fg.Helix(5.0, 10.0, 3, 8)
		if len(helix) != 24 {
			t.Errorf("Expected 24 points for helix, got %d", len(helix))
		}
	})

	// Test transformations
	t.Run("Transformations", func(t *testing.T) {
		shape := fg.Circle(5.0)

		// Test translate
		translated := fg.Translate(shape, dxfmath.NewVec3(10, 10, 0))
		if len(translated) != 32 {
			t.Errorf("Expected 32 points after translation, got %d", len(translated))
		}

		// Test scale
		scaled := fg.Scale(shape, 2.0)
		if len(scaled) != 32 {
			t.Errorf("Expected 32 points after scaling, got %d", len(scaled))
		}
	})

	// Test turtle graphics
	t.Run("Turtle", func(t *testing.T) {
		path := fg.Turtle(func(t *Turtle) {
			t.Move(10)
			t.Left(math.Pi / 2) // 90 degrees
			t.Move(10)
			t.Left(math.Pi / 2)
			t.Move(10)
			t.Left(math.Pi / 2)
			t.Move(10)
		})

		// Should have 4 points for a square
		if len(path) != 4 {
			t.Errorf("Expected 4 points for turtle square, got %d", len(path))
		}
	})

	// Test box function
	t.Run("Box", func(t *testing.T) {
		box := fg.Box(10.0, 5.0)
		if len(box) != 4 {
			t.Errorf("Expected 4 points for box, got %d", len(box))
		}

		// Test closed box
		closedBox := fg.Box(10.0, 5.0, true)
		if len(closedBox) != 5 {
			t.Errorf("Expected 5 points for closed box, got %d", len(closedBox))
		}
	})

	// Test close polygon utility
	t.Run("ClosePolygon", func(t *testing.T) {
		triangle := fg.Triangle(10.0)
		closedTriangle := fg.ClosePolygon(triangle)
		if len(closedTriangle) != 4 {
			t.Errorf("Expected 4 points for closed triangle, got %d", len(closedTriangle))
		}

		// Test already closed polygon
		square := fg.Square(10.0, true)
		alreadyClosed := fg.ClosePolygon(square)
		if len(alreadyClosed) != 5 {
			t.Errorf("Expected 5 points for already closed square, got %d", len(alreadyClosed))
		}
	})

	// Test rotation form
	t.Run("RotationForm", func(t *testing.T) {
		profile := []dxfmath.Vec3{
			dxfmath.NewVec3(1, 0, 0),
			dxfmath.NewVec3(2, 0, 0),
		}

		faces := fg.RotationForm(profile, 360, 8, "z")
		if len(faces) != 16 {
			t.Errorf("Expected 16 faces for rotation form, got %d", len(faces))
		}
	})

	// Test extrude with twist and scale
	t.Run("ExtrudeTwistScale", func(t *testing.T) {
		profile := []dxfmath.Vec3{
			dxfmath.NewVec3(1, 1, 0),
			dxfmath.NewVec3(1, -1, 0),
			dxfmath.NewVec3(-1, -1, 0),
			dxfmath.NewVec3(-1, 1, 0),
		}

		faces := fg.ExtrudeTwistScale(profile, 10, 90, 1.5, 4)
		if len(faces) != 16 { // 4 profiles * 4 faces each
			t.Errorf("Expected 16 faces for twisted extrusion, got %d", len(faces))
		}
	})
}
