package path

import (
	"math"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func TestPathBasicOperations(t *testing.T) {
	p := NewPath()

	// Test empty path
	if len(p.Commands()) != 0 {
		t.Errorf("Expected empty path, got %d commands", len(p.Commands()))
	}

	// Test basic path construction
	p.MoveTo(0, 0)
	p.LineTo(10, 0)
	p.LineTo(10, 10)
	p.LineTo(0, 10)
	p.Close()

	commands := p.Commands()
	if len(commands) != 5 {
		t.Errorf("Expected 5 commands, got %d", len(commands))
	}

	// Test vertices
	vertices := p.Vertices()
	if len(vertices) != 5 {
		t.Errorf("Expected 5 vertices, got %d", len(vertices))
	}

	// Test properties
	if !p.HasLines() {
		t.Error("Expected path to have lines")
	}

	if p.HasCurves() {
		t.Error("Expected path to have no curves")
	}

	if !p.IsClosed() {
		t.Error("Expected path to be closed")
	}

	// Test bounding box
	min, max := p.BoundingBox()
	if math.Abs(min.X()-0) > 1e-6 || math.Abs(min.Y()-0) > 1e-6 {
		t.Errorf("Unexpected min point: (%.3f, %.3f)", min.X(), min.Y())
	}
	if math.Abs(max.X()-10) > 1e-6 || math.Abs(max.Y()-10) > 1e-6 {
		t.Errorf("Unexpected max point: (%.3f, %.3f)", max.X(), max.Y())
	}
}

func TestPathCurves(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.Curve3To(10, 0, 5, -5)          // Quadratic curve
	p.Curve4To(20, 0, 15, 10, 25, -5) // Cubic curve

	if !p.HasCurves() {
		t.Error("Expected path to have curves")
	}

	// The original path might not have lines until approximated
	// Test that approximation creates lines
	approx := p.Approximate(0.1, 4)
	if !approx.HasLines() {
		t.Error("Expected approximated path to have lines")
	}
}

func TestPathApproximation(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.Curve4To(10, 0, 5, 5, 10, 0) // Simple cubic curve

	// Approximate with different parameters
	approx := p.Approximate(0.1, 4)

	if len(approx.Commands()) == 0 {
		t.Error("Expected non-empty approximation")
	}

	// Approximated path should have more line segments than original
	if len(approx.Commands()) <= len(p.Commands()) {
		t.Error("Expected approximation to have more commands than original")
	}
}

func TestPathSubPaths(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(10, 0)
	p.LineTo(10, 10)
	p.MoveTo(20, 20) // New sub-path
	p.LineTo(30, 20)
	p.LineTo(30, 30)

	if !p.HasSubPaths() {
		t.Error("Expected path to have sub-paths")
	}

	subPaths := p.SubPaths()
	if len(subPaths) != 2 {
		t.Errorf("Expected 2 sub-paths, got %d", len(subPaths))
	}

	// First sub-path should have 3 commands (MoveTo + 2 LineTo)
	if len(subPaths[0].Commands()) != 3 {
		t.Errorf("Expected first sub-path to have 3 commands, got %d", len(subPaths[0].Commands()))
	}

	// Second sub-path should have 3 commands (MoveTo + 2 LineTo)
	if len(subPaths[1].Commands()) != 3 {
		t.Errorf("Expected second sub-path to have 3 commands, got %d", len(subPaths[1].Commands()))
	}
}

func TestPathTransform(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(10, 0)
	p.LineTo(10, 10)

	// Create scaling matrix
	scale := dxfmath.Scale(2, 2, 1)

	// Transform path
	transformed := p.Transform(&scale)

	// Check that vertices are scaled
	vertices := transformed.Vertices()
	if len(vertices) != 3 {
		t.Errorf("Expected 3 vertices, got %d", len(vertices))
	}

	// First vertex should be (0,0) * 2 = (0,0)
	if math.Abs(vertices[0].X()-0) > 1e-6 || math.Abs(vertices[0].Y()-0) > 1e-6 {
		t.Errorf("Unexpected first vertex: (%.3f, %.3f)", vertices[0].X(), vertices[0].Y())
	}

	// Second vertex should be (10,0) * 2 = (20,0)
	if math.Abs(vertices[1].X()-20) > 1e-6 || math.Abs(vertices[1].Y()-0) > 1e-6 {
		t.Errorf("Unexpected second vertex: (%.3f, %.3f)", vertices[1].X(), vertices[1].Y())
	}

	// Third vertex should be (10,10) * 2 = (20,20)
	if math.Abs(vertices[2].X()-20) > 1e-6 || math.Abs(vertices[2].Y()-20) > 1e-6 {
		t.Errorf("Unexpected third vertex: (%.3f, %.3f)", vertices[2].X(), vertices[2].Y())
	}
}

func TestPathClone(t *testing.T) {
	original := NewPath()
	original.MoveTo(1, 2)
	original.LineTo(3, 4)
	original.Curve3To(6, 8, 5, 5)

	clone := original.Clone()

	// Check that clone has same commands
	if len(clone.Commands()) != len(original.Commands()) {
		t.Errorf("Expected same number of commands, got %d vs %d", len(clone.Commands()), len(original.Commands()))
	}

	// Modify clone and verify original is unchanged
	clone.MoveTo(10, 10)

	if len(original.Commands()) != 3 {
		t.Errorf("Original should be unchanged, got %d commands", len(original.Commands()))
	}

	if len(clone.Commands()) != 4 {
		t.Errorf("Clone should have 4 commands, got %d", len(clone.Commands()))
	}
}

func TestBezierCalculations(t *testing.T) {
	p := NewPath()

	// Test quadratic Bezier calculation
	start := dxfmath.NewVec2(0, 0)
	ctrl := dxfmath.NewVec2(5, 10)
	end := dxfmath.NewVec2(10, 0)

	// At t=0, should be start
	point := p.quadraticBezierPoint(start, ctrl, end, 0.0)
	if point.Distance(start) > 1e-6 {
		t.Errorf("At t=0, should be start point")
	}

	// At t=1, should be end
	point = p.quadraticBezierPoint(start, ctrl, end, 1.0)
	if point.Distance(end) > 1e-6 {
		t.Errorf("At t=1, should be end point")
	}

	// Test cubic Bezier calculation
	ctrl1 := dxfmath.NewVec2(3, 10)
	ctrl2 := dxfmath.NewVec2(7, -10)

	// At t=0, should be start
	point = p.cubicBezierPoint(start, ctrl1, ctrl2, end, 0.0)
	if point.Distance(start) > 1e-6 {
		t.Errorf("At t=0, should be start point")
	}

	// At t=1, should be end
	point = p.cubicBezierPoint(start, ctrl1, ctrl2, end, 1.0)
	if point.Distance(end) > 1e-6 {
		t.Errorf("At t=1, should be end point")
	}
}
