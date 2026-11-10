package units

import (
	"math"
	"testing"
)

func TestConvert(t *testing.T) {
	got, err := Convert(1000, Millimeters, Meters)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-1) > 1e-9 {
		t.Fatalf("1000mm -> %v m", got)
	}
	in, err := Convert(1, Inches, Millimeters)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(in-25.4) > 1e-9 {
		t.Fatalf("1in -> %v mm", in)
	}
	back, err := Convert(got, Meters, Millimeters)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(back-1000) > 1e-9 {
		t.Fatalf("round trip: %v", back)
	}
}

func TestUnitlessPolicy(t *testing.T) {
	if _, err := Convert(5, Unitless, Meters); err == nil {
		t.Fatal("unitless -> meters must error")
	}
	if got, err := Convert(5, Unitless, Unitless); err != nil || got != 5 {
		t.Fatalf("unitless identity: %v %v", got, err)
	}
}

func TestINSUNITSMapping(t *testing.T) {
	if u, ok := FromINSUNITS(4); !ok || u != Millimeters {
		t.Fatalf("INSUNITS 4: %v %v", u, ok)
	}
	if ToINSUNITS(Millimeters) != 4 {
		t.Fatal("ToINSUNITS")
	}
	if !ValidateINSUNITS(0) || !ValidateINSUNITS(20) || ValidateINSUNITS(99) {
		t.Fatal("ValidateINSUNITS")
	}
}
