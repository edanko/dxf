package factory

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

// Entity interface for factory creation
type Entity interface {
	IsEntity() bool
	Format(Formatter)
	SetBlockRecord(handle.Handler)
	Layer() *table.Layer
	SetLayer(*table.Layer)
	SetLtscale(float64)
	BBox() ([]float64, []float64)
	SetColor(color.ColorNumber)
	handle.Handler

	// Factory support
	DXFType() string
	SetDocument(*drawing.Drawing)
	GetDocument() *drawing.Drawing
	LoadAttributes(attribs map[string]interface{}) error
	GetAttributes() map[string]interface{}
}

// Tag represents a DXF tag
type Tag struct {
	GroupCode int
	Value     interface{}
}

// Formatter interface for entity output
type Formatter interface {
	WriteString(groupCode int, value string)
	WriteInt(groupCode int, value int)
	WriteFloat(groupCode int, value float64)
	WriteBool(groupCode int, value bool)
	Write3DPoint(groupCode int, x, y, z float64)
}

// EntityCreator interface for creating entities
type EntityCreator interface {
	New() Entity
	NewWithAttribs(attribs map[string]interface{}) Entity
	LoadFromTags(tags []Tag) (Entity, error)
	DXFType() string
}

// Factory manages entity creation and registration
type Factory struct {
	entityTypes map[string]EntityCreator
	defaultType EntityCreator
}

// Global factory instance
var DefaultFactory = NewFactory()

// NewFactory creates a new factory instance
func NewFactory() *Factory {
	return &Factory{
		entityTypes: make(map[string]EntityCreator),
	}
}

// Register registers an entity creator for a DXF type
func (f *Factory) Register(dxfType string, creator EntityCreator) {
	f.entityTypes[strings.ToUpper(dxfType)] = creator
}

// Create creates a new entity of the specified type with attributes
func (f *Factory) Create(dxfType string, attribs map[string]interface{}) (Entity, error) {
	creator, exists := f.entityTypes[strings.ToUpper(dxfType)]
	if !exists {
		if f.defaultType != nil {
			return f.defaultType.NewWithAttribs(attribs), nil
		}
		return nil, fmt.Errorf("unknown entity type: %s", dxfType)
	}
	return creator.NewWithAttribs(attribs), nil
}

// CreateNew creates a new entity of the specified type without attributes
func (f *Factory) CreateNew(dxfType string) (Entity, error) {
	creator, exists := f.entityTypes[strings.ToUpper(dxfType)]
	if !exists {
		if f.defaultType != nil {
			return f.defaultType.New(), nil
		}
		return nil, fmt.Errorf("unknown entity type: %s", dxfType)
	}
	return creator.New(), nil
}

// LoadFromTags creates an entity from parsed DXF tags
func (f *Factory) LoadFromTags(dxfType string, tags []Tag) (Entity, error) {
	creator, exists := f.entityTypes[strings.ToUpper(dxfType)]
	if !exists {
		if f.defaultType != nil {
			return f.defaultType.LoadFromTags(tags)
		}
		return nil, fmt.Errorf("unknown entity type: %s", dxfType)
	}
	return creator.LoadFromTags(tags)
}

// GetRegisteredTypes returns a list of all registered entity types
func (f *Factory) GetRegisteredTypes() []string {
	types := make([]string, 0, len(f.entityTypes))
	for t := range f.entityTypes {
		types = append(types, t)
	}
	return types
}

// IsRegistered checks if an entity type is registered
func (f *Factory) IsRegistered(dxfType string) bool {
	_, exists := f.entityTypes[strings.ToUpper(dxfType)]
	return exists
}

// SetDefault sets the default entity creator for unknown types
func (f *Factory) SetDefault(creator EntityCreator) {
	f.defaultType = creator
}

// RegisterGlobal registers an entity type with the default factory
func RegisterGlobal(dxfType string, creator EntityCreator) {
	DefaultFactory.Register(dxfType, creator)
}

// CreateGlobal creates an entity using the default factory
func CreateGlobal(dxfType string, attribs map[string]interface{}) (Entity, error) {
	return DefaultFactory.Create(dxfType, attribs)
}

// LoadFromTagsGlobal creates an entity from tags using the default factory
func LoadFromTagsGlobal(dxfType string, tags []Tag) (Entity, error) {
	return DefaultFactory.LoadFromTags(dxfType, tags)
}
