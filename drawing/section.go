package drawing

// SectionType represents Section names (code 2)
type SectionType int

// Section name: code 2
const (
	HEADER SectionType = iota
	CLASSES
	TABLES
	BLOCKS
	ENTITIES
	OBJECTS
)

// SectionTypeValue converts string to SectionType.
// If string is unknown SectionType, it returns -1.
func SectionTypeValue(s string) SectionType {
	switch s {
	case "HEADER":
		return HEADER
	case "CLASSES":
		return CLASSES
	case "TABLES":
		return TABLES
	case "BLOCKS":
		return BLOCKS
	case "ENTITIES":
		return ENTITIES
	case "OBJECTS":
		return OBJECTS
	default:
		return -1
	}
}
