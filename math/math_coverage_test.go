package math

import (
	"math"
	"testing"
)

func TestBoundingBoxExtended(t *testing.T) {
	t.Run("NewBoundingBoxFromPointsEmpty", func(t *testing.T) {
		bb := NewBoundingBoxFromPoints([]Vec2{})
		if !bb.IsEmpty() {
			t.Error("Empty points should create empty bounding box")
		}
	})

	t.Run("NewBoundingBoxFrom3DPointsEmpty", func(t *testing.T) {
		bb := NewBoundingBoxFrom3DPoints([]Vec3{})
		if bb.Min.X() != 0 || bb.Max.X() != 0 {
			t.Error("Empty 3D points should create zero bounding box")
		}
	})

	t.Run("NewBoundingBoxFrom3DPoints", func(t *testing.T) {
		points := []Vec3{
			{1, 2, 3},
			{-1, -2, -3},
			{5, 5, 5},
		}
		bb := NewBoundingBoxFrom3DPoints(points)
		if bb.Min.X() != -1 || bb.Min.Y() != -2 {
			t.Errorf("Min should be [-1,-2], got [%f,%f]", bb.Min.X(), bb.Min.Y())
		}
		if bb.Max.X() != 5 || bb.Max.Y() != 5 {
			t.Errorf("Max should be [5,5], got [%f,%f]", bb.Max.X(), bb.Max.Y())
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		if bb.IsEmpty() {
			t.Error("Valid bbox should not be empty")
		}

		emptyBB := BoundingBox{}
		if !emptyBB.IsEmpty() {
			t.Error("Default bbox should be empty")
		}
	})

	t.Run("Area", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 5})
		if bb.Area() != 50 {
			t.Errorf("Area should be 50, got %f", bb.Area())
		}

		emptyBB := BoundingBox{}
		if emptyBB.Area() != 0 {
			t.Error("Empty bbox area should be 0")
		}
	})

	t.Run("Size", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 5})
		size := bb.Size()
		if size.X() != 10 || size.Y() != 5 {
			t.Errorf("Size should be [10,5], got [%f,%f]", size.X(), size.Y())
		}
	})

	t.Run("ContainsBoundingBox", func(t *testing.T) {
		outer := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		inner := NewBoundingBox(Vec2{2, 2}, Vec2{5, 5})

		if !outer.ContainsBoundingBox(inner) {
			t.Error("Outer should contain inner bbox")
		}

		overlapping := NewBoundingBox(Vec2{5, 5}, Vec2{15, 15})
		if outer.ContainsBoundingBox(overlapping) {
			t.Error("Overlapping bbox should not be contained")
		}
	})

	t.Run("Intersection", func(t *testing.T) {
		bb1 := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		bb2 := NewBoundingBox(Vec2{5, 5}, Vec2{15, 15})

		intersect := bb1.Intersection(bb2)
		if intersect.Min.X() != 5 || intersect.Max.X() != 10 {
			t.Errorf("Intersection should be [5,5]-[10,10], got [%f,%f]-[%f,%f]",
				intersect.Min.X(), intersect.Min.Y(), intersect.Max.X(), intersect.Max.Y())
		}

		nonIntersect := NewBoundingBox(Vec2{20, 20}, Vec2{30, 30})
		if !nonIntersect.Intersection(bb1).IsEmpty() {
			t.Error("Non-intersecting bboxes should have empty intersection")
		}
	})

	t.Run("Expand", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{5, 5}, Vec2{10, 10})
		expanded := bb.Expand(2)

		if expanded.Min.X() != 3 || expanded.Max.X() != 12 {
			t.Errorf("Expanded bbox should extend by margin, got [%f,%f]-[%f,%f]",
				expanded.Min.X(), expanded.Min.Y(), expanded.Max.X(), expanded.Max.Y())
		}

		if expanded.IsEmpty() {
			t.Error("Expanded valid bbox should not be empty")
		}

		emptyBB := BoundingBox{}
		if !emptyBB.Expand(5).IsEmpty() {
			t.Error("Expanded empty bbox should remain empty")
		}
	})

	t.Run("ExpandToPoint", func(t *testing.T) {
		emptyBB := BoundingBox{}
		expanded := emptyBB.ExpandToPoint(Vec2{5, 5})

		if expanded.Min.X() != 5 || expanded.Max.X() != 5 {
			t.Errorf("Expanded empty bbox to point should equal point, got [%f,%f]-[%f,%f]",
				expanded.Min.X(), expanded.Min.Y(), expanded.Max.X(), expanded.Max.Y())
		}

		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		expanded = bb.ExpandToPoint(Vec2{15, 15})
		if expanded.Max.X() != 15 {
			t.Errorf("ExpandToPoint should extend max, got Max.X()=%f", expanded.Max.X())
		}
	})

	t.Run("ClosestPoint", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})

		inside := Vec2{5, 5}
		if bb.ClosestPoint(inside) != inside {
			t.Error("ClosestPoint for inside point should return same point")
		}

		outside := Vec2{15, 5}
		closest := bb.ClosestPoint(outside)
		if closest.X() != 10 || closest.Y() != 5 {
			t.Errorf("ClosestPoint for right of bbox should clamp X to 10, got [%f,%f]",
				closest.X(), closest.Y())
		}

		cornerOutside := Vec2{15, 15}
		closest = bb.ClosestPoint(cornerOutside)
		if closest.X() != 10 || closest.Y() != 10 {
			t.Errorf("ClosestPoint for corner outside should clamp to corner, got [%f,%f]",
				closest.X(), closest.Y())
		}

		_ = bb
	})

	t.Run("DistanceToPoint", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})

		inside := Vec2{5, 5}
		if bb.DistanceToPoint(inside) != 0 {
			t.Errorf("Distance to inside point should be 0, got %f", bb.DistanceToPoint(inside))
		}

		outside := Vec2{15, 5}
		dist := bb.DistanceToPoint(outside)
		if math.Abs(dist-5) > 1e-10 {
			t.Errorf("Distance to point (15,5) should be 5, got %f", dist)
		}
	})

	t.Run("Translate", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		translated := bb.Translate(Vec2{5, -5})

		if translated.Min.X() != 5 || translated.Min.Y() != -5 {
			t.Errorf("Translated bbox should be shifted, got [%f,%f]-[%f,%f]",
				translated.Min.X(), translated.Min.Y(), translated.Max.X(), translated.Max.Y())
		}

		emptyBB := BoundingBox{}
		if !emptyBB.Translate(Vec2{5, 5}).IsEmpty() {
			t.Error("Translated empty bbox should remain empty")
		}
	})

	t.Run("Scale", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		scaled := bb.Scale(Vec2{2, 2})

		if scaled.Width() != 20 {
			t.Errorf("Scaled width should be 20, got %f", scaled.Width())
		}

		emptyBB := BoundingBox{}
		if !emptyBB.Scale(Vec2{2, 2}).IsEmpty() {
			t.Error("Scaled empty bbox should remain empty")
		}
	})

	t.Run("Corners", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		corners := bb.Corners()

		if len(corners) != 4 {
			t.Errorf("Expected 4 corners, got %d", len(corners))
		}

		expected := []Vec2{
			{0, 0},
			{10, 0},
			{10, 10},
			{0, 10},
		}
		for i, c := range expected {
			if !corners[i].IsEqual(c, 1e-10) {
				t.Errorf("Corner %d should be %v, got %v", i, c, corners[i])
			}
		}

		emptyBB := BoundingBox{}
		if len(emptyBB.Corners()) != 0 {
			t.Error("Empty bbox corners should be empty slice")
		}
	})

	t.Run("Perimeter", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 5})
		if bb.Perimeter() != 30 {
			t.Errorf("Perimeter should be 30, got %f", bb.Perimeter())
		}
	})

	t.Run("AspectRatio", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 5})
		if math.Abs(bb.AspectRatio()-2) > 1e-10 {
			t.Errorf("AspectRatio should be 2, got %f", bb.AspectRatio())
		}

		zeroHeightBB := NewBoundingBox(Vec2{0, 0}, Vec2{10, 0})
		if zeroHeightBB.AspectRatio() != 0 {
			t.Error("AspectRatio with zero height should be 0")
		}
	})

	t.Run("IsSquare", func(t *testing.T) {
		squareBB := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		if !squareBB.IsSquare(1e-10) {
			t.Error("10x10 bbox should be square")
		}

		rectBB := NewBoundingBox(Vec2{0, 0}, Vec2{10, 5})
		if rectBB.IsSquare(1e-10) {
			t.Error("10x5 bbox should not be square")
		}

		if !rectBB.IsSquare(10) {
			t.Error("10x5 bbox should be square within tolerance 10")
		}

		emptyBB := BoundingBox{}
		if emptyBB.IsSquare(1e-10) {
			t.Error("Empty bbox should not be square")
		}
	})

	t.Run("LongestSide", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 5})
		if bb.LongestSide() != 10 {
			t.Errorf("LongestSide should be 10, got %f", bb.LongestSide())
		}
	})

	t.Run("ShortestSide", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 5})
		if bb.ShortestSide() != 5 {
			t.Errorf("ShortestSide should be 5, got %f", bb.ShortestSide())
		}

		emptyBB := BoundingBox{}
		if emptyBB.ShortestSide() != 0 {
			t.Error("ShortestSide of empty bbox should be 0")
		}
	})

	t.Run("String", func(t *testing.T) {
		bb := NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
		str := bb.String()
		if len(str) == 0 {
			t.Error("String should not be empty")
		}
	})
}

func TestBoundingBox3DExtended(t *testing.T) {
	t.Run("NewBoundingBox3DFromPointsEmpty", func(t *testing.T) {
		bb := NewBoundingBox3DFromPoints([]Vec3{})
		if bb.Min.X() != 0 || bb.Min.Y() != 0 || bb.Min.Z() != 0 {
			t.Error("Empty points should create zero bounding box")
		}
	})

	t.Run("Center", func(t *testing.T) {
		bb := NewBoundingBox3D(Vec3{0, 0, 0}, Vec3{10, 20, 30})
		center := bb.Center()
		if center.X() != 5 || center.Y() != 10 || center.Z() != 15 {
			t.Errorf("Center should be [5,10,15], got [%f,%f,%f]", center.X(), center.Y(), center.Z())
		}
	})

	t.Run("ContainsPoint3DEdges", func(t *testing.T) {
		bb := NewBoundingBox3D(Vec3{0, 0, 0}, Vec3{10, 10, 10})

		if !bb.ContainsPoint3D(Vec3{0, 0, 0}) {
			t.Error("Min corner should be contained")
		}
		if !bb.ContainsPoint3D(Vec3{10, 10, 10}) {
			t.Error("Max corner should be contained")
		}
		if !bb.ContainsPoint3D(Vec3{5, 5, 5}) {
			t.Error("Center should be contained")
		}
		if bb.ContainsPoint3D(Vec3{-1, 5, 5}) {
			t.Error("Point outside min X should not be contained")
		}
		if bb.ContainsPoint3D(Vec3{11, 5, 5}) {
			t.Error("Point outside max X should not be contained")
		}
	})

	t.Run("Size", func(t *testing.T) {
		bb := NewBoundingBox3D(Vec3{0, 0, 0}, Vec3{3, 4, 5})
		size := bb.Size()
		if size.X() != 3 || size.Y() != 4 || size.Z() != 5 {
			t.Errorf("Size should be [3,4,5], got [%f,%f,%f]", size.X(), size.Y(), size.Z())
		}
	})

	t.Run("Volume", func(t *testing.T) {
		bb := NewBoundingBox3D(Vec3{0, 0, 0}, Vec3{2, 3, 4})
		volume := bb.Volume()
		expected := 2.0 * 3.0 * 4.0
		if math.Abs(volume-expected) > 1e-10 {
			t.Errorf("Volume should be %f, got %f", expected, volume)
		}
	})

	t.Run("String", func(t *testing.T) {
		bb := NewBoundingBox3D(Vec3{0, 0, 0}, Vec3{10, 10, 10})
		str := bb.String()
		if len(str) == 0 {
			t.Error("String should not be empty")
		}
	})
}

func TestMatrix44Extended(t *testing.T) {
	t.Run("NewMatrix44FromSlice", func(t *testing.T) {
		slice := []float64{1, 0, 0, 0, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0, 0, 4}
		m := NewMatrix44FromSlice(slice)
		if m[0] != 1 || m[5] != 2 || m[10] != 3 || m[15] != 4 {
			t.Error("NewMatrix44FromSlice should copy values")
		}
	})

	t.Run("GetSet", func(t *testing.T) {
		m := NewMatrix44()
		m.Set(1, 1, 5)
		if m.Get(1, 1) != 5 {
			t.Errorf("Get(1,1) should be 5, got %f", m.Get(1, 1))
		}

		if m.Get(5, 5) != 0 {
			t.Error("Get out of bounds should return 0")
		}
	})

	t.Run("GetRow", func(t *testing.T) {
		m := NewMatrix44()
		m.SetRow(1, Vec4{1, 2, 3, 4})
		row := m.GetRow(1)
		if row[0] != 1 || row[1] != 2 || row[2] != 3 || row[3] != 4 {
			t.Errorf("GetRow(1) should be [1,2,3,4], got [%f,%f,%f,%f]", row[0], row[1], row[2], row[3])
		}
	})

	t.Run("GetColumn", func(t *testing.T) {
		m := NewMatrix44()
		m.SetColumn(1, Vec4{1, 2, 3, 4})
		col := m.GetColumn(1)
		if col[0] != 1 || col[1] != 2 || col[2] != 3 || col[3] != 4 {
			t.Errorf("GetColumn(1) should be [1,2,3,4], got [%f,%f,%f,%f]", col[0], col[1], col[2], col[3])
		}
	})

	t.Run("Add", func(t *testing.T) {
		m1 := NewMatrix44()
		m1[0] = 1
		m2 := NewMatrix44()
		m2[0] = 2
		sum := m1.Add(m2)
		if sum[0] != 3 {
			t.Errorf("Add result[0] should be 3, got %f", sum[0])
		}
	})

	t.Run("Sub", func(t *testing.T) {
		m1 := NewMatrix44()
		m1[0] = 5
		m2 := NewMatrix44()
		m2[0] = 2
		diff := m1.Sub(m2)
		if diff[0] != 3 {
			t.Errorf("Sub result[0] should be 3, got %f", diff[0])
		}
	})

	t.Run("MulScalar", func(t *testing.T) {
		m := NewMatrix44()
		m[0] = 2
		scaled := m.MulScalar(3)
		if scaled[0] != 6 {
			t.Errorf("MulScalar result[0] should be 6, got %f", scaled[0])
		}
	})

	t.Run("Transpose", func(t *testing.T) {
		m := NewMatrix44()
		m.Set(0, 1, 5)
		transposed := m.Transpose()
		if transposed.Get(1, 0) != 5 {
			t.Error("Transpose should swap rows and columns")
		}
	})

	t.Run("Determinant", func(t *testing.T) {
		m := NewMatrix44()
		det := m.Determinant()
		if math.Abs(det-1) > 1e-10 {
			t.Errorf("Identity matrix determinant should be 1, got %f", det)
		}
	})

	t.Run("Inverse", func(t *testing.T) {
		m := Translation(5, 10, 15)
		inv, ok := m.Inverse()
		if !ok {
			t.Error("Translation matrix should be invertible")
		}

		result := m.Mul(inv)
		if !result.IsIdentity(1e-10) {
			t.Error("Matrix * Inverse should be identity")
		}
	})

	t.Run("InverseSingular", func(t *testing.T) {
		m := NewMatrix44()
		_, ok := m.Inverse()
		if !ok {
			t.Error("Identity matrix should be invertible")
		}

		singular := Matrix44{}
		_, ok = singular.Inverse()
		if ok {
			t.Error("Zero matrix should not be invertible")
		}
	})

	t.Run("IsIdentity", func(t *testing.T) {
		if !NewMatrix44().IsIdentity(1e-10) {
			t.Error("NewMatrix44 should be identity")
		}

		m := NewMatrix44()
		m[0] = 2
		if m.IsIdentity(1e-10) {
			t.Error("Modified matrix should not be identity")
		}
		if !m.IsIdentity(1) {
			t.Error("Modified matrix should be identity within tolerance 1")
		}
	})

	t.Run("TransformVector", func(t *testing.T) {
		m := Scale(2, 3, 4)
		v := NewVec3(1, 1, 1)
		result := m.TransformVector(v)
		if result.X() != 2 || result.Y() != 3 || result.Z() != 4 {
			t.Errorf("TransformVector should scale, got [%f,%f,%f]", result.X(), result.Y(), result.Z())
		}
	})

	t.Run("IsEqual", func(t *testing.T) {
		m1 := NewMatrix44()
		m2 := NewMatrix44()
		if !m1.IsEqual(m2, 1e-10) {
			t.Error("Two identity matrices should be equal")
		}

		m3 := NewMatrix44()
		m3[0] = 1.00000001
		if !m1.IsEqual(m3, 1e-8) {
			t.Error("Matrices should be equal within tolerance 1e-8")
		}
	})

	t.Run("ToSlice", func(t *testing.T) {
		m := NewMatrix44()
		slice := m.ToSlice()
		if len(slice) != 16 {
			t.Errorf("ToSlice should return 16 elements, got %d", len(slice))
		}
	})

	t.Run("RotationX", func(t *testing.T) {
		m := RotationX(math.Pi / 2)
		v := NewVec3(0, 1, 0)
		result := m.MulVec(v)
		if math.Abs(result[1]) > 1e-10 || math.Abs(result[2]-1) > 1e-10 {
			t.Errorf("RotationX(90deg) should rotate Y to Z, got [%f,%f,%f]", result[0], result[1], result[2])
		}
	})

	t.Run("RotationY", func(t *testing.T) {
		m := RotationY(math.Pi / 2)
		v := NewVec3(1, 0, 0)
		result := m.MulVec(v)
		if math.Abs(result[0]) > 1e-10 || math.Abs(result[2]+1) > 1e-10 {
			t.Errorf("RotationY(90deg) should rotate X to -Z, got [%f,%f,%f]", result[0], result[1], result[2])
		}
	})

	t.Run("RotationZ", func(t *testing.T) {
		m := RotationZ(math.Pi / 2)
		v := NewVec3(1, 0, 0)
		result := m.MulVec(v)
		if math.Abs(result[0]) > 1e-10 || math.Abs(result[1]-1) > 1e-10 {
			t.Errorf("RotationZ(90deg) should rotate X to Y, got [%f,%f,%f]", result[0], result[1], result[2])
		}
	})

	t.Run("RotationAxis", func(t *testing.T) {
		m := RotationAxis(NewVec3(0, 0, 1), math.Pi/2)
		v := NewVec3(1, 0, 0)
		result := m.MulVec(v)
		if math.Abs(result[0]) > 1e-10 || math.Abs(result[1]-1) > 1e-10 {
			t.Errorf("RotationAxis around Z should match RotationZ, got [%f,%f,%f]", result[0], result[1], result[2])
		}
	})

	t.Run("ScaleUniform", func(t *testing.T) {
		m := ScaleUniform(5)
		v := NewVec3(1, 2, 3)
		result := m.MulVec(v)
		if result.X() != 5 || result.Y() != 10 || result.Z() != 15 {
			t.Errorf("ScaleUniform(5) should scale by 5, got [%f,%f,%f]", result.X(), result.Y(), result.Z())
		}
	})
}

func TestOCSExtended(t *testing.T) {
	t.Run("ToWCS2D", func(t *testing.T) {
		ocs := NewOCSIdentity()
		result := ocs.ToWCS2D(NewVec2(5, 10))
		if result.X() != 5 || result.Y() != 10 || result.Z() != 0 {
			t.Errorf("Identity OCS ToWCS2D should preserve XY, got [%f,%f,%f]", result.X(), result.Y(), result.Z())
		}
	})

	t.Run("ToOCS2D", func(t *testing.T) {
		ocs := NewOCSIdentity()
		result := ocs.ToOCS2D(NewVec3(5, 10, 15))
		if result.X() != 5 || result.Y() != 10 {
			t.Errorf("Identity OCS ToOCS2D should preserve XY, got [%f,%f]", result.X(), result.Y())
		}
	})

	t.Run("DirectionToWCS", func(t *testing.T) {
		ocs := NewOCSIdentity()
		dir := NewVec3(1, 0, 0)
		result := ocs.DirectionToWCS(dir)
		if result.X() != 1 || result.Y() != 0 || result.Z() != 0 {
			t.Errorf("Identity OCS DirectionToWCS should preserve direction, got [%f,%f,%f]", result.X(), result.Y(), result.Z())
		}
	})

	t.Run("DirectionToOCS", func(t *testing.T) {
		ocs := NewOCSIdentity()
		dir := NewVec3(1, 0, 0)
		result := ocs.DirectionToOCS(dir)
		if result.X() != 1 || result.Y() != 0 || result.Z() != 0 {
			t.Errorf("Identity OCS DirectionToOCS should preserve direction, got [%f,%f,%f]", result.X(), result.Y(), result.Z())
		}
	})

	t.Run("AngleToWCS", func(t *testing.T) {
		ocs := NewOCSIdentity()
		angle := ocs.AngleToWCS(math.Pi / 4)
		if math.Abs(angle-math.Pi/4) > 1e-10 {
			t.Errorf("Identity OCS AngleToWCS should preserve angle, got %f", angle)
		}
	})

	t.Run("Extrusion", func(t *testing.T) {
		ocs := NewOCSIdentity()
		extrusion := ocs.Extrusion()
		if extrusion.X() != 0 || extrusion.Y() != 0 || extrusion.Z() != 1 {
			t.Errorf("Identity OCS extrusion should be [0,0,1], got [%f,%f,%f]", extrusion.X(), extrusion.Y(), extrusion.Z())
		}
	})

	t.Run("String", func(t *testing.T) {
		ocs := NewOCSIdentity()
		str := ocs.String()
		if len(str) == 0 {
			t.Error("OCS String should not be empty")
		}
	})
}
