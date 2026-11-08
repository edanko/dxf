package object

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

func TestBaseObjectHandle(t *testing.T) {
	obj := &BaseObject{}
	if obj.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", obj.Handle())
	}
}

func TestBaseObjectSetHandle(t *testing.T) {
	obj := &BaseObject{}
	obj.SetHandle(handle.NewHandleGenerator())

	if obj.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestObjectObjectIsObject(t *testing.T) {
	obj := &objectObject{}
	if !obj.IsObject() {
		t.Error("objectObject should return true for IsObject()")
	}
}

func TestObjectObjectHandle(t *testing.T) {
	obj := &objectObject{}
	if obj.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", obj.Handle())
	}
}

func TestObjectObjectSetHandle(t *testing.T) {
	obj := &objectObject{}
	obj.SetHandle(handle.NewHandleGenerator())

	if obj.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestObjectObjectFormat(t *testing.T) {
	obj := &objectObject{}
	obj.SetHandle(handle.NewHandleGenerator())

	f := format.NewASCII()
	obj.Format(f)
	result := f.Output()

	if !strings.Contains(result, "OBJECT") {
		t.Error("Format should contain 'OBJECT'")
	}
	if !strings.Contains(result, obj.Handle()) {
		t.Error("Format should contain handle")
	}
}

func TestObjectInterface(t *testing.T) {
	var obj Object = &objectObject{}
	if !obj.IsObject() {
		t.Error("objectObject should satisfy Object interface")
	}
}

func TestXRecordNew(t *testing.T) {
	xrec := NewXRecord()
	if xrec == nil {
		t.Error("NewXRecord returned nil")
	}
	if xrec.IsObject() != true {
		t.Error("XRecord should return true for IsObject()")
	}
}

func TestXRecordNewWithOwner(t *testing.T) {
	xrec := NewXRecordWithOwner("123")
	if xrec.owner != "123" {
		t.Errorf("Expected owner '123', got '%s'", xrec.owner)
	}
}

func TestXRecordHandle(t *testing.T) {
	xrec := NewXRecord()
	if xrec.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", xrec.Handle())
	}
}

func TestXRecordSetHandle(t *testing.T) {
	xrec := NewXRecord()
	xrec.SetHandle("ABC")
	if xrec.Handle() != "ABC" {
		t.Errorf("Expected handle 'ABC', got %s", xrec.Handle())
	}
}

func TestXRecordSetOwner(t *testing.T) {
	xrec := NewXRecord()
	xrec.SetOwner("DEF")
	if xrec.owner != "DEF" {
		t.Errorf("Expected owner 'DEF', got '%s'", xrec.owner)
	}
}

func TestXRecordAddTag(t *testing.T) {
	xrec := NewXRecord()
	xrec.AddTag(1, "test")

	if len(xrec.tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(xrec.tags))
	}

	tag, ok := xrec.GetTag(1)
	if !ok {
		t.Error("GetTag should return the added tag")
	}
	if tag.value != "test" {
		t.Errorf("Expected tag value 'test', got '%s'", tag.value)
	}
}

func TestXRecordAddData(t *testing.T) {
	xrec := NewXRecord()
	data := map[int]string{
		1: "value1",
		2: "value2",
		3: "value3",
	}
	xrec.AddData(data)

	if len(xrec.tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(xrec.tags))
	}

	val, ok := xrec.GetTagValue(1)
	if !ok || val != "value1" {
		t.Errorf("GetTagValue(1) = '%s', ok=%v", val, ok)
	}
}

func TestXRecordGetTagNotFound(t *testing.T) {
	xrec := NewXRecord()
	_, ok := xrec.GetTag(999)
	if ok {
		t.Error("GetTag should return false for non-existent tag")
	}
}

func TestXRecordGetTagValueNotFound(t *testing.T) {
	xrec := NewXRecord()
	_, ok := xrec.GetTagValue(999)
	if ok {
		t.Error("GetTagValue should return false for non-existent tag")
	}
}

func TestXRecordRemoveTag(t *testing.T) {
	xrec := NewXRecord()
	xrec.AddTag(1, "test1")
	xrec.AddTag(2, "test2")
	xrec.AddTag(3, "test3")

	if len(xrec.tags) != 3 {
		t.Errorf("Expected 3 tags before remove, got %d", len(xrec.tags))
	}

	xrec.RemoveTag(2)

	if len(xrec.tags) != 2 {
		t.Errorf("Expected 2 tags after remove, got %d", len(xrec.tags))
	}

	_, ok := xrec.GetTag(2)
	if ok {
		t.Error("Removed tag should not be found")
	}
}

func TestXRecordClear(t *testing.T) {
	xrec := NewXRecord()
	xrec.AddTag(1, "test")
	xrec.AddTag(2, "test2")

	xrec.Clear()

	if len(xrec.tags) != 0 {
		t.Errorf("Expected 0 tags after Clear, got %d", len(xrec.tags))
	}
}

func TestXRecordGetTags(t *testing.T) {
	xrec := NewXRecord()
	xrec.AddTag(1, "a")
	xrec.AddTag(2, "b")

	tags := xrec.GetTags()

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags in map, got %d", len(tags))
	}
	if tags[1] != "a" {
		t.Errorf("Expected tags[1]='a', got '%s'", tags[1])
	}
	if tags[2] != "b" {
		t.Errorf("Expected tags[2]='b', got '%s'", tags[2])
	}
}

func TestXRecordSchema(t *testing.T) {
	xrec := NewXRecord()
	xrec.SetSchema(5)

	if xrec.schema != 5 {
		t.Errorf("Expected schema 5, got %d", xrec.schema)
	}
}

func TestXRecordCloning(t *testing.T) {
	xrec := NewXRecord()
	xrec.SetCloning(1)

	if xrec.cloning != 1 {
		t.Errorf("Expected cloning 1, got %d", xrec.cloning)
	}
}

func TestXRecordFormat(t *testing.T) {
	xrec := NewXRecord()
	xrec.SetHandle("ABC")
	xrec.AddTag(1, "test")

	f := format.NewASCII()
	xrec.Format(f)
	result := f.Output()

	if !strings.Contains(result, "XRECORD") {
		t.Error("Format should contain 'XRECORD'")
	}
	if !strings.Contains(result, "ABC") {
		t.Error("Format should contain handle 'ABC'")
	}
}

func TestXRecordFormatWithOwner(t *testing.T) {
	xrec := NewXRecordWithOwner("OWNER123")
	xrec.SetHandle("HANDLE")
	xrec.AddTag(1, "data")

	f := format.NewASCII()
	xrec.Format(f)
	result := f.Output()

	if !strings.Contains(result, "OWNER123") {
		t.Error("Format should contain owner")
	}
}

func TestXRecordFormatWithSchema(t *testing.T) {
	xrec := NewXRecord()
	xrec.SetSchema(2)
	xrec.AddTag(1, "test")

	f := format.NewASCII()
	xrec.Format(f)

	if xrec.schema != 2 {
		t.Errorf("Expected schema 2, got %d", xrec.schema)
	}
}

func TestXRecordTag(t *testing.T) {
	tag := &XRecordTag{groupCode: 10, value: "test"}
	if tag.groupCode != 10 {
		t.Errorf("Expected groupCode 10, got %d", tag.groupCode)
	}
	if tag.value != "test" {
		t.Errorf("Expected value 'test', got '%s'", tag.value)
	}
}

func TestDictionaryNew(t *testing.T) {
	dict := NewDictionary()
	if dict == nil {
		t.Error("NewDictionary returned nil")
	}
	if !dict.IsObject() {
		t.Error("Dictionary should return true for IsObject()")
	}
	if dict.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", dict.Handle())
	}
}

func TestDictionarySetHandle(t *testing.T) {
	dict := NewDictionary()
	dict.SetHandle(handle.NewHandleGenerator())

	if dict.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestDictionaryHardOwned(t *testing.T) {
	dict := NewDictionary()
	if !dict.HardOwned {
		t.Error("Dictionary should default to HardOwned=true")
	}

	dict.HardOwned = false
	if dict.HardOwned {
		t.Error("Dictionary.HardOwned should be settable to false")
	}
}

func TestDictionaryOwner(t *testing.T) {
	dict := NewDictionary()
	dict.Owner = "ABC123"

	if dict.Owner != "ABC123" {
		t.Errorf("Expected Owner 'ABC123', got '%s'", dict.Owner)
	}
}

func TestDictionaryAddEntry(t *testing.T) {
	dict := NewDictionary()
	dict.AddEntry("key1", "value1")

	entry, ok := dict.GetEntry("key1")
	if !ok {
		t.Error("GetEntry should return true for added key")
	}
	if entry.Value != "value1" {
		t.Errorf("Expected entry.Value 'value1', got '%s'", entry.Value)
	}
}

func TestDictionaryAddMultipleEntries(t *testing.T) {
	dict := NewDictionary()
	dict.AddEntry("key1", "value1")
	dict.AddEntry("key2", "value2")
	dict.AddEntry("key3", "value3")

	entries := dict.GetEntries()
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestDictionaryGetEntryNotFound(t *testing.T) {
	dict := NewDictionary()
	_, ok := dict.GetEntry("nonexistent")
	if ok {
		t.Error("GetEntry should return false for non-existent key")
	}
}

func TestDictionaryRemoveEntry(t *testing.T) {
	dict := NewDictionary()
	dict.AddEntry("key1", "value1")
	dict.AddEntry("key2", "value2")

	dict.RemoveEntry("key1")

	_, ok := dict.GetEntry("key1")
	if ok {
		t.Error("GetEntry should return false after RemoveEntry")
	}

	entry, ok := dict.GetEntry("key2")
	if !ok {
		t.Error("GetEntry should still return true for key2")
	}
	if entry.Value != "value2" {
		t.Errorf("Expected entry.Value 'value2', got '%s'", entry.Value)
	}
}

func TestDictionaryGetEntries(t *testing.T) {
	dict := NewDictionary()
	dict.AddEntry("a", "1")
	dict.AddEntry("b", "2")

	entries := dict.GetEntries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
	if entries["a"].Value != "1" {
		t.Errorf("Expected entries['a'].Value '1', got '%s'", entries["a"].Value)
	}
	if entries["b"].Value != "2" {
		t.Errorf("Expected entries['b'].Value '2', got '%s'", entries["b"].Value)
	}
}

func TestDictionaryFormat(t *testing.T) {
	dict := NewDictionary()
	hg := handle.NewHandleGenerator()
	dict.SetHandle(hg)
	dict.HardOwned = true
	dict.AddEntry("testKey", "testValue")

	f := format.NewASCII()
	dict.Format(f)
	result := f.Output()

	if !strings.Contains(result, "DICTIONARY") {
		t.Error("Format should contain 'DICTIONARY'")
	}
	if !strings.Contains(result, dict.Handle()) {
		t.Error("Format should contain handle")
	}
	if !strings.Contains(result, "AcDbDictionary") {
		t.Error("Format should contain 'AcDbDictionary'")
	}
	if !strings.Contains(result, "testKey") {
		t.Error("Format should contain entry key 'testKey'")
	}
	if !strings.Contains(result, "testValue") {
		t.Error("Format should contain entry value 'testValue'")
	}
}

func TestDictionaryFormatWithOwner(t *testing.T) {
	dict := NewDictionary()
	hg := handle.NewHandleGenerator()
	dict.SetHandle(hg)
	dict.Owner = "OWNER123"

	f := format.NewASCII()
	dict.Format(f)
	result := f.Output()

	if !strings.Contains(result, "OWNER123") {
		t.Error("Format should contain owner 'OWNER123'")
	}
}

func TestDictionaryBBox(t *testing.T) {
	dict := NewDictionary()
	min, max := dict.BBox()

	if len(min) != 3 || len(max) != 3 {
		t.Error("BBox should return 3-element slices")
	}
	if min[0] != 0 || max[0] != 0 {
		t.Error("BBox should return zeros for dictionary")
	}
}

func TestDictionaryWithDefaultNew(t *testing.T) {
	dict := NewDictionaryWithDefault()
	if dict == nil {
		t.Error("NewDictionaryWithDefault returned nil")
	}
	if dict.DefaultName != "" {
		t.Errorf("Expected empty DefaultName, got '%s'", dict.DefaultName)
	}
}

func TestDictionaryWithDefaultSetGetDefaultEntry(t *testing.T) {
	dict := NewDictionaryWithDefault()
	dict.SetDefaultEntry("defaultKey")

	if dict.GetDefaultEntry() != "defaultKey" {
		t.Errorf("Expected 'defaultKey', got '%s'", dict.GetDefaultEntry())
	}
}

func TestDictionaryWithDefaultFormat(t *testing.T) {
	dict := NewDictionaryWithDefault()
	hg := handle.NewHandleGenerator()
	dict.SetHandle(hg)
	dict.SetDefaultEntry("defaultKey")

	f := format.NewASCII()
	dict.Format(f)
	result := f.Output()

	if !strings.Contains(result, "DICTIONARY") {
		t.Error("Format should contain 'DICTIONARY'")
	}
	if !strings.Contains(result, "defaultKey") {
		t.Error("Format should contain default entry name 'defaultKey'")
	}
}

func TestDictionaryVarNew(t *testing.T) {
	dictVar := NewDictionaryVar()
	if dictVar == nil {
		t.Error("NewDictionaryVar returned nil")
	}
}

func TestDictionaryVarSetGetValue(t *testing.T) {
	dictVar := NewDictionaryVar()
	dictVar.SetValue("testValue")

	if dictVar.GetValue() != "testValue" {
		t.Errorf("Expected 'testValue', got '%s'", dictVar.GetValue())
	}
}

func TestDictionaryVarSchema(t *testing.T) {
	dictVar := NewDictionaryVar()
	dictVar.Schema = 5

	if dictVar.Schema != 5 {
		t.Errorf("Expected Schema 5, got %d", dictVar.Schema)
	}
}

func TestDictionaryVarFormat(t *testing.T) {
	dictVar := NewDictionaryVar()
	hg := handle.NewHandleGenerator()
	dictVar.SetHandle(hg)
	dictVar.SetValue("varValue")
	dictVar.Schema = 3

	f := format.NewASCII()
	dictVar.Format(f)
	result := f.Output()

	if !strings.Contains(result, "DICTIONARY") {
		t.Error("Format should contain 'DICTIONARY'")
	}
	if !strings.Contains(result, "varValue") {
		t.Error("Format should contain value 'varValue'")
	}
}

func TestGroupNew(t *testing.T) {
	group := NewGroup("testGroup", "test description")
	if group == nil {
		t.Error("NewGroup returned nil")
	}
	if group.Name != "testGroup" {
		t.Errorf("Expected Name 'testGroup', got '%s'", group.Name)
	}
	if group.Description != "test description" {
		t.Errorf("Expected Description 'test description', got '%s'", group.Description)
	}
	if !group.IsObject() {
		t.Error("Group should return true for IsObject()")
	}
}

func TestGroupHandle(t *testing.T) {
	group := NewGroup("test", "desc")
	if group.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", group.Handle())
	}

	group.SetHandle(handle.NewHandleGenerator())
	if group.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestGroupSelectable(t *testing.T) {
	group := NewGroup("test", "desc")
	if !group.selectable {
		t.Error("Group should default to selectable=true")
	}
}

func TestDictionaryEntry(t *testing.T) {
	entry := &DictionaryEntry{Key: "key", Value: "value"}
	if entry.Key != "key" {
		t.Errorf("Expected Key 'key', got '%s'", entry.Key)
	}
	if entry.Value != "value" {
		t.Errorf("Expected Value 'value', got '%s'", entry.Value)
	}
}

func TestAcDbPlaceHolderNew(t *testing.T) {
	ph := NewAcDbPlaceHolder()
	if ph == nil {
		t.Error("NewAcDbPlaceHolder returned nil")
	}
	if !ph.IsObject() {
		t.Error("AcDbPlaceHolder should return true for IsObject()")
	}
}

func TestAcDbPlaceHolderHandle(t *testing.T) {
	ph := NewAcDbPlaceHolder()
	if ph.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", ph.Handle())
	}

	ph.SetHandle(handle.NewHandleGenerator())
	if ph.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestAcDbPlaceHolderFormat(t *testing.T) {
	ph := NewAcDbPlaceHolder()
	hg := handle.NewHandleGenerator()
	ph.SetHandle(hg)

	f := format.NewASCII()
	ph.Format(f)
	result := f.Output()

	if !strings.Contains(result, "ACDBPLACEHOLDER") {
		t.Error("Format should contain 'ACDBPLACEHOLDER'")
	}
	if !strings.Contains(result, ph.Handle()) {
		t.Error("Format should contain handle")
	}
}

func TestAcDbDictionaryWDFLTNew(t *testing.T) {
	owner := NewDictionary()
	dict, ph := NewAcDbDictionaryWDFLT(owner)
	if dict == nil {
		t.Error("NewAcDbDictionaryWDFLT returned nil dict")
	}
	if ph == nil {
		t.Error("NewAcDbDictionaryWDFLT returned nil placeholder")
	}
	if !dict.IsObject() {
		t.Error("AcDbDictionaryWDFLT should return true for IsObject()")
	}
}

func TestAcDbDictionaryWDFLTHandle(t *testing.T) {
	dict, _ := NewAcDbDictionaryWDFLT(nil)
	if dict.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", dict.Handle())
	}

	dict.SetHandle(handle.NewHandleGenerator())
	if dict.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestAcDbDictionaryWDFLTAddItem(t *testing.T) {
	owner := NewDictionary()
	dict, _ := NewAcDbDictionaryWDFLT(owner)

	ph2 := NewAcDbPlaceHolder()
	err := dict.AddItem("Item2", ph2)
	if err != nil {
		t.Errorf("AddItem should succeed, got error: %v", err)
	}
}

func TestAcDbDictionaryWDFLTAddItemDuplicate(t *testing.T) {
	owner := NewDictionary()
	dict, _ := NewAcDbDictionaryWDFLT(owner)

	err := dict.AddItem("Normal", NewAcDbPlaceHolder())
	if err == nil {
		t.Error("AddItem should return error for duplicate key")
	}
}

func TestAcDbDictionaryWDFLTFormat(t *testing.T) {
	owner := NewDictionary()
	dict, _ := NewAcDbDictionaryWDFLT(owner)
	hg := handle.NewHandleGenerator()
	dict.SetHandle(hg)

	f := format.NewASCII()
	dict.Format(f)
	result := f.Output()

	if !strings.Contains(result, "ACDBDICTIONARYWDFLT") {
		t.Error("Format should contain 'ACDBDICTIONARYWDFLT'")
	}
	if !strings.Contains(result, dict.Handle()) {
		t.Error("Format should contain handle")
	}
	if !strings.Contains(result, "AcDbDictionary") {
		t.Error("Format should contain 'AcDbDictionary'")
	}
}
