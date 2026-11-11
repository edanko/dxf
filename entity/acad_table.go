package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/table"
)

// AcadTable represents an ACAD_TABLE entity (R2000+)
type AcadTable struct {
	*entity

	// Basic properties
	InsertPoint []float64 // 10, 20, 30 - Insertion point
	TableName   string    // 1 - Table name
	StyleName   string    // 7 - Table style name
	Direction   int       // 70 - Table direction (0=Up, 1=Down, 2=Left, 3=Right)
	RowHeight   float64   // 40 - Row height
	ColWidth    float64   // 41 - Column width
	RowCount    int       // 91 - Number of rows
	ColCount    int       // 92 - Number of columns
	Flags       int       // legacy table flags (not emitted)

	// ACAD_TABLE wire layout
	BlockName                    string    // 2 - anonymous table block name (*T...)
	TableStyleHandle             string    // 342 - TABLESTYLE object handle
	BlockRecordHandle            string    // 343 - associated block record handle
	HorizontalDirection          []float64 // 11, 21, 31 - horizontal direction vector
	TableValue                   int       // 90 - table value flag
	Version                      int       // 280 - table data version
	OverrideFlag                 int       // 93
	BorderColorOverrideFlag      int       // 94
	BorderLineweightOverrideFlag int       // 95
	BorderVisibilityOverrideFlag int       // 96
	RowHeights                   []float64 // 141 - per-row heights
	ColWidths                    []float64 // 142 - per-column widths

	// Table content
	Rows               []*TableRow // Row data
	TableStyleOverride *TableStyle // Style overrides

	// 3D properties
	Extrusion []float64 // 210, 220, 230 - Extrusion direction
	Normal    []float64 // 210, 220, 230 - Normal vector
	Scale     []float64 // 41, 42, 43 - Scale factors
}

// TableRow represents a single row in a table
type TableRow struct {
	Cells  []*TableCell // Row cells
	Height float64      // Row height override
	Format *RowFormat   // Row formatting
}

// TableCell represents a single cell in a table
type TableCell struct {
	Text       string        // Cell text content
	CellType   CellType      // Type of cell (171)
	Alignment  CellAlignment // Text alignment (170)
	TextColor  int           // Text color
	BgColor    int           // Background color
	GridColor  int           // Grid line color
	GridWeight int           // Grid line weight
	MergedCell *MergedCell   // Merged cell information
	CellStyle  *CellStyle    // Cell style override
	DataValue  any           // Data value for binding

	// Cell coordinates and raw wire attributes
	Row          int     // zero-based row index
	Col          int     // zero-based column index
	Flags        int     // 172 - cell flags
	Merged       bool    // 173 - cell is covered by a merged master
	AutoFit      bool    // 174 - autofit option
	BorderWidth  int     // 175 - merged cell width (column span)
	BorderHeight int     // 176 - merged cell height (row span)
	OverrideFlag int     // 91 - cell override flag
	VirtualEdge  int     // 178 - virtual edge flag
	ExtendedFlag int     // 92 - extended cell flags
	ValueType    int     // 93 - cell value type
	DataType     int     // 90 - cell data type
	ValueFlag    int     // 94 - value block flag
	Rotation     float64 // 145 - rotation for block/text cells

	TextStyleName string  // 7 - cell text style override
	TextHeight    float64 // 140 - cell text height override

	BlockRecordHandle string  // 340 - block table record handle (block cells)
	BlockScale        float64 // 144 - block scale (block cells)
}

// CellType represents the type of table cell
type CellType int

const (
	CellTypeText      CellType = iota // Text cell
	CellTypeBlock                     // Block cell
	CellTypeAttribute                 // Attribute cell
	CellTypeFormula                   // Formula cell
)

// CellAlignment represents text alignment in a cell
type CellAlignment int

const (
	CellAlignTopLeft      CellAlignment = iota // Top-left
	CellAlignTopCenter                         // Top-center
	CellAlignTopRight                          // Top-right
	CellAlignMiddleLeft                        // Middle-left
	CellAlignMiddleCenter                      // Middle-center
	CellAlignMiddleRight                       // Middle-right
	CellAlignBottomLeft                        // Bottom-left
	CellAlignBottomCenter                      // Bottom-center
	CellAlignBottomRight                       // Bottom-right
)

// MergedCell represents merged cell information
type MergedCell struct {
	RowSpan    int        // Number of rows spanned
	ColSpan    int        // Number of columns spanned
	MasterCell *TableCell // Reference to master cell
}

// CellStyle represents cell-specific style overrides
type CellStyle struct {
	TextStyle      string        // Text style name
	TextHeight     float64       // Text height override
	TextColor      int           // Text color override
	BgColor        int           // Background color override
	Alignment      CellAlignment // Text alignment override
	BorderOverride bool          // Whether borders override table style
	TopBorder      *BorderStyle  // Top border style
	RightBorder    *BorderStyle  // Right border style
	BottomBorder   *BorderStyle  // Bottom border style
	LeftBorder     *BorderStyle  // Left border style
}

// BorderStyle represents a cell border
type BorderStyle struct {
	LineStyle int  // Linetype
	Color     int  // Color
	Weight    int  // Line weight
	Visible   bool // Visibility
}

// RowFormat represents row-specific formatting
type RowFormat struct {
	Height    float64 // Row height override
	BgColor   int     // Background color
	TextStyle string  // Text style
	TextColor int     // Text color
}

// TableStyle represents table style
type TableStyle struct {
	Name                string  // Style name
	TextStyle           string  // Default text style
	TextHeight          float64 // Default text height
	TextColor           int     // Default text color
	BgColor             int     // Default background color
	GridColor           int     // Default grid color
	GridWeight          int     // Default grid weight
	HeaderRowHeight     float64 // Header row height
	HeaderBgColor       int     // Header background color
	HeaderTextColor     int     // Header text color
	RowHeight           float64 // Data row height
	MinRowHeight        float64 // Minimum row height
	CellMarginH         float64 // Horizontal cell margin
	CellMarginV         float64 // Vertical cell margin
	FlowDirection       int     // 0=Down, 1=Up, 2=Left, 3=Right
	BreakOnPage         bool    // Break table across pages
	RepeatTopLabels     bool    // Repeat top row labels
	BottomLabels        bool    // Bottom row labels
	TitleRow            bool    // Title row
	TitleSuppressed     bool    // Title row suppression
	HeaderRowSuppressed bool    // Column header suppression
}

// IsEntity is for Entity interface
func (t *AcadTable) IsEntity() bool {
	return true
}

// NewAcadTable creates a new AcadTable entity
func NewAcadTable() *AcadTable {
	t := &AcadTable{
		entity:              NewEntity(ACADTABLE),
		InsertPoint:         []float64{0.0, 0.0, 0.0},
		TableName:           "",
		StyleName:           "",
		Direction:           0, // Up
		RowHeight:           1.0,
		ColWidth:            1.0,
		RowCount:            0,
		ColCount:            0,
		Flags:               0,
		Rows:                make([]*TableRow, 0),
		HorizontalDirection: []float64{1.0, 0.0, 0.0},
		Extrusion:           []float64{0.0, 0.0, 1.0},
		Normal:              []float64{0.0, 0.0, 1.0},
		Scale:               []float64{1.0, 1.0, 1.0},
	}
	t.SetRecordTypeName("ACAD_TABLE")
	return t
}

// SetInsertPoint sets the insertion point
func (t *AcadTable) SetInsertPoint(x, y, z float64) {
	if len(t.InsertPoint) >= 3 {
		t.InsertPoint[0] = x
		t.InsertPoint[1] = y
		t.InsertPoint[2] = z
	}
}

// SetTableSize sets the table dimensions
func (t *AcadTable) SetTableSize(rows, cols int) {
	t.RowCount = rows
	t.ColCount = cols
}

// SetTableName sets the table name
func (t *AcadTable) SetTableName(name string) {
	t.TableName = name
}

// SetStyle sets the table style name
func (t *AcadTable) SetStyle(styleName string) {
	t.StyleName = styleName
}

// SetDirection sets the table flow direction
func (t *AcadTable) SetDirection(direction int) {
	t.Direction = direction
}

// SetCellSize sets default cell dimensions
func (t *AcadTable) SetCellSize(width, height float64) {
	t.ColWidth = width
	t.RowHeight = height
}

// SetExtrusion sets the 3D extrusion direction
func (t *AcadTable) SetExtrusion(x, y, z float64) {
	if len(t.Extrusion) >= 3 {
		t.Extrusion[0] = x
		t.Extrusion[1] = y
		t.Extrusion[2] = z
	}
}

// SetNormal sets the surface normal vector
func (t *AcadTable) SetNormal(x, y, z float64) {
	if len(t.Normal) >= 3 {
		t.Normal[0] = x
		t.Normal[1] = y
		t.Normal[2] = z
	}
}

// SetScale sets the scale factors
func (t *AcadTable) SetScale(x, y, z float64) {
	if len(t.Scale) >= 3 {
		t.Scale[0] = x
		t.Scale[1] = y
		t.Scale[2] = z
	}
}

// AddRow adds a new row to the table
func (t *AcadTable) AddRow() *TableRow {
	row := &TableRow{
		Cells: make([]*TableCell, t.ColCount),
	}
	t.Rows = append(t.Rows, row)
	return row
}

// AddColumn adds a new column (expands existing rows)
func (t *AcadTable) AddColumn() {
	for _, row := range t.Rows {
		row.Cells = append(row.Cells, &TableCell{})
	}
	t.ColCount++
}

// SetMergedCell creates a merged cell spanning multiple rows/columns
func (t *AcadTable) SetMergedCell(startRow, startCol, rowSpan, colSpan int) {
	if startRow >= 0 && startRow < len(t.Rows) && startCol >= 0 && startCol < len(t.Rows[startRow].Cells) {
		masterCell := &TableCell{
			Text:      "",
			CellType:  CellTypeText,
			Alignment: CellAlignTopLeft,
			MergedCell: &MergedCell{
				RowSpan: rowSpan,
				ColSpan: colSpan,
			},
		}

		// Mark the spanned cells
		for r := startRow; r < startRow+rowSpan && r < len(t.Rows); r++ {
			for c := startCol; c < startCol+colSpan && c < len(t.Rows[r].Cells); c++ {
				if t.Rows[r].Cells[c] == nil {
					t.Rows[r].Cells[c] = &TableCell{}
				}
				t.Rows[r].Cells[c].MergedCell = &MergedCell{
					RowSpan:    0, // Spanned cell
					ColSpan:    0,
					MasterCell: masterCell,
				}
			}
		}

		if startRow < len(t.Rows) && startCol < len(t.Rows[startRow].Cells) {
			if t.Rows[startRow].Cells[startCol] == nil {
				t.Rows[startRow].Cells[startCol] = &TableCell{}
			}
			t.Rows[startRow].Cells[startCol] = masterCell
		}
	}
}

// SetCell sets the text content of a cell
func (t *AcadTable) SetCell(rowIdx, colIdx int, text string) {
	if rowIdx >= 0 && rowIdx < len(t.Rows) && colIdx >= 0 && colIdx < len(t.Rows[rowIdx].Cells) {
		if t.Rows[rowIdx].Cells[colIdx] == nil {
			t.Rows[rowIdx].Cells[colIdx] = &TableCell{}
		}
		t.Rows[rowIdx].Cells[colIdx].Text = text
		t.Rows[rowIdx].Cells[colIdx].CellType = CellTypeText
	}
}

// SetCellStyle sets cell-specific formatting
func (t *AcadTable) SetCellStyle(rowIdx, colIdx int, style *CellStyle) {
	if rowIdx >= 0 && rowIdx < len(t.Rows) && colIdx >= 0 && colIdx < len(t.Rows[rowIdx].Cells) {
		if t.Rows[rowIdx].Cells[colIdx] == nil {
			t.Rows[rowIdx].Cells[colIdx] = &TableCell{}
		}
		t.Rows[rowIdx].Cells[colIdx].CellStyle = style
	}
}

// SetLayer assigns the entity to a layer
func (t *AcadTable) SetLayer(layer *table.Layer) {
	t.entity.layer = layer
}

// BBox returns the bounding box of the table
func (t *AcadTable) BBox() ([]float64, []float64) {
	if len(t.InsertPoint) < 3 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}

	x, y, z := t.InsertPoint[0], t.InsertPoint[1], t.InsertPoint[2]

	// Calculate table width and height from per-row/column sizes when present.
	width := 0.0
	if len(t.ColWidths) > 0 {
		for _, w := range t.ColWidths {
			width += w
		}
	} else {
		width = float64(t.ColCount) * t.ColWidth
	}
	height := 0.0
	if len(t.RowHeights) > 0 {
		for _, h := range t.RowHeights {
			height += h
		}
	} else {
		height = float64(t.RowCount) * t.RowHeight
	}

	// Return min and max corners
	return []float64{x, y, z}, []float64{x + width, y + height, z}
}

// Format writes data to formatter
func (t *AcadTable) Format(f format.Formatter) {
	t.entity.Format(f)

	// ACAD_TABLE reuses the block reference subclass for placement.
	f.WriteString(100, "AcDbBlockReference")
	if t.BlockName != "" {
		f.WriteString(2, t.BlockName)
	}
	for i := range 3 {
		val := 0.0
		if i < len(t.InsertPoint) {
			val = t.InsertPoint[i]
		}
		f.WriteFloat(10+i*10, val)
	}

	f.WriteString(100, "AcDbTable")
	f.WriteInt(280, t.Version)
	if t.TableStyleHandle != "" {
		f.WriteString(342, t.TableStyleHandle)
	}
	if t.BlockRecordHandle != "" {
		f.WriteString(343, t.BlockRecordHandle)
	}
	for i := range 3 {
		val := 0.0
		if i < len(t.HorizontalDirection) {
			val = t.HorizontalDirection[i]
		} else if i == 0 {
			val = 1.0
		}
		f.WriteFloat(11+i*10, val)
	}
	if t.TableValue != 0 {
		f.WriteInt(90, t.TableValue)
	}
	f.WriteInt(91, t.RowCount)
	f.WriteInt(92, t.ColCount)
	if t.OverrideFlag != 0 {
		f.WriteInt(93, t.OverrideFlag)
	}
	if t.BorderColorOverrideFlag != 0 {
		f.WriteInt(94, t.BorderColorOverrideFlag)
	}
	if t.BorderLineweightOverrideFlag != 0 {
		f.WriteInt(95, t.BorderLineweightOverrideFlag)
	}
	if t.BorderVisibilityOverrideFlag != 0 {
		f.WriteInt(96, t.BorderVisibilityOverrideFlag)
	}
	for _, h := range t.RowHeights {
		f.WriteFloat(141, h)
	}
	for _, w := range t.ColWidths {
		f.WriteFloat(142, w)
	}

	for _, row := range t.Rows {
		for _, cell := range row.Cells {
			if cell != nil {
				t.formatCell(f, cell)
			}
		}
	}

	// Table-level style overrides.
	if style := t.TableStyleOverride; style != nil {
		if style.FlowDirection != 0 {
			f.WriteInt(70, style.FlowDirection)
		}
		if style.CellMarginH != 0 {
			f.WriteFloat(40, style.CellMarginH)
		}
		if style.CellMarginV != 0 {
			f.WriteFloat(41, style.CellMarginV)
		}
		if style.TitleSuppressed {
			f.WriteInt(280, 1)
		}
		if style.TextStyle != "" {
			f.WriteString(7, style.TextStyle)
		}
		if style.TextHeight != 0 {
			f.WriteFloat(140, style.TextHeight)
		}
	}

	// 3D properties
	if len(t.Extrusion) == 3 {
		for i := range 3 {
			f.WriteFloat(210+i, t.Extrusion[i])
		}
	}
	if len(t.Normal) == 3 {
		for i := range 3 {
			f.WriteFloat(210+i, t.Normal[i])
		}
	}
	if len(t.Scale) == 3 {
		for i := range 3 {
			f.WriteFloat(40+i, t.Scale[i])
		}
	}
}

// formatCell writes one ACAD_TABLE cell record.
func (t *AcadTable) formatCell(f format.Formatter, cell *TableCell) {
	typeCode := 1
	if cell.CellType == CellTypeBlock {
		typeCode = 2
	}
	f.WriteInt(171, typeCode)
	f.WriteInt(172, cell.Flags)
	if cell.Merged {
		f.WriteInt(173, 1)
	} else {
		f.WriteInt(173, 0)
	}
	if cell.AutoFit {
		f.WriteInt(174, 1)
	} else {
		f.WriteInt(174, 0)
	}
	width := cell.BorderWidth
	if width == 0 {
		width = 1
	}
	height := cell.BorderHeight
	if height == 0 {
		height = 1
	}
	f.WriteInt(175, width)
	f.WriteInt(176, height)
	f.WriteInt(91, cell.OverrideFlag)
	f.WriteInt(178, cell.VirtualEdge)
	f.WriteFloat(145, cell.Rotation)

	if cell.CellType == CellTypeBlock {
		if cell.BlockRecordHandle != "" {
			f.WriteString(340, cell.BlockRecordHandle)
		}
		f.WriteFloat(144, cell.BlockScale)
		f.WriteInt(170, int(cell.Alignment))
	}
	f.WriteInt(92, cell.ExtendedFlag)
	f.WriteString(301, "CELL_VALUE")
	valueType := cell.ValueType
	dataType := cell.DataType
	if valueType == 0 {
		if cell.Text != "" {
			valueType, dataType = 2, 4
		} else {
			valueType, dataType = 3, 0
		}
	}
	f.WriteInt(93, valueType)
	f.WriteInt(90, dataType)
	if cell.Text != "" {
		writeCellText(f, cell.Text)
	}
	f.WriteInt(94, cell.ValueFlag)
	f.WriteString(300, "")
	f.WriteString(302, cell.Text)
	f.WriteString(304, "ACVALUE_END")
}

// writeCellText emits the code 1/2/3 text chunks used by ACAD_TABLE cells.
func writeCellText(f format.Formatter, text string) {
	const chunkSize = 250
	if len(text) <= chunkSize {
		f.WriteString(1, text)
		return
	}
	for len(text) > chunkSize {
		f.WriteString(2, text[:chunkSize])
		text = text[chunkSize:]
	}
	f.WriteString(1, text)
}

// String outputs data using default formatter
func (t *AcadTable) String() string {
	f := format.NewASCII()
	return t.FormatString(f)
}

// FormatString outputs data using given formatter
func (t *AcadTable) FormatString(f format.Formatter) string {
	t.Format(f)
	return f.Output()
}

// AcadTableEditor provides utility methods for building tables
type AcadTableEditor struct {
	table *AcadTable
}

// NewAcadTableEditor creates a new table editor
func NewAcadTableEditor(table *AcadTable) *AcadTableEditor {
	return &AcadTableEditor{
		table: table,
	}
}

// SetHeader creates a header row with bold text
func (e *AcadTableEditor) SetHeader(texts []string) {
	if len(texts) > 0 {
		e.table.SetTableSize(len(texts), len(texts))
		headerRow := e.table.AddRow()
		for i, text := range texts {
			e.table.SetCell(0, i, text)
			if headerRow.Cells[i] != nil {
				headerRow.Cells[i].TextColor = 1 // White/black contrast
				headerRow.Cells[i].BgColor = 5   // Blue background
				style := &CellStyle{
					TextColor: 1,
					BgColor:   5,
				}
				e.table.SetCellStyle(0, i, style)
			}
		}
	}
}

// AddDataRow adds a data row with the given texts
func (e *AcadTableEditor) AddDataRow(texts []string) {
	if len(texts) > 0 {
		e.table.SetTableSize(e.table.RowCount+1, len(texts))
		rowIdx := e.table.RowCount
		e.table.AddRow()
		for i, text := range texts {
			e.table.SetCell(rowIdx, i, text)
		}
	}
}

// AddMergedCells creates merged cells in the current row
func (e *AcadTableEditor) AddMergedCells(rowIdx int, cells []struct {
	StartCol int
	EndCol   int
	Text     string
}) {
	if rowIdx < len(e.table.Rows) {
		row := e.table.Rows[rowIdx]
		for _, cell := range cells {
			e.table.SetMergedCell(rowIdx, cell.StartCol, 1, cell.EndCol-cell.StartCol+1)
			if row.Cells[cell.StartCol] != nil {
				row.Cells[cell.StartCol].Text = cell.Text
			}
		}
	}
}

// GetTable returns the constructed table
func (e *AcadTableEditor) GetTable() *AcadTable {
	return e.table
}
