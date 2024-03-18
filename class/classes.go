// CLASSES section
package class

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// Class represents each CLASS.
type Class struct {
}

// Format writes data to formatter.
func (c *Class) Format(f format.Formatter) {
	f.WriteString(0, "CLASS")
}

// Classes represents CLASSES section.
type Classes []*Class

// New creates a new Classes.
func New() Classes {
	c := make([]*Class, 0)
	return c
}

// Format writes CLASSES data to formatter.
func (cs Classes) Format(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "CLASSES")
	for _, c := range cs {
		c.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

// SetHandle sets handles to each class.
func (cs Classes) SetHandle(_ *handle.HandleGenerator) {}
