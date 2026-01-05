package lldxf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Encoding constants
const (
	DefaultEncoding = "utf8"

	// Common code pages
	CodePageANSI = "1252"
	CodePageUTF8 = "utf8"

	// MIF encoded character patterns
	MIFPattern     = `\\M\+[1-5][A-F0-9]{4}`
	UnicodePattern = `\\U\+[A-F0-9]{4}`
)

// EncodingInfo contains information about text encoding in DXF
type EncodingInfo struct {
	Name      string
	IsUnicode bool
	IsMIF     bool
}

// DetectEncoding attempts to detect the text encoding from DXF header variables
func DetectEncoding(headerVars map[string]string) (string, float64) {
	// Check $DWGCODEPAGE variable first
	if codepage, exists := headerVars["$DWGCODEPAGE"]; exists {
		return codepage, 0.8
	}

	// Check $ACADVER variable for version-based detection
	if acadver, exists := headerVars["$ACADVER"]; exists {
		// DXF R2007+ uses UTF-8 by default
		if strings.Contains(acadver, "AC1021") || strings.Contains(acadver, "AC1024") {
			return CodePageUTF8, 0.9
		}
		// Earlier versions often use ANSI code pages
		return CodePageANSI, 0.7
	}

	// Default to UTF-8 if no clear indication
	return CodePageUTF8, 0.5
}

// EncodingDetector provides encoding detection and conversion for DXF text
type EncodingDetector struct {
	encoding    string
	confidence  float64
	hasMIF      bool
	unicodeOnly bool
}

// NewEncodingDetector creates a new encoding detector
func NewEncodingDetector() *EncodingDetector {
	return &EncodingDetector{
		encoding:    CodePageUTF8,
		confidence:  0.5,
		hasMIF:      false,
		unicodeOnly: true,
	}
}

// DetectFromHeader analyzes DXF header variables to determine encoding
func (ed *EncodingDetector) DetectFromHeader(headerVars map[string]string) {
	encoding, confidence := DetectEncoding(headerVars)
	ed.encoding = encoding
	ed.confidence = confidence

	// Check for MIF encoding in header
	for _, value := range headerVars {
		if hasMIFEncoding(value) {
			ed.hasMIF = true
			ed.unicodeOnly = false
			break
		}
	}
}

// hasMIFEncoding checks if a string contains MIF encoded characters
func hasMIFEncoding(s string) bool {
	matched, _ := regexp.MatchString(MIFPattern, s)
	return matched
}

// hasUnicodeEncoding checks if a string contains Unicode escape sequences
func hasUnicodeEncoding(s string) bool {
	matched, _ := regexp.MatchString(UnicodePattern, s)
	return matched
}

// getEncodingByName returns a simple decoder for supported encodings
// For now, we only support UTF-8 and basic ANSI (Windows-1252) conversion
func getEncodingByName(encoding string) interface{} {
	switch encoding {
	case CodePageUTF8:
		return nil // UTF-8 needs no conversion
	case CodePageANSI:
		// For basic ANSI support, we'd need a more sophisticated approach
		// For now, return nil to pass through unchanged
		return nil
	default:
		return nil
	}
}

// ConvertToUTF8 converts text from detected encoding to UTF-8
func (ed *EncodingDetector) ConvertToUTF8(input string) (string, error) {
	if ed.encoding == CodePageUTF8 || ed.unicodeOnly {
		return input, nil
	}

	// For now, only basic support - in a full implementation would need
	// proper encoding conversion libraries
	switch ed.encoding {
	case CodePageANSI:
		// Simple ANSI to UTF-8 approximation for basic Latin characters
		// A full implementation would use golang.org/x/text/encoding
		return input, nil
	default:
		return input, fmt.Errorf("unsupported encoding conversion from %s", ed.encoding)
	}
}

// ConvertToEncoding converts UTF-8 text to target encoding for DXF output
func (ed *EncodingDetector) ConvertToEncoding(input string) ([]byte, error) {
	if ed.encoding == CodePageUTF8 || ed.unicodeOnly {
		return []byte(input), nil
	}

	// For now, only basic support
	switch ed.encoding {
	case CodePageANSI:
		// Simple UTF-8 to ANSI approximation
		// A full implementation would use golang.org/x/text/encoding
		return []byte(input), nil
	default:
		return []byte(input), fmt.Errorf("unsupported encoding conversion to %s", ed.encoding)
	}
}

// processDXFUnicodeEscapes processes DXF-specific Unicode escape sequences
func processDXFUnicodeEscapes(input string) string {
	// Handle MIF encoding
	if hasMIFEncoding(input) {
		input = decodeMIFSequences(input)
	}

	// Handle Unicode escapes
	input = decodeUnicodeSequences(input)

	// Replace problematic DXF characters
	return strings.ReplaceAll(input, "\x00", "")
}

// decodeMIFSequences decodes MIF \M+xxxx sequences
func decodeMIFSequences(input string) string {
	mifRe := regexp.MustCompile(MIFPattern)

	return mifRe.ReplaceAllStringFunc(input, func(match string) string {
		if len(match) < 5 {
			return match
		}

		// Extract hex code
		hexStr := match[3:]
		if code, err := strconv.ParseInt(hexStr, 16, 64); err == nil {
			return string(rune(code))
		}

		return match
	})
}

// decodeUnicodeSequences decodes \U+xxxx escape sequences
func decodeUnicodeSequences(input string) string {
	unicodeRe := regexp.MustCompile(UnicodePattern)

	return unicodeRe.ReplaceAllStringFunc(input, func(match string) string {
		if len(match) < 3 {
			return match
		}

		// Extract hex code
		hexStr := match[2:]
		if code, err := strconv.ParseInt(hexStr, 16, 64); err == nil {
			return string(rune(code))
		}

		return match
	})
}

// GetEncodingName returns the detected encoding name
func (ed *EncodingDetector) GetEncodingName() string {
	return ed.encoding
}

// SetEncoding forces a specific encoding
func (ed *EncodingDetector) SetEncoding(encoding string) error {
	switch encoding {
	case CodePageUTF8, CodePageANSI:
		ed.encoding = encoding
		ed.confidence = 0.8
		return nil
	default:
		return fmt.Errorf("unsupported encoding: %s", encoding)
	}
}

// GetConfidence returns the detection confidence
func (ed *EncodingDetector) GetConfidence() float64 {
	return ed.confidence
}

// IsUnicode returns true if using Unicode encoding
func (ed *EncodingDetector) IsUnicode() bool {
	return ed.unicodeOnly
}

// HasMIF returns true if MIF encoding was detected
func (ed *EncodingDetector) HasMIF() bool {
	return ed.hasMIF
}

// ValidateEncoding checks if an encoding is valid for DXF
func ValidateEncoding(encoding string) error {
	if encoding == "" {
		return fmt.Errorf("empty encoding")
	}

	// Check if encoding is supported
	supportedEncodings := []string{CodePageUTF8, CodePageANSI}

	for _, supported := range supportedEncodings {
		if strings.EqualFold(encoding, supported) {
			return nil
		}
	}

	return fmt.Errorf("unsupported DXF encoding: %s", encoding)
}

// String returns string representation
func (ed *EncodingDetector) String() string {
	return fmt.Sprintf("EncodingDetector{encoding=%s, confidence=%.2f, hasMIF=%v}",
		ed.encoding, ed.confidence, ed.hasMIF)
}
