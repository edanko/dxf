package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

type Field struct {
	*entity
	EvaluatorID       string
	FieldCode         string
	FieldCodeOverflow string
	NChildFields      int
	DataSets          []FieldDataSet
	FormatString      string
	FormatOverflow    string
}

type FieldDataSet struct {
	Key         string
	ValueKey    string
	DataType    int
	LongValue   int
	DoubleValue float64
	IDValue     string
	BinaryData  []byte
}

func NewField() *Field {
	return &Field{
		entity:            NewEntity(0),
		EvaluatorID:       "",
		FieldCode:         "",
		FieldCodeOverflow: "",
		NChildFields:      0,
		DataSets:          make([]FieldDataSet, 0),
		FormatString:      "",
		FormatOverflow:    "",
	}
}

func (f *Field) IsEntity() bool {
	return true
}

func (f *Field) Format(fr format.Formatter) {
	fr.WriteString(0, "FIELD")
	fr.WriteString(5, f.Handle())
	fr.WriteString(100, "AcDbField")
	fr.WriteString(1, f.EvaluatorID)
	fr.WriteString(2, f.FieldCode)
	if f.FieldCodeOverflow != "" {
		fr.WriteString(3, f.FieldCodeOverflow)
	}
	fr.WriteInt(90, f.NChildFields)
	for _, ds := range f.DataSets {
		fr.WriteInt(93, 1)
		fr.WriteString(6, ds.Key)
		fr.WriteString(7, ds.ValueKey)
		fr.WriteInt(90, ds.DataType)
		switch ds.DataType {
		case 1:
			fr.WriteInt(91, ds.LongValue)
		case 2:
			fr.WriteFloat(140, ds.DoubleValue)
		case 3:
			if ds.IDValue != "" {
				fr.WriteString(330, ds.IDValue)
			}
		case 4:
			fr.WriteInt(92, len(ds.BinaryData))
			fr.WriteString(310, "")
		}
	}
	fr.WriteString(301, f.FormatString)
	if f.FormatOverflow != "" {
		fr.WriteString(9, f.FormatOverflow)
	}
}

func (f *Field) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (f *Field) Copy() Entity {
	field := NewField()
	field.entity = f.entity
	field.EvaluatorID = f.EvaluatorID
	field.FieldCode = f.FieldCode
	field.FieldCodeOverflow = f.FieldCodeOverflow
	field.NChildFields = f.NChildFields
	field.DataSets = make([]FieldDataSet, len(f.DataSets))
	for i, ds := range f.DataSets {
		field.DataSets[i] = ds
	}
	field.FormatString = f.FormatString
	field.FormatOverflow = f.FormatOverflow
	return field
}

func (f *Field) Validate() error {
	return nil
}

func (f *Field) AddDataSet(ds FieldDataSet) {
	f.DataSets = append(f.DataSets, ds)
}

func (f *Field) ClearDataSets() {
	f.DataSets = make([]FieldDataSet, 0)
}

func (f *Field) Handle() string {
	if f.entity != nil {
		return f.entity.Handle()
	}
	return ""
}

func (f *Field) SetHandle(hg *handle.HandleGenerator) {
	if f.entity != nil {
		f.entity.SetHandle(hg)
	}
}
