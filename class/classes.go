// CLASSES section
package class

import (
	"fmt"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

// Class represents each CLASS.
type Class struct {
	DXFName       string // 1
	CppName       string // 2
	AppName       string // 3
	Flag          int    // 90
	InstanceCount int    // 91
	WasProxy      int    // 280
	IsEntity      int    // 281
}

// Format writes data to formatter.
func (c *Class) Format(f format.Formatter) {
	f.WriteString(0, "CLASS")
	f.WriteString(1, c.DXFName)
	f.WriteString(2, c.CppName)
	f.WriteString(3, c.AppName)
	f.WriteInt(90, c.Flag)
	f.WriteInt(91, c.InstanceCount)
	f.WriteInt(280, c.WasProxy)
	f.WriteInt(281, c.IsEntity)
}

// Classes represents CLASSES section.
type Classes []*Class

// New creates a new Classes.
func New() Classes {
	c := make([]*Class, 0)
	return c
}

// Parse parses a CLASSES section, retaining every class record.
func Parse(tags lldxf.Tags) Classes {
	classes := make(Classes, 0)
	var current *Class
	for _, tag := range tags {
		if tag.Code() == 0 {
			if current != nil {
				classes = append(classes, current)
			}
			if tagString(tag) == "CLASS" {
				current = &Class{}
			} else {
				current = nil
			}
			continue
		}
		if current == nil {
			continue
		}
		switch int(tag.Code()) {
		case 1:
			current.DXFName = tagString(tag)
		case 2:
			current.CppName = tagString(tag)
		case 3:
			current.AppName = tagString(tag)
		case 90:
			current.Flag = tagInt(tag)
		case 91:
			current.InstanceCount = tagInt(tag)
		case 280:
			current.WasProxy = tagInt(tag)
		case 281:
			current.IsEntity = tagInt(tag)
		}
	}
	if current != nil {
		classes = append(classes, current)
	}
	return classes
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

func tagString(tag lldxf.Tag) string {
	return formatValue(tag.Value())
}

func tagInt(tag lldxf.Tag) int {
	switch v := tag.Value().(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		var n int
		for _, r := range tagString(tag) {
			if r < '0' || r > '9' {
				break
			}
			n = n*10 + int(r-'0')
		}
		return n
	}
}

func formatValue(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
