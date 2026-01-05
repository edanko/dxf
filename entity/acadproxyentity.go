package entity

import (
	"encoding/hex"
	"strings"

	"github.com/edanko/dxf/format"
)

type AcadProxyEntity struct {
	*entity
	ClassID          int
	ApplicationID    int
	GraphicsDataSize int
	GraphicsData     []byte
	UnknownDataSize  int
	UnknownData      []byte
	EntityDataSize   int
	EntityData       []byte
	ObjectIDs        []string
	FormatFlag       int
	OriginalFormat   int
	HasBBox          bool
	BBoxMinX         float64
	BBoxMinY         float64
	BBoxMaxX         float64
	BBoxMaxY         float64
}

func NewAcadProxyEntity() *AcadProxyEntity {
	return &AcadProxyEntity{
		entity:           NewEntity(ACADPROXYENTITY),
		ClassID:          498,
		ApplicationID:    0,
		GraphicsDataSize: 0,
		GraphicsData:     make([]byte, 0),
		UnknownDataSize:  0,
		UnknownData:      make([]byte, 0),
		EntityDataSize:   0,
		EntityData:       make([]byte, 0),
		ObjectIDs:        make([]string, 0),
		FormatFlag:       0,
		OriginalFormat:   0,
		HasBBox:          false,
		BBoxMinX:         0,
		BBoxMinY:         0,
		BBoxMaxX:         0,
		BBoxMaxY:         0,
	}
}

func (a *AcadProxyEntity) IsEntity() bool {
	return true
}

func (a *AcadProxyEntity) Format(f format.Formatter) {
	a.entity.Format(f)
	f.WriteString(100, "AcDbProxyEntity")
	f.WriteInt(90, a.ClassID)
	f.WriteInt(91, a.ApplicationID)

	if len(a.GraphicsData) > 0 {
		f.WriteInt(92, len(a.GraphicsData))
		a.writeBinaryData(f, 310, a.GraphicsData)
	}

	if len(a.UnknownData) > 0 {
		f.WriteInt(96, len(a.UnknownData))
		a.writeBinaryData(f, 311, a.UnknownData)
	}

	if len(a.EntityData) > 0 {
		var bitSize int
		for _, b := range a.EntityData {
			for i := 7; i >= 0; i-- {
				if (b>>i)&1 == 1 {
					bitSize++
				}
			}
		}
		f.WriteInt(93, bitSize)
		a.writeBinaryData(f, 310, a.EntityData)
	}

	for _, id := range a.ObjectIDs {
		f.WriteString(330, id)
	}
	f.WriteInt(94, 0)
	f.WriteInt(95, a.FormatFlag)
	f.WriteInt(70, a.OriginalFormat)
}

func (a *AcadProxyEntity) writeBinaryData(f format.Formatter, code int, data []byte) {
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

func (a *AcadProxyEntity) BBox() ([]float64, []float64) {
	if a.HasBBox {
		return []float64{a.BBoxMinX, a.BBoxMinY, 0}, []float64{a.BBoxMaxX, a.BBoxMaxY, 0}
	}
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (a *AcadProxyEntity) Copy() Entity {
	proxy := NewAcadProxyEntity()
	proxy.entity = a.entity
	proxy.ClassID = a.ClassID
	proxy.ApplicationID = a.ApplicationID
	proxy.GraphicsDataSize = a.GraphicsDataSize
	proxy.GraphicsData = make([]byte, len(a.GraphicsData))
	for i, b := range a.GraphicsData {
		proxy.GraphicsData[i] = b
	}
	proxy.UnknownDataSize = a.UnknownDataSize
	proxy.UnknownData = make([]byte, len(a.UnknownData))
	for i, b := range a.UnknownData {
		proxy.UnknownData[i] = b
	}
	proxy.EntityDataSize = a.EntityDataSize
	proxy.EntityData = make([]byte, len(a.EntityData))
	for i, b := range a.EntityData {
		proxy.EntityData[i] = b
	}
	proxy.ObjectIDs = make([]string, len(a.ObjectIDs))
	for i, id := range a.ObjectIDs {
		proxy.ObjectIDs[i] = id
	}
	proxy.FormatFlag = a.FormatFlag
	proxy.OriginalFormat = a.OriginalFormat
	proxy.HasBBox = a.HasBBox
	proxy.BBoxMinX = a.BBoxMinX
	proxy.BBoxMinY = a.BBoxMinY
	proxy.BBoxMaxX = a.BBoxMaxX
	proxy.BBoxMaxY = a.BBoxMaxY
	return proxy
}

func (a *AcadProxyEntity) Validate() error {
	return nil
}

func (a *AcadProxyEntity) SetGraphicsData(data []byte) {
	a.GraphicsData = make([]byte, len(data))
	for i, b := range data {
		a.GraphicsData[i] = b
	}
}

func (a *AcadProxyEntity) SetEntityData(data []byte) {
	a.EntityData = make([]byte, len(data))
	for i, b := range data {
		a.EntityData[i] = b
	}
}

func (a *AcadProxyEntity) AddObjectID(id string) {
	a.ObjectIDs = append(a.ObjectIDs, id)
}

func (a *AcadProxyEntity) ClearObjectIDs() {
	a.ObjectIDs = make([]string, 0)
}

func (a *AcadProxyEntity) SetBoundingBox(minX, minY, maxX, maxY float64) {
	a.HasBBox = true
	a.BBoxMinX = minX
	a.BBoxMinY = minY
	a.BBoxMaxX = maxX
	a.BBoxMaxY = maxY
}
