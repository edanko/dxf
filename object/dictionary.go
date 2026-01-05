package object

import (
	"github.com/edanko/dxf/format"
)

// Dictionary represents a DXF DICTIONARY object
type Dictionary struct {
	*BaseObject

	// Core properties
	HardOwned bool   `dxf:"280"` // Hard ownership flag (0=soft, 1=hard)
	Owner     string `dxf:"330"` // Owner handle (usually "0" for root dictionary)

	// Dictionary entries
	entries map[string]*DictionaryEntry

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
	DefaultName string `dxf:"1"` // Name of default entry
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
	d.entries[key] = entry
}

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
	// Basic header
	f.WriteString(0, "DICTIONARY")
	f.WriteString(5, d.Handle())
	f.WriteString(100, "AcDbDictionary")

	// Write hard ownership flag
	f.WriteInt(280, func() int {
		if d.HardOwned {
			return 1
		} else {
			return 0
		}
	}())

	// Write owner handle if specified
	if d.Owner != "" {
		f.WriteString(330, d.Owner)
	}

	// Write dictionary entries
	for key, entry := range d.entries {
		f.WriteString(3, key)           // Entry name
		f.WriteString(350, entry.Value) // Soft-owner handle
	}

	// Note: Default entry handling is done in DictionaryWithDefault.Format()
}

// FormatWithDefault writes DICTIONARYWDFLT data to formatter
func (d *DictionaryWithDefault) Format(f format.Formatter) {
	// DictionaryWithDefault is a subclass - write base dictionary first
	d.Dictionary.Format(f)

	// Add default entry specification
	if d.DefaultName != "" {
		f.WriteInt(280, 1)              // Schema flag for dictionary with default
		f.WriteString(1, d.DefaultName) // Default entry name
	}
}

// FormatVar writes DICTIONARYVAR data to formatter
func (d *DictionaryVar) Format(f format.Formatter) {
	// DictionaryVar is a subclass - write base dictionary first
	d.Dictionary.Format(f)

	// Add variable specification
	f.WriteInt(280, d.Schema) // Schema number
	f.WriteString(1, d.Value) // Variable value
}

// BBox returns bounding box for dictionary (n/a for object entity)
func (d *Dictionary) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}
