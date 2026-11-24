package format

import (
	"bytes"
	"io"
	"strconv"

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
//
// Version-derived flags and the float precision are cached when version or
// codepage change: every written tag consults them, and resolving
// format.Profile by version string per tag dominated serialization. The
// scratch slice is reused for the per-tag text so that writing a tag does not
// allocate.
type ASCII struct {
	buffer   bytes.Buffer
	float    string
	version  string
	codepage string
	prec     int
	encode   bool
	scratch  []byte
}

// NewASCII creates a new ASCII formatter.
func NewASCII() *ASCII {
	var b bytes.Buffer
	f := &ASCII{
		buffer:  b,
		float:   "%.6f",
		prec:    6,
		version: AC1021, // Default to R2004 for MPOLYGON support
	}
	f.refresh()
	return f
}

// NewASCIIWithVersion creates a new ASCII formatter with specified version.
func NewASCIIWithVersion(version string) *ASCII {
	if version == "" {
		version = AC1021
	}
	f := &ASCII{
		buffer:  bytes.Buffer{},
		float:   "%.6f",
		prec:    6,
		version: version,
	}
	f.refresh()
	return f
}

// refresh recomputes the flags derived from version and codepage.
func (f *ASCII) refresh() {
	f.encode = !Profile(f.version).UTF8 && f.codepage != ""
}

// SetVersion sets the DXF version reported by the formatter.
func (f *ASCII) SetVersion(version string) {
	if version != "" {
		f.version = version
		f.refresh()
	}
}

// SetCodepage sets the codepage used to encode strings for pre-R2007 output.
func (f *ASCII) SetCodepage(name string) {
	f.codepage = name
	f.refresh()
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
	f.float = "%." + strconv.Itoa(p) + "f"
	f.prec = p
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
	if f.encode {
		val = string(lldxf.EncodeCodepageBytes(val, f.codepage))
	}
	b := make([]byte, 0, len(val)+16)
	b = strconv.AppendInt(b, int64(num), 10)
	b = append(b, '\n')
	b = append(b, val...)
	b = append(b, '\n')
	return string(b)
}

// Int outputs given code & int in DXF format.
func (f *ASCII) Int(num int, val int) string {
	b := make([]byte, 0, 24)
	b = strconv.AppendInt(b, int64(num), 10)
	b = append(b, '\n')
	b = strconv.AppendInt(b, int64(val), 10)
	b = append(b, '\n')
	return string(b)
}

// Float outputs given code & floating point in DXF format.
func (f *ASCII) Float(num int, val float64) string {
	b := make([]byte, 0, 32)
	b = strconv.AppendInt(b, int64(num), 10)
	b = append(b, '\n')
	b = strconv.AppendFloat(b, val, 'f', f.prec, 64)
	b = append(b, '\n')
	return string(b)
}

// appendTagHeader appends "<code>\n" to dst.
func appendTagHeader(dst []byte, num int) []byte {
	dst = strconv.AppendInt(dst, int64(num), 10)
	return append(dst, '\n')
}

// WriteString appends string data to the buffer.
func (f *ASCII) WriteString(num int, val string) {
	if f.encode {
		val = string(lldxf.EncodeCodepageBytes(val, f.codepage))
	}
	b := appendTagHeader(f.scratch[:0], num)
	b = append(b, val...)
	b = append(b, '\n')
	f.scratch = b
	f.buffer.Write(b)
}

// WriteInt appends int data to the buffer.
func (f *ASCII) WriteInt(num int, val int) {
	b := appendTagHeader(f.scratch[:0], num)
	b = strconv.AppendInt(b, int64(val), 10)
	b = append(b, '\n')
	f.scratch = b
	f.buffer.Write(b)
}

// WriteFloat appends floating point data to the buffer.
func (f *ASCII) WriteFloat(num int, val float64) {
	b := appendTagHeader(f.scratch[:0], num)
	b = strconv.AppendFloat(b, val, 'f', f.prec, 64)
	b = append(b, '\n')
	f.scratch = b
	f.buffer.Write(b)
}

// WriteBool appends boolean data to the buffer.
func (f *ASCII) WriteBool(num int, val bool) {
	intVal := 0
	if val {
		intVal = 1
	}
	f.WriteInt(num, intVal)
}

// Version returns the DXF version
func (f *ASCII) Version() string {
	return f.version
}
