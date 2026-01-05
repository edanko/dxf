package interfaces

import "github.com/edanko/dxf/entity"

// Document represents a DXF drawing document for entity binding
type Document interface {
	FileName() string
	AddEntity(e entity.Entity) error
	GetEntity(handle string) (entity.Entity, bool)
}

// Entity represents a DXF entity interface without circular imports
type Entity interface {
	IsEntity() bool
	Handle() string
	DXFType() string
	GetDocument() Document
	SetDocument(Document)
	GetAttributes() map[string]interface{}
	LoadAttributes(attribs map[string]interface{}) error
}
