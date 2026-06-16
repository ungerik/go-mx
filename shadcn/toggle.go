//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// ToggleVariant selects a toggle's visual style. Class strings are transcribed
// verbatim from shadcn/ui's toggle.tsx (new-york-v4, Tailwind v4).
type ToggleVariant string //#enum

const (
	// ToggleDefault is the default toggle style, transparent until pressed.
	ToggleDefault ToggleVariant = "default"
	// ToggleOutline styles the toggle with a visible border.
	ToggleOutline ToggleVariant = "outline"
)

// Valid indicates if t is any of the valid values for ToggleVariant
func (t ToggleVariant) Valid() bool {
	switch t {
	case
		ToggleDefault,
		ToggleOutline:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for ToggleVariant
func (t ToggleVariant) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ToggleVariant", t)
	}
	return nil
}

// Enums returns all valid values for ToggleVariant
func (ToggleVariant) Enums() []ToggleVariant {
	return []ToggleVariant{
		ToggleDefault,
		ToggleOutline,
	}
}

// EnumStrings returns all valid values for ToggleVariant as strings
func (ToggleVariant) EnumStrings() []string {
	return []string{
		"default",
		"outline",
	}
}

// String implements the fmt.Stringer interface for ToggleVariant
func (t ToggleVariant) String() string {
	return string(t)
}

// ToggleSize selects a toggle's size. Distinct from [ButtonSize] because the
// toggle's size table is its own (no icon sizes, different padding).
type ToggleSize string //#enum

const (
	// ToggleSizeDefault is the default toggle size.
	ToggleSizeDefault ToggleSize = "default"
	// ToggleSizeSM is the small toggle size.
	ToggleSizeSM ToggleSize = "sm"
	// ToggleSizeLG is the large toggle size.
	ToggleSizeLG ToggleSize = "lg"
)

// Valid indicates if t is any of the valid values for ToggleSize
func (t ToggleSize) Valid() bool {
	switch t {
	case
		ToggleSizeDefault,
		ToggleSizeSM,
		ToggleSizeLG:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for ToggleSize
func (t ToggleSize) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ToggleSize", t)
	}
	return nil
}

// Enums returns all valid values for ToggleSize
func (ToggleSize) Enums() []ToggleSize {
	return []ToggleSize{
		ToggleSizeDefault,
		ToggleSizeSM,
		ToggleSizeLG,
	}
}

// EnumStrings returns all valid values for ToggleSize as strings
func (ToggleSize) EnumStrings() []string {
	return []string{
		"default",
		"sm",
		"lg",
	}
}

// String implements the fmt.Stringer interface for ToggleSize
func (t ToggleSize) String() string {
	return string(t)
}

// toggleVariants resolves a toggle's base + variant + size classes, declared
// the same way shadcn/ui's toggle.tsx declares them with cva. The Radix
// data-[state=on]:* selectors are rewritten to the native aria-pressed:*
// variant since this port renders a <button aria-pressed> rather than wrapping
// Radix's TogglePrimitive.Root.
var toggleVariants = cva.New(cva.Config{
	Base: "inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium hover:bg-muted hover:text-muted-foreground disabled:pointer-events-none disabled:opacity-50 aria-pressed:bg-accent aria-pressed:text-accent-foreground [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none transition-[color,box-shadow] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive whitespace-nowrap",
	Variants: map[string]map[string]string{
		"variant": {
			"default": "bg-transparent",
			"outline": "border border-input bg-transparent shadow-xs hover:bg-accent hover:text-accent-foreground",
		},
		"size": {
			"default": "h-9 px-2 min-w-9",
			"sm":      "h-8 px-1.5 min-w-8",
			"lg":      "h-10 px-2.5 min-w-10",
		},
	},
	DefaultVariants: map[string]string{"variant": "default", "size": "default"},
})

// ToggleClasses returns the merged base + variant + size toggle class string.
// It is the equivalent of shadcn/ui's exported toggleVariants and mirrors
// [ButtonClasses]: use it to style a non-button element (or a [ToggleGroupItem])
// with the toggle look. An empty variant or size resolves to the default.
func ToggleClasses(variant ToggleVariant, size ToggleSize) string {
	return Cn(toggleVariants(map[string]string{
		"variant": normToggleVariant(variant),
		"size":    normToggleSize(size),
	}))
}

// normToggleVariant maps an empty or unknown variant to the default.
func normToggleVariant(v ToggleVariant) string {
	if v == ToggleOutline {
		return string(ToggleOutline)
	}
	return string(ToggleDefault)
}

// normToggleSize maps an empty or unknown size to the default.
func normToggleSize(s ToggleSize) string {
	switch s {
	case ToggleSizeSM, ToggleSizeLG:
		return string(s)
	default:
		return string(ToggleSizeDefault)
	}
}

// toggleAriaPressedFlip is the default onclick that flips aria-pressed in
// place. Toggle uses it when the caller passes neither an onclick nor any
// htmx attribute.
const toggleAriaPressedFlip = /*js*/ `this.setAttribute('aria-pressed', this.getAttribute('aria-pressed')!=='true')`

// Toggle renders a shadcn/ui toggle as a <button aria-pressed>. variant and
// size may be "" to select the defaults. attribsChildren are go-mx attributes
// and children; a caller-supplied class is merged with the variant classes.
//
// Default attributes (overridable): type="button" (a server-rendered <button>
// without an explicit type would otherwise submit an enclosing form), and
// aria-pressed="false". When no caller onclick is supplied and no htmx
// attribute is present, Toggle adds a default onclick that flips aria-pressed
// in place — pass html.OnClick("…") to replace it, or pass any hx.* attribute
// (e.g. hx.Post(...)) to opt into a server round-trip; in either case Toggle
// stays out of the way.
func Toggle(variant ToggleVariant, size ToggleSize, attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("aria-pressed") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-pressed", "false"))
	}
	if e.AttribIndex("onclick") < 0 && !hasHX(e) {
		e.Attribs = append(e.Attribs, html.OnClick(toggleAriaPressedFlip))
	}
	e.Attribs = append(e.Attribs,
		html.DataAttr("variant", normToggleVariant(variant)),
		html.DataAttr("size", normToggleSize(size)),
	)
	return finish(e, "toggle", ToggleClasses(variant, size))
}
