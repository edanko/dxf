package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// Material represents a MATERIAL entity for rendering properties
type Material struct {
	*entity
	Name                     string  // 1 - material name
	Description              string  // 2 - material description
	AmbientColorMethod       int     // 70 - 0=use current, 1=override
	AmbientColorFactor       float64 // 40 - 0.0 to 1.0
	AmbientColorValue        int     // 90 - AcCmEntityColor
	DiffuseColorMethod       int     // 71 - 0=use current, 1=override
	DiffuseColorFactor       float64 // 41 - 0.0 to 1.0
	DiffuseColorValue        int     // 91 - AcCmEntityColor
	DiffuseMapBlendFactor    float64 // 42 - 0.0 to 1.0
	DiffuseMapSource         int     // 72 - 0=scene, 1=image file
	DiffuseMapFileName       string  // 3
	DiffuseMapProjection     int     // 73 - 1=Planar, 2=Box, 3=Cylinder, 4=Sphere
	DiffuseMapTiling         int     // 74 - 1=Tile, 2=Crop, 3=Clamp
	DiffuseMapAutoTransform  int     // 75 - bitfield
	SpecularGlossFactor      float64 // 44 - 0.0 to 1.0
	SpecularColorMethod      int     // 73 - 0=use current, 1=override
	SpecularColorFactor      float64 // 45 - 0.0 to 1.0
	SpecularColorValue       int     // 92 - AcCmEntityColor
	SpecularMapBlendFactor   float64 // 46 - 0.0 to 1.0
	SpecularMapSource        int     // 77 - 0=scene, 1=image file
	SpecularMapFileName      string  // 4
	SpecularMapProjection    int     // 78 - 1=Planar, 2=Box, 3=Cylinder, 4=Sphere
	SpecularMapTiling        int     // 79 - 1=Tile, 2=Crop, 3=Clamp
	SpecularMapAutoTransform int     // 80 - bitfield
	ReflectionMapBlendFactor float64 // 47 - 0.0 to 1.0
	ReflectionMapSource      int     // 82 - 0=scene, 1=image file
	ReflectionMapFileName    string  // 5
	ReflectionMapProjection  int     // 83 - 1=Planar, 2=Box, 3=Cylinder, 4=Sphere
	ReflectionMapTiling      int     // 84 - 1=Tile, 2=Crop, 3=Clamp
	OpacityMapBlendFactor    float64 // 48 - 0.0 to 1.0
	OpacityMapSource         int     // 85 - 0=scene, 1=image file
	OpacityMapFileName       string  // 6
	OpacityMapProjection     int     // 86 - 1=Planar, 2=Box, 3=Cylinder, 4=Sphere
	OpacityMapTiling         int     // 87 - 1=Tile, 2=Crop, 3=Clamp
	BumpMapBlendFactor       float64 // 49 - 0.0 to 1.0
	BumpMapSource            int     // 88 - 0=scene, 1=image file
	BumpMapFileName          string  // 7
	BumpMapProjection        int     // 89 - 1=Planar, 2=Box, 3=Cylinder, 4=Sphere
	BumpMapTiling            int     // 90 - 1=Tile, 2=Crop, 3=Clamp
}

// NewMaterial creates a new Material entity
func NewMaterial() *Material {
	m := &Material{
		entity:                   NewEntity(MATERIAL),
		Name:                     "",
		Description:              "",
		AmbientColorMethod:       0,
		AmbientColorFactor:       1.0,
		AmbientColorValue:        0,
		DiffuseColorMethod:       0,
		DiffuseColorFactor:       1.0,
		DiffuseColorValue:        -1023410177,
		DiffuseMapBlendFactor:    1.0,
		DiffuseMapSource:         1,
		DiffuseMapFileName:       "",
		DiffuseMapProjection:     1,
		DiffuseMapTiling:         1,
		DiffuseMapAutoTransform:  1,
		SpecularGlossFactor:      0.5,
		SpecularColorMethod:      0,
		SpecularColorFactor:      1.0,
		SpecularColorValue:       0,
		SpecularMapBlendFactor:   1.0,
		SpecularMapSource:        1,
		SpecularMapFileName:      "",
		SpecularMapProjection:    1,
		SpecularMapTiling:        1,
		SpecularMapAutoTransform: 1,
		ReflectionMapBlendFactor: 1.0,
		ReflectionMapSource:      1,
		ReflectionMapFileName:    "",
		ReflectionMapProjection:  1,
		ReflectionMapTiling:      1,
		OpacityMapBlendFactor:    1.0,
		OpacityMapSource:         1,
		OpacityMapFileName:       "",
		OpacityMapProjection:     1,
		OpacityMapTiling:         1,
		BumpMapBlendFactor:       1.0,
		BumpMapSource:            1,
		BumpMapFileName:          "",
		BumpMapProjection:        1,
		BumpMapTiling:            1,
	}
	return m
}

// IsEntity is for Entity interface.
func (m *Material) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (m *Material) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbMaterial")
	f.WriteString(1, m.Name)
	f.WriteString(2, m.Description)
	f.WriteInt(70, m.AmbientColorMethod)
	f.WriteFloat(40, m.AmbientColorFactor)
	f.WriteInt(90, m.AmbientColorValue)
	f.WriteInt(71, m.DiffuseColorMethod)
	f.WriteFloat(41, m.DiffuseColorFactor)
	f.WriteInt(91, m.DiffuseColorValue)
	f.WriteFloat(42, m.DiffuseMapBlendFactor)
	f.WriteInt(72, m.DiffuseMapSource)
	f.WriteString(3, m.DiffuseMapFileName)
	f.WriteInt(73, m.DiffuseMapProjection)
	f.WriteInt(74, m.DiffuseMapTiling)
	f.WriteInt(75, m.DiffuseMapAutoTransform)
	f.WriteFloat(44, m.SpecularGlossFactor)
	f.WriteInt(73, m.SpecularColorMethod)
	f.WriteFloat(45, m.SpecularColorFactor)
	f.WriteInt(92, m.SpecularColorValue)
	f.WriteFloat(46, m.SpecularMapBlendFactor)
	f.WriteInt(77, m.SpecularMapSource)
	f.WriteString(4, m.SpecularMapFileName)
	f.WriteInt(78, m.SpecularMapProjection)
	f.WriteInt(79, m.SpecularMapTiling)
	f.WriteInt(80, m.SpecularMapAutoTransform)
	f.WriteFloat(47, m.ReflectionMapBlendFactor)
	f.WriteInt(82, m.ReflectionMapSource)
	f.WriteString(5, m.ReflectionMapFileName)
	f.WriteInt(83, m.ReflectionMapProjection)
	f.WriteInt(84, m.ReflectionMapTiling)
	f.WriteFloat(48, m.OpacityMapBlendFactor)
	f.WriteInt(85, m.OpacityMapSource)
	f.WriteString(6, m.OpacityMapFileName)
	f.WriteInt(86, m.OpacityMapProjection)
	f.WriteInt(87, m.OpacityMapTiling)
	f.WriteFloat(49, m.BumpMapBlendFactor)
	f.WriteInt(88, m.BumpMapSource)
	f.WriteString(7, m.BumpMapFileName)
	f.WriteInt(89, m.BumpMapProjection)
	f.WriteInt(90, m.BumpMapTiling)
}

// BBox returns the bounding box of the entity
func (m *Material) BBox() ([]float64, []float64) {
	// Material is a metadata entity, return empty bounding box
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the Material
func (m *Material) Transform(mt *math.Matrix44) error {
	// Material doesn't have geometric data to transform
	return nil
}

// Copy creates a deep copy of the Material entity
func (m *Material) Copy() Entity {
	mat := NewMaterial()
	mat.entity = m.entity
	mat.Name = m.Name
	mat.Description = m.Description
	mat.AmbientColorMethod = m.AmbientColorMethod
	mat.AmbientColorFactor = m.AmbientColorFactor
	mat.AmbientColorValue = m.AmbientColorValue
	mat.DiffuseColorMethod = m.DiffuseColorMethod
	mat.DiffuseColorFactor = m.DiffuseColorFactor
	mat.DiffuseColorValue = m.DiffuseColorValue
	mat.DiffuseMapBlendFactor = m.DiffuseMapBlendFactor
	mat.DiffuseMapSource = m.DiffuseMapSource
	mat.DiffuseMapFileName = m.DiffuseMapFileName
	mat.DiffuseMapProjection = m.DiffuseMapProjection
	mat.DiffuseMapTiling = m.DiffuseMapTiling
	mat.DiffuseMapAutoTransform = m.DiffuseMapAutoTransform
	mat.SpecularGlossFactor = m.SpecularGlossFactor
	mat.SpecularColorMethod = m.SpecularColorMethod
	mat.SpecularColorFactor = m.SpecularColorFactor
	mat.SpecularColorValue = m.SpecularColorValue
	mat.SpecularMapBlendFactor = m.SpecularMapBlendFactor
	mat.SpecularMapSource = m.SpecularMapSource
	mat.SpecularMapFileName = m.SpecularMapFileName
	mat.SpecularMapProjection = m.SpecularMapProjection
	mat.SpecularMapTiling = m.SpecularMapTiling
	mat.SpecularMapAutoTransform = m.SpecularMapAutoTransform
	mat.ReflectionMapBlendFactor = m.ReflectionMapBlendFactor
	mat.ReflectionMapSource = m.ReflectionMapSource
	mat.ReflectionMapFileName = m.ReflectionMapFileName
	mat.ReflectionMapProjection = m.ReflectionMapProjection
	mat.ReflectionMapTiling = m.ReflectionMapTiling
	mat.OpacityMapBlendFactor = m.OpacityMapBlendFactor
	mat.OpacityMapSource = m.OpacityMapSource
	mat.OpacityMapFileName = m.OpacityMapFileName
	mat.OpacityMapProjection = m.OpacityMapProjection
	mat.OpacityMapTiling = m.OpacityMapTiling
	mat.BumpMapBlendFactor = m.BumpMapBlendFactor
	mat.BumpMapSource = m.BumpMapSource
	mat.BumpMapFileName = m.BumpMapFileName
	mat.BumpMapProjection = m.BumpMapProjection
	mat.BumpMapTiling = m.BumpMapTiling
	return mat
}

// Validate validates the Material entity
func (m *Material) Validate() error {
	// Clamp color factors to valid range
	if m.AmbientColorFactor < 0 || m.AmbientColorFactor > 1 {
		m.AmbientColorFactor = 1.0
	}
	if m.DiffuseColorFactor < 0 || m.DiffuseColorFactor > 1 {
		m.DiffuseColorFactor = 1.0
	}
	if m.SpecularGlossFactor < 0 || m.SpecularGlossFactor > 1 {
		m.SpecularGlossFactor = 0.5
	}
	if m.SpecularColorFactor < 0 || m.SpecularColorFactor > 1 {
		m.SpecularColorFactor = 1.0
	}
	return nil
}
