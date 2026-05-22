package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Table renders a shadcn/ui table. The <table> is wrapped in a
// data-slot="table-container" <div> that provides horizontal overflow
// scrolling; caller attributes and children are placed on the inner <table>.
// Compose it with [TableHeader], [TableBody], [TableFooter], [TableRow],
// [TableHead], [TableCell] and [TableCaption].
func Table(attribsChildren ...any) *mx.Element {
	table := finish(html.Table(attribsChildren...), "table", "w-full caption-bottom text-sm")
	return finish(html.Div(table), "table-container", "relative w-full overflow-x-auto")
}

// TableHeader renders a table's <thead>.
func TableHeader(attribsChildren ...any) *mx.Element {
	return finish(html.THead(attribsChildren...), "table-header", "[&_tr]:border-b")
}

// TableBody renders a table's <tbody>.
func TableBody(attribsChildren ...any) *mx.Element {
	return finish(html.TBody(attribsChildren...), "table-body", "[&_tr:last-child]:border-0")
}

// TableFooter renders a table's <tfoot>.
func TableFooter(attribsChildren ...any) *mx.Element {
	return finish(html.TFoot(attribsChildren...), "table-footer",
		"bg-muted/50 border-t font-medium [&>tr]:last:border-b-0")
}

// TableRow renders a table's <tr>.
func TableRow(attribsChildren ...any) *mx.Element {
	return finish(html.TR(attribsChildren...), "table-row",
		"hover:bg-muted/50 data-[state=selected]:bg-muted border-b transition-colors")
}

// TableHead renders a header cell <th>.
func TableHead(attribsChildren ...any) *mx.Element {
	return finish(html.TH(attribsChildren...), "table-head",
		"text-foreground h-10 px-2 text-left align-middle font-medium whitespace-nowrap [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]")
}

// TableCell renders a data cell <td>.
func TableCell(attribsChildren ...any) *mx.Element {
	return finish(html.TD(attribsChildren...), "table-cell",
		"p-2 align-middle whitespace-nowrap [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]")
}

// TableCaption renders a table's <caption>.
func TableCaption(attribsChildren ...any) *mx.Element {
	return finish(html.Caption(attribsChildren...), "table-caption", "text-muted-foreground mt-4 text-sm")
}
