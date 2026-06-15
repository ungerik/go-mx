//go:generate go -C ../tools tool go-enum ../pdf/$GOFILE

package pdf

import (
	"fmt"

	"codeberg.org/go-pdf/fpdf"
)

// Defaults applied by the Renderer constructors and Document so text can be
// drawn without explicit setup.
const (
	DefaultFontFamily = Helvetica
	DefaultFontSize   = 12.0
)

// Orientation is a page orientation passed to the Renderer constructors and to
// the Page component. fpdf matches on the first letter, case-insensitively, so
// Portrait and Landscape are the only two distinct orientations.
type Orientation string //#enum

const (
	// Portrait is the upright "portrait" page orientation (taller than wide).
	Portrait Orientation = fpdf.OrientationPortrait // "portrait"
	// Landscape is the sideways "landscape" page orientation (wider than tall).
	Landscape Orientation = fpdf.OrientationLandscape // "landscape"
)

// Valid indicates if o is any of the valid values for Orientation
func (o Orientation) Valid() bool {
	switch o {
	case
		Portrait,
		Landscape:
		return true
	}
	return false
}

// Validate returns an error if o is none of the valid values for Orientation
func (o Orientation) Validate() error {
	if !o.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.Orientation", o)
	}
	return nil
}

// Enums returns all valid values for Orientation
func (Orientation) Enums() []Orientation {
	return []Orientation{
		Portrait,
		Landscape,
	}
}

// EnumStrings returns all valid values for Orientation as strings
func (Orientation) EnumStrings() []string {
	return []string{
		fpdf.OrientationPortrait,
		fpdf.OrientationLandscape,
	}
}

// String implements the fmt.Stringer interface for Orientation
func (o Orientation) String() string {
	return string(o)
}

// Unit is the document measurement unit. All coordinates, widths and heights
// passed to components are expressed in this unit; font sizes are always in
// points regardless of the document unit. These four are the only distinct
// units fpdf supports.
type Unit string //#enum

const (
	// UnitPoint is the typographic point unit ("pt", 1/72 inch).
	UnitPoint Unit = fpdf.UnitPoint // "pt"
	// UnitMillimeter is the millimeter unit ("mm").
	UnitMillimeter Unit = fpdf.UnitMillimeter // "mm"
	// UnitCentimeter is the centimeter unit ("cm").
	UnitCentimeter Unit = fpdf.UnitCentimeter // "cm"
	// UnitInch is the inch unit ("inch").
	UnitInch Unit = fpdf.UnitInch // "inch"
)

// Valid indicates if u is any of the valid values for Unit
func (u Unit) Valid() bool {
	switch u {
	case
		UnitPoint,
		UnitMillimeter,
		UnitCentimeter,
		UnitInch:
		return true
	}
	return false
}

// Validate returns an error if u is none of the valid values for Unit
func (u Unit) Validate() error {
	if !u.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.Unit", u)
	}
	return nil
}

// Enums returns all valid values for Unit
func (Unit) Enums() []Unit {
	return []Unit{
		UnitPoint,
		UnitMillimeter,
		UnitCentimeter,
		UnitInch,
	}
}

// EnumStrings returns all valid values for Unit as strings
func (Unit) EnumStrings() []string {
	return []string{
		fpdf.UnitPoint,
		fpdf.UnitMillimeter,
		fpdf.UnitCentimeter,
		fpdf.UnitInch,
	}
}

// String implements the fmt.Stringer interface for Unit
func (u Unit) String() string {
	return string(u)
}

// PageSize is a standard page size accepted by fpdf's string-based page setup.
// The constants below are all of them; fully custom dimensions go through the
// raw fpdf API (fpdf.NewCustom / AddPageFormat with a SizeType). fpdf
// lower-cases the value, so these upper-case forms and their lower-case
// spellings are equivalent.
type PageSize string //#enum

const (
	// A1 is the ISO 216 A1 page size (594 × 841 mm).
	A1 PageSize = "A1" // 594 × 841 mm
	// A2 is the ISO 216 A2 page size (420 × 594 mm).
	A2 PageSize = "A2" // 420 × 594 mm
	// A3 is the ISO 216 A3 page size (297 × 420 mm).
	A3 PageSize = fpdf.PageSizeA3 // 297 × 420 mm
	// A4 is the ISO 216 A4 page size (210 × 297 mm).
	A4 PageSize = fpdf.PageSizeA4 // 210 × 297 mm
	// A5 is the ISO 216 A5 page size (148 × 210 mm).
	A5 PageSize = fpdf.PageSizeA5 // 148 × 210 mm
	// A6 is the ISO 216 A6 page size (105 × 148 mm).
	A6 PageSize = "A6" // 105 × 148 mm
	// A7 is the ISO 216 A7 page size (74 × 105 mm).
	A7 PageSize = "A7" // 74 × 105 mm
	// A8 is the ISO 216 A8 page size (52 × 74 mm).
	A8 PageSize = "A8" // 52 × 74 mm
	// Letter is the US Letter page size (8.5 × 11 in).
	Letter PageSize = fpdf.PageSizeLetter // 8.5 × 11 in
	// Legal is the US Legal page size (8.5 × 14 in).
	Legal PageSize = fpdf.PageSizeLegal // 8.5 × 14 in
	// Tabloid is the US Tabloid/Ledger page size (11 × 17 in).
	Tabloid PageSize = "Tabloid" // 11 × 17 in
)

// Valid indicates if p is any of the valid values for PageSize
func (p PageSize) Valid() bool {
	switch p {
	case
		A1,
		A2,
		A3,
		A4,
		A5,
		A6,
		A7,
		A8,
		Letter,
		Legal,
		Tabloid:
		return true
	}
	return false
}

// Validate returns an error if p is none of the valid values for PageSize
func (p PageSize) Validate() error {
	if !p.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.PageSize", p)
	}
	return nil
}

// Enums returns all valid values for PageSize
func (PageSize) Enums() []PageSize {
	return []PageSize{
		A1,
		A2,
		A3,
		A4,
		A5,
		A6,
		A7,
		A8,
		Letter,
		Legal,
		Tabloid,
	}
}

// EnumStrings returns all valid values for PageSize as strings
func (PageSize) EnumStrings() []string {
	return []string{
		"A1",
		"A2",
		fpdf.PageSizeA3,
		fpdf.PageSizeA4,
		fpdf.PageSizeA5,
		"A6",
		"A7",
		"A8",
		fpdf.PageSizeLetter,
		fpdf.PageSizeLegal,
		"Tabloid",
	}
}

// String implements the fmt.Stringer interface for PageSize
func (p PageSize) String() string {
	return string(p)
}

// Standard (core) font families, which need no font files. These are untyped
// string constants because any family registered with the raw fpdf AddFont /
// AddUTF8Font API (or [Renderer.LoadUTF8FontBytes]) is an equally valid family
// name and can be passed as a plain string.
const (
	Helvetica    = "Helvetica"    // sans-serif (alias of Arial)
	Arial        = "Arial"        // sans-serif (alias of Helvetica)
	Times        = "Times"        // serif
	Courier      = "Courier"      // fixed-width
	Symbol       = "Symbol"       // symbolic
	ZapfDingbats = "ZapfDingbats" // symbolic
)

// FontStyle selects bold/italic/underline/strike-out. The four are independent
// fpdf flags (B, I, U, S); every combination is enumerated below in canonical
// B-I-U-S order, so the values serve both as a closed enum and as the result of
// concatenating the single-flag constants — StyleBold + StyleItalic equals
// StyleBoldItalic. Concatenate in B-I-U-S order to stay within the enumerated
// set. Bold and italic do not apply to the Symbol and ZapfDingbats families.
type FontStyle string //#enum

const (
	// StyleRegular is the unstyled (regular) font style.
	StyleRegular FontStyle = "" // regular
	// StyleBold is the bold (B) font style.
	StyleBold FontStyle = "B" // bold
	// StyleItalic is the italic (I) font style.
	StyleItalic FontStyle = "I" // italic
	// StyleUnderline is the underline (U) font style.
	StyleUnderline FontStyle = "U" // underline
	// StyleStrikeOut is the strike-out (S) font style.
	StyleStrikeOut FontStyle = "S" // strike-out

	// StyleBoldItalic combines the bold and italic styles (BI).
	StyleBoldItalic FontStyle = "BI" // bold + italic
	// StyleBoldUnderline combines the bold and underline styles (BU).
	StyleBoldUnderline FontStyle = "BU" // bold + underline
	// StyleBoldStrikeOut combines the bold and strike-out styles (BS).
	StyleBoldStrikeOut FontStyle = "BS" // bold + strike-out
	// StyleItalicUnderline combines the italic and underline styles (IU).
	StyleItalicUnderline FontStyle = "IU" // italic + underline
	// StyleItalicStrikeOut combines the italic and strike-out styles (IS).
	StyleItalicStrikeOut FontStyle = "IS" // italic + strike-out
	// StyleUnderlineStrikeOut combines the underline and strike-out styles (US).
	StyleUnderlineStrikeOut FontStyle = "US" // underline + strike-out

	// StyleBoldItalicUnderline combines the bold, italic and underline styles (BIU).
	StyleBoldItalicUnderline FontStyle = "BIU" // bold + italic + underline
	// StyleBoldItalicStrikeOut combines the bold, italic and strike-out styles (BIS).
	StyleBoldItalicStrikeOut FontStyle = "BIS" // bold + italic + strike-out
	// StyleBoldUnderlineStrikeOut combines the bold, underline and strike-out styles (BUS).
	StyleBoldUnderlineStrikeOut FontStyle = "BUS" // bold + underline + strike-out
	// StyleItalicUnderlineStrikeOut combines the italic, underline and strike-out styles (IUS).
	StyleItalicUnderlineStrikeOut FontStyle = "IUS" // italic + underline + strike-out

	// StyleBoldItalicUnderlineStrikeOut combines all four styles (BIUS).
	StyleBoldItalicUnderlineStrikeOut FontStyle = "BIUS" // all four
)

// Valid indicates if f is any of the valid values for FontStyle
func (f FontStyle) Valid() bool {
	switch f {
	case
		StyleRegular,
		StyleBold,
		StyleItalic,
		StyleUnderline,
		StyleStrikeOut,
		StyleBoldItalic,
		StyleBoldUnderline,
		StyleBoldStrikeOut,
		StyleItalicUnderline,
		StyleItalicStrikeOut,
		StyleUnderlineStrikeOut,
		StyleBoldItalicUnderline,
		StyleBoldItalicStrikeOut,
		StyleBoldUnderlineStrikeOut,
		StyleItalicUnderlineStrikeOut,
		StyleBoldItalicUnderlineStrikeOut:
		return true
	}
	return false
}

// Validate returns an error if f is none of the valid values for FontStyle
func (f FontStyle) Validate() error {
	if !f.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.FontStyle", f)
	}
	return nil
}

// Enums returns all valid values for FontStyle
func (FontStyle) Enums() []FontStyle {
	return []FontStyle{
		StyleRegular,
		StyleBold,
		StyleItalic,
		StyleUnderline,
		StyleStrikeOut,
		StyleBoldItalic,
		StyleBoldUnderline,
		StyleBoldStrikeOut,
		StyleItalicUnderline,
		StyleItalicStrikeOut,
		StyleUnderlineStrikeOut,
		StyleBoldItalicUnderline,
		StyleBoldItalicStrikeOut,
		StyleBoldUnderlineStrikeOut,
		StyleItalicUnderlineStrikeOut,
		StyleBoldItalicUnderlineStrikeOut,
	}
}

// EnumStrings returns all valid values for FontStyle as strings
func (FontStyle) EnumStrings() []string {
	return []string{
		"",
		"B",
		"I",
		"U",
		"S",
		"BI",
		"BU",
		"BS",
		"IU",
		"IS",
		"US",
		"BIU",
		"BIS",
		"BUS",
		"IUS",
		"BIUS",
	}
}

// String implements the fmt.Stringer interface for FontStyle
func (f FontStyle) String() string {
	return string(f)
}

// HAlign is the horizontal text alignment inside a cell. Horizontal and vertical
// alignment are two independent single-choice axes rather than a flag set, so
// they are modelled as two enums ([HAlign] and [VAlign]) instead of one large
// combined enumeration. fpdf treats left as the default, so AlignLeft and the
// empty value render identically.
type HAlign string //#enum

const (
	// AlignLeft aligns text to the left edge of the cell (fpdf default).
	AlignLeft HAlign = "L" // left (default)
	// AlignCenter centers text horizontally within the cell.
	AlignCenter HAlign = "C" // center
	// AlignRight aligns text to the right edge of the cell.
	AlignRight HAlign = "R" // right
	// AlignJustify justifies text to both edges of the cell (MultiCell only).
	AlignJustify HAlign = "J" // justified (MultiCell only)
)

// Valid indicates if h is any of the valid values for HAlign
func (h HAlign) Valid() bool {
	switch h {
	case
		AlignLeft,
		AlignCenter,
		AlignRight,
		AlignJustify:
		return true
	}
	return false
}

// Validate returns an error if h is none of the valid values for HAlign
func (h HAlign) Validate() error {
	if !h.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.HAlign", h)
	}
	return nil
}

// Enums returns all valid values for HAlign
func (HAlign) Enums() []HAlign {
	return []HAlign{
		AlignLeft,
		AlignCenter,
		AlignRight,
		AlignJustify,
	}
}

// EnumStrings returns all valid values for HAlign as strings
func (HAlign) EnumStrings() []string {
	return []string{
		"L",
		"C",
		"R",
		"J",
	}
}

// String implements the fmt.Stringer interface for HAlign
func (h HAlign) String() string {
	return string(h)
}

// VAlign is the vertical text alignment inside a cell. fpdf treats middle as the
// default, so AlignMiddle and the empty value render identically.
type VAlign string //#enum

const (
	// AlignTop aligns text to the top edge of the cell.
	AlignTop VAlign = "T" // top
	// AlignMiddle centers text vertically within the cell (fpdf default).
	AlignMiddle VAlign = "M" // middle (default)
	// AlignBottom aligns text to the bottom edge of the cell.
	AlignBottom VAlign = "B" // bottom
	// AlignBaseline aligns text to the font baseline.
	AlignBaseline VAlign = "A" // baseline
)

// Valid indicates if v is any of the valid values for VAlign
func (v VAlign) Valid() bool {
	switch v {
	case
		AlignTop,
		AlignMiddle,
		AlignBottom,
		AlignBaseline:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for VAlign
func (v VAlign) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.VAlign", v)
	}
	return nil
}

// Enums returns all valid values for VAlign
func (VAlign) Enums() []VAlign {
	return []VAlign{
		AlignTop,
		AlignMiddle,
		AlignBottom,
		AlignBaseline,
	}
}

// EnumStrings returns all valid values for VAlign as strings
func (VAlign) EnumStrings() []string {
	return []string{
		"T",
		"M",
		"B",
		"A",
	}
}

// String implements the fmt.Stringer interface for VAlign
func (v VAlign) String() string {
	return string(v)
}

// Border selects which cell edges are stroked. The four edges are independent
// fpdf flags (L, T, R, B); every combination is enumerated below in canonical
// L-T-R-B order, so the values serve both as a closed enum and as the result of
// concatenating the single-edge constants — BorderLeft + BorderRight equals
// BorderLeftRight. Concatenate in L-T-R-B order to stay within the enumerated
// set. BorderNone strokes nothing; BorderFull is fpdf's "1" shorthand that
// strokes all four edges as a single rectangle — a more compact content stream
// than the otherwise equivalent BorderLeftTopRightBottom, which strokes four
// separate lines.
type Border string //#enum

const (
	// BorderNone strokes no cell edge.
	BorderNone Border = "" // no border
	// BorderFull strokes all four edges as a single rectangle (fpdf's "1" shorthand).
	BorderFull Border = "1" // all four edges as a single rectangle stroke

	// BorderLeft strokes the left edge.
	BorderLeft Border = "L" // left edge
	// BorderTop strokes the top edge.
	BorderTop Border = "T" // top edge
	// BorderRight strokes the right edge.
	BorderRight Border = "R" // right edge
	// BorderBottom strokes the bottom edge.
	BorderBottom Border = "B" // bottom edge

	// BorderLeftTop strokes the left and top edges.
	BorderLeftTop Border = "LT" // left + top
	// BorderLeftRight strokes the left and right edges.
	BorderLeftRight Border = "LR" // left + right
	// BorderLeftBottom strokes the left and bottom edges.
	BorderLeftBottom Border = "LB" // left + bottom
	// BorderTopRight strokes the top and right edges.
	BorderTopRight Border = "TR" // top + right
	// BorderTopBottom strokes the top and bottom edges.
	BorderTopBottom Border = "TB" // top + bottom
	// BorderRightBottom strokes the right and bottom edges.
	BorderRightBottom Border = "RB" // right + bottom

	// BorderLeftTopRight strokes the left, top and right edges.
	BorderLeftTopRight Border = "LTR" // left + top + right
	// BorderLeftTopBottom strokes the left, top and bottom edges.
	BorderLeftTopBottom Border = "LTB" // left + top + bottom
	// BorderLeftRightBottom strokes the left, right and bottom edges.
	BorderLeftRightBottom Border = "LRB" // left + right + bottom
	// BorderTopRightBottom strokes the top, right and bottom edges.
	BorderTopRightBottom Border = "TRB" // top + right + bottom

	// BorderLeftTopRightBottom strokes all four edges as separate lines.
	BorderLeftTopRightBottom Border = "LTRB" // all four edges as separate lines
)

// Valid indicates if b is any of the valid values for Border
func (b Border) Valid() bool {
	switch b {
	case
		BorderNone,
		BorderFull,
		BorderLeft,
		BorderTop,
		BorderRight,
		BorderBottom,
		BorderLeftTop,
		BorderLeftRight,
		BorderLeftBottom,
		BorderTopRight,
		BorderTopBottom,
		BorderRightBottom,
		BorderLeftTopRight,
		BorderLeftTopBottom,
		BorderLeftRightBottom,
		BorderTopRightBottom,
		BorderLeftTopRightBottom:
		return true
	}
	return false
}

// Validate returns an error if b is none of the valid values for Border
func (b Border) Validate() error {
	if !b.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.Border", b)
	}
	return nil
}

// Enums returns all valid values for Border
func (Border) Enums() []Border {
	return []Border{
		BorderNone,
		BorderFull,
		BorderLeft,
		BorderTop,
		BorderRight,
		BorderBottom,
		BorderLeftTop,
		BorderLeftRight,
		BorderLeftBottom,
		BorderTopRight,
		BorderTopBottom,
		BorderRightBottom,
		BorderLeftTopRight,
		BorderLeftTopBottom,
		BorderLeftRightBottom,
		BorderTopRightBottom,
		BorderLeftTopRightBottom,
	}
}

// EnumStrings returns all valid values for Border as strings
func (Border) EnumStrings() []string {
	return []string{
		"",
		"1",
		"L",
		"T",
		"R",
		"B",
		"LT",
		"LR",
		"LB",
		"TR",
		"TB",
		"RB",
		"LTR",
		"LTB",
		"LRB",
		"TRB",
		"LTRB",
	}
}

// String implements the fmt.Stringer interface for Border
func (b Border) String() string {
	return string(b)
}

// DrawOp is the paint operation for vector shapes: stroke the outline, fill the
// interior, or both — the three canonical combinations of fill and stroke. fpdf
// also accepts the even-odd variants ("F*", "FD*") and raw PDF path operators,
// reachable through the embedded renderer for the rare cases that need them.
type DrawOp string //#enum

const (
	// Stroke strokes the shape outline with the draw color.
	Stroke DrawOp = "D" // stroke the outline with the draw color
	// FillShape fills the shape interior with the fill color.
	FillShape DrawOp = "F" // fill the interior with the fill color
	// FillStroke fills the interior and then strokes the outline.
	FillStroke DrawOp = "FD" // fill, then stroke
)

// Valid indicates if d is any of the valid values for DrawOp
func (d DrawOp) Valid() bool {
	switch d {
	case
		Stroke,
		FillShape,
		FillStroke:
		return true
	}
	return false
}

// Validate returns an error if d is none of the valid values for DrawOp
func (d DrawOp) Validate() error {
	if !d.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.DrawOp", d)
	}
	return nil
}

// Enums returns all valid values for DrawOp
func (DrawOp) Enums() []DrawOp {
	return []DrawOp{
		Stroke,
		FillShape,
		FillStroke,
	}
}

// EnumStrings returns all valid values for DrawOp as strings
func (DrawOp) EnumStrings() []string {
	return []string{
		"D",
		"F",
		"FD",
	}
}

// String implements the fmt.Stringer interface for DrawOp
func (d DrawOp) String() string {
	return string(d)
}

// LnPos selects where the cursor moves after a Cell, matching fpdf's ln
// argument.
type LnPos int //#enum

const (
	// LnRight moves the cursor to the right of the cell (fpdf default).
	LnRight LnPos = 0 // to the right of the cell (default)
	// LnNewline moves the cursor to the start of the next line.
	LnNewline LnPos = 1 // to the start of the next line
	// LnBelow moves the cursor directly below the cell.
	LnBelow LnPos = 2 // directly below the cell
)

// Valid indicates if l is any of the valid values for LnPos
func (l LnPos) Valid() bool {
	switch l {
	case
		LnRight,
		LnNewline,
		LnBelow:
		return true
	}
	return false
}

// Validate returns an error if l is none of the valid values for LnPos
func (l LnPos) Validate() error {
	if !l.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.LnPos", l)
	}
	return nil
}

// Enums returns all valid values for LnPos
func (LnPos) Enums() []LnPos {
	return []LnPos{
		LnRight,
		LnNewline,
		LnBelow,
	}
}

// EnumStrings returns all valid values for LnPos as strings
func (LnPos) EnumStrings() []string {
	return []string{
		"0",
		"1",
		"2",
	}
}

// LineCapStyle is the shape drawn at the ends of open paths.
type LineCapStyle string //#enum

const (
	// CapButt ends open paths flush at the endpoint with no extension.
	CapButt LineCapStyle = "butt"
	// CapRound ends open paths with a semicircular cap centered on the endpoint.
	CapRound LineCapStyle = "round"
	// CapSquare ends open paths with a square cap extending half the line width past the endpoint.
	CapSquare LineCapStyle = "square"
)

// Valid indicates if l is any of the valid values for LineCapStyle
func (l LineCapStyle) Valid() bool {
	switch l {
	case
		CapButt,
		CapRound,
		CapSquare:
		return true
	}
	return false
}

// Validate returns an error if l is none of the valid values for LineCapStyle
func (l LineCapStyle) Validate() error {
	if !l.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.LineCapStyle", l)
	}
	return nil
}

// Enums returns all valid values for LineCapStyle
func (LineCapStyle) Enums() []LineCapStyle {
	return []LineCapStyle{
		CapButt,
		CapRound,
		CapSquare,
	}
}

// EnumStrings returns all valid values for LineCapStyle as strings
func (LineCapStyle) EnumStrings() []string {
	return []string{
		"butt",
		"round",
		"square",
	}
}

// String implements the fmt.Stringer interface for LineCapStyle
func (l LineCapStyle) String() string {
	return string(l)
}

// LineJoinStyle is the shape drawn where path segments meet.
type LineJoinStyle string //#enum

const (
	// JoinMiter joins segments with a sharp, extended corner.
	JoinMiter LineJoinStyle = "miter"
	// JoinRound joins segments with a rounded corner.
	JoinRound LineJoinStyle = "round"
	// JoinBevel joins segments with a flattened corner.
	JoinBevel LineJoinStyle = "bevel"
)

// Valid indicates if l is any of the valid values for LineJoinStyle
func (l LineJoinStyle) Valid() bool {
	switch l {
	case
		JoinMiter,
		JoinRound,
		JoinBevel:
		return true
	}
	return false
}

// Validate returns an error if l is none of the valid values for LineJoinStyle
func (l LineJoinStyle) Validate() error {
	if !l.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.LineJoinStyle", l)
	}
	return nil
}

// Enums returns all valid values for LineJoinStyle
func (LineJoinStyle) Enums() []LineJoinStyle {
	return []LineJoinStyle{
		JoinMiter,
		JoinRound,
		JoinBevel,
	}
}

// EnumStrings returns all valid values for LineJoinStyle as strings
func (LineJoinStyle) EnumStrings() []string {
	return []string{
		"miter",
		"round",
		"bevel",
	}
}

// String implements the fmt.Stringer interface for LineJoinStyle
func (l LineJoinStyle) String() string {
	return string(l)
}

// ImageType identifies the encoding of an in-memory image passed to
// [ImageReader] / [ImageBytes], where there is no filename to infer it from.
// fpdf supports these three raster formats only.
type ImageType string //#enum

const (
	// ImagePNG is the PNG image encoding.
	ImagePNG ImageType = "png"
	// ImageJPEG is the JPEG image encoding.
	ImageJPEG ImageType = "jpg"
	// ImageGIF is the GIF image encoding.
	ImageGIF ImageType = "gif"
)

// Valid indicates if i is any of the valid values for ImageType
func (i ImageType) Valid() bool {
	switch i {
	case
		ImagePNG,
		ImageJPEG,
		ImageGIF:
		return true
	}
	return false
}

// Validate returns an error if i is none of the valid values for ImageType
func (i ImageType) Validate() error {
	if !i.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.ImageType", i)
	}
	return nil
}

// Enums returns all valid values for ImageType
func (ImageType) Enums() []ImageType {
	return []ImageType{
		ImagePNG,
		ImageJPEG,
		ImageGIF,
	}
}

// EnumStrings returns all valid values for ImageType as strings
func (ImageType) EnumStrings() []string {
	return []string{
		"png",
		"jpg",
		"gif",
	}
}

// String implements the fmt.Stringer interface for ImageType
func (i ImageType) String() string {
	return string(i)
}

// Point is a coordinate in document units, re-exported from fpdf so it can be
// passed straight to the embedded renderer's polygon and curve methods.
type Point = fpdf.PointType

// Pt builds a [Point] from x and y coordinates.
func Pt(x, y float64) Point {
	return Point{X: x, Y: y}
}
