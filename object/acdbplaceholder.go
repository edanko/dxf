package object

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// AcDbPlaceHolder represents ACDBPLACEHOLDER Object.
type AcDbPlaceHolder struct {
	handle string
	owner  handle.Handler
}

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

// Format writes data to formatter.
func (p *AcDbPlaceHolder) Format(f format.Formatter) {
	f.WriteString(0, "ACDBPLACEHOLDER")
	f.WriteString(5, p.handle)
	if p.owner != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, p.owner.Handle())
		f.WriteString(102, "}")
		f.WriteString(330, p.owner.Handle())
	}
}

// Handle returns a handle value.
func (p *AcDbPlaceHolder) Handle() string {
	return p.handle
}

// SetHandle sets a handle.
func (p *AcDbPlaceHolder) SetHandle(hg *handle.HandleGenerator) {
	p.handle = hg.Next()
}
