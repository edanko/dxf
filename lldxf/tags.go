package lldxf

import (
	"fmt"
	"strconv"
	"strings"
)

// GroupCode represents a DXF group code
type GroupCode int

// Tag interface for all DXF tag types
type Tag interface {
	Code() GroupCode
	Value() interface{}
	String() string
	Clone() Tag
}

// DXFTag represents a single DXF tag (group code, value pair)
type DXFTag struct {
	code  GroupCode
	value interface{}
}

// NewDXFTag creates a new DXF tag with automatic type casting
func NewDXFTag(code GroupCode, value interface{}) Tag {
	// Special case: if value is already a coordinate array and code is a point code, preserve it
	if coords, ok := value.([]float64); ok && isPointCode(code) {
		return &DXFVertex{DXFTag{code: code, value: coords}}
	}

	castValue := castTagValue(code, value)

	// Create specialized tag types based on group code
	switch {
	case isBinaryDataCode(code):
		return &DXFBinaryTag{DXFTag{code: code, value: castValue}}
	case isPointCode(code):
		return &DXFVertex{DXFTag{code: code, value: castValue}}
	default:
		return &DXFTag{code: code, value: castValue}
	}
}

// Code returns the group code
func (t *DXFTag) Code() GroupCode {
	return t.code
}

// Value returns the tag value
func (t *DXFTag) Value() interface{} {
	return t.value
}

// String returns the DXF string representation
func (t *DXFTag) String() string {
	var valueStr string
	switch v := t.value.(type) {
	case float64:
		valueStr = strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		valueStr = strconv.FormatFloat(float64(v), 'f', -1, 64)
	case string:
		valueStr = v
	case []byte:
		valueStr = string(v)
	default:
		valueStr = fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%d\n%s\n", t.code, valueStr)
}

// Clone creates a deep copy of the tag
func (t *DXFTag) Clone() Tag {
	return &DXFTag{code: t.code, value: t.value}
}

// DXFBinaryTag represents binary data tags (310-319, 1004)
type DXFBinaryTag struct {
	DXFTag
}

// String returns the binary tag string representation
func (t *DXFBinaryTag) String() string {
	return fmt.Sprintf("%d\n%s\n", t.code, t.value)
}

// Clone creates a deep copy of the binary tag
func (t *DXFBinaryTag) Clone() Tag {
	return &DXFBinaryTag{DXFTag{code: t.code, value: t.value}}
}

// DXFVertex represents coordinate tags (10-19, 110-119, 210-219, etc.)
type DXFVertex struct {
	DXFTag
}

// String returns the vertex tag string representation
func (t *DXFVertex) String() string {
	var valueStr string
	switch v := t.value.(type) {
	case float64:
		valueStr = strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		valueStr = strconv.FormatFloat(float64(v), 'f', -1, 64)
	case string:
		valueStr = v
	case []byte:
		valueStr = string(v)
	default:
		valueStr = fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%d\n%s\n", t.code, valueStr)
}

// Clone creates a deep copy of the vertex tag
func (t *DXFVertex) Clone() Tag {
	return &DXFVertex{DXFTag{code: t.code, value: t.value}}
}

// Tags represents a collection of DXF tags
type Tags []Tag

// NewTags creates an empty Tags collection
func NewTags() Tags {
	return make(Tags, 0)
}

// Add appends a tag to the collection
func (t Tags) Add(tag Tag) Tags {
	return append(t, tag)
}

// GetFirstValue returns the value of the first tag with the specified group code
func (t Tags) GetFirstValue(code GroupCode) (interface{}, bool) {
	for _, tag := range t {
		if tag.Code() == code {
			return tag.Value(), true
		}
	}
	return nil, false
}

// FindAll returns all tags with the specified group code
func (t Tags) FindAll(code GroupCode) Tags {
	result := make(Tags, 0)
	for _, tag := range t {
		if tag.Code() == code {
			result = append(result, tag)
		}
	}
	return result
}

// SetFirst updates the first tag with the specified group code, or adds it if not found
func (t Tags) SetFirst(tag Tag) {
	for i, existing := range t {
		if existing.Code() == tag.Code() {
			t[i] = tag
			return
		}
	}
	t.Add(tag)
}

// GetHandle returns the entity handle (group code 5 or 105)
func (t Tags) GetHandle() string {
	if handle, found := t.GetFirstValue(5); found {
		return fmt.Sprintf("%v", handle)
	}
	if handle, found := t.GetFirstValue(105); found {
		return fmt.Sprintf("%v", handle)
	}
	return ""
}

// DXFType returns the entity type (group code 0)
func (t Tags) DXFType() string {
	if dtype, found := t.GetFirstValue(0); found {
		return fmt.Sprintf("%v", dtype)
	}
	return ""
}

// String returns the string representation of all tags
func (t Tags) String() string {
	var builder strings.Builder
	for _, tag := range t {
		builder.WriteString(tag.String())
	}
	return builder.String()
}

// castTagValue converts tag values to appropriate types based on group code
func castTagValue(code GroupCode, value interface{}) interface{} {
	switch {
	case isFloatCode(code):
		return castToFloat(value)
	case isInt16Code(code):
		return castToInt16(value)
	case isInt32Code(code):
		return castToInt32(value)
	case isInt64Code(code):
		return castToInt64(value)
	case isBinaryDataCode(code):
		return castToBinary(value)
	default:
		return castToString(value)
	}
}

// Helper functions for group code classification

func isBinaryDataCode(code GroupCode) bool {
	binaryCodes := []GroupCode{310, 311, 312, 313, 314, 315, 316, 317, 318, 319, 1004}
	for _, c := range binaryCodes {
		if code == c {
			return true
		}
	}
	return false
}

func isPointCode(code GroupCode) bool {
	pointCodes := []GroupCode{10, 11, 12, 13, 14, 15, 16, 17, 18, 110, 111, 112, 210, 211, 212, 213, 1010, 1011, 1012, 1013}
	for _, c := range pointCodes {
		if code == c {
			return true
		}
	}
	return false
}

func isFloatCode(code GroupCode) bool {
	// Group codes that typically contain floating-point values
	// Based on DXF specification
	floatCodes := []GroupCode{
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19, // X coordinates
		20, 21, 22, 23, 24, 25, 26, 27, 28, 29, // Y coordinates
		30, 31, 32, 33, 34, 35, 36, 37, 38, 39, // Z coordinates
		40, 41, 42, 43, 44, 45, 46, 47, 48, 49, // Radii, angles, etc.
		50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
		110, 111, 112, 113, 114, 115, 116, 117, 118, 119, // More X coords
		120, 121, 122, 123, 124, 125, 126, 127, 128, 129, // More Y coords
		130, 131, 132, 133, 134, 135, 136, 137, 138, 139, // More Z coords
		140, 141, 142, 143, 144, 145, 146, 147, 148, 149, // Dimension floats
		210, 211, 212, 213, 214, 215, 216, 217, 218, 219, // Extrusion vectors
		460, 461, 462, 463, 464, 465, 466, 467, 468, 469, // More floats
		1010, 1011, 1012, 1013, // XDATA floats
		1040, 1041, 1042, // XDATA floats
	}
	for _, c := range floatCodes {
		if code == c {
			return true
		}
	}
	return false
}

func isInt16Code(code GroupCode) bool {
	// Group codes that typically contain 16-bit integers
	// Based on DXF specification
	int16Codes := []GroupCode{
		60, 61, 62, 63, 64, 65, 66, 67, 68, 69, // Status, color, etc.
		70, 71, 72, 73, 74, 75, 76, 77, 78, 79, // Flags, counts
		170, 171, 172, 173, 174, 175, 176, 177, 178, 179, // More flags
		270, 271, 272, 273, 274, 275, 276, 277, 278, 279, // More flags
		280, 281, 282, 283, 284, 285, 286, 287, 288, 289, // Booleans, small ints
		290, 291, 292, 293, 294, 295, 296, 297, 298, 299, // Booleans
		370, 371, 372, 373, 374, 375, 376, 377, 378, 379, // More flags
		380, 381, 382, 383, 384, 385, 386, 387, 388, 389, // More small ints
		400, 401, 402, 403, 404, 405, 406, 407, 408, 409, // More flags
		1060, 1061, 1062, 1063, 1064, 1065, 1066, 1067, 1068, 1069, // XDATA ints
		1070, // XDATA int
	}
	for _, c := range int16Codes {
		if code == c {
			return true
		}
	}
	return false
}

func isInt32Code(code GroupCode) bool {
	// Group codes that typically contain 32-bit integers
	// Based on DXF specification
	int32Codes := []GroupCode{
		90, 91, 92, 93, 94, 95, 96, 97, 98, 99, // Counts, handles
		100, 101, 102, 103, 104, // String codes (actually stored as string)
		105,                // Handle
		106, 107, 108, 109, // String codes
		420, 421, 422, 423, 424, 425, 426, 427, 428, 429, // Annotation related
		430, 431, 432, 433, 434, 435, 436, 437, 438, 439, // Dictionary handles
		440, 441, 442, 443, 444, 445, 446, 447, 448, 449, // More ints
		450, 451, 452, 453, 454, 455, 456, 457, 458, 459, // More ints
		480, 481, 482, 483, 484, 485, 486, 487, 488, 489, // More handles
		1071, // XDATA int32
	}
	for _, c := range int32Codes {
		if code == c {
			return true
		}
	}
	return false
}

func isInt64Code(code GroupCode) bool {
	// Group codes that typically contain 64-bit integers
	// Based on DXF specification
	int64Codes := []GroupCode{
		160, 161, 162, 163, 164, 165, 166, 167, 168, 169, // Lengths, counts
		170, 171, 172, 173, 174, 175, 176, 177, 178, 179, // Moved to int16
	}
	for _, c := range int64Codes {
		if code == c {
			return true
		}
	}
	return false
}

// Type casting functions

func castToFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return 0.0
	default:
		return 0.0
	}
}

func castToInt16(value interface{}) int16 {
	switch v := value.(type) {
	case int16:
		return v
	case int:
		return int16(v)
	case int64:
		return int16(v)
	case float64:
		return int16(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 16); err == nil {
			return int16(i)
		}
		return 0
	default:
		return 0
	}
}

func castToInt32(value interface{}) int32 {
	switch v := value.(type) {
	case int32:
		return v
	case int:
		return int32(v)
	case int64:
		return int32(v)
	case float64:
		return int32(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int32(i)
		}
		return 0
	default:
		return 0
	}
}

func castToInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
		return 0
	default:
		return 0
	}
}

func castToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func castToBinary(value interface{}) []byte {
	switch v := value.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}
