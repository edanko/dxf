package entity

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

// Trace represents a TRACE entity (filled trace/outline)
type Trace struct {
	*entity

	// Core properties
	Points    []float64 `dxf:"10"` // 4 required points for filled trace
	Plane     int       `dxf:"70"` // 0=XY, 1=YZ, 2=ZX (extrusion plane)
	Thickness float64   `dxf:"39"` // Line thickness
}

// NewTrace creates a new TRACE entity
func NewTrace(points []float64, plane int, thickness float64) *Trace {
	// Validate input
	if len(points) < 4 {
		// Create empty trace for invalid input
		return &Trace{
			entity:    NewEntity(TRACE),
			Points:    make([]float64, 4),
			Plane:     0,
			Thickness: 0.0,
		}
	}

	// Ensure at least 4 points
	validPoints := make([]float64, 4)
	copy(validPoints, points)
	if len(points) < 4 {
		// Fill remaining with first point
		for i := len(points); i < 4; i++ {
			validPoints[i] = points[0]
		}
	}

	return &Trace{
		entity:    NewEntity(TRACE),
		Points:    validPoints,
		Plane:     plane,
		Thickness: thickness,
	}
}

// IsEntity returns true for TRACE entities
func (t *Trace) IsEntity() bool {
	return true
}

// Format writes TRACE entity data to formatter
func (t *Trace) Format(f format.Formatter) {
	t.entity.Format(f)
	f.WriteString(100, "AcDbTrace")

	// Write points (4 points minimum)
	for i, point := range t.Points {
		f.WriteFloat(10+i, point)
	}

	// Write extrusion plane
	f.WriteInt(70, t.Plane)

	// Write thickness
	if t.Thickness != 0.0 {
		f.WriteFloat(39, t.Thickness)
	}
}

// BBox returns bounding box for TRACE
func (t *Trace) BBox() ([]float64, []float64) {
	if len(t.Points) == 0 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}

	minX, minY := t.Points[0], t.Points[1]
	maxX, maxY := t.Points[0], t.Points[1]

	// Handle 4-point filled polygon
	numPoints := len(t.Points)
	for i := 0; i < numPoints; i += 2 {
		if i+1 < numPoints {
			x := t.Points[i]
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}

			y := t.Points[i+1]
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	// Handle Z coordinates based on plane
	var zMin, zMax float64
	if t.Plane == 0 { // XY plane
		if len(t.Points) > 2 {
			zMin = t.Points[2]
		} else {
			zMin = 0.0
		}
		zMax = zMin
	} else {
		zMin, zMax = 0.0, 0.0 // For other planes, ignore Z for now
	}

	return []float64{minX, minY, zMin}, []float64{maxX, maxY, zMax}
}

// SetColor sets the color for the TRACE
func (t *Trace) SetColor(c color.ColorNumber) {
	t.entity.color = c
}

// SetLayer sets the layer for the TRACE
func (t *Trace) SetLayer(l *table.Layer) {
	t.entity.layer = l
}

// SetLtscale sets the linetype scale for the TRACE
func (t *Trace) SetLtscale(scale float64) {
	t.entity.ltscale = scale
}

// SetBlockRecord sets the block record for the TRACE
func (t *Trace) SetBlockRecord(br handle.Handler) {
	t.entity.blockRecord = br
}

// DXFType returns the entity type string
func (t *Trace) DXFType() string {
	return "TRACE"
}

// GetAttributes returns entity attributes as a map
func (t *Trace) GetAttributes() map[string]interface{} {
	attrs := make(map[string]interface{})

	// Basic attributes
	attrs["points"] = t.Points
	attrs["plane"] = t.Plane
	attrs["thickness"] = t.Thickness
	attrs["entity_type"] = "TRACE"

	// Inherited attributes
	attrs["handle"] = t.Handle()
	attrs["layer"] = t.Layer()
	attrs["color"] = t.entity.color
	attrs["ltscale"] = t.entity.ltscale

	return attrs
}

// LoadAttributes loads entity attributes from a map
func (t *Trace) LoadAttributes(attrs map[string]interface{}) error {
	// Load points attribute
	if points, ok := attrs["points"].([]float64); ok {
		if len(points) >= 4 {
			t.Points = make([]float64, 4)
			copy(t.Points, points[:4])
		}
	}

	// Load plane attribute
	if plane, ok := attrs["plane"].(int); ok {
		t.Plane = plane
	}

	// Load thickness attribute
	if thickness, ok := attrs["thickness"].(float64); ok {
		t.Thickness = thickness
	}

	return nil
}
