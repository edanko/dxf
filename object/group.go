package object

import (
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// Group represents GROUP Object.
type Group struct {
	Name          string
	Description   string
	handle        string
	owner         handle.Handler
	entities      []entity.Entity
	entityHandles []string
	selectable    bool
}

// SetHandleString sets an explicit handle loaded from a file.
func (g *Group) SetHandleString(h string) { g.handle = h }

// SetEntityHandles records entity handles until the entity database is bound.
func (g *Group) SetEntityHandles(hs []string) { g.entityHandles = hs }

// EntityHandles returns the entity handles recorded for the group.
func (g *Group) EntityHandles() []string { return g.entityHandles }

// Handle returns the group's handle.
func (g *Group) Handle() string {
	return g.handle
}

// SetHandle sets the group's handle.
func (g *Group) SetHandle(hg *handle.HandleGenerator) {
	if g.handle == "" {
		g.handle = hg.Next()
	}
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
	d.AddItem(g.Name, g)
	return nil
}

// Format writes data to formatter.
func (g *Group) Format(f format.Formatter) {
	f.WriteString(0, "GROUP")
	f.WriteString(5, g.handle)
	if g.owner != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, g.owner.Handle())
		f.WriteString(102, "}")
		f.WriteString(330, g.owner.Handle())
	}
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
	for _, h := range g.entityHandles {
		f.WriteString(340, h)
	}
}

// AddEntity adds entities to Group.
func (g *Group) AddEntity(es ...entity.Entity) {
	for _, e := range es {
		e.SetBlockRecord(g)
	}
	g.entities = append(g.entities, es...)
}
