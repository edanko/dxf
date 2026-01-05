package lldxf

import (
	"bytes"
	"testing"
)

func TestBinaryLoaderSignature(t *testing.T) {
	validSignature := []byte("AutoCAD Binary DXF\r\n\x1a\x00")
	_, err := NewBinaryLoader(validSignature)
	if err != nil {
		t.Errorf("Valid signature should not error: %v", err)
	}

	invalidSignature := []byte("Not a DXF file")
	_, err = NewBinaryLoader(invalidSignature)
	if err == nil {
		t.Error("Invalid signature should error")
	}
}

func TestBinaryLoaderVersionDetection(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "Valid binary DXF header",
			data:    append([]byte("AutoCAD Binary DXF\r\n\x1a\x00"), bytes.Repeat([]byte{0}, 200)...),
			wantErr: false,
		},
		{
			name:    "Too short",
			data:    []byte("AutoCAD Binary DXF\r\n"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBinaryLoader(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBinaryLoader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBinaryLoaderGroupCodes(t *testing.T) {
	data := []byte("AutoCAD Binary DXF\r\n\x1a\x00") // Signature
	data = append(data, []byte{
		0x0F, 0x00, // Group code 15 (float)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, // 1.0
	}...)

	loader, err := NewBinaryLoader(data)
	if err != nil {
		t.Fatalf("Failed to create binary loader: %v", err)
	}

	tags, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("Failed to load all tags: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	if tags[0].Code != 15 {
		t.Errorf("Expected group code 15, got %d", tags[0].Code)
	}
}
