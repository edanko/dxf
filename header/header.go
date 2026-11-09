// HEADER section
package header

import (
	"fmt"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/insunit"
	"github.com/edanko/dxf/lldxf"
)

// Variable is a header variable retained without a typed model.
type Variable struct {
	Name string
	Tags lldxf.Tags
}

// ConversionReport records how a document was normalized or converted.
type ConversionReport struct {
	SourceVersion     string
	NormalizedVersion string
	Dropped           []string
	Converted         []string
}

// Header contains information written in HEADER section.
type Header struct {
	Version           string
	SourceVersion     string
	NormalizedVersion string
	TargetVersion     string
	Codepage          string
	InsBase           []float64
	InsUnit           insunit.Unit
	InsLUnit          insunit.Type
	ExtMin            []float64
	ExtMax            []float64
	LtScale           float64
	handseed          string
	Extra             []Variable
	Conversion        *ConversionReport
}

// New creates a new Header.
func New() *Header {
	h := new(Header)
	h.Version = "AC1015"
	h.SourceVersion = "AC1015"
	h.NormalizedVersion = "AC1015"
	h.InsBase = make([]float64, 3)
	h.ExtMin = make([]float64, 3)
	h.ExtMax = make([]float64, 3)
	h.LtScale = 1.0
	return h
}

// Format writes HEADER information to formatter.
func (h *Header) Format(f format.Formatter) {
	h.FormatAt(f, "")
}

// FormatAt writes the HEADER section declaring the requested target version
// when it is supported; otherwise the header's own $ACADVER.
// writeVariableTag emits a raw header variable tag with its native type: binary
// formatters must receive ints and floats as numbers (int16/int32 sizing per
// group code), while ASCII output formats every value textually.
func writeVariableTag(f format.Formatter, tag lldxf.Tag) {
	code := int(tag.Code())
	if bf, ok := f.(interface {
		WriteValue(int, interface{})
	}); ok {
		bf.WriteValue(code, tag.Value())
		return
	}
	f.WriteString(code, tagString(tag))
}

func (h *Header) FormatAt(f format.Formatter, target string) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "HEADER")
	version := h.Version
	if h.TargetVersion != "" {
		version = h.TargetVersion
	}
	if target != "" && format.IsSupportedVersion(target) {
		version = target
	}
	f.WriteString(9, "$ACADVER")
	f.WriteString(1, version)
	if cp := h.Codepage; cp != "" && !format.Profile(version).UTF8 {
		f.WriteString(9, "$DWGCODEPAGE")
		f.WriteString(3, cp)
	}
	f.WriteString(9, "$INSBASE")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.InsBase[i])
	}
	f.WriteString(9, "$INSUNITS")
	h.InsUnit.Format(f)
	f.WriteString(9, "$LUNITS")
	h.InsLUnit.Format(f)

	f.WriteString(9, "$EXTMIN")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.ExtMin[i])
	}
	f.WriteString(9, "$EXTMAX")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.ExtMax[i])
	}
	f.WriteString(9, "$LTSCALE")
	f.WriteFloat(40, h.LtScale)
	f.WriteString(9, "$HANDSEED")
	if h.handseed != "" {
		f.WriteString(5, h.handseed)
	}
	dropCodes := format.Profile(version).DropHeaderCodes
	for _, v := range h.Extra {
		drop, notes := filterVariableTags(v.Tags, dropCodes)
		for _, n := range notes {
			h.noteConversionDrop(n)
		}
		if drop {
			continue
		}
		f.WriteString(9, v.Name)
		for _, tag := range filteredTags(v.Tags, dropCodes) {
			writeVariableTag(f, tag)
		}
	}
	f.WriteString(0, "ENDSEC")
}

// SetHandle sets $HANDSEED.

// SetHandle sets $HANDSEED.
func (h *Header) SetHandle(hg *handle.HandleGenerator) {
	h.handseed = hg.String()
}

// dropsForR12 reports whether a header variable uses any group code that the
// restricted release cannot store (e.g. true-color 420 in R12). The check is
// by group code: values are free text and carry no version semantics.
// filterVariableTags returns the tags of a header variable that the target
// release can store. When every value tag is dropped, the caller must skip the
// whole variable (drop=true). Dropped features are returned as notes for the
// conversion report.
func filterVariableTags(tags lldxf.Tags, dropCodes map[int]bool) (bool, []string) {
	if len(dropCodes) == 0 {
		return false, nil
	}
	var kept lldxf.Tags
	var notes []string
	for _, tag := range tags {
		if dropCodes[int(tag.Code())] {
			notes = append(notes, fmt.Sprintf("header variable code %d", int(tag.Code())))
			continue
		}
		kept = append(kept, tag)
	}
	return len(kept) == 0 && len(tags) > 0, notes
}

// filteredTags returns the tags of a header variable that the target release
// can store.
func filteredTags(tags lldxf.Tags, dropCodes map[int]bool) lldxf.Tags {
	if len(dropCodes) == 0 {
		return tags
	}
	var kept lldxf.Tags
	for _, tag := range tags {
		if dropCodes[int(tag.Code())] {
			continue
		}
		kept = append(kept, tag)
	}
	return kept
}

// noteConversionDrop records a dropped feature in the header conversion report.
func (h *Header) noteConversionDrop(note string) {
	if h.Conversion == nil {
		h.Conversion = &ConversionReport{SourceVersion: h.SourceVersion, NormalizedVersion: h.TargetVersion}
	}
	for _, n := range h.Conversion.Dropped {
		if n == note {
			return
		}
	}
	h.Conversion.Dropped = append(h.Conversion.Dropped, note)
}
