package table

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

// fixedHandle is a handle.Handler with a value loaded from a DXF file.
type fixedHandle string

func (h fixedHandle) Handle() string                   { return string(h) }
func (h fixedHandle) SetHandle(*handle.HandleGenerator) {}

// RawSymbolTable retains a table entry that has no typed model yet.
type RawSymbolTable struct {
	typeName string
	name     string
	handle   string
	owner    handle.Handler
	tags     lldxf.Tags
}

// IsSymbolTable is for SymbolTable interface.
func (r *RawSymbolTable) IsSymbolTable() bool { return true }

// Name returns the entry name (code 2).
func (r *RawSymbolTable) Name() string { return r.name }

// Handle returns the entry handle.
func (r *RawSymbolTable) Handle() string { return r.handle }

// SetHandle sets the entry handle when it has not been loaded.
func (r *RawSymbolTable) SetHandle(hg *handle.HandleGenerator) {
	if r.handle == "" {
		r.handle = hg.Next()
	}
}

// SetOwner sets the table owner.
func (r *RawSymbolTable) SetOwner(h handle.Handler) { r.owner = h }

// Format writes the raw entry.
func (r *RawSymbolTable) Format(f format.Formatter) {
	f.WriteString(0, r.typeName)
	for _, tag := range r.tags {
		if tag.Code() == 0 {
			continue
		}
		f.WriteString(int(tag.Code()), tagString(tag))
	}
}

// ParseTables parses a TABLES section into typed tables, retaining unknown
// table and entry types as raw entries.
func ParseTables(tags lldxf.Tags) Tables {
	tables := make(Tables, 0)
	i := 0
	for i < len(tags) {
		if !isTag(tags[i], 0, "TABLE") {
			i++
			continue
		}
		i++
		t := NewTable("")
		entryTags := make([]lldxf.Tags, 0)
		var current lldxf.Tags
		for i < len(tags) {
			tag := tags[i]
			if tag.Code() == 0 {
				value := strings.ToUpper(tagString(tag))
				switch value {
				case "ENDTAB":
					i++
					goto tableDone
				default:
					if len(current) > 0 {
						entryTags = append(entryTags, current)
					}
					current = lldxf.Tags{tag}
					i++
					continue
				}
			}
			if len(current) == 0 {
				// Table header tags (name, handle, count, subclass markers).
				switch tag.Code() {
				case 2:
					t.name = tagString(tag)
				case 5:
					t.handle = tagString(tag)
				}
			} else {
				current = append(current, tag)
			}
			i++
		}
	tableDone:
		if len(current) > 0 {
			entryTags = append(entryTags, current)
		}
		for _, et := range entryTags {
			t.Add(ParseTableEntry(t.name, et))
		}
		tables = append(tables, t)
	}
	return tables
}

// ParseTableEntry parses one symbol-table entry.
func ParseTableEntry(tableName string, tags lldxf.Tags) SymbolTable {
	typeName := strings.ToUpper(tags.DXFType())
	name := tagStringByCode(tags, 2)
	if name == "" {
		name = typeName
	}
	var st SymbolTable

	switch typeName {
	case "LAYER":
		ltName := tagStringByCode(tags, 6)
		if ltName == "" {
			ltName = "Continuous"
		}
		l := NewLayer(name, color.ColorNumber(tagIntByCode(tags, 62)), NewLineType(ltName, ""))
		l.handle = tags.GetHandle()
		l.flag = tagIntByCode(tags, 70)
		if w, ok := tagIntByCodeOK(tags, 370); ok {
			l.lineWidth = w
		}
		if ps := tagStringByCode(tags, 390); ps != "" {
			l.PlotStyle = fixedHandle(ps)
		}
		st = l
	case "LTYPE":
		lengths := make([]float64, 0)
		for _, tag := range tags {
			if tag.Code() == 49 {
				lengths = append(lengths, tagFloat(tag))
			}
		}
		lt := NewLineType(name, tagStringByCode(tags, 3), lengths...)
		lt.handle = tags.GetHandle()
		st = lt
	case "STYLE":
		s := NewStyle(name)
		s.handle = tags.GetHandle()
		s.FixedTextHeight = tagFloatByCode(tags, 40)
		s.WidthFactor = tagFloatByCode(tags, 41)
		s.LastHeightUsed = tagFloatByCode(tags, 42)
		s.ObliqueAngle = tagFloatByCode(tags, 50)
		s.FontName = tagStringByCode(tags, 3)
		s.BigFontName = tagStringByCode(tags, 4)
		st = s
	case "BLOCK_RECORD":
		b := NewBlockRecord(name)
		b.handle = tags.GetHandle()
		st = b
	case "APPID":
		a := NewAppID(name)
		a.handle = tags.GetHandle()
		st = a
	case "DIMSTYLE":
		d := NewDimStyle(name)
		d.handle = tags.GetHandle()
		st = d
	case "UCS":
		u := NewUCS(name)
		u.handle = tags.GetHandle()
		st = u
	case "VIEW":
		v := NewView(name)
		v.handle = tags.GetHandle()
		st = v
	case "VPORT":
		v := NewViewport(name)
		v.handle = tags.GetHandle()
		v.LowerLeft = point2(tags, 10)
		v.UpperRight = point2(tags, 11)
		v.ViewCenter = point2(tags, 12)
		v.SnapBase = point2(tags, 13)
		v.SnapSpacing = point2(tags, 14)
		v.GridSpacing = point2(tags, 15)
		v.ViewDirection = point3(tags, 16)
		v.ViewTarget = point3(tags, 17)
		v.Height = tagFloatByCode(tags, 40)
		v.AspectRatio = tagFloatByCode(tags, 41)
		v.LensLength = tagFloatByCode(tags, 42)
		v.FrontClip = tagFloatByCode(tags, 43)
		v.BackClip = tagFloatByCode(tags, 44)
		v.SnapAngle = tagFloatByCode(tags, 50)
		v.TwistAngle = tagFloatByCode(tags, 51)
		st = v
	default:
		st = &RawSymbolTable{
			typeName: typeName,
			name:     name,
			handle:   tags.GetHandle(),
			tags:     tags,
		}
	}
	if st != nil {
		st.SetOwner(nil)
	}
	return st
}

func isTag(tag lldxf.Tag, code int, value string) bool {
	return int(tag.Code()) == code && strings.EqualFold(tagString(tag), value)
}

func tagString(tag lldxf.Tag) string {
	return fmt.Sprintf("%v", tag.Value())
}

func tagStringByCode(tags lldxf.Tags, code int) string {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			return tagString(tag)
		}
	}
	return ""
}

func tagFloat(tag lldxf.Tag) float64 {
	switch v := tag.Value().(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		var f float64
		fmt.Sscanf(tagString(tag), "%f", &f)
		return f
	}
}

func tagFloatByCode(tags lldxf.Tags, code int) float64 {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			return tagFloat(tag)
		}
	}
	return 0
}

func tagIntByCode(tags lldxf.Tags, code int) int {
	v, _ := tagIntByCodeOK(tags, code)
	return v
}

func tagIntByCodeOK(tags lldxf.Tags, code int) (int, bool) {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			switch v := tag.Value().(type) {
			case int:
				return v, true
			case int64:
				return int(v), true
			case float64:
				return int(v), true
			default:
				var n int
				if _, err := fmt.Sscanf(tagString(tag), "%d", &n); err == nil {
					return n, true
				}
			}
		}
	}
	return 0, false
}

func point2(tags lldxf.Tags, base int) []float64 {
	return []float64{tagFloatByCode(tags, base), tagFloatByCode(tags, base+10)}
}

func point3(tags lldxf.Tags, base int) []float64 {
	return []float64{tagFloatByCode(tags, base), tagFloatByCode(tags, base+10), tagFloatByCode(tags, base+20)}
}

// Find returns the named table, or nil.
func (ts Tables) Find(name string) *Table {
	for _, t := range ts {
		if strings.EqualFold(t.Name(), name) {
			return t
		}
	}
	return nil
}

// Name returns the table name.
func (t *Table) Name() string { return t.name }

// SetHandleString sets an explicit table handle.
func (t *Table) SetHandleString(h string) { t.handle = h }

// Entries returns the table entries.
func (t *Table) Entries() []SymbolTable { return t.tables }
