package math

// DefaultTolerance is the default absolute tolerance used by geometric
// comparisons and intersection routines. Callers that need a different
// precision can pass an explicit tolerance to the APIs that accept one.
const DefaultTolerance = 1e-9

// Epsilon is kept as an alias of DefaultTolerance for code that predates the
// tolerance policy. New code should prefer DefaultTolerance.
const Epsilon = DefaultTolerance

// ApproxEqual reports whether a and b differ by less than tolerance.
func ApproxEqual(a, b, tolerance float64) bool {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	return absFloat(a-b) <= tolerance
}

// ApproxZero reports whether a is within tolerance of zero.
func ApproxZero(a, tolerance float64) bool {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	return absFloat(a) <= tolerance
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
