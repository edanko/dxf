package object

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// BaseObject is the base struct for DXF objects
type BaseObject struct {
	handle string
}

func (b *BaseObject) Handle() string {
	return b.handle
}

func (b *BaseObject) SetHandle(hg *handle.HandleGenerator) {
	b.handle = hg.Next()
}

// Object is the base interface for DXF objects
type Object interface {
	IsObject() bool
	Format(format.Formatter)
	handle.Handler
}

// objectObject implements the Object interface
type objectObject struct {
	handle string
}

// IsObject returns true for object objects
func (o *objectObject) IsObject() bool {
	return true
}

// Handle returns object handle
func (o *objectObject) Handle() string {
	return o.handle
}

// SetHandle sets object handle
func (o *objectObject) SetHandle(hg *handle.HandleGenerator) {
	o.handle = hg.Next()
}

// Format writes basic object data to formatter
func (o *objectObject) Format(f format.Formatter) {
	f.WriteString(0, "OBJECT")
	f.WriteString(5, o.handle)
}
