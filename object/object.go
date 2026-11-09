package object

import (
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

// BaseObject is the base struct for DXF objects
type BaseObject struct {
	handle   string
	wireType string     // code-0 marker emitted by Format (e.g. DICTIONARYVAR)
	reactors []string   // owner handles from the {ACAD_REACTORS group (code 102/330)
	retained lldxf.Tags // tags the typed model does not emit, kept for round trip
}

func (b *BaseObject) Handle() string {
	return b.handle
}

func (b *BaseObject) SetHandle(hg *handle.HandleGenerator) {
	if b.handle == "" {
		b.handle = hg.Next()
	}
}

// WireType returns the code-0 marker Format emits for this object. Models
// that do not pin a wire type (DICTIONARY) report an empty string and their
// Format methods emit the conventional marker.
func (b *BaseObject) WireType() string {
	return b.wireType
}

// SetWireType pins the code-0 marker emitted on save.
func (b *BaseObject) SetWireType(t string) { b.wireType = t }

// RetainedTags returns tags kept for round-trip fidelity.
func (b *BaseObject) RetainedTags() lldxf.Tags { return b.retained }

// SetRetainedTags replaces the retained tag stream.
func (b *BaseObject) SetRetainedTags(tags lldxf.Tags) { b.retained = tags }

// SetReactorHandles stores the reactor owner handles parsed from the record's
// {ACAD_REACTORS group. Format re-emits the group when it is non-empty.
func (b *BaseObject) SetReactorHandles(hs []string) { b.reactors = hs }

// ReactorHandles returns the parsed reactor links.
func (b *BaseObject) ReactorHandles() []string { return b.reactors }

// formatReactors writes the {ACAD_REACTORS group from the parsed links.
func (b *BaseObject) formatReactors(f format.Formatter) {
	if len(b.reactors) == 0 {
		return
	}
	f.WriteString(102, "{ACAD_REACTORS")
	for _, r := range b.reactors {
		f.WriteString(330, r)
	}
	f.WriteString(102, "}")
}

// RetainUnhandled records codes the typed model does not emit. emitted lists
// the codes Format writes itself; 0/5 are always handled by the base.
func (b *BaseObject) RetainUnhandled(tags lldxf.Tags, emitted map[int]bool) {
	raw := make(lldxf.Tags, 0)
	inReactor := false
	keep330 := !emitted[330] // the model emits its own owner link when set
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 102 {
			v := strings.ToUpper(strings.TrimSpace(lldxf.FormatValue(tag)))
			if v == "{ACAD_REACTORS" {
				inReactor = true
				if len(b.reactors) == 0 {
					raw = append(raw, tag)
				}
				continue
			}
			if v == "}" && inReactor {
				inReactor = false
				if len(b.reactors) == 0 {
					raw = append(raw, tag)
				}
				continue
			}
			if !inReactor && len(b.reactors) == 0 {
				raw = append(raw, tag)
			}
			continue
		}
		if code == 0 || code == 5 {
			continue
		}
		if inReactor {
			if len(b.reactors) == 0 {
				raw = append(raw, tag)
			}
			continue
		}
		if code == 330 && !keep330 {
			continue // owner link re-emitted from the model's own field
		}
		if emitted[code] {
			continue // every occurrence is re-emitted from the model's fields
		}
		raw = append(raw, tag)
	}
	b.retained = raw
}

// formatRetained writes the retained tags after the typed body.
func (b *BaseObject) formatRetained(f format.Formatter) {
	for _, tag := range b.retained {
		f.WriteString(int(tag.Code()), lldxf.FormatValue(tag))
	}
}

// Object is the base interface for DXF objects
type Object interface {
	IsObject() bool
	Format(format.Formatter)
	handle.Handler
}

// objectObject implements the Object interface
type objectObject struct {
	handle string
}

// IsObject returns true for object objects
func (o *objectObject) IsObject() bool {
	return true
}

// Handle returns object handle
func (o *objectObject) Handle() string {
	return o.handle
}

// SetHandle sets object handle
func (o *objectObject) SetHandle(hg *handle.HandleGenerator) {
	o.handle = hg.Next()
}

// Format writes basic object data to formatter
func (o *objectObject) Format(f format.Formatter) {
	f.WriteString(0, "OBJECT")
	f.WriteString(5, o.handle)
}
