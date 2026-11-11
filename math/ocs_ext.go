package math

import "math"

// ocs_ext.go extends OCS with the plane basis accessors and the inverse angle
// conversion (OCS.AngleToWCS in ocs.go is the forward direction). The plane
// basis follows the convention of ToWCS in ocs.go (ezdxf OCS.from_wcs /
// to_wcs, src/ezdxf/src/ezdxf/math/ucs.py:27): the X axis is perpendicular to
// the extrusion, the Y axis completes the right-handed basis.

// ocsPlaneBasis returns the OCS plane basis axes as used by ToWCS/ToOCS,
// following ezdxf OCS.__init__ (src/ezdxf/src/ezdxf/math/ucs.py:37): the X
// axis is the reference axis crossed into the extrusion (world Y for
// extrusions close to Z, world Z otherwise), the Y axis completes the
// right-handed basis (extrusion × X).
func ocsPlaneBasis(extrusion Vec3) (xAxis, yAxis Vec3) {
	if math.Abs(extrusion[0]) < 0.015625 && math.Abs(extrusion[1]) < 0.015625 {
		// Extrusion is close to the Z axis: ezdxf uses Y_AXIS.cross(Az).
		xAxis = Vec3{0, 1, 0}.Cross(extrusion).Normalized()
	} else {
		// ezdxf uses Z_AXIS.cross(Az).
		xAxis = Vec3{0, 0, 1}.Cross(extrusion).Normalized()
	}
	yAxis = extrusion.Cross(xAxis).Normalized()
	return xAxis, yAxis
}

// AxisVectors returns the plane basis axes of the OCS: the X axis is
// perpendicular to the extrusion, the Y axis completes the right-handed
// basis (extrusion × X). The extrusion itself is the OCS Z axis and is
// available via Extrusion.
func (ocs OCS) AxisVectors() (xAxis, yAxis Vec3) {
	if ocs.IsWCSAligned() {
		return Vec3{1, 0, 0}, Vec3{0, 1, 0}
	}
	return ocsPlaneBasis(ocs.Extrusion())
}

// AngleToOCS transforms a WCS angle (radians, measured about the WCS Z axis)
// into the OCS angle of the same direction. It is the exact inverse of
// AngleToWCS: for the +Z extrusion (WCS aligned) both are the identity,
// otherwise the direction is projected onto the OCS plane basis, mirroring
// angles at the axis (ezdxf OCSParts.from_world_angle semantics:
// normalize(2·axis_angle − angle) for the axis-angle form).
func (ocs OCS) AngleToOCS(wcsAngle float64) float64 {
	if ocs.IsWCSAligned() {
		return wcsAngle
	}
	xAxis, yAxis := ocsPlaneBasis(ocs.Extrusion())
	// 2D projections of the basis: the forward transform in AngleToWCS
	// maps (cos θ, sin θ) through M = [xAxis.YX | yAxis.YX]; invert via
	// the adjugate, with the sign of det(M) restoring the orientation.
	m00, m01 := xAxis[0], yAxis[0]
	m10, m11 := xAxis[1], yAxis[1]
	det := m00*m11 - m01*m10
	sign := 1.0
	if det < 0 {
		sign = -1
	}
	c := math.Cos(wcsAngle)
	s := math.Sin(wcsAngle)
	// adj(M) · (c, s) = (m11·c − m01·s, −m10·c + m00·s)
	return math.Atan2(sign*(-m10*c+m00*s), sign*(m11*c-m01*s))
}
