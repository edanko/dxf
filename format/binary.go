package format

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"
)

// Binary is Formatter for binary DXF format.
type Binary struct {
	buffer     bytes.Buffer
	float      string // Precision format (unused in binary, kept for compatibility)
	dxfversion string
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

	// Write string as null-terminated UTF-16 LE
	utf16Str := utf16.Encode([]rune(val))
	for _, r := range utf16Str {
		if err := binary.Write(&f.buffer, binary.LittleEndian, r); err != nil {
			return "" // Silent fail for interface compatibility
		}
	}

	// Write null terminator
	if err := binary.Write(&f.buffer, binary.LittleEndian, uint16(0)); err != nil {
		return "" // Silent fail for interface compatibility
	}

	return ""
}

// Int outputs given code & int in binary DXF format.
func (f *Binary) Int(num int, val int) string {
	// Write group code (2 bytes, little-endian)
	if err := binary.Write(&f.buffer, binary.LittleEndian, int16(num)); err != nil {
		return "" // Silent fail for interface compatibility
	}

	// Determine the appropriate integer size based on the group code
	switch {
	case num >= 70 && num <= 79:
		// 16-bit integer
		if err := binary.Write(&f.buffer, binary.LittleEndian, int16(val)); err != nil {
			return "" // Silent fail for interface compatibility
		}
	case num >= 90 && num <= 99:
		// 32-bit integer
		if err := binary.Write(&f.buffer, binary.LittleEndian, int32(val)); err != nil {
			return "" // Silent fail for interface compatibility
		}
	case num >= 270 && num <= 289:
		// 16-bit integer
		if err := binary.Write(&f.buffer, binary.LittleEndian, int16(val)); err != nil {
			return "" // Silent fail for interface compatibility
		}
	case num >= 370 && num <= 389:
		// 16-bit integer
		if err := binary.Write(&f.buffer, binary.LittleEndian, int16(val)); err != nil {
			return "" // Silent fail for interface compatibility
		}
	case num >= 400 && num <= 409:
		// 16-bit integer
		if err := binary.Write(&f.buffer, binary.LittleEndian, int16(val)); err != nil {
			return "" // Silent fail for interface compatibility
		}
	default:
		// Default to 16-bit for other codes
		if err := binary.Write(&f.buffer, binary.LittleEndian, int16(val)); err != nil {
			return "" // Silent fail for interface compatibility
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

// WriteBinaryData outputs binary data for codes 310-319, 1004
func (f *Binary) WriteBinaryData(num int, data []byte) error {
	// Write group code (2 bytes, little-endian)
	if err := binary.Write(&f.buffer, binary.LittleEndian, int16(num)); err != nil {
		return fmt.Errorf("failed to write group code: %w", err)
	}

	// Write data length as int32
	if err := binary.Write(&f.buffer, binary.LittleEndian, int32(len(data))); err != nil {
		return fmt.Errorf("failed to write binary data length: %w", err)
	}

	// Write binary data
	if _, err := f.buffer.Write(data); err != nil {
		return fmt.Errorf("failed to write binary data: %w", err)
	}

	return nil
}

// WriteString appends string data to buffer (returns empty string for compatibility)
func (f *Binary) WriteString(num int, val string) {
	f.String(num, val)
}

// WriteInt appends int data to buffer (returns empty string for compatibility)
func (f *Binary) WriteInt(num int, val int) {
	f.Int(num, val)
}

// WriteFloat appends floating point data to buffer (returns empty string for compatibility)
func (f *Binary) WriteFloat(num int, val float64) {
	f.Float(num, val)
}

// WriteBool appends boolean data to buffer (returns empty string for compatibility)
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

// Version returns the DXF version
func (f *Binary) Version() string {
	return f.dxfversion
}
