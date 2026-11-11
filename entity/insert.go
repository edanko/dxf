package entity

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// INSERT entity flags
const (
	InsertFlagAnonymous      = 1   // Anonymous block
	InsertFlagHasAttribs     = 2   // Has attributes attached
	InsertFlagAttribsFollow  = 4   // Attributes follow insert (unified)
	InsertFlagExternalXref   = 8   // External reference
	InsertFlagXrefOverlay    = 16  // Xref overlay
	InsertFlagXrefResolved   = 32  // Xref resolved
	InsertFlagXrefDependent  = 64  // Xref dependent
	InsertFlagMultipleInsert = 128 // Multiple insert (MINSERT)
)

// Insert represents a DXF INSERT entity (block reference)
// INSERT inserts a block reference at a specified location with transformation
type Insert struct {
	*entity

	// Basic properties
	blockName     string       // Block name to reference
	insertPoint   dxfmath.Vec3 // Insertion point in WCS
	layerOverride string       // Layer override (empty = use block's layer)

	// Transformation parameters
	xScale    float64      // X scale factor (default = 1)
	yScale    float64      // Y scale factor (default = 1)
	zScale    float64      // Z scale factor (default = 1)
	rotation  float64      // Rotation angle in degrees (default = 0)
	extrusion dxfmath.Vec3 // Extrusion direction (defines OCS)

	// Grid parameters (for MINSERT)
	columnCount   int     // Number of columns (default = 1)
	rowCount      int     // Number of rows (default = 1)
	columnSpacing float64 // Column spacing (default = 0)
	rowSpacing    float64 // Row spacing (default = 0)

	// Block reference data
	block      *block.Block // Referenced block definition
	attributes []*Attrib    // Attached attributes
	flags      int          // INSERT flags

	// Advanced options
	// XCLIP boundary (102 {XCLIP_BOUNDS} group)
	xclipEnabled bool
	xclipType    int // 1 rectangular, 2 polyline
	xclipPoints  [][2]float64
	hasAttribs   bool // Whether insert has attributes
	isMultiple   bool // Whether this is a multi-insert (MINSERT)
	isResolved   bool // Whether XREF is resolved
	isExternal   bool // Whether this is an external reference
}

// NewInsert creates a new INSERT entity
func NewInsert() *Insert {
	return &Insert{
		entity:        NewEntity(INSERT),
		blockName:     "",
		insertPoint:   dxfmath.NewVec3(0, 0, 0),
		layerOverride: "",
		xScale:        1.0,
		yScale:        1.0,
		zScale:        1.0,
		rotation:      0.0,
		extrusion:     dxfmath.NewVec3(0, 0, 1), // Default Z-axis
		columnCount:   1,
		rowCount:      1,
		columnSpacing: 0.0,
		rowSpacing:    0.0,
		block:         nil,
		attributes:    make([]*Attrib, 0),
		flags:         0,
		hasAttribs:    false,
		isMultiple:    false,
		isResolved:    false,
		isExternal:    false,
	}
}

// NewInsertFromBlock creates an INSERT from a block definition
func NewInsertFromBlock(blk *block.Block, insertPoint dxfmath.Vec3) *Insert {
	insert := NewInsert()
	insert.blockName = blk.Name
	insert.insertPoint = insertPoint
	insert.block = blk
	return insert
}

// IsEntity returns true for INSERT entities
// XClip boundary types.
const (
	XClipRectangular = 1
	XClipPolyline    = 2
)

// SetXClipRect clips the inserted block reference to an axis-aligned rectangle.
func (i *Insert) SetXClipRect(x0, y0, x1, y1 float64) {
	i.xclipEnabled = true
	i.xclipType = XClipRectangular
	i.xclipPoints = [][2]float64{{x0, y0}, {x1, y1}}
}

// SetXClipPolyline clips the inserted block reference to a closed polyline.
func (i *Insert) SetXClipPolyline(points [][2]float64) {
	i.xclipEnabled = true
	i.xclipType = XClipPolyline
	i.xclipPoints = append([][2]float64(nil), points...)
}

// RemoveXClip removes the clipping boundary.
func (i *Insert) RemoveXClip() {
	i.xclipEnabled = false
	i.xclipType = 0
	i.xclipPoints = nil
}

// IsXClipped reports whether a clipping boundary is set.
func (i *Insert) IsXClipped() bool { return i.xclipEnabled }

// XClipType returns the boundary type (XClipRectangular or XClipPolyline).
func (i *Insert) XClipType() int { return i.xclipType }

// XClipBoundaryPath returns a copy of the boundary vertices.
func (i *Insert) XClipBoundaryPath() [][2]float64 {
	out := make([][2]float64, len(i.xclipPoints))
	for k, p := range i.xclipPoints {
		out[k] = p
	}
	return out
}

func (i *Insert) IsEntity() bool {
	return true
}

// Format writes INSERT entity data to DXF format
func (i *Insert) Format(f format.Formatter) {
	i.entity.Format(f)
	f.WriteString(100, "AcDbBlockReference")

	// Write block name (2)
	if i.blockName != "" {
		f.WriteString(2, i.blockName)
	}

	// Write insertion point (10,20,30)
	f.WriteFloat(10, i.insertPoint.X())
	f.WriteFloat(20, i.insertPoint.Y())
	f.WriteFloat(30, i.insertPoint.Z())

	// Write transformation parameters
	if i.xScale != 1.0 || i.yScale != 1.0 || i.zScale != 1.0 {
		f.WriteFloat(41, i.xScale)
		f.WriteFloat(42, i.yScale)
		f.WriteFloat(43, i.zScale)
	}

	if i.rotation != 0.0 {
		f.WriteFloat(50, i.rotation)
	}

	// Write extrusion direction (210,220,230)
	if !i.extrusion.IsEqual(dxfmath.NewVec3(0, 0, 1), 1e-9) {
		f.WriteFloat(210, i.extrusion.X())
		f.WriteFloat(220, i.extrusion.Y())
		f.WriteFloat(230, i.extrusion.Z())
	}

	// Write layer override
	if i.layerOverride != "" {
		f.WriteString(8, i.layerOverride)
	}

	// Write flags
	if i.flags != 0 {
		f.WriteInt(70, i.flags)
	}

	// Write grid parameters for MINSERT
	if i.isMultiple {
		f.WriteInt(71, i.columnCount)
		f.WriteFloat(44, i.columnSpacing)
		f.WriteInt(72, i.rowCount)
		f.WriteFloat(45, i.rowSpacing)
	}

	// Write attributes if any
	if len(i.attributes) > 0 {
		f.WriteInt(66, 1) // Attributes follow flag
		last := len(i.attributes) - 1
		for idx, attrib := range i.attributes {
			attrib.Format(f)
			if idx == last && attrib.EndHandle() != "" {
				f.WriteString(0, "SEQEND")
				f.WriteString(5, attrib.EndHandle())
			}
		}
	}
	// Write the XCLIP boundary group (102 {XCLIP_BOUNDS} ... 102 }).
	if i.xclipEnabled && len(i.xclipPoints) > 0 {
		f.WriteString(102, "{XCLIP_BOUNDS")
		f.WriteInt(71, i.xclipType)
		f.WriteInt(91, len(i.xclipPoints))
		for _, p := range i.xclipPoints {
			f.WriteFloat(10, p[0])
			f.WriteFloat(20, p[1])
		}
		f.WriteString(102, "}")
	}
}

// SetScaleX sets the X scale factor (code 41).
func (i *Insert) SetScaleX(v float64) { i.xScale = v }

// SetScaleY sets the Y scale factor (code 42).
func (i *Insert) SetScaleY(v float64) { i.yScale = v }

// SetScaleZ sets the Z scale factor (code 43).
func (i *Insert) SetScaleZ(v float64) { i.zScale = v }

// SetExtrusionComponent sets one component of the extrusion direction.
func (i *Insert) SetExtrusionComponent(idx int, v float64) {
	cur := i.extrusion
	if len(cur) < 3 {
		cur = dxfmath.Vec3{0, 0, 1}
	}
	cur[idx] = v
	i.extrusion = cur
}

// Flags returns the insert flags (code 70).
func (i *Insert) Flags() int { return i.flags }

// SetFlags sets the insert flags (code 70).
func (i *Insert) SetFlags(v int) { i.flags = v }

// SetBlockName sets the block name to reference
func (i *Insert) SetBlockName(name string) {
	i.blockName = name
}

// SetInsertPoint sets the insertion point
func (i *Insert) SetInsertPoint(point dxfmath.Vec3) {
	i.insertPoint = point
}

// SetScale sets uniform scaling
func (i *Insert) SetScale(scale float64) {
	i.xScale = scale
	i.yScale = scale
	i.zScale = scale
}

// SetNonUniformScale sets individual scale factors
func (i *Insert) SetNonUniformScale(x, y, z float64) {
	i.xScale = x
	i.yScale = y
	i.zScale = z
}

// SetRotation sets the rotation angle
func (i *Insert) SetRotation(angle float64) {
	i.rotation = angle
}

// SetExtrusion sets the extrusion direction (OCS definition)
func (i *Insert) SetExtrusion(extrusion dxfmath.Vec3) {
	normalized := extrusion.Normalized()
	if !normalized.IsEqual(dxfmath.NewVec3(0, 0, 0), 1e-9) {
		i.extrusion = normalized
	} else {
		i.extrusion = dxfmath.NewVec3(0, 0, 1)
	}
}

// SetGrid sets grid parameters for multi-insert
func (i *Insert) SetGrid(cols, rows int, colSpacing, rowSpacing float64) {
	i.columnCount = cols
	i.rowCount = rows
	i.columnSpacing = colSpacing
	i.rowSpacing = rowSpacing
	i.isMultiple = cols > 1 || rows > 1
}

// BlockDef returns the resolved block definition, when bound.
func (i *Insert) BlockDef() *block.Block { return i.block }

// SetBlockDef binds a resolved block definition to this insert.
func (i *Insert) SetBlockDef(b *block.Block) { i.block = b }

// ColumnCount returns the MINSERT column count (code 71).
func (i *Insert) ColumnCount() int { return i.columnCount }

// RowCount returns the MINSERT row count (code 72).
func (i *Insert) RowCount() int { return i.rowCount }

// ColumnSpacing returns the MINSERT column spacing (code 44).
func (i *Insert) ColumnSpacing() float64 { return i.columnSpacing }

// RowSpacing returns the MINSERT row spacing (code 45).
func (i *Insert) RowSpacing() float64 { return i.rowSpacing }

// SetLayerOverride sets the layer override
func (i *Insert) SetLayerOverride(layer string) {
	i.layerOverride = layer
}

// SetAnonymous sets the anonymous flag
func (i *Insert) SetAnonymous(anonymous bool) {
	if anonymous {
		i.flags |= InsertFlagAnonymous
	} else {
		i.flags &= ^InsertFlagAnonymous
	}
}

// SetHasAttribs sets the has attributes flag
func (i *Insert) SetHasAttribs(hasAttribs bool) {
	i.hasAttribs = hasAttribs
	if hasAttribs {
		i.flags |= InsertFlagHasAttribs
	} else {
		i.flags &= ^InsertFlagHasAttribs
	}
}

// SetAttribsFollow sets the attributes follow flag
func (i *Insert) SetAttribsFollow(follow bool) {
	if follow {
		i.flags |= InsertFlagAttribsFollow
	} else {
		i.flags &= ^InsertFlagAttribsFollow
	}
}

// SetExternalXref sets the external XREF flag
func (i *Insert) SetExternalXref(external bool) {
	i.isExternal = external
	if external {
		i.flags |= InsertFlagExternalXref
	} else {
		i.flags &= ^InsertFlagExternalXref
	}
}

// SetXrefOverlay sets the XREF overlay flag
func (i *Insert) SetXrefOverlay(overlay bool) {
	if overlay {
		i.flags |= InsertFlagXrefOverlay
	} else {
		i.flags &= ^InsertFlagXrefOverlay
	}
}

// SetXrefResolved sets the XREF resolved flag
func (i *Insert) SetXrefResolved(resolved bool) {
	i.isResolved = resolved
	if resolved {
		i.flags |= InsertFlagXrefResolved
	} else {
		i.flags &= ^InsertFlagXrefResolved
	}
}

// SetXrefDependent sets the XREF dependent flag
func (i *Insert) SetXrefDependent(dependent bool) {
	if dependent {
		i.flags |= InsertFlagXrefDependent
	} else {
		i.flags &= ^InsertFlagXrefDependent
	}
}

// SetMultiple sets the multiple insert flag
func (i *Insert) SetMultiple(multiple bool) {
	i.isMultiple = multiple
	if multiple {
		i.flags |= InsertFlagMultipleInsert
	} else {
		i.flags &= ^InsertFlagMultipleInsert
	}
}

// AddAttribute adds an attribute to the insert
func (i *Insert) AddAttribute(attrib *Attrib) {
	i.attributes = append(i.attributes, attrib)
	i.hasAttribs = true
	i.flags |= InsertFlagHasAttribs
}

// RemoveAttribute removes an attribute by index
func (i *Insert) RemoveAttribute(index int) {
	if index >= 0 && index < len(i.attributes) {
		i.attributes = append(i.attributes[:index], i.attributes[index+1:]...)
		if len(i.attributes) == 0 {
			i.hasAttribs = false
			i.flags &= ^InsertFlagHasAttribs
		}
	}
}

// GetAttribute returns an attribute by tag
func (i *Insert) GetAttribute(tag string) *Attrib {
	for _, attrib := range i.attributes {
		if attrib.tag == tag {
			return attrib
		}
	}
	return nil
}

// Subentities returns the attribute records that follow the INSERT in the DXF
// stream. SEQEND is structural and is not included.
func (i *Insert) Subentities() []Entity {
	subs := make([]Entity, 0, len(i.attributes))
	for _, attrib := range i.attributes {
		subs = append(subs, attrib)
	}
	return subs
}

// GetAllAttributes returns all attached attributes
func (i *Insert) GetAllAttributes() []*Attrib {
	return i.attributes
}

// BlockName returns the referenced block name
func (i *Insert) BlockName() string {
	return i.blockName
}

// InsertPoint returns the insertion point
func (i *Insert) InsertPoint() dxfmath.Vec3 {
	return i.insertPoint
}

// Scale returns the scale factors as a vector
func (i *Insert) Scale() dxfmath.Vec3 {
	return dxfmath.NewVec3(i.xScale, i.yScale, i.zScale)
}

// Rotation returns the rotation angle in degrees
func (i *Insert) Rotation() float64 {
	return i.rotation
}

// Extrusion returns the extrusion direction
func (i *Insert) Extrusion() dxfmath.Vec3 {
	return i.extrusion
}

// IsUniformScale returns true if all scale factors are equal
func (i *Insert) IsUniformScale() bool {
	return i.xScale == i.yScale && i.yScale == i.zScale
}

// IsMultiple returns true if this is a multi-insert
func (i *Insert) IsMultiple() bool {
	return i.isMultiple
}

// BBox returns bounding box for INSERT entity
func (i *Insert) BBox() ([]float64, []float64) {
	// For INSERT, we need to get the block definition
	if i.block == nil {
		// Return empty bounding box if no block
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}

	basePoint := i.block.Coord
	insertPoint := i.insertPoint

	mins := []float64{
		insertPoint.X() + basePoint[0],
		insertPoint.Y() + basePoint[1],
		insertPoint.Z() + basePoint[2],
	}
	maxs := []float64{
		insertPoint.X() + basePoint[0],
		insertPoint.Y() + basePoint[1],
		insertPoint.Z() + basePoint[2],
	}

	// Widen the box with the bounding boxes of the block content entities,
	// scaled and rotated around the insertion point.
	content := i.block.Entities()
	for _, ent := range content {
		e, ok := ent.(Entity)
		if !ok {
			continue
		}
		bb, ok := e.(Bounder)
		if !ok {
			continue
		}
		cmin, cmax := bb.BBox()
		if len(cmin) < 2 || len(cmax) < 2 {
			continue
		}
		sx, sy, sz := i.xScale, i.yScale, i.zScale
		for _, lo := range [][3]float64{{cmin[0], cmin[1], cmin[2]}, {cmax[0], cmax[1], cmin[2]}} {
			x := insertPoint.X() + (lo[0]+basePoint[0])*sx
			y := insertPoint.Y() + (lo[1]+basePoint[1])*sy
			z := insertPoint.Z() + lo[2]*sz
			for k, v := range []float64{x, y, z} {
				if v < mins[k] {
					mins[k] = v
				}
				if v > maxs[k] {
					maxs[k] = v
				}
			}
		}
	}

	return mins, maxs
}

// Transform applies an affine matrix to the insert. The insertion point,
// scale factors, and rotation compose with the matrix (2D rotation component
// only). Extrusion direction transforms as a vector.
func (i *Insert) Transform(m dxfmath.Matrix44) error {
	i.insertPoint = dxfmath.Vec3(TransformPoint(m, []float64{i.insertPoint.X(), i.insertPoint.Y(), i.insertPoint.Z()}))
	sx, sy, sz := ScaleFactors(m)
	i.xScale *= sx
	i.yScale *= sy
	i.zScale *= sz
	if rot := math.Atan2(m.Get(1, 0), m.Get(0, 0)); rot != 0 {
		i.rotation += rot * 180 / math.Pi
	}
	if len(i.extrusion) >= 3 {
		d := TransformDirection(m, []float64{i.extrusion[0], i.extrusion[1], i.extrusion[2]})
		n := dxfmath.NewVec3(d[0], d[1], d[2]).Normalized()
		i.extrusion = n
	}
	return nil
}

// insertMatrix composes the block-to-world transform: scale, Z rotation,
// then translation to the insertion point (plus the block definition base
// point offset).
// InsertMatrix returns the composed insert placement matrix (base point,
// scale, rotation, insertion point) mapping block-definition coordinates to
// the insert coordinate system.
func (i *Insert) InsertMatrix() dxfmath.Matrix44 { return i.insertMatrix() }

func (i *Insert) insertMatrix() dxfmath.Matrix44 {
	scl := dxfmath.Scale(i.xScale, i.yScale, i.zScale)
	rot := dxfmath.RotationZ(i.rotation * math.Pi / 180)
	trl := dxfmath.Translation(i.insertPoint.X(), i.insertPoint.Y(), i.insertPoint.Z())
	off := dxfmath.NewMatrix44()
	if i.block != nil && len(i.block.Coord) >= 3 {
		off = dxfmath.Translation(i.block.Coord[0], i.block.Coord[1], i.block.Coord[2])
	}
	// Application order: base offset, then scale, rotation, insertion point.
	// The composition places the first-applied factor rightmost (the library's
	// MulVec applies the composed matrix directly).
	m := trl.Mul(rot).Mul(scl).Mul(off)
	return m
}

// Explode expands the insert: every block-definition entity is deep-copied
// and transformed into the insert coordinate system (block base point, scale,
// rotation, insertion point). Nested inserts expand one level.
func (i *Insert) Explode() ([]Entity, error) {
	if i.block == nil {
		return nil, fmt.Errorf("insert %q: unresolved block definition", i.blockName)
	}
	m := i.insertMatrix()
	var out []Entity
	for _, be := range i.block.Entities() {
		ent, ok := be.(Entity)
		if !ok {
			continue
		}
		if w, ok := ent.(Unwrapper); ok {
			ent = w.Unwrap()
		}
		cp, ok := ent.(Copier)
		var clone Entity
		if ok {
			clone = cp.Copy()
		} else {
			// Entities without a deep-copy implementation are emitted as
			// shallow clones: the tag data is shared, the geometry mutated by
			// the insert transform must not be aliased. Types that own nested
			// slices (POLYLINE vertices) implement Copier and take the deep path.
			clone = ent
		}
		if tr, ok := clone.(Transformer); ok {
			if err := tr.Transform(m); err != nil {
				return nil, fmt.Errorf("insert %q: transform block entity %s: %w", i.blockName, ent.DXFType(), err)
			}
		}
		out = append(out, clone)
	}
	return out, nil
}

// Copy creates a deep copy of the Insert entity
func (i *Insert) Copy() Entity {
	insert := NewInsert()
	insert.entity = i.entity.clone()
	insert.blockName = i.blockName
	insert.insertPoint = i.insertPoint
	insert.layerOverride = i.layerOverride
	insert.xScale = i.xScale
	insert.yScale = i.yScale
	insert.zScale = i.zScale
	insert.rotation = i.rotation
	insert.extrusion = i.extrusion
	insert.columnCount = i.columnCount
	insert.rowCount = i.rowCount
	insert.columnSpacing = i.columnSpacing
	insert.rowSpacing = i.rowSpacing
	insert.block = i.block
	insert.flags = i.flags
	insert.hasAttribs = i.hasAttribs
	insert.isMultiple = i.isMultiple
	insert.isResolved = i.isResolved
	insert.isExternal = i.isExternal

	// Deep copy attributes
	if len(i.attributes) > 0 {
		insert.attributes = make([]*Attrib, len(i.attributes))
		for idx, attr := range i.attributes {
			if attr != nil {
				insert.attributes[idx] = attr.Copy().(*Attrib)
			}
		}
	}

	return insert
}
