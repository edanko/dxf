package entity

import (
	"math"
	"strings"
	"testing"
)

// TestCalculateSolarPositionKnownDate verifies the NOAA algorithm against
// physical properties at a known date and location: June 15 2023 (declination
// about +23.3 degrees) at latitude 40 N around solar noon, the sun is due
// south and its elevation approaches 90 - |lat - decl|.
func TestCalculateSolarPositionKnownDate(t *testing.T) {
	// Time is UTC (timezone 0): scan all hours of the day to find the peak
	// elevation and its azimuth.
	peakElevation := -math.MaxFloat64
	azimuthAtPeak := 0.0
	for hour := 0; hour < 24; hour++ {
		elev, az, err := CalculateSolarPosition(2023, 6, 15, hour, 0, 0, 40, -105)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if math.IsNaN(elev) || math.IsNaN(az) || math.IsInf(elev, 0) || math.IsInf(az, 0) {
			t.Fatalf("non-finite result: elev=%v az=%v", elev, az)
		}
		if elev > peakElevation {
			peakElevation = elev
			azimuthAtPeak = az
		}
	}
	if peakElevation < 71 || peakElevation > 75 {
		t.Errorf("peak elevation = %v, want ~73.3", peakElevation)
	}
	if math.Abs(azimuthAtPeak-180) > 6 {
		t.Errorf("azimuth at peak = %v, want ~180 (due south)", azimuthAtPeak)
	}
}

// TestCalculateSolarPositionEquinoxEquator verifies the near-zenith sun at
// the equator on the March equinox.
func TestCalculateSolarPositionEquinoxEquator(t *testing.T) {
	// 2024-03-20, equator, longitude 0: sun crosses the zenith around 12:04
	// UTC; elevation must approach 90 near noon.
	peak := -math.MaxFloat64
	for minute := 0; minute < 60; minute++ {
		elev, _, err := CalculateSolarPosition(2024, 3, 20, 12, minute, 0, 0, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if elev > peak {
			peak = elev
		}
	}
	if peak < 88 {
		t.Errorf("peak elevation at equator equinox = %v, want >= 88", peak)
	}
}

// TestCalculateSolarPositionNight verifies the sun is below the horizon at
// night.
func TestCalculateSolarPositionNight(t *testing.T) {
	// Midnight UTC at longitude 0 is local solar midnight: the sun is below
	// the horizon.
	elev, _, err := CalculateSolarPosition(2023, 6, 15, 0, 0, 0, 40, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elev > 0 {
		t.Errorf("midnight elevation = %v, want negative", elev)
	}
	// Direction vector must remain finite and consistent.
	dir := &Sun{Time: 0, JulianDay: 2460111}
	v := dir.GetDirectionVector(40, -105)
	norm := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if math.Abs(norm-1) > 1e-6 {
		t.Errorf("direction vector norm = %v", norm)
	}
}

// TestCalculateSolarPositionInvalidDates verifies errors naming the offending
// field for invalid month and day values.
func TestCalculateSolarPositionInvalidDates(t *testing.T) {
	cases := []struct {
		year, month, day int
		wantField        string
	}{
		{2023, 13, 15, "month"},
		{2023, 0, 15, "month"},
		{2023, 2, 30, "day"},
		{2023, 4, 31, "day"},
		{2023, 6, 0, "day"},
	}
	for _, tc := range cases {
		_, _, err := CalculateSolarPosition(tc.year, tc.month, tc.day, 12, 0, 0, 40, 0)
		if err == nil {
			t.Errorf("%d-%d-%d: expected error", tc.year, tc.month, tc.day)
			continue
		}
		if !strings.Contains(err.Error(), tc.wantField) {
			t.Errorf("%d-%d-%d: error %v should name field %q", tc.year, tc.month, tc.day, err, tc.wantField)
		}
	}

	// Leap year allows February 29; non-leap does not.
	if _, _, err := CalculateSolarPosition(2024, 2, 29, 12, 0, 0, 0, 0); err != nil {
		t.Errorf("2024-02-29 should be valid: %v", err)
	}
	if _, _, err := CalculateSolarPosition(2023, 2, 29, 12, 0, 0, 0, 0); err == nil {
		t.Error("2023-02-29 should be invalid")
	}

	// Sun.CalculatePositionAt delegates validation.
	if err := ValidateDateTime(2023, 6, 15, 25, 0, 0); err == nil || !strings.Contains(err.Error(), "hour") {
		t.Errorf("hour validation: %v", err)
	}
	if err := ValidateDateTime(2023, 6, 15, 12, 0, 60); err == nil || !strings.Contains(err.Error(), "second") {
		t.Errorf("second validation: %v", err)
	}
}

// TestSunCalculatePositionFromJulianDay verifies the absolute Julian day
// interpretation matches the direct date-based calculation.
func TestSunCalculatePositionFromJulianDay(t *testing.T) {
	sun := NewSunWithDateTime(2023, 6, 15, 12, 30, 0)
	if sun.JulianDay != 2460111 {
		t.Errorf("Julian day for 2023-06-15 = %d, want 2460111", sun.JulianDay)
	}
	elev, az := sun.CalculatePosition(40, -105)
	wantElev, wantAz, err := CalculateSolarPosition(2023, 6, 15, 12, 30, 0, 40, -105)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(elev-wantElev) > 1e-9 || math.Abs(az-wantAz) > 1e-9 {
		t.Errorf("Julian-day path = (%v, %v), want (%v, %v)", elev, az, wantElev, wantAz)
	}
}

// TestSunDayOfYearPosition verifies the day-of-year interpretation stays
// finite and physically plausible.
func TestSunDayOfYearPosition(t *testing.T) {
	sun := NewSun()
	sun.JulianDay = 166 // June 15 as day of year
	sun.Time = 43200
	elev, az := sun.CalculatePosition(40, 0)
	if math.IsNaN(elev) || math.IsNaN(az) {
		t.Fatalf("non-finite result: elev=%v az=%v", elev, az)
	}
	if elev < 65 || elev > 80 {
		t.Errorf("elevation = %v, want ~73", elev)
	}
}
