package table

import (
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

// EntryTag is one retained tag pair (group code, string value).
type EntryTag struct {
	Code  int
	Value string
}

// EntryBase is the shared base for symbol-table records: handle, owner,
// flags (code 70), xdata and extension-dictionary links (retained as tags),
// and raw-tag retention for codes the typed model does not emit.
type EntryBase struct {
	handle    string
	owner     handle.Handler
	flags     int // 70
	xdictH    string
	reactions []string
	rawTags   []EntryTag
}

// Handle returns the entry handle (code 5).
func (b *EntryBase) Handle() string { return b.handle }

// SetHandleValue sets an explicit handle loaded from a file.
func (b *EntryBase) SetHandleValue(h string) { b.handle = h }

// SetHandleString sets an explicit handle, used by deterministic repair.
func (b *EntryBase) SetHandleString(h string) { b.handle = h }

// SetHandle assigns a generated handle when the entry has none.
func (b *EntryBase) SetHandle(hg *handle.HandleGenerator) {
	if b.handle == "" {
		b.handle = hg.Next()
	}
}

// Owner returns the owning table handler (code 330).
func (b *EntryBase) Owner() handle.Handler { return b.owner }

// SetOwner sets the owning table.
func (b *EntryBase) SetOwner(h handle.Handler) { b.owner = h }

// SetOwnerHandle sets the owner from a raw handle string.
func (b *EntryBase) SetOwnerHandle(h string) {
	if h != "" {
		b.owner = fixedHandle(h)
	}
}

// Flags returns the record flags (code 70).
func (b *EntryBase) Flags() int { return b.flags }

// SetFlags sets the record flags.
func (b *EntryBase) SetFlags(v int) { b.flags = v }

// ExtDictHandle returns the extension dictionary handle (code 350).
func (b *EntryBase) ExtDictHandle() string { return b.xdictH }

// SetExtDictHandle sets the extension dictionary link.
func (b *EntryBase) SetExtDictHandle(h string) { b.xdictH = h }

// ReactorHandles returns the reactor owner handles (code 102 block).
func (b *EntryBase) ReactorHandles() []string { return b.reactions }

// SetReactorHandles sets the reactor links.
func (b *EntryBase) SetReactorHandles(hs []string) { b.reactions = hs }

// RawTags returns tags retained for round-trip fidelity.
func (b *EntryBase) RawTags() []EntryTag { return b.rawTags }

// SetRawTags replaces the retained tag stream.
func (b *EntryBase) SetRawTags(tags []EntryTag) { b.rawTags = tags }

// FormatCommon writes the common prologue of every symbol-table record:
// type, handle, reactors block, owner, and both subclass markers. The
// trailing 350 extension-dictionary link is emitted by FormatCommonTail.
func (b *EntryBase) FormatCommon(f format.Formatter, typeName, subclassName string) {
	f.WriteString(0, typeName)
	if b.handle != "" {
		f.WriteString(5, b.handle)
	}
	if len(b.reactions) > 0 {
		f.WriteString(102, "{ACAD_REACTORS")
		for _, r := range b.reactions {
			f.WriteString(330, r)
		}
		f.WriteString(102, "}")
	}
	if b.owner != nil && b.owner.Handle() != "" {
		f.WriteString(330, b.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	if subclassName != "" {
		f.WriteString(100, subclassName)
	}
}

// FormatCommonTail writes the extension dictionary link and retained raw tags.
func (b *EntryBase) FormatCommonTail(f format.Formatter) {
	if b.xdictH != "" {
		f.WriteString(350, b.xdictH)
	}
	for _, t := range b.rawTags {
		f.WriteString(t.Code, t.Value)
	}
}

// RetainUnhandled records codes the typed entry does not emit. emitted lists
// the codes Format writes itself; code 0/5 are handled by the base. Code-102
// {ACAD_REACTORS groups (and the owner 330 inside them) are kept whole when
// the model has no reactor links; standalone 330 tags and repeated occurrences
// of emitted codes are dropped.
func (b *EntryBase) RetainUnhandled(tags lldxf.Tags, emitted map[int]bool) {
	raw := make([]EntryTag, 0)
	inReactor := false
	keepReactors := len(b.reactions) == 0
	for _, tag := range tags {
		code := int(tag.Code())
		v := lldxf.FormatValue(tag)
		if code == 102 {
			u := strings.ToUpper(strings.TrimSpace(v))
			if u == "{ACAD_REACTORS" {
				inReactor = true
				if keepReactors {
					raw = append(raw, EntryTag{Code: code, Value: v})
				}
				continue
			}
			if u == "}" && inReactor {
				inReactor = false
				if keepReactors {
					raw = append(raw, EntryTag{Code: code, Value: v})
				}
				continue
			}
			if !inReactor && keepReactors {
				raw = append(raw, EntryTag{Code: code, Value: v})
			}
			continue
		}
		if code == 0 || code == 5 {
			continue
		}
		if inReactor {
			if keepReactors {
				raw = append(raw, EntryTag{Code: code, Value: v})
			}
			continue
		}
		if code == 330 {
			continue // owner link handled by the base
		}
		if code == 350 && b.xdictH != "" {
			continue
		}
		if emitted[code] {
			continue // every occurrence is re-emitted from the model's fields
		}
		raw = append(raw, EntryTag{Code: code, Value: v})
	}
	b.rawTags = raw
}
