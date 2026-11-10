package math

import (
	"math"
	"testing"
)

func TestTolerancePolicy(t *testing.T) {
	if !ApproxEqual(1, 1+5e-10, DefaultTolerance) {
		t.Error("default tolerance should accept 5e-10")
	}
	if ApproxEqual(1, 1+5e-6, DefaultTolerance) {
		t.Error("default tolerance should reject 5e-6")
	}
	if !ApproxEqual(1, 1+5e-6, 1e-3) {
		t.Error("custom tolerance 1e-3 should accept 5e-6")
	}
	if !ApproxZero(-4e-10, DefaultTolerance) {
		t.Error("default tolerance should treat 4e-10 as zero")
	}
	if !ApproxZero(1, 10) {
		t.Error("tolerance 10 should treat 1 as zero")
	}
	// A non-positive tolerance falls back to the documented default.
	if !ApproxEqual(2, 2+5e-10, 0) {
		t.Error("tolerance 0 should fall back to DefaultTolerance")
	}
}

func TestZeroVectorNormalizationIsSafe(t *testing.T) {
	for name, got := range map[string]Vec3{
		"vec3": Vec3{}.Normalized(),
	} {
		if math.IsNaN(got.X()) || math.IsNaN(got.Y()) || math.IsNaN(got.Z()) {
			t.Errorf("%s zero normalization produced NaN: %v", name, got)
		}
		if got != (Vec3{}) {
			t.Errorf("%s zero normalization = %v, want zero vector", name, got)
		}
	}
	if v := (Vec2{}).Normalized(); v != (Vec2{}) {
		t.Errorf("Vec2 zero normalization = %v, want zero vector", v)
	}
	// In-place normalization must not panic or produce NaN either.
	v3 := Vec3{}
	v3.Normalize()
	v2 := Vec2{}
	v2.Normalize()
	if math.IsNaN(v3[0]) || math.IsNaN(v2[0]) {
		t.Error("in-place normalization of a zero vector produced NaN")
	}
	// A unit vector survives normalization, and the cross product is perpendicular.
	a := Vec3{1, 0, 0}
	b := Vec3{0, 1, 0}
	c := a.Cross(b)
	if math.Abs(c.Dot(a)) > DefaultTolerance || math.Abs(c.Dot(b)) > DefaultTolerance {
		t.Errorf("cross product %v is not perpendicular to inputs", c)
	}
	if c.Length() != 1 {
		t.Errorf("cross product of two unit axes has length %v, want 1", c.Length())
	}
}

func TestMatrixDecompositionRoundTrip(t *testing.T) {
	m := Translation(3, -2, 5).
		Mul(RotationZ(math.Pi / 6)).
		Mul(RotationX(0.2)).
		Mul(Scale(2, 0.5, 4))

	d, err := m.Decompose()
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	wantTranslation := Vec3{3, -2, 5}
	if got := d.Translation; got.Distance(wantTranslation) > 1e-9 {
		t.Errorf("translation = %v, want %v", got, wantTranslation)
	}
	wantScale := Vec3{2, 0.5, 4}
	if got := d.Scale; got.Distance(wantScale) > 1e-9 {
		t.Errorf("scale = %v, want %v", got, wantScale)
	}
	// Rebuilding the matrix from the parts must reproduce the input.
	rebuilt := Translation(d.Translation.X(), d.Translation.Y(), d.Translation.Z()).
		Mul(d.Rotation).
		Mul(Scale(d.Scale[0], d.Scale[1], d.Scale[2]))
	for i := range 16 {
		if math.Abs(m[i]-rebuilt[i]) > 1e-9 {
			t.Fatalf("rebuilt matrix differs at index %d: %v vs %v", i, rebuilt[i], m[i])
		}
	}
	// The rotation part must stay orthonormal.
	for i, axis := range []Vec3{
		{d.Rotation[0], d.Rotation[4], d.Rotation[8]},
		{d.Rotation[1], d.Rotation[5], d.Rotation[9]},
		{d.Rotation[2], d.Rotation[6], d.Rotation[10]},
	} {
		if math.Abs(axis.Length()-1) > 1e-9 {
			t.Errorf("rotation column %d length = %v, want 1", i, axis.Length())
		}
	}
	wantOrigin := Vec3{3, -2, 5}
	if got := m.MulVec(Vec3{}); got.Distance(wantOrigin) > 1e-9 {
		t.Errorf("transformed origin = %v, want %v", got, wantOrigin)
	}
}

func TestSingularMatrixErrors(t *testing.T) {
	singular := Matrix44{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 1,
	}
	if _, err := singular.Decompose(); err == nil {
		t.Error("Decompose of a singular matrix must fail")
	} else if err != ErrSingularMatrix {
		t.Errorf("Decompose error = %v, want ErrSingularMatrix", err)
	}
	if _, err := singular.InverseChecked(); err != ErrSingularMatrix {
		t.Errorf("InverseChecked error = %v, want ErrSingularMatrix", err)
	}
	if _, err := NewMatrix44().InverseChecked(); err != nil {
		t.Errorf("InverseChecked of the identity failed: %v", err)
	}
	// A flat projection (z collapsed) is singular.
	flat := Scale(1, 1, 0)
	if _, err := flat.Decompose(); err != ErrSingularMatrix {
		t.Errorf("flat scale decomposition = %v, want ErrSingularMatrix", err)
	}
}

func TestBoundingBox3DOperations(t *testing.T) {
	a := NewBoundingBox3D(Vec3{0, 0, 0}, Vec3{1, 1, 1})
	b := NewBoundingBox3D(Vec3{0.5, 0.5, 0.5}, Vec3{2, 2, 2})

	union := a.Union(b)
	if union.Min != (Vec3{0, 0, 0}) || union.Max != (Vec3{2, 2, 2}) {
		t.Errorf("union = %v..%v", union.Min, union.Max)
	}
	inter := a.Intersection(b)
	if inter.Min != (Vec3{0.5, 0.5, 0.5}) || inter.Max != (Vec3{1, 1, 1}) {
		t.Errorf("intersection = %v..%v", inter.Min, inter.Max)
	}
	if !a.IntersectsBoundingBox(b) {
		t.Error("overlapping boxes must intersect")
	}
	gap := NewBoundingBox3D(Vec3{5, 5, 5}, Vec3{6, 6, 6})
	if a.IntersectsBoundingBox(gap) {
		t.Error("disjoint boxes must not intersect")
	}
	if !gap.Intersection(a).IsEmpty() {
		t.Error("disjoint box intersection must be empty")
	}
	if !a.ContainsBoundingBox(NewBoundingBox3D(Vec3{0.2, 0.2, 0.2}, Vec3{0.8, 0.8, 0.8})) {
		t.Error("containment failed")
	}
	if a.ContainsBoundingBox(b) {
		t.Error("partial overlap is not containment")
	}
	expanded := a.Expand(1)
	if expanded.Min != (Vec3{-1, -1, -1}) || expanded.Max != (Vec3{2, 2, 2}) {
		t.Errorf("expand = %v..%v", expanded.Min, expanded.Max)
	}
	translated := a.Translate(Vec3{1, 0, 0})
	if translated.Min != (Vec3{1, 0, 0}) || translated.Max != (Vec3{2, 1, 1}) {
		t.Errorf("translate = %v..%v", translated.Min, translated.Max)
	}
	if len(a.Corners()) != 8 {
		t.Errorf("corner count = %d, want 8", len(a.Corners()))
	}
	// A 90 degree rotation about Z maps the unit box to X in [-1,0].
	rotated := a.Transform(RotationZ(math.Pi / 2))
	if math.Abs(rotated.Min.X()+1) > 1e-12 || math.Abs(rotated.Max.X()) > 1e-12 {
		t.Errorf("rotated box X range = %v..%v, want -1..0", rotated.Min.X(), rotated.Max.X())
	}
	scaled := a.Transform(Scale(2, 3, 1))
	if math.Abs(scaled.Max.X()-2) > 1e-12 || math.Abs(scaled.Max.Y()-3) > 1e-12 || math.Abs(scaled.Max.Z()-1) > 1e-12 {
		t.Errorf("scaled box max = %v, want (2,3,1)", scaled.Max)
	}
	if (BoundingBox3D{}).Union(a) != a {
		t.Error("union with an empty box must return the other box")
	}
	if !(BoundingBox3D{}).IsEmpty() {
		t.Error("zero box must report empty")
	}
}

func TestBoundingBox2DTransform(t *testing.T) {
	box := NewBoundingBoxFromPoints([]Vec2{{0, 0}, {2, 1}})
	rotated := box.Transform(RotationZ(math.Pi / 2))
	if math.Abs(rotated.Min.X()+1) > 1e-12 || math.Abs(rotated.Max.X()) > 1e-12 {
		t.Errorf("rotated box X range = %v..%v, want -1..0", rotated.Min.X(), rotated.Max.X())
	}
	if math.Abs(rotated.Min.Y()) > 1e-12 || math.Abs(rotated.Max.Y()-2) > 1e-12 {
		t.Errorf("rotated box Y range = %v..%v, want 0..2", rotated.Min.Y(), rotated.Max.Y())
	}
}
