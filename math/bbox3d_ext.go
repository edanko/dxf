package math

import "math"

// IsEmpty reports whether the 3D box has no extent.
func (bb BoundingBox3D) IsEmpty() bool {
	return bb.Min == bb.Max && bb.Min == (Vec3{})
}

// Union returns the smallest box containing both inputs.
func (bb BoundingBox3D) Union(other BoundingBox3D) BoundingBox3D {
	if bb.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return bb
	}
	return BoundingBox3D{
		Min: Vec3{
			math.Min(bb.Min.X(), other.Min.X()),
			math.Min(bb.Min.Y(), other.Min.Y()),
			math.Min(bb.Min.Z(), other.Min.Z()),
		},
		Max: Vec3{
			math.Max(bb.Max.X(), other.Max.X()),
			math.Max(bb.Max.Y(), other.Max.Y()),
			math.Max(bb.Max.Z(), other.Max.Z()),
		},
	}
}

// Intersection returns the overlapping box. The result is empty when the
// boxes do not overlap.
func (bb BoundingBox3D) Intersection(other BoundingBox3D) BoundingBox3D {
	min := Vec3{
		math.Max(bb.Min.X(), other.Min.X()),
		math.Max(bb.Min.Y(), other.Min.Y()),
		math.Max(bb.Min.Z(), other.Min.Z()),
	}
	max := Vec3{
		math.Min(bb.Max.X(), other.Max.X()),
		math.Min(bb.Max.Y(), other.Max.Y()),
		math.Min(bb.Max.Z(), other.Max.Z()),
	}
	if min.X() > max.X() || min.Y() > max.Y() || min.Z() > max.Z() {
		return BoundingBox3D{}
	}
	return BoundingBox3D{Min: min, Max: max}
}

// IntersectsBoundingBox reports whether two boxes overlap.
func (bb BoundingBox3D) IntersectsBoundingBox(other BoundingBox3D) bool {
	if bb.IsEmpty() || other.IsEmpty() {
		return false
	}
	return bb.Min.X() <= other.Max.X() && bb.Max.X() >= other.Min.X() &&
		bb.Min.Y() <= other.Max.Y() && bb.Max.Y() >= other.Min.Y() &&
		bb.Min.Z() <= other.Max.Z() && bb.Max.Z() >= other.Min.Z()
}

// ContainsBoundingBox reports whether other is fully inside bb.
func (bb BoundingBox3D) ContainsBoundingBox(other BoundingBox3D) bool {
	if other.IsEmpty() {
		return true
	}
	if bb.IsEmpty() {
		return false
	}
	return other.Min.X() >= bb.Min.X() && other.Max.X() <= bb.Max.X() &&
		other.Min.Y() >= bb.Min.Y() && other.Max.Y() <= bb.Max.Y() &&
		other.Min.Z() >= bb.Min.Z() && other.Max.Z() <= bb.Max.Z()
}

// Expand grows the box by margin on every side.
func (bb BoundingBox3D) Expand(margin float64) BoundingBox3D {
	if bb.IsEmpty() {
		return bb
	}
	delta := Vec3{margin, margin, margin}
	return BoundingBox3D{Min: bb.Min.Sub(delta), Max: bb.Max.Add(delta)}
}

// ExpandToPoint grows the box to include point.
func (bb BoundingBox3D) ExpandToPoint(point Vec3) BoundingBox3D {
	if bb.IsEmpty() {
		return BoundingBox3D{Min: point, Max: point}
	}
	return BoundingBox3D{
		Min: Vec3{
			math.Min(bb.Min.X(), point.X()),
			math.Min(bb.Min.Y(), point.Y()),
			math.Min(bb.Min.Z(), point.Z()),
		},
		Max: Vec3{
			math.Max(bb.Max.X(), point.X()),
			math.Max(bb.Max.Y(), point.Y()),
			math.Max(bb.Max.Z(), point.Z()),
		},
	}
}

// Translate moves the box by vec.
func (bb BoundingBox3D) Translate(vec Vec3) BoundingBox3D {
	if bb.IsEmpty() {
		return bb
	}
	return BoundingBox3D{Min: bb.Min.Add(vec), Max: bb.Max.Add(vec)}
}

// Corners returns the eight corner points of the box.
func (bb BoundingBox3D) Corners() []Vec3 {
	return []Vec3{
		{bb.Min.X(), bb.Min.Y(), bb.Min.Z()},
		{bb.Max.X(), bb.Min.Y(), bb.Min.Z()},
		{bb.Max.X(), bb.Max.Y(), bb.Min.Z()},
		{bb.Min.X(), bb.Max.Y(), bb.Min.Z()},
		{bb.Min.X(), bb.Min.Y(), bb.Max.Z()},
		{bb.Max.X(), bb.Min.Y(), bb.Max.Z()},
		{bb.Max.X(), bb.Max.Y(), bb.Max.Z()},
		{bb.Min.X(), bb.Max.Y(), bb.Max.Z()},
	}
}

// Transform applies m to the box and returns the axis-aligned box enclosing
// every transformed corner.
func (bb BoundingBox3D) Transform(m Matrix44) BoundingBox3D {
	if bb.IsEmpty() {
		return bb
	}
	corners := bb.Corners()
	transformed := make([]Vec3, len(corners))
	for i, corner := range corners {
		transformed[i] = m.MulVec(corner)
	}
	return NewBoundingBox3DFromPoints(transformed)
}

// Transform applies m to the 2D box (z = 0) and returns the axis-aligned box
// enclosing every transformed corner.
func (bb BoundingBox) Transform(m Matrix44) BoundingBox {
	if bb.IsEmpty() {
		return bb
	}
	points := bb.Corners()
	transformed := make([]Vec2, len(points))
	for i, point := range points {
		p := m.MulVec(Vec3{point.X(), point.Y(), 0})
		transformed[i] = Vec2{p.X(), p.Y()}
	}
	return NewBoundingBoxFromPoints(transformed)
}
