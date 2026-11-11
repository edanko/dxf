// Package blockadapter bridges entity.Entity values into block.BlockEntity.
//
// The adapter exists because block cannot import entity (entity imports block).
// Keeping it in one place avoids the two copies that previously lived in the
// root package and drawing package.
package blockadapter

import (
	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
)

// Adapter supplies the raw-tag methods for entities that lack them.
type Adapter struct {
	entity.Entity
	raw []block.TagPair
}

// Unwrap exposes the adapted entity for consumers that need the concrete type.
func (a *Adapter) Unwrap() entity.Entity { return a.Entity }

// RawTags returns the retained group codes.
func (a *Adapter) RawTags() []block.TagPair { return a.raw }

// SetRawTags stores retained group codes.
func (a *Adapter) SetRawTags(tags []block.TagPair) { a.raw = tags }

// LayerName forwards the recorded layer name of the adapted entity.
func (a *Adapter) LayerName() string {
	if namer, ok := a.Entity.(entity.LayerNamer); ok {
		return namer.LayerName()
	}
	return ""
}

// Adapt returns e as a block.BlockEntity, wrapping it only when necessary.
func Adapt(e entity.Entity) block.BlockEntity {
	if be, ok := e.(block.BlockEntity); ok {
		return be
	}
	return &Adapter{Entity: e}
}
