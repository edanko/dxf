package object

import (
	"github.com/edanko/dxf/format"
)

// Dictionary represents a DXF DICTIONARY object
type Dictionary struct {
	*BaseObject

	// Core properties
	HardOwned bool            `dxf:"280"` // Hard ownership flag (0=soft, 1=hard)
	noFlag    bool            // source record carries no code-280: Format must not emit one
	hardKeys  map[string]bool // entry keys loaded from code-360 (emitted as 360)
	Owner     string          `dxf:"330"` // Owner handle (usually "0" for root dictionary)

	// Dictionary entries
	entries map[string]*DictionaryEntry
	// entryOrder keeps insertion order for deterministic output.
	entryOrder []string

	// Default entry (for DictionaryWithDefault)
	DefaultEntry *DictionaryEntry `dxf:"340"` // Handle to default entry

	// Schema information (for DictionaryVar)
	Schema int `dxf:"280"` // Schema number

	// Value information (for DictionaryVar)
	Value string `dxf:"1"` // Variable value
}

// DictionaryEntry represents a single dictionary entry
type DictionaryEntry struct {
	Key   string `dxf:"3"`   // Entry name (group code 3)
	Value string `dxf:"350"` // Soft-owner handle (group code 350)
	// HardOwner string `dxf:"360"` // Hard-owner handle (alternative to 350)
}

// DictionaryWithDefault represents a DICTIONARYWDFLT object
type DictionaryWithDefault struct {
	*Dictionary
	DefaultName string `dxf:"1"`   // Name of default entry
	Schema      int    `dxf:"280"` // Schema flag emitted after the subclass marker
	hasSchema   bool   // source record carries a schema 280 (second flag)
}

// DictionaryVar represents a DICTIONARYVAR object
type DictionaryVar struct {
	*Dictionary
}

// NewDictionary creates a new dictionary
func NewDictionary() *Dictionary {
	return &Dictionary{
		BaseObject: &BaseObject{},
		entries:    make(map[string]*DictionaryEntry),
		HardOwned:  true, // Default to hard ownership for reliability
	}
}

// NewDictionaryWithDefault creates a new dictionary with default entry
func NewDictionaryWithDefault() *DictionaryWithDefault {
	return &DictionaryWithDefault{
		Dictionary: NewDictionary(),
	}
}

// NewDictionaryVar creates a new dictionary variable
func NewDictionaryVar() *DictionaryVar {
	return &DictionaryVar{
		Dictionary: NewDictionary(),
	}
}

// IsObject returns true for dictionary objects
func (d *Dictionary) IsObject() bool {
	return true
}

// AddEntry adds an entry to the dictionary
func (d *Dictionary) AddEntry(key string, value string) {
	entry := &DictionaryEntry{
		Key:   key,
		Value: value,
	}
	if _, exists := d.entries[key]; !exists {
		d.entryOrder = append(d.entryOrder, key)
	}
	d.entries[key] = entry
}

// SetHandleString sets an explicit handle loaded from a file.
func (d *Dictionary) SetHandleString(h string) { d.SetHandleValue(h) }

// SetHandleValue sets the underlying handle string.
func (d *Dictionary) SetHandleValue(h string) {
	if d.BaseObject == nil {
		d.BaseObject = &BaseObject{}
	}
	d.BaseObject.handle = h
}

// SetOwnerString sets the owner handle loaded from a file.
func (d *Dictionary) SetOwnerString(h string) { d.Owner = h }

// SetNoFlag suppresses the code-280 hard-ownership flag in Format, matching
// source records that carry no 280 at all.
func (d *Dictionary) SetNoFlag(v bool) { d.noFlag = v }

// SetHardOwnerKeys records entry keys loaded from code-360; Format re-emits
// those links as 360 to preserve the wire spelling.
func (d *Dictionary) SetHardOwnerKeys(keys map[string]bool) { d.hardKeys = keys }

// AddItem adds an object to the dictionary (compatibility method)
func (d *Dictionary) AddItem(key string, obj Object) {
	// Add object using its handle as value
	d.AddEntry(key, obj.Handle())
}

// GetEntry returns an entry from the dictionary
func (d *Dictionary) GetEntry(key string) (*DictionaryEntry, bool) {
	entry, exists := d.entries[key]
	return entry, exists
}

// RemoveEntry removes an entry from the dictionary
func (d *Dictionary) RemoveEntry(key string) {
	delete(d.entries, key)
	for i, k := range d.entryOrder {
		if k == key {
			d.entryOrder = append(d.entryOrder[:i], d.entryOrder[i+1:]...)
			break
		}
	}
}

// GetEntries returns all dictionary entries
func (d *Dictionary) GetEntries() map[string]*DictionaryEntry {
	result := make(map[string]*DictionaryEntry)
	for k, v := range d.entries {
		result[k] = v
	}
	return result
}

// SetDefaultEntry sets the default entry for DictionaryWithDefault
func (d *DictionaryWithDefault) SetDefaultEntry(key string) {
	d.DefaultName = key
}

// GetDefaultEntry returns default entry name for DictionaryWithDefault
func (d *DictionaryWithDefault) GetDefaultEntry() string {
	return d.DefaultName
}

// SetValue sets the variable value for DictionaryVar
func (d *DictionaryVar) SetValue(value string) {
	d.Value = value
}

// GetValue returns the variable value for DictionaryVar
func (d *DictionaryVar) GetValue() string {
	return d.Value
}

// Format writes DICTIONARY data to formatter
func (d *Dictionary) Format(f format.Formatter) {
	d.writeHeader(f, 0)
	for _, key := range d.entryOrder {
		entry, ok := d.entries[key]
		if !ok {
			continue
		}
		f.WriteString(3, key) // Entry name
		if d.hardKeys[key] {
			f.WriteString(360, entry.Value) // Hard-owner handle (wire spelling)
		} else {
			f.WriteString(350, entry.Value) // Soft-owner handle
		}
	}

	d.formatRetained(f)
}

// writeHeader emits the code-0 marker, handle, hard-ownership flag (280),
// owner (330), and AcDbDictionary subclass marker. skip280 suppresses the
// flag for sub-class writers that emit their own 280 values. after280 places
// the marker before the flag, matching ACDBDICTIONARYWDFLT records.
func (d *Dictionary) writeHeader(f format.Formatter, mode int) {
	wire := d.WireType()
	if wire == "" {
		wire = "DICTIONARY"
	}
	f.WriteString(0, wire)
	f.WriteString(5, d.Handle())
	if mode == 2 {
		f.WriteString(100, "AcDbDictionary")
	}
	if mode != 1 && !d.noFlag {
		flag := 0
		if d.HardOwned {
			flag = 1
		}
		f.WriteInt(280, flag)
	}
	if d.Owner != "" {
		f.WriteString(330, d.Owner)
	}
	if mode == 0 || mode == 1 {
		f.WriteString(100, "AcDbDictionary")
	}
}

// Format writes DICTIONARYWDFLT data to formatter. The schema flag (second
// code-280 in the record) is emitted by the retained tag stream.
func (d *DictionaryWithDefault) Format(f format.Formatter) {
	d.Dictionary.writeHeader(f, 2)
	for _, key := range d.entryOrder {
		entry, ok := d.entries[key]
		if !ok {
			continue
		}
		f.WriteString(3, key)
		if d.hardKeys[key] {
			f.WriteString(360, entry.Value)
		} else {
			f.WriteString(350, entry.Value)
		}
	}
	f.WriteString(100, "AcDbDictionaryWithDefault")
	if d.hasSchema { // Schema flag (second 280 of the record)
		f.WriteInt(280, d.Schema)
	}
	if d.DefaultName != "" {
		f.WriteString(1, d.DefaultName) // Default entry name
	}
	d.formatRetained(f)
}

// Format writes DICTIONARYVAR data to formatter. The schema flag is emitted
// before the AcDbDictionary subclass marker by AutoCAD writers.
func (d *DictionaryVar) Format(f format.Formatter) {
	wire := d.WireType()
	if wire == "" {
		wire = "DICTIONARYVAR"
	}
	f.WriteString(0, wire)
	f.WriteString(5, d.Handle())
	if !d.noFlag { // Schema number (first 280 of the record)
		f.WriteInt(280, d.Schema)
	}
	if d.Owner != "" {
		f.WriteString(330, d.Owner)
	}
	f.WriteString(100, "AcDbDictionary")
	f.WriteString(1, d.Value) // Variable value
	d.formatRetained(f)
}

// BBox returns bounding box for dictionary (n/a for object entity)
func (d *Dictionary) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}
