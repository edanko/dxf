package format

import (
	"bytes"
	"fmt"
	"io"

	"github.com/edanko/dxf/lldxf"
)

// DXF version constants
const (
	AC1009 = "AC1009" // R12
	AC1014 = "AC1014" // R14
	AC1015 = "AC1015" // R2000
	AC1018 = "AC1018" // R2004
	AC1021 = "AC1021" // R2007
	AC1024 = "AC1024" // R2010
	AC1027 = "AC1027" // R2013
	AC1032 = "AC1032" // R2018
)

// ASCII is Formatter for ASCII format.
type ASCII struct {
	buffer   bytes.Buffer
	float    string
	version  string
	codepage string
}

// NewASCII creates a new ASCII formatter.
func NewASCII() *ASCII {
	var b bytes.Buffer
	return &ASCII{
		buffer:  b,
		float:   "%.6f",
		version: AC1021, // Default to R2004 for MPOLYGON support
	}
}

// NewASCIIWithVersion creates a new ASCII formatter with specified version.
func NewASCIIWithVersion(version string) *ASCII {
	if version == "" {
		version = AC1021
	}
	return &ASCII{
		buffer:  bytes.Buffer{},
		float:   "%.6f",
		version: version,
	}
}

// SetVersion sets the DXF version reported by the formatter.
func (f *ASCII) SetVersion(version string) {
	if version != "" {
		f.version = version
	}
}

// SetCodepage sets the codepage used to encode strings for pre-R2007 output.
func (f *ASCII) SetCodepage(name string) { f.codepage = name }

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

// String outputs given code & string in DXF format. For pre-R2007 targets the
// text is encoded in the formatter's codepage; for R2007+ it is emitted as
// UTF-8.
func (f *ASCII) String(num int, val string) string {
	if !Profile(f.version).UTF8 && f.codepage != "" {
		val = string(lldxf.EncodeCodepageBytes(val, f.codepage))
	}
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
	return f.version
}
