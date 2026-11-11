package dxf

import (
	"fmt"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
	"github.com/edanko/dxf/object"
	"github.com/edanko/dxf/table"
)

// ObjectEntity adapts a typed OBJECTS-section model so a record that appears
// in the ENTITIES section can be handled by the entity loader while keeping
// its section placement on save. The embedded object owns the wire format and
// tag retention; the adapter supplies the entity-only methods.
type ObjectEntity struct {
	object.Object
}

// IsEntity marks the adapter as an entity.
func (e *ObjectEntity) IsEntity() bool { return true }

// TypedObject returns the underlying typed object model.
func (e *ObjectEntity) TypedObject() object.Object { return e.Object }

// DXFType returns the record's code-0 name.
func (e *ObjectEntity) DXFType() string { return object.TypeName(e.Object) }

// SetBlockRecord is a no-op: object records have no block record owner.
func (e *ObjectEntity) SetBlockRecord(handle.Handler) {}

// Layer returns nil; object records do not belong to a layer.
func (e *ObjectEntity) Layer() *table.Layer { return nil }

// SetLayer is a no-op for object records.
func (e *ObjectEntity) SetLayer(*table.Layer) {}

// SetLtscale is a no-op for object records.
func (e *ObjectEntity) SetLtscale(float64) {}

// BBox returns a degenerate box; object records carry no geometry.
func (e *ObjectEntity) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// SetColor is a no-op for object records.
func (e *ObjectEntity) SetColor(color.ColorNumber) {}

// GetAttributes exposes an empty attribute map for factory compatibility.
func (e *ObjectEntity) GetAttributes() map[string]any { return map[string]any{} }

// LoadAttributes accepts an attribute map without mutating the record.
func (e *ObjectEntity) LoadAttributes(map[string]any) error { return nil }

// loadObjectEntity parses an ENTITIES-section record with the typed object
// models. Records without a typed model stay on the retained-record path.
func loadObjectEntity(tags lldxf.Tags) (entity.Entity, error) {
	o := object.ParseObject(tags)
	if o == nil {
		return nil, fmt.Errorf("cannot parse object-as-entity %q", tags.DXFType())
	}
	return &ObjectEntity{Object: o}, nil
}

var _ entity.Entity = (*ObjectEntity)(nil)
