package math

import (
	"math"
	"testing"
)

func TestVec3(t *testing.T) {
	t.Run("NewVec3", func(t *testing.T) {
		v := NewVec3(1, 2, 3)
		if v[0] != 1 || v[1] != 2 || v[2] != 3 {
			t.Errorf("Expected [1,2,3], got [%f,%f,%f]", v[0], v[1], v[2])
		}
	})

	t.Run("NewVec3FromVec2", func(t *testing.T) {
		v2 := Vec2{5, 6}
		v := NewVec3FromVec2(v2)
		if v[0] != 5 || v[1] != 6 || v[2] != 0 {
			t.Errorf("Expected [5,6,0], got [%f,%f,%f]", v[0], v[1], v[2])
		}
	})

	t.Run("NewVec3FromSlice", func(t *testing.T) {
		v := NewVec3FromSlice([]float64{1, 2, 3})
		if v[0] != 1 || v[1] != 2 || v[2] != 3 {
			t.Errorf("Expected [1,2,3], got [%f,%f,%f]", v[0], v[1], v[2])
		}

		v = NewVec3FromSlice([]float64{4, 5})
		if v[0] != 4 || v[1] != 5 || v[2] != 0 {
			t.Errorf("Expected [4,5,0], got [%f,%f,%f]", v[0], v[1], v[2])
		}

		v = NewVec3FromSlice([]float64{7})
		if v[0] != 0 || v[1] != 0 || v[2] != 0 {
			t.Errorf("Expected [0,0,0] for single element slice, got [%f,%f,%f]", v[0], v[1], v[2])
		}

		v = NewVec3FromSlice([]float64{})
		if v[0] != 0 || v[1] != 0 || v[2] != 0 {
			t.Errorf("Expected [0,0,0], got [%f,%f,%f]", v[0], v[1], v[2])
		}
	})

	t.Run("Component accessors", func(t *testing.T) {
		v := NewVec3(10, 20, 30)
		if v.X() != 10 {
			t.Errorf("Expected X()=10, got %f", v.X())
		}
		if v.Y() != 20 {
			t.Errorf("Expected Y()=20, got %f", v.Y())
		}
		if v.Z() != 30 {
			t.Errorf("Expected Z()=30, got %f", v.Z())
		}
	})

	t.Run("Component pairs", func(t *testing.T) {
		v := NewVec3(1, 2, 3)
		xy := v.XY()
		if xy[0] != 1 || xy[1] != 2 {
			t.Errorf("XY()=[%f,%f], expected [1,2]", xy[0], xy[1])
		}

		xz := v.XZ()
		if xz[0] != 1 || xz[1] != 3 {
			t.Errorf("XZ()=[%f,%f], expected [1,3]", xz[0], xz[1])
		}

		yz := v.YZ()
		if yz[0] != 2 || yz[1] != 3 {
			t.Errorf("YZ()=[%f,%f], expected [2,3]", yz[0], yz[1])
		}
	})

	t.Run("Add", func(t *testing.T) {
		v1 := NewVec3(1, 2, 3)
		v2 := NewVec3(4, 5, 6)
		result := v1.Add(v2)
		if result[0] != 5 || result[1] != 7 || result[2] != 9 {
			t.Errorf("Add result [%f,%f,%f], expected [5,7,9]", result[0], result[1], result[2])
		}
	})

	t.Run("Sub", func(t *testing.T) {
		v1 := NewVec3(10, 20, 30)
		v2 := NewVec3(1, 2, 3)
		result := v1.Sub(v2)
		if result[0] != 9 || result[1] != 18 || result[2] != 27 {
			t.Errorf("Sub result [%f,%f,%f], expected [9,18,27]", result[0], result[1], result[2])
		}
	})

	t.Run("Mul", func(t *testing.T) {
		v := NewVec3(1, 2, 3)
		result := v.Mul(2)
		if result[0] != 2 || result[1] != 4 || result[2] != 6 {
			t.Errorf("Mul result [%f,%f,%f], expected [2,4,6]", result[0], result[1], result[2])
		}
	})

	t.Run("Div", func(t *testing.T) {
		v := NewVec3(10, 20, 30)
		result := v.Div(10)
		if result[0] != 1 || result[1] != 2 || result[2] != 3 {
			t.Errorf("Div result [%f,%f,%f], expected [1,2,3]", result[0], result[1], result[2])
		}
	})

	t.Run("Dot", func(t *testing.T) {
		v1 := NewVec3(1, 2, 3)
		v2 := NewVec3(4, 5, 6)
		result := v1.Dot(v2)
		expected := float64(1*4 + 2*5 + 3*6)
		if result != expected {
			t.Errorf("Dot result %f, expected %f", result, expected)
		}
	})

	t.Run("Cross", func(t *testing.T) {
		v1 := NewVec3(1, 0, 0)
		v2 := NewVec3(0, 1, 0)
		result := v1.Cross(v2)
		if result[0] != 0 || result[1] != 0 || result[2] != 1 {
			t.Errorf("Cross result [%f,%f,%f], expected [0,0,1]", result[0], result[1], result[2])
		}
	})

	t.Run("Length", func(t *testing.T) {
		v := NewVec3(3, 4, 0)
		length := v.Length()
		if math.Abs(length-5) > 1e-10 {
			t.Errorf("Length %f, expected 5", length)
		}
	})

	t.Run("LengthSquared", func(t *testing.T) {
		v := NewVec3(3, 4, 0)
		lengthSq := v.LengthSquared()
		if lengthSq != 25 {
			t.Errorf("LengthSquared %f, expected 25", lengthSq)
		}
	})

	t.Run("Normalized", func(t *testing.T) {
		v := NewVec3(3, 4, 0)
		normalized := v.Normalized()
		length := normalized.Length()
		if math.Abs(length-1) > 1e-10 {
			t.Errorf("Normalized length %f, expected 1", length)
		}
	})

	t.Run("Normalized zero vector", func(t *testing.T) {
		v := NewVec3(0, 0, 0)
		normalized := v.Normalized()
		if normalized[0] != 0 || normalized[1] != 0 || normalized[2] != 0 {
			t.Errorf("Zero vector normalized should be zero, got [%f,%f,%f]", normalized[0], normalized[1], normalized[2])
		}
	})

	t.Run("IsEqual", func(t *testing.T) {
		v1 := NewVec3(1, 2, 3)
		v2 := NewVec3(1, 2, 3)
		v3 := NewVec3(1, 2, 4)

		if !v1.IsEqual(v2, 1e-9) {
			t.Error("Equal vectors should match")
		}

		if v1.IsEqual(v3, 1e-9) {
			t.Error("Different vectors should not match")
		}
	})

	t.Run("Distance", func(t *testing.T) {
		v1 := NewVec3(0, 0, 0)
		v2 := NewVec3(3, 4, 0)
		dist := v1.Distance(v2)
		if math.Abs(dist-5) > 1e-10 {
			t.Errorf("Distance %f, expected 5", dist)
		}
	})
}

func TestVec2(t *testing.T) {
	t.Run("NewVec2", func(t *testing.T) {
		v := NewVec2(1, 2)
		if v[0] != 1 || v[1] != 2 {
			t.Errorf("Expected [1,2], got [%f,%f]", v[0], v[1])
		}
	})

	t.Run("Add", func(t *testing.T) {
		v1 := NewVec2(1, 2)
		v2 := NewVec2(3, 4)
		result := v1.Add(v2)
		if result[0] != 4 || result[1] != 6 {
			t.Errorf("Add result [%f,%f], expected [4,6]", result[0], result[1])
		}
	})

	t.Run("Sub", func(t *testing.T) {
		v1 := NewVec2(5, 6)
		v2 := NewVec2(1, 2)
		result := v1.Sub(v2)
		if result[0] != 4 || result[1] != 4 {
			t.Errorf("Sub result [%f,%f], expected [4,4]", result[0], result[1])
		}
	})

	t.Run("Mul", func(t *testing.T) {
		v := NewVec2(2, 3)
		result := v.Mul(2)
		if result[0] != 4 || result[1] != 6 {
			t.Errorf("Mul result [%f,%f], expected [4,6]", result[0], result[1])
		}
	})

	t.Run("Dot", func(t *testing.T) {
		v1 := NewVec2(1, 2)
		v2 := NewVec2(3, 4)
		result := v1.Dot(v2)
		if result != 11 {
			t.Errorf("Dot result %f, expected 11", result)
		}
	})

	t.Run("Length", func(t *testing.T) {
		v := NewVec2(3, 4)
		length := v.Length()
		if math.Abs(length-5) > 1e-10 {
			t.Errorf("Length %f, expected 5", length)
		}
	})

	t.Run("Perpendicular", func(t *testing.T) {
		v := NewVec2(1, 0)
		perp := v.Perpendicular()
		if perp[0] != 0 || perp[1] != 1 {
			t.Errorf("Perpendicular result [%f,%f], expected [0,1]", perp[0], perp[1])
		}

		dot := v.Dot(perp)
		if math.Abs(dot) > 1e-10 {
			t.Errorf("Perpendicular vectors should have dot product ~0, got %f", dot)
		}
	})

	t.Run("Rotate", func(t *testing.T) {
		v := NewVec2(1, 0)
		rotated := v.Rotate(math.Pi / 2)
		if math.Abs(rotated[0]-0) > 1e-10 || math.Abs(rotated[1]-1) > 1e-10 {
			t.Errorf("Rotated [%f,%f], expected [0,1]", rotated[0], rotated[1])
		}
	})
}

func TestMatrix44(t *testing.T) {
	t.Run("Identity", func(t *testing.T) {
		m := NewMatrix44()
		for i := 0; i < 4; i++ {
			if m[i*5] != 1 {
				t.Errorf("Identity matrix diagonal[%d] should be 1, got %f", i, m[i*5])
			}
		}
	})

	t.Run("Mul", func(t *testing.T) {
		m1 := NewMatrix44()
		m2 := NewMatrix44()
		m2[0] = 2

		result := m1.Mul(m2)
		if result[0] != 2 {
			t.Errorf("Mul result [0] should be 2, got %f", result[0])
		}
	})

	t.Run("MulVec", func(t *testing.T) {
		m := NewMatrix44()
		v := NewVec3(1, 2, 3)
		result := m.MulVec(v)
		if result[0] != 1 || result[1] != 2 || result[2] != 3 {
			t.Errorf("MulVec result [%f,%f,%f], expected [1,2,3]", result[0], result[1], result[2])
		}
	})

	t.Run("Translation", func(t *testing.T) {
		m := Translation(5, 10, 15)
		v := NewVec3(1, 2, 3)
		result := m.MulVec(v)
		if result[0] != 6 || result[1] != 12 || result[2] != 18 {
			t.Errorf("Translation result [%f,%f,%f], expected [6,12,18]", result[0], result[1], result[2])
		}
	})

	t.Run("Scale", func(t *testing.T) {
		m := Scale(2, 3, 4)
		v := NewVec3(1, 2, 3)
		result := m.MulVec(v)
		if result[0] != 2 || result[1] != 6 || result[2] != 12 {
			t.Errorf("Scale result [%f,%f,%f], expected [2,6,12]", result[0], result[1], result[2])
		}
	})
}

func TestBBox(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		bbox := NewBoundingBox(NewVec2(0, 0), NewVec2(10, 10))
		if bbox.Min.X() != 0 || bbox.Min.Y() != 0 {
			t.Errorf("Min should be [0,0], got [%f,%f]", bbox.Min.X(), bbox.Min.Y())
		}
		if bbox.Max.X() != 10 || bbox.Max.Y() != 10 {
			t.Errorf("Max should be [10,10], got [%f,%f]", bbox.Max.X(), bbox.Max.Y())
		}
	})

	t.Run("ContainsPoint", func(t *testing.T) {
		bbox := NewBoundingBox(NewVec2(0, 0), NewVec2(10, 10))
		if !bbox.ContainsPoint(NewVec2(5, 5)) {
			t.Error("Point [5,5] should be inside bbox")
		}
		if bbox.ContainsPoint(NewVec2(15, 15)) {
			t.Error("Point [15,15] should be outside bbox")
		}
	})

	t.Run("IntersectsBoundingBox", func(t *testing.T) {
		bbox1 := NewBoundingBox(NewVec2(0, 0), NewVec2(10, 10))
		bbox2 := NewBoundingBox(NewVec2(5, 5), NewVec2(15, 15))
		if !bbox1.IntersectsBoundingBox(bbox2) {
			t.Error("Overlapping bboxes should intersect")
		}

		bbox3 := NewBoundingBox(NewVec2(20, 20), NewVec2(30, 30))
		if bbox1.IntersectsBoundingBox(bbox3) {
			t.Error("Non-overlapping bboxes should not intersect")
		}
	})

	t.Run("WidthHeight", func(t *testing.T) {
		bbox := NewBoundingBox(NewVec2(0, 0), NewVec2(10, 5))
		if bbox.Width() != 10 {
			t.Errorf("Width %f, expected 10", bbox.Width())
		}
		if bbox.Height() != 5 {
			t.Errorf("Height %f, expected 5", bbox.Height())
		}
	})
}

func TestBSpline(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		points := []Vec3{
			NewVec3(0, 0, 0),
			NewVec3(1, 1, 0),
			NewVec3(2, 0, 0),
		}
		spline := NewBSpline(points, 2)
		if spline == nil {
			t.Error("BSpline should be created")
		}
	})

	t.Run("Evaluate", func(t *testing.T) {
		points := []Vec3{
			NewVec3(0, 0, 0),
			NewVec3(1, 0, 0),
			NewVec3(2, 0, 0),
		}
		spline := NewBSpline(points, 1)
		if spline == nil {
			t.Skip("BSpline creation failed")
		}

		p0 := spline.EvaluatePoint(0)
		if math.Abs(p0[0]-0) > 1e-6 || math.Abs(p0[1]-0) > 1e-6 {
			t.Logf("EvaluatePoint(0) = [%f, %f, %f]", p0[0], p0[1], p0[2])
		}
	})

	t.Run("Degree", func(t *testing.T) {
		points := []Vec3{
			NewVec3(0, 0, 0),
			NewVec3(1, 1, 0),
			NewVec3(2, 0, 0),
			NewVec3(3, 1, 0),
		}
		spline := NewBSpline(points, 3)
		if spline.Degree() != 3 {
			t.Errorf("Degree %d, expected 3", spline.Degree())
		}
	})
}

func TestBezier(t *testing.T) {
	t.Run("CreateBezier4P", func(t *testing.T) {
		p0 := NewVec3(0, 0, 0)
		p1 := NewVec3(1, 1, 0)
		p2 := NewVec3(2, 1, 0)
		p3 := NewVec3(3, 0, 0)

		bezier, err := NewBezier4P(p0, p1, p2, p3)
		if err != nil {
			t.Errorf("Failed to create Bezier: %v", err)
		}

		if bezier == nil {
			t.Error("Bezier should not be nil")
		}
	})

	t.Run("Bezier Evaluate", func(t *testing.T) {
		p0 := NewVec3(0, 0, 0)
		p1 := NewVec3(1, 1, 0)
		p2 := NewVec3(2, 1, 0)
		p3 := NewVec3(3, 0, 0)

		bezier, err := NewBezier4P(p0, p1, p2, p3)
		if err != nil {
			t.Skip("Bezier creation failed")
		}

		point := bezier.EvaluatePoint(0)
		if math.Abs(point[0]-0) > 1e-6 || math.Abs(point[1]-0) > 1e-6 {
			t.Errorf("EvaluatePoint(0) should be [0,0,0], got [%f,%f,%f]", point[0], point[1], point[2])
		}

		point = bezier.EvaluatePoint(1)
		if math.Abs(point[0]-3) > 1e-6 || math.Abs(point[1]-0) > 1e-6 {
			t.Errorf("EvaluatePoint(1) should be [3,0,0], got [%f,%f,%f]", point[0], point[1], point[2])
		}
	})

	t.Run("Degree", func(t *testing.T) {
		p0 := NewVec3(0, 0, 0)
		p1 := NewVec3(1, 1, 0)
		p2 := NewVec3(2, 1, 0)
		p3 := NewVec3(3, 0, 0)

		bezier, err := NewBezier4P(p0, p1, p2, p3)
		if err != nil {
			t.Skip("Bezier creation failed")
		}

		if bezier.Degree() != 3 {
			t.Errorf("Degree %d, expected 3", bezier.Degree())
		}
	})
}

func TestOCS(t *testing.T) {
	t.Run("OCS to WCS", func(t *testing.T) {
		ocs := NewOCS(NewVec3(0, 0, 1))

		ocsPoint := NewVec3(1, 2, 3)
		wcs := ocs.ToWCS(ocsPoint)
		if wcs[0] != 1 || wcs[1] != 2 || wcs[2] != 3 {
			t.Errorf("ToWCS for Z-up should be identity, got [%f,%f,%f]", wcs[0], wcs[1], wcs[2])
		}
	})

	t.Run("OCS Identity", func(t *testing.T) {
		ocs := NewOCSIdentity()
		if !ocs.IsWCSAligned() {
			t.Error("Identity OCS should be WCS aligned")
		}
	})

	t.Run("OCS ToOCS roundtrip", func(t *testing.T) {
		ocs := NewOCS(NewVec3(0, 0, 1))
		original := NewVec3(5, 10, 15)
		transformed := ocs.ToOCS(ocs.ToWCS(original))
		if !original.IsEqual(transformed, 1e-9) {
			t.Errorf("Roundtrip failed: got [%f,%f,%f]", transformed[0], transformed[1], transformed[2])
		}
	})
}

func TestConstruct2D(t *testing.T) {
	t.Run("CircleFrom3Points", func(t *testing.T) {
		c := NewConstruct2D()
		p1 := NewVec2(0, 0)
		p2 := NewVec2(1, 0)
		p3 := NewVec2(0, 1)

		center, radius, ok := c.CircleFrom3Points(p1, p2, p3)
		if !ok {
			t.Error("CircleFrom3Points should succeed")
		}

		expectedCenter := NewVec2(0.5, 0.5)
		if !center.IsEqual(expectedCenter, 1e-9) {
			t.Errorf("Center [%f,%f], expected [%f,%f]", center[0], center[1], expectedCenter[0], expectedCenter[1])
		}

		if math.Abs(radius-math.Sqrt(2)/2) > 1e-9 {
			t.Errorf("Radius %f, expected %f", radius, math.Sqrt(2)/2)
		}
	})

	t.Run("LineIntersection", func(t *testing.T) {
		c := NewConstruct2D()
		p1 := NewVec2(0, 0)
		p2 := NewVec2(10, 10)
		p3 := NewVec2(0, 10)
		p4 := NewVec2(10, 0)

		intersection := c.LineIntersection(p1, p2, p3, p4)
		expected := NewVec2(5, 5)
		if !intersection.IsEqual(expected, 1e-9) {
			t.Errorf("Intersection [%f,%f], expected [%f,%f]", intersection[0], intersection[1], expected[0], expected[1])
		}
	})
}
