package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
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

const buttonBaseClasses = "inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap transition-all outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4"

var buttonVariantClasses = map[ButtonVariant]string{
	ButtonDefault:     "bg-primary text-primary-foreground hover:bg-primary/90",
	ButtonDestructive: "bg-destructive text-white hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:bg-destructive/60 dark:focus-visible:ring-destructive/40",
	ButtonOutline:     "border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50",
	ButtonSecondary:   "bg-secondary text-secondary-foreground hover:bg-secondary/80",
	ButtonGhost:       "hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50",
	ButtonLink:        "text-primary underline-offset-4 hover:underline",
}

var buttonSizeClasses = map[ButtonSize]string{
	SizeDefault: "h-9 px-4 py-2 has-[>svg]:px-3",
	SizeXS:      "h-6 gap-1 rounded-md px-2 text-xs has-[>svg]:px-1.5 [&_svg:not([class*='size-'])]:size-3",
	SizeSM:      "h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5",
	SizeLG:      "h-10 rounded-md px-6 has-[>svg]:px-4",
	SizeIcon:    "size-9",
	SizeIconXS:  "size-6 rounded-md [&_svg:not([class*='size-'])]:size-3",
	SizeIconSM:  "size-8",
	SizeIconLG:  "size-10",
}

// ButtonClasses returns the merged base + variant + size button class string.
// It is the equivalent of shadcn/ui's exported buttonVariants: use it to give
// the button look to a non-button element or to an [AlertDialogTrigger], e.g.
// html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeSM)).
//
// An empty variant or size resolves to the default; an unknown value falls
// back to the default classes.
func ButtonClasses(variant ButtonVariant, size ButtonSize) string {
	if variant == "" {
		variant = ButtonDefault
	}
	if size == "" {
		size = SizeDefault
	}
	v, ok := buttonVariantClasses[variant]
	if !ok {
		v = buttonVariantClasses[ButtonDefault]
	}
	s, ok := buttonSizeClasses[size]
	if !ok {
		s = buttonSizeClasses[SizeDefault]
	}
	return Cn(buttonBaseClasses, v, s)
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
