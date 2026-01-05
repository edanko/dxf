package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// AppID represents an APPID (Application ID) table entry
type AppID struct {
	*entity
	Name  string // 2 - application name
	Flags int    // 70 - flags
}

// NewAppID creates a new AppID entity
func NewAppID() *AppID {
	a := &AppID{
		entity: NewEntity(APPID),
		Name:   "",
		Flags:  0,
	}
	return a
}

// NewAppIDWithName creates an AppID with the specified name
func NewAppIDWithName(name string) *AppID {
	a := NewAppID()
	a.Name = name
	return a
}

// IsEntity is for Entity interface.
func (a *AppID) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (a *AppID) Format(f format.Formatter) {
	a.entity.Format(f)
	f.WriteString(100, "AcDbRegAppTableRecord")
	f.WriteString(100, "AcDbRegAppTableRecord")
	f.WriteString(2, a.Name)
	f.WriteInt(70, a.Flags)
}

// BBox returns the bounding box of the entity
func (a *AppID) BBox() ([]float64, []float64) {
	// AppID is a table entry, return empty bounding box
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the AppID
func (a *AppID) Transform(m *math.Matrix44) error {
	// AppID doesn't have geometric data to transform
	return nil
}

// Copy creates a deep copy of the AppID entity
func (a *AppID) Copy() Entity {
	app := NewAppID()
	app.entity = a.entity
	app.Name = a.Name
	app.Flags = a.Flags
	return app
}

// Validate validates the AppID entity
func (a *AppID) Validate() error {
	if a.Name == "" {
		a.Name = "UNKNOWN"
	}
	return nil
}
