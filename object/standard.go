package object

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// StdObject is the shared typed model for standard OBJECTS-section records
// whose full field set is not individually modelled. It exposes the object
// identity, owner, and the fields that matter to callers (name, file path,
// block record, image size, entity links, sort order), and retains the whole
// tag stream verbatim so every unmodelled code round-trips unchanged.
type StdObject struct {
	*BaseObject

	typeName string

	owner       string
	name        string
	description string
	flags       int

	path       string
	blockRec   string
	entityRefs []string

	imageSizeX float64
	imageSizeY float64
	pixelSizeX float64
	pixelSizeY float64

	sortOrder []string          // entity handles in source order
	sortBy    map[string]string // entity handle -> sort handle
}

// IsObject is for the Object interface.
func (s *StdObject) IsObject() bool { return true }

// DXFType returns the code-0 marker of the record.
func (s *StdObject) DXFType() string { return s.typeName }

// Owner returns the owning object/dictionary handle.
func (s *StdObject) Owner() string { return s.owner }

// Name returns the record name/description where the type defines one.
func (s *StdObject) Name() string { return s.name }

// Description returns the record description where the type defines one.
func (s *StdObject) Description() string { return s.description }

// Flags returns the record's primary bit-coded flags field.
func (s *StdObject) Flags() int { return s.flags }

// FilePath returns the referenced file path (IMAGEDEF, UNDERLAYDEFINITION,
// PDFDEFINITION, ...).
func (s *StdObject) FilePath() string { return s.path }

// BlockRecordHandle returns the associated BLOCK_RECORD handle (LAYOUT,
// SORTENTSTABLE).
func (s *StdObject) BlockRecordHandle() string { return s.blockRec }

// EntityHandles returns the entity links recorded by IDBUFFER/FIELDLIST.
func (s *StdObject) EntityHandles() []string { return s.entityRefs }

// ImageSize returns the source image size in pixels.
func (s *StdObject) ImageSize() (float64, float64) { return s.imageSizeX, s.imageSizeY }

// PixelSize returns the default size of one pixel in drawing units.
func (s *StdObject) PixelSize() (float64, float64) { return s.pixelSizeX, s.pixelSizeY }

// SortOrder returns the entity-handle to sort-handle mapping of a
// SORTENTSTABLE, in source order.
func (s *StdObject) SortOrder() map[string]string { return s.sortBy }

// ExtDictHandle returns the retained extension-dictionary link (code 350).
func (s *StdObject) ExtDictHandle() string {
	for _, tag := range s.RetainedTags() {
		if int(tag.Code()) == 350 {
			return lldxf.FormatValue(tag)
		}
	}
	return ""
}

// Format writes the record. Loaded records re-emit their retained tag stream
// verbatim; a freshly created record writes just its identity and owner.
func (s *StdObject) Format(f format.Formatter) {
	wire := s.WireType()
	if wire == "" {
		wire = s.typeName
	}
	f.WriteString(0, wire)
	if s.Handle() != "" {
		f.WriteString(5, s.Handle())
	}
	if ret := s.RetainedTags(); len(ret) > 0 {
		for _, tag := range ret {
			f.WriteString(int(tag.Code()), lldxf.FormatValue(tag))
		}
		return
	}
	if s.owner != "" {
		f.WriteString(330, s.owner)
	}
}

// stdRetained keeps every tag except the record type (0) and the record handle
// (the first code-5). Later code-5 tags (SORTENTSTABLE sort handles) are kept.
func stdRetained(tags lldxf.Tags) lldxf.Tags {
	out := make(lldxf.Tags, 0, len(tags))
	seenHandle := false
	for _, tag := range tags {
		switch int(tag.Code()) {
		case 0:
			continue
		case 5:
			if !seenHandle {
				seenHandle = true
				continue
			}
		}
		out = append(out, tag)
	}
	return out
}

// newStdObject builds the shared typed model from a parsed record.
func newStdObject(typeName string, tags lldxf.Tags) *StdObject {
	s := &StdObject{
		BaseObject: &BaseObject{},
		typeName:   typeName,
		owner:      ownerOfRecord(tags),
	}
	s.SetWireType(typeName)
	s.BaseObject.handle = tags.GetHandle()
	if reactors, _ := reactorsOfRecord(tags); len(reactors) > 0 {
		s.SetReactorHandles(reactors)
	}
	s.SetRetainedTags(stdRetained(tags))

	switch typeName {
	case "LAYOUT":
		s.name = tagAfterMarker(tags, "AcDbLayout", 1)
		s.blockRec = handleAfterMarker(tags, "AcDbLayout", 330)
		s.flags = intAfterMarker(tags, "AcDbLayout", 70)
	case "PLOTSETTINGS":
		s.name = tagStringByCode(tags, 1)
	case "MLINESTYLE":
		s.name = tagStringByCode(tags, 2)
		s.description = tagStringByCode(tags, 3)
		s.flags = intByCode(tags, 70)
	case "MLEADERSTYLE":
		s.name = tagStringByCode(tags, 3)
		s.flags = intByCode(tags, 179)
	case "MATERIAL":
		s.name = tagStringByCode(tags, 1)
		s.description = tagStringByCode(tags, 2)
	case "VISUALSTYLE":
		s.name = tagStringByCode(tags, 2)
	case "TABLESTYLE":
		s.name = tagStringByCode(tags, 3)
		s.flags = intByCode(tags, 7)
	case "SORTENTSTABLE":
		s.blockRec = handleAfterMarker(tags, "AcDbSortentsTable", 330)
		s.loadSortOrder(tags)
	case "IDBUFFER":
		s.entityRefs = handlesAfterMarker(tags, "AcDbIdBuffer", 330)
	case "FIELDLIST":
		s.entityRefs = handlesAfterMarker(tags, "AcDbIdSet", 330)
	case "IMAGEDEF":
		s.path = tagStringByCode(tags, 1)
		s.imageSizeX = floatAfterMarker(tags, "AcDbRasterImageDef", 10)
		s.imageSizeY = floatAfterMarker(tags, "AcDbRasterImageDef", 20)
		s.pixelSizeX = floatAfterMarker(tags, "AcDbRasterImageDef", 11)
		s.pixelSizeY = floatAfterMarker(tags, "AcDbRasterImageDef", 21)
	case "IMAGEDEF_REACTOR":
		s.blockRec = handleAfterMarker(tags, "AcDbRasterImageDefReactor", 330)
	case "UNDERLAYDEFINITION", "PDFDEFINITION", "DWFDEFINITION", "DGNDEFINITION":
		s.path = tagStringByCode(tags, 1)
		s.name = tagStringByCode(tags, 2)
	}
	return s
}

// loadSortOrder records SORTENTSTABLE (331 entity, 5 sort handle) pairs in
// source order.
func (s *StdObject) loadSortOrder(tags lldxf.Tags) {
	s.sortBy = make(map[string]string)
	inTable := false
	pending := ""
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 100 {
			if strings.EqualFold(tagString(tag), "AcDbSortentsTable") {
				inTable = true
			}
			continue
		}
		if !inTable {
			continue
		}
		switch code {
		case 331:
			pending = tagString(tag)
		case 5:
			if pending != "" {
				s.sortBy[pending] = tagString(tag)
				s.sortOrder = append(s.sortOrder, pending)
				pending = ""
			}
		}
	}
}

// tagAfterMarker returns the first tag with the given code that follows the
// named subclass marker.
func tagAfterMarker(tags lldxf.Tags, marker string, code int) string {
	in := false
	for _, tag := range tags {
		if trigger, ok := markerTrigger(tag, marker); ok {
			in = trigger
			continue
		}
		if in && int(tag.Code()) == code {
			return tagString(tag)
		}
	}
	return ""
}

// handleAfterMarker returns the first code-330 handle after the marker.
func handleAfterMarker(tags lldxf.Tags, marker string, code int) string {
	return tagAfterMarker(tags, marker, code)
}

// handlesAfterMarker returns every handle of the given code after the marker.
func handlesAfterMarker(tags lldxf.Tags, marker string, code int) []string {
	in := false
	out := make([]string, 0)
	for _, tag := range tags {
		if trigger, ok := markerTrigger(tag, marker); ok {
			in = trigger
			continue
		}
		if in && int(tag.Code()) == code {
			out = append(out, tagString(tag))
		}
	}
	return out
}

// intAfterMarker returns the first integer of the given code after the marker.
func intAfterMarker(tags lldxf.Tags, marker string, code int) int {
	in := false
	for _, tag := range tags {
		if trigger, ok := markerTrigger(tag, marker); ok {
			in = trigger
			continue
		}
		if in && int(tag.Code()) == code {
			if n, ok := toInt(tagString(tag)); ok {
				return n
			}
		}
	}
	return 0
}

// floatAfterMarker returns the first float of the given code after the marker.
func floatAfterMarker(tags lldxf.Tags, marker string, code int) float64 {
	in := false
	for _, tag := range tags {
		if trigger, ok := markerTrigger(tag, marker); ok {
			in = trigger
			continue
		}
		if in && int(tag.Code()) == code {
			if v, ok := tagFloat(tag); ok {
				return v
			}
		}
	}
	return 0
}

// markerTrigger reports whether tag is a code-100 subclass marker and whether
// it names the wanted subclass.
func markerTrigger(tag lldxf.Tag, marker string) (bool, bool) {
	if int(tag.Code()) != 100 {
		return false, false
	}
	return strings.EqualFold(tagString(tag), marker), true
}

// intByCode returns the first integer with the given code.
func intByCode(tags lldxf.Tags, code int) int {
	v, _ := intByCodeOK(tags, code)
	return v
}

// intByCodeOK returns the first integer with the given code.
func intByCodeOK(tags lldxf.Tags, code int) (int, bool) {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			if n, ok := toInt(tagString(tag)); ok {
				return n, true
			}
		}
	}
	return 0, false
}

// Typed wrappers. Each is a distinct Go type so callers can switch on the
// concrete standard object; all behaviour is promoted from StdObject.
type Layout struct{ *StdObject }

type PlotSettings struct{ *StdObject }

type MLineStyle struct{ *StdObject }

type MLeaderStyle struct{ *StdObject }

type Material struct{ *StdObject }

type VisualStyle struct{ *StdObject }

type TableStyle struct{ *StdObject }

type SortEntsTable struct{ *StdObject }

type IDBuffer struct{ *StdObject }

type Field struct{ *StdObject }

type FieldList struct{ *StdObject }

type GeoData struct{ *StdObject }

type Sun struct{ *StdObject }

type SpatialFilter struct{ *StdObject }

type RasterVariables struct{ *StdObject }

type WipeoutVariables struct{ *StdObject }

type ImageDef struct{ *StdObject }

type ImageDefReactor struct{ *StdObject }

type UnderlayDefinition struct{ *StdObject }

type PdfDefinition struct{ *StdObject }

type DwfDefinition struct{ *StdObject }

type DgnDefinition struct{ *StdObject }

// isStdObjectType reports whether typeName has a typed standard-object model.
func isStdObjectType(typeName string) bool {
	switch typeName {
	case "LAYOUT", "PLOTSETTINGS", "MLINESTYLE", "MLEADERSTYLE", "MATERIAL",
		"VISUALSTYLE", "TABLESTYLE", "SORTENTSTABLE", "IDBUFFER", "FIELD",
		"FIELDLIST", "GEODATA", "SUN", "SPATIAL_FILTER", "RASTERVARIABLES",
		"WIPEOUTVARIABLES", "IMAGEDEF", "IMAGEDEF_REACTOR", "UNDERLAYDEFINITION",
		"PDFDEFINITION", "DWFDEFINITION", "DGNDEFINITION":
		return true
	}
	return false
}

// wrapStd returns the typed concrete model for a standard object type.
func wrapStd(typeName string, tags lldxf.Tags) Object {
	s := newStdObject(typeName, tags)
	switch typeName {
	case "LAYOUT":
		return &Layout{s}
	case "PLOTSETTINGS":
		return &PlotSettings{s}
	case "MLINESTYLE":
		return &MLineStyle{s}
	case "MLEADERSTYLE":
		return &MLeaderStyle{s}
	case "MATERIAL":
		return &Material{s}
	case "VISUALSTYLE":
		return &VisualStyle{s}
	case "TABLESTYLE":
		return &TableStyle{s}
	case "SORTENTSTABLE":
		return &SortEntsTable{s}
	case "IDBUFFER":
		return &IDBuffer{s}
	case "FIELD":
		return &Field{s}
	case "FIELDLIST":
		return &FieldList{s}
	case "GEODATA":
		return &GeoData{s}
	case "SUN":
		return &Sun{s}
	case "SPATIAL_FILTER":
		return &SpatialFilter{s}
	case "RASTERVARIABLES":
		return &RasterVariables{s}
	case "WIPEOUTVARIABLES":
		return &WipeoutVariables{s}
	case "IMAGEDEF":
		return &ImageDef{s}
	case "IMAGEDEF_REACTOR":
		return &ImageDefReactor{s}
	case "UNDERLAYDEFINITION":
		return &UnderlayDefinition{s}
	case "PDFDEFINITION":
		return &PdfDefinition{s}
	case "DWFDEFINITION":
		return &DwfDefinition{s}
	case "DGNDEFINITION":
		return &DgnDefinition{s}
	}
	return s
}

// tagFloat coerces a tag value to float64.
func tagFloat(tag lldxf.Tag) (float64, bool) {
	switch v := tag.Value().(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		var f float64
		if _, err := fmt.Sscanf(tagString(tag), "%f", &f); err != nil {
			return 0, false
		}
		return f, true
	}
}
