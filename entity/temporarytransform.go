package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

type TemporaryTransform struct {
	*entity
	TransformationMatrix math.Matrix44
	OperationType        int
	AffectedEntity       string
}

func NewTemporaryTransform() *TemporaryTransform {
	return &TemporaryTransform{
		entity:               NewEntity(TEMPORARYTRANSFORM),
		TransformationMatrix: math.NewMatrix44(),
		OperationType:        0,
		AffectedEntity:       "",
	}
}

func (t *TemporaryTransform) IsEntity() bool {
	return true
}

func (t *TemporaryTransform) Format(f format.Formatter) {
	t.entity.Format(f)
	f.WriteString(100, "AcDbTemporaryTransform")
	for i := 0; i < 16; i++ {
		f.WriteFloat(40+i, t.TransformationMatrix[i])
	}
	f.WriteInt(70, t.OperationType)
	if t.AffectedEntity != "" {
		f.WriteString(330, t.AffectedEntity)
	}
}

func (t *TemporaryTransform) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (t *TemporaryTransform) Copy() Entity {
	transform := NewTemporaryTransform()
	transform.entity = t.entity
	transform.TransformationMatrix = t.TransformationMatrix
	transform.OperationType = t.OperationType
	transform.AffectedEntity = t.AffectedEntity
	return transform
}

func (t *TemporaryTransform) Validate() error {
	return nil
}

func (t *TemporaryTransform) SetMatrix(m math.Matrix44) {
	t.TransformationMatrix = m
}

func (t *TemporaryTransform) GetMatrix() math.Matrix44 {
	return t.TransformationMatrix
}
