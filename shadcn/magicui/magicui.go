//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/svg"
)

// UpstreamCommit is the magicuidesign/magicui commit the components and
// [CSS] were transcribed from.
const UpstreamCommit = "a54d9edd7bd01dfb4fa72641df2b84da449a5e35"

//go:embed css/*.css
var cssFiles embed.FS

// CSS holds the Tailwind v4 rules every component of this package needs:
// the @theme --animate-* entries with their @keyframes, the @utility rules
// of the scroll-driven animations, and the prefers-reduced-motion block
// that switches them all off. Add it to the app's Tailwind input, after
// @import "tailwindcss"; without it the components render but do not move.
//
// It is assembled from one file per component (css/*.css in the package
// source), in file-name order.
var CSS = assembleCSS()

const cssHeader = `/*
 * Magic UI animations for github.com/ungerik/go-mx/shadcn/magicui.
 * Transcribed from magicuidesign/magicui at ` + UpstreamCommit + `
 * (MIT License, see THIRD-PARTY-LICENSES.md), with the motion-library
 * animations re-expressed as CSS. Add it to the Tailwind CSS input of the
 * app, after @import "tailwindcss".
 */
`

func assembleCSS() string {
	names, err := fs.Glob(cssFiles, "css/*.css")
	if err != nil {
		panic(err) // only for a malformed pattern
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString(cssHeader)
	for _, name := range names {
		data, err := cssFiles.ReadFile(name)
		if err != nil {
			panic(err) // embedded files are always readable
		}
		b.WriteString("\n")
		b.Write(data)
	}
	return b.String()
}

// Direction is the direction an element moves in or comes from.
type Direction string //#enum

const (
	// DirectionUp moves upwards.
	DirectionUp Direction = "up"
	// DirectionDown moves downwards.
	DirectionDown Direction = "down"
	// DirectionLeft moves to the left.
	DirectionLeft Direction = "left"
	// DirectionRight moves to the right.
	DirectionRight Direction = "right"
)

// Valid indicates if d is any of the valid values for Direction
func (d Direction) Valid() bool {
	switch d {
	case
		DirectionUp,
		DirectionDown,
		DirectionLeft,
		DirectionRight:
		return true
	}
	return false
}

// Validate returns an error if d is none of the valid values for Direction
func (d Direction) Validate() error {
	if !d.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.Direction", d)
	}
	return nil
}

// ValidOr returns d if it is any of the valid values for Direction, else defaultVal
func (d Direction) ValidOr(defaultVal Direction) Direction {
	if d.Valid() {
		return d
	}
	return defaultVal
}

// Enums returns all valid values for Direction
func (Direction) Enums() []Direction {
	return []Direction{
		DirectionUp,
		DirectionDown,
		DirectionLeft,
		DirectionRight,
	}
}

// EnumStrings returns all valid values for Direction as strings
func (Direction) EnumStrings() []string {
	return []string{
		"up",
		"down",
		"left",
		"right",
	}
}

// String implements the fmt.Stringer interface for Direction
func (d Direction) String() string {
	return string(d)
}

// checkCSSValue returns an error if value could escape the CSS declaration
// it is placed in: a ';', a brace, an angle bracket, a backslash, a quote,
// a comment delimiter or an unbalanced parenthesis or bracket that would
// swallow the declarations after it, or a url( that would load a resource.
func checkCSSValue(name, value string) error {
	if strings.ContainsAny(value, ";{}<>\\\"'") ||
		strings.Contains(value, "/*") || strings.Contains(value, "*/") ||
		strings.Contains(strings.ToLower(value), "url(") ||
		!balancedCSSBrackets(value) {
		return fmt.Errorf("magicui: invalid CSS value %q for %s", value, name)
	}
	return nil
}

// balancedCSSBrackets reports whether every '(' and '[' in value is closed
// by its matching ')' or ']' in order. CSS keeps a ';' that follows an open
// function or block inside the value, so an unclosed one would swallow the
// declarations after it, and a stray closer would end a function early.
func balancedCSSBrackets(value string) bool {
	var open []rune
	for _, r := range value {
		switch r {
		case '(', '[':
			open = append(open, r)
		case ')', ']':
			want := '('
			if r == ']' {
				want = '['
			}
			if len(open) == 0 || open[len(open)-1] != want {
				return false
			}
			open = open[:len(open)-1]
		}
	}
	return len(open) == 0
}

// styleVars formats name/value pairs as CSS declarations for a style
// attribute, skipping pairs with an empty value. It returns an error for a
// value [checkCSSValue] rejects.
func styleVars(pairs ...string) (string, error) {
	var b strings.Builder
	for i := 0; i+1 < len(pairs); i += 2 {
		name, value := pairs[i], pairs[i+1]
		if value == "" {
			continue
		}
		if err := checkCSSValue(name, value); err != nil {
			return "", err
		}
		if b.Len() > 0 {
			b.WriteString(";")
		}
		b.WriteString(name)
		b.WriteString(":")
		b.WriteString(value)
	}
	return b.String(), nil
}

// finish completes a component's root element like [shadcn.StyleSlot] and
// also merges style: the component's own declarations come first and the
// caller's style attributes after them, so a caller declaration wins. The
// caller's style values are read at render time, with the render context.
func finish(e *mx.Element, slot, baseClasses, style string) *mx.Element {
	if e.Err != nil {
		return e
	}
	merged := mergedStyle{own: style}
	attribs := e.Attribs[:0:0]
	for _, a := range e.Attribs {
		if mx.IsNilAttrib(a) || a.AttribName() != "style" {
			attribs = append(attribs, a)
			continue
		}
		merged.caller = append(merged.caller, a)
	}
	if merged.own != "" || len(merged.caller) > 0 {
		attribs = append(attribs, merged)
	}
	e.Attribs = attribs
	return shadcn.StyleSlot(e, slot, baseClasses)
}

// mergedStyle is the style attribute [finish] gives a component: its own
// declarations followed by those of the caller's style attributes, which
// are valued with the render context.
type mergedStyle struct {
	own    string
	caller []mx.Attrib
}

func (mergedStyle) AttribName() string { return "style" }

func (s mergedStyle) AttribValue(ctx context.Context) (string, error) {
	declarations := make([]string, 0, 1+len(s.caller))
	if s.own != "" {
		declarations = append(declarations, s.own)
	}
	for _, a := range s.caller {
		value, err := a.AttribValue(ctx)
		if err != nil {
			return "", err
		}
		if value = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(value), ";")); value != "" {
			declarations = append(declarations, value)
		}
	}
	return strings.Join(declarations, ";"), nil
}

// seconds formats d as a CSS time in seconds, def if d is not positive.
func seconds(d, def time.Duration) string {
	if d <= 0 {
		d = def
	}
	return strconv.FormatFloat(d.Seconds(), 'f', -1, 64) + "s"
}

// num formats f as a CSS number without exponent or trailing zeros.
func num(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// px formats f as a CSS length in pixels.
func px(f float64) string {
	return num(f) + "px"
}

// or returns value, or def if value is empty.
func or[T comparable](value, def T) T {
	var zero T
	if value == zero {
		return def
	}
	return value
}

// newRand returns the deterministic generator a component draws its
// "random" positions from, so the same seed renders the same markup.
func newRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
}

// ariaHidden hides decorative markup and duplicated content from assistive
// technology.
var ariaHidden = mx.NewAttrib("aria-hidden", "true")

// anys converts components to the ...any of an element constructor.
func anys(components []mx.Component) []any {
	result := make([]any, len(components))
	for i, c := range components {
		result[i] = c
	}
	return result
}

// svgRefID returns id, or a process-unique id from [mx.UniqueID] if
// id is empty. The SVG components of this package (the patterns, NoiseTexture,
// Backlight and the device mockups) use it for the element they reference as url(#id), so an id
// with a character outside [mx.ValidIDRune], which could end the url(), is
// an error.
func svgRefID(id string) (string, error) {
	if id == "" {
		return mx.UniqueID().AttribValue(context.Background())
	}
	if i := strings.IndexFunc(id, func(r rune) bool { return !mx.ValidIDRune(r) }); i >= 0 {
		return "", fmt.Errorf("magicui: invalid id %q: only letters, digits, '_' and '-' are allowed", id)
	}
	return id, nil
}

// derefOr returns *p, or def if p is nil.
func derefOr[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

// patternSVG builds the aria-hidden <svg> root shared by the pattern
// components: the caller's attributes, then parts followed by the caller's
// children, finished with slot and baseClasses.
func patternSVG(slot, baseClasses string, parts []mx.Component, attribsChildren []any) *mx.Element {
	root := svg.SVG(append([]any{ariaHidden}, attribsChildren...)...)
	root.Children = append(parts, root.Children...)
	return finish(root, slot, baseClasses, "")
}

// secondsRounded formats s as a CSS time in seconds, rounded to
// milliseconds.
func secondsRounded(s float64) string {
	return num(math.Round(s*1000)/1000) + "s"
}

// numRounded formats f as a CSS number rounded to a millionth, so 0.24-0.03
// renders as 0.21 rather than 0.21000000000000002.
func numRounded(f float64) string {
	return num(math.Round(f*1e6) / 1e6)
}

// childText returns the concatenated string children of e, the text of the
// text components ([TextAnimate], [TextReveal], [SpinningText],
// [Text3DFlip], [LineShadowText]), or an error naming component if a child
// is not text.
func childText(component string, e *mx.Element) (string, error) {
	var text strings.Builder
	for _, child := range e.Children {
		t, ok := child.(mx.Text)
		if !ok {
			return "", fmt.Errorf("magicui: %s children must be text, got %T", component, child)
		}
		text.WriteString(string(t))
	}
	return text.String(), nil
}
