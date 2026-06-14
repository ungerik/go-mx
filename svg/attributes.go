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

// Value is the set of types accepted by the attribute constructors in this
// package: strings are passed through unchanged, while numbers are formatted
// with fmt.Sprint so number literals can be passed directly, e.g. svg.CX(50)
// or svg.StrokeWidth(1.5) instead of svg.CX("50").
type Value interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Attrib creates an attribute with the given name and value.
// The value may be a string or any number type (see Value), so svg.Width("100%")
// and svg.Width(100) both work. Strings pass through unchanged; float values are
// formatted as plain decimals (never scientific notation like "1e-07") to stay
// canonical across SVG processors.
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
func ID[T Value](value T) mx.Attrib { return Attrib("id", value) }

// Class maps the SVG class attribute.
func Class(classes ...string) mx.Attrib {
	return Attrib("class", strings.Join(classes, " "))
}

// Style maps the SVG style attribute.
func Style[T Value](value T) mx.Attrib { return Attrib("style", value) }

// Lang maps the SVG lang attribute.
func Lang[T Value](value T) mx.Attrib { return Attrib("lang", value) }

// TabIndex maps the SVG tabindex attribute.
func TabIndex[T Value](value T) mx.Attrib { return Attrib("tabindex", value) }

// Role maps the SVG role attribute.
func Role[T Value](value T) mx.Attrib { return Attrib("role", value) }

// Version maps the SVG version attribute.
func Version[T Value](value T) mx.Attrib { return Attrib("version", value) }

// References (xlink:href is deprecated in favor of href)

// Href maps the SVG href attribute.
func Href[T Value](value T) mx.Attrib { return Attrib("href", value) }

// XLinkHref maps the SVG xlink:href attribute.
func XLinkHref[T Value](value T) mx.Attrib { return Attrib("xlink:href", value) }

// CrossOrigin maps the SVG crossorigin attribute.
func CrossOrigin[T Value](value T) mx.Attrib {
	return Attrib("crossorigin", value)
}

// Viewport and coordinate system

// ViewBox maps the SVG viewBox attribute, whose value is always exactly the
// four numbers min-x, min-y, width and height (width and height must be
// non-negative). They are formatted as plain decimals, so ViewBox(0, 0, 24, 24)
// renders viewBox="0 0 24 24".
func ViewBox(minX, minY, width, height float64) mx.Attrib {
	return Attrib("viewBox", attribStringValue(minX)+" "+attribStringValue(minY)+" "+
		attribStringValue(width)+" "+attribStringValue(height))
}

// PreserveAspectRatio maps the SVG preserveAspectRatio attribute.
func PreserveAspectRatio[T Value](value T) mx.Attrib {
	return Attrib("preserveAspectRatio", value)
}

// ZoomAndPan maps the SVG zoomAndPan attribute.
func ZoomAndPan[T Value](value T) mx.Attrib { return Attrib("zoomAndPan", value) }

// Geometry

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

// D maps the SVG d attribute.
func D[T Value](value T) mx.Attrib { return Attrib("d", value) }

// Points maps the SVG points attribute.
func Points[T Value](value T) mx.Attrib { return Attrib("points", value) }

// Rotate maps the SVG rotate attribute.
func Rotate[T Value](value T) mx.Attrib { return Attrib("rotate", value) }

// PathLength maps the SVG pathLength attribute.
func PathLength[T Value](value T) mx.Attrib { return Attrib("pathLength", value) }

// Fill and stroke presentation

// Fill maps the SVG fill attribute.
func Fill[T Value](value T) mx.Attrib { return Attrib("fill", value) }

// FillOpacity maps the SVG fill-opacity attribute.
func FillOpacity[T Value](value T) mx.Attrib { return Attrib("fill-opacity", value) }

// FillRule maps the SVG fill-rule attribute.
func FillRule[T Value](value T) mx.Attrib { return Attrib("fill-rule", value) }

// Stroke maps the SVG stroke attribute.
func Stroke[T Value](value T) mx.Attrib { return Attrib("stroke", value) }

// StrokeWidth maps the SVG stroke-width attribute.
func StrokeWidth[T Value](value T) mx.Attrib { return Attrib("stroke-width", value) }

// StrokeOpacity maps the SVG stroke-opacity attribute.
func StrokeOpacity[T Value](value T) mx.Attrib { return Attrib("stroke-opacity", value) }

// StrokeLineCap maps the SVG stroke-linecap attribute.
func StrokeLineCap[T Value](value T) mx.Attrib { return Attrib("stroke-linecap", value) }

// StrokeLineJoin maps the SVG stroke-linejoin attribute.
func StrokeLineJoin[T Value](value T) mx.Attrib { return Attrib("stroke-linejoin", value) }

// StrokeDashArray maps the SVG stroke-dasharray attribute.
func StrokeDashArray[T Value](value T) mx.Attrib { return Attrib("stroke-dasharray", value) }

// StrokeDashOffset maps the SVG stroke-dashoffset attribute.
func StrokeDashOffset[T Value](value T) mx.Attrib { return Attrib("stroke-dashoffset", value) }

// StrokeMiterLimit maps the SVG stroke-miterlimit attribute.
func StrokeMiterLimit[T Value](value T) mx.Attrib { return Attrib("stroke-miterlimit", value) }

// General presentation

// Opacity maps the SVG opacity attribute.
func Opacity[T Value](value T) mx.Attrib { return Attrib("opacity", value) }

// Color maps the SVG color attribute.
func Color[T Value](value T) mx.Attrib { return Attrib("color", value) }

// Cursor maps the SVG cursor attribute.
func Cursor[T Value](value T) mx.Attrib { return Attrib("cursor", value) }

// Display maps the SVG display attribute.
func Display[T Value](value T) mx.Attrib { return Attrib("display", value) }

// Visibility maps the SVG visibility attribute.
func Visibility[T Value](value T) mx.Attrib { return Attrib("visibility", value) }

// Overflow maps the SVG overflow attribute.
func Overflow[T Value](value T) mx.Attrib { return Attrib("overflow", value) }

// PointerEvents maps the SVG pointer-events attribute.
func PointerEvents[T Value](value T) mx.Attrib { return Attrib("pointer-events", value) }

// Transform maps the SVG transform attribute.
func Transform[T Value](value T) mx.Attrib { return Attrib("transform", value) }

// TransformOrigin maps the SVG transform-origin attribute.
func TransformOrigin[T Value](value T) mx.Attrib {
	return Attrib("transform-origin", value)
}

// VectorEffect maps the SVG vector-effect attribute.
func VectorEffect[T Value](value T) mx.Attrib { return Attrib("vector-effect", value) }

// ShapeRendering maps the SVG shape-rendering attribute.
func ShapeRendering[T Value](value T) mx.Attrib { return Attrib("shape-rendering", value) }

// ImageRendering maps the SVG image-rendering attribute.
func ImageRendering[T Value](value T) mx.Attrib { return Attrib("image-rendering", value) }

// PaintOrder maps the SVG paint-order attribute.
func PaintOrder[T Value](value T) mx.Attrib { return Attrib("paint-order", value) }

// Isolation maps the SVG isolation attribute.
func Isolation[T Value](value T) mx.Attrib { return Attrib("isolation", value) }

// MixBlendMode maps the SVG mix-blend-mode attribute.
func MixBlendMode[T Value](value T) mx.Attrib { return Attrib("mix-blend-mode", value) }

// ColorInterpolation maps the SVG color-interpolation attribute.
func ColorInterpolation[T Value](value T) mx.Attrib {
	return Attrib("color-interpolation", value)
}

// ColorInterpolationFilters maps the SVG color-interpolation-filters attribute.
func ColorInterpolationFilters[T Value](value T) mx.Attrib {
	return Attrib("color-interpolation-filters", value)
}

// References to <clipPath>, <mask> and <filter> (names suffixed to avoid
// colliding with the element constructors).

// ClipPathAttr maps the SVG clip-path attribute.
func ClipPathAttr[T Value](value T) mx.Attrib { return Attrib("clip-path", value) }

// ClipRule maps the SVG clip-rule attribute.
func ClipRule[T Value](value T) mx.Attrib { return Attrib("clip-rule", value) }

// MaskAttr maps the SVG mask attribute.
func MaskAttr[T Value](value T) mx.Attrib { return Attrib("mask", value) }

// FilterAttr maps the SVG filter attribute.
func FilterAttr[T Value](value T) mx.Attrib { return Attrib("filter", value) }

// Markers

// MarkerStart maps the SVG marker-start attribute.
func MarkerStart[T Value](value T) mx.Attrib { return Attrib("marker-start", value) }

// MarkerMid maps the SVG marker-mid attribute.
func MarkerMid[T Value](value T) mx.Attrib { return Attrib("marker-mid", value) }

// MarkerEnd maps the SVG marker-end attribute.
func MarkerEnd[T Value](value T) mx.Attrib { return Attrib("marker-end", value) }

// MarkerWidth maps the SVG markerWidth attribute.
func MarkerWidth[T Value](value T) mx.Attrib { return Attrib("markerWidth", value) }

// MarkerHeight maps the SVG markerHeight attribute.
func MarkerHeight[T Value](value T) mx.Attrib {
	return Attrib("markerHeight", value)
}

// MarkerUnits maps the SVG markerUnits attribute.
func MarkerUnits[T Value](value T) mx.Attrib { return Attrib("markerUnits", value) }

// RefX maps the SVG refX attribute.
func RefX[T Value](value T) mx.Attrib { return Attrib("refX", value) }

// RefY maps the SVG refY attribute.
func RefY[T Value](value T) mx.Attrib { return Attrib("refY", value) }

// Orient maps the SVG orient attribute.
func Orient[T Value](value T) mx.Attrib { return Attrib("orient", value) }

// Text and fonts

// FontFamily maps the SVG font-family attribute.
func FontFamily[T Value](value T) mx.Attrib { return Attrib("font-family", value) }

// FontSize maps the SVG font-size attribute.
func FontSize[T Value](value T) mx.Attrib { return Attrib("font-size", value) }

// FontSizeAdjust maps the SVG font-size-adjust attribute.
func FontSizeAdjust[T Value](value T) mx.Attrib { return Attrib("font-size-adjust", value) }

// FontStretch maps the SVG font-stretch attribute.
func FontStretch[T Value](value T) mx.Attrib { return Attrib("font-stretch", value) }

// FontStyle maps the SVG font-style attribute.
func FontStyle[T Value](value T) mx.Attrib { return Attrib("font-style", value) }

// FontVariant maps the SVG font-variant attribute.
func FontVariant[T Value](value T) mx.Attrib { return Attrib("font-variant", value) }

// FontWeight maps the SVG font-weight attribute.
func FontWeight[T Value](value T) mx.Attrib { return Attrib("font-weight", value) }

// TextAnchor maps the SVG text-anchor attribute.
func TextAnchor[T Value](value T) mx.Attrib { return Attrib("text-anchor", value) }

// DominantBaseline maps the SVG dominant-baseline attribute.
func DominantBaseline[T Value](value T) mx.Attrib { return Attrib("dominant-baseline", value) }

// AlignmentBaseline maps the SVG alignment-baseline attribute.
func AlignmentBaseline[T Value](value T) mx.Attrib {
	return Attrib("alignment-baseline", value)
}

// BaselineShift maps the SVG baseline-shift attribute.
func BaselineShift[T Value](value T) mx.Attrib { return Attrib("baseline-shift", value) }

// LetterSpacing maps the SVG letter-spacing attribute.
func LetterSpacing[T Value](value T) mx.Attrib { return Attrib("letter-spacing", value) }

// WordSpacing maps the SVG word-spacing attribute.
func WordSpacing[T Value](value T) mx.Attrib { return Attrib("word-spacing", value) }

// TextDecoration maps the SVG text-decoration attribute.
func TextDecoration[T Value](value T) mx.Attrib { return Attrib("text-decoration", value) }

// WritingMode maps the SVG writing-mode attribute.
func WritingMode[T Value](value T) mx.Attrib { return Attrib("writing-mode", value) }

// Direction maps the SVG direction attribute.
func Direction[T Value](value T) mx.Attrib { return Attrib("direction", value) }

// TextLength maps the SVG textLength attribute.
func TextLength[T Value](value T) mx.Attrib { return Attrib("textLength", value) }

// LengthAdjust maps the SVG lengthAdjust attribute.
func LengthAdjust[T Value](value T) mx.Attrib { return Attrib("lengthAdjust", value) }

// StartOffset maps the SVG startOffset attribute.
func StartOffset[T Value](value T) mx.Attrib { return Attrib("startOffset", value) }

// Method maps the SVG method attribute.
func Method[T Value](value T) mx.Attrib { return Attrib("method", value) }

// Spacing maps the SVG spacing attribute.
func Spacing[T Value](value T) mx.Attrib { return Attrib("spacing", value) }

// Side maps the SVG side attribute.
func Side[T Value](value T) mx.Attrib { return Attrib("side", value) }

// Gradients and patterns

// GradientUnits maps the SVG gradientUnits attribute.
func GradientUnits[T Value](value T) mx.Attrib { return Attrib("gradientUnits", value) }

// GradientTransform maps the SVG gradientTransform attribute.
func GradientTransform[T Value](value T) mx.Attrib {
	return Attrib("gradientTransform", value)
}

// SpreadMethod maps the SVG spreadMethod attribute.
func SpreadMethod[T Value](value T) mx.Attrib { return Attrib("spreadMethod", value) }

// Offset maps the SVG offset attribute.
func Offset[T Value](value T) mx.Attrib { return Attrib("offset", value) }

// StopColor maps the SVG stop-color attribute.
func StopColor[T Value](value T) mx.Attrib { return Attrib("stop-color", value) }

// StopOpacity maps the SVG stop-opacity attribute.
func StopOpacity[T Value](value T) mx.Attrib { return Attrib("stop-opacity", value) }

// PatternUnits maps the SVG patternUnits attribute.
func PatternUnits[T Value](value T) mx.Attrib { return Attrib("patternUnits", value) }

// PatternContentUnits maps the SVG patternContentUnits attribute.
func PatternContentUnits[T Value](value T) mx.Attrib {
	return Attrib("patternContentUnits", value)
}

// PatternTransform maps the SVG patternTransform attribute.
func PatternTransform[T Value](value T) mx.Attrib {
	return Attrib("patternTransform", value)
}

// clipPath and mask units

// ClipPathUnits maps the SVG clipPathUnits attribute.
func ClipPathUnits[T Value](value T) mx.Attrib { return Attrib("clipPathUnits", value) }

// MaskUnits maps the SVG maskUnits attribute.
func MaskUnits[T Value](value T) mx.Attrib { return Attrib("maskUnits", value) }

// MaskContentUnits maps the SVG maskContentUnits attribute.
func MaskContentUnits[T Value](value T) mx.Attrib {
	return Attrib("maskContentUnits", value)
}

// Filters and filter primitives

// FilterUnits maps the SVG filterUnits attribute.
func FilterUnits[T Value](value T) mx.Attrib { return Attrib("filterUnits", value) }

// PrimitiveUnits maps the SVG primitiveUnits attribute.
func PrimitiveUnits[T Value](value T) mx.Attrib { return Attrib("primitiveUnits", value) }

// In maps the SVG in attribute.
func In[T Value](value T) mx.Attrib { return Attrib("in", value) }

// In2 maps the SVG in2 attribute.
func In2[T Value](value T) mx.Attrib { return Attrib("in2", value) }

// Result maps the SVG result attribute.
func Result[T Value](value T) mx.Attrib { return Attrib("result", value) }

// StdDeviation maps the SVG stdDeviation attribute.
func StdDeviation[T Value](value T) mx.Attrib { return Attrib("stdDeviation", value) }

// Mode maps the SVG mode attribute.
func Mode[T Value](value T) mx.Attrib { return Attrib("mode", value) }

// Type maps the SVG type attribute.
func Type[T Value](value T) mx.Attrib { return Attrib("type", value) }

// Values maps the SVG values attribute.
func Values[T Value](value T) mx.Attrib { return Attrib("values", value) }

// Operator maps the SVG operator attribute.
func Operator[T Value](value T) mx.Attrib { return Attrib("operator", value) }

// K1 maps the SVG k1 attribute.
func K1[T Value](value T) mx.Attrib { return Attrib("k1", value) }

// K2 maps the SVG k2 attribute.
func K2[T Value](value T) mx.Attrib { return Attrib("k2", value) }

// K3 maps the SVG k3 attribute.
func K3[T Value](value T) mx.Attrib { return Attrib("k3", value) }

// K4 maps the SVG k4 attribute.
func K4[T Value](value T) mx.Attrib { return Attrib("k4", value) }

// Scale maps the SVG scale attribute.
func Scale[T Value](value T) mx.Attrib { return Attrib("scale", value) }

// XChannelSelector maps the SVG xChannelSelector attribute.
func XChannelSelector[T Value](value T) mx.Attrib {
	return Attrib("xChannelSelector", value)
}

// YChannelSelector maps the SVG yChannelSelector attribute.
func YChannelSelector[T Value](value T) mx.Attrib {
	return Attrib("yChannelSelector", value)
}

// BaseFrequency maps the SVG baseFrequency attribute.
func BaseFrequency[T Value](value T) mx.Attrib { return Attrib("baseFrequency", value) }

// NumOctaves maps the SVG numOctaves attribute.
func NumOctaves[T Value](value T) mx.Attrib { return Attrib("numOctaves", value) }

// Seed maps the SVG seed attribute.
func Seed[T Value](value T) mx.Attrib { return Attrib("seed", value) }

// StitchTiles maps the SVG stitchTiles attribute.
func StitchTiles[T Value](value T) mx.Attrib { return Attrib("stitchTiles", value) }

// EdgeMode maps the SVG edgeMode attribute.
func EdgeMode[T Value](value T) mx.Attrib { return Attrib("edgeMode", value) }

// KernelMatrix maps the SVG kernelMatrix attribute.
func KernelMatrix[T Value](value T) mx.Attrib { return Attrib("kernelMatrix", value) }

// Order maps the SVG order attribute.
func Order[T Value](value T) mx.Attrib { return Attrib("order", value) }

// Divisor maps the SVG divisor attribute.
func Divisor[T Value](value T) mx.Attrib { return Attrib("divisor", value) }

// Bias maps the SVG bias attribute.
func Bias[T Value](value T) mx.Attrib { return Attrib("bias", value) }

// TargetX maps the SVG targetX attribute.
func TargetX[T Value](value T) mx.Attrib { return Attrib("targetX", value) }

// TargetY maps the SVG targetY attribute.
func TargetY[T Value](value T) mx.Attrib { return Attrib("targetY", value) }

// PreserveAlpha maps the SVG preserveAlpha attribute.
func PreserveAlpha[T Value](value T) mx.Attrib { return Attrib("preserveAlpha", value) }

// KernelUnitLength maps the SVG kernelUnitLength attribute.
func KernelUnitLength[T Value](value T) mx.Attrib {
	return Attrib("kernelUnitLength", value)
}

// SurfaceScale maps the SVG surfaceScale attribute.
func SurfaceScale[T Value](value T) mx.Attrib { return Attrib("surfaceScale", value) }

// SpecularConstant maps the SVG specularConstant attribute.
func SpecularConstant[T Value](value T) mx.Attrib {
	return Attrib("specularConstant", value)
}

// SpecularExponent maps the SVG specularExponent attribute.
func SpecularExponent[T Value](value T) mx.Attrib {
	return Attrib("specularExponent", value)
}

// DiffuseConstant maps the SVG diffuseConstant attribute.
func DiffuseConstant[T Value](value T) mx.Attrib {
	return Attrib("diffuseConstant", value)
}

// Azimuth maps the SVG azimuth attribute.
func Azimuth[T Value](value T) mx.Attrib { return Attrib("azimuth", value) }

// Elevation maps the SVG elevation attribute.
func Elevation[T Value](value T) mx.Attrib { return Attrib("elevation", value) }

// PointsAtX maps the SVG pointsAtX attribute.
func PointsAtX[T Value](value T) mx.Attrib { return Attrib("pointsAtX", value) }

// PointsAtY maps the SVG pointsAtY attribute.
func PointsAtY[T Value](value T) mx.Attrib { return Attrib("pointsAtY", value) }

// PointsAtZ maps the SVG pointsAtZ attribute.
func PointsAtZ[T Value](value T) mx.Attrib { return Attrib("pointsAtZ", value) }

// LimitingConeAngle maps the SVG limitingConeAngle attribute.
func LimitingConeAngle[T Value](value T) mx.Attrib {
	return Attrib("limitingConeAngle", value)
}

// Radius maps the SVG radius attribute.
func Radius[T Value](value T) mx.Attrib { return Attrib("radius", value) }

// TableValues maps the SVG tableValues attribute.
func TableValues[T Value](value T) mx.Attrib { return Attrib("tableValues", value) }

// Slope maps the SVG slope attribute.
func Slope[T Value](value T) mx.Attrib { return Attrib("slope", value) }

// Intercept maps the SVG intercept attribute.
func Intercept[T Value](value T) mx.Attrib { return Attrib("intercept", value) }

// Amplitude maps the SVG amplitude attribute.
func Amplitude[T Value](value T) mx.Attrib { return Attrib("amplitude", value) }

// Exponent maps the SVG exponent attribute.
func Exponent[T Value](value T) mx.Attrib { return Attrib("exponent", value) }

// FloodColor maps the SVG flood-color attribute.
func FloodColor[T Value](value T) mx.Attrib { return Attrib("flood-color", value) }

// FloodOpacity maps the SVG flood-opacity attribute.
func FloodOpacity[T Value](value T) mx.Attrib {
	return Attrib("flood-opacity", value)
}

// LightingColor maps the SVG lighting-color attribute.
func LightingColor[T Value](value T) mx.Attrib {
	return Attrib("lighting-color", value)
}

// Animation

// AttributeName maps the SVG attributeName attribute.
func AttributeName[T Value](value T) mx.Attrib { return Attrib("attributeName", value) }

// AttributeType maps the SVG attributeType attribute.
func AttributeType[T Value](value T) mx.Attrib { return Attrib("attributeType", value) }

// Begin maps the SVG begin attribute.
func Begin[T Value](value T) mx.Attrib { return Attrib("begin", value) }

// End maps the SVG end attribute.
func End[T Value](value T) mx.Attrib { return Attrib("end", value) }

// Dur maps the SVG dur attribute.
func Dur[T Value](value T) mx.Attrib { return Attrib("dur", value) }

// From maps the SVG from attribute.
func From[T Value](value T) mx.Attrib { return Attrib("from", value) }

// To maps the SVG to attribute.
func To[T Value](value T) mx.Attrib { return Attrib("to", value) }

// By maps the SVG by attribute.
func By[T Value](value T) mx.Attrib { return Attrib("by", value) }

// RepeatCount maps the SVG repeatCount attribute.
func RepeatCount[T Value](value T) mx.Attrib { return Attrib("repeatCount", value) }

// RepeatDur maps the SVG repeatDur attribute.
func RepeatDur[T Value](value T) mx.Attrib { return Attrib("repeatDur", value) }

// CalcMode maps the SVG calcMode attribute.
func CalcMode[T Value](value T) mx.Attrib { return Attrib("calcMode", value) }

// KeyTimes maps the SVG keyTimes attribute.
func KeyTimes[T Value](value T) mx.Attrib { return Attrib("keyTimes", value) }

// KeySplines maps the SVG keySplines attribute.
func KeySplines[T Value](value T) mx.Attrib { return Attrib("keySplines", value) }

// KeyPoints maps the SVG keyPoints attribute.
func KeyPoints[T Value](value T) mx.Attrib { return Attrib("keyPoints", value) }

// Additive maps the SVG additive attribute.
func Additive[T Value](value T) mx.Attrib { return Attrib("additive", value) }

// Accumulate maps the SVG accumulate attribute.
func Accumulate[T Value](value T) mx.Attrib { return Attrib("accumulate", value) }

// Restart maps the SVG restart attribute.
func Restart[T Value](value T) mx.Attrib { return Attrib("restart", value) }

// Min maps the SVG min attribute.
func Min[T Value](value T) mx.Attrib { return Attrib("min", value) }

// Max maps the SVG max attribute.
func Max[T Value](value T) mx.Attrib { return Attrib("max", value) }

// Origin maps the SVG origin attribute.
func Origin[T Value](value T) mx.Attrib { return Attrib("origin", value) }

// PathAttr sets the "path" attribute (e.g. on <animateMotion>); suffixed to
// avoid colliding with the Path element constructor.
func PathAttr[T Value](value T) mx.Attrib { return Attrib("path", value) }

// Conditional processing

// SystemLanguage maps the SVG systemLanguage attribute.
func SystemLanguage[T Value](value T) mx.Attrib {
	return Attrib("systemLanguage", value)
}

// RequiredExtensions maps the SVG requiredExtensions attribute.
func RequiredExtensions[T Value](value T) mx.Attrib {
	return Attrib("requiredExtensions", value)
}

// RequiredFeatures maps the SVG requiredFeatures attribute.
func RequiredFeatures[T Value](value T) mx.Attrib {
	return Attrib("requiredFeatures", value)
}
