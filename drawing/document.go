package drawing

import (
	"strings"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/class"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/header"
	"github.com/edanko/dxf/lldxf"
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
		if strings.EqualFold(o.Handle(), handle) {
			return o, true
		}
	}
	return nil, false
}

// RootDictionary returns the OBJECTS-section dictionary with no owner.
func (d *Drawing) RootDictionary() (*object.Dictionary, bool) {
	for _, o := range d.objectsSection {
		if dict, ok := o.(*object.Dictionary); ok && (dict.Owner == "" || dict.Owner == "0") {
			return dict, true
		}
	}
	return nil, false
}

// GetObjectByDictionaryKey resolves a root-dictionary entry by key.
func (d *Drawing) GetObjectByDictionaryKey(key string) (object.Object, bool) {
	root, ok := d.RootDictionary()
	if !ok {
		return nil, false
	}
	return d.GetObjectByKey(root.Handle(), key)
}

// GetObjectByKey resolves a dictionary entry by dictionary handle and key.
func (d *Drawing) GetObjectByKey(dictHandle, key string) (object.Object, bool) {
	dictObj, ok := d.GetObject(dictHandle)
	if !ok {
		return nil, false
	}
	dict, ok := dictObj.(*object.Dictionary)
	if !ok {
		return nil, false
	}
	entry, ok := dict.GetEntry(key)
	if !ok {
		return nil, false
	}
	return d.GetObject(entry.Value)
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

// SetComments records code 999 comments per section. It replaces any
// positioned comments loaded from a file; use SetPositionedComments to restore
// exact positions.
func (d *Drawing) SetComments(comments map[string][]string) {
	d.comments = comments
	d.positionedComments = nil
}

// SetPositionedComments records code 999 comments anchored to record positions
// inside their sections.
func (d *Drawing) SetPositionedComments(comments []lldxf.PositionedComment) {
	d.positionedComments = append([]lldxf.PositionedComment(nil), comments...)
}

// PositionedComments returns the positioned comments currently retained.
func (d *Drawing) PositionedComments() []lldxf.PositionedComment {
	return append([]lldxf.PositionedComment(nil), d.positionedComments...)
}

// Comments returns code 999 comments per section.
func (d *Drawing) Comments() map[string][]string { return d.comments }

// SetRecoveryReport records the outcome of a recovery load.
func (d *Drawing) SetRecoveryReport(r *RecoveryReport) { d.recoveryReport = r }

// RecoveryReport returns the recovery report, or nil for a normal load.
func (d *Drawing) RecoveryReport() *RecoveryReport { return d.recoveryReport }

// SetSourceVersion sets the source DXF version declaration from the original file.
func (d *Drawing) SetSourceVersion(v string) { d.sourceVersion = v }

// SourceVersion returns the source DXF version.
func (d *Drawing) SourceVersion() string { return d.sourceVersion }

// SetNormalizedVersion sets the normalized DXF version (after load normalization).
func (d *Drawing) SetNormalizedVersion(v string) { d.normalizedVersion = v }

// NormalizedVersion returns the normalized DXF version.
func (d *Drawing) NormalizedVersion() string { return d.normalizedVersion }

// SetTargetVersion sets the target version for writing (if different from current).
func (d *Drawing) SetTargetVersion(v string) { d.targetVersion = v }

// TargetVersion returns the target version for writing.
func (d *Drawing) TargetVersion() string { return d.targetVersion }

// SetCodepage sets the DWG codepage from the source file.
func (d *Drawing) SetCodepage(p string) { d.codepage = p }

// Codepage returns the DWG codepage.
func (d *Drawing) Codepage() string { return d.codepage }

// SetConversion records how the document was normalized or converted.
func (d *Drawing) SetConversion(r *header.ConversionReport) { d.conversion = r }

// Conversion returns the conversion report, or nil if no normalization occurred.
func (d *Drawing) Conversion() *header.ConversionReport { return d.conversion }

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
	if setter, ok := e.(entity.LayerNameSetter); ok {
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
		observe(b.EndHandle())
		observeBlockEntities(b, observe)
	}
	for _, e := range d.entitiesSection {
		observe(e.Handle())
	}
	for _, o := range d.objectsSection {
		observe(o.Handle())
	}
}

func isKnownSection(name string) bool {
	switch name {
	case "HEADER", "CLASSES", "TABLES", "BLOCKS", "ENTITIES", "OBJECTS":
		return true
	}
	return false
}

// RenameNamesForCodepage re-decodes names of layers, styles, line types, and
// blocks that were read as raw codepage bytes by the ASCII tagger.
// observeBlockRecords feeds every handle in the block's retained raw tags to
// the generator so nested blocks and unbuildable records never collide.
func observeBlockRecords(b *block.Block, observe func(string)) {
	for _, t := range b.RawTags() {
		if t.Code == 5 {
			observe(t.Value)
		}
	}
}

func observeBlockEntities(b *block.Block, observe func(string)) {
	for _, e := range b.Entities() {
		observe(e.Handle())
		if nb, ok := e.(*block.Block); ok {
			observe(nb.Handle())
			observe(nb.EndHandle())
			observeBlockRecords(nb, observe)
			observeBlockEntities(nb, observe)
		}
	}
}

func (d *Drawing) RenameNamesForCodepage(codepage string) {
	decode := func(name string) string {
		if name == "" {
			return name
		}
		return lldxf.DecodeCodepageBytes([]byte(name), codepage)
	}

	newLayers := make(map[string]*table.Layer, len(d.Layers))
	for name, l := range d.Layers {
		n := decode(name)
		if n != name {
			l.SetName(n)
		}
		newLayers[n] = l
	}
	d.Layers = newLayers
	if d.CurrentLayer != nil {
		d.CurrentLayer = d.Layers[decode(d.CurrentLayer.Name())]
	}

	newStyles := make(map[string]*table.Style, len(d.Styles))
	for name, st := range d.Styles {
		n := decode(name)
		if n != name {
			st.SetName(n)
		}
		newStyles[n] = st
	}
	d.Styles = newStyles
	if d.CurrentStyle != nil {
		d.CurrentStyle = d.Styles[decode(d.CurrentStyle.Name())]
	}

	for _, tbl := range d.tablesSection {
		switch tbl.Name() {
		case "LAYER":
			for _, e := range tbl.Entries() {
				if l, ok := e.(*table.Layer); ok {
					l.SetName(decode(l.Name()))
				}
			}
		case "STYLE":
			for _, e := range tbl.Entries() {
				if st, ok := e.(*table.Style); ok {
					st.SetName(decode(st.Name()))
				}
			}
		case "LTYPE":
			for _, e := range tbl.Entries() {
				if lt, ok := e.(*table.LineType); ok {
					lt.SetName(decode(lt.Name()))
				}
			}
		}
	}

	for _, b := range d.blocksSection {
		b.Name = decode(b.Name)
		b.Description = decode(b.Description)
	}
}
