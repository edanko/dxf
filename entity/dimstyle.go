package entity

import (
	"fmt"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
)

// DimStyle represents a DXF DIMSTYLE table entry
type DimStyle struct {
	*entity

	// Core properties
	Name     string `dxf:"2"`
	Flags    int    `dxf:"70"`
	DimPost  string `dxf:"3"` // Primary dimension text suffix/prefix
	DimAPost string `dxf:"4"` // Alternate units text suffix/prefix

	// Arrow and block references
	DimBlk    string `dxf:"5"`                 // Arrow block name
	DimBlk1   string `dxf:"6"`                 // First arrow block name
	DimBlk2   string `dxf:"7"`                 // Second arrow block name
	DimLdrBlk string `dxf:"virtual,dimldrblk"` // Leader arrow block

	// Handle versions of arrow blocks
	DimBlkHandle    string `dxf:"342"` // Main arrow block handle
	DimBlk1Handle   string `dxf:"343"` // First arrow block handle
	DimBlk2Handle   string `dxf:"344"` // Second arrow block handle
	DimLdrBlkHandle string `dxf:"341"` // Leader arrow block handle

	// Scale and size parameters
	DimScale float64 `dxf:"40"` // Overall scale factor
	DimAsz   float64 `dxf:"41"` // Arrow size
	DimExo   float64 `dxf:"42"` // Extension line offset
	DimDli   float64 `dxf:"43"` // Dimension line increment
	DimExe   float64 `dxf:"44"` // Extension line extension
	DimRnd   float64 `dxf:"45"` // Rounding value
	DimDle   float64 `dxf:"46"` // Dimension line extension
	DimTp    float64 `dxf:"47"` // Upper tolerance
	DimTm    float64 `dxf:"48"` // Lower tolerance

	// Text properties
	DimTxt   float64 `dxf:"140"`              // Text height
	DimTvp   float64 `dxf:"145"`              // Text vertical position
	DimTfac  float64 `dxf:"146"`              // Tolerance text height factor
	DimGap   float64 `dxf:"147"`              // Gap from dimension line
	DimTxsty string  `dxf:"virtual,dimtxsty"` // Text style name

	// Handle version of text style
	DimTxstyHandle string `dxf:"340"` // Text style handle

	// Colors and appearance
	DimClrd     color.ColorNumber `dxf:"176"` // Dimension line color
	DimClre     color.ColorNumber `dxf:"177"` // Extension line color
	DimClrt     color.ColorNumber `dxf:"178"` // Text color
	DimTfill    int               `dxf:"69"`  // Text fill (R2007+)
	DimTfillclr color.ColorNumber `dxf:"70"`  // Text fill color (R2007+)

	// Formatting and units (R2000+)
	DimAltrnd int `dxf:"148"` // Alternate units rounding
	DimDec    int `dxf:"271"` // Primary units decimal places
	DimTdec   int `dxf:"272"` // Tolerance decimal places
	DimAltu   int `dxf:"273"` // Alternate units format
	DimAlttd  int `dxf:"274"` // Alternate tolerance decimal places
	DimAunit  int `dxf:"275"` // Angular unit format
	DimFrac   int `dxf:"276"` // Fraction format
	DimLunit  int `dxf:"277"` // Linear unit format
	DimDecSep int `dxf:"278"` // Decimal separator

	// Text position and alignment
	DimTad  int `dxf:"77"`  // Text vertical position
	DimTih  int `dxf:"73"`  // Text inside horizontal
	DimToh  int `dxf:"74"`  // Text outside horizontal
	DimJust int `dxf:"280"` // Text horizontal justification

	// Extension line control (R2007+)
	DimSe1   int    `dxf:"75"`               // Suppress extension line 1
	DimSe2   int    `dxf:"76"`               // Suppress extension line 2
	DimLtype string `dxf:"virtual,dimltype"` // Dimension line linetype
	DimLtEx1 string `dxf:"virtual,dimltex1"` // Extension line 1 linetype
	DimLtEx2 string `dxf:"virtual,dimltex2"` // Extension line 2 linetype

	// Handle versions of linetypes
	DimLtypeHandle string `dxf:"345"` // Dimension line linetype handle
	DimLtEx1Handle string `dxf:"346"` // Extension line 1 linetype handle
	DimLtEx2Handle string `dxf:"347"` // Extension line 2 linetype handle

	// Advanced options
	DimTol    int     `dxf:"71"`  // Tolerance display
	DimLim    int     `dxf:"72"`  // Limits display
	DimZin    int     `dxf:"78"`  // Zero suppression
	DimAzin   int     `dxf:"79"`  // Alternate zero suppression
	DimArcsym int     `dxf:"90"`  // Arc symbol display
	DimAlt    int     `dxf:"170"` // Alternate units
	DimAltd   int     `dxf:"171"` // Alternate units decimal places
	DimTofl   int     `dxf:"172"` // Text outside extension lines
	DimSah    int     `dxf:"173"` // Separate arrows
	DimTix    int     `dxf:"174"` // Force text inside
	DimSoxd   int     `dxf:"175"` // Suppress outside dimension lines
	DimAdec   int     `dxf:"179"` // Angular decimal places
	DimTolj   int     `dxf:"283"` // Tolerance vertical justification
	DimTzin   int     `dxf:"284"` // Tolerance zero suppression
	DimAltz   int     `dxf:"285"` // Alternate unit zero suppression
	DimAltzz  int     `dxf:"286"` // Alternate tolerance zero suppression
	DimUpt    int     `dxf:"288"` // User text position on dimension line
	DimAtfit  int     `dxf:"289"` // Text and arrow fitting
	DimFxlon  int     `dxf:"290"` // Fixed extension line length
	DimFxL    float64 `dxf:"49"`  // Fixed extension line length
	DimJogang float64 `dxf:"50"`  // Jog angle
	DimLfac   float64 `dxf:"144"` // Length factor
	DimAlttf  float64 `dxf:"143"` // Alternate units scale factor
	DimTsz    float64 `dxf:"142"` // Tick size
}

// NewDimStyle creates a new dimension style
func NewDimStyle() *DimStyle {
	return &DimStyle{
		entity: NewEntity(DIMSTYLE),

		// Default values based on AutoCAD standard style
		Name:        "Standard",
		Flags:       0,
		DimScale:    1.0,
		DimAsz:      2.5,
		DimExo:      0.625,
		DimDli:      0.0,
		DimExe:      1.25,
		DimTxt:      2.5,
		DimTvp:      0.0,
		DimTfac:     1.0,
		DimGap:      0.625,
		DimRnd:      0.0,
		DimDle:      0.0,
		DimClrd:     0,
		DimClre:     0,
		DimClrt:     0,
		DimJust:     1,
		DimTad:      1,
		DimTih:      0,
		DimToh:      0,
		DimTol:      0,
		DimLim:      0,
		DimZin:      0,
		DimAzin:     0,
		DimTfill:    0,
		DimTfillclr: 0,

		// R2000+ defaults
		DimDec:    4,
		DimTdec:   4,
		DimDecSep: 44, // Period
		DimAlt:    0,
		DimAltu:   2, // Decimal
		DimAltd:   4,
		DimAunit:  0, // Decimal degrees
		DimFrac:   0, // No fractions
		DimLunit:  2, // Decimal
		DimAltrnd: 0,

		// R2007+ defaults
		DimLtype: "BYBLOCK",
		DimLtEx1: "BYBLOCK",
		DimLtEx2: "BYBLOCK",
		DimSe1:   0,
		DimSe2:   0,
		DimFxlon: 0,
		DimFxL:   0.0,

		// Advanced defaults
		DimArcsym: 0,
		DimSah:    0,
		DimTix:    0,
		DimSoxd:   0,
		DimAdec:   2,
		DimTolj:   0,
		DimTzin:   0,
		DimAltz:   0,
		DimAltzz:  0,
		DimUpt:    0,
		DimAtfit:  3,
		DimJogang: 0.0,
		DimLfac:   1.0,
		DimAlttf:  1.0,
		DimTsz:    0.0,
	}
}

// IsEntity returns true for DIMSTYLE entities
func (s *DimStyle) IsEntity() bool {
	return true
}

// DXFType returns the DXF type string
func (s *DimStyle) DXFType() string {
	return "DIMSTYLE"
}

// Format writes DIMSTYLE data to formatter
func (s *DimStyle) Format(f format.Formatter) {
	// Basic header
	f.WriteString(0, "DIMSTYLE")
	s.entity.Format(f)
	f.WriteString(100, "AcDbDimStyleTableRecord")
	f.WriteString(105, s.Name)

	// Core properties
	f.WriteInt(70, s.Flags)
	if s.DimPost != "" {
		f.WriteString(3, s.DimPost)
	}
	if s.DimAPost != "" {
		f.WriteString(4, s.DimAPost)
	}

	// Arrow block names
	if s.DimBlk != "" {
		f.WriteString(5, s.DimBlk)
	}
	if s.DimBlk1 != "" {
		f.WriteString(6, s.DimBlk1)
	}
	if s.DimBlk2 != "" {
		f.WriteString(7, s.DimBlk2)
	}
	if s.DimLdrBlk != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, s.DimLdrBlkHandle)
		f.WriteString(102, "}")
	}

	// Scale and size parameters
	f.WriteFloat(40, s.DimScale)
	f.WriteFloat(41, s.DimAsz)
	f.WriteFloat(42, s.DimExo)
	f.WriteFloat(43, s.DimDli)
	f.WriteFloat(44, s.DimExe)
	f.WriteFloat(45, s.DimRnd)
	f.WriteFloat(46, s.DimDle)
	f.WriteFloat(47, s.DimTp)
	f.WriteFloat(48, s.DimTm)

	// Text properties
	f.WriteFloat(140, s.DimTxt)
	f.WriteFloat(145, s.DimTvp)
	f.WriteFloat(146, s.DimTfac)
	f.WriteFloat(147, s.DimGap)
	if s.DimTxsty != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, s.DimTxstyHandle)
		f.WriteString(102, "}")
	}

	// Colors
	f.WriteInt(176, int(s.DimClrd))
	f.WriteInt(177, int(s.DimClre))
	f.WriteInt(178, int(s.DimClrt))

	// Formatting options
	f.WriteInt(271, s.DimDec)
	f.WriteInt(272, s.DimTdec)
	f.WriteInt(273, s.DimAltu)
	f.WriteInt(274, s.DimAlttd)
	f.WriteInt(275, s.DimAunit)
	f.WriteInt(276, s.DimFrac)
	f.WriteInt(277, s.DimLunit)
	f.WriteInt(278, s.DimDecSep)

	// Text positioning
	f.WriteInt(71, s.DimTol)
	f.WriteInt(72, s.DimLim)
	f.WriteInt(73, s.DimTih)
	f.WriteInt(74, s.DimToh)
	f.WriteInt(77, s.DimTad)
	f.WriteInt(78, s.DimZin)
	f.WriteInt(79, s.DimAzin)

	// Advanced options
	f.WriteInt(170, s.DimAlt)
	f.WriteInt(171, s.DimAltd)
	f.WriteInt(172, s.DimTofl)
	f.WriteInt(173, s.DimSah)
	f.WriteInt(174, s.DimTix)
	f.WriteInt(175, s.DimSoxd)
	f.WriteInt(179, s.DimAdec)
	f.WriteInt(280, s.DimJust)

	// R2007+ features
	if s.DimTfill != 0 {
		f.WriteInt(69, s.DimTfill)
		f.WriteInt(70, int(s.DimTfillclr))
	}
	if s.DimLtype != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, s.DimLtypeHandle)
		f.WriteString(102, "}")
	}
	if s.DimLtEx1 != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, s.DimLtEx1Handle)
		f.WriteString(102, "}")
	}
	if s.DimLtEx2 != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, s.DimLtEx2Handle)
		f.WriteString(102, "}")
	}
	f.WriteInt(75, s.DimSe1)
	f.WriteInt(76, s.DimSe2)
	f.WriteInt(90, s.DimArcsym)
	f.WriteInt(173, s.DimTix)

	// Final options
	f.WriteInt(283, s.DimTolj)
	f.WriteInt(284, s.DimTzin)
	f.WriteInt(285, s.DimAltz)
	f.WriteInt(286, s.DimAltzz)
	f.WriteInt(287, s.DimUpt)
	f.WriteInt(288, s.DimAtfit)
	f.WriteInt(289, s.DimFxlon)
	if s.DimFxL != 0.0 {
		f.WriteFloat(49, s.DimFxL)
	}
	if s.DimJogang != 0.0 {
		f.WriteFloat(50, s.DimJogang)
	}
	f.WriteFloat(142, s.DimTsz)
	f.WriteFloat(143, s.DimAlttf)
	f.WriteFloat(144, s.DimLfac)
	if s.DimAltrnd != 0 {
		f.WriteInt(148, s.DimAltrnd)
	}

	// Colors
	f.WriteInt(176, int(s.DimClrd))
	f.WriteInt(177, int(s.DimClre))
	f.WriteInt(178, int(s.DimClrt))

	// Formatting options
	f.WriteString(271, fmt.Sprintf("%d", s.DimDec))
	f.WriteString(272, fmt.Sprintf("%d", s.DimTdec))
	f.WriteString(273, fmt.Sprintf("%d", s.DimAltu))
	f.WriteString(274, fmt.Sprintf("%d", s.DimAlttd))
	f.WriteString(275, fmt.Sprintf("%d", s.DimAunit))
	f.WriteString(276, fmt.Sprintf("%d", s.DimFrac))
	f.WriteString(277, fmt.Sprintf("%d", s.DimLunit))
	f.WriteString(278, fmt.Sprintf("%d", s.DimDecSep))

	// Text positioning
	f.WriteInt(71, s.DimTol)
	f.WriteInt(72, s.DimLim)
	f.WriteInt(73, s.DimTih)
	f.WriteInt(74, s.DimToh)
	f.WriteInt(77, s.DimTad)
	f.WriteString(78, fmt.Sprintf("%d", s.DimZin))
	f.WriteString(79, fmt.Sprintf("%d", s.DimAzin))

	// Advanced options
	f.WriteString(170, fmt.Sprintf("%d", s.DimAlt))
	f.WriteString(171, fmt.Sprintf("%d", s.DimAltd))
	f.WriteString(172, fmt.Sprintf("%d", s.DimTofl))
	f.WriteString(173, fmt.Sprintf("%d", s.DimSah))
	f.WriteString(174, fmt.Sprintf("%d", s.DimTix))
	f.WriteString(175, fmt.Sprintf("%d", s.DimSoxd))
	f.WriteString(179, fmt.Sprintf("%d", s.DimAdec))
	f.WriteString(280, fmt.Sprintf("%d", s.DimJust))

	// R2007+ features
	if s.DimTfill != 0 {
		f.WriteString(69, fmt.Sprintf("%d", s.DimTfill))
		f.WriteString(70, fmt.Sprintf("%d", s.DimTfillclr))
	}
	if s.DimLtype != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, s.DimLtypeHandle)
		f.WriteString(102, "}")
	}
	if s.DimLtEx1 != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, s.DimLtEx1Handle)
		f.WriteString(102, "}")
	}
	if s.DimLtEx2 != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, s.DimLtEx2Handle)
		f.WriteString(102, "}")
	}
	f.WriteString(75, fmt.Sprintf("%d", s.DimSe1))
	f.WriteString(76, fmt.Sprintf("%d", s.DimSe2))
	f.WriteString(90, fmt.Sprintf("%d", s.DimArcsym))
	f.WriteString(173, fmt.Sprintf("%d", s.DimTix))

	// Final options
	f.WriteString(283, fmt.Sprintf("%d", s.DimTolj))
	f.WriteString(284, fmt.Sprintf("%d", s.DimTzin))
	f.WriteString(285, fmt.Sprintf("%d", s.DimAltz))
	f.WriteString(286, fmt.Sprintf("%d", s.DimAltzz))
	f.WriteString(288, fmt.Sprintf("%d", s.DimUpt))
	f.WriteString(289, fmt.Sprintf("%d", s.DimAtfit))
	f.WriteString(290, fmt.Sprintf("%d", s.DimFxlon))
	if s.DimFxL != 0.0 {
		f.WriteString(49, fmt.Sprintf("%.6f", s.DimFxL))
	}
	if s.DimJogang != 0.0 {
		f.WriteString(50, fmt.Sprintf("%.6f", s.DimJogang))
	}
	f.WriteString(142, fmt.Sprintf("%.6f", s.DimTsz))
	f.WriteString(143, fmt.Sprintf("%.6f", s.DimAlttf))
	f.WriteString(144, fmt.Sprintf("%.6f", s.DimLfac))
	if s.DimAltrnd != 0 {
		f.WriteString(148, fmt.Sprintf("%d", s.DimAltrnd))
	}
}

// BBox returns bounding box for dimstyle (n/a for table entity)
func (s *DimStyle) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}
