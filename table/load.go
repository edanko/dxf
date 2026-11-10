package table

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

type ownerHandleSetter interface {
	SetOwnerHandle(string)
}

type tagRetainer interface {
	RetainTags(lldxf.Tags)
}

// point2First returns the first pair of codes base/base+1 (repeats dropped).
func point2First(tags lldxf.Tags, base int) []float64 {
	return []float64{tagFloatByCodeFirst(tags, base), tagFloatByCodeFirst(tags, base+1)}
}

// point3First returns the first triple of codes base/base+1/base+2.
func point3First(tags lldxf.Tags, base int) []float64 {
	return []float64{
		tagFloatByCodeFirst(tags, base),
		tagFloatByCodeFirst(tags, base+1),
		tagFloatByCodeFirst(tags, base+2),
	}
}

// tagFloatByCodeFirst returns the value of the first tag with the given code.
func tagFloatByCodeFirst(tags lldxf.Tags, code int) float64 {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			return tagFloat(tag)
		}
	}
	return 0
}

// fixedHandle is a handle.Handler with a value loaded from a DXF file.
type fixedHandle string

func (h fixedHandle) Handle() string                    { return string(h) }
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

// SetName renames the raw entry.
func (r *RawSymbolTable) SetName(name string) { r.name = name }

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
		dimstyle := false
		inDimTail := false     // inside the DIMSTYLE header tail (71 count + 340 links)
		var dimTail lldxf.Tags // tail captured at the table header (71 count + 340 links)
		for i < len(tags) {
			tag := tags[i]
			if tag.Code() == 0 {
				value := strings.ToUpper(tagString(tag))
				switch value {
				case "ENDTAB":
					i++
					goto tableDone
				default:
					if inDimTail {
						inDimTail = false
					}
					if len(current) > 0 {
						entryTags = append(entryTags, current)
					}
					current = lldxf.Tags{tag}
					i++
					continue
				}
			}
			if inDimTail {
				dimTail = append(dimTail, tag)
				i++
				continue
			}
			if len(current) == 0 {
				// Table header tags (name, handle, count, subclass markers).
				switch int(tag.Code()) {
				case 2:
					t.name = tagString(tag)
				case 5:
					t.handle = tagString(tag)
				case 100:
					if strings.EqualFold(strings.ToUpper(tagString(tag)), "AcDbDimStyleTable") {
						dimstyle = true
					}
				case 71:
					if dimstyle && len(current) == 0 {
						inDimTail = true
						dimTail = lldxf.Tags{tag}
					}
				}
			} else {
				current = append(current, tag)
			}
			i++
		}
	tableDone:
		if len(dimTail) > 0 && len(current) > 0 {
			current = append(current, dimTail...)
			dimTail = nil
		}
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
	typeName := recordTypeName(tags)
	if typeName == "" {
		typeName = strings.ToUpper(tags.DXFType())
	}
	name := tagStringByCode(tags, 2)
	var st SymbolTable

	switch typeName {
	case "LAYER":
		ltName := tagStringByCode(tags, 6)
		if ltName == "" {
			ltName = "Continuous"
		}
		l := NewLayer(name, color.ColorNumber(tagIntByCode(tags, 62)), NewLineType(ltName, ""))
		l.SetHandleValue(tags.GetHandle())
		if v, ok := tagIntByCodeOK(tags, 70); ok {
			l.SetFlags(v)
		}
		if w, ok := tagIntByCodeOK(tags, 370); ok {
			l.lineWidth = w
		}
		if t, ok := tagIntByCodeOK(tags, 440); ok {
			l.SetTransparency(t)
		}
		if p, ok := tagIntByCodeOK(tags, 290); ok {
			l.SetPlottable(p != 0)
		}
		if tc, ok := tagIntByCodeOK(tags, 420); ok {
			l.SetTrueColor(tc)
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
		lt.SetHandleValue(tags.GetHandle())
		if v70, ok := tagIntByCodeOK(tags, 70); ok {
			lt.SetFlags(v70)
		}
		if a, ok := tagIntByCodeOK(tags, 72); ok {
			lt.SetAlignment(a)
		}
		if n, ok := tagIntByCodeOK(tags, 73); ok && n > len(lengths) {
			// Complex linetype: per-element records follow each code 49.
			lt.lengths = nil
			var e LineTypeElement
			inElement := false
			for _, tag := range tags {
				switch int(tag.Code()) {
				case 49:
					if inElement {
						lt.AddElement(e)
					}
					e = LineTypeElement{DashLength: tagFloat(tag)}
					inElement = true
				case 74:
					e.SegLen = int(tagValueInt(tag))
				case 75:
					e.ShapeNum = int(tagValueInt(tag))
				case 9:
					e.TextValue = tagString(tag)
				case 50:
					e.Rotation = tagFloat(tag)
				case 46:
					e.XScale = tagFloat(tag)
				case 44:
					e.XOffset = tagFloat(tag)
				case 45:
					e.YOffset = tagFloat(tag)
				}
			}
			if inElement {
				lt.AddElement(e)
			}
		}
		st = lt
	case "STYLE":
		s := NewStyle(name)
		s.SetHandleValue(tags.GetHandle())
		s.FixedTextHeight = tagFloatByCode(tags, 40)
		s.WidthFactor = tagFloatByCode(tags, 41)
		s.LastHeightUsed = tagFloatByCode(tags, 42)
		s.ObliqueAngle = tagFloatByCode(tags, 50)
		s.FontName = tagStringByCode(tags, 3)
		s.BigFontName = tagStringByCode(tags, 4)
		if v70, ok := tagIntByCodeOK(tags, 70); ok {
			s.SetFlags(v70)
		}
		st = s
	case "BLOCK_RECORD":
		b := NewBlockRecord(name)
		b.SetHandleValue(tags.GetHandle())
		if v70, ok := tagIntByCodeOK(tags, 70); ok {
			b.SetFlags(v70)
		}
		if v, ok := tagIntByCodeOK(tags, 280); ok {
			b.SetBlockScaling(v)
		}
		if v, ok := tagIntByCodeOK(tags, 281); ok {
			b.SetExploding(v)
		}
		if v := tagStringByCode(tags, 305); v != "" {
			b.SetXRefPath(v)
		}
		st = b
	case "APPID":
		a := NewAppID(name)
		a.SetHandleValue(tags.GetHandle())
		if v70, ok := tagIntByCodeOK(tags, 70); ok {
			a.SetFlags(v70)
		}
		st = a
	case "DIMSTYLE":
		d := NewDimStyle(name)
		d.SetHandleValue(tags.GetHandle())
		if v70, ok := tagIntByCodeOK(tags, 70); ok {
			d.SetFlags(v70)
		}
		for _, tag := range tags {
			code := int(tag.Code())
			if code == 0 || code == 2 || code == 5 || code == 102 || code == 105 || code == 330 || code == 70 || code == 100 || code == 350 {
				continue
			}
			d.SetVar(code, tagString(tag))
		}

		st = d
	case "UCS":
		u := NewUCS(name)
		u.SetHandleValue(tags.GetHandle())
		if v70, ok := tagIntByCodeOK(tags, 70); ok {
			u.SetFlags(v70)
		}
		st = u
	case "VIEW":
		v := NewView(name)
		v.SetHandleValue(tags.GetHandle())
		if v70, ok := tagIntByCodeOK(tags, 70); ok {
			v.SetFlags(v70)
		}
		st = v
	case "VPORT":
		v := NewViewport(name)
		v.SetHandleValue(tags.GetHandle())
		v.LowerLeft = point2First(tags, 10)
		v.UpperRight = point2First(tags, 11)
		v.ViewCenter = point2First(tags, 12)
		v.SnapBase = point2First(tags, 13)
		v.SnapSpacing = point2First(tags, 14)
		v.GridSpacing = point2First(tags, 15)
		v.ViewDirection = point3First(tags, 16)
		v.ViewTarget = point3First(tags, 17)
		v.Height = tagFloatByCodeFirst(tags, 40)
		v.AspectRatio = tagFloatByCodeFirst(tags, 41)
		v.LensLength = tagFloatByCodeFirst(tags, 42)
		v.FrontClip = tagFloatByCodeFirst(tags, 43)
		v.BackClip = tagFloatByCodeFirst(tags, 44)
		v.SnapAngle = tagFloatByCodeFirst(tags, 50)
		v.TwistAngle = tagFloatByCodeFirst(tags, 51)
		if f2, ok := tagIntByCodeOK(tags, 70); ok {
			v.SetFlags(f2)
		}
		if v71, ok := tagIntByCodeOK(tags, 71); ok {
			v.ViewportType = v71
		}
		if v72, ok := tagIntByCodeOK(tags, 72); ok {
			v.UCSIconSetting = v72
		}
		if v73, ok := tagIntByCodeOK(tags, 73); ok {
			v.SnapOn = v73
		}
		if v74, ok := tagIntByCodeOK(tags, 74); ok {
			v.GridOn = v74
		}
		if v75, ok := tagIntByCodeOK(tags, 75); ok {
			v.ZoomFlags = v75
		}
		st = v
	default:
		st = &RawSymbolTable{
			typeName: typeName,
			name:     name,
			handle:   tags.GetHandle(),
			tags:     tags,
		}
	}
	if oh := tagStringByCode(tags, 330); oh != "" {
		if ownerSetter, ok := st.(ownerHandleSetter); ok {
			ownerSetter.SetOwnerHandle(oh)
		}
	}
	if retainer, ok := st.(tagRetainer); ok {
		retainer.RetainTags(tags)
	}
	return st
}

func isTag(tag lldxf.Tag, code int, value string) bool {
	return int(tag.Code()) == code && strings.EqualFold(tagString(tag), value)
}

func tagString(tag lldxf.Tag) string {
	return lldxf.FormatValue(tag)
}

// recordTypeName returns the entry's type marker: the first code-0 value that
// is not a group-code number. DIMSTYLE records inside the DIMSTYLE table start
// mid-table (code 105 handle, no leading marker), and their stream contains
// numeric code-0 pairs (the table tail's 71 count followed by a 340 link);
// skipping those keeps record typing robust for such streams.
func recordTypeName(tags lldxf.Tags) string {
	for _, tag := range tags {
		if int(tag.Code()) != 0 {
			continue
		}
		v := strings.TrimSpace(tagString(tag))
		if v == "" || isNumericToken(v) {
			continue
		}
		return strings.ToUpper(v)
	}
	return ""
}

func isNumericToken(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
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

// tagValueInt coerces a tag value to int (tagger stores small ints as int8/int16/int32).
func tagValueInt(tag lldxf.Tag) int64 {
	switch v := tag.Value().(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		var n int64
		fmt.Sscanf(tagString(tag), "%d", &n)
		return n
	}
}
