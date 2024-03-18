package block

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

// Block represents each BLOCK.
type Block struct {
	Name        string
	Description string
	handle      string
	endhandle   string
	layer       *table.Layer
	Flag        int
	Coord       []float64
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
	f.WriteString(5, b.handle)
	f.WriteString(100, "AcDbEntity")
	f.WriteString(8, b.layer.Name())
	f.WriteString(100, "AcDbBlockBegin")
	f.WriteString(2, b.Name)
	f.WriteInt(70, b.Flag)
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, b.Coord[i])
	}
	f.WriteString(3, b.Name)
	f.WriteString(1, b.Description)
	f.WriteString(0, "ENDBLK")
	f.WriteString(5, b.endhandle)
	f.WriteString(100, "AcDbEntity")
	f.WriteString(8, b.layer.Name())
	f.WriteString(100, "AcDbBlockEnd")
}

// Handle returns a handle value of BLOCK.
func (b *Block) Handle() string {
	return b.handle
}

// SetHandle sets handles to BLOCK and ENDBLK.
func (b *Block) SetHandle(hg *handle.HandleGenerator) {
	b.handle = hg.Next()
	b.endhandle = hg.Next()
}

// Layer returns BLOCK's Layer.
func (b *Block) Layer() *table.Layer {
	return b.layer
}

// SetLayer sets Layer to BLOCK.
func (b *Block) SetLayer(layer *table.Layer) {
	b.layer = layer
}
