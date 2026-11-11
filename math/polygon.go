package math

import (
	"math"
	"slices"
)

// PolygonOrientation describes the winding direction of a polygon.
type PolygonOrientation int

const (
	// PolygonDegenerate means the polygon has no signed area.
	PolygonDegenerate PolygonOrientation = iota
	// PolygonCounterClockwise means positive signed area.
	PolygonCounterClockwise
	// PolygonClockwise means negative signed area.
	PolygonClockwise
)

// PolygonSignedArea returns the signed area of a polygon. The sign is positive
// for counter-clockwise winding.
func PolygonSignedArea(polygon []Vec2) float64 {
	if len(polygon) < 3 {
		return 0
	}
	var area float64
	for i := range polygon {
		j := (i + 1) % len(polygon)
		area += polygon[i][0]*polygon[j][1] - polygon[j][0]*polygon[i][1]
	}
	return area / 2
}

// PolygonArea returns the absolute area of a polygon regardless of winding.
func PolygonArea(polygon []Vec2) float64 {
	return math.Abs(PolygonSignedArea(polygon))
}

// PolygonCentroid returns the area-weighted centroid of a polygon. For
// degenerate polygons it returns the average vertex.
func PolygonCentroid(polygon []Vec2) Vec2 {
	if len(polygon) == 0 {
		return Vec2{}
	}
	area := PolygonSignedArea(polygon)
	if math.Abs(area) <= DefaultTolerance {
		var sum Vec2
		for _, point := range polygon {
			sum = sum.Add(point)
		}
		return sum.Div(float64(len(polygon)))
	}
	var cx, cy float64
	for i := range polygon {
		j := (i + 1) % len(polygon)
		cross := polygon[i][0]*polygon[j][1] - polygon[j][0]*polygon[i][1]
		cx += (polygon[i][0] + polygon[j][0]) * cross
		cy += (polygon[i][1] + polygon[j][1]) * cross
	}
	return Vec2{cx / (6 * area), cy / (6 * area)}
}

// PolygonWinding reports the winding direction of a polygon.
func PolygonWinding(polygon []Vec2) PolygonOrientation {
	area := PolygonSignedArea(polygon)
	if math.Abs(area) <= DefaultTolerance {
		return PolygonDegenerate
	}
	if area > 0 {
		return PolygonCounterClockwise
	}
	return PolygonClockwise
}

// PointInPolygon reports whether point lies inside a simple polygon using the
// even-odd ray casting rule. Points on an edge are treated as inside.
func PointInPolygon(point Vec2, polygon []Vec2) bool {
	if len(polygon) < 3 {
		return false
	}
	inside := false
	for i, j := 0, len(polygon)-1; i < len(polygon); j, i = i, i+1 {
		pi := polygon[i]
		pj := polygon[j]
		if pointOnSegment(point, pj, pi, DefaultTolerance) {
			return true
		}
		intersects := (pi[1] > point[1]) != (pj[1] > point[1])
		if intersects {
			x := (pj[0]-pi[0])*(point[1]-pi[1])/(pj[1]-pi[1]) + pi[0]
			if point[0] < x {
				inside = !inside
			}
		}
	}
	return inside
}

// ConvexHull2D returns the convex hull of points in counter-clockwise order
// using Andrew's monotone chain algorithm.
func ConvexHull2D(points []Vec2) []Vec2 {
	if len(points) < 3 {
		out := make([]Vec2, len(points))
		copy(out, points)
		return out
	}
	sorted := make([]Vec2, len(points))
	copy(sorted, points)
	sortVec2(sorted)

	lower := make([]Vec2, 0, len(sorted))
	for _, point := range sorted {
		for len(lower) >= 2 && crossVec2(lower[len(lower)-2], lower[len(lower)-1], point) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, point)
	}
	upper := make([]Vec2, 0, len(sorted))
	for _, point := range slices.Backward(sorted) {

		for len(upper) >= 2 && crossVec2(upper[len(upper)-2], upper[len(upper)-1], point) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, point)
	}
	lower = lower[:len(lower)-1]
	upper = upper[:len(upper)-1]
	return append(lower, upper...)
}

// ClipPolygonToRect clips a polygon against an axis-aligned rectangle using
// Sutherland-Hodgman. The result may be empty.
func ClipPolygonToRect(polygon []Vec2, rect BoundingBox) []Vec2 {
	clip := []Vec2{
		{rect.Min.X(), rect.Min.Y()},
		{rect.Max.X(), rect.Min.Y()},
		{rect.Max.X(), rect.Max.Y()},
		{rect.Min.X(), rect.Max.Y()},
	}
	return ClipPolygonToConvex(polygon, clip)
}

// ClipPolygonToConvex clips subject against a convex clip polygon. The clip
// polygon must be counter-clockwise; it is reversed automatically otherwise.
func ClipPolygonToConvex(subject, clip []Vec2) []Vec2 {
	if len(subject) < 3 || len(clip) < 3 {
		return nil
	}
	if PolygonWinding(clip) == PolygonClockwise {
		reversed := make([]Vec2, len(clip))
		for i := range clip {
			reversed[i] = clip[len(clip)-1-i]
		}
		clip = reversed
	}
	output := make([]Vec2, len(subject))
	copy(output, subject)
	for i := range clip {
		if len(output) == 0 {
			return nil
		}
		a := clip[i]
		b := clip[(i+1)%len(clip)]
		input := output
		output = make([]Vec2, 0, len(input))
		for j := range input {
			current := input[j]
			previous := input[(j+len(input)-1)%len(input)]
			currentInside := crossVec2(a, b, current) >= -DefaultTolerance
			previousInside := crossVec2(a, b, previous) >= -DefaultTolerance
			if currentInside {
				if !previousInside {
					if point, ok := lineSegmentIntersection(previous, current, a, b); ok {
						output = append(output, point)
					}
				}
				output = append(output, current)
			} else if previousInside {
				if point, ok := lineSegmentIntersection(previous, current, a, b); ok {
					output = append(output, point)
				}
			}
		}
	}
	return output
}

func pointOnSegment(point, a, b Vec2, tolerance float64) bool {
	ab := b.Sub(a)
	ap := point.Sub(a)
	cross := ab[0]*ap[1] - ab[1]*ap[0]
	if math.Abs(cross) > tolerance*maxFloat(ab.Length(), 1) {
		return false
	}
	dot := ap.Dot(ab)
	if dot < -tolerance {
		return false
	}
	return dot <= ab.LengthSquared()+tolerance
}

func crossVec2(a, b, c Vec2) float64 {
	return (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
}

func lineSegmentIntersection(a0, a1, b0, b1 Vec2) (Vec2, bool) {
	d1 := a1.Sub(a0)
	d2 := b1.Sub(b0)
	denom := d1[0]*d2[1] - d1[1]*d2[0]
	if math.Abs(denom) <= DefaultTolerance {
		return Vec2{}, false
	}
	diff := b0.Sub(a0)
	t := (diff[0]*d2[1] - diff[1]*d2[0]) / denom
	return a0.Add(d1.Mul(t)), true
}

func sortVec2(points []Vec2) {
	for i := 1; i < len(points); i++ {
		for j := i; j > 0; j-- {
			if points[j][0] < points[j-1][0] || (points[j][0] == points[j-1][0] && points[j][1] < points[j-1][1]) {
				points[j], points[j-1] = points[j-1], points[j]
				continue
			}
			break
		}
	}
}
