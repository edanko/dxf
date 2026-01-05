package math

import (
	"fmt"
	"math"
)

// UCS represents an enhanced User Coordinate System with full transformation support
type UCS struct {
	origin Vec3
	xAxis  Vec3
	yAxis  Vec3
	zAxis  Vec3
}

// NewUCS creates a new UCS from origin and axes
func NewUCS(origin, xAxis, yAxis, zAxis Vec3) *UCS {
	return &UCS{
		origin: origin,
		xAxis:  xAxis.Normalized(),
		yAxis:  yAxis.Normalized(),
		zAxis:  zAxis.Normalized(),
	}
}

// WorldUCS returns the standard world coordinate system
func WorldUCS() *UCS {
	return &UCS{
		origin: Vec3{0, 0, 0},
		xAxis:  Vec3{1, 0, 0},
		yAxis:  Vec3{0, 1, 0},
		zAxis:  Vec3{0, 0, 1},
	}
}

// UCSFromEulerAngles creates UCS from origin and Euler angles (in radians)
func UCSFromEulerAngles(origin Vec3, yaw, pitch, roll float64) *UCS {
	// Create rotation matrices
	yawMatrix := RotationY(yaw)
	pitchMatrix := RotationX(pitch)
	rollMatrix := RotationZ(roll)

	// Combined rotation: R = Ry * Rx * Rz
	rotation := yawMatrix.Mul(pitchMatrix).Mul(rollMatrix)

	// Extract axes from rotation matrix
	xAxis := Vec3{rotation.Get(0, 0), rotation.Get(1, 0), rotation.Get(2, 0)}
	yAxis := Vec3{rotation.Get(0, 1), rotation.Get(1, 1), rotation.Get(2, 1)}
	zAxis := Vec3{rotation.Get(0, 2), rotation.Get(1, 2), rotation.Get(2, 2)}

	return &UCS{
		origin: origin,
		xAxis:  xAxis.Normalized(),
		yAxis:  yAxis.Normalized(),
		zAxis:  zAxis.Normalized(),
	}
}

// UCSFromAxisAndAngle creates UCS from origin, primary axis, and rotation angle
func UCSFromAxisAndAngle(origin, axis Vec3, angle float64) *UCS {
	// Create rotation matrix using Rodrigues' rotation formula
	k := axis.Normalized()
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	// Rotation matrix components
	rotation := Matrix44{
		cos + k.X()*k.X()*(1-cos), k.X()*k.Y()*(1-cos) - k.Z()*sin, 0,
		k.Y()*k.X()*(1-cos) + k.Z()*sin, cos + k.Y()*k.Y()*(1-cos), 0,
		k.X() * sin, k.Y() * sin, cos + k.Z()*k.Z()*(1-cos),
		0, 0, 0, 1,
	}

	// Extract axes
	xAxis := Vec3{rotation.Get(0, 0), rotation.Get(1, 0), rotation.Get(2, 0)}
	yAxis := Vec3{rotation.Get(0, 1), rotation.Get(1, 1), rotation.Get(2, 1)}
	zAxis := Vec3{rotation.Get(0, 2), rotation.Get(1, 2), rotation.Get(2, 2)}

	return &UCS{
		origin: origin,
		xAxis:  xAxis.Normalized(),
		yAxis:  yAxis.Normalized(),
		zAxis:  zAxis.Normalized(),
	}
}

// ToWorld transforms a point from UCS to world coordinates
func (ucs *UCS) ToWorld(point Vec3) Vec3 {
	// Transform: world_point = origin + point.x * x_axis + point.y * y_axis + point.z * z_axis
	return ucs.origin.Add(
		ucs.xAxis.Mul(point.X()).Add(
			ucs.yAxis.Mul(point.Y()).Add(
				ucs.zAxis.Mul(point.Z()),
			),
		),
	)
}

// ToUCS transforms a point from world to UCS coordinates
func (ucs *UCS) ToUCS(point Vec3) Vec3 {
	// Transform: ucs_point = (world_point - origin) expressed in UCS basis
	worldPoint := point.Sub(ucs.origin)

	// Project onto UCS axes
	x := worldPoint.Dot(ucs.xAxis)
	y := worldPoint.Dot(ucs.yAxis)
	z := worldPoint.Dot(ucs.zAxis)

	return Vec3{x, y, z}
}

// TransformVector transforms a vector from UCS to world coordinates
func (ucs *UCS) TransformVector(vector Vec3) Vec3 {
	// Transform: world_vector = vector.x * x_axis + vector.y * y_axis + vector.z * z_axis
	return ucs.xAxis.Mul(vector.X()).Add(
		ucs.yAxis.Mul(vector.Y()).Add(
			ucs.zAxis.Mul(vector.Z()),
		),
	)
}

// InverseTransformVector transforms a vector from world to UCS coordinates
func (ucs *UCS) InverseTransformVector(vector Vec3) Vec3 {
	// Transform: ucs_vector = world_vector expressed in UCS basis
	x := vector.Dot(ucs.xAxis)
	y := vector.Dot(ucs.yAxis)
	z := vector.Dot(ucs.zAxis)

	return Vec3{x, y, z}
}

// Matrix returns the transformation matrix from UCS to world
func (ucs *UCS) Matrix() Matrix44 {
	return Matrix44FromColumns(
		ucs.xAxis,
		ucs.yAxis,
		ucs.zAxis,
		ucs.origin,
	)
}

// InverseMatrix returns the transformation matrix from world to UCS
func (ucs *UCS) InverseMatrix() Matrix44 {
	matrix := ucs.Matrix()
	result, _ := matrix.Inverse()
	return result
}

// TransformDirection transforms a direction vector (ignoring translation)
func (ucs *UCS) TransformDirection(direction Vec3) Vec3 {
	return ucs.TransformVector(direction)
}

// InverseTransformDirection transforms a direction from world to UCS
func (ucs *UCS) InverseTransformDirection(direction Vec3) Vec3 {
	return ucs.InverseTransformVector(direction)
}

// Rotate rotates the UCS around its own axes
func (ucs *UCS) Rotate(yaw, pitch, roll float64) *UCS {
	// Create rotation matrices for local rotation
	yawMatrix := RotationY(yaw)
	pitchMatrix := RotationX(pitch)
	rollMatrix := RotationZ(roll)

	// Combined rotation
	rotation := yawMatrix.Mul(pitchMatrix).Mul(rollMatrix)

	// Transform existing axes
	newXAxis := rotation.TransformVector(ucs.xAxis)
	newYAxis := rotation.TransformVector(ucs.yAxis)
	newZAxis := rotation.TransformVector(ucs.zAxis)

	return &UCS{
		origin: ucs.origin,
		xAxis:  newXAxis.Normalized(),
		yAxis:  newYAxis.Normalized(),
		zAxis:  newZAxis.Normalized(),
	}
}

// Translate translates the UCS origin
func (ucs *UCS) Translate(offset Vec3) *UCS {
	return &UCS{
		origin: ucs.origin.Add(offset),
		xAxis:  ucs.xAxis,
		yAxis:  ucs.yAxis,
		zAxis:  ucs.zAxis,
	}
}

// Scale scales the UCS axes (maintains orthogonality if scale is uniform)
func (ucs *UCS) Scale(scale Vec3) *UCS {
	return &UCS{
		origin: ucs.origin,
		xAxis:  ucs.xAxis.Mul(scale.X()).Normalized(),
		yAxis:  ucs.yAxis.Mul(scale.Y()).Normalized(),
		zAxis:  ucs.zAxis.Mul(scale.Z()).Normalized(),
	}
}

// IsOrthogonal checks if UCS axes are orthogonal
func (ucs *UCS) IsOrthogonal() bool {
	tolerance := 1e-10

	// Check if axes are perpendicular
	xyOrthogonal := math.Abs(ucs.xAxis.Dot(ucs.yAxis)) < tolerance
	xzOrthogonal := math.Abs(ucs.xAxis.Dot(ucs.zAxis)) < tolerance
	yzOrthogonal := math.Abs(ucs.yAxis.Dot(ucs.zAxis)) < tolerance

	return xyOrthogonal && xzOrthogonal && yzOrthogonal
}

// IsRightHanded checks if UCS forms a right-handed coordinate system
func (ucs *UCS) IsRightHanded() bool {
	// For right-handed system: x × y = z
	cross := ucs.xAxis.Cross(ucs.yAxis)
	dot := cross.Dot(ucs.zAxis)
	return dot > 0
}

// String returns string representation
func (ucs *UCS) String() string {
	return fmt.Sprintf("UCS{origin: %v, xAxis: %v, yAxis: %v, zAxis: %v}",
		ucs.origin, ucs.xAxis, ucs.yAxis, ucs.zAxis)
}

// Matrix44FromColumns creates a 4x4 matrix from column vectors and origin
func Matrix44FromColumns(xAxis, yAxis, zAxis, origin Vec3) Matrix44 {
	return Matrix44{
		xAxis.X(), yAxis.X(), zAxis.X(), origin.X(),
		xAxis.Y(), yAxis.Y(), zAxis.Y(), origin.Y(),
		xAxis.Z(), yAxis.Z(), zAxis.Z(), origin.Z(),
		0, 0, 0, 1,
	}
}
