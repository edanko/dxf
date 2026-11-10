package math

import (
	"math"
	"sort"
)

// Segment is a finite 2D line segment.
type Segment struct {
	Start Vec2
	End   Vec2
}

// ClipSegmentRange clips the segment p0->p1 against portal using the Liang-
// Barsky algorithm and returns the visible parameter interval [t0, t1] where
// the clipped points are p0.Lerp(p1, t0) and p0.Lerp(p1, t1). ok is false when
// the segment is entirely outside the portal.
func ClipSegmentRange(p0, p1 Vec2, portal BoundingBox) (t0, t1 float64, ok bool) {
	if portal.IsEmpty() {
		return 0, 0, false
	}
	dx := p1.X() - p0.X()
	dy := p1.Y() - p0.Y()
	lo, hi := 0.0, 1.0

	clip := func(p, q float64) bool {
		if math.Abs(p) <= DefaultTolerance {
			return q >= -DefaultTolerance
		}
		r := q / p
		if p < 0 {
			if r > hi {
				return false
			}
			if r > lo {
				lo = r
			}
		} else {
			if r < lo {
				return false
			}
			if r < hi {
				hi = r
			}
		}
		return true
	}

	if !clip(-dx, p0.X()-portal.Min.X()) {
		return 0, 0, false
	}
	if !clip(dx, portal.Max.X()-p0.X()) {
		return 0, 0, false
	}
	if !clip(-dy, p0.Y()-portal.Min.Y()) {
		return 0, 0, false
	}
	if !clip(dy, portal.Max.Y()-p0.Y()) {
		return 0, 0, false
	}
	if lo > hi+DefaultTolerance {
		return 0, 0, false
	}
	if lo < 0 {
		lo = 0
	}
	if hi > 1 {
		hi = 1
	}
	return lo, hi, true
}

// ClipSegmentToPortal returns the portion of p0->p1 that lies inside portal.
// The boolean is false when the segment is entirely outside.
func ClipSegmentToPortal(p0, p1 Vec2, portal BoundingBox) (Segment, bool) {
	t0, t1, ok := ClipSegmentRange(p0, p1, portal)
	if !ok {
		return Segment{}, false
	}
	return Segment{Start: p0.Lerp(p1, t0), End: p0.Lerp(p1, t1)}, true
}

// ClipSegmentsToPortal clips each segment and returns only the visible
// portions. Order is preserved.
func ClipSegmentsToPortal(segments []Segment, portal BoundingBox) []Segment {
	out := make([]Segment, 0, len(segments))
	for _, s := range segments {
		if clipped, ok := ClipSegmentToPortal(s.Start, s.End, portal); ok {
			out = append(out, clipped)
		}
	}
	return out
}

// QuadraticBezier is a quadratic Bézier curve with start P0, control C and end
// P1.
type QuadraticBezier struct {
	P0 Vec2
	C  Vec2
	P1 Vec2
}

// CubicBezier is a cubic Bézier curve with endpoints P0/P1 and controls C1/C2.
type CubicBezier struct {
	P0 Vec2
	C1 Vec2
	C2 Vec2
	P1 Vec2
}

// Point evaluates the quadratic curve at t.
func (q QuadraticBezier) Point(t float64) Vec2 {
	mt := 1 - t
	return q.P0.Mul(mt * mt).Add(q.C.Mul(2 * mt * t)).Add(q.P1.Mul(t * t))
}

// Point evaluates the cubic curve at t.
func (c CubicBezier) Point(t float64) Vec2 {
	mt := 1 - t
	return c.P0.Mul(mt * mt * mt).
		Add(c.C1.Mul(3 * mt * mt * t)).
		Add(c.C2.Mul(3 * mt * t * t)).
		Add(c.P1.Mul(t * t * t))
}

// Split divides the quadratic curve at t into two curves.
func (q QuadraticBezier) Split(t float64) (QuadraticBezier, QuadraticBezier) {
	a := q.P0.Lerp(q.C, t)
	b := q.C.Lerp(q.P1, t)
	p := a.Lerp(b, t)
	return QuadraticBezier{q.P0, a, p}, QuadraticBezier{p, b, q.P1}
}

// Split divides the cubic curve at t into two curves.
func (c CubicBezier) Split(t float64) (CubicBezier, CubicBezier) {
	a := c.P0.Lerp(c.C1, t)
	b := c.C1.Lerp(c.C2, t)
	d := c.C2.Lerp(c.P1, t)
	e := a.Lerp(b, t)
	f := b.Lerp(d, t)
	g := e.Lerp(f, t)
	return CubicBezier{c.P0, a, e, g}, CubicBezier{g, f, d, c.P1}
}

// SubCurve returns the portion of the quadratic curve on [t0, t1].
func (q QuadraticBezier) SubCurve(t0, t1 float64) QuadraticBezier {
	if t1 < t0 {
		t0, t1 = t1, t0
	}
	_, tail := q.Split(t0)
	if t1 <= t0 {
		return QuadraticBezier{tail.P0, tail.P0, tail.P0}
	}
	f := (t1 - t0) / (1 - t0)
	sub, _ := tail.Split(f)
	return sub
}

// SubCurve returns the portion of the cubic curve on [t0, t1].
func (c CubicBezier) SubCurve(t0, t1 float64) CubicBezier {
	if t1 < t0 {
		t0, t1 = t1, t0
	}
	_, tail := c.Split(t0)
	if t1 <= t0 {
		return CubicBezier{tail.P0, tail.P0, tail.P0, tail.P0}
	}
	f := (t1 - t0) / (1 - t0)
	sub, _ := tail.Split(f)
	return sub
}

// ClipQuadraticToPortal returns the visible portions of q inside portal as up
// to a few sub-curves, in order along the curve. Curves entirely outside
// return nil.
func ClipQuadraticToPortal(q QuadraticBezier, portal BoundingBox) []QuadraticBezier {
	intervals := curveIntervals(portal, []float64{q.P0.X(), q.C.X(), q.P1.X()}, []float64{q.P0.Y(), q.C.Y(), q.P1.Y()}, q.Point)
	return mapIntervals(intervals, q.SubCurve)
}

// ClipCubicToPortal returns the visible portions of c inside portal as up to a
// few sub-curves, in order along the curve. Curves entirely outside return nil.
func ClipCubicToPortal(c CubicBezier, portal BoundingBox) []CubicBezier {
	intervals := curveIntervals(
		portal,
		[]float64{c.P0.X(), c.C1.X(), c.C2.X(), c.P1.X()},
		[]float64{c.P0.Y(), c.C1.Y(), c.C2.Y(), c.P1.Y()},
		c.Point,
	)
	return mapIntervals(intervals, c.SubCurve)
}

type interval struct{ t0, t1 float64 }

// curveIntervals finds the t intervals where a polynomial curve lies inside
// portal. xs/ys are the 1D Bernstein control values for x and y.
func curveIntervals(portal BoundingBox, xs, ys []float64, point func(float64) Vec2) []interval {
	if portal.IsEmpty() || len(xs) < 3 || len(ys) < 3 {
		return nil
	}
	// Build 1D polynomial coefficients in power basis for x(t) and y(t).
	xc := bernsteinToPower(xs)
	yc := bernsteinToPower(ys)

	candidates := []float64{0, 1}
	candidates = append(candidates, cubicRoots(xc[0], xc[1], xc[2], xc[3]-portal.Min.X())...)
	candidates = append(candidates, cubicRoots(xc[0], xc[1], xc[2], xc[3]-portal.Max.X())...)
	candidates = append(candidates, cubicRoots(yc[0], yc[1], yc[2], yc[3]-portal.Min.Y())...)
	candidates = append(candidates, cubicRoots(yc[0], yc[1], yc[2], yc[3]-portal.Max.Y())...)

	valid := make([]float64, 0, len(candidates))
	for _, t := range candidates {
		if t >= -1e-9 && t <= 1+1e-9 {
			valid = append(valid, math.Max(0, math.Min(1, t)))
		}
	}
	sort.Float64s(valid)

	var out []interval
	inStart := math.Inf(1)
	for i := 0; i+1 < len(valid); i++ {
		a, b := valid[i], valid[i+1]
		if b-a <= 1e-12 {
			continue
		}
		mid := (a + b) / 2
		p := point(mid)
		inside := p.X() >= portal.Min.X()-DefaultTolerance && p.X() <= portal.Max.X()+DefaultTolerance &&
			p.Y() >= portal.Min.Y()-DefaultTolerance && p.Y() <= portal.Max.Y()+DefaultTolerance
		if inside {
			if math.IsInf(inStart, 1) {
				inStart = a
			}
		} else if !math.IsInf(inStart, 1) {
			out = append(out, interval{inStart, a})
			inStart = math.Inf(1)
		}
	}
	if !math.IsInf(inStart, 1) {
		out = append(out, interval{inStart, 1})
	}
	return out
}

func mapIntervals[T any](intervals []interval, sub func(t0, t1 float64) T) []T {
	if len(intervals) == 0 {
		return nil
	}
	out := make([]T, 0, len(intervals))
	for _, iv := range intervals {
		out = append(out, sub(iv.t0, iv.t1))
	}
	return out
}

// bernsteinToPower converts control values (degree 2 or 3) to power-basis
// coefficients [a,b,c,d] for the polynomial.
func bernsteinToPower(v []float64) [4]float64 {
	switch len(v) {
	case 3:
		// quadratic: (1-t)^2 p0 + 2(1-t)t p1 + t^2 p2
		return [4]float64{
			0,
			v[0] - 2*v[1] + v[2],
			2 * (v[1] - v[0]),
			v[0],
		}
	case 4:
		// cubic: (1-t)^3 p0 + 3(1-t)^2 t p1 + 3(1-t)t^2 p2 + t^3 p3
		return [4]float64{
			-v[0] + 3*v[1] - 3*v[2] + v[3],
			3*v[0] - 6*v[1] + 3*v[2],
			-3*v[0] + 3*v[1],
			v[0],
		}
	default:
		return [4]float64{}
	}
}

// cubicRoots returns the real roots of a*t^3+b*t^2+c*t+d in [0,1].
func cubicRoots(a, b, c, d float64) []float64 {
	if math.Abs(a) <= 1e-12 {
		return quadraticRoots(b, c, d)
	}
	// Critical points from the derivative: 3a t^2 + 2b t + c.
	crit := quadraticRootsRaw(3*a, 2*b, c)
	pts := []float64{0, 1}
	for _, t := range crit {
		if t > 0 && t < 1 {
			pts = append(pts, t)
		}
	}
	sort.Float64s(pts)
	roots := make([]float64, 0, 3)
	for i := 0; i+1 < len(pts); i++ {
		lo, hi := pts[i], pts[i+1]
		flo := poly(a, b, c, d, lo)
		fhi := poly(a, b, c, d, hi)
		if math.Abs(flo) <= 1e-12 {
			roots = append(roots, lo)
			continue
		}
		if math.Abs(fhi) <= 1e-12 {
			roots = append(roots, hi)
			continue
		}
		if flo*fhi < 0 {
			roots = append(roots, bisect(a, b, c, d, lo, hi))
		}
	}
	return roots
}

func quadraticRoots(a, b, c float64) []float64 {
	if math.Abs(a) <= 1e-12 {
		if math.Abs(b) <= 1e-12 {
			return nil
		}
		return []float64{-c / b}
	}
	return quadraticRootsRaw(a, b, c)
}

func quadraticRootsRaw(a, b, c float64) []float64 {
	if math.Abs(a) <= 1e-12 {
		if math.Abs(b) <= 1e-12 {
			return nil
		}
		return []float64{-c / b}
	}
	disc := b*b - 4*a*c
	if disc < 0 {
		return nil
	}
	if disc == 0 {
		return []float64{-b / (2 * a)}
	}
	s := math.Sqrt(disc)
	return []float64{(-b - s) / (2 * a), (-b + s) / (2 * a)}
}

func poly(a, b, c, d, t float64) float64 {
	return ((a*t+b)*t+c)*t + d
}

func bisect(a, b, c, d, lo, hi float64) float64 {
	flo := poly(a, b, c, d, lo)
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		fmid := poly(a, b, c, d, mid)
		if flo*fmid <= 0 {
			hi = mid
		} else {
			lo = mid
			flo = fmid
		}
	}
	return (lo + hi) / 2
}