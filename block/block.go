package block

import (
	"strings"

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
	DXFType() string
	RawTags() []TagPair
	SetRawTags([]TagPair)
}

// TagPair is a group code/value pair retained on an entity. It mirrors
// entity.TagPair; the interface method signatures are structural, so any
// embedded entity satisfies it without importing the entity package.
type TagPair struct {
	Code  int
	Value string
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
	rawTags     []TagPair
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
	if len(b.rawTags) > 0 {
		b.formatRaw(f)
		return
	}
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

// formatRaw writes the block from its retained group codes. The raw tags
// contain BLOCK header, entity boundary tags (code 0), and the ENDBLK trailer
// in source order; entities are emitted at their boundary positions so nested
// blocks keep their exact structure.
func (b *Block) formatRaw(f format.Formatter) {
	f.WriteString(0, "BLOCK")
	headerEnd := 1
	for ; headerEnd < len(b.rawTags); headerEnd++ {
		t := b.rawTags[headerEnd]
		if t.Code == 0 && isBlockRecordTag(t.Value) {
			break
		}
		f.WriteString(t.Code, t.Value)
	}
	_ = headerEnd

	// Emit entities at their boundary positions. The raw stream contains the
	// exact source tags of this block record; a nested BLOCK's full record is
	// replaced by its parsed representation (recursively), and an ENDBLK tag
	// closes this block when all entities are emitted.
	entities := b.entities
	next := 0
	i := headerEnd
	for i < len(b.rawTags) {
		t := b.rawTags[i]
		if t.Code == 0 {
			end := rawTagsEnd(b.rawTags, i)
			if next < len(entities) && strings.EqualFold(t.Value, entities[next].DXFType()) {
				entities[next].Format(f)
				next++
				i = end
				continue
			}
			if strings.EqualFold(t.Value, "ENDBLK") && next >= len(entities) {
				break // trailer begins here
			}
			for j := i; j < end; j++ {
				f.WriteString(b.rawTags[j].Code, b.rawTags[j].Value)
			}
			i = end
			continue
		}
		f.WriteString(t.Code, t.Value)
		i++
	}

	// Emit any entities with no boundary in the raw stream.
	for ; next < len(entities); next++ {
		entities[next].Format(f)
	}

	// ENDBLK trailer: prefer the LAST top-level ENDBLK tag (the parent's own
	// terminator); nested blocks keep their own terminators inside the loop.
	// Scan from the end so a parent never steals a nested block's terminator.
	trailer := len(b.rawTags)
	depth := 0
	for j := i; j < len(b.rawTags); j++ {
		t := b.rawTags[j]
		if t.Code != 0 {
			continue
		}
		if strings.EqualFold(t.Value, "BLOCK") {
			depth++
		}
		if strings.EqualFold(t.Value, "ENDBLK") {
			if depth == 0 {
				trailer = j
			} else {
				depth--
			}
		}
	}
	for j := i; false && j < len(b.rawTags); j++ {
		t := b.rawTags[j]
		if t.Code != 0 {
			continue
		}
		if strings.EqualFold(t.Value, "BLOCK") {
			depth++
		}
		if strings.EqualFold(t.Value, "ENDBLK") {
			if depth == 0 {
				trailer = j
			} else {
				depth--
			}
		}
	}
	for ; trailer < len(b.rawTags); trailer++ {
		t := b.rawTags[trailer]
		f.WriteString(t.Code, t.Value)
	}
}

func formatRawLegacy(b *Block, f format.Formatter) {
	entities := b.entities
	next := 0
	i := 0
	for i < len(b.rawTags) {
		t := b.rawTags[i]
		if t.Code == 5 && next < len(entities) {
			f.WriteString(5, entities[next].Handle())
			next++
			i++
			continue
		}
		if t.Code == 0 {
			if next < len(entities) && strings.EqualFold(t.Value, entities[next].DXFType()) {
				// Entity boundary: emit the entity (nested blocks recursively).
				entities[next].Format(f)
				next++
				i = rawTagsEnd(b.rawTags, i)
				continue
			}
			if next < len(entities) && isBlockRecordTag(t.Value) {
				// Boundary desync: the next entity is a nested block whose
				// record tags are missing here; emit it at this boundary.
				if _, ok := entities[next].(*Block); ok {
					entities[next].Format(f)
					next++
					continue
				}
			}
			f.WriteString(t.Code, t.Value)
			i++
			continue
		}
		if t.Code == 5 {
			if b.handle != "" {
				f.WriteString(5, b.handle)
			}
			i++
			continue
		}
		f.WriteString(t.Code, t.Value)
		i++
	}
}

func isBlockRecordTag(v string) bool {
	return strings.EqualFold(v, "BLOCK") || strings.EqualFold(v, "ENDBLK")
}

// rawTagsEnd returns the index just after the record that starts at i:
// everything up to and including the matching ENDBLK (nested blocks balanced).
func rawTagsEnd(tags []blockTag, start int) int {
	startType := strings.ToUpper(tags[start].Value)
	if startType != "BLOCK" {
		// Ordinary entity record: ends at the next code-0 tag.
		for i := start + 1; i < len(tags); i++ {
			if tags[i].Code == 0 {
				return i
			}
		}
		return len(tags)
	}
	depth := 0
	for i := start + 1; i < len(tags); i++ {
		if tags[i].Code != 0 {
			continue
		}
		if strings.EqualFold(tags[i].Value, "BLOCK") {
			depth++
		}
		if strings.EqualFold(tags[i].Value, "ENDBLK") {
			if depth == 0 {
				return i + 1
			}
			depth--
		}
	}
	return len(tags)
}

type blockTag = TagPair

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

// EndHandle returns the ENDBLK handle.
func (b *Block) EndHandle() string { return b.endhandle }

// SetLayerName records the layer name until the layer table is resolved.
func (b *Block) SetLayerName(name string) { b.layerName = name }

// IsEntity is for BlockEntity: a nested block acts as an entity of its parent.
func (b *Block) IsEntity() bool { return true }

// DXFType returns the record type written at group code 0.
func (b *Block) DXFType() string { return "BLOCK" }

// RawTags returns the retained BLOCK/ENDBLK group codes.
func (b *Block) RawTags() []TagPair { return b.rawTags }

// SetRawTags stores the BLOCK/ENDBLK group codes of the definition.
func (b *Block) SetRawTags(tags []TagPair) { b.rawTags = tags }

// AddEntity appends an entity contained in the block.
func (b *Block) AddEntity(e BlockEntity) { b.entities = append(b.entities, e) }

// SetEntities replaces the entities contained in the block.
func (b *Block) SetEntities(es []BlockEntity) { b.entities = es }

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
