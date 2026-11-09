package entity

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/format"

	"github.com/edanko/dxf/lldxf"
)

// RetainedRecordType is the shared typed model for DXF records that have no
// dedicated Go type (ACAD_PROXY_ENTITY, third-party app entities, and any
// unregistered standard type). It embeds the common base so handle, layer,
// color, xdata, and raw-tag retention work uniformly, and keeps the complete
// record's tags for verbatim re-emission.
type RetainedRecordType struct {
	*entity
	typeName string
}

// NewRetainedRecord creates a retained record for dxfType with its full tag
// stream (code 0 included; it is re-emitted verbatim by the base).
func NewRetainedRecord(dxfType string, tags lldxf.Tags) *RetainedRecordType {
	e := &entity{
		Type:    ACADPROXYENTITY,
		ltscale: 1.0,
		rawTags: make([]TagPair, 0, len(tags)),
	}
	for _, tag := range tags {
		if int(tag.Code()) == 0 {
			continue // code 0 is emitted from the type name
		}
		e.rawTags = append(e.rawTags, TagPair{Code: int(tag.Code()), Value: tagValueText(tag)})
	}
	e.SetRawWholeRecord(true)
	e.SetRecordTypeName(strings.ToUpper(dxfType))
	r := &RetainedRecordType{entity: e, typeName: strings.ToUpper(dxfType)}
	if h := tags.GetHandle(); h != "" {
		e.SetHandleString(h)
	}
	return r
}

// DXFType returns the original type name of the retained record.
func (r *RetainedRecordType) DXFType() string { return r.typeName }

// Format writes the retained record through the shared base: code 0 with its
// type name, handle, then every retained tag verbatim. Retained records keep
// no typed common attributes of their own (all codes are in rawTags).
func (r *RetainedRecordType) Format(f format.Formatter) {
	f.WriteString(0, r.typeName)
	if h := r.Handle(); h != "" {
		f.WriteString(5, h)
	}
	for _, tag := range r.rawTags {
		if int(tag.Code) == 5 {
			continue // emitted from the base handle field
		}
		f.WriteString(tag.Code, tag.Value)
	}
}

// IsEntity implements Entity.
func (r *RetainedRecordType) IsEntity() bool { return true }

// BBox returns a degenerate box; retained records have no evaluated geometry.
func (r *RetainedRecordType) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// tagValueText renders a tag value as its canonical text form.
func tagValueText(tag lldxf.Tag) string {
	switch v := tag.Value().(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetAttributes returns the retained group codes keyed by code string.
func (r *RetainedRecordType) GetAttributes() map[string]interface{} {
	attrs := make(map[string]interface{}, len(r.rawTags))
	for _, tag := range r.rawTags {
		attrs[fmt.Sprintf("%d", tag.Code)] = tag.Value
	}
	return attrs
}

// LoadAttributes replaces the retained group codes.
func (r *RetainedRecordType) LoadAttributes(attribs map[string]interface{}) error {
	r.rawTags = make([]TagPair, 0, len(attribs))
	for k, v := range attribs {
		var code int
		if _, err := fmt.Sscanf(k, "%d", &code); err != nil {
			continue
		}
		r.rawTags = append(r.rawTags, TagPair{Code: code, Value: fmt.Sprintf("%v", v)})
	}
	return nil
}

// NewRetainedLoader returns the registry loader for unregistered types: a
// retained record that keeps every tag of the source record.
func NewRetainedLoader(tags lldxf.Tags) (Entity, error) {
	dxfType := tags.DXFType()
	if dxfType == "" {
		return nil, fmt.Errorf("record has no DXF type tag")
	}
	return NewRetainedRecord(dxfType, tags), nil
}
