package lldxf

import (
	"strings"
	"testing"
)

func TestDetectEncodingWithDWGCodePage(t *testing.T) {
	headerVars := map[string]string{
		"$DWGCODEPAGE": "ANSI_1252",
	}

	encoding, confidence := DetectEncoding(headerVars)

	if encoding != "ANSI_1252" {
		t.Errorf("Expected encoding 'ANSI_1252', got '%s'", encoding)
	}

	if confidence != 0.8 {
		t.Errorf("Expected confidence 0.8, got %f", confidence)
	}
}

func TestDetectEncodingWithR2007Version(t *testing.T) {
	testCases := []struct {
		name     string
		acadver  string
		expected string
		conf     float64
	}{
		{"R2007 AC1021", "AC1021", CodePageUTF8, 0.9},
		{"R2010 AC1024", "AC1024", CodePageUTF8, 0.9},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			headerVars := map[string]string{
				"$ACADVER": tc.acadver,
			}

			encoding, confidence := DetectEncoding(headerVars)

			if encoding != tc.expected {
				t.Errorf("Expected encoding '%s', got '%s'", tc.expected, encoding)
			}

			if confidence != tc.conf {
				t.Errorf("Expected confidence %f, got %f", tc.conf, confidence)
			}
		})
	}
}

func TestDetectEncodingOlderVersion(t *testing.T) {
	headerVars := map[string]string{
		"$ACADVER": "AC1009",
	}

	encoding, confidence := DetectEncoding(headerVars)

	if encoding != CodePageANSI {
		t.Errorf("Expected encoding '%s', got '%s'", CodePageANSI, encoding)
	}

	if confidence != 0.7 {
		t.Errorf("Expected confidence 0.7, got %f", confidence)
	}
}

func TestDetectEncodingDefault(t *testing.T) {
	headerVars := map[string]string{}

	encoding, confidence := DetectEncoding(headerVars)

	if encoding != CodePageUTF8 {
		t.Errorf("Expected encoding '%s', got '%s'", CodePageUTF8, encoding)
	}

	if confidence != 0.5 {
		t.Errorf("Expected confidence 0.5, got %f", confidence)
	}
}

func TestDetectEncodingPriority(t *testing.T) {
	headerVars := map[string]string{
		"$DWGCODEPAGE": "utf8",
		"$ACADVER":     "AC1009",
	}

	encoding, _ := DetectEncoding(headerVars)

	if encoding != "utf8" {
		t.Errorf("Expected DWGCODEPAGE to take priority, got '%s'", encoding)
	}
}

func TestNewEncodingDetector(t *testing.T) {
	detector := NewEncodingDetector()

	if detector.encoding != CodePageUTF8 {
		t.Errorf("Expected default encoding '%s', got '%s'", CodePageUTF8, detector.encoding)
	}

	if detector.confidence != 0.5 {
		t.Errorf("Expected default confidence 0.5, got %f", detector.confidence)
	}

	if detector.hasMIF {
		t.Error("Expected hasMIF to be false by default")
	}

	if !detector.unicodeOnly {
		t.Error("Expected unicodeOnly to be true by default")
	}
}

func TestEncodingDetectorDetectFromHeader(t *testing.T) {
	detector := NewEncodingDetector()

	headerVars := map[string]string{
		"$ACADVER":     "AC1021",
		"$DWGCODEPAGE": "utf8",
	}

	detector.DetectFromHeader(headerVars)

	if detector.encoding != "utf8" {
		t.Errorf("Expected encoding 'utf8', got '%s'", detector.encoding)
	}

	if detector.confidence != 0.8 {
		t.Errorf("Expected confidence 0.8, got %f", detector.confidence)
	}
}

func TestEncodingDetectorDetectFromHeaderMIF(t *testing.T) {
	detector := NewEncodingDetector()

	headerVars := map[string]string{
		"$ACADVER": "AC1009",
		"TEST_VAR": "\\M+10041",
	}

	detector.DetectFromHeader(headerVars)

	if !detector.hasMIF {
		t.Error("Expected hasMIF to be true after detecting MIF encoding")
	}

	if detector.unicodeOnly {
		t.Error("Expected unicodeOnly to be false when MIF is detected")
	}
}

func TestHasMIFEncoding(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{`\M+10041`, true},
		{`\M+200C0`, true},
		{`\M+3FFFF`, true},
		{"normal text", false},
		{`\U+0041`, false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := hasMIFEncoding(tc.input)
			if result != tc.expected {
				t.Errorf("hasMIFEncoding(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestHasUnicodeEncoding(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"\\U+0041", true},
		{"\\U+00C0", true},
		{"\\U+FFFF", true},
		{"normal text", false},
		{"\\M+0041", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := hasUnicodeEncoding(tc.input)
			if result != tc.expected {
				t.Errorf("hasUnicodeEncoding(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestConvertToUTF8UTF8(t *testing.T) {
	detector := NewEncodingDetector()
	detector.encoding = CodePageUTF8

	input := "Hello, World!"
	result, err := detector.ConvertToUTF8(input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != input {
		t.Errorf("Expected '%s', got '%s'", input, result)
	}
}

func TestConvertToUTF8UnicodeOnly(t *testing.T) {
	detector := NewEncodingDetector()
	detector.encoding = CodePageANSI
	detector.unicodeOnly = true

	input := "Unicode text"
	result, err := detector.ConvertToUTF8(input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != input {
		t.Errorf("Expected '%s', got '%s'", input, result)
	}
}

func TestConvertToEncodingUTF8(t *testing.T) {
	detector := NewEncodingDetector()
	detector.encoding = CodePageUTF8

	input := "Test string"
	result, err := detector.ConvertToEncoding(input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if string(result) != input {
		t.Errorf("Expected '%s', got '%s'", input, string(result))
	}
}

func TestDecodeMIFSequences(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{`\M+10041`, "𐁁"},
		{`\M+200C0`, "𠃀"},
		{`\M+300E0`, "𰃠"},
		{`\M+10430`, "𐐰"},
		{`\M+10041\M+20042`, "𐁁𠁂"},
		{"no encoding", "no encoding"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := decodeMIFSequences(tc.input)
			if result != tc.expected {
				t.Errorf("decodeMIFSequences(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestDecodeUnicodeSequences(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"\\U+0041", "A"},
		{"\\U+00C0", "À"},
		{"\\U+00E0", "à"},
		{"\\U+03B1", "α"},
		{"\\U+0041\\U+0042", "AB"},
		{"no encoding", "no encoding"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := decodeUnicodeSequences(tc.input)
			if result != tc.expected {
				t.Errorf("decodeUnicodeSequences(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestProcessDXFUnicodeEscapes(t *testing.T) {
	input := "Hello\\U+0021\\M+0041\\U+0020World\x00"
	result := processDXFUnicodeEscapes(input)

	if strings.Contains(result, "\x00") {
		t.Error("Null characters should be removed")
	}

	if !strings.Contains(result, "!") {
		t.Error("Unicode escape should be decoded")
	}
}

func TestGetEncodingName(t *testing.T) {
	detector := NewEncodingDetector()
	detector.encoding = CodePageANSI

	if detector.GetEncodingName() != CodePageANSI {
		t.Errorf("Expected '%s', got '%s'", CodePageANSI, detector.GetEncodingName())
	}
}

func TestSetEncoding(t *testing.T) {
	detector := NewEncodingDetector()

	err := detector.SetEncoding(CodePageUTF8)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if detector.encoding != CodePageUTF8 {
		t.Errorf("Expected encoding '%s', got '%s'", CodePageUTF8, detector.encoding)
	}

	if detector.confidence != 0.8 {
		t.Errorf("Expected confidence 0.8, got %f", detector.confidence)
	}

	err = detector.SetEncoding("invalid")
	if err == nil {
		t.Error("Expected error for invalid encoding")
	}
}

func TestGetConfidence(t *testing.T) {
	detector := NewEncodingDetector()
	detector.confidence = 0.75

	if detector.GetConfidence() != 0.75 {
		t.Errorf("Expected 0.75, got %f", detector.GetConfidence())
	}
}

func TestIsUnicode(t *testing.T) {
	detector := NewEncodingDetector()

	if !detector.IsUnicode() {
		t.Error("Expected IsUnicode to return true by default")
	}

	detector.unicodeOnly = false
	if detector.IsUnicode() {
		t.Error("Expected IsUnicode to return false when unicodeOnly is false")
	}
}

func TestHasMIF(t *testing.T) {
	detector := NewEncodingDetector()

	if detector.HasMIF() {
		t.Error("Expected HasMIF to return false by default")
	}

	detector.hasMIF = true
	if !detector.HasMIF() {
		t.Error("Expected HasMIF to return true when set")
	}
}

func TestValidateEncoding(t *testing.T) {
	err := ValidateEncoding(CodePageUTF8)
	if err != nil {
		t.Errorf("Unexpected error for valid encoding: %v", err)
	}

	err = ValidateEncoding(CodePageANSI)
	if err != nil {
		t.Errorf("Unexpected error for valid encoding: %v", err)
	}

	err = ValidateEncoding("")
	if err == nil {
		t.Error("Expected error for empty encoding")
	}

	err = ValidateEncoding("invalid")
	if err == nil {
		t.Error("Expected error for invalid encoding")
	}

	err = ValidateEncoding("utf8")
	if err != nil {
		t.Errorf("utf8 should be valid: %v", err)
	}
}

func TestEncodingDetectorString(t *testing.T) {
	detector := NewEncodingDetector()
	detector.encoding = CodePageANSI
	detector.confidence = 0.75
	detector.hasMIF = true

	result := detector.String()

	if !strings.Contains(result, "EncodingDetector") {
		t.Error("String should contain 'EncodingDetector'")
	}

	if !strings.Contains(result, "1252") {
		t.Error("String should contain encoding name")
	}

	if !strings.Contains(result, "0.75") {
		t.Error("String should contain confidence")
	}

	if !strings.Contains(result, "true") {
		t.Error("String should contain hasMIF value")
	}
}

func TestGetEncodingByName(t *testing.T) {
	result := getEncodingByName(CodePageUTF8)
	if result != nil {
		t.Error("UTF-8 should return nil (no conversion needed)")
	}

	result = getEncodingByName(CodePageANSI)
	if result != nil {
		t.Error("ANSI should return nil (basic support)")
	}

	result = getEncodingByName("invalid")
	if result != nil {
		t.Error("Invalid encoding should return nil")
	}
}

func TestConvertToUTF8ANSI(t *testing.T) {
	detector := NewEncodingDetector()
	detector.encoding = CodePageANSI
	detector.unicodeOnly = false

	input := "Test"
	result, err := detector.ConvertToUTF8(input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != input {
		t.Errorf("Expected '%s', got '%s'", input, result)
	}
}

func TestConvertToEncodingANSI(t *testing.T) {
	detector := NewEncodingDetector()
	detector.encoding = CodePageANSI
	detector.unicodeOnly = false

	input := "Test"
	result, err := detector.ConvertToEncoding(input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if string(result) != input {
		t.Errorf("Expected '%s', got '%s'", input, string(result))
	}
}

func TestConvertToEncodingUnsupported(t *testing.T) {
	detector := NewEncodingDetector()
	detector.encoding = "unsupported"
	detector.unicodeOnly = false

	input := "Test"
	result, err := detector.ConvertToEncoding(input)

	if err == nil {
		t.Error("Expected error for unsupported encoding")
	}

	if string(result) != input {
		t.Errorf("Expected '%s', got '%s'", input, string(result))
	}
}
