package entity

// VersionGate documents a group code that only exists in DXF versions at or
// after MinVersion.
type VersionGate struct {
	Code       int
	MinVersion string
	Note       string
}

// AttributeConformance is the per-type attribute matrix row. It records the
// reference used to derive the documented group codes, the codes this package
// exposes as typed fields, defaults that must be emitted when a field is unset,
// and version gates for fields introduced after a save target.
type AttributeConformance struct {
	Type         string
	Reference    string
	Codes        []int
	Defaults     map[int]string
	VersionGates []VersionGate
	Golden       string
}

// EntityConformance is the entity attribute matrix. Every registered entity
// type must have a row and a golden fixture name; a missing row or fixture is
// a test failure.
var EntityConformance = map[string]AttributeConformance{
	"3DFACE":     conformance("ezdxf.entities.Face3d", "3dface.dxf", 10, 11, 12, 13, 70),
	"3DSOLID":    conformance("ezdxf.entities.Solid3d", "3dsolid.dxf", 70, 1, 10, 20, 30),
	"ACAD_TABLE": conformance("ezdxf.entities.AcadTable", "acad_table_simple.dxf", 1, 2, 11, 21, 31, 90, 91, 92, 140, 141),
	"ARC":        conformance("ezdxf.entities.Arc", "arc.dxf", 10, 20, 30, 40, 50, 51),
	"ARC_DIMENSION": func() AttributeConformance {
		row := conformance("ezdxf.entities.ArcDimension", "dimension.dxf", 2, 3, 10, 11, 13, 14, 15, 16, 40, 50, 51)
		row.VersionGates = []VersionGate{{Code: 560, MinVersion: "AC1021", Note: "dimension text rotation"}}
		return row
	}(),
	"ATTDEF": conformance("ezdxf.entities.AttDef", "attdef.dxf", 1, 2, 3, 10, 20, 30, 40, 70, 72, 73),
	"ATTRIB": conformance("ezdxf.entities.Attrib", "attrib.dxf", 1, 2, 10, 20, 30, 40, 70, 72, 73),
	"BODY":   conformance("ezdxf.entities.Body", "body.dxf", 70, 1, 10, 20, 30, 40, 41, 42, 43),
	"CIRCLE": conformance("ezdxf.entities.Circle", "circle.dxf", 10, 20, 30, 40),
	"DIMENSION": func() AttributeConformance {
		row := conformance("ezdxf.entities.Dimension", "dimension.dxf", 2, 3, 10, 11, 13, 14, 15, 16, 40, 50, 51, 70)
		row.VersionGates = []VersionGate{{Code: 560, MinVersion: "AC1021", Note: "dimension text rotation"}}
		return row
	}(),
	"ELLIPSE": conformance("ezdxf.entities.Ellipse", "ellipse.dxf", 10, 20, 30, 11, 21, 31, 40, 41, 42),
	"EXTRUDED_SURFACE": func() AttributeConformance {
		row := conformance("ezdxf.entities.ExtrudedSurface", "extruded_surface.dxf", 70, 10, 20, 30, 40, 41, 42)
		row.VersionGates = []VersionGate{{Code: 70, MinVersion: "AC1021", Note: "surface modeling version"}}
		return row
	}(),
	"FIELDLIST": conformance("ezdxf.entities.FieldList", "fieldlist.dxf", 1, 2, 90, 310),
	"GEODATA":   conformance("ezdxf.entities.GeoData", "geodata.dxf", 70, 1, 10, 20, 30, 40, 41),
	"HATCH":     conformance("ezdxf.entities.Hatch", "hatches_1.dxf", 2, 10, 20, 30, 41, 52, 70, 71, 75, 76, 78, 91, 92, 93, 97, 98, 450, 451, 452, 460, 461, 462, 463, 470),
	"HELIX":     conformance("ezdxf.entities.Helix", "helix.dxf", 10, 20, 30, 11, 21, 31, 40, 41, 42, 90),
	"IDBUFFER":  conformance("ezdxf.entities.IDBuffer", "idbuffer.dxf", 1, 2, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99),
	"IMAGE":     conformance("ezdxf.entities.Image", "image.dxf", 10, 20, 30, 11, 21, 31, 12, 22, 32, 13, 23, 33, 40, 41, 50, 70, 280, 281, 282, 283, 340),
	"INSERT":    conformance("ezdxf.entities.Insert", "insert.dxf", 2, 10, 20, 30, 41, 42, 43, 44, 45, 50, 66, 70, 71, 210, 220, 230),
	"LARGE_RADIAL_DIMENSION": func() AttributeConformance {
		row := conformance("ezdxf.entities.LargeRadialDimension", "dimension.dxf", 2, 3, 10, 11, 13, 14, 15, 16, 40, 50, 51)
		row.VersionGates = []VersionGate{{Code: 560, MinVersion: "AC1021", Note: "dimension text rotation"}}
		return row
	}(),
	"LAYOUT": conformance("ezdxf.entities.Layout", "layout.dxf", 1, 70, 71, 10, 20, 11, 21, 12, 22, 14, 24, 15, 25, 146, 13, 23, 16, 26, 36),
	"LEADER": conformance("ezdxf.entities.Leader", "leader.dxf", 3, 10, 20, 30, 40, 41, 71, 72, 73, 74, 75, 76, 77, 210, 211, 212, 213, 214, 215, 216, 217, 290, 291, 340),
	"LINE":   conformance("ezdxf.entities.Line", "line.dxf", 10, 20, 30, 11, 21, 31, 39, 210, 220, 230),
	"LOFTED_SURFACE": func() AttributeConformance {
		row := conformance("ezdxf.entities.LoftedSurface", "lofted_surface.dxf", 70, 10, 20, 30, 40, 41, 42)
		row.VersionGates = []VersionGate{{Code: 70, MinVersion: "AC1021", Note: "surface modeling version"}}
		return row
	}(),
	"LWPOLYLINE":        conformance("ezdxf.entities.LWPolyline", "lwpolyline.dxf", 10, 20, 40, 41, 42, 70, 90, 43, 38, 39, 210, 220, 230),
	"MESH":              conformance("ezdxf.entities.Mesh", "mesh.dxf", 71, 72, 91, 92, 93, 94, 95, 96, 97, 98, 10, 20, 30),
	"MESHVERTEXCACHE":   conformance("ezdxf.entities.MeshVertexCache", "mesh_vertex_cache.dxf", 70, 10, 20, 30),
	"MESH_VERTEX_CACHE": conformance("ezdxf.entities.MeshVertexCache", "mesh_vertex_cache.dxf", 70, 10, 20, 30),
	"MLEADER":           conformance("ezdxf.entities.MultiLeader", "mleader.dxf", 1, 2, 3, 10, 20, 30, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 170, 171, 172, 173, 174, 175, 176, 177, 290, 291, 292, 293, 294, 295, 296, 297, 298, 299, 300, 301, 302, 303, 304, 305, 306, 307, 308, 309, 310, 311, 312, 313, 314, 315, 316, 317, 318, 319, 320, 330, 340, 341, 342, 343, 344, 345, 346, 347, 348, 349),
	"MLINE":             conformance("ezdxf.entities.MLine", "mline.dxf", 2, 3, 10, 20, 30, 40, 70, 71, 72, 73, 74, 340, 210, 220, 230),
	"MPOLYGON":          conformance("ezdxf.entities.MPolygon", "mpolygon.dxf", 2, 10, 20, 30, 41, 52, 70, 71, 72, 73, 74, 75, 76, 77, 78, 91, 92, 93, 97, 98, 450, 451, 452, 460, 461, 462, 463),
	"MTEXT":             conformance("ezdxf.entities.MText", "mtext_complex_inline_codes.dxf", 1, 3, 7, 10, 20, 30, 11, 21, 31, 40, 41, 42, 43, 44, 45, 50, 71, 72, 73, 90, 210, 220, 230),
	"MULTILEADER":       conformance("ezdxf.entities.MultiLeader", "mleader.dxf", 1, 2, 3, 10, 20, 30, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 170, 171, 172, 173, 174, 175, 176, 177, 290, 291, 292, 293, 294, 295, 296, 297, 298, 299, 300, 301, 302, 303, 304, 305, 306, 307, 308, 309, 310, 311, 312, 313, 314, 315, 316, 317, 318, 319, 320, 330, 340, 341, 342, 343, 344, 345, 346, 347, 348, 349),
	"PLANESURFACE":      conformance("ezdxf.entities.PlaneSurface", "planesurface.dxf", 70, 10, 20, 30, 40, 41, 42),
	"PLOTSETTINGS":      conformance("ezdxf.entities.PlotSettings", "plotsettings.dxf", 1, 2, 4, 6, 7, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 72, 73, 74, 75, 76, 77, 78, 147, 148, 149, 340, 390),
	"POINT":             conformance("ezdxf.entities.Point", "point.dxf", 10, 20, 30, 39, 50, 210, 220, 230),
	"POLYLINE":          conformance("ezdxf.entities.Polyline", "polyline.dxf", 10, 20, 30, 39, 40, 41, 70, 71, 72, 73, 74, 75, 210, 220, 230),
	"RASTERVARIABLES":   conformance("ezdxf.entities.RasterVariables", "rastervariables.dxf", 70, 71, 72, 1),
	"RAY":               conformance("ezdxf.entities.Ray", "ray.dxf", 10, 20, 30, 11, 21, 31),
	"REGION":            conformance("ezdxf.entities.Region", "region.dxf", 70, 1, 10, 20, 30, 40, 41, 42),
	"REVOLVED_SURFACE": func() AttributeConformance {
		row := conformance("ezdxf.entities.RevolvedSurface", "revolved_surface.dxf", 70, 10, 20, 30, 40, 41, 42)
		row.VersionGates = []VersionGate{{Code: 70, MinVersion: "AC1021", Note: "surface modeling version"}}
		return row
	}(),
	"SHAPE":          conformance("ezdxf.entities.Shape", "shape.dxf", 2, 10, 20, 30, 39, 40, 41, 50, 51, 70, 210, 220, 230),
	"SOLID":          conformance("ezdxf.entities.Solid", "solid.dxf", 10, 11, 12, 13, 20, 21, 22, 23, 30, 31, 32, 33, 39, 210, 220, 230),
	"SPATIAL_FILTER": conformance("ezdxf.entities.SpatialFilter", "spatial_filter.dxf", 70, 1, 10, 20, 30, 11, 21, 31, 40, 41, 71, 72),
	"SPLINE":         conformance("ezdxf.entities.Spline", "spline.dxf", 10, 20, 30, 40, 41, 42, 70, 71, 72, 73, 74),
	"SWEPT_SURFACE": func() AttributeConformance {
		row := conformance("ezdxf.entities.SweptSurface", "swept_surface.dxf", 70, 10, 20, 30, 40, 41, 42)
		row.VersionGates = []VersionGate{{Code: 70, MinVersion: "AC1021", Note: "surface modeling version"}}
		return row
	}(),
	"TEXT":             conformance("ezdxf.entities.Text", "text.dxf", 1, 7, 10, 20, 30, 11, 21, 31, 39, 40, 41, 50, 51, 71, 72, 73, 210, 220, 230),
	"TOLERANCE":        conformance("ezdxf.entities.Tolerance", "tolerance.dxf", 1, 3, 10, 11, 40, 50, 71, 72, 73),
	"TRACE":            conformance("ezdxf.entities.Trace", "trace.dxf", 10, 11, 12, 13, 20, 21, 22, 23, 30, 31, 32, 33, 39, 210, 220, 230),
	"UNDERLAY":         conformance("ezdxf.entities.Underlay", "underlay.dxf", 10, 20, 30, 11, 21, 31, 12, 22, 32, 13, 23, 33, 40, 41, 50, 70, 280, 281, 282, 340),
	"VERTEX":           conformance("ezdxf.entities.Vertex", "vertex.dxf", 10, 20, 30, 40, 41, 42, 70, 71, 72, 73, 74, 210, 220, 230),
	"VIEWPORT":         conformance("ezdxf.entities.Viewport", "viewport.dxf", 10, 20, 30, 40, 41, 68, 69, 12, 22, 13, 23, 14, 24, 15, 25, 16, 26, 17, 27, 42, 43, 44, 45, 50, 51, 72, 90, 340, 1),
	"WIPEOUT":          conformance("ezdxf.entities.Wipeout", "wipeout_door.dxf", 10, 20, 30, 11, 21, 31, 12, 22, 32, 13, 23, 33, 70, 280, 281, 282, 283, 340),
	"WIPEOUTVARIABLES": conformance("ezdxf.entities.WipeoutVariables", "wipeoutvariables.dxf", 70),
	"XLINE":            conformance("ezdxf.entities.XLine", "xline.dxf", 10, 20, 30, 11, 21, 31),
}

func conformance(reference, golden string, codes ...int) AttributeConformance {
	return AttributeConformance{
		Reference: reference,
		Codes:     codes,
		Golden:    golden,
	}
}

// ConformanceFor returns the matrix row for a DXF type name.
func ConformanceFor(dxfType string) (AttributeConformance, bool) {
	row, ok := EntityConformance[dxfType]
	return row, ok
}

// MissingEntityConformance reports registered types without a matrix row or a
// golden fixture name.
func MissingEntityConformance() []string {
	var missing []string
	for _, name := range RegisteredTypes() {
		row, ok := EntityConformance[name]
		if !ok || row.Golden == "" || len(row.Codes) == 0 {
			missing = append(missing, name)
		}
	}
	return missing
}
