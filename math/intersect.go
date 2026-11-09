package math

// IntersectionKind classifies the result of an intersection query.
type IntersectionKind int

const (
	// IntersectionNone means the primitives do not intersect.
	IntersectionNone IntersectionKind = iota
	// IntersectionPoint means the primitives meet at a single point.
	IntersectionPoint
	// IntersectionPoints means the primitives meet at several isolated points.
	IntersectionPoints
	// IntersectionTangent means the primitives touch tangentially.
	IntersectionTangent
	// IntersectionCoincident means the primitives lie on the same line/plane.
	IntersectionCoincident
	// IntersectionOverlap means the primitives overlap along a segment.
	IntersectionOverlap
)

// Intersection is the typed result of an intersection query. Callers should
// inspect Kind before using Points.
type Intersection struct {
	Kind   IntersectionKind
	Points []Vec3
}

// Empty reports whether the query found no intersection.
func (i Intersection) Empty() bool {
	return i.Kind == IntersectionNone || len(i.Points) == 0
}

// IntersectLineLine2D intersects two infinite 2D lines. Parallel lines are
// reported as coincident when they are the same line, otherwise as no
// intersection.
func IntersectLineLine2D(a0, a1, b0, b1 Vec2, tolerance float64) Intersection {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	d1 := a1.Sub(a0)
	d2 := b1.Sub(b0)
	denom := d1[0]*d2[1] - d1[1]*d2[0]
	diff := b0.Sub(a0)

	if absFloat(denom) <= tolerance {
		if absFloat(diff[0]*d1[1]-diff[1]*d1[0]) <= tolerance*maxFloat(d1.Length(), 1) {
			return Intersection{Kind: IntersectionCoincident}
		}
		return Intersection{Kind: IntersectionNone}
	}

	t := (diff[0]*d2[1] - diff[1]*d2[0]) / denom
	point := a0.Add(d1.Mul(t))
	return Intersection{Kind: IntersectionPoint, Points: []Vec3{{point.X(), point.Y(), 0}}}
}

// IntersectSegmentSegment2D intersects two finite 2D segments.
func IntersectSegmentSegment2D(a0, a1, b0, b1 Vec2, tolerance float64) Intersection {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	d1 := a1.Sub(a0)
	d2 := b1.Sub(b0)
	denom := d1[0]*d2[1] - d1[1]*d2[0]
	diff := b0.Sub(a0)

	if absFloat(denom) <= tolerance {
		if absFloat(diff[0]*d1[1]-diff[1]*d1[0]) > tolerance*maxFloat(d1.Length(), 1) {
			return Intersection{Kind: IntersectionNone}
		}
		// Collinear: project both segments onto the first direction.
		lengthSq := d1.LengthSquared()
		if lengthSq <= tolerance*tolerance {
			return Intersection{Kind: IntersectionNone}
		}
		t0 := b0.Sub(a0).Dot(d1) / lengthSq
		t1 := b1.Sub(a0).Dot(d1) / lengthSq
		// Overlap window on the first segment's parameter range [0, 1]; an empty
		// window (lo > hi) means the collinear segments are disjoint.
		lo := maxFloat(minFloat(t0, t1), 0)
		hi := minFloat(maxFloat(t0, t1), 1)
		if lo > hi+tolerance {
			return Intersection{Kind: IntersectionNone}
		}
		if absFloat(lo-hi) <= tolerance {
			point := a0.Add(d1.Mul(lo))
			return Intersection{Kind: IntersectionPoint, Points: []Vec3{{point.X(), point.Y(), 0}}}
		}
		return Intersection{Kind: IntersectionOverlap, Points: []Vec3{
			{a0.Add(d1.Mul(lo)).X(), a0.Add(d1.Mul(lo)).Y(), 0},
			{a0.Add(d1.Mul(hi)).X(), a0.Add(d1.Mul(hi)).Y(), 0},
		}}
	}

	t := (diff[0]*d2[1] - diff[1]*d2[0]) / denom
	u := (diff[0]*d1[1] - diff[1]*d1[0]) / denom
	if t < -tolerance || t > 1+tolerance || u < -tolerance || u > 1+tolerance {
		return Intersection{Kind: IntersectionNone}
	}
	point := a0.Add(d1.Mul(t))
	return Intersection{Kind: IntersectionPoint, Points: []Vec3{{point.X(), point.Y(), 0}}}
}

// IntersectLineLine3D intersects two infinite 3D lines. When the lines are
// skew but closer than tolerance they are reported as a single approximate
// point.
func IntersectLineLine3D(a0, a1, b0, b1 Vec3, tolerance float64) Intersection {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	d1 := a1.Sub(a0)
	d2 := b1.Sub(b0)
	r := a0.Sub(b0)
	a := d1.Dot(d1)
	e := d2.Dot(d2)
	f := d2.Dot(r)
	if a <= tolerance*tolerance || e <= tolerance*tolerance {
		return Intersection{Kind: IntersectionNone}
	}
	b := d1.Dot(d2)
	c := d1.Dot(r)
	denom := a*e - b*b
	if absFloat(denom) <= tolerance {
		if (r.Sub(d1.Mul(c / a))).Length() <= tolerance {
			return Intersection{Kind: IntersectionCoincident}
		}
		return Intersection{Kind: IntersectionNone}
	}
	t := (b*f - c*e) / denom
	u := (a*f - b*c) / denom
	p1 := a0.Add(d1.Mul(t))
	p2 := b0.Add(d2.Mul(u))
	if p1.Distance(p2) <= tolerance {
		mid := p1.Add(p2).Mul(0.5)
		return Intersection{Kind: IntersectionPoint, Points: []Vec3{mid}}
	}
	return Intersection{Kind: IntersectionNone}
}

// IntersectRayPlane intersects a ray with a plane.
func IntersectRayPlane(origin, direction Vec3, plane Plane, tolerance float64) Intersection {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	denom := direction.Dot(plane.Normal)
	dist := plane.Point.Sub(origin).Dot(plane.Normal)
	if absFloat(denom) <= tolerance {
		if absFloat(dist) <= tolerance {
			return Intersection{Kind: IntersectionCoincident}
		}
		return Intersection{Kind: IntersectionNone}
	}
	t := dist / denom
	if t < -tolerance {
		return Intersection{Kind: IntersectionNone}
	}
	return Intersection{Kind: IntersectionPoint, Points: []Vec3{origin.Add(direction.Mul(t))}}
}

// IntersectLineBezier2D intersects a finite 2D segment with a Bezier curve.
// The curve is flattened to the requested tolerance before segment tests.
func IntersectLineBezier2D(a0, a1 Vec2, bezier *Bezier, tolerance float64) Intersection {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	if bezier == nil {
		return Intersection{Kind: IntersectionNone}
	}
	samples := bezier.Flattening(tolerance, 64)
	if len(samples) < 2 {
		return Intersection{Kind: IntersectionNone}
	}
	var points []Vec3
	merge := mergeTolerance(samples, tolerance)
	for i := 0; i+1 < len(samples); i++ {
		b0 := samples[i].XY()
		b1 := samples[i+1].XY()
		hit := IntersectSegmentSegment2D(a0, a1, b0, b1, tolerance)
		points = appendUniquePoints(points, hit.Points, merge)
	}
	return classifyPoints(points)
}

// IntersectBezierBezier2D intersects two Bezier curves by flattening both.
func IntersectBezierBezier2D(a, b *Bezier, tolerance float64) Intersection {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	if a == nil || b == nil {
		return Intersection{Kind: IntersectionNone}
	}
	as := a.Flattening(tolerance, 64)
	bs := b.Flattening(tolerance, 64)
	if len(as) < 2 || len(bs) < 2 {
		return Intersection{Kind: IntersectionNone}
	}
	var points []Vec3
	merge := mergeTolerance(as, tolerance)
	if other := mergeTolerance(bs, tolerance); other > merge {
		merge = other
	}
	for i := 0; i+1 < len(as); i++ {
		for j := 0; j+1 < len(bs); j++ {
			hit := IntersectSegmentSegment2D(as[i].XY(), as[i+1].XY(), bs[j].XY(), bs[j+1].XY(), tolerance)
			points = appendUniquePoints(points, hit.Points, merge)
		}
	}
	return classifyPoints(points)
}

// mergeTolerance is the distance below which two hits found on a flattened
// polyline are the same crossing. Flattened segments are far longer than the
// flattening tolerance, so a crossing exactly at a shared segment endpoint is
// reported by both neighbouring segments at points up to one segment length
// apart; that length is the merge distance.
func mergeTolerance(samples []Vec3, tolerance float64) float64 {
	merge := tolerance
	for i := 0; i+1 < len(samples); i++ {
		if length := samples[i].Distance(samples[i+1]); length > merge {
			merge = length
		}
	}
	return merge
}

func classifyPoints(points []Vec3) Intersection {
	switch len(points) {
	case 0:
		return Intersection{Kind: IntersectionNone}
	case 1:
		return Intersection{Kind: IntersectionPoint, Points: points}
	default:
		return Intersection{Kind: IntersectionPoints, Points: points}
	}
}

func appendUniquePoints(dst, src []Vec3, tolerance float64) []Vec3 {
	for _, point := range src {
		found := false
		for _, existing := range dst {
			if existing.Distance(point) <= tolerance {
				found = true
				break
			}
		}
		if !found {
			dst = append(dst, point)
		}
	}
	return dst
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
