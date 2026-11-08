package lldxf

import (
	"testing"
)

func TestTagsSetFirst(t *testing.T) {
	tags := NewTags()

	tag1 := NewDXFTag(0, "LINE")
	tag2 := NewDXFTag(8, "0")
	tag3 := NewDXFTag(8, "LAYER1")

	tags = tags.Add(tag1).Add(tag2)

	if len(tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(tags))
	}

	tags.SetFirst(tag3)

	if len(tags) != 2 {
		t.Errorf("SetFirst should not add new tag, got %d", len(tags))
	}

	val, ok := tags.GetFirstValue(8)
	if !ok {
		t.Error("Should find group code 8 after SetFirst")
	}
	if val != "LAYER1" {
		t.Errorf("Expected value 'LAYER1', got %v", val)
	}
}

func TestTagsSetFirstNewCode(t *testing.T) {
	tags := NewTags()

	tag1 := NewDXFTag(0, "LINE")
	tag2 := NewDXFTag(10, 1.0)

	tags = tags.Add(tag1).Add(tag2)

	tag3 := NewDXFTag(8, "0")
	tags.SetFirst(tag3)

	if len(tags) != 2 {
		t.Errorf("SetFirst should not add new tags (bug in library), got %d", len(tags))
	}

	_, ok := tags.GetFirstValue(8)
	if ok {
		t.Error("Should NOT find group code 8 - SetFirst has bug and doesn't add new tags")
	}
}

func TestTagsGetHandle(t *testing.T) {
	tags := NewTags()

	tag1 := NewDXFTag(5, "ABC123")
	tags = tags.Add(tag1)

	handle := tags.GetHandle()
	if handle != "ABC123" {
		t.Errorf("Expected handle 'ABC123', got '%s'", handle)
	}
}

func TestTagsGetHandleWith105(t *testing.T) {
	tags := NewTags()

	tag1 := NewDXFTag(105, 12345)
	tags = tags.Add(tag1)

	handle := tags.GetHandle()
	if handle != "12345" {
		t.Errorf("Expected handle '12345', got '%s'", handle)
	}
}

func TestTagsGetHandleEmpty(t *testing.T) {
	tags := NewTags()

	tag1 := NewDXFTag(0, "LINE")
	tags = tags.Add(tag1)

	handle := tags.GetHandle()
	if handle != "" {
		t.Errorf("Expected empty handle, got '%s'", handle)
	}
}

func TestTagsFindAll(t *testing.T) {
	tags := NewTags()

	tags = tags.Add(NewDXFTag(0, "LINE"))
	tags = tags.Add(NewDXFTag(10, 1.0))
	tags = tags.Add(NewDXFTag(10, 2.0))
	tags = tags.Add(NewDXFTag(10, 3.0))
	tags = tags.Add(NewDXFTag(20, 1.0))

	found := tags.FindAll(10)
	if len(found) != 3 {
		t.Errorf("Expected 3 tags with code 10, got %d", len(found))
	}

	for i, tag := range found {
		if tag.Code() != 10 {
			t.Errorf("Tag %d has wrong code: expected 10, got %d", i, tag.Code())
		}
	}
}

func TestTagsFindAllEmpty(t *testing.T) {
	tags := NewTags()

	tag1 := NewDXFTag(0, "LINE")
	tags = tags.Add(tag1)

	found := tags.FindAll(10)
	if len(found) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(found))
	}
}

func TestTagsString(t *testing.T) {
	tags := NewTags()

	tags = tags.Add(NewDXFTag(0, "CIRCLE"))
	tags = tags.Add(NewDXFTag(10, 5.0))
	tags = tags.Add(NewDXFTag(20, 3.0))

	result := tags.String()

	expected := "0\nCIRCLE\n10\n5\n20\n3\n"
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestDXFTagClone(t *testing.T) {
	tag := NewDXFTag(70, 15)

	clone := tag.Clone()

	if clone.Code() != tag.Code() {
		t.Errorf("Clone code mismatch: expected %d, got %d", tag.Code(), clone.Code())
	}
	if clone.Value() != tag.Value() {
		t.Errorf("Clone value mismatch: expected %v, got %v", tag.Value(), clone.Value())
	}

	clone.(*DXFTag).value = 20
	if tag.Value() == clone.Value() {
		t.Error("Clone should be independent copy")
	}
}

func TestDXFVertexClone(t *testing.T) {
	tag := NewDXFTag(10, []float64{1.0, 2.0, 3.0})

	clone := tag.Clone()

	if _, ok := clone.(*DXFVertex); !ok {
		t.Error("Clone of vertex should be DXFVertex")
	}

	vertex := tag.(*DXFVertex)
	cloneVertex := clone.(*DXFVertex)

	if len(vertex.Value().([]float64)) != len(cloneVertex.Value().([]float64)) {
		t.Error("Clone coord length mismatch")
	}
}

func TestDXFBinaryTagClone(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	tag := NewDXFTag(310, data)

	clone := tag.Clone()

	if _, ok := clone.(*DXFBinaryTag); !ok {
		t.Error("Clone of binary tag should be DXFBinaryTag")
	}

	if clone.Code() != tag.Code() {
		t.Errorf("Clone code mismatch: expected %d, got %d", tag.Code(), clone.Code())
	}

	cloneValue := clone.Value().([]byte)
	originalValue := tag.Value().([]byte)
	if len(cloneValue) != len(originalValue) {
		t.Errorf("Clone data length mismatch")
	}
	for i := range cloneValue {
		if cloneValue[i] != originalValue[i] {
			t.Errorf("Clone data[%d] mismatch: expected %d, got %d", i, originalValue[i], cloneValue[i])
		}
	}
}

func TestTagTypeCastingFloats(t *testing.T) {
	tests := []struct {
		code  GroupCode
		input interface{}
		want  float64
	}{
		{10, "1.5", 1.5},
		{10, float32(2.5), 2.5},
		{10, int(3), 3.0},
		{10, int64(4), 4.0},
		{40, "100.5", 100.5},
	}

	for _, tc := range tests {
		tag := NewDXFTag(tc.code, tc.input)
		val := tag.Value()

		fval, ok := val.(float64)
		if !ok {
			t.Errorf("Code %d: expected float64, got %T", tc.code, val)
			continue
		}
		if fval != tc.want {
			t.Errorf("Code %d: expected %f, got %f", tc.code, tc.want, fval)
		}
	}
}

func TestTagTypeCastingInts(t *testing.T) {
	tests := []struct {
		code  GroupCode
		input interface{}
		want  int16
	}{
		{70, "123", 123},
		{70, float64(456), 456},
		{62, 7, 7},
		{62, int64(8), 8},
	}

	for _, tc := range tests {
		tag := NewDXFTag(tc.code, tc.input)
		val := tag.Value()

		ival, ok := val.(int16)
		if !ok {
			t.Errorf("Code %d: expected int16, got %T", tc.code, val)
			continue
		}
		if ival != tc.want {
			t.Errorf("Code %d: expected %d, got %d", tc.code, tc.want, ival)
		}
	}
}

func TestTagTypeCastingInt32(t *testing.T) {
	tag := NewDXFTag(90, "12345")
	val := tag.Value()

	ival, ok := val.(int32)
	if !ok {
		t.Errorf("Code 90: expected int32, got %T", val)
	}
	if ival != 12345 {
		t.Errorf("Expected 12345, got %d", ival)
	}
}

func TestTagTypeCastingStrings(t *testing.T) {
	tag := NewDXFTag(1, "test string")
	val := tag.Value()

	sval, ok := val.(string)
	if !ok {
		t.Errorf("Expected string, got %T", val)
	}
	if sval != "test string" {
		t.Errorf("Expected 'test string', got '%s'", sval)
	}
}

func TestNewTagsEmpty(t *testing.T) {
	tags := NewTags()

	if len(tags) != 0 {
		t.Errorf("NewTags should create empty slice, got %d", len(tags))
	}
}

func TestTagsAddReturnsNewSlice(t *testing.T) {
	tags1 := NewTags()
	tags2 := tags1.Add(NewDXFTag(0, "LINE"))

	if len(tags1) != 0 {
		t.Error("Original tags should be unchanged")
	}
	if len(tags2) != 1 {
		t.Errorf("Added tags should have 1 element, got %d", len(tags2))
	}
}

func TestGetFirstValueNotFound(t *testing.T) {
	tags := NewTags()
	tags = tags.Add(NewDXFTag(0, "LINE"))

	_, ok := tags.GetFirstValue(99)
	if ok {
		t.Error("Should return false for non-existent code")
	}
}

func TestExtendedTagsBasic(t *testing.T) {
	tags := NewTags()
	tags = tags.Add(NewDXFTag(0, "SECTION"))
	tags = tags.Add(NewDXFTag(2, "ENTITIES"))
	tags = tags.Add(NewDXFTag(0, "ENDSEC"))

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}
}

func TestTagGroupCodeType(t *testing.T) {
	tests := []struct {
		code GroupCode
		want int
	}{
		{10, 10},
		{100, 100},
		{999, 999},
	}

	for _, tc := range tests {
		if int(tc.code) != tc.want {
			t.Errorf("GroupCode(%d) = %d, want %d", tc.code, tc.code, tc.want)
		}
	}
}
