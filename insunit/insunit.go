package insunit

import (
	"github.com/edanko/dxf/format"
)

// Unit is the drawing unit for AutoCAD DesignCenter Blocks
type Unit uint8

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

func (u Unit) Format(f format.Formatter) {
	f.WriteInt(70, int(u))
}

type Type int8

const (
	Scientific Type = iota - 1 // We want decimal to be default of zero.
	Decimal                    // This will be zero.
	Engineering
	Architectural
	Fractional
	WindowsDesktop
)

func (t Type) Format(f format.Formatter) {
	// When defining the constants we subtracted two from the values
	// to insure that Decimal was the zero value
	f.WriteInt(70, int(t+2))
}
