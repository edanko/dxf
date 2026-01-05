package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// XRecordTag represents a single DXF tag in an XRecord
type XRecordTag struct {
	Code  int
	Value interface{}
}

// XRecord represents an XRECORD entity for storing arbitrary DXF data
type XRecord struct {
	*entity
	Cloning int          // 280 - cloning flags
	Tags    []XRecordTag // arbitrary DXF tags
}

// NewXRecord creates a new XRecord entity
func NewXRecord() *XRecord {
	x := &XRecord{
		entity:  NewEntity(XRECORD),
		Cloning: 1,
		Tags:    []XRecordTag{},
	}
	return x
}

// IsEntity is for Entity interface.
func (x *XRecord) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (x *XRecord) Format(f format.Formatter) {
	x.entity.Format(f)
	f.WriteString(100, "AcDbXrecord")
	f.WriteInt(280, x.Cloning)

	// Write arbitrary tags
	for _, tag := range x.Tags {
		switch v := tag.Value.(type) {
		case string:
			f.WriteString(tag.Code, v)
		case float64:
			f.WriteFloat(tag.Code, v)
		case int:
			f.WriteInt(tag.Code, v)
		case bool:
			if v {
				f.WriteInt(tag.Code, 1)
			} else {
				f.WriteInt(tag.Code, 0)
			}
		case []float64:
			for i, val := range v {
				f.WriteFloat(tag.Code+i, val)
			}
		}
	}
}

// BBox returns the bounding box of the entity
func (x *XRecord) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the XRecord
func (x *XRecord) Transform(m *math.Matrix44) error {
	return nil
}

// Copy creates a deep copy of the XRecord entity
func (x *XRecord) Copy() Entity {
	xr := NewXRecord()
	xr.entity = x.entity
	xr.Cloning = x.Cloning
	xr.Tags = append(xr.Tags, x.Tags...)
	return xr
}

// Validate validates the XRecord entity
func (x *XRecord) Validate() error {
	if x.Cloning < 0 || x.Cloning > 5 {
		x.Cloning = 1
	}
	return nil
}

// AddTag adds a tag to the XRecord
func (x *XRecord) AddTag(code int, value interface{}) {
	x.Tags = append(x.Tags, XRecordTag{Code: code, Value: value})
}

// AddString adds a string tag
func (x *XRecord) AddString(code int, value string) {
	x.AddTag(code, value)
}

// AddFloat adds a float tag
func (x *XRecord) AddFloat(code int, value float64) {
	x.AddTag(code, value)
}

// AddInt adds an integer tag
func (x *XRecord) AddInt(code int, value int) {
	x.AddTag(code, value)
}

// AddPoint adds a 2D point (two float tags)
func (x *XRecord) AddPoint(code int, xVal, yVal float64) {
	x.AddFloat(code, xVal)
	x.AddFloat(code+1, yVal)
}

// AddPoint3D adds a 3D point (three float tags)
func (x *XRecord) AddPoint3D(code int, xVal, yVal, zVal float64) {
	x.AddFloat(code, xVal)
	x.AddFloat(code+1, yVal)
	x.AddFloat(code+2, zVal)
}

// Clear removes all tags
func (x *XRecord) Clear() {
	x.Tags = []XRecordTag{}
}

// Len returns the number of tags
func (x *XRecord) Len() int {
	return len(x.Tags)
}
