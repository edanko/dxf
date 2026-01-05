package entity

import (
	"encoding/hex"
	"strings"

	"github.com/edanko/dxf/format"
)

type VBAProject struct {
	*entity
	Data []byte
}

func NewVBAProject() *VBAProject {
	return &VBAProject{
		entity: NewEntity(VBAPROJECT),
		Data:   make([]byte, 0),
	}
}

func (v *VBAProject) IsEntity() bool {
	return true
}

func (v *VBAProject) Format(f format.Formatter) {
	v.entity.Format(f)
	f.WriteString(100, "AcDbVbaProject")
	f.WriteInt(90, len(v.Data))
	if len(v.Data) > 0 {
		v.writeBinaryData(f, 310, v.Data)
	}
}

func (v *VBAProject) writeBinaryData(f format.Formatter, code int, data []byte) {
	hexStr := strings.ToUpper(hex.EncodeToString(data))
	for len(hexStr) > 0 {
		chunk := hexStr
		if len(hexStr) > 254 {
			chunk = hexStr[:254]
			hexStr = hexStr[254:]
		} else {
			hexStr = ""
		}
		f.WriteString(code, chunk)
	}
}

func (v *VBAProject) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (v *VBAProject) Copy() Entity {
	proj := NewVBAProject()
	proj.entity = v.entity
	proj.Data = make([]byte, len(v.Data))
	for i, b := range v.Data {
		proj.Data[i] = b
	}
	return proj
}

func (v *VBAProject) Validate() error {
	return nil
}

func (v *VBAProject) SetData(data []byte) {
	v.Data = make([]byte, len(data))
	for i, b := range data {
		v.Data[i] = b
	}
}

func (v *VBAProject) GetData() []byte {
	result := make([]byte, len(v.Data))
	for i, b := range v.Data {
		result[i] = b
	}
	return result
}

func (v *VBAProject) Clear() {
	v.Data = make([]byte, 0)
}
