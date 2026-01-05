package entity

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// Sun represents a DXF SUN entity for realistic outdoor lighting
type Sun struct {
	*entity
	handle string

	// Core properties
	Version int  `dxf:"90"`  // Sun version
	Status  bool `dxf:"290"` // Sun status (on/off)

	// Color properties
	Color     color.ColorNumber `dxf:"63"`  // Sun color index
	TrueColor int               `dxf:"421"` // True color (RGB)

	// Light properties
	Intensity   float64 `dxf:"40"`  // Sun intensity (0.0 to 1.0)
	CastShadows bool    `dxf:"291"` // Cast shadows

	// Date and time properties
	JulianDay           int     `dxf:"91"`  // Julian day (1-366)
	Time                float64 `dxf:"92"`  // Time in seconds past midnight
	DaylightSavingsTime bool    `dxf:"292"` // Daylight savings time flag

	// Shadow properties
	ShadowType     ShadowType `dxf:"70"`  // Shadow type (0=ray traced, 1=shadow maps)
	ShadowMapSize  int        `dxf:"71"`  // Shadow map size
	ShadowSoftness int        `dxf:"280"` // Shadow softness
}

// NewSun creates a new SUN entity with default values
func NewSun() *Sun {
	hg := handle.NewHandleGenerator()
	sun := &Sun{
		entity:              NewEntity(SUN),
		handle:              hg.Next(),
		Version:             1,
		Status:              true,
		Color:               color.White, // Default white sunlight
		TrueColor:           16777215,    // Default RGB white
		Intensity:           1.0,
		CastShadows:         true,
		JulianDay:           2456922, // Default to a reasonable date
		Time:                43200.0, // Default to 12:00 PM (12*3600)
		DaylightSavingsTime: false,
		ShadowType:          ShadowTypeRayTraced,
		ShadowMapSize:       256,
		ShadowSoftness:      1,
	}
	return sun
}

// NewSunWithDateTime creates a sun with specific date and time
func NewSunWithDateTime(year, month, day, hour, minute, second int) *Sun {
	sun := NewSun()

	// Calculate Julian day (simplified calculation)
	// For simplicity, using a more basic calculation
	yearAdj := year
	monthAdj := month
	if month <= 2 {
		yearAdj = year - 1
		monthAdj = month + 12
	}
	a := yearAdj / 100
	b := a / 4
	c := 2 - a + b
	julianDay := int(365.25*float64(yearAdj) + 30.6001*float64(monthAdj) + float64(day) + float64(c) + 1720995)

	sun.JulianDay = julianDay
	sun.Time = float64(hour*3600 + minute*60 + second)
	return sun
}

// SetTime sets the time of day (in seconds past midnight)
func (s *Sun) SetTime(hour, minute, second int) {
	s.Time = float64(hour)*3600 + float64(minute*60) + float64(second)
}

// GetTimeHoursMinutes extracts time as hours and minutes
func (s *Sun) GetTimeHoursMinutes() (hours, minutes, seconds int) {
	totalSeconds := int(s.Time)
	hours = totalSeconds / 3600
	minutes = (totalSeconds % 3600) / 60
	seconds = totalSeconds % 60
	return
}

// SetShadowProperties configures shadow casting
func (s *Sun) SetShadowProperties(cast bool, shadowType ShadowType, mapSize, softness int) {
	s.CastShadows = cast
	s.ShadowType = shadowType
	s.ShadowMapSize = mapSize
	s.ShadowSoftness = softness
}

// CalculatePosition calculates sun position based on date, time, and location
func (s *Sun) CalculatePosition(latitude, longitude float64) (elevation, azimuth float64) {
	// Simplified sun position calculation
	// In a real implementation, this would use more complex astronomical calculations

	// Convert Julian day to day of year
	dayOfYear := s.JulianDay % 1000 // Simplified

	// Solar declination (simplified)
	declination := 23.45 * math.Sin((360.0/365.0*float64(dayOfYear-81))*math.Pi/180.0)

	// Hour angle
	hours := s.Time / 3600.0
	hourAngle := 15.0 * (hours - 12.0) // 15 degrees per hour from solar noon

	// Solar elevation angle
	latitudeRad := latitude * math.Pi / 180.0
	elevation = math.Asin(
		math.Sin(declination*math.Pi/180.0)*math.Sin(latitudeRad)+
			math.Cos(declination*math.Pi/180.0)*math.Cos(latitudeRad)*math.Cos(hourAngle*math.Pi/180.0),
	) * 180.0 / math.Pi

	// Solar azimuth angle
	azimuth = math.Atan2(
		-math.Cos(declination*math.Pi/180.0)*math.Sin(hourAngle*math.Pi/180.0),
		math.Cos(declination*math.Pi/180.0)*math.Cos(latitudeRad)*math.Sin(hourAngle*math.Pi/180.0)-
			math.Sin(declination*math.Pi/180.0)*math.Sin(latitudeRad),
	) * 180.0 / math.Pi

	if azimuth < 0 {
		azimuth += 360.0
	}

	return elevation, azimuth
}

// GetDirectionVector calculates sun direction from position
func (s *Sun) GetDirectionVector(latitude, longitude float64) [3]float64 {
	elevation, azimuth := s.CalculatePosition(latitude, longitude)

	// Convert spherical to Cartesian coordinates
	elevRad := elevation * math.Pi / 180.0
	azimRad := azimuth * math.Pi / 180.0

	x := math.Cos(elevRad) * math.Sin(azimRad)
	y := math.Cos(elevRad) * math.Cos(azimRad)
	z := math.Sin(elevRad)

	return [3]float64{x, y, z}
}

// IsOn returns true if sun is enabled
func (s *Sun) IsOn() bool {
	return s.Status
}

// TurnOn turns the sun on
func (s *Sun) TurnOn() {
	s.Status = true
}

// TurnOff turns the sun off
func (s *Sun) TurnOff() {
	s.Status = false
}

// Validate checks if sun entity has valid properties
func (s *Sun) Validate() []string {
	errors := make([]string, 0)

	// Check Julian day - allow both day-of-year (1-366) and absolute Julian days (>2000000)
	if s.JulianDay < 1 || (s.JulianDay > 366 && s.JulianDay < 2000000) {
		errors = append(errors, fmt.Sprintf("Julian day must be between 1-366 (day of year) or >2000000 (absolute Julian day), got: %d", s.JulianDay))
	}

	// Check time
	if s.Time < 0 || s.Time >= 86400 {
		errors = append(errors, fmt.Sprintf("Time must be between 0 and 86400 seconds, got: %f", s.Time))
	}

	// Check color
	if s.Color < 0 || s.Color > 255 {
		errors = append(errors, fmt.Sprintf("Color index must be between 0 and 255, got: %d", s.Color))
	}

	// Check intensity
	if s.Intensity < 0 || s.Intensity > 1 {
		errors = append(errors, fmt.Sprintf("Intensity must be between 0 and 1, got: %f", s.Intensity))
	}

	// Check shadow properties
	if s.ShadowType < ShadowTypeRayTraced || s.ShadowType > ShadowTypeShadowMaps {
		errors = append(errors, fmt.Sprintf("Invalid shadow type: %d", s.ShadowType))
	}

	if s.ShadowMapSize <= 0 {
		errors = append(errors, "Shadow map size must be positive")
	}

	return errors
}

// IsEntity is for Entity interface.
func (s *Sun) IsEntity() bool {
	return true
}

// Handle returns the sun's handle
func (s *Sun) Handle() string {
	return s.handle
}

// SetHandle sets the sun's handle
func (s *Sun) SetHandle(h *handle.HandleGenerator) {
	s.handle = h.Next()
}

// SetBlockRecord sets block record for sun
func (s *Sun) SetBlockRecord(h handle.Handler) {
	s.entity.SetBlockRecord(h)
}

// Format writes data to formatter.
func (s *Sun) Format(f format.Formatter) {
	s.entity.Format(f)
	f.WriteString(100, "AcDbSun")

	// Write version and status
	f.WriteInt(90, s.Version)
	if s.Status {
		f.WriteInt(290, 1)
	} else {
		f.WriteInt(290, 0)
	}

	// Write color properties
	f.WriteInt(63, int(s.Color))
	f.WriteInt(421, s.TrueColor)

	// Write intensity
	f.WriteFloat(40, s.Intensity)

	// Write shadow properties
	if s.CastShadows {
		f.WriteInt(291, 1)
	} else {
		f.WriteInt(291, 0)
	}

	// Write date and time properties
	f.WriteInt(91, s.JulianDay)
	f.WriteFloat(92, s.Time)

	if s.DaylightSavingsTime {
		f.WriteInt(292, 1)
	} else {
		f.WriteInt(292, 0)
	}

	// Write shadow type and properties
	f.WriteInt(70, int(s.ShadowType))
	f.WriteInt(71, s.ShadowMapSize)
	f.WriteInt(280, s.ShadowSoftness)
}

// Clone creates a copy of the sun entity
func (s *Sun) Clone() Entity {
	hg := handle.NewHandleGenerator()
	clone := &Sun{
		entity: NewEntity(SUN),
		handle: hg.Next(),

		// Copy sun properties
		Version:             s.Version,
		Status:              s.Status,
		Color:               s.Color,
		TrueColor:           s.TrueColor,
		Intensity:           s.Intensity,
		CastShadows:         s.CastShadows,
		JulianDay:           s.JulianDay,
		Time:                s.Time,
		DaylightSavingsTime: s.DaylightSavingsTime,
		ShadowType:          s.ShadowType,
		ShadowMapSize:       s.ShadowMapSize,
		ShadowSoftness:      s.ShadowSoftness,
	}

	// Copy base entity fields
	clone.entity.Type = s.entity.Type
	clone.entity.handle = s.entity.handle
	clone.entity.blockRecord = s.entity.blockRecord
	clone.entity.owner = s.entity.owner
	clone.entity.layer = s.entity.layer
	clone.entity.ltscale = s.entity.ltscale
	clone.entity.color = s.entity.color

	return clone
}

// String returns string representation of the sun
func (s *Sun) String() string {
	hours, minutes, _ := s.GetTimeHoursMinutes()
	timeStr := fmt.Sprintf("%02d:%02d", hours, minutes)
	return fmt.Sprintf("SUN: JD %d at %s, intensity %.2f", s.JulianDay, timeStr, s.Intensity)
}

// BBox returns the bounding box for sun (infinite since sun is distant light)
func (s *Sun) BBox() ([]float64, []float64) {
	// Sun affects entire drawing, so return infinite bounds
	return []float64{-1e10, -1e10, -1e10}, []float64{1e10, 1e10, 1e10}
}
