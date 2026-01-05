package entity

import (
	"fmt"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/table"
)

// AppData represents an application data entity (R2000+)
type AppData struct {
	*entity

	// Basic properties
	ApplicationName string // 1 - Application name
	RegistryName    string // 2 - Registry name (for Windows registry)
	Data            string // 3 - Data content
	Flags           int    // 70 - AppData flags
}

// NewAppData creates a new AppData entity
func NewAppData() *AppData {
	return &AppData{
		entity: NewEntity(APPDATA),
	}
}

// IsEntity implements Entity interface
func (a *AppData) IsEntity() bool {
	return true
}

// SetApplicationName sets the application name
func (a *AppData) SetApplicationName(name string) {
	a.ApplicationName = name
}

// SetRegistryName sets the registry name
func (a *AppData) SetRegistryName(name string) {
	a.RegistryName = name
}

// SetData sets the data content
func (a *AppData) SetData(data string) {
	a.Data = data
}

// SetFlags sets the AppData flags
func (a *AppData) SetFlags(flags int) {
	a.Flags = flags
}

// GetApplicationName returns the application name
func (a *AppData) GetApplicationName() string {
	return a.ApplicationName
}

// GetRegistryName returns the registry name
func (a *AppData) GetRegistryName() string {
	return a.RegistryName
}

// GetData returns the data content
func (a *AppData) GetData() string {
	return a.Data
}

// GetFlags returns the AppData flags
func (a *AppData) GetFlags() int {
	return a.Flags
}

// Format formats the AppData entity for DXF output
func (a *AppData) Format(f format.Formatter) string {
	// Simple formatting for AppData
	result := f.String(100, "APPDATA")
	result += f.String(1, a.ApplicationName)
	result += f.String(2, a.RegistryName)
	result += f.String(3, a.Data)
	result += f.Int(70, a.Flags)
	result += f.Int(0, 0) // Terminator
	result += f.String(0, "")

	return result
}

// FormatString formats the AppData entity using default formatter
func (a *AppData) FormatString() string {
	return fmt.Sprintf("AppData{Application: %s, Registry: %s, Data: %s, Flags: %d}",
		a.ApplicationName, a.RegistryName, a.Data, a.Flags)
}

// SetLayer sets the layer for AppData entity
func (a *AppData) SetLayer(layer *table.Layer) {
	a.entity.SetLayer(layer)
}

// Layer returns the current layer
func (a *AppData) Layer() *table.Layer {
	return a.entity.Layer()
}

// SetColor sets the color for AppData entity
func (a *AppData) SetColor(color color.ColorNumber) {
	a.entity.SetColor(color)
}

// Color returns the current color
func (a *AppData) Color() color.ColorNumber {
	return a.entity.color
}
