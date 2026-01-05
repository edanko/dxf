package entity

import (
	"github.com/edanko/dxf/format"
)

type DXFGroups struct {
	*entity
	Name          string
	Description   string
	Unnamed       bool
	Selectable    bool
	Hidden        bool
	EntityHandles []string
}

func NewDXFGroups() *DXFGroups {
	return &DXFGroups{
		entity:        NewEntity(DXFGROUPS),
		Name:          "",
		Description:   "",
		Unnamed:       false,
		Selectable:    true,
		Hidden:        false,
		EntityHandles: make([]string, 0),
	}
}

func (d *DXFGroups) IsEntity() bool {
	return true
}

func (d *DXFGroups) Format(f format.Formatter) {
	d.entity.Format(f)
	f.WriteString(100, "AcDbGroup")
	f.WriteString(1, d.Name)
	f.WriteString(2, d.Description)
	var flags int
	if d.Unnamed {
		flags |= 1
	}
	if !d.Selectable {
		flags |= 2
	}
	if d.Hidden {
		flags |= 4
	}
	f.WriteInt(70, flags)
	for _, handle := range d.EntityHandles {
		f.WriteString(5, handle)
	}
}

func (d *DXFGroups) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (d *DXFGroups) Copy() Entity {
	groups := NewDXFGroups()
	groups.entity = d.entity
	groups.Name = d.Name
	groups.Description = d.Description
	groups.Unnamed = d.Unnamed
	groups.Selectable = d.Selectable
	groups.Hidden = d.Hidden
	groups.EntityHandles = make([]string, len(d.EntityHandles))
	for i, h := range d.EntityHandles {
		groups.EntityHandles[i] = h
	}
	return groups
}

func (d *DXFGroups) Validate() error {
	return nil
}

func (d *DXFGroups) AddEntity(handle string) {
	d.EntityHandles = append(d.EntityHandles, handle)
}

func (d *DXFGroups) RemoveEntity(handle string) {
	for i, h := range d.EntityHandles {
		if h == handle {
			d.EntityHandles = append(d.EntityHandles[:i], d.EntityHandles[i+1:]...)
			return
		}
	}
}

func (d *DXFGroups) ClearEntities() {
	d.EntityHandles = make([]string, 0)
}

func (d *DXFGroups) Len() int {
	return len(d.EntityHandles)
}
