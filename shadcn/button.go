package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// ButtonVariant selects a button's visual style. Class strings are transcribed
// verbatim from shadcn/ui's button.tsx (new-york-v4, Tailwind v4).
type ButtonVariant string

// ButtonSize selects a button's size.
type ButtonSize string

const (
	ButtonDefault     ButtonVariant = "default"
	ButtonDestructive ButtonVariant = "destructive"
	ButtonOutline     ButtonVariant = "outline"
	ButtonSecondary   ButtonVariant = "secondary"
	ButtonGhost       ButtonVariant = "ghost"
	ButtonLink        ButtonVariant = "link"
)

const (
	SizeDefault ButtonSize = "default"
	SizeXS      ButtonSize = "xs"
	SizeSM      ButtonSize = "sm"
	SizeLG      ButtonSize = "lg"
	SizeIcon    ButtonSize = "icon"
	SizeIconXS  ButtonSize = "icon-xs"
	SizeIconSM  ButtonSize = "icon-sm"
	SizeIconLG  ButtonSize = "icon-lg"
)

// buttonVariants resolves a button's base + variant + size classes, declared
// the same way shadcn/ui's button.tsx declares them with cva.
var buttonVariants = cva.New(cva.Config{
	Base: "inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap transition-all outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
	Variants: map[string]map[string]string{
		"variant": {
			"default":     "bg-primary text-primary-foreground hover:bg-primary/90",
			"destructive": "bg-destructive text-white hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:bg-destructive/60 dark:focus-visible:ring-destructive/40",
			"outline":     "border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50",
			"secondary":   "bg-secondary text-secondary-foreground hover:bg-secondary/80",
			"ghost":       "hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50",
			"link":        "text-primary underline-offset-4 hover:underline",
		},
		"size": {
			"default": "h-9 px-4 py-2 has-[>svg]:px-3",
			"xs":      "h-6 gap-1 rounded-md px-2 text-xs has-[>svg]:px-1.5 [&_svg:not([class*='size-'])]:size-3",
			"sm":      "h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5",
			"lg":      "h-10 rounded-md px-6 has-[>svg]:px-4",
			"icon":    "size-9",
			"icon-xs": "size-6 rounded-md [&_svg:not([class*='size-'])]:size-3",
			"icon-sm": "size-8",
			"icon-lg": "size-10",
		},
	},
	DefaultVariants: map[string]string{"variant": "default", "size": "default"},
})

// ButtonClasses returns the merged base + variant + size button class string.
// It is the equivalent of shadcn/ui's exported buttonVariants: use it to give
// the button look to a non-button element or to an [AlertDialogTrigger], e.g.
// html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeSM)).
//
// An empty variant or size resolves to the default; an unknown value falls
// back to the default classes.
func ButtonClasses(variant ButtonVariant, size ButtonSize) string {
	return Cn(buttonVariants(map[string]string{
		"variant": normButtonVariant(variant),
		"size":    normButtonSize(size),
	}))
}

// normButtonVariant maps an empty or unknown variant to the default.
func normButtonVariant(v ButtonVariant) string {
	switch v {
	case ButtonDestructive, ButtonOutline, ButtonSecondary, ButtonGhost, ButtonLink:
		return string(v)
	default:
		return string(ButtonDefault)
	}
}

// normButtonSize maps an empty or unknown size to the default.
func normButtonSize(s ButtonSize) string {
	switch s {
	case SizeXS, SizeSM, SizeLG, SizeIcon, SizeIconXS, SizeIconSM, SizeIconLG:
		return string(s)
	default:
		return string(SizeDefault)
	}
}

// Button renders a shadcn/ui button. variant and size may be "" to select the
// defaults. attribsChildren are go-mx attributes and children; a caller-
// supplied class attribute is merged with the variant classes.
//
// The button defaults to type="button" (a server-rendered <button> without an
// explicit type would otherwise submit an enclosing form); pass html.Type to
// override.
func Button(variant ButtonVariant, size ButtonSize, attribsChildren ...any) *mx.Element {
	if variant == "" {
		variant = ButtonDefault
	}
	if size == "" {
		size = SizeDefault
	}
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	e.Attribs = append(e.Attribs,
		html.DataAttr("variant", string(variant)),
		html.DataAttr("size", string(size)),
	)
	return finish(e, "button", ButtonClasses(variant, size))
}
