package lldxf

import (
	"bytes"
	"testing"
)

func TestNewTagCompiler(t *testing.T) {
	tagger := NewASCIITagger(bytes.NewReader([]byte("0\nSECTION\n")))
	compiler := NewTagCompiler(tagger)

	if compiler == nil {
		t.Fatal("NewTagCompiler returned nil")
	}

	if compiler.Tag() != nil {
		t.Error("NewTagCompiler should have nil current tag")
	}

	if compiler.Err() != nil {
		t.Error("NewTagCompiler should have nil error")
	}
}

func TestTagCompilerNext(t *testing.T) {
	input := "0\nSECTION\n8\nLAYER\n"
	tagger := NewASCIITagger(bytes.NewReader([]byte(input)))
	compiler := NewTagCompiler(tagger)

	if !compiler.Next() {
		t.Fatal("Next() returned false unexpectedly")
	}

	tag := compiler.Tag()
	if tag == nil {
		t.Fatal("Tag() returned nil after Next()")
	}

	if tag.Code() != 0 {
		t.Errorf("Expected code 0, got %d", tag.Code())
	}

	val, ok := tag.Value().(string)
	if !ok {
		t.Fatal("Value is not a string")
	}

	if val != "SECTION" {
		t.Errorf("Expected 'SECTION', got %q", val)
	}
}

func TestTagCompilerPointAssembly2D(t *testing.T) {
	input := "10\n1.5\n20\n2.5\n"
	tagger := NewASCIITagger(bytes.NewReader([]byte(input)))
	compiler := NewTagCompiler(tagger)

	if !compiler.Next() {
		t.Fatal("Next() returned false")
	}

	tag := compiler.Tag()
	coords, ok := tag.Value().([]float64)
	if !ok {
		t.Fatalf("Expected []float64, got %T", tag.Value())
	}

	if len(coords) != 2 {
		t.Errorf("Expected 2 coordinates, got %d", len(coords))
	}

	if coords[0] != 1.5 {
		t.Errorf("Expected X=1.5, got %f", coords[0])
	}

	if coords[1] != 2.5 {
		t.Errorf("Expected Y=2.5, got %f", coords[1])
	}
}

func TestTagCompilerPointAssembly3D(t *testing.T) {
	input := "10\n1.5\n20\n2.5\n30\n3.5\n"
	tagger := NewASCIITagger(bytes.NewReader([]byte(input)))
	compiler := NewTagCompiler(tagger)

	if !compiler.Next() {
		t.Fatal("Next() returned false")
	}

	tag := compiler.Tag()
	coords, ok := tag.Value().([]float64)
	if !ok {
		t.Fatalf("Expected []float64, got %T", tag.Value())
	}

	if len(coords) != 3 {
		t.Errorf("Expected 3 coordinates, got %d", len(coords))
	}

	if coords[2] != 3.5 {
		t.Errorf("Expected Z=3.5, got %f", coords[2])
	}
}

func TestTagCompilerError(t *testing.T) {
	input := "0\nSECTION\n"
	tagger := NewASCIITagger(bytes.NewReader([]byte(input)))
	compiler := NewTagCompiler(tagger)

	if !compiler.Next() {
		t.Fatal("First Next() failed")
	}

	tag := compiler.Tag()
	if tag.Code() != 0 {
		t.Errorf("Expected code 0, got %d", tag.Code())
	}

	if compiler.Next() {
		t.Error("Second Next() should have failed due to EOF")
	}
}

func TestTagCompilerLookahead(t *testing.T) {
	input := "20\n2.5\n0\nSECTION\n"
	tagger := NewASCIITagger(bytes.NewReader([]byte(input)))
	compiler := NewTagCompiler(tagger)

	if !compiler.Next() {
		t.Fatal("Next() returned false")
	}

	tag := compiler.Tag()
	if tag.Code() != 20 {
		t.Errorf("Expected code 20, got %d", tag.Code())
	}

	if !compiler.Next() {
		t.Fatal("Second Next() returned false")
	}

	tag = compiler.Tag()
	if tag.Code() != 0 {
		t.Errorf("Expected code 0, got %d", tag.Code())
	}
}

func TestRobustTaggerMaxTags(t *testing.T) {
	input := "0\nA\n1\nB\n2\nC\n"
	r := bytes.NewReader([]byte(input))
	opts := DefaultTaggerOptions()
	opts.MaxTags = 2

	tagger := NewRobustTagger(r, opts)

	count := 0
	for tagger.Next() {
		count++
		if count >= opts.MaxTags {
			break
		}
	}

	if count != opts.MaxTags {
		t.Errorf("Expected %d tags, got %d", opts.MaxTags, count)
	}
}

func TestRobustTaggerTag(t *testing.T) {
	input := "0\nSECTION\n"
	r := bytes.NewReader([]byte(input))
	tagger := NewRobustTagger(r, nil)

	if tagger.Tag() != nil {
		t.Error("Initial Tag() should return nil")
	}

	tagger.Next()

	if tagger.Tag() == nil {
		t.Error("Tag() should return current tag after Next()")
	}
}

func TestRobustTaggerErr(t *testing.T) {
	r := bytes.NewReader([]byte(""))
	tagger := NewRobustTagger(r, nil)

	if tagger.Err() != nil {
		t.Error("Initial Err() should return nil")
	}
}

func TestIsValidDXFString(t *testing.T) {
	testCases := []struct {
		input    string
		code     int
		expected bool
	}{
		{"Hello", 1, true},
		{"", 1, true},
		{"With\ttab", 1, true},
		{"With\nnewline", 1, true},
		{"With\rreturn", 1, true},
		{"With\x00null", 1, false},
		{"With\x01ctrl", 1, false},
		{"Hello", 310, true},
	}

	for _, tc := range testCases {
		result := isValidDXFString(tc.input, tc.code)
		if result != tc.expected {
			t.Errorf("isValidDXFString(%q, %d) = %v, expected %v", tc.input, tc.code, result, tc.expected)
		}
	}
}

func TestLoadTagsWithCompiler(t *testing.T) {
	input := "0\nSECTION\n8\nLAYER\n0\nENDSEC\n"
	tags, err := LoadTagsWithCompiler(bytes.NewReader([]byte(input)))

	if err != nil {
		t.Fatalf("LoadTagsWithCompiler failed: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	if tags[0].Code() != 0 {
		t.Errorf("First tag code should be 0, got %d", tags[0].Code())
	}
}

func TestLoadTagsRobust(t *testing.T) {
	input := "0\nSECTION\n8\nLAYER\n0\nENDSEC\n"
	tags, err := LoadTagsRobust(bytes.NewReader([]byte(input)), nil)

	if err != nil {
		t.Fatalf("LoadTagsRobust failed: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}
}

func TestLoadTagsRobustWithOptions(t *testing.T) {
	input := "0\nSECTION\n999\nCOMMENT\n8\nLAYER\n"
	opts := DefaultTaggerOptions()
	opts.SkipComments = true

	tags, err := LoadTagsRobust(bytes.NewReader([]byte(input)), opts)

	if err != nil {
		t.Fatalf("LoadTagsRobust failed: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags (comment skipped), got %d", len(tags))
	}
}

func TestDetectDXFFormatBinary(t *testing.T) {
	signature := []byte("AutoCAD Binary DXF\r\n\x1a\x00")
	data := make([]byte, 22)
	copy(data, signature)

	format, err := DetectDXFFormat(data)

	if err != nil {
		t.Fatalf("DetectDXFFormat failed: %v", err)
	}

	if format != "binary" {
		t.Errorf("Expected 'binary', got %q", format)
	}
}

func TestDetectDXFFormatASCII(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"0\\nSECTION", []byte("0\nSECTION\n" + string(make([]byte, 20)))},
		{"0\\r\\nSECTION", []byte("0\r\nSECTION\n" + string(make([]byte, 20)))},
		{"with spaces", []byte(" 0\nSECTION\n" + string(make([]byte, 20)))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			format, err := DetectDXFFormat(tc.data)
			if err != nil {
				t.Fatalf("DetectDXFFormat failed: %v", err)
			}
			if format != "ascii" {
				t.Errorf("Expected 'ascii', got %q", format)
			}
		})
	}
}

func TestDetectDXFFormatUnknown(t *testing.T) {
	data := []byte("Some random content")

	format, err := DetectDXFFormat(data)

	if err == nil {
		t.Error("DetectDXFFormat should return error for unknown format")
	}

	if format != "" {
		t.Errorf("Expected empty format, got %q", format)
	}
}

func TestDetectDXFFormatTooShort(t *testing.T) {
	data := []byte("short")

	format, err := DetectDXFFormat(data)

	if err == nil {
		t.Error("DetectDXFFormat should return error for short data")
	}

	if format != "" {
		t.Errorf("Expected empty format, got %q", format)
	}
}

func TestCreateTaggerBinary(t *testing.T) {
	r := bytes.NewReader([]byte{})
	tagger := CreateTagger(r, "binary")

	if _, ok := tagger.(*BinaryTagger); !ok {
		t.Error("CreateTagger should return BinaryTagger for 'binary' format")
	}
}

func TestCreateTaggerASCII(t *testing.T) {
	r := bytes.NewReader([]byte{})
	tagger := CreateTagger(r, "ascii")

	if _, ok := tagger.(*asciiTagger); !ok {
		t.Error("CreateTagger should return asciiTagger for 'ascii' format")
	}
}

func TestCreateTaggerDefault(t *testing.T) {
	r := bytes.NewReader([]byte{})
	tagger := CreateTagger(r, "unknown")

	if _, ok := tagger.(*asciiTagger); !ok {
		t.Error("CreateTagger should return asciiTagger for unknown format")
	}
}

func TestMin(t *testing.T) {
	if min(1, 2) != 1 {
		t.Error("min(1, 2) should return 1")
	}

	if min(2, 1) != 1 {
		t.Error("min(2, 1) should return 1")
	}

	if min(5, 5) != 5 {
		t.Error("min(5, 5) should return 5")
	}
}
