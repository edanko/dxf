package object

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

type dxfTypeReader interface {
	DXFType() string
}

// fixedHandle is a handle.Handler with a value loaded from a DXF file.
type fixedHandle string

func (h fixedHandle) Handle() string                    { return string(h) }
func (h fixedHandle) SetHandle(*handle.HandleGenerator) {}

// RawObject retains an object type that has no typed model yet.
type RawObject struct {
	TypeName string
	handle   string
	tags     lldxf.Tags
}

// NewRawObject creates a raw object.
func NewRawObject(typeName string, tags lldxf.Tags) *RawObject {
	return &RawObject{TypeName: typeName, handle: tags.GetHandle(), tags: tags}
}

// IsObject is for Object interface.
func (o *RawObject) IsObject() bool { return true }

// Handle returns the object handle.
func (o *RawObject) Handle() string { return o.handle }

// SetHandle sets the handle when it has not been loaded.
func (o *RawObject) SetHandle(hg *handle.HandleGenerator) {
	if o.handle == "" {
		o.handle = hg.Next()
	}
}

// SetHandleString sets an explicit handle loaded from a file.
func (o *RawObject) SetHandleString(h string) { o.handle = h }

// Format writes the raw object.
func (o *RawObject) Format(f format.Formatter) {
	f.WriteString(0, o.TypeName)
	for _, tag := range o.tags {
		if tag.Code() == 0 {
			continue
		}
		f.WriteString(int(tag.Code()), tagString(tag))
	}
}

// DXFType returns the object type name.
func (o *RawObject) DXFType() string { return o.TypeName }

// RawTags returns the retained tags of the raw object.
func (o *RawObject) RawTags() lldxf.Tags { return o.tags }

// dictEmitCounts gives, per code, how many occurrences Dictionary.Format
// re-emits from its own state; retainObjectTags keeps only the surplus.
func dictEmitCounts(tags lldxf.Tags) map[int]int {
	nSoft := 0
	for _, tag := range tags {
		if int(tag.Code()) == 350 {
			nSoft++
		}
	}
	return map[int]int{0: 1, 5: 1, 330: 1, 280: 1, 3: len(entryKeys(tags)), 350: nSoft}
}

// wdfltEmitCounts matches DictionaryWithDefault.Format (flag + schema flag).
func wdfltEmitCounts(tags lldxf.Tags) map[int]int {
	n280 := min(countIntByCode(tags, 280),
		// Format emits at most the flag and the schema flag
		2)
	nSoftW := 0
	for _, tag := range tags {
		if int(tag.Code()) == 350 {
			nSoftW++
		}
	}
	return map[int]int{0: 1, 5: 1, 330: 1, 280: n280, 1: 1, 3: len(entryKeys(tags)), 350: nSoftW}
}

// dictVarEmitCounts matches DictionaryVar.Format (schema flag + value).
func dictVarEmitCounts(tags lldxf.Tags) map[int]int {
	n280 := min(countIntByCode(tags, 280),
		// Format emits the schema flag only
		1)
	return map[int]int{0: 1, 5: 1, 330: 1, 280: n280, 1: 1}
}

// entryKeys returns the dictionary entry keys in wire order.
func entryKeys(tags lldxf.Tags) []string {
	keys := make([]string, 0)
	key := ""
	for _, tag := range tags {
		switch int(tag.Code()) {
		case 3:
			key = tagString(tag)
		case 350, 360:
			if key != "" {
				keys = append(keys, key)
				key = ""
			}
		}
	}
	return keys
}

// groupEmitCounts matches Group.Format.
func groupEmitCounts(tags lldxf.Tags) map[int]int {
	n340 := 0
	for _, tag := range tags {
		if int(tag.Code()) == 340 {
			n340++
		}
	}
	return map[int]int{0: 1, 5: 1, 100: 1, 330: 1, 70: 1, 71: 1, 300: 1, 340: n340}
}

var placeholderEmitted = map[int]bool{0: true, 5: true, 100: true, 330: true}

// reactorsOfRecord returns the owner handles inside the record's
// {ACAD_REACTORS groups (code 102 block, code-330 members) and the number of
// group tags they span (each group: 102 opener, its 330 members, 102 closer).
func reactorsOfRecord(tags lldxf.Tags) ([]string, int) {
	out := make([]string, 0)
	groupTags := 0
	inReactor := false
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 102 {
			v := strings.ToUpper(strings.TrimSpace(tagString(tag)))
			switch v {
			case "{ACAD_REACTORS":
				inReactor = true
				groupTags++
			case "}":
				inReactor = false
				groupTags++
			}
			continue
		}
		if inReactor && code == 330 {
			out = append(out, tagString(tag))
			groupTags++
		}
	}
	return out, groupTags
}

// ownerOfRecord returns the record owner: the first code-330 outside a
// {ACAD_REACTORS group.
func ownerOfRecord(tags lldxf.Tags) string {
	inReactor := false
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 102 {
			v := strings.ToUpper(strings.TrimSpace(tagString(tag)))
			switch v {
			case "{ACAD_REACTORS":
				inReactor = true
			case "}":
				inReactor = false
			}
			continue
		}
		if !inReactor && code == 330 {
			return tagString(tag)
		}
	}
	return ""
}

// extDictOfRecord returns the handle inside a 102 {ACAD_XDICTIONARY group.
func extDictOfRecord(tags lldxf.Tags) string {
	inXDict := false
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 102 {
			v := strings.ToUpper(strings.TrimSpace(tagString(tag)))
			switch v {
			case "{ACAD_XDICTIONARY":
				inXDict = true
			case "}":
				inXDict = false
			}
			continue
		}
		if inXDict && code == 360 {
			return tagString(tag)
		}
	}
	return ""
}

// retainObjectTags keeps every tag the typed model does not emit, in wire
// order. Code-102 {ACAD_REACTORS groups are parsed into the model's reactor
// links (retained as a structured list and re-emitted by Format), so they are
// consumed here; standalone code-330 tags inside a reactor group belong to
// that group and are dropped with it.
func retainObjectTags(tags lldxf.Tags, emitCount map[int]int, reactions []string, groupTags int) lldxf.Tags {
	raw := make(lldxf.Tags, 0)
	inReactor := false
	keepReactors := len(reactions) == 0 // no structured links: keep the group verbatim
	consumedTags := 0                   // owner links re-emitted from the model's own field
	emitSeen := map[int]int{}           // occurrences of emitted codes already claimed by Format
	emitHard := emitHardKeys(tags)      // 360 links the model re-emits from its entries
	for i, tag := range tags {
		code := int(tag.Code())
		if code == 102 {
			v := strings.ToUpper(strings.TrimSpace(tagString(tag)))
			if v == "{ACAD_REACTORS" {
				inReactor = true
				if keepReactors {
					raw = append(raw, tag)
				}
				continue
			}
			if v == "}" && inReactor {
				inReactor = false
				if keepReactors {
					raw = append(raw, tag)
				}
				continue
			}
			if !inReactor && keepReactors {
				raw = append(raw, tag)
			}
			continue
		}
		if code == 0 || code == 5 {
			continue
		}
		if inReactor {
			if keepReactors {
				raw = append(raw, tag)
			}
			continue
		}
		if consumedTags < groupTags && code == 330 {
			consumedTags++ // an owner link re-emitted from the model's own field
			continue
		}
		if code == 100 && isModelledSubclass(tagString(tag)) {
			continue // re-emitted by the model's Format
		}
		if code == 360 && emitHard[keyAtLink(tags, i)] {
			continue // hard-owner link re-emitted from the model's own entries
		}
		if n, ok := emitCount[code]; ok && emitSeen[code] < n {
			emitSeen[code]++ // re-emitted from the model's own fields
			continue
		}
		raw = append(raw, tag)
	}
	return raw
}

// keyAtLink returns the entry key that precedes the link tag at index i,
// walking back over a code-3/105 pair (key, optional 105) to the 3 tag.
func keyAtLink(tags lldxf.Tags, i int) string {
	for j := i - 1; j >= 0; j-- {
		switch int(tags[j].Code()) {
		case 3:
			return tagString(tags[j])
		case 105, 100:
			continue
		default:
			return ""
		}
	}
	return ""
}

// emitHard reports the entry keys loaded from code-360 (hard-owner links).
func emitHardKeys(tags lldxf.Tags) map[string]bool {
	out := map[string]bool{}
	key := ""
	for _, tag := range tags {
		switch int(tag.Code()) {
		case 3:
			key = tagString(tag)
		case 350:
			key = ""
		case 360:
			if key != "" {
				out[key] = true
				key = ""
			}
		}
	}
	return out
}

// isModelledSubclass reports code-100 subclass markers the typed dictionary
// models re-emit in Format.
func isModelledSubclass(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "acdbdictionary", "acdbdictionarywithdefault", "acdbdictionaryvar":
		return true
	}
	return false
}

func tagIntValue(tag lldxf.Tag) int {
	n, _ := toInt(tagString(tag))
	return n
}

// ParseObject parses one OBJECTS-section object, returning a typed model for
// the core object types and a raw object for everything else.
// parseDictionary builds the shared DICTIONARY model used both by plain
// dictionaries and by typed collection records.
func parseDictionary(tags lldxf.Tags) *Dictionary {
	d := NewDictionary()
	d.SetHandleValue(tags.GetHandle())
	d.SetOwnerString(ownerOfRecord(tags))
	d.SetExtDictHandle(extDictOfRecord(tags))
	reactors, groupTags := reactorsOfRecord(tags)
	d.SetReactorHandles(reactors)
	// The first code-280 is the hard-ownership flag; a second occurrence
	// (AcDbDictionaryWithDefault schema flag) is retained verbatim.
	// When the record has no 280 at all, Format must not invent one:
	// keep the default and mark the flag emitted via a sentinel below.
	count280 := countIntByCode(tags, 280)
	if c, ok := firstIntByCode(tags, 280); ok {
		d.HardOwned = c != 0
	}
	if count280 == 0 {
		d.SetNoFlag(true) // Format must not invent a code-280
	}
	// 360 hard-owner links keep their wire code: record which keys used it.
	hardKey := map[string]bool{}
	lastKey := ""
	for _, tag := range tags {
		switch int(tag.Code()) {
		case 3:
			lastKey = tagString(tag)
		case 350:
			lastKey = ""
		case 360:
			if lastKey != "" {
				hardKey[lastKey] = true
				lastKey = ""
			}
		}
	}
	d.SetHardOwnerKeys(hardKey)
	var key string
	for _, tag := range tags {
		switch int(tag.Code()) {
		case 3:
			key = tagString(tag)
		case 350, 360:
			if key != "" {
				d.AddEntry(key, tagString(tag))
				key = ""
			}
		}
	}
	d.SetRetainedTags(retainObjectTags(tags, dictEmitCounts(tags), reactors, groupTags))
	return d
}

func ParseObject(tags lldxf.Tags) Object {
	typeName := strings.ToUpper(tags.DXFType())
	switch typeName {
	case "DICTIONARY":
		return parseDictionary(tags)
	case "MLEADERSTYLECOLLECTION", "MLINESTYLECOLLECTION":
		d := parseDictionary(tags)
		d.SetWireType(typeName)
		return WrapCollection(d, typeName)
	case "ACDBDICTIONARYWDFLT":
		d := NewDictionaryWithDefault()
		d.SetHandleValue(tags.GetHandle())
		d.Owner = ownerOfRecord(tags)
		d.SetExtDictHandle(extDictOfRecord(tags))
		r2, gt2 := reactorsOfRecord(tags)
		d.SetReactorHandles(r2)
		d.SetWireType(typeName) // wire marker is ACDBDICTIONARYWDFLT
		// The first code-280 is the hard-ownership flag; a second 280 after
		// the AcDbDictionaryWithDefault marker is the schema flag.
		if c, ok := firstIntByCode(tags, 280); ok {
			d.HardOwned = c != 0
		}
		inWDFLT := false
		for _, tag := range tags {
			code := int(tag.Code())
			if code == 100 && strings.EqualFold(tagString(tag), "AcDbDictionaryWithDefault") {
				inWDFLT = true
				continue
			}
			if inWDFLT && code == 280 {
				d.Schema = tagIntValue(tag)
			}
		}
		var key string
		hardKeyW := map[string]bool{}
		for _, tag := range tags {
			switch int(tag.Code()) {
			case 1:
				d.DefaultName = tagString(tag)
			case 3:
				key = tagString(tag)
			case 350:
				if key != "" {
					d.AddEntry(key, tagString(tag))
					key = ""
				}
			case 360:
				if key != "" {
					d.AddEntry(key, tagString(tag))
					hardKeyW[key] = true
					key = ""
				}
			}
		}
		d.SetHardOwnerKeys(hardKeyW)
		count280w := countIntByCode(tags, 280)
		if c, ok := firstIntByCode(tags, 280); ok {
			d.HardOwned = c != 0
		}
		if count280w == 0 {
			d.SetNoFlag(true)
		}
		d.SetRetainedTags(retainObjectTags(tags, wdfltEmitCounts(tags), r2, gt2))
		return d
	case "DICTIONARYVAR":
		v := NewDictionaryVar()
		v.SetHandleValue(tags.GetHandle())
		v.Owner = ownerOfRecord(tags)
		r3, gt3 := reactorsOfRecord(tags)
		v.SetReactorHandles(r3)
		v.SetWireType(typeName) // wire marker is DICTIONARYVAR
		inVar := false
		for _, tag := range tags {
			code := int(tag.Code())
			if code == 100 && strings.EqualFold(tagString(tag), "AcDbDictionaryVar") {
				inVar = true
				continue
			}
			if inVar && code == 280 && v.Schema == 0 {
				if n, ok := toInt(tagString(tag)); ok {
					v.Schema = n
				}
			}
		}
		v.Value = tagStringByCode(tags, 1)
		count280v := countIntByCode(tags, 280)
		if count280v == 0 {
			v.SetNoFlag(true)
		}
		v.SetRetainedTags(retainObjectTags(tags, dictVarEmitCounts(tags), r3, gt3))
		return v
	case "GROUP":
		g := NewGroup("", "", nil...)
		g.SetHandleString(tags.GetHandle())
		if owner := ownerOfRecord(tags); owner != "" {
			g.owner = fixedHandle(owner)
		}
		r4, gt4 := reactorsOfRecord(tags)
		g.SetReactorHandles(r4, gt4)
		g.Description = tagStringByCode(tags, 300)
		if v, ok := tagIntByCode(tags, 71); ok {
			g.selectable = v != 0
		}
		var handles []string
		for _, tag := range tags {
			if tag.Code() == 340 {
				handles = append(handles, tagString(tag))
			}
		}
		g.SetEntityHandles(handles)
		g.SetRetainedTags(retainObjectTags(tags, groupEmitCounts(tags), r4, gt4))
		return g
	case "XRECORD":
		x := NewXRecord()
		x.SetHandleString(tags.GetHandle())
		x.owner = ownerOfRecord(tags)
		inData := false
		seen280 := false
		for _, tag := range tags {
			code := int(tag.Code())
			if code == 0 || code == 5 || code == 330 {
				continue
			}
			if code == 100 && strings.EqualFold(tagString(tag), "AcDbXRecord") {
				inData = true
				continue
			}
			if !inData {
				continue
			}
			switch code {
			case 280:
				if !seen280 {
					seen280 = true
					if n, ok := toInt(tagString(tag)); ok {
						x.cloning = n
					}
					continue
				}
				if n, ok := toInt(tagString(tag)); ok {
					x.schema = n
				}
			default:
				x.AddTag(code, tagString(tag))
			}
		}
		return x
	case "ACDBPLACEHOLDER":
		p := NewAcDbPlaceHolder()
		p.SetHandleString(tags.GetHandle())
		if owner := ownerOfRecord(tags); owner != "" {
			p.SetOwnerHandle(fixedHandle(owner))
		}
		r5, gt5 := reactorsOfRecord(tags)
		p.SetReactorHandles(r5, gt5)
		p.RetainUnhandled(tags, placeholderEmitted)
		return p
	default:
		if isStdObjectType(typeName) {
			return wrapStd(typeName, tags)
		}
		return NewRawObject(typeName, tags)
	}
}

// TypeName returns the DXF type name for any object.
func TypeName(o Object) string {
	switch v := o.(type) {
	case *RawObject:
		return v.TypeName
	case *Dictionary:
		return wireOr(v.BaseObject, "DICTIONARY")
	case *DictionaryWithDefault:
		return wireOr(v.BaseObject, "DICTIONARYWDFLT")
	case *DictionaryVar:
		return wireOr(v.BaseObject, "DICTIONARYVAR")
	case *Group:
		return "GROUP"
	case *XRecord:
		return "XRECORD"
	case *AcDbPlaceHolder:
		return "ACDBPLACEHOLDER"
	case *AcDbDictionaryWDFLT:
		return "ACDBDICTIONARYWDFLT"
	case *ObjectCollection:
		return v.CollectionType()
	case *MLeaderStyleCollection:
		return "MLEADERSTYLECOLLECTION"
	case *MLineStyleCollection:
		return "MLINESTYLECOLLECTION"
	default:
		if t, ok := o.(dxfTypeReader); ok {
			return t.DXFType()
		}
		return fmt.Sprintf("%T", o)
	}
}

// wireOr returns the pinned wire type when set, falling back to a default.
func wireOr(b *BaseObject, def string) string {
	if b != nil && b.WireType() != "" {
		return b.WireType()
	}
	return def
}

func tagString(tag lldxf.Tag) string {
	return formatTagValue(tag.Code(), tag.Value())
}

// formatTagValue renders a tag value for the string model. Binary data codes
// (310-319, 1004) carry hex payloads on the wire; if a byte slice reached the
// model (e.g. from a binary DXF stream), it is re-encoded as uppercase hex so
// repeated round trips neither grow nor corrupt the payload.
func formatTagValue(code lldxf.GroupCode, v any) string {
	return lldxf.FormatValue(lldxf.NewDXFTag(code, v))
}

func tagStringByCode(tags lldxf.Tags, code int) string {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			return tagString(tag)
		}
	}
	return ""
}

// stringsByCode returns every string tag with the given code in wire order.
func stringsByCode(tags lldxf.Tags, code int) []string {
	out := make([]string, 0)
	for _, tag := range tags {
		if int(tag.Code()) == code {
			out = append(out, tagString(tag))
		}
	}
	return out
}

// countIntByCode returns the number of tags with the given code.
func countIntByCode(tags lldxf.Tags, code int) int {
	n := 0
	for _, tag := range tags {
		if int(tag.Code()) == code {
			n++
		}
	}
	return n
}

// firstIntByCode returns the value of the first tag with the given code.
func firstIntByCode(tags lldxf.Tags, code int) (int, bool) {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			return toInt(tagString(tag))
		}
	}
	return 0, false
}

func toInt(s string) (int, bool) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err == nil {
		return n, true
	}
	return 0, false
}

func tagIntByCode(tags lldxf.Tags, code int) (int, bool) {
	for _, tag := range tags {
		if int(tag.Code()) != code {
			continue
		}
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
	return 0, false
}
