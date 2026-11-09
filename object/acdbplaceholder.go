package object

import (
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

// AcDbPlaceHolder represents ACDBPLACEHOLDER Object.
type AcDbPlaceHolder struct {
	handle    string
	owner     handle.Handler
	reactors  []string // parsed {ACAD_REACTORS links; Format re-emits the group
	groupTags int      // number of reactor-group tags in the source record
	retained  lldxf.Tags
}

// SetReactorHandles stores the parsed reactor links and their source tag span.
func (p *AcDbPlaceHolder) SetReactorHandles(hs []string, groupTags int) {
	p.reactors = hs
	p.groupTags = groupTags
}

// SetRetainedTags records tags the typed model does not emit.
func (p *AcDbPlaceHolder) SetRetainedTags(tags lldxf.Tags) { p.retained = tags }

// RetainedTags returns the retained tag stream.
func (p *AcDbPlaceHolder) RetainedTags() lldxf.Tags { return p.retained }

// IsObject is for Object interface.
func (p *AcDbPlaceHolder) IsObject() bool {
	return true
}

// NewAcDbPlaceHolder creates a new AcDbPlaceHolder.
func NewAcDbPlaceHolder() *AcDbPlaceHolder {
	p := &AcDbPlaceHolder{
		owner: nil,
	}
	return p
}

// Format writes data to formatter. The owner link is emitted as a reactor
// group plus the standalone 330, matching AutoCAD writers.
func (p *AcDbPlaceHolder) Format(f format.Formatter) {
	f.WriteString(0, "ACDBPLACEHOLDER")
	f.WriteString(5, p.handle)
	if len(p.reactors) > 0 {
		f.WriteString(102, "{ACAD_REACTORS")
		for _, r := range p.reactors {
			f.WriteString(330, r)
		}
		f.WriteString(102, "}")
	}
	if p.owner != nil {
		f.WriteString(330, p.owner.Handle())
	}
	for _, tag := range p.retained {
		f.WriteString(int(tag.Code()), lldxf.FormatValue(tag))
	}
}

// RetainUnhandled keeps tags the typed model does not emit. The owner link
// (reactor group and standalone 330) is emitted by Format, so reactor groups
// and code-330 tags are dropped; everything else is kept verbatim.
func (p *AcDbPlaceHolder) RetainUnhandled(tags lldxf.Tags, emitted map[int]bool) {
	raw := make(lldxf.Tags, 0)
	inReactor := false
	consumedTags := 0
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 102 {
			v := strings.ToUpper(strings.TrimSpace(lldxf.FormatValue(tag)))
			if v == "{ACAD_REACTORS" {
				inReactor = true
				if len(p.reactors) == 0 {
					raw = append(raw, tag)
				}
				continue
			}
			if v == "}" && inReactor {
				inReactor = false
				if len(p.reactors) == 0 {
					raw = append(raw, tag)
				}
				continue
			}
			if !inReactor && len(p.reactors) == 0 {
				raw = append(raw, tag)
			}
			continue
		}
		if code == 0 || code == 5 {
			continue
		}
		if inReactor {
			if len(p.reactors) == 0 {
				raw = append(raw, tag)
			}
			continue
		}
		if consumedTags < p.groupTags && code == 330 {
			consumedTags++ // a reactor member re-emitted as part of the group
			continue
		}
		if code == 330 {
			continue // owner link emitted from the model's own field
		}
		if emitted[code] {
			continue
		}
		raw = append(raw, tag)
	}
	p.retained = raw
}

// Handle returns a handle value.
func (p *AcDbPlaceHolder) Handle() string {
	return p.handle
}

// SetHandle sets a handle.
func (p *AcDbPlaceHolder) SetHandle(hg *handle.HandleGenerator) {
	if p.handle == "" {
		p.handle = hg.Next()
	}
}

// SetHandleString sets an explicit handle loaded from a file.
func (p *AcDbPlaceHolder) SetHandleString(h string) { p.handle = h }

// SetOwnerHandle sets an explicit owner handle loaded from a file.
func (p *AcDbPlaceHolder) SetOwnerHandle(h handle.Handler) { p.owner = h }
