package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

type TagPair struct {
	Code  int
	Value string
}

type DXFEntity struct {
	TypeName string
	Tags     []TagPair
}

func NewDXFEntity(typeName string) *DXFEntity {
	return &DXFEntity{
		TypeName: typeName,
		Tags:     make([]TagPair, 0),
	}
}

func (d *DXFEntity) IsEntity() bool {
	return true
}

func (d *DXFEntity) Format(f format.Formatter) {
	f.WriteString(0, d.TypeName)
	for _, tag := range d.Tags {
		f.WriteString(tag.Code, tag.Value)
	}
}

func (d *DXFEntity) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (d *DXFEntity) SetBlockRecord(h handle.Handler) {}
func (d *DXFEntity) Layer() *table.Layer             { return nil }
func (d *DXFEntity) SetLayer(l *table.Layer)         {}
func (d *DXFEntity) SetLtscale(v float64)            {}
func (d *DXFEntity) SetColor(c interface{})          {}

func (d *DXFEntity) Handle() string {
	for _, tag := range d.Tags {
		if tag.Code == 5 {
			return tag.Value
		}
	}
	return ""
}

func (d *DXFEntity) SetHandle(hg *handle.HandleGenerator) {}

func (d *DXFEntity) DXFType() string {
	return d.TypeName
}

func (d *DXFEntity) GetAttributes() map[string]interface{} {
	return make(map[string]interface{})
}

func (d *DXFEntity) LoadAttributes(attribs map[string]interface{}) error {
	return nil
}
