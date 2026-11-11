// Drawing unit definitions and conversion through meters.
package units

import "fmt"

// Unit is a drawing unit matching the $INSUNITS codes.
type Unit int

const (
	Unitless Unit = iota
	Inches
	Feet
	Miles
	Millimeters
	Centimeters
	Meters
	Kilometers
	Microinches
	Mils
	Yards
	Angstroms
	Nanometers
	Microns
	Decimeters
	Decameters
	Hectometers
	Gigameters
	Astronomical
	LightYears
	Parsecs
)

var metersPerUnit = map[Unit]float64{
	Inches:       0.0254,
	Feet:         0.03048,
	Miles:        1609.344,
	Millimeters:  0.001,
	Centimeters:  0.01,
	Meters:       1,
	Kilometers:   1000,
	Microinches:  0.0000254,
	Mils:         0.0000254,
	Yards:        0.9144,
	Angstroms:    1e-10,
	Nanometers:   1e-9,
	Microns:      1e-6,
	Decimeters:   0.1,
	Decameters:   10,
	Hectometers:  100,
	Gigameters:   1e9,
	Astronomical: 1.495978707e11,
	LightYears:   9.4607304725808e15,
	Parsecs:      3.0856775814913673e16,
}

// MetersPerUnit returns the length of one unit in meters.
func MetersPerUnit(u Unit) (float64, error) {
	m, ok := metersPerUnit[u]
	if !ok {
		return 0, fmt.Errorf("units: unsupported unit %d", u)
	}
	return m, nil
}

// Convert converts length between units through meters. Unitless drawings
// reject conversion unless from == to (documented identity policy).
func Convert(length float64, from, to Unit) (float64, error) {
	if from == Unitless || to == Unitless {
		if from == to {
			return length, nil
		}
		return 0, fmt.Errorf("units: cannot convert unitless drawing to unit %d", to)
	}
	fm, err := MetersPerUnit(from)
	if err != nil {
		return 0, err
	}
	tm, err := MetersPerUnit(to)
	if err != nil {
		return 0, err
	}
	return length * fm / tm, nil
}

// FromINSUNITS maps an $INSUNITS code to a Unit.
func FromINSUNITS(code int) (Unit, bool) {
	u := Unit(code)
	if u < Unitless || u > Parsecs {
		return 0, false
	}
	if u != Unitless {
		if _, ok := metersPerUnit[u]; !ok {
			return 0, false
		}
	}
	return u, true
}

// ToINSUNITS maps a Unit to its $INSUNITS code.
func ToINSUNITS(u Unit) int { return int(u) }

// ValidateINSUNITS reports whether code is a known $INSUNITS value.
func ValidateINSUNITS(code int) bool {
	_, ok := FromINSUNITS(code)
	return ok
}
