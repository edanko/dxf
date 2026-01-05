package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// BlockRecord represents a BLOCK_RECORD table entry
type BlockRecord struct {
	*entity
	Name          string // 2 - block name
	LayoutHandle  string // 340 - handle to associated layout
	Explode       int    // 280 - can explode (0=no, 1=yes)
	UniformScale  int    // 281 - scale uniformly (0=no, 1=yes)
	Units         int    // 70 - units (see ezdxf/units.py)
	BitmapPreview []byte // 310 - binary data for bitmap preview
}

// NewBlockRecord creates a new BlockRecord entity
func NewBlockRecord() *BlockRecord {
	b := &BlockRecord{
		entity:        NewEntity(BLOCKRECORD),
		Name:          "",
		LayoutHandle:  "",
		Explode:       1,
		UniformScale:  0,
		Units:         0,
		BitmapPreview: nil,
	}
	return b
}

// NewBlockRecordWithName creates a block record with the specified name
func NewBlockRecordWithName(name string) *BlockRecord {
	b := NewBlockRecord()
	b.Name = name
	return b
}

// IsEntity is for Entity interface.
func (b *BlockRecord) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (b *BlockRecord) Format(f format.Formatter) {
	b.entity.Format(f)
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbBlockTableRecord")
	f.WriteString(2, b.Name)
	if b.LayoutHandle != "" {
		f.WriteString(340, b.LayoutHandle)
	}
	f.WriteInt(280, b.Explode)
	f.WriteInt(281, b.UniformScale)
	f.WriteInt(70, b.Units)
}

// BBox returns the bounding box of the entity
func (b *BlockRecord) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the BlockRecord
func (b *BlockRecord) Transform(m *math.Matrix44) error {
	return nil
}

// Copy creates a deep copy of the BlockRecord entity
func (b *BlockRecord) Copy() Entity {
	br := NewBlockRecord()
	br.entity = b.entity
	br.Name = b.Name
	br.LayoutHandle = b.LayoutHandle
	br.Explode = b.Explode
	br.UniformScale = b.UniformScale
	br.Units = b.Units
	return br
}

// Validate validates the BlockRecord entity
func (b *BlockRecord) Validate() error {
	if b.Name == "" {
		b.Name = "UNNAMED"
	}
	if b.Explode < 0 || b.Explode > 1 {
		b.Explode = 1
	}
	if b.UniformScale < 0 || b.UniformScale > 1 {
		b.UniformScale = 0
	}
	if b.Units < 0 || b.Units > 25 {
		b.Units = 0
	}
	return nil
}

// CanExplode returns true if the block can be exploded
func (b *BlockRecord) CanExplode() bool {
	return b.Explode != 0
}

// SetExplode sets the explode flag
func (b *BlockRecord) SetExplode(explode bool) {
	if explode {
		b.Explode = 1
	} else {
		b.Explode = 0
	}
}

// IsUniformScale returns true if scaling should be uniform
func (b *BlockRecord) IsUniformScale() bool {
	return b.UniformScale != 0
}

// SetUniformScale sets the uniform scale flag
func (b *BlockRecord) SetUniformScale(uniform bool) {
	if uniform {
		b.UniformScale = 1
	} else {
		b.UniformScale = 0
	}
}

// IsModelSpace returns true if this is a model space block
func (b *BlockRecord) IsModelSpace() bool {
	return b.Name == "*MODEL_SPACE"
}

// IsPaperSpace returns true if this is a paper space block
func (b *BlockRecord) IsPaperSpace() bool {
	return b.Name == "*PAPER_SPACE"
}

// IsAnonymous returns true if this is an anonymous block
func (b *BlockRecord) IsAnonymous() bool {
	return len(b.Name) > 0 && b.Name[0] == '*'
}
