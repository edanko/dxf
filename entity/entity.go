package entity

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
	"github.com/edanko/dxf/xdata"
)

// Entity is interface for DXF Entities.
type Entity interface {
	IsEntity() bool
	Format(format.Formatter)
	SetBlockRecord(handle.Handler)
	Layer() *table.Layer
	SetLayer(*table.Layer)
	SetLtscale(float64)
	BBox() ([]float64, []float64)
	SetColor(color.ColorNumber)

	handle.Handler

	// Factory support
	DXFType() string
	GetAttributes() map[string]interface{}
	LoadAttributes(attribs map[string]interface{}) error
}

// entity is common part of Entities.
// It is embedded in each entities to implement Entity interface.
type entity struct {
	Type        EntityType        // 0
	handle      string            // 5
	blockRecord handle.Handler    // 102 330
	owner       handle.Handler    // 330
	layer       *table.Layer      // 8
	layerName   string            // 8 fallback when no layer pointer is resolved
	ltscale     float64           // 48
	color       color.ColorNumber // 62
	xdata       *xdata.XData      // XDATA (1000-1071)
	rawTags     []TagPair         // unhandled tags retained for round trips
}

// NewEntity creates a new entity.
func NewEntity(t EntityType) *entity {
	e := &entity{
		Type:        t,
		blockRecord: nil,
		owner:       nil,
		ltscale:     1.0,
		color:       0,
	}
	return e
}

// SetColor sets a color to an entity.
func (e *entity) SetColor(cl color.ColorNumber) {
	e.color = cl
}

// Format writes data to formatter.
func (e *entity) Format(f format.Formatter) {
	f.WriteString(0, EntityTypeString(e.Type))
	if e.handle != "" {
		f.WriteString(5, e.handle)
	}
	if e.blockRecord != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, e.blockRecord.Handle())
		f.WriteString(102, "}")
	}
	if e.owner != nil {
		f.WriteString(330, e.owner.Handle())
	}
	f.WriteString(100, "AcDbEntity")
	layerName := e.layerName
	if e.layer != nil {
		layerName = e.layer.Name()
	}
	if layerName != "" {
		f.WriteString(8, layerName)
	}
	if e.ltscale != 1.0 {
		f.WriteFloat(48, e.ltscale)
	}
	f.WriteInt(62, int(e.color))
	for _, tag := range e.rawTags {
		f.WriteString(tag.Code, tag.Value)
	}
}

// Handle returns a handle value of TABLE.
func (e *entity) Handle() string {
	return e.handle
}

// SetHandle sets handles to TABLE itself and each SymbolTable.
func (e *entity) SetHandle(hg *handle.HandleGenerator) {
	if e.handle == "" {
		e.handle = hg.Next()
	}
}

// SetHandleString sets an explicit handle, preserving handles loaded from a file.
func (e *entity) SetHandleString(h string) {
	e.handle = h
}

// SetLayerName records the layer name for cases where the layer table has not
// been resolved yet (for example while loading a document).
func (e *entity) SetLayerName(name string) {
	e.layerName = name
}

// LayerName returns the resolved or recorded layer name.
func (e *entity) LayerName() string {
	if e.layer != nil {
		return e.layer.Name()
	}
	return e.layerName
}

// SetRawTags stores unhandled group codes so they survive a load/save round trip.
func (e *entity) SetRawTags(tags []TagPair) {
	e.rawTags = tags
}

// RawTags returns the unhandled group codes retained on this entity.
func (e *entity) RawTags() []TagPair {
	return e.rawTags
}

// SetBlockRecord sets BLOCK_RECORD to entity (code 330).
func (e *entity) SetBlockRecord(h handle.Handler) {
	e.blockRecord = h
}

// SetOwner sets an owner.
func (e *entity) SetOwner(h handle.Handler) {
	e.owner = h
}

// Layer returns entity's Layer.
func (e *entity) Layer() *table.Layer {
	return e.layer
}

// SetLayer sets Layer to entity.
func (e *entity) SetLayer(l *table.Layer) {
	e.layer = l
}

// SetLtscale sets Layer to entity.
func (e *entity) SetLtscale(v float64) {
	e.ltscale = v
}

// SetEntityType sets entity type.
func (e *entity) SetEntityType(t EntityType) {
	e.Type = t
}

// DXFType returns the DXF type string for this entity
func (e *entity) DXFType() string {
	return EntityTypeString(e.Type)
}

// GetAttributes returns entity attributes as a map
func (e *entity) GetAttributes() map[string]interface{} {
	attributes := make(map[string]interface{})
	if e.layer != nil {
		attributes["layer"] = e.layer.Name()
	}
	attributes["color"] = int(e.color)
	if e.ltscale != 1.0 {
		attributes["ltscale"] = e.ltscale
	}
	return attributes
}

// LoadAttributes loads attributes from a map
func (e *entity) LoadAttributes(attribs map[string]interface{}) error {
	if layerName, ok := attribs["layer"].(string); ok {
		// Note: This requires access to the document's layer table
		// For now, just store the name
		_ = layerName
	}
	if colorVal, ok := attribs["color"].(int); ok {
		e.color = color.ColorNumber(colorVal)
	}
	if ltscaleVal, ok := attribs["ltscale"].(float64); ok {
		e.ltscale = ltscaleVal
	}
	return nil
}
