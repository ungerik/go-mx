//go:generate go -C ../tools tool go-enum ../svg/$GOFILE

// This file defines the SVG attributes whose value is a strict, closed set of
// keyword tokens as enum types. Each type is a string that implements mx.Attrib,
// so the typed constants are used directly as element attributes, e.g.
// svg.Path(svg.D("…"), svg.FillRuleEvenodd). A conversion such as
// FillRule("nonzero") also works for dynamic values. AttribValue returns the
// value together with the result of the generated Validate method, so rendering
// an element that holds an invalid enum value fails with a descriptive error.
//
// The go:generate directive runs the go-enum tool pinned in the nested tools
// module (kept out of the shipped go-mx dependency tree). Run it with:
//
//	go generate ./svg/...
//
// go-enum appends the Valid, Validate, Enums, EnumStrings and String methods for
// every //#enum type; the hand-written AttribName/AttribValue methods are left
// untouched.

package svg

import (
	"context"
	"fmt"

	"github.com/ungerik/go-mx"
)

// Fill and stroke

// FillRule is the SVG fill-rule presentation attribute (an enumerated keyword).
type FillRule string //#enum

const (
	// FillRuleNonzero fills using the nonzero winding-number rule.
	FillRuleNonzero FillRule = "nonzero"
	// FillRuleEvenodd fills using the even-odd rule.
	FillRuleEvenodd FillRule = "evenodd"
)

// Valid indicates if v is any of the valid values for FillRule
func (v FillRule) Valid() bool {
	switch v {
	case
		FillRuleNonzero,
		FillRuleEvenodd:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for FillRule
func (v FillRule) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.FillRule", v)
	}
	return nil
}

// Enums returns all valid values for FillRule
func (FillRule) Enums() []FillRule {
	return []FillRule{
		FillRuleNonzero,
		FillRuleEvenodd,
	}
}

// EnumStrings returns all valid values for FillRule as strings
func (FillRule) EnumStrings() []string {
	return []string{
		"nonzero",
		"evenodd",
	}
}

// String implements the fmt.Stringer interface for FillRule
func (v FillRule) String() string {
	return string(v)
}

// AttribName returns the "fill-rule" attribute name.
func (v FillRule) AttribName() string { return "fill-rule" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v FillRule) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ClipRule is the SVG clip-rule presentation attribute (an enumerated keyword).
type ClipRule string //#enum

const (
	// ClipRuleNonzero clips using the nonzero winding-number rule.
	ClipRuleNonzero ClipRule = "nonzero"
	// ClipRuleEvenodd clips using the even-odd rule.
	ClipRuleEvenodd ClipRule = "evenodd"
)

// Valid indicates if v is any of the valid values for ClipRule
func (v ClipRule) Valid() bool {
	switch v {
	case
		ClipRuleNonzero,
		ClipRuleEvenodd:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for ClipRule
func (v ClipRule) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.ClipRule", v)
	}
	return nil
}

// Enums returns all valid values for ClipRule
func (ClipRule) Enums() []ClipRule {
	return []ClipRule{
		ClipRuleNonzero,
		ClipRuleEvenodd,
	}
}

// EnumStrings returns all valid values for ClipRule as strings
func (ClipRule) EnumStrings() []string {
	return []string{
		"nonzero",
		"evenodd",
	}
}

// String implements the fmt.Stringer interface for ClipRule
func (v ClipRule) String() string {
	return string(v)
}

// AttribName returns the "clip-rule" attribute name.
func (v ClipRule) AttribName() string { return "clip-rule" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v ClipRule) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// StrokeLineCap is the SVG stroke-linecap presentation attribute.
type StrokeLineCap string //#enum

const (
	// StrokeLineCapButt ends the stroke flush at the path endpoint with no extension.
	StrokeLineCapButt StrokeLineCap = "butt"
	// StrokeLineCapRound ends the stroke with a half-circle extending past the endpoint.
	StrokeLineCapRound StrokeLineCap = "round"
	// StrokeLineCapSquare ends the stroke with a square extending half the stroke width past the endpoint.
	StrokeLineCapSquare StrokeLineCap = "square"
)

// Valid indicates if v is any of the valid values for StrokeLineCap
func (v StrokeLineCap) Valid() bool {
	switch v {
	case
		StrokeLineCapButt,
		StrokeLineCapRound,
		StrokeLineCapSquare:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for StrokeLineCap
func (v StrokeLineCap) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.StrokeLineCap", v)
	}
	return nil
}

// Enums returns all valid values for StrokeLineCap
func (StrokeLineCap) Enums() []StrokeLineCap {
	return []StrokeLineCap{
		StrokeLineCapButt,
		StrokeLineCapRound,
		StrokeLineCapSquare,
	}
}

// EnumStrings returns all valid values for StrokeLineCap as strings
func (StrokeLineCap) EnumStrings() []string {
	return []string{
		"butt",
		"round",
		"square",
	}
}

// String implements the fmt.Stringer interface for StrokeLineCap
func (v StrokeLineCap) String() string {
	return string(v)
}

// AttribName returns the "stroke-linecap" attribute name.
func (v StrokeLineCap) AttribName() string { return "stroke-linecap" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v StrokeLineCap) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// StrokeLineJoin is the SVG stroke-linejoin presentation attribute.
type StrokeLineJoin string //#enum

const (
	// StrokeLineJoinArcs joins path segments with an arc that continues the stroke outlines.
	StrokeLineJoinArcs StrokeLineJoin = "arcs"
	// StrokeLineJoinBevel joins path segments with a flattened (beveled) corner.
	StrokeLineJoinBevel StrokeLineJoin = "bevel"
	// StrokeLineJoinMiter joins path segments with a sharp mitered corner.
	StrokeLineJoinMiter StrokeLineJoin = "miter"
	// StrokeLineJoinMiterClip mitres the corner but clips it at the miter limit instead of falling back to bevel.
	StrokeLineJoinMiterClip StrokeLineJoin = "miter-clip"
	// StrokeLineJoinRound joins path segments with a rounded corner.
	StrokeLineJoinRound StrokeLineJoin = "round"
)

// Valid indicates if v is any of the valid values for StrokeLineJoin
func (v StrokeLineJoin) Valid() bool {
	switch v {
	case
		StrokeLineJoinArcs,
		StrokeLineJoinBevel,
		StrokeLineJoinMiter,
		StrokeLineJoinMiterClip,
		StrokeLineJoinRound:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for StrokeLineJoin
func (v StrokeLineJoin) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.StrokeLineJoin", v)
	}
	return nil
}

// Enums returns all valid values for StrokeLineJoin
func (StrokeLineJoin) Enums() []StrokeLineJoin {
	return []StrokeLineJoin{
		StrokeLineJoinArcs,
		StrokeLineJoinBevel,
		StrokeLineJoinMiter,
		StrokeLineJoinMiterClip,
		StrokeLineJoinRound,
	}
}

// EnumStrings returns all valid values for StrokeLineJoin as strings
func (StrokeLineJoin) EnumStrings() []string {
	return []string{
		"arcs",
		"bevel",
		"miter",
		"miter-clip",
		"round",
	}
}

// String implements the fmt.Stringer interface for StrokeLineJoin
func (v StrokeLineJoin) String() string {
	return string(v)
}

// AttribName returns the "stroke-linejoin" attribute name.
func (v StrokeLineJoin) AttribName() string { return "stroke-linejoin" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v StrokeLineJoin) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// General presentation

// Visibility is the SVG visibility presentation attribute.
type Visibility string //#enum

const (
	// VisibilityVisible renders the element.
	VisibilityVisible Visibility = "visible"
	// VisibilityHidden hides the element while keeping its layout box.
	VisibilityHidden Visibility = "hidden"
	// VisibilityCollapse hides the element (treated like hidden outside of table contexts).
	VisibilityCollapse Visibility = "collapse"
)

// Valid indicates if v is any of the valid values for Visibility
func (v Visibility) Valid() bool {
	switch v {
	case
		VisibilityVisible,
		VisibilityHidden,
		VisibilityCollapse:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Visibility
func (v Visibility) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Visibility", v)
	}
	return nil
}

// Enums returns all valid values for Visibility
func (Visibility) Enums() []Visibility {
	return []Visibility{
		VisibilityVisible,
		VisibilityHidden,
		VisibilityCollapse,
	}
}

// EnumStrings returns all valid values for Visibility as strings
func (Visibility) EnumStrings() []string {
	return []string{
		"visible",
		"hidden",
		"collapse",
	}
}

// String implements the fmt.Stringer interface for Visibility
func (v Visibility) String() string {
	return string(v)
}

// AttribName returns the "visibility" attribute name.
func (v Visibility) AttribName() string { return "visibility" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Visibility) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Overflow is the SVG overflow presentation attribute.
type Overflow string //#enum

const (
	// OverflowVisible lets content overflow the element's bounds without clipping.
	OverflowVisible Overflow = "visible"
	// OverflowHidden clips content to the element's bounds.
	OverflowHidden Overflow = "hidden"
	// OverflowScroll clips content to the element's bounds (a scroll mechanism may be provided).
	OverflowScroll Overflow = "scroll"
	// OverflowAuto lets the user agent decide how to handle overflow (typically visible for SVG).
	OverflowAuto Overflow = "auto"
)

// Valid indicates if v is any of the valid values for Overflow
func (v Overflow) Valid() bool {
	switch v {
	case
		OverflowVisible,
		OverflowHidden,
		OverflowScroll,
		OverflowAuto:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Overflow
func (v Overflow) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Overflow", v)
	}
	return nil
}

// Enums returns all valid values for Overflow
func (Overflow) Enums() []Overflow {
	return []Overflow{
		OverflowVisible,
		OverflowHidden,
		OverflowScroll,
		OverflowAuto,
	}
}

// EnumStrings returns all valid values for Overflow as strings
func (Overflow) EnumStrings() []string {
	return []string{
		"visible",
		"hidden",
		"scroll",
		"auto",
	}
}

// String implements the fmt.Stringer interface for Overflow
func (v Overflow) String() string {
	return string(v)
}

// AttribName returns the "overflow" attribute name.
func (v Overflow) AttribName() string { return "overflow" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Overflow) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// PointerEvents is the SVG pointer-events presentation attribute.
type PointerEvents string //#enum

const (
	// PointerEventsBoundingBox makes the element's bounding box capture pointer events.
	PointerEventsBoundingBox PointerEvents = "bounding-box"
	// PointerEventsVisiblePainted captures pointer events on painted, visible fill and stroke areas.
	PointerEventsVisiblePainted PointerEvents = "visiblePainted"
	// PointerEventsVisibleFill captures pointer events on the visible fill area regardless of paint.
	PointerEventsVisibleFill PointerEvents = "visibleFill"
	// PointerEventsVisibleStroke captures pointer events on the visible stroke area regardless of paint.
	PointerEventsVisibleStroke PointerEvents = "visibleStroke"
	// PointerEventsVisible captures pointer events on visible fill and stroke areas regardless of paint.
	PointerEventsVisible PointerEvents = "visible"
	// PointerEventsPainted captures pointer events on painted fill and stroke areas regardless of visibility.
	PointerEventsPainted PointerEvents = "painted"
	// PointerEventsFill captures pointer events on the fill area regardless of paint or visibility.
	PointerEventsFill PointerEvents = "fill"
	// PointerEventsStroke captures pointer events on the stroke area regardless of paint or visibility.
	PointerEventsStroke PointerEvents = "stroke"
	// PointerEventsAll captures pointer events anywhere within the element regardless of paint or visibility.
	PointerEventsAll PointerEvents = "all"
	// PointerEventsNone makes the element ignore pointer events.
	PointerEventsNone PointerEvents = "none"
)

// Valid indicates if v is any of the valid values for PointerEvents
func (v PointerEvents) Valid() bool {
	switch v {
	case
		PointerEventsBoundingBox,
		PointerEventsVisiblePainted,
		PointerEventsVisibleFill,
		PointerEventsVisibleStroke,
		PointerEventsVisible,
		PointerEventsPainted,
		PointerEventsFill,
		PointerEventsStroke,
		PointerEventsAll,
		PointerEventsNone:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for PointerEvents
func (v PointerEvents) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.PointerEvents", v)
	}
	return nil
}

// Enums returns all valid values for PointerEvents
func (PointerEvents) Enums() []PointerEvents {
	return []PointerEvents{
		PointerEventsBoundingBox,
		PointerEventsVisiblePainted,
		PointerEventsVisibleFill,
		PointerEventsVisibleStroke,
		PointerEventsVisible,
		PointerEventsPainted,
		PointerEventsFill,
		PointerEventsStroke,
		PointerEventsAll,
		PointerEventsNone,
	}
}

// EnumStrings returns all valid values for PointerEvents as strings
func (PointerEvents) EnumStrings() []string {
	return []string{
		"bounding-box",
		"visiblePainted",
		"visibleFill",
		"visibleStroke",
		"visible",
		"painted",
		"fill",
		"stroke",
		"all",
		"none",
	}
}

// String implements the fmt.Stringer interface for PointerEvents
func (v PointerEvents) String() string {
	return string(v)
}

// AttribName returns the "pointer-events" attribute name.
func (v PointerEvents) AttribName() string { return "pointer-events" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v PointerEvents) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ShapeRendering is the SVG shape-rendering presentation attribute.
type ShapeRendering string //#enum

const (
	// ShapeRenderingAuto lets the user agent balance speed, edge crispness and geometric precision.
	ShapeRenderingAuto ShapeRendering = "auto"
	// ShapeRenderingOptimizeSpeed favors rendering speed, disabling anti-aliasing.
	ShapeRenderingOptimizeSpeed ShapeRendering = "optimizeSpeed"
	// ShapeRenderingCrispEdges favors sharp contrast edges over geometric accuracy.
	ShapeRenderingCrispEdges ShapeRendering = "crispEdges"
	// ShapeRenderingGeometricPrecision favors geometric accuracy over speed and crisp edges.
	ShapeRenderingGeometricPrecision ShapeRendering = "geometricPrecision"
)

// Valid indicates if v is any of the valid values for ShapeRendering
func (v ShapeRendering) Valid() bool {
	switch v {
	case
		ShapeRenderingAuto,
		ShapeRenderingOptimizeSpeed,
		ShapeRenderingCrispEdges,
		ShapeRenderingGeometricPrecision:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for ShapeRendering
func (v ShapeRendering) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.ShapeRendering", v)
	}
	return nil
}

// Enums returns all valid values for ShapeRendering
func (ShapeRendering) Enums() []ShapeRendering {
	return []ShapeRendering{
		ShapeRenderingAuto,
		ShapeRenderingOptimizeSpeed,
		ShapeRenderingCrispEdges,
		ShapeRenderingGeometricPrecision,
	}
}

// EnumStrings returns all valid values for ShapeRendering as strings
func (ShapeRendering) EnumStrings() []string {
	return []string{
		"auto",
		"optimizeSpeed",
		"crispEdges",
		"geometricPrecision",
	}
}

// String implements the fmt.Stringer interface for ShapeRendering
func (v ShapeRendering) String() string {
	return string(v)
}

// AttribName returns the "shape-rendering" attribute name.
func (v ShapeRendering) AttribName() string { return "shape-rendering" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v ShapeRendering) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// VectorEffect is the SVG vector-effect presentation attribute.
type VectorEffect string //#enum

const (
	// VectorEffectNone applies no vector effect.
	VectorEffectNone VectorEffect = "none"
	// VectorEffectNonScalingStroke keeps the stroke width constant regardless of transforms or zoom.
	VectorEffectNonScalingStroke VectorEffect = "non-scaling-stroke"
	// VectorEffectNonScalingSize keeps the element's size constant regardless of transforms or zoom.
	VectorEffectNonScalingSize VectorEffect = "non-scaling-size"
	// VectorEffectNonRotation keeps the element unrotated regardless of transforms.
	VectorEffectNonRotation VectorEffect = "non-rotation"
	// VectorEffectFixedPosition keeps the element at a fixed position regardless of transforms.
	VectorEffectFixedPosition VectorEffect = "fixed-position"
)

// Valid indicates if v is any of the valid values for VectorEffect
func (v VectorEffect) Valid() bool {
	switch v {
	case
		VectorEffectNone,
		VectorEffectNonScalingStroke,
		VectorEffectNonScalingSize,
		VectorEffectNonRotation,
		VectorEffectFixedPosition:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for VectorEffect
func (v VectorEffect) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.VectorEffect", v)
	}
	return nil
}

// Enums returns all valid values for VectorEffect
func (VectorEffect) Enums() []VectorEffect {
	return []VectorEffect{
		VectorEffectNone,
		VectorEffectNonScalingStroke,
		VectorEffectNonScalingSize,
		VectorEffectNonRotation,
		VectorEffectFixedPosition,
	}
}

// EnumStrings returns all valid values for VectorEffect as strings
func (VectorEffect) EnumStrings() []string {
	return []string{
		"none",
		"non-scaling-stroke",
		"non-scaling-size",
		"non-rotation",
		"fixed-position",
	}
}

// String implements the fmt.Stringer interface for VectorEffect
func (v VectorEffect) String() string {
	return string(v)
}

// AttribName returns the "vector-effect" attribute name.
func (v VectorEffect) AttribName() string { return "vector-effect" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v VectorEffect) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Isolation is the SVG isolation presentation attribute.
type Isolation string //#enum

const (
	// IsolationAuto isolates the element only when a property forcing a stacking context requires it.
	IsolationAuto Isolation = "auto"
	// IsolationIsolate creates a new stacking context so blend modes do not reach below it.
	IsolationIsolate Isolation = "isolate"
)

// Valid indicates if v is any of the valid values for Isolation
func (v Isolation) Valid() bool {
	switch v {
	case
		IsolationAuto,
		IsolationIsolate:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Isolation
func (v Isolation) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Isolation", v)
	}
	return nil
}

// Enums returns all valid values for Isolation
func (Isolation) Enums() []Isolation {
	return []Isolation{
		IsolationAuto,
		IsolationIsolate,
	}
}

// EnumStrings returns all valid values for Isolation as strings
func (Isolation) EnumStrings() []string {
	return []string{
		"auto",
		"isolate",
	}
}

// String implements the fmt.Stringer interface for Isolation
func (v Isolation) String() string {
	return string(v)
}

// AttribName returns the "isolation" attribute name.
func (v Isolation) AttribName() string { return "isolation" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Isolation) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ColorInterpolation is the SVG color-interpolation presentation attribute.
type ColorInterpolation string //#enum

const (
	// ColorInterpolationAuto lets the user agent choose the color space for interpolation.
	ColorInterpolationAuto ColorInterpolation = "auto"
	// ColorInterpolationSRGB interpolates colors in the sRGB color space.
	ColorInterpolationSRGB ColorInterpolation = "sRGB"
	// ColorInterpolationLinearRGB interpolates colors in the linearized RGB color space.
	ColorInterpolationLinearRGB ColorInterpolation = "linearRGB"
)

// Valid indicates if v is any of the valid values for ColorInterpolation
func (v ColorInterpolation) Valid() bool {
	switch v {
	case
		ColorInterpolationAuto,
		ColorInterpolationSRGB,
		ColorInterpolationLinearRGB:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for ColorInterpolation
func (v ColorInterpolation) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.ColorInterpolation", v)
	}
	return nil
}

// Enums returns all valid values for ColorInterpolation
func (ColorInterpolation) Enums() []ColorInterpolation {
	return []ColorInterpolation{
		ColorInterpolationAuto,
		ColorInterpolationSRGB,
		ColorInterpolationLinearRGB,
	}
}

// EnumStrings returns all valid values for ColorInterpolation as strings
func (ColorInterpolation) EnumStrings() []string {
	return []string{
		"auto",
		"sRGB",
		"linearRGB",
	}
}

// String implements the fmt.Stringer interface for ColorInterpolation
func (v ColorInterpolation) String() string {
	return string(v)
}

// AttribName returns the "color-interpolation" attribute name.
func (v ColorInterpolation) AttribName() string { return "color-interpolation" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v ColorInterpolation) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// ColorInterpolationFilters is the SVG color-interpolation-filters presentation
// attribute.
type ColorInterpolationFilters string //#enum

const (
	// ColorInterpolationFiltersAuto lets the user agent choose the color space for filter interpolation.
	ColorInterpolationFiltersAuto ColorInterpolationFilters = "auto"
	// ColorInterpolationFiltersSRGB interpolates filter colors in the sRGB color space.
	ColorInterpolationFiltersSRGB ColorInterpolationFilters = "sRGB"
	// ColorInterpolationFiltersLinearRGB interpolates filter colors in the linearized RGB color space (the filter default).
	ColorInterpolationFiltersLinearRGB ColorInterpolationFilters = "linearRGB"
)

// Valid indicates if v is any of the valid values for ColorInterpolationFilters
func (v ColorInterpolationFilters) Valid() bool {
	switch v {
	case
		ColorInterpolationFiltersAuto,
		ColorInterpolationFiltersSRGB,
		ColorInterpolationFiltersLinearRGB:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for ColorInterpolationFilters
func (v ColorInterpolationFilters) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.ColorInterpolationFilters", v)
	}
	return nil
}

// Enums returns all valid values for ColorInterpolationFilters
func (ColorInterpolationFilters) Enums() []ColorInterpolationFilters {
	return []ColorInterpolationFilters{
		ColorInterpolationFiltersAuto,
		ColorInterpolationFiltersSRGB,
		ColorInterpolationFiltersLinearRGB,
	}
}

// EnumStrings returns all valid values for ColorInterpolationFilters as strings
func (ColorInterpolationFilters) EnumStrings() []string {
	return []string{
		"auto",
		"sRGB",
		"linearRGB",
	}
}

// String implements the fmt.Stringer interface for ColorInterpolationFilters
func (v ColorInterpolationFilters) String() string {
	return string(v)
}

// AttribName returns the "color-interpolation-filters" attribute name.
func (v ColorInterpolationFilters) AttribName() string { return "color-interpolation-filters" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v ColorInterpolationFilters) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// Blend modes

// MixBlendMode is the SVG mix-blend-mode presentation attribute (a <blend-mode>).
type MixBlendMode string //#enum

const (
	// MixBlendModeNormal paints the element on top with no blending.
	MixBlendModeNormal MixBlendMode = "normal"
	// MixBlendModeMultiply multiplies the element's colors with the backdrop, darkening the result.
	MixBlendModeMultiply MixBlendMode = "multiply"
	// MixBlendModeScreen inverts, multiplies and inverts again, lightening the result.
	MixBlendModeScreen MixBlendMode = "screen"
	// MixBlendModeOverlay combines multiply and screen depending on the backdrop.
	MixBlendModeOverlay MixBlendMode = "overlay"
	// MixBlendModeDarken keeps the darker of the element and backdrop colors per channel.
	MixBlendModeDarken MixBlendMode = "darken"
	// MixBlendModeLighten keeps the lighter of the element and backdrop colors per channel.
	MixBlendModeLighten MixBlendMode = "lighten"
	// MixBlendModeColorDodge brightens the backdrop to reflect the element's color.
	MixBlendModeColorDodge MixBlendMode = "color-dodge"
	// MixBlendModeColorBurn darkens the backdrop to reflect the element's color.
	MixBlendModeColorBurn MixBlendMode = "color-burn"
	// MixBlendModeHardLight applies multiply or screen depending on the element's color.
	MixBlendModeHardLight MixBlendMode = "hard-light"
	// MixBlendModeSoftLight applies a softer dodge or burn depending on the element's color.
	MixBlendModeSoftLight MixBlendMode = "soft-light"
	// MixBlendModeDifference subtracts the darker color from the lighter one per channel.
	MixBlendModeDifference MixBlendMode = "difference"
	// MixBlendModeExclusion is like difference but with lower contrast.
	MixBlendModeExclusion MixBlendMode = "exclusion"
	// MixBlendModeHue keeps the element's hue with the backdrop's saturation and luminosity.
	MixBlendModeHue MixBlendMode = "hue"
	// MixBlendModeSaturation keeps the element's saturation with the backdrop's hue and luminosity.
	MixBlendModeSaturation MixBlendMode = "saturation"
	// MixBlendModeColor keeps the element's hue and saturation with the backdrop's luminosity.
	MixBlendModeColor MixBlendMode = "color"
	// MixBlendModeLuminosity keeps the element's luminosity with the backdrop's hue and saturation.
	MixBlendModeLuminosity MixBlendMode = "luminosity"
)

// Valid indicates if v is any of the valid values for MixBlendMode
func (v MixBlendMode) Valid() bool {
	switch v {
	case
		MixBlendModeNormal,
		MixBlendModeMultiply,
		MixBlendModeScreen,
		MixBlendModeOverlay,
		MixBlendModeDarken,
		MixBlendModeLighten,
		MixBlendModeColorDodge,
		MixBlendModeColorBurn,
		MixBlendModeHardLight,
		MixBlendModeSoftLight,
		MixBlendModeDifference,
		MixBlendModeExclusion,
		MixBlendModeHue,
		MixBlendModeSaturation,
		MixBlendModeColor,
		MixBlendModeLuminosity:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for MixBlendMode
func (v MixBlendMode) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.MixBlendMode", v)
	}
	return nil
}

// Enums returns all valid values for MixBlendMode
func (MixBlendMode) Enums() []MixBlendMode {
	return []MixBlendMode{
		MixBlendModeNormal,
		MixBlendModeMultiply,
		MixBlendModeScreen,
		MixBlendModeOverlay,
		MixBlendModeDarken,
		MixBlendModeLighten,
		MixBlendModeColorDodge,
		MixBlendModeColorBurn,
		MixBlendModeHardLight,
		MixBlendModeSoftLight,
		MixBlendModeDifference,
		MixBlendModeExclusion,
		MixBlendModeHue,
		MixBlendModeSaturation,
		MixBlendModeColor,
		MixBlendModeLuminosity,
	}
}

// EnumStrings returns all valid values for MixBlendMode as strings
func (MixBlendMode) EnumStrings() []string {
	return []string{
		"normal",
		"multiply",
		"screen",
		"overlay",
		"darken",
		"lighten",
		"color-dodge",
		"color-burn",
		"hard-light",
		"soft-light",
		"difference",
		"exclusion",
		"hue",
		"saturation",
		"color",
		"luminosity",
	}
}

// String implements the fmt.Stringer interface for MixBlendMode
func (v MixBlendMode) String() string {
	return string(v)
}

// AttribName returns the "mix-blend-mode" attribute name.
func (v MixBlendMode) AttribName() string { return "mix-blend-mode" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v MixBlendMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Mode is the SVG mode attribute of <feBlend> (a <blend-mode>).
type Mode string //#enum

const (
	// ModeNormal paints the input on top with no blending.
	ModeNormal Mode = "normal"
	// ModeMultiply multiplies the input colors with the backdrop, darkening the result.
	ModeMultiply Mode = "multiply"
	// ModeScreen inverts, multiplies and inverts again, lightening the result.
	ModeScreen Mode = "screen"
	// ModeOverlay combines multiply and screen depending on the backdrop.
	ModeOverlay Mode = "overlay"
	// ModeDarken keeps the darker of the input and backdrop colors per channel.
	ModeDarken Mode = "darken"
	// ModeLighten keeps the lighter of the input and backdrop colors per channel.
	ModeLighten Mode = "lighten"
	// ModeColorDodge brightens the backdrop to reflect the input color.
	ModeColorDodge Mode = "color-dodge"
	// ModeColorBurn darkens the backdrop to reflect the input color.
	ModeColorBurn Mode = "color-burn"
	// ModeHardLight applies multiply or screen depending on the input color.
	ModeHardLight Mode = "hard-light"
	// ModeSoftLight applies a softer dodge or burn depending on the input color.
	ModeSoftLight Mode = "soft-light"
	// ModeDifference subtracts the darker color from the lighter one per channel.
	ModeDifference Mode = "difference"
	// ModeExclusion is like difference but with lower contrast.
	ModeExclusion Mode = "exclusion"
	// ModeHue keeps the input's hue with the backdrop's saturation and luminosity.
	ModeHue Mode = "hue"
	// ModeSaturation keeps the input's saturation with the backdrop's hue and luminosity.
	ModeSaturation Mode = "saturation"
	// ModeColor keeps the input's hue and saturation with the backdrop's luminosity.
	ModeColor Mode = "color"
	// ModeLuminosity keeps the input's luminosity with the backdrop's hue and saturation.
	ModeLuminosity Mode = "luminosity"
)

// Valid indicates if v is any of the valid values for Mode
func (v Mode) Valid() bool {
	switch v {
	case
		ModeNormal,
		ModeMultiply,
		ModeScreen,
		ModeOverlay,
		ModeDarken,
		ModeLighten,
		ModeColorDodge,
		ModeColorBurn,
		ModeHardLight,
		ModeSoftLight,
		ModeDifference,
		ModeExclusion,
		ModeHue,
		ModeSaturation,
		ModeColor,
		ModeLuminosity:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Mode
func (v Mode) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Mode", v)
	}
	return nil
}

// Enums returns all valid values for Mode
func (Mode) Enums() []Mode {
	return []Mode{
		ModeNormal,
		ModeMultiply,
		ModeScreen,
		ModeOverlay,
		ModeDarken,
		ModeLighten,
		ModeColorDodge,
		ModeColorBurn,
		ModeHardLight,
		ModeSoftLight,
		ModeDifference,
		ModeExclusion,
		ModeHue,
		ModeSaturation,
		ModeColor,
		ModeLuminosity,
	}
}

// EnumStrings returns all valid values for Mode as strings
func (Mode) EnumStrings() []string {
	return []string{
		"normal",
		"multiply",
		"screen",
		"overlay",
		"darken",
		"lighten",
		"color-dodge",
		"color-burn",
		"hard-light",
		"soft-light",
		"difference",
		"exclusion",
		"hue",
		"saturation",
		"color",
		"luminosity",
	}
}

// String implements the fmt.Stringer interface for Mode
func (v Mode) String() string {
	return string(v)
}

// AttribName returns the "mode" attribute name.
func (v Mode) AttribName() string { return "mode" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Mode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Text and fonts

// FontStyle is the SVG font-style presentation attribute.
type FontStyle string //#enum

const (
	// FontStyleNormal selects the upright (roman) face.
	FontStyleNormal FontStyle = "normal"
	// FontStyleItalic selects the italic face.
	FontStyleItalic FontStyle = "italic"
	// FontStyleOblique selects the oblique (slanted) face.
	FontStyleOblique FontStyle = "oblique"
)

// Valid indicates if v is any of the valid values for FontStyle
func (v FontStyle) Valid() bool {
	switch v {
	case
		FontStyleNormal,
		FontStyleItalic,
		FontStyleOblique:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for FontStyle
func (v FontStyle) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.FontStyle", v)
	}
	return nil
}

// Enums returns all valid values for FontStyle
func (FontStyle) Enums() []FontStyle {
	return []FontStyle{
		FontStyleNormal,
		FontStyleItalic,
		FontStyleOblique,
	}
}

// EnumStrings returns all valid values for FontStyle as strings
func (FontStyle) EnumStrings() []string {
	return []string{
		"normal",
		"italic",
		"oblique",
	}
}

// String implements the fmt.Stringer interface for FontStyle
func (v FontStyle) String() string {
	return string(v)
}

// AttribName returns the "font-style" attribute name.
func (v FontStyle) AttribName() string { return "font-style" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v FontStyle) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// TextAnchor is the SVG text-anchor presentation attribute.
type TextAnchor string //#enum

const (
	// TextAnchorStart aligns text so it begins at the given position.
	TextAnchorStart TextAnchor = "start"
	// TextAnchorMiddle centers text on the given position.
	TextAnchorMiddle TextAnchor = "middle"
	// TextAnchorEnd aligns text so it ends at the given position.
	TextAnchorEnd TextAnchor = "end"
)

// Valid indicates if v is any of the valid values for TextAnchor
func (v TextAnchor) Valid() bool {
	switch v {
	case
		TextAnchorStart,
		TextAnchorMiddle,
		TextAnchorEnd:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for TextAnchor
func (v TextAnchor) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.TextAnchor", v)
	}
	return nil
}

// Enums returns all valid values for TextAnchor
func (TextAnchor) Enums() []TextAnchor {
	return []TextAnchor{
		TextAnchorStart,
		TextAnchorMiddle,
		TextAnchorEnd,
	}
}

// EnumStrings returns all valid values for TextAnchor as strings
func (TextAnchor) EnumStrings() []string {
	return []string{
		"start",
		"middle",
		"end",
	}
}

// String implements the fmt.Stringer interface for TextAnchor
func (v TextAnchor) String() string {
	return string(v)
}

// AttribName returns the "text-anchor" attribute name.
func (v TextAnchor) AttribName() string { return "text-anchor" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v TextAnchor) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// DominantBaseline is the SVG dominant-baseline presentation attribute.
type DominantBaseline string //#enum

const (
	// DominantBaselineAuto uses the dominant baseline appropriate for the script and writing mode.
	DominantBaselineAuto DominantBaseline = "auto"
	// DominantBaselineTextBottom aligns to the bottom of the text content area.
	DominantBaselineTextBottom DominantBaseline = "text-bottom"
	// DominantBaselineAlphabetic aligns to the alphabetic baseline used by Latin scripts.
	DominantBaselineAlphabetic DominantBaseline = "alphabetic"
	// DominantBaselineIdeographic aligns to the ideographic (under-side) baseline used by CJK scripts.
	DominantBaselineIdeographic DominantBaseline = "ideographic"
	// DominantBaselineMiddle aligns to the middle baseline, halfway up the x-height.
	DominantBaselineMiddle DominantBaseline = "middle"
	// DominantBaselineCentral aligns to the central baseline, midway between ascender and descender.
	DominantBaselineCentral DominantBaseline = "central"
	// DominantBaselineMathematical aligns to the mathematical baseline.
	DominantBaselineMathematical DominantBaseline = "mathematical"
	// DominantBaselineHanging aligns to the hanging baseline used by Indic scripts.
	DominantBaselineHanging DominantBaseline = "hanging"
	// DominantBaselineTextTop aligns to the top of the text content area.
	DominantBaselineTextTop DominantBaseline = "text-top"
)

// Valid indicates if v is any of the valid values for DominantBaseline
func (v DominantBaseline) Valid() bool {
	switch v {
	case
		DominantBaselineAuto,
		DominantBaselineTextBottom,
		DominantBaselineAlphabetic,
		DominantBaselineIdeographic,
		DominantBaselineMiddle,
		DominantBaselineCentral,
		DominantBaselineMathematical,
		DominantBaselineHanging,
		DominantBaselineTextTop:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for DominantBaseline
func (v DominantBaseline) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.DominantBaseline", v)
	}
	return nil
}

// Enums returns all valid values for DominantBaseline
func (DominantBaseline) Enums() []DominantBaseline {
	return []DominantBaseline{
		DominantBaselineAuto,
		DominantBaselineTextBottom,
		DominantBaselineAlphabetic,
		DominantBaselineIdeographic,
		DominantBaselineMiddle,
		DominantBaselineCentral,
		DominantBaselineMathematical,
		DominantBaselineHanging,
		DominantBaselineTextTop,
	}
}

// EnumStrings returns all valid values for DominantBaseline as strings
func (DominantBaseline) EnumStrings() []string {
	return []string{
		"auto",
		"text-bottom",
		"alphabetic",
		"ideographic",
		"middle",
		"central",
		"mathematical",
		"hanging",
		"text-top",
	}
}

// String implements the fmt.Stringer interface for DominantBaseline
func (v DominantBaseline) String() string {
	return string(v)
}

// AttribName returns the "dominant-baseline" attribute name.
func (v DominantBaseline) AttribName() string { return "dominant-baseline" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v DominantBaseline) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// AlignmentBaseline is the SVG alignment-baseline presentation attribute.
type AlignmentBaseline string //#enum

const (
	// AlignmentBaselineBaseline aligns to the dominant baseline of the parent.
	AlignmentBaselineBaseline AlignmentBaseline = "baseline"
	// AlignmentBaselineTextBottom aligns to the bottom of the text content area.
	AlignmentBaselineTextBottom AlignmentBaseline = "text-bottom"
	// AlignmentBaselineAlphabetic aligns to the alphabetic baseline used by Latin scripts.
	AlignmentBaselineAlphabetic AlignmentBaseline = "alphabetic"
	// AlignmentBaselineIdeographic aligns to the ideographic (under-side) baseline used by CJK scripts.
	AlignmentBaselineIdeographic AlignmentBaseline = "ideographic"
	// AlignmentBaselineMiddle aligns to the middle baseline, halfway up the x-height.
	AlignmentBaselineMiddle AlignmentBaseline = "middle"
	// AlignmentBaselineCentral aligns to the central baseline, midway between ascender and descender.
	AlignmentBaselineCentral AlignmentBaseline = "central"
	// AlignmentBaselineMathematical aligns to the mathematical baseline.
	AlignmentBaselineMathematical AlignmentBaseline = "mathematical"
	// AlignmentBaselineTextTop aligns to the top of the text content area.
	AlignmentBaselineTextTop AlignmentBaseline = "text-top"
)

// Valid indicates if v is any of the valid values for AlignmentBaseline
func (v AlignmentBaseline) Valid() bool {
	switch v {
	case
		AlignmentBaselineBaseline,
		AlignmentBaselineTextBottom,
		AlignmentBaselineAlphabetic,
		AlignmentBaselineIdeographic,
		AlignmentBaselineMiddle,
		AlignmentBaselineCentral,
		AlignmentBaselineMathematical,
		AlignmentBaselineTextTop:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for AlignmentBaseline
func (v AlignmentBaseline) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.AlignmentBaseline", v)
	}
	return nil
}

// Enums returns all valid values for AlignmentBaseline
func (AlignmentBaseline) Enums() []AlignmentBaseline {
	return []AlignmentBaseline{
		AlignmentBaselineBaseline,
		AlignmentBaselineTextBottom,
		AlignmentBaselineAlphabetic,
		AlignmentBaselineIdeographic,
		AlignmentBaselineMiddle,
		AlignmentBaselineCentral,
		AlignmentBaselineMathematical,
		AlignmentBaselineTextTop,
	}
}

// EnumStrings returns all valid values for AlignmentBaseline as strings
func (AlignmentBaseline) EnumStrings() []string {
	return []string{
		"baseline",
		"text-bottom",
		"alphabetic",
		"ideographic",
		"middle",
		"central",
		"mathematical",
		"text-top",
	}
}

// String implements the fmt.Stringer interface for AlignmentBaseline
func (v AlignmentBaseline) String() string {
	return string(v)
}

// AttribName returns the "alignment-baseline" attribute name.
func (v AlignmentBaseline) AttribName() string { return "alignment-baseline" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v AlignmentBaseline) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// WritingMode is the SVG writing-mode presentation attribute.
type WritingMode string //#enum

const (
	// WritingModeHorizontalTB lays out text horizontally, top to bottom.
	WritingModeHorizontalTB WritingMode = "horizontal-tb"
	// WritingModeVerticalRL lays out text vertically, right column to left.
	WritingModeVerticalRL WritingMode = "vertical-rl"
	// WritingModeVerticalLR lays out text vertically, left column to right.
	WritingModeVerticalLR WritingMode = "vertical-lr"
)

// Valid indicates if v is any of the valid values for WritingMode
func (v WritingMode) Valid() bool {
	switch v {
	case
		WritingModeHorizontalTB,
		WritingModeVerticalRL,
		WritingModeVerticalLR:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for WritingMode
func (v WritingMode) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.WritingMode", v)
	}
	return nil
}

// Enums returns all valid values for WritingMode
func (WritingMode) Enums() []WritingMode {
	return []WritingMode{
		WritingModeHorizontalTB,
		WritingModeVerticalRL,
		WritingModeVerticalLR,
	}
}

// EnumStrings returns all valid values for WritingMode as strings
func (WritingMode) EnumStrings() []string {
	return []string{
		"horizontal-tb",
		"vertical-rl",
		"vertical-lr",
	}
}

// String implements the fmt.Stringer interface for WritingMode
func (v WritingMode) String() string {
	return string(v)
}

// AttribName returns the "writing-mode" attribute name.
func (v WritingMode) AttribName() string { return "writing-mode" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v WritingMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Direction is the SVG direction presentation attribute.
type Direction string //#enum

const (
	// DirectionLTR sets the inline base direction left-to-right.
	DirectionLTR Direction = "ltr"
	// DirectionRTL sets the inline base direction right-to-left.
	DirectionRTL Direction = "rtl"
)

// Valid indicates if v is any of the valid values for Direction
func (v Direction) Valid() bool {
	switch v {
	case
		DirectionLTR,
		DirectionRTL:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Direction
func (v Direction) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Direction", v)
	}
	return nil
}

// Enums returns all valid values for Direction
func (Direction) Enums() []Direction {
	return []Direction{
		DirectionLTR,
		DirectionRTL,
	}
}

// EnumStrings returns all valid values for Direction as strings
func (Direction) EnumStrings() []string {
	return []string{
		"ltr",
		"rtl",
	}
}

// String implements the fmt.Stringer interface for Direction
func (v Direction) String() string {
	return string(v)
}

// AttribName returns the "direction" attribute name.
func (v Direction) AttribName() string { return "direction" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Direction) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// LengthAdjust is the SVG lengthAdjust attribute of <text>/<textPath>.
type LengthAdjust string //#enum

const (
	// LengthAdjustSpacing fits text to textLength by adjusting only spacing between glyphs.
	LengthAdjustSpacing LengthAdjust = "spacing"
	// LengthAdjustSpacingAndGlyphs fits text to textLength by stretching both spacing and glyphs.
	LengthAdjustSpacingAndGlyphs LengthAdjust = "spacingAndGlyphs"
)

// Valid indicates if v is any of the valid values for LengthAdjust
func (v LengthAdjust) Valid() bool {
	switch v {
	case
		LengthAdjustSpacing,
		LengthAdjustSpacingAndGlyphs:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for LengthAdjust
func (v LengthAdjust) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.LengthAdjust", v)
	}
	return nil
}

// Enums returns all valid values for LengthAdjust
func (LengthAdjust) Enums() []LengthAdjust {
	return []LengthAdjust{
		LengthAdjustSpacing,
		LengthAdjustSpacingAndGlyphs,
	}
}

// EnumStrings returns all valid values for LengthAdjust as strings
func (LengthAdjust) EnumStrings() []string {
	return []string{
		"spacing",
		"spacingAndGlyphs",
	}
}

// String implements the fmt.Stringer interface for LengthAdjust
func (v LengthAdjust) String() string {
	return string(v)
}

// AttribName returns the "lengthAdjust" attribute name.
func (v LengthAdjust) AttribName() string { return "lengthAdjust" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v LengthAdjust) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Method is the SVG method attribute of <textPath>.
type Method string //#enum

const (
	// MethodAlign renders glyphs along the path without distorting them.
	MethodAlign Method = "align"
	// MethodStretch stretches the glyph outlines to follow the path's curvature.
	MethodStretch Method = "stretch"
)

// Valid indicates if v is any of the valid values for Method
func (v Method) Valid() bool {
	switch v {
	case
		MethodAlign,
		MethodStretch:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Method
func (v Method) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Method", v)
	}
	return nil
}

// Enums returns all valid values for Method
func (Method) Enums() []Method {
	return []Method{
		MethodAlign,
		MethodStretch,
	}
}

// EnumStrings returns all valid values for Method as strings
func (Method) EnumStrings() []string {
	return []string{
		"align",
		"stretch",
	}
}

// String implements the fmt.Stringer interface for Method
func (v Method) String() string {
	return string(v)
}

// AttribName returns the "method" attribute name.
func (v Method) AttribName() string { return "method" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Method) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Spacing is the SVG spacing attribute of <textPath>.
type Spacing string //#enum

const (
	// SpacingAuto lets the user agent adjust glyph spacing to render text smoothly along the path.
	SpacingAuto Spacing = "auto"
	// SpacingExact lays out glyphs with their exact advance, ignoring path curvature.
	SpacingExact Spacing = "exact"
)

// Valid indicates if v is any of the valid values for Spacing
func (v Spacing) Valid() bool {
	switch v {
	case
		SpacingAuto,
		SpacingExact:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Spacing
func (v Spacing) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Spacing", v)
	}
	return nil
}

// Enums returns all valid values for Spacing
func (Spacing) Enums() []Spacing {
	return []Spacing{
		SpacingAuto,
		SpacingExact,
	}
}

// EnumStrings returns all valid values for Spacing as strings
func (Spacing) EnumStrings() []string {
	return []string{
		"auto",
		"exact",
	}
}

// String implements the fmt.Stringer interface for Spacing
func (v Spacing) String() string {
	return string(v)
}

// AttribName returns the "spacing" attribute name.
func (v Spacing) AttribName() string { return "spacing" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Spacing) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Side is the SVG side attribute of <textPath>.
type Side string //#enum

const (
	// SideLeft renders text on the left side of the path (the default).
	SideLeft Side = "left"
	// SideRight renders text on the right side of the path, effectively reversing it.
	SideRight Side = "right"
)

// Valid indicates if v is any of the valid values for Side
func (v Side) Valid() bool {
	switch v {
	case
		SideLeft,
		SideRight:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Side
func (v Side) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Side", v)
	}
	return nil
}

// Enums returns all valid values for Side
func (Side) Enums() []Side {
	return []Side{
		SideLeft,
		SideRight,
	}
}

// EnumStrings returns all valid values for Side as strings
func (Side) EnumStrings() []string {
	return []string{
		"left",
		"right",
	}
}

// String implements the fmt.Stringer interface for Side
func (v Side) String() string {
	return string(v)
}

// AttribName returns the "side" attribute name.
func (v Side) AttribName() string { return "side" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Side) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Units

// GradientUnits is the SVG gradientUnits attribute.
type GradientUnits string //#enum

const (
	// GradientUnitsUserSpaceOnUse interprets gradient coordinates in the current user coordinate system.
	GradientUnitsUserSpaceOnUse GradientUnits = "userSpaceOnUse"
	// GradientUnitsObjectBoundingBox interprets gradient coordinates as fractions of the referencing element's bounding box.
	GradientUnitsObjectBoundingBox GradientUnits = "objectBoundingBox"
)

// Valid indicates if v is any of the valid values for GradientUnits
func (v GradientUnits) Valid() bool {
	switch v {
	case
		GradientUnitsUserSpaceOnUse,
		GradientUnitsObjectBoundingBox:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for GradientUnits
func (v GradientUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.GradientUnits", v)
	}
	return nil
}

// Enums returns all valid values for GradientUnits
func (GradientUnits) Enums() []GradientUnits {
	return []GradientUnits{
		GradientUnitsUserSpaceOnUse,
		GradientUnitsObjectBoundingBox,
	}
}

// EnumStrings returns all valid values for GradientUnits as strings
func (GradientUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"objectBoundingBox",
	}
}

// String implements the fmt.Stringer interface for GradientUnits
func (v GradientUnits) String() string {
	return string(v)
}

// AttribName returns the "gradientUnits" attribute name.
func (v GradientUnits) AttribName() string { return "gradientUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v GradientUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// PatternUnits is the SVG patternUnits attribute.
type PatternUnits string //#enum

const (
	// PatternUnitsUserSpaceOnUse interprets the pattern's x, y, width and height in the current user coordinate system.
	PatternUnitsUserSpaceOnUse PatternUnits = "userSpaceOnUse"
	// PatternUnitsObjectBoundingBox interprets the pattern's x, y, width and height as fractions of the referencing element's bounding box.
	PatternUnitsObjectBoundingBox PatternUnits = "objectBoundingBox"
)

// Valid indicates if v is any of the valid values for PatternUnits
func (v PatternUnits) Valid() bool {
	switch v {
	case
		PatternUnitsUserSpaceOnUse,
		PatternUnitsObjectBoundingBox:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for PatternUnits
func (v PatternUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.PatternUnits", v)
	}
	return nil
}

// Enums returns all valid values for PatternUnits
func (PatternUnits) Enums() []PatternUnits {
	return []PatternUnits{
		PatternUnitsUserSpaceOnUse,
		PatternUnitsObjectBoundingBox,
	}
}

// EnumStrings returns all valid values for PatternUnits as strings
func (PatternUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"objectBoundingBox",
	}
}

// String implements the fmt.Stringer interface for PatternUnits
func (v PatternUnits) String() string {
	return string(v)
}

// AttribName returns the "patternUnits" attribute name.
func (v PatternUnits) AttribName() string { return "patternUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v PatternUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// PatternContentUnits is the SVG patternContentUnits attribute.
type PatternContentUnits string //#enum

const (
	// PatternContentUnitsUserSpaceOnUse interprets the pattern's content coordinates in the current user coordinate system.
	PatternContentUnitsUserSpaceOnUse PatternContentUnits = "userSpaceOnUse"
	// PatternContentUnitsObjectBoundingBox interprets the pattern's content coordinates as fractions of the referencing element's bounding box.
	PatternContentUnitsObjectBoundingBox PatternContentUnits = "objectBoundingBox"
)

// Valid indicates if v is any of the valid values for PatternContentUnits
func (v PatternContentUnits) Valid() bool {
	switch v {
	case
		PatternContentUnitsUserSpaceOnUse,
		PatternContentUnitsObjectBoundingBox:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for PatternContentUnits
func (v PatternContentUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.PatternContentUnits", v)
	}
	return nil
}

// Enums returns all valid values for PatternContentUnits
func (PatternContentUnits) Enums() []PatternContentUnits {
	return []PatternContentUnits{
		PatternContentUnitsUserSpaceOnUse,
		PatternContentUnitsObjectBoundingBox,
	}
}

// EnumStrings returns all valid values for PatternContentUnits as strings
func (PatternContentUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"objectBoundingBox",
	}
}

// String implements the fmt.Stringer interface for PatternContentUnits
func (v PatternContentUnits) String() string {
	return string(v)
}

// AttribName returns the "patternContentUnits" attribute name.
func (v PatternContentUnits) AttribName() string { return "patternContentUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v PatternContentUnits) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// ClipPathUnits is the SVG clipPathUnits attribute.
type ClipPathUnits string //#enum

const (
	// ClipPathUnitsUserSpaceOnUse interprets clip path coordinates in the current user coordinate system.
	ClipPathUnitsUserSpaceOnUse ClipPathUnits = "userSpaceOnUse"
	// ClipPathUnitsObjectBoundingBox interprets clip path coordinates as fractions of the referencing element's bounding box.
	ClipPathUnitsObjectBoundingBox ClipPathUnits = "objectBoundingBox"
)

// Valid indicates if v is any of the valid values for ClipPathUnits
func (v ClipPathUnits) Valid() bool {
	switch v {
	case
		ClipPathUnitsUserSpaceOnUse,
		ClipPathUnitsObjectBoundingBox:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for ClipPathUnits
func (v ClipPathUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.ClipPathUnits", v)
	}
	return nil
}

// Enums returns all valid values for ClipPathUnits
func (ClipPathUnits) Enums() []ClipPathUnits {
	return []ClipPathUnits{
		ClipPathUnitsUserSpaceOnUse,
		ClipPathUnitsObjectBoundingBox,
	}
}

// EnumStrings returns all valid values for ClipPathUnits as strings
func (ClipPathUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"objectBoundingBox",
	}
}

// String implements the fmt.Stringer interface for ClipPathUnits
func (v ClipPathUnits) String() string {
	return string(v)
}

// AttribName returns the "clipPathUnits" attribute name.
func (v ClipPathUnits) AttribName() string { return "clipPathUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v ClipPathUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// MaskUnits is the SVG maskUnits attribute.
type MaskUnits string //#enum

const (
	// MaskUnitsUserSpaceOnUse interprets the mask's x, y, width and height in the current user coordinate system.
	MaskUnitsUserSpaceOnUse MaskUnits = "userSpaceOnUse"
	// MaskUnitsObjectBoundingBox interprets the mask's x, y, width and height as fractions of the masked element's bounding box.
	MaskUnitsObjectBoundingBox MaskUnits = "objectBoundingBox"
)

// Valid indicates if v is any of the valid values for MaskUnits
func (v MaskUnits) Valid() bool {
	switch v {
	case
		MaskUnitsUserSpaceOnUse,
		MaskUnitsObjectBoundingBox:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for MaskUnits
func (v MaskUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.MaskUnits", v)
	}
	return nil
}

// Enums returns all valid values for MaskUnits
func (MaskUnits) Enums() []MaskUnits {
	return []MaskUnits{
		MaskUnitsUserSpaceOnUse,
		MaskUnitsObjectBoundingBox,
	}
}

// EnumStrings returns all valid values for MaskUnits as strings
func (MaskUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"objectBoundingBox",
	}
}

// String implements the fmt.Stringer interface for MaskUnits
func (v MaskUnits) String() string {
	return string(v)
}

// AttribName returns the "maskUnits" attribute name.
func (v MaskUnits) AttribName() string { return "maskUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v MaskUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// MaskContentUnits is the SVG maskContentUnits attribute.
type MaskContentUnits string //#enum

const (
	// MaskContentUnitsUserSpaceOnUse interprets the mask's content coordinates in the current user coordinate system.
	MaskContentUnitsUserSpaceOnUse MaskContentUnits = "userSpaceOnUse"
	// MaskContentUnitsObjectBoundingBox interprets the mask's content coordinates as fractions of the masked element's bounding box.
	MaskContentUnitsObjectBoundingBox MaskContentUnits = "objectBoundingBox"
)

// Valid indicates if v is any of the valid values for MaskContentUnits
func (v MaskContentUnits) Valid() bool {
	switch v {
	case
		MaskContentUnitsUserSpaceOnUse,
		MaskContentUnitsObjectBoundingBox:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for MaskContentUnits
func (v MaskContentUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.MaskContentUnits", v)
	}
	return nil
}

// Enums returns all valid values for MaskContentUnits
func (MaskContentUnits) Enums() []MaskContentUnits {
	return []MaskContentUnits{
		MaskContentUnitsUserSpaceOnUse,
		MaskContentUnitsObjectBoundingBox,
	}
}

// EnumStrings returns all valid values for MaskContentUnits as strings
func (MaskContentUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"objectBoundingBox",
	}
}

// String implements the fmt.Stringer interface for MaskContentUnits
func (v MaskContentUnits) String() string {
	return string(v)
}

// AttribName returns the "maskContentUnits" attribute name.
func (v MaskContentUnits) AttribName() string { return "maskContentUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v MaskContentUnits) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// FilterUnits is the SVG filterUnits attribute.
type FilterUnits string //#enum

const (
	// FilterUnitsUserSpaceOnUse interprets the filter region's x, y, width and height in the current user coordinate system.
	FilterUnitsUserSpaceOnUse FilterUnits = "userSpaceOnUse"
	// FilterUnitsObjectBoundingBox interprets the filter region's x, y, width and height as fractions of the referencing element's bounding box.
	FilterUnitsObjectBoundingBox FilterUnits = "objectBoundingBox"
)

// Valid indicates if v is any of the valid values for FilterUnits
func (v FilterUnits) Valid() bool {
	switch v {
	case
		FilterUnitsUserSpaceOnUse,
		FilterUnitsObjectBoundingBox:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for FilterUnits
func (v FilterUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.FilterUnits", v)
	}
	return nil
}

// Enums returns all valid values for FilterUnits
func (FilterUnits) Enums() []FilterUnits {
	return []FilterUnits{
		FilterUnitsUserSpaceOnUse,
		FilterUnitsObjectBoundingBox,
	}
}

// EnumStrings returns all valid values for FilterUnits as strings
func (FilterUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"objectBoundingBox",
	}
}

// String implements the fmt.Stringer interface for FilterUnits
func (v FilterUnits) String() string {
	return string(v)
}

// AttribName returns the "filterUnits" attribute name.
func (v FilterUnits) AttribName() string { return "filterUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v FilterUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// PrimitiveUnits is the SVG primitiveUnits attribute.
type PrimitiveUnits string //#enum

const (
	// PrimitiveUnitsUserSpaceOnUse interprets filter primitive coordinates and lengths in the current user coordinate system.
	PrimitiveUnitsUserSpaceOnUse PrimitiveUnits = "userSpaceOnUse"
	// PrimitiveUnitsObjectBoundingBox interprets filter primitive coordinates and lengths as fractions of the referencing element's bounding box.
	PrimitiveUnitsObjectBoundingBox PrimitiveUnits = "objectBoundingBox"
)

// Valid indicates if v is any of the valid values for PrimitiveUnits
func (v PrimitiveUnits) Valid() bool {
	switch v {
	case
		PrimitiveUnitsUserSpaceOnUse,
		PrimitiveUnitsObjectBoundingBox:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for PrimitiveUnits
func (v PrimitiveUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.PrimitiveUnits", v)
	}
	return nil
}

// Enums returns all valid values for PrimitiveUnits
func (PrimitiveUnits) Enums() []PrimitiveUnits {
	return []PrimitiveUnits{
		PrimitiveUnitsUserSpaceOnUse,
		PrimitiveUnitsObjectBoundingBox,
	}
}

// EnumStrings returns all valid values for PrimitiveUnits as strings
func (PrimitiveUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"objectBoundingBox",
	}
}

// String implements the fmt.Stringer interface for PrimitiveUnits
func (v PrimitiveUnits) String() string {
	return string(v)
}

// AttribName returns the "primitiveUnits" attribute name.
func (v PrimitiveUnits) AttribName() string { return "primitiveUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v PrimitiveUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// MarkerUnits is the SVG markerUnits attribute.
type MarkerUnits string //#enum

const (
	// MarkerUnitsUserSpaceOnUse interprets marker dimensions in the current user coordinate system.
	MarkerUnitsUserSpaceOnUse MarkerUnits = "userSpaceOnUse"
	// MarkerUnitsStrokeWidth scales marker dimensions by the stroke width of the referencing element.
	MarkerUnitsStrokeWidth MarkerUnits = "strokeWidth"
)

// Valid indicates if v is any of the valid values for MarkerUnits
func (v MarkerUnits) Valid() bool {
	switch v {
	case
		MarkerUnitsUserSpaceOnUse,
		MarkerUnitsStrokeWidth:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for MarkerUnits
func (v MarkerUnits) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.MarkerUnits", v)
	}
	return nil
}

// Enums returns all valid values for MarkerUnits
func (MarkerUnits) Enums() []MarkerUnits {
	return []MarkerUnits{
		MarkerUnitsUserSpaceOnUse,
		MarkerUnitsStrokeWidth,
	}
}

// EnumStrings returns all valid values for MarkerUnits as strings
func (MarkerUnits) EnumStrings() []string {
	return []string{
		"userSpaceOnUse",
		"strokeWidth",
	}
}

// String implements the fmt.Stringer interface for MarkerUnits
func (v MarkerUnits) String() string {
	return string(v)
}

// AttribName returns the "markerUnits" attribute name.
func (v MarkerUnits) AttribName() string { return "markerUnits" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v MarkerUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Gradients and filter primitives

// SpreadMethod is the SVG spreadMethod attribute of gradients.
type SpreadMethod string //#enum

const (
	// SpreadMethodPad extends the gradient's end colors to fill the remaining area.
	SpreadMethodPad SpreadMethod = "pad"
	// SpreadMethodReflect repeats the gradient with alternating direction to fill the remaining area.
	SpreadMethodReflect SpreadMethod = "reflect"
	// SpreadMethodRepeat tiles the gradient in the same direction to fill the remaining area.
	SpreadMethodRepeat SpreadMethod = "repeat"
)

// Valid indicates if v is any of the valid values for SpreadMethod
func (v SpreadMethod) Valid() bool {
	switch v {
	case
		SpreadMethodPad,
		SpreadMethodReflect,
		SpreadMethodRepeat:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for SpreadMethod
func (v SpreadMethod) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.SpreadMethod", v)
	}
	return nil
}

// Enums returns all valid values for SpreadMethod
func (SpreadMethod) Enums() []SpreadMethod {
	return []SpreadMethod{
		SpreadMethodPad,
		SpreadMethodReflect,
		SpreadMethodRepeat,
	}
}

// EnumStrings returns all valid values for SpreadMethod as strings
func (SpreadMethod) EnumStrings() []string {
	return []string{
		"pad",
		"reflect",
		"repeat",
	}
}

// String implements the fmt.Stringer interface for SpreadMethod
func (v SpreadMethod) String() string {
	return string(v)
}

// AttribName returns the "spreadMethod" attribute name.
func (v SpreadMethod) AttribName() string { return "spreadMethod" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v SpreadMethod) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// EdgeMode is the SVG edgeMode attribute of <feConvolveMatrix>/<feGaussianBlur>.
type EdgeMode string //#enum

const (
	// EdgeModeDuplicate extends the input image by duplicating its edge pixels.
	EdgeModeDuplicate EdgeMode = "duplicate"
	// EdgeModeWrap extends the input image by tiling it.
	EdgeModeWrap EdgeMode = "wrap"
	// EdgeModeNone treats pixels outside the input image as transparent black.
	EdgeModeNone EdgeMode = "none"
)

// Valid indicates if v is any of the valid values for EdgeMode
func (v EdgeMode) Valid() bool {
	switch v {
	case
		EdgeModeDuplicate,
		EdgeModeWrap,
		EdgeModeNone:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for EdgeMode
func (v EdgeMode) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.EdgeMode", v)
	}
	return nil
}

// Enums returns all valid values for EdgeMode
func (EdgeMode) Enums() []EdgeMode {
	return []EdgeMode{
		EdgeModeDuplicate,
		EdgeModeWrap,
		EdgeModeNone,
	}
}

// EnumStrings returns all valid values for EdgeMode as strings
func (EdgeMode) EnumStrings() []string {
	return []string{
		"duplicate",
		"wrap",
		"none",
	}
}

// String implements the fmt.Stringer interface for EdgeMode
func (v EdgeMode) String() string {
	return string(v)
}

// AttribName returns the "edgeMode" attribute name.
func (v EdgeMode) AttribName() string { return "edgeMode" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v EdgeMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// StitchTiles is the SVG stitchTiles attribute of <feTurbulence>.
type StitchTiles string //#enum

const (
	// StitchTilesStitch adjusts noise frequencies so adjacent tiles join without seams.
	StitchTilesStitch StitchTiles = "stitch"
	// StitchTilesNoStitch leaves tile edges unmatched, allowing visible seams.
	StitchTilesNoStitch StitchTiles = "noStitch"
)

// Valid indicates if v is any of the valid values for StitchTiles
func (v StitchTiles) Valid() bool {
	switch v {
	case
		StitchTilesStitch,
		StitchTilesNoStitch:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for StitchTiles
func (v StitchTiles) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.StitchTiles", v)
	}
	return nil
}

// Enums returns all valid values for StitchTiles
func (StitchTiles) Enums() []StitchTiles {
	return []StitchTiles{
		StitchTilesStitch,
		StitchTilesNoStitch,
	}
}

// EnumStrings returns all valid values for StitchTiles as strings
func (StitchTiles) EnumStrings() []string {
	return []string{
		"stitch",
		"noStitch",
	}
}

// String implements the fmt.Stringer interface for StitchTiles
func (v StitchTiles) String() string {
	return string(v)
}

// AttribName returns the "stitchTiles" attribute name.
func (v StitchTiles) AttribName() string { return "stitchTiles" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v StitchTiles) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// XChannelSelector is the SVG xChannelSelector attribute of <feDisplacementMap>.
type XChannelSelector string //#enum

const (
	// XChannelSelectorR uses the red channel of the displacement map for the x displacement.
	XChannelSelectorR XChannelSelector = "R"
	// XChannelSelectorG uses the green channel of the displacement map for the x displacement.
	XChannelSelectorG XChannelSelector = "G"
	// XChannelSelectorB uses the blue channel of the displacement map for the x displacement.
	XChannelSelectorB XChannelSelector = "B"
	// XChannelSelectorA uses the alpha channel of the displacement map for the x displacement.
	XChannelSelectorA XChannelSelector = "A"
)

// Valid indicates if v is any of the valid values for XChannelSelector
func (v XChannelSelector) Valid() bool {
	switch v {
	case
		XChannelSelectorR,
		XChannelSelectorG,
		XChannelSelectorB,
		XChannelSelectorA:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for XChannelSelector
func (v XChannelSelector) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.XChannelSelector", v)
	}
	return nil
}

// Enums returns all valid values for XChannelSelector
func (XChannelSelector) Enums() []XChannelSelector {
	return []XChannelSelector{
		XChannelSelectorR,
		XChannelSelectorG,
		XChannelSelectorB,
		XChannelSelectorA,
	}
}

// EnumStrings returns all valid values for XChannelSelector as strings
func (XChannelSelector) EnumStrings() []string {
	return []string{
		"R",
		"G",
		"B",
		"A",
	}
}

// String implements the fmt.Stringer interface for XChannelSelector
func (v XChannelSelector) String() string {
	return string(v)
}

// AttribName returns the "xChannelSelector" attribute name.
func (v XChannelSelector) AttribName() string { return "xChannelSelector" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v XChannelSelector) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// YChannelSelector is the SVG yChannelSelector attribute of <feDisplacementMap>.
type YChannelSelector string //#enum

const (
	// YChannelSelectorR uses the red channel of the displacement map for the y displacement.
	YChannelSelectorR YChannelSelector = "R"
	// YChannelSelectorG uses the green channel of the displacement map for the y displacement.
	YChannelSelectorG YChannelSelector = "G"
	// YChannelSelectorB uses the blue channel of the displacement map for the y displacement.
	YChannelSelectorB YChannelSelector = "B"
	// YChannelSelectorA uses the alpha channel of the displacement map for the y displacement.
	YChannelSelectorA YChannelSelector = "A"
)

// Valid indicates if v is any of the valid values for YChannelSelector
func (v YChannelSelector) Valid() bool {
	switch v {
	case
		YChannelSelectorR,
		YChannelSelectorG,
		YChannelSelectorB,
		YChannelSelectorA:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for YChannelSelector
func (v YChannelSelector) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.YChannelSelector", v)
	}
	return nil
}

// Enums returns all valid values for YChannelSelector
func (YChannelSelector) Enums() []YChannelSelector {
	return []YChannelSelector{
		YChannelSelectorR,
		YChannelSelectorG,
		YChannelSelectorB,
		YChannelSelectorA,
	}
}

// EnumStrings returns all valid values for YChannelSelector as strings
func (YChannelSelector) EnumStrings() []string {
	return []string{
		"R",
		"G",
		"B",
		"A",
	}
}

// String implements the fmt.Stringer interface for YChannelSelector
func (v YChannelSelector) String() string {
	return string(v)
}

// AttribName returns the "yChannelSelector" attribute name.
func (v YChannelSelector) AttribName() string { return "yChannelSelector" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v YChannelSelector) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// Animation

// AttributeType is the SVG attributeType attribute of animation elements.
type AttributeType string //#enum

const (
	// AttributeTypeCSS animates the target as a CSS property.
	AttributeTypeCSS AttributeType = "CSS"
	// AttributeTypeXML animates the target as an XML presentation attribute.
	AttributeTypeXML AttributeType = "XML"
	// AttributeTypeAuto lets the user agent decide whether the target is a CSS property or XML attribute.
	AttributeTypeAuto AttributeType = "auto"
)

// Valid indicates if v is any of the valid values for AttributeType
func (v AttributeType) Valid() bool {
	switch v {
	case
		AttributeTypeCSS,
		AttributeTypeXML,
		AttributeTypeAuto:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for AttributeType
func (v AttributeType) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.AttributeType", v)
	}
	return nil
}

// Enums returns all valid values for AttributeType
func (AttributeType) Enums() []AttributeType {
	return []AttributeType{
		AttributeTypeCSS,
		AttributeTypeXML,
		AttributeTypeAuto,
	}
}

// EnumStrings returns all valid values for AttributeType as strings
func (AttributeType) EnumStrings() []string {
	return []string{
		"CSS",
		"XML",
		"auto",
	}
}

// String implements the fmt.Stringer interface for AttributeType
func (v AttributeType) String() string {
	return string(v)
}

// AttribName returns the "attributeType" attribute name.
func (v AttributeType) AttribName() string { return "attributeType" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v AttributeType) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// CalcMode is the SVG calcMode attribute of animation elements.
type CalcMode string //#enum

const (
	// CalcModeDiscrete jumps from one value to the next without interpolation.
	CalcModeDiscrete CalcMode = "discrete"
	// CalcModeLinear interpolates values linearly between keyframes.
	CalcModeLinear CalcMode = "linear"
	// CalcModePaced interpolates linearly at an even pace across the whole animation.
	CalcModePaced CalcMode = "paced"
	// CalcModeSpline interpolates along the cubic Bezier curves given by keySplines.
	CalcModeSpline CalcMode = "spline"
)

// Valid indicates if v is any of the valid values for CalcMode
func (v CalcMode) Valid() bool {
	switch v {
	case
		CalcModeDiscrete,
		CalcModeLinear,
		CalcModePaced,
		CalcModeSpline:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for CalcMode
func (v CalcMode) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.CalcMode", v)
	}
	return nil
}

// Enums returns all valid values for CalcMode
func (CalcMode) Enums() []CalcMode {
	return []CalcMode{
		CalcModeDiscrete,
		CalcModeLinear,
		CalcModePaced,
		CalcModeSpline,
	}
}

// EnumStrings returns all valid values for CalcMode as strings
func (CalcMode) EnumStrings() []string {
	return []string{
		"discrete",
		"linear",
		"paced",
		"spline",
	}
}

// String implements the fmt.Stringer interface for CalcMode
func (v CalcMode) String() string {
	return string(v)
}

// AttribName returns the "calcMode" attribute name.
func (v CalcMode) AttribName() string { return "calcMode" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v CalcMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Additive is the SVG additive attribute of animation elements.
type Additive string //#enum

const (
	// AdditiveReplace overrides the underlying attribute value with the animation value.
	AdditiveReplace Additive = "replace"
	// AdditiveSum adds the animation value to the underlying attribute value.
	AdditiveSum Additive = "sum"
)

// Valid indicates if v is any of the valid values for Additive
func (v Additive) Valid() bool {
	switch v {
	case
		AdditiveReplace,
		AdditiveSum:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Additive
func (v Additive) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Additive", v)
	}
	return nil
}

// Enums returns all valid values for Additive
func (Additive) Enums() []Additive {
	return []Additive{
		AdditiveReplace,
		AdditiveSum,
	}
}

// EnumStrings returns all valid values for Additive as strings
func (Additive) EnumStrings() []string {
	return []string{
		"replace",
		"sum",
	}
}

// String implements the fmt.Stringer interface for Additive
func (v Additive) String() string {
	return string(v)
}

// AttribName returns the "additive" attribute name.
func (v Additive) AttribName() string { return "additive" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Additive) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Accumulate is the SVG accumulate attribute of animation elements.
type Accumulate string //#enum

const (
	// AccumulateNone restarts each repeat iteration from the base value.
	AccumulateNone Accumulate = "none"
	// AccumulateSum adds each repeat iteration's end value to the next, accumulating over repeats.
	AccumulateSum Accumulate = "sum"
)

// Valid indicates if v is any of the valid values for Accumulate
func (v Accumulate) Valid() bool {
	switch v {
	case
		AccumulateNone,
		AccumulateSum:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Accumulate
func (v Accumulate) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Accumulate", v)
	}
	return nil
}

// Enums returns all valid values for Accumulate
func (Accumulate) Enums() []Accumulate {
	return []Accumulate{
		AccumulateNone,
		AccumulateSum,
	}
}

// EnumStrings returns all valid values for Accumulate as strings
func (Accumulate) EnumStrings() []string {
	return []string{
		"none",
		"sum",
	}
}

// String implements the fmt.Stringer interface for Accumulate
func (v Accumulate) String() string {
	return string(v)
}

// AttribName returns the "accumulate" attribute name.
func (v Accumulate) AttribName() string { return "accumulate" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Accumulate) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Restart is the SVG restart attribute of animation elements.
type Restart string //#enum

const (
	// RestartAlways allows the animation to restart at any time.
	RestartAlways Restart = "always"
	// RestartWhenNotActive allows a restart only while the animation is not currently active.
	RestartWhenNotActive Restart = "whenNotActive"
	// RestartNever prevents the animation from restarting once begun.
	RestartNever Restart = "never"
)

// Valid indicates if v is any of the valid values for Restart
func (v Restart) Valid() bool {
	switch v {
	case
		RestartAlways,
		RestartWhenNotActive,
		RestartNever:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Restart
func (v Restart) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.Restart", v)
	}
	return nil
}

// Enums returns all valid values for Restart
func (Restart) Enums() []Restart {
	return []Restart{
		RestartAlways,
		RestartWhenNotActive,
		RestartNever,
	}
}

// EnumStrings returns all valid values for Restart as strings
func (Restart) EnumStrings() []string {
	return []string{
		"always",
		"whenNotActive",
		"never",
	}
}

// String implements the fmt.Stringer interface for Restart
func (v Restart) String() string {
	return string(v)
}

// AttribName returns the "restart" attribute name.
func (v Restart) AttribName() string { return "restart" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Restart) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Misc

// CrossOrigin is the SVG crossorigin attribute.
type CrossOrigin string //#enum

const (
	// CrossOriginAnonymous makes the cross-origin request without credentials.
	CrossOriginAnonymous CrossOrigin = "anonymous"
	// CrossOriginUseCredentials makes the cross-origin request with credentials.
	CrossOriginUseCredentials CrossOrigin = "use-credentials"
)

// Valid indicates if v is any of the valid values for CrossOrigin
func (v CrossOrigin) Valid() bool {
	switch v {
	case
		CrossOriginAnonymous,
		CrossOriginUseCredentials:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for CrossOrigin
func (v CrossOrigin) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.CrossOrigin", v)
	}
	return nil
}

// Enums returns all valid values for CrossOrigin
func (CrossOrigin) Enums() []CrossOrigin {
	return []CrossOrigin{
		CrossOriginAnonymous,
		CrossOriginUseCredentials,
	}
}

// EnumStrings returns all valid values for CrossOrigin as strings
func (CrossOrigin) EnumStrings() []string {
	return []string{
		"anonymous",
		"use-credentials",
	}
}

// String implements the fmt.Stringer interface for CrossOrigin
func (v CrossOrigin) String() string {
	return string(v)
}

// AttribName returns the "crossorigin" attribute name.
func (v CrossOrigin) AttribName() string { return "crossorigin" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v CrossOrigin) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ZoomAndPan is the SVG zoomAndPan attribute.
type ZoomAndPan string //#enum

const (
	// ZoomAndPanDisable disables user-initiated zooming and panning of the viewport.
	ZoomAndPanDisable ZoomAndPan = "disable"
	// ZoomAndPanMagnify allows user-initiated zooming and panning of the viewport.
	ZoomAndPanMagnify ZoomAndPan = "magnify"
)

// Valid indicates if v is any of the valid values for ZoomAndPan
func (v ZoomAndPan) Valid() bool {
	switch v {
	case
		ZoomAndPanDisable,
		ZoomAndPanMagnify:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for ZoomAndPan
func (v ZoomAndPan) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type svg.ZoomAndPan", v)
	}
	return nil
}

// Enums returns all valid values for ZoomAndPan
func (ZoomAndPan) Enums() []ZoomAndPan {
	return []ZoomAndPan{
		ZoomAndPanDisable,
		ZoomAndPanMagnify,
	}
}

// EnumStrings returns all valid values for ZoomAndPan as strings
func (ZoomAndPan) EnumStrings() []string {
	return []string{
		"disable",
		"magnify",
	}
}

// String implements the fmt.Stringer interface for ZoomAndPan
func (v ZoomAndPan) String() string {
	return string(v)
}

// AttribName returns the "zoomAndPan" attribute name.
func (v ZoomAndPan) AttribName() string { return "zoomAndPan" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v ZoomAndPan) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Compile-time checks that every enum type is usable directly as an attribute.
var (
	_ mx.Attrib = FillRule("")
	_ mx.Attrib = ClipRule("")
	_ mx.Attrib = StrokeLineCap("")
	_ mx.Attrib = StrokeLineJoin("")
	_ mx.Attrib = Visibility("")
	_ mx.Attrib = Overflow("")
	_ mx.Attrib = PointerEvents("")
	_ mx.Attrib = ShapeRendering("")
	_ mx.Attrib = VectorEffect("")
	_ mx.Attrib = Isolation("")
	_ mx.Attrib = ColorInterpolation("")
	_ mx.Attrib = ColorInterpolationFilters("")
	_ mx.Attrib = MixBlendMode("")
	_ mx.Attrib = Mode("")
	_ mx.Attrib = FontStyle("")
	_ mx.Attrib = TextAnchor("")
	_ mx.Attrib = DominantBaseline("")
	_ mx.Attrib = AlignmentBaseline("")
	_ mx.Attrib = WritingMode("")
	_ mx.Attrib = Direction("")
	_ mx.Attrib = LengthAdjust("")
	_ mx.Attrib = Method("")
	_ mx.Attrib = Spacing("")
	_ mx.Attrib = Side("")
	_ mx.Attrib = GradientUnits("")
	_ mx.Attrib = PatternUnits("")
	_ mx.Attrib = PatternContentUnits("")
	_ mx.Attrib = ClipPathUnits("")
	_ mx.Attrib = MaskUnits("")
	_ mx.Attrib = MaskContentUnits("")
	_ mx.Attrib = FilterUnits("")
	_ mx.Attrib = PrimitiveUnits("")
	_ mx.Attrib = MarkerUnits("")
	_ mx.Attrib = SpreadMethod("")
	_ mx.Attrib = EdgeMode("")
	_ mx.Attrib = StitchTiles("")
	_ mx.Attrib = XChannelSelector("")
	_ mx.Attrib = YChannelSelector("")
	_ mx.Attrib = AttributeType("")
	_ mx.Attrib = CalcMode("")
	_ mx.Attrib = Additive("")
	_ mx.Attrib = Accumulate("")
	_ mx.Attrib = Restart("")
	_ mx.Attrib = CrossOrigin("")
	_ mx.Attrib = ZoomAndPan("")
)
