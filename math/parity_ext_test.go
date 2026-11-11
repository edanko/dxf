package math

import (
	"math"
	"testing"
)

// Tests for the change-5 parity scenarios: offset joins, triangulation of
// degenerate inputs and polygons with holes, caller-supplied tolerances,
// bulge/arc round trips, matrix decomposition, and BBox3D operations.

func sumTriangleArea(vertices []Point2D, indices []int) float64 {
	total := 0.0
	for i := 0; i+2 < len(indices); i += 3 {
		a := vertices[indices[i]]
		b := vertices[indices[i+1]]
		c := vertices[indices[i+2]]
		total += math.Abs((b.X-a.X)*(c.Y-a.Y)-(c.X-a.X)*(b.Y-a.Y)) / 2
	}
	return total
}

func TestOffsetPolylineParallelAndJoins(t *testing.T) {
	points := []Vec2{{0, 0}, {10, 0}, {10, 10}}
	// Positive distance offsets to the left normal of travel: the first
	// segment (direction +X) moves by (0, +distance).
	offset, err := OffsetPolyline(points, 2, JoinRound)
	if err != nil {
		t.Fatalf("OffsetPolyline: %v", err)
	}
	if len(offset) < 4 {
		t.Fatalf("round join expected arc samples, got %d points", len(offset))
	}
	if first := offset[0]; math.Abs(first[1]-2) > 1e-9 || math.Abs(first[0]) > 1e-9 {
		t.Errorf("first point %v not parallel at distance 2", first)
	}
	if last := offset[len(offset)-1]; last[1] < 10 {
		t.Errorf("endpoint orientation not preserved: last %v", last)
	}

	miter, err := OffsetPolyline(points, 2, JoinMiter)
	if err != nil {
		t.Fatalf("OffsetPolyline miter: %v", err)
	}
	miterPoint := false
	for _, p := range miter {
		if math.Abs(p.Distance(Vec2{10, 0})-2*math.Sqrt2) < 1e-9 {
			miterPoint = true
		}
	}
	if !miterPoint {
		t.Errorf("miter join: no point at distance 2·sqrt(2) from the corner: %v", miter)
	}
	roundArc := false
	for _, p := range offset {
		if p.Distance(Vec2{10, 0}) <= 2+1e-9 {
			roundArc = true
			break
		}
	}
	if !roundArc {
		t.Errorf("round join: no sample on the join arc")
	}
}

func TestOffsetBezierAndSplineParallel(t *testing.T) {
	bezier, err := NewBezier3P(Vec3{0, 0, 0}, Vec3{5, 5, 0}, Vec3{10, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	pts, err := OffsetBezier2D(bezier, 1, 1e-6, JoinRound)
	if err != nil {
		t.Fatalf("OffsetBezier2D: %v", err)
	}
	if len(pts) < 8 {
		t.Fatalf("expected sampled offset, got %d points", len(pts))
	}
	curveSamples := bezier.Flattening(1e-6, 64)
	for _, p := range pts {
		// Every sampled point must sit ~1 away from the curve, measured as
		// the distance to the flattening polyline (point-to-sample distance
		// would carry the chord deviation of the flattened curve itself).
		minDist := math.Inf(1)
		for i := 0; i+1 < len(curveSamples); i++ {
			a := Vec2{curveSamples[i][0], curveSamples[i][1]}
			b := Vec2{curveSamples[i+1][0], curveSamples[i+1][1]}
			if d := pointToSegmentDist2D(p, a, b); d < minDist {
				minDist = d
			}
		}
		if math.Abs(minDist-1) > 1e-3 {
			// Measured against the flattening polyline, so the chord
			// deviation of the flattened curve (chords ~0.007 long) is
			// included in the bound.
			t.Fatalf("offset point %v at distance %v from curve, want 1", p, minDist)
		}
	}

	spline := NewBSpline([]Vec3{{0, 0, 0}, {4, 4, 0}, {8, 0, 0}, {12, -4, 0}}, 3)
	spts, err := OffsetSpline2D(spline, 2, 1e-6, JoinRound)
	if err != nil {
		t.Fatalf("OffsetSpline2D: %v", err)
	}
	if len(spts) < 8 {
		t.Fatalf("expected sampled spline offset, got %d points", len(spts))
	}
}

func TestTriangulatePolygonWithHole(t *testing.T) {
	exterior := []Point2D{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	hole := []Point2D{{4, 4}, {6, 4}, {6, 6}, {4, 6}}
	indices, err := TriangulatePolygonSafe(exterior, [][]Point2D{hole})
	if err != nil {
		t.Fatalf("TriangulatePolygonSafe: %v", err)
	}
	vertices := append(append([]Point2D{}, exterior...), hole...)
	if len(indices) == 0 {
		t.Fatalf("no triangles produced")
	}
	if !ValidateTriangulation(vertices, indices) {
		t.Errorf("mesh contains overlapping triangles")
	}
	area := sumTriangleArea(vertices, indices)
	// The centered 2x2 hole has no exterior vertex on its row; the bridge
	// must still be a zero-width needle so the triangle area sum equals the
	// full 10x10 minus 2x2 region.
	if math.Abs(area-96) > 1e-9 {
		t.Errorf("triangle area sum %v, want 96", area)
	}
	if issues := DiagnoseTriangulation(vertices, indices); len(issues) != 0 {
		t.Errorf("unexpected diagnostics on valid mesh: %v", issues)
	}
}

func TestTriangulateCenteredHoleNoExteriorBay(t *testing.T) {
	// Regression: a hole with no exterior vertex on its row used to be
	// bridged through a diagonal notch, leaving a bay outside the region and
	// a triangle area shortfall. Every triangle centroid must stay inside the
	// outer ring and outside the hole, and the areas must add up exactly.
	exterior := []Point2D{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	hole := []Point2D{{4, 4}, {6, 4}, {6, 6}, {4, 6}}
	indices, err := TriangulatePolygonSafe(exterior, [][]Point2D{hole})
	if err != nil {
		t.Fatalf("TriangulatePolygonSafe: %v", err)
	}
	vertices := append(append([]Point2D{}, exterior...), hole...)
	if !ValidateTriangulation(vertices, indices) {
		t.Fatalf("mesh contains overlapping triangles")
	}
	if got := sumTriangleArea(vertices, indices); math.Abs(got-96) > 1e-9 {
		t.Errorf("triangle area sum %v, want 96 (outer 100 minus hole 4)", got)
	}
	shape := NewShape(
		[]Vec2{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
		[]Vec2{{4, 4}, {6, 4}, {6, 6}, {4, 6}},
	)
	for tri := 0; tri < len(indices)/3; tri++ {
		a := vertices[indices[tri*3]]
		b := vertices[indices[tri*3+1]]
		c := vertices[indices[tri*3+2]]
		centroid := Vec2{(a.X + b.X + c.X) / 3, (a.Y + b.Y + c.Y) / 3}
		if !shape.Contains(centroid) {
			t.Errorf("triangle %d centroid %v outside outer boundary or inside hole", tri, centroid)
		}
	}
}

func TestTriangulateDisjointHolesAreaParity(t *testing.T) {
	// Multiple holes exercise the deterministic bridge-target search: both
	// holes are bridged without losing any region area.
	exterior := []Point2D{{0, 0}, {20, 0}, {20, 10}, {0, 10}}
	holes := [][]Point2D{
		{{2, 2}, {4, 2}, {4, 4}, {2, 4}},
		{{14, 5}, {17, 5}, {17, 8}, {14, 8}},
	}
	indices, err := TriangulatePolygonSafe(exterior, holes)
	if err != nil {
		t.Fatalf("TriangulatePolygonSafe: %v", err)
	}
	vertices := append([]Point2D{}, exterior...)
	for _, hole := range holes {
		vertices = append(vertices, hole...)
	}
	if !ValidateTriangulation(vertices, indices) {
		t.Fatalf("mesh contains overlapping triangles")
	}
	if got := sumTriangleArea(vertices, indices); math.Abs(got-187) > 1e-9 {
		t.Errorf("triangle area sum %v, want 187 (outer 200 minus holes 13)", got)
	}
	if issues := DiagnoseTriangulation(vertices, indices); len(issues) != 0 {
		t.Errorf("unexpected diagnostics on valid mesh: %v", issues)
	}
}

func TestTriangulateDegenerateInput(t *testing.T) {
	// Collinear points: no valid triangles, no panic.
	collinear := []Point2D{{0, 0}, {1, 0}, {2, 0}, {3, 0}}
	indices, err := TriangulatePolygonSafe(collinear, nil)
	if err != nil {
		t.Fatalf("collinear input: unexpected error %v", err)
	}
	if len(indices) != 0 {
		t.Errorf("collinear input produced %d indices, want 0", len(indices))
	}
	if issues := DiagnoseTriangulation(collinear, indices); len(issues) != 0 {
		t.Errorf("empty mesh over collinear vertices produced issues: %v", issues)
	}
	// Duplicate vertices are removed by SanitizePolygon (Vec2 flavor): the
	// repeated closing vertex and the repeated middle vertex collapse, so
	// the canonical open ring has three vertices.
	dup := SanitizePolygon([]Vec2{{0, 0}, {1, 0}, {1, 0}, {1, 1}, {0, 0}}, 1e-9)
	if len(dup) != 3 {
		t.Errorf("SanitizePolygon returned %d vertices: %v", len(dup), dup)
	}
	// Collinear middles are removed.
	line := SanitizePolygon([]Vec2{{0, 0}, {1, 0}, {2, 0}, {2, 2}}, 1e-9)
	if len(line) != 3 {
		t.Errorf("SanitizePolygon kept collinear middle: %v", line)
	}
	// Zero-area triangles are reported by DiagnoseTriangulation.
	verts := []Point2D{{0, 0}, {1, 0}, {2, 0}, {0, 1}}
	bad := DiagnoseTriangulation(verts, []int{0, 1, 2, 0, 2, 3})
	found := false
	for _, issue := range bad {
		if issue.Kind == IssueDegenerateTriangle {
			found = true
		}
	}
	if !found {
		t.Errorf("zero-area triangle not reported, got %v", bad)
	}
}

func TestCustomToleranceGoverns(t *testing.T) {
	// A 1e-7 gap between two collinear segments.
	a0, a1 := Vec2{0, 0}, Vec2{1, 0}
	b0, b1 := Vec2{1.0000001, 0}, Vec2{2, 0}
	if IntersectSegmentSegment2D(a0, a1, b0, b1, 1e-6).Empty() {
		t.Error("1e-6 tolerance should bridge the 1e-7 gap")
	}
	if !IntersectSegmentSegment2D(a0, a1, b0, b1, 1e-9).Empty() {
		t.Error("1e-9 tolerance must not bridge the 1e-7 gap")
	}
	if !ApproxEqual(1, 1+1e-9, 1e-6) {
		t.Error("ApproxEqual with custom tolerance failed")
	}
	if ApproxEqual(1, 1+1e-6, 1e-9) {
		t.Error("custom tighter tolerance must govern")
	}
	if !ApproxZero(1e-12, 1e-9) || ApproxZero(1e-8, 1e-9) {
		t.Error("ApproxZero with custom tolerance failed")
	}
	if DefaultTolerance <= 0 {
		t.Error("DefaultTolerance must be positive")
	}
	// Zero tolerance falls back to the default.
	if IntersectLineLine2D(Vec2{0, 0}, Vec2{1, 1}, Vec2{0, 1}, Vec2{1, 0}, 0).Empty() {
		t.Error("crossing lines with default tolerance must intersect")
	}
}

func TestBulgeArcRoundTrip(t *testing.T) {
	start, end := Vec2{0, 0}, Vec2{4, 0}
	for _, bulge := range []Bulge{0.5, -0.5, 1, -1, 0.25, NewBulgeFromAngle(math.Pi / 2)} {
		params, err := bulge.ArcParameters(start, end)
		if err != nil {
			t.Fatalf("ArcParameters(%v): %v", bulge, err)
		}
		back := NewBulgeFromArc(params)
		if math.Abs(float64(back)-float64(bulge)) > 1e-9 {
			t.Errorf("bulge %v: arc round trip gave %v", bulge, back)
		}
	}
	// Sagitta round trip.
	sag := SagittaFromBulge(0.5, 4)
	if math.Abs(float64(BulgeFromSagitta(sag, 4))-0.5) > 1e-9 {
		t.Errorf("sagitta round trip failed: %v", sag)
	}
}

func TestDecomposeRoundTrip(t *testing.T) {
	translation := Translation(3, -2, 5)
	rotation := RotationZ(0.4).Mul(RotationY(0.2))
	scaling := Scale(2, 3, 0.5)
	matrix := translation.Mul(rotation).Mul(scaling)

	dec, err := matrix.Decompose()
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	if dec.Translation.Distance(Vec3{3, -2, 5}) > 1e-9 {
		t.Errorf("translation %v, want (3,-2,5)", dec.Translation)
	}
	for j, want := range []float64{2, 3, 0.5} {
		if math.Abs(dec.Scale[j]-want) > 1e-9 {
			t.Errorf("scale %d = %v, want %v", j, dec.Scale[j], want)
		}
	}
	// The rotation columns are unit and orthogonal; shear is zero.
	for j := range 3 {
		col := Vec3{dec.Rotation.Get(0, j), dec.Rotation.Get(1, j), dec.Rotation.Get(2, j)}
		if math.Abs(col.Length()-1) > 1e-9 {
			t.Errorf("rotation column %d not unit: %v", j, col)
		}
		for k := range j {
			colK := Vec3{dec.Rotation.Get(0, k), dec.Rotation.Get(1, k), dec.Rotation.Get(2, k)}
			if math.Abs(col.Dot(colK)) > 1e-9 {
				t.Errorf("rotation columns %d, %d not orthogonal", j, k)
			}
		}
	}
	if math.Abs(dec.Shear[0])+math.Abs(dec.Shear[1])+math.Abs(dec.Shear[2]) > 1e-9 {
		t.Errorf("shear of a pure T*R*S matrix: %v", dec.Shear)
	}
	inv, err := matrix.InverseChecked()
	if err != nil {
		t.Fatalf("InverseChecked: %v", err)
	}
	if !matrix.Mul(inv).IsIdentity(1e-9) {
		t.Error("M · M⁻¹ is not identity")
	}
	if _, err := (Matrix44{}).InverseChecked(); err == nil {
		t.Error("singular matrix must fail InverseChecked")
	}
}

func TestBBox3DOperations(t *testing.T) {
	a := NewBoundingBox3D(Vec3{0, 0, 0}, Vec3{2, 2, 2})
	b := NewBoundingBox3D(Vec3{1, 1, 1}, Vec3{3, 3, 3})
	u := a.Union(b)
	if u.Min != (Vec3{0, 0, 0}) || u.Max != (Vec3{3, 3, 3}) {
		t.Errorf("union %v-%v", u.Min, u.Max)
	}
	inter := a.Intersection(b)
	if inter.Min != (Vec3{1, 1, 1}) || inter.Max != (Vec3{2, 2, 2}) {
		t.Errorf("intersection %v-%v", inter.Min, inter.Max)
	}
	far := NewBoundingBox3D(Vec3{5, 5, 5}, Vec3{6, 6, 6})
	if !a.Intersection(far).IsEmpty() {
		t.Error("disjoint intersection must be empty")
	}
	if a.IntersectsBoundingBox(far) {
		t.Error("disjoint boxes must not intersect")
	}
	if !a.IntersectsBoundingBox(b) {
		t.Error("overlapping boxes must intersect")
	}
	if a.ContainsBoundingBox(b) {
		t.Error("a does not contain b")
	}
	if !a.ContainsBoundingBox(NewBoundingBox3D(Vec3{0.5, 0.5, 0.5}, Vec3{1.5, 1.5, 1.5})) {
		t.Error("inner box must be contained")
	}
	empty := BoundingBox3D{}
	if !empty.IsEmpty() {
		t.Error("zero box must be empty")
	}
	shifted := a.Translate(Vec3{1, 0, 0})
	if shifted.Min != (Vec3{1, 0, 0}) || shifted.Max != (Vec3{3, 2, 2}) {
		t.Errorf("translate %v-%v", shifted.Min, shifted.Max)
	}
	corners := a.Corners()
	if len(corners) != 8 {
		t.Fatalf("corners %d, want 8", len(corners))
	}
	grown := a.Expand(1)
	if grown.Min != (Vec3{-1, -1, -1}) {
		t.Errorf("expand min %v", grown.Min)
	}
	rotated := NewBoundingBox3D(Vec3{-1, -1, 0}, Vec3{1, 1, 0}).Transform(RotationZ(math.Pi / 2))
	if math.Abs(rotated.Max[0]-1) > 1e-9 || math.Abs(rotated.Max[1]-1) > 1e-9 {
		t.Errorf("rotated bbox %v-%v", rotated.Min, rotated.Max)
	}
}

// pointToSegmentDist2D returns the distance from p to the segment a-b.
func pointToSegmentDist2D(p, a, b Vec2) float64 {
	ax, ay := a[0], a[1]
	bx, by := b[0], b[1]
	dx, dy := bx-ax, by-ay
	lengthSq := dx*dx + dy*dy
	if lengthSq == 0 {
		return p.Distance(a)
	}
	t := ((p[0]-ax)*dx + (p[1]-ay)*dy) / lengthSq
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return math.Hypot(p[0]-(ax+t*dx), p[1]-(ay+t*dy))
}
