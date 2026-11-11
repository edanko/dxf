package entity

import (
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// LwPolyline represents LWPOLYLINE Entity.
type LwPolyline struct {
	*entity
	Num           int     // 90
	Closed        bool    // 70 bit 1
	ConstantWidth float64 // 43
	Vertices      [][]float64
	Bulges        []float64
}

// IsEntity is for Entity interface.
func (l *LwPolyline) IsEntity() bool {
	return true
}

// NewLwPolyline creates a new LwPolyline.
func NewLwPolyline(size int) *LwPolyline {
	vs := make([][]float64, size)
	for i := range size {
		vs[i] = make([]float64, 2)
	}
	l := &LwPolyline{
		entity:   NewEntity(LWPOLYLINE),
		Num:      size,
		Closed:   false,
		Vertices: vs,
		Bulges:   make([]float64, size),
	}
	return l
}

// Format writes data to formatter.
func (l *LwPolyline) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbPolyline")
	f.WriteInt(90, l.Num)
	flag := 0
	if l.Closed {
		flag |= 1
	}
	f.WriteInt(70, flag)
	if l.ConstantWidth != 0 {
		f.WriteFloat(43, l.ConstantWidth)
	}
	for i := 0; i < l.Num; i++ {
		for j := range 2 {
			f.WriteFloat((j+1)*10, l.Vertices[i][j])
		}
		if l.Bulges[i] != 0 {
			f.WriteFloat(42, l.Bulges[i])
		}
	}
}

// Close closes LwPolyline.
func (l *LwPolyline) Close() {
	l.Closed = true
}

func (l *LwPolyline) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for _, p := range l.Vertices {
		for i := range p {
			if p[i] < mins[i] {
				mins[i] = p[i]
			}
			if p[i] > maxs[i] {
				maxs[i] = p[i]
			}
		}
	}
	return mins, maxs
}

// Transform applies an affine matrix to all lightweight polyline vertices.
func (l *LwPolyline) Transform(m dxfmath.Matrix44) error {
	for i := range l.Vertices {
		l.Vertices[i] = TransformPoint(m, l.Vertices[i])
	}
	return nil
}

// Copy creates a deep copy of the LwPolyline entity
func (l *LwPolyline) Copy() Entity {
	lwpolyline := NewLwPolyline(l.Num)
	lwpolyline.entity = l.entity.clone()
	lwpolyline.Closed = l.Closed
	lwpolyline.ConstantWidth = l.ConstantWidth
	lwpolyline.Bulges = make([]float64, len(l.Bulges))
	copy(lwpolyline.Bulges, l.Bulges)
	for i, v := range l.Vertices {
		lwpolyline.Vertices[i] = []float64{v[0], v[1]}
	}
	return lwpolyline
}
