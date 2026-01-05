package object

import (
	"github.com/edanko/dxf/format"
)

// XRecord represents an XRECORD object for storing application-specific data
type XRecord struct {
	*BaseObject

	// Core properties
	handle string `dxf:"5"`   // Entity handle
	owner  string `dxf:"330"` // Owner handle (usually "0" for root dictionary)

	// XRecord data
	tags []*XRecordTag `dxf:"-"`

	// Optional schema information
	schema  int `dxf:"280"` // Schema number
	cloning int `dxf:"280"` // Cloning flag (0=not cloneable, 1=cloneable)
}

// XRecordTag represents a single tag in an XRECORD
type XRecordTag struct {
	groupCode int    `dxf:"-"`
	value     string `dxf:"-"`
}

// NewXRecord creates a new XRECORD object
func NewXRecord() *XRecord {
	return &XRecord{
		BaseObject: &BaseObject{},
		tags:       make([]*XRecordTag, 0),
		cloning:    0, // Default to not cloneable
	}
}

// NewXRecordWithOwner creates a new XRECORD with specified owner
func NewXRecordWithOwner(owner string) *XRecord {
	xrec := NewXRecord()
	xrec.owner = owner
	return xrec
}

// IsObject returns true for XRECORD objects
func (x *XRecord) IsObject() bool {
	return true
}

// Handle returns the object handle
func (x *XRecord) Handle() string {
	return x.handle
}

// SetHandle sets the object handle
func (x *XRecord) SetHandle(h string) {
	x.handle = h
}

// SetOwner sets the owner handle
func (x *XRecord) SetOwner(owner string) {
	x.owner = owner
}

// AddTag adds a tag to the XRECORD
func (x *XRecord) AddTag(groupCode int, value string) {
	tag := &XRecordTag{
		groupCode: groupCode,
		value:     value,
	}
	x.tags = append(x.tags, tag)
}

// AddData adds multiple tags from a map
func (x *XRecord) AddData(data map[int]string) {
	for groupCode, value := range data {
		x.AddTag(groupCode, value)
	}
}

// GetTag returns a tag by group code
func (x *XRecord) GetTag(groupCode int) (*XRecordTag, bool) {
	for _, tag := range x.tags {
		if tag.groupCode == groupCode {
			return tag, true
		}
	}
	return nil, false
}

// GetTagValue returns a tag value by group code
func (x *XRecord) GetTagValue(groupCode int) (string, bool) {
	if tag, exists := x.GetTag(groupCode); exists {
		return tag.value, true
	}
	return "", false
}

// RemoveTag removes a tag by group code
func (x *XRecord) RemoveTag(groupCode int) {
	for i, tag := range x.tags {
		if tag.groupCode == groupCode {
			x.tags = append(x.tags[:i], x.tags[i+1:]...)
			return
		}
	}
}

// Clear removes all tags from the XRECORD
func (x *XRecord) Clear() {
	x.tags = make([]*XRecordTag, 0)
}

// GetTags returns all tags as a map for easier access
func (x *XRecord) GetTags() map[int]string {
	result := make(map[int]string)
	for _, tag := range x.tags {
		result[tag.groupCode] = tag.value
	}
	return result
}

// SetSchema sets the schema number for XRECORD validation
func (x *XRecord) SetSchema(schema int) {
	x.schema = schema
}

// SetCloning sets the cloning flag
func (x *XRecord) SetCloning(cloning int) {
	x.cloning = cloning
}

// Format writes XRECORD data to DXF formatter
func (x *XRecord) Format(f format.Formatter) {
	// Write common object data
	f.WriteString(0, "XRECORD")
	f.WriteString(5, x.handle)

	if x.owner != "" {
		f.WriteString(330, x.owner)
	}

	f.WriteString(100, "AcDbXRecord")
	cloning := 0
	if x.cloning != 0 {
		cloning = 1
	}
	f.WriteInt(280, cloning)

	// Write schema if specified
	if x.schema != 0 {
		f.WriteInt(280, x.schema)
	}

	// Write all tags
	for _, tag := range x.tags {
		switch {
		case tag.groupCode >= 0 && tag.groupCode <= 9:
			// String values
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 10 && tag.groupCode <= 39:
			// Point/coordinate values
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 40 && tag.groupCode <= 59:
			// Floating point values
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 60 && tag.groupCode <= 79:
			// 16-bit integer values
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 90 && tag.groupCode <= 99:
			// 32-bit integer values
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 100 && tag.groupCode <= 107:
			// String values (app data)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 160 && tag.groupCode <= 169:
			// 16-bit integers (misc)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 170 && tag.groupCode <= 179:
			// 16-bit integers (misc)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 270 && tag.groupCode <= 279:
			// 16-bit integers (misc)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 280 && tag.groupCode <= 289:
			// 16-bit integers (misc)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 290 && tag.groupCode <= 299:
			// 16-bit integers (misc)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 300 && tag.groupCode <= 309:
			// String values
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 310 && tag.groupCode <= 319:
			// Binary data
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 320 && tag.groupCode <= 369:
			// String values (app data)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 370 && tag.groupCode <= 379:
			// 16-bit integers (misc)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 380 && tag.groupCode <= 389:
			// 16-bit integers (misc)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 390 && tag.groupCode <= 399:
			// String values
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 400 && tag.groupCode <= 409:
			// 16-bit integers (misc)
			f.WriteString(tag.groupCode, tag.value)
		case tag.groupCode >= 410 && tag.groupCode <= 419:
			// String values (app data)
			f.WriteString(tag.groupCode, tag.value)
		default:
			// Default to string for unknown group codes
			f.WriteString(tag.groupCode, tag.value)
		}
	}
}
