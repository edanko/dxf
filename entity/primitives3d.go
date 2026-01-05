package entity

import (
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

// Create3DCube creates a cube solid from center and size
func Create3DCube(center dxfmath.Vec3, size float64) *Solid3d {
	solid := NewSolid3d()
	solid.SetUID("Cube")

	// Set geometric properties
	volume := size * size * size
	surfaceArea := 6.0 * size * size
	solid.SetVolume(volume)
	solid.SetSurfaceArea(surfaceArea)
	solid.SetCenter(center)

	return solid
}

// Create3DSphere creates a sphere solid
func Create3DSphere(center dxfmath.Vec3, radius float64, segments int) *Solid3d {
	solid := NewSolid3d()
	solid.SetUID("Sphere")

	// Calculate volume and surface area
	volume := (4.0 / 3.0) * math.Pi * radius * radius * radius
	surfaceArea := 4.0 * math.Pi * radius * radius

	solid.SetVolume(volume)
	solid.SetSurfaceArea(surfaceArea)
	solid.SetCenter(center)

	return solid
}

// Create3DCylinder creates a cylinder solid
func Create3DCylinder(center dxfmath.Vec3, radius, height float64, segments int) *Solid3d {
	solid := NewSolid3d()
	solid.SetUID("Cylinder")

	// Calculate volume and surface area
	volume := math.Pi * radius * radius * height
	surfaceArea := 2.0 * math.Pi * radius * (radius + height)

	solid.SetVolume(volume)
	solid.SetSurfaceArea(surfaceArea)
	solid.SetCenter(center)

	return solid
}

// Create3DCone creates a cone solid
func Create3DCone(center dxfmath.Vec3, radius1, radius2, height float64, segments int) *Solid3d {
	solid := NewSolid3d()
	solid.SetUID("Cone")

	// Calculate volume (frustum of cone)
	var volume float64
	if radius2 == 0 {
		// Pointed cone
		volume = (1.0 / 3.0) * math.Pi * radius1 * radius1 * height
	} else {
		// Truncated cone
		volume = (1.0 / 3.0) * math.Pi * height * (radius1*radius1 + radius1*radius2 + radius2*radius2)
	}

	// Calculate surface area (lateral + top + bottom)
	var surfaceArea float64
	if radius2 == 0 {
		// Pointed cone
		slantHeight := math.Sqrt(radius1*radius1 + height*height)
		lateralArea := math.Pi * radius1 * slantHeight
		bottomArea := math.Pi * radius1 * radius1
		surfaceArea = lateralArea + bottomArea
	} else {
		// Truncated cone
		lateralArea := math.Pi * (radius1 + radius2) * math.Sqrt((radius1-radius2)*(radius1-radius2)+height*height)
		topArea := math.Pi * radius2 * radius2
		bottomArea := math.Pi * radius1 * radius1
		surfaceArea = lateralArea + topArea + bottomArea
	}

	solid.SetVolume(volume)
	solid.SetSurfaceArea(surfaceArea)
	solid.SetCenter(center)

	return solid
}

// Create3DTorus creates a torus solid
func Create3DTorus(center dxfmath.Vec3, majorRadius, minorRadius float64, majorSegments, minorSegments int) *Solid3d {
	solid := NewSolid3d()
	solid.SetUID("Torus")

	// Calculate volume and surface area
	volume := 2.0 * math.Pi * math.Pi * majorRadius * minorRadius * minorRadius
	surfaceArea := 4.0 * math.Pi * math.Pi * majorRadius * minorRadius

	solid.SetVolume(volume)
	solid.SetSurfaceArea(surfaceArea)
	solid.SetCenter(center)

	return solid
}
