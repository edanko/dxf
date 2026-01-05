package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// HELIX entity handedness
const (
	HelixHandednessLeft  = 0 // Left-handed helix
	HelixHandednessRight = 1 // Right-handed helix
)

// HELIX constrain types
const (
	HelixConstrainTurns      = 0 // Constrain turn height
	HelixConstrainTurnHeight = 1 // Constrain turn height
	HelixConstrainBoth       = 2 // Constrain both turn height and height
)

// HELIX represents a DXF HELIX entity (3D spiral/helix)
type Helix struct {
	*entity

	// Basic properties
	startPoint dxfmath.Vec3 // 10,20,30 - Helix start point
	axisVector dxfmath.Vec3 // 11,21,31 - Helix axis vector
	radius     float64      // 40 - Helix radius
	turns      float64      // 90 - Number of turns
	turnHeight float64      // 41 - Turn height

	// Advanced properties
	handedness    int     // 290 - Handedness (0=left, 1=right)
	constrainType int     // 280 - Constrain type
	baseRadius    float64 // 70 - Base radius (if constrained)
	baseHeight    float64 // 71 - Base height (if constrained)
}

// NewHelix creates a new HELIX entity
func NewHelix() *Helix {
	return &Helix{
		entity:        NewEntity(HELIX),
		startPoint:    dxfmath.NewVec3(0, 0, 0),
		axisVector:    dxfmath.NewVec3(0, 0, 1),
		radius:        1.0,
		turns:         1.0,
		turnHeight:    0.0,
		handedness:    HelixHandednessRight,
		constrainType: HelixConstrainTurns,
		baseRadius:    0.0,
		baseHeight:    0.0,
	}
}

// IsEntity returns true for HELIX entities
func (h *Helix) IsEntity() bool {
	return true
}

// Format writes HELIX entity data to DXF format
func (h *Helix) Format(f format.Formatter) {
	h.entity.Format(f)
	f.WriteString(100, "AcDbHelix")

	// Write start point (10,20,30)
	f.WriteFloat(10, h.startPoint.X())
	f.WriteFloat(20, h.startPoint.Y())
	f.WriteFloat(30, h.startPoint.Z())

	// Write axis vector (11,21,31)
	f.WriteFloat(11, h.axisVector.X())
	f.WriteFloat(21, h.axisVector.Y())
	f.WriteFloat(31, h.axisVector.Z())

	// Write radius (40)
	f.WriteFloat(40, h.radius)

	// Write turns (90)
	f.WriteFloat(90, h.turns)

	// Write turn height (41)
	f.WriteFloat(41, h.turnHeight)

	// Write handedness (290)
	if h.handedness != HelixHandednessRight {
		f.WriteInt(290, h.handedness)
	}

	// Write constrain type (280)
	if h.constrainType != HelixConstrainTurns {
		f.WriteInt(280, h.constrainType)
	}

	// Write base radius (70)
	if h.baseRadius != 0.0 {
		f.WriteFloat(70, h.baseRadius)
	}

	// Write base height (71)
	if h.baseHeight != 0.0 {
		f.WriteFloat(71, h.baseHeight)
	}
}

// SetStartPoint sets the helix start point
func (h *Helix) SetStartPoint(point dxfmath.Vec3) {
	h.startPoint = point
}

// SetAxisVector sets the helix axis vector
func (h *Helix) SetAxisVector(vector dxfmath.Vec3) {
	h.axisVector = vector
}

// SetRadius sets the helix radius
func (h *Helix) SetRadius(radius float64) {
	h.radius = radius
}

// SetTurns sets the number of turns
func (h *Helix) SetTurns(turns float64) {
	h.turns = turns
}

// SetTurnHeight sets the turn height
func (h *Helix) SetTurnHeight(height float64) {
	h.turnHeight = height
}

// SetHandedness sets the helix handedness
func (h *Helix) SetHandedness(handedness int) {
	h.handedness = handedness
}

// SetConstrainType sets the constraint type
func (h *Helix) SetConstrainType(constrainType int) {
	h.constrainType = constrainType
}

// SetBaseRadius sets the base radius
func (h *Helix) SetBaseRadius(radius float64) {
	h.baseRadius = radius
}

// SetBaseHeight sets the base height
func (h *Helix) SetBaseHeight(height float64) {
	h.baseHeight = height
}

// StartPoint returns the helix start point
func (h *Helix) StartPoint() dxfmath.Vec3 {
	return h.startPoint
}

// AxisVector returns the helix axis vector
func (h *Helix) AxisVector() dxfmath.Vec3 {
	return h.axisVector
}

// Radius returns the helix radius
func (h *Helix) Radius() float64 {
	return h.radius
}

// Turns returns the number of turns
func (h *Helix) Turns() float64 {
	return h.turns
}

// TurnHeight returns the turn height
func (h *Helix) TurnHeight() float64 {
	return h.turnHeight
}

// Handedness returns the helix handedness
func (h *Helix) Handedness() int {
	return h.handedness
}

// ConstrainType returns the constraint type
func (h *Helix) ConstrainType() int {
	return h.constrainType
}

// BaseRadius returns the base radius
func (h *Helix) BaseRadius() float64 {
	return h.baseRadius
}

// BaseHeight returns the base height
func (h *Helix) BaseHeight() float64 {
	return h.baseHeight
}

// EndPoint calculates the end point of the helix
func (h *Helix) EndPoint() dxfmath.Vec3 {
	totalHeight := h.TotalHeight()
	return h.startPoint.Add(h.axisVector.Mul(totalHeight))
}

// TotalHeight calculates the total height of the helix
func (h *Helix) TotalHeight() float64 {
	return h.turns * h.turnHeight
}

// TotalLength calculates the total length of the helix
func (h *Helix) TotalLength() float64 {
	if h.turns == 0 || h.radius == 0 {
		return 0.0
	}

	// Length of one turn: sqrt((2πr)² + turnHeight²)
	oneTurnLength := math.Sqrt(math.Pow(2*math.Pi*h.radius, 2) + math.Pow(h.turnHeight, 2))

	return h.turns * oneTurnLength
}

// BBox calculates bounding box of the helix
func (h *Helix) BBox() ([]float64, []float64) {
	if h.turns == 0 || h.radius == 0 {
		return []float64{}, []float64{}
	}

	// Calculate helix path endpoints
	totalHeight := h.TotalHeight()
	endPoint := h.startPoint.Add(h.axisVector.Mul(totalHeight))

	// Calculate bounding box
	minX := math.Min(h.startPoint.X(), endPoint.X())
	maxX := math.Max(h.startPoint.X(), endPoint.X())
	minY := math.Min(h.startPoint.Y(), endPoint.Y())
	maxY := math.Max(h.startPoint.Y(), endPoint.Y())
	minZ := math.Min(h.startPoint.Z(), endPoint.Z())
	maxZ := math.Max(h.startPoint.Z(), endPoint.Z())

	// Add radius to bounding box
	minX -= h.radius
	maxX += h.radius
	minY -= h.radius
	maxY += h.radius
	minZ -= h.radius
	maxZ += h.radius

	return []float64{minX, minY, minZ}, []float64{maxX, maxY, maxZ}
}

// IsLeftHanded returns true if helix is left-handed
func (h *Helix) IsLeftHanded() bool {
	return h.handedness == HelixHandednessLeft
}

// IsRightHanded returns true if helix is right-handed
func (h *Helix) IsRightHanded() bool {
	return h.handedness == HelixHandednessRight
}

// CreateSpringHelix creates a simple spring helix
func CreateSpringHelix(startPoint dxfmath.Vec3, radius, turns, turnHeight float64, handedness int) *Helix {
	helix := NewHelix()
	helix.SetStartPoint(startPoint)
	helix.SetRadius(radius)
	helix.SetTurns(turns)
	helix.SetTurnHeight(turnHeight)
	helix.SetHandedness(handedness)
	return helix
}

// CreateTaperedHelix creates a helix with varying radius
func CreateTaperedHelix(startPoint dxfmath.Vec3, startRadius, endRadius, turns, turnHeight float64, handedness int) *Helix {
	helix := NewHelix()
	helix.SetStartPoint(startPoint)
	helix.SetRadius(startRadius) // Will be overridden during generation
	helix.SetTurns(turns)
	helix.SetTurnHeight(turnHeight)
	helix.SetHandedness(handedness)
	helix.SetConstrainType(HelixConstrainBoth) // Allow base radius and height
	helix.SetBaseRadius(startRadius)           // Set base radius constraint
	helix.SetBaseHeight(endRadius)             // Set base height constraint (creates taper)
	return helix
}

// CreateConstrainedHelix creates a helix with constraints
func CreateConstrainedHelix(startPoint dxfmath.Vec3, axisVector dxfmath.Vec3, radius, turns, turnHeight float64, handedness int) *Helix {
	helix := NewHelix()
	helix.SetStartPoint(startPoint)
	helix.SetAxisVector(axisVector)
	helix.SetRadius(radius)
	helix.SetTurns(turns)
	helix.SetTurnHeight(turnHeight)
	helix.SetHandedness(handedness)
	helix.SetConstrainType(HelixConstrainBoth) // Apply constraints
	helix.SetBaseRadius(radius)                // Constrain base radius
	helix.SetBaseHeight(turnHeight)            // Constrain base height
	return helix
}
