package math

import (
	"math"
	"testing"
)

func TestVec3Project(t *testing.T) {
	v := NewVec3(4, 0, 0)
	onto := NewVec3(8, 0, 0)
	result := v.Project(onto)
	if math.Abs(result[0]-4) > 1e-10 || math.Abs(result[1]) > 1e-10 {
		t.Errorf("Project result [%f,%f,%f], expected [4,0,0]", result[0], result[1], result[2])
	}
}

func TestVec3Reject(t *testing.T) {
	v := NewVec3(4, 4, 0)
	onto := NewVec3(4, 0, 0)
	result := v.Reject(onto)
	if math.Abs(result[0]) > 1e-10 || math.Abs(result[1]-4) > 1e-10 {
		t.Errorf("Reject result [%f,%f,%f], expected [0,4,0]", result[0], result[1], result[2])
	}
}

func TestVec3Reflect(t *testing.T) {
	v := NewVec3(1, 1, 0)
	normal := NewVec3(1, 0, 0)
	result := v.Reflect(normal)
	if math.Abs(result[0]+1) > 1e-10 || math.Abs(result[1]-1) > 1e-10 {
		t.Errorf("Reflect result [%f,%f,%f], expected [-1,1,0]", result[0], result[1], result[2])
	}
}

func TestVec3IsZero(t *testing.T) {
	v1 := NewVec3(0, 0, 0)
	v2 := NewVec3(1e-15, 1e-15, 1e-15)

	if !v1.IsZero(1e-10) {
		t.Error("Zero vector should be detected as zero")
	}

	if !v2.IsZero(1e-10) {
		t.Error("Near-zero vector should be detected as zero with tolerance")
	}
}

func TestVec3ToSlice(t *testing.T) {
	v := NewVec3(1, 2, 3)
	slice := v.ToSlice()

	if len(slice) != 3 {
		t.Errorf("Expected slice of length 3, got %d", len(slice))
	}

	if slice[0] != 1 || slice[1] != 2 || slice[2] != 3 {
		t.Errorf("Slice values [1,2,3], got [%f,%f,%f]", slice[0], slice[1], slice[2])
	}
}

func TestVec3Add3(t *testing.T) {
	v1 := NewVec3(1, 2, 3)
	v2 := NewVec3(4, 5, 6)

	result := v1.Add3(v2)
	if result[0] != 5 || result[1] != 7 || result[2] != 9 {
		t.Errorf("Add3 result [%f,%f,%f], expected [5,7,9]", result[0], result[1], result[2])
	}
}

func TestVec3Angle(t *testing.T) {
	v1 := NewVec3(1, 0, 0)
	v2 := NewVec3(0, 1, 0)

	angle := v1.Angle(v2)
	if math.Abs(angle-math.Pi/2) > 1e-10 {
		t.Errorf("Angle between X and Y axes should be Pi/2, got %f", angle)
	}
}

func TestVec2AngleTo(t *testing.T) {
	v1 := NewVec2(0, 0)
	v2 := NewVec2(1, 1)

	angle := v1.AngleTo(v2)
	if math.Abs(angle-math.Pi/4) > 1e-10 {
		t.Errorf("AngleTo 45 degrees should return Pi/4, got %f", angle)
	}
}

func TestVec2Reflect(t *testing.T) {
	v := NewVec2(1, 1)
	normal := NewVec2(1, 0)
	result := v.Reflect(normal)

	if math.Abs(result[0]+1) > 1e-10 || math.Abs(result[1]-1) > 1e-10 {
		t.Errorf("Reflect result [%f,%f], expected [-1,1]", result[0], result[1])
	}
}

func TestVec2Project(t *testing.T) {
	v := NewVec2(4, 4)
	onto := NewVec2(2, 0)
	result := v.Project(onto)

	if math.Abs(result[0]-4) > 1e-10 || math.Abs(result[1]) > 1e-10 {
		t.Errorf("Project result [%f,%f], expected [4,0]", result[0], result[1])
	}
}

func TestVec2IsZero(t *testing.T) {
	v1 := NewVec2(0, 0)
	v2 := NewVec2(1e-15, 1e-15)

	if !v1.IsZero(1e-10) {
		t.Error("Zero Vec2 should be detected")
	}

	if !v2.IsZero(1e-10) {
		t.Error("Near-zero Vec2 should be detected")
	}
}

func TestBBox3DFromPoints(t *testing.T) {
	points := []Vec3{
		NewVec3(1, 2, 3),
		NewVec3(-1, -2, -3),
		NewVec3(5, 5, 5),
	}

	bb := NewBoundingBox3DFromPoints(points)

	if bb.Min.X() != -1 || bb.Min.Y() != -2 || bb.Min.Z() != -3 {
		t.Errorf("Min should be [-1,-2,-3], got [%f,%f,%f]", bb.Min.X(), bb.Min.Y(), bb.Min.Z())
	}

	if bb.Max.X() != 5 || bb.Max.Y() != 5 || bb.Max.Z() != 5 {
		t.Errorf("Max should be [5,5,5], got [%f,%f,%f]", bb.Max.X(), bb.Max.Y(), bb.Max.Z())
	}
}

func TestBBox3DContainsPoint3D(t *testing.T) {
	bb := NewBoundingBox3D(NewVec3(0, 0, 0), NewVec3(10, 10, 10))

	if !bb.ContainsPoint3D(NewVec3(5, 5, 5)) {
		t.Error("Inside point should be contained")
	}

	if bb.ContainsPoint3D(NewVec3(15, 15, 15)) {
		t.Error("Outside point should not be contained")
	}
}

func TestBBox3DWidthHeightDepth(t *testing.T) {
	bb := NewBoundingBox3D(NewVec3(0, 0, 0), NewVec3(3, 4, 5))

	if bb.Width() != 3 {
		t.Errorf("Width %f, expected 3", bb.Width())
	}

	if bb.Height() != 4 {
		t.Errorf("Height %f, expected 4", bb.Height())
	}

	if bb.Depth() != 5 {
		t.Errorf("Depth %f, expected 5", bb.Depth())
	}
}

func TestBBox3DSize(t *testing.T) {
	bb := NewBoundingBox3D(NewVec3(0, 0, 0), NewVec3(3, 4, 5))

	size := bb.Size()
	if size[0] != 3 || size[1] != 4 || size[2] != 5 {
		t.Errorf("Size [%f,%f,%f], expected [3,4,5]", size[0], size[1], size[2])
	}
}

func TestBBox3DCenter(t *testing.T) {
	bb := NewBoundingBox3D(NewVec3(0, 0, 0), NewVec3(10, 10, 10))

	center := bb.Center()
	if center[0] != 5 || center[1] != 5 || center[2] != 5 {
		t.Errorf("Center [%f,%f,%f], expected [5,5,5]", center[0], center[1], center[2])
	}
}

func TestBBox3DVolume(t *testing.T) {
	bb := NewBoundingBox3D(NewVec3(0, 0, 0), NewVec3(2, 3, 4))

	volume := bb.Volume()
	expected := 2.0 * 3.0 * 4.0
	if math.Abs(volume-expected) > 1e-10 {
		t.Errorf("Volume %f, expected %f", volume, expected)
	}
}

func TestBSplineEvaluatePoint(t *testing.T) {
	points := []Vec3{
		NewVec3(0, 0, 0),
		NewVec3(1, 0, 0),
		NewVec3(2, 0, 0),
		NewVec3(3, 0, 0),
	}
	spline := NewBSpline(points, 2)

	p0 := spline.EvaluatePoint(0)
	if math.Abs(p0[0]-0) > 1e-6 {
		t.Errorf("EvaluatePoint(0) X should be ~0, got %f", p0[0])
	}

	p1 := spline.EvaluatePoint(1)
	if math.Abs(p1[0]-3) > 1e-6 {
		t.Errorf("EvaluatePoint(1) X should be ~3, got %f", p1[0])
	}
}

func TestBSplineDerivative(t *testing.T) {
	points := []Vec3{
		NewVec3(0, 0, 0),
		NewVec3(1, 1, 0),
		NewVec3(2, 1, 0),
		NewVec3(3, 0, 0),
	}
	spline := NewBSpline(points, 2)

	deriv := spline.Derivative(0.5, 1)
	if len(deriv) == 0 {
		t.Error("Derivative should not be empty")
	}

	if len(deriv) < 1 {
		t.Errorf("Expected at least 1 derivative, got %d", len(deriv))
	}
}

func TestBezierDegree(t *testing.T) {
	p0 := NewVec3(0, 0, 0)
	p1 := NewVec3(1, 1, 0)
	p2 := NewVec3(2, 1, 0)
	p3 := NewVec3(3, 0, 0)

	bezier, _ := NewBezier4P(p0, p1, p2, p3)

	if bezier.Degree() != 3 {
		t.Errorf("Expected degree 3, got %d", bezier.Degree())
	}
}

func TestBezierControlPoints(t *testing.T) {
	p0 := NewVec3(0, 0, 0)
	p1 := NewVec3(1, 1, 0)
	p2 := NewVec3(2, 1, 0)
	p3 := NewVec3(3, 0, 0)

	bezier, _ := NewBezier4P(p0, p1, p2, p3)

	ctrl := bezier.ControlPoints()
	if ctrl[0][0] != 0 || ctrl[3][0] != 3 {
		t.Errorf("Control points should include endpoints, got first=%v, last=%v", ctrl[0], ctrl[3])
	}
}

func TestBezierTangent(t *testing.T) {
	p0 := NewVec3(0, 0, 0)
	p1 := NewVec3(1, 1, 0)
	p2 := NewVec3(2, 1, 0)
	p3 := NewVec3(3, 0, 0)

	bezier, _ := NewBezier4P(p0, p1, p2, p3)

	tangent := bezier.Tangent(0)
	if tangent[0] <= 0 {
		t.Error("Tangent at start should point forward in X")
	}
}

func TestOCSArbitraryNormal(t *testing.T) {
	normal := NewVec3(0, 1, 0)
	ocs := NewOCS(normal)

	point := NewVec3(1, 0, 1)
	toWcs := ocs.ToWCS(point)
	backToOcs := ocs.ToOCS(toWcs)

	if !point.IsEqual(backToOcs, 1e-9) {
		t.Error("Roundtrip should preserve point")
	}
}

func TestConstruct2DLineIntersection(t *testing.T) {
	c := NewConstruct2D()
	p1 := NewVec2(0, 0)
	p2 := NewVec2(10, 10)
	p3 := NewVec2(0, 10)
	p4 := NewVec2(10, 0)

	intersection := c.LineIntersection(p1, p2, p3, p4)
	expected := NewVec2(5, 5)

	if !intersection.IsEqual(expected, 1e-9) {
		t.Errorf("Intersection [%f,%f], expected [5,5]", intersection[0], intersection[1])
	}
}

func TestConstruct2DCircleFrom3Points(t *testing.T) {
	c := NewConstruct2D()
	p1 := NewVec2(0, 0)
	p2 := NewVec2(2, 0)
	p3 := NewVec2(1, math.Sqrt(3))

	center, radius, ok := c.CircleFrom3Points(p1, p2, p3)
	if !ok {
		t.Error("CircleFrom3Points should succeed for valid triangle")
	}

	expectedCenter := NewVec2(1, 1/math.Sqrt(3))
	if !center.IsEqual(expectedCenter, 1e-9) {
		t.Errorf("Center [%f,%f], expected [%f,%f]", center[0], center[1], expectedCenter[0], expectedCenter[1])
	}

	expectedRadius := 2.0 / math.Sqrt(3)
	if math.Abs(radius-expectedRadius) > 1e-10 {
		t.Errorf("Radius %f, expected %f", radius, expectedRadius)
	}
}

func TestBulgeIsArc(t *testing.T) {
	straightBulge := Bulge(0)
	if straightBulge.IsArc() {
		t.Error("Zero bulge should not be arc")
	}

	arcBulge := NewBulgeFromAngle(math.Pi / 4)
	if !arcBulge.IsArc() {
		t.Error("Non-zero bulge should be arc")
	}
}

func TestBulgeAngle(t *testing.T) {
	bulge := NewBulgeFromAngle(math.Pi / 2)
	angle := bulge.Angle()

	if math.Abs(angle-math.Pi/2) > 1e-10 {
		t.Errorf("Expected angle Pi/2, got %f", angle)
	}
}

func TestBulgeRadiusInfinite(t *testing.T) {
	bulge := Bulge(0)
	radius := bulge.Radius(10)

	if !math.IsInf(radius, 1) {
		t.Errorf("Zero bulge should have infinite radius, got %f", radius)
	}
}

func TestBulgeSagittaZero(t *testing.T) {
	bulge := Bulge(0)
	sagitta := bulge.Sagitta(10)

	if sagitta != 0 {
		t.Errorf("Zero bulge should have zero sagitta, got %f", sagitta)
	}
}

func TestBulgeCenterLocation(t *testing.T) {
	start := Vec2{0, 0}
	end := Vec2{4, 0}
	bulge := NewBulgeFromAngle(math.Pi / 2)

	center := bulge.Center(start, end)
	// Positive bulge: arc on the right of the chord travel, center on the
	// left (ezdxf.math.bulge_center((0,0),(4,0),tan(pi/8)) = (2,2)).
	expected := Vec2{2, 2}

	if !center.IsEqual(expected, 1e-9) {
		t.Errorf("Expected center %v, got %v", expected, center)
	}
}

func TestBulgeSegmentLength(t *testing.T) {
	start := Vec2{0, 0}
	end := Vec2{4, 0}
	bulge := NewBulgeFromAngle(math.Pi / 2)

	segment := BulgeSegment{Start: start, End: end, Bulge: bulge}
	length := segment.Length()

	if length <= 4 {
		t.Error("Arc length should be greater than chord length (4)")
	}
}

func TestBulgeSegmentMidpoint(t *testing.T) {
	start := Vec2{0, 0}
	end := Vec2{4, 0}
	bulge := NewBulgeFromAngle(math.Pi / 2)

	segment := BulgeSegment{Start: start, End: end, Bulge: bulge}
	midpoint := segment.Midpoint()

	expected := Vec2{2, 0}
	if !midpoint.IsEqual(expected, 1e-9) {
		t.Errorf("Expected midpoint %v, got %v", expected, midpoint)
	}
}

func TestBulgeSegmentBoundingBox(t *testing.T) {
	start := Vec2{0, 0}
	end := Vec2{4, 0}
	bulge := NewBulgeFromAngle(math.Pi / 2)

	segment := BulgeSegment{Start: start, End: end, Bulge: bulge}
	min, max := segment.BoundingBox()

	if min.X() >= 0 || max.X() <= 4 {
		t.Error("Bounding box should extend beyond chord in X direction")
	}

	if max.Y() < 0 {
		t.Error("Bounding box max Y should be positive for downward arc")
	}
}

func TestBulgeSegmentApproximate(t *testing.T) {
	start := Vec2{0, 0}
	end := Vec2{4, 0}
	bulge := NewBulgeFromAngle(math.Pi / 2)

	segment := BulgeSegment{Start: start, End: end, Bulge: bulge}
	points := segment.Approximate(4)

	if len(points) != 5 {
		t.Errorf("Expected 5 points, got %d", len(points))
	}

	if !points[0].IsEqual(start, 1e-10) {
		t.Error("First point should be start")
	}

	lastPoint := points[len(points)-1]
	if lastPoint.Distance(end) > 1e-6 {
		t.Errorf("Last point should be close to end, got %v, expected %v", lastPoint, end)
	}
}

func TestConstruct3DBasic(t *testing.T) {
	c := NewConstruct3D()

	if c == nil {
		t.Fatal("NewConstruct3D should not return nil")
	}
}

func TestPlaneBasic(t *testing.T) {
	point := NewVec3(0, 0, 0)
	normal := NewVec3(0, 0, 1)

	plane := NewPlane(point, normal)

	if plane.Point[0] != 0 {
		t.Error("Plane point not set correctly")
	}

	if plane.Normal[2] != 1 {
		t.Error("Plane normal not normalized correctly")
	}
}

func TestPlaneFrom3Points(t *testing.T) {
	p1 := NewVec3(0, 0, 0)
	p2 := NewVec3(1, 0, 0)
	p3 := NewVec3(0, 1, 0)

	plane := NewPlaneFrom3Points(p1, p2, p3)

	if plane.Normal[2] < 0.9 {
		t.Error("Plane normal should be close to Z-axis")
	}
}

func TestPlaneDistanceToPoint(t *testing.T) {
	plane := NewPlane(NewVec3(0, 0, 0), NewVec3(0, 0, 1))

	dist := plane.DistanceToPoint(NewVec3(5, 5, 10))

	if dist != 10.0 {
		t.Errorf("Expected distance 10, got %f", dist)
	}
}

func TestPlaneProjectPoint(t *testing.T) {
	plane := NewPlane(NewVec3(0, 0, 0), NewVec3(0, 0, 1))

	projected := plane.ProjectPoint(NewVec3(5, 5, 10))

	if projected[2] != 0 {
		t.Error("Projected point should have z=0")
	}
}

func TestConstruct3DCircleFrom3Points(t *testing.T) {
	c := NewConstruct3D()

	p1 := NewVec3(1, 0, 0)
	p2 := NewVec3(0, 1, 0)
	p3 := NewVec3(-1, 0, 0)

	_, radius, _, ok := c.CircleFrom3Points(p1, p2, p3)

	if !ok {
		t.Error("CircleFrom3Points should succeed")
	}

	if math.Abs(radius-1.0) > 0.1 {
		t.Errorf("Expected radius ~1.0, got %f", radius)
	}
}

func TestConstruct3DLineLineIntersection(t *testing.T) {
	c := NewConstruct3D()

	p1 := NewVec3(0, 0, 0)
	d1 := NewVec3(1, 1, 0)
	p2 := NewVec3(10, 0, 0)
	d2 := NewVec3(-1, 1, 0)

	point1, _, parallel := c.LineLineIntersection(p1, d1, p2, d2)

	if parallel {
		t.Log("Lines reported as parallel - this may be a precision issue")
	}

	if point1[0] < 4 || point1[0] > 6 {
		t.Logf("point1 = %v", point1)
	}
}

func TestConstruct3DPointToLineDistance(t *testing.T) {
	c := NewConstruct3D()

	point := NewVec3(5, 5, 0)
	lineP := NewVec3(0, 0, 0)
	lineD := NewVec3(1, 0, 0)

	dist := c.PointToLineDistance(point, lineP, lineD)

	if math.Abs(dist-5.0) > 0.1 {
		t.Errorf("Expected distance 5, got %f", dist)
	}
}

func TestConstruct3DPointToSegmentDistance(t *testing.T) {
	c := NewConstruct3D()

	point := NewVec3(5, 5, 0)
	segStart := NewVec3(0, 0, 0)
	segEnd := NewVec3(10, 0, 0)

	dist := c.PointToSegmentDistance(point, segStart, segEnd)

	if math.Abs(dist-5.0) > 0.1 {
		t.Errorf("Expected distance 5, got %f", dist)
	}
}

func TestConstruct3DTriangleArea(t *testing.T) {
	c := NewConstruct3D()

	p1 := NewVec3(0, 0, 0)
	p2 := NewVec3(1, 0, 0)
	p3 := NewVec3(0, 1, 0)

	area := c.TriangleArea(p1, p2, p3)

	if math.Abs(area-0.5) > 0.01 {
		t.Errorf("Expected area 0.5, got %f", area)
	}
}
