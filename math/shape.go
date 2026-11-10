package math

// Shape is a closed boundary made of one outer ring and zero or more hole
// rings. Rings are stored exactly as supplied; orientation is preserved and
// exposed through Orientation. Shape is independent of the DXF document model.
type Shape struct {
	outer []Vec2
	holes [][]Vec2
}

// NewShape builds a Shape from an outer ring and optional hole rings. The
// input slices are copied so later caller mutation cannot change the shape.
func NewShape(outer []Vec2, holes ...[]Vec2) *Shape {
	s := &Shape{
		outer: cloneRing(outer),
		holes: make([][]Vec2, len(holes)),
	}
	for i, h := range holes {
		s.holes[i] = cloneRing(h)
	}
	return s
}

func cloneRing(ring []Vec2) []Vec2 {
	if ring == nil {
		return nil
	}
	out := make([]Vec2, len(ring))
	copy(out, ring)
	return out
}

// Outer returns a copy of the outer boundary ring.
func (s *Shape) Outer() []Vec2 { return cloneRing(s.outer) }

// Holes returns copies of the hole rings.
func (s *Shape) Holes() [][]Vec2 {
	if s == nil {
		return nil
	}
	out := make([][]Vec2, len(s.holes))
	for i, h := range s.holes {
		out[i] = cloneRing(h)
	}
	return out
}

// Orientation returns the winding direction of the outer ring.
func (s *Shape) Orientation() PolygonOrientation {
	if s == nil {
		return PolygonDegenerate
	}
	return PolygonWinding(s.outer)
}

// Area returns the enclosed area: the absolute outer area minus the absolute
// area of every hole.
func (s *Shape) Area() float64 {
	if s == nil {
		return 0
	}
	area := PolygonArea(s.outer)
	for _, h := range s.holes {
		area -= PolygonArea(h)
	}
	if area < 0 {
		return 0
	}
	return area
}

// Contains reports whether point lies inside the outer ring and outside every
// hole ring. Points on a hole boundary are treated as outside the shape.
func (s *Shape) Contains(point Vec2) bool {
	if s == nil {
		return false
	}
	if !PointInPolygon(point, s.outer) {
		return false
	}
	for _, h := range s.holes {
		if PointInPolygon(point, h) {
			return false
		}
	}
	return true
}

// Centroid returns the area-weighted centroid of the outer ring.
func (s *Shape) Centroid() Vec2 {
	if s == nil {
		return Vec2{}
	}
	return PolygonCentroid(s.outer)
}

// BoundingBox returns the bounding box of the outer ring.
func (s *Shape) BoundingBox() BoundingBox {
	if s == nil {
		return BoundingBox{}
	}
	return NewBoundingBoxFromPoints(s.outer)
}

// Normalized returns a copy of the shape with a counter-clockwise outer ring
// and clockwise hole rings, which is the conventional fill orientation.
func (s *Shape) Normalized() *Shape {
	if s == nil {
		return nil
	}
	out := &Shape{outer: normalizeRing(s.outer, true)}
	out.holes = make([][]Vec2, len(s.holes))
	for i, h := range s.holes {
		out.holes[i] = normalizeRing(h, false)
	}
	return out
}

func normalizeRing(ring []Vec2, wantCCW bool) []Vec2 {
	clone := cloneRing(ring)
	ccw := PolygonWinding(clone) == PolygonCounterClockwise
	if ccw != wantCCW {
		for i, j := 0, len(clone)-1; i < j; i, j = i+1, j-1 {
			clone[i], clone[j] = clone[j], clone[i]
		}
	}
	return clone
}