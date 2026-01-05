package entity

import (
	"github.com/edanko/dxf/format"
)

type IDBuffer struct {
	*entity
	Handles []string
}

func NewIDBuffer() *IDBuffer {
	return &IDBuffer{
		entity:  NewEntity(IDBUFFER),
		Handles: make([]string, 0),
	}
}

func (b *IDBuffer) IsEntity() bool {
	return true
}

func (b *IDBuffer) Format(f format.Formatter) {
	b.entity.Format(f)
	f.WriteString(100, "AcDbIdBuffer")
	for _, handle := range b.Handles {
		f.WriteString(330, handle)
	}
}

func (b *IDBuffer) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (b *IDBuffer) Copy() Entity {
	buf := NewIDBuffer()
	buf.entity = b.entity
	buf.Handles = make([]string, len(b.Handles))
	for i, h := range b.Handles {
		buf.Handles[i] = h
	}
	return buf
}

func (b *IDBuffer) Validate() error {
	return nil
}

func (b *IDBuffer) AddHandle(handle string) {
	b.Handles = append(b.Handles, handle)
}

func (b *IDBuffer) ClearHandles() {
	b.Handles = make([]string, 0)
}
