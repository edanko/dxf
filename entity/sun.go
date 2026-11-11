package entity

import (
	"fmt"
	"math"
	"time"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	dxfmath "github.com/edanko/dxf/math"
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

// NewSunWithDateTime creates a sun with specific date and time. The integer
// Julian day stored on the entity is the noon Julian day of the calendar
// date (computed with the shared math package civil-to-Julian conversion),
// and the time of day is stored separately as seconds past midnight.
func NewSunWithDateTime(year, month, day, hour, minute, second int) *Sun {
	sun := NewSun()
	jd := dxfmath.CivilToJulian(dxfmath.CivilDateTime{
		Year:   year,
		Month:  month,
		Day:    day,
		Hour:   12,
		Minute: 0,
		Second: 0,
	})
	sun.JulianDay = int(jd)
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

// CalculatePosition calculates the sun position from the stored date and
// time using the NOAA General Solar Position Calculations. An absolute
// Julian day (>2000000) is converted back to its calendar date; a
// day-of-year value (1-366) is used directly. The stored time of day is
// treated as UTC (timezone 0).
func (s *Sun) CalculatePosition(latitude, longitude float64) (elevation, azimuth float64) {
	h, m, sec := s.GetTimeHoursMinutes()
	if s.JulianDay > 2000000 {
		civil := dxfmath.JulianToCivil(float64(s.JulianDay))
		// JulianToCivil returns the noon instant of the stored day; the
		// time of day comes from the separate Time field.
		elevation, azimuth, _ = CalculateSolarPosition(civil.Year, civil.Month, civil.Day, h, m, sec, latitude, longitude)
		return elevation, azimuth
	}
	elevation, azimuth = solarPositionFromDayOfYear(s.JulianDay, s.Time/3600.0, latitude, longitude)
	return elevation, azimuth
}

// CalculateSolarPosition computes the sun's elevation and azimuth angles for
// a calendar date, time of day, and geographic location using the documented
// NOAA General Solar Position Calculations ("General Solar Position
// Calculations", NOAA Global Monitoring Division): a fractional-year
// expansion for the equation of time and solar declination, a time offset
// from the equation of time and longitude, and the standard zenith/azimuth
// spherical formulas. The time of day is treated as UTC (timezone 0), so
// solar noon occurs when the true solar time is 12:00.
//
// It returns an error naming the offending field when the date or time is
// invalid (for example month 13 or day 30 of February).
func CalculateSolarPosition(year, month, day, hour, minute, second int, latitude, longitude float64) (elevation, azimuth float64, err error) {
	if err := ValidateDateTime(year, month, day, hour, minute, second); err != nil {
		return 0, 0, err
	}
	hourFraction := float64(hour) + float64(minute)/60.0 + float64(second)/3600.0
	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	elevation, azimuth = solarPositionFromDayOfYear(t.YearDay(), hourFraction, latitude, longitude)
	return elevation, azimuth, nil
}

// ValidateDateTime reports an error naming the offending field when the
// calendar date or wall-clock time is invalid.
func ValidateDateTime(year, month, day, hour, minute, second int) error {
	if month < 1 || month > 12 {
		return fmt.Errorf("invalid month: %d (must be 1-12)", month)
	}
	days := daysInMonth(year, month)
	if day < 1 || day > days {
		return fmt.Errorf("invalid day: %d (month %d of year %d has %d days)", day, month, year, days)
	}
	if hour < 0 || hour > 23 {
		return fmt.Errorf("invalid hour: %d (must be 0-23)", hour)
	}
	if minute < 0 || minute > 59 {
		return fmt.Errorf("invalid minute: %d (must be 0-59)", minute)
	}
	if second < 0 || second > 59 {
		return fmt.Errorf("invalid second: %d (must be 0-59)", second)
	}
	return nil
}

// daysInMonth returns the number of days in the given month of a year.
func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}

// solarPositionFromDayOfYear implements the NOAA General Solar Position
// Calculations from a day of year (1-366) and the fractional hour of the
// day. See CalculateSolarPosition for the documented model.
func solarPositionFromDayOfYear(dayOfYear int, hourFraction, latitude, longitude float64) (elevation, azimuth float64) {
	// Fractional year in radians (NOAA): gamma = 2*pi/365*(d-1+(h-12)/24).
	gamma := 2 * math.Pi / 365.0 * (float64(dayOfYear-1) + (hourFraction-12.0)/24.0)

	// Equation of time in minutes.
	eqTime := 229.18 * (0.000075 +
		0.001868*math.Cos(gamma) -
		0.032077*math.Sin(gamma) -
		0.014615*math.Cos(2*gamma) -
		0.040849*math.Sin(2*gamma))

	// Solar declination in radians.
	decl := 0.006918 -
		0.399912*math.Cos(gamma) +
		0.070257*math.Sin(gamma) -
		0.006758*math.Cos(2*gamma) +
		0.000907*math.Sin(2*gamma) -
		0.002697*math.Cos(3*gamma) +
		0.00148*math.Sin(3*gamma)

	// Time offset and true solar time in minutes (timezone 0: the stored
	// time is treated as UTC).
	timeOffset := eqTime + 4.0*longitude
	tst := hourFraction*60.0 + timeOffset

	// Hour angle in degrees.
	ha := tst/4.0 - 180.0

	latRad := latitude * math.Pi / 180.0
	haRad := ha * math.Pi / 180.0
	declRad := decl

	// Zenith angle.
	cosZenith := math.Sin(latRad)*math.Sin(declRad) +
		math.Cos(latRad)*math.Cos(declRad)*math.Cos(haRad)
	if cosZenith > 1 {
		cosZenith = 1
	} else if cosZenith < -1 {
		cosZenith = -1
	}
	zenith := math.Acos(cosZenith)
	elevation = 90.0 - zenith*180.0/math.Pi

	// Azimuth measured clockwise from north (NOAA formula plus 180 degrees
	// so that solar noon on the northern-hemisphere equator side reports
	// due south).
	azRad := math.Atan2(
		math.Sin(haRad),
		math.Cos(haRad)*math.Sin(latRad)-math.Tan(declRad)*math.Cos(latRad),
	)
	azimuth = azRad*180.0/math.Pi + 180.0
	if azimuth < 0 {
		azimuth += 360.0
	}
	if azimuth >= 360.0 {
		azimuth -= 360.0
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
