package object

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

// fixedHandle is a handle.Handler with a value loaded from a DXF file.
type fixedHandle string

func (h fixedHandle) Handle() string                   { return string(h) }
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

// ParseObject parses one OBJECTS-section object, returning a typed model for
// the core object types and a raw object for everything else.
func ParseObject(tags lldxf.Tags) Object {
	typeName := strings.ToUpper(tags.DXFType())
	switch typeName {
	case "DICTIONARY":
		d := NewDictionary()
		d.SetHandleValue(tags.GetHandle())
		d.SetOwnerString(tagStringByCode(tags, 330))
		if v, ok := tagIntByCode(tags, 280); ok {
			d.HardOwned = v != 0
		}
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
		return d
	case "ACDBDICTIONARYWDFLT":
		return NewRawObject(typeName, tags)
	case "GROUP":
		g := NewGroup("", "", nil...)
		g.SetHandleString(tags.GetHandle())
		if owner := tagStringByCode(tags, 330); owner != "" {
			g.owner = fixedHandle(owner)
		}
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
		return g
	case "XRECORD":
		x := NewXRecord()
		x.SetHandleString(tags.GetHandle())
		x.owner = tagStringByCode(tags, 330)
		inData := false
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
				if v, ok := tagIntByCode(tags, 280); ok {
					x.cloning = v
				}
			default:
				x.AddTag(code, tagString(tag))
			}
		}
		return x
	case "ACDBPLACEHOLDER":
		p := NewAcDbPlaceHolder()
		p.SetHandleString(tags.GetHandle())
		if owner := tagStringByCode(tags, 330); owner != "" {
			p.SetOwnerHandle(fixedHandle(owner))
		}
		return p
	default:
		return NewRawObject(typeName, tags)
	}
}

// TypeName returns the DXF type name for any object.
func TypeName(o Object) string {
	switch v := o.(type) {
	case *RawObject:
		return v.TypeName
	case *Dictionary:
		return "DICTIONARY"
	case *DictionaryWithDefault:
		return "DICTIONARYWDFLT"
	case *DictionaryVar:
		return "DICTIONARYVAR"
	case *Group:
		return "GROUP"
	case *XRecord:
		return "XRECORD"
	case *AcDbPlaceHolder:
		return "ACDBPLACEHOLDER"
	case *AcDbDictionaryWDFLT:
		return "ACDBDICTIONARYWDFLT"
	default:
		return fmt.Sprintf("%T", o)
	}
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
