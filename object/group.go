package object

import (
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// Group represents GROUP Object.
type Group struct {
	Name        string
	Description string
	handle      string
	owner       handle.Handler
	entities    []entity.Entity
	selectable  bool
}

// IsObject is for Object interface.
func (g *Group) IsObject() bool {
	return true
}

// NewGroup creates a new Group.
func NewGroup(name, desc string, es ...entity.Entity) *Group {
	g := &Group{
		Name:        name,
		Description: desc,
		owner:       nil,
		entities:    es,
		selectable:  true,
	}
	return g
}

// SetOwner sets an owner(Dictionary).
func (g *Group) SetOwner(d *Dictionary) error {
	g.owner = d
	return d.AddItem(g.Name, g)
}

// Format writes data to formatter.
func (g *Group) Format(f format.Formatter) {
	f.WriteString(0, "GROUP")
	f.WriteString(5, g.handle)
	f.WriteString(102, "{ACAD_REACTORS")
	f.WriteString(330, g.owner.Handle())
	f.WriteString(102, "}")
	f.WriteString(330, g.owner.Handle())
	f.WriteString(100, "AcDbGroup")
	f.WriteString(300, g.Description)
	f.WriteInt(70, 0)
	if g.selectable {
		f.WriteInt(71, 1)
	} else {
		f.WriteInt(71, 0)
	}
	for _, e := range g.entities {
		f.WriteString(340, e.Handle())
	}
}

// Handle returns a handle value.
func (g *Group) Handle() string {
	return g.handle
}

// SetHandle sets a handle.
func (g *Group) SetHandle(hg *handle.HandleGenerator) {
	g.handle = hg.Next()
}

// AddEntity adds entities to Group.
func (g *Group) AddEntity(es ...entity.Entity) {
	for _, e := range es {
		e.SetBlockRecord(g)
	}
	g.entities = append(g.entities, es...)
}
