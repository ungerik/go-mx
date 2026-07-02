package pdf

import (
	"io"
	"math"

	"github.com/domonda/go-errs"
)

// Routines in this file are translated from the work of Moritz Wagner and
// Andreas Würmser.

// TransformMatrix is used for generalized transformations of text, drawings
// and images.
type TransformMatrix struct {
	A, B, C, D, E, F float64
}

// TransformBegin sets up a transformation context for subsequent text,
// drawings and images. The typical usage is to immediately follow a call to
// this method with a call to one or more of the transformation methods such as
// TransformScale(), TransformSkew(), etc. This is followed by text, drawing or
// image output and finally a call to TransformEnd(). All transformation
// contexts must be properly ended prior to outputting the document.
func (r *Renderer) TransformBegin() {
	r.transformNest++
	r.out("q")
}

// TransformScaleX scales the width of the following text, drawings and images.
// scaleWd is the percentage scaling factor. (x, y) is center of scaling.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformScaleX(scaleWd, x, y float64) {
	r.TransformScale(scaleWd, 100, x, y)
}

// TransformScaleY scales the height of the following text, drawings and
// images. scaleHt is the percentage scaling factor. (x, y) is center of
// scaling.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformScaleY(scaleHt, x, y float64) {
	r.TransformScale(100, scaleHt, x, y)
}

// TransformScaleXY uniformly scales the width and height of the following
// text, drawings and images. s is the percentage scaling factor for both width
// and height. (x, y) is center of scaling.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformScaleXY(s, x, y float64) {
	r.TransformScale(s, s, x, y)
}

// TransformScale generally scales the following text, drawings and images.
// scaleWd and scaleHt are the percentage scaling factors for width and height.
// (x, y) is center of scaling.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformScale(scaleWd, scaleHt, x, y float64) {
	if scaleWd == 0 || scaleHt == 0 {
		r.err = errs.New("scale factor cannot be zero")
		return
	}
	y = (r.h - y) * r.k
	x *= r.k
	scaleWd /= 100
	scaleHt /= 100
	r.Transform(TransformMatrix{scaleWd, 0, 0,
		scaleHt, x * (1 - scaleWd), y * (1 - scaleHt)})
}

// TransformMirrorHorizontal horizontally mirrors the following text, drawings
// and images. x is the axis of reflection.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformMirrorHorizontal(x float64) {
	r.TransformScale(-100, 100, x, r.y)
}

// TransformMirrorVertical vertically mirrors the following text, drawings and
// images. y is the axis of reflection.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformMirrorVertical(y float64) {
	r.TransformScale(100, -100, r.x, y)
}

// TransformMirrorPoint symmetrically mirrors the following text, drawings and
// images on the point specified by (x, y).
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformMirrorPoint(x, y float64) {
	r.TransformScale(-100, -100, x, y)
}

// TransformMirrorLine symmetrically mirrors the following text, drawings and
// images on the line defined by angle and the point (x, y). angles is
// specified in degrees and measured counter-clockwise from the 3 o'clock
// position.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformMirrorLine(angle, x, y float64) {
	r.TransformScale(-100, 100, x, y)
	r.TransformRotate(-2*(angle-90), x, y)
}

// TransformTranslateX moves the following text, drawings and images
// horizontally by the amount specified by tx.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformTranslateX(tx float64) {
	r.TransformTranslate(tx, 0)
}

// TransformTranslateY moves the following text, drawings and images vertically
// by the amount specified by ty.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformTranslateY(ty float64) {
	r.TransformTranslate(0, ty)
}

// TransformTranslate moves the following text, drawings and images
// horizontally and vertically by the amounts specified by tx and ty.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformTranslate(tx, ty float64) {
	r.Transform(TransformMatrix{1, 0, 0, 1, tx * r.k, -ty * r.k})
}

// TransformRotate rotates the following text, drawings and images around the
// center point (x, y). angle is specified in degrees and measured
// counter-clockwise from the 3 o'clock position.
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformRotate(angle, x, y float64) {
	y = (r.h - y) * r.k
	x *= r.k
	angle = angle * math.Pi / 180
	var tm TransformMatrix
	tm.A = math.Cos(angle)
	tm.B = math.Sin(angle)
	tm.C = -tm.B
	tm.D = tm.A
	tm.E = x + tm.B*y - tm.A*x
	tm.F = y - tm.A*y - tm.B*x
	r.Transform(tm)
}

// TransformSkewX horizontally skews the following text, drawings and images
// keeping the point (x, y) stationary. angleX ranges from -90 degrees (skew to
// the left) to 90 degrees (skew to the right).
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformSkewX(angleX, x, y float64) {
	r.TransformSkew(angleX, 0, x, y)
}

// TransformSkewY vertically skews the following text, drawings and images
// keeping the point (x, y) stationary. angleY ranges from -90 degrees (skew to
// the bottom) to 90 degrees (skew to the top).
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformSkewY(angleY, x, y float64) {
	r.TransformSkew(0, angleY, x, y)
}

// TransformSkew generally skews the following text, drawings and images
// keeping the point (x, y) stationary. angleX ranges from -90 degrees (skew to
// the left) to 90 degrees (skew to the right). angleY ranges from -90 degrees
// (skew to the bottom) to 90 degrees (skew to the top).
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformSkew(angleX, angleY, x, y float64) {
	if angleX <= -90 || angleX >= 90 || angleY <= -90 || angleY >= 90 {
		r.err = errs.New("skew values must be between -90° and 90°")
		return
	}
	x *= r.k
	y = (r.h - y) * r.k
	var tm TransformMatrix
	tm.A = 1
	tm.B = math.Tan(angleY * math.Pi / 180)
	tm.C = math.Tan(angleX * math.Pi / 180)
	tm.D = 1
	tm.E = -tm.C * y
	tm.F = -tm.B * x
	r.Transform(tm)
}

// Transform generally transforms the following text, drawings and images
// according to the specified matrix. It is typically easier to use the various
// methods such as TransformRotate() and TransformMirrorVertical() instead.
func (r *Renderer) Transform(tm TransformMatrix) {
	switch {
	case r.transformNest > 0:
		r.outf("%.5f %.5f %.5f %.5f %.5f %.5f cm",
			tm.A, tm.B, tm.C, tm.D, tm.E, tm.F)
	case r.err == nil:
		r.err = errs.New("transformation context is not active")
	}
}

// TransformEnd applies a transformation that was begun with a call to TransformBegin().
//
// The TransformBegin() example demonstrates this method.
func (r *Renderer) TransformEnd() {
	if r.transformNest > 0 {
		r.transformNest--
		r.out("Q")
	} else {
		r.err = errs.New("error attempting to end transformation operation out of sequence")
	}
}

// NewRenderer creates a Renderer for the given page orientation, measurement
// unit and page size, with a Helvetica 12pt default font already selected so
// text can be drawn without further setup.
func NewRenderer(orientation Orientation, unit Unit, size PageSize) *Renderer {
	r := New(string(orientation), string(unit), string(size), "")
	r.translate = r.UnicodeTranslatorFromDescriptor("")
	r.SetFont(DefaultFontFamily, string(StyleRegular), DefaultFontSize)
	return r
}

// NewRendererA4Portrait creates a Renderer for an A4 portrait document
// measured in millimeters.
func NewRendererA4Portrait() *Renderer {
	return NewRenderer(Portrait, UnitMillimeter, A4)
}

// NewRendererA4Landscape creates a Renderer for an A4 landscape document
// measured in millimeters.
func NewRendererA4Landscape() *Renderer {
	return NewRenderer(Landscape, UnitMillimeter, A4)
}

// NewRendererLetterPortrait creates a Renderer for a US Letter portrait
// document measured in inches.
func NewRendererLetterPortrait() *Renderer {
	return NewRenderer(Portrait, UnitInch, Letter)
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
