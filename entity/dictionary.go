package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// DictionaryEntry represents a key-value pair in a dictionary
type DictionaryEntry struct {
	Name   string // 3 - entry name
	Handle string // 350 - entry handle (soft-owner)
}

// Dictionary represents a DICTIONARY entity for storing named references to other entities
type Dictionary struct {
	*entity
	HardOwned int               // 280 - hard-owner flag
	Cloning   int               // 281 - cloning flag
	Entries   []DictionaryEntry // dictionary entries
}

// NewDictionary creates a new Dictionary entity
func NewDictionary() *Dictionary {
	d := &Dictionary{
		entity:    NewEntity(DICTIONARY),
		HardOwned: 0,
		Cloning:   1,
		Entries:   []DictionaryEntry{},
	}
	return d
}

// IsEntity is for Entity interface.
func (d *Dictionary) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (d *Dictionary) Format(f format.Formatter) {
	d.entity.Format(f)
	f.WriteString(100, "AcDbDictionary")
	f.WriteInt(280, d.HardOwned)
	f.WriteInt(281, d.Cloning)

	for _, entry := range d.Entries {
		f.WriteString(3, entry.Name)
		f.WriteString(350, entry.Handle)
	}
}

// BBox returns the bounding box of the entity
func (d *Dictionary) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the Dictionary
func (d *Dictionary) Transform(m *math.Matrix44) error {
	return nil
}

// Copy creates a deep copy of the Dictionary entity
func (d *Dictionary) Copy() Entity {
	dict := NewDictionary()
	dict.entity = d.entity
	dict.HardOwned = d.HardOwned
	dict.Cloning = d.Cloning
	dict.Entries = append(dict.Entries, d.Entries...)
	return dict
}

// Validate validates the Dictionary entity
func (d *Dictionary) Validate() error {
	if d.HardOwned < 0 || d.HardOwned > 1 {
		d.HardOwned = 0
	}
	if d.Cloning < 0 || d.Cloning > 5 {
		d.Cloning = 1
	}
	return nil
}

// AddEntry adds an entry to the dictionary
func (d *Dictionary) AddEntry(name, handle string) {
	d.Entries = append(d.Entries, DictionaryEntry{Name: name, Handle: handle})
}

// RemoveEntry removes an entry by name
func (d *Dictionary) RemoveEntry(name string) {
	for i, entry := range d.Entries {
		if entry.Name == name {
			d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
			return
		}
	}
}

// GetEntry returns an entry by name
func (d *Dictionary) GetEntry(name string) *DictionaryEntry {
	for i, entry := range d.Entries {
		if entry.Name == name {
			return &d.Entries[i]
		}
	}
	return nil
}

// HasEntry returns true if the dictionary contains the entry
func (d *Dictionary) HasEntry(name string) bool {
	for _, entry := range d.Entries {
		if entry.Name == name {
			return true
		}
	}
	return false
}

// Clear removes all entries
func (d *Dictionary) Clear() {
	d.Entries = []DictionaryEntry{}
}

// Len returns the number of entries
func (d *Dictionary) Len() int {
	return len(d.Entries)
}

// DictionaryWithDefault represents a DICTIONARY_WITH_DEFAULT entity
type DictionaryWithDefault struct {
	*entity
	HardOwned     int               // 280 - hard-owner flag
	Cloning       int               // 281 - cloning flag
	Entries       []DictionaryEntry // dictionary entries
	DefaultHandle string            // 360 - default entry handle
}

// NewDictionaryWithDefault creates a new DictionaryWithDefault entity
func NewDictionaryWithDefault() *DictionaryWithDefault {
	d := &DictionaryWithDefault{
		entity:        NewEntity(DICTIONARYWITHDEFAULT),
		HardOwned:     0,
		Cloning:       1,
		Entries:       []DictionaryEntry{},
		DefaultHandle: "",
	}
	return d
}

// IsEntity is for Entity interface.
func (d *DictionaryWithDefault) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (d *DictionaryWithDefault) Format(f format.Formatter) {
	d.entity.Format(f)
	f.WriteString(100, "AcDbDictionary")
	f.WriteInt(280, d.HardOwned)
	f.WriteInt(281, d.Cloning)

	for _, entry := range d.Entries {
		f.WriteString(3, entry.Name)
		f.WriteString(350, entry.Handle)
	}

	if d.DefaultHandle != "" {
		f.WriteString(360, d.DefaultHandle)
	}
}

// BBox returns the bounding box of the entity
func (d *DictionaryWithDefault) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the DictionaryWithDefault
func (d *DictionaryWithDefault) Transform(m *math.Matrix44) error {
	return nil
}

// Copy creates a deep copy of the DictionaryWithDefault entity
func (d *DictionaryWithDefault) Copy() Entity {
	dict := NewDictionaryWithDefault()
	dict.entity = d.entity
	dict.HardOwned = d.HardOwned
	dict.Cloning = d.Cloning
	dict.Entries = append(dict.Entries, d.Entries...)
	dict.DefaultHandle = d.DefaultHandle
	return dict
}

// Validate validates the DictionaryWithDefault entity
func (d *DictionaryWithDefault) Validate() error {
	if d.HardOwned < 0 || d.HardOwned > 1 {
		d.HardOwned = 0
	}
	if d.Cloning < 0 || d.Cloning > 5 {
		d.Cloning = 1
	}
	return nil
}

// AddEntry adds an entry to the dictionary
func (d *DictionaryWithDefault) AddEntry(name, handle string) {
	d.Entries = append(d.Entries, DictionaryEntry{Name: name, Handle: handle})
}

// RemoveEntry removes an entry by name
func (d *DictionaryWithDefault) RemoveEntry(name string) {
	for i, entry := range d.Entries {
		if entry.Name == name {
			d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
			return
		}
	}
}

// Clear removes all entries
func (d *DictionaryWithDefault) Clear() {
	d.Entries = []DictionaryEntry{}
}

// Len returns the number of entries
func (d *DictionaryWithDefault) Len() int {
	return len(d.Entries)
}
