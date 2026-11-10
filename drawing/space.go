// High-level authoring facade: layout spaces with ezdxf-style Add<Type>
// methods, attribute options, block-definition spaces, groups, selection
// helpers, and document extents.

package drawing

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/table"
)

// attrSetter exposes the shared base attributes that are not part of the
// minimal Entity interface; every concrete entity satisfies it.
type attrSetter interface {
	SetLayerName(string)
	SetLinetype(string)
	SetLineWeight(int)
	SetColorRGB(int)
	SetInvisible(bool)
}

// Space is a drawing area (model space, a paper-space layout, or a block
// definition) that accepts entities through Add and the Add<Type> factory
// methods. Entities added to a space receive an automatic handle and, on
// save, an owner handle pointing at the BLOCK_RECORD of the space.
type Space struct {
	drawing   *Drawing
	spaceName string // layout name ("Model", layout name, or block name)
	blockName string // "*Model_Space", layout block name, ...
	owner     ownerRef
	isBlock   bool
}

// ownerRef pins a table entry's handle string as an owner handler.
type ownerRef struct{ h string }

// Handle returns the pinned handle string.
func (o ownerRef) Handle() string { return o.h }

// SetHandle is a no-op; the handle was captured from the table entry.
func (o ownerRef) SetHandle(hg *handle.HandleGenerator) {}

// Name returns the layout ("Model", layout name) or block name.
func (s *Space) Name() string { return s.spaceName }

// BlockName returns the DXF block owning this space.
func (s *Space) BlockName() string { return s.blockName }

// Drawing returns the drawing this space belongs to.
func (s *Space) Drawing() *Drawing { return s.drawing }

// Model returns the model-space authoring space.
func (d *Drawing) Model() *Space {
	return &Space{drawing: d, spaceName: "Model", blockName: "*Model_Space"}
}

// SpaceByName returns the authoring space for a layout name or an existing
// block definition name.
func (d *Drawing) SpaceByName(name string) (*Space, error) {
	if name == "" {
		return nil, fmt.Errorf("empty space name")
	}
	if name == "Model" {
		return d.Model(), nil
	}
	if le := d.layouts[name]; le != nil {
		return &Space{drawing: d, spaceName: name, blockName: le.blockName}, nil
	}
	b, err := d.GetBlock(name)
	if err != nil {
		return nil, fmt.Errorf("space %q: no layout or block", name)
	}
	return d.BlockSpace(b), nil
}

// BlockSpace returns the authoring space for a block definition. Entities
// added to it are stored inside the block, between BLOCK and ENDBLK.
func (d *Drawing) BlockSpace(b *block.Block) *Space {
	return &Space{drawing: d, spaceName: b.Name, blockName: b.Name, isBlock: true}
}

// EntityOption configures common DXF attributes of an entity at add time,
// mirroring ezdxf's dxfattribs dictionary.
type EntityOption func(entity.Entity) error

// WithLayer assigns the layer by name, creating it when missing.
func (d *Drawing) WithLayer(name string) EntityOption {
	return func(e entity.Entity) error {
		l, ok := d.Layers[name]
		if !ok {
			var lt *table.LineType
			if st, err := d.tablesSection[1].Lookup("Continuous"); err == nil {
				lt, _ = st.(*table.LineType)
			}
			var err error
			l, err = d.AddLayer(name, color.White, lt)
			if err != nil {
				return fmt.Errorf("with-layer %q: %w", name, err)
			}
		}
		e.SetLayer(l)
		if as, ok := e.(attrSetter); ok {
			as.SetLayerName(name)
		}
		return nil
	}
}

// WithColor assigns the color number (code 62).
func WithColor(c color.ColorNumber) EntityOption {
	return func(e entity.Entity) error {
		e.SetColor(c)
		return nil
	}
}

// WithLinetype assigns the linetype name (code 6).
func WithLinetype(name string) EntityOption {
	return func(e entity.Entity) error {
		as, ok := e.(attrSetter)
		if !ok {
			return fmt.Errorf("with-linetype: %T does not support attributes", e)
		}
		as.SetLinetype(name)
		return nil
	}
}

// WithLineweight assigns the lineweight (code 370).
func WithLineweight(w int) EntityOption {
	return func(e entity.Entity) error {
		as, ok := e.(attrSetter)
		if !ok {
			return fmt.Errorf("with-lineweight: %T does not support attributes", e)
		}
		as.SetLineWeight(w)
		return nil
	}
}

// WithTrueColor assigns the RGB true color (code 420).
func WithTrueColor(rgb int) EntityOption {
	return func(e entity.Entity) error {
		as, ok := e.(attrSetter)
		if !ok {
			return fmt.Errorf("with-true-color: %T does not support attributes", e)
		}
		as.SetColorRGB(rgb)
		return nil
	}
}

// WithInvisible marks the entity invisible (code 60).
func WithInvisible(e entity.Entity) error {
	as, ok := e.(attrSetter)
	if !ok {
		return fmt.Errorf("with-invisible: %T does not support attributes", e)
	}
	as.SetInvisible(true)
	return nil
}

// WithLineScale assigns the linetype scale (code 48).
func WithLineScale(scale float64) EntityOption {
	return func(e entity.Entity) error {
		e.SetLtscale(scale)
		return nil
	}
}

// WithAttribs attaches attribute values to an INSERT (key = tag).
func (s *Space) WithAttribs(values map[string]string) EntityOption {
	return func(e entity.Entity) error {
		ins, ok := e.(*entity.Insert)
		if !ok {
			return fmt.Errorf("with-attribs: %T is not an INSERT", e)
		}
		ip := ins.InsertPoint()
		for tag, text := range values {
			attrib := entity.NewAttribWithValues(tag, text, ip[0], ip[1], ip[2])
			ins.AddAttribute(attrib)
		}
		ins.SetHasAttribs(true)
		ins.SetAttribsFollow(true)
		return nil
	}
}

// Add registers an entity in this space and applies the options.
func (s *Space) Add(e entity.Entity, opts ...EntityOption) error {
	if e == nil {
		return fmt.Errorf("cannot add nil entity")
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(e); err != nil {
			return err
		}
	}
	if s.isBlock {
		b, err := s.drawing.GetBlock(s.blockName)
		if err != nil {
			return fmt.Errorf("space %q: %w", s.spaceName, err)
		}
		b.AddEntity(asBlockEntity(e))
		return nil
	}
	if err := s.drawing.AddEntity(e); err != nil {
		return err
	}
	s.drawing.recordTransactionActivity(1, 0)
	if s.blockName != "*Model_Space" {
		s.drawing.assignLayout(e.Handle(), s.spaceName)
	}
	return nil
}

// blockEntity is the interface satisfied by entities storable in a block.
type blockEntity interface {
	IsEntity() bool
	Format(f format.Formatter)
	Handle() string
	DXFType() string
}

// AddGeneric creates an entity of the given DXF type from an attribute map
// and adds it to this space, mirroring ezdxf's layout.add(entity) plus the
// attribs dictionary. Coordinate keys are flat names ("x", "y", "z",
// "x1".."z2", "center" via x/y/z, "radius", "text", "tag", "vertices" as
// []float64).
func (s *Space) AddGeneric(dxfType string, attribs map[string]any, opts ...EntityOption) (entity.Entity, error) {
	e, err := newEntityByType(dxfType, attribs)
	if err != nil {
		return nil, err
	}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

func attrFloat(attribs map[string]any, key string) float64 {
	if v, ok := attribs[key].(float64); ok {
		return v
	}
	if v, ok := attribs[key].(int); ok {
		return float64(v)
	}
	return 0
}

func attrFloats(attribs map[string]any, key string) []float64 {
	if v, ok := attribs[key].([]float64); ok {
		return v
	}
	return nil
}

func attrString(attribs map[string]any, key string) string {
	if v, ok := attribs[key].(string); ok {
		return v
	}
	return ""
}

// newEntityByType builds a standard entity from flat attributes; it is the
// generic counterpart of the typed Add<Type> helpers.
func newEntityByType(dxfType string, a map[string]any) (entity.Entity, error) {
	switch strings.ToUpper(dxfType) {
	case "POINT":
		e := entity.NewPoint()
		e.Coord = []float64{attrFloat(a, "x"), attrFloat(a, "y"), attrFloat(a, "z")}
		return e, nil
	case "LINE":
		e := entity.NewLine()
		e.Start = []float64{attrFloat(a, "x1"), attrFloat(a, "y1"), attrFloat(a, "z1")}
		e.End = []float64{attrFloat(a, "x2"), attrFloat(a, "y2"), attrFloat(a, "z2")}
		return e, nil
	case "CIRCLE":
		e := entity.NewCircle()
		e.Center = []float64{attrFloat(a, "x"), attrFloat(a, "y"), attrFloat(a, "z")}
		e.Radius = attrFloat(a, "radius")
		return e, nil
	case "ARC":
		base := entity.NewCircle()
		base.Center = []float64{attrFloat(a, "x"), attrFloat(a, "y"), attrFloat(a, "z")}
		base.Radius = attrFloat(a, "radius")
		e := entity.NewArc(base)
		e.Angle = []float64{attrFloat(a, "startangle"), attrFloat(a, "endangle")}
		return e, nil
	case "TEXT":
		e := entity.NewText()
		e.Coord1 = []float64{attrFloat(a, "x"), attrFloat(a, "y"), attrFloat(a, "z")}
		e.Coord2 = []float64{0, 0, 0}
		e.Height = attrFloat(a, "height")
		e.Value = attrString(a, "text")
		return e, nil
	case "MTEXT":
		e := entity.NewMText()
		e.Coord1 = []float64{attrFloat(a, "x"), attrFloat(a, "y"), attrFloat(a, "z")}
		e.RectWidth = attrFloat(a, "width")
		e.SetHeight(attrFloat(a, "height"))
		e.Value = attrString(a, "text")
		return e, nil
	case "ELLIPSE":
		e := entity.NewEllipse()
		e.Center = []float64{attrFloat(a, "x"), attrFloat(a, "y"), attrFloat(a, "z")}
		e.Major = attrFloat(a, "majoraxis")
		e.Minor = attrFloat(a, "minoraxis")
		return e, nil
	case "LWPOLYLINE":
		pts := attrFloats(a, "vertices")
		count := len(pts) / 2
		e := entity.NewLwPolyline(count)
		for i := 0; i < count && i < len(e.Vertices); i++ {
			e.Vertices[i] = []float64{pts[i*2], pts[i*2+1], 0}
		}
		return e, nil
	case "POLYLINE":
		pts := attrFloats(a, "vertices")
		e := entity.NewPolyline()
		for i := 0; i+1 < len(pts); i += 2 {
			e.AddVertex(pts[i], pts[i+1], 0)
		}
		return e, nil
	case "SPLINE":
		pts := attrFloats(a, "vertices")
		e := entity.NewSpline(3)
		for i := 0; i+2 < len(pts); i += 3 {
			e.Controls = append(e.Controls, []float64{pts[i], pts[i+1], pts[i+2]})
		}
		return e, nil
	case "INSERT":
		e := entity.NewInsert()
		e.SetInsertPoint(dxfmath.NewVec3(attrFloat(a, "x"), attrFloat(a, "y"), attrFloat(a, "z")))
		e.SetBlockName(attrString(a, "blockname"))
		return e, nil
	}
	return nil, fmt.Errorf("AddGeneric: no generic builder for entity type %q", dxfType)
}

// blockEntityAdapter adapts an entity.Entity to block.BlockEntity. Library
// entities that satisfy the interface directly are passed through; the rest
// are wrapped with the raw-tag methods (mirroring the loader adapter).
type blockEntityAdapter struct {
	entity.Entity
	raw []block.TagPair
}

// Unwrap exposes the adapted entity for consumers that need the concrete type.
func (a *blockEntityAdapter) Unwrap() entity.Entity { return a.Entity }

// RawTags returns the retained group codes.
func (a *blockEntityAdapter) RawTags() []block.TagPair { return a.raw }

// SetRawTags stores retained group codes.
func (a *blockEntityAdapter) SetRawTags(tags []block.TagPair) { a.raw = tags }

func asBlockEntity(e entity.Entity) block.BlockEntity {
	if be, ok := e.(block.BlockEntity); ok {
		return be
	}
	return &blockEntityAdapter{Entity: e}
}

// assignLayout records the layout an entity belongs to.
func (d *Drawing) assignLayout(entityHandle, layoutName string) {
	if d.entityLayout == nil {
		d.entityLayout = make(map[string]string)
	}
	d.entityLayout[entityHandle] = layoutName
}

// blockRecordTable returns the BLOCK_RECORD symbol table.
func (d *Drawing) blockRecordTable() *table.Table {
	for _, t := range d.tablesSection {
		if t.Name() == "BLOCK_RECORD" {
			return t
		}
	}
	return nil
}

// BlockRecordHandle returns the BLOCK_RECORD handle for a block name.
func (d *Drawing) BlockRecordHandle(name string) (string, bool) {
	return d.blockRecordHandle(name)
}

// blockRecordHandle returns the BLOCK_RECORD handle for a block name.
func (d *Drawing) blockRecordHandle(name string) (string, bool) {
	t := d.blockRecordTable()
	if t == nil {
		return "", false
	}
	st, err := t.Lookup(name)
	if err != nil {
		return "", false
	}
	return st.Handle(), true
}

// ownerSetter exposes the owner accessors that the minimal Entity
// interface omits; every standard entity satisfies it.
type ownerSetter interface {
	SetOwner(handle.Handler)
	Owner() handle.Handler
}

// assignOwners sets the owner (code 330) of every entity to the BLOCK_RECORD
// of its space; entity handles of block-internal entities are allocated.
func (d *Drawing) assignOwners() {
	for _, b := range d.blocksSection {
		for _, be := range b.Entities() {
			var target entity.Entity
			if u, ok := be.(interface{ Unwrap() entity.Entity }); ok {
				target = u.Unwrap()
			} else if ent, ok := be.(entity.Entity); ok {
				target = ent
			} else {
				continue
			}
			if sh, ok := target.(interface{ SetHandle(*handle.HandleGenerator) }); ok && target.Handle() == "" {
				sh.SetHandle(d.handleGenarator)
			}
			if h, ok := d.blockRecordHandle(b.Name); ok {
				if setter, ok := target.(interface{ SetOwner(handle.Handler) }); ok {
					setter.SetOwner(ownerRef{h})
				}
			}
		}
	}
	for _, e := range d.entitiesSection {
		owner, ok := e.(ownerSetter)
		if !ok {
			continue
		}
		if owner.Owner() != nil {
			continue
		}
		blockName := "*Model_Space"
		if layout := d.entityLayout[e.Handle()]; layout != "" {
			if le := d.layouts[layout]; le != nil {
				blockName = le.blockName
			}
		}
		if h, ok := d.blockRecordHandle(blockName); ok {
			owner.SetOwner(ownerRef{h})
		}
	}
}

// EntitySpace returns the layout name an entity was added to ("Model" when
// unassigned).
func (d *Drawing) EntitySpace(handleString string) string {
	if layout := d.entityLayout[handleString]; layout != "" {
		return layout
	}
	return "Model"
}

// layoutNameFilter is shared by selection helpers.
func normalizeLayoutName(name string) string {
	if strings.EqualFold(name, "model") {
		return "Model"
	}
	return name
}
