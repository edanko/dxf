package block

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

// BlockEntity is the subset of entity behavior needed by a block definition.
// It avoids an import cycle between the block and entity packages.
type BlockEntity interface {
	IsEntity() bool
	Format(format.Formatter)
	Handle() string
}

// Block represents each BLOCK.
type Block struct {
	Name        string
	Description string
	handle      string
	endhandle   string
	layer       *table.Layer
	layerName   string
	Flag        int
	Coord       []float64
	entities    []BlockEntity
}

// NewBlock create a new Block.
func NewBlock(name, desc string) *Block {
	b := &Block{
		Name:        name,
		Description: desc,
		Flag:        0,
		Coord:       []float64{0.0, 0.0, 0.0},
	}
	return b
}

// Format writes data to formatter.
func (b *Block) Format(f format.Formatter) {
	f.WriteString(0, "BLOCK")
	if b.handle != "" {
		f.WriteString(5, b.handle)
	}
	f.WriteString(100, "AcDbEntity")
	layerName := b.layerName
	if b.layer != nil {
		layerName = b.layer.Name()
	}
	if layerName != "" {
		f.WriteString(8, layerName)
	}
	f.WriteString(100, "AcDbBlockBegin")
	f.WriteString(2, b.Name)
	f.WriteInt(70, b.Flag)
	for i := 0; i < 3 && i < len(b.Coord); i++ {
		f.WriteFloat((i+1)*10, b.Coord[i])
	}
	f.WriteString(3, b.Name)
	f.WriteString(1, b.Description)
	for _, e := range b.entities {
		e.Format(f)
	}
	f.WriteString(0, "ENDBLK")
	if b.endhandle != "" {
		f.WriteString(5, b.endhandle)
	}
	f.WriteString(100, "AcDbEntity")
	if layerName != "" {
		f.WriteString(8, layerName)
	}
	f.WriteString(100, "AcDbBlockEnd")
}

// Handle returns a handle value of BLOCK.
func (b *Block) Handle() string {
	return b.handle
}

// SetHandle sets handles to BLOCK and ENDBLK.
func (b *Block) SetHandle(hg *handle.HandleGenerator) {
	if b.handle == "" {
		b.handle = hg.Next()
	}
	if b.endhandle == "" {
		b.endhandle = hg.Next()
	}
}

// SetHandleString sets the BLOCK handle explicitly.
func (b *Block) SetHandleString(h string) { b.handle = h }

// SetEndHandleString sets the ENDBLK handle explicitly.
func (b *Block) SetEndHandleString(h string) { b.endhandle = h }

// SetLayerName records the layer name until the layer table is resolved.
func (b *Block) SetLayerName(name string) { b.layerName = name }

// AddEntity appends an entity contained in the block.
func (b *Block) AddEntity(e BlockEntity) { b.entities = append(b.entities, e) }

// Entities returns the entities contained in the block.
func (b *Block) Entities() []BlockEntity { return b.entities }

// Layer returns BLOCK's Layer.
func (b *Block) Layer() *table.Layer {
	return b.layer
}

// SetLayer sets Layer to BLOCK.
func (b *Block) SetLayer(layer *table.Layer) {
	b.layer = layer
}
