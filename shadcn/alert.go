package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// AlertVariant selects an alert's visual style. Class strings are transcribed
// verbatim from shadcn/ui's alert.tsx (new-york-v4, Tailwind v4).
type AlertVariant string // TODO use go-enum

const (
	// AlertDefault is the default alert style, rendered on the card background.
	AlertDefault AlertVariant = "default"
	// AlertDestructive styles the alert to signal an error or destructive condition.
	AlertDestructive AlertVariant = "destructive"
)

// alertVariants resolves an alert's base + variant classes, declared the same
// way shadcn/ui's alert.tsx declares them with cva.
var alertVariants = cva.New(cva.Config{
	Base: "relative grid w-full grid-cols-[0_1fr] items-start gap-y-0.5 rounded-lg border px-4 py-3 text-sm has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr] has-[>svg]:gap-x-3 [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current",
	Variants: map[string]map[string]string{
		"variant": {
			"default":     "bg-card text-card-foreground",
			"destructive": "bg-card text-destructive *:data-[slot=alert-description]:text-destructive/90 [&>svg]:text-current",
		},
	},
	DefaultVariants: map[string]string{"variant": "default"},
})

// Alert renders a shadcn/ui alert container. variant may be "" for the
// default. A role="alert" attribute is added unless the caller supplies a role.
func Alert(variant AlertVariant, attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("alert"))
	}
	v := string(AlertDefault)
	if variant == AlertDestructive {
		v = string(AlertDestructive)
	}
	return finish(e, "alert", alertVariants(map[string]string{"variant": v}))
}

// AlertTitle renders the title row of an [Alert]. shadcn/ui uses a div here
// (not a heading element), which this port keeps for parity.
func AlertTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "alert-title",
		"col-start-2 line-clamp-1 min-h-4 font-medium tracking-tight")
}

// AlertDescription renders the description row of an [Alert].
func AlertDescription(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "alert-description",
		"col-start-2 grid justify-items-start gap-1 text-sm text-muted-foreground [&_p]:leading-relaxed")
}
