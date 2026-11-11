package object

import (
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

// Group represents GROUP Object.
type Group struct {
	Name          string
	Description   string
	handle        string
	owner         handle.Handler
	reactors      []string // parsed {ACAD_REACTORS links; Format re-emits the group
	groupTags     int      // number of reactor-group tags in the source record
	entities      []entity.Entity
	entityHandles []string
	selectable    bool
	retained      lldxf.Tags
}

// SetReactorHandles stores the parsed reactor links and their source tag span.
func (g *Group) SetReactorHandles(hs []string, groupTags int) {
	g.reactors = hs
	g.groupTags = groupTags
}

// SetRetainedTags records tags the typed model does not emit.
func (g *Group) SetRetainedTags(tags lldxf.Tags) { g.retained = tags }

// RetainedTags returns the retained tag stream.
func (g *Group) RetainedTags() lldxf.Tags { return g.retained }

// SetHandleString sets an explicit handle loaded from a file.
func (g *Group) SetHandleString(h string) { g.handle = h }

// SetEntityHandles records entity handles until the entity database is bound.
func (g *Group) SetEntityHandles(hs []string) { g.entityHandles = hs }

// EntityHandles returns the entity handles recorded for the group.
func (g *Group) EntityHandles() []string { return g.entityHandles }

// Entities returns the group member entities attached at runtime.
func (g *Group) Entities() []entity.Entity { return g.entities }

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
	if len(g.reactors) > 0 {
		f.WriteString(102, "{ACAD_REACTORS")
		for _, r := range g.reactors {
			f.WriteString(330, r)
		}
		f.WriteString(102, "}")
	}
	if g.owner != nil {
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
	for _, tag := range g.retained {
		f.WriteString(int(tag.Code()), lldxf.FormatValue(tag))
	}
}

// AddEntity adds entities to Group.
func (g *Group) AddEntity(es ...entity.Entity) {
	for _, e := range es {
		e.SetBlockRecord(g)
	}
	g.entities = append(g.entities, es...)
}
