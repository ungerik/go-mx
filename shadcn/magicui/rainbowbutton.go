//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// RainbowButtonVariant selects a [RainbowButton]'s look.
type RainbowButtonVariant string //#enum

const (
	// RainbowButtonDefault is a dark button (white in dark mode) with a
	// rainbow border and a blurred rainbow glow below it.
	RainbowButtonDefault RainbowButtonVariant = "default"
	// RainbowButtonOutline is a light button (near-black in dark mode)
	// with the rainbow on its bottom edge and the same glow.
	RainbowButtonOutline RainbowButtonVariant = "outline"
)

// Valid indicates if r is any of the valid values for RainbowButtonVariant
func (r RainbowButtonVariant) Valid() bool {
	switch r {
	case
		RainbowButtonDefault,
		RainbowButtonOutline:
		return true
	}
	return false
}

// Validate returns an error if r is none of the valid values for RainbowButtonVariant
func (r RainbowButtonVariant) Validate() error {
	if !r.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.RainbowButtonVariant", r)
	}
	return nil
}

// ValidOr returns r if it is any of the valid values for RainbowButtonVariant, else defaultVal
func (r RainbowButtonVariant) ValidOr(defaultVal RainbowButtonVariant) RainbowButtonVariant {
	if r.Valid() {
		return r
	}
	return defaultVal
}

// Enums returns all valid values for RainbowButtonVariant
func (RainbowButtonVariant) Enums() []RainbowButtonVariant {
	return []RainbowButtonVariant{
		RainbowButtonDefault,
		RainbowButtonOutline,
	}
}

// EnumStrings returns all valid values for RainbowButtonVariant as strings
func (RainbowButtonVariant) EnumStrings() []string {
	return []string{
		"default",
		"outline",
	}
}

// String implements the fmt.Stringer interface for RainbowButtonVariant
func (r RainbowButtonVariant) String() string {
	return string(r)
}

// RainbowButtonSize selects a [RainbowButton]'s size.
type RainbowButtonSize string //#enum

const (
	// RainbowButtonSizeDefault is the default height of 2.25rem.
	RainbowButtonSizeDefault RainbowButtonSize = "default"
	// RainbowButtonSizeSM is a small button with extra-small text.
	RainbowButtonSizeSM RainbowButtonSize = "sm"
	// RainbowButtonSizeLG is a large button.
	RainbowButtonSizeLG RainbowButtonSize = "lg"
	// RainbowButtonSizeIcon is a square button for a single icon.
	RainbowButtonSizeIcon RainbowButtonSize = "icon"
)

// Valid indicates if r is any of the valid values for RainbowButtonSize
func (r RainbowButtonSize) Valid() bool {
	switch r {
	case
		RainbowButtonSizeDefault,
		RainbowButtonSizeSM,
		RainbowButtonSizeLG,
		RainbowButtonSizeIcon:
		return true
	}
	return false
}

// Validate returns an error if r is none of the valid values for RainbowButtonSize
func (r RainbowButtonSize) Validate() error {
	if !r.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.RainbowButtonSize", r)
	}
	return nil
}

// ValidOr returns r if it is any of the valid values for RainbowButtonSize, else defaultVal
func (r RainbowButtonSize) ValidOr(defaultVal RainbowButtonSize) RainbowButtonSize {
	if r.Valid() {
		return r
	}
	return defaultVal
}

// Enums returns all valid values for RainbowButtonSize
func (RainbowButtonSize) Enums() []RainbowButtonSize {
	return []RainbowButtonSize{
		RainbowButtonSizeDefault,
		RainbowButtonSizeSM,
		RainbowButtonSizeLG,
		RainbowButtonSizeIcon,
	}
}

// EnumStrings returns all valid values for RainbowButtonSize as strings
func (RainbowButtonSize) EnumStrings() []string {
	return []string{
		"default",
		"sm",
		"lg",
		"icon",
	}
}

// String implements the fmt.Stringer interface for RainbowButtonSize
func (r RainbowButtonSize) String() string {
	return string(r)
}

// rainbowButtonVariants resolves a rainbow button's classes, declared the
// same way upstream's rainbow-button.tsx declares them with cva.
var rainbowButtonVariants = cva.New(cva.Config{
	Base: "relative cursor-pointer group transition-all animate-rainbow inline-flex items-center justify-center gap-2 shrink-0 rounded-sm outline-none focus-visible:ring-[3px] aria-invalid:border-destructive text-sm font-medium whitespace-nowrap disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 [&_svg]:shrink-0",
	Variants: map[string]map[string]string{
		"variant": {
			"default": "border-0 bg-[linear-gradient(#121213,#121213),linear-gradient(#121213_50%,rgba(18,18,19,0.6)_80%,rgba(18,18,19,0)),linear-gradient(90deg,var(--rainbow-color-1),var(--rainbow-color-5),var(--rainbow-color-3),var(--rainbow-color-4),var(--rainbow-color-2))] bg-[length:200%] text-primary-foreground [background-clip:padding-box,border-box,border-box] [background-origin:border-box] [border:calc(0.125rem)_solid_transparent] before:absolute before:bottom-[-20%] before:left-1/2 before:z-0 before:h-1/5 before:w-3/5 before:-translate-x-1/2 before:animate-rainbow before:bg-[linear-gradient(90deg,var(--rainbow-color-1),var(--rainbow-color-5),var(--rainbow-color-3),var(--rainbow-color-4),var(--rainbow-color-2))] before:bg-[length:200%] before:[filter:blur(0.75rem)] dark:bg-[linear-gradient(#fff,#fff),linear-gradient(#fff_50%,rgba(255,255,255,0.6)_80%,rgba(0,0,0,0)),linear-gradient(90deg,var(--rainbow-color-1),var(--rainbow-color-5),var(--rainbow-color-3),var(--rainbow-color-4),var(--rainbow-color-2))]",
			"outline": "border border-input border-b-transparent bg-[linear-gradient(#ffffff,#ffffff),linear-gradient(#ffffff_50%,rgba(18,18,19,0.6)_80%,rgba(18,18,19,0)),linear-gradient(90deg,var(--rainbow-color-1),var(--rainbow-color-5),var(--rainbow-color-3),var(--rainbow-color-4),var(--rainbow-color-2))] bg-[length:200%] text-accent-foreground [background-clip:padding-box,border-box,border-box] [background-origin:border-box] before:absolute before:bottom-[-20%] before:left-1/2 before:z-0 before:h-1/5 before:w-3/5 before:-translate-x-1/2 before:animate-rainbow before:bg-[linear-gradient(90deg,var(--rainbow-color-1),var(--rainbow-color-5),var(--rainbow-color-3),var(--rainbow-color-4),var(--rainbow-color-2))] before:bg-[length:200%] before:[filter:blur(0.75rem)] dark:bg-[linear-gradient(#0a0a0a,#0a0a0a),linear-gradient(#0a0a0a_50%,rgba(255,255,255,0.6)_80%,rgba(0,0,0,0)),linear-gradient(90deg,var(--rainbow-color-1),var(--rainbow-color-5),var(--rainbow-color-3),var(--rainbow-color-4),var(--rainbow-color-2))]",
		},
		"size": {
			"default": "h-9 px-4 py-2",
			"sm":      "h-8 rounded-xl px-3 text-xs",
			"lg":      "h-11 rounded-xl px-8",
			"icon":    "size-9",
		},
	},
	DefaultVariants: map[string]string{"variant": "default", "size": "default"},
})

// RainbowButtonClasses returns the merged class string of a [RainbowButton],
// the equivalent of upstream's exported rainbowButtonVariants. Use it to
// give the rainbow look to another element, such as a link — upstream's
// asChild prop does that, and go-mx has no Slot:
//
//	html.A(html.HRef("/pricing"), html.Class(magicui.RainbowButtonClasses("", "")), "Pricing")
//
// An empty or unknown variant or size resolves to the default.
func RainbowButtonClasses(variant RainbowButtonVariant, size RainbowButtonSize) string {
	return shadcn.Cn(rainbowButtonVariants(map[string]string{
		"variant": string(variant.ValidOr(RainbowButtonDefault)),
		"size":    string(size.ValidOr(RainbowButtonSizeDefault)),
	}))
}

// RainbowButtonOptions configures a [RainbowButton]. The zero value is
// upstream's default variant and size.
type RainbowButtonOptions struct {
	// Variant is the look; "" means [RainbowButtonDefault].
	Variant RainbowButtonVariant
	// Size is the size; "" means [RainbowButtonSizeDefault].
	Size RainbowButtonSize
}

// RainbowButton renders a button whose border and glow cycle through the
// rainbow colors --rainbow-color-1 to --rainbow-color-5 (defined by [CSS] on :root) at the
// speed of the CSS variable --speed (2s). Its classes come from
// [RainbowButtonClasses]; a caller class wins over them.
//
// Divergences from upstream: there is no asChild prop (use
// [RainbowButtonClasses] on the other element), data-slot is
// "rainbow-button" instead of "button", and like [shadcn.Button] it
// defaults to type="button"; pass html.Type to override.
//
// Needs [CSS] in the app's Tailwind input.
func RainbowButton(opts RainbowButtonOptions, attribsChildren ...any) *mx.Element {
	root := html.Button(attribsChildren...)
	if root.AttribIndex("type") < 0 {
		root.Attribs = append(root.Attribs, html.Type("button"))
	}
	return finish(root, "rainbow-button", RainbowButtonClasses(opts.Variant, opts.Size), "")
}
