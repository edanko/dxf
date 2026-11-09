package drawing

import (
	"sort"
	"strings"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/header"
	"github.com/edanko/dxf/table"
)

// r12Tables lists the symbol tables R12 supports.
var r12Tables = map[string]bool{
	"$ACAD_PROXY_TABLES": true,
	"LTYPE":              true,
	"STYLE":              true,
	"VIEW":               true,
	"UCS":                true,
	"APPID":              true,
	"DIMSTYLE":           true,
	"BLOCK_RECORD":       true,
	"LAYER":              true,
	"VPORT":              true,
}

// r12EntityTypes lists entity types R12 supports.
var r12EntityTypes = map[string]bool{
	"POLYLINE": true, "VERTEX": true, "3DFACE": true, "LINE": true,
	"ENDMK": true, "INSERT": true, "ARC": true, "SHAPE": true,
	"SOLID": true, "TRACE": true, "SEQEND": true, "TEXT": true,
	"CIRCLE": true, "POINT": true, "3DLINE": true,
}

// targetOf reports the version a formatter writes.
func targetOf(f format.Formatter) string { return f.Version() }

// recordConversionDrop merges dropped-content notes into the conversion report.
func (d *Drawing) recordConversionDrop(target string, dropped []string) {
	source := d.normalizedVersion
	if source == "" && d.headerSection != nil {
		source = d.headerSection.Version
	}
	if d.conversion == nil {
		d.conversion = &header.ConversionReport{SourceVersion: source, NormalizedVersion: target}
	}
	d.conversion.Dropped = append(d.conversion.Dropped, dedup(dropped)...)
}

// sanitizeTables returns an emit function that writes only R12-supported tables.
func (d *Drawing) sanitizeTables(dropped *[]string) func(f format.Formatter) {
	return func(f format.Formatter) {
		f.WriteString(0, "SECTION")
		f.WriteString(2, "TABLES")
		for _, tbl := range d.tablesSection {
			name := strings.ToUpper(tbl.Name())
			if !r12Tables[name] {
				*dropped = append(*dropped, "table "+name)
				continue
			}
			var kept []table.SymbolTable
			for _, entry := range tbl.Entries() {
				if lt, ok := entry.(*table.LineType); ok && strings.EqualFold(lt.Name(), "ByBlock") {
					*dropped = append(*dropped, "linetype ByBlock")
					continue
				}
				dropR12LayerFields(entry, dropped)
				kept = append(kept, entry)
			}
			tbl.SetEntries(kept)
			tbl.Format(f)
		}
		f.WriteString(0, "ENDSEC")
	}
}

// dropR12LayerFields records layer features R12 cannot store.
func dropR12LayerFields(entry table.SymbolTable, dropped *[]string) {
	l, ok := entry.(*table.Layer)
	if !ok {
		return
	}
	if l.Color < 0 {
		*dropped = append(*dropped, "layer transparency on "+l.Name())
	}
	if w := l.LineWidth(); w >= 0 {
		*dropped = append(*dropped, "lineweight on layer "+l.Name())
	}
	if l.PlotStyle != nil && l.PlotStyle.Handle() != "" {
		*dropped = append(*dropped, "plot style on layer "+l.Name())
	}
}

// sanitizeEntities returns an emit function that drops non-R12 entities.
func (d *Drawing) sanitizeEntities(dropped *[]string) func(f format.Formatter) {
	return func(f format.Formatter) {
		f.WriteString(0, "SECTION")
		f.WriteString(2, "ENTITIES")
		for _, e := range d.entitiesSection {
			if !isR12Entity(e.DXFType()) {
				*dropped = append(*dropped, "entity "+e.DXFType())
				continue
			}
			dropR12EntityFields(e, dropped)
			e.Format(f)
		}
		f.WriteString(0, "ENDSEC")
	}
}

// sanitizeBlocks drops block-embedded entities R12 cannot store.
func (d *Drawing) sanitizeBlocks(dropped *[]string) func(f format.Formatter) {
	return func(f format.Formatter) {
		for _, blk := range d.blocksSection {
			var kept []block.BlockEntity
			for _, e := range blk.Entities() {
				if !isR12Entity(e.DXFType()) {
					*dropped = append(*dropped, "entity "+e.DXFType()+" in block "+blk.Name)
					continue
				}
				dropR12BlockEntityFields(e, dropped)
				kept = append(kept, e)
			}
			if len(kept) != len(blk.Entities()) {
				blk.SetEntities(kept)
			}
		}
		d.blocksSection.Format(f)
	}
}

func dropR12BlockEntityFields(e block.BlockEntity, dropped *[]string) {
	dropRawTagsCodes(e, e.DXFType(), dropped)
}

func isR12Entity(t string) bool { return r12EntityTypes[strings.ToUpper(t)] }

// dropR12EntityFields records true-color and transparency fields R12 cannot store.
func dropR12EntityFields(e entity.Entity, dropped *[]string) {
	if g, ok := e.(interface{ ColorRGB() int }); ok && g.ColorRGB() != 0 {
		*dropped = append(*dropped, "true color on "+e.DXFType())
	}
	if g, ok := e.(interface{ Transparency() float64 }); ok && g.Transparency() != 0 {
		*dropped = append(*dropped, "transparency on "+e.DXFType())
	}
	dropRawTagsCodes(rawAdapter{e}, e.DXFType(), dropped)
}

// rawAdapter exposes an entity's retained group codes through the structural
// interface used by dropRawTagsCodes.
type rawAdapter struct{ e entity.Entity }

func (a rawAdapter) RawTags() []block.TagPair {
	r, ok := a.e.(interface{ RawTags() []entity.TagPair })
	if !ok {
		return nil
	}
	out := make([]block.TagPair, 0, len(r.RawTags()))
	for _, t := range r.RawTags() {
		out = append(out, block.TagPair{Code: t.Code, Value: t.Value})
	}
	return out
}

func (a rawAdapter) SetRawTags(tags []block.TagPair) {
	s, ok := a.e.(interface{ SetRawTags([]entity.TagPair) })
	if !ok {
		return
	}
	out := make([]entity.TagPair, 0, len(tags))
	for _, t := range tags {
		out = append(out, entity.TagPair{Code: t.Code, Value: t.Value})
	}
	s.SetRawTags(out)
}

// r12RawCodes are group codes R12 cannot store (true color, transparency).
var r12RawCodes = map[int]string{420: "true color", 440: "transparency"}

type rawTagHolder interface {
	RawTags() []block.TagPair
	SetRawTags([]block.TagPair)
}

func dropRawTagsCodes(r rawTagHolder, typeName string, dropped *[]string) {
	tags := r.RawTags()
	var kept []block.TagPair
	for _, tag := range tags {
		if what, bad := r12RawCodes[tag.Code]; bad {
			*dropped = append(*dropped, what+" on "+typeName)
			continue
		}
		kept = append(kept, tag)
	}
	if len(kept) != len(tags) {
		r.SetRawTags(kept)
	}
}

// dedup returns sorted unique notes.
func dedup(notes []string) []string {
	set := map[string]bool{}
	for _, n := range notes {
		set[n] = true
	}
	out := make([]string, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
