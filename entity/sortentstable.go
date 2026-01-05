package entity

import (
	"github.com/edanko/dxf/format"
)

type SortEntsTable struct {
	*entity
	BlockRecordHandle string
	Table             map[string]string
}

func NewSortEntsTable() *SortEntsTable {
	return &SortEntsTable{
		entity:            NewEntity(SORTENTSTABLE),
		BlockRecordHandle: "",
		Table:             make(map[string]string),
	}
}

func (s *SortEntsTable) IsEntity() bool {
	return true
}

func (s *SortEntsTable) Format(f format.Formatter) {
	s.entity.Format(f)
	f.WriteString(100, "AcDbSortentsTable")
	f.WriteString(330, s.BlockRecordHandle)
	for handle, sortHandle := range s.Table {
		f.WriteString(331, handle)
		f.WriteString(5, sortHandle)
	}
}

func (s *SortEntsTable) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (s *SortEntsTable) Copy() Entity {
	tbl := NewSortEntsTable()
	tbl.entity = s.entity
	tbl.BlockRecordHandle = s.BlockRecordHandle
	tbl.Table = make(map[string]string)
	for k, v := range s.Table {
		tbl.Table[k] = v
	}
	return tbl
}

func (s *SortEntsTable) Validate() error {
	return nil
}

func (s *SortEntsTable) Len() int {
	return len(s.Table)
}

func (s *SortEntsTable) Append(handle, sortHandle string) {
	s.Table[handle] = sortHandle
}

func (s *SortEntsTable) Clear() {
	s.Table = make(map[string]string)
}

func (s *SortEntsTable) SetHandles(handles map[string]string) {
	s.Table = make(map[string]string)
	for k, v := range handles {
		s.Table[k] = v
	}
}

func (s *SortEntsTable) RemoveHandle(handle string) {
	delete(s.Table, handle)
}

func (s *SortEntsTable) GetSortHandle(handle string) (string, bool) {
	sortHandle, ok := s.Table[handle]
	return sortHandle, ok
}
