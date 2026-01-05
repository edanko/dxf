package entity

import (
	"errors"
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// XLine represents a DXF XLINE entity (bi-infinite construction line)
// An XLINE extends infinitely in both directions from a start point
type XLine struct {
	*entity
	start      dxfmath.Vec3 // Start point in WCS
	unitVector dxfmath.Vec3 // Unit direction vector (normalized)
}

// NewXLine creates a new XLINE entity
func NewXLine() *XLine {
	return &XLine{
		entity:     NewEntity(XLINE),
		start:      dxfmath.NewVec3(0, 0, 0),
		unitVector: dxfmath.NewVec3(0, 0, 1), // Default: Z-axis
	}
}

// NewXLineFromPointDirection creates an XLINE from a point and direction vector
func NewXLineFromPointDirection(start, direction dxfmath.Vec3) *XLine {
	// Normalize direction vector
	unitVec := direction.Normalized()
	if unitVec.IsEqual(dxfmath.NewVec3(0, 0, 0), 1e-9) {
		// Fallback to Z-axis if direction is zero
		unitVec = dxfmath.NewVec3(0, 0, 1)
	}

	return &XLine{
		entity:     NewEntity(XLINE),
		start:      start,
		unitVector: unitVec,
	}
}

// NewXLineFromPointDirectionValid creates an XLINE with validation
func NewXLineFromPointDirectionValid(start, direction dxfmath.Vec3) (*XLine, error) {
	// Check if direction vector is zero
	if direction.IsEqual(dxfmath.NewVec3(0, 0, 0), 1e-9) {
		return nil, errors.New("direction vector cannot be zero")
	}

	unitVec := direction.Normalized()
	return &XLine{
		entity:     NewEntity(XLINE),
		start:      start,
		unitVector: unitVec,
	}, nil
}

// IsEntity returns true for XLINE entities
func (x *XLine) IsEntity() bool {
	return true
}

// Format writes XLINE entity data to DXF format
func (x *XLine) Format(f format.Formatter) {
	x.entity.Format(f)
	f.WriteString(100, "AcDbXline")

	// Write start point (10,20,30)
	f.WriteFloat(10, x.start.X())
	f.WriteFloat(20, x.start.Y())
	f.WriteFloat(30, x.start.Z())

	// Write unit direction vector (11,21,31)
	f.WriteFloat(11, x.unitVector.X())
	f.WriteFloat(21, x.unitVector.Y())
	f.WriteFloat(31, x.unitVector.Z())
}

// SetStart sets the start point
func (x *XLine) SetStart(start dxfmath.Vec3) {
	x.start = start
}

// SetUnitVector sets the unit direction vector (will be normalized)
func (x *XLine) SetUnitVector(vector dxfmath.Vec3) {
	normalized := vector.Normalized()
	if !normalized.IsEqual(dxfmath.NewVec3(0, 0, 0), 1e-9) {
		x.unitVector = normalized
	}
	// Keep existing vector if normalization fails
}

// Start returns the start point
func (x *XLine) Start() dxfmath.Vec3 {
	return x.start
}

// UnitVector returns the unit direction vector
func (x *XLine) UnitVector() dxfmath.Vec3 {
	return x.unitVector
}

// BBox returns a bounding box for XLINE
// For infinite lines, this returns a large bounding box based on start point and direction
func (x *XLine) BBox() ([]float64, []float64) {
	// For infinite lines, return a large bounding box centered around start point
	// Use a reasonable size for display purposes
	bounds := 1000.0

	// Calculate points far along the line
	point1 := x.start.Add(x.unitVector.Mul(bounds))
	point2 := x.start.Add(x.unitVector.Mul(-bounds))

	// Calculate min and max coordinates
	minX := math.Min(point1.X(), point2.X())
	minY := math.Min(point1.Y(), point2.Y())
	minZ := math.Min(point1.Z(), point2.Z())
	maxX := math.Max(point1.X(), point2.X())
	maxY := math.Max(point1.Y(), point2.Y())
	maxZ := math.Max(point1.Z(), point2.Z())

	mins := []float64{minX, minY, minZ}
	maxs := []float64{maxX, maxY, maxZ}

	return mins, maxs
}

// Transform applies a transformation matrix to the XLINE
func (x *XLine) Transform(matrix dxfmath.Matrix44) {
	// Transform start point
	x.start = matrix.TransformVector(x.start)

	// Transform direction vector and normalize
	transformedDir := matrix.TransformVector(x.unitVector)
	normalized := transformedDir.Normalized()
	if !normalized.IsEqual(dxfmath.NewVec3(0, 0, 0), 1e-9) {
		x.unitVector = normalized
	}
}

// Translate moves the XLINE by the specified offset
func (x *XLine) Translate(dx, dy, dz float64) {
	offset := dxfmath.NewVec3(dx, dy, dz)
	x.start = x.start.Add(offset)
}

// GetPointsForDrawing returns points for drawing the infinite line within a bounds
// This is useful for rendering infinite line as a finite segment
func (x *XLine) GetPointsForDrawing(infiniteLineLength float64) (dxfmath.Vec3, dxfmath.Vec3) {
	halfLength := infiniteLineLength / 2.0
	point1 := x.start.Add(x.unitVector.Mul(-halfLength))
	point2 := x.start.Add(x.unitVector.Mul(halfLength))
	return point1, point2
}

// DistanceToPoint calculates the shortest distance from XLINE to a point
func (x *XLine) DistanceToPoint(point dxfmath.Vec3) float64 {
	// Vector from start point to target point
	toPoint := point.Sub(x.start)

	// Project onto direction vector
	projectionLength := toPoint.Dot(x.unitVector)

	// Find closest point on line
	closestPoint := x.start.Add(x.unitVector.Mul(projectionLength))

	// Return distance
	return point.Sub(closestPoint).Length()
}

// IsParallelTo checks if this XLINE is parallel to another XLINE
func (x *XLine) IsParallelTo(other *XLine) bool {
	// Two lines are parallel if their direction vectors are parallel
	// Check if cross product is close to zero
	crossProduct := x.unitVector.Cross(other.unitVector)
	return crossProduct.Length() < 1e-9
}

// IsParallelToRay checks if this XLINE is parallel to a Ray
func (x *XLine) IsParallelToRay(ray *Ray) bool {
	return x.IsParallelTo(ray.XLine)
}

// IntersectWith finds the intersection point between two XLINEs
// Returns error if lines are parallel
func (x *XLine) IntersectWith(other *XLine) (dxfmath.Vec3, error) {
	if x.IsParallelTo(other) {
		return dxfmath.NewVec3(0, 0, 0), errors.New("lines are parallel, no intersection")
	}

	// Use 2D intersection in the XY plane
	return x.intersect2D(x.start.X(), x.start.Y(), x.unitVector.X(), x.unitVector.Y(),
		other.start.X(), other.start.Y(), other.unitVector.X(), other.unitVector.Y())
}

// IntersectWithRay finds the intersection point between XLINE and Ray
// Returns error if line is parallel to ray or intersection is behind ray start
func (x *XLine) IntersectWithRay(ray *Ray) (dxfmath.Vec3, error) {
	// First find line-line intersection
	point, err := x.IntersectWith(ray.XLine)
	if err != nil {
		return point, err
	}

	// Check if intersection is in the ray direction (positive projection from ray start)
	toIntersection := point.Sub(ray.start)
	projection := toIntersection.Dot(ray.unitVector)
	if projection < -1e-9 {
		return dxfmath.NewVec3(0, 0, 0), errors.New("intersection is behind ray start")
	}

	return point, nil
}

// intersect2D performs 2D line intersection calculation
func (x *XLine) intersect2D(x1, y1, dx1, dy1, x2, y2, dx2, dy2 float64) (dxfmath.Vec3, error) {
	// Solve the system:
	// x1 + t1*dx1 = x2 + t2*dx2
	// y1 + t1*dy1 = y2 + t2*dy2

	denominator := dx1*dy2 - dy1*dx2
	if math.Abs(denominator) < 1e-9 {
		return dxfmath.NewVec3(0, 0, 0), errors.New("lines are parallel")
	}

	// Calculate t1 for the first line
	t1 := ((x2-x1)*dy2 - (y2-y1)*dx2) / denominator

	// Calculate intersection point
	intersectX := x1 + t1*dx1
	intersectY := y1 + t1*dy1

	return dxfmath.NewVec3(intersectX, intersectY, 0), nil
}

// OrthogonalAt returns a perpendicular XLINE at the given point
func (x *XLine) OrthogonalAt(point dxfmath.Vec3) *XLine {
	// Create perpendicular direction vector
	perpDir := dxfmath.NewVec3(-x.unitVector.Y(), x.unitVector.X(), 0)
	if perpDir.IsEqual(dxfmath.NewVec3(0, 0, 0), 1e-9) {
		// If original line is vertical, use horizontal
		perpDir = dxfmath.NewVec3(1, 0, 0)
	}

	return NewXLineFromPointDirection(point, perpDir)
}

// Angle returns the angle of the XLINE in the XY plane (radians from X-axis)
func (x *XLine) Angle() float64 {
	return math.Atan2(x.unitVector.Y(), x.unitVector.X())
}

// Slope returns the slope in the XY plane (dy/dx), returns NaN for vertical lines
func (x *XLine) Slope() float64 {
	if math.Abs(x.unitVector.X()) < 1e-9 {
		return math.NaN() // Vertical line
	}
	return x.unitVector.Y() / x.unitVector.X()
}

// IsVertical returns true if the line is vertical in the XY plane
func (x *XLine) IsVertical() bool {
	return math.Abs(x.unitVector.X()) < 1e-9
}

// IsHorizontal returns true if the line is horizontal in the XY plane
func (x *XLine) IsHorizontal() bool {
	return math.Abs(x.unitVector.Y()) < 1e-9
}

// YAtX returns the Y coordinate for a given X coordinate (if line is not vertical)
func (x *XLine) YAtX(xCoord float64) (float64, error) {
	if x.IsVertical() {
		return 0, errors.New("line is vertical, YAtX undefined")
	}

	// Use point-slope form: y - y1 = m(x - x1)
	slope := x.Slope()
	y := x.start.Y() + slope*(xCoord-x.start.X())
	return y, nil
}

// XAtY returns the X coordinate for a given Y coordinate (if line is not horizontal)
func (x *XLine) XAtY(yCoord float64) (float64, error) {
	if x.IsHorizontal() {
		return 0, errors.New("line is horizontal, XAtY undefined")
	}

	// Rearrange point-slope form
	slope := x.Slope()
	if math.IsNaN(slope) {
		// Vertical line
		return x.start.X(), nil
	}
	xCoord := x.start.X() + (yCoord-x.start.Y())/slope
	return xCoord, nil
}

// ClosestPoint returns the closest point on XLINE to the given point
func (x *XLine) ClosestPoint(point dxfmath.Vec3) dxfmath.Vec3 {
	toPoint := point.Sub(x.start)
	projectionLength := toPoint.Dot(x.unitVector)
	return x.start.Add(x.unitVector.Mul(projectionLength))
}

// Ray represents a DXF RAY entity (semi-infinite ray)
// A RAY extends infinitely in one direction from a start point
type Ray struct {
	*XLine // Embed XLine for shared functionality
}

// NewRay creates a new RAY entity
func NewRay() *Ray {
	return &Ray{
		XLine: NewXLine(), // Start with default XLine
	}
}

// NewRayFromPointDirection creates a RAY from a point and direction vector
func NewRayFromPointDirection(start, direction dxfmath.Vec3) *Ray {
	return &Ray{
		XLine: NewXLineFromPointDirection(start, direction),
	}
}

// IsEntity returns true for RAY entities
func (r *Ray) IsEntity() bool {
	return true
}

// Format writes RAY entity data to DXF format
func (r *Ray) Format(f format.Formatter) {
	r.entity.Format(f)
	f.WriteString(100, "AcDbRay")

	// Write start point (10,20,30)
	f.WriteFloat(10, r.start.X())
	f.WriteFloat(20, r.start.Y())
	f.WriteFloat(30, r.start.Z())

	// Write unit direction vector (11,21,31)
	f.WriteFloat(11, r.unitVector.X())
	f.WriteFloat(21, r.unitVector.Y())
	f.WriteFloat(31, r.unitVector.Z())
}

// BBox returns a bounding box for RAY
// For rays, this returns a bounding box from start point extending in direction
func (r *Ray) BBox() ([]float64, []float64) {
	// For rays, return a bounding box from start point extending forward
	bounds := 1000.0
	endPoint := r.start.Add(r.unitVector.Mul(bounds))

	// Calculate min and max coordinates
	minX := math.Min(r.start.X(), endPoint.X())
	minY := math.Min(r.start.Y(), endPoint.Y())
	minZ := math.Min(r.start.Z(), endPoint.Z())
	maxX := math.Max(r.start.X(), endPoint.X())
	maxY := math.Max(r.start.Y(), endPoint.Y())
	maxZ := math.Max(r.start.Z(), endPoint.Z())

	mins := []float64{minX, minY, minZ}
	maxs := []float64{maxX, maxY, maxZ}

	return mins, maxs
}

// GetPointsForDrawing returns points for drawing the ray within a bounds
func (r *Ray) GetPointsForDrawing(infiniteLineLength float64) (dxfmath.Vec3, dxfmath.Vec3) {
	endPoint := r.start.Add(r.unitVector.Mul(infiniteLineLength))
	return r.start, endPoint
}

// ClosestPoint returns the closest point on RAY to the given point
func (r *Ray) ClosestPoint(point dxfmath.Vec3) dxfmath.Vec3 {
	toPoint := point.Sub(r.start)
	projectionLength := toPoint.Dot(r.unitVector)

	// For rays, if projection is negative, closest point is the start point
	if projectionLength < 0 {
		return r.start
	}

	return r.start.Add(r.unitVector.Mul(projectionLength))
}

// DistanceToPoint calculates the shortest distance from RAY to a point
func (r *Ray) DistanceToPoint(point dxfmath.Vec3) float64 {
	closestPoint := r.ClosestPoint(point)
	return point.Sub(closestPoint).Length()
}

// IsParallelTo checks if this Ray is parallel to another Ray
func (r *Ray) IsParallelTo(other *Ray) bool {
	return r.XLine.IsParallelTo(other.XLine)
}

// IsParallelToXLine checks if this Ray is parallel to an XLINE
func (r *Ray) IsParallelToXLine(xline *XLine) bool {
	return r.XLine.IsParallelTo(xline)
}

// IntersectWith finds the intersection point between two Rays
// Returns error if rays are parallel or if intersection is not in the forward direction of both rays
func (r *Ray) IntersectWith(other *Ray) (dxfmath.Vec3, error) {
	// First find line-line intersection
	point, err := r.XLine.IntersectWith(other.XLine)
	if err != nil {
		return point, err
	}

	// Check if intersection is in forward direction for both rays
	if !r.isPointForward(point) {
		return dxfmath.NewVec3(0, 0, 0), errors.New("intersection is behind first ray start")
	}

	if !other.isPointForward(point) {
		return dxfmath.NewVec3(0, 0, 0), errors.New("intersection is behind second ray start")
	}

	return point, nil
}

// IntersectWithXLine finds the intersection point between Ray and XLINE
// Returns error if ray is parallel to line
func (r *Ray) IntersectWithXLine(xline *XLine) (dxfmath.Vec3, error) {
	return xline.IntersectWithRay(r)
}

// isPointForward checks if a point is in the forward direction of the ray
func (r *Ray) isPointForward(point dxfmath.Vec3) bool {
	toPoint := point.Sub(r.start)
	projection := toPoint.Dot(r.unitVector)
	return projection >= -1e-9
}

// OrthogonalAt returns a perpendicular Ray at the given point
func (r *Ray) OrthogonalAt(point dxfmath.Vec3) *Ray {
	perpXLine := r.XLine.OrthogonalAt(point)
	return &Ray{XLine: perpXLine}
}

// Angle returns the angle of the Ray in the XY plane (radians from X-axis)
func (r *Ray) Angle() float64 {
	return r.XLine.Angle()
}

// Slope returns the slope in the XY plane (dy/dx), returns NaN for vertical rays
func (r *Ray) Slope() float64 {
	return r.XLine.Slope()
}

// IsVertical returns true if the ray is vertical in the XY plane
func (r *Ray) IsVertical() bool {
	return r.XLine.IsVertical()
}

// IsHorizontal returns true if the ray is horizontal in the XY plane
func (r *Ray) IsHorizontal() bool {
	return r.XLine.IsHorizontal()
}

// YAtX returns the Y coordinate for a given X coordinate (if ray is not vertical)
// Returns error if the X coordinate is behind the ray start
func (r *Ray) YAtX(xCoord float64) (float64, error) {
	y, err := r.XLine.YAtX(xCoord)
	if err != nil {
		return y, err
	}

	// Check if this point is in the forward direction
	point := dxfmath.NewVec3(xCoord, y, 0)
	if !r.isPointForward(point) {
		return 0, errors.New("X coordinate is behind ray start")
	}

	return y, nil
}

// XAtY returns the X coordinate for a given Y coordinate (if ray is not horizontal)
// Returns error if the X coordinate is behind the ray start
func (r *Ray) XAtY(yCoord float64) (float64, error) {
	x, err := r.XLine.XAtY(yCoord)
	if err != nil {
		return x, err
	}

	// Check if this point is in the forward direction
	point := dxfmath.NewVec3(x, yCoord, 0)
	if !r.isPointForward(point) {
		return 0, errors.New("Y coordinate is behind ray start")
	}

	return x, nil
}
