package entity

import (
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
	hasAttribs bool // Whether insert has attributes
	isMultiple bool // Whether this is a multi-insert (MINSERT)
	isResolved bool // Whether XREF is resolved
	isExternal bool // Whether this is an external reference
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
		for _, attrib := range i.attributes {
			attrib.Format(f)
		}
	}
}

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
	i.isMultiple = (cols > 1 || rows > 1)
}

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

	// Use block coordinate as base point (simplified)
	// A full implementation would transform all block entities
	basePoint := i.block.Coord
	insertPoint := i.insertPoint

	// Apply transformation (simplified - just use insert point + block base)
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

	return mins, maxs
}
