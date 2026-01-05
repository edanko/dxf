package entity

import (
	dxfmath "github.com/edanko/dxf/math"
	"math"
)

// Curve represents a geometric curve interface for 3D modeling
type Curve interface {
	// GetBounds returns the parameter bounds of the curve
	GetBounds() (float64, float64)

	// Evaluate returns the point on the curve at parameter t
	Evaluate(t float64) dxfmath.Vec3

	// EvaluateDerivative returns the derivative at parameter t
	EvaluateDerivative(t float64) dxfmath.Vec3

	// Length returns the approximate length of the curve
	Length() float64

	// Type returns the type of curve
	Type() string
}

// LineCurve represents a straight line curve
type LineCurve struct {
	startPoint dxfmath.Vec3
	endPoint   dxfmath.Vec3
}

// NewLineCurve creates a new line curve
func NewLineCurve(start, end dxfmath.Vec3) *LineCurve {
	return &LineCurve{
		startPoint: start,
		endPoint:   end,
	}
}

func (lc *LineCurve) GetBounds() (float64, float64) {
	return 0.0, 1.0
}

func (lc *LineCurve) Evaluate(t float64) dxfmath.Vec3 {
	return dxfmath.NewVec3(
		lc.startPoint.X()+t*(lc.endPoint.X()-lc.startPoint.X()),
		lc.startPoint.Y()+t*(lc.endPoint.Y()-lc.startPoint.Y()),
		lc.startPoint.Z()+t*(lc.endPoint.Z()-lc.startPoint.Z()),
	)
}

func (lc *LineCurve) EvaluateDerivative(t float64) dxfmath.Vec3 {
	return lc.endPoint.Sub(lc.startPoint)
}

func (lc *LineCurve) Length() float64 {
	return lc.endPoint.Sub(lc.startPoint).Length()
}

func (lc *LineCurve) Type() string {
	return "Line"
}

// ArcCurve represents an arc curve
type ArcCurve struct {
	center     dxfmath.Vec3
	radius     float64
	normal     dxfmath.Vec3
	startAngle float64
	endAngle   float64
}

// NewArcCurve creates a new arc curve
func NewArcCurve(center dxfmath.Vec3, radius float64, normal dxfmath.Vec3, startAngle, endAngle float64) *ArcCurve {
	return &ArcCurve{
		center:     center,
		radius:     radius,
		normal:     normal,
		startAngle: startAngle,
		endAngle:   endAngle,
	}
}

func (ac *ArcCurve) GetBounds() (float64, float64) {
	return ac.startAngle, ac.endAngle
}

func (ac *ArcCurve) Evaluate(t float64) dxfmath.Vec3 {
	angle := ac.startAngle + t*(ac.endAngle-ac.startAngle)

	// Create point in XY plane, then transform to plane
	x := ac.radius * math.Cos(angle)
	y := ac.radius * math.Sin(angle)
	z := 0.0

	// For now, assume XY plane and ignore normal transformation
	return dxfmath.NewVec3(
		ac.center.X()+x,
		ac.center.Y()+y,
		ac.center.Z()+z,
	)
}

func (ac *ArcCurve) EvaluateDerivative(t float64) dxfmath.Vec3 {
	angle := ac.startAngle + t*(ac.endAngle-ac.startAngle)

	// Tangent vector
	dx := -ac.radius * math.Sin(angle)
	dy := ac.radius * math.Cos(angle)

	// For now, assume XY plane
	return dxfmath.NewVec3(dx, dy, 0.0)
}

func (ac *ArcCurve) Length() float64 {
	angleDiff := ac.endAngle - ac.startAngle
	return ac.radius * math.Abs(angleDiff)
}

func (ac *ArcCurve) Type() string {
	return "Arc"
}

// CircleCurve represents a closed circle curve
type CircleCurve struct {
	center dxfmath.Vec3
	radius float64
	normal dxfmath.Vec3
}

// NewCircleCurve creates a new circle curve
func NewCircleCurve(center dxfmath.Vec3, radius float64, normal dxfmath.Vec3) *CircleCurve {
	return &CircleCurve{
		center: center,
		radius: radius,
		normal: normal,
	}
}

func (cc *CircleCurve) GetBounds() (float64, float64) {
	return 0.0, 2.0 * dxfmath.Pi
}

func (cc *CircleCurve) Evaluate(t float64) dxfmath.Vec3 {
	angle := t * 2.0 * math.Pi

	x := cc.radius * math.Cos(angle)
	y := cc.radius * math.Sin(angle)

	// For now, assume XY plane and ignore normal transformation
	return dxfmath.NewVec3(
		cc.center.X()+x,
		cc.center.Y()+y,
		cc.center.Z(),
	)
}

func (cc *CircleCurve) EvaluateDerivative(t float64) dxfmath.Vec3 {
	angle := t * 2.0 * math.Pi

	dx := -cc.radius * math.Sin(angle)
	dy := cc.radius * math.Cos(angle)

	return dxfmath.NewVec3(dx, dy, 0.0)
}

func (cc *CircleCurve) Length() float64 {
	return 2.0 * math.Pi * cc.radius
}

func (cc *CircleCurve) Type() string {
	return "Circle"
}
