package entity

import (
	"github.com/edanko/dxf/format"
)

type Placeholder struct {
	*entity
}

func NewPlaceholder() *Placeholder {
	return &Placeholder{
		entity: NewEntity(PLACEHOLDER),
	}
}

func (p *Placeholder) IsEntity() bool {
	return true
}

func (p *Placeholder) Format(f format.Formatter) {
	p.entity.Format(f)
	f.WriteString(100, "AcDbPlaceholder")
}

func (p *Placeholder) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (p *Placeholder) Copy() Entity {
	ph := NewPlaceholder()
	ph.entity = p.entity
	return ph
}

func (p *Placeholder) Validate() error {
	return nil
}
