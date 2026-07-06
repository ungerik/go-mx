// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
 * Copyright (c) 2013-2014 Kurt Jung (Gmail: kurt.w.jung)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package pdf

// Version: 1.7
// Date:    2011-06-18
// Author:  Olivier PLATHEY
// Port to Go: Kurt Jung, 2013-07-15

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"maps"
	"math"
	"os"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/domonda/go-errs"
)

// Renderer is the principal structure for creating a single PDF document
type Renderer struct {
	isCurrentUTF8    bool                       // is current font used in utf-8 mode
	isRTL            bool                       // is is right to left mode enabled
	page             int                        // current page number
	n                int                        // current object number
	offsets          []int                      // array of object offsets
	buffer           bytes.Buffer               // buffer holding in-memory PDF
	pages            []*bytes.Buffer            // slice[page] of page content; 1-based
	state            int                        // current document state
	compress         bool                       // compression flag
	k                float64                    // scale factor (number of points in user unit)
	defOrientation   Orientation                // default orientation
	curOrientation   Orientation                // current orientation
	defPageSize      Size                       // default page size
	defPageBoxes     map[string]PageBox         // default page size
	curPageSize      Size                       // current page size
	pageSizes        map[int]Size               // used for pages with non default sizes or orientations
	pageBoxes        map[int]map[string]PageBox // used to define the crop, trim, bleed and art boxes
	unit             Unit                       // unit of measure for all rendered objects except fonts
	wPt, hPt         float64                    // dimensions of current page in points
	w, h             float64                    // dimensions of current page in user unit
	lMargin          float64                    // left margin
	tMargin          float64                    // top margin
	rMargin          float64                    // right margin
	bMargin          float64                    // page break margin
	cMargin          float64                    // cell margin
	x, y             float64                    // current position in user unit
	lasth            float64                    // height of last printed cell
	lineWidth        float64                    // line width in user unit
	fontpath         string                     // path containing fonts
	fontLoader       FontLoader                 // used to load font files from arbitrary locations
	coreFonts        map[string]bool            // array of core font names
	fonts            map[string]fontDef         // array of used fonts
	fontFiles        map[string]fontFile        // array of font files
	diffs            []string                   // array of encoding differences
	fontFamily       string                     // current font family
	fontStyle        string                     // current font style
	underline        bool                       // underlining flag
	strikeout        bool                       // strike out flag
	currentFont      fontDef                    // current font info
	fontSizePt       float64                    // current font size in points
	fontSize         float64                    // current font size in user unit
	ws               float64                    // word spacing
	images           map[string]*ImageInfo      // array of used images
	aliasMap         map[string]string          // map of alias->replacement
	pageLinks        [][]pageLink               // pageLinks[page][link], both 1-based
	links            []internalPageLink         // array of internal links
	attachments      []Attachment               // slice of content to embed globally
	pageAttachments  [][]annotationAttach       // 1-based array of annotation for file attachments (per page)
	outlines         []outline                  // array of outlines
	outlineRoot      int                        // root of outlines
	autoPageBreak    bool                       // automatic page breaking
	acceptPageBreak  func() bool                // returns true to accept page break
	pageBreakTrigger float64                    // threshold used to trigger page breaks
	inHeader         bool                       // flag set when processing header
	headerFnc        func()                     // function provided by app and called to write header
	headerHomeMode   bool                       // set position to home after headerFnc is called
	inFooter         bool                       // flag set when processing footer
	footerFnc        func()                     // function provided by app and called to write footer
	footerFncLpi     func(bool)                 // function provided by app and called to write footer with last page flag
	zoomMode         string                     // zoom display mode
	layoutMode       string                     // layout display mode
	nXMP             int                        // XMP object number
	xmp              []byte                     // XMP metadata
	producer         string                     // producer
	title            string                     // title
	subject          string                     // subject
	author           string                     // author
	lang             string                     // lang
	keywords         string                     // keywords
	creator          string                     // creator
	creationDate     time.Time                  // override for document CreationDate value
	modDate          time.Time                  // override for document ModDate value
	aliasNbPagesStr  string                     // alias for total number of pages
	pdfVersion       pdfVersion                 // PDF version number
	capStyle         int                        // line cap style: butt 0, round 1, square 2
	joinStyle        int                        // line segment join style: miter 0, round 1, bevel 2
	dashArray        []float64                  // dash array
	dashPhase        float64                    // dash phase
	blendList        []blendMode                // slice[idx] of alpha transparency modes, 1-based
	blendMap         map[string]int             // map into blendList
	blendMode        BlendMode                  // current blend mode
	alpha            float64                    // current transpacency
	gradientList     []gradient                 // slice[idx] of gradient records
	clipNest         int                        // Number of active clipping contexts
	transformNest    int                        // Number of active transformation contexts
	err              error                      // Set if error occurs during life cycle of instance
	protect          protect                    // document protection structure
	layer            layerRec                   // manages optional layers in document
	catalogSort      bool                       // sort resource catalogs in document
	nJs              int                        // JavaScript object number
	javascript       *string                    // JavaScript code to include in the PDF
	colorFlag        bool                       // indicates whether fill and text colors are different
	color            struct {
		// Composite values of colors
		draw, fill, text colorState
	}
	spotColorMap           map[string]spotColor // Map of named ink-based colors
	outputIntents          []OutputIntent       // OutputIntents
	outputIntentStartN     int                  // Start object number for
	userUnderlineThickness float64              // A custom user underline thickness multiplier.

	// translate maps UTF-8 to the encoding of the current standard (core)
	// font, which uses cp1252 (Western Europe). Applied automatically by the
	// text components via tr. Has no effect on UTF-8 fonts added with
	// AddUTF8Font; reset it with SetTranslator when switching font encodings.
	translate func(string) string
	// lineHeight is the default line height in document units used by flowing
	// text components (Text, Paragraph, NewLine) when no explicit height is
	// given. Zero means "derive from the current font size".
	lineHeight float64
}

var gl struct {
	catalogSort  bool
	noCompress   bool // Initial zero value indicates compression
	creationDate time.Time
	modDate      time.Time
}

func newRenderer(orientation Orientation, unit Unit, pageSize PageSize, fontDirStr string, size Size) (r *Renderer) {
	orientation = cmp.Or(orientation, OrientationPortrait)
	if err := orientation.Validate(); err != nil {
		return &Renderer{err: err}
	}
	normalizedUnit, ok := NormalizeUnit(cmp.Or(unit, UnitMillimeter))
	if !ok {
		return &Renderer{err: fmt.Errorf("invalid value %#v for type pdf.Unit", unit)}
	}
	unit = normalizedUnit
	pageSize = cmp.Or(pageSize, PageSizeA4)
	if size.Wd <= 0 || size.Ht <= 0 {
		normalizedPageSize, ok := NormalizePageSize(pageSize)
		if !ok {
			return &Renderer{err: fmt.Errorf("invalid value %#v for type pdf.PageSize", pageSize)}
		}
		pageSize = normalizedPageSize
	}
	fontDirStr = cmp.Or(fontDirStr, ".")

	k, _ := unit.pointsPerUnit()
	defPageSize := size
	if size.Wd <= 0 || size.Ht <= 0 {
		pt, ok := pageSize.Size()
		if !ok {
			return &Renderer{err: fmt.Errorf("unknown page size %s", pageSize)}
		}
		defPageSize = Size{pt.Wd / k, pt.Ht / k}
	}
	w, h := orientation.pageSize(defPageSize)
	margin := 28.35 / k

	r = &Renderer{
		n:               2,
		pages:           []*bytes.Buffer{bytes.NewBufferString("")}, // pages[0] is unused (1-based)
		pageSizes:       make(map[int]Size),
		pageBoxes:       make(map[int]map[string]PageBox),
		defPageBoxes:    make(map[string]PageBox),
		fonts:           make(map[string]fontDef),
		fontFiles:       make(map[string]fontFile),
		diffs:           make([]string, 0, 8),
		images:          make(map[string]*ImageInfo),
		pageLinks:       [][]pageLink{{}},         // pageLinks[0] is unused (1-based)
		links:           []internalPageLink{{}},   // links[0] is unused (1-based)
		pageAttachments: [][]annotationAttach{{}}, //
		aliasMap:        make(map[string]string),
		fontpath:        fontDirStr,
		coreFonts: map[string]bool{
			"courier":      true,
			"helvetica":    true,
			"times":        true,
			"symbol":       true,
			"zapfdingbats": true,
		},
		fontSizePt:       12,
		fontSize:         12 / k,
		k:                k,
		unit:             unit,
		defPageSize:      defPageSize,
		curPageSize:      defPageSize,
		defOrientation:   orientation,
		curOrientation:   orientation,
		w:                w,
		h:                h,
		wPt:              w * k,
		hPt:              h * k,
		lMargin:          margin,
		tMargin:          margin,
		rMargin:          margin,
		cMargin:          margin / 10,
		lineWidth:        0.567 / k,
		autoPageBreak:    true,
		bMargin:          2 * margin,
		pageBreakTrigger: h - 2*margin,
		zoomMode:         "default",
		layoutMode:       "default",
		compress:         !gl.noCompress,
		spotColorMap:     make(map[string]spotColor),
		blendList:        make([]blendMode, 1, 8), // blendList[0] is unused (1-based)
		blendMap:         make(map[string]int),
		blendMode:        BlendModeNormal,
		alpha:            1,
		gradientList:     make([]gradient, 1, 8), // gradientList[0] is unused (1-based)
		pdfVersion:       pdfVers1_3,
		producer:         utf8toutf16("FPDF " + cnFpdfVersion),
		layer: layerRec{
			list:          make([]layer, 0),
			currentLayer:  -1,
			openLayerPane: false,
		},
		catalogSort:            gl.catalogSort,
		creationDate:           gl.creationDate,
		modDate:                gl.modDate,
		userUnderlineThickness: 1,
	}
	// Default draw, fill and text colors (black). text equals fill so
	// colorFlag stays at its false zero value.
	r.color.draw = r.rgbColorValue(0, 0, 0, "G", "RG")
	r.color.fill = r.rgbColorValue(0, 0, 0, "g", "rg")
	r.color.text = r.rgbColorValue(0, 0, 0, "g", "rg")
	r.acceptPageBreak = func() bool {
		return r.autoPageBreak
	}
	// Install the cp1252 translator for the core fonts so that the text
	// components render non-ASCII cp1252 text correctly regardless of which
	// constructor created the renderer. UTF-8 fonts replace it with the
	// identity via LoadUTF8FontBytes / SetTranslator.
	r.translate = r.UnicodeTranslatorFromDescriptor("")
	return r
}

// NewCustom returns a pointer to a new Renderer instance. Its methods are
// subsequently called to produce a single PDF document. NewCustom() is an
// alternative to New() that provides additional customization. The PageSize()
// example demonstrates this method.
func NewCustom(init *Init) (r *Renderer) {
	return newRenderer(init.Orientation, init.Unit, init.PageSize, init.FontDirStr, init.Size)
}

// New returns a pointer to a new Renderer instance. Its methods are subsequently
// called to produce a single PDF document.
//
// orientation specifies the default page orientation: [OrientationPortrait] or
// [OrientationLandscape]. The zero value is replaced with [OrientationPortrait].
//
// unit specifies the document measurement unit. The zero value is replaced with
// [UnitMillimeter].
//
// pageSize specifies the default page size. The zero value is replaced with [PageSizeA4].
// Fully custom dimensions can be passed through [NewCustom] via [Init.Size].
//
// fontDirStr specifies the file system location in which font resources will
// be found. An empty string is replaced with ".". This argument only needs to
// reference an actual directory if a font other than one of the core
// fonts is used. The core fonts are "courier", "helvetica" (also called
// "arial"), "times", and "zapfdingbats" (also called "symbol").
func New(orientation Orientation, unit Unit, pageSize PageSize, fontDirStr string) (r *Renderer) {
	return newRenderer(orientation, unit, pageSize, fontDirStr, Size{0, 0})
}

// Ok returns true if no processing errors have occurred.
func (r *Renderer) Ok() bool {
	return r.err == nil
}

// Err returns true if a processing error has occurred.
func (r *Renderer) Err() bool {
	return r.err != nil
}

// ClearError unsets the internal Renderer error. This method should be used with
// care, as an internal error condition usually indicates an unrecoverable
// problem with the generation of a document. It is intended to deal with cases
// in which an error is used to select an alternate form of the document.
func (r *Renderer) ClearError() {
	r.err = nil
}

// SetErrorf sets the internal Renderer error with formatted text to halt PDF
// generation; this may facilitate error handling by application. If an error
// condition is already set, this call is ignored.
//
// See the documentation for printing in the standard fmt package for details
// about fmtStr and args.
func (r *Renderer) SetErrorf(fmtStr string, args ...any) {
	if r.err == nil {
		r.err = errs.Errorf(fmtStr, args...)
	}
}

// String satisfies the fmt.Stringer interface and summarizes the Renderer
// instance.
func (r *Renderer) String() string {
	return "Renderer " + cnFpdfVersion
}

// SetError sets an error to halt PDF generation. This may facilitate error
// handling by application. See also Ok(), Err() and Error().
func (r *Renderer) SetError(err error) {
	if r.err == nil && err != nil {
		r.err = err
	}
}

// Error returns the internal Renderer error; this will be nil if no error has occurred.
func (r *Renderer) Error() error {
	return r.err
}

// GetPageSize returns the current page's width and height. This is the paper's
// size. To compute the size of the area being used, subtract the margins (see
// GetMargins()).
func (r *Renderer) GetPageSize() (width, height float64) {
	width = r.w
	height = r.h
	return width, height
}

// GetMargins returns the left, top, right, and bottom margins. The first three
// are set with the SetMargins() method. The bottom margin is set with the
// SetAutoPageBreak() method.
func (r *Renderer) GetMargins() (left, top, right, bottom float64) {
	left = r.lMargin
	top = r.tMargin
	right = r.rMargin
	bottom = r.bMargin
	return left, top, right, bottom
}

// SetMargins defines the left, top and right margins. By default, they equal 1
// cm. Call this method to change them. If the value of the right margin is
// less than zero, it is set to the same as the left margin.
func (r *Renderer) SetMargins(left, top, right float64) {
	r.lMargin = left
	r.tMargin = top
	if right < 0 {
		right = left
	}
	r.rMargin = right
}

// SetLeftMargin defines the left margin. The method can be called before
// creating the first page. If the current abscissa gets out of page, it is
// brought back to the margin.
func (r *Renderer) SetLeftMargin(margin float64) {
	r.lMargin = margin
	if r.page > 0 && r.x < margin {
		r.x = margin
	}
}

// GetCellMargin returns the cell margin. This is the amount of space before
// and after the text within a cell that's left blank, and is in units passed
// to New(). It defaults to 1mm.
func (r *Renderer) GetCellMargin() float64 {
	return r.cMargin
}

// SetCellMargin sets the cell margin. This is the amount of space before and
// after the text within a cell that's left blank, and is in units passed to
// New().
func (r *Renderer) SetCellMargin(margin float64) {
	r.cMargin = margin
}

// SetPageBoxRec sets the page box for the current page, and any following
// pages. Allowable types are trim, trimbox, crop, cropbox, bleed, bleedbox,
// art and artbox box types are case insensitive. See SetPageBox() for a method
// that specifies the coordinates and extent of the page box individually.
func (r *Renderer) SetPageBoxRec(t string, pb PageBox) {
	switch strings.ToLower(t) {
	case "trim":
		fallthrough
	case "trimbox":
		t = "TrimBox"
	case "crop":
		fallthrough
	case "cropbox":
		t = "CropBox"
	case "bleed":
		fallthrough
	case "bleedbox":
		t = "BleedBox"
	case "art":
		fallthrough
	case "artbox":
		t = "ArtBox"
	default:
		r.err = fmt.Errorf("%s is not a valid page box type", t)
		return
	}

	pb.X = pb.X * r.k
	pb.Y = pb.Y * r.k
	pb.Wd = (pb.Wd * r.k) + pb.X
	pb.Ht = (pb.Ht * r.k) + pb.Y

	if r.page > 0 {
		r.pageBoxes[r.page][t] = pb
	}

	// always override. page defaults are supplied in addPage function
	r.defPageBoxes[t] = pb
}

// SetPageBox sets the page box for the current page, and any following pages.
// Allowable types are trim, trimbox, crop, cropbox, bleed, bleedbox, art and
// artbox box types are case insensitive.
func (r *Renderer) SetPageBox(t string, x, y, wd, ht float64) {
	r.SetPageBoxRec(t, PageBox{Size{Wd: wd, Ht: ht}, Point{X: x, Y: y}})
}

// SetPage sets the current page to that of a valid page in the PDF document.
// pageNum is one-based. The SetPage() example demonstrates this method.
func (r *Renderer) SetPage(pageNum int) {
	if (pageNum > 0) && (pageNum < len(r.pages)) {
		r.page = pageNum
	}
}

// PageCount returns the number of pages currently in the document. Since page
// numbers in gofpdf are one-based, the page count is the same as the page
// number of the current last page.
func (r *Renderer) PageCount() int {
	return len(r.pages) - 1
}

// GetFontLocation returns the location in the file system of the font and font
// definition files.
func (r *Renderer) GetFontLocation() string {
	return r.fontpath
}

// SetFontLocation sets the location in the file system of the font and font
// definition files.
func (r *Renderer) SetFontLocation(fontDirStr string) {
	r.fontpath = fontDirStr
}

// GetFontLoader returns the loader used to read font files (.json and .z) from
// an arbitrary source.
func (r *Renderer) GetFontLoader() FontLoader {
	return r.fontLoader
}

// SetFontLoader sets a loader used to read font files (.json and .z) from an
// arbitrary source. If a font loader has been specified, it is used to load
// the named font resources when AddFont() is called. If this operation fails,
// an attempt is made to load the resources from the configured font directory
// (see SetFontLocation()).
func (r *Renderer) SetFontLoader(loader FontLoader) {
	r.fontLoader = loader
}

// SetHeaderFuncMode sets the function that lets the application render the
// page header. See SetHeaderFunc() for more details. The value for homeMode
// should be set to true to have the current position set to the left and top
// margin after the header function is called.
func (r *Renderer) SetHeaderFuncMode(fnc func(), homeMode bool) {
	r.headerFnc = fnc
	r.headerHomeMode = homeMode
}

// SetHeaderFunc sets the function that lets the application render the page
// header. The specified function is automatically called by AddPage() and
// should not be called directly by the application. The implementation in Renderer
// is empty, so you have to provide an appropriate function if you want page
// headers. fnc will typically be a closure that has access to the Renderer
// instance and other document generation variables.
//
// A header is a convenient place to put background content that repeats on
// each page such as a watermark. When this is done, remember to reset the X
// and Y values so the normal content begins where expected. Including a
// watermark on each page is demonstrated in the example for TransformRotate.
//
// This method is demonstrated in the example for AddPage().
func (r *Renderer) SetHeaderFunc(fnc func()) {
	r.headerFnc = fnc
}

// SetFooterFunc sets the function that lets the application render the page
// footer. The specified function is automatically called by AddPage() and
// Close() and should not be called directly by the application. The
// implementation in Renderer is empty, so you have to provide an appropriate
// function if you want page footers. fnc will typically be a closure that has
// access to the Renderer instance and other document generation variables. See
// SetFooterFuncLpi for a similar function that passes a last page indicator.
//
// This method is demonstrated in the example for AddPage().
func (r *Renderer) SetFooterFunc(fnc func()) {
	r.footerFnc = fnc
	r.footerFncLpi = nil
}

// SetFooterFuncLpi sets the function that lets the application render the page
// footer. The specified function is automatically called by AddPage() and
// Close() and should not be called directly by the application. It is passed a
// boolean that is true if the last page of the document is being rendered. The
// implementation in Renderer is empty, so you have to provide an appropriate
// function if you want page footers. fnc will typically be a closure that has
// access to the Renderer instance and other document generation variables.
func (r *Renderer) SetFooterFuncLpi(fnc func(lastPage bool)) {
	r.footerFncLpi = fnc
	r.footerFnc = nil
}

// SetTopMargin defines the top margin. The method can be called before
// creating the first page.
func (r *Renderer) SetTopMargin(margin float64) {
	r.tMargin = margin
}

// SetRightMargin defines the right margin. The method can be called before
// creating the first page.
func (r *Renderer) SetRightMargin(margin float64) {
	r.rMargin = margin
}

// GetAutoPageBreak returns true if automatic pages breaks are enabled, false
// otherwise. This is followed by the triggering limit from the bottom of the
// page. This value applies only if automatic page breaks are enabled.
func (r *Renderer) GetAutoPageBreak() (auto bool, margin float64) {
	auto = r.autoPageBreak
	margin = r.bMargin
	return auto, margin
}

// SetAutoPageBreak enables or disables the automatic page breaking mode. When
// enabling, the second parameter is the distance from the bottom of the page
// that defines the triggering limit. By default, the mode is on and the margin
// is 2 cm.
func (r *Renderer) SetAutoPageBreak(auto bool, margin float64) {
	r.autoPageBreak = auto
	r.bMargin = margin
	r.pageBreakTrigger = r.h - margin
}

// GetDisplayMode returns the current display mode. See SetDisplayMode() for details.
func (r *Renderer) GetDisplayMode() (zoomStr, layoutStr string) {
	return r.zoomMode, r.layoutMode
}

// SetDisplayMode sets advisory display directives for the document viewer.
// Pages can be displayed entirely on screen, occupy the full width of the
// window, use real size, be scaled by a specific zooming factor or use viewer
// default (configured in the Preferences menu of Adobe Reader). The page
// layout can be specified so that pages are displayed individually or in
// pairs.
//
// zoomStr can be "fullpage" to display the entire page on screen, "fullwidth"
// to use maximum width of window, "real" to use real size (equivalent to 100%
// zoom) or "default" to use viewer default mode.
//
// layoutStr can be "single" (or "SinglePage") to display one page at once,
// "continuous" (or "OneColumn") to display pages continuously, "two" (or
// "TwoColumnLeft") to display two pages on two columns with odd-numbered pages
// on the left, or "TwoColumnRight" to display two pages on two columns with
// odd-numbered pages on the right, or "TwoPageLeft" to display pages two at a
// time with odd-numbered pages on the left, or "TwoPageRight" to display pages
// two at a time with odd-numbered pages on the right, or "default" to use
// viewer default mode.
func (r *Renderer) SetDisplayMode(zoomStr, layoutStr string) {
	if r.err != nil {
		return
	}
	layoutStr = cmp.Or(layoutStr, "default")
	switch zoomStr {
	case "fullpage", "fullwidth", "real", "default":
		r.zoomMode = zoomStr
	default:
		r.err = fmt.Errorf("incorrect zoom display mode: %s", zoomStr)
		return
	}
	switch layoutStr {
	case "single", "continuous", "two", "default", "SinglePage", "OneColumn",
		"TwoColumnLeft", "TwoColumnRight", "TwoPageLeft", "TwoPageRight":
		r.layoutMode = layoutStr
	default:
		r.err = fmt.Errorf("incorrect layout display mode: %s", layoutStr)
		return
	}
}

// SetDefaultCompression controls the default setting of the internal
// compression flag. See SetCompression() for more details. Compression is on
// by default.
func SetDefaultCompression(compress bool) {
	gl.noCompress = !compress
}

// GetCompression returns whether page compression is enabled.
func (r *Renderer) GetCompression() bool {
	return r.compress
}

// SetCompression activates or deactivates page compression with zlib. When
// activated, the internal representation of each page is compressed, which
// leads to a compression ratio of about 2 for the resulting document.
// Compression is on by default.
func (r *Renderer) SetCompression(compress bool) {
	r.compress = compress
}

// GetProducer returns the producer of the document as ISO-8859-1 or UTF-16BE.
func (r *Renderer) GetProducer() string {
	return r.producer
}

// SetProducer defines the producer of the document. isUTF8 indicates if the string
// is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (r *Renderer) SetProducer(producerStr string, isUTF8 bool) {
	if isUTF8 {
		producerStr = utf8toutf16(producerStr)
	}
	r.producer = producerStr
}

// GetTitle returns the title of the document as ISO-8859-1 or UTF-16BE.
func (r *Renderer) GetTitle() string {
	return r.title
}

// SetTitle defines the title of the document. isUTF8 indicates if the string
// is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (r *Renderer) SetTitle(titleStr string, isUTF8 bool) {
	if isUTF8 {
		titleStr = utf8toutf16(titleStr)
	}
	r.title = titleStr
}

// GetSubject returns the subject of the document as ISO-8859-1 or UTF-16BE.
func (r *Renderer) GetSubject() string {
	return r.subject
}

// SetSubject defines the subject of the document. isUTF8 indicates if the
// string is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (r *Renderer) SetSubject(subjectStr string, isUTF8 bool) {
	if isUTF8 {
		subjectStr = utf8toutf16(subjectStr)
	}
	r.subject = subjectStr
}

// GetAuthor returns the author of the document as ISO-8859-1 or UTF-16BE.
func (r *Renderer) GetAuthor() string {
	return r.author
}

// SetAuthor defines the author of the document. isUTF8 indicates if the string
// is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (r *Renderer) SetAuthor(authorStr string, isUTF8 bool) {
	if isUTF8 {
		authorStr = utf8toutf16(authorStr)
	}
	r.author = authorStr
}

// GetLang returns the natural language of the document (e.g. "de-CH").
func (r *Renderer) GetLang() string {
	return r.lang
}

// SetLang defines the natural language of the document (e.g. "de-CH").
func (r *Renderer) SetLang(lang string) {
	r.lang = lang
}

// GetKeywords returns the keywords of the document as ISO-8859-1 or UTF-16BE.
func (r *Renderer) GetKeywords() string {
	return r.keywords
}

// SetKeywords defines the keywords of the document. keywordStr is a
// space-delimited string, for example "invoice August". isUTF8 indicates if
// the string is encoded
func (r *Renderer) SetKeywords(keywordsStr string, isUTF8 bool) {
	if isUTF8 {
		keywordsStr = utf8toutf16(keywordsStr)
	}
	r.keywords = keywordsStr
}

// GetCreator returns the creator of the document as ISO-8859-1 or UTF-16BE.
func (r *Renderer) GetCreator() string {
	return r.creator
}

// SetCreator defines the creator of the document. isUTF8 indicates if the
// string is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (r *Renderer) SetCreator(creatorStr string, isUTF8 bool) {
	if isUTF8 {
		creatorStr = utf8toutf16(creatorStr)
	}
	r.creator = creatorStr
}

// GetXmpMetadata returns the XMP metadata that will be embedded with the document.
func (r *Renderer) GetXmpMetadata() []byte {
	return []byte(string(r.xmp))
}

// SetXmpMetadata defines XMP metadata that will be embedded with the document.
func (r *Renderer) SetXmpMetadata(xmpStream []byte) {
	r.xmp = xmpStream
}

// AddOutputIntent adds an output intent with ICC color profile
func (r *Renderer) AddOutputIntent(outputIntent OutputIntent) {
	r.outputIntents = append(r.outputIntents, outputIntent)
	if r.pdfVersion < pdfVers1_4 {
		r.pdfVersion = pdfVers1_4
	}
}

// AliasNbPages defines an alias for the total number of pages. It will be
// substituted as the document is closed. An empty string is replaced with the
// string "{nb}".
//
// See the example for AddPage() for a demonstration of this method.
func (r *Renderer) AliasNbPages(aliasStr string) {
	aliasStr = cmp.Or(aliasStr, "{nb}")
	r.aliasNbPagesStr = aliasStr
}

// RTL enables right-to-left mode
func (r *Renderer) RTL() {
	r.isRTL = true
}

// LTR disables right-to-left mode
func (r *Renderer) LTR() {
	r.isRTL = false
}

// open begins a document
func (r *Renderer) open() {
	r.state = 1
}

// Close terminates the PDF document. It is not necessary to call this method
// explicitly because Output(), OutputAndClose() and OutputFileAndClose() do it
// automatically. If the document contains no page, AddPage() is called to
// prevent the generation of an invalid document.
func (r *Renderer) Close() {
	if r.err == nil {
		switch {
		case r.clipNest > 0:
			r.err = errs.New("clip procedure must be explicitly ended")
		case r.transformNest > 0:
			r.err = errs.New("transformation procedure must be explicitly ended")
		}
	}
	if r.err != nil {
		return
	}
	if r.state == 3 {
		return
	}
	if r.page == 0 {
		r.AddPage()
		if r.err != nil {
			return
		}
	}
	// Page footer
	r.inFooter = true
	switch {
	case r.footerFnc != nil:
		r.footerFnc()
	case r.footerFncLpi != nil:
		r.footerFncLpi(true)
	}
	r.inFooter = false

	// Close page
	r.endpage()
	// Close document
	r.enddoc()
}

// PageSize returns the width and height of the specified page in the units
// established in New(). These return values are followed by the unit of
// measure itself. If pageNum is zero or otherwise out of bounds, it returns
// the default page size, that is, the size of the page that would be added by
// AddPage().
func (r *Renderer) PageSize(pageNum int) (wd, ht float64, unit Unit) {
	sz, ok := r.pageSizes[pageNum]
	if ok {
		sz.Wd, sz.Ht = sz.Wd/r.k, sz.Ht/r.k
	} else {
		sz = r.defPageSize // user units
	}
	return sz.Wd, sz.Ht, r.unit
}

// AddPageFormat adds a new page with non-default orientation or size. See
// AddPage() for more details.
//
// See New() for a description of orientation.
//
// size specifies the size of the new page in the units established in New().
//
// The PageSize() example demonstrates this method.
func (r *Renderer) AddPageFormat(orientation Orientation, size Size) {
	if r.err != nil {
		return
	}
	if r.page != len(r.pages)-1 {
		r.page = len(r.pages) - 1
	}
	if r.state == 0 {
		r.open()
	}
	familyStr := r.fontFamily
	style := r.fontStyle
	if r.underline {
		style += "U"
	}
	if r.strikeout {
		style += "S"
	}
	fontsize := r.fontSizePt
	lw := r.lineWidth
	dc := r.color.draw
	fc := r.color.fill
	tc := r.color.text
	cf := r.colorFlag

	if r.page > 0 {
		r.inFooter = true
		// Page footer avoid double call on footer.
		switch {
		case r.footerFnc != nil:
			r.footerFnc()
		case r.footerFncLpi != nil:
			r.footerFncLpi(false) // not last page.
		}
		r.inFooter = false
		// Close page
		r.endpage()
	}
	// Start new page
	r.beginpage(orientation, size)
	// 	Set line cap style to current value
	// r.out("2 J")
	r.outf("%d J", r.capStyle)
	// 	Set line join style to current value
	r.outf("%d j", r.joinStyle)
	// Set line width
	r.lineWidth = lw
	r.outf("%.2f w", lw*r.k)
	// Set dash pattern
	if len(r.dashArray) > 0 {
		r.outputDashPattern()
	}
	// 	Set font
	if familyStr != "" {
		r.SetFont(familyStr, style, fontsize)
		if r.err != nil {
			return
		}
	}
	// 	Set colors
	r.color.draw = dc
	if dc.str != "0 G" {
		r.out(dc.str)
	}
	r.color.fill = fc
	if fc.str != "0 g" {
		r.out(fc.str)
	}
	r.color.text = tc
	r.colorFlag = cf
	// 	Page header
	if r.headerFnc != nil {
		r.inHeader = true
		r.headerFnc()
		r.inHeader = false
		if r.headerHomeMode {
			r.SetHomeXY()
		}
	}
	// 	Restore line width
	if r.lineWidth != lw {
		r.lineWidth = lw
		r.outf("%.2f w", lw*r.k)
	}
	// Restore font
	if familyStr != "" {
		r.SetFont(familyStr, style, fontsize)
		if r.err != nil {
			return
		}
	}
	// Restore colors
	if r.color.draw.str != dc.str {
		r.color.draw = dc
		r.out(dc.str)
	}
	if r.color.fill.str != fc.str {
		r.color.fill = fc
		r.out(fc.str)
	}
	r.color.text = tc
	r.colorFlag = cf
}

// AddPage adds a new page to the document. If a page is already present, the
// Footer() method is called first to output the footer. Then the page is
// added, the current position set to the top-left corner according to the left
// and top margins, and Header() is called to display the header.
//
// The font which was set before calling is automatically restored. There is no
// need to call SetFont() again if you want to continue with the same font. The
// same is true for colors and line width.
//
// The origin of the coordinate system is at the top-left corner and increasing
// ordinates go downwards.
//
// See AddPageFormat() for a version of this method that allows the page size
// and orientation to be different than the default.
func (r *Renderer) AddPage() {
	if r.err != nil {
		return
	}
	// dbg("AddPage")
	r.AddPageFormat(r.defOrientation, r.defPageSize)
}

// PageNo returns the current page number.
//
// See the example for AddPage() for a demonstration of this method.
func (r *Renderer) PageNo() int {
	return r.page
}

func colorComp(v int) (int, float64) {
	v = min(max(v, 0), 255)
	return v, float64(v) / 255.0
}

func (r *Renderer) rgbColorValue(red, green, blue int, grayStr, fullStr string) (clr colorState) {
	clr.ir, clr.r = colorComp(red)
	clr.ig, clr.g = colorComp(green)
	clr.ib, clr.b = colorComp(blue)
	clr.mode = colorModeRGB
	clr.gray = clr.ir == clr.ig && clr.r == clr.b
	if len(grayStr) > 0 {
		if clr.gray {
			clr.str = fmt.Sprintf("%.3f %s", clr.r, grayStr)
		} else {
			clr.str = fmt.Sprintf("%.3f %.3f %.3f %s", clr.r, clr.g, clr.b, fullStr)
		}
	} else {
		clr.str = fmt.Sprintf("%.3f %.3f %.3f", clr.r, clr.g, clr.b)
	}
	return clr
}

// SetDrawColor defines the color used for all drawing operations (lines,
// rectangles and cell borders). It is expressed in RGB components (0 - 255).
// The method can be called before the first page is created. The value is
// retained from page to page.
func (r *Renderer) SetDrawColor(red, green, blue int) {
	r.setDrawColor(red, green, blue)
}

func (r *Renderer) setDrawColor(red, green, blue int) {
	r.color.draw = r.rgbColorValue(red, green, blue, "G", "RG")
	if r.page > 0 {
		r.out(r.color.draw.str)
	}
}

// GetDrawColor returns the most recently set draw color as RGB components (0 -
// 255). This will not be the current value if a draw color of some other type
// (for example, spot) has been more recently set.
func (r *Renderer) GetDrawColor() (int, int, int) {
	return r.color.draw.ir, r.color.draw.ig, r.color.draw.ib
}

// SetFillColor defines the color used for all filling operations (filled
// rectangles and cell backgrounds). It is expressed in RGB components (0
// -255). The method can be called before the first page is created and the
// value is retained from page to page.
func (r *Renderer) SetFillColor(red, green, blue int) {
	r.setFillColor(red, green, blue)
}

func (r *Renderer) setFillColor(red, green, blue int) {
	r.color.fill = r.rgbColorValue(red, green, blue, "g", "rg")
	r.colorFlag = r.color.fill.str != r.color.text.str
	if r.page > 0 {
		r.out(r.color.fill.str)
	}
}

// GetFillColor returns the most recently set fill color as RGB components (0 -
// 255). This will not be the current value if a fill color of some other type
// (for example, spot) has been more recently set.
func (r *Renderer) GetFillColor() (int, int, int) {
	return r.color.fill.ir, r.color.fill.ig, r.color.fill.ib
}

// SetTextColor defines the color used for text. It is expressed in RGB
// components (0 - 255). The method can be called before the first page is
// created. The value is retained from page to page.
func (r *Renderer) SetTextColor(red, green, blue int) {
	r.setTextColor(red, green, blue)
}

func (r *Renderer) setTextColor(red, green, blue int) {
	r.color.text = r.rgbColorValue(red, green, blue, "g", "rg")
	r.colorFlag = r.color.fill.str != r.color.text.str
}

// GetTextColor returns the most recently set text color as RGB components (0 -
// 255). This will not be the current value if a text color of some other type
// (for example, spot) has been more recently set.
func (r *Renderer) GetTextColor() (int, int, int) {
	return r.color.text.ir, r.color.text.ig, r.color.text.ib
}

// GetStringWidth returns the length of a string in user units. A font must be
// currently selected.
func (r *Renderer) GetStringWidth(s string) float64 {
	if r.err != nil {
		return 0
	}
	w := r.GetStringSymbolWidth(s)
	return float64(w) * r.fontSize / 1000
}

// GetStringSymbolWidth returns the length of a string in glyf units. A font must be
// currently selected.
func (r *Renderer) GetStringSymbolWidth(s string) int {
	if r.err != nil {
		return 0
	}
	w := 0
	if r.isCurrentUTF8 {
		for _, char := range s {
			intChar := int(char)
			switch {
			case len(r.currentFont.Cw) > intChar && r.currentFont.Cw[intChar] > 0:
				if r.currentFont.Cw[intChar] != 65535 {
					w += r.currentFont.Cw[intChar]
				}
			case r.currentFont.Desc.MissingWidth != 0:
				w += r.currentFont.Desc.MissingWidth
			default:
				w += 500
			}
		}
	} else {
		for _, ch := range []byte(s) {
			if ch == 0 {
				break
			}
			w += r.currentFont.Cw[ch]
		}
	}
	return w
}

// SetLineWidth defines the line width. By default, the value equals 0.2 mm.
// The method can be called before the first page is created. The value is
// retained from page to page.
func (r *Renderer) SetLineWidth(width float64) {
	r.setLineWidth(width)
}

func (r *Renderer) setLineWidth(width float64) {
	r.lineWidth = width
	if r.page > 0 {
		r.out(fmtF64(width*r.k, 2) + " w")
	}
}

// GetLineWidth returns the current line thickness.
func (r *Renderer) GetLineWidth() float64 {
	return r.lineWidth
}

// GetLineCapStyle returns the current line cap style.
func (r *Renderer) GetLineCapStyle() string {
	switch r.capStyle {
	case 1:
		return "round"
	case 2:
		return "square"
	default:
		return "butt"
	}
}

// SetLineCapStyle defines the line cap style. styleStr should be "butt",
// "round" or "square". A square style projects from the end of the line. The
// method can be called before the first page is created. The value is
// retained from page to page.
func (r *Renderer) SetLineCapStyle(styleStr string) {
	var capStyle int
	switch styleStr {
	case "round":
		capStyle = 1
	case "square":
		capStyle = 2
	default:
		capStyle = 0
	}
	r.capStyle = capStyle
	if r.page > 0 {
		r.outf("%d J", r.capStyle)
	}
}

// GetLineJoinStyle returns the current line join style.
func (r *Renderer) GetLineJoinStyle() string {
	switch r.joinStyle {
	case 1:
		return "round"
	case 2:
		return "bevel"
	default:
		return "miter"
	}
}

// SetLineJoinStyle defines the line cap style. styleStr should be "miter",
// "round" or "bevel". The method can be called before the first page
// is created. The value is retained from page to page.
func (r *Renderer) SetLineJoinStyle(styleStr string) {
	var joinStyle int
	switch styleStr {
	case "round":
		joinStyle = 1
	case "bevel":
		joinStyle = 2
	default:
		joinStyle = 0
	}
	r.joinStyle = joinStyle
	if r.page > 0 {
		r.outf("%d j", r.joinStyle)
	}
}

// SetDashPattern sets the dash pattern that is used to draw lines. The
// dashArray elements are numbers that specify the lengths, in units
// established in New(), of alternating dashes and gaps. The dash phase
// specifies the distance into the dash pattern at which to start the dash. The
// dash pattern is retained from page to page. Call this method with an empty
// array to restore solid line drawing.
//
// The Beziergon() example demonstrates this method.
func (r *Renderer) SetDashPattern(dashArray []float64, dashPhase float64) {
	scaled := make([]float64, len(dashArray))
	for i, value := range dashArray {
		scaled[i] = value * r.k
	}
	dashPhase *= r.k

	r.dashArray = scaled
	r.dashPhase = dashPhase
	if r.page > 0 {
		r.outputDashPattern()
	}

}

func (r *Renderer) outputDashPattern() {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, value := range r.dashArray {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(strconv.FormatFloat(value, 'f', 2, 64))
	}
	buf.WriteString("] ")
	buf.WriteString(strconv.FormatFloat(r.dashPhase, 'f', 2, 64))
	buf.WriteString(" d")
	r.outbuf(&buf)
}

// Line draws a line between points (x1, y1) and (x2, y2) using the current
// draw color, line width and cap style.
func (r *Renderer) Line(x1, y1, x2, y2 float64) {
	// r.outf("%.2f %.2f m %.2f %.2f l S", x1*r.k, (r.h-y1)*r.k, x2*r.k, (r.h-y2)*r.k)
	const prec = 2
	r.putF64(x1*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y1)*r.k, prec)
	r.put(" m ")
	r.putF64(x2*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y2)*r.k, prec)
	r.put(" l S\n")
}

// fillDrawOp corrects path painting operators
func fillDrawOp(styleStr string) (opStr string) {
	switch strings.ToUpper(styleStr) {
	case "", "D":
		// Stroke the path.
		opStr = "S"
	case "F":
		// fill the path, using the nonzero winding number rule
		opStr = "f"
	case "F*":
		// fill the path, using the even-odd rule
		opStr = "f*"
	case "FD", "DF":
		// fill and then stroke the path, using the nonzero winding number rule
		opStr = "B"
	case "FD*", "DF*":
		// fill and then stroke the path, using the even-odd rule
		opStr = "B*"
	default:
		opStr = styleStr
	}
	return opStr
}

// Rect outputs a rectangle of width w and height h with the upper left corner
// positioned at point (x, y).
//
// It can be drawn (border only), filled (with no border) or both. styleStr can
// be "F" for filled, "D" for outlined only, or "DF" or "FD" for outlined and
// filled. An empty string will be replaced with "D". Drawing uses the current
// draw color and line width centered on the rectangle's perimeter. Filling
// uses the current fill color.
func (r *Renderer) Rect(x, y, w, h float64, styleStr string) {
	// r.outf("%.2f %.2f %.2f %.2f re %s", x*r.k, (r.h-y)*r.k, w*r.k, -h*r.k, fillDrawOp(styleStr))
	const prec = 2
	r.putF64(x*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y)*r.k, prec)
	r.put(" ")
	r.putF64(w*r.k, prec)
	r.put(" ")
	r.putF64(-h*r.k, prec)
	r.put(" re " + fillDrawOp(styleStr) + "\n")
}

// RoundedRect outputs a rectangle of width w and height h with the upper left
// corner positioned at point (x, y). It can be drawn (border only), filled
// (with no border) or both. styleStr can be "F" for filled, "D" for outlined
// only, or "DF" or "FD" for outlined and filled. An empty string will be
// replaced with "D". Drawing uses the current draw color and line width
// centered on the rectangle's perimeter. Filling uses the current fill color.
// The rounded corners of the rectangle are specified by radius r. corners is a
// string that includes "1" to round the upper left corner, "2" to round the
// upper right corner, "3" to round the lower right corner, and "4" to round
// the lower left corner. The RoundedRect example demonstrates this method.
func (r *Renderer) RoundedRect(x, y, w, h, radius float64, corners string, stylestr string) {
	// This routine was adapted by Brigham Thompson from a script by Christophe Prugnaud
	var rTL, rTR, rBR, rBL float64 // zero means no rounded corner
	if strings.Contains(corners, "1") {
		rTL = radius
	}
	if strings.Contains(corners, "2") {
		rTR = radius
	}
	if strings.Contains(corners, "3") {
		rBR = radius
	}
	if strings.Contains(corners, "4") {
		rBL = radius
	}
	r.RoundedRectExt(x, y, w, h, rTL, rTR, rBR, rBL, stylestr)
}

// RoundedRectExt behaves the same as RoundedRect() but supports a different
// radius for each corner. A zero radius means squared corner. See
// RoundedRect() for more details. This method is demonstrated in the
// RoundedRect() example.
func (r *Renderer) RoundedRectExt(x, y, w, h, rTL, rTR, rBR, rBL float64, stylestr string) {
	r.roundedRectPath(x, y, w, h, rTL, rTR, rBR, rBL)
	r.out(fillDrawOp(stylestr))
	r.out("Q")
}

// Circle draws a circle centered on point (x, y) with radius r.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the circle's perimeter.
// Filling uses the current fill color.
func (r *Renderer) Circle(x, y, radius float64, styleStr string) {
	r.Ellipse(x, y, radius, radius, 0, styleStr)
}

// Ellipse draws an ellipse centered at point (x, y). rx and ry specify its
// horizontal and vertical radii.
//
// degRotate specifies the counter-clockwise angle in degrees that the ellipse
// will be rotated.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the ellipse's perimeter.
// Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (r *Renderer) Ellipse(x, y, rx, ry, degRotate float64, styleStr string) {
	r.arc(x, y, rx, ry, degRotate, 0, 360, styleStr, false)
}

// Polygon draws a closed figure defined by a series of vertices specified by
// points. The x and y fields of the points use the units established in New().
// The last point in the slice will be implicitly joined to the first to close
// the polygon.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the ellipse's perimeter.
// Filling uses the current fill color.
func (r *Renderer) Polygon(points []Point, styleStr string) {
	if len(points) > 2 {
		const prec = 5
		for j, pt := range points {
			if j == 0 {
				r.point(pt.X, pt.Y)
			} else {
				// r.outf("%.5f %.5f l ", pt.X*r.k, (r.h-pt.Y)*r.k)
				r.putF64(pt.X*r.k, prec)
				r.put(" ")
				r.putF64((r.h-pt.Y)*r.k, prec)
				r.put(" l \n")
			}
		}
		// r.outf("%.5f %.5f l ", points[0].X*r.k, (r.h-points[0].Y)*r.k)
		r.putF64(points[0].X*r.k, prec)
		r.put(" ")
		r.putF64((r.h-points[0].Y)*r.k, prec)
		r.put(" l \n")
		r.DrawPath(styleStr)
	}
}

// Beziergon draws a closed figure defined by a series of cubic Bézier curve
// segments. The first point in the slice defines the starting point of the
// figure. Each three following points p1, p2, p3 represent a curve segment to
// the point p3 using p1 and p2 as the Bézier control points.
//
// The x and y fields of the points use the units established in New().
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the ellipse's perimeter.
// Filling uses the current fill color.
func (r *Renderer) Beziergon(points []Point, styleStr string) {

	// Thanks, Robert Lillack, for contributing this function.

	if len(points) < 4 {
		return
	}
	r.point(points[0].XY())

	points = points[1:]
	for len(points) >= 3 {
		cx0, cy0 := points[0].XY()
		cx1, cy1 := points[1].XY()
		x1, y1 := points[2].XY()
		r.curve(cx0, cy0, cx1, cy1, x1, y1)
		points = points[3:]
	}

	r.DrawPath(styleStr)
}

// point outputs current point
func (r *Renderer) point(x, y float64) {
	// r.outf("%.2f %.2f m", x*r.k, (r.h-y)*r.k)
	r.putF64(x*r.k, 2)
	r.put(" ")
	r.putF64((r.h-y)*r.k, 2)
	r.put(" m\n")
}

// curve outputs a single cubic Bézier curve segment from current point
func (r *Renderer) curve(cx0, cy0, cx1, cy1, x, y float64) {
	// Thanks, Robert Lillack, for straightening this out
	// r.outf("%.5f %.5f %.5f %.5f %.5f %.5f c", cx0*r.k, (r.h-cy0)*r.k, cx1*r.k,
	// 	(r.h-cy1)*r.k, x*r.k, (r.h-y)*r.k)
	const prec = 5
	r.putF64(cx0*r.k, prec)
	r.put(" ")
	r.putF64((r.h-cy0)*r.k, prec)
	r.put(" ")
	r.putF64(cx1*r.k, prec)
	r.put(" ")
	r.putF64((r.h-cy1)*r.k, prec)
	r.put(" ")
	r.putF64(x*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y)*r.k, prec)
	r.put(" c\n")
}

// Curve draws a single-segment quadratic Bézier curve. The curve starts at
// the point (x0, y0) and ends at the point (x1, y1). The control point (cx,
// cy) specifies the curvature. At the start point, the curve is tangent to the
// straight line between the start point and the control point. At the end
// point, the curve is tangent to the straight line between the end point and
// the control point.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the curve's
// path. Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (r *Renderer) Curve(x0, y0, cx, cy, x1, y1 float64, styleStr string) {
	r.point(x0, y0)
	// r.outf("%.5f %.5f %.5f %.5f v %s", cx*r.k, (r.h-cy)*r.k, x1*r.k, (r.h-y1)*r.k,
	// 	fillDrawOp(styleStr))
	const prec = 5
	r.putF64(cx*r.k, prec)
	r.put(" ")
	r.putF64((r.h-cy)*r.k, prec)
	r.put(" ")
	r.putF64(x1*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y1)*r.k, prec)
	r.put(" v " + fillDrawOp(styleStr) + "\n")
}

// CurveCubic draws a single-segment cubic Bézier curve. This routine performs
// the same function as CurveBezierCubic() but has a nonstandard argument order.
// It is retained to preserve backward compatibility.
func (r *Renderer) CurveCubic(x0, y0, cx0, cy0, x1, y1, cx1, cy1 float64, styleStr string) {
	// r.point(x0, y0)
	// r.outf("%.5f %.5f %.5f %.5f %.5f %.5f c %s", cx0*r.k, (r.h-cy0)*r.k,
	// cx1*r.k, (r.h-cy1)*r.k, x1*r.k, (r.h-y1)*r.k, fillDrawOp(styleStr))
	r.CurveBezierCubic(x0, y0, cx0, cy0, cx1, cy1, x1, y1, styleStr)
}

// CurveBezierCubic draws a single-segment cubic Bézier curve. The curve starts at
// the point (x0, y0) and ends at the point (x1, y1). The control points (cx0,
// cy0) and (cx1, cy1) specify the curvature. At the start point, the curve is
// tangent to the straight line between the start point and the control point
// (cx0, cy0). At the end point, the curve is tangent to the straight line
// between the end point and the control point (cx1, cy1).
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the curve's
// path. Filling uses the current fill color.
//
// This routine performs the same function as CurveCubic() but uses standard
// argument order.
//
// The Circle() example demonstrates this method.
func (r *Renderer) CurveBezierCubic(x0, y0, cx0, cy0, cx1, cy1, x1, y1 float64, styleStr string) {
	r.point(x0, y0)
	//	r.outf("%.5f %.5f %.5f %.5f %.5f %.5f c %s", cx0*r.k, (r.h-cy0)*r.k,
	//		cx1*r.k, (r.h-cy1)*r.k, x1*r.k, (r.h-y1)*r.k, fillDrawOp(styleStr))
	const prec = 5
	r.putF64(cx0*r.k, prec)
	r.put(" ")
	r.putF64((r.h-cy0)*r.k, prec)
	r.put(" ")
	r.putF64(cx1*r.k, prec)
	r.put(" ")
	r.putF64((r.h-cy1)*r.k, prec)
	r.put(" ")
	r.putF64(x1*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y1)*r.k, prec)
	r.put(" c " + fillDrawOp(styleStr) + "\n")
}

// Arc draws an elliptical arc centered at point (x, y). rx and ry specify its
// horizontal and vertical radii.
//
// degRotate specifies the angle that the arc will be rotated. degStart and
// degEnd specify the starting and ending angle of the arc. All angles are
// specified in degrees and measured counter-clockwise from the 3 o'clock
// position.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the arc's
// path. Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (r *Renderer) Arc(x, y, rx, ry, degRotate, degStart, degEnd float64, styleStr string) {
	r.arc(x, y, rx, ry, degRotate, degStart, degEnd, styleStr, false)
}

// GetAlpha returns the alpha blending channel, which consists of the
// alpha transparency value and the blend mode. See SetAlpha for more
// details.
func (r *Renderer) GetAlpha() (alpha float64, mode BlendMode) {
	return r.alpha, r.blendMode
}

// SetAlpha sets the alpha blending channel. The blending effect applies to
// text, drawings and images.
//
// alpha must be a value between 0.0 (fully transparent) to 1.0 (fully opaque).
// Values outside of this range result in an error.
//
// mode must be a valid [BlendMode]. The zero value is replaced with
// [BlendModeNormal].
//
// To reset normal rendering after applying a blending mode, call this method
// with alpha set to 1.0 and mode set to [BlendModeNormal].
func (r *Renderer) SetAlpha(alpha float64, mode BlendMode) {
	if r.err != nil {
		return
	}
	mode = cmp.Or(mode, BlendModeNormal)
	if !mode.Valid() {
		r.err = fmt.Errorf("unrecognized blend mode %q", mode)
		return
	}
	if alpha < 0.0 || alpha > 1.0 {
		r.err = fmt.Errorf("alpha value (0.0 - 1.0) is out of range: %.3f", alpha)
		return
	}
	r.alpha = alpha
	r.blendMode = mode
	modeStr := mode.String()
	alphaStr := fmt.Sprintf("%.3f", alpha)
	keyStr := fmt.Sprintf("%s %s", alphaStr, modeStr)
	pos, ok := r.blendMap[keyStr]
	if !ok {
		pos = len(r.blendList) // at least 1
		r.blendList = append(r.blendList, blendMode{alphaStr, alphaStr, modeStr, 0})
		r.blendMap[keyStr] = pos
	}
	if len(r.blendMap) > 0 && r.pdfVersion < pdfVers1_4 {
		r.pdfVersion = pdfVers1_4
	}
	r.outf("/GS%d gs", pos)
}

func (r *Renderer) gradientClipStart(x, y, w, h float64) {
	{
		const prec = 2
		// Save current graphic state and set clipping area
		// r.outf("q %.2f %.2f %.2f %.2f re W n", x*r.k, (r.h-y)*r.k, w*r.k, -h*r.k)
		r.put("q ")
		r.putF64(x*r.k, prec)
		r.put(" ")
		r.putF64((r.h-y)*r.k, prec)
		r.put(" ")
		r.putF64(w*r.k, prec)
		r.put(" ")
		r.putF64(-h*r.k, prec)
		r.put(" re W n\n")
	}
	{
		const prec = 5
		// Set up transformation matrix for gradient
		// r.outf("%.5f 0 0 %.5f %.5f %.5f cm", w*r.k, h*r.k, x*r.k, (r.h-(y+h))*r.k)
		r.putF64(w*r.k, prec)
		r.put(" 0 0 ")
		r.putF64(h*r.k, prec)
		r.put(" ")
		r.putF64(x*r.k, prec)
		r.put(" ")
		r.putF64((r.h-(y+h))*r.k, prec)
		r.put(" cm\n")
	}
}

func (r *Renderer) gradientClipEnd() {
	// Restore previous graphic state
	r.out("Q")
}

func (r *Renderer) gradient(tp, r1, g1, b1, r2, g2, b2 int, x1, y1, x2, y2, radius float64) {
	pos := len(r.gradientList)
	clr1 := r.rgbColorValue(r1, g1, b1, "", "")
	clr2 := r.rgbColorValue(r2, g2, b2, "", "")
	r.gradientList = append(r.gradientList, gradient{tp, clr1.str, clr2.str,
		x1, y1, x2, y2, radius, 0})
	r.outf("/Sh%d sh", pos)
}

// LinearGradient draws a rectangular area with a blending of one color to
// another. The rectangle is of width w and height h. Its upper left corner is
// positioned at point (x, y).
//
// Each color is specified with three component values, one each for red, green
// and blue. The values range from 0 to 255. The first color is specified by
// (r1, g1, b1) and the second color by (r2, g2, b2).
//
// The blending is controlled with a gradient vector that uses normalized
// coordinates in which the lower left corner is position (0, 0) and the upper
// right corner is (1, 1). The vector's origin and destination are specified by
// the points (x1, y1) and (x2, y2). In a linear gradient, blending occurs
// perpendicularly to the vector. The vector does not necessarily need to be
// anchored on the rectangle edge. Color 1 is used up to the origin of the
// vector and color 2 is used beyond the vector's end point. Between the points
// the colors are gradually blended.
func (r *Renderer) LinearGradient(x, y, w, h float64, r1, g1, b1, r2, g2, b2 int, x1, y1, x2, y2 float64) {
	r.gradientClipStart(x, y, w, h)
	r.gradient(2, r1, g1, b1, r2, g2, b2, x1, y1, x2, y2, 0)
	r.gradientClipEnd()
}

// RadialGradient draws a rectangular area with a blending of one color to
// another. The rectangle is of width w and height h. Its upper left corner is
// positioned at point (x, y).
//
// Each color is specified with three component values, one each for red, green
// and blue. The values range from 0 to 255. The first color is specified by
// (r1, g1, b1) and the second color by (r2, g2, b2).
//
// The blending is controlled with a point and a circle, both specified with
// normalized coordinates in which the lower left corner of the rendered
// rectangle is position (0, 0) and the upper right corner is (1, 1). Color 1
// begins at the origin point specified by (x1, y1). Color 2 begins at the
// circle specified by the center point (x2, y2) and radius r. Colors are
// gradually blended from the origin to the circle. The origin and the circle's
// center do not necessarily have to coincide, but the origin must be within
// the circle to avoid rendering problems.
//
// The LinearGradient() example demonstrates this method.
func (r *Renderer) RadialGradient(x, y, w, h float64, r1, g1, b1, r2, g2, b2 int, x1, y1, x2, y2, radius float64) {
	r.gradientClipStart(x, y, w, h)
	r.gradient(3, r1, g1, b1, r2, g2, b2, x1, y1, x2, y2, radius)
	r.gradientClipEnd()
}

// ClipRect begins a rectangular clipping operation. The rectangle is of width
// w and height h. Its upper left corner is positioned at point (x, y). outline
// is true to draw a border with the current draw color and line width centered
// on the rectangle's perimeter. Only the outer half of the border will be
// shown. After calling this method, all rendering operations (for example,
// Image(), LinearGradient(), etc) will be clipped by the specified rectangle.
// Call ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (r *Renderer) ClipRect(x, y, w, h float64, outline bool) {
	r.clipNest++
	// r.outf("q %.2f %.2f %.2f %.2f re W %s", x*r.k, (r.h-y)*r.k, w*r.k, -h*r.k, strIf(outline, "S", "n"))
	const prec = 2
	r.put("q ")
	r.putF64(x*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y)*r.k, prec)
	r.put(" ")
	r.putF64(w*r.k, prec)
	r.put(" ")
	r.putF64(-h*r.k, prec)
	op := "n"
	if outline {
		op = "S"
	}
	r.put(" re W " + op + "\n")
}

// ClipText begins a clipping operation in which rendering is confined to the
// character string specified by txtStr. The origin (x, y) is on the left of
// the first character at the baseline. The current font is used. outline is
// true to draw a border with the current draw color and line width centered on
// the perimeters of the text characters. Only the outer half of the border
// will be shown. After calling this method, all rendering operations (for
// example, Image(), LinearGradient(), etc) will be clipped. Call ClipEnd() to
// restore unclipped operations.
func (r *Renderer) ClipText(x, y float64, txtStr string, outline bool) {
	r.clipNest++
	// r.outf("q BT %.5f %.5f Td %d Tr (%s) Tj ET", x*r.k, (r.h-y)*r.k, intIf(outline, 5, 7), r.escape(txtStr))
	const prec = 5
	r.put("q BT ")
	r.putF64(x*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y)*r.k, prec)
	r.put(" Td ")
	renderMode := 7 // add to clipping path
	if outline {
		renderMode = 5 // stroke and add to clipping path
	}
	r.putInt(renderMode)
	r.put(" Tr (")
	r.put(r.escape(txtStr))
	r.put(") Tj ET\n")
}

func (r *Renderer) clipArc(x1, y1, x2, y2, x3, y3 float64) {
	h := r.h
	// r.outf("%.5f %.5f %.5f %.5f %.5f %.5f c ", x1*r.k, (h-y1)*r.k,
	// 	x2*r.k, (h-y2)*r.k, x3*r.k, (h-y3)*r.k)
	const prec = 5
	r.putF64(x1*r.k, prec)
	r.put(" ")
	r.putF64((h-y1)*r.k, prec)
	r.put(" ")
	r.putF64(x2*r.k, prec)
	r.put(" ")
	r.putF64((h-y2)*r.k, prec)
	r.put(" ")
	r.putF64(x3*r.k, prec)
	r.put(" ")
	r.putF64((h-y3)*r.k, prec)
	r.put(" c \n")
}

// ClipRoundedRect begins a rectangular clipping operation. The rectangle is of
// width w and height h. Its upper left corner is positioned at point (x, y).
// The rounded corners of the rectangle are specified by radius r. outline is
// true to draw a border with the current draw color and line width centered on
// the rectangle's perimeter. Only the outer half of the border will be shown.
// After calling this method, all rendering operations (for example, Image(),
// LinearGradient(), etc) will be clipped by the specified rectangle. Call
// ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (r *Renderer) ClipRoundedRect(x, y, w, h, radius float64, outline bool) {
	r.ClipRoundedRectExt(x, y, w, h, radius, radius, radius, radius, outline)
}

// ClipRoundedRectExt behaves the same as ClipRoundedRect() but supports a
// different radius for each corner, given by rTL (top-left), rTR (top-right)
// rBR (bottom-right), rBL (bottom-left). See ClipRoundedRect() for more
// details. This method is demonstrated in the ClipText() example.
func (r *Renderer) ClipRoundedRectExt(x, y, w, h, rTL, rTR, rBR, rBL float64, outline bool) {
	r.clipNest++
	r.roundedRectPath(x, y, w, h, rTL, rTR, rBR, rBL)
	op := "n"
	if outline {
		op = "S"
	}
	r.outf(" W %s", op)
}

// add a rectangle path with rounded corners.
// routine shared by RoundedRect() and ClipRoundedRect(), which add the
// drawing operation
func (r *Renderer) roundedRectPath(x, y, w, h, rTL, rTR, rBR, rBL float64) {
	k := r.k
	hp := r.h
	myArc := (4.0 / 3.0) * (math.Sqrt2 - 1.0)
	// r.outf("q %.5f %.5f m", (x+rTL)*k, (hp-y)*k)
	const prec = 5
	r.put("q ")
	r.putF64((x+rTL)*k, prec)
	r.put(" ")
	r.putF64((hp-y)*k, prec)
	r.put(" m\n")
	xc := x + w - rTR
	yc := y + rTR
	// r.outf("%.5f %.5f l", xc*k, (hp-y)*k)
	r.putF64(xc*k, prec)
	r.put(" ")
	r.putF64((hp-y)*k, prec)
	r.put(" l\n")
	if rTR != 0 {
		r.clipArc(xc+rTR*myArc, yc-rTR, xc+rTR, yc-rTR*myArc, xc+rTR, yc)
	}
	xc = x + w - rBR
	yc = y + h - rBR
	// r.outf("%.5f %.5f l", (x+w)*k, (hp-yc)*k)
	r.putF64((x+w)*k, prec)
	r.put(" ")
	r.putF64((hp-yc)*k, prec)
	r.put(" l\n")
	if rBR != 0 {
		r.clipArc(xc+rBR, yc+rBR*myArc, xc+rBR*myArc, yc+rBR, xc, yc+rBR)
	}
	xc = x + rBL
	yc = y + h - rBL
	// r.outf("%.5f %.5f l", xc*k, (hp-(y+h))*k)
	r.putF64(xc*k, prec)
	r.put(" ")
	r.putF64((hp-(y+h))*k, prec)
	r.put(" l\n")
	if rBL != 0 {
		r.clipArc(xc-rBL*myArc, yc+rBL, xc-rBL, yc+rBL*myArc, xc-rBL, yc)
	}
	xc = x + rTL
	yc = y + rTL
	// r.outf("%.5f %.5f l", x*k, (hp-yc)*k)
	r.putF64(x*k, prec)
	r.put(" ")
	r.putF64((hp-yc)*k, prec)
	r.put(" l\n")
	if rTL != 0 {
		r.clipArc(xc-rTL, yc-rTL*myArc, xc-rTL*myArc, yc-rTL, xc, yc-rTL)
	}
}

// ClipEllipse begins an elliptical clipping operation. The ellipse is centered
// at (x, y). Its horizontal and vertical radii are specified by rx and ry.
// outline is true to draw a border with the current draw color and line width
// centered on the ellipse's perimeter. Only the outer half of the border will
// be shown. After calling this method, all rendering operations (for example,
// Image(), LinearGradient(), etc) will be clipped by the specified ellipse.
// Call ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (r *Renderer) ClipEllipse(x, y, rx, ry float64, outline bool) {
	r.clipNest++
	lx := (4.0 / 3.0) * rx * (math.Sqrt2 - 1)
	ly := (4.0 / 3.0) * ry * (math.Sqrt2 - 1)
	k := r.k
	h := r.h
	//	r.outf("q %.5f %.5f m %.5f %.5f %.5f %.5f %.5f %.5f c",
	//		(x+rx)*k, (h-y)*k,
	//		(x+rx)*k, (h-(y-ly))*k,
	//		(x+lx)*k, (h-(y-ry))*k,
	//		x*k, (h-(y-ry))*k)
	const prec = 5
	r.put("q ")
	r.putF64((x+rx)*k, prec)
	r.put(" ")
	r.putF64((h-y)*k, prec)
	r.put(" m ")
	r.putF64((x+rx)*k, prec)
	r.put(" ")
	r.putF64((h-(y-ly))*k, prec)
	r.put(" ")
	r.putF64((x+lx)*k, prec)
	r.put(" ")
	r.putF64((h-(y-ry))*k, prec)
	r.put(" ")
	r.putF64(x*k, prec)
	r.put(" ")
	r.putF64((h-(y-ry))*k, prec)
	r.put(" c\n")

	//	r.outf("%.5f %.5f %.5f %.5f %.5f %.5f c",
	//		(x-lx)*k, (h-(y-ry))*k,
	//		(x-rx)*k, (h-(y-ly))*k,
	//		(x-rx)*k, (h-y)*k)
	r.putF64((x-lx)*k, prec)
	r.put(" ")
	r.putF64((h-(y-ry))*k, prec)
	r.put(" ")
	r.putF64((x-rx)*k, prec)
	r.put(" ")
	r.putF64((h-(y-ly))*k, prec)
	r.put(" ")
	r.putF64((x-rx)*k, prec)
	r.put(" ")
	r.putF64((h-y)*k, prec)
	r.put(" c\n")

	//	r.outf("%.5f %.5f %.5f %.5f %.5f %.5f c",
	//		(x-rx)*k, (h-(y+ly))*k,
	//		(x-lx)*k, (h-(y+ry))*k,
	//		x*k, (h-(y+ry))*k)
	r.putF64((x-rx)*k, prec)
	r.put(" ")
	r.putF64((h-(y+ly))*k, prec)
	r.put(" ")
	r.putF64((x-lx)*k, prec)
	r.put(" ")
	r.putF64((h-(y+ry))*k, prec)
	r.put(" ")
	r.putF64(x*k, prec)
	r.put(" ")
	r.putF64((h-(y+ry))*k, prec)
	r.put(" c\n")

	//	r.outf("%.5f %.5f %.5f %.5f %.5f %.5f c W %s",
	//		(x+lx)*k, (h-(y+ry))*k,
	//		(x+rx)*k, (h-(y+ly))*k,
	//		(x+rx)*k, (h-y)*k,
	//		strIf(outline, "S", "n"))
	r.putF64((x+lx)*k, prec)
	r.put(" ")
	r.putF64((h-(y+ry))*k, prec)
	r.put(" ")
	r.putF64((x+rx)*k, prec)
	r.put(" ")
	r.putF64((h-(y+ly))*k, prec)
	r.put(" ")
	r.putF64((x+rx)*k, prec)
	r.put(" ")
	r.putF64((h-y)*k, prec)
	op := "n"
	if outline {
		op = "S"
	}
	r.put(" c W " + op + "\n")
}

// ClipCircle begins a circular clipping operation. The circle is centered at
// (x, y) and has radius r. outline is true to draw a border with the current
// draw color and line width centered on the circle's perimeter. Only the outer
// half of the border will be shown. After calling this method, all rendering
// operations (for example, Image(), LinearGradient(), etc) will be clipped by
// the specified circle. Call ClipEnd() to restore unclipped operations.
//
// The ClipText() example demonstrates this method.
func (r *Renderer) ClipCircle(x, y, radius float64, outline bool) {
	r.ClipEllipse(x, y, radius, radius, outline)
}

// ClipPolygon begins a clipping operation within a polygon. The figure is
// defined by a series of vertices specified by points. The x and y fields of
// the points use the units established in New(). The last point in the slice
// will be implicitly joined to the first to close the polygon. outline is true
// to draw a border with the current draw color and line width centered on the
// polygon's perimeter. Only the outer half of the border will be shown. After
// calling this method, all rendering operations (for example, Image(),
// LinearGradient(), etc) will be clipped by the specified polygon. Call
// ClipEnd() to restore unclipped operations.
//
// The ClipText() example demonstrates this method.
func (r *Renderer) ClipPolygon(points []Point, outline bool) {
	r.clipNest++
	var s bytes.Buffer
	h := r.h
	k := r.k
	fmt.Fprintf(&s, "q ")
	for j, pt := range points {
		op := "l"
		if j == 0 {
			op = "m"
		}
		fmt.Fprintf(&s, "%.5f %.5f %s ", pt.X*k, (h-pt.Y)*k, op)
	}
	op := "n"
	if outline {
		op = "S"
	}
	fmt.Fprintf(&s, "h W %s", op)
	r.out(s.String())
}

// ClipEnd ends a clipping operation that was started with a call to
// ClipRect(), ClipRoundedRect(), ClipText(), ClipEllipse(), ClipCircle() or
// ClipPolygon(). Clipping operations can be nested. The document cannot be
// successfully output while a clipping operation is active.
//
// The ClipText() example demonstrates this method.
func (r *Renderer) ClipEnd() {
	if r.err == nil {
		if r.clipNest > 0 {
			r.clipNest--
			r.out("Q")
		} else {
			r.err = errs.New("error attempting to end clip operation out of sequence")
		}
	}
}

// AddFont imports a TrueType, OpenType or Type1 font and makes it available.
// It is necessary to generate a font definition file first with the makefont
// utility. It is not necessary to call this function for the core PDF fonts
// (courier, helvetica, times, zapfdingbats).
//
// The JSON definition file (and the font file itself when embedding) must be
// present in the font directory. If it is not found, the error "Could not
// include font definition file" is set.
//
// family specifies the font family. The name can be chosen arbitrarily. If it
// is a standard family name, it will override the corresponding font. This
// string is used to subsequently set the font with the SetFont method.
//
// style specifies the font style. Acceptable values are (case insensitive) the
// empty string for regular style, "B" for bold, "I" for italic, or "BI" or
// "IB" for bold and italic combined.
//
// fileStr specifies the base name with ".json" extension of the font
// definition file to be added. The file will be loaded from the font directory
// specified in the call to New() or SetFontLocation().
func (r *Renderer) AddFont(familyStr, styleStr, fileStr string) {
	r.addFont(fontFamilyEscape(familyStr), styleStr, fileStr, false)
}

// AddUTF8Font imports a TrueType font with utf-8 symbols and makes it available.
// It is necessary to generate a font definition file first with the makefont
// utility. It is not necessary to call this function for the core PDF fonts
// (courier, helvetica, times, zapfdingbats).
//
// The JSON definition file (and the font file itself when embedding) must be
// present in the font directory. If it is not found, the error "Could not
// include font definition file" is set.
//
// family specifies the font family. The name can be chosen arbitrarily. If it
// is a standard family name, it will override the corresponding font. This
// string is used to subsequently set the font with the SetFont method.
//
// style specifies the font style. Acceptable values are (case insensitive) the
// empty string for regular style, "B" for bold, "I" for italic, or "BI" or
// "IB" for bold and italic combined.
//
// fileStr specifies the base name with ".json" extension of the font
// definition file to be added. The file will be loaded from the font directory
// specified in the call to New() or SetFontLocation().
func (r *Renderer) AddUTF8Font(familyStr, styleStr, fileStr string) {
	r.addFont(fontFamilyEscape(familyStr), styleStr, fileStr, true)
}

func (r *Renderer) addFont(familyStr, styleStr, fileStr string, isUTF8 bool) {
	if fileStr == "" {
		if isUTF8 {
			fileStr = strings.ReplaceAll(familyStr, " ", "") + strings.ToLower(styleStr) + ".ttf"
		} else {
			fileStr = strings.ReplaceAll(familyStr, " ", "") + strings.ToLower(styleStr) + ".json"
		}
	}
	if isUTF8 {
		fontKey := getFontKey(familyStr, styleStr)
		_, ok := r.fonts[fontKey]
		if ok {
			return
		}
		var ttfStat os.FileInfo
		var err error
		fileStr = path.Join(r.fontpath, fileStr)
		ttfStat, err = os.Stat(fileStr)
		if err != nil {
			r.SetError(err)
			return
		}
		originalSize := ttfStat.Size()
		Type := "UTF8"
		var utf8Bytes []byte
		utf8Bytes, err = os.ReadFile(fileStr)
		if err != nil {
			r.SetError(err)
			return
		}
		reader := fileReader{readerPosition: 0, array: utf8Bytes}
		utf8File := newUTF8Font(&reader)
		err = utf8File.parseFile()
		if err != nil {
			r.SetError(err)
			return
		}

		desc := FontDesc{
			Ascent:       int(utf8File.Ascent),
			Descent:      int(utf8File.Descent),
			CapHeight:    utf8File.CapHeight,
			Flags:        utf8File.Flags,
			FontBBox:     utf8File.Bbox,
			ItalicAngle:  utf8File.ItalicAngle,
			StemV:        utf8File.StemV,
			MissingWidth: int(math.Round(utf8File.DefaultWidth)),
		}

		var sbarr map[int]int
		if r.aliasNbPagesStr == "" {
			sbarr = makeSubsetRange(57)
		} else {
			sbarr = makeSubsetRange(32)
		}
		def := fontDef{
			Tp:        Type,
			Name:      fontKey,
			Desc:      desc,
			Up:        int(math.Round(utf8File.UnderlinePosition)),
			Ut:        int(math.Round(utf8File.UnderlineThickness)),
			Cw:        utf8File.CharWidths,
			usedRunes: sbarr,
			File:      fileStr,
			utf8File:  utf8File,
		}
		def.i, _ = generateFontID(def)
		r.fonts[fontKey] = def
		r.fontFiles[fontKey] = fontFile{
			length1:  originalSize,
			fontType: "UTF8",
		}
		r.fontFiles[fileStr] = fontFile{
			fontType: "UTF8",
		}
	} else {
		if r.fontLoader != nil {
			reader, err := r.fontLoader.Open(fileStr)
			if err == nil {
				r.AddFontFromReader(familyStr, styleStr, reader)
				if closer, ok := reader.(io.Closer); ok {
					closer.Close()
				}
				return
			}
		}

		fileStr = path.Join(r.fontpath, fileStr)
		file, err := os.Open(fileStr)
		if err != nil {
			r.err = err
			return
		}
		defer file.Close()

		r.AddFontFromReader(familyStr, styleStr, file)
	}
}

func makeSubsetRange(end int) map[int]int {
	answer := make(map[int]int)
	for i := range end {
		answer[i] = 0
	}
	return answer
}

// AddFontFromBytes imports a TrueType, OpenType or Type1 font from static
// bytes within the executable and makes it available for use in the generated
// document.
//
// family specifies the font family. The name can be chosen arbitrarily. If it
// is a standard family name, it will override the corresponding font. This
// string is used to subsequently set the font with the SetFont method.
//
// style specifies the font style. Acceptable values are (case insensitive) the
// empty string for regular style, "B" for bold, "I" for italic, or "BI" or
// "IB" for bold and italic combined.
//
// jsonFileBytes contain all bytes of JSON file.
//
// zFileBytes contain all bytes of Z file.
func (r *Renderer) AddFontFromBytes(familyStr, styleStr string, jsonFileBytes, zFileBytes []byte) {
	r.addFontFromBytes(fontFamilyEscape(familyStr), styleStr, jsonFileBytes, zFileBytes, nil)
}

// AddUTF8FontFromBytes  imports a TrueType font with utf-8 symbols from static
// bytes within the executable and makes it available for use in the generated
// document.
//
// family specifies the font family. The name can be chosen arbitrarily. If it
// is a standard family name, it will override the corresponding font. This
// string is used to subsequently set the font with the SetFont method.
//
// style specifies the font style. Acceptable values are (case insensitive) the
// empty string for regular style, "B" for bold, "I" for italic, or "BI" or
// "IB" for bold and italic combined.
//
// jsonFileBytes contain all bytes of JSON file.
//
// zFileBytes contain all bytes of Z file.
func (r *Renderer) AddUTF8FontFromBytes(familyStr, styleStr string, utf8Bytes []byte) {
	r.addFontFromBytes(fontFamilyEscape(familyStr), styleStr, nil, nil, utf8Bytes)
}

func (r *Renderer) addFontFromBytes(familyStr, styleStr string, jsonFileBytes, zFileBytes, utf8Bytes []byte) {
	if r.err != nil {
		return
	}

	// load font key
	var ok bool
	fontkey := getFontKey(familyStr, styleStr)
	_, ok = r.fonts[fontkey]

	if ok {
		return
	}

	if utf8Bytes != nil {

		// if styleStr == "IB" {
		// 	styleStr = "BI"
		// }

		Type := "UTF8"
		reader := fileReader{readerPosition: 0, array: utf8Bytes}

		utf8File := newUTF8Font(&reader)

		err := utf8File.parseFile()
		if err != nil {
			r.SetError(err)
			return
		}
		desc := FontDesc{
			Ascent:       int(utf8File.Ascent),
			Descent:      int(utf8File.Descent),
			CapHeight:    utf8File.CapHeight,
			Flags:        utf8File.Flags,
			FontBBox:     utf8File.Bbox,
			ItalicAngle:  utf8File.ItalicAngle,
			StemV:        utf8File.StemV,
			MissingWidth: int(math.Round(utf8File.DefaultWidth)),
		}

		var sbarr map[int]int
		if r.aliasNbPagesStr == "" {
			sbarr = makeSubsetRange(57)
		} else {
			sbarr = makeSubsetRange(32)
		}
		def := fontDef{
			Tp:        Type,
			Name:      fontkey,
			Desc:      desc,
			Up:        int(math.Round(utf8File.UnderlinePosition)),
			Ut:        int(math.Round(utf8File.UnderlineThickness)),
			Cw:        utf8File.CharWidths,
			utf8File:  utf8File,
			usedRunes: sbarr,
		}
		def.i, _ = generateFontID(def)
		r.fonts[fontkey] = def
	} else {
		// load font definitions
		var info fontDef
		err := json.Unmarshal(jsonFileBytes, &info)

		if err != nil {
			r.err = err
		}

		if r.err != nil {
			return
		}

		if info.i, err = generateFontID(info); err != nil {
			r.err = err
			return
		}

		// search existing encodings
		if len(info.Diff) > 0 {
			n := -1

			for j, str := range r.diffs {
				if str == info.Diff {
					n = j + 1
					break
				}
			}

			if n < 0 {
				r.diffs = append(r.diffs, info.Diff)
				n = len(r.diffs)
			}

			info.DiffN = n
		}

		// embed font
		if len(info.File) > 0 {
			if info.Tp == "TrueType" {
				r.fontFiles[info.File] = fontFile{
					length1:  int64(info.OriginalSize),
					embedded: true,
					content:  zFileBytes,
				}
			} else {
				r.fontFiles[info.File] = fontFile{
					length1:  int64(info.Size1),
					length2:  int64(info.Size2),
					embedded: true,
					content:  zFileBytes,
				}
			}
		}

		r.fonts[fontkey] = info
	}
}

// getFontKey is used by AddFontFromReader and GetFontDesc
func getFontKey(familyStr, styleStr string) string {
	familyStr = strings.ToLower(familyStr)
	styleStr = strings.ToUpper(styleStr)
	if styleStr == "IB" {
		styleStr = "BI"
	}
	return familyStr + styleStr
}

// AddFontFromReader imports a TrueType, OpenType or Type1 font and makes it
// available using a reader that satisifies the io.Reader interface. See
// AddFont for details about familyStr and styleStr.
func (r *Renderer) AddFontFromReader(familyStr, styleStr string, rd io.Reader) {
	if r.err != nil {
		return
	}
	// dbg("Adding family [%s], style [%s]", familyStr, styleStr)
	familyStr = fontFamilyEscape(familyStr)
	var ok bool
	fontkey := getFontKey(familyStr, styleStr)
	_, ok = r.fonts[fontkey]
	if ok {
		return
	}
	info := r.loadfont(rd)
	if r.err != nil {
		return
	}
	if len(info.Diff) > 0 {
		// Search existing encodings
		n := -1
		for j, str := range r.diffs {
			if str == info.Diff {
				n = j + 1
				break
			}
		}
		if n < 0 {
			r.diffs = append(r.diffs, info.Diff)
			n = len(r.diffs)
		}
		info.DiffN = n
	}
	// dbg("font [%s], type [%s]", info.File, info.Tp)
	if len(info.File) > 0 {
		// Embedded font
		if info.Tp == "TrueType" {
			r.fontFiles[info.File] = fontFile{length1: int64(info.OriginalSize)}
		} else {
			r.fontFiles[info.File] = fontFile{length1: int64(info.Size1), length2: int64(info.Size2)}
		}
	}
	r.fonts[fontkey] = info
}

// GetFontDesc returns the font descriptor, which can be used for
// example to find the baseline of a font. If familyStr is empty
// current font descriptor will be returned.
// See FontDesc for documentation about the font descriptor.
// See AddFont for details about familyStr and styleStr.
func (r *Renderer) GetFontDesc(familyStr, styleStr string) FontDesc {
	if familyStr == "" {
		return r.currentFont.Desc
	}
	return r.fonts[getFontKey(fontFamilyEscape(familyStr), styleStr)].Desc
}

// SetFont sets the font used to print character strings. It is mandatory to
// call this method at least once before printing text or the resulting
// document will not be valid.
//
// The font can be either a standard one or a font added via the AddFont()
// method or AddFontFromReader() method. Standard fonts use the Windows
// encoding cp1252 (Western Europe).
//
// The method can be called before the first page is created and the font is
// kept from page to page. If you just wish to change the current font size, it
// is simpler to call SetFontSize().
//
// Note: the font definition file must be accessible. An error is set if the
// file cannot be read.
//
// familyStr specifies the font family. It can be either a name defined by
// AddFont(), AddFontFromReader() or one of the standard families (case
// insensitive): "Courier" for fixed-width, "Helvetica" or "Arial" for sans
// serif, "Times" for serif, "Symbol" or "ZapfDingbats" for symbolic.
//
// styleStr can be "B" (bold), "I" (italic), "U" (underscore), "S" (strike-out)
// or any combination. The default value (specified with an empty string) is
// regular. Bold and italic styles do not apply to Symbol and ZapfDingbats.
//
// size is the font size measured in points. The default value is the current
// size. If no size has been specified since the beginning of the document, the
// value taken is 12.
func (r *Renderer) SetFont(familyStr, styleStr string, size float64) {
	// dbg("SetFont x %.2f, lMargin %.2f", r.x, r.lMargin)

	if r.err != nil {
		return
	}
	// dbg("SetFont")
	familyStr = fontFamilyEscape(familyStr)
	var ok bool
	familyStr = cmp.Or(strings.ToLower(familyStr), r.fontFamily)
	styleStr = strings.ToUpper(styleStr)
	r.underline = strings.Contains(styleStr, "U")
	if r.underline {
		styleStr = strings.ReplaceAll(styleStr, "U", "")
	}
	r.strikeout = strings.Contains(styleStr, "S")
	if r.strikeout {
		styleStr = strings.ReplaceAll(styleStr, "S", "")
	}
	if styleStr == "IB" {
		styleStr = "BI"
	}
	if size == 0.0 {
		size = r.fontSizePt
	}

	// Test if font is already loaded
	fontKey := familyStr + styleStr
	_, ok = r.fonts[fontKey]
	if !ok {
		// Test if one of the core fonts
		if familyStr == "arial" {
			familyStr = "helvetica"
		}
		_, ok = r.coreFonts[familyStr]
		if ok {
			if familyStr == "symbol" {
				familyStr = "zapfdingbats"
			}
			if familyStr == "zapfdingbats" {
				styleStr = ""
			}
			fontKey = familyStr + styleStr
			_, ok = r.fonts[fontKey]
			if !ok {
				rdr := r.coreFontReader(familyStr, styleStr)
				if r.err == nil {
					defer rdr.Close()
					r.AddFontFromReader(familyStr, styleStr, rdr)
				}
				if r.err != nil {
					return
				}
			}
		} else {
			r.err = fmt.Errorf("undefined font: %s %s", familyStr, styleStr)
			return
		}
	}
	// Select it
	r.fontFamily = familyStr
	r.fontStyle = styleStr
	r.fontSizePt = size
	r.fontSize = size / r.k
	r.currentFont = r.fonts[fontKey]
	if r.currentFont.Tp == "UTF8" {
		r.isCurrentUTF8 = true
	} else {
		r.isCurrentUTF8 = false
	}
	if r.page > 0 {
		r.outf("BT /F%s %.2f Tf ET", r.currentFont.i, r.fontSizePt)
	}
}

// GetFontFamily returns the family of the current font. See SetFont() for details.
func (r *Renderer) GetFontFamily() string {
	return r.fontFamily
}

// GetFontStyle returns the style of the current font. See SetFont() for details.
func (r *Renderer) GetFontStyle() string {
	styleStr := r.fontStyle

	if r.underline {
		styleStr += "U"
	}
	if r.strikeout {
		styleStr += "S"
	}

	return styleStr
}

// SetFontStyle sets the style of the current font. See also SetFont()
func (r *Renderer) SetFontStyle(styleStr string) {
	r.SetFont(r.fontFamily, styleStr, r.fontSizePt)
}

// SetFontSize defines the size of the current font. Size is specified in
// points (1/ 72 inch). See also SetFontUnitSize().
func (r *Renderer) SetFontSize(size float64) {
	r.fontSizePt = size
	r.fontSize = size / r.k
	if r.page > 0 {
		r.outf("BT /F%s %.2f Tf ET", r.currentFont.i, r.fontSizePt)
	}
}

// SetFontUnitSize defines the size of the current font. Size is specified in
// the unit of measure specified in New(). See also SetFontSize().
func (r *Renderer) SetFontUnitSize(size float64) {
	r.fontSizePt = size * r.k
	r.fontSize = size
	if r.page > 0 {
		r.outf("BT /F%s %.2f Tf ET", r.currentFont.i, r.fontSizePt)
	}
}

// GetFontSize returns the size of the current font in points followed by the
// size in the unit of measure specified in New(). The second value can be used
// as a line height value in drawing operations.
func (r *Renderer) GetFontSize() (ptSize, unitSize float64) {
	return r.fontSizePt, r.fontSize
}

// AddLink creates a new internal link and returns its identifier. An internal
// link is a clickable area which directs to another place within the document.
// The identifier can then be passed to Cell(), Write(), Image() or Link(). The
// destination is defined with SetLink().
func (r *Renderer) AddLink() int {
	r.links = append(r.links, internalPageLink{})
	return len(r.links) - 1
}

// SetLink defines the page and position a link points to. See AddLink().
func (r *Renderer) SetLink(link int, y float64, page int) {
	if y == -1 {
		y = r.y
	}
	if page == -1 {
		page = r.page
	}
	r.links[link] = internalPageLink{page, y}
}

// newLink adds a new clickable link on current page
func (r *Renderer) newLink(x, y, w, h float64, link int, linkStr string) {
	// linkList, ok := r.pageLinks[r.page]
	// if !ok {
	// linkList = make([]linkType, 0, 8)
	// r.pageLinks[r.page] = linkList
	// }
	r.pageLinks[r.page] = append(r.pageLinks[r.page],
		pageLink{x * r.k, r.hPt - y*r.k, w * r.k, h * r.k, link, linkStr})
}

// Link puts a link on a rectangular area of the page. Text or image links are
// generally put via Cell(), Write() or Image(), but this method can be useful
// for instance to define a clickable area inside an image. link is the value
// returned by AddLink().
func (r *Renderer) Link(x, y, w, h float64, link int) {
	r.newLink(x, y, w, h, link, "")
}

// LinkString puts a link on a rectangular area of the page. Text or image
// links are generally put via Cell(), Write() or Image(), but this method can
// be useful for instance to define a clickable area inside an image. linkStr
// is the target URL.
func (r *Renderer) LinkString(x, y, w, h float64, linkStr string) {
	r.newLink(x, y, w, h, 0, linkStr)
}

// Bookmark sets a bookmark that will be displayed in a sidebar outline. txtStr
// is the title of the bookmark. level specifies the level of the bookmark in
// the outline; 0 is the top level, 1 is just below, and so on. y specifies the
// vertical position of the bookmark destination in the current page; -1
// indicates the current position.
func (r *Renderer) Bookmark(txtStr string, level int, y float64) {
	if y == -1 {
		y = r.y
	}
	if r.isCurrentUTF8 {
		txtStr = utf8toutf16(txtStr)
	}
	r.outlines = append(r.outlines, outline{text: txtStr, level: level, y: y, p: r.PageNo(), prev: -1, last: -1, next: -1, first: -1})
}

// Text prints a character string. The origin (x, y) is on the left of the
// first character at the baseline. This method permits a string to be placed
// precisely on the page, but it is usually easier to use Cell(), MultiCell()
// or Write() which are the standard methods to print text.
func (r *Renderer) Text(x, y float64, txtStr string) {
	var txt2 string
	if r.isCurrentUTF8 {
		if r.isRTL {
			txtStr = reverseText(txtStr)
			x -= r.GetStringWidth(txtStr)
		}
		txt2 = r.escape(utf8toutf16(txtStr, false))
		for _, uni := range txtStr {
			r.currentFont.usedRunes[int(uni)] = int(uni)
		}
	} else {
		txt2 = r.escape(txtStr)
	}
	s := fmt.Sprintf("BT %.2f %.2f Td (%s) Tj ET", x*r.k, (r.h-y)*r.k, txt2)
	if r.underline && txtStr != "" {
		s += " " + r.dounderline(x, y, txtStr)
	}
	if r.strikeout && txtStr != "" {
		s += " " + r.dostrikeout(x, y, txtStr)
	}
	if r.colorFlag {
		s = fmt.Sprintf("q %s %s Q", r.color.text.str, s)
	}
	r.out(s)
}

// GetWordSpacing returns the spacing between words of following text.
func (r *Renderer) GetWordSpacing() float64 {
	return r.ws
}

// SetWordSpacing sets spacing between words of following text. See the
// WriteAligned() example for a demonstration of its use.
func (r *Renderer) SetWordSpacing(space float64) {
	r.ws = space
	r.out(fmt.Sprintf("%.5f Tw", space*r.k))
}

// SetTextRenderingMode sets the rendering mode of following text.
// The mode can be as follows:
// 0: Fill text
// 1: Stroke text
// 2: Fill, then stroke text
// 3: Neither fill nor stroke text (invisible)
// 4: Fill text and add to path for clipping
// 5: Stroke text and add to path for clipping
// 6: Fills then stroke text and add to path for clipping
// 7: Add text to path for clipping
// This method is demonstrated in the SetTextRenderingMode example.
func (r *Renderer) SetTextRenderingMode(mode int) {
	if mode >= 0 && mode <= 7 {
		r.out(fmt.Sprintf("%d Tr", mode))
	}
}

// SetAcceptPageBreakFunc allows the application to control where page breaks
// occur.
//
// fnc is an application function (typically a closure) that is called by the
// library whenever a page break condition is met. The break is issued if true
// is returned. The default implementation returns a value according to the
// mode selected by SetAutoPageBreak. The function provided should not be
// called by the application.
//
// See the example for SetLeftMargin() to see how this function can be used to
// manage multiple columns.
func (r *Renderer) SetAcceptPageBreakFunc(fnc func() bool) {
	r.acceptPageBreak = fnc
}

// CellFormat prints a rectangular cell with optional borders, background color
// and character string. The upper-left corner of the cell corresponds to the
// current position. The text can be aligned or centered. After the call, the
// current position moves to the right or to the next line. It is possible to
// put a link on the text.
//
// An error will be returned if a call to SetFont() has not already taken
// place before this method is called.
//
// If automatic page breaking is enabled and the cell goes beyond the limit, a
// page break is done before outputting.
//
// w and h specify the width and height of the cell. If w is 0, the cell
// extends up to the right margin. Specifying 0 for h will result in no output,
// but the current position will be advanced by w.
//
// txtStr specifies the text to display.
//
// borderStr specifies how the cell border will be drawn. An empty string
// indicates no border, "1" indicates a full border, and one or more of "L",
// "T", "R" and "B" indicate the left, top, right and bottom sides of the
// border.
//
// ln indicates where the current position should go after the call. Possible
// values are 0 (to the right), 1 (to the beginning of the next line), and 2
// (below). Putting 1 is equivalent to putting 0 and calling Ln() just after.
//
// alignStr specifies how the text is to be positioned within the cell.
// Horizontal alignment is controlled by including "L", "C" or "R" (left,
// center, right) in alignStr. Vertical alignment is controlled by including
// "T", "M", "B" or "A" (top, middle, bottom, baseline) in alignStr. The default
// alignment is left middle.
//
// fill is true to paint the cell background or false to leave it transparent.
//
// link is the identifier returned by AddLink() or 0 for no internal link.
//
// linkStr is a target URL or empty for no external link. A non--zero value for
// link takes precedence over linkStr.
func (r *Renderer) CellFormat(w, h float64, txtStr, borderStr string, ln int,
	alignStr string, fill bool, link int, linkStr string) {
	// dbg("CellFormat. h = %.2f, borderStr = %s", h, borderStr)
	if r.err != nil {
		return
	}

	if r.currentFont.Name == "" {
		r.err = errs.New("font has not been set; unable to render text")
		return
	}

	borderStr = strings.ToUpper(borderStr)
	k := r.k
	if r.y+h > r.pageBreakTrigger && !r.inHeader && !r.inFooter && r.acceptPageBreak() {
		// Automatic page break
		x := r.x
		ws := r.ws
		// dbg("auto page break, x %.2f, ws %.2f", x, ws)
		if ws > 0 {
			r.ws = 0
			r.out("0 Tw")
		}
		r.AddPageFormat(r.curOrientation, r.curPageSize)
		if r.err != nil {
			return
		}
		r.x = x
		if ws > 0 {
			r.ws = ws
			// r.outf("%.3f Tw", ws*k)
			r.putF64(ws*k, 3)
			r.put(" Tw\n")
		}
	}
	if w == 0 {
		w = r.w - r.rMargin - r.x
	}
	var s bytes.Buffer
	if h > 0 && (fill || borderStr == "1") {
		var op string
		if fill {
			if borderStr == "1" {
				op = "B"
				// dbg("border is '1', fill")
			} else {
				op = "f"
				// dbg("border is empty, fill")
			}
		} else {
			// dbg("border is '1', no fill")
			op = "S"
		}
		/// dbg("(CellFormat) r.x %.2f r.k %.2f", r.x, r.k)
		fmt.Fprintf(&s, "%.2f %.2f %.2f %.2f re %s ", r.x*k, (r.h-r.y)*k, w*k, -h*k, op)
	}
	if len(borderStr) > 0 && borderStr != "1" {
		// fmt.Printf("border is '%s', no fill\n", borderStr)
		x := r.x
		y := r.y
		left := x * k
		top := (r.h - y) * k
		right := (x + w) * k
		bottom := (r.h - (y + h)) * k
		if strings.Contains(borderStr, "L") {
			fmt.Fprintf(&s, "%.2f %.2f m %.2f %.2f l S ", left, top, left, bottom)
		}
		if strings.Contains(borderStr, "T") {
			fmt.Fprintf(&s, "%.2f %.2f m %.2f %.2f l S ", left, top, right, top)
		}
		if strings.Contains(borderStr, "R") {
			fmt.Fprintf(&s, "%.2f %.2f m %.2f %.2f l S ", right, top, right, bottom)
		}
		if strings.Contains(borderStr, "B") {
			fmt.Fprintf(&s, "%.2f %.2f m %.2f %.2f l S ", left, bottom, right, bottom)
		}
	}
	if len(txtStr) > 0 {
		var dx, dy float64
		// Horizontal alignment
		switch {
		case strings.Contains(alignStr, "R"):
			dx = w - r.cMargin - r.GetStringWidth(txtStr)
		case strings.Contains(alignStr, "C"):
			dx = (w - r.GetStringWidth(txtStr)) / 2
		default:
			dx = r.cMargin
		}

		// Vertical alignment
		switch {
		case strings.Contains(alignStr, "T"):
			dy = (r.fontSize - h) / 2.0
		case strings.Contains(alignStr, "B"):
			dy = (h - r.fontSize) / 2.0
		case strings.Contains(alignStr, "A"):
			var descent float64
			d := r.currentFont.Desc
			if d.Descent == 0 {
				// not defined (standard font?), use average of 19%
				descent = -0.19 * r.fontSize
			} else {
				descent = float64(d.Descent) * r.fontSize / float64(d.Ascent-d.Descent)
			}
			dy = (h-r.fontSize)/2.0 - descent
		default:
			dy = 0
		}
		if r.colorFlag {
			fmt.Fprintf(&s, "q %s ", r.color.text.str)
		}
		//If multibyte, Tw has no effect - do word spacing using an adjustment before each space
		if (r.ws != 0 || alignStr == "J") && r.isCurrentUTF8 { // && r.ws != 0
			if r.isRTL {
				txtStr = reverseText(txtStr)
			}
			wmax := int(math.Ceil((w - 2*r.cMargin) * 1000 / r.fontSize))
			for _, uni := range txtStr {
				r.currentFont.usedRunes[int(uni)] = int(uni)
			}
			space := r.escape(utf8toutf16(" ", false))
			strSize := r.GetStringSymbolWidth(txtStr)
			fmt.Fprintf(&s, "BT 0 Tw %.2f %.2f Td [", (r.x+dx)*k, (r.h-(r.y+.5*h+.3*r.fontSize))*k)
			t := strings.Split(txtStr, " ")
			shift := float64((wmax - strSize)) / float64(len(t)-1)
			numt := len(t)
			for i := range numt {
				tx := t[i]
				tx = "(" + r.escape(utf8toutf16(tx, false)) + ")"
				fmt.Fprintf(&s, "%s ", tx)
				if (i + 1) < numt {
					fmt.Fprintf(&s, "%.3f(%s) ", -shift, space)
				}
			}
			fmt.Fprintf(&s, "] TJ ET")
		} else {
			var txt2 string
			if r.isCurrentUTF8 {
				if r.isRTL {
					txtStr = reverseText(txtStr)
				}
				txt2 = r.escape(utf8toutf16(txtStr, false))
				for _, uni := range txtStr {
					r.currentFont.usedRunes[int(uni)] = int(uni)
				}
			} else {

				txt2 = strings.ReplaceAll(txtStr, "\\", "\\\\")
				txt2 = strings.ReplaceAll(txt2, "(", "\\(")
				txt2 = strings.ReplaceAll(txt2, ")", "\\)")
			}
			bt := (r.x + dx) * k
			td := (r.h - (r.y + dy + .5*h + .3*r.fontSize)) * k
			fmt.Fprintf(&s, "BT %.2f %.2f Td (%s)Tj ET", bt, td, txt2)
			//BT %.2F %.2F Td (%s) Tj ET',(r.x+dx)*k,(r.h-(r.y+.5*h+.3*r.FontSize))*k,txt2);
		}

		if r.underline {
			fmt.Fprintf(&s, " %s", r.dounderline(r.x+dx, r.y+dy+.5*h+.3*r.fontSize, txtStr))
		}
		if r.strikeout {
			fmt.Fprintf(&s, " %s", r.dostrikeout(r.x+dx, r.y+dy+.5*h+.3*r.fontSize, txtStr))
		}
		if r.colorFlag {
			fmt.Fprintf(&s, " Q")
		}
		if link > 0 || len(linkStr) > 0 {
			r.newLink(r.x+dx, r.y+dy+.5*h-.5*r.fontSize, r.GetStringWidth(txtStr), r.fontSize, link, linkStr)
		}
	}
	str := s.String()
	if len(str) > 0 {
		r.out(str)
	}
	r.lasth = h
	if ln > 0 {
		// Go to next line
		r.y += h
		if ln == 1 {
			r.x = r.lMargin
		}
	} else {
		r.x += w
	}
}

// Revert string to use in RTL languages
func reverseText(text string) string {
	oldText := []rune(text)
	newText := make([]rune, len(oldText))
	length := len(oldText) - 1
	for i, r := range oldText {
		newText[length-i] = r
	}
	return string(newText)
}

// Cell is a simpler version of CellFormat with no fill, border, links or
// special alignment. The Cell_strikeout() example demonstrates this method.
func (r *Renderer) Cell(w, h float64, txtStr string) {
	r.CellFormat(w, h, txtStr, "", 0, "L", false, 0, "")
}

// Cellf is a simpler printf-style version of CellFormat with no fill, border,
// links or special alignment. See documentation for the fmt package for
// details on fmtStr and args.
func (r *Renderer) Cellf(w, h float64, fmtStr string, args ...any) {
	r.CellFormat(w, h, fmt.Sprintf(fmtStr, args...), "", 0, "L", false, 0, "")
}

// SplitLines splits text into several lines using the current font. Each line
// has its length limited to a maximum width given by w. This function can be
// used to determine the total height of wrapped text for vertical placement
// purposes.
//
// This method is useful for codepage-based fonts only. For UTF-8 encoded text,
// use SplitText().
//
// You can use MultiCell if you want to print a text on several lines in a
// simple way.
func (r *Renderer) SplitLines(txt []byte, w float64) [][]byte {
	// Function contributed by Bruno Michel
	lines := [][]byte{}
	cw := r.currentFont.Cw
	wmax := int(math.Ceil((w - 2*r.cMargin) * 1000 / r.fontSize))
	s := bytes.ReplaceAll(txt, []byte("\r"), []byte{})
	nb := len(s)
	for nb > 0 && s[nb-1] == '\n' {
		nb--
	}
	s = s[0:nb]
	sep := -1
	i := 0
	j := 0
	l := 0
	for i < nb {
		c := s[i]
		l += cw[c]
		if c == ' ' || c == '\t' || c == '\n' {
			sep = i
		}
		if c == '\n' || l > wmax {
			if sep == -1 {
				if i == j {
					i++
				}
				sep = i
			} else {
				i = sep + 1
			}
			lines = append(lines, s[j:sep])
			sep = -1
			j = i
			l = 0
		} else {
			i++
		}
	}
	if i != j {
		lines = append(lines, s[j:i])
	}
	return lines
}

// MultiCell supports printing text with line breaks. They can be automatic (as
// soon as the text reaches the right border of the cell) or explicit (via the
// \n character). As many cells as necessary are output, one below the other.
//
// Text can be aligned, centered or justified. The cell block can be framed and
// the background painted. See CellFormat() for more details.
//
// The current position after calling MultiCell() is the beginning of the next
// line, equivalent to calling CellFormat with ln equal to 1.
//
// w is the width of the cells. A value of zero indicates cells that reach to
// the right margin.
//
// h indicates the line height of each cell in the unit of measure specified in New().
//
// Note: this method has a known bug that treats UTF-8 fonts differently than
// non-UTF-8 fonts. With UTF-8 fonts, all trailing newlines in txtStr are
// removed. With a non-UTF-8 font, if txtStr has one or more trailing newlines,
// only the last is removed. In the next major module version, the UTF-8 logic
// will be changed to match the non-UTF-8 logic. To prepare for that change,
// applications that use UTF-8 fonts and depend on having all trailing newlines
// removed should call strings.TrimRight(txtStr, "\r\n") before calling this
// method.
func (r *Renderer) MultiCell(w, h float64, txtStr, borderStr, alignStr string, fill bool) {
	if r.err != nil {
		return
	}
	// dbg("MultiCell")
	alignStr = cmp.Or(alignStr, "J")
	cw := r.currentFont.Cw
	if w == 0 {
		w = r.w - r.rMargin - r.x
	}
	wmax := int(math.Ceil((w - 2*r.cMargin) * 1000 / r.fontSize))
	s := strings.ReplaceAll(txtStr, "\r", "")
	srune := []rune(s)

	// remove extra line breaks
	var nb int
	if r.isCurrentUTF8 {
		nb = len(srune)
		for nb > 0 && srune[nb-1] == '\n' {
			nb--
		}
		srune = srune[0:nb]
	} else {
		nb = len(s)
		bytes2 := []byte(s)

		// for nb > 0 && bytes2[nb-1] == '\n' {

		// Prior to August 2019, if s ended with a newline, this code stripped it.
		// After that date, to be compatible with the UTF-8 code above, *all*
		// trailing newlines were removed. Because this regression caused at least
		// one application to break (see issue #333), the original behavior has been
		// reinstated with a caveat included in the documentation.
		if nb > 0 && bytes2[nb-1] == '\n' {
			nb--
		}
		s = s[0:nb]
	}
	// dbg("[%s]\n", s)
	var b, b2 string
	b = "0"
	if len(borderStr) > 0 {
		if borderStr == "1" {
			borderStr = "LTRB"
			b = "LRT"
			b2 = "LR"
		} else {
			b2 = ""
			if strings.Contains(borderStr, "L") {
				b2 += "L"
			}
			if strings.Contains(borderStr, "R") {
				b2 += "R"
			}
			if strings.Contains(borderStr, "T") {
				b = b2 + "T"
			} else {
				b = b2
			}
		}
	}
	sep := -1
	i := 0
	j := 0
	l := 0
	ls := 0
	ns := 0
	nl := 1
	for i < nb {
		// Get next character
		var c rune
		if r.isCurrentUTF8 {
			c = srune[i]
		} else {
			c = rune(s[i])
		}
		if c == '\n' {
			// Explicit line break
			if r.ws > 0 {
				r.ws = 0
				r.out("0 Tw")
			}

			if r.isCurrentUTF8 {
				newAlignStr := alignStr
				if newAlignStr == "J" {
					if r.isRTL {
						newAlignStr = "R"
					} else {
						newAlignStr = "L"
					}
				}
				r.CellFormat(w, h, string(srune[j:i]), b, 2, newAlignStr, fill, 0, "")
			} else {
				r.CellFormat(w, h, s[j:i], b, 2, alignStr, fill, 0, "")
			}
			i++
			sep = -1
			j = i
			l = 0
			ns = 0
			nl++
			if len(borderStr) > 0 && nl == 2 {
				b = b2
			}
			continue
		}
		// Chinese text (CJK Unified Ideographs, U+4E00–U+9FA5) has no word
		// spaces, so every such character is a permissible break point.
		if c == ' ' || (c >= 0x4e00 && c <= 0x9fa5) {
			sep = i
			ls = l
			ns++
		}
		if int(c) >= len(cw) {
			r.err = fmt.Errorf("character outside the supported range: %s", string(c))
			return
		}
		switch cw[int(c)] {
		case 0: // marker width 0 is used for missing symbols
			l += r.currentFont.Desc.MissingWidth
		case 65535: // marker width 65535 is used for zero-width symbols
		default:
			l += cw[int(c)]
		}
		if l > wmax {
			// Automatic line break
			if sep == -1 {
				if i == j {
					i++
				}
				if r.ws > 0 {
					r.ws = 0
					r.out("0 Tw")
				}
				if r.isCurrentUTF8 {
					r.CellFormat(w, h, string(srune[j:i]), b, 2, alignStr, fill, 0, "")
				} else {
					r.CellFormat(w, h, s[j:i], b, 2, alignStr, fill, 0, "")
				}
			} else {
				if alignStr == "J" {
					if ns > 1 {
						r.ws = float64((wmax-ls)/1000) * r.fontSize / float64(ns-1)
					} else {
						r.ws = 0
					}
					// r.outf("%.3f Tw", r.ws*r.k)
					r.putF64(r.ws*r.k, 3)
					r.put(" Tw\n")
				}
				if r.isCurrentUTF8 {
					r.CellFormat(w, h, string(srune[j:sep]), b, 2, alignStr, fill, 0, "")
				} else {
					r.CellFormat(w, h, s[j:sep], b, 2, alignStr, fill, 0, "")
				}
				i = sep + 1
			}
			sep = -1
			j = i
			l = 0
			ns = 0
			nl++
			if len(borderStr) > 0 && nl == 2 {
				b = b2
			}
		} else {
			i++
		}
	}
	// Last chunk
	if r.ws > 0 {
		r.ws = 0
		r.out("0 Tw")
	}
	if len(borderStr) > 0 && strings.Contains(borderStr, "B") {
		b += "B"
	}
	if r.isCurrentUTF8 {
		if alignStr == "J" {
			if r.isRTL {
				alignStr = "R"
			} else {
				alignStr = ""
			}
		}
		r.CellFormat(w, h, string(srune[j:i]), b, 2, alignStr, fill, 0, "")
	} else {
		r.CellFormat(w, h, s[j:i], b, 2, alignStr, fill, 0, "")
	}
	r.x = r.lMargin
}

// write outputs text in flowing mode
func (r *Renderer) write(h float64, txtStr string, link int, linkStr string) {
	// dbg("Write")
	cw := r.currentFont.Cw
	w := r.w - r.rMargin - r.x
	wmax := (w - 2*r.cMargin) * 1000 / r.fontSize
	s := strings.ReplaceAll(txtStr, "\r", "")
	var nb int
	if r.isCurrentUTF8 {
		nb = len([]rune(s))
		if nb == 1 && s == " " {
			r.x += r.GetStringWidth(s)
			return
		}
	} else {
		nb = len(s)
	}
	sep := -1
	i := 0
	j := 0
	l := 0.0
	nl := 1
	for i < nb {
		// Get next character
		var c rune
		if r.isCurrentUTF8 {
			c = []rune(s)[i]
		} else {
			c = rune(byte(s[i]))
		}
		if c == '\n' {
			// Explicit line break
			if r.isCurrentUTF8 {
				r.CellFormat(w, h, string([]rune(s)[j:i]), "", 2, "", false, link, linkStr)
			} else {
				r.CellFormat(w, h, s[j:i], "", 2, "", false, link, linkStr)
			}
			i++
			sep = -1
			j = i
			l = 0.0
			if nl == 1 {
				r.x = r.lMargin
				w = r.w - r.rMargin - r.x
				wmax = (w - 2*r.cMargin) * 1000 / r.fontSize
			}
			nl++
			continue
		}
		if c == ' ' {
			sep = i
		}
		l += float64(cw[int(c)])
		if l > wmax {
			// Automatic line break
			if sep == -1 {
				if r.x > r.lMargin {
					// Move to next line
					r.x = r.lMargin
					r.y += h
					w = r.w - r.rMargin - r.x
					wmax = (w - 2*r.cMargin) * 1000 / r.fontSize
					i++
					nl++
					continue
				}
				if i == j {
					i++
				}
				if r.isCurrentUTF8 {
					r.CellFormat(w, h, string([]rune(s)[j:i]), "", 2, "", false, link, linkStr)
				} else {
					r.CellFormat(w, h, s[j:i], "", 2, "", false, link, linkStr)
				}
			} else {
				if r.isCurrentUTF8 {
					r.CellFormat(w, h, string([]rune(s)[j:sep]), "", 2, "", false, link, linkStr)
				} else {
					r.CellFormat(w, h, s[j:sep], "", 2, "", false, link, linkStr)
				}
				i = sep + 1
			}
			sep = -1
			j = i
			l = 0.0
			if nl == 1 {
				r.x = r.lMargin
				w = r.w - r.rMargin - r.x
				wmax = (w - 2*r.cMargin) * 1000 / r.fontSize
			}
			nl++
		} else {
			i++
		}
	}
	// Last chunk
	if i != j {
		if r.isCurrentUTF8 {
			r.CellFormat(l/1000*r.fontSize, h, string([]rune(s)[j:]), "", 0, "", false, link, linkStr)
		} else {
			r.CellFormat(l/1000*r.fontSize, h, s[j:], "", 0, "", false, link, linkStr)
		}
	}
}

// Write prints text from the current position. When the right margin is
// reached (or the \n character is met) a line break occurs and text continues
// from the left margin. Upon method exit, the current position is left just at
// the end of the text.
//
// It is possible to put a link on the text.
//
// h indicates the line height in the unit of measure specified in New().
func (r *Renderer) Write(h float64, txtStr string) {
	r.write(h, txtStr, 0, "")
}

// Writef is like Write but uses printf-style formatting. See the documentation
// for package fmt for more details on fmtStr and args.
func (r *Renderer) Writef(h float64, fmtStr string, args ...any) {
	r.write(h, fmt.Sprintf(fmtStr, args...), 0, "")
}

// WriteLinkString writes text that when clicked launches an external URL. See
// Write() for argument details.
func (r *Renderer) WriteLinkString(h float64, displayStr, targetStr string) {
	r.write(h, displayStr, 0, targetStr)
}

// WriteLinkID writes text that when clicked jumps to another location in the
// PDF. linkID is an identifier returned by AddLink(). See Write() for argument
// details.
func (r *Renderer) WriteLinkID(h float64, displayStr string, linkID int) {
	r.write(h, displayStr, linkID, "")
}

// WriteAligned is an implementation of Write that makes it possible to align
// text.
//
// width indicates the width of the box the text will be drawn in. This is in
// the unit of measure specified in New(). If it is set to 0, the bounding box
// of the page will be taken (pageWidth - leftMargin - rightMargin).
//
// lineHeight indicates the line height in the unit of measure specified in
// New().
//
// alignStr sees to horizontal alignment of the given textStr. The options are
// "L", "C" and "R" (Left, Center, Right). The default is "L".
func (r *Renderer) WriteAligned(width, lineHeight float64, textStr, alignStr string) {
	lMargin, _, rMargin, _ := r.GetMargins()

	pageWidth, _ := r.GetPageSize()
	if width == 0 {
		width = pageWidth - (lMargin + rMargin)
	}

	var lines []string

	if r.isCurrentUTF8 {
		lines = r.SplitText(textStr, width)
	} else {
		for _, line := range r.SplitLines([]byte(textStr), width) {
			lines = append(lines, string(line))
		}
	}

	for _, lineBt := range lines {
		lineStr := string(lineBt)
		lineWidth := r.GetStringWidth(lineStr)

		switch alignStr {
		case "C":
			r.SetLeftMargin(lMargin + ((width - lineWidth) / 2))
			r.Write(lineHeight, lineStr)
			r.SetLeftMargin(lMargin)
		case "R":
			r.SetLeftMargin(lMargin + (width - lineWidth) - 2.01*r.cMargin)
			r.Write(lineHeight, lineStr)
			r.SetLeftMargin(lMargin)
		default:
			r.SetRightMargin(pageWidth - lMargin - width)
			r.Write(lineHeight, lineStr)
			r.SetRightMargin(rMargin)
		}
	}
}

// Ln performs a line break. The current abscissa goes back to the left margin
// and the ordinate increases by the amount passed in parameter. A negative
// value of h indicates the height of the last printed cell.
//
// This method is demonstrated in the example for MultiCell.
func (r *Renderer) Ln(h float64) {
	r.x = r.lMargin
	if h < 0 {
		r.y += r.lasth
	} else {
		r.y += h
	}
}

// ImageTypeFromMime returns the image type used in various image-related
// functions (for example, Image()) that is associated with the specified MIME
// type. For example, "jpg" is returned if mimeStr is "image/jpeg". An error is
// set if the specified MIME type is not supported.
func (r *Renderer) ImageTypeFromMime(mimeStr string) (tp string) {
	switch mimeStr {
	case "image/png":
		tp = "png"
	case "image/jpg":
		tp = "jpg"
	case "image/jpeg":
		tp = "jpg"
	case "image/gif":
		tp = "gif"
	default:
		r.SetErrorf("unsupported image type: %s", mimeStr)
	}
	return tp
}

func (r *Renderer) imageOut(info *ImageInfo, x, y, w, h float64, allowNegativeX, flow bool, link int, linkStr string) {
	// Automatic width and height calculation if needed
	if w == 0 && h == 0 {
		// Put image at 96 dpi
		w = -96
		h = -96
	}
	if w == -1 {
		// Set image width to whatever value for dpi we read
		// from the image or that was set manually
		w = -info.dpi
	}
	if h == -1 {
		// Set image height to whatever value for dpi we read
		// from the image or that was set manually
		h = -info.dpi
	}
	if w < 0 {
		w = -info.w * 72.0 / w / r.k
	}
	if h < 0 {
		h = -info.h * 72.0 / h / r.k
	}
	if w == 0 {
		w = h * info.w / info.h
	}
	if h == 0 {
		h = w * info.h / info.w
	}
	// Flowing mode
	if flow {
		if r.y+h > r.pageBreakTrigger && !r.inHeader && !r.inFooter && r.acceptPageBreak() {
			// Automatic page break
			x2 := r.x
			r.AddPageFormat(r.curOrientation, r.curPageSize)
			if r.err != nil {
				return
			}
			r.x = x2
		}
		y = r.y
		r.y += h
	}
	if !allowNegativeX {
		if x < 0 {
			x = r.x
		}
	}
	// dbg("h %.2f", h)
	// q 85.04 0 0 NaN 28.35 NaN cm /I2 Do Q
	// r.outf("q %.5f 0 0 %.5f %.5f %.5f cm /I%s Do Q", w*r.k, h*r.k, x*r.k, (r.h-(y+h))*r.k, info.i)
	const prec = 5
	r.put("q ")
	r.putF64(w*r.k, prec)
	r.put(" 0 0 ")
	r.putF64(h*r.k, prec)
	r.put(" ")
	r.putF64(x*r.k, prec)
	r.put(" ")
	r.putF64((r.h-(y+h))*r.k, prec)
	r.put(" cm /I" + info.i + " Do Q\n")
	if link > 0 || len(linkStr) > 0 {
		r.newLink(x, y, w, h, link, linkStr)
	}
}

// Image puts a JPEG, PNG or GIF image in the current page.
//
// Deprecated in favor of ImageOptions -- see that function for
// details on the behavior of arguments
func (r *Renderer) Image(imageNameStr string, x, y, w, h float64, flow bool, tp string, link int, linkStr string) {
	options := ImageOptions{
		ReadDpi:   false,
		ImageType: tp,
	}
	r.ImageOptions(imageNameStr, x, y, w, h, flow, options, link, linkStr)
}

// ImageOptions puts a JPEG, PNG or GIF image in the current page. The size it
// will take on the page can be specified in different ways. If both w and h
// are 0, the image is rendered at 96 dpi. If either w or h is zero, it will be
// calculated from the other dimension so that the aspect ratio is maintained.
// If w and/or h are -1, the dpi for that dimension will be read from the
// ImageInfo object. PNG files can contain dpi information, and if present,
// this information will be populated in the ImageInfo object and used in
// Width, Height, and Extent calculations. Otherwise, the SetDpi function can
// be used to change the dpi from the default of 72.
//
// If w and h are any other negative value, their absolute values
// indicate their dpi extents.
//
// Supported JPEG formats are 24 bit, 32 bit and gray scale. Supported PNG
// formats are 24 bit, indexed color, and 8 bit indexed gray scale. If a GIF
// image is animated, only the first frame is rendered. Transparency is
// supported. It is possible to put a link on the image.
//
// imageNameStr may be the name of an image as registered with a call to either
// RegisterImageReader() or RegisterImage(). In the first case, the image is
// loaded using an io.Reader. This is generally useful when the image is
// obtained from some other means than as a disk-based file. In the second
// case, the image is loaded as a file. Alternatively, imageNameStr may
// directly specify a sufficiently qualified filename.
//
// However the image is loaded, if it is used more than once only one copy is
// embedded in the file.
//
// If x is negative, the current abscissa is used.
//
// If flow is true, the current y value is advanced after placing the image and
// a page break may be made if necessary.
//
// If link refers to an internal page anchor (that is, it is non-zero; see
// AddLink()), the image will be a clickable internal link. Otherwise, if
// linkStr specifies a URL, the image will be a clickable external link.
func (r *Renderer) ImageOptions(imageNameStr string, x, y, w, h float64, flow bool, options ImageOptions, link int, linkStr string) {
	if r.err != nil {
		return
	}
	info := r.RegisterImageOptions(imageNameStr, options)
	if r.err != nil {
		return
	}
	r.imageOut(info, x, y, w, h, options.AllowNegativePosition, flow, link, linkStr)
}

// RegisterImageReader registers an image, reading it from Reader r, adding it
// to the PDF file but not adding it to the page.
//
// This function is now deprecated in favor of RegisterImageOptionsReader
func (r *Renderer) RegisterImageReader(imgName, tp string, rd io.Reader) (info *ImageInfo) {
	options := ImageOptions{
		ReadDpi:   false,
		ImageType: tp,
	}
	return r.RegisterImageOptionsReader(imgName, options, rd)
}

// ImageOptions provides a place to hang any options we want to use while
// parsing an image.
//
// ImageType's possible values are (case insensitive):
// "JPG", "JPEG", "PNG" and "GIF". If empty, the type is inferred from
// the file extension.
//
// ReadDpi defines whether to attempt to automatically read the image
// dpi information from the image file. Normally, this should be set
// to true (understanding that not all images will have this info
// available). However, for backwards compatibility with previous
// versions of the API, it defaults to false.
//
// AllowNegativePosition can be set to true in order to prevent the default
// coercion of negative x values to the current x position.
type ImageOptions struct {
	ImageType             string
	ReadDpi               bool
	AllowNegativePosition bool
}

// RegisterImageOptionsReader registers an image, reading it from Reader r, adding it
// to the PDF file but not adding it to the page. Use Image() with the same
// name to add the image to the page. Note that tp should be specified in this
// case.
//
// See Image() for restrictions on the image and the options parameters.
func (r *Renderer) RegisterImageOptionsReader(imgName string, options ImageOptions, rd io.Reader) (info *ImageInfo) {
	// Thanks, Ivan Daniluk, for generalizing this code to use the Reader interface.
	if r.err != nil {
		return info
	}
	info, ok := r.images[imgName]
	if ok {
		return info
	}

	// First use of this image, get info
	if options.ImageType == "" {
		r.err = errors.New("image type should be specified if reading from custom reader")
		return info
	}
	options.ImageType = strings.ToLower(options.ImageType)
	if options.ImageType == "jpeg" {
		options.ImageType = "jpg"
	}
	switch options.ImageType {
	case "jpg":
		info = r.parsejpg(rd)
	case "png":
		info = r.parsepng(rd, options.ReadDpi)
	case "gif":
		info = r.parsegif(rd)
	default:
		r.err = fmt.Errorf("unsupported image type: %s", options.ImageType)
	}
	if r.err != nil {
		return info
	}

	if info.i, r.err = generateImageID(info); r.err != nil {
		return info
	}
	r.images[imgName] = info

	return info
}

// RegisterImage registers an image, adding it to the PDF file but not adding
// it to the page. Use Image() with the same filename to add the image to the
// page. Note that Image() calls this function, so this function is only
// necessary if you need information about the image before placing it.
//
// This function is now deprecated in favor of RegisterImageOptions.
// See Image() for restrictions on the image and the "tp" parameters.
func (r *Renderer) RegisterImage(fileStr, tp string) (info *ImageInfo) {
	options := ImageOptions{
		ReadDpi:   false,
		ImageType: tp,
	}
	return r.RegisterImageOptions(fileStr, options)
}

// RegisterImageOptions registers an image, adding it to the PDF file but not
// adding it to the page. Use Image() with the same filename to add the image
// to the page. Note that Image() calls this function, so this function is only
// necessary if you need information about the image before placing it. See
// Image() for restrictions on the image and the "tp" parameters.
func (r *Renderer) RegisterImageOptions(fileStr string, options ImageOptions) (info *ImageInfo) {
	info, ok := r.images[fileStr]
	if ok {
		return info
	}

	file, err := os.Open(fileStr)
	if err != nil {
		r.err = err
		return info
	}
	defer file.Close()

	// First use of this image, get info
	if options.ImageType == "" {
		pos := strings.LastIndex(fileStr, ".")
		if pos < 0 {
			r.err = fmt.Errorf("image file has no extension and no type was specified: %s", fileStr)
			return info
		}
		options.ImageType = fileStr[pos+1:]
	}

	return r.RegisterImageOptionsReader(fileStr, options, file)
}

// GetImageInfo returns information about the registered image specified by
// imageStr. If the image has not been registered, nil is returned. The
// internal error is not modified by this method.
func (r *Renderer) GetImageInfo(imageStr string) (info *ImageInfo) {
	return r.images[imageStr]
}

// GetConversionRatio returns the conversion ratio based on the unit given when
// creating the PDF.
func (r *Renderer) GetConversionRatio() float64 {
	return r.k
}

// GetXY returns the abscissa and ordinate of the current position.
//
// Note: the value returned for the abscissa will be affected by the current
// cell margin. To account for this, you may need to either add the value
// returned by GetCellMargin() to it or call SetCellMargin(0) to remove the
// cell margin.
func (r *Renderer) GetXY() (float64, float64) {
	return r.x, r.y
}

// GetX returns the abscissa of the current position.
//
// Note: the value returned will be affected by the current cell margin. To
// account for this, you may need to either add the value returned by
// GetCellMargin() to it or call SetCellMargin(0) to remove the cell margin.
func (r *Renderer) GetX() float64 {
	return r.x
}

// SetX defines the abscissa of the current position. If the passed value is
// negative, it is relative to the right of the page.
func (r *Renderer) SetX(x float64) {
	if x >= 0 {
		r.x = x
	} else {
		r.x = r.w + x
	}
}

// GetY returns the ordinate of the current position.
func (r *Renderer) GetY() float64 {
	return r.y
}

// SetY moves the current abscissa back to the left margin and sets the
// ordinate. If the passed value is negative, it is relative to the bottom of
// the page.
func (r *Renderer) SetY(y float64) {
	// dbg("SetY x %.2f, lMargin %.2f", r.x, r.lMargin)
	r.x = r.lMargin
	if y >= 0 {
		r.y = y
	} else {
		r.y = r.h + y
	}
}

// SetHomeXY is a convenience method that sets the current position to the left
// and top margins.
func (r *Renderer) SetHomeXY() {
	r.SetY(r.tMargin)
	r.SetX(r.lMargin)
}

// SetXY defines the abscissa and ordinate of the current position. If the
// passed values are negative, they are relative respectively to the right and
// bottom of the page.
func (r *Renderer) SetXY(x, y float64) {
	r.SetY(y)
	r.SetX(x)
}

// SetProtection applies certain constraints on the finished PDF document.
//
// actionFlag is a bitflag that controls various document operations.
// CnProtectPrint allows the document to be printed. CnProtectModify allows a
// document to be modified by a PDF editor. CnProtectCopy allows text and
// images to be copied into the system clipboard. CnProtectAnnotForms allows
// annotations and forms to be added by a PDF editor. These values can be
// combined by or-ing them together, for example,
// CnProtectCopy|CnProtectModify. This flag is advisory; not all PDF readers
// implement the constraints that this argument attempts to control.
//
// userPassStr specifies the password that will need to be provided to view the
// contents of the PDF. The permissions specified by actionFlag will apply.
//
// ownerPassStr specifies the password that will need to be provided to gain
// full access to the document regardless of the actionFlag value. An empty
// string for this argument will be replaced with a random value, effectively
// prohibiting full access to the document.
func (r *Renderer) SetProtection(actionFlag byte, userPassStr, ownerPassStr string) {
	if r.err != nil {
		return
	}
	r.protect.setProtection(actionFlag, userPassStr, ownerPassStr)
}

// OutputAndClose sends the PDF document to the writer specified by w. This
// method will close both f and w, even if an error is detected and no document
// is produced.
func (r *Renderer) OutputAndClose(w io.WriteCloser) error {
	_ = r.Output(w)
	err := w.Close()
	if err != nil {
		return errs.Errorf("could not close writer: %w", err)
	}
	return r.err
}

// OutputFileAndClose creates or truncates the file specified by fileStr and
// writes the PDF document to it. This method will close f and the newly
// written file, even if an error is detected and no document is produced.
//
// Most examples demonstrate the use of this method.
func (r *Renderer) OutputFileAndClose(fileStr string) error {
	if r.err != nil {
		return r.err
	}

	pdfFile, err := os.Create(fileStr)
	if err != nil {
		r.err = err
		return r.err
	}

	_ = r.Output(pdfFile)

	err = pdfFile.Close()
	if err != nil {
		return errs.Errorf("could not close output file: %w", err)
	}

	return r.err
}

// Output sends the PDF document to the writer specified by w. No output will
// take place if an error has occurred in the document generation process. w
// remains open after this function returns. After returning, f is in a closed
// state and its methods should not be called.
func (r *Renderer) Output(w io.Writer) error {
	if r.err != nil {
		return r.err
	}
	// dbg("Output")
	if r.state < 3 {
		r.Close()
	}
	_, err := r.buffer.WriteTo(w)
	if err != nil {
		r.err = err
	}
	return r.err
}

func (r *Renderer) standardPageSize(pageSize PageSize) Size {
	pt, ok := pageSize.Size()
	if !ok {
		r.err = fmt.Errorf("unknown page size %s", pageSize)
		return Size{}
	}
	return Size{pt.Wd / r.k, pt.Ht / r.k}
}

func (r *Renderer) beginpage(orientation Orientation, size Size) {
	if r.err != nil {
		return
	}
	r.page++
	// add the default page boxes, if any exist, to the page
	r.pageBoxes[r.page] = make(map[string]PageBox)
	maps.Copy(r.pageBoxes[r.page], r.defPageBoxes)
	r.pages = append(r.pages, bytes.NewBufferString(""))
	r.pageLinks = append(r.pageLinks, make([]pageLink, 0))
	r.pageAttachments = append(r.pageAttachments, []annotationAttach{})
	r.state = 2
	r.x = r.lMargin
	r.y = r.tMargin
	r.fontFamily = ""
	// Check page size and orientation
	orientation = cmp.Or(orientation, r.defOrientation)
	if !orientation.Valid() {
		r.err = fmt.Errorf("incorrect orientation: %s", orientation)
		return
	}
	if orientation != r.curOrientation || size.Wd != r.curPageSize.Wd || size.Ht != r.curPageSize.Ht {
		r.w, r.h = orientation.pageSize(size)
		r.wPt = r.w * r.k
		r.hPt = r.h * r.k
		r.pageBreakTrigger = r.h - r.bMargin
		r.curOrientation = orientation
		r.curPageSize = size
	}
	if orientation != r.defOrientation || size.Wd != r.defPageSize.Wd || size.Ht != r.defPageSize.Ht {
		r.pageSizes[r.page] = Size{r.wPt, r.hPt}
	}
}

func (r *Renderer) endpage() {
	r.EndLayer()
	r.state = 1
}

// Load a font definition file from the given Reader
func (r *Renderer) loadfont(rd io.Reader) (def fontDef) {
	if r.err != nil {
		return def
	}
	// dbg("Loading font [%s]", fontStr)
	var buf bytes.Buffer
	_, err := buf.ReadFrom(rd)
	if err != nil {
		r.err = err
		return def
	}
	err = json.Unmarshal(buf.Bytes(), &def)
	if err != nil {
		r.err = err
		return def
	}

	if def.i, err = generateFontID(def); err != nil {
		r.err = err
	}
	// dump(def)
	return def
}

// Escape special characters in strings
func (r *Renderer) escape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}

// textstring formats a text string
func (r *Renderer) textstring(s string) string {
	if r.protect.encrypted {
		b := []byte(s)
		r.protect.rc4(uint32(r.n), &b)
		s = string(b)
	}
	return "(" + r.escape(s) + ")"
}

func blankCount(str string) (count int) {
	l := len(str)
	for j := range l {
		if byte(' ') == str[j] {
			count++
		}
	}
	return count
}

// GetUnderlineThickness returns the current text underline thickness multiplier.
func (r *Renderer) GetUnderlineThickness() float64 {
	return r.userUnderlineThickness
}

// SetUnderlineThickness accepts a multiplier for adjusting the text underline
// thickness, defaulting to 1. See SetUnderlineThickness example.
func (r *Renderer) SetUnderlineThickness(thickness float64) {
	r.userUnderlineThickness = thickness
}

// Underline text
func (r *Renderer) dounderline(x, y float64, txt string) string {
	up := float64(r.currentFont.Up)
	ut := float64(r.currentFont.Ut) * r.userUnderlineThickness
	w := r.GetStringWidth(txt) + r.ws*float64(blankCount(txt))
	return fmt.Sprintf("%.2f %.2f %.2f %.2f re f", x*r.k,
		(r.h-(y-up/1000*r.fontSize))*r.k, w*r.k, -ut/1000*r.fontSizePt)
}

func (r *Renderer) dostrikeout(x, y float64, txt string) string {
	up := float64(r.currentFont.Up)
	ut := float64(r.currentFont.Ut)
	w := r.GetStringWidth(txt) + r.ws*float64(blankCount(txt))
	return fmt.Sprintf("%.2f %.2f %.2f %.2f re f", x*r.k,
		(r.h-(y+4*up/1000*r.fontSize))*r.k, w*r.k, -ut/1000*r.fontSizePt)
}

func (r *Renderer) newImageInfo() *ImageInfo {
	// default dpi to 72 unless told otherwise
	return &ImageInfo{scale: r.k, dpi: 72}
}

// parsejpg extracts info from io.Reader with JPEG data
// Thank you, Bruno Michel, for providing this code.
func (r *Renderer) parsejpg(rd io.Reader) (info *ImageInfo) {
	info = r.newImageInfo()
	var (
		data bytes.Buffer
		err  error
	)
	_, err = data.ReadFrom(rd)
	if err != nil {
		r.err = err
		return info
	}
	info.data = data.Bytes()

	config, err := jpeg.DecodeConfig(bytes.NewReader(info.data))
	if err != nil {
		r.err = err
		return info
	}
	info.w = float64(config.Width)
	info.h = float64(config.Height)
	info.f = "DCTDecode"
	info.bpc = 8
	switch config.ColorModel {
	case color.GrayModel:
		info.cs = "DeviceGray"
	case color.YCbCrModel:
		info.cs = "DeviceRGB"
	case color.CMYKModel:
		info.cs = "DeviceCMYK"
	default:
		r.err = fmt.Errorf("image JPEG buffer has unsupported color space (%v)", config.ColorModel)
		return info
	}
	return info
}

// parsepng extracts info from a PNG data
func (r *Renderer) parsepng(rd io.Reader, readdpi bool) (info *ImageInfo) {
	data, err := io.ReadAll(rd)
	if err != nil {
		r.err = err
		return info
	}
	return r.parsepngstream(bytes.NewBuffer(data), readdpi)
}

// parsegif extracts info from a GIF data (via PNG conversion)
func (r *Renderer) parsegif(rd io.Reader) (info *ImageInfo) {
	data, err := io.ReadAll(rd)
	if err != nil {
		r.err = err
		return info
	}
	var img image.Image
	img, err = gif.Decode(bytes.NewReader(data))
	if err != nil {
		r.err = err
		return info
	}
	pngBuf := new(bytes.Buffer)
	err = png.Encode(pngBuf, img)
	if err != nil {
		r.err = err
		return info
	}
	return r.parsepngstream(bytes.NewBuffer(pngBuf.Bytes()), false)
}

// newobj begins a new object
func (r *Renderer) newobj() {
	// dbg("newobj")
	r.n++
	for j := len(r.offsets); j <= r.n; j++ {
		r.offsets = append(r.offsets, 0)
	}
	r.offsets[r.n] = r.buffer.Len()
	r.outf("%d 0 obj", r.n)
}

func (r *Renderer) putstream(b []byte) {
	// dbg("putstream")
	if r.protect.encrypted {
		r.protect.rc4(uint32(r.n), &b)
	}
	r.out("stream")
	r.out(string(b))
	r.out("endstream")
}

// out; Add a line to the document
func (r *Renderer) out(s string) {
	if r.state == 2 {
		r.pages[r.page].WriteString(s)
		r.pages[r.page].WriteString("\n")
	} else {
		r.buffer.WriteString(s)
		r.buffer.WriteString("\n")
	}
}

func (r *Renderer) put(s string) {
	if r.state == 2 {
		r.pages[r.page].WriteString(s)
	} else {
		r.buffer.WriteString(s)
	}
}

// outbuf adds a buffered line to the document. Unlike the bytes.Buffer write
// methods, ReadFrom can fail (RawWriteBuf accepts arbitrary readers); such an
// error is recorded in the renderer's error state.
func (r *Renderer) outbuf(rd io.Reader) {
	if r.state == 2 {
		if _, err := r.pages[r.page].ReadFrom(rd); err != nil {
			r.SetError(err)
		}
		r.pages[r.page].WriteString("\n")
	} else {
		if _, err := r.buffer.ReadFrom(rd); err != nil {
			r.SetError(err)
		}
		r.buffer.WriteString("\n")
	}
}

// RawWriteStr writes a string directly to the PDF generation buffer. This is a
// low-level function that is not required for normal PDF construction. An
// understanding of the PDF specification is needed to use this method
// correctly.
func (r *Renderer) RawWriteStr(str string) {
	r.out(str)
}

// RawWriteBuf writes the contents of the specified buffer directly to the PDF
// generation buffer. This is a low-level function that is not required for
// normal PDF construction. An understanding of the PDF specification is needed
// to use this method correctly.
func (r *Renderer) RawWriteBuf(rd io.Reader) {
	r.outbuf(rd)
}

// outf adds a formatted line to the document
func (r *Renderer) outf(fmtStr string, args ...any) {
	r.out(fmt.Sprintf(fmtStr, args...))
}

func (r *Renderer) putF64(v float64, prec int) {
	r.put(fmtF64(v, prec))
}

func fmtF64(v float64, prec int) string {
	return strconv.FormatFloat(v, 'f', prec, 64)
}

func (r *Renderer) putInt(v int) {
	r.put(strconv.Itoa(v))
}

// SetDefaultCatalogSort sets the default value of the catalog sort flag that
// will be used when initializing a new Renderer instance. See SetCatalogSort() for
// more details.
func SetDefaultCatalogSort(flag bool) {
	gl.catalogSort = flag
}

// GetCatalogSort returns the document's internal catalog sort flag.
func (r *Renderer) GetCatalogSort() bool {
	return r.catalogSort
}

// SetCatalogSort sets a flag that will be used, if true, to consistently order
// the document's internal resource catalogs. This method is typically only
// used for test purposes to facilitate PDF comparison.
func (r *Renderer) SetCatalogSort(flag bool) {
	r.catalogSort = flag
}

// SetDefaultCreationDate sets the default value of the document creation date
// that will be used when initializing a new Renderer instance. See
// SetCreationDate() for more details.
func SetDefaultCreationDate(tm time.Time) {
	gl.creationDate = tm
}

// SetDefaultModificationDate sets the default value of the document modification date
// that will be used when initializing a new Renderer instance. See
// SetCreationDate() for more details.
func SetDefaultModificationDate(tm time.Time) {
	gl.modDate = tm
}

// GetCreationDate returns the document's internal CreationDate value.
func (r *Renderer) GetCreationDate() time.Time {
	return r.creationDate
}

// SetCreationDate fixes the document's internal CreationDate value. By
// default, the time when the document is generated is used for this value.
// This method is typically only used for testing purposes to facilitate PDF
// comparison. Specify a zero-value time to revert to the default behavior.
func (r *Renderer) SetCreationDate(tm time.Time) {
	r.creationDate = tm
}

// GetModificationDate returns the document's internal ModDate value.
func (r *Renderer) GetModificationDate() time.Time {
	return r.modDate
}

// SetModificationDate fixes the document's internal ModDate value.
// See `SetCreationDate` for more details.
func (r *Renderer) SetModificationDate(tm time.Time) {
	r.modDate = tm
}

// GetJavascript returns the Adobe JavaScript for the document.
//
// GetJavascript returns an empty string if no javascript was
// previously defined.
func (r *Renderer) GetJavascript() string {
	if r.javascript == nil {
		return ""
	}
	return *r.javascript
}

// SetJavascript adds Adobe JavaScript to the document.
func (r *Renderer) SetJavascript(script string) {
	r.javascript = &script
}

// RegisterAlias adds an (alias, replacement) pair to the document so we can
// replace all occurrences of that alias after writing but before the document
// is closed. Functions ExampleFpdf_RegisterAlias() and
// ExampleFpdf_RegisterAlias_utf8() in fpdf_test.go demonstrate this method.
func (r *Renderer) RegisterAlias(alias, replacement string) {
	// Note: map[string]string assignments embed literal escape ("\00") sequences
	// into utf16 key and value strings. Consequently, subsequent search/replace
	// operations will fail unexpectedly if utf8toutf16() conversions take place
	// here. Instead, conversions are deferred until the actual search/replace
	// operation takes place when the PDF output is generated.
	r.aliasMap[alias] = replacement
}

func (r *Renderer) replaceAliases() {
	// Replace longer aliases first (ties broken lexicographically) so that an
	// alias containing another alias as a substring is never corrupted by the
	// shorter one's replacement, and so the output does not depend on the map
	// iteration order.
	aliases := slices.SortedFunc(maps.Keys(r.aliasMap), func(a, b string) int {
		if d := len(b) - len(a); d != 0 {
			return d
		}
		return strings.Compare(a, b)
	})
	for mode := range 2 {
		for _, alias := range aliases {
			replacement := r.aliasMap[alias]
			if mode == 1 {
				alias = utf8toutf16(alias, false)
				replacement = utf8toutf16(replacement, false)
			}
			for n := 1; n <= r.page; n++ {
				s := r.pages[n].String()
				if strings.Contains(s, alias) {
					s = strings.ReplaceAll(s, alias, replacement)
					r.pages[n].Truncate(0)
					r.pages[n].WriteString(s)
				}
			}
		}
	}
}

func (r *Renderer) putpages() {
	var wPt, hPt float64
	var pageSize Size
	var ok bool
	nb := r.page
	if len(r.aliasNbPagesStr) > 0 {
		// Replace number of pages
		r.RegisterAlias(r.aliasNbPagesStr, fmt.Sprintf("%d", nb))
	}
	r.replaceAliases()
	if r.defOrientation == OrientationPortrait {
		wPt = r.defPageSize.Wd * r.k
		hPt = r.defPageSize.Ht * r.k
	} else {
		wPt = r.defPageSize.Ht * r.k
		hPt = r.defPageSize.Wd * r.k
	}
	pagesObjectNumbers := make([]int, nb+1) // 1-based
	for n := 1; n <= nb; n++ {
		// Page
		r.newobj()
		pagesObjectNumbers[n] = r.n // save for /Kids
		r.out("<</Type /Page")
		r.out("/Parent 1 0 R")
		pageSize, ok = r.pageSizes[n]
		if ok {
			r.outf("/MediaBox [0 0 %.2f %.2f]", pageSize.Wd, pageSize.Ht)
		}
		// Sorted so the box order does not depend on map iteration order.
		for _, t := range slices.Sorted(maps.Keys(r.pageBoxes[n])) {
			pb := r.pageBoxes[n][t]
			r.outf("/%s [%.2f %.2f %.2f %.2f]", t, pb.X, pb.Y, pb.Wd, pb.Ht)
		}
		r.out("/Resources 2 0 R")
		// Links
		if len(r.pageLinks[n])+len(r.pageAttachments[n]) > 0 {
			var annots bytes.Buffer
			fmt.Fprintf(&annots, "/Annots [")
			for _, pl := range r.pageLinks[n] {
				fmt.Fprintf(&annots, "<</Type /Annot /Subtype /Link /Rect [%.2f %.2f %.2f %.2f] /Border [0 0 0] ",
					pl.x, pl.y, pl.x+pl.wd, pl.y-pl.ht)
				if pl.link == 0 {
					fmt.Fprintf(&annots, "/A <</S /URI /URI %s>>>>", r.textstring(pl.linkStr))
				} else {
					l := r.links[pl.link]
					var sz Size
					var h float64
					sz, ok = r.pageSizes[l.page]
					if ok {
						h = sz.Ht
					} else {
						h = hPt
					}
					// dbg("h [%.2f], l.y [%.2f] r.k [%.2f]\n", h, l.y, r.k)
					fmt.Fprintf(&annots, "/Dest [%d 0 R /XYZ 0 %.2f null]>>", 1+2*l.page, h-l.y*r.k)
				}
			}
			r.putAttachmentAnnotationLinks(&annots, n)
			fmt.Fprintf(&annots, "]")
			r.out(annots.String())
		}
		if r.pdfVersion > pdfVers1_3 {
			r.out("/Group <</Type /Group /S /Transparency /CS /DeviceRGB>>")
		}
		r.outf("/Contents %d 0 R>>", r.n+1)
		r.out("endobj")
		// Page content
		r.newobj()
		if r.compress {
			mem := xmem.compress(r.pages[n].Bytes())
			data := mem.Bytes()
			r.outf("<</Filter /FlateDecode /Length %d>>", len(data))
			r.putstream(data)
			xmem.release(mem)
		} else {
			r.outf("<</Length %d>>", r.pages[n].Len())
			r.putstream(r.pages[n].Bytes())
		}
		r.out("endobj")
	}
	// Pages root
	r.offsets[1] = r.buffer.Len()
	r.out("1 0 obj")
	r.out("<</Type /Pages")
	var kids bytes.Buffer
	fmt.Fprintf(&kids, "/Kids [")
	for i := 1; i <= nb; i++ {
		fmt.Fprintf(&kids, "%d 0 R ", pagesObjectNumbers[i])
	}
	fmt.Fprintf(&kids, "]")
	r.out(kids.String())
	r.outf("/Count %d", nb)
	r.outf("/MediaBox [0 0 %.2f %.2f]", wPt, hPt)
	r.out(">>")
	r.out("endobj")
}

func (r *Renderer) putfonts() {
	if r.err != nil {
		return
	}
	nf := r.n
	for _, diff := range r.diffs {
		// Encodings
		r.newobj()
		r.outf("<</Type /Encoding /BaseEncoding /WinAnsiEncoding /Differences [%s]>>", diff)
		r.out("endobj")
	}
	{
		var fileList []string
		var info fontFile
		var file string
		for file = range r.fontFiles {
			fileList = append(fileList, file)
		}
		if r.catalogSort {
			sort.SliceStable(fileList, func(i, j int) bool { return fileList[i] < fileList[j] })
		}
		for _, file = range fileList {
			info = r.fontFiles[file]
			if info.fontType != "UTF8" {
				r.newobj()
				info.n = r.n
				r.fontFiles[file] = info

				var font []byte

				if info.embedded {
					font = info.content
				} else {
					var err error
					font, err = r.loadFontFile(file)
					if err != nil {
						r.err = err
						return
					}
				}
				compressed := file[len(file)-2:] == ".z"
				if !compressed && info.length2 > 0 {
					buf := font[6:info.length1]
					buf = append(buf, font[6+info.length1+6:info.length2]...)
					font = buf
				}
				r.outf("<</Length %d", len(font))
				if compressed {
					r.out("/Filter /FlateDecode")
				}
				r.outf("/Length1 %d", info.length1)
				if info.length2 > 0 {
					r.outf("/Length2 %d /Length3 0", info.length2)
				}
				r.out(">>")
				r.putstream(font)
				r.out("endobj")
			}
		}
	}
	{
		keyList := make([]string, 0, len(r.fonts))
		for key := range r.fonts {
			keyList = append(keyList, key)
		}
		if r.catalogSort {
			sort.SliceStable(keyList, func(i, j int) bool { return keyList[i] < keyList[j] })
		}
		for _, key := range keyList {
			font := r.fonts[key]
			// Font objects
			font.N = r.n + 1
			r.fonts[key] = font
			tp := font.Tp
			name := font.Name
			switch tp {
			case "Core":
				// Core font
				r.newobj()
				r.out("<</Type /Font")
				r.outf("/BaseFont /%s", name)
				r.out("/Subtype /Type1")
				if name != "Symbol" && name != "ZapfDingbats" {
					r.out("/Encoding /WinAnsiEncoding")
				}
				r.out(">>")
				r.out("endobj")

			case "Type1":
				fallthrough

			case "TrueType":
				// Additional Type1 or TrueType/OpenType font
				r.newobj()
				r.out("<</Type /Font")
				r.outf("/BaseFont /%s", name)
				r.outf("/Subtype /%s", tp)
				r.out("/FirstChar 32 /LastChar 255")
				r.outf("/Widths %d 0 R", r.n+1)
				r.outf("/FontDescriptor %d 0 R", r.n+2)
				if font.DiffN > 0 {
					r.outf("/Encoding %d 0 R", nf+font.DiffN)
				} else {
					r.out("/Encoding /WinAnsiEncoding")
				}
				r.out(">>")
				r.out("endobj")
				// Widths
				r.newobj()
				var s bytes.Buffer
				s.WriteString("[")
				for j := 32; j < 256; j++ {
					fmt.Fprintf(&s, "%d ", font.Cw[j])
				}
				s.WriteString("]")
				r.out(s.String())
				r.out("endobj")
				// Descriptor
				r.newobj()
				s.Truncate(0)
				fmt.Fprintf(&s, "<</Type /FontDescriptor /FontName /%s ", name)
				fmt.Fprintf(&s, "/Ascent %d ", font.Desc.Ascent)
				fmt.Fprintf(&s, "/Descent %d ", font.Desc.Descent)
				fmt.Fprintf(&s, "/CapHeight %d ", font.Desc.CapHeight)
				fmt.Fprintf(&s, "/Flags %d ", font.Desc.Flags)
				fmt.Fprintf(&s, "/FontBBox [%d %d %d %d] ", font.Desc.FontBBox.Xmin, font.Desc.FontBBox.Ymin,
					font.Desc.FontBBox.Xmax, font.Desc.FontBBox.Ymax)
				fmt.Fprintf(&s, "/ItalicAngle %d ", font.Desc.ItalicAngle)
				fmt.Fprintf(&s, "/StemV %d ", font.Desc.StemV)
				fmt.Fprintf(&s, "/MissingWidth %d ", font.Desc.MissingWidth)
				suffix := ""
				if tp != "Type1" {
					suffix = "2"
				}
				fmt.Fprintf(&s, "/FontFile%s %d 0 R>>", suffix, r.fontFiles[font.File].n)
				r.out(s.String())
				r.out("endobj")

			case "UTF8":
				fontName := "utf8" + font.Name
				usedRunes := font.usedRunes
				delete(usedRunes, 0)
				utf8FontStream, err := font.utf8File.GenerateCutFont(usedRunes)
				if err != nil {
					r.SetError(err)
					return
				}
				utf8FontSize := len(utf8FontStream)
				CodeSignDictionary := font.utf8File.CodeSymbolDictionary
				delete(CodeSignDictionary, 0)

				r.newobj()
				r.out(fmt.Sprintf("<</Type /Font\n/Subtype /Type0\n/BaseFont /%s\n/Encoding /Identity-H\n/DescendantFonts [%d 0 R]\n/ToUnicode %d 0 R>>\n"+"endobj", fontName, r.n+1, r.n+2))

				r.newobj()
				r.out("<</Type /Font\n/Subtype /CIDFontType2\n/BaseFont /" + fontName + "\n" +
					"/CIDSystemInfo " + strconv.Itoa(r.n+2) + " 0 R\n/FontDescriptor " + strconv.Itoa(r.n+3) + " 0 R")
				if font.Desc.MissingWidth != 0 {
					r.out("/DW " + strconv.Itoa(font.Desc.MissingWidth))
				}
				r.generateCIDFontMap(&font, font.utf8File.LastRune)
				r.out("/CIDToGIDMap " + strconv.Itoa(r.n+4) + " 0 R>>")
				r.out("endobj")

				r.newobj()
				r.out("<</Length " + strconv.Itoa(len(toUnicode)) + ">>")
				r.putstream([]byte(toUnicode))
				r.out("endobj")

				// CIDInfo
				r.newobj()
				r.out("<</Registry (Adobe)\n/Ordering (UCS)\n/Supplement 0>>")
				r.out("endobj")

				// Font descriptor
				r.newobj()
				var s bytes.Buffer
				fmt.Fprintf(&s, "<</Type /FontDescriptor /FontName /%s\n /Ascent %d", fontName, font.Desc.Ascent)
				fmt.Fprintf(&s, " /Descent %d", font.Desc.Descent)
				fmt.Fprintf(&s, " /CapHeight %d", font.Desc.CapHeight)
				v := font.Desc.Flags
				v = v | 4
				v = v &^ 32
				fmt.Fprintf(&s, " /Flags %d", v)
				fmt.Fprintf(&s, "/FontBBox [%d %d %d %d] ", font.Desc.FontBBox.Xmin, font.Desc.FontBBox.Ymin,
					font.Desc.FontBBox.Xmax, font.Desc.FontBBox.Ymax)
				fmt.Fprintf(&s, " /ItalicAngle %d", font.Desc.ItalicAngle)
				fmt.Fprintf(&s, " /StemV %d", font.Desc.StemV)
				fmt.Fprintf(&s, " /MissingWidth %d", font.Desc.MissingWidth)
				fmt.Fprintf(&s, "/FontFile2 %d 0 R", r.n+2)
				fmt.Fprintf(&s, ">>")
				r.out(s.String())
				r.out("endobj")

				// Embed CIDToGIDMap
				cidToGidMap := make([]byte, 256*256*2)

				for cc, glyph := range CodeSignDictionary {
					cidToGidMap[cc*2] = byte(glyph >> 8)
					cidToGidMap[cc*2+1] = byte(glyph & 0xFF)
				}

				mem := xmem.compress(cidToGidMap)
				cidToGidMap = mem.Bytes()
				r.newobj()
				r.out("<</Length " + strconv.Itoa(len(cidToGidMap)) + "/Filter /FlateDecode>>")
				r.putstream(cidToGidMap)
				r.out("endobj")
				xmem.release(mem)

				//Font file
				mem = xmem.compress(utf8FontStream)
				compressedFontStream := mem.Bytes()
				r.newobj()
				r.out("<</Length " + strconv.Itoa(len(compressedFontStream)))
				r.out("/Filter /FlateDecode")
				r.out("/Length1 " + strconv.Itoa(utf8FontSize))
				r.out(">>")
				r.putstream(compressedFontStream)
				r.out("endobj")
				xmem.release(mem)

			default:
				r.err = fmt.Errorf("unsupported font type: %s", tp)
				return
			}
		}
	}
}

func (r *Renderer) generateCIDFontMap(font *fontDef, LastRune int) {
	rangeID := 0
	cidArray := make(map[int]*cidWidthRange)
	cidArrayKeys := make([]int, 0)
	prevCid := -2
	prevWidth := -1
	interval := false
	startCid := 1
	cwLen := LastRune + 1

	// for each character
	for cid := startCid; cid < cwLen; cid++ {
		if font.Cw[cid] == 0x00 {
			continue
		}
		width := font.Cw[cid]
		if width == 65535 {
			width = 0
		}
		if numb, ok := font.usedRunes[cid]; cid > 255 && (!ok || numb == 0) {
			continue
		}

		if cid == prevCid+1 {
			if width == prevWidth {
				if width == cidArray[rangeID].firstWidth() {
					cidArray[rangeID].appendWidth(width)
				} else {
					cidArray[rangeID].pop()
					rangeID = prevCid
					cidArray[rangeID] = newCIDWidthRange()
					cidArrayKeys = append(cidArrayKeys, rangeID)
					cidArray[rangeID].appendWidth(prevWidth)
					cidArray[rangeID].appendWidth(width)
				}
				interval = true
				cidArray[rangeID].interval = true
			} else {
				if interval {
					// new range
					rangeID = cid
					cidArray[rangeID] = newCIDWidthRange()
					cidArrayKeys = append(cidArrayKeys, rangeID)
					cidArray[rangeID].appendWidth(width)
				} else {
					cidArray[rangeID].appendWidth(width)
				}
				interval = false
			}
		} else {
			rangeID = cid
			cidArray[rangeID] = newCIDWidthRange()
			cidArrayKeys = append(cidArrayKeys, rangeID)
			cidArray[rangeID].appendWidth(width)
			interval = false
		}
		prevCid = cid
		prevWidth = width

	}
	previousKey := -1
	nextKey := -1
	isInterval := false
	for g := 0; g < len(cidArrayKeys); {
		key := cidArrayKeys[g]
		ws := *cidArray[key]
		cws := ws.entryCount()
		if (key == nextKey) && (!isInterval) && (!ws.interval || cws < 4) {
			if cidArray[key].interval {
				cidArray[key].interval = false
			}
			cidArray[previousKey] = mergeCIDWidthRanges(cidArray[previousKey], cidArray[key])
			if i := slices.Index(cidArrayKeys, key); i >= 0 {
				cidArrayKeys = slices.Delete(cidArrayKeys, i, i+1)
			}
		} else {
			g++
			previousKey = key
		}
		nextKey = key + cws
		if ws.interval {
			if cws > 3 {
				isInterval = true
			} else {
				isInterval = false
			}
			cidArray[key].interval = false
			nextKey--
		} else {
			isInterval = false
		}
	}
	var w bytes.Buffer
	for _, k := range cidArrayKeys {
		ws := cidArray[k]
		if len(arrayCountValues(ws.widths)) == 1 {
			fmt.Fprintf(&w, " %d %d %d", k, k+len(ws.widths)-1, ws.firstWidth())
		} else {
			fmt.Fprintf(&w, " %d [ %s ]\n", k, implode(" ", ws.widths))
		}
	}
	r.out("/W [" + w.String() + " ]")
}

// cidWidthRange is an ordered list of glyph widths for one entry in a PDF CID
// /W array, optionally marked as an interval run.
type cidWidthRange struct {
	widths   []int
	interval bool
}

func newCIDWidthRange() *cidWidthRange {
	return &cidWidthRange{widths: make([]int, 0)}
}

func (r *cidWidthRange) appendWidth(w int) {
	r.widths = append(r.widths, w)
}

func (r *cidWidthRange) pop() {
	r.widths = r.widths[:len(r.widths)-1]
}

func (r *cidWidthRange) firstWidth() int {
	if len(r.widths) == 0 {
		return 0
	}
	return r.widths[0]
}

// entryCount matches the legacy PHP-array entry count: one slot per width plus
// one when the interval marker is present.
func (r cidWidthRange) entryCount() int {
	n := len(r.widths)
	if r.interval {
		n++
	}
	return n
}

func mergeCIDWidthRanges(a, b *cidWidthRange) *cidWidthRange {
	switch {
	case a == nil && b == nil:
		return newCIDWidthRange()
	case b == nil:
		return &cidWidthRange{
			widths:   slices.Clone(a.widths),
			interval: a.interval,
		}
	case a == nil:
		return &cidWidthRange{
			widths:   slices.Clone(b.widths),
			interval: b.interval,
		}
	default:
		merged := &cidWidthRange{
			widths:   slices.Clone(a.widths),
			interval: a.interval,
		}
		merged.widths = append(merged.widths, b.widths...)
		if !merged.interval && b.interval {
			merged.interval = true
		}
		return merged
	}
}

func implode(sep string, arr []int) string {
	var s bytes.Buffer
	for i := 0; i < len(arr)-1; i++ {
		fmt.Fprintf(&s, "%v", arr[i])
		fmt.Fprintf(&s, "%s", sep)
	}
	if len(arr) > 0 {
		fmt.Fprintf(&s, "%v", arr[len(arr)-1])
	}
	return s.String()
}

// arrayCountValues counts the occurrences of each item in the $mp array.
func arrayCountValues(mp []int) map[int]int {
	answer := make(map[int]int)
	for _, v := range mp {
		answer[v] = answer[v] + 1
	}
	return answer
}

func (r *Renderer) loadFontFile(name string) ([]byte, error) {
	if r.fontLoader != nil {
		reader, err := r.fontLoader.Open(name)
		if err == nil {
			data, err := io.ReadAll(reader)
			if closer, ok := reader.(io.Closer); ok {
				closer.Close()
			}
			return data, err
		}
	}
	return os.ReadFile(path.Join(r.fontpath, name))
}

func (r *Renderer) putimages() {
	keyList := make([]string, 0, len(r.images))
	for key := range r.images {
		keyList = append(keyList, key)
	}

	// Sort the keyList []string by the corresponding image's width, with the
	// content hash as tie-breaker so that equal-width images do not leak the
	// random map iteration order into the output.
	if r.catalogSort {
		sort.SliceStable(keyList, func(i, j int) bool {
			a, b := r.images[keyList[i]], r.images[keyList[j]]
			if a.w != b.w {
				return a.w < b.w
			}
			return a.i < b.i
		})
	}

	// Maintain a list of inserted image SHA-1 hashes, with their
	// corresponding object ID number.
	insertedImages := map[string]int{}

	for _, key := range keyList {
		image := r.images[key]

		// Check if this image has already been inserted using it's SHA-1 hash.
		insertedImageObjN, isFound := insertedImages[image.i]

		// If found, skip inserting the image as a new object, and
		// use the object ID from the insertedImages map.
		// If not, insert the image into the PDF and store the object ID.
		if isFound {
			image.n = insertedImageObjN
		} else {
			r.putimage(image)
			insertedImages[image.i] = image.n
		}
	}
}

func (r *Renderer) putimage(info *ImageInfo) {
	r.newobj()
	info.n = r.n
	r.out("<</Type /XObject")
	r.out("/Subtype /Image")
	r.outf("/Width %d", int(info.w))
	r.outf("/Height %d", int(info.h))
	if info.cs == "Indexed" {
		r.outf("/ColorSpace [/Indexed /DeviceRGB %d %d 0 R]", len(info.pal)/3-1, r.n+1)
	} else {
		r.outf("/ColorSpace /%s", info.cs)
		if info.cs == "DeviceCMYK" {
			r.out("/Decode [1 0 1 0 1 0 1 0]")
		}
	}
	r.outf("/BitsPerComponent %d", info.bpc)
	if len(info.f) > 0 {
		r.outf("/Filter /%s", info.f)
	}
	if len(info.dp) > 0 {
		r.outf("/DecodeParms <<%s>>", info.dp)
	}
	if len(info.trns) > 0 {
		var trns bytes.Buffer
		for _, v := range info.trns {
			fmt.Fprintf(&trns, "%d %d ", v, v)
		}
		r.outf("/Mask [%s]", trns.String())
	}
	if info.smask != nil {
		r.outf("/SMask %d 0 R", r.n+1)
	}
	r.outf("/Length %d>>", len(info.data))
	r.putstream(info.data)
	r.out("endobj")
	// 	Soft mask
	if len(info.smask) > 0 {
		smask := &ImageInfo{
			w:     info.w,
			h:     info.h,
			cs:    "DeviceGray",
			bpc:   8,
			f:     info.f,
			dp:    fmt.Sprintf("/Predictor 15 /Colors 1 /BitsPerComponent 8 /Columns %d", int(info.w)),
			data:  info.smask,
			scale: r.k,
		}
		r.putimage(smask)
	}
	// 	Palette
	if info.cs == "Indexed" {
		r.newobj()
		if r.compress {
			mem := xmem.compress(info.pal)
			pal := mem.Bytes()
			r.outf("<</Filter /FlateDecode /Length %d>>", len(pal))
			r.putstream(pal)
			xmem.release(mem)
		} else {
			r.outf("<</Length %d>>", len(info.pal))
			r.putstream(info.pal)
		}
		r.out("endobj")
	}
}

func (r *Renderer) putxobjectdict() {
	keyList := make([]string, 0, len(r.images))
	for key := range r.images {
		keyList = append(keyList, key)
	}
	if r.catalogSort {
		sort.SliceStable(keyList, func(i, j int) bool { return r.images[keyList[i]].i < r.images[keyList[j]].i })
	}
	for _, key := range keyList {
		image := r.images[key]
		r.outf("/I%s %d 0 R", image.i, image.n)
	}
}

func (r *Renderer) putresourcedict() {
	r.out("/ProcSet [/PDF /Text /ImageB /ImageC /ImageI]")
	r.out("/Font <<")
	{
		keyList := make([]string, 0, len(r.fonts))
		for key := range r.fonts {
			keyList = append(keyList, key)
		}
		if r.catalogSort {
			sort.SliceStable(keyList, func(i, j int) bool { return r.fonts[keyList[i]].i < r.fonts[keyList[j]].i })
		}
		for _, key := range keyList {
			font := r.fonts[key]
			r.outf("/F%s %d 0 R", font.i, font.N)
		}
	}
	r.out(">>")
	r.out("/XObject <<")
	r.putxobjectdict()
	r.out(">>")
	count := len(r.blendList)
	if count > 1 {
		r.out("/ExtGState <<")
		for j := 1; j < count; j++ {
			r.outf("/GS%d %d 0 R", j, r.blendList[j].objNum)
		}
		r.out(">>")
	}
	count = len(r.gradientList)
	if count > 1 {
		r.out("/Shading <<")
		for j := 1; j < count; j++ {
			r.outf("/Sh%d %d 0 R", j, r.gradientList[j].objNum)
		}
		r.out(">>")
	}
	// Layers
	r.layerPutResourceDict()
	r.spotColorPutResourceDict()
}

func (r *Renderer) putBlendModes() {
	count := len(r.blendList)
	for j := 1; j < count; j++ {
		bl := r.blendList[j]
		r.newobj()
		r.blendList[j].objNum = r.n
		r.outf("<</Type /ExtGState /ca %s /CA %s /BM /%s>>",
			bl.fillStr, bl.strokeStr, bl.modeStr)
		r.out("endobj")
	}
}

func (r *Renderer) putGradients() {
	count := len(r.gradientList)
	for j := 1; j < count; j++ {
		var f1 int
		gr := r.gradientList[j]
		if gr.tp == 2 || gr.tp == 3 {
			r.newobj()
			r.outf("<</FunctionType 2 /Domain [0.0 1.0] /C0 [%s] /C1 [%s] /N 1>>", gr.clr1Str, gr.clr2Str)
			r.out("endobj")
			f1 = r.n
		}
		r.newobj()
		r.outf("<</ShadingType %d /ColorSpace /DeviceRGB", gr.tp)
		switch gr.tp {
		case 2:
			r.outf("/Coords [%.5f %.5f %.5f %.5f] /Function %d 0 R /Extend [true true]>>",
				gr.x1, gr.y1, gr.x2, gr.y2, f1)
		case 3:
			r.outf("/Coords [%.5f %.5f 0 %.5f %.5f %.5f] /Function %d 0 R /Extend [true true]>>",
				gr.x1, gr.y1, gr.x2, gr.y2, gr.r, f1)
		}
		r.out("endobj")
		r.gradientList[j].objNum = r.n
	}
}

func (r *Renderer) putjavascript() {
	if r.javascript == nil {
		return
	}
	r.newobj()
	r.nJs = r.n
	r.out("<<")
	r.outf("/Names [(EmbeddedJS) %d 0 R]", r.n+1)
	r.out(">>")
	r.out("endobj")
	r.newobj()
	r.out("<<")
	r.out("/S /JavaScript")
	r.outf("/JS %s", r.textstring(*r.javascript))
	r.out(">>")
	r.out("endobj")
}

func (r *Renderer) putresources() {
	if r.err != nil {
		return
	}
	r.layerPutLayers()
	r.putBlendModes()
	r.putGradients()
	r.putSpotColors()
	r.putfonts()
	if r.err != nil {
		return
	}
	r.putimages()
	// 	Resource dictionary
	r.offsets[2] = r.buffer.Len()
	r.out("2 0 obj")
	r.out("<<")
	r.putresourcedict()
	r.out(">>")
	r.out("endobj")
	r.putjavascript()
	if r.protect.encrypted {
		r.newobj()
		r.protect.objNum = r.n
		r.out("<<")
		r.out("/Filter /Standard")
		r.out("/V 1")
		r.out("/R 2")
		r.outf("/O (%s)", r.escape(string(r.protect.oValue)))
		r.outf("/U (%s)", r.escape(string(r.protect.uValue)))
		r.outf("/P %d", r.protect.pValue)
		r.out(">>")
		r.out("endobj")
	}
}

// returns Now() if tm is zero
func timeOrNow(tm time.Time) time.Time {
	if tm.IsZero() {
		return time.Now()
	}
	return tm
}

func (r *Renderer) putinfo() {
	if len(r.producer) > 0 {
		r.outf("/Producer %s", r.textstring(r.producer))
	}
	if len(r.title) > 0 {
		r.outf("/Title %s", r.textstring(r.title))
	}
	if len(r.subject) > 0 {
		r.outf("/Subject %s", r.textstring(r.subject))
	}
	if len(r.author) > 0 {
		r.outf("/Author %s", r.textstring(r.author))
	}
	if len(r.keywords) > 0 {
		r.outf("/Keywords %s", r.textstring(r.keywords))
	}
	if len(r.creator) > 0 {
		r.outf("/Creator %s", r.textstring(r.creator))
	}
	// With XMP metadata present (PDF/A mode) the dates carry the timezone,
	// fully specifying the instants so they can be consistent with the XMP
	// dates as PDF/A requires; otherwise the legacy format is kept for byte
	// parity with the fpdf baseline.
	withTimezone := len(r.xmp) != 0
	creation := timeOrNow(r.creationDate)
	r.outf("/CreationDate %s", r.textstring(pdfDate(creation, withTimezone)))
	mod := timeOrNow(r.modDate)
	r.outf("/ModDate %s", r.textstring(pdfDate(mod, withTimezone)))
}

func (r *Renderer) putcatalog() {
	r.out("/Type /Catalog")
	r.out("/Pages 1 0 R")
	r.putOutputIntents()
	if r.lang != "" {
		r.outf("/Lang (%s)", r.lang)
	}
	switch r.zoomMode {
	case "fullpage":
		r.out("/OpenAction [3 0 R /Fit]")
	case "fullwidth":
		r.out("/OpenAction [3 0 R /FitH null]")
	case "real":
		r.out("/OpenAction [3 0 R /XYZ null null 1]")
	}
	// } 	else if !is_string($this->zoomMode))
	// 		$this->out('/OpenAction [3 0 R /XYZ null null '.fmt.Sprintf('%.2f',$this->zoomMode/100).']');
	switch r.layoutMode {
	case "single", "SinglePage":
		r.out("/PageLayout /SinglePage")
	case "continuous", "OneColumn":
		r.out("/PageLayout /OneColumn")
	case "two", "TwoColumnLeft":
		r.out("/PageLayout /TwoColumnLeft")
	case "TwoColumnRight":
		r.out("/PageLayout /TwoColumnRight")
	case "TwoPageLeft", "TwoPageRight":
		if r.pdfVersion < pdfVers1_5 {
			r.pdfVersion = pdfVers1_5
		}
		r.out("/PageLayout /" + r.layoutMode)
	}
	// Bookmarks
	if len(r.outlines) > 0 {
		r.outf("/Outlines %d 0 R", r.outlineRoot)
		r.out("/PageMode /UseOutlines")
	}
	// Layers
	r.layerPutCatalog()
	// XMP metadata
	if len(r.xmp) != 0 {
		r.outf("/Metadata %d 0 R", r.nXMP)
	}
	// Name dictionary :
	//	-> Javascript
	//	-> Embedded files
	r.out("/Names <<")
	// JavaScript
	if r.javascript != nil {
		r.outf("/JavaScript %d 0 R", r.nJs)
	}
	// Embedded files
	r.outf("/EmbeddedFiles %s", r.getEmbeddedFiles())
	r.out(">>")
	// Associated files (PDF/A-3)
	if af := r.getAssociatedFiles(); af != "" {
		r.outf("/AF %s", af)
	}
}

func (r *Renderer) putheader() {
	r.outf("%%PDF-%s", r.pdfVersion)
	r.out("%µ¶")
}

func (r *Renderer) puttrailer() {
	r.outf("/Size %d", r.n+1)
	r.outf("/Root %d 0 R", r.n)
	r.outf("/Info %d 0 R", r.n-1)
	switch {
	case r.protect.encrypted:
		r.outf("/Encrypt %d 0 R", r.protect.objNum)
		r.out("/ID [()()]")
	case len(r.xmp) != 0:
		// PDF/A requires a file identifier in the trailer. The hash of the
		// document bytes so far is deterministic for fixed dates.
		id := checksum(r.buffer.Bytes())
		r.outf("/ID [<%s> <%s>]", id, id)
	}
}

func (r *Renderer) putxmp() {
	if len(r.xmp) == 0 {
		return
	}
	r.newobj()
	r.nXMP = r.n
	r.outf("<< /Type /Metadata /Subtype /XML /Length %d >>", len(r.xmp))
	r.putstream(r.xmp)
	r.out("endobj")
}

func (r *Renderer) putbookmarks() {
	nb := len(r.outlines)
	if nb == 0 {
		return
	}
	lru := make(map[int]int)
	level := 0
	for i, o := range r.outlines {
		if o.level > 0 {
			parent := lru[o.level-1]
			r.outlines[i].parent = parent
			r.outlines[parent].last = i
			if o.level > level {
				r.outlines[parent].first = i
			}
		} else {
			r.outlines[i].parent = nb
		}
		if o.level <= level && i > 0 {
			prev := lru[o.level]
			r.outlines[prev].next = i
			r.outlines[i].prev = prev
		}
		lru[o.level] = i
		level = o.level
	}
	n := r.n + 1
	for _, o := range r.outlines {
		r.newobj()
		r.outf("<</Title %s", r.textstring(o.text))
		r.outf("/Parent %d 0 R", n+o.parent)
		if o.prev != -1 {
			r.outf("/Prev %d 0 R", n+o.prev)
		}
		if o.next != -1 {
			r.outf("/Next %d 0 R", n+o.next)
		}
		if o.first != -1 {
			r.outf("/First %d 0 R", n+o.first)
		}
		if o.last != -1 {
			r.outf("/Last %d 0 R", n+o.last)
		}
		r.outf("/Dest [%d 0 R /XYZ 0 %.2f null]", 1+2*o.p, (r.h-o.y)*r.k)
		r.out("/Count 0>>")
		r.out("endobj")
	}
	r.newobj()
	r.outlineRoot = r.n
	r.outf("<</Type /Outlines /First %d 0 R", n)
	r.outf("/Last %d 0 R>>", n+lru[0])
	r.out("endobj")
}

func (r *Renderer) putOutputIntents() {
	if len(r.outputIntents) <= 0 {
		return
	}

	r.out("/OutputIntents [")
	for index, oi := range r.outputIntents {
		infoSegment := ""
		if oi.Info != "" {
			infoSegment = fmt.Sprintf("/Info (%s) ", oi.Info)
		}
		r.outf(
			`<< /Type /OutputIntent /S /%s /OutputConditionIdentifier (%s) %s/DestOutputProfile %d 0 R >>`,
			oi.SubtypeIdent, oi.OutputConditionIdentifier, infoSegment, r.outputIntentStartN+index,
		)
	}
	r.out("]")
}

func (r *Renderer) putOutputIntentStreams() {
	if len(r.outputIntents) <= 0 {
		return
	}

	r.outputIntentStartN = r.n + 1
	for _, oi := range r.outputIntents {
		r.newobj()
		mem := xmem.compress(oi.ICCProfile)
		compressedICC := mem.Bytes()
		r.outf("<< /N 3 /Alternate /DeviceRGB /Length %d /Filter /FlateDecode >>", len(compressedICC))
		r.putstream(compressedICC)
		r.out("endobj")

		xmem.release(mem)
	}
}

func (r *Renderer) enddoc() {
	if r.err != nil {
		return
	}
	r.layerEndDoc()
	r.putheader()
	// Embedded files
	r.putAttachments()
	r.putAnnotationsAttachments()
	r.putpages()
	r.putresources()
	if r.err != nil {
		return
	}
	// Bookmarks
	r.putbookmarks()
	// Metadata
	r.putxmp()
	// 	Info
	r.newobj()
	r.out("<<")
	r.putinfo()
	r.out(">>")
	r.out("endobj")
	// Output intent color profile streams
	r.putOutputIntentStreams()
	// 	Catalog
	r.newobj()
	r.out("<<")
	r.putcatalog()
	r.out(">>")
	r.out("endobj")
	// Cross-ref
	o := r.buffer.Len()
	r.out("xref")
	r.outf("0 %d", r.n+1)
	r.out("0000000000 65535 f ")
	for j := 1; j <= r.n; j++ {
		r.outf("%010d 00000 n ", r.offsets[j])
	}
	// Trailer
	r.out("trailer")
	r.out("<<")
	r.puttrailer()
	r.out(">>")
	r.out("startxref")
	r.outf("%d", o)
	r.out("%%EOF")
	r.state = 3
}

// Path Drawing

// MoveTo moves the stylus to (x, y) without drawing the path from the
// previous point. Paths must start with a MoveTo to set the original
// stylus location or the result is undefined.
//
// Create a "path" by moving a virtual stylus around the page (with
// MoveTo, LineTo, CurveTo, CurveBezierCubicTo, ArcTo & ClosePath)
// then draw it or  fill it in (with DrawPath). The main advantage of
// using the path drawing routines rather than multiple Renderer.Line is
// that PDF creates nice line joins at the angles, rather than just
// overlaying the lines.
func (r *Renderer) MoveTo(x, y float64) {
	r.point(x, y)
	r.x, r.y = x, y
}

// LineTo creates a line from the current stylus location to (x, y), which
// becomes the new stylus location. Note that this only creates the line in
// the path; it does not actually draw the line on the page.
//
// The MoveTo() example demonstrates this method.
func (r *Renderer) LineTo(x, y float64) {
	// r.outf("%.2f %.2f l", x*r.k, (r.h-y)*r.k)
	const prec = 2
	r.putF64(x*r.k, prec)
	r.put(" ")

	r.putF64((r.h-y)*r.k, prec)
	r.put(" l\n")

	r.x, r.y = x, y
}

// CurveTo creates a single-segment quadratic Bézier curve. The curve starts at
// the current stylus location and ends at the point (x, y). The control point
// (cx, cy) specifies the curvature. At the start point, the curve is tangent
// to the straight line between the current stylus location and the control
// point. At the end point, the curve is tangent to the straight line between
// the end point and the control point.
//
// The MoveTo() example demonstrates this method.
func (r *Renderer) CurveTo(cx, cy, x, y float64) {
	// r.outf("%.5f %.5f %.5f %.5f v", cx*r.k, (r.h-cy)*r.k, x*r.k, (r.h-y)*r.k)
	const prec = 5
	r.putF64(cx*r.k, prec)
	r.put(" ")
	r.putF64((r.h-cy)*r.k, prec)
	r.put(" ")
	r.putF64(x*r.k, prec)
	r.put(" ")
	r.putF64((r.h-y)*r.k, prec)
	r.put(" v\n")
	r.x, r.y = x, y
}

// CurveBezierCubicTo creates a single-segment cubic Bézier curve. The curve
// starts at the current stylus location and ends at the point (x, y). The
// control points (cx0, cy0) and (cx1, cy1) specify the curvature. At the
// current stylus, the curve is tangent to the straight line between the
// current stylus location and the control point (cx0, cy0). At the end point,
// the curve is tangent to the straight line between the end point and the
// control point (cx1, cy1).
//
// The MoveTo() example demonstrates this method.
func (r *Renderer) CurveBezierCubicTo(cx0, cy0, cx1, cy1, x, y float64) {
	r.curve(cx0, cy0, cx1, cy1, x, y)
	r.x, r.y = x, y
}

// ClosePath creates a line from the current location to the last MoveTo point
// (if not the same) and mark the path as closed so the first and last lines
// join nicely.
//
// The MoveTo() example demonstrates this method.
func (r *Renderer) ClosePath() {
	r.outf("h")
}

// DrawPath actually draws the path on the page.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D".
// Path-painting operators as defined in the PDF specification are also
// allowed: "S" (Stroke the path), "s" (Close and stroke the path),
// "f" (fill the path, using the nonzero winding number), "f*"
// (Fill the path, using the even-odd rule), "B" (Fill and then stroke
// the path, using the nonzero winding number rule), "B*" (Fill and
// then stroke the path, using the even-odd rule), "b" (Close, fill,
// and then stroke the path, using the nonzero winding number rule) and
// "b*" (Close, fill, and then stroke the path, using the even-odd
// rule).
// Drawing uses the current draw color, line width, and cap style
// centered on the
// path. Filling uses the current fill color.
//
// The MoveTo() example demonstrates this method.
func (r *Renderer) DrawPath(styleStr string) {
	r.outf("%s", fillDrawOp(styleStr))
}

// ArcTo draws an elliptical arc centered at point (x, y). rx and ry specify its
// horizontal and vertical radii. If the start of the arc is not at
// the current position, a connecting line will be drawn.
//
// degRotate specifies the angle that the arc will be rotated. degStart and
// degEnd specify the starting and ending angle of the arc. All angles are
// specified in degrees and measured counter-clockwise from the 3 o'clock
// position.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the arc's
// path. Filling uses the current fill color.
//
// The MoveTo() example demonstrates this method.
func (r *Renderer) ArcTo(x, y, rx, ry, degRotate, degStart, degEnd float64) {
	r.arc(x, y, rx, ry, degRotate, degStart, degEnd, "", true)
}

func (r *Renderer) arc(x, y, rx, ry, degRotate, degStart, degEnd float64,
	styleStr string, path bool) {
	x *= r.k
	y = (r.h - y) * r.k
	rx *= r.k
	ry *= r.k
	segments := max(int(degEnd-degStart)/60, 2)
	angleStart := degStart * math.Pi / 180
	angleEnd := degEnd * math.Pi / 180
	angleTotal := angleEnd - angleStart
	dt := angleTotal / float64(segments)
	dtm := dt / 3
	if degRotate != 0 {
		a := -degRotate * math.Pi / 180
		sin, cos := math.Sincos(a)
		//	r.outf("q %.5f %.5f %.5f %.5f %.5f %.5f cm",
		//		math.Cos(a), -1*math.Sin(a),
		//		math.Sin(a), math.Cos(a), x, y)
		const prec = 5
		r.put("q ")
		r.putF64(cos, prec)
		r.put(" ")
		r.putF64(-1*sin, prec)
		r.put(" ")
		r.putF64(sin, prec)
		r.put(" ")
		r.putF64(cos, prec)
		r.put(" ")
		r.putF64(x, prec)
		r.put(" ")
		r.putF64(y, prec)
		r.put(" cm\n")

		x = 0
		y = 0
	}
	t := angleStart
	a0 := x + rx*math.Cos(t)
	b0 := y + ry*math.Sin(t)
	c0 := -rx * math.Sin(t)
	d0 := ry * math.Cos(t)
	sx := a0 / r.k // start point of arc
	sy := r.h - (b0 / r.k)
	if path {
		if r.x != sx || r.y != sy {
			// Draw connecting line to start point
			r.LineTo(sx, sy)
		}
	} else {
		r.point(sx, sy)
	}
	for j := 1; j <= segments; j++ {
		// Draw this bit of the total curve
		t = (float64(j) * dt) + angleStart
		a1 := x + rx*math.Cos(t)
		b1 := y + ry*math.Sin(t)
		c1 := -rx * math.Sin(t)
		d1 := ry * math.Cos(t)
		r.curve((a0+(c0*dtm))/r.k,
			r.h-((b0+(d0*dtm))/r.k),
			(a1-(c1*dtm))/r.k,
			r.h-((b1-(d1*dtm))/r.k),
			a1/r.k,
			r.h-(b1/r.k))
		a0 = a1
		b0 = b1
		c0 = c1
		d0 = d1
		if path {
			r.x = a1 / r.k
			r.y = r.h - (b1 / r.k)
		}
	}
	if !path {
		r.out(fillDrawOp(styleStr))
	}
	if degRotate != 0 {
		r.out("Q")
	}
}
