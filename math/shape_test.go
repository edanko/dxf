package math

import (
	"math"
	"testing"
)

func square(x0, y0, size float64) []Vec2 {
	return []Vec2{
		{x0, y0},
		{x0 + size, y0},
		{x0 + size, y0 + size},
		{x0, y0 + size},
	}
}

func TestShapeAreaWithHole(t *testing.T) {
	outer := square(0, 0, 10)
	hole := square(3, 3, 4)
	s := NewShape(outer, hole)
	want := 100.0 - 16.0
	if got := s.Area(); math.Abs(got-want) > 1e-9 {
		t.Fatalf("area = %v, want %v", got, want)
	}
}

func TestShapeContainment(t *testing.T) {
	s := NewShape(square(0, 0, 10), square(3, 3, 4))
	cases := []struct {
		p    Vec2
		want bool
	}{
		{Vec2{1, 1}, true},   // in outer, outside hole
		{Vec2{5, 5}, false},  // inside hole
		{Vec2{-1, 5}, false}, // outside outer
		{Vec2{9.5, 9.5}, true},
	}
	for _, c := range cases {
		if got := s.Contains(c.p); got != c.want {
			t.Errorf("Contains(%v) = %v, want %v", c.p, got, c.want)
		}
	}
}

func TestShapePreservesOrientation(t *testing.T) {
	ccw := square(0, 0, 10)
	s := NewShape(ccw)
	if s.Orientation() != PolygonCounterClockwise {
		t.Fatalf("orientation = %v, want CCW", s.Orientation())
	}
	// Reverse the ring: orientation must be preserved as supplied.
	cw := square(0, 0, 10)
	for i, j := 0, len(cw)-1; i < j; i, j = i+1, j-1 {
		cw[i], cw[j] = cw[j], cw[i]
	}
	s2 := NewShape(cw)
	if s2.Orientation() != PolygonClockwise {
		t.Fatalf("orientation = %v, want CW", s2.Orientation())
	}
}

func TestShapeNormalizedOrientation(t *testing.T) {
	outer := square(0, 0, 10)
	for i, j := 0, len(outer)-1; i < j; i, j = i+1, j-1 {
		outer[i], outer[j] = outer[j], outer[i]
	}
	s := NewShape(outer, square(3, 3, 4)).Normalized()
	if s.Orientation() != PolygonCounterClockwise {
		t.Fatalf("normalized outer orientation = %v, want CCW", s.Orientation())
	}
	if PolygonWinding(s.Holes()[0]) != PolygonClockwise {
		t.Fatalf("normalized hole orientation = %v, want CW", PolygonWinding(s.Holes()[0]))
	}
}

func TestShapeCopiesInput(t *testing.T) {
	outer := square(0, 0, 10)
	s := NewShape(outer)
	outer[0] = Vec2{999, 999}
	if s.Outer()[0] == (Vec2{999, 999}) {
		t.Fatal("shape aliased caller input slice")
	}
}

func TestShapeAreaNoHoles(t *testing.T) {
	s := NewShape(square(0, 0, 5))
	if got := s.Area(); math.Abs(got-25) > 1e-9 {
		t.Fatalf("area = %v, want 25", got)
	}
}