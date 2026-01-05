package entity

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
	"math"
)

// Arrow style types
const (
	ArrowTypeNone        = 0
	ArrowTypeClosed      = 1
	ArrowTypeOpen        = 2
	ArrowTypeDot         = 3
	ArrowTypeArch        = 4
	ArrowTypeTick        = 5
	ArrowTypeOpen90      = 6
	ArrowTypeOpen180     = 7
	ArrowTypeOrigin      = 8
	ArrowTypeDotSmall    = 9
	ArrowTypeDotBlank    = 10
	ArrowTypeSmall       = 11
	ArrowTypeBox         = 12
	ArrowTypeBoxBlank    = 13
	ArrowTypeTriangle    = 14
	ArrowTypeBoxFilled   = 15
	ArrowTypeDiamond     = 16
	ArrowTypeOblique     = 17
	ArrowTypeClosedBlank = 18
)

// Arrow represents a professional arrow for dimensions and leaders
type Arrow struct {
	*entity

	// Basic properties
	Type   int     // Arrow type (see ArrowType constants)
	Size   float64 // Arrow size
	Length float64 // Arrow length (optional, calculated from size if 0)
	Width  float64 // Arrow width at base
	Color  int     // Arrow color index

	// Geometry
	Direction dxfmath.Vec3 // Arrow direction vector
	BasePoint dxfmath.Vec3 // Arrow base point (start point)
	TipPoint  dxfmath.Vec3 // Arrow tip point (calculated)

	// Advanced properties
	Filled       bool    // Whether arrow is filled
	FilledSize   float64 // Size of filled portion (for certain types)
	HollowSize   float64 // Size of hollow portion (for certain types)
	Oblique      bool    // Whether arrow is oblique (angled)
	ObliqueAngle float64 // Oblique angle in radians

	// Style properties
	Thickness float64 // Arrow line thickness
	Scale     float64 // Arrow scale factor
	Rotation  float64 // Arrow rotation (independent of direction)

	// Custom arrow data (for user-defined arrows)
	CustomPath   []dxfmath.Vec3 // Custom arrow outline path
	CustomFilled bool           // Whether custom arrow is filled
}

// NewArrow creates a new arrow entity
func NewArrow() *Arrow {
	return &Arrow{
		entity:       NewEntity(0), // Will use custom entity type
		Type:         ArrowTypeClosed,
		Size:         1.0,
		Length:       0.0, // Auto-calculate
		Width:        0.0, // Auto-calculate based on type
		Color:        0,   // ByLayer
		Direction:    dxfmath.NewVec3(1, 0, 0),
		BasePoint:    dxfmath.NewVec3(0, 0, 0),
		TipPoint:     dxfmath.NewVec3(1, 0, 0),
		Filled:       true,
		FilledSize:   1.0,
		HollowSize:   0.5,
		Oblique:      false,
		ObliqueAngle: 0.0,
		Thickness:    0.0,
		Scale:        1.0,
		Rotation:     0.0,
		CustomPath:   make([]dxfmath.Vec3, 0),
		CustomFilled: true,
	}
}

// IsEntity is for Entity interface
func (a *Arrow) IsEntity() bool {
	return true
}

// SetType sets the arrow type
func (a *Arrow) SetType(arrowType int) {
	a.Type = arrowType
	a.calculateGeometry()
}

// SetSize sets the arrow size
func (a *Arrow) SetSize(size float64) {
	a.Size = size
	a.calculateGeometry()
}

// SetColor sets arrow color (implements Entity interface)
func (a *Arrow) SetColor(color color.ColorNumber) {
	a.entity.color = color
}

// SetDirection sets the arrow direction
func (a *Arrow) SetDirection(direction dxfmath.Vec3) {
	a.Direction = direction
	a.calculateGeometry()
}

// SetBasePoint sets the arrow base point
func (a *Arrow) SetBasePoint(point dxfmath.Vec3) {
	a.BasePoint = point
	a.calculateGeometry()
}

// SetFilled sets whether arrow is filled
func (a *Arrow) SetFilled(filled bool) {
	a.Filled = filled
}

// SetThickness sets the arrow line thickness
func (a *Arrow) SetThickness(thickness float64) {
	a.Thickness = thickness
}

// calculateGeometry calculates arrow geometry based on type and properties
func (a *Arrow) calculateGeometry() {
	// Calculate tip point based on base point, direction, and size
	if a.Length == 0.0 {
		// Use size-based length calculation
		a.Length = a.Size * 2.0 // Default: 2x size
	}

	dir := a.Direction
	a.TipPoint = a.BasePoint.Add(dir.Mul(a.Length))
}

// GenerateOutline generates arrow outline path for rendering
func (a *Arrow) GenerateOutline() []dxfmath.Vec3 {
	if len(a.CustomPath) > 0 {
		return a.CustomPath
	}

	switch a.Type {
	case ArrowTypeClosed:
		return a.generateClosedArrow()
	case ArrowTypeOpen:
		return a.generateOpenArrow()
	case ArrowTypeDot:
		return a.generateDotArrow()
	case ArrowTypeTriangle:
		return a.generateTriangleArrow()
	case ArrowTypeBox:
		return a.generateBoxArrow()
	case ArrowTypeDiamond:
		return a.generateDiamondArrow()
	case ArrowTypeOblique:
		return a.generateObliqueArrow()
	default:
		return a.generateClosedArrow()
	}
}

// generateClosedArrow generates a standard closed arrow outline
func (a *Arrow) generateClosedArrow() []dxfmath.Vec3 {
	scale := a.Size

	// Calculate perpendicular vector for arrow width
	dir := a.Direction
	perp := dxfmath.NewVec3(-dir.Y(), dir.X(), 0).Normalized().Mul(scale * 0.4)

	// Arrow points
	base1 := a.BasePoint.Add(perp)
	base2 := a.BasePoint.Sub(perp)
	tip := a.TipPoint

	return []dxfmath.Vec3{base1, tip, base2, base1}
}

// generateOpenArrow generates an open arrow outline
func (a *Arrow) generateOpenArrow() []dxfmath.Vec3 {
	scale := a.Size

	// Calculate perpendicular vector for arrow width
	dir := a.Direction
	perp := dxfmath.NewVec3(-dir.Y(), dir.X(), 0).Normalized().Mul(scale * 0.4)

	// Arrow points
	base1 := a.BasePoint.Add(perp)
	base2 := a.BasePoint.Sub(perp)
	arrowTip := a.TipPoint
	v := dir.Normalized().Mul(scale * 0.7)

	// Open arrow has V-shape at tip
	return []dxfmath.Vec3{base1, arrowTip.Sub(v), arrowTip, arrowTip.Add(v), base2, base1}
}

// generateDotArrow generates a dot arrow outline
func (a *Arrow) generateDotArrow() []dxfmath.Vec3 {
	// Dot arrow is just a circle at base point
	return a.generateCircle(a.BasePoint, a.Size*0.3)
}

// generateTriangleArrow generates a triangle arrow outline
func (a *Arrow) generateTriangleArrow() []dxfmath.Vec3 {
	scale := a.Size

	// Calculate perpendicular vector for arrow width
	dir := a.Direction
	perp := dxfmath.NewVec3(-dir.Y(), dir.X(), 0).Normalized().Mul(scale * 0.5)

	// Triangle points
	base1 := a.BasePoint.Add(perp)
	base2 := a.BasePoint.Sub(perp)
	tip := a.TipPoint

	return []dxfmath.Vec3{base1, tip, base2, base1}
}

// generateBoxArrow generates a box arrow outline
func (a *Arrow) generateBoxArrow() []dxfmath.Vec3 {
	scale := a.Size

	// Calculate perpendicular vector for arrow width
	dir := a.Direction
	perp := dxfmath.NewVec3(-dir.Y(), dir.X(), 0).Normalized().Mul(scale * 0.5)

	// Box points
	base1 := a.BasePoint.Add(perp)
	base2 := a.BasePoint.Sub(perp)
	boxDepth := dir.Normalized().Mul(scale * 0.3)
	boxTop1 := base1.Add(boxDepth)
	boxTop2 := base2.Add(boxDepth)
	tip := a.TipPoint

	return []dxfmath.Vec3{base1, base2, boxTop2, boxTop1, tip, base1}
}

// generateDiamondArrow generates a diamond arrow outline
func (a *Arrow) generateDiamondArrow() []dxfmath.Vec3 {
	scale := a.Size

	// Calculate perpendicular vector for arrow width
	dir := a.Direction
	perp := dxfmath.NewVec3(-dir.Y(), dir.X(), 0).Normalized().Mul(scale * 0.4)

	// Diamond points
	midPoint := a.BasePoint.Add(dir.Normalized().Mul(scale * 0.5))
	base1 := a.BasePoint.Add(perp.Mul(0.5))
	base2 := a.BasePoint.Sub(perp.Mul(0.5))
	midTop1 := midPoint.Add(perp)
	midTop2 := midPoint.Sub(perp)
	tip := a.TipPoint

	return []dxfmath.Vec3{base1, midTop1, tip, midTop2, base2, base1}
}

// generateObliqueArrow generates an oblique arrow outline
func (a *Arrow) generateObliqueArrow() []dxfmath.Vec3 {
	scale := a.Size

	// Oblique arrow has angled sides
	dir := a.Direction
	perp := dxfmath.NewVec3(-dir.Y(), dir.X(), 0).Normalized()

	// Create angled sides
	sideOffset := perp.Mul(scale * 0.3)
	obliqueOffset := dir.Normalized().Mul(scale * 0.2).Add(perp.Mul(scale * 0.1))

	base1 := a.BasePoint.Add(sideOffset)
	base2 := a.BasePoint.Sub(sideOffset)
	tip := a.TipPoint

	return []dxfmath.Vec3{base1, tip.Add(obliqueOffset), tip, tip.Sub(obliqueOffset), base2, base1}
}

// generateCircle generates a circle outline
func (a *Arrow) generateCircle(center dxfmath.Vec3, radius float64) []dxfmath.Vec3 {
	points := make([]dxfmath.Vec3, 0)
	segments := 16

	for i := 0; i <= segments; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(segments)
		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		z := center.Z()
		points = append(points, dxfmath.NewVec3(x, y, z))
	}

	return points
}

// GetLength returns the arrow length
func (a *Arrow) GetLength() float64 {
	return a.Length
}

// GetSize returns the arrow size
func (a *Arrow) GetSize() float64 {
	return a.Size
}

// GetTipPoint returns the arrow tip point
func (a *Arrow) GetTipPoint() dxfmath.Vec3 {
	return a.TipPoint
}

// GetBasePoint returns the arrow base point
func (a *Arrow) GetBasePoint() dxfmath.Vec3 {
	return a.BasePoint
}

// Format writes arrow data to formatter
func (a *Arrow) Format(f format.Formatter) {
	a.entity.Format(f)
	f.WriteString(100, "AcDbArrow")

	// Write arrow type
	f.WriteInt(70, a.Type)

	// Write arrow size
	f.WriteFloat(40, a.Size)

	// Write arrow color
	if a.Color != 0 {
		f.WriteInt(62, a.Color)
	}

	// Write geometry
	f.WriteFloat(10, a.BasePoint.X())
	f.WriteFloat(20, a.BasePoint.Y())
	f.WriteFloat(30, a.BasePoint.Z())

	f.WriteFloat(11, a.TipPoint.X())
	f.WriteFloat(21, a.TipPoint.Y())
	f.WriteFloat(31, a.TipPoint.Z())

	// Write direction
	f.WriteFloat(50, a.Direction.X())
	f.WriteFloat(51, a.Direction.Y())
	f.WriteFloat(52, a.Direction.Z())

	// Write advanced properties
	if a.Filled {
		f.WriteInt(290, 1)
	}
	if a.Oblique {
		f.WriteInt(291, 1)
		f.WriteFloat(41, a.ObliqueAngle)
	}
}

// BBox returns bounding box of arrow
func (a *Arrow) BBox() ([]float64, []float64) {
	minX, minY, minZ := a.BasePoint.X(), a.BasePoint.Y(), a.BasePoint.Z()
	maxX, maxY, maxZ := a.TipPoint.X(), a.TipPoint.Y(), a.TipPoint.Z()

	// Include outline points for accurate bounds
	outline := a.GenerateOutline()
	for _, point := range outline {
		if point.X() < minX {
			minX = point.X()
		}
		if point.Y() < minY {
			minY = point.Y()
		}
		if point.Z() < minZ {
			minZ = point.Z()
		}
		if point.X() > maxX {
			maxX = point.X()
		}
		if point.Y() > maxY {
			maxY = point.Y()
		}
		if point.Z() > maxZ {
			maxZ = point.Z()
		}
	}

	return []float64{minX, minY, minZ}, []float64{maxX, maxY, maxZ}
}

// CreateStandardArrow creates a standard closed arrow
func CreateStandardArrow(basePoint, tipPoint dxfmath.Vec3, size float64) *Arrow {
	arrow := NewArrow()
	arrow.SetType(ArrowTypeClosed)
	arrow.SetSize(size)
	arrow.SetBasePoint(basePoint)
	arrow.SetDirection(tipPoint.Sub(basePoint))
	arrow.SetFilled(true)
	return arrow
}

// CreateOpenArrow creates an open arrow
func CreateOpenArrow(basePoint, tipPoint dxfmath.Vec3, size float64) *Arrow {
	arrow := NewArrow()
	arrow.SetType(ArrowTypeOpen)
	arrow.SetSize(size)
	arrow.SetBasePoint(basePoint)
	arrow.SetDirection(tipPoint.Sub(basePoint))
	arrow.SetFilled(false)
	return arrow
}

// CreateDotArrow creates a dot arrow
func CreateDotArrow(point dxfmath.Vec3, size float64) *Arrow {
	arrow := NewArrow()
	arrow.SetType(ArrowTypeDot)
	arrow.SetSize(size)
	arrow.SetBasePoint(point)
	arrow.SetFilled(true)
	return arrow
}

// SetColorNumber sets the arrow color using ColorNumber type
func (a *Arrow) SetColorNumber(colorNumber color.ColorNumber) {
	a.Color = int(colorNumber)
}

// CreateTriangleArrow creates a triangle arrow
func CreateTriangleArrow(basePoint, tipPoint dxfmath.Vec3, size float64) *Arrow {
	arrow := NewArrow()
	arrow.SetType(ArrowTypeTriangle)
	arrow.SetSize(size)
	arrow.SetBasePoint(basePoint)
	arrow.SetDirection(tipPoint.Sub(basePoint))
	arrow.SetFilled(true)
	return arrow
}
