package object

import (
	"fmt"
	"sort"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// Dictionary represents DICTIONARY Object.
type Dictionary struct {
	handle string
	item   map[string]handle.Handler
}

// IsObject is for Object interface.
func (d *Dictionary) IsObject() bool {
	return true
}

// NewDictionary creates a new Dictionary.
func NewDictionary() *Dictionary {
	ds := make(map[string]handle.Handler)
	d := &Dictionary{
		item: ds,
	}
	return d
}

// Format writes data to formatter.
func (d *Dictionary) Format(f format.Formatter) {
	f.WriteString(0, "DICTIONARY")
	f.WriteString(5, d.handle)
	f.WriteString(100, "AcDbDictionary")
	f.WriteInt(281, 1)
	keys := make([]string, len(d.item))
	{
		idx := 0
		for k := range d.item {
			keys[idx] = k
			idx++
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		f.WriteString(3, k)
		f.WriteString(350, d.item[k].Handle())
	}
}

// Handle returns a handle value.
func (d *Dictionary) Handle() string {
	return d.handle
}

// SetHandle sets a handle.
func (d *Dictionary) SetHandle(hg *handle.HandleGenerator) {
	d.handle = hg.Next()
	for _, val := range d.item {
		val.SetHandle(hg)
	}
}

// AddItem adds new a new item to Dictionary.
func (d *Dictionary) AddItem(key string, value handle.Handler) error {
	if _, exist := d.item[key]; exist {
		return fmt.Errorf("key %s already exists", key)
	}
	d.item[key] = value
	return nil
}
