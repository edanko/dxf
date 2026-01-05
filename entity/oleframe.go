package entity

import (
	"encoding/hex"
	"strings"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

type OLE2Frame struct {
	*entity
	BinaryData   []byte
	MinPoint     dxfmath.Vec3
	MaxPoint     dxfmath.Vec3
	OLEClass     string
	ItemName     string
	Flag         int
	Mode         int
	Rotation     float64
	Width        float64
	Height       float64
	UploadTime   float64
	MonikerBytes []byte
}

func NewOLE2Frame() *OLE2Frame {
	return &OLE2Frame{
		entity:       NewEntity(OLE2FRAME),
		BinaryData:   make([]byte, 0),
		MinPoint:     dxfmath.NewVec3(0, 0, 0),
		MaxPoint:     dxfmath.NewVec3(1, 1, 0),
		OLEClass:     "",
		ItemName:     "",
		Flag:         0,
		Mode:         0,
		Rotation:     0,
		Width:        1,
		Height:       1,
		UploadTime:   0,
		MonikerBytes: make([]byte, 0),
	}
}

func (o *OLE2Frame) IsEntity() bool {
	return true
}

func (o *OLE2Frame) Format(f format.Formatter) {
	o.entity.Format(f)
	f.WriteString(100, "AcDbOle2Frame")
	f.WriteString(1, o.OLEClass)
	f.WriteString(2, o.ItemName)
	f.WriteInt(70, o.Flag)
	f.WriteInt(71, o.Mode)
	f.WriteFloat(90, o.UploadTime)
	f.WriteFloat(10, o.MinPoint.X())
	f.WriteFloat(20, o.MinPoint.Y())
	f.WriteFloat(30, o.MinPoint.Z())
	f.WriteFloat(11, o.MaxPoint.X())
	f.WriteFloat(21, o.MaxPoint.Y())
	f.WriteFloat(31, o.MaxPoint.Z())
	f.WriteFloat(40, o.Width)
	f.WriteFloat(41, o.Height)
	f.WriteFloat(42, o.Rotation)
	if len(o.MonikerBytes) > 0 {
		o.writeBinaryData(f, 310, o.MonikerBytes)
	}
	if len(o.BinaryData) > 0 {
		o.writeBinaryData(f, 310, o.BinaryData)
	}
}

func (o *OLE2Frame) writeBinaryData(f format.Formatter, code int, data []byte) {
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

func (o *OLE2Frame) BBox() ([]float64, []float64) {
	return []float64{o.MinPoint.X(), o.MinPoint.Y(), o.MinPoint.Z()}, []float64{o.MaxPoint.X(), o.MaxPoint.Y(), o.MaxPoint.Z()}
}

func (o *OLE2Frame) Copy() Entity {
	ole := NewOLE2Frame()
	ole.entity = o.entity
	ole.BinaryData = make([]byte, len(o.BinaryData))
	for i, b := range o.BinaryData {
		ole.BinaryData[i] = b
	}
	ole.MinPoint = o.MinPoint
	ole.MaxPoint = o.MaxPoint
	ole.OLEClass = o.OLEClass
	ole.ItemName = o.ItemName
	ole.Flag = o.Flag
	ole.Mode = o.Mode
	ole.Rotation = o.Rotation
	ole.Width = o.Width
	ole.Height = o.Height
	ole.UploadTime = o.UploadTime
	ole.MonikerBytes = make([]byte, len(o.MonikerBytes))
	for i, b := range o.MonikerBytes {
		ole.MonikerBytes[i] = b
	}
	return ole
}

func (o *OLE2Frame) Validate() error {
	return nil
}

func (o *OLE2Frame) SetBinaryData(data []byte) {
	o.BinaryData = make([]byte, len(data))
	for i, b := range data {
		o.BinaryData[i] = b
	}
}

func (o *OLE2Frame) GetBinaryData() []byte {
	result := make([]byte, len(o.BinaryData))
	for i, b := range o.BinaryData {
		result[i] = b
	}
	return result
}
