package svg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
)

// Attribs is an alias for mx.Attribs, a map of attribute names to values that
// can be passed as element attributes.
type Attribs = mx.Attribs

// Value is the type set accepted by the mixed-type attribute constructors in
// this package — those whose SVG value may be either a number or a string,
// because the spec allows a bare number as well as a length with a unit, a
// percentage, or a keyword. Strings pass through unchanged; numbers are
// formatted as plain decimals, so svg.CX(50), svg.Width("100%") and
// svg.RX("auto") all work.
//
// Attributes are typed according to what their SVG value can be:
//   - always a single plain number → float64 (e.g. StrokeMiterLimit), or int
//     for integer-only attributes (e.g. NumOctaves)
//   - a list of plain numbers → ...float64, rendered space-separated like ViewBox
//   - number or string (length/percentage/unit/keyword) → generic over Value
//   - never numeric (keyword, color, URL, path data, transform, timing) → string
type Value interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Attrib creates an attribute with the given name and value. The value may be a
// string or any number type (see Value), so svg.Width("100%") and svg.Width(100)
// both work. Strings pass through unchanged; float values are formatted as plain
// decimals (never scientific notation like "1e-07") to stay canonical across SVG
// processors. It also serves as an escape hatch for attributes not covered by a
// dedicated constructor.
func Attrib[T Value](name string, value T) mx.Attribute {
	return mx.Attribute{Name: name, Value: attribStringValue(value)}
}

// attribStringValue formats an attribute value as a string. Strings pass through
// and integer types go through fmt.Sprint, but floats use strconv.FormatFloat
// with the 'f' format so small or large magnitudes render as plain decimals
// instead of fmt's scientific notation (e.g. 0.00005 not "5e-05").
func attribStringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprint(value)
	}
}

// joinNums formats the numbers as plain decimals joined by sep. The buffer is
// sliced from a stack-allocated array so no heap allocation happens while the
// formatted output fits in 256 bytes; longer lists fall back to append's growth.
func joinNums(sep string, values ...float64) string {
	var arr [256]byte
	buf := arr[:0]
	for i, v := range values {
		if i > 0 {
			buf = append(buf, sep...)
		}
		buf = strconv.AppendFloat(buf, v, 'f', -1, 64)
	}
	return string(buf)
}

// numListAttrib renders a space-separated list of numbers, the form used by
// viewBox, points, stdDeviation and similar attributes.
func numListAttrib(name string, values ...float64) mx.Attrib {
	return mx.Attribute{Name: name, Value: joinNums(" ", values...)}
}

// XMLNS sets xmlns to the SVG namespace, required on a standalone <svg> root.
const XMLNS = mx.ConstAttrib("xmlns=" + NS)

// XMLNSXLink declares the deprecated xlink namespace, needed when using XLinkHref.
const XMLNSXLink = mx.ConstAttrib("xmlns:xlink=" + XLinkNS)

// See https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute
//
// Some presentation and animation attribute names collide with element names
// (filter, mask, clip-path, path). Element constructors keep the clean name;
// the colliding attributes carry an "Attr" suffix: FilterAttr, MaskAttr,
// ClipPathAttr, PathAttr.

// Core attributes

// ID maps the SVG id attribute.
func ID(value string) mx.Attrib { return Attrib("id", value) }

// Class maps the SVG class attribute.
func Class(classes ...string) mx.Attrib {
	return Attrib("class", strings.Join(classes, " "))
}

// Style maps the SVG style attribute.
func Style(value string) mx.Attrib { return Attrib("style", value) }

// Lang maps the SVG lang attribute.
func Lang(value string) mx.Attrib { return Attrib("lang", value) }

// TabIndex maps the SVG tabindex attribute.
func TabIndex(value int) mx.Attrib { return Attrib("tabindex", value) }

// Role maps the SVG role attribute.
func Role(value string) mx.Attrib { return Attrib("role", value) }

// Version maps the SVG version attribute.
func Version(value string) mx.Attrib { return Attrib("version", value) }

// References (xlink:href is deprecated in favor of href)

// Href maps the SVG href attribute.
func Href(value string) mx.Attrib { return Attrib("href", value) }

// XLinkHref maps the SVG xlink:href attribute.
func XLinkHref(value string) mx.Attrib { return Attrib("xlink:href", value) }

// Viewport and coordinate system

// ViewBox maps the SVG viewBox attribute, whose value is always exactly the
// four numbers min-x, min-y, width and height (width and height must be
// non-negative). They are formatted as plain decimals, so ViewBox(0, 0, 24, 24)
// renders viewBox="0 0 24 24".
func ViewBox(minX, minY, width, height float64) mx.Attrib {
	return numListAttrib("viewBox", minX, minY, width, height)
}

// PreserveAspectRatio maps the SVG preserveAspectRatio attribute.
func PreserveAspectRatio(value string) mx.Attrib {
	return Attrib("preserveAspectRatio", value)
}

// Geometry
//
// Coordinate and size attributes are <length-percentage> values: a bare number
// (user units), a length with a unit ("10px") or a percentage ("50%"). x, y, dx
// and dy additionally accept a space-separated list on <text>/<tspan>, and rx,
// ry, width and height also accept the "auto" keyword — all reachable through
// the string side of the generic Value.

// X maps the SVG x attribute.
func X[T Value](value T) mx.Attrib { return Attrib("x", value) }

// Y maps the SVG y attribute.
func Y[T Value](value T) mx.Attrib { return Attrib("y", value) }

// X1 maps the SVG x1 attribute.
func X1[T Value](value T) mx.Attrib { return Attrib("x1", value) }

// Y1 maps the SVG y1 attribute.
func Y1[T Value](value T) mx.Attrib { return Attrib("y1", value) }

// X2 maps the SVG x2 attribute.
func X2[T Value](value T) mx.Attrib { return Attrib("x2", value) }

// Y2 maps the SVG y2 attribute.
func Y2[T Value](value T) mx.Attrib { return Attrib("y2", value) }

// CX maps the SVG cx attribute.
func CX[T Value](value T) mx.Attrib { return Attrib("cx", value) }

// CY maps the SVG cy attribute.
func CY[T Value](value T) mx.Attrib { return Attrib("cy", value) }

// R maps the SVG r attribute.
func R[T Value](value T) mx.Attrib { return Attrib("r", value) }

// RX maps the SVG rx attribute.
func RX[T Value](value T) mx.Attrib { return Attrib("rx", value) }

// RY maps the SVG ry attribute.
func RY[T Value](value T) mx.Attrib { return Attrib("ry", value) }

// FX maps the SVG fx attribute.
func FX[T Value](value T) mx.Attrib { return Attrib("fx", value) }

// FY maps the SVG fy attribute.
func FY[T Value](value T) mx.Attrib { return Attrib("fy", value) }

// FR maps the SVG fr attribute.
func FR[T Value](value T) mx.Attrib { return Attrib("fr", value) }

// Width maps the SVG width attribute.
func Width[T Value](value T) mx.Attrib { return Attrib("width", value) }

// Height maps the SVG height attribute.
func Height[T Value](value T) mx.Attrib { return Attrib("height", value) }

// DX maps the SVG dx attribute.
func DX[T Value](value T) mx.Attrib { return Attrib("dx", value) }

// DY maps the SVG dy attribute.
func DY[T Value](value T) mx.Attrib { return Attrib("dy", value) }

// D maps the SVG d attribute (path data).
func D(value string) mx.Attrib { return Attrib("d", value) }

// Points maps the SVG points attribute: the list of coordinate numbers of a
// <polyline> or <polygon>, rendered space-separated, e.g.
// Points(0, 0, 10, 0, 10, 10) renders points="0 0 10 0 10 10".
func Points(coords ...float64) mx.Attrib { return numListAttrib("points", coords...) }

// Rotate maps the SVG rotate attribute. On <text>/<tspan> it is a list of
// per-glyph angles (Rotate(45) or, as a list, Rotate("0 90 180")); on
// <animateMotion> it is a number or the keyword "auto"/"auto-reverse".
func Rotate[T Value](value T) mx.Attrib { return Attrib("rotate", value) }

// PathLength maps the SVG pathLength attribute.
func PathLength(value float64) mx.Attrib { return Attrib("pathLength", value) }

// Fill and stroke presentation

// Fill maps the SVG fill attribute (a paint: color, url() or keyword).
func Fill(value string) mx.Attrib { return Attrib("fill", value) }

// FillOpacity maps the SVG fill-opacity attribute (a number or a percentage).
func FillOpacity[T Value](value T) mx.Attrib { return Attrib("fill-opacity", value) }

// Stroke maps the SVG stroke attribute (a paint: color, url() or keyword).
func Stroke(value string) mx.Attrib { return Attrib("stroke", value) }

// StrokeWidth maps the SVG stroke-width attribute (a <length-percentage>).
func StrokeWidth[T Value](value T) mx.Attrib { return Attrib("stroke-width", value) }

// StrokeOpacity maps the SVG stroke-opacity attribute (a number or a percentage).
func StrokeOpacity[T Value](value T) mx.Attrib { return Attrib("stroke-opacity", value) }

// StrokeDashArray maps the SVG stroke-dasharray attribute: a list of lengths or
// percentages, or the keyword "none" (e.g. "4 2" or "none").
func StrokeDashArray(value string) mx.Attrib { return Attrib("stroke-dasharray", value) }

// StrokeDashOffset maps the SVG stroke-dashoffset attribute (a <length-percentage>).
func StrokeDashOffset[T Value](value T) mx.Attrib { return Attrib("stroke-dashoffset", value) }

// StrokeMiterLimit maps the SVG stroke-miterlimit attribute.
func StrokeMiterLimit(value float64) mx.Attrib { return Attrib("stroke-miterlimit", value) }

// General presentation

// Opacity maps the SVG opacity attribute (a number or a percentage).
func Opacity[T Value](value T) mx.Attrib { return Attrib("opacity", value) }

// Color maps the SVG color attribute.
func Color(value string) mx.Attrib { return Attrib("color", value) }

// Cursor maps the SVG cursor attribute.
func Cursor(value string) mx.Attrib { return Attrib("cursor", value) }

// Display maps the SVG display attribute.
func Display(value string) mx.Attrib { return Attrib("display", value) }

// Transform maps the SVG transform attribute (a transform-list).
func Transform(value string) mx.Attrib { return Attrib("transform", value) }

// TransformOrigin maps the SVG transform-origin attribute.
func TransformOrigin(value string) mx.Attrib { return Attrib("transform-origin", value) }

// ImageRendering maps the SVG image-rendering attribute.
func ImageRendering(value string) mx.Attrib { return Attrib("image-rendering", value) }

// PaintOrder maps the SVG paint-order attribute.
func PaintOrder(value string) mx.Attrib { return Attrib("paint-order", value) }

// References to <clipPath>, <mask> and <filter> (names suffixed to avoid
// colliding with the element constructors).

// ClipPathAttr maps the SVG clip-path attribute.
func ClipPathAttr(value string) mx.Attrib { return Attrib("clip-path", value) }

// MaskAttr maps the SVG mask attribute.
func MaskAttr(value string) mx.Attrib { return Attrib("mask", value) }

// FilterAttr maps the SVG filter attribute.
func FilterAttr(value string) mx.Attrib { return Attrib("filter", value) }

// Markers

// MarkerStart maps the SVG marker-start attribute.
func MarkerStart(value string) mx.Attrib { return Attrib("marker-start", value) }

// MarkerMid maps the SVG marker-mid attribute.
func MarkerMid(value string) mx.Attrib { return Attrib("marker-mid", value) }

// MarkerEnd maps the SVG marker-end attribute.
func MarkerEnd(value string) mx.Attrib { return Attrib("marker-end", value) }

// MarkerWidth maps the SVG markerWidth attribute (a <length-percentage>).
func MarkerWidth[T Value](value T) mx.Attrib { return Attrib("markerWidth", value) }

// MarkerHeight maps the SVG markerHeight attribute (a <length-percentage>).
func MarkerHeight[T Value](value T) mx.Attrib { return Attrib("markerHeight", value) }

// RefX maps the SVG refX attribute (a <length-percentage> or a keyword such as
// "left"/"center"/"right").
func RefX[T Value](value T) mx.Attrib { return Attrib("refX", value) }

// RefY maps the SVG refY attribute (a <length-percentage> or a keyword such as
// "top"/"center"/"bottom").
func RefY[T Value](value T) mx.Attrib { return Attrib("refY", value) }

// Orient maps the SVG orient attribute (a number/angle or the keyword
// "auto"/"auto-start-reverse").
func Orient[T Value](value T) mx.Attrib { return Attrib("orient", value) }

// Text and fonts

// FontFamily maps the SVG font-family attribute.
func FontFamily(value string) mx.Attrib { return Attrib("font-family", value) }

// FontSize maps the SVG font-size attribute (a <length-percentage> or a keyword
// such as "medium"/"larger").
func FontSize[T Value](value T) mx.Attrib { return Attrib("font-size", value) }

// FontSizeAdjust maps the SVG font-size-adjust attribute (a number or the
// keyword "none").
func FontSizeAdjust[T Value](value T) mx.Attrib { return Attrib("font-size-adjust", value) }

// FontStretch maps the SVG font-stretch attribute (a keyword such as
// "condensed" or a percentage like "75%").
func FontStretch(value string) mx.Attrib { return Attrib("font-stretch", value) }

// FontVariant maps the SVG font-variant attribute.
func FontVariant(value string) mx.Attrib { return Attrib("font-variant", value) }

// FontWeight maps the SVG font-weight attribute (a number 1–1000 or a keyword
// such as "bold"/"normal").
func FontWeight[T Value](value T) mx.Attrib { return Attrib("font-weight", value) }

// BaselineShift maps the SVG baseline-shift attribute (a <length-percentage> or
// a keyword such as "sub"/"super").
func BaselineShift[T Value](value T) mx.Attrib { return Attrib("baseline-shift", value) }

// LetterSpacing maps the SVG letter-spacing attribute (a <length> or the keyword
// "normal").
func LetterSpacing[T Value](value T) mx.Attrib { return Attrib("letter-spacing", value) }

// WordSpacing maps the SVG word-spacing attribute (a <length> or the keyword
// "normal").
func WordSpacing[T Value](value T) mx.Attrib { return Attrib("word-spacing", value) }

// TextDecoration maps the SVG text-decoration attribute.
func TextDecoration(value string) mx.Attrib { return Attrib("text-decoration", value) }

// TextLength maps the SVG textLength attribute (a <length-percentage>).
func TextLength[T Value](value T) mx.Attrib { return Attrib("textLength", value) }

// StartOffset maps the SVG startOffset attribute (a <length-percentage>).
func StartOffset[T Value](value T) mx.Attrib { return Attrib("startOffset", value) }

// Gradients and patterns

// GradientTransform maps the SVG gradientTransform attribute (a transform-list).
func GradientTransform(value string) mx.Attrib {
	return Attrib("gradientTransform", value)
}

// Offset maps the SVG offset attribute of a gradient <stop> (a number or a
// percentage).
func Offset[T Value](value T) mx.Attrib { return Attrib("offset", value) }

// StopColor maps the SVG stop-color attribute.
func StopColor(value string) mx.Attrib { return Attrib("stop-color", value) }

// StopOpacity maps the SVG stop-opacity attribute (a number or a percentage).
func StopOpacity[T Value](value T) mx.Attrib { return Attrib("stop-opacity", value) }

// PatternTransform maps the SVG patternTransform attribute (a transform-list).
func PatternTransform(value string) mx.Attrib {
	return Attrib("patternTransform", value)
}

// Filters and filter primitives

// In maps the SVG in attribute (a keyword like "SourceGraphic" or a result name).
func In(value string) mx.Attrib { return Attrib("in", value) }

// In2 maps the SVG in2 attribute (a keyword like "SourceGraphic" or a result name).
func In2(value string) mx.Attrib { return Attrib("in2", value) }

// Result maps the SVG result attribute.
func Result(value string) mx.Attrib { return Attrib("result", value) }

// StdDeviation maps the SVG stdDeviation attribute: one number for both axes, or
// two for the x and y standard deviations, e.g. StdDeviation(2) or
// StdDeviation(2, 3).
func StdDeviation(values ...float64) mx.Attrib { return numListAttrib("stdDeviation", values...) }

// Type maps the SVG type attribute.
func Type(value string) mx.Attrib { return Attrib("type", value) }

// Values maps the SVG values attribute. On filter primitives it is a list of
// numbers, on animation elements a ';'-separated list of values of any type, so
// it is passed as a preformatted string.
func Values(value string) mx.Attrib { return Attrib("values", value) }

// Operator maps the SVG operator attribute.
func Operator(value string) mx.Attrib { return Attrib("operator", value) }

// K1 maps the SVG k1 attribute.
func K1(value float64) mx.Attrib { return Attrib("k1", value) }

// K2 maps the SVG k2 attribute.
func K2(value float64) mx.Attrib { return Attrib("k2", value) }

// K3 maps the SVG k3 attribute.
func K3(value float64) mx.Attrib { return Attrib("k3", value) }

// K4 maps the SVG k4 attribute.
func K4(value float64) mx.Attrib { return Attrib("k4", value) }

// Scale maps the SVG scale attribute.
func Scale(value float64) mx.Attrib { return Attrib("scale", value) }

// BaseFrequency maps the SVG baseFrequency attribute: one number for both axes,
// or two for the x and y base frequencies, e.g. BaseFrequency(0.05) or
// BaseFrequency(0.05, 0.1).
func BaseFrequency(values ...float64) mx.Attrib { return numListAttrib("baseFrequency", values...) }

// NumOctaves maps the SVG numOctaves attribute.
func NumOctaves(value int) mx.Attrib { return Attrib("numOctaves", value) }

// Seed maps the SVG seed attribute.
func Seed(value float64) mx.Attrib { return Attrib("seed", value) }

// KernelMatrix maps the SVG kernelMatrix attribute: the list of order-x × order-y
// numbers of the convolution kernel, rendered space-separated.
func KernelMatrix(values ...float64) mx.Attrib { return numListAttrib("kernelMatrix", values...) }

// Order maps the SVG order attribute: one number for a square kernel, or two for
// the x and y kernel sizes, e.g. Order(3) or Order(3, 2).
func Order(values ...float64) mx.Attrib { return numListAttrib("order", values...) }

// Divisor maps the SVG divisor attribute.
func Divisor(value float64) mx.Attrib { return Attrib("divisor", value) }

// Bias maps the SVG bias attribute.
func Bias(value float64) mx.Attrib { return Attrib("bias", value) }

// TargetX maps the SVG targetX attribute.
func TargetX(value int) mx.Attrib { return Attrib("targetX", value) }

// TargetY maps the SVG targetY attribute.
func TargetY(value int) mx.Attrib { return Attrib("targetY", value) }

// PreserveAlpha maps the SVG preserveAlpha attribute (the literal "true" or
// "false").
func PreserveAlpha(value string) mx.Attrib { return Attrib("preserveAlpha", value) }

// KernelUnitLength maps the SVG kernelUnitLength attribute: one number for both
// axes, or two for the x and y intended distances, e.g. KernelUnitLength(1) or
// KernelUnitLength(1, 1).
func KernelUnitLength(values ...float64) mx.Attrib {
	return numListAttrib("kernelUnitLength", values...)
}

// SurfaceScale maps the SVG surfaceScale attribute.
func SurfaceScale(value float64) mx.Attrib { return Attrib("surfaceScale", value) }

// SpecularConstant maps the SVG specularConstant attribute.
func SpecularConstant(value float64) mx.Attrib {
	return Attrib("specularConstant", value)
}

// SpecularExponent maps the SVG specularExponent attribute.
func SpecularExponent(value float64) mx.Attrib {
	return Attrib("specularExponent", value)
}

// DiffuseConstant maps the SVG diffuseConstant attribute.
func DiffuseConstant(value float64) mx.Attrib {
	return Attrib("diffuseConstant", value)
}

// Azimuth maps the SVG azimuth attribute.
func Azimuth(value float64) mx.Attrib { return Attrib("azimuth", value) }

// Elevation maps the SVG elevation attribute.
func Elevation(value float64) mx.Attrib { return Attrib("elevation", value) }

// PointsAtX maps the SVG pointsAtX attribute.
func PointsAtX(value float64) mx.Attrib { return Attrib("pointsAtX", value) }

// PointsAtY maps the SVG pointsAtY attribute.
func PointsAtY(value float64) mx.Attrib { return Attrib("pointsAtY", value) }

// PointsAtZ maps the SVG pointsAtZ attribute.
func PointsAtZ(value float64) mx.Attrib { return Attrib("pointsAtZ", value) }

// LimitingConeAngle maps the SVG limitingConeAngle attribute.
func LimitingConeAngle(value float64) mx.Attrib {
	return Attrib("limitingConeAngle", value)
}

// Radius maps the SVG radius attribute: one number for both axes, or two for the
// x and y radii, e.g. Radius(2) or Radius(2, 3).
func Radius(values ...float64) mx.Attrib { return numListAttrib("radius", values...) }

// TableValues maps the SVG tableValues attribute: the list of transfer-function
// numbers, rendered space-separated.
func TableValues(values ...float64) mx.Attrib { return numListAttrib("tableValues", values...) }

// Slope maps the SVG slope attribute.
func Slope(value float64) mx.Attrib { return Attrib("slope", value) }

// Intercept maps the SVG intercept attribute.
func Intercept(value float64) mx.Attrib { return Attrib("intercept", value) }

// Amplitude maps the SVG amplitude attribute.
func Amplitude(value float64) mx.Attrib { return Attrib("amplitude", value) }

// Exponent maps the SVG exponent attribute.
func Exponent(value float64) mx.Attrib { return Attrib("exponent", value) }

// FloodColor maps the SVG flood-color attribute.
func FloodColor(value string) mx.Attrib { return Attrib("flood-color", value) }

// FloodOpacity maps the SVG flood-opacity attribute (a number or a percentage).
func FloodOpacity[T Value](value T) mx.Attrib { return Attrib("flood-opacity", value) }

// LightingColor maps the SVG lighting-color attribute.
func LightingColor(value string) mx.Attrib {
	return Attrib("lighting-color", value)
}

// Animation

// AttributeName maps the SVG attributeName attribute.
func AttributeName(value string) mx.Attrib { return Attrib("attributeName", value) }

// Begin maps the SVG begin attribute (a begin-value-list).
func Begin(value string) mx.Attrib { return Attrib("begin", value) }

// End maps the SVG end attribute (an end-value-list).
func End(value string) mx.Attrib { return Attrib("end", value) }

// Dur maps the SVG dur attribute (a clock-value or the keyword "indefinite").
func Dur(value string) mx.Attrib { return Attrib("dur", value) }

// From maps the SVG from attribute.
func From(value string) mx.Attrib { return Attrib("from", value) }

// To maps the SVG to attribute.
func To(value string) mx.Attrib { return Attrib("to", value) }

// By maps the SVG by attribute.
func By(value string) mx.Attrib { return Attrib("by", value) }

// RepeatCount maps the SVG repeatCount attribute (a number or the keyword
// "indefinite").
func RepeatCount[T Value](value T) mx.Attrib { return Attrib("repeatCount", value) }

// RepeatDur maps the SVG repeatDur attribute (a clock-value or "indefinite").
func RepeatDur(value string) mx.Attrib { return Attrib("repeatDur", value) }

// KeyTimes maps the SVG keyTimes attribute: the time fractions in [0,1], one per
// animation value, rendered as a ';'-separated list, e.g. KeyTimes(0, 0.5, 1)
// renders keyTimes="0;0.5;1".
func KeyTimes(values ...float64) mx.Attrib {
	return mx.Attribute{Name: "keyTimes", Value: joinNums(";", values...)}
}

// KeySplines maps the SVG keySplines attribute: the cubic Bézier control points
// for calcMode="spline" animation, four numbers (x1 y1 x2 y2) per spline. The
// flat list is grouped four per spline, the four numbers within a spline joined
// by spaces and the splines by ';', e.g. KeySplines(0, 0, 1, 1, 0.5, 0, 0.5, 1)
// renders keySplines="0 0 1 1;0.5 0 0.5 1".
//
// The number of values must be a non-empty multiple of four. If it is not,
// KeySplines returns an mx.ErrAttrib, so rendering the enclosing element fails
// with an error instead of emitting a malformed keySplines list.
func KeySplines(values ...float64) mx.Attrib {
	if len(values) == 0 || len(values)%4 != 0 {
		return mx.ErrAttribf("keySplines", "keySplines needs 4 values (x1 y1 x2 y2) per spline, got %d", len(values))
	}
	// Buffer sliced from a stack array; no heap allocation while it fits 256 bytes.
	var arr [1024]byte
	buf := arr[:0]
	for i, v := range values {
		if i > 0 {
			// ';' between splines (every fourth value), ' ' within a spline.
			if i%4 == 0 {
				buf = append(buf, ';')
			} else {
				buf = append(buf, ' ')
			}
		}
		buf = strconv.AppendFloat(buf, v, 'f', -1, 64)
	}
	return mx.Attribute{Name: "keySplines", Value: string(buf)}
}

// KeyPoints maps the SVG keyPoints attribute: the distance fractions in [0,1]
// used with calcMode="paced", rendered as a ';'-separated list, e.g.
// KeyPoints(0, 0.5, 1) renders keyPoints="0;0.5;1".
func KeyPoints(values ...float64) mx.Attrib {
	return mx.Attribute{Name: "keyPoints", Value: joinNums(";", values...)}
}

// Min maps the SVG min attribute (a clock-value or the keyword "media").
func Min(value string) mx.Attrib { return Attrib("min", value) }

// Max maps the SVG max attribute (a clock-value or the keyword "media").
func Max(value string) mx.Attrib { return Attrib("max", value) }

// Origin maps the SVG origin attribute.
func Origin(value string) mx.Attrib { return Attrib("origin", value) }

// PathAttr sets the "path" attribute (e.g. on <animateMotion>); suffixed to
// avoid colliding with the Path element constructor.
func PathAttr(value string) mx.Attrib { return Attrib("path", value) }

// Conditional processing

// SystemLanguage maps the SVG systemLanguage attribute.
func SystemLanguage(value string) mx.Attrib {
	return Attrib("systemLanguage", value)
}

// RequiredExtensions maps the SVG requiredExtensions attribute.
func RequiredExtensions(value string) mx.Attrib {
	return Attrib("requiredExtensions", value)
}

// RequiredFeatures maps the SVG requiredFeatures attribute.
func RequiredFeatures(value string) mx.Attrib {
	return Attrib("requiredFeatures", value)
}
