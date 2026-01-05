package lldxf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"
	"unsafe"
)

const (
	BinaryDXFSignature = "AutoCAD Binary DXF\r\n\x1a\x00"
)

// BinaryTag represents a tag in binary DXF format
type BinaryTag struct {
	Code  int
	Value interface{}
}

// BinaryLoader loads binary DXF files
type BinaryLoader struct {
	data     []byte
	pos      int
	version  string
	encoding string
}

// NewBinaryLoader creates a new binary DXF loader
func NewBinaryLoader(data []byte) (*BinaryLoader, error) {
	if len(data) < 22 {
		return nil, fmt.Errorf("file too small for binary DXF")
	}

	if string(data[:22]) != BinaryDXFSignature {
		return nil, fmt.Errorf("not a binary DXF file")
	}

	loader := &BinaryLoader{
		data:     data,
		pos:      22,       // Skip signature
		version:  "AC1009", // Default to R12
		encoding: "cp1252", // Default encoding
	}

	// Detect DXF version and encoding
	if err := loader.detectVersion(); err != nil {
		return nil, fmt.Errorf("failed to detect DXF version: %w", err)
	}

	return loader, nil
}

// detectVersion attempts to detect the DXF version and encoding
func (bl *BinaryLoader) detectVersion() error {
	// Look for $ACADVER in the first 1024 bytes
	maxSearch := 1024
	if len(bl.data) < maxSearch {
		maxSearch = len(bl.data)
	}

	acadVer := []byte("$ACADVER")
	pos := bytes.Index(bl.data[22:maxSearch], acadVer)
	if pos != -1 {
		pos += 22 + 10 // Skip signature and variable name
		if pos < len(bl.data) {
			// Check if this is a 2-byte group code
			if bl.data[pos] != 65 { // Not 'A' = 2-byte group code
				pos++
			}
			if pos+6 <= len(bl.data) {
				bl.version = string(bl.data[pos : pos+6])
			}
		}
	}

	// Determine encoding based on version
	if bl.version >= "AC1021" { // AutoCAD 2007+
		bl.encoding = "utf8"
	} else {
		// Look for $DWGCODEPAGE for older versions
		dwgCodePage := []byte("$DWGCODEPAGE")
		pos := bytes.Index(bl.data[22:maxSearch], dwgCodePage)
		if pos != -1 {
			pos += 22 + 14 // Skip signature and variable name
			if pos < len(bl.data) {
				if bl.data[pos] != 65 { // Not 'A' = 2-byte group code
					pos++
				}
				if pos+3 <= len(bl.data) {
					codepage := string(bl.data[pos : pos+3])
					bl.encoding = bl.codepageToEncoding(codepage)
				}
			}
		}
	}

	bl.pos = 22 // Reset position after detection
	return nil
}

// codepageToEncoding converts DXF codepage to Go encoding name
func (bl *BinaryLoader) codepageToEncoding(codepage string) string {
	switch codepage {
	case "ANSI_1252":
		return "cp1252"
	case "ANSI_1250":
		return "cp1250"
	case "ANSI_1251":
		return "cp1251"
	default:
		return "cp1252" // Default
	}
}

// Next reads the next tag from the binary DXF file
func (bl *BinaryLoader) Next() (*BinaryTag, error) {
	if bl.pos >= len(bl.data) {
		return nil, io.EOF
	}

	// Read group code (2 bytes, little-endian)
	if bl.pos+2 > len(bl.data) {
		return nil, fmt.Errorf("unexpected end of file reading group code")
	}

	groupCode := int(binary.LittleEndian.Uint16(bl.data[bl.pos : bl.pos+2]))
	bl.pos += 2

	// Determine value type and read accordingly
	var value interface{}
	var err error

	switch {
	case bl.pos >= len(bl.data):
		return nil, fmt.Errorf("unexpected end of file after group code")

	// String values
	case groupCode >= 0 && groupCode <= 9:
		value, err = bl.readString()

	// Layer names, handles, etc.
	case groupCode >= 10 && groupCode <= 39:
		value, err = bl.readFloat()

	// More floats (coordinates, etc.)
	case groupCode >= 40 && groupCode <= 59:
		value, err = bl.readFloat()

	// More coordinates
	case groupCode >= 60 && groupCode <= 79:
		value, err = bl.readInt16()

	// 32-bit integers
	case groupCode >= 90 && groupCode <= 99:
		value, err = bl.readInt32()

	// Handles and other strings
	case groupCode >= 100 && groupCode <= 109:
		value, err = bl.readString()

	// More strings
	case groupCode >= 110 && groupCode <= 149:
		value, err = bl.readFloat()

	// 16-bit integers
	case groupCode >= 160 && groupCode <= 169:
		value, err = bl.readInt16()

	// More floats
	case groupCode >= 170 && groupCode <= 179:
		value, err = bl.readInt16()

	// More floats
	case groupCode >= 210 && groupCode <= 239:
		value, err = bl.readFloat()

	// More 16-bit integers
	case groupCode >= 270 && groupCode <= 289:
		value, err = bl.readInt16()

	// More 16-bit integers
	case groupCode >= 290 && groupCode <= 299:
		value, err = bl.readBoolean()

	// More 16-bit integers
	case groupCode >= 300 && groupCode <= 309:
		value, err = bl.readString()

	// More strings
	case groupCode >= 310 && groupCode <= 319:
		value, err = bl.readBinaryData()

	// More strings
	case groupCode >= 320 && groupCode <= 329:
		value, err = bl.readString()

	// More strings
	case groupCode >= 330 && groupCode <= 369:
		value, err = bl.readString()

	// 16-bit integers
	case groupCode >= 370 && groupCode <= 379:
		value, err = bl.readInt16()

	// 16-bit integers
	case groupCode >= 380 && groupCode <= 389:
		value, err = bl.readInt16()

	// 16-bit integers
	case groupCode >= 390 && groupCode <= 399:
		value, err = bl.readString()

	// 16-bit integers
	case groupCode >= 400 && groupCode <= 409:
		value, err = bl.readInt16()

	// 16-bit integers
	case groupCode >= 410 && groupCode <= 419:
		value, err = bl.readString()

	// Strings and integers
	case groupCode >= 420 && groupCode <= 429:
		if groupCode == 420 {
			value, err = bl.readInt32() // True color
		} else {
			value, err = bl.readString()
		}

	// More integers and strings
	case groupCode >= 430 && groupCode <= 439:
		value, err = bl.readString()

	// 16-bit integers
	case groupCode >= 440 && groupCode <= 449:
		value, err = bl.readInt16()

	// 32-bit integers
	case groupCode >= 450 && groupCode <= 459:
		value, err = bl.readInt32()

	// 16-bit integers
	case groupCode >= 460 && groupCode <= 469:
		value, err = bl.readInt16()

	// 32-bit integers
	case groupCode >= 470 && groupCode <= 479:
		value, err = bl.readInt32()

	// Strings
	case groupCode >= 480 && groupCode <= 481:
		value, err = bl.readString()

	// Strings
	case groupCode >= 999:
		value, err = bl.readString()

	default:
		// Default to string for unknown group codes
		value, err = bl.readString()
	}

	if err != nil {
		return nil, fmt.Errorf("error reading value for group code %d: %w", groupCode, err)
	}

	return &BinaryTag{
		Code:  groupCode,
		Value: value,
	}, nil
}

// readString reads a null-terminated UTF-16LE string
func (bl *BinaryLoader) readString() (string, error) {
	if bl.pos >= len(bl.data) {
		return "", fmt.Errorf("unexpected end of file reading string")
	}

	// Read UTF-16 characters until null terminator
	var utf16Data []uint16
	for {
		if bl.pos+2 > len(bl.data) {
			return "", fmt.Errorf("unexpected end of file in string")
		}

		char := binary.LittleEndian.Uint16(bl.data[bl.pos : bl.pos+2])
		bl.pos += 2

		if char == 0 {
			break
		}

		utf16Data = append(utf16Data, char)
	}

	// Convert UTF-16 to Go string
	runes := utf16.Decode(utf16Data)
	return string(runes), nil
}

// readFloat reads a 64-bit double precision float
func (bl *BinaryLoader) readFloat() (float64, error) {
	if bl.pos+8 > len(bl.data) {
		return 0, fmt.Errorf("unexpected end of file reading float")
	}

	bits := binary.LittleEndian.Uint64(bl.data[bl.pos : bl.pos+8])
	bl.pos += 8

	return float64frombits(bits), nil
}

// readInt16 reads a 16-bit integer
func (bl *BinaryLoader) readInt16() (int, error) {
	if bl.pos+2 > len(bl.data) {
		return 0, fmt.Errorf("unexpected end of file reading int16")
	}

	val := binary.LittleEndian.Uint16(bl.data[bl.pos : bl.pos+2])
	bl.pos += 2

	return int(val), nil
}

// readInt32 reads a 32-bit integer
func (bl *BinaryLoader) readInt32() (int, error) {
	if bl.pos+4 > len(bl.data) {
		return 0, fmt.Errorf("unexpected end of file reading int32")
	}

	val := binary.LittleEndian.Uint32(bl.data[bl.pos : bl.pos+4])
	bl.pos += 4

	return int(val), nil
}

// readBoolean reads a boolean value (stored as int16)
func (bl *BinaryLoader) readBoolean() (bool, error) {
	val, err := bl.readInt16()
	if err != nil {
		return false, err
	}
	return val != 0, nil
}

// readBinaryData reads binary data for codes 310-319
func (bl *BinaryLoader) readBinaryData() ([]byte, error) {
	if bl.pos+4 > len(bl.data) {
		return nil, fmt.Errorf("unexpected end of file reading binary data length")
	}

	// Read length as int32
	length := int(binary.LittleEndian.Uint32(bl.data[bl.pos : bl.pos+4]))
	bl.pos += 4

	if bl.pos+length > len(bl.data) {
		return nil, fmt.Errorf("unexpected end of file reading binary data")
	}

	data := make([]byte, length)
	copy(data, bl.data[bl.pos:bl.pos+length])
	bl.pos += length

	return data, nil
}

// LoadAll loads all tags from the binary DXF file
func (bl *BinaryLoader) LoadAll() ([]*BinaryTag, error) {
	var tags []*BinaryTag

	for {
		tag, err := bl.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error loading tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// Helper function to convert bits to float64 (from math package)
func float64frombits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}
