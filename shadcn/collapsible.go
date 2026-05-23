package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Collapsible renders a shadcn/ui collapsible region as a native <details>
// element. shadcn's Collapsible is Radix-driven; this port replaces Radix with
// <details>/<summary>, which the browser provides for free — no JavaScript,
// no client framework.
//
// The element carries the group base class so descendants of [CollapsibleTrigger]
// (e.g. a chevron icon) can react to the open state with the group-open:
// variant — the native equivalent of shadcn's data-[state=open]: selector.
//
// shadcn's open-state height animation is not reproduced: a native <details>
// snaps open. It can be brought back in plain CSS with @starting-style and
// transition-behavior: allow-discrete; this package does not emit that CSS.
//
// Pass html.Open to start in the open state.
func Collapsible(attribsChildren ...any) *mx.Element {
	return finish(html.Details(attribsChildren...), "collapsible", "group")
}

// CollapsibleTrigger renders the always-visible disclosure trigger as a
// <summary>. The default browser disclosure marker (the ▶ triangle) is hidden
// so callers can supply their own chevron — typically with the group-open:
// variant to rotate it when [Collapsible] is open.
func CollapsibleTrigger(attribsChildren ...any) *mx.Element {
	return finish(html.Summary(attribsChildren...), "collapsible-trigger",
		"list-none cursor-pointer [&::-webkit-details-marker]:hidden")
}

// CollapsibleContent renders the content shown when [Collapsible] is open as a
// plain <div>. shadcn's Radix-driven height animation is dropped; the content
// snaps in.
func CollapsibleContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "collapsible-content", "")
}
