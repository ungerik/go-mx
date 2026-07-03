package pdf

import (
	"context"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
)

// Best-effort SVG rendering: SVG draws an element tree built with the go-mx
// svg package directly into the PDF as native vector graphics.
//
// This file holds the public component, the element walker, the style model
// and the shape and text painting. Path data lives in svgpath.go, colors in
// svgcolor.go and transforms in svgtransform.go.

// SVG renders an SVG element tree, as built with the go-mx svg package, into
// the box at (x, y) of width w and height h in document units — the vector
// counterpart of [Image]. The root element must be an <svg> element
// (svg.Root or svg.SVG). A zero w or h is derived from the other dimension
// and the SVG's aspect ratio (from its width/height attributes or viewBox);
// if both are zero the SVG's intrinsic size is used as document units.
//
// The rendering is a best effort, not a full SVG implementation: supported
// are the shape elements (rect, circle, ellipse, line, polyline, polygon and
// path with the full d syntax), the containers svg, g, a and switch, basic
// text and tspan, and the common presentation attributes — fill (with
// fill-rule and fill-opacity), stroke (with width, opacity, linecap,
// linejoin, dasharray and dashoffset), opacity, color/currentColor,
// transform, display, visibility, font and text-anchor properties, and
// inline style="…" declarations of the same properties. Colors may be hex,
// rgb()/rgba(), hsl()/hsla() or CSS color keywords. One SVG user unit maps
// to one document unit, scaled by the viewBox-to-box mapping
// (preserveAspectRatio is honored); content is clipped to the box like an
// SVG viewport.
//
// Everything else — gradients, patterns, clip paths, masks, filters,
// markers, symbol/use references, images, animation, CSS stylesheets and
// foreignObject — is skipped silently, and group opacity is approximated by
// multiplying it into the children instead of compositing. Malformed
// attribute values (colors, lengths, transforms, path data) return an error.
//
// Text is drawn with the renderer's fonts: generic font-family values map to
// the matching core font (sans-serif→Helvetica, serif→Times,
// monospace→Courier) and any family registered on the renderer (for example
// via LoadUTF8FontBytes) is matched by name; an unknown family keeps the
// current font. Like [Save], the graphics state is restored afterwards.
func SVG(root *mx.Element, x, y, w, h float64) Component {
	return ComponentFunc(func(ctx context.Context, r *Renderer) error {
		if root == nil {
			return nil
		}
		if root.Err != nil {
			return root.Err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if root.Name != "svg" {
			return errs.Errorf("pdf.SVG needs an <svg> root element, got <%s>", root.Name)
		}
		r.ensurePage()
		if err := r.Error(); err != nil {
			return err
		}
		attrs, err := svgAttribMap(ctx, root)
		if err != nil {
			return err
		}
		w, h, err = resolveSVGBoxSize(attrs, w, h)
		if err != nil {
			return err
		}

		sr := &svgRenderer{r: r}
		saved := sr.saveState()
		err = sr.renderRoot(ctx, root, attrs, x, y, w, h)
		sr.restoreState(saved)
		if err != nil {
			return err
		}
		return r.Error()
	})
}

// resolveSVGBoxSize fills in a zero target width or height from the SVG's
// intrinsic size: its width/height attributes (percentages are ignored) or,
// missing those, its viewBox dimensions.
func resolveSVGBoxSize(attrs map[string]string, w, h float64) (float64, float64, error) {
	intrinsic := func(name string) float64 {
		v, ok := attrs[name]
		if !ok || strings.Contains(v, "%") {
			return 0
		}
		f, err := parseSVGLength(v, 0)
		if err != nil || f <= 0 {
			return 0
		}
		return f
	}
	iw, ih := intrinsic("width"), intrinsic("height")
	if iw <= 0 || ih <= 0 {
		if vals, err := parseSVGNumberList(attrs["viewBox"]); err == nil && len(vals) == 4 && vals[2] > 0 && vals[3] > 0 {
			if iw <= 0 {
				iw = vals[2]
			}
			if ih <= 0 {
				ih = vals[3]
			}
		}
	}
	switch {
	case w > 0 && h > 0:
		return w, h, nil
	case w > 0 && iw > 0 && ih > 0:
		return w, w * ih / iw, nil
	case h > 0 && iw > 0 && ih > 0:
		return h * iw / ih, h, nil
	case w <= 0 && h <= 0 && iw > 0 && ih > 0:
		return iw, ih, nil
	}
	return 0, 0, errs.New("pdf.SVG cannot determine the size: pass w and h, or give the SVG a width/height or viewBox")
}

// svgViewport is the size of the current SVG viewport in user units, the
// reference for percentage lengths.
type svgViewport struct {
	w, h float64
}

// diag is the reference for percentages of lengths that have no horizontal
// or vertical orientation, per the SVG specification.
func (vp svgViewport) diag() float64 {
	return math.Sqrt((vp.w*vp.w + vp.h*vp.h) / 2)
}

// svgRenderer carries the rendering pass state that is not part of the
// inherited style: the target renderer and whether alpha or dash state was
// touched and needs restoring.
type svgRenderer struct {
	r            *Renderer
	alphaTouched bool
	dashTouched  bool
}

// svgSavedState is the graphics state captured before and restored after
// rendering an SVG, mirroring what Save restores plus the alpha and dash
// state (readable here because the engine is part of this package).
type svgSavedState struct {
	family, fontStyle     string
	fontSizePt            float64
	textColor, fill, draw Color
	lineWidth             float64
	capStyle, joinStyle   string
	x, y                  float64
	alpha                 float64
	blendMode             BlendMode
	dashArray             []float64
	dashPhase             float64
}

func (sr *svgRenderer) saveState() svgSavedState {
	r := sr.r
	var s svgSavedState
	s.family = r.GetFontFamily()
	s.fontStyle = r.GetFontStyle()
	s.fontSizePt, _ = r.GetFontSize()
	s.textColor.R, s.textColor.G, s.textColor.B = r.GetTextColor()
	s.fill.R, s.fill.G, s.fill.B = r.GetFillColor()
	s.draw.R, s.draw.G, s.draw.B = r.GetDrawColor()
	s.lineWidth = r.GetLineWidth()
	s.capStyle = r.GetLineCapStyle()
	s.joinStyle = r.GetLineJoinStyle()
	s.x, s.y = r.GetXY()
	s.alpha, s.blendMode = r.GetAlpha()
	if s.blendMode == "" {
		// SetAlpha was never called: the effective state is fully opaque.
		s.alpha, s.blendMode = 1, BlendModeNormal
	}
	s.dashArray = slices.Clone(r.dashArray)
	s.dashPhase = r.dashPhase

	// SVG content defines its own transparency and dashing, so start from
	// the defaults if the ambient state differs.
	if s.alpha != 1 {
		r.SetAlpha(1, BlendModeNormal)
		sr.alphaTouched = true
	}
	if len(s.dashArray) > 0 {
		r.SetDashPattern([]float64{}, 0)
		sr.dashTouched = true
	}
	return s
}

func (sr *svgRenderer) restoreState(s svgSavedState) {
	r := sr.r
	r.SetFont(s.family, s.fontStyle, s.fontSizePt)
	r.SetTextColor(s.textColor.R, s.textColor.G, s.textColor.B)
	r.SetFillColor(s.fill.R, s.fill.G, s.fill.B)
	r.SetDrawColor(s.draw.R, s.draw.G, s.draw.B)
	r.SetLineWidth(s.lineWidth)
	r.SetLineCapStyle(s.capStyle)
	r.SetLineJoinStyle(s.joinStyle)
	r.SetXY(s.x, s.y)
	if sr.alphaTouched {
		r.SetAlpha(s.alpha, s.blendMode)
	}
	if sr.dashTouched {
		// SetDashPattern would scale by k again; the captured values are
		// already scaled, so restore the fields directly.
		r.dashArray = s.dashArray
		r.dashPhase = s.dashPhase
		if r.page > 0 {
			r.outputDashPattern()
		}
	}
}

// withAlpha runs paint with the given alpha and resets to opaque afterwards,
// so the PDF alpha state never leaks across the q/Q scopes of transforms and
// viewports.
func (sr *svgRenderer) withAlpha(alpha float64, paint func()) {
	if alpha >= 1 {
		paint()
		return
	}
	sr.alphaTouched = true
	sr.r.SetAlpha(alpha, BlendModeNormal)
	paint()
	sr.r.SetAlpha(1, BlendModeNormal)
}

// svgAttribMap resolves an element's attributes into a name→value map,
// keeping the first occurrence of duplicate names.
func svgAttribMap(ctx context.Context, el *mx.Element) (map[string]string, error) {
	if len(el.Attribs) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(el.Attribs))
	for _, a := range el.Attribs {
		name := a.AttribName()
		if _, exists := m[name]; exists {
			continue
		}
		value, err := a.AttribValue(ctx)
		if err != nil {
			return nil, err
		}
		m[name] = value
	}
	return m, nil
}

func (sr *svgRenderer) renderRoot(ctx context.Context, root *mx.Element, attrs map[string]string, x, y, w, h float64) error {
	st := svgDefaultStyle()
	if err := st.apply(attrs, svgViewport{w, h}); err != nil {
		return err
	}
	if st.displayNone {
		return nil
	}
	if t := attrs["transform"]; t != "" {
		m, err := parseSVGTransform(t)
		if err != nil {
			return err
		}
		sr.r.TransformBegin()
		defer sr.r.TransformEnd()
		sr.r.transformSVG(m)
	}
	return sr.renderViewport(ctx, root, attrs, st, x, y, w, h)
}

// renderViewport establishes the SVG viewport at (x, y) of size w×h in the
// current user space: it clips to the viewport, maps the viewBox into it
// honoring preserveAspectRatio, and renders the element's children.
func (sr *svgRenderer) renderViewport(ctx context.Context, el *mx.Element, attrs map[string]string, st svgStyle, x, y, w, h float64) error {
	minX, minY, vbW, vbH := 0.0, 0.0, w, h
	if vb, ok := attrs["viewBox"]; ok {
		vals, err := parseSVGNumberList(vb)
		if err != nil || len(vals) != 4 || vals[2] < 0 || vals[3] < 0 {
			return errs.Errorf("invalid SVG viewBox %q", vb)
		}
		if vals[2] == 0 || vals[3] == 0 {
			return nil // spec: a zero viewBox width or height disables rendering
		}
		minX, minY, vbW, vbH = vals[0], vals[1], vals[2], vals[3]
	}
	sx, sy := w/vbW, h/vbH
	tx, ty := 0.0, 0.0
	alignX, alignY, none, meet, err := parseSVGPreserveAspectRatio(attrs["preserveAspectRatio"])
	if err != nil {
		return err
	}
	if !none {
		s := math.Min(sx, sy)
		if !meet {
			s = math.Max(sx, sy)
		}
		sx, sy = s, s
		tx = (w - vbW*s) * alignX
		ty = (h - vbH*s) * alignY
	}
	m := svgTranslate(x+tx, y+ty).mul(svgScale(sx, sy)).mul(svgTranslate(-minX, -minY))

	r := sr.r
	r.ClipRect(x, y, w, h, false)
	defer r.ClipEnd()
	r.TransformBegin()
	defer r.TransformEnd()
	r.transformSVG(m)
	return sr.renderChildren(ctx, el.Children, st, svgViewport{vbW, vbH})
}

// parseSVGPreserveAspectRatio parses the preserveAspectRatio attribute into
// alignment factors (0, 0.5 or 1 per axis), the "none" flag and meet/slice.
// An empty value is the default "xMidYMid meet".
func parseSVGPreserveAspectRatio(s string) (alignX, alignY float64, none, meet bool, err error) {
	fields := strings.Fields(s)
	if len(fields) > 0 && fields[0] == "defer" {
		fields = fields[1:] // only relevant for <image> references
	}
	if len(fields) == 0 {
		return 0.5, 0.5, false, true, nil
	}
	align := fields[0]
	meet = true
	if len(fields) > 1 {
		switch fields[1] {
		case "meet":
		case "slice":
			meet = false
		default:
			return 0, 0, false, false, errs.Errorf("invalid preserveAspectRatio %q", s)
		}
	}
	if align == "none" {
		return 0, 0, true, meet, nil
	}
	factor := func(part string) (float64, bool) {
		switch part {
		case "Min":
			return 0, true
		case "Mid":
			return 0.5, true
		case "Max":
			return 1, true
		}
		return 0, false
	}
	var okX, okY bool
	if len(align) == 8 && align[0] == 'x' && align[4] == 'Y' {
		alignX, okX = factor(align[1:4])
		alignY, okY = factor(align[5:8])
	}
	if !okX || !okY {
		return 0, 0, false, false, errs.Errorf("invalid preserveAspectRatio %q", s)
	}
	return alignX, alignY, false, meet, nil
}

// svgRenderedElements are the element names the walker interprets; children
// with any other name are skipped silently (best effort).
var svgRenderedElements = map[string]bool{
	"svg": true, "g": true, "a": true, "switch": true,
	"rect": true, "circle": true, "ellipse": true, "line": true,
	"polyline": true, "polygon": true, "path": true,
	"text": true,
}

func (sr *svgRenderer) renderChildren(ctx context.Context, children mx.Components, st svgStyle, vp svgViewport) error {
	for _, child := range children {
		var err error
		switch c := child.(type) {
		case *mx.Element:
			err = sr.renderElement(ctx, c, st, vp)
		case mx.Components:
			err = sr.renderChildren(ctx, c, st, vp)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// renderElement renders one SVG element with the inherited style st (passed
// by value: changes apply to the element and its subtree only).
func (sr *svgRenderer) renderElement(ctx context.Context, el *mx.Element, st svgStyle, vp svgViewport) error {
	if el.Err != nil {
		return el.Err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !svgRenderedElements[el.Name] {
		return nil
	}
	attrs, err := svgAttribMap(ctx, el)
	if err != nil {
		return err
	}
	if err = st.apply(attrs, vp); err != nil {
		return err
	}
	if st.displayNone {
		return nil
	}
	if t := attrs["transform"]; t != "" {
		m, err := parseSVGTransform(t)
		if err != nil {
			return err
		}
		sr.r.TransformBegin()
		defer sr.r.TransformEnd()
		sr.r.transformSVG(m)
	}

	switch el.Name {
	case "svg":
		x, err := svgLengthAttr(attrs, "x", vp.w, 0)
		if err != nil {
			return err
		}
		y, err := svgLengthAttr(attrs, "y", vp.h, 0)
		if err != nil {
			return err
		}
		w, err := svgLengthAttr(attrs, "width", vp.w, vp.w)
		if err != nil {
			return err
		}
		h, err := svgLengthAttr(attrs, "height", vp.h, vp.h)
		if err != nil {
			return err
		}
		if w < 0 || h < 0 {
			return errs.Errorf("negative SVG viewport size %g x %g", w, h)
		}
		if w == 0 || h == 0 {
			return nil
		}
		return sr.renderViewport(ctx, el, attrs, st, x, y, w, h)

	case "g", "a":
		return sr.renderChildren(ctx, el.Children, st, vp)

	case "switch":
		// Best effort without conditional-processing support: render the
		// first element child.
		if first := firstElementChild(el.Children); first != nil {
			return sr.renderElement(ctx, first, st, vp)
		}
		return nil

	case "rect":
		return sr.renderRect(attrs, st, vp)
	case "circle":
		return sr.renderCircle(attrs, st, vp)
	case "ellipse":
		return sr.renderEllipse(attrs, st, vp)
	case "line":
		return sr.renderLine(attrs, st, vp)
	case "polyline", "polygon":
		return sr.renderPoly(attrs, st, el.Name == "polygon")
	case "path":
		return sr.renderPath(attrs, st)
	case "text":
		return sr.renderText(ctx, el, attrs, st, vp)
	}
	return nil
}

func firstElementChild(children mx.Components) *mx.Element {
	for _, child := range children {
		switch c := child.(type) {
		case *mx.Element:
			return c
		case mx.Components:
			if e := firstElementChild(c); e != nil {
				return e
			}
		}
	}
	return nil
}

// Shape elements

func (sr *svgRenderer) renderRect(attrs map[string]string, st svgStyle, vp svgViewport) error {
	x, err := svgLengthAttr(attrs, "x", vp.w, 0)
	if err != nil {
		return err
	}
	y, err := svgLengthAttr(attrs, "y", vp.h, 0)
	if err != nil {
		return err
	}
	w, err := svgLengthAttr(attrs, "width", vp.w, 0)
	if err != nil {
		return err
	}
	h, err := svgLengthAttr(attrs, "height", vp.h, 0)
	if err != nil {
		return err
	}
	if w < 0 || h < 0 {
		return errs.Errorf("negative SVG rect size %g x %g", w, h)
	}
	if w == 0 || h == 0 {
		return nil
	}
	rx, hasRX, err := svgOptionalLengthAttr(attrs, "rx", vp.w)
	if err != nil {
		return err
	}
	ry, hasRY, err := svgOptionalLengthAttr(attrs, "ry", vp.h)
	if err != nil {
		return err
	}
	if rx < 0 || ry < 0 {
		return errs.Errorf("negative SVG rect corner radius %g/%g", rx, ry)
	}
	// The auto rules: a missing radius takes the other one's value.
	if !hasRX {
		rx = ry
	}
	if !hasRY {
		ry = rx
	}
	rx = math.Min(rx, w/2)
	ry = math.Min(ry, h/2)
	sr.paintShape(&st, true, func(sink pathSink) {
		svgRoundedRectPath(sink, x, y, w, h, rx, ry)
	})
	return nil
}

func (sr *svgRenderer) renderCircle(attrs map[string]string, st svgStyle, vp svgViewport) error {
	cx, err := svgLengthAttr(attrs, "cx", vp.w, 0)
	if err != nil {
		return err
	}
	cy, err := svgLengthAttr(attrs, "cy", vp.h, 0)
	if err != nil {
		return err
	}
	rad, err := svgLengthAttr(attrs, "r", vp.diag(), 0)
	if err != nil {
		return err
	}
	if rad < 0 {
		return errs.Errorf("negative SVG circle radius %g", rad)
	}
	if rad == 0 {
		return nil
	}
	sr.paintShape(&st, true, func(sink pathSink) {
		svgEllipsePath(sink, cx, cy, rad, rad)
	})
	return nil
}

func (sr *svgRenderer) renderEllipse(attrs map[string]string, st svgStyle, vp svgViewport) error {
	cx, err := svgLengthAttr(attrs, "cx", vp.w, 0)
	if err != nil {
		return err
	}
	cy, err := svgLengthAttr(attrs, "cy", vp.h, 0)
	if err != nil {
		return err
	}
	rx, hasRX, err := svgOptionalLengthAttr(attrs, "rx", vp.w)
	if err != nil {
		return err
	}
	ry, hasRY, err := svgOptionalLengthAttr(attrs, "ry", vp.h)
	if err != nil {
		return err
	}
	if !hasRX {
		rx = ry
	}
	if !hasRY {
		ry = rx
	}
	if rx < 0 || ry < 0 {
		return errs.Errorf("negative SVG ellipse radius %g/%g", rx, ry)
	}
	if rx == 0 || ry == 0 {
		return nil
	}
	sr.paintShape(&st, true, func(sink pathSink) {
		svgEllipsePath(sink, cx, cy, rx, ry)
	})
	return nil
}

func (sr *svgRenderer) renderLine(attrs map[string]string, st svgStyle, vp svgViewport) error {
	x1, err := svgLengthAttr(attrs, "x1", vp.w, 0)
	if err != nil {
		return err
	}
	y1, err := svgLengthAttr(attrs, "y1", vp.h, 0)
	if err != nil {
		return err
	}
	x2, err := svgLengthAttr(attrs, "x2", vp.w, 0)
	if err != nil {
		return err
	}
	y2, err := svgLengthAttr(attrs, "y2", vp.h, 0)
	if err != nil {
		return err
	}
	// A line has no interior: it is never filled.
	sr.paintShape(&st, false, func(sink pathSink) {
		sink.moveTo(x1, y1)
		sink.lineTo(x2, y2)
	})
	return nil
}

func (sr *svgRenderer) renderPoly(attrs map[string]string, st svgStyle, closed bool) error {
	coords, err := parseSVGNumberList(attrs["points"])
	if err != nil {
		return errs.Errorf("invalid SVG points %q: %w", attrs["points"], err)
	}
	coords = coords[:len(coords)/2*2] // an odd trailing coordinate is dropped
	if len(coords) < 4 {
		return nil
	}
	sr.paintShape(&st, true, func(sink pathSink) {
		sink.moveTo(coords[0], coords[1])
		for i := 2; i < len(coords); i += 2 {
			sink.lineTo(coords[i], coords[i+1])
		}
		if closed {
			sink.closePath()
		}
	})
	return nil
}

func (sr *svgRenderer) renderPath(attrs map[string]string, st svgStyle) error {
	d := strings.TrimSpace(attrs["d"])
	if d == "" {
		return nil
	}
	// Parse into a recording first so malformed path data fails before any
	// path construction reaches the content stream.
	var rec recordedPath
	if err := renderSVGPathData(d, &rec); err != nil {
		return err
	}
	sr.paintShape(&st, true, rec.replay)
	return nil
}

// recordedPath records path segments for validation and replay.
type recordedPath struct {
	segs []recordedPathSeg
}

type recordedPathSeg struct {
	op     byte // 'M', 'L', 'C' or 'Z'
	coords [6]float64
}

func (p *recordedPath) moveTo(x, y float64) {
	p.segs = append(p.segs, recordedPathSeg{op: 'M', coords: [6]float64{x, y}})
}

func (p *recordedPath) lineTo(x, y float64) {
	p.segs = append(p.segs, recordedPathSeg{op: 'L', coords: [6]float64{x, y}})
}

func (p *recordedPath) cubicTo(cx0, cy0, cx1, cy1, x, y float64) {
	p.segs = append(p.segs, recordedPathSeg{op: 'C', coords: [6]float64{cx0, cy0, cx1, cy1, x, y}})
}

func (p *recordedPath) closePath() {
	p.segs = append(p.segs, recordedPathSeg{op: 'Z'})
}

func (p *recordedPath) replay(sink pathSink) {
	for _, s := range p.segs {
		switch s.op {
		case 'M':
			sink.moveTo(s.coords[0], s.coords[1])
		case 'L':
			sink.lineTo(s.coords[0], s.coords[1])
		case 'C':
			sink.cubicTo(s.coords[0], s.coords[1], s.coords[2], s.coords[3], s.coords[4], s.coords[5])
		case 'Z':
			sink.closePath()
		}
	}
}

// paintShape fills and/or strokes the path produced by emit according to the
// style. emit may be called twice when fill and stroke need different alpha
// values.
func (sr *svgRenderer) paintShape(st *svgStyle, fillable bool, emit func(pathSink)) {
	if !st.visible {
		return
	}
	r := sr.r
	fillColor, fillAlpha, hasFill := resolveSVGPaint(st.fill, st.color, st.opacity*st.fillOpacity)
	hasFill = hasFill && fillable
	strokeColor, strokeAlpha, hasStroke := resolveSVGPaint(st.stroke, st.color, st.opacity*st.strokeOpacity)
	hasStroke = hasStroke && st.strokeWidth > 0
	if !hasFill && !hasStroke {
		return
	}
	if hasFill {
		r.SetFillColor(fillColor.R, fillColor.G, fillColor.B)
	}
	dashed := false
	if hasStroke {
		r.SetDrawColor(strokeColor.R, strokeColor.G, strokeColor.B)
		r.SetLineWidth(st.strokeWidth)
		r.SetLineCapStyle(st.lineCap)
		r.SetLineJoinStyle(st.lineJoin)
		if len(st.dashArray) > 0 {
			r.SetDashPattern(st.dashArray, st.dashOffset)
			sr.dashTouched = true
			dashed = true
		}
	}
	fillOp, fillStrokeOp := "F", "FD"
	if st.fillRule == "evenodd" {
		fillOp, fillStrokeOp = "F*", "FD*"
	}
	sink := rendererPathSink{r}
	switch {
	case hasFill && hasStroke && fillAlpha == strokeAlpha:
		sr.withAlpha(fillAlpha, func() {
			emit(sink)
			r.DrawPath(fillStrokeOp)
		})
	default:
		if hasFill {
			sr.withAlpha(fillAlpha, func() {
				emit(sink)
				r.DrawPath(fillOp)
			})
		}
		if hasStroke {
			sr.withAlpha(strokeAlpha, func() {
				emit(sink)
				r.DrawPath("D")
			})
		}
	}
	if dashed {
		r.SetDashPattern([]float64{}, 0)
	}
}

// resolveSVGPaint resolves a parsed paint against the current color (for
// currentColor) and the effective opacity. ok is false when nothing is to be
// painted.
func resolveSVGPaint(p svgPaint, current Color, opacity float64) (c Color, alpha float64, ok bool) {
	if p.none {
		return Color{}, 0, false
	}
	c = p.color
	if p.currentColor {
		c = current
	}
	alpha = math.Min(1, math.Max(0, opacity*p.alpha))
	return c, alpha, alpha > 0
}

// Text elements

type svgTextCursor struct {
	x, y float64
}

func (sr *svgRenderer) renderText(ctx context.Context, el *mx.Element, attrs map[string]string, st svgStyle, vp svgViewport) error {
	cur := &svgTextCursor{}
	if err := positionSVGTextCursor(attrs, vp, cur); err != nil {
		return err
	}
	return sr.renderTextContent(ctx, el.Children, st, vp, cur)
}

// positionSVGTextCursor applies the x, y, dx and dy attributes of <text> or
// <tspan> to the cursor. Only the first value of a coordinate list is used.
func positionSVGTextCursor(attrs map[string]string, vp svgViewport, cur *svgTextCursor) error {
	get := func(name string, ref float64) (float64, bool, error) {
		v, ok := attrs[name]
		if !ok {
			return 0, false, nil
		}
		if fields := strings.FieldsFunc(v, func(r rune) bool {
			return unicode.IsSpace(r) || r == ','
		}); len(fields) > 0 {
			v = fields[0]
		}
		f, err := parseSVGLength(v, ref)
		return f, true, err
	}
	if x, ok, err := get("x", vp.w); err != nil {
		return err
	} else if ok {
		cur.x = x
	}
	if y, ok, err := get("y", vp.h); err != nil {
		return err
	} else if ok {
		cur.y = y
	}
	if dx, ok, err := get("dx", vp.w); err != nil {
		return err
	} else if ok {
		cur.x += dx
	}
	if dy, ok, err := get("dy", vp.h); err != nil {
		return err
	} else if ok {
		cur.y += dy
	}
	return nil
}

func (sr *svgRenderer) renderTextContent(ctx context.Context, children mx.Components, st svgStyle, vp svgViewport, cur *svgTextCursor) error {
	for _, child := range children {
		switch c := child.(type) {
		case mx.Text:
			sr.drawTextRun(string(c), &st, cur)
			if err := sr.r.Error(); err != nil {
				return err
			}
		case mx.Components:
			if err := sr.renderTextContent(ctx, c, st, vp, cur); err != nil {
				return err
			}
		case *mx.Element:
			if c.Err != nil {
				return c.Err
			}
			if c.Name != "tspan" {
				continue // textPath and other text children are unsupported
			}
			attrs, err := svgAttribMap(ctx, c)
			if err != nil {
				return err
			}
			spanStyle := st
			if err = spanStyle.apply(attrs, vp); err != nil {
				return err
			}
			if spanStyle.displayNone {
				continue
			}
			if err = positionSVGTextCursor(attrs, vp, cur); err != nil {
				return err
			}
			if err = sr.renderTextContent(ctx, c.Children, spanStyle, vp, cur); err != nil {
				return err
			}
		}
	}
	return nil
}

// drawTextRun draws one run of text at the cursor and advances it by the
// text width. The run is anchored per text-anchor and painted with the fill
// paint; stroked text is not supported.
func (sr *svgRenderer) drawTextRun(s string, st *svgStyle, cur *svgTextCursor) {
	text := normalizeSVGText(s)
	if text == "" {
		return
	}
	r := sr.r
	sr.selectSVGFont(st)
	translated := r.tr(text)
	width := r.GetStringWidth(translated)
	x := cur.x
	switch st.textAnchor {
	case "middle":
		x -= width / 2
	case "end":
		x -= width
	}
	fillColor, fillAlpha, hasFill := resolveSVGPaint(st.fill, st.color, st.opacity*st.fillOpacity)
	if st.visible && hasFill {
		r.SetTextColor(fillColor.R, fillColor.G, fillColor.B)
		sr.withAlpha(fillAlpha, func() {
			r.Text(x, cur.y, translated)
		})
	}
	cur.x = x + width
}

// normalizeSVGText collapses whitespace runs into single spaces, keeping one
// leading/trailing space so runs split across tspans keep their separation.
func normalizeSVGText(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	joined := strings.Join(fields, " ")
	if unicode.IsSpace(rune(s[0])) {
		joined = " " + joined
	}
	if unicode.IsSpace(rune(s[len(s)-1])) {
		joined += " "
	}
	return joined
}

// selectSVGFont selects the best matching registered font for the style's
// font properties and sets the font size in user units. An unknown family
// keeps the current font.
func (sr *svgRenderer) selectSVGFont(st *svgStyle) {
	r := sr.r
	style := ""
	if st.bold {
		style += "B"
	}
	if st.italic {
		style += "I"
	}
	deco := ""
	if st.underline {
		deco += "U"
	}
	if st.strikeout {
		deco += "S"
	}
	for _, fam := range st.fontFamilies {
		family := strings.ToLower(strings.Trim(strings.TrimSpace(fam), `'"`))
		switch family {
		case "sans-serif", "ui-sans-serif", "system-ui", "arial", "helvetica neue":
			family = "helvetica"
		case "serif", "ui-serif", "times new roman":
			family = "times"
		case "monospace", "ui-monospace", "courier new":
			family = "courier"
		}
		for _, tryStyle := range svgFontStyleFallbacks(style) {
			if svgFontRegistered(r, family, tryStyle) {
				r.SetFont(family, tryStyle+deco, 0)
				r.SetFontUnitSize(st.fontSize)
				return
			}
		}
	}
	r.SetFontUnitSize(st.fontSize)
}

// svgFontStyleFallbacks lists the styles to try in order, degrading to less
// specific variants for fonts registered without bold or italic faces.
func svgFontStyleFallbacks(style string) []string {
	switch style {
	case "BI":
		return []string{"BI", "B", "I", ""}
	case "B", "I":
		return []string{style, ""}
	}
	return []string{""}
}

// svgFontRegistered reports whether the family is usable with SetFont: either
// already registered (e.g. a loaded UTF-8 font) or a core font that SetFont
// loads on demand.
func svgFontRegistered(r *Renderer, family, style string) bool {
	if _, ok := r.fonts[getFontKey(family, style)]; ok {
		return true
	}
	if _, ok := r.coreFonts[family]; ok {
		// The Latin core fonts have all four faces; symbol fonts only one.
		return style == "" || family == "helvetica" || family == "times" || family == "courier"
	}
	return false
}

// Style model

// svgStyle is the inherited presentation state, the SVG counterpart of the
// renderer's graphics state. It is copied down the tree, so an element's
// changes apply to it and its subtree only.
type svgStyle struct {
	displayNone   bool // not inherited, but pruning the subtree makes that moot
	visible       bool
	color         Color // the CSS color property, for currentColor
	fill          svgPaint
	fillRule      string // "nonzero" or "evenodd"
	fillOpacity   float64
	stroke        svgPaint
	strokeOpacity float64
	strokeWidth   float64
	lineCap       string // butt, round or square
	lineJoin      string // miter, round or bevel
	dashArray     []float64
	dashOffset    float64
	opacity       float64 // accumulated group opacity (best-effort approximation)
	fontFamilies  []string
	fontSize      float64
	bold          bool
	italic        bool
	underline     bool
	strikeout     bool
	textAnchor    string // start, middle or end
}

func svgDefaultStyle() svgStyle {
	return svgStyle{
		visible:       true,
		color:         Black,
		fill:          svgPaint{color: Black, alpha: 1},
		fillRule:      "nonzero",
		fillOpacity:   1,
		stroke:        svgPaint{none: true, alpha: 1},
		strokeOpacity: 1,
		strokeWidth:   1,
		lineCap:       "butt",
		lineJoin:      "miter",
		opacity:       1,
		fontSize:      16, // CSS "medium"
		textAnchor:    "start",
	}
}

// apply merges an element's presentation attributes and inline style="…"
// declarations into the style; style declarations win, matching CSS. The
// `opacity` property is special: st.opacity carries the accumulated parent
// opacity, and the element's own opacity multiplies into it exactly once. So
// the accumulated value is snapshotted and st.opacity reset to 1 while the
// property is read (setProperty overrides it, style winning over attribute
// like any other property), then the element's value is multiplied back in.
func (st *svgStyle) apply(attrs map[string]string, vp svgViewport) error {
	inheritedOpacity := st.opacity
	st.opacity = 1
	for name, value := range attrs {
		if name == "style" {
			continue
		}
		if err := st.setProperty(name, value, vp); err != nil {
			return err
		}
	}
	if style, ok := attrs["style"]; ok {
		for decl := range strings.SplitSeq(style, ";") {
			name, value, found := strings.Cut(decl, ":")
			if !found {
				continue
			}
			if err := st.setProperty(strings.TrimSpace(name), strings.TrimSpace(value), vp); err != nil {
				return err
			}
		}
	}
	st.opacity *= inheritedOpacity
	return nil
}

// setProperty applies one presentation property. Properties and keyword
// values outside the supported subset are ignored; malformed color, length
// and number values return an error.
func (st *svgStyle) setProperty(name, value string, vp svgViewport) error {
	value = strings.TrimSpace(value)
	if value == "" || value == "inherit" {
		return nil
	}
	switch name {
	case "display":
		st.displayNone = value == "none"
	case "visibility":
		switch value {
		case "hidden", "collapse":
			st.visible = false
		case "visible":
			st.visible = true
		}
	case "color":
		c, _, err := parseSVGColor(value)
		if err != nil {
			return err
		}
		st.color = c
	case "fill":
		p, err := parseSVGPaint(value)
		if err != nil {
			return err
		}
		st.fill = p
	case "fill-rule":
		if value == "nonzero" || value == "evenodd" {
			st.fillRule = value
		}
	case "fill-opacity":
		return setSVGOpacity(&st.fillOpacity, value)
	case "stroke":
		p, err := parseSVGPaint(value)
		if err != nil {
			return err
		}
		st.stroke = p
	case "stroke-opacity":
		return setSVGOpacity(&st.strokeOpacity, value)
	case "stroke-width":
		w, err := parseSVGLength(value, vp.diag())
		if err != nil {
			return err
		}
		if w < 0 {
			return errs.Errorf("negative stroke-width %q", value)
		}
		st.strokeWidth = w
	case "stroke-linecap":
		if value == "butt" || value == "round" || value == "square" {
			st.lineCap = value
		}
	case "stroke-linejoin":
		switch value {
		case "miter", "round", "bevel":
			st.lineJoin = value
		}
	case "stroke-dasharray":
		return st.setDashArray(value, vp)
	case "stroke-dashoffset":
		offset, err := parseSVGLength(value, vp.diag())
		if err != nil {
			return err
		}
		st.dashOffset = offset
	case "opacity":
		// The element's own opacity; apply() resets st.opacity to 1 first and
		// multiplies the accumulated parent opacity back in, so this overrides
		// (style winning over attribute) instead of multiplying per source.
		return setSVGOpacity(&st.opacity, value)
	case "font-family":
		st.fontFamilies = strings.Split(value, ",")
	case "font-size":
		// Percentages and em are relative to the inherited font size.
		if em, ok := strings.CutSuffix(value, "em"); ok {
			f, err := strconv.ParseFloat(strings.TrimSpace(em), 64)
			if err != nil {
				return errs.Errorf("invalid font-size %q", value)
			}
			st.fontSize = f * st.fontSize
			return nil
		}
		size, err := parseSVGLength(value, st.fontSize)
		if err != nil {
			return nil // keyword sizes like "medium" keep the inherited size
		}
		if size > 0 {
			st.fontSize = size
		}
	case "font-weight":
		switch value {
		case "bold", "bolder":
			st.bold = true
		case "normal", "lighter":
			st.bold = false
		default:
			if weight, err := strconv.ParseFloat(value, 64); err == nil {
				st.bold = weight >= 600
			}
		}
	case "font-style":
		switch value {
		case "italic", "oblique":
			st.italic = true
		case "normal":
			st.italic = false
		}
	case "text-decoration":
		st.underline = strings.Contains(value, "underline")
		st.strikeout = strings.Contains(value, "line-through")
	case "text-anchor":
		switch value {
		case "start", "middle", "end":
			st.textAnchor = value
		}
	}
	return nil
}

func setSVGOpacity(target *float64, value string) error {
	o, err := parseAlphaValue(value)
	if err != nil {
		return err
	}
	*target = o
	return nil
}

func (st *svgStyle) setDashArray(value string, vp svgViewport) error {
	if value == "none" {
		st.dashArray = nil
		return nil
	}
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
	dashes := make([]float64, 0, len(fields))
	allZero := true
	for _, f := range fields {
		d, err := parseSVGLength(f, vp.diag())
		if err != nil {
			return err
		}
		if d < 0 {
			return errs.Errorf("negative value in stroke-dasharray %q", value)
		}
		if d != 0 {
			allZero = false
		}
		dashes = append(dashes, d)
	}
	if len(dashes) == 0 || allZero {
		st.dashArray = nil
		return nil
	}
	// An odd number of values repeats the list once, per the spec.
	if len(dashes)%2 != 0 {
		dashes = append(dashes, dashes...)
	}
	st.dashArray = dashes
	return nil
}

// Length parsing

// parseSVGLength parses an SVG length: a bare number in user units, a
// percentage of ref, or a number with one of the CSS absolute units
// converted at their CSS px ratios (1px = 1 user unit).
func parseSVGLength(s string, ref float64) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errs.New("empty SVG length")
	}
	if p, ok := strings.CutSuffix(s, "%"); ok {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return 0, errs.Errorf("invalid SVG percentage %q", s)
		}
		return v / 100 * ref, nil
	}
	factor := 1.0
	for _, unit := range []struct {
		suffix string
		factor float64
	}{
		{"px", 1},
		{"pt", 96.0 / 72},
		{"pc", 16},
		{"mm", 96.0 / 25.4},
		{"cm", 96.0 / 2.54},
		{"in", 96},
	} {
		if p, ok := strings.CutSuffix(s, unit.suffix); ok {
			s, factor = strings.TrimSpace(p), unit.factor
			break
		}
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errs.Errorf("invalid SVG length %q", s)
	}
	return v * factor, nil
}

// svgLengthAttr parses the named attribute as a length, returning def when
// the attribute is missing or "auto".
func svgLengthAttr(attrs map[string]string, name string, ref, def float64) (float64, error) {
	v, ok := attrs[name]
	if !ok || strings.TrimSpace(v) == "" || strings.TrimSpace(v) == "auto" {
		return def, nil
	}
	return parseSVGLength(v, ref)
}

// svgOptionalLengthAttr is svgLengthAttr for attributes whose absence has
// meaning (like the rx/ry auto rules): ok reports whether a value was given.
func svgOptionalLengthAttr(attrs map[string]string, name string, ref float64) (value float64, ok bool, err error) {
	v, exists := attrs[name]
	if !exists || strings.TrimSpace(v) == "" || strings.TrimSpace(v) == "auto" {
		return 0, false, nil
	}
	value, err = parseSVGLength(v, ref)
	return value, err == nil, err
}
