package math

import (
	"fmt"
	"math"
)

// BulgeFromSagitta returns the bulge of an arc whose chord length is chord and
// whose sagitta (bulge height, measured from the chord midpoint to the arc) is
// sagitta. A positive sagitta bulges to the left of the travel direction.
//
// The relation is b = 2*s/c, which follows from s/c = tan(theta/4)/2.
func BulgeFromSagitta(sagitta, chord float64) Bulge {
	if chord == 0 {
		return 0
	}
	return Bulge(2 * sagitta / chord)
}

// SagittaFromBulge returns the sagitta of the arc described by bulge over a
// chord of the given length.
func SagittaFromBulge(bulge Bulge, chord float64) float64 {
	return float64(bulge) * chord / 2
}

// BulgeFromThreePoints returns the bulge of the arc that starts at start, ends
// at end, and passes through mid. It fails when the three points are
// collinear, in which case the connecting primitive is a straight line
// (bulge 0).
func BulgeFromThreePoints(start, end, mid Vec2) (Bulge, error) {
	chord := end.Sub(start)
	length := chord.Length()
	if length <= DefaultTolerance {
		return 0, fmt.Errorf("bulge: chord length is zero")
	}
	// Signed distance of mid from the chord line. A positive bulge puts the arc
	// to the right of the chord's travel direction, so the sagitta carries the
	// opposite sign of the chord-to-mid cross product.
	toMid := mid.Sub(start)
	sagitta := -(chord[0]*toMid[1] - chord[1]*toMid[0]) / length
	if math.Abs(sagitta) <= DefaultTolerance {
		return 0, nil // collinear: a straight segment
	}
	return BulgeFromSagitta(sagitta, length), nil
}

// ArcParameters describes a circular arc in WCS XY, the reverse form of a
// bulge segment.
type ArcParameters struct {
	Center     Vec2
	Radius     float64
	StartAngle float64 // radians, includes start
	EndAngle   float64 // radians, includes end
	Sweep      float64 // signed sweep angle, sign matches the bulge sign
	Clockwise  bool
}

// ArcParameters converts a bulge segment into a circular arc. It reports an
// error for straight segments, which have no arc.
func (b Bulge) ArcParameters(start, end Vec2) (ArcParameters, error) {
	var a ArcParameters
	if !b.IsArc() {
		return a, fmt.Errorf("bulge: straight segment has no arc parameters")
	}
	chord := end.Sub(start)
	chordLength := chord.Length()
	if chordLength <= DefaultTolerance {
		return a, fmt.Errorf("bulge: chord length is zero")
	}
	angle := b.Angle() // signed included angle, 4*atan(b)
	radius := b.Radius(chordLength)
	if math.IsInf(radius, 0) || math.IsNaN(radius) {
		return a, fmt.Errorf("bulge: degenerate arc radius")
	}
	center := b.Center(start, end)

	startAngle := math.Atan2(start[1]-center[1], start[0]-center[0])
	sweep := angle
	endAngle := startAngle + sweep
	return ArcParameters{
		Center:     center,
		Radius:     radius,
		StartAngle: normalizeAngle(startAngle),
		EndAngle:   normalizeAngle(endAngle),
		Sweep:      sweep,
		Clockwise:  b.IsClockwise(),
	}, nil
}

// NewBulgeFromArc returns the bulge of the arc described by arcParameters. A
// bulge encodes only the included angle of the arc over its chord, so the sweep
// is what determines the value; the center and radius fix where the arc sits,
// not its shape. Sweep carries signed sweeps beyond pi (major arcs); the
// normalized StartAngle/EndAngle pair alone cannot represent those, so Sweep is
// preferred and the angle difference is only the fallback.
func NewBulgeFromArc(arc ArcParameters) Bulge {
	sweep := arc.Sweep
	if sweep == 0 {
		sweep = normalizeAngle(arc.EndAngle - arc.StartAngle)
	}
	bulge := NewBulgeFromAngle(sweep)
	// A clockwise arc has a negative bulge by convention; the sweep sign
	// already encodes that, so no extra flip is needed.
	if bulge.Angle() == 0 {
		return 0
	}
	return bulge
}

// BulgeFromArcSweep returns the bulge for an arc that includes the given signed
// sweep angle over a chord of the given length.
func BulgeFromArcSweep(sweep, chord float64) Bulge {
	if chord == 0 {
		return 0
	}
	// b = tan(theta/4) and sagitta = r(1-cos(theta/2)); express through the
	// chord so the caller does not need the radius.
	half := sweep / 2
	if math.Abs(math.Sin(half)) <= DefaultTolerance {
		return 0
	}
	sagitta := (chord / (2 * math.Sin(half))) * (1 - math.Cos(half))
	return BulgeFromSagitta(sagitta, chord)
}

func normalizeAngle(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a < -math.Pi {
		a += 2 * math.Pi
	}
	return a
}
