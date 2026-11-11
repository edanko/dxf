package math

import (
	"math"
	"time"
)

// JulianEpoch is the Julian date of the Unix epoch, 1970-01-01T00:00:00Z.
const JulianEpoch = 2440587.5

// JulianJ2000 is the Julian date of 2000-01-01T12:00:00Z.
const JulianJ2000 = 2451545.0

// CivilDateTime is a UTC civil date and time. Second may contain a fraction.
type CivilDateTime struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second float64
}

// JulianDate converts a time to a Julian date in UTC.
func JulianDate(t time.Time) float64 {
	utc := t.UTC()
	seconds := utc.Sub(time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)).Seconds()
	return JulianJ2000 + seconds/86400
}

// TimeFromJulian converts a Julian date to a UTC time.
func TimeFromJulian(jd float64) time.Time {
	seconds := (jd - JulianJ2000) * 86400
	ns := int64(math.Round(seconds * 1e9))
	return time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC).Add(time.Duration(ns))
}

// CivilToJulian converts a UTC civil date-time to a Julian date.
func CivilToJulian(c CivilDateTime) float64 {
	base := time.Date(c.Year, time.Month(c.Month), c.Day, c.Hour, c.Minute, 0, 0, time.UTC)
	ns := int64(math.Round(c.Second * 1e9))
	return JulianDate(base.Add(time.Duration(ns)))
}

// JulianToCivil converts a Julian date to a UTC civil date-time.
func JulianToCivil(jd float64) CivilDateTime {
	t := TimeFromJulian(jd)
	second := float64(t.Second()) + float64(t.Nanosecond())/1e9
	return CivilDateTime{
		Year:   t.Year(),
		Month:  int(t.Month()),
		Day:    t.Day(),
		Hour:   t.Hour(),
		Minute: t.Minute(),
		Second: second,
	}
}
