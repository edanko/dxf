package header

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/insunit"
	"github.com/edanko/dxf/lldxf"
)

var typedHeaderVariables = map[string]bool{
	"$ACADVER":  true,
	"$INSBASE":  true,
	"$INSUNITS": true,
	"$LUNITS":   true,
	"$EXTMIN":   true,
	"$EXTMAX":   true,
	"$LTSCALE":  true,
	"$HANDSEED": true,
}

// Parse parses a HEADER section into a Header, retaining variables that do not
// have a typed field in Extra.
func Parse(tags lldxf.Tags) *Header {
	h := New()
	var currentName string
	var currentTags lldxf.Tags

	flush := func() {
		if currentName == "" {
			return
		}
		if !typedHeaderVariables[currentName] {
			h.Extra = append(h.Extra, Variable{Name: currentName, Tags: currentTags})
		}
		currentName = ""
		currentTags = nil
	}

	for _, tag := range tags {
		if tag.Code() == 9 {
			flush()
			currentName = tagString(tag)
			continue
		}
		if currentName == "" {
			continue
		}
		currentTags = append(currentTags, tag)
	}
	flush()

	return h
}

// tagString formats a tag value as a string.
func tagString(tag lldxf.Tag) string {
	return fmt.Sprintf("%v", tag.Value())
}

// ParseTyped fills the typed fields of h from a HEADER tag stream.
func ParseTyped(tags lldxf.Tags) *Header {
	h := Parse(tags)
	var name string
	var values lldxf.Tags
	flush := func() {
		switch name {
		case "$ACADVER":
			if v := firstString(values); v != "" {
				h.Version = v
			}
		case "$INSBASE":
			h.InsBase = coords(values, 10, 3)
		case "$INSUNITS":
			if v, ok := firstInt(values); ok {
				h.InsUnit = insunit.Unit(v)
			}
		case "$LUNITS":
			if v, ok := firstInt(values); ok {
				h.InsLUnit = insunit.Type(v - 2)
			}
		case "$EXTMIN":
			h.ExtMin = coords(values, 10, 3)
		case "$EXTMAX":
			h.ExtMax = coords(values, 10, 3)
		case "$LTSCALE":
			if v, ok := firstFloat(values); ok {
				h.LtScale = v
			}
		case "$HANDSEED":
			if v := firstString(values); v != "" {
				h.handseed = v
			}
		}
		name = ""
		values = nil
	}
	for _, tag := range tags {
		if tag.Code() == 9 {
			flush()
			name = tagString(tag)
			continue
		}
		values = append(values, tag)
	}
	flush()
	return h
}

func firstString(tags lldxf.Tags) string {
	for _, tag := range tags {
		return tagString(tag)
	}
	return ""
}

func firstInt(tags lldxf.Tags) (int, bool) {
	for _, tag := range tags {
		switch v := tag.Value().(type) {
		case int:
			return v, true
		case int64:
			return int(v), true
		case float64:
			return int(v), true
		default:
			var n int
			if _, err := fmt.Sscanf(tagString(tag), "%d", &n); err == nil {
				return n, true
			}
		}
	}
	return 0, false
}

func firstFloat(tags lldxf.Tags) (float64, bool) {
	for _, tag := range tags {
		switch v := tag.Value().(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		case int64:
			return float64(v), true
		default:
			var f float64
			if _, err := fmt.Sscanf(tagString(tag), "%f", &f); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}

func coords(tags lldxf.Tags, base, count int) []float64 {
	result := make([]float64, count)
	for i := 0; i < count; i++ {
		for _, tag := range tags {
			if int(tag.Code()) == base+i*10 {
				f, _ := firstFloat(lldxf.Tags{tag})
				result[i] = f
				break
			}
		}
	}
	return result
}

// VariableNames returns the names of retained extra variables.
func (h *Header) VariableNames() []string {
	names := make([]string, 0, len(h.Extra))
	for _, v := range h.Extra {
		names = append(names, strings.ToUpper(v.Name))
	}
	return names
}

// Variable returns a retained variable by name.
func (h *Header) Variable(name string) (Variable, bool) {
	for _, v := range h.Extra {
		if strings.EqualFold(v.Name, name) {
			return v, true
		}
	}
	return Variable{}, false
}
