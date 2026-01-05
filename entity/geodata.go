package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// GeoData represents a GEODATA entity for geographic data
type GeoData struct {
	*entity
	Version              int       // 90 - version (1=R2009, 2=R2010)
	BlockRecordHandle    string    // 330 - handle to host block table record
	CoordinateType       int       // 70 - 0=unknown, 1=local grid, 2=projected, 3=geographic
	DesignPoint          math.Vec3 // 10, 20, 30 - design point in WCS
	ReferencePoint       math.Vec3 // 11, 21, 31 - reference point in coordinate system
	HorizontalUnitScale  float64   // 40 - horizontal unit scale to meters
	HorizontalUnits      int       // 91 - horizontal units enumeration
	VerticalUnitScale    float64   // 41 - vertical unit scale to meters
	VerticalUnits        int       // 92 - vertical units enumeration
	UpDirection          math.Vec3 // 210, 220, 230 - up direction
	NorthDirection       math.Vec2 // 12, 22 - north direction (2D)
	Easting              float64   // 13 - easting
	Northing             float64   // 14 - northing
	Elevation            float64   // 15 - elevation
	UnitScale            float64   // 42 - coordinate unit scale
	CoordinatesProjected bool      // 73 - projected coordinates flag
	Obsolete             float64   // 44 - obsolete
	Translation          math.Vec3 // 46, 47, 48 - translation offset
	RScript              string    // 1 - registration script
	GeoRSS               string    // 16 - GeoRSS tag
}

// NewGeoData creates a new GeoData entity
func NewGeoData() *GeoData {
	g := &GeoData{
		entity:               NewEntity(GEODATA),
		Version:              2,
		BlockRecordHandle:    "",
		CoordinateType:       3,
		DesignPoint:          math.Vec3{0, 0, 0},
		ReferencePoint:       math.Vec3{0, 0, 0},
		HorizontalUnitScale:  1.0,
		HorizontalUnits:      1,
		VerticalUnitScale:    1.0,
		VerticalUnits:        1,
		UpDirection:          math.Vec3{0, 0, 1},
		NorthDirection:       math.Vec2{0, 1},
		Easting:              0,
		Northing:             0,
		Elevation:            0,
		UnitScale:            1.0,
		CoordinatesProjected: false,
		Obsolete:             0,
		Translation:          math.Vec3{0, 0, 0},
		RScript:              "",
		GeoRSS:               "",
	}
	return g
}

// IsEntity is for Entity interface.
func (g *GeoData) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (g *GeoData) Format(f format.Formatter) {
	g.entity.Format(f)
	f.WriteString(100, "AcDbGeoData")
	f.WriteInt(90, g.Version)
	f.WriteString(330, g.BlockRecordHandle)
	f.WriteInt(70, g.CoordinateType)
	f.WriteFloat(10, g.DesignPoint.X())
	f.WriteFloat(20, g.DesignPoint.Y())
	f.WriteFloat(30, g.DesignPoint.Z())
	f.WriteFloat(11, g.ReferencePoint.X())
	f.WriteFloat(21, g.ReferencePoint.Y())
	f.WriteFloat(31, g.ReferencePoint.Z())
	f.WriteFloat(40, g.HorizontalUnitScale)
	f.WriteInt(91, g.HorizontalUnits)
	f.WriteFloat(41, g.VerticalUnitScale)
	f.WriteInt(92, g.VerticalUnits)
	f.WriteFloat(210, g.UpDirection.X())
	f.WriteFloat(220, g.UpDirection.Y())
	f.WriteFloat(230, g.UpDirection.Z())
	f.WriteFloat(12, g.NorthDirection.X())
	f.WriteFloat(22, g.NorthDirection.Y())
	f.WriteFloat(13, g.Easting)
	f.WriteFloat(14, g.Northing)
	f.WriteFloat(15, g.Elevation)
	f.WriteFloat(42, g.UnitScale)
	if g.CoordinatesProjected {
		f.WriteInt(73, 1)
	}
	f.WriteFloat(44, g.Obsolete)
	f.WriteFloat(46, g.Translation.X())
	f.WriteFloat(47, g.Translation.Y())
	f.WriteFloat(48, g.Translation.Z())
	f.WriteString(1, g.RScript)
	f.WriteString(16, g.GeoRSS)
}

// BBox returns the bounding box of the entity
func (g *GeoData) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the GeoData
func (g *GeoData) Transform(m *math.Matrix44) error {
	transformVec3 := func(v math.Vec3) math.Vec3 {
		x := v[0]*m[0] + v[1]*m[1] + v[2]*m[2] + m[3]
		y := v[0]*m[4] + v[1]*m[5] + v[2]*m[6] + m[7]
		z := v[0]*m[8] + v[1]*m[9] + v[2]*m[10] + m[11]
		return math.Vec3{x, y, z}
	}
	g.DesignPoint = transformVec3(g.DesignPoint)
	g.ReferencePoint = transformVec3(g.ReferencePoint)
	g.UpDirection = transformVec3(g.UpDirection)
	g.Translation = transformVec3(g.Translation)
	return nil
}

// Copy creates a deep copy of the GeoData entity
func (g *GeoData) Copy() Entity {
	geo := NewGeoData()
	geo.entity = g.entity
	geo.Version = g.Version
	geo.BlockRecordHandle = g.BlockRecordHandle
	geo.CoordinateType = g.CoordinateType
	geo.DesignPoint = g.DesignPoint
	geo.ReferencePoint = g.ReferencePoint
	geo.HorizontalUnitScale = g.HorizontalUnitScale
	geo.HorizontalUnits = g.HorizontalUnits
	geo.VerticalUnitScale = g.VerticalUnitScale
	geo.VerticalUnits = g.VerticalUnits
	geo.UpDirection = g.UpDirection
	geo.NorthDirection = g.NorthDirection
	geo.Easting = g.Easting
	geo.Northing = g.Northing
	geo.Elevation = g.Elevation
	geo.UnitScale = g.UnitScale
	geo.CoordinatesProjected = g.CoordinatesProjected
	geo.Obsolete = g.Obsolete
	geo.Translation = g.Translation
	geo.RScript = g.RScript
	geo.GeoRSS = g.GeoRSS
	return geo
}

// Validate validates the GeoData entity
func (g *GeoData) Validate() error {
	if g.Version < 1 || g.Version > 2 {
		g.Version = 2
	}
	if g.CoordinateType < 0 || g.CoordinateType > 3 {
		g.CoordinateType = 3
	}
	if g.HorizontalUnitScale == 0 {
		g.HorizontalUnitScale = 1.0
	}
	if g.VerticalUnitScale == 0 {
		g.VerticalUnitScale = 1.0
	}
	return nil
}

// SetGeographic sets coordinate type to geographic (lat/lon)
func (g *GeoData) SetGeographic() {
	g.CoordinateType = 3
}

// SetProjected sets coordinate type to projected
func (g *GeoData) SetProjected() {
	g.CoordinateType = 2
}

// SetLocalGrid sets coordinate type to local grid
func (g *GeoData) SetLocalGrid() {
	g.CoordinateType = 1
}
