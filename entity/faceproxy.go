package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

type FaceProxy struct {
	*entity
	OriginalType      string
	Flags             int
	Color             int
	LineWeight        int
	PlotStyleHandle   string
	MaterialHandle    string
	Transform         math.Matrix44
	ExtentsMin        math.Vec3
	ExtentsMax        math.Vec3
	ProxyBBox         math.BoundingBox
	AcisData          []string
	Entropy           float64
	EmbeddedObjectPtr int
}

func NewFaceProxy() *FaceProxy {
	return &FaceProxy{
		entity:            NewEntity(FACEPROXY),
		OriginalType:      "",
		Flags:             0,
		Color:             0,
		LineWeight:        -1,
		PlotStyleHandle:   "",
		MaterialHandle:    "",
		Transform:         math.NewMatrix44(),
		ExtentsMin:        math.NewVec3(0, 0, 0),
		ExtentsMax:        math.NewVec3(0, 0, 0),
		ProxyBBox:         math.BoundingBox{},
		AcisData:          make([]string, 0),
		Entropy:           0,
		EmbeddedObjectPtr: 0,
	}
}

func (f *FaceProxy) IsEntity() bool {
	return true
}

func (f *FaceProxy) Format(fm format.Formatter) {
	f.entity.Format(fm)
	fm.WriteString(100, "AcDbFaceProxy")
	fm.WriteString(2, f.OriginalType)
	fm.WriteInt(70, f.Flags)
	fm.WriteInt(62, f.Color)
	fm.WriteInt(370, f.LineWeight)
	if f.PlotStyleHandle != "" {
		fm.WriteString(390, f.PlotStyleHandle)
	}
	if f.MaterialHandle != "" {
		fm.WriteString(347, f.MaterialHandle)
	}
	for i := 0; i < 16; i++ {
		fm.WriteFloat(40+i, f.Transform[i])
	}
	fm.WriteFloat(10, f.ExtentsMin.X())
	fm.WriteFloat(20, f.ExtentsMin.Y())
	fm.WriteFloat(30, f.ExtentsMin.Z())
	fm.WriteFloat(11, f.ExtentsMax.X())
	fm.WriteFloat(21, f.ExtentsMax.Y())
	fm.WriteFloat(31, f.ExtentsMax.Z())
	if f.Entropy != 0 {
		fm.WriteFloat(90, f.Entropy*1000000)
	}
	fm.WriteInt(95, f.EmbeddedObjectPtr)
	if len(f.AcisData) > 0 {
		fm.WriteInt(93, len(f.AcisData))
		for _, line := range f.AcisData {
			fm.WriteString(1, line)
		}
	}
}

func (f *FaceProxy) BBox() ([]float64, []float64) {
	return []float64{f.ExtentsMin.X(), f.ExtentsMin.Y(), f.ExtentsMin.Z()},
		[]float64{f.ExtentsMax.X(), f.ExtentsMax.Y(), f.ExtentsMax.Z()}
}

func (f *FaceProxy) Copy() Entity {
	fp := NewFaceProxy()
	fp.entity = f.entity
	fp.OriginalType = f.OriginalType
	fp.Flags = f.Flags
	fp.Color = f.Color
	fp.LineWeight = f.LineWeight
	fp.PlotStyleHandle = f.PlotStyleHandle
	fp.MaterialHandle = f.MaterialHandle
	fp.Transform = f.Transform
	fp.ExtentsMin = f.ExtentsMin
	fp.ExtentsMax = f.ExtentsMax
	fp.ProxyBBox = f.ProxyBBox
	fp.AcisData = make([]string, len(f.AcisData))
	for i, line := range f.AcisData {
		fp.AcisData[i] = line
	}
	fp.Entropy = f.Entropy
	fp.EmbeddedObjectPtr = f.EmbeddedObjectPtr
	return fp
}

func (f *FaceProxy) Validate() error {
	return nil
}

func (f *FaceProxy) AddAcisLine(line string) {
	f.AcisData = append(f.AcisData, line)
}

func (f *FaceProxy) ClearAcisData() {
	f.AcisData = make([]string, 0)
}
