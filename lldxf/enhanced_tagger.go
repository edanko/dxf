package lldxf

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// TagCompiler handles coordinate assembly and value compilation
type TagCompiler struct {
	lookaheadTags []Tag
	currentTag    Tag
	err           error
	source        Tagger
}

// NewTagCompiler creates a new tag compiler
func NewTagCompiler(source Tagger) *TagCompiler {
	return &TagCompiler{
		source: source,
	}
}

// Next advances to the next compiled tag
func (tc *TagCompiler) Next() bool {
	if tc.err != nil {
		return false
	}

	// Return any buffered tags first
	if len(tc.lookaheadTags) > 0 {
		tc.currentTag = tc.lookaheadTags[0]
		tc.lookaheadTags = tc.lookaheadTags[1:]
		return true
	}

	if !tc.source.Next() {
		tc.err = tc.source.Err()
		return false
	}

	tag := tc.source.Tag()
	code := tag.Code()

	// Handle coordinate assembly for point codes
	if isPrimaryPointCode(code) {
		return tc.assemblePoint(tag)
	}

	tc.currentTag = tag
	return true
}

// assemblePoint assembles 2D/3D coordinates from consecutive point tags
func (tc *TagCompiler) assemblePoint(xTag Tag) bool {
	primaryCode := xTag.Code()

	// Get Y coordinate (mandatory)
	if !tc.source.Next() {
		tc.err = fmt.Errorf("missing Y coordinate for point code %d", primaryCode)
		return false
	}

	yTag := tc.source.Tag()
	expectedYCode := primaryCode + 10
	if yTag.Code() != expectedYCode {
		tc.err = fmt.Errorf("expected Y coordinate code %d, got %d", expectedYCode, yTag.Code())
		return false
	}

	// Get Z coordinate (optional)
	var zTag Tag = nil
	if tc.source.Next() {
		zTag = tc.source.Tag()
		expectedZCode := primaryCode + 20
		if zTag.Code() != expectedZCode {
			// Not a Z coordinate, put it back
			tc.lookaheadTags = append(tc.lookaheadTags, zTag)
			zTag = nil
		}
	}

	// Create vertex tag
	var coords []float64
	coords = append(coords, castToFloat(xTag.Value()))
	coords = append(coords, castToFloat(yTag.Value()))
	if zTag != nil {
		coords = append(coords, castToFloat(zTag.Value()))
	}

	tc.currentTag = NewDXFTag(primaryCode, coords)
	return true
}

// Tag returns the current compiled tag
func (tc *TagCompiler) Tag() Tag {
	return tc.currentTag
}

// Err returns any error that occurred
func (tc *TagCompiler) Err() error {
	return tc.err
}

// isPrimaryPointCode checks if a code is a primary coordinate code (X coordinate)
func isPrimaryPointCode(code GroupCode) bool {
	// Primary coordinate codes: 10-19, 110-119, 210-219, 1010-1019, etc.
	return (code >= 10 && code <= 19) ||
		(code >= 110 && code <= 119) ||
		(code >= 210 && code <= 219) ||
		(code >= 1010 && code <= 1019) ||
		(code >= 220 && code <= 229) ||
		(code >= 330 && code <= 339)
}

// BinaryTagger implements Tagger for binary DXF files
type BinaryTagger struct {
	reader io.Reader
	tag    Tag
	err    error
}

// NewBinaryTagger creates a new binary DXF tagger
func NewBinaryTagger(r io.Reader) Tagger {
	return &BinaryTagger{reader: r}
}

// Next advances to the next binary tag
func (bt *BinaryTagger) Next() bool {
	if bt.err != nil {
		return false
	}

	// Read group code (2 bytes, little-endian)
	var code int16
	if err := binary.Read(bt.reader, binary.LittleEndian, &code); err != nil {
		if err == io.EOF {
			return false
		}
		bt.err = fmt.Errorf("failed to read group code: %w", err)
		return false
	}

	// Read value based on group code
	var value interface{}

	switch {
	case isInt16Code(GroupCode(code)):
		var val int16
		if err := binary.Read(bt.reader, binary.LittleEndian, &val); err != nil {
			bt.err = fmt.Errorf("failed to read int16 value: %w", err)
			return false
		}
		value = val

	case isInt32Code(GroupCode(code)):
		var val int32
		if err := binary.Read(bt.reader, binary.LittleEndian, &val); err != nil {
			bt.err = fmt.Errorf("failed to read int32 value: %w", err)
			return false
		}
		value = val

	case isInt64Code(GroupCode(code)):
		var val int64
		if err := binary.Read(bt.reader, binary.LittleEndian, &val); err != nil {
			bt.err = fmt.Errorf("failed to read int64 value: %w", err)
			return false
		}
		value = val

	case isFloatCode(GroupCode(code)):
		var val float64
		if err := binary.Read(bt.reader, binary.LittleEndian, &val); err != nil {
			bt.err = fmt.Errorf("failed to read float64 value: %w", err)
			return false
		}
		value = val

	case isBinaryDataCode(GroupCode(code)):
		// Read byte length first
		var byteLength int32
		if err := binary.Read(bt.reader, binary.LittleEndian, &byteLength); err != nil {
			bt.err = fmt.Errorf("failed to read binary data length: %w", err)
			return false
		}

		// Read binary data
		data := make([]byte, byteLength)
		if _, err := io.ReadFull(bt.reader, data); err != nil {
			bt.err = fmt.Errorf("failed to read binary data: %w", err)
			return false
		}
		value = data

	default:
		// String value - read null-terminated string
		var builder strings.Builder
		for {
			var b byte
			if err := binary.Read(bt.reader, binary.LittleEndian, &b); err != nil {
				bt.err = fmt.Errorf("failed to read string byte: %w", err)
				return false
			}
			if b == 0 {
				break
			}
			builder.WriteByte(b)

			// Prevent infinite loops
			if builder.Len() > 2048 {
				bt.err = fmt.Errorf("string value too long")
				return false
			}
		}
		value = builder.String()
	}

	bt.tag = NewDXFTag(GroupCode(code), value)
	return true
}

// Tag returns the current tag
func (bt *BinaryTagger) Tag() Tag {
	return bt.tag
}

// Err returns any error that occurred
func (bt *BinaryTagger) Err() error {
	return bt.err
}

// RobustTagger provides enhanced error handling and comment support
type RobustTagger struct {
	tagger     Tagger
	currentTag Tag
	err        error
	options    *TaggerOptions
}

// TaggerOptions configures tagger behavior
type TaggerOptions struct {
	SkipComments  bool
	IgnoreErrors  bool
	MaxTags       int
	ValidateCodes bool
}

// DefaultTaggerOptions returns default tagger options
func DefaultTaggerOptions() *TaggerOptions {
	return &TaggerOptions{
		SkipComments:  true,
		IgnoreErrors:  false,
		MaxTags:       1000000, // Safety limit
		ValidateCodes: true,
	}
}

// NewRobustTagger creates a new robust ASCII tagger
func NewRobustTagger(r io.Reader, options *TaggerOptions) Tagger {
	if options == nil {
		options = DefaultTaggerOptions()
	}

	return &RobustTagger{
		tagger:  NewASCIITagger(r),
		options: options,
	}
}

// Next advances to the next tag with error handling
func (rt *RobustTagger) Next() bool {
	if rt.err != nil {
		return false
	}

	tagCount := 0

	for {
		if !rt.tagger.Next() {
			rt.err = rt.tagger.Err()
			return false
		}

		tag := rt.tagger.Tag()
		tagCount++

		// Safety check
		if tagCount > rt.options.MaxTags {
			rt.err = fmt.Errorf("exceeded maximum tag count of %d", rt.options.MaxTags)
			return false
		}

		// Skip comments (group code 999)
		if rt.options.SkipComments && tag.Code() == 999 {
			continue
		}

		// Validate group codes
		if rt.options.ValidateCodes {
			if err := rt.validateTag(tag); err != nil {
				if rt.options.IgnoreErrors {
					continue
				}
				rt.err = err
				return false
			}
		}

		rt.currentTag = tag
		return true
	}
}

// validateTag performs basic tag validation
func (rt *RobustTagger) validateTag(tag Tag) error {
	code := int(tag.Code())

	// Check for valid group code range
	if code < 0 || code > 1071 {
		return fmt.Errorf("invalid group code: %d", code)
	}

	// Check for valid string values
	if str, ok := tag.Value().(string); ok {
		// Check for non-printable characters (except for certain codes)
		if !isValidDXFString(str, code) {
			return fmt.Errorf("invalid string value for group code %d", code)
		}
	}

	return nil
}

// isValidDXFString checks if a string is valid for DXF
func isValidDXFString(s string, groupCode int) bool {
	// Binary data codes can contain any bytes
	if isBinaryDataCode(GroupCode(groupCode)) {
		return true
	}

	// Check for valid UTF-8
	if !utf8.ValidString(s) {
		return false
	}

	// Check for control characters (except specific cases)
	for _, r := range s {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}

	return true
}

// Tag returns the current tag
func (rt *RobustTagger) Tag() Tag {
	return rt.currentTag
}

// Err returns any error that occurred
func (rt *RobustTagger) Err() error {
	return rt.err
}

// LoadTagsWithCompiler loads and compiles tags from a reader
func LoadTagsWithCompiler(r io.Reader) (Tags, error) {
	tagger := NewASCIITagger(r)
	compiler := NewTagCompiler(tagger)

	var tags Tags

	for compiler.Next() {
		tags = tags.Add(compiler.Tag())
	}

	if err := compiler.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// LoadTagsRobust loads tags with enhanced error handling
func LoadTagsRobust(r io.Reader, options *TaggerOptions) (Tags, error) {
	tagger := NewRobustTagger(r, options)

	var tags Tags

	for tagger.Next() {
		tags = tags.Add(tagger.Tag())
	}

	if err := tagger.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// DetectDXFFormat detects whether a file is ASCII or binary DXF
func DetectDXFFormat(data []byte) (string, error) {
	if len(data) < 22 {
		return "", fmt.Errorf("insufficient data to detect format")
	}

	// Check for binary DXF signature (AutoCAD Binary DXF\r\n\x1a\x00)
	signature := "AutoCAD Binary DXF\r\n\x1a\x00"
	if len(data) >= 22 && string(data[:22]) == signature {
		return "binary", nil
	}

	// Check for common ASCII DXF patterns
	content := string(data[:min(len(data), 1000)])

	// Look for section markers
	if strings.Contains(content, "0\nSECTION") ||
		strings.Contains(content, " 0\nSECTION") ||
		strings.Contains(content, "0\r\nSECTION") {
		return "ascii", nil
	}

	return "", fmt.Errorf("unable to detect DXF format")
}

// CreateTagger creates appropriate tagger based on format
func CreateTagger(r io.Reader, format string) Tagger {
	switch format {
	case "binary":
		return NewBinaryTagger(r)
	case "ascii":
		return NewASCIITagger(r)
	default:
		return NewASCIITagger(r) // Default to ASCII
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
