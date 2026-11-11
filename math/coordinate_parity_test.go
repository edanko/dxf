package math

import (
	"math"
	"testing"
)

func assertVecClose(t *testing.T, label string, got, want Vec3, tolerance float64) {
	t.Helper()
	if got.Distance(want) > tolerance {
		t.Fatalf("%s = %v, want %v (delta %v)", label, got, want, got.Distance(want))
	}
}

func normalizeAngleDifference(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

func TestOCSPointAndDirectionRoundTrips(t *testing.T) {
	extrusions := []struct {
		name string
		vec  Vec3
	}{
		{"extrusion-wcs", Vec3{0, 0, 1}},
		{"extrusion-plus", Vec3{0, 0, -1}},
		{"extrusion-negative-x", Vec3{1, 0, 0}},
		{"extrusion-tilted", Vec3{0.3, 0.4, 0.85}},
	}
	points := []Vec3{{1, 2, 3}, {-4, 5, 0.5}, {0, 0, 0}}
	directions := []Vec3{{1, 2, 3}, {0, 0, -1}, {-3, 0.5, 2}}

	for _, tc := range extrusions {
		t.Run(tc.name, func(t *testing.T) {
			ocs := NewOCS(tc.vec)
			for _, p := range points {
				assertVecClose(t, "ToOCS(ToWCS(point))", ocs.ToOCS(ocs.ToWCS(p)), p, 1e-9)
			}
			for _, d := range directions {
				got := ocs.DirectionToOCS(ocs.DirectionToWCS(d))
				assertVecClose(t, "DirectionToOCS(DirectionToWCS(direction))", got, d, 1e-9)
			}
		})
	}
}

func TestOCSAxisVectors(t *testing.T) {
	for _, extrusion := range []Vec3{{0, 0, 1}, {0, 0, -1}, {1, 0, 0}, {0.3, 0.4, 0.85}} {
		ocs := NewOCS(extrusion)
		n := ocs.Extrusion()
		xAxis, yAxis := ocs.AxisVectors()
		if math.Abs(xAxis.Length()-1) > 1e-9 {
			t.Errorf("x axis %v is not unit length", xAxis)
		}
		if math.Abs(yAxis.Length()-1) > 1e-9 {
			t.Errorf("y axis %v is not unit length", yAxis)
		}
		if math.Abs(xAxis.Dot(n)) > 1e-9 {
			t.Errorf("x axis %v is not perpendicular to extrusion %v", xAxis, n)
		}
		if math.Abs(xAxis.Dot(yAxis)) > 1e-9 {
			t.Errorf("axes %v and %v are not perpendicular", xAxis, yAxis)
		}
		if math.Abs(yAxis.Dot(n)) > 1e-9 {
			t.Errorf("y axis %v is not perpendicular to extrusion %v", yAxis, n)
		}
		// Right-handed: xAxis × yAxis = extrusion (ezdxf Matrix44.ucs).
		if xAxis.Cross(yAxis).Distance(n) > 1e-9 {
			t.Errorf("%v × %v is not the extrusion %v", xAxis, yAxis, n)
		}
	}
}

// TestOCSAxisMatchesEzdxfProbe checks the plane basis against the ezdxf
// 1.4.6b0 OCS.to_wcs probe (src/ezdxf/src/ezdxf/math/ucs.py:37) for the OCS
// X and Y axes (to_wcs((1,0,0)) and to_wcs((0,1,0))).
func TestOCSAxisMatchesEzdxfProbe(t *testing.T) {
	cases := []struct {
		extrusion Vec3
		x, y      Vec3
	}{
		{Vec3{0, 0, 1}, Vec3{1, 0, 0}, Vec3{0, 1, 0}},
		{Vec3{0, 0, -1}, Vec3{-1, 0, 0}, Vec3{0, 1, 0}},
		{Vec3{1, 0, 0}, Vec3{0, 1, 0}, Vec3{0, 0, 1}},
		{Vec3{0.3, 0.4, 0.85}, Vec3{-0.8, 0.6, 0}, Vec3{-0.5171605290946617, -0.6895473721262156, 0.5070201265633938}},
	}
	for _, tc := range cases {
		ocs := NewOCS(tc.extrusion)
		xAxis, yAxis := ocs.AxisVectors()
		assertVecClose(t, "axis vectors x", xAxis, tc.x, 1e-9)
		assertVecClose(t, "axis vectors y", yAxis, tc.y, 1e-9)
		assertVecClose(t, "ToWCS((1,0,0))", ocs.ToWCS(Vec3{1, 0, 0}), tc.x, 1e-9)
		assertVecClose(t, "ToWCS((0,1,0))", ocs.ToWCS(Vec3{0, 1, 0}), tc.y, 1e-9)
	}
}

// TestOCSAngleToWCSMatchesEzdxfProbe checks AngleToWCS against the ezdxf
// probe of the WCS XY angle of to_wcs(Vec3.from_deg_angle(θ))
// (normalized to degrees in [0, 360)).
func TestOCSAngleToWCSMatchesEzdxfProbe(t *testing.T) {
	cases := []struct {
		extrusion  Vec3
		ocsDegrees []float64
		wantDegs   []float64
	}{
		{Vec3{0, 0, 1}, []float64{0, 30, 90, 150}, []float64{0, 30, 90, 150}},
		{Vec3{0, 0, -1}, []float64{0, 30, 90, 150}, []float64{180, 150, 90, 30}},
		{Vec3{1, 0, 0}, []float64{0, 30, 90, 150}, []float64{90, 90, 90, 270}},
		{Vec3{0.3, 0.4, 0.85}, []float64{0, 30, 90, 150}, []float64{143.130102354156, 169.586782925365, 233.130102354156, 296.673421782947}},
	}
	for _, tc := range cases {
		ocs := NewOCS(tc.extrusion)
		for i, deg := range tc.ocsDegrees {
			got := ocs.AngleToWCS(deg * math.Pi / 180)
			gotDeg := math.Mod(got*180/math.Pi, 360)
			if gotDeg < 0 {
				gotDeg += 360
			}
			if math.Abs(gotDeg-tc.wantDegs[i]) > 1e-9 {
				t.Errorf("extrusion %v: AngleToWCS(%v°) = %v°, want %v°",
					tc.extrusion, deg, gotDeg, tc.wantDegs[i])
			}
		}
	}
}

func TestOCSAngleRoundTrip(t *testing.T) {
	angles := []float64{0, math.Pi / 6, math.Pi / 2, math.Pi, -math.Pi / 3, 2}
	// Extrusions whose OCS plane projects non-degenerately onto the WCS XY
	// plane (the X-axis extrusion projects the basis Y onto Z, collapsing
	// every angle to ±90°, so it has no inverse).
	for _, extrusion := range []Vec3{{0, 0, 1}, {0, 0, -1}, {0.3, 0.4, 0.85}} {
		ocs := NewOCS(extrusion)
		for _, angle := range angles {
			back := ocs.AngleToOCS(ocs.AngleToWCS(angle))
			if diff := normalizeAngleDifference(back - angle); math.Abs(diff) > 1e-9 {
				t.Errorf("extrusion %v: AngleToOCS(AngleToWCS(%v)) = %v, off by %v",
					extrusion, angle, back, diff)
			}
		}
	}
}

func TestUCSPointVectorDirectionRoundTrips(t *testing.T) {
	spaces := []struct {
		name string
		ucs  *UCS
	}{
		{"world", WorldUCS()},
		{"rotated-z", NewUCS(Vec3{1, 2, 3}, Vec3{0, 1, 0}, Vec3{-1, 0, 0}, Vec3{0, 0, 1})},
		{"tilted", NewUCS(Vec3{1, -2, 3}, Vec3{1, 1, 0}, Vec3{-1, 1, 0}, Vec3{0, 0, 1})},
		{"from-axis-angle", UCSFromAxisAndAngle(Vec3{2, 2, 0}, Vec3{0.5, 0.5, 1}, 0.7)},
	}
	points := []Vec3{{4, 5, 6}, {-1, 0, 2.5}}
	vectors := []Vec3{{1, 0, 0}, {2, -3, 4}}

	for _, tc := range spaces {
		t.Run(tc.name, func(t *testing.T) {
			ucs := tc.ucs
			if !ucs.IsOrthogonal() {
				t.Fatalf("frame of %s is not orthogonal", tc.name)
			}
			if !ucs.IsRightHanded() {
				t.Fatalf("frame of %s is not right handed", tc.name)
			}
			for _, p := range points {
				assertVecClose(t, "ToUCS(ToWorld(point))", ucs.ToUCS(ucs.ToWorld(p)), p, 1e-9)
				if got := ucs.Matrix().MulVec(p); got.Distance(ucs.ToWorld(p)) > 1e-9 {
					t.Errorf("Matrix().MulVec(point) = %v, want %v", got, ucs.ToWorld(p))
				}
			}
			for _, v := range vectors {
				assertVecClose(t, "InverseTransformVector(TransformVector(v))",
					ucs.InverseTransformVector(ucs.TransformVector(v)), v, 1e-9)
				assertVecClose(t, "InverseTransformDirection(TransformDirection(d))",
					ucs.InverseTransformDirection(ucs.TransformDirection(v)), v, 1e-9)
			}
		})
	}
}

func TestUCSFromEulerAnglesIsOrthogonalAndRightHanded(t *testing.T) {
	ucs := UCSFromEulerAngles(Vec3{1, 2, 3}, math.Pi/6, math.Pi/5, math.Pi/7)
	if !ucs.IsOrthogonal() {
		t.Fatal("euler frame is not orthogonal")
	}
	if !ucs.IsRightHanded() {
		t.Fatal("euler frame is not right handed")
	}
	point := Vec3{7, -3, 2}
	assertVecClose(t, "ToUCS(ToWorld(point))", ucs.ToUCS(ucs.ToWorld(point)), point, 1e-9)
	vector := Vec3{1, 2, 3}
	assertVecClose(t, "vector round trip", ucs.InverseTransformVector(ucs.TransformVector(vector)), vector, 1e-9)
}

func TestWorldUCSIdentity(t *testing.T) {
	ucs := WorldUCS()
	point := Vec3{1.5, -2.5, 3.5}
	assertVecClose(t, "WorldUCS().ToWorld(point)", ucs.ToWorld(point), point, 1e-12)
	assertVecClose(t, "WorldUCS().ToUCS(point)", ucs.ToUCS(point), point, 1e-12)
}
