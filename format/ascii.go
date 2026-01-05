package format

import (
	"bytes"
	"fmt"
	"io"
)

// DXF version constants
const (
	AC1009 = "AC1009"
	AC1014 = "AC1014"
	AC1015 = "AC1015"
	AC1018 = "AC1018"
	AC1021 = "AC1021" // R2004
	AC1024 = "AC1024" // R2007
	AC1027 = "AC1027" // R2010
	AC1032 = "AC1032" // R2013
)

// ASCII is Formatter for ASCII format.
type ASCII struct {
	buffer bytes.Buffer
	float  string
}

// NewASCII creates a new ASCII formatter.
func NewASCII() *ASCII {
	var b bytes.Buffer
	return &ASCII{
		buffer: b,
		float:  "%.6f",
	}
}

// NewASCIIWithVersion creates a new ASCII formatter with specified version.
func NewASCIIWithVersion(version string) *ASCII {
	return &ASCII{
		buffer: bytes.Buffer{},
		float:  "%.6f",
	}
}

// Reset resets the buffer.
func (f *ASCII) Reset() {
	f.buffer.Reset()
}

// WriteTo writes data stored in the buffer to w.
func (f *ASCII) WriteTo(w io.Writer) (int64, error) {
	return f.buffer.WriteTo(w)
}

// SetPrecision sets precision part for outputting floating point values.
func (f *ASCII) SetPrecision(p int) {
	f.float = fmt.Sprintf("%%.%df", p)
}

// Output outputs data stored in the buffer in DXF format.
func (f *ASCII) Output() string {
	rtn := f.buffer.String()
	f.buffer.Reset()
	return rtn
}

// String outputs given code & string in DXF format.
func (f *ASCII) String(num int, val string) string {
	return fmt.Sprintf("%d\n%s\n", num, val)
}

// Int outputs given code & int in DXF format.
func (f *ASCII) Int(num int, val int) string {
	return fmt.Sprintf("%d\n%d\n", num, val)
}

// Float outputs given code & floating point in DXF format.
func (f *ASCII) Float(num int, val float64) string {
	return fmt.Sprintf(fmt.Sprintf("%d\n%s\n", num, f.float), val)
}

// WriteString appends string data to the buffer.
func (f *ASCII) WriteString(num int, val string) {
	f.buffer.WriteString(f.String(num, val))
}

// WriteInt appends int data to the buffer.
func (f *ASCII) WriteInt(num int, val int) {
	f.buffer.WriteString(f.Int(num, val))
}

// WriteFloat appends floating point data to the buffer.
func (f *ASCII) WriteFloat(num int, val float64) {
	f.buffer.WriteString(f.Float(num, val))
}

// WriteBool appends boolean data to the buffer.
func (f *ASCII) WriteBool(num int, val bool) {
	intVal := 0
	if val {
		intVal = 1
	}
	f.buffer.WriteString(f.Int(num, intVal))
}

// Version returns the DXF version
func (f *ASCII) Version() string {
	return AC1021 // Default to R2004 for MPOLYGON support
}
