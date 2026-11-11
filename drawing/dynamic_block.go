// Dynamic block read helpers: resolve an INSERT to the original dynamic block
// definition, following the anonymous block reference stored by CAD
// applications. Only reading is supported; the helpers never modify the
// drawing.
package drawing

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/table"
)

const (
	// DynamicBlockGUIDAppID is the XDATA application id stored on a dynamic
	// block record.
	DynamicBlockGUIDAppID = "AcDbDynamicBlockGUID"
	// DynamicBlockRepBTagAppID is the XDATA application id that links an
	// anonymous block record to the original dynamic block record.
	DynamicBlockRepBTagAppID = "AcDbBlockRepBTag"
	// DynamicBlockTrueNameAppID is the XDATA application id that carries the
	// original dynamic block name.
	DynamicBlockTrueNameAppID = "AcDbDynamicBlockTrueName"
)

// DynamicBlockInfo describes the original dynamic block definition behind an
// INSERT. Anonymous references point at a generated *U block whose block
// record links back to the named dynamic definition.
type DynamicBlockInfo struct {
	// Name is the name referenced by the INSERT.
	Name string
	// DefinitionName is the name of the original dynamic block definition.
	DefinitionName string
	// BlockRecordHandle is the handle of the definition's BLOCK_RECORD.
	BlockRecordHandle string
	// GUID is the dynamic block GUID (empty when absent).
	GUID string
	// TrueName is the original dynamic block name from XDATA, if present.
	TrueName string
	// Anonymous reports whether the INSERT referenced a generated *U block.
	Anonymous bool
	// Metadata holds all XDATA applications on the block record, keyed by
	// application id, in source order.
	Metadata map[string][]string
}

// ResolveDynamicBlock returns the dynamic block definition for an INSERT.
// It returns nil without an error for ordinary (non-dynamic) inserts and for
// inserts whose block is not present. A dynamic reference whose anonymous
// block record no longer links to its definition returns an error.
func (d *Drawing) ResolveDynamicBlock(ins *entity.Insert) (*DynamicBlockInfo, error) {
	if d == nil || ins == nil {
		return nil, nil
	}
	name := ins.BlockName()
	if name == "" {
		return nil, nil
	}
	rec, ok := blockRecordByName(d.blockRecordTable(), name)
	if !ok {
		return nil, nil
	}
	info := &DynamicBlockInfo{
		Name:              name,
		DefinitionName:    name,
		BlockRecordHandle: rec.Handle(),
		Anonymous:         isAnonymousBlockName(name),
		Metadata:          xdataByApp(rec.RawTags()),
	}
	if values := info.Metadata[DynamicBlockGUIDAppID]; len(values) > 0 {
		info.GUID = values[0]
	}
	if values := info.Metadata[DynamicBlockTrueNameAppID]; len(values) > 0 {
		info.TrueName = values[0]
		info.DefinitionName = values[0]
	}
	if info.GUID != "" {
		if info.TrueName == "" {
			info.TrueName = name
		}
		return info, nil
	}

	// Indirect reference: the anonymous block record carries the handle of
	// the original dynamic block record.
	ref := xdataValueByCode(rec.RawTags(), DynamicBlockRepBTagAppID, 1005)
	if ref == "" {
		return nil, nil
	}
	definition, ok := blockRecordByHandle(d.blockRecordTable(), ref)
	if !ok {
		return nil, fmt.Errorf("dynamic block %q: definition handle %q not found", name, ref)
	}
	info.BlockRecordHandle = definition.Handle()
	info.DefinitionName = definition.Name()
	definitionMetadata := xdataByApp(definition.RawTags())
	if values := definitionMetadata[DynamicBlockGUIDAppID]; len(values) > 0 {
		info.GUID = values[0]
	}
	if values := definitionMetadata[DynamicBlockTrueNameAppID]; len(values) > 0 {
		info.TrueName = values[0]
		info.DefinitionName = values[0]
	}
	if info.GUID == "" {
		return nil, fmt.Errorf("dynamic block %q: definition %q carries no dynamic block GUID", name, definition.Name())
	}
	return info, nil
}

func isAnonymousBlockName(name string) bool {
	switch name {
	case "*Model_Space", "*Paper_Space":
		return false
	}
	return strings.HasPrefix(name, "*")
}

func blockRecordByName(t *table.Table, name string) (*table.BlockRecord, bool) {
	if t == nil {
		return nil, false
	}
	st, err := t.Lookup(name)
	if err != nil {
		return nil, false
	}
	rec, ok := st.(*table.BlockRecord)
	return rec, ok
}

func blockRecordByHandle(t *table.Table, h string) (*table.BlockRecord, bool) {
	if t == nil || h == "" {
		return nil, false
	}
	for _, entry := range t.Entries() {
		rec, ok := entry.(*table.BlockRecord)
		if ok && strings.EqualFold(rec.Handle(), h) {
			return rec, true
		}
	}
	return nil, false
}

// xdataByApp groups retained XDATA values by application id. Each 1001 tag
// starts a new application block; all following codes up to the next 1001
// belong to that application.
func xdataByApp(tags []table.EntryTag) map[string][]string {
	out := make(map[string][]string)
	app := ""
	for _, tag := range tags {
		switch tag.Code {
		case 1001:
			app = strings.TrimSpace(tag.Value)
		case 1000, 1002, 1003, 1004, 1005, 1010, 1011, 1012, 1013, 1020, 1030, 1040, 1070, 1071:
			if app != "" {
				out[app] = append(out[app], tag.Value)
			}
		}
	}
	return out
}

// xdataValueByCode returns the first value of code inside the XDATA block for
// app.
func xdataValueByCode(tags []table.EntryTag, app string, code int) string {
	current := ""
	for _, tag := range tags {
		switch tag.Code {
		case 1001:
			current = strings.TrimSpace(tag.Value)
		default:
			if current == app && tag.Code == code {
				return strings.TrimSpace(tag.Value)
			}
		}
	}
	return ""
}
