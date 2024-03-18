package entity

type Solid struct {
	*entity
	FirstPoint          []float64 //10, 20, 30
	SecondPoint         []float64 //11, 21, 31
	ThirdPoint          []float64 //12 ,22, 32
	FourthPoint         []float64 // 13,23, 33
	Thickness           float64   //0, 0, 1
	StretchingDirection []float64 // 210,220, 230
}

func NewSolid() *Solid {
	return &Solid{
		entity:              NewEntity(SOLID),
		FirstPoint:          []float64{0.0, 0.0, 0.0},
		SecondPoint:         []float64{0.0, 0.0, 0.0},
		ThirdPoint:          []float64{0.0, 0.0, 0.0},
		FourthPoint:         []float64{0.0, 0.0, 0.0},
		Thickness:           0.0,
		StretchingDirection: []float64{0.0, 0.0, 0.0},
	}
}

func (s Solid) IsEntity() bool {
	return true
}

func (s Solid) BBox() ([]float64, []float64) {
	panic("implement me")
}
