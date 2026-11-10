package format

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/edanko/dxf/lldxf"
)

// Binary is Formatter for binary DXF format.
type Binary struct {
	buffer     bytes.Buffer
	float      string // Precision format (unused in binary, kept for compatibility)
	dxfversion string
	codepage   string
}

// NewBinary creates a new binary formatter.
func NewBinary(version string) *Binary {
	return &Binary{
		dxfversion: version,
	}
}

// Reset resets buffer.
func (f *Binary) Reset() {
	f.buffer.Reset()
}

// WriteTo writes data stored in buffer to w.
func (f *Binary) WriteTo(w io.Writer) (int64, error) {
	return f.buffer.WriteTo(w)
}

// SetPrecision sets precision part for outputting floating point values.
// In binary format, precision is handled by the binary encoding itself.
func (f *Binary) SetPrecision(p int) {
	// No-op for binary format - precision is inherent in float64 encoding
	f.float = fmt.Sprintf("%%.%df", p)
}

// Output outputs data stored in buffer in DXF format.
func (f *Binary) Output() string {
	rtn := f.buffer.String()
	f.buffer.Reset()
	return rtn
}

// String outputs given code & string in binary DXF format.
func (f *Binary) String(num int, val string) string {
	// Write group code (2 bytes, little-endian)
	if err := binary.Write(&f.buffer, binary.LittleEndian, int16(num)); err != nil {
		return "" // Silent fail for interface compatibility
	}

	if Profile(f.dxfversion).UTF8 {
		// R2007+: null-terminated UTF-8 bytes.
		f.buffer.WriteString(val)
		f.buffer.WriteByte(0x00)
	} else {
		// Pre-R2007: null-terminated codepage bytes.
		for _, b := range lldxf.EncodeCodepageBytes(val, f.codepage) {
			f.buffer.WriteByte(b)
		}
		f.buffer.WriteByte(0x00)
	}

	return ""
}

// Int outputs given code & int in binary DXF format.
func (f *Binary) Int(num int, val int) string {
	// Write group code (2 bytes, little-endian)
	if err := binary.Write(&f.buffer, binary.LittleEndian, int16(num)); err != nil {
		return "" // Silent fail for interface compatibility
	}

	switch lldxf.BinaryKindFor(lldxf.GroupCode(num)) {
	case lldxf.BinaryInt64:
		if err := binary.Write(&f.buffer, binary.LittleEndian, int64(val)); err != nil {
			return ""
		}
	case lldxf.BinaryInt32:
		if err := binary.Write(&f.buffer, binary.LittleEndian, int32(val)); err != nil {
			return ""
		}
	default:
		// int16 for 16-bit codes and boolean codes
		if err := binary.Write(&f.buffer, binary.LittleEndian, int16(val)); err != nil {
			return ""
		}
	}

	return ""
}

// Float outputs given code & floating point in binary DXF format.
func (f *Binary) Float(num int, val float64) string {
	// Write group code (2 bytes, little-endian)
	if err := binary.Write(&f.buffer, binary.LittleEndian, int16(num)); err != nil {
		return "" // Silent fail for interface compatibility
	}

	// Write 64-bit double precision float
	if err := binary.Write(&f.buffer, binary.LittleEndian, val); err != nil {
		return "" // Silent fail for interface compatibility
	}

	return ""
}

// WriteInt64 outputs given code & 64-bit integer in binary DXF format.
func (f *Binary) WriteInt64(num int, val int64) {
	if err := binary.Write(&f.buffer, binary.LittleEndian, int16(num)); err != nil {
		return
	}
	if err := binary.Write(&f.buffer, binary.LittleEndian, val); err != nil {
		return
	}
}

// WriteBinaryData outputs binary data for codes 310-319, 1004: a one-byte
// length followed by the raw bytes, matching AutoCAD Binary DXF.
func (f *Binary) WriteBinaryData(num int, data []byte) error {
	if err := binary.Write(&f.buffer, binary.LittleEndian, int16(num)); err != nil {
		return fmt.Errorf("failed to write group code: %w", err)
	}
	if len(data) > 255 {
		data = data[:255]
	}
	f.buffer.WriteByte(byte(len(data)))
	if _, err := f.buffer.Write(data); err != nil {
		return fmt.Errorf("failed to write binary data: %w", err)
	}
	return nil
}

// BytesWritten reports how many bytes are currently buffered.
func (f *Binary) BytesWritten() int {
	return f.buffer.Len()
}

// SetCodepage sets the codepage used to encode strings for pre-R2007 output.
func (f *Binary) SetCodepage(name string) { f.codepage = name }

// WriteString appends string data to buffer. Raw text values (retained group
// codes from a parsed stream, subclass markers, handles) are re-typed by group
// code so the binary stream matches what a tagger reads back: boolean codes
// (290-299) become int16 flags and integer codes become sized ints.
func (f *Binary) WriteString(num int, val string) {
	f.WriteValue(num, val)
}

// WriteInt appends int data to buffer.
func (f *Binary) WriteInt(num int, val int) {
	f.Int(num, val)
}

// WriteFloat appends floating point data to buffer.
func (f *Binary) WriteFloat(num int, val float64) {
	f.Float(num, val)
}

// WriteBool appends boolean data to buffer as an int16 flag.
func (f *Binary) WriteBool(num int, val bool) {
	intVal := 0
	if val {
		intVal = 1
	}
	f.Int(num, intVal)
}

// WriteSignature writes the binary DXF signature
func (f *Binary) WriteSignature() error {
	signature := "AutoCAD Binary DXF\r\n\x1a\x00"
	if _, err := f.buffer.WriteString(signature); err != nil {
		return fmt.Errorf("failed to write binary DXF signature: %w", err)
	}
	return nil
}

// WriteValue writes a tag value using the encoding for its group code.
func (f *Binary) WriteValue(num int, val any) {
	switch lldxf.BinaryKindFor(lldxf.GroupCode(num)) {
	case lldxf.BinaryFloat:
		f.WriteFloat(num, toFloat64(val))
	case lldxf.BinaryInt16, lldxf.BinaryInt32, lldxf.BinaryInt64:
		if lldxf.IsBinaryBoolCode(lldxf.GroupCode(num)) {
			b := false
			switch v := val.(type) {
			case bool:
				b = v
			case string:
				b = v == "true" || v == "1" || toInt(v) != 0
			default:
				b = toInt(val) != 0
			}
			f.WriteBool(num, b)
		} else {
			f.WriteInt(num, toInt(val))
		}
	case lldxf.BinaryRaw:
		data, _ := val.([]byte)
		f.WriteBinaryData(num, data)
	default:
		f.String(num, toString(val))
	}
}

func toFloat64(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		var x float64
		fmt.Sscanf(n, "%g", &x)
		return x
	}
	return 0
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int16:
		return int(n)
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		var x int
		fmt.Sscanf(n, "%d", &x)
		return x
	}
	return 0
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// Version returns the DXF version
func (f *Binary) Version() string {
	return f.dxfversion
}
