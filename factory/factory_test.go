package factory

import (
	"testing"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

type mockEntity struct{}

func (m *mockEntity) IsEntity() bool                                      { return true }
func (m *mockEntity) Handle() string                                      { return "" }
func (m *mockEntity) SetHandle(hg *handle.HandleGenerator)                {}
func (m *mockEntity) DXFType() string                                     { return "" }
func (m *mockEntity) GetAttributes() map[string]interface{}               { return nil }
func (m *mockEntity) LoadAttributes(attribs map[string]interface{}) error { return nil }
func (m *mockEntity) BBox() ([]float64, []float64)                        { return nil, nil }
func (m *mockEntity) Layer() *table.Layer                                 { return nil }
func (m *mockEntity) SetLayer(layer *table.Layer)                         {}
func (m *mockEntity) Format(f format.Formatter)                           {}
func (m *mockEntity) SetBlockRecord(blockRecord handle.Handler)           {}
func (m *mockEntity) SetLtscale(ltscale float64)                          {}
func (m *mockEntity) SetColor(c color.ColorNumber)                        {}

type mockCreator struct {
	name string
}

func (m *mockCreator) New() entity.Entity {
	return &mockEntity{}
}

func (m *mockCreator) NewWithAttribs(attribs map[string]interface{}) entity.Entity {
	return &mockEntity{}
}

func (m *mockCreator) LoadFromTags(tags []Tag) (entity.Entity, error) {
	return &mockEntity{}, nil
}

func (m *mockCreator) DXFType() string {
	return m.name
}

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Error("NewFactory should return non-nil")
	}
	if factory.entityTypes == nil {
		t.Error("entityTypes map should be initialized")
	}
}

func TestDefaultFactory(t *testing.T) {
	if DefaultFactory == nil {
		t.Error("DefaultFactory should not be nil")
	}
}

func TestFactoryRegister(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "TEST"}

	factory.Register("TEST", creator)

	if len(factory.entityTypes) != 1 {
		t.Errorf("Expected 1 registered type, got %d", len(factory.entityTypes))
	}
}

func TestFactoryRegisterCaseInsensitive(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "TEST"}

	factory.Register("test", creator)
	factory.Register("Test", creator)
	factory.Register("TEST", creator)

	if len(factory.entityTypes) != 1 {
		t.Errorf("Expected 1 registered type (case-insensitive), got %d", len(factory.entityTypes))
	}
}

func TestNewTag(t *testing.T) {
	tag := Tag{GroupCode: 10, Value: "test"}
	if tag.GroupCode != 10 {
		t.Errorf("Expected GroupCode 10, got %d", tag.GroupCode)
	}
	if tag.Value != "test" {
		t.Errorf("Expected Value 'test', got %v", tag.Value)
	}
}

func TestTagTypes(t *testing.T) {
	stringTag := Tag{GroupCode: 1, Value: "hello"}
	if stringTag.Value != "hello" {
		t.Error("String tag should hold 'hello'")
	}

	intTag := Tag{GroupCode: 62, Value: 5}
	if intTag.Value != 5 {
		t.Error("Int tag should hold 5")
	}

	floatTag := Tag{GroupCode: 40, Value: 3.14}
	if floatTag.Value != 3.14 {
		t.Error("Float tag should hold 3.14")
	}

	boolTag := Tag{GroupCode: 290, Value: true}
	if boolTag.Value != true {
		t.Error("Bool tag should hold true")
	}
}

func TestFactoryCreate(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "MOCK"}

	factory.Register("MOCK", creator)

	entity, err := factory.Create("MOCK", nil)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	if entity == nil {
		t.Error("Create should return non-nil entity")
	}
}

func TestFactoryCreateUnknown(t *testing.T) {
	factory := NewFactory()

	_, err := factory.Create("UNKNOWN", nil)
	if err == nil {
		t.Error("Create should return error for unknown type")
	}
}

func TestFactoryCreateNew(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "MOCK"}

	factory.Register("MOCK", creator)

	entity, err := factory.CreateNew("MOCK")
	if err != nil {
		t.Errorf("CreateNew failed: %v", err)
	}
	if entity == nil {
		t.Error("CreateNew should return non-nil entity")
	}
}

func TestFactoryCreateNewUnknown(t *testing.T) {
	factory := NewFactory()

	_, err := factory.CreateNew("UNKNOWN")
	if err == nil {
		t.Error("CreateNew should return error for unknown type")
	}
}

func TestFactoryLoadFromTags(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "MOCK"}

	factory.Register("MOCK", creator)

	tags := []Tag{{GroupCode: 1, Value: "test"}}
	entity, err := factory.LoadFromTags("MOCK", tags)
	if err != nil {
		t.Errorf("LoadFromTags failed: %v", err)
	}
	if entity == nil {
		t.Error("LoadFromTags should return non-nil entity")
	}
}

func TestFactoryLoadFromTagsUnknown(t *testing.T) {
	factory := NewFactory()

	tags := []Tag{{GroupCode: 1, Value: "test"}}
	_, err := factory.LoadFromTags("UNKNOWN", tags)
	if err == nil {
		t.Error("LoadFromTags should return error for unknown type")
	}
}

func TestFactoryGetRegisteredTypes(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "TYPE1"}
	creator2 := &mockCreator{name: "TYPE2"}

	factory.Register("TYPE1", creator)
	factory.Register("TYPE2", creator2)

	types := factory.GetRegisteredTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 registered types, got %d", len(types))
	}
}

func TestFactoryIsRegistered(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "REGISTERED"}

	factory.Register("REGISTERED", creator)

	if !factory.IsRegistered("REGISTERED") {
		t.Error("IsRegistered should return true for registered type")
	}
	if factory.IsRegistered("NOTREGISTERED") {
		t.Error("IsRegistered should return false for unregistered type")
	}
}

func TestFactorySetDefault(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "DEFAULT"}

	factory.SetDefault(creator)

	if factory.defaultType != creator {
		t.Error("defaultType should be set")
	}
}

func TestFactoryCreateWithDefault(t *testing.T) {
	factory := NewFactory()
	defaultCreator := &mockCreator{name: "DEFAULT"}
	factory.SetDefault(defaultCreator)

	entity, err := factory.Create("UNKNOWN", nil)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	if entity == nil {
		t.Error("Create with default should return non-nil entity")
	}
}

func TestFactoryCreateNewWithDefault(t *testing.T) {
	factory := NewFactory()
	defaultCreator := &mockCreator{name: "DEFAULT"}
	factory.SetDefault(defaultCreator)

	entity, err := factory.CreateNew("UNKNOWN")
	if err != nil {
		t.Errorf("CreateNew failed: %v", err)
	}
	if entity == nil {
		t.Error("CreateNew with default should return non-nil entity")
	}
}

func TestFactoryLoadFromTagsWithDefault(t *testing.T) {
	factory := NewFactory()
	defaultCreator := &mockCreator{name: "DEFAULT"}
	factory.SetDefault(defaultCreator)

	tags := []Tag{{GroupCode: 1, Value: "test"}}
	entity, err := factory.LoadFromTags("UNKNOWN", tags)
	if err != nil {
		t.Errorf("LoadFromTags failed: %v", err)
	}
	if entity == nil {
		t.Error("LoadFromTags with default should return non-nil entity")
	}
}

func TestRegisterGlobal(t *testing.T) {
	RegisterGlobal("GLOBAL", &mockCreator{name: "GLOBAL"})

	if !DefaultFactory.IsRegistered("GLOBAL") {
		t.Error("Global registration should work")
	}
}

func TestCreateGlobal(t *testing.T) {
	RegisterGlobal("GLOBAL", &mockCreator{name: "GLOBAL"})

	entity, err := CreateGlobal("GLOBAL", nil)
	if err != nil {
		t.Errorf("CreateGlobal failed: %v", err)
	}
	if entity == nil {
		t.Error("CreateGlobal should return non-nil entity")
	}
}

func TestLoadFromTagsGlobal(t *testing.T) {
	RegisterGlobal("GLOBALTAGS", &mockCreator{name: "GLOBALTAGS"})

	tags := []Tag{{GroupCode: 1, Value: "test"}}
	entity, err := LoadFromTagsGlobal("GLOBALTAGS", tags)
	if err != nil {
		t.Errorf("LoadFromTagsGlobal failed: %v", err)
	}
	if entity == nil {
		t.Error("LoadFromTagsGlobal should return non-nil entity")
	}
}

func TestFactoryCreateCaseInsensitive(t *testing.T) {
	factory := NewFactory()
	creator := &mockCreator{name: "MIXED"}

	factory.Register("MixedCase", creator)

	entity1, err := factory.Create("mixedcase", nil)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	if entity1 == nil {
		t.Error("Create should work with lowercase")
	}

	entity2, err := factory.Create("MIXEDCASE", nil)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	if entity2 == nil {
		t.Error("Create should work with uppercase")
	}
}
