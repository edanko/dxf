package math

import "math"

// OCS represents an Object Coordinate System for DXF entities
// The OCS is defined by an extrusion vector (normal to the XY plane)
type OCS struct {
	extrusion Vec3 // Normal vector to the OCS XY plane (Z-axis in OCS)
}

// NewOCS creates a new OCS with the given extrusion vector
// If extrusion is zero vector, uses WCS (world coordinate system)
func NewOCS(extrusion Vec3) OCS {
	if extrusion.IsZero(1e-9) {
		extrusion = Vec3{0, 0, 1} // Default to WCS Z-axis
	}
	return OCS{extrusion: extrusion.Normalized()}
}

// NewOCSIdentity creates an identity OCS (aligned with WCS)
func NewOCSIdentity() OCS {
	return OCS{extrusion: Vec3{0, 0, 1}}
}

// Extrusion returns the extrusion vector (Z-axis in OCS)
func (ocs OCS) Extrusion() Vec3 {
	return ocs.extrusion
}

// ToWCS transforms a point from OCS to World Coordinate System
func (ocs OCS) ToWCS(point Vec3) Vec3 {
	if ocs.extrusion[0] == 0 && ocs.extrusion[1] == 0 && ocs.extrusion[2] == 1 {
		// Identity transformation for WCS-aligned OCS
		return point
	}

	// Calculate arbitrary X-axis perpendicular to extrusion
	var xAxis Vec3
	if math.Abs(ocs.extrusion[0]) < 0.015625 && math.Abs(ocs.extrusion[1]) < 0.015625 {
		// Extrusion is close to Z-axis, use world Y-axis
		xAxis = ocs.extrusion.Cross(Vec3{0, 1, 0}).Normalized()
	} else {
		// Use world Z-axis
		xAxis = ocs.extrusion.Cross(Vec3{0, 0, 1}).Normalized()
	}

	// Y-axis is perpendicular to both extrusion and X-axis
	yAxis := ocs.extrusion.Cross(xAxis).Normalized()

	// Transform: point = x*xAxis + y*yAxis + z*extrusion
	return xAxis.Mul(point[0]).Add(yAxis.Mul(point[1])).Add(ocs.extrusion.Mul(point[2]))
}

// ToOCS transforms a point from World Coordinate System to OCS
func (ocs OCS) ToOCS(point Vec3) Vec3 {
	if ocs.extrusion[0] == 0 && ocs.extrusion[1] == 0 && ocs.extrusion[2] == 1 {
		// Identity transformation for WCS-aligned OCS
		return point
	}

	// Calculate arbitrary X-axis perpendicular to extrusion
	var xAxis Vec3
	if math.Abs(ocs.extrusion[0]) < 0.015625 && math.Abs(ocs.extrusion[1]) < 0.015625 {
		// Extrusion is close to Z-axis, use world Y-axis
		xAxis = ocs.extrusion.Cross(Vec3{0, 1, 0}).Normalized()
	} else {
		// Use world Z-axis
		xAxis = ocs.extrusion.Cross(Vec3{0, 0, 1}).Normalized()
	}

	// Y-axis is perpendicular to both extrusion and X-axis
	yAxis := ocs.extrusion.Cross(xAxis).Normalized()

	// Transform back using dot products
	return Vec3{
		point.Dot(xAxis),
		point.Dot(yAxis),
		point.Dot(ocs.extrusion),
	}
}

// ToWCS2D transforms a 2D point from OCS to WCS (Z=0)
func (ocs OCS) ToWCS2D(point Vec2) Vec3 {
	return ocs.ToWCS(Vec3{point[0], point[1], 0})
}

// ToOCS2D transforms a 3D point from WCS to OCS, returning only XY coordinates
func (ocs OCS) ToOCS2D(point Vec3) Vec2 {
	ocsPoint := ocs.ToOCS(point)
	return Vec2{ocsPoint[0], ocsPoint[1]}
}

// DirectionToWCS transforms a direction vector from OCS to WCS
func (ocs OCS) DirectionToWCS(direction Vec3) Vec3 {
	// For directions, we don't apply translation
	return ocs.ToWCS(direction)
}

// DirectionToOCS transforms a direction vector from WCS to OCS
func (ocs OCS) DirectionToOCS(direction Vec3) Vec3 {
	// For directions, we don't apply translation
	return ocs.ToOCS(direction)
}

// AngleToWCS transforms an angle from OCS to WCS
func (ocs OCS) AngleToWCS(angle float64) float64 {
	if ocs.extrusion[0] == 0 && ocs.extrusion[1] == 0 && ocs.extrusion[2] == 1 {
		// Identity transformation for WCS-aligned OCS
		return angle
	}

	// Calculate arbitrary X-axis perpendicular to extrusion
	var xAxis Vec3
	if math.Abs(ocs.extrusion[0]) < 0.015625 && math.Abs(ocs.extrusion[1]) < 0.015625 {
		// Extrusion is close to Z-axis, use world Y-axis
		xAxis = ocs.extrusion.Cross(Vec3{0, 1, 0}).Normalized()
	} else {
		// Use world Z-axis
		xAxis = ocs.extrusion.Cross(Vec3{0, 0, 1}).Normalized()
	}

	// Transform angle by rotating the X-axis
	// Calculate angle between transformed X-axis and world X-axis
	transformedX := xAxis.Mul(math.Cos(angle)).Add(ocs.extrusion.Cross(xAxis).Mul(math.Sin(angle)))

	return math.Atan2(transformedX[1], transformedX[0])
}

// IsWCSAligned returns true if the OCS is aligned with World Coordinate System
func (ocs OCS) IsWCSAligned() bool {
	return ocs.extrusion[0] == 0 && ocs.extrusion[1] == 0 && ocs.extrusion[2] == 1
}

// String returns string representation of OCS
func (ocs OCS) String() string {
	return "OCS(Extrusion: " + ocs.extrusion.String() + ")"
}
