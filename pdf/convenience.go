package pdf

import "io"

// Renderer convenience helpers for the component layer: constructors with
// go-mx defaults, UTF-8 text translation, line height for flowing text, and
// automatic first-page creation.

// NewRenderer creates a Renderer for the given page orientation, measurement
// unit and page size, with a Helvetica 12pt default font already selected so
// text can be drawn without further setup.
func NewRenderer(orientation Orientation, unit Unit, size PageSize) *Renderer {
	r := newRenderer(orientation, unit, size, "", Size{0, 0})
	r.SetFont(DefaultFontFamily, string(StyleRegular), DefaultFontSize)
	return r
}

// NewRendererA4Portrait creates a Renderer for an A4 portrait document
// measured in millimeters.
func NewRendererA4Portrait() *Renderer {
	return NewRenderer(OrientationPortrait, UnitMillimeter, PageSizeA4)
}

// NewRendererA4Landscape creates a Renderer for an A4 landscape document
// measured in millimeters.
func NewRendererA4Landscape() *Renderer {
	return NewRenderer(OrientationLandscape, UnitMillimeter, PageSizeA4)
}

// NewRendererLetterPortrait creates a Renderer for a US Letter portrait
// document measured in inches.
func NewRendererLetterPortrait() *Renderer {
	return NewRenderer(OrientationPortrait, UnitInch, PageSizeLetter)
}

// Str applies the current font's UTF-8 translation to s. The text components
// call this automatically; use it when passing strings to the raw renderer
// methods.
func (r *Renderer) Str(s string) string {
	return r.tr(s)
}

// SetTranslator replaces the UTF-8 translator, e.g. after switching to a font
// with a different code page. Pass the result of
// UnicodeTranslatorFromDescriptor or the identity for UTF-8 fonts.
func (r *Renderer) SetTranslator(translate func(string) string) {
	r.translate = translate
}

// LoadUTF8FontBytes registers a UTF-8 TrueType font from in-memory bytes under
// the given family and style — the in-memory counterpart of the file-based
// AddUTF8Font, so non-Latin text needs no font file on disk. It also switches
// the renderer's translator to identity, because UTF-8 fonts take UTF-8 strings
// directly and the cp1252 translation used for the core fonts would corrupt
// them; call SetTranslator if you later go back to a core font. Select the font
// afterwards with the Font component or SetFont.
func (r *Renderer) LoadUTF8FontBytes(family string, style FontStyle, ttf []byte) {
	r.AddUTF8FontFromBytes(family, string(style), ttf)
	r.SetTranslator(returnStringUnchanged)
}

// LoadUTF8FontReader is [Renderer.LoadUTF8FontBytes] reading the font from src.
// A read error is recorded on the renderer and surfaces from the next Error().
func (r *Renderer) LoadUTF8FontReader(family string, style FontStyle, src io.Reader) {
	ttf, err := io.ReadAll(src)
	if err != nil {
		r.SetError(err)
		return
	}
	r.LoadUTF8FontBytes(family, style, ttf)
}

// LineHeight returns the default line height for flowing text in document
// units, resolving the "auto" zero value to 1.15× the current font size.
func (r *Renderer) LineHeight() float64 {
	return r.lineHt(r.lineHeight)
}

// SetLineHeight sets the default line height in document units for flowing
// text. A value <= 0 restores automatic height derived from the font size.
func (r *Renderer) SetLineHeight(h float64) {
	r.lineHeight = h
}

// lineHt resolves an explicit height: h if positive, otherwise the configured
// default, otherwise 1.15× the current font size converted to document units.
func (r *Renderer) lineHt(h float64) float64 {
	if h > 0 {
		return h
	}
	if r.lineHeight > 0 {
		return r.lineHeight
	}
	_, unitSize := r.GetFontSize()
	return unitSize * 1.15
}

// ContentWidth returns the printable width of a page in document units: the
// page width minus the left and right margins.
func (r *Renderer) ContentWidth() float64 {
	left, _, right, _ := r.GetMargins()
	width, _ := r.GetPageSize()
	return width - left - right
}

// ContentHeight returns the printable height of a page in document units: the
// page height minus the top margin and the bottom (auto page break) margin.
func (r *Renderer) ContentHeight() float64 {
	_, top, _, bottom := r.GetMargins()
	_, height := r.GetPageSize()
	return height - top - bottom
}

// RemainingHeight returns the vertical space in document units between the
// current cursor position and the bottom (auto page break) margin — the height
// a block must fit into to avoid a page break.
func (r *Renderer) RemainingHeight() float64 {
	_, _, _, bottom := r.GetMargins()
	_, height := r.GetPageSize()
	return height - bottom - r.GetY()
}

// ensurePage adds the first page if none has been started yet, so a component
// can draw without the caller having to remember an explicit AddPage / Page.
func (r *Renderer) ensurePage() {
	if r.PageNo() == 0 {
		r.AddPage()
	}
}

// tr translates s for the current font and is used by the text components.
func (r *Renderer) tr(s string) string {
	if r.translate == nil {
		return s
	}
	return r.translate(s)
}
