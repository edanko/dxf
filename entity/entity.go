package entity

import (
	"fmt"

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
	GetAttributes() map[string]any
	LoadAttributes(attribs map[string]any) error
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

	// Common DXF attributes carried by the shared base (EntityBase).
	linetype     string    // 6
	trueColor    int       // 420 RGB true color, 0 = not set
	lineweight   int       // 370; -3 = unset (ByLayer), emitted only when set
	invisible    bool      // 60
	transparency float64   // 440 transparency 0..1, 0 = none
	extrusion    []float64 // 210, 220, 230 extrusion direction
	thickness    float64   // 39 thickness (LINE, CIRCLE, ARC, LWPOLYLINE, ...)

	reactors []string // 102 {ACAD_REACTORS 330 ... 102 }
	xdict    string   // 102 {ACAD_XDICTIONARY 360 handle 102 }

	rawWholeRecord   bool   // rawTags carry every retained code of the record
	typeNameOverride string // explicit code 0 name for retained records
}

// SetRawWholeRecord marks the raw tags as the complete DXF record
// (whole-record retention). The record's type name and handle are emitted from
// the base fields, and every other retained tag is written verbatim.
func (e *entity) SetRawWholeRecord(v bool) { e.rawWholeRecord = v }

// RawWholeRecord reports whether whole-record retention is active.
func (e *entity) RawWholeRecord() bool { return e.rawWholeRecord }

// recordTypeName returns the DXF type name for code 0, preferring an explicit
// override set by retained records with non-standard names.
func (e *entity) recordTypeName() string {
	if e.typeNameOverride != "" {
		return e.typeNameOverride
	}
	return EntityTypeString(e.Type)
}

// SetRecordTypeName overrides the code 0 name emitted for this record.
func (e *entity) SetRecordTypeName(name string) { e.typeNameOverride = name }

// NewEntity creates a new entity.
func NewEntity(t EntityType) *entity {
	e := &entity{
		lineweight:  -3,
		Type:        t,
		blockRecord: nil,
		owner:       nil,
		ltscale:     1.0,
		color:       0,
	}
	return e
}

// IsEntity implements Entity for every type embedding the base.
func (e *entity) IsEntity() bool { return true }

// SetColor sets a color to an entity.
func (e *entity) SetColor(cl color.ColorNumber) {
	e.color = cl
}

// Format writes data to formatter. When the owning entity registered raw tags
// for the whole record (fully unhandled records), the base emits the complete
// raw stream verbatim; otherwise rawTags carry only codes this type does not
// model and are appended after the typed fields.
func (e *entity) Format(f format.Formatter) {
	f.WriteString(0, e.recordTypeName())
	if e.handle != "" {
		f.WriteString(5, e.handle)
	}
	if len(e.reactors) > 0 {
		f.WriteString(102, "{ACAD_REACTORS")
		for _, r := range e.reactors {
			f.WriteString(330, r)
		}
		f.WriteString(102, "}")
	}
	if e.xdict != "" {
		f.WriteString(102, "{ACAD_XDICTIONARY")
		f.WriteString(360, e.xdict)
		f.WriteString(102, "}")
	}
	if e.blockRecord != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, e.blockRecord.Handle())
		f.WriteString(102, "}")
	}
	f.WriteString(100, "AcDbEntity")
	layerName := e.layerName
	if e.layer != nil {
		layerName = e.layer.Name()
	}
	if layerName != "" {
		f.WriteString(8, layerName)
	}
	if e.linetype != "" {
		f.WriteString(6, e.linetype)
	}
	if e.ltscale != 1.0 {
		f.WriteFloat(48, e.ltscale)
	}
	f.WriteInt(62, int(e.color))
	if e.trueColor != 0 {
		f.WriteInt(420, e.trueColor)
	}
	if e.transparency != 0 {
		f.WriteInt(440, int(e.transparency*100+0.5))
	}
	if e.invisible {
		f.WriteInt(60, 1)
	}
	if e.lineweight > -3 {
		f.WriteInt(370, e.lineweight)
	}
	if e.owner != nil {
		f.WriteString(330, e.owner.Handle())
	}
	e.formatXData(f)
	for _, tag := range e.rawTags {
		f.WriteString(tag.Code, tag.Value)
	}
}

// formatXData writes retained extended data (1001 app IDs and their tags).
func (e *entity) formatXData(f format.Formatter) {
	if e.xdata == nil {
		return
	}
	for _, appID := range e.xdata.AppIDs() {
		data, ok := e.xdata.GetAppData(appID)
		if !ok || len(data.Tags) == 0 {
			continue
		}
		f.WriteString(1001, data.AppID)
		for _, tag := range data.Tags {
			f.WriteString(tag.Code, fmt.Sprintf("%v", tag.Value))
		}
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

// Owner returns the entity's owner handler (code 330).
func (e *entity) Owner() handle.Handler { return e.owner }

// SetReactorHandles stores persistent reactor handles (102 {ACAD_REACTORS).
func (e *entity) SetReactorHandles(hs []string) { e.reactors = hs }

// ReactorHandles returns persistent reactor handles.
func (e *entity) ReactorHandles() []string { return e.reactors }

// SetExtDictHandle stores the extension dictionary handle
// (102 {ACAD_XDICTIONARY 360 handle).
func (e *entity) SetExtDictHandle(h string) { e.xdict = h }

// ExtDictHandle returns the extension dictionary handle.
func (e *entity) ExtDictHandle() string { return e.xdict }

// Linetype returns the entity linetype name (code 6).
func (e *entity) Linetype() string { return e.linetype }

// SetLinetype sets the entity linetype name (code 6).
func (e *entity) SetLinetype(name string) { e.linetype = name }

// ColorRGB returns the true color RGB value (code 420), 0 when unset.
// Color returns the color index (group code 62).
func (e *entity) Color() color.ColorNumber { return e.color }

func (e *entity) ColorRGB() int { return e.trueColor }

// SetColorRGB sets the true color RGB value (code 420).
func (e *entity) SetColorRGB(rgb int) { e.trueColor = rgb }

// LineWeight returns the entity lineweight (code 370); -3 means unset.
func (e *entity) LineWeight() int { return e.lineweight }

// SetLineWeight sets the entity lineweight (code 370).
func (e *entity) SetLineWeight(w int) { e.lineweight = w }

// Invisible reports whether the entity is marked invisible (code 60).
func (e *entity) Invisible() bool { return e.invisible }

// SetInvisible sets the visibility flag (code 60).
func (e *entity) SetInvisible(v bool) { e.invisible = v }

// Transparency returns the entity transparency 0..1 (code 440).
func (e *entity) Transparency() float64 { return e.transparency }

// SetTransparency sets the entity transparency 0..1 (code 440).
func (e *entity) SetTransparency(t float64) { e.transparency = t }

// Extrusion returns the extrusion direction (codes 210, 220, 230), nil when unset.
func (e *entity) Extrusion() []float64 { return e.extrusion }

// SetExtrusion sets the extrusion direction (codes 210, 220, 230).
func (e *entity) SetExtrusion(v []float64) { e.extrusion = v }

// Thickness returns the entity thickness (code 39). Types that model
// thickness per-record (CIRCLE, ARC, TRACE, SHAPE, ...) emit it in their own
// Format; types without a dedicated field rely on this base value.
func (e *entity) Thickness() float64 { return e.thickness }

// SetThickness sets the entity thickness (code 39).
func (e *entity) SetThickness(v float64) { e.thickness = v }

// BaseExtrusion returns the extrusion direction stored on the shared base,
// nil when unset. Types that model 210/220/230 as their own field (CIRCLE,
// INSERT, SPLINE, ...) emit theirs; types without one rely on this value via
// FormatBaseExtrusion.
func (e *entity) BaseExtrusion() []float64 { return e.extrusion }

// FormatBaseExtrusion writes the base extrusion direction when set. Callers
// that do not model 210/220/230 themselves invoke it after their subclass
// marker so DXF ordering (common, subclass, geometry, extrusion) is kept.
func (e *entity) FormatBaseExtrusion(f format.Formatter) {
	if len(e.extrusion) == 3 {
		f.WriteFloat(210, e.extrusion[0])
		f.WriteFloat(220, e.extrusion[1])
		f.WriteFloat(230, e.extrusion[2])
	}
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

// DXFType returns the DXF type string for this entity. Retained records with
// non-standard names override it via typeNameOverride.
func (e *entity) DXFType() string {
	if e.typeNameOverride != "" {
		return e.typeNameOverride
	}
	name := EntityTypeString(e.Type)
	// MULTILEADER is the standard DXF spelling; MLEADER is an alias.
	if name == "MULTI_LEADER" {
		return "MULTILEADER"
	}
	return name
}

// GetAttributes returns entity attributes as a map
func (e *entity) GetAttributes() map[string]any {
	attributes := make(map[string]any)
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
func (e *entity) LoadAttributes(attribs map[string]any) error {
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
