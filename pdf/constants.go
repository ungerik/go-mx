//go:generate go -C ../tools tool go-enum ../pdf/$GOFILE

package pdf

import (
	"fmt"
	"strings"
)

// Defaults applied by the Renderer constructors and Document so text can be
// drawn without explicit setup.
const (
	DefaultFontFamily = Helvetica
	DefaultFontSize   = 12.0
)

// Orientation is a page orientation passed to the Renderer constructors and to
// the Page component.
type Orientation string //#enum

const (
	// OrientationPortrait is the upright "portrait" page orientation (taller than wide).
	OrientationPortrait Orientation = "portrait"
	// OrientationLandscape is the sideways "landscape" page orientation (wider than tall).
	OrientationLandscape Orientation = "landscape"
)

// Valid indicates if o is any of the valid values for Orientation
func (o Orientation) Valid() bool {
	switch o {
	case
		OrientationPortrait,
		OrientationLandscape:
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
		OrientationPortrait,
		OrientationLandscape,
	}
}

// EnumStrings returns all valid values for Orientation as strings
func (Orientation) EnumStrings() []string {
	return []string{
		"portrait",
		"landscape",
	}
}

// String implements the fmt.Stringer interface for Orientation
func (o Orientation) String() string {
	return string(o)
}

// pageSize returns width and height in user units for size at orientation o.
func (o Orientation) pageSize(size Size) (w, h float64) {
	if o == OrientationLandscape {
		return size.Ht, size.Wd
	}
	return size.Wd, size.Ht
}

// Unit is the document measurement unit. All coordinates, widths and heights
// passed to components are expressed in this unit; font sizes are always in
// points regardless of the document unit. These four are the only distinct
// units fpdf supports. [New] accepts the canonical spellings below and legacy
// aliases "point" (for pt) and "in" (for inch), compared case-insensitively.
type Unit string //#enum

const (
	// UnitPoint is the typographic point unit ("pt", 1/72 inch).
	UnitPoint Unit = "pt"
	// UnitMillimeter is the millimeter unit ("mm").
	UnitMillimeter Unit = "mm"
	// UnitCentimeter is the centimeter unit ("cm").
	UnitCentimeter Unit = "cm"
	// UnitInch is the inch unit ("inch").
	UnitInch Unit = "inch"
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
		"pt",
		"mm",
		"cm",
		"inch",
	}
}

// String implements the fmt.Stringer interface for Unit
func (u Unit) String() string {
	return string(u)
}

// pointsPerUnit returns the number of points per document unit.
func (u Unit) pointsPerUnit() (float64, bool) {
	u, ok := NormalizeUnit(u)
	if !ok {
		return 0, false
	}
	switch u {
	case UnitPoint:
		return 1.0, true
	case UnitMillimeter:
		return 72.0 / 25.4, true
	case UnitCentimeter:
		return 72.0 / 2.54, true
	case UnitInch:
		return 72.0, true
	default:
		return 0, false
	}
}

// unitAliases maps legacy fpdf spellings (compared case-insensitively) to
// canonical [Unit] values.
var unitAliases = map[string]Unit{
	"point": UnitPoint,
	"in":    UnitInch,
}

// NormalizeUnit returns the canonical [Unit] and true when u is a known unit
// or legacy alias ("point", "in", compared case-insensitively). It is a plain
// constructor function, not a Unit method, so the go-enum code generation
// leaves it untouched.
func NormalizeUnit(u Unit) (Unit, bool) {
	if u.Valid() {
		return u, true
	}
	s := strings.ToLower(string(u))
	if canon, ok := unitAliases[s]; ok {
		return canon, true
	}
	for _, canon := range Unit("").Enums() {
		if strings.EqualFold(s, string(canon)) {
			return canon, true
		}
	}
	return "", false
}

// PageSize is a standard page size accepted by the Renderer constructors and
// [PageFormat]. Fully custom dimensions use [NewCustom] / [Renderer.AddPageFormat]
// with a [Size]. Names are matched case-insensitively (ISO sizes use an
// uppercase A and digit, US sizes use title case).
type PageSize string //#enum

const (
	// PageSizeA1 is the ISO 216 A1 page size (594 × 841 mm).
	PageSizeA1 PageSize = "A1" // 594 × 841 mm
	// PageSizeA2 is the ISO 216 A2 page size (420 × 594 mm).
	PageSizeA2 PageSize = "A2" // 420 × 594 mm
	// PageSizeA3 is the ISO 216 A3 page size (297 × 420 mm).
	PageSizeA3 PageSize = "A3" // 297 × 420 mm
	// PageSizeA4 is the ISO 216 A4 page size (210 × 297 mm).
	PageSizeA4 PageSize = "A4" // 210 × 297 mm
	// PageSizeA5 is the ISO 216 A5 page size (148 × 210 mm).
	PageSizeA5 PageSize = "A5" // 148 × 210 mm
	// PageSizeA6 is the ISO 216 A6 page size (105 × 148 mm).
	PageSizeA6 PageSize = "A6" // 105 × 148 mm
	// PageSizeA7 is the ISO 216 A7 page size (74 × 105 mm).
	PageSizeA7 PageSize = "A7" // 74 × 105 mm
	// PageSizeA8 is the ISO 216 A8 page size (52 × 74 mm).
	PageSizeA8 PageSize = "A8" // 52 × 74 mm
	// PageSizeLetter is the US Letter page size (8.5 × 11 in).
	PageSizeLetter PageSize = "Letter" // 8.5 × 11 in
	// PageSizeLegal is the US Legal page size (8.5 × 14 in).
	PageSizeLegal PageSize = "Legal" // 8.5 × 14 in
	// PageSizeTabloid is the US Tabloid/Ledger page size (11 × 17 in).
	PageSizeTabloid PageSize = "Tabloid" // 11 × 17 in
)

// Valid indicates if p is any of the valid values for PageSize
func (p PageSize) Valid() bool {
	switch p {
	case
		PageSizeA1,
		PageSizeA2,
		PageSizeA3,
		PageSizeA4,
		PageSizeA5,
		PageSizeA6,
		PageSizeA7,
		PageSizeA8,
		PageSizeLetter,
		PageSizeLegal,
		PageSizeTabloid:
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
		PageSizeA1,
		PageSizeA2,
		PageSizeA3,
		PageSizeA4,
		PageSizeA5,
		PageSizeA6,
		PageSizeA7,
		PageSizeA8,
		PageSizeLetter,
		PageSizeLegal,
		PageSizeTabloid,
	}
}

// EnumStrings returns all valid values for PageSize as strings
func (PageSize) EnumStrings() []string {
	return []string{
		"A1",
		"A2",
		"A3",
		"A4",
		"A5",
		"A6",
		"A7",
		"A8",
		"Letter",
		"Legal",
		"Tabloid",
	}
}

// String implements the fmt.Stringer interface for PageSize
func (p PageSize) String() string {
	return string(p)
}

// pageSizeByFold maps a lowercased page size name to its canonical [PageSize].
var pageSizeByFold = func() map[string]PageSize {
	enums := PageSize("").Enums()
	m := make(map[string]PageSize, len(enums))
	for _, p := range enums {
		m[strings.ToLower(string(p))] = p
	}
	return m
}()

// NormalizePageSize returns the canonical [PageSize] and true when p is a
// known name (matched case-insensitively, as fpdf did with strings.ToLower).
// It is a plain constructor function, not a PageSize method, so the go-enum
// code generation leaves it untouched.
func NormalizePageSize(p PageSize) (PageSize, bool) {
	if p.Valid() {
		return p, true
	}
	if canon, ok := pageSizeByFold[strings.ToLower(string(p))]; ok {
		return canon, true
	}
	return "", false
}

// Size returns the page width and height in points (1/72 inch) for p in
// portrait orientation. The second result is false when p is not a known size.
func (p PageSize) Size() (Size, bool) {
	p, ok := NormalizePageSize(p)
	if !ok {
		return Size{}, false
	}
	sz, ok := stdPageSizesPt[p]
	return sz, ok
}

// stdPageSizesPt maps each standard page size to its width and height in points.
var stdPageSizesPt = map[PageSize]Size{
	PageSizeA1:      {1683.78, 2383.94},
	PageSizeA2:      {1190.55, 1683.78},
	PageSizeA3:      {841.89, 1190.55},
	PageSizeA4:      {595.28, 841.89},
	PageSizeA5:      {420.94, 595.28},
	PageSizeA6:      {297.64, 420.94},
	PageSizeA7:      {209.76, 297.64},
	PageSizeA8:      {147.40, 209.76},
	PageSizeLetter:  {612, 792},
	PageSizeLegal:   {612, 1008},
	PageSizeTabloid: {792, 1224},
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

// BlendMode is a PDF extended graphics state blend mode (PDF 1.4+), as used by
// [Renderer.SetAlpha].
type BlendMode string //#enum

const (
	// BlendModeNormal is the default blend mode (source over destination).
	BlendModeNormal BlendMode = "Normal"
	// BlendModeMultiply multiplies source and destination color values.
	BlendModeMultiply BlendMode = "Multiply"
	// BlendModeScreen screens source and destination color values.
	BlendModeScreen BlendMode = "Screen"
	// BlendModeOverlay combines Multiply and Screen according to destination.
	BlendModeOverlay BlendMode = "Overlay"
	// BlendModeDarken keeps the darker of source and destination.
	BlendModeDarken BlendMode = "Darken"
	// BlendModeLighten keeps the lighter of source and destination.
	BlendModeLighten BlendMode = "Lighten"
	// BlendModeColorDodge brightens destination to reflect source.
	BlendModeColorDodge BlendMode = "ColorDodge"
	// BlendModeColorBurn darkens destination to reflect source.
	BlendModeColorBurn BlendMode = "ColorBurn"
	// BlendModeHardLight combines Multiply and Screen according to source.
	BlendModeHardLight BlendMode = "HardLight"
	// BlendModeSoftLight softens HardLight.
	BlendModeSoftLight BlendMode = "SoftLight"
	// BlendModeDifference subtracts darker from lighter channel values.
	BlendModeDifference BlendMode = "Difference"
	// BlendModeExclusion produces a lower-contrast Difference.
	BlendModeExclusion BlendMode = "Exclusion"
	// BlendModeHue uses the hue of source with destination saturation and luminosity.
	BlendModeHue BlendMode = "Hue"
	// BlendModeSaturation uses the saturation of source with destination hue and luminosity.
	BlendModeSaturation BlendMode = "Saturation"
	// BlendModeColor uses the hue and saturation of source with destination luminosity.
	BlendModeColor BlendMode = "Color"
	// BlendModeLuminosity uses the luminosity of source with destination hue and saturation.
	BlendModeLuminosity BlendMode = "Luminosity"
)

// Valid indicates if b is any of the valid values for BlendMode
func (b BlendMode) Valid() bool {
	switch b {
	case
		BlendModeNormal,
		BlendModeMultiply,
		BlendModeScreen,
		BlendModeOverlay,
		BlendModeDarken,
		BlendModeLighten,
		BlendModeColorDodge,
		BlendModeColorBurn,
		BlendModeHardLight,
		BlendModeSoftLight,
		BlendModeDifference,
		BlendModeExclusion,
		BlendModeHue,
		BlendModeSaturation,
		BlendModeColor,
		BlendModeLuminosity:
		return true
	}
	return false
}

// Validate returns an error if b is none of the valid values for BlendMode
func (b BlendMode) Validate() error {
	if !b.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.BlendMode", b)
	}
	return nil
}

// Enums returns all valid values for BlendMode
func (BlendMode) Enums() []BlendMode {
	return []BlendMode{
		BlendModeNormal,
		BlendModeMultiply,
		BlendModeScreen,
		BlendModeOverlay,
		BlendModeDarken,
		BlendModeLighten,
		BlendModeColorDodge,
		BlendModeColorBurn,
		BlendModeHardLight,
		BlendModeSoftLight,
		BlendModeDifference,
		BlendModeExclusion,
		BlendModeHue,
		BlendModeSaturation,
		BlendModeColor,
		BlendModeLuminosity,
	}
}

// EnumStrings returns all valid values for BlendMode as strings
func (BlendMode) EnumStrings() []string {
	return []string{
		"Normal",
		"Multiply",
		"Screen",
		"Overlay",
		"Darken",
		"Lighten",
		"ColorDodge",
		"ColorBurn",
		"HardLight",
		"SoftLight",
		"Difference",
		"Exclusion",
		"Hue",
		"Saturation",
		"Color",
		"Luminosity",
	}
}

// String implements the fmt.Stringer interface for BlendMode
func (b BlendMode) String() string {
	return string(b)
}

// AFRelationship expresses how an embedded file relates to the document it is
// attached to (the /AFRelationship key of a file specification, ISO 32000-2).
// An [Attachment] with a non-empty Relationship is listed in the document
// catalog's /AF array as an associated file, as PDF/A-3 hybrid formats like
// ZUGFeRD/Factur-X require. The empty string means the attachment is a plain
// embedded file without an associated-file declaration.
type AFRelationship string //#enum

const (
	// AFRelationshipSource declares the file as the source material of the document.
	AFRelationshipSource AFRelationship = "Source"
	// AFRelationshipData declares the file as data underlying the document,
	// used by Factur-X for the MINIMUM and BASIC WL profiles.
	AFRelationshipData AFRelationship = "Data"
	// AFRelationshipAlternative declares the file as an alternative
	// representation of the document, used by ZUGFeRD/Factur-X for the
	// machine-readable invoice XML.
	AFRelationshipAlternative AFRelationship = "Alternative"
	// AFRelationshipSupplement declares the file as supplemental material.
	AFRelationshipSupplement AFRelationship = "Supplement"
	// AFRelationshipEncryptedPayload declares the file as an encrypted payload.
	AFRelationshipEncryptedPayload AFRelationship = "EncryptedPayload"
	// AFRelationshipFormData declares the file as form data for the document.
	AFRelationshipFormData AFRelationship = "FormData"
	// AFRelationshipSchema declares the file as a schema for the document data.
	AFRelationshipSchema AFRelationship = "Schema"
	// AFRelationshipUnspecified declares no particular relationship.
	AFRelationshipUnspecified AFRelationship = "Unspecified"
)

// Valid indicates if a is any of the valid values for AFRelationship
func (a AFRelationship) Valid() bool {
	switch a {
	case
		AFRelationshipSource,
		AFRelationshipData,
		AFRelationshipAlternative,
		AFRelationshipSupplement,
		AFRelationshipEncryptedPayload,
		AFRelationshipFormData,
		AFRelationshipSchema,
		AFRelationshipUnspecified:
		return true
	}
	return false
}

// Validate returns an error if a is none of the valid values for AFRelationship
func (a AFRelationship) Validate() error {
	if !a.Valid() {
		return fmt.Errorf("invalid value %#v for type pdf.AFRelationship", a)
	}
	return nil
}

// Enums returns all valid values for AFRelationship
func (AFRelationship) Enums() []AFRelationship {
	return []AFRelationship{
		AFRelationshipSource,
		AFRelationshipData,
		AFRelationshipAlternative,
		AFRelationshipSupplement,
		AFRelationshipEncryptedPayload,
		AFRelationshipFormData,
		AFRelationshipSchema,
		AFRelationshipUnspecified,
	}
}

// EnumStrings returns all valid values for AFRelationship as strings
func (AFRelationship) EnumStrings() []string {
	return []string{
		"Source",
		"Data",
		"Alternative",
		"Supplement",
		"EncryptedPayload",
		"FormData",
		"Schema",
		"Unspecified",
	}
}

// String implements the fmt.Stringer interface for AFRelationship
func (a AFRelationship) String() string {
	return string(a)
}

// Pt builds a [Point] from x and y coordinates.
func Pt(x, y float64) Point {
	return Point{X: x, Y: y}
}
