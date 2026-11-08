package drawing

import (
	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/class"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/header"
	"github.com/edanko/dxf/object"
	"github.com/edanko/dxf/table"
)

// RawTag is a group code/value pair retained for opaque sections.
type RawTag struct {
	Code  int
	Value string
}

// RecoveryReport describes what a recovery load repaired or skipped.
type RecoveryReport struct {
	Issues            []string
	RecoveredSections []string
	SkippedEntities   int
}

// SetHeader replaces the HEADER section.
func (d *Drawing) SetHeader(h *header.Header) {
	if h != nil {
		d.headerSection = h
	}
}

// Classes returns the CLASSES section.
func (d *Drawing) Classes() class.Classes { return d.classesSection }

// SetClasses replaces the CLASSES section.
func (d *Drawing) SetClasses(c class.Classes) { d.classesSection = c }

// SetTables replaces the TABLES section and rebuilds the layer/style maps.
func (d *Drawing) SetTables(t table.Tables) {
	d.tablesSection = t
	d.Layers = make(map[string]*table.Layer)
	d.Styles = make(map[string]*table.Style)
	var firstLayer *table.Layer
	var firstStyle *table.Style
	for _, tbl := range t {
		switch tbl.Name() {
		case "LAYER":
			for _, entry := range tbl.Entries() {
				if l, ok := entry.(*table.Layer); ok {
					d.Layers[l.Name()] = l
					if firstLayer == nil {
						firstLayer = l
					}
				}
			}
		case "STYLE":
			for _, entry := range tbl.Entries() {
				if s, ok := entry.(*table.Style); ok {
					d.Styles[s.Name()] = s
					if firstStyle == nil {
						firstStyle = s
					}
				}
			}
		}
	}
	// Resolve layer line-type placeholders now that all tables are known.
	if lt := t.Find("LTYPE"); lt != nil {
		for _, l := range d.Layers {
			if l.LineType == nil {
				continue
			}
			if found, err := lt.Contains(l.LineType.Name()); err == nil {
				if resolved, ok := found.(*table.LineType); ok {
					l.SetLineType(resolved)
				}
			}
		}
	}
	if l := d.Layers["0"]; l != nil {
		d.CurrentLayer = l
	} else if firstLayer != nil {
		d.CurrentLayer = firstLayer
	}
	if s := d.Styles["STANDARD"]; s != nil {
		d.CurrentStyle = s
	} else if firstStyle != nil {
		d.CurrentStyle = firstStyle
	}
}

// Blocks returns the BLOCKS section.
func (d *Drawing) Blocks() block.Blocks { return d.blocksSection }

// SetBlocks replaces the BLOCKS section.
func (d *Drawing) SetBlocks(b block.Blocks) { d.blocksSection = b }

// Objects returns the OBJECTS section.
func (d *Drawing) Objects() object.Objects { return d.objectsSection }

// SetObjects replaces the OBJECTS section.
func (d *Drawing) SetObjects(o object.Objects) { d.objectsSection = o }

// GetObject returns an object by handle.
func (d *Drawing) GetObject(handle string) (object.Object, bool) {
	for _, o := range d.objectsSection {
		if o.Handle() == handle {
			return o, true
		}
	}
	return nil, false
}

// SetOpaqueSections stores sections that are retained as raw tags.
func (d *Drawing) SetOpaqueSections(sections map[string][]RawTag) {
	d.opaqueSections = sections
}

// OpaqueSections returns the retained opaque sections.
func (d *Drawing) OpaqueSections() map[string][]RawTag { return d.opaqueSections }

// SetSectionOrder records the order sections appeared in the source file.
func (d *Drawing) SetSectionOrder(order []string) { d.sectionOrder = order }

// SectionOrder returns the source section order.
func (d *Drawing) SectionOrder() []string { return d.sectionOrder }

// SetComments records code 999 comments per section.
func (d *Drawing) SetComments(comments map[string][]string) { d.comments = comments }

// Comments returns code 999 comments per section.
func (d *Drawing) Comments() map[string][]string { return d.comments }

// SetRecoveryReport records the outcome of a recovery load.
func (d *Drawing) SetRecoveryReport(r *RecoveryReport) { d.recoveryReport = r }

// RecoveryReport returns the recovery report, or nil for a normal load.
func (d *Drawing) RecoveryReport() *RecoveryReport { return d.recoveryReport }

// SetEntityLayer resolves an entity layer by name, falling back to layer 0.
func (d *Drawing) SetEntityLayer(e entity.Entity, name string) {
	if e == nil {
		return
	}
	if name == "" {
		name = "0"
	}
	if l := d.Layers[name]; l != nil {
		e.SetLayer(l)
		return
	}
	if setter, ok := e.(interface{ SetLayerName(string) }); ok {
		setter.SetLayerName(name)
	}
}

// observeHandles advances the handle generator past every handle already
// present in the document, so newly created objects never reuse a handle.
func (d *Drawing) observeHandles() {
	observe := func(h string) {
		if h != "" && d.handleGenarator != nil {
			d.handleGenarator.Observe(h)
		}
	}
	for _, t := range d.tablesSection {
		observe(t.Handle())
		for _, e := range t.Entries() {
			observe(e.Handle())
		}
	}
	for _, b := range d.blocksSection {
		observe(b.Handle())
		for _, e := range b.Entities() {
			observe(e.Handle())
		}
	}
	for _, e := range d.entitiesSection {
		observe(e.Handle())
	}
	for _, o := range d.objectsSection {
		observe(o.Handle())
	}
}

func (d *Drawing) writeComments(section string) {
	if d.formatter == nil {
		return
	}
	for _, comment := range d.comments[section] {
		d.formatter.WriteString(999, comment)
	}
}

func (d *Drawing) writeOpaqueSections() {
	if len(d.opaqueSections) == 0 || d.formatter == nil {
		return
	}
	written := make(map[string]bool)
	write := func(name string, tags []RawTag) {
		d.formatter.WriteString(0, "SECTION")
		d.formatter.WriteString(2, name)
		for _, tag := range tags {
			d.formatter.WriteString(tag.Code, tag.Value)
		}
		d.formatter.WriteString(0, "ENDSEC")
		written[name] = true
	}
	for _, name := range d.sectionOrder {
		if tags, ok := d.opaqueSections[name]; ok && !isKnownSection(name) {
			write(name, tags)
		}
	}
	for name, tags := range d.opaqueSections {
		if !written[name] && !isKnownSection(name) {
			write(name, tags)
		}
	}
}

func isKnownSection(name string) bool {
	switch name {
	case "HEADER", "CLASSES", "TABLES", "BLOCKS", "ENTITIES", "OBJECTS":
		return true
	}
	return false
}
