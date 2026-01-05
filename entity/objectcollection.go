package entity

import (
	"github.com/edanko/dxf/format"
)

type ObjectCollection struct {
	*entity
	Name          string
	Description   string
	ObjectTypes   []string
	EntityHandles []string
}

func NewObjectCollection() *ObjectCollection {
	return &ObjectCollection{
		entity:        NewEntity(OBJECTCOLLECTION),
		Name:          "",
		Description:   "",
		ObjectTypes:   make([]string, 0),
		EntityHandles: make([]string, 0),
	}
}

func (o *ObjectCollection) IsEntity() bool {
	return true
}

func (o *ObjectCollection) Format(f format.Formatter) {
	o.entity.Format(f)
	f.WriteString(100, "AcDbObjectCollection")
	f.WriteString(1, o.Name)
	f.WriteString(2, o.Description)
	f.WriteInt(90, len(o.ObjectTypes))
	for _, objType := range o.ObjectTypes {
		f.WriteString(100, objType)
	}
	for _, handle := range o.EntityHandles {
		f.WriteString(330, handle)
	}
}

func (o *ObjectCollection) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (o *ObjectCollection) Copy() Entity {
	coll := NewObjectCollection()
	coll.entity = o.entity
	coll.Name = o.Name
	coll.Description = o.Description
	coll.ObjectTypes = make([]string, len(o.ObjectTypes))
	for i, t := range o.ObjectTypes {
		coll.ObjectTypes[i] = t
	}
	coll.EntityHandles = make([]string, len(o.EntityHandles))
	for i, h := range o.EntityHandles {
		coll.EntityHandles[i] = h
	}
	return coll
}

func (o *ObjectCollection) Validate() error {
	return nil
}

func (o *ObjectCollection) AddObjectType(objType string) {
	o.ObjectTypes = append(o.ObjectTypes, objType)
}

func (o *ObjectCollection) AddEntity(handle string) {
	o.EntityHandles = append(o.EntityHandles, handle)
}

func (o *ObjectCollection) ClearObjectTypes() {
	o.ObjectTypes = make([]string, 0)
}

func (o *ObjectCollection) ClearEntities() {
	o.EntityHandles = make([]string, 0)
}
