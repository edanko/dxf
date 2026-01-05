package entity

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// LightType represents different types of lights
type LightType int

const (
	LightTypeDistant LightType = 1
	LightTypePoint   LightType = 2
	LightTypeSpot    LightType = 3
)

// String returns string representation of light type
func (lt LightType) String() string {
	switch lt {
	case LightTypeDistant:
		return "Distant"
	case LightTypePoint:
		return "Point"
	case LightTypeSpot:
		return "Spot"
	default:
		return "Unknown"
	}
}

// AttenuationType represents light attenuation types
type AttenuationType int

const (
	AttenuationNone          AttenuationType = 0
	AttenuationInverseLinear AttenuationType = 1
	AttenuationInverseSquare AttenuationType = 2
)

// String returns string representation of attenuation type
func (at AttenuationType) String() string {
	switch at {
	case AttenuationNone:
		return "None"
	case AttenuationInverseLinear:
		return "Inverse Linear"
	case AttenuationInverseSquare:
		return "Inverse Square"
	default:
		return "Unknown"
	}
}

// ShadowType represents shadow types
type ShadowType int

const (
	ShadowTypeRayTraced  ShadowType = 0
	ShadowTypeShadowMaps ShadowType = 1
)

// String returns string representation of shadow type
func (st ShadowType) String() string {
	switch st {
	case ShadowTypeRayTraced:
		return "Ray Traced"
	case ShadowTypeShadowMaps:
		return "Shadow Maps"
	default:
		return "Unknown"
	}
}

// Light represents a DXF LIGHT entity for 3D rendering
type Light struct {
	*entity
	handle string

	// Core properties
	Version   int       `dxf:"90"`  // Light version
	Name      string    `dxf:"1"`   // Light name
	Type      LightType `dxf:"70"`  // Light type (1=distant, 2=point, 3=spot)
	Status    bool      `dxf:"290"` // Light status (on/off)
	PlotGlyph bool      `dxf:"291"` // Plot glyph in drawing

	// Light properties
	Intensity float64    `dxf:"40"` // Light intensity (0.0 to 1.0)
	Location  [3]float64 `dxf:"10"` // Light position (X, Y, Z)
	Target    [3]float64 `dxf:"11"` // Light target (X, Y, Z)

	// Attenuation properties
	AttenuationType      AttenuationType `dxf:"72"`  // Attenuation type
	UseAttenuationLimits bool            `dxf:"292"` // Use attenuation limits
	AttenuationStart     float64         `dxf:"41"`  // Attenuation start limit
	AttenuationEnd       float64         `dxf:"42"`  // Attenuation end limit

	// Spotlight properties (for LightTypeSpot)
	HotspotAngle float64 `dxf:"50"` // Hotspot angle in degrees
	FalloffAngle float64 `dxf:"51"` // Falloff angle in degrees

	// Shadow properties
	CastShadows       bool       `dxf:"293"` // Cast shadows
	ShadowType        ShadowType `dxf:"73"`  // Shadow type
	ShadowMapSize     int        `dxf:"91"`  // Shadow map size
	ShadowMapSoftness int        `dxf:"280"` // Shadow map softness
}

// NewLight creates a new LIGHT entity
func NewLight() *Light {
	hg := handle.NewHandleGenerator()
	light := &Light{
		entity:               NewEntity(LIGHT),
		handle:               hg.Next(),
		Version:              0,
		Name:                 "",
		Type:                 LightTypeDistant, // Default to distant light
		Status:               true,
		PlotGlyph:            false,
		Intensity:            1.0,
		Location:             [3]float64{0, 0, 0},
		Target:               [3]float64{0, 0, -1}, // Default target pointing down
		AttenuationType:      AttenuationNone,
		UseAttenuationLimits: false,
		AttenuationStart:     1.0,
		AttenuationEnd:       0.0,
		HotspotAngle:         45.0, // Default spotlight cone
		FalloffAngle:         60.0, // Default falloff cone
		CastShadows:          true,
		ShadowType:           ShadowTypeRayTraced,
		ShadowMapSize:        512,
		ShadowMapSoftness:    2,
	}
	return light
}

// NewDistantLight creates a distant light (like the sun)
func NewDistantLight(name string, direction [3]float64) *Light {
	light := NewLight()
	light.Name = name
	light.Type = LightTypeDistant
	light.Target = direction // For distant lights, target is direction
	return light
}

// NewPointLight creates a point light
func NewPointLight(name string, position [3]float64, intensity float64) *Light {
	light := NewLight()
	light.Name = name
	light.Type = LightTypePoint
	light.Location = position
	light.Intensity = intensity
	light.AttenuationType = AttenuationInverseSquare
	return light
}

// NewSpotLight creates a spotlight
func NewSpotLight(name string, position, target [3]float64, hotspot, falloff float64) *Light {
	light := NewLight()
	light.Name = name
	light.Type = LightTypeSpot
	light.Location = position
	light.Target = target
	light.HotspotAngle = hotspot
	light.FalloffAngle = falloff
	light.AttenuationType = AttenuationInverseSquare
	return light
}

// IsOn returns true if the light is turned on
func (l *Light) IsOn() bool {
	return l.Status
}

// TurnOn turns the light on
func (l *Light) TurnOn() {
	l.Status = true
}

// TurnOff turns the light off
func (l *Light) TurnOff() {
	l.Status = false
}

// SetIntensity sets the light intensity (0.0 to 1.0)
func (l *Light) SetIntensity(intensity float64) {
	l.Intensity = math.Max(0.0, math.Min(1.0, intensity))
}

// SetLocation sets the light position
func (l *Light) SetLocation(x, y, z float64) {
	l.Location = [3]float64{x, y, z}
}

// SetTarget sets the light target/direction
func (l *Light) SetTarget(x, y, z float64) {
	l.Target = [3]float64{x, y, z}
}

// SetAttenuation sets attenuation parameters
func (l *Light) SetAttenuation(attType AttenuationType, start, end float64) {
	l.AttenuationType = attType
	l.AttenuationStart = start
	l.AttenuationEnd = end
	l.UseAttenuationLimits = (attType != AttenuationNone)
}

// SetSpotlightAngles sets spotlight cone angles (for spot lights)
func (l *Light) SetSpotlightAngles(hotspot, falloff float64) {
	l.HotspotAngle = math.Max(0, math.Min(180, hotspot))
	l.FalloffAngle = math.Max(0, math.Min(180, falloff))
}

// SetShadowProperties configures shadow casting
func (l *Light) SetShadowProperties(cast bool, shadowType ShadowType, mapSize, softness int) {
	l.CastShadows = cast
	l.ShadowType = shadowType
	l.ShadowMapSize = mapSize
	l.ShadowMapSoftness = softness
}

// GetDirection returns the light direction (for distant and spot lights)
func (l *Light) GetDirection() [3]float64 {
	if l.Type == LightTypePoint {
		return [3]float64{0, 0, 0} // Point lights have no direction
	}

	// Calculate direction from location to target
	dx := l.Target[0] - l.Location[0]
	dy := l.Target[1] - l.Location[1]
	dz := l.Target[2] - l.Location[2]

	// Normalize vector
	length := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if length > 0 {
		return [3]float64{dx / length, dy / length, dz / length}
	}
	return [3]float64{0, 0, -1} // Default pointing down
}

// DistanceTo calculates distance from light to a point
func (l *Light) DistanceTo(point [3]float64) float64 {
	dx := point[0] - l.Location[0]
	dy := point[1] - l.Location[1]
	dz := point[2] - l.Location[2]
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// GetAttenuationFactor calculates attenuation factor at a given distance
func (l *Light) GetAttenuationFactor(distance float64) float64 {
	if !l.UseAttenuationLimits || l.AttenuationType == AttenuationNone {
		return 1.0
	}

	if distance <= l.AttenuationStart {
		return 1.0
	}

	if distance >= l.AttenuationEnd {
		return 0.0
	}

	switch l.AttenuationType {
	case AttenuationInverseLinear:
		return l.AttenuationStart / distance
	case AttenuationInverseSquare:
		return (l.AttenuationStart * l.AttenuationStart) / (distance * distance)
	default:
		return 1.0
	}
}

// Validate checks if the light entity has valid properties
func (l *Light) Validate() []string {
	errors := make([]string, 0)

	// Check name
	if l.Name == "" {
		errors = append(errors, "Light name cannot be empty")
	}

	// Check type
	if l.Type < LightTypeDistant || l.Type > LightTypeSpot {
		errors = append(errors, fmt.Sprintf("Invalid light type: %d", l.Type))
	}

	// Check intensity
	if l.Intensity < 0 || l.Intensity > 1 {
		errors = append(errors, fmt.Sprintf("Light intensity must be between 0 and 1, got: %f", l.Intensity))
	}

	// Check attenuation
	if l.AttenuationType < AttenuationNone || l.AttenuationType > AttenuationInverseSquare {
		errors = append(errors, fmt.Sprintf("Invalid attenuation type: %d", l.AttenuationType))
	}

	// Check attenuation limits
	if l.UseAttenuationLimits {
		if l.AttenuationEnd <= l.AttenuationStart {
			errors = append(errors, "Attenuation end must be greater than start")
		}
		if l.AttenuationStart < 0 {
			errors = append(errors, "Attenuation start cannot be negative")
		}
	}

	// Check spotlight angles
	if l.Type == LightTypeSpot {
		if l.HotspotAngle < 0 || l.HotspotAngle > 180 {
			errors = append(errors, fmt.Sprintf("Hotspot angle must be between 0 and 180, got: %f", l.HotspotAngle))
		}
		if l.FalloffAngle < 0 || l.FalloffAngle > 180 {
			errors = append(errors, fmt.Sprintf("Falloff angle must be between 0 and 180, got: %f", l.FalloffAngle))
		}
		if l.FalloffAngle < l.HotspotAngle {
			errors = append(errors, "Falloff angle must be greater than or equal to hotspot angle")
		}
	}

	// Check shadow properties
	if l.ShadowType < ShadowTypeRayTraced || l.ShadowType > ShadowTypeShadowMaps {
		errors = append(errors, fmt.Sprintf("Invalid shadow type: %d", l.ShadowType))
	}

	if l.ShadowMapSize <= 0 {
		errors = append(errors, "Shadow map size must be positive")
	}

	return errors
}

// IsEntity is for Entity interface.
func (l *Light) IsEntity() bool {
	return true
}

// Handle returns the light's handle
func (l *Light) Handle() string {
	return l.handle
}

// SetHandle sets the light's handle
func (l *Light) SetHandle(h *handle.HandleGenerator) {
	l.handle = h.Next()
}

// SetBlockRecord sets block record for light
func (l *Light) SetBlockRecord(h handle.Handler) {
	l.entity.SetBlockRecord(h)
}

// Format writes data to formatter.
func (l *Light) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLight")

	// Write version and name
	f.WriteInt(90, l.Version)
	f.WriteString(1, l.Name)

	// Write light type and status
	f.WriteInt(70, int(l.Type))
	// Write boolean status as integer (0 or 1)
	if l.Status {
		f.WriteInt(290, 1)
	} else {
		f.WriteInt(290, 0)
	}

	if l.PlotGlyph {
		f.WriteInt(291, 1)
	} else {
		f.WriteInt(291, 0)
	}

	// Write intensity and location
	f.WriteFloat(40, l.Intensity)
	f.WriteFloat(10, l.Location[0])
	f.WriteFloat(20, l.Location[1])
	f.WriteFloat(30, l.Location[2])

	// Write target/direction
	f.WriteFloat(11, l.Target[0])
	f.WriteFloat(21, l.Target[1])
	f.WriteFloat(31, l.Target[2])

	// Write attenuation properties
	f.WriteInt(72, int(l.AttenuationType))
	if l.UseAttenuationLimits {
		f.WriteInt(292, 1)
	} else {
		f.WriteInt(292, 0)
	}
	if l.UseAttenuationLimits || l.AttenuationType != AttenuationNone {
		f.WriteFloat(41, l.AttenuationStart)
		f.WriteFloat(42, l.AttenuationEnd)
	}

	// Write spotlight properties if applicable
	if l.Type == LightTypeSpot {
		f.WriteFloat(50, l.HotspotAngle)
		f.WriteFloat(51, l.FalloffAngle)
	}

	// Write shadow properties
	if l.CastShadows {
		f.WriteInt(293, 1)
	} else {
		f.WriteInt(293, 0)
	}
	f.WriteInt(73, int(l.ShadowType))
	f.WriteInt(91, l.ShadowMapSize)
	f.WriteInt(280, l.ShadowMapSoftness)
}

// Clone creates a copy of the light entity
func (l *Light) Clone() Entity {
	hg := handle.NewHandleGenerator()
	clone := &Light{
		entity: NewEntity(LIGHT),
		handle: hg.Next(),

		// Copy entity fields
		Version:              l.Version,
		Name:                 l.Name,
		Type:                 l.Type,
		Status:               l.Status,
		PlotGlyph:            l.PlotGlyph,
		Intensity:            l.Intensity,
		Location:             l.Location,
		Target:               l.Target,
		AttenuationType:      l.AttenuationType,
		UseAttenuationLimits: l.UseAttenuationLimits,
		AttenuationStart:     l.AttenuationStart,
		AttenuationEnd:       l.AttenuationEnd,
		HotspotAngle:         l.HotspotAngle,
		FalloffAngle:         l.FalloffAngle,
		CastShadows:          l.CastShadows,
		ShadowType:           l.ShadowType,
		ShadowMapSize:        l.ShadowMapSize,
		ShadowMapSoftness:    l.ShadowMapSoftness,
	}

	// Copy base entity fields
	clone.entity.Type = l.entity.Type
	clone.entity.handle = l.entity.handle
	clone.entity.blockRecord = l.entity.blockRecord
	clone.entity.owner = l.entity.owner
	clone.entity.layer = l.entity.layer
	clone.entity.ltscale = l.entity.ltscale
	clone.entity.color = l.entity.color

	return clone
}

// String returns string representation of the light
func (l *Light) String() string {
	return fmt.Sprintf("LIGHT(%s): %s at (%.2f,%.2f,%.2f)",
		l.Name, l.Type.String(), l.Location[0], l.Location[1], l.Location[2])
}

// BBox returns the bounding box for the light
func (l *Light) BBox() ([]float64, []float64) {
	// For lights, we use a simple bbox around the location
	minX := l.Location[0] - 1.0
	maxX := l.Location[0] + 1.0
	minY := l.Location[1] - 1.0
	maxY := l.Location[1] + 1.0
	minZ := l.Location[2] - 1.0
	maxZ := l.Location[2] + 1.0

	return []float64{minX, minY, minZ}, []float64{maxX, maxY, maxZ}
}
