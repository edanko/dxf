// BLOCK section
package block

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/table"
)

// Blocks represents BLOCKS section.
type Blocks []*Block

// New creates a new Blocks.
func New(layer *table.Layer) Blocks {
	b := make([]*Block, 3)
	b[0] = NewBlock("*Model_Space", "")
	b[0].SetLayer(layer)
	b[1] = NewBlock("*Paper_Space", "")
	b[1].SetLayer(layer)
	b[2] = NewBlock("*Paper_Space0", "")
	b[2].SetLayer(layer)
	return b
}

// Format writes BLOCKS data to formatter.
func (bs Blocks) Format(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "BLOCKS")
	for _, b := range bs {
		b.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

// Add adds a new block to BLOCKS section.
func (bs Blocks) Add(b *Block) Blocks {
	bs = append(bs, b)
	return bs
}

// SetHandle sets handles to each block.
func (bs Blocks) SetHandle(v *int) {
	for _, b := range bs {
		b.SetHandle(v)
	}
}
