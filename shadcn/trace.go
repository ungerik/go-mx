//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Trace is a Go port of shadcn/ui's Trace, a vertical trail of an agent's
// steps or tool calls — the shape an agent transcript streamed over
// [mx.SSEResponse] shows its work in. Class strings are reconstructed from
// the post-July-2026 registry (trace.tsx skeleton + the style-vega.css
// .cn-trace* rules) — see "Upstream restructure (July 2026)" in TODOS.md.
//
// Rows are [TraceLine]s, either directly in the Trace or as the only child of
// an <a> or <button> (or of a trigger such as [SheetTrigger]), which then
// gets the row's focus ring. A [TraceCollapsible] row discloses details.
//
// Upstream's shimmer utility on a running row's title comes from shadcn's own
// tailwind.css, not from Tailwind, so by default the title pulses with
// animate-pulse instead (not under prefers-reduced-motion); with
// [UseUtilities] it shimmers.
func Trace(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "trace",
		"flex w-full min-w-0 flex-col *:has-[>[data-slot=trace-line]]:flex *:has-[>[data-slot=trace-line]]:w-fit *:has-[>[data-slot=trace-line]]:max-w-full *:has-[>[data-slot=trace-line]]:min-w-0 *:has-[>[data-slot=trace-line]]:text-left *:has-[>[data-slot=trace-line]]:outline-none gap-2 *:has-[>[data-slot=trace-line]]:rounded-md *:has-[>[data-slot=trace-line]]:focus-visible:ring-3 *:has-[>[data-slot=trace-line]]:focus-visible:ring-ring/50")
}

// TraceStatus is the state of a [TraceLine]'s step. A finished step has none.
type TraceStatus string //#enum

const (
	// TraceRunning pulses the row's [TraceTitle].
	TraceRunning TraceStatus = "running"
	// TraceError colors the row, and a [TraceCollapsibleTrigger] holding it,
	// destructive.
	TraceError TraceStatus = "error"
	// TracePending dims a step that has not started.
	TracePending TraceStatus = "pending"
)

// Valid indicates if t is any of the valid values for TraceStatus
func (t TraceStatus) Valid() bool {
	switch t {
	case
		TraceRunning,
		TraceError,
		TracePending:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for TraceStatus
func (t TraceStatus) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.TraceStatus", t)
	}
	return nil
}

// ValidOr returns t if it is any of the valid values for TraceStatus, else defaultVal
func (t TraceStatus) ValidOr(defaultVal TraceStatus) TraceStatus {
	if t.Valid() {
		return t
	}
	return defaultVal
}

// Enums returns all valid values for TraceStatus
func (TraceStatus) Enums() []TraceStatus {
	return []TraceStatus{
		TraceRunning,
		TraceError,
		TracePending,
	}
}

// EnumStrings returns all valid values for TraceStatus as strings
func (TraceStatus) EnumStrings() []string {
	return []string{
		"running",
		"error",
		"pending",
	}
}

// String implements the fmt.Stringer interface for TraceStatus
func (t TraceStatus) String() string {
	return string(t)
}

// TraceLine renders one step as a row of [TraceIcon], [TraceTitle] and
// [TraceDescription]. status is "" for a finished step, which renders no
// data-status; an unknown value renders none either.
//
// Inside an <a>, <button> or <summary> (a [TraceCollapsibleTrigger]) the row
// turns to the foreground color on hover. Upstream lists only a and button,
// because its collapsible trigger is a <button>; this port adds summary.
func TraceLine(status TraceStatus, attribsChildren ...any) *mx.Element {
	e := html.Span(attribsChildren...)
	if status.Valid() {
		e.Attribs = append(e.Attribs, html.DataAttr("status", string(status)))
	}
	return finish(e, "trace-line",
		"flex w-fit max-w-full min-w-0 items-center text-muted-foreground transition-colors in-[a,button,summary]:cursor-pointer in-[a:hover,button:hover,summary:hover]:text-foreground data-[status=error]:text-destructive data-[status=error]:in-[a:hover,button:hover,summary:hover]:text-destructive data-[status=pending]:opacity-60 "+utilityClasses("data-[status=running]:*:data-[slot=trace-title]:shimmer motion-reduce:shimmer-none", "motion-safe:data-[status=running]:*:data-[slot=trace-title]:animate-pulse")+" gap-2 text-sm")
}

// traceIconClasses is shared by [TraceIcon] and the chevron of
// [TraceCollapsibleTrigger].
const traceIconClasses = "flex shrink-0 items-center justify-center [&_svg]:pointer-events-none size-4 [&_svg:not([class*='size-'])]:size-4 [[data-slot=trace-title]~&]:-ml-1"

// TraceIcon holds a [TraceLine]'s icon. It is hidden from assistive
// technology, like upstream. An icon after the [TraceTitle] (a trailing
// chevron) moves closer to it.
func TraceIcon(attribsChildren ...any) *mx.Element {
	e := html.Span(attribsChildren...)
	if e.AttribIndex("aria-hidden") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-hidden", "true"))
	}
	return finish(e, "trace-icon", traceIconClasses)
}

// TraceTitle renders a [TraceLine]'s title, truncated to one line.
func TraceTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "trace-title", "min-w-0 truncate")
}

// TraceDescription renders a [TraceLine]'s secondary text, such as a
// duration, after the title.
func TraceDescription(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "trace-description",
		"shrink-0 text-muted-foreground/70 tabular-nums")
}

// TraceCollapsible renders a step that discloses its details: a native
// <details> ([Collapsible]) holding a [TraceCollapsibleTrigger] and a
// [TraceCollapsibleContent]. Pass html.Open to start open. Upstream's
// controlled open/onOpenChange has no counterpart; the browser owns the
// state.
func TraceCollapsible(attribsChildren ...any) *mx.Element {
	return finish(Collapsible(attribsChildren...), "trace-collapsible",
		"flex w-full min-w-0 flex-col")
}

// TraceCollapsibleTrigger renders the always-visible row of a
// [TraceCollapsible] as a <summary> ([CollapsibleTrigger]) holding a
// [TraceLine], and appends the chevron that turns when the step is open.
// The turn keys on the <details> directly around it (upstream's
// group-aria-expanded becomes [details[open]>summary>&]), so a nested
// TraceCollapsible does not turn the outer chevron.
func TraceCollapsibleTrigger(attribsChildren ...any) *mx.Element {
	chevron := finish(html.Span(html.Attrib("aria-hidden", "true"), iconChevronRight()), "trace-icon",
		Cn(traceIconClasses, "transition-transform [details[open]>summary>&]:rotate-90"))
	return finish(CollapsibleTrigger(append(attribsChildren, chevron)...), "trace-collapsible-trigger",
		"group/trace-collapsible-trigger flex w-fit max-w-full min-w-0 cursor-pointer items-center text-left text-muted-foreground transition-colors outline-none hover:text-foreground has-[[data-status=error]]:text-destructive gap-1 rounded-md focus-visible:ring-3 focus-visible:ring-ring/50")
}

// TraceCollapsibleContent renders the details of a [TraceCollapsible],
// indented behind a rule. As upstream does with className, a caller class
// goes to the inner <div> that holds the children; every other attribute
// goes to the outer [CollapsibleContent]. Upstream's height transition is
// not reproduced: a native <details> snaps open.
func TraceCollapsibleContent(attribsChildren ...any) *mx.Element {
	e := CollapsibleContent(attribsChildren...)
	inner := html.Div(takeClassAttribs(e)...)
	inner.Children = e.Children
	e.Children = mx.Components{finish(inner, "trace-collapsible-content-inner",
		"min-w-0 wrap-break-word text-muted-foreground mt-2 mb-1 ml-[calc(--spacing(2)-1px)] border-l pr-1 pl-4 leading-relaxed")}
	return finish(e, "trace-collapsible-content", "min-w-0 overflow-hidden")
}
