package lldxf

// PositionedComment is a code 999 comment anchored to its original location in
// a section. Index counts code 0 records in the section body before the
// comment; 0 means the comment appeared before the first record.
type PositionedComment struct {
	Section string
	Index   int
	Text    string
}
