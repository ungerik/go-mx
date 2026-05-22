package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AlertVariant selects an alert's visual style. Class strings are transcribed
// verbatim from shadcn/ui's alert.tsx (new-york-v4, Tailwind v4).
type AlertVariant string

const (
	AlertDefault     AlertVariant = "default"
	AlertDestructive AlertVariant = "destructive"
)

const alertBaseClasses = "relative grid w-full grid-cols-[0_1fr] items-start gap-y-0.5 rounded-lg border px-4 py-3 text-sm has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr] has-[>svg]:gap-x-3 [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current"

var alertVariantClasses = map[AlertVariant]string{
	AlertDefault:     "bg-card text-card-foreground",
	AlertDestructive: "bg-card text-destructive *:data-[slot=alert-description]:text-destructive/90 [&>svg]:text-current",
}

// Alert renders a shadcn/ui alert container. variant may be "" for the
// default. A role="alert" attribute is added unless the caller supplies a role.
func Alert(variant AlertVariant, attribsChildren ...any) *mx.Element {
	if variant == "" {
		variant = AlertDefault
	}
	v, ok := alertVariantClasses[variant]
	if !ok {
		v = alertVariantClasses[AlertDefault]
	}
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("alert"))
	}
	return finish(e, "alert", Cn(alertBaseClasses, v))
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
