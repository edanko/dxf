package render

import (
	"math"
)

// ArrowType represents different arrow styles for dimensions
type ArrowType int

const (
	ArrowNone ArrowType = iota
	ArrowClosedBlank
	ArrowClosedFilled
	ArrowOpen
	ArrowOpen30
	ArrowOpen90
	ArrowCircle
	ArrowDot
	ArrowArchTick
	ArrowOblique
	ArrowBox
	ArrowBoxFilled
	ArrowDiamond
	ArrowDotSmall
)

// Arrow represents a dimension arrow with geometry and rendering capabilities
type Arrow struct {
	ArrowType ArrowType
	Size      float64
	Angle     float64
	Vertices  []Vec3
}

// Vec3 represents a 3D point (alias for external packages that might need this)
type Vec3 struct {
	X, Y, Z float64
}

// NewVec3 creates a new 3D point
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// Add adds two vectors
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

// Mul multiplies vector by scalar
func (v Vec3) Mul(scalar float64) Vec3 {
	return Vec3{X: v.X * scalar, Y: v.Y * scalar, Z: v.Z * scalar}
}

// NewArrow creates a new arrow with specified type and size
func NewArrow(arrowType ArrowType, size float64, angle float64) *Arrow {
	vertices := generateArrowVertices(arrowType, size)

	return &Arrow{
		ArrowType: arrowType,
		Size:      size,
		Angle:     angle,
		Vertices:  vertices,
	}
}

// generateArrowVertices generates geometry for different arrow types
func generateArrowVertices(arrowType ArrowType, size float64) []Vec3 {
	switch arrowType {
	case ArrowNone:
		return nil
	case ArrowClosedBlank:
		return generateClosedArrowBlank(size)
	case ArrowClosedFilled:
		return generateClosedArrowFilled(size)
	case ArrowOpen:
		return generateOpenArrow(size, 30.0)
	case ArrowOpen30:
		return generateOpenArrow(size, 30.0)
	case ArrowOpen90:
		return generateOpenArrow(size, 90.0)
	case ArrowCircle:
		return generateCircleArrow(size)
	case ArrowDot:
		return generateDotArrow(size)
	case ArrowArchTick:
		return generateArchTick(size)
	case ArrowOblique:
		return generateObliqueArrow(size)
	default:
		return generateOpenArrow(size, 30.0) // Default
	}
}

// generateClosedArrowBlank generates a closed blank arrow (triangle outline)
func generateClosedArrowBlank(size float64) []Vec3 {
	return []Vec3{
		{X: -size, Y: 0, Z: 0},
		{X: size, Y: 0, Z: 0},
	}
}

// generateClosedArrowFilled generates a filled closed arrow (filled triangle)
func generateClosedArrowFilled(size float64) []Vec3 {
	return []Vec3{
		{X: -size, Y: 0, Z: 0},
		{X: size, Y: 0, Z: 0},
	}
}

// generateOpenArrow generates an open arrow
func generateOpenArrow(size float64, angle float64) []Vec3 {
	h := math.Sin(angle*math.Pi/180.0) * size
	return []Vec3{
		{X: -size, Y: h, Z: 0},
		{X: 0, Y: 0, Z: 0},
		{X: -size, Y: -h, Z: 0},
	}
}

// generateCircleArrow generates a circle arrow
func generateCircleArrow(size float64) []Vec3 {
	return []Vec3{
		{X: 0, Y: 0, Z: 0}, // Center
	}
}

// generateDotArrow generates a dot arrow
func generateDotArrow(size float64) []Vec3 {
	return []Vec3{
		{X: 0, Y: 0, Z: 0}, // Center
	}
}

// generateArchTick generates an architectural tick arrow
func generateArchTick(size float64) []Vec3 {
	width := size * 0.15
	return []Vec3{
		{X: -width, Y: 0, Z: 0},
		{X: width, Y: 0, Z: 0},
	}
}

// generateObliqueArrow generates an oblique stroke arrow
func generateObliqueArrow(size float64) []Vec3 {
	s2 := size / 2.0
	return []Vec3{
		{X: -s2, Y: -s2, Z: 0},
		{X: s2, Y: s2, Z: 0},
	}
}

// Transform applies rotation to arrow vertices
func (a *Arrow) Rotate(angle float64) {
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	for i := range a.Vertices {
		v := a.Vertices[i]
		x := v.X*cos - v.Y*sin
		y := v.X*sin + v.Y*cos
		a.Vertices[i] = Vec3{X: x, Y: y, Z: v.Z}
	}
}

// Translate translates arrow by specified vector
func (a *Arrow) Translate(vec Vec3) {
	for i := range a.Vertices {
		a.Vertices[i] = a.Vertices[i].Add(vec)
	}
}

// GetArrowTypeName returns the string name of an arrow type
func GetArrowTypeName(arrowType ArrowType) string {
	switch arrowType {
	case ArrowNone:
		return "NONE"
	case ArrowClosedBlank:
		return "CLOSED_BLANK"
	case ArrowClosedFilled:
		return "CLOSED_FILLED"
	case ArrowOpen:
		return "OPEN"
	case ArrowOpen30:
		return "OPEN_30"
	case ArrowOpen90:
		return "OPEN_90"
	case ArrowCircle:
		return "CIRCLE"
	case ArrowDot:
		return "DOT"
	case ArrowArchTick:
		return "ARCH_TICK"
	case ArrowOblique:
		return "OBLIQUE"
	default:
		return "UNKNOWN"
	}
}
