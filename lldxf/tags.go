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
	Value() any
	String() string
	Clone() Tag
}

// DXFTag represents a single DXF tag (group code, value pair)
type DXFTag struct {
	code  GroupCode
	value any
}

// SetValue replaces the tag's value.
func (t *DXFTag) SetValue(v any) { t.value = v }

// NewDXFTag creates a new DXF tag with automatic type casting
func NewDXFTag(code GroupCode, value any) Tag {
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
func (t *DXFTag) Value() any {
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
// String renders the tag for ASCII DXF output: binary payloads are hex-encoded
// uppercase, matching the wire form.
func (t *DXFBinaryTag) String() string {
	if b, ok := t.value.([]byte); ok {
		return fmt.Sprintf("%d\n%s\n", t.code, UpperHex(b))
	}
	return fmt.Sprintf("%d\n%s\n", t.code, t.value)
}

// BinaryTagFromHex builds a binary data tag (codes 260/261/310-319/1004) from
// an ASCII hex payload string (as found in DXF text streams), decoding it to
// raw bytes. Non-hex payloads are kept as their raw byte form.
func BinaryTagFromHex(code GroupCode, hexPayload string) Tag {
	data := decodeHexPayload(hexPayload)
	if data == nil {
		data = []byte(hexPayload)
	}
	return &DXFBinaryTag{DXFTag{code: code, value: data}}
}

// IsBinaryDataCode is documented on the unexported predicate.
var _ = IsBinaryDataCode

// FormatValue renders a tag value in its canonical text form for the string
// model and ASCII output: binary data payloads (codes 260/261, 310-319, 1004)
// become uppercase hex matching the wire form; other values render as %v would.
// Keeping every tag-to-string conversion on this path makes retained tags stable
// across repeated round trips.
//
// The value is converted with a type switch rather than fmt.Sprintf: %v is
// defined as 'g' with the shortest representation for floats and base-10 for
// integers, which strconv renders identically without fmt's reflection cost.
// This runs once per retained tag on every parse, so the difference is material.
func FormatValue(tag Tag) string {
	if b, ok := tag.Value().([]byte); ok {
		if _, isBinaryTag := tag.(*DXFBinaryTag); isBinaryTag || isBinaryDataCode(tag.Code()) {
			return UpperHex(b)
		}
	}
	return FormatAnyValue(tag.Value())
}

// FormatAnyValue renders a tag value exactly as fmt's %v verb would, without
// fmt's reflection cost. It is the canonical text form for tag values, shared
// by tag formatting and by the tags helpers that read a single scalar.
func FormatAnyValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(x), 'g', -1, 32)
	case int:
		return strconv.Itoa(x)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// UpperHex returns the uppercase hexadecimal encoding of b, the wire form of
// DXF binary payloads (codes 260/261, 310-319, 1004). Encoding directly avoids
// the hex.EncodeToString plus strings.ToUpper pair, which allocates twice and
// rescans the output.
func UpperHex(b []byte) string {
	const digits = "0123456789ABCDEF"
	var sb strings.Builder
	sb.Grow(len(b) * 2)
	for _, v := range b {
		sb.WriteByte(digits[v>>4])
		sb.WriteByte(digits[v&0x0f])
	}
	return sb.String()
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
func (t Tags) GetFirstValue(code GroupCode) (any, bool) {
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
		return FormatAnyValue(handle)
	}
	if handle, found := t.GetFirstValue(105); found {
		return FormatAnyValue(handle)
	}
	return ""
}

// DXFType returns the entity type (group code 0)
func (t Tags) DXFType() string {
	if dtype, found := t.GetFirstValue(0); found {
		return FormatAnyValue(dtype)
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
func castTagValue(code GroupCode, value any) any {
	// Point codes carry floats, but ACIS payloads (3DSOLID/REGION/BODY) reuse
	// them for text lines. Preserve strings that do not parse as numbers so
	// the payload survives round trips; numeric parsing stays unchanged.
	if isPointCode(code) || (code >= 10 && code <= 17) {
		if str, ok := value.(string); ok && str != "" {
			if _, err := strconv.ParseFloat(str, 64); err != nil {
				return str
			}
		}
	}
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

// IsBinaryDataCode reports group codes carrying hex-encoded binary payloads
// (310-319, 1004).
func IsBinaryDataCode(code GroupCode) bool { return isBinaryDataCode(code) }

func isBinaryDataCode(code GroupCode) bool {
	// Codes 260/261 (binary flags) and 310-319 (binary chunks), plus the
	// XDATA binary chunk 1004.
	if code >= 310 && code <= 319 {
		return true
	}
	return code == 260 || code == 261 || code == 1004
}

func isPointCode(code GroupCode) bool {
	// Coordinate tags: primary X/Y/Z series and their extrusion counterparts.
	if code >= 10 && code <= 18 {
		return true
	}
	if code >= 110 && code <= 112 {
		return true
	}
	if code >= 210 && code <= 213 {
		return true
	}
	return code >= 1010 && code <= 1013
}

func isFloatCode(code GroupCode) bool {
	// Group codes carrying floating-point values, per the DXF specification:
	// coordinates (10-59, 110-149), extrusion vectors (210-219), and the
	// remaining float ranges (460-469, 1010-1013, 1040-1042).
	switch {
	case code >= 10 && code <= 59:
		return true
	case code >= 110 && code <= 149:
		return true
	case code >= 210 && code <= 219:
		return true
	case code >= 460 && code <= 469:
		return true
	case code >= 1010 && code <= 1013:
		return true
	case code >= 1040 && code <= 1042:
		return true
	}
	return false
}

func isInt16Code(code GroupCode) bool {
	// Group codes carrying 16-bit integers: flags, counts, ACI colors and
	// booleans, including the XDATA 16-bit integer range.
	switch {
	case code >= 60 && code <= 79:
		return true
	case code >= 170 && code <= 179:
		return true
	case code >= 270 && code <= 299:
		return true
	case code >= 370 && code <= 389:
		return true
	case code >= 400 && code <= 409:
		return true
	case code >= 1060 && code <= 1070:
		return true
	}
	return false
}

func isInt32Code(code GroupCode) bool {
	// Group codes carrying 32-bit integers: counts, handles and true colors,
	// plus the XDATA 32-bit integer code.
	switch {
	case code >= 90 && code <= 99:
		return true
	case code >= 420 && code <= 459:
		return true
	case code >= 480 && code <= 489:
		return true
	}
	return code == 1071
}

func isInt64Code(code GroupCode) bool {
	// Group codes carrying 64-bit integers (lengths and counts in binary DXF).
	return code >= 160 && code <= 169
}

// Type casting functions

func castToFloat(value any) float64 {
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

func castToInt16(value any) int16 {
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

func castToInt32(value any) int32 {
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

func castToInt64(value any) int64 {
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

func castToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// decodeHexPayload converts an ASCII DXF binary payload (hex digits, optional
// spaces) to raw bytes. Non-hex input is returned as nil so callers can fall
// back to the string form.
func decodeHexPayload(s string) []byte {
	// Decode in one pass instead of first copying the digits without spaces:
	// the common payload has no separators, so the copy was pure overhead.
	out := make([]byte, 0, len(s)/2)
	hi := -1
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' {
			continue
		}
		d, ok := hexDigit(c)
		if !ok {
			return nil
		}
		if hi < 0 {
			hi = int(d)
			continue
		}
		out = append(out, byte(hi<<4|int(d)))
		hi = -1
	}
	if hi >= 0 {
		return nil // odd number of hex digits
	}
	return out
}

func hexDigit(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

func castToBinary(value any) []byte {
	switch v := value.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}

// BinaryValueKind classifies how a group code's value is encoded in binary DXF.
type BinaryValueKind int

const (
	BinaryString BinaryValueKind = iota // null-terminated UTF-16LE string
	BinaryFloat                         // 64-bit double
	BinaryInt16                         // 16-bit little-endian integer
	BinaryInt32                         // 32-bit little-endian integer
	BinaryInt64                         // 64-bit little-endian integer
	BinaryRaw                           // int32 byte count followed by raw bytes
)

// BinaryKindFor returns the binary encoding kind for a group code, matching
// castTagValue: the value type a tag holds after casting is determined here.
func BinaryKindFor(code GroupCode) BinaryValueKind {
	switch {
	case isBinaryDataCode(code):
		return BinaryRaw
	case isFloatCode(code):
		return BinaryFloat
	case isInt64Code(code):
		return BinaryInt64
	case isInt32Code(code):
		return BinaryInt32
	case isInt16Code(code):
		return BinaryInt16
	case IsBinaryBoolCode(code):
		return BinaryInt16
	default:
		return BinaryString
	}
}

// IsBinaryBoolCode reports whether a code stores a boolean (as int16) in binary DXF.
func IsBinaryBoolCode(code GroupCode) bool {
	return code >= 290 && code <= 299
}
