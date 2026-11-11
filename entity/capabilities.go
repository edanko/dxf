package entity

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/xdata"
)

// Capability interfaces for optional entity operations. Callers type-assert:
// a type that does not implement an interface simply does not support the
// operation; there are no no-op implementations hiding failures.

// Transformer applies an affine transform to an entity's geometry in place.
type Transformer interface {
	Transform(m math.Matrix44) error
}

// Copier produces a deep copy of an entity.
type Copier interface {
	Copy() Entity
}

// Bounder computes an axis-aligned bounding box (min, max corners).
type Bounder interface {
	BBox() ([]float64, []float64)
}

// Exploder decomposes an entity into constituent entities (INSERT -> expanded
// content, POLYLINE -> LINE segments, DIMENSION -> geometry, ...).
type Exploder interface {
	Explode() ([]Entity, error)
}

// SubentityOwner exposes child records of composite entities (POLYLINE
// vertices, INSERT attribs, MLEADER leader lines, dimensions).
type SubentityOwner interface {
	Subentities() []Entity
}

// Validator reports entity problems for audit; an empty slice means valid.
type Validator interface {
	Validate() []string
}

// LayerNamer exposes the resolved or recorded layer name without requiring the
// layer table to be bound. It is useful for semantic inventories and audits.
type LayerNamer interface {
	LayerName() string
}

// LayerNameSetter updates the recorded layer name.
type LayerNameSetter interface {
	SetLayerName(string)
}

// Unwrapper exposes the concrete entity behind an adapter.
type Unwrapper interface {
	Unwrap() Entity
}

// RawTagReader exposes retained DXF group codes.
type RawTagReader interface {
	RawTags() []TagPair
}

// RawTagWriter replaces retained DXF group codes.
type RawTagWriter interface {
	SetRawTags([]TagPair)
}

// HandleStringSetter sets a fixed entity handle.
type HandleStringSetter interface {
	SetHandleString(string)
}

// HandleGeneratorSetter installs a handle generator.
type HandleGeneratorSetter interface {
	SetHandle(*handle.HandleGenerator)
}

// OwnerSetter sets the owning entity or table handle.
type OwnerSetter interface {
	SetOwner(handle.Handler)
}

// OwnerGetter returns the owning handle.
type OwnerGetter interface {
	Owner() handle.Handler
}

// ColorRGBReader exposes a true-color value.
type ColorRGBReader interface {
	ColorRGB() int
}

// ColorReader exposes a color number.
type ColorReader interface {
	Color() color.ColorNumber
}

// ColorSetter updates a color number.
type ColorSetter interface {
	SetColor(color.ColorNumber)
}

// RGBReader exposes an 8-bit RGB triple.
type RGBReader interface {
	RGB() (uint8, uint8, uint8)
}

// ColorRGBSetter updates a true-color value.
type ColorRGBSetter interface {
	SetColorRGB(int)
}

// TransparencyReader exposes alpha transparency.
type TransparencyReader interface {
	Transparency() float64
}

// TransparencySetter updates alpha transparency.
type TransparencySetter interface {
	SetTransparency(float64)
}

// InvisibleReader reports entity visibility.
type InvisibleReader interface {
	Invisible() bool
}

// InvisibleSetter updates entity visibility.
type InvisibleSetter interface {
	SetInvisible(bool)
}

// LineWeightReader exposes a lineweight value.
type LineWeightReader interface {
	LineWeight() int
}

// LineWeightSetter updates a lineweight value.
type LineWeightSetter interface {
	SetLineWeight(int)
}

// LinetypeReader exposes a linetype name.
type LinetypeReader interface {
	Linetype() string
}

// LinetypeSetter updates a linetype name.
type LinetypeSetter interface {
	SetLinetype(string)
}

// BlockNameReader exposes an INSERT block name.
type BlockNameReader interface {
	BlockName() string
}

// StyleReader exposes an entity style name.
type StyleReader interface {
	GetStyle() string
}

// XDataAppIDReader exposes XDATA application identifiers.
type XDataAppIDReader interface {
	XDataAppIDs() []string
}

// XDataObjectReader exposes attached XDATA.
type XDataObjectReader interface {
	GetXDataObject() *xdata.XData
}

// XDataObjectSetter attaches XDATA.
type XDataObjectSetter interface {
	SetXDataObject(*xdata.XData)
}

// ReactorReader exposes persistent reactor handles.
type ReactorReader interface {
	ReactorHandles() []string
}

// ReactorSetter updates persistent reactor handles.
type ReactorSetter interface {
	SetReactorHandles([]string)
}

// ExtDictReader exposes an extension dictionary handle.
type ExtDictReader interface {
	ExtDictHandle() string
}

// ExtDictSetter updates the extension dictionary handle.
type ExtDictSetter interface {
	SetExtDictHandle(string)
}

// RecordTypeNamer overrides the wire type name of a retained record.
type RecordTypeNamer interface {
	SetRecordTypeName(string)
}

// ThicknessSetter updates entity thickness.
type ThicknessSetter interface {
	SetThickness(float64)
}

// TextReader exposes single-line text content.
type TextReader interface {
	Text() string
}

// ValueReader exposes a generic string value.
type ValueReader interface {
	Value() string
}

// RotationReaderSetter reads and writes a rotation angle in degrees.
type RotationReaderSetter interface {
	GetRotation() float64
	SetRotation(float64)
}

// RawTagReadWriter reads and replaces the retained raw tag stream.
type RawTagReadWriter interface {
	RawTagReader
	RawTagWriter
}

// ACISMetadataReader exposes shallow ACIS payload metadata.
type ACISMetadataReader interface {
	ACISInfo() ACISMetadata
}

// ACISBinaryReader reports whether an entity carries binary SAB data.
type ACISBinaryReader interface {
	HasBinaryACIS() bool
}

// ACISBinaryDataSetter attaches binary SAB data to an ACIS entity.
type ACISBinaryDataSetter interface {
	SetACISBinaryData([]byte)
}

// ACISBinaryDataReader returns the retained binary SAB payload.
type ACISBinaryDataReader interface {
	ACISBinaryData() []byte
}
