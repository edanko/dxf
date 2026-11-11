package math

import (
	"fmt"
	"math"
	"sort"
)

// clampedUniformKnots returns an open uniform (clamped) knot vector for count
// control points of the given degree, normalized so the parameter domain is
// [0,1]. The first degree+1 knots equal 0 and the last degree+1 equal 1.
func clampedUniformKnots(count, degree int) []float64 {
	n := count - 1
	knots := make([]float64, count+degree+1)
	for i := range knots {
		switch {
		case i <= degree:
			knots[i] = 0
		case i >= n+1:
			knots[i] = 1
		default:
			knots[i] = float64(i-degree) / float64(n+1-degree)
		}
	}
	return knots
}

// NewClampedBSpline creates a clamped (open uniform) B-spline: the curve
// interpolates its first and last control points. The parameter domain of
// BSpline.EvaluatePoint is [0,1].
func NewClampedBSpline(controlPoints []Vec3, degree int) (*BSpline, error) {
	if len(controlPoints) < 2 {
		return nil, fmt.Errorf("need at least 2 control points")
	}
	if degree < 1 {
		return nil, fmt.Errorf("degree must be at least 1, got %d", degree)
	}
	if degree >= len(controlPoints) {
		return nil, fmt.Errorf("degree %d must be smaller than control point count %d", degree, len(controlPoints))
	}
	knots := clampedUniformKnots(len(controlPoints), degree)
	return NewBSplineWithKnots(controlPoints, nil, knots, degree), nil
}

// NewPeriodicBSpline creates a uniform, unclamped ("periodic" form) B-spline:
// the knot vector has no repeated end knots, so the curve is C(degree-1)
// continuous everywhere and does not interpolate its end control points.
// Callers wanting a closed curve append the first degree control points to the
// end of controlPoints.
func NewPeriodicBSpline(controlPoints []Vec3, degree int) (*BSpline, error) {
	if len(controlPoints) < 2 {
		return nil, fmt.Errorf("need at least 2 control points")
	}
	if degree < 1 {
		return nil, fmt.Errorf("degree must be at least 1, got %d", degree)
	}
	if degree >= len(controlPoints) {
		return nil, fmt.Errorf("degree %d must be smaller than control point count %d", degree, len(controlPoints))
	}
	knots := make([]float64, len(controlPoints)+degree+1)
	for i := range knots {
		knots[i] = float64(i)
	}
	weights := make([]float64, len(controlPoints))
	for i := range weights {
		weights[i] = 1
	}
	return NewBSplineWithKnots(controlPoints, weights, knots, degree), nil
}

// IsClamped reports whether the knot vector repeats the first and last knot
// value degree+1 times, which makes the curve interpolate its first and last
// control points.
func (bs *BSpline) IsClamped() bool {
	knots := bs.Knots()
	degree := bs.Degree()
	if len(knots) < 2*degree+2 {
		return false
	}
	last := len(knots) - 1
	low, high := knots[0], knots[last]
	for i := 0; i <= degree; i++ {
		if absFloat(knots[i]-low) > DefaultTolerance || absFloat(knots[last-i]-high) > DefaultTolerance {
			return false
		}
	}
	return true
}

// KnotDomain returns the parameter interval where the curve is defined, in the
// curve's own knot coordinates: [knot[degree], knot[control point count]].
// BSpline.EvaluatePoint maps t in [0,1] onto the full knot range, so t is only
// meaningful for knot vectors normalized to [0,1] (clamped forms); sample other
// forms directly at parameters inside this interval.
func (bs *BSpline) KnotDomain() (float64, float64) {
	knots := bs.Knots()
	degree := bs.Degree()
	count := len(bs.controlPoints)
	if len(knots) < count+degree+1 || count <= degree {
		return 0, 0
	}
	return knots[degree], knots[count]
}

// InsertKnot returns a new curve with knot u inserted once. The shape is
// unchanged: the curve is geometrically identical to the receiver, evaluated at
// the same normalized parameter. Insertion uses Boehm's algorithm on
// homogeneous (weighted) control points, so rational curves stay exact.
func (bs *BSpline) InsertKnot(u float64) (*BSpline, error) {
	return bs.insertKnots([]float64{u})
}

// InsertKnots returns a new curve with every knot in us inserted. Knots may
// repeat, which raises continuity at that parameter to zero.
func (bs *BSpline) InsertKnots(us []float64) (*BSpline, error) {
	return bs.insertKnots(us)
}

func (bs *BSpline) insertKnots(us []float64) (*BSpline, error) {
	knots := bs.Knots()
	degree := bs.Degree()
	weights := bs.Weights()
	cps := bs.ControlPoints()
	if len(knots) != len(cps)+degree+1 {
		return nil, fmt.Errorf("knot vector length %d does not match %d control points of degree %d", len(knots), len(cps), degree)
	}
	low, high := knots[degree], knots[len(cps)]
	for _, u := range us {
		if u < low-DefaultTolerance || u > high+DefaultTolerance {
			return nil, fmt.Errorf("knot %v outside parameter domain [%v, %v]", u, low, high)
		}
	}
	sorted := make([]float64, len(us))
	copy(sorted, us)
	sort.Float64s(sorted)

	curCP, curWeights, curKnots := cps, weights, knots
	curDegree := degree
	for _, u := range sorted {
		var err error
		curCP, curWeights, curKnots, curDegree, err = insertOneKnot(curCP, curWeights, curKnots, curDegree, u)
		if err != nil {
			return nil, err
		}
	}
	return NewBSplineWithKnots(curCP, curWeights, curKnots, curDegree), nil
}

// insertOneKnot applies one Boehm knot insertion step and returns the
// resulting control points, weights, knot vector and degree.
func insertOneKnot(cps []Vec3, weights []float64, knots []float64, degree int, u float64) ([]Vec3, []float64, []float64, int, error) {
	n := len(cps) - 1
	if degree < 1 || n < degree {
		return nil, nil, nil, 0, fmt.Errorf("cannot insert knot into degree %d with %d control points", degree, n+1)
	}
	if len(weights) != len(cps) {
		weights = defaultWeights(len(cps))
	}
	// Homogeneous control points keep the insertion exact for rational curves.
	hw := make([]Vec4, len(cps))
	for i, cp := range cps {
		w := weights[i]
		hw[i] = Vec4{cp.X() * w, cp.Y() * w, cp.Z() * w, w}
	}

	span := knotSpanForInsertion(knots, degree, n, u)
	if span < degree || span > n {
		return nil, nil, nil, 0, fmt.Errorf("knot %v sits on a domain boundary", u)
	}
	if mult := knotMultiplicity(knots, u); mult > degree {
		return nil, nil, nil, 0, fmt.Errorf("knot %v has multiplicity %d, maximum for degree %d is %d", u, mult, degree, degree)
	}

	newCount := len(cps) + 1
	out := make([]Vec4, newCount)
	// Copy the leading control points that are unaffected.
	for i := 0; i <= span-degree; i++ {
		out[i] = hw[i]
	}
	// Blend the control points whose basis functions are affected.
	for i := span - degree + 1; i <= span; i++ {
		denominator := knots[i+degree] - knots[i]
		alpha := 1.0
		if absFloat(denominator) > DefaultTolerance {
			alpha = (u - knots[i]) / denominator
		}
		out[i] = blendHom(hw[i], hw[i-1], alpha)
	}
	// Shift the trailing control points.
	for i := span + 1; i < newCount; i++ {
		out[i] = hw[i-1]
	}

	newKnots := make([]float64, len(knots)+1)
	copy(newKnots, knots[:span+1])
	newKnots[span+1] = u
	copy(newKnots[span+2:], knots[span+1:])

	newCP := make([]Vec3, newCount)
	newWeights := make([]float64, newCount)
	for i, h := range out {
		w := h[3]
		newWeights[i] = w
		if absFloat(w) > DefaultTolerance {
			newCP[i] = Vec3{h[0] / w, h[1] / w, h[2] / w}
		} else {
			newCP[i] = Vec3{h[0], h[1], h[2]}
		}
	}
	return newCP, newWeights, newKnots, degree, nil
}

// knotSpanForInsertion returns the index k with knots[k] <= u < knots[k+1],
// matching BSpline.findKnotSpan: n is the last control point index.
func knotSpanForInsertion(knots []float64, degree, n int, u float64) int {
	if u >= knots[n+1] {
		return n
	}
	if u <= knots[degree] {
		return degree
	}
	low, high := degree, n+1
	mid := (low + high) / 2
	for u < knots[mid] || u >= knots[mid+1] {
		if u < knots[mid] {
			high = mid
		} else {
			low = mid
		}
		mid = (low + high) / 2
	}
	return mid
}

// knotMultiplicity counts how many times u already appears in the knot vector.
func knotMultiplicity(knots []float64, u float64) int {
	count := 0
	for _, knot := range knots {
		if absFloat(knot-u) <= DefaultTolerance {
			count++
		}
	}
	return count
}

func blendHom(a, b Vec4, alpha float64) Vec4 {
	return Vec4{
		alpha*a[0] + (1-alpha)*b[0],
		alpha*a[1] + (1-alpha)*b[1],
		alpha*a[2] + (1-alpha)*b[2],
		alpha*a[3] + (1-alpha)*b[3],
	}
}

func defaultWeights(count int) []float64 {
	w := make([]float64, count)
	for i := range w {
		w[i] = 1
	}
	return w
}

// InsertKnot returns a NURBS curve with knot u inserted once. The shape is
// unchanged. Both the rational model and its underlying B-spline are rebuilt so
// the two stay consistent.
func (nurbs *NURBS) InsertKnot(u float64) (*NURBS, error) {
	cps := make([]Vec3, len(nurbs.controlPoints))
	copy(cps, nurbs.controlPoints)
	weights := make([]float64, len(nurbs.weights))
	copy(weights, nurbs.weights)
	knots := make([]float64, len(nurbs.knots))
	copy(knots, nurbs.knots)

	low, high := knots[nurbs.degree], knots[len(cps)]
	if u < low-DefaultTolerance || u > high+DefaultTolerance {
		return nil, fmt.Errorf("knot %v outside parameter domain [%v, %v]", u, low, high)
	}

	newCP, newWeights, newKnots, _, err := insertOneKnot(cps, weights, knots, nurbs.degree, u)
	if err != nil {
		return nil, err
	}
	return NewNURBS(newCP, newWeights, newKnots, nurbs.degree), nil
}

// NewRationalCircleArc returns an exact rational quadratic arc of the given
// radius, sweeping sweep radians counter-clockwise from startAngle around
// center in the plane z = center.Z. sweep must not exceed pi/2 for one segment.
// The parameter domain of the result is [0,1].
func NewRationalCircleArc(center Vec3, radius, startAngle, sweep float64) (*NURBS, error) {
	if radius <= 0 {
		return nil, fmt.Errorf("radius must be positive, got %v", radius)
	}
	if sweep <= 0 || sweep > math.Pi/2+DefaultTolerance {
		return nil, fmt.Errorf("arc sweep must be in (0, pi/2], got %v", sweep)
	}
	end := startAngle + sweep
	mid := startAngle + sweep/2
	c, s := math.Cos(startAngle), math.Sin(startAngle)
	c2, s2 := math.Cos(end), math.Sin(end)
	cm, sm := math.Cos(mid), math.Sin(mid)
	w := math.Cos(sweep / 2)
	cps := []Vec3{
		{center.X() + radius*c, center.Y() + radius*s, center.Z()},
		{center.X() + radius*cm/w, center.Y() + radius*sm/w, center.Z()},
		{center.X() + radius*c2, center.Y() + radius*s2, center.Z()},
	}
	weights := []float64{1, w, 1}
	knots := []float64{0, 0, 0, 1, 1, 1}
	nurbs := NewNURBS(cps, weights, knots, 2)
	if nurbs == nil {
		return nil, fmt.Errorf("failed to build rational arc")
	}
	return nurbs, nil
}

// NewRationalCircle returns an exact rational quadratic full circle of the
// given radius centered at center, built from four quarter arcs. The parameter
// domain of the result is [0,1] and every evaluated point lies on the circle.
func NewRationalCircle(center Vec3, radius float64) (*NURBS, error) {
	if radius <= 0 {
		return nil, fmt.Errorf("radius must be positive, got %v", radius)
	}
	const w = math.Sqrt2 / 2 // cos(pi/4)
	cps := []Vec3{
		{center.X() + radius, center.Y(), center.Z()},
		{center.X() + radius, center.Y() + radius, center.Z()},
		{center.X(), center.Y() + radius, center.Z()},
		{center.X() - radius, center.Y() + radius, center.Z()},
		{center.X() - radius, center.Y(), center.Z()},
		{center.X() - radius, center.Y() - radius, center.Z()},
		{center.X(), center.Y() - radius, center.Z()},
		{center.X() + radius, center.Y() - radius, center.Z()},
		{center.X() + radius, center.Y(), center.Z()},
	}
	weights := []float64{1, w, 1, w, 1, w, 1, w, 1}
	knots := []float64{0, 0, 0, 0.25, 0.25, 0.5, 0.5, 0.75, 0.75, 1, 1, 1}
	nurbs := NewNURBS(cps, weights, knots, 2)
	if nurbs == nil {
		return nil, fmt.Errorf("failed to build rational circle")
	}
	return nurbs, nil
}
