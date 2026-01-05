package path

import (
	"math"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func TestPathTools(t *testing.T) {
	// Create a simple path
	path := NewPath()
	path.MoveTo(0, 0)
	path.LineTo(10, 0)
	path.LineTo(10, 10)
	path.LineTo(0, 10)
	path.Close()

	tools := NewTools(path)

	// Test bounding box
	t.Run("BoundingBox", func(t *testing.T) {
		bbox := tools.BoundingBox()
		min := bbox.Min
		max := bbox.Max

		if min.X() != 0 || min.Y() != 0 {
			t.Errorf("Expected min (0, 0), got (%.2f, %.2f)", min.X(), min.Y())
		}
		if max.X() != 10 || max.Y() != 10 {
			t.Errorf("Expected max (10, 10), got (%.2f, %.2f)", max.X(), max.Y())
		}
	})

	// Test length
	t.Run("Length", func(t *testing.T) {
		length := tools.Length()
		expected := 40.0 // 10 + 10 + 10 + 10 (square perimeter)
		if math.Abs(length-expected) > 0.1 {
			t.Errorf("Expected length %.1f, got %.1f", expected, length)
		}
	})

	// Test is closed
	t.Run("IsClosed", func(t *testing.T) {
		if !tools.IsClosed() {
			t.Error("Expected path to be closed")
		}
	})

	// Test simplification
	t.Run("Simplify", func(t *testing.T) {
		// Create path with extra points
		complexPath := NewPath()
		complexPath.MoveTo(0, 0)
		complexPath.LineTo(5, 0) // Extra point
		complexPath.LineTo(10, 0)
		complexPath.LineTo(10, 10)
		complexPath.LineTo(0, 10)
		complexPath.Close()

		complexTools := NewTools(complexPath)
		simplified := complexTools.Simplify(1.0)

		// Should have fewer points after simplification
		simplifiedVerts := simplified.Vertices()
		complexVerts := complexPath.Vertices()
		if len(simplifiedVerts) >= len(complexVerts) {
			t.Errorf("Simplification should reduce vertex count")
		}
	})

	// Test transformation
	t.Run("Transform", func(t *testing.T) {
		// Create translation matrix
		matrix := dxfmath.Translation(10, 20, 0)

		transformed := tools.Transform(&matrix)

		// Debug: Check if transformation is working
		commands := transformed.Commands()
		if len(commands) > 0 {
			if moveTo, ok := commands[0].(*PathMoveTo); ok {
				t.Logf("Transformed MoveTo: (%.2f, %.2f)", moveTo.End.X(), moveTo.End.Y())
			}
		}

		bbox := NewTools(transformed).BoundingBox()
		min := bbox.Min
		max := bbox.Max

		// Should be translated by (10, 20)
		if min.X() != 10 || min.Y() != 20 {
			t.Errorf("Expected translated min (10, 20), got (%.2f, %.2f)", min.X(), min.Y())
		}
		if max.X() != 20 || max.Y() != 30 {
			t.Errorf("Expected translated max (20, 30), got (%.2f, %.2f)", max.X(), max.Y())
		}
	})

	// Test curve path
	t.Run("CurvePath", func(t *testing.T) {
		curvePath := NewPath()
		curvePath.MoveTo(0, 0)
		curvePath.Curve3To(10, 10, 5, 0) // Quadratic Bezier

		curveTools := NewTools(curvePath)
		length := curveTools.Length()

		if length <= 10 { // Should be longer than straight line
			t.Errorf("Curve length should be greater than straight line distance")
		}

		// Test cubic curve
		curvePath.Curve4To(20, 10, 15, 15, 10, 5)
		cubicLength := curveTools.Length()

		if cubicLength <= 10 { // Should be longer than straight line
			t.Errorf("Cubic curve length should be greater than straight line distance")
		}
	})

	// Test offset
	t.Run("Offset", func(t *testing.T) {
		offsetPaths := tools.Offset(2.0)
		if len(offsetPaths) == 0 {
			t.Error("Offset should return at least one path")
			return
		}
		offsetPath := offsetPaths[0]
		bbox := NewTools(offsetPath).BoundingBox()
		min := bbox.Min
		max := bbox.Max

		// Offset should expand the bounding box
		if min.X() >= 0 || min.Y() >= 0 {
			t.Errorf("Offset should expand bounds in negative direction")
		}
		if max.X() <= 10 || max.Y() <= 10 {
			t.Errorf("Offset should expand bounds in positive direction")
		}
	})
}

func TestPathConversion(t *testing.T) {
	t.Run("CircleToPath", func(t *testing.T) {
		// Create a circle path manually
		circlePath := NewPath()
		segments := 16
		radius := 5.0

		// Start point
		startX := radius
		startY := 0.0
		circlePath.MoveTo(startX, startY)

		// Generate circle points
		for i := 1; i <= segments; i++ {
			angle := 2 * math.Pi * float64(i) / float64(segments)
			x := radius * math.Cos(angle)
			y := radius * math.Sin(angle)
			circlePath.LineTo(x, y)
		}

		tools := NewTools(circlePath)
		bbox := tools.BoundingBox()
		min := bbox.Min
		max := bbox.Max

		// Circle should fit in a square from (-5, -5) to (5, 5)
		if min.X() < -5.1 || min.Y() < -5.1 {
			t.Errorf("Circle bounds too small: min (%.2f, %.2f)", min.X(), min.Y())
		}
		if max.X() > 5.1 || max.Y() > 5.1 {
			t.Errorf("Circle bounds too large: max (%.2f, %.2f)", max.X(), max.Y())
		}
	})

	t.Run("PolylineToPath", func(t *testing.T) {
		// Create polyline path
		polyPath := NewPath()
		points := [][]float64{
			{0, 0},
			{10, 0},
			{10, 10},
			{0, 10},
			{0, 0}, // closed
		}

		// Add points manually
		for i, point := range points {
			if i == 0 {
				polyPath.MoveTo(point[0], point[1])
			} else {
				polyPath.LineTo(point[0], point[1])
			}
		}

		tools := NewTools(polyPath)
		if !tools.IsClosed() {
			t.Error("Polyline should be closed")
		}

		length := tools.Length()
		expected := 40.0 // perimeter of square
		if math.Abs(length-expected) > 0.1 {
			t.Errorf("Expected polyline length %.1f, got %.1f", expected, length)
		}
	})
}

func TestPathOperations(t *testing.T) {
	t.Run("ReversePath", func(t *testing.T) {
		// Create path
		path := NewPath()
		path.MoveTo(0, 0)
		path.LineTo(10, 0)
		path.LineTo(10, 10)

		tools := NewTools(path)
		reversed := tools.Reverse()

		// Reversed path should start from end of original
		commands := reversed.Commands()
		if len(commands) < 2 {
			t.Error("Reversed path should have commands")
		}

		// First command should be MoveTo to original end point
		if cmd, ok := commands[0].(*PathMoveTo); ok {
			if math.Abs(cmd.End.X()-10) > 0.01 || math.Abs(cmd.End.Y()-10) > 0.01 {
				t.Errorf("Reversed path should start from original end point")
			}
		} else {
			t.Error("First command should be MoveTo")
		}
	})

	t.Run("DouglasPeucker", func(t *testing.T) {
		// Test line simplification
		points := []dxfmath.Vec2{
			dxfmath.NewVec2(0, 0),
			dxfmath.NewVec2(5, 0), // Collinear - should be removed
			dxfmath.NewVec2(10, 0),
			dxfmath.NewVec2(10, 5), // Off-collinear - should be kept
			dxfmath.NewVec2(10, 10),
		}

		tools := NewTools(NewPath())
		simplified := tools.douglasPeucker(points, 0.1)

		// Should have fewer points
		if len(simplified) >= len(points) {
			t.Error("Simplification should reduce point count")
		}

		// Should preserve start and end points
		if simplified[0].Distance(points[0]) > 0.01 {
			t.Error("Should preserve start point")
		}
		lastPoint := simplified[len(simplified)-1]
		if lastPoint.Distance(points[len(points)-1]) > 0.01 {
			t.Error("Should preserve end point")
		}
	})
}
