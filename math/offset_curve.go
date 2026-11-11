package math

import (
	"fmt"
	"math"
)

// JoinStyle selects how an offset curve joins consecutive segments.
type JoinStyle int

const (
	// JoinMiter extends the offset segments until they meet.
	JoinMiter JoinStyle = iota
	// JoinRound connects offset segments with a circular arc.
	JoinRound
	// JoinBevel connects offset segments with a straight line.
	JoinBevel
)

// OffsetPolyline offsets an open polyline by a signed distance. Positive
// distances offset to the left of the travel direction.
func OffsetPolyline(points []Vec2, distance float64, join JoinStyle) ([]Vec2, error) {
	if len(points) < 2 {
		return nil, fmt.Errorf("need at least 2 points for polyline offset")
	}
	if math.Abs(distance) <= DefaultTolerance {
		out := make([]Vec2, len(points))
		copy(out, points)
		return out, nil
	}
	// Flattening curves emits repeated samples at subdivision boundaries;
	// zero-length segments would corrupt normals and joins.
	points = removeConsecutiveDuplicates(points, DefaultTolerance)
	if len(points) < 2 {
		return nil, fmt.Errorf("polyline offset: degenerate polyline")
	}

	normals := make([]Vec2, len(points)-1)
	offsets := make([]Vec2, len(points)-1)
	for i := 0; i+1 < len(points); i++ {
		dir := points[i+1].Sub(points[i]).Normalized()
		normals[i] = NewVec2(-dir[1], dir[0])
		offsets[i] = normals[i].Mul(distance)
	}

	result := make([]Vec2, 0, len(points)*2)
	result = append(result, points[0].Add(offsets[0]))

	for i := 1; i+1 < len(points); i++ {
		prev := offsets[i-1]
		next := offsets[i]
		start := points[i].Add(prev)
		end := points[i].Add(next)

		// Deflections below ~1.4 mrad carry no join: the flattened curve
		// deflects by fractions of a degree at every sample, and a full
		// arc per sample would explode into a petal fan.
		if prev.Normalized().Dot(next.Normalized()) > 1-1e-6 {
			result = append(result, start)
			continue
		}

		switch join {
		case JoinRound:
			result = append(result, start)
			result = append(result, roundJoin(points[i], prev, next)...)
			result = append(result, end)
		case JoinBevel:
			result = append(result, start, end)
		default:
			miter, ok := miterJoin(points[i], prev, next, distance)
			if ok {
				result = append(result, start, miter, end)
			} else {
				result = append(result, start, end)
			}
		}
	}

	last := len(points) - 1
	result = append(result, points[last].Add(offsets[last-1]))
	return removeConsecutiveDuplicates(result, DefaultTolerance), nil
}

// OffsetBezier2D offsets a Bezier curve by flattening it to tolerance and
// offsetting the resulting polyline.
func OffsetBezier2D(bezier *Bezier, distance, tolerance float64, join JoinStyle) ([]Vec2, error) {
	if bezier == nil {
		return nil, fmt.Errorf("nil bezier")
	}
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	samples := bezier.Flattening(tolerance, 64)
	return OffsetPolyline(samplesXY(samples), distance, join)
}

// OffsetSpline2D offsets a B-spline by flattening it to tolerance.
func OffsetSpline2D(spline *BSpline, distance, tolerance float64, join JoinStyle) ([]Vec2, error) {
	if spline == nil {
		return nil, fmt.Errorf("nil spline")
	}
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	samples := spline.Flattening(tolerance)
	return OffsetPolyline(samplesXY(samples), distance, join)
}

func miterJoin(vertex, prevOffset, nextOffset Vec2, distance float64) (Vec2, bool) {
	a := vertex.Add(prevOffset)
	b := vertex.Add(nextOffset)
	d1 := prevOffset.Perpendicular()
	d2 := nextOffset.Perpendicular()
	denom := d1[0]*d2[1] - d1[1]*d2[0]
	if math.Abs(denom) <= DefaultTolerance {
		return Vec2{}, false
	}
	diff := b.Sub(a)
	t := (diff[0]*d2[1] - diff[1]*d2[0]) / denom
	point := a.Add(d1.Mul(t))
	limit := math.Abs(distance) * 4
	if math.Abs(distance) <= DefaultTolerance {
		limit = 4
	}
	if point.Distance(vertex) > limit {
		return Vec2{}, false
	}
	return point, true
}

func roundJoin(vertex, prevOffset, nextOffset Vec2) []Vec2 {
	startAngle := math.Atan2(prevOffset[1], prevOffset[0])
	endAngle := math.Atan2(nextOffset[1], nextOffset[0])
	delta := endAngle - startAngle
	for delta > math.Pi {
		delta -= 2 * math.Pi
	}
	for delta < -math.Pi {
		delta += 2 * math.Pi
	}
	const segments = 8
	radius := prevOffset.Length()
	points := make([]Vec2, 0, segments-1)
	for i := 1; i < segments; i++ {
		angle := startAngle + delta*float64(i)/segments
		points = append(points, Vec2{
			vertex[0] + radius*math.Cos(angle),
			vertex[1] + radius*math.Sin(angle),
		})
	}
	return points
}

func samplesXY(samples []Vec3) []Vec2 {
	out := make([]Vec2, len(samples))
	for i, sample := range samples {
		out[i] = sample.XY()
	}
	return out
}

func removeConsecutiveDuplicates(points []Vec2, tolerance float64) []Vec2 {
	if len(points) == 0 {
		return points
	}
	out := make([]Vec2, 0, len(points))
	out = append(out, points[0])
	for _, point := range points[1:] {
		if point.Distance(out[len(out)-1]) > tolerance {
			out = append(out, point)
		}
	}
	return out
}
