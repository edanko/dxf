// Constructors for image/underlay OBJECTS-section records and the
// StdObject page accessor. Records are built with their retained tag stream
// so they round-trip through Load and Format unchanged.
package object

import (
	"strconv"
	"strings"

	"github.com/edanko/dxf/lldxf"
)

// NewImageDef builds an IMAGEDEF object describing a raster image file.
func NewImageDef(objHandle, filePath string, width, height, pixelSizeX, pixelSizeY float64) *ImageDef {
	s := &StdObject{
		BaseObject: &BaseObject{},
		typeName:   "IMAGEDEF",
		path:       filePath,
		imageSizeX: width,
		imageSizeY: height,
		pixelSizeX: pixelSizeX,
		pixelSizeY: pixelSizeY,
	}
	s.SetWireType("IMAGEDEF")
	s.BaseObject.handle = objHandle
	retained := lldxf.NewTags()
	retained = retained.Add(lldxf.NewDXFTag(100, "AcDbRasterImageDef"))
	retained = retained.Add(lldxf.NewDXFTag(1, filePath))
	retained = retained.Add(lldxf.NewDXFTag(10, width))
	retained = retained.Add(lldxf.NewDXFTag(20, height))
	retained = retained.Add(lldxf.NewDXFTag(11, pixelSizeX))
	retained = retained.Add(lldxf.NewDXFTag(21, pixelSizeY))
	s.SetRetainedTags(retained)
	return &ImageDef{StdObject: s}
}

// NewImageDefReactor builds an IMAGEDEF_REACTOR linking an image definition
// to its owning entity.
func NewImageDefReactor(objHandle, blockRec string) *ImageDefReactor {
	s := &StdObject{BaseObject: &BaseObject{}, typeName: "IMAGEDEF_REACTOR", blockRec: blockRec}
	s.SetWireType("IMAGEDEF_REACTOR")
	s.BaseObject.handle = objHandle
	retained := lldxf.NewTags()
	retained = retained.Add(lldxf.NewDXFTag(100, "AcDbRasterImageDefReactor"))
	retained = retained.Add(lldxf.NewDXFTag(90, 0))
	retained = retained.Add(lldxf.NewDXFTag(330, blockRec))
	s.SetRetainedTags(retained)
	return &ImageDefReactor{StdObject: s}
}

// NewRasterVariables builds a RASTERVARIABLES object with display defaults.
func NewRasterVariables(objHandle string, visible, frameDisplay, units int) *RasterVariables {
	s := &StdObject{BaseObject: &BaseObject{}, typeName: "RASTERVARIABLES"}
	s.SetWireType("RASTERVARIABLES")
	s.BaseObject.handle = objHandle
	retained := lldxf.NewTags()
	retained = retained.Add(lldxf.NewDXFTag(2, "AcDbRasterVariables"))
	retained = retained.Add(lldxf.NewDXFTag(90, visible))
	retained = retained.Add(lldxf.NewDXFTag(91, frameDisplay))
	retained = retained.Add(lldxf.NewDXFTag(280, units))
	s.SetRetainedTags(retained)
	return &RasterVariables{StdObject: s}
}

// NewUnderlayDef builds an underlay definition object. kind is one of
// PDFDEFINITION, DWFDEFINITION, DGNDEFINITION or UNDERLAYDEFINITION.
func NewUnderlayDef(objHandle, kind, name, filePath string, page, flags, contrast, fade int) *UnderlayDefinition {
	s := &StdObject{BaseObject: &BaseObject{}, typeName: kind, page: page}
	s.SetWireType(kind)
	s.BaseObject.handle = objHandle
	retained := lldxf.NewTags()
	retained = retained.Add(lldxf.NewDXFTag(100, underlayMarker(kind)))
	retained = retained.Add(lldxf.NewDXFTag(1, filePath))
	retained = retained.Add(lldxf.NewDXFTag(2, name))
	retained = retained.Add(lldxf.NewDXFTag(70, flags))
	retained = retained.Add(lldxf.NewDXFTag(280, contrast))
	retained = retained.Add(lldxf.NewDXFTag(281, fade))
	if page != 0 {
		retained = retained.Add(lldxf.NewDXFTag(300, page))
	}
	s.SetRetainedTags(retained)
	return &UnderlayDefinition{StdObject: s}
}

// underlayMarker maps an underlay definition record type to its subclass
// marker.
func underlayMarker(kind string) string {
	switch kind {
	case "PDFDEFINITION":
		return "AcDbPdfDefinition"
	case "DWFDEFINITION":
		return "AcDbDwfDefinition"
	case "DGNDEFINITION":
		return "AcDbDgnDefinition"
	default:
		return "AcDbUnderlayDefinition"
	}
}

// Page returns the page number of an underlay definition (code 300).
func (s *StdObject) Page() int {
	for _, t := range s.retained {
		if t.Code() == 300 {
			switch v := t.Value().(type) {
			case int:
				return v
			case float64:
				return int(v)
			case string:
				n, err := strconv.Atoi(strings.TrimSpace(v))
				if err != nil {
					f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
					if err == nil {
						return int(f)
					}
					continue
				}
				return n
			}
		}
	}
	return s.page
}
