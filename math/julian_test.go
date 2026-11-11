package math

import (
	"math"
	"testing"
	"time"
)

func TestJulianJ2000Epoch(t *testing.T) {
	jd := CivilToJulian(CivilDateTime{Year: 2000, Month: 1, Day: 1, Hour: 12})
	if math.Abs(jd-JulianJ2000) > 1.0/86400 {
		t.Fatalf("J2000 = %.12f, want %.1f", jd, JulianJ2000)
	}
	got := JulianToCivil(JulianJ2000)
	want := CivilDateTime{Year: 2000, Month: 1, Day: 1, Hour: 12}
	if got.Year != want.Year || got.Month != want.Month || got.Day != want.Day || got.Hour != want.Hour || got.Minute != want.Minute || math.Abs(got.Second) > 1e-6 {
		t.Fatalf("civil from J2000 = %+v", got)
	}
}

func TestJulianRoundTrip(t *testing.T) {
	original := time.Date(2026, 11, 11, 8, 30, 15, 250000000, time.UTC)
	jd := JulianDate(original)
	got := TimeFromJulian(jd)
	if d := got.Sub(original); d > time.Second || d < -time.Second {
		t.Fatalf("round trip delta = %v", d)
	}
	civil := CivilToJulian(CivilDateTime{Year: 2026, Month: 11, Day: 11, Hour: 8, Minute: 30, Second: 15.25})
	if math.Abs(civil-jd) > 1.0/86400 {
		t.Fatalf("civil julian = %.12f, time julian = %.12f", civil, jd)
	}
}
