package lldxf

import (
	"testing"
)

func TestDXFTag(t *testing.T) {
	// Test basic tag creation
	tag := NewDXFTag(10, 1.5)
	if tag.Code() != 10 {
		t.Errorf("Expected code 10, got %d", tag.Code())
	}

	if tag.Value() != 1.5 {
		t.Errorf("Expected value 1.5, got %v (%T)", tag.Value(), tag.Value())
	}

	// Check if it's a vertex tag
	if _, isVertex := tag.(*DXFVertex); !isVertex {
		t.Error("Expected DXFVertex for point code")
	}

	// Test string representation
	expected := "10\n1.5\n"
	if tag.String() != expected {
		t.Errorf("Expected string '%s', got '%s'", expected, tag.String())
	}
}

func TestTagsCollection(t *testing.T) {
	tags := NewTags()

	tag1 := NewDXFTag(0, "LINE")
	tag2 := NewDXFTag(10, 0.0)
	tag3 := NewDXFTag(20, 0.0)

	tags = tags.Add(tag1).Add(tag2).Add(tag3)

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	// Test GetFirstValue
	if val, ok := tags.GetFirstValue(0); ok {
		if val != "LINE" {
			t.Errorf("Expected 'LINE', got %v", val)
		}
	} else {
		t.Error("Should find group code 0")
	}

	// Test FindAll
	pointTags := tags.FindAll(10)
	if len(pointTags) != 1 {
		t.Errorf("Expected 1 point tag, got %d", len(pointTags))
	}

	// Test DXFType
	if tags.DXFType() != "LINE" {
		t.Errorf("Expected entity type 'LINE', got '%s'", tags.DXFType())
	}
}

func TestTagTypes(t *testing.T) {
	// Test vertex tag
	vertexTag := NewDXFTag(10, 1.0)
	if _, ok := vertexTag.(*DXFVertex); !ok {
		t.Error("Expected DXFVertex for point code")
	}

	// Test binary tag
	binaryTag := NewDXFTag(310, []byte("test"))
	if _, ok := binaryTag.(*DXFBinaryTag); !ok {
		t.Error("Expected DXFBinaryTag for binary code")
	}

	// Test regular tag
	regularTag := NewDXFTag(62, 7)
	if _, ok := regularTag.(*DXFTag); !ok {
		t.Error("Expected DXFTag for regular code")
	}
}

func TestTagTypeCasting(t *testing.T) {
	// Test float casting
	floatTag := NewDXFTag(10, "1.5")
	if val := floatTag.Value(); val != 1.5 {
		t.Errorf("Expected float 1.5, got %v (%T)", val, val)
	}

	// Test int casting
	intTag := NewDXFTag(70, "123")
	if val := intTag.Value(); val != int16(123) {
		t.Errorf("Expected int 123, got %v (%T)", val, val)
	}
}

func TestTagCollector(t *testing.T) {
	collector := NewTagCollector()

	tag1 := NewDXFTag(0, "CIRCLE")
	tag2 := NewDXFTag(10, 5.0)
	tag3 := NewDXFTag(20, 3.0)

	collector.WriteTags([]Tag{tag1, tag2, tag3})

	if len(collector.Tags()) != 3 {
		t.Errorf("Expected 3 tags in collector, got %d", len(collector.Tags()))
	}

	// Test string output
	output := collector.String()
	expected := "0\nCIRCLE\n10\n5\n20\n3\n"
	if output != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, output)
	}
}
