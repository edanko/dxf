package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

type LinkedEntities struct {
	*entity
	Name        string
	Description string
	ParentID    string
	ChildIDs    []string
	Flags       int
	LinkType    int
}

func NewLinkedEntities() *LinkedEntities {
	return &LinkedEntities{
		entity:      NewEntity(LINKEDENTITIES),
		Name:        "",
		Description: "",
		ParentID:    "",
		ChildIDs:    make([]string, 0),
		Flags:       0,
		LinkType:    0,
	}
}

func (l *LinkedEntities) IsEntity() bool {
	return true
}

func (l *LinkedEntities) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLinkedEntities")
	f.WriteString(1, l.Name)
	f.WriteString(2, l.Description)
	if l.ParentID != "" {
		f.WriteString(330, l.ParentID)
	}
	for _, id := range l.ChildIDs {
		f.WriteString(331, id)
	}
	f.WriteInt(90, l.Flags)
	f.WriteInt(70, l.LinkType)
}

func (l *LinkedEntities) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (l *LinkedEntities) Copy() Entity {
	linked := NewLinkedEntities()
	linked.entity = l.entity
	linked.Name = l.Name
	linked.Description = l.Description
	linked.ParentID = l.ParentID
	linked.ChildIDs = make([]string, len(l.ChildIDs))
	for i, id := range l.ChildIDs {
		linked.ChildIDs[i] = id
	}
	linked.Flags = l.Flags
	linked.LinkType = l.LinkType
	return linked
}

func (l *LinkedEntities) Validate() error {
	return nil
}

func (l *LinkedEntities) AddChild(id string) {
	l.ChildIDs = append(l.ChildIDs, id)
}

func (l *LinkedEntities) RemoveChild(id string) {
	for i, childID := range l.ChildIDs {
		if childID == id {
			l.ChildIDs = append(l.ChildIDs[:i], l.ChildIDs[i+1:]...)
			return
		}
	}
}

func (l *LinkedEntities) ClearChildren() {
	l.ChildIDs = make([]string, 0)
}

func (l *LinkedEntities) GetTransformationMatrix() math.Matrix44 {
	return math.NewMatrix44()
}
