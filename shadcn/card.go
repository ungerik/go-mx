package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Card renders a shadcn/ui card container. Compose it with the card part
// functions: [CardHeader], [CardTitle], [CardDescription], [CardAction],
// [CardContent] and [CardFooter].
func Card(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "card",
		"bg-card text-card-foreground flex flex-col gap-6 rounded-xl border py-6 shadow-sm")
}

// CardHeader renders a card's header row. When it contains a [CardAction] the
// header switches to a two-column grid.
func CardHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "card-header",
		"@container/card-header grid auto-rows-min grid-rows-[auto_auto] items-start gap-1.5 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6")
}

// CardTitle renders a card's title. shadcn/ui uses a div here, not a heading
// element, which this port keeps for parity.
func CardTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "card-title", "leading-none font-semibold")
}

// CardDescription renders a card's description text.
func CardDescription(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "card-description", "text-muted-foreground text-sm")
}

// CardAction renders an action slot in the card header, placed in the header
// grid's second column.
func CardAction(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "card-action",
		"col-start-2 row-span-2 row-start-1 self-start justify-self-end")
}

// CardContent renders a card's main content area.
func CardContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "card-content", "px-6")
}

// CardFooter renders a card's footer row.
func CardFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "card-footer", "flex items-center px-6 [.border-t]:pt-6")
}
