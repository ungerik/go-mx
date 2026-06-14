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
	FillRuleNonzero FillRule = "nonzero"
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

func (v FillRule) AttribName() string                          { return "fill-rule" }
func (v FillRule) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ClipRule is the SVG clip-rule presentation attribute (an enumerated keyword).
type ClipRule string //#enum

const (
	ClipRuleNonzero ClipRule = "nonzero"
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

func (v ClipRule) AttribName() string                          { return "clip-rule" }
func (v ClipRule) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// StrokeLineCap is the SVG stroke-linecap presentation attribute.
type StrokeLineCap string //#enum

const (
	StrokeLineCapButt   StrokeLineCap = "butt"
	StrokeLineCapRound  StrokeLineCap = "round"
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

func (v StrokeLineCap) AttribName() string                          { return "stroke-linecap" }
func (v StrokeLineCap) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// StrokeLineJoin is the SVG stroke-linejoin presentation attribute.
type StrokeLineJoin string //#enum

const (
	StrokeLineJoinArcs      StrokeLineJoin = "arcs"
	StrokeLineJoinBevel     StrokeLineJoin = "bevel"
	StrokeLineJoinMiter     StrokeLineJoin = "miter"
	StrokeLineJoinMiterClip StrokeLineJoin = "miter-clip"
	StrokeLineJoinRound     StrokeLineJoin = "round"
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

func (v StrokeLineJoin) AttribName() string                          { return "stroke-linejoin" }
func (v StrokeLineJoin) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// General presentation

// Visibility is the SVG visibility presentation attribute.
type Visibility string //#enum

const (
	VisibilityVisible  Visibility = "visible"
	VisibilityHidden   Visibility = "hidden"
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

func (v Visibility) AttribName() string                          { return "visibility" }
func (v Visibility) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Overflow is the SVG overflow presentation attribute.
type Overflow string //#enum

const (
	OverflowVisible Overflow = "visible"
	OverflowHidden  Overflow = "hidden"
	OverflowScroll  Overflow = "scroll"
	OverflowAuto    Overflow = "auto"
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

func (v Overflow) AttribName() string                          { return "overflow" }
func (v Overflow) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// PointerEvents is the SVG pointer-events presentation attribute.
type PointerEvents string //#enum

const (
	PointerEventsBoundingBox    PointerEvents = "bounding-box"
	PointerEventsVisiblePainted PointerEvents = "visiblePainted"
	PointerEventsVisibleFill    PointerEvents = "visibleFill"
	PointerEventsVisibleStroke  PointerEvents = "visibleStroke"
	PointerEventsVisible        PointerEvents = "visible"
	PointerEventsPainted        PointerEvents = "painted"
	PointerEventsFill           PointerEvents = "fill"
	PointerEventsStroke         PointerEvents = "stroke"
	PointerEventsAll            PointerEvents = "all"
	PointerEventsNone           PointerEvents = "none"
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

func (v PointerEvents) AttribName() string                          { return "pointer-events" }
func (v PointerEvents) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ShapeRendering is the SVG shape-rendering presentation attribute.
type ShapeRendering string //#enum

const (
	ShapeRenderingAuto               ShapeRendering = "auto"
	ShapeRenderingOptimizeSpeed      ShapeRendering = "optimizeSpeed"
	ShapeRenderingCrispEdges         ShapeRendering = "crispEdges"
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

func (v ShapeRendering) AttribName() string                          { return "shape-rendering" }
func (v ShapeRendering) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// VectorEffect is the SVG vector-effect presentation attribute.
type VectorEffect string //#enum

const (
	VectorEffectNone             VectorEffect = "none"
	VectorEffectNonScalingStroke VectorEffect = "non-scaling-stroke"
	VectorEffectNonScalingSize   VectorEffect = "non-scaling-size"
	VectorEffectNonRotation      VectorEffect = "non-rotation"
	VectorEffectFixedPosition    VectorEffect = "fixed-position"
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

func (v VectorEffect) AttribName() string                          { return "vector-effect" }
func (v VectorEffect) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Isolation is the SVG isolation presentation attribute.
type Isolation string //#enum

const (
	IsolationAuto    Isolation = "auto"
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

func (v Isolation) AttribName() string                          { return "isolation" }
func (v Isolation) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ColorInterpolation is the SVG color-interpolation presentation attribute.
type ColorInterpolation string //#enum

const (
	ColorInterpolationAuto      ColorInterpolation = "auto"
	ColorInterpolationSRGB      ColorInterpolation = "sRGB"
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

func (v ColorInterpolation) AttribName() string { return "color-interpolation" }
func (v ColorInterpolation) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// ColorInterpolationFilters is the SVG color-interpolation-filters presentation
// attribute.
type ColorInterpolationFilters string //#enum

const (
	ColorInterpolationFiltersAuto      ColorInterpolationFilters = "auto"
	ColorInterpolationFiltersSRGB      ColorInterpolationFilters = "sRGB"
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

func (v ColorInterpolationFilters) AttribName() string { return "color-interpolation-filters" }
func (v ColorInterpolationFilters) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// Blend modes

// MixBlendMode is the SVG mix-blend-mode presentation attribute (a <blend-mode>).
type MixBlendMode string //#enum

const (
	MixBlendModeNormal     MixBlendMode = "normal"
	MixBlendModeMultiply   MixBlendMode = "multiply"
	MixBlendModeScreen     MixBlendMode = "screen"
	MixBlendModeOverlay    MixBlendMode = "overlay"
	MixBlendModeDarken     MixBlendMode = "darken"
	MixBlendModeLighten    MixBlendMode = "lighten"
	MixBlendModeColorDodge MixBlendMode = "color-dodge"
	MixBlendModeColorBurn  MixBlendMode = "color-burn"
	MixBlendModeHardLight  MixBlendMode = "hard-light"
	MixBlendModeSoftLight  MixBlendMode = "soft-light"
	MixBlendModeDifference MixBlendMode = "difference"
	MixBlendModeExclusion  MixBlendMode = "exclusion"
	MixBlendModeHue        MixBlendMode = "hue"
	MixBlendModeSaturation MixBlendMode = "saturation"
	MixBlendModeColor      MixBlendMode = "color"
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

func (v MixBlendMode) AttribName() string                          { return "mix-blend-mode" }
func (v MixBlendMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Mode is the SVG mode attribute of <feBlend> (a <blend-mode>).
type Mode string //#enum

const (
	ModeNormal     Mode = "normal"
	ModeMultiply   Mode = "multiply"
	ModeScreen     Mode = "screen"
	ModeOverlay    Mode = "overlay"
	ModeDarken     Mode = "darken"
	ModeLighten    Mode = "lighten"
	ModeColorDodge Mode = "color-dodge"
	ModeColorBurn  Mode = "color-burn"
	ModeHardLight  Mode = "hard-light"
	ModeSoftLight  Mode = "soft-light"
	ModeDifference Mode = "difference"
	ModeExclusion  Mode = "exclusion"
	ModeHue        Mode = "hue"
	ModeSaturation Mode = "saturation"
	ModeColor      Mode = "color"
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

func (v Mode) AttribName() string                          { return "mode" }
func (v Mode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Text and fonts

// FontStyle is the SVG font-style presentation attribute.
type FontStyle string //#enum

const (
	FontStyleNormal  FontStyle = "normal"
	FontStyleItalic  FontStyle = "italic"
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

func (v FontStyle) AttribName() string                          { return "font-style" }
func (v FontStyle) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// TextAnchor is the SVG text-anchor presentation attribute.
type TextAnchor string //#enum

const (
	TextAnchorStart  TextAnchor = "start"
	TextAnchorMiddle TextAnchor = "middle"
	TextAnchorEnd    TextAnchor = "end"
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

func (v TextAnchor) AttribName() string                          { return "text-anchor" }
func (v TextAnchor) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// DominantBaseline is the SVG dominant-baseline presentation attribute.
type DominantBaseline string //#enum

const (
	DominantBaselineAuto         DominantBaseline = "auto"
	DominantBaselineTextBottom   DominantBaseline = "text-bottom"
	DominantBaselineAlphabetic   DominantBaseline = "alphabetic"
	DominantBaselineIdeographic  DominantBaseline = "ideographic"
	DominantBaselineMiddle       DominantBaseline = "middle"
	DominantBaselineCentral      DominantBaseline = "central"
	DominantBaselineMathematical DominantBaseline = "mathematical"
	DominantBaselineHanging      DominantBaseline = "hanging"
	DominantBaselineTextTop      DominantBaseline = "text-top"
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

func (v DominantBaseline) AttribName() string { return "dominant-baseline" }
func (v DominantBaseline) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// AlignmentBaseline is the SVG alignment-baseline presentation attribute.
type AlignmentBaseline string //#enum

const (
	AlignmentBaselineBaseline     AlignmentBaseline = "baseline"
	AlignmentBaselineTextBottom   AlignmentBaseline = "text-bottom"
	AlignmentBaselineAlphabetic   AlignmentBaseline = "alphabetic"
	AlignmentBaselineIdeographic  AlignmentBaseline = "ideographic"
	AlignmentBaselineMiddle       AlignmentBaseline = "middle"
	AlignmentBaselineCentral      AlignmentBaseline = "central"
	AlignmentBaselineMathematical AlignmentBaseline = "mathematical"
	AlignmentBaselineTextTop      AlignmentBaseline = "text-top"
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

func (v AlignmentBaseline) AttribName() string { return "alignment-baseline" }
func (v AlignmentBaseline) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// WritingMode is the SVG writing-mode presentation attribute.
type WritingMode string //#enum

const (
	WritingModeHorizontalTB WritingMode = "horizontal-tb"
	WritingModeVerticalRL   WritingMode = "vertical-rl"
	WritingModeVerticalLR   WritingMode = "vertical-lr"
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

func (v WritingMode) AttribName() string                          { return "writing-mode" }
func (v WritingMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Direction is the SVG direction presentation attribute.
type Direction string //#enum

const (
	DirectionLTR Direction = "ltr"
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

func (v Direction) AttribName() string                          { return "direction" }
func (v Direction) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// LengthAdjust is the SVG lengthAdjust attribute of <text>/<textPath>.
type LengthAdjust string //#enum

const (
	LengthAdjustSpacing          LengthAdjust = "spacing"
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

func (v LengthAdjust) AttribName() string                          { return "lengthAdjust" }
func (v LengthAdjust) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Method is the SVG method attribute of <textPath>.
type Method string //#enum

const (
	MethodAlign   Method = "align"
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

func (v Method) AttribName() string                          { return "method" }
func (v Method) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Spacing is the SVG spacing attribute of <textPath>.
type Spacing string //#enum

const (
	SpacingAuto  Spacing = "auto"
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

func (v Spacing) AttribName() string                          { return "spacing" }
func (v Spacing) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Side is the SVG side attribute of <textPath>.
type Side string //#enum

const (
	SideLeft  Side = "left"
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

func (v Side) AttribName() string                          { return "side" }
func (v Side) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Units

// GradientUnits is the SVG gradientUnits attribute.
type GradientUnits string //#enum

const (
	GradientUnitsUserSpaceOnUse    GradientUnits = "userSpaceOnUse"
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

func (v GradientUnits) AttribName() string                          { return "gradientUnits" }
func (v GradientUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// PatternUnits is the SVG patternUnits attribute.
type PatternUnits string //#enum

const (
	PatternUnitsUserSpaceOnUse    PatternUnits = "userSpaceOnUse"
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

func (v PatternUnits) AttribName() string                          { return "patternUnits" }
func (v PatternUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// PatternContentUnits is the SVG patternContentUnits attribute.
type PatternContentUnits string //#enum

const (
	PatternContentUnitsUserSpaceOnUse    PatternContentUnits = "userSpaceOnUse"
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

func (v PatternContentUnits) AttribName() string { return "patternContentUnits" }
func (v PatternContentUnits) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// ClipPathUnits is the SVG clipPathUnits attribute.
type ClipPathUnits string //#enum

const (
	ClipPathUnitsUserSpaceOnUse    ClipPathUnits = "userSpaceOnUse"
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

func (v ClipPathUnits) AttribName() string                          { return "clipPathUnits" }
func (v ClipPathUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// MaskUnits is the SVG maskUnits attribute.
type MaskUnits string //#enum

const (
	MaskUnitsUserSpaceOnUse    MaskUnits = "userSpaceOnUse"
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

func (v MaskUnits) AttribName() string                          { return "maskUnits" }
func (v MaskUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// MaskContentUnits is the SVG maskContentUnits attribute.
type MaskContentUnits string //#enum

const (
	MaskContentUnitsUserSpaceOnUse    MaskContentUnits = "userSpaceOnUse"
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

func (v MaskContentUnits) AttribName() string { return "maskContentUnits" }
func (v MaskContentUnits) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// FilterUnits is the SVG filterUnits attribute.
type FilterUnits string //#enum

const (
	FilterUnitsUserSpaceOnUse    FilterUnits = "userSpaceOnUse"
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

func (v FilterUnits) AttribName() string                          { return "filterUnits" }
func (v FilterUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// PrimitiveUnits is the SVG primitiveUnits attribute.
type PrimitiveUnits string //#enum

const (
	PrimitiveUnitsUserSpaceOnUse    PrimitiveUnits = "userSpaceOnUse"
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

func (v PrimitiveUnits) AttribName() string                          { return "primitiveUnits" }
func (v PrimitiveUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// MarkerUnits is the SVG markerUnits attribute.
type MarkerUnits string //#enum

const (
	MarkerUnitsUserSpaceOnUse MarkerUnits = "userSpaceOnUse"
	MarkerUnitsStrokeWidth    MarkerUnits = "strokeWidth"
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

func (v MarkerUnits) AttribName() string                          { return "markerUnits" }
func (v MarkerUnits) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Gradients and filter primitives

// SpreadMethod is the SVG spreadMethod attribute of gradients.
type SpreadMethod string //#enum

const (
	SpreadMethodPad     SpreadMethod = "pad"
	SpreadMethodReflect SpreadMethod = "reflect"
	SpreadMethodRepeat  SpreadMethod = "repeat"
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

func (v SpreadMethod) AttribName() string                          { return "spreadMethod" }
func (v SpreadMethod) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// EdgeMode is the SVG edgeMode attribute of <feConvolveMatrix>/<feGaussianBlur>.
type EdgeMode string //#enum

const (
	EdgeModeDuplicate EdgeMode = "duplicate"
	EdgeModeWrap      EdgeMode = "wrap"
	EdgeModeNone      EdgeMode = "none"
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

func (v EdgeMode) AttribName() string                          { return "edgeMode" }
func (v EdgeMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// StitchTiles is the SVG stitchTiles attribute of <feTurbulence>.
type StitchTiles string //#enum

const (
	StitchTilesStitch   StitchTiles = "stitch"
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

func (v StitchTiles) AttribName() string                          { return "stitchTiles" }
func (v StitchTiles) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// XChannelSelector is the SVG xChannelSelector attribute of <feDisplacementMap>.
type XChannelSelector string //#enum

const (
	XChannelSelectorR XChannelSelector = "R"
	XChannelSelectorG XChannelSelector = "G"
	XChannelSelectorB XChannelSelector = "B"
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

func (v XChannelSelector) AttribName() string { return "xChannelSelector" }
func (v XChannelSelector) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// YChannelSelector is the SVG yChannelSelector attribute of <feDisplacementMap>.
type YChannelSelector string //#enum

const (
	YChannelSelectorR YChannelSelector = "R"
	YChannelSelectorG YChannelSelector = "G"
	YChannelSelectorB YChannelSelector = "B"
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

func (v YChannelSelector) AttribName() string { return "yChannelSelector" }
func (v YChannelSelector) AttribValue(context.Context) (string, error) {
	return string(v), v.Validate()
}

// Animation

// AttributeType is the SVG attributeType attribute of animation elements.
type AttributeType string //#enum

const (
	AttributeTypeCSS  AttributeType = "CSS"
	AttributeTypeXML  AttributeType = "XML"
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

func (v AttributeType) AttribName() string                          { return "attributeType" }
func (v AttributeType) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// CalcMode is the SVG calcMode attribute of animation elements.
type CalcMode string //#enum

const (
	CalcModeDiscrete CalcMode = "discrete"
	CalcModeLinear   CalcMode = "linear"
	CalcModePaced    CalcMode = "paced"
	CalcModeSpline   CalcMode = "spline"
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

func (v CalcMode) AttribName() string                          { return "calcMode" }
func (v CalcMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Additive is the SVG additive attribute of animation elements.
type Additive string //#enum

const (
	AdditiveReplace Additive = "replace"
	AdditiveSum     Additive = "sum"
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

func (v Additive) AttribName() string                          { return "additive" }
func (v Additive) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Accumulate is the SVG accumulate attribute of animation elements.
type Accumulate string //#enum

const (
	AccumulateNone Accumulate = "none"
	AccumulateSum  Accumulate = "sum"
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

func (v Accumulate) AttribName() string                          { return "accumulate" }
func (v Accumulate) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Restart is the SVG restart attribute of animation elements.
type Restart string //#enum

const (
	RestartAlways        Restart = "always"
	RestartWhenNotActive Restart = "whenNotActive"
	RestartNever         Restart = "never"
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

func (v Restart) AttribName() string                          { return "restart" }
func (v Restart) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Misc

// CrossOrigin is the SVG crossorigin attribute.
type CrossOrigin string //#enum

const (
	CrossOriginAnonymous      CrossOrigin = "anonymous"
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

func (v CrossOrigin) AttribName() string                          { return "crossorigin" }
func (v CrossOrigin) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ZoomAndPan is the SVG zoomAndPan attribute.
type ZoomAndPan string //#enum

const (
	ZoomAndPanDisable ZoomAndPan = "disable"
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

func (v ZoomAndPan) AttribName() string                          { return "zoomAndPan" }
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
