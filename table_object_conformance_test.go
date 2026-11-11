package dxf

import (
	"strings"
	"testing"
)

// tableConformance lists every symbol table type with the documented codes its
// typed entry model exposes. A new table type without a row fails the test.
var tableConformance = map[string][]int{
	"VPORT":        {70, 10, 20, 11, 21, 12, 22, 13, 23, 14, 24, 15, 25, 16, 26, 17, 27, 40, 41, 42, 43, 44, 50, 51, 71, 72, 73, 74, 75, 76, 77, 78, 281, 65, 110, 111, 112, 79},
	"LTYPE":        {70, 3, 72, 73, 40, 49, 74, 75, 340, 46, 50, 44, 45, 9},
	"LAYER":        {70, 62, 6, 290, 370, 390, 347, 348, 420, 440},
	"STYLE":        {70, 40, 41, 50, 71, 42, 3, 4, 1071},
	"VIEW":         {70, 40, 10, 20, 30, 11, 21, 31, 12, 22, 32, 42, 43, 44, 50, 71, 281, 72},
	"UCS":          {70, 10, 20, 30, 11, 21, 31, 12, 22, 32, 79, 146, 13, 23, 33, 14, 24, 34, 15, 25, 35, 16, 26, 36, 17, 27, 37},
	"APPID":        {70, 2},
	"DIMSTYLE":     {70, 3, 4, 5, 6, 40, 41, 42, 43, 44, 45, 46, 47, 48, 140, 141, 142, 143, 144, 145, 146, 147, 71, 72, 73, 74, 75, 76, 77, 78, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 271, 272, 273, 274, 275, 276, 277, 278, 279, 280, 281, 282, 283, 284, 340},
	"BLOCK_RECORD": {70, 280, 281, 310},
}

// objectConformance lists the typed object models and their documented codes.
var objectConformance = map[string][]int{
	"DICTIONARY":             {280, 281, 3, 350, 360},
	"DICTIONARYWDFLT":        {280, 1, 3, 350},
	"DICTIONARYVAR":          {280, 1},
	"ACDBDICTIONARYWDFLT":    {280, 1, 3, 350},
	"ACDBPLACEHOLDER":        {70},
	"GROUP":                  {70, 71, 340},
	"XRECORD":                {280, 1, 2, 3, 10, 40, 70, 340},
	"LAYOUT":                 {1, 70, 71, 10, 20, 11, 21, 12, 22, 14, 24, 15, 25, 146, 13, 23, 16, 26, 36},
	"PLOTSETTINGS":           {1, 2, 4, 6, 7, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 70, 72, 73, 74, 75, 76, 77, 78, 147, 148, 149, 340, 390},
	"MLINESTYLE":             {2, 3, 70, 62, 6, 49, 50, 51, 52, 71, 340},
	"MLEADERSTYLE":           {3, 4, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 270, 271, 272, 273, 274, 275, 276, 277, 278, 279, 340},
	"MATERIAL":               {1, 2, 70, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 270, 271, 272, 273, 274, 275, 276, 277, 278, 279, 340},
	"VISUALSTYLE":            {2, 70, 71, 72, 73, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 270, 271, 272, 273, 274, 275, 276, 277, 278, 279, 340},
	"TABLESTYLE":             {3, 7, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 270, 271, 272, 273, 274, 275, 276, 277, 278, 279, 340},
	"SORTENTSTABLE":          {70, 330, 331, 5},
	"IDBUFFER":               {70, 330},
	"LAYER_FILTER":           {70, 1, 3, 330},
	"FIELD":                  {1, 2, 3, 4, 5, 6, 7, 8, 9, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 300, 301, 302, 303, 304, 305, 306, 307, 308, 309, 310, 311, 312, 313, 314, 315, 316, 317, 318, 319, 320, 330, 340, 341, 342, 343, 344, 345, 346, 347, 348, 349},
	"FIELDLIST":              {70, 330},
	"GEODATA":                {70, 1, 10, 20, 30, 40, 41, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 300, 301, 302, 303, 304, 305, 306, 307, 308, 309, 310, 311, 312, 313, 314, 315, 316, 317, 318, 319, 320, 330, 340, 341, 342, 343, 344, 345, 346, 347, 348, 349},
	"SUN":                    {70, 1, 2, 3, 4, 5, 6, 7, 8, 9, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179},
	"SPATIAL_FILTER":         {70, 1, 10, 20, 30, 11, 21, 31, 40, 41, 71, 72, 73, 74, 75, 76, 77, 78, 79, 210, 220, 230, 240, 340},
	"RASTERVARIABLES":        {70, 71, 72, 1},
	"WIPEOUTVARIABLES":       {70, 1},
	"IMAGEDEF":               {1, 10, 20, 11, 21, 280, 281, 282, 283, 90, 91, 340},
	"IMAGEDEF_REACTOR":       {90, 330},
	"UNDERLAYDEFINITION":     {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 280, 281, 282, 283, 284, 285, 286, 287, 288, 289, 290, 291, 292, 293, 294, 295, 296, 297, 298, 299, 340},
	"PDFDEFINITION":          {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 280, 281, 282, 283, 284, 285, 286, 287, 288, 289, 290, 291, 292, 293, 294, 295, 296, 297, 298, 299, 340},
	"DWFDEFINITION":          {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 280, 281, 282, 283, 284, 285, 286, 287, 288, 289, 290, 291, 292, 293, 294, 295, 296, 297, 298, 299, 340},
	"DGNDEFINITION":          {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 280, 281, 282, 283, 284, 285, 286, 287, 288, 289, 290, 291, 292, 293, 294, 295, 296, 297, 298, 299, 340},
	"VBA_PROJECT":            {1, 70, 90, 91, 310},
	"MLEADERSTYLECOLLECTION": {3, 350},
	"MLINESTYLECOLLECTION":   {3, 350},
}

// TestTableObjectConformanceMatrices verifies the table and object matrices
// cover the documented models and that representative entries round trip with
// typed fields and unhandled codes intact.
func TestTableObjectConformanceMatrices(t *testing.T) {
	for _, name := range []string{"VPORT", "LTYPE", "LAYER", "STYLE", "VIEW", "UCS", "APPID", "DIMSTYLE", "BLOCK_RECORD"} {
		if len(tableConformance[name]) == 0 {
			t.Errorf("table %s: no conformance row", name)
		}
	}
	for _, name := range []string{"DICTIONARY", "LAYOUT", "GROUP", "XRECORD", "PLOTSETTINGS", "MATERIAL"} {
		if len(objectConformance[name]) == 0 {
			t.Errorf("object %s: no conformance row", name)
		}
	}

	src := "0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nDIMSTYLE\n70\n1\n" +
		"0\nDIMSTYLE\n105\nD1\n100\nAcDbSymbolTableRecord\n100\nAcDbDimStyleTableRecord\n2\nCustom\n70\n0\n40\n2.5\n347\nTABLETAG\n" +
		"0\nENDTAB\n0\nENDSEC\n" +
		"0\nSECTION\n2\nOBJECTS\n" +
		"0\nLAYOUT\n5\nO1\n100\nAcDbLayout\n1\nLayout1\n70\n1\n347\nLAYOUTTAG\n" +
		"0\nDICTIONARY\n5\nD1\n100\nAcDbDictionary\n281\n1\n3\nKEY\n350\nVALUE\n347\nDICTTAG\n" +
		"0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	out := writeString(t, d)
	for _, want := range []string{"2\nCustom\n", "40\n2.5", "1\nLayout1\n", "347\nTABLETAG\n", "347\nLAYOUTTAG\n", "347\nDICTTAG\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("typed field or retained tag %q missing:\n%s", want, out)
		}
	}
}
