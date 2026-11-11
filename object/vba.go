package object

import (
	"encoding/hex"
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// VBAProject is the typed model of a VBA_PROJECT object. The payload is a
// binary blob split across group-code 310 chunks; the declared byte count is
// carried by group code 90.
type VBAProject struct {
	*StdObject
	declaredSize int
	data         []byte
}

// Data returns a copy of the VBA project payload.
func (v *VBAProject) Data() []byte { return append([]byte(nil), v.data...) }

// DeclaredSize returns the byte count recorded in the source record (code 90).
func (v *VBAProject) DeclaredSize() int { return v.declaredSize }

// SetData replaces the payload and declared size.
func (v *VBAProject) SetData(data []byte) {
	v.data = append([]byte(nil), data...)
	v.declaredSize = len(v.data)
}

func newVBAProject(tags lldxf.Tags) *VBAProject {
	v := &VBAProject{StdObject: newStdObject("VBA_PROJECT", tags)}
	v.declaredSize = intByCode(tags, 90)
	for _, tag := range tags {
		if int(tag.Code()) != 310 {
			continue
		}
		chunk, err := hex.DecodeString(strings.TrimSpace(tagString(tag)))
		if err != nil {
			chunk = []byte(tagString(tag))
		}
		v.data = append(v.data, chunk...)
	}
	return v
}

// Format writes the object. Loaded records keep their full retained tag
// stream; a newly constructed project writes the header, byte count, and
// binary chunks.
func (v *VBAProject) Format(f format.Formatter) {
	if len(v.RetainedTags()) > 0 {
		v.StdObject.Format(f)
		return
	}
	wire := v.WireType()
	if wire == "" {
		wire = "VBA_PROJECT"
	}
	f.WriteString(0, wire)
	if v.Handle() != "" {
		f.WriteString(5, v.Handle())
	}
	if v.StdObject.owner != "" {
		f.WriteString(330, v.StdObject.owner)
	}
	f.WriteString(100, "AcDbVbaProject")
	f.WriteInt(90, len(v.data))
	remaining := v.data
	for len(remaining) > 0 {
		n := min(len(remaining), 127)
		f.WriteString(310, strings.ToUpper(hex.EncodeToString(remaining[:n])))
		remaining = remaining[n:]
	}
}
