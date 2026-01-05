// OBJECTS section
package object

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// Objects represents OBJECTS section.
type Objects []Object

// New creates a new Objects.
func New() Objects {
	o := make([]Object, 0)
	return o
}

// Format writes OBJECTS data to formatter.
func (os Objects) Format(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "OBJECTS")
	for _, o := range os {
		o.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

// Add adds a new object to OBJECTS section.
func (os Objects) Add(o Object) Objects {
	os = append(os, o)
	return os
}

// SetHandle sets handles to each object.
func (os Objects) SetHandle(hg *handle.HandleGenerator) {
	for _, o := range os {
		o.SetHandle(hg)
	}
}
