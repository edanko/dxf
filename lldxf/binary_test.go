package lldxf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBinaryLoader(t *testing.T) {
	// Test with existing test file that should be ASCII
	testFile := filepath.Join("..", "testdata", "mypoint.dxf")

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Test format detection
	format, err := DetectDXFFormat(data)
	if err != nil {
		t.Fatalf("Failed to detect format: %v", err)
	}

	if format != "ascii" {
		t.Errorf("Expected format 'ascii', got '%s'", format)
	}

	// Test creating binary loader (should fail for ASCII)
	_, err = NewBinaryLoader(data)
	if err == nil {
		t.Error("Expected NewBinaryLoader to fail for ASCII data")
	}
}

func TestBinarySignature(t *testing.T) {
	// Test binary signature detection
	binarySig := []byte("AutoCAD Binary DXF\r\n\x1a\x00")

	format, err := DetectDXFFormat(binarySig)
	if err != nil {
		t.Fatalf("Failed to detect binary format: %v", err)
	}

	if format != "binary" {
		t.Errorf("Expected format 'binary', got '%s'", format)
	}

	// Test creating binary loader with minimal valid data
	testData := make([]byte, 1024)
	copy(testData, binarySig)

	loader, err := NewBinaryLoader(testData)
	if err != nil {
		t.Fatalf("Failed to create binary loader: %v", err)
	}

	// Test reading first tag
	tag, err := loader.Next()
	if err != nil {
		t.Fatalf("Failed to read first tag: %v", err)
	}

	if tag == nil {
		t.Error("Expected non-nil tag")
	}
}

func TestBinaryTagger(t *testing.T) {
	// Test creating binary tagger with ASCII data should work gracefully
	asciiData := "0\nSECTION\n2\nHEADER\n0\nENDSEC\n0\nEOF\n"
	r := strings.NewReader(asciiData)

	// This should create a binary tagger but it will fail to read binary data
	// The test verifies that it handles the error gracefully
	tagger := NewBinaryTagger(r)

	// Should try to iterate but fail gracefully since this isn't binary data
	count := 0
	for tagger.Next() {
		tag := tagger.Tag()
		if tag == nil {
			t.Error("Expected non-nil tag")
		}
		count++
	}

	// Expect an error since this is ASCII data being read as binary
	if err := tagger.Err(); err == nil {
		t.Error("Expected error when reading ASCII data with binary tagger")
	}
}
