package table

// TableType represents Table names (code 2)
type TableType int

// Table name: code 2
const (
	VPORT TableType = iota
	LTYPE
	LAYER
	STYLE
	VIEW
	UCS
	APPID
	DIMSTYLE
	BLOCK_RECORD
)

// TableTypeValue converts string to TableType.
// If string is unknown TableType, it returns -1.
func TableTypeValue(t string) TableType {
	switch t {
	case "VPORT":
		return VPORT
	case "LTYPE":
		return LTYPE
	case "LAYER":
		return LAYER
	case "STYLE":
		return STYLE
	case "VIEW":
		return VIEW
	case "UCS":
		return UCS
	case "APPID":
		return APPID
	case "DIMSTYLE":
		return DIMSTYLE
	case "BLOCK_RECORD":
		return BLOCK_RECORD
	default:
		return -1
	}
}
