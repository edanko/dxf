package math

import (
	"math"
)

// chordParameterization creates parameter vector using chord length method
func chordParameterization(points []Vec3) []float64 {
	if len(points) < 2 {
		return []float64{0.0}
	}

	params := make([]float64, len(points))
	params[0] = 0.0

	for i := 1; i < len(points); i++ {
		dist := points[i-1].Distance(points[i])
		if i == 1 {
			params[i] = dist
		} else {
			params[i] = params[i-1] + dist
		}
	}

	// Normalize to [0,1]
	total := params[len(params)-1]
	if total > 0.0 {
		for i := range params {
			params[i] /= total
		}
	}

	return params
}

// centripetalParameterization creates parameter vector using centripetal method
func centripetalParameterization(points []Vec3) []float64 {
	if len(points) < 2 {
		return []float64{0.0}
	}

	params := make([]float64, len(points))
	params[0] = 0.0

	for i := 1; i < len(points); i++ {
		dist := math.Sqrt(points[i-1].Distance(points[i]))
		if i == 1 {
			params[i] = dist
		} else {
			params[i] = params[i-1] + dist
		}
	}

	// Normalize to [0,1]
	total := params[len(params)-1]
	if total > 0.0 {
		for i := range params {
			params[i] /= total
		}
	}

	return params
}

// uniformParameterization creates uniform parameter vector
func uniformParameterization(n int) []float64 {
	if n < 1 {
		return []float64{0.0}
	}

	params := make([]float64, n)
	for i := 0; i < n; i++ {
		params[i] = float64(i) / float64(n-1)
	}

	return params
}

// knotsFromParameterization creates knot vector from parameterization
func knotsFromParameterization(params []float64, degree int) []float64 {
	n := len(params) - 1
	m := degree + 1
	numKnots := n + m

	knots := make([]float64, numKnots)

	// First m knots are 0.0
	for i := 0; i < m; i++ {
		knots[i] = 0.0
	}

	// Internal knots from parameterization
	for i := 0; i < n; i++ {
		knotIndex := m + i
		if knotIndex < numKnots {
			knots[knotIndex] = params[i]
		}
	}

	// Last m knots are 1.0
	for i := n + 1; i < numKnots; i++ {
		knots[i] = 1.0
	}

	return knots
}

// OpenUniformKnotVector creates open uniform knot vector
func generateOpenUniformKnotVector(count, order int, normalize bool) []float64 {
	if count < 1 {
		return []float64{0.0}
	}

	n := count - 1
	m := order
	numKnots := n + m

	knots := make([]float64, numKnots)

	for i := 0; i < numKnots; i++ {
		if i < m {
			knots[i] = 0.0
		} else if i >= n+m {
			knots[i] = 1.0
		} else {
			knots[i] = float64(i-m) / float64(n)
		}
	}

	if normalize && len(knots) > 0 {
		first := knots[0]
		last := knots[len(knots)-1]
		if last-first != 0.0 {
			scale := 1.0 / (last - first)
			for i := range knots {
				knots[i] = (knots[i] - first) * scale
			}
		}
	}

	return knots
}

// normalizeKnots normalizes knot vector to [0,1] range
func normalizeKnots(knots []float64) []float64 {
	if len(knots) == 0 {
		return knots
	}

	first := knots[0]
	last := knots[len(knots)-1]

	if last-first == 0.0 {
		return knots
	}

	scale := 1.0 / (last - first)
	normalized := make([]float64, len(knots))

	for i, knot := range knots {
		normalized[i] = (knot - first) * scale
	}

	return normalized
}

// nurbsArcParameters generates NURBS parameters for circular arc approximation
func nurbsArcParameters(startAngle, endAngle float64, segments int) ([]Vec3, []float64, []float64) {
	if segments < 1 {
		segments = 1
	}

	angleRange := endAngle - startAngle

	controlPoints := make([]Vec3, segments+3)
	weights := make([]float64, segments+3)
	knots := make([]float64, segments+3)

	// Generate control points for quadratic NURBS arc representation
	for i := 0; i <= segments+2; i++ {
		t := float64(i) / float64(segments+2)
		angle := startAngle + angleRange*t

		// Weight for NURBS circular arc
		weight := 1.0
		if i == 0 || i == segments+2 {
			weight = 1.0 // Endpoints have weight 1.0
		} else {
			weight = math.Cos(math.Pi * float64(i) / float64(segments+2)) // Weighted middle points
		}

		x := math.Cos(angle)
		y := math.Sin(angle)

		controlPoints[i] = NewVec3(x, y, 0)
		weights[i] = weight
		knots[i] = float64(i) / float64(segments+2)
	}

	return controlPoints, weights, knots
}

// nurbsEllipseParameters generates NURBS parameters for elliptical arc approximation
func nurbsEllipseParameters(majorAxis, minorAxis, rotation float64, segments int) ([]Vec3, []float64, []float64) {
	if segments < 1 {
		segments = 1
	}

	controlPoints := make([]Vec3, segments+3)
	weights := make([]float64, segments+3)
	knots := make([]float64, segments+3)

	// Generate control points for quadratic NURBS ellipse representation
	for i := 0; i <= segments+2; i++ {
		t := float64(i) / float64(segments+2)
		angle := 2.0 * math.Pi * t

		// Ellipse point before rotation
		x := majorAxis * math.Cos(angle)
		y := minorAxis * math.Sin(angle)

		// Apply rotation
		rotX := x*math.Cos(rotation) - y*math.Sin(rotation)
		rotY := x*math.Sin(rotation) + y*math.Cos(rotation)

		// Weight for NURBS ellipse
		weight := 1.0
		if i > 0 && i < segments+2 {
			weight = math.Cos(math.Pi * float64(i) / float64(segments+2))
		}

		controlPoints[i] = NewVec3(rotX, rotY, 0)
		weights[i] = weight
		knots[i] = float64(i) / float64(segments+2)
	}

	return controlPoints, weights, knots
}

// generateExtendedKnotVector creates extended knot vector for curve extension
func generateExtendedKnotVector(knots []float64, order int, atStart bool) []float64 {
	m := order
	numNewKnots := m * 2 // Add order knots at each end

	extended := make([]float64, len(knots)+numNewKnots)

	if atStart {
		// Add knots at beginning
		for i := 0; i < m; i++ {
			extended[i] = knots[0] - float64(m-i)
		}
		copy(extended[m:], knots)
	} else {
		// Add knots at end
		copy(extended, knots)
		last := knots[len(knots)-1]
		for i := 0; i < m; i++ {
			extended[len(knots)+i] = last + float64(i+1)
		}
	}

	return extended
}
