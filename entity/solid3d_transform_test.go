package entity

import (
	"math"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

// matrixFromRows builds a Matrix44 from explicit rows (row-major with
// translation in column 3).
func matrixFromRows(r0, r1, r2, r3 [4]float64) dxfmath.Matrix44 {
	return dxfmath.Matrix44{r0[0], r0[1], r0[2], r0[3], r1[0], r1[1], r1[2], r1[3], r2[0], r2[1], r2[2], r2[3], r3[0], r3[1], r3[2], r3[3]}
}

func vecAlmost(t *testing.T, got dxfmath.Vec3, want dxfmath.Vec3, tol float64, msg string) {
	t.Helper()
	if d := got.Distance(want); d > tol {
		t.Errorf("%s: got %v want %v (distance %g > %g)", msg, got, want, d, tol)
	}
}

// TestSolid3dTransformGeometryAndPayload verifies that a transform moves the
// defining points and directions while the ACIS payload stays untouched.
func TestSolid3dTransformGeometryAndPayload(t *testing.T) {
	s := NewSolid3d()
	s.SetCenter(dxfmath.NewVec3(1, 2, 3))
	s.boundingBox = NewACISBoundingBox(dxfmath.NewVec3(0, 1, 2), dxfmath.NewVec3(2, 3, 4))
	s.entity.SetExtrusion([]float64{0, 0, 1})
	s.SetSATLines([]string{"ACIS 214", "test solid"})

	// Rotate 90 degrees around Z, scale Z by 2, translate by (10, 20, 30).
	m := matrixFromRows(
		[4]float64{0, -1, 0, 10},
		[4]float64{1, 0, 0, 20},
		[4]float64{0, 0, 2, 30},
		[4]float64{0, 0, 0, 1})

	payloadBefore := append([]string(nil), s.SATLines()...)
	if err := s.Transform(m); err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}

	// Center: (1,2,3) -> (-2+10, 1+20, 6+30) = (8, 21, 36).
	vecAlmost(t, s.GetCenter(), dxfmath.NewVec3(8, 21, 36), 1e-9, "center")

	// Bounding box corners transform accordingly: (x,y,2z)+(10,20,30) for
	// corners x∈{0,2}, y∈{1,3}, z∈{2,4} gives x'∈{7,9}, y'∈{20,22}, z'∈{34,38}.
	bb := s.GetBoundingBox()
	vecAlmost(t, bb.Min, dxfmath.NewVec3(7, 20, 34), 1e-9, "bbox min")
	vecAlmost(t, bb.Max, dxfmath.NewVec3(9, 22, 38), 1e-9, "bbox max")

	// Extrusion direction (0,0,1) rotates/scales to (0,0,1) normalized.
	if e := s.entity.Extrusion(); e[0] != 0 || e[1] != 0 || math.Abs(e[2]-1) > 1e-9 {
		t.Errorf("extrusion direction wrong: %v", e)
	}

	// ACIS payload preserved verbatim.
	payloadAfter := s.SATLines()
	if len(payloadAfter) != len(payloadBefore) {
		t.Fatalf("payload line count changed: %d -> %d", len(payloadBefore), len(payloadAfter))
	}
	for i := range payloadBefore {
		if payloadBefore[i] != payloadAfter[i] {
			t.Errorf("payload line %d changed: %q -> %q", i, payloadBefore[i], payloadAfter[i])
		}
	}

	// Composed placement maps local points through the old placement and m.
	local := dxfmath.NewVec3(0.5, 0.5, 0.5)
	// The composed transform must equal m applied AFTER the original
	// (identity) placement, so applying it to a local point equals m*p.
	vecAlmost(t, s.Placement.Apply(local), m.MulVec(local), 1e-9, "composed placement")
}

// TestSolid3dTransformRoundTrip verifies that transforming by M and then by
// M^-1 restores the original defining geometry within tolerance.
func TestSolid3dTransformRoundTrip(t *testing.T) {
	s := NewSolid3d()
	s.SetCenter(dxfmath.NewVec3(1, 2, 3))
	s.boundingBox = NewACISBoundingBox(dxfmath.NewVec3(0, 1, 2), dxfmath.NewVec3(2, 3, 4))

	m := matrixFromRows(
		[4]float64{0, -1, 0, 10},
		[4]float64{1, 0, 0, 20},
		[4]float64{0, 0, 2, 30},
		[4]float64{0, 0, 0, 1})
	inv, ok := m.Inverse()
	if !ok {
		t.Fatal("test matrix should be invertible")
	}

	if err := s.Transform(m); err != nil {
		t.Fatalf("first Transform: %v", err)
	}
	if err := s.Transform(inv); err != nil {
		t.Fatalf("inverse Transform: %v", err)
	}

	vecAlmost(t, s.GetCenter(), dxfmath.NewVec3(1, 2, 3), 1e-7, "center after round trip")
	bb := s.GetBoundingBox()
	vecAlmost(t, bb.Min, dxfmath.NewVec3(0, 1, 2), 1e-7, "bbox min after round trip")
	vecAlmost(t, bb.Max, dxfmath.NewVec3(2, 3, 4), 1e-7, "bbox max after round trip")
}

// TestSolid3dTransformSingularMatrix verifies rejection of singular matrices
// without modifying the solid.
func TestSolid3dTransformSingularMatrix(t *testing.T) {
	s := NewSolid3d()
	s.SetCenter(dxfmath.NewVec3(1, 2, 3))
	s.SetSATLines([]string{"ACIS 214", "test solid"})

	// Zero X scale collapses the solid: singular.
	m := matrixFromRows(
		[4]float64{0, 0, 0, 10},
		[4]float64{0, 1, 0, 20},
		[4]float64{0, 0, 1, 30},
		[4]float64{0, 0, 0, 1})

	err := s.Transform(m)
	if err == nil {
		t.Fatal("expected error for singular matrix")
	}
	if msg := err.Error(); !contains(msg, "singular") {
		t.Errorf("error should name singular matrix: %v", msg)
	}

	// Solid unchanged.
	vecAlmost(t, s.GetCenter(), dxfmath.NewVec3(1, 2, 3), 1e-12, "center after rejected transform")
	if s.Placement == nil {
		t.Fatal("placement transform should remain set")
	}
	// The placement must still be the identity from NewSolid3d.
	identity := NewACASTransform()
	for i := range identity.Matrix {
		if math.Abs(s.Placement.Matrix[i]-identity.Matrix[i]) > 1e-12 {
			t.Fatalf("placement changed after rejected transform at %d", i)
		}
	}
}

// TestSolid3dSetTransformInverse verifies SetTransform derives a real inverse
// instead of an identity placeholder.
func TestSolid3dSetTransformInverse(t *testing.T) {
	s := NewSolid3d()
	tr := &ACASTransform{Matrix: dxfmath.Matrix44{
		2, 0, 0, 5,
		0, 2, 0, 6,
		0, 0, 2, 7,
		0, 0, 0, 1,
	}}
	s.SetTransform(tr)
	if s.InverseTransform == nil {
		t.Fatal("inverse transform not derived")
	}
	// Applying the inverse to a transformed point returns the local point.
	world := tr.Apply(dxfmath.NewVec3(1, 2, 3))
	local := s.InverseTransform.Apply(world)
	vecAlmost(t, local, dxfmath.NewVec3(1, 2, 3), 1e-9, "inverse of placement")
}
