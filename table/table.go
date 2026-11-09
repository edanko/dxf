package table

import (
	"errors"
	"fmt"
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// Table represents each TABLE.
type Table struct {
	name   string // 2
	handle string // 5
	size   int    // 70
	tables []SymbolTable
}

// NewTable creates a new Table.
func NewTable(name string) *Table {
	t := new(Table)
	t.name = name
	return t
}

// Format writes data to formatter.
func (t *Table) Format(f format.Formatter) {
	f.WriteString(0, "TABLE")
	f.WriteString(2, t.name)
	f.WriteString(5, t.handle)
	f.WriteString(100, "AcDbSymbolTable")
	f.WriteInt(70, len(t.tables))
	if t.name == "DIMSTYLE" {
		f.WriteString(100, "AcDbDimStyleTable")
		// AutoCAD writes the style count and 340 links in the table header,
		// before the records; ParseTables captures that tail and attaches it
		// to the last record, which emits it verbatim at its end.
	}
	for _, e := range t.tables {
		e.Format(f)
	}
	f.WriteString(0, "ENDTAB")
}

// SetEntries replaces the symbol table records and updates the declared size.
func (t *Table) SetEntries(entries []SymbolTable) {
	t.tables = entries
	t.size = len(entries)
}

// Handle returns a handle value of TABLE.
func (t *Table) Handle() string {
	return t.handle
}

// SetHandle sets handles to TABLE itself and each SymbolTable.
func (t *Table) SetHandle(hg *handle.HandleGenerator) {
	if t.handle == "" {
		t.handle = hg.Next()
	}
	for i := 0; i < t.size; i++ {
		t.tables[i].SetHandle(hg)
	}
}

// Add adds a new SymbolTable to TABLE.
func (t *Table) Add(st SymbolTable) {
	t.tables = append(t.tables, st)
	st.SetOwner(t)
	t.size++
}

// Clear deletes existing SymbolTables.
func (t *Table) Clear() {
	t.tables = make([]SymbolTable, 0)
	t.size = 0
}

// Contains reports if TABLE has the named SymbolTable.
func (t *Table) Contains(name string) (SymbolTable, error) {
	for _, st := range t.tables {
		if strings.EqualFold(st.Name(), name) {
			return st, nil
		}
	}
	return nil, fmt.Errorf("%s doesn't exist", name)
}

// Lookup returns the entry with the given name (case-insensitive).
func (t *Table) Lookup(name string) (SymbolTable, error) { return t.Contains(name) }

// Len returns the number of entries.
func (t *Table) Len() int { return len(t.tables) }

// ErrDuplicateEntry is returned when adding an entry whose name exists.
var ErrDuplicateEntry = errors.New("duplicate table entry name")

// ErrEntryNotFound is returned by Remove/Rename for unknown names.
var ErrEntryNotFound = errors.New("table entry not found")

// AddChecked adds an entry, rejecting duplicate names (case-insensitive).
func (t *Table) AddChecked(st SymbolTable) error {
	for _, e := range t.tables {
		if strings.EqualFold(e.Name(), st.Name()) {
			return fmt.Errorf("%w: %s in %s", ErrDuplicateEntry, st.Name(), t.name)
		}
	}
	t.Add(st)
	return nil
}

// Remove deletes the named entry.
func (t *Table) Remove(name string) error {
	for i, e := range t.tables {
		if strings.EqualFold(e.Name(), name) {
			t.tables = append(t.tables[:i], t.tables[i+1:]...)
			t.size = len(t.tables)
			return nil
		}
	}
	return fmt.Errorf("%w: %s in %s", ErrEntryNotFound, name, t.name)
}

// Rename changes an entry's name, rejecting collisions.
func (t *Table) Rename(oldName, newName string) error {
	for _, e := range t.tables {
		if strings.EqualFold(e.Name(), newName) && !strings.EqualFold(e.Name(), oldName) {
			return fmt.Errorf("%w: %s in %s", ErrDuplicateEntry, newName, t.name)
		}
	}
	e, err := t.Lookup(oldName)
	if err != nil {
		return fmt.Errorf("%w: %s in %s", ErrEntryNotFound, oldName, t.name)
	}
	rn, ok := e.(interface{ SetName(string) })
	if !ok {
		return fmt.Errorf("entry %s does not support rename", oldName)
	}
	rn.SetName(newName)
	return nil
}
