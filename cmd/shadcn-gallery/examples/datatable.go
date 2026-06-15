package examples

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// dtFilterScript wires the toolbar input to hide table rows whose email cell
// doesn't match — the one live piece of the DataTable recipe. Sorting and
// pagination would be server-side (HTMX / query params) in a real app; here the
// sort buttons and Previous/Next are rendered but inert.
const dtFilterScript = `if(!window.dtFilter){window.dtFilter=function(i){var q=i.value.toLowerCase();i.closest('[data-datatable]').querySelectorAll('tbody>tr').forEach(function(r){r.hidden=(r.dataset.email||'').toLowerCase().indexOf(q)<0;});};}`

// dtRow builds one data row with a selection checkbox, the data cells and a
// row-actions dropdown (unique id per row).
func dtRow(i int, status, email, amount string, selected bool) mx.Component {
	id := "dt-actions-" + strconv.Itoa(i)
	cells := []any{html.DataAttr("email", email)}
	if selected {
		cells = append(cells, html.DataAttr("state", "selected"))
	}
	check := shadcn.Checkbox(html.Attrib("aria-label", "Select row"))
	if selected {
		check = shadcn.Checkbox(html.Attrib("aria-label", "Select row"), html.Checked)
	}
	cells = append(cells,
		shadcn.TableCell(check),
		shadcn.TableCell(html.Class("capitalize"), status),
		shadcn.TableCell(html.Class("lowercase"), email),
		shadcn.TableCell(html.Class("text-right font-medium"), amount),
		shadcn.TableCell(html.Class("text-right"),
			shadcn.DropdownMenu(
				shadcn.DropdownMenuTrigger(id,
					html.Class(shadcn.ButtonClasses(shadcn.ButtonGhost, shadcn.SizeIcon)+" size-8"),
					navIcon("M12 5h.01", "M12 12h.01", "M12 19h.01")),
				shadcn.DropdownMenuContent(id, "",
					shadcn.DropdownMenuItem("Copy email"),
					shadcn.DropdownMenuItem("View customer"),
					shadcn.DropdownMenuSeparator(),
					shadcn.DropdownMenuItem("Delete"),
				),
			),
		),
	)
	return shadcn.TableRow(cells...)
}

// DataTableDemo renders a data table with a live email filter, a column-toggle dropdown, selectable rows, row-action menus, and pagination controls.
func DataTableDemo() mx.Component {
	return html.Div(html.DataAttr("datatable", ""), html.Class("w-full"),
		html.Script(mx.Raw(dtFilterScript)),
		html.DivClass("flex items-center gap-2 py-4",
			shadcn.Input(html.Class("max-w-sm"),
				html.Placeholder("Filter emails..."), html.OnInput("dtFilter(this)")),
			shadcn.DropdownMenu(html.Class("ml-auto"),
				shadcn.DropdownMenuTrigger("dt-columns",
					html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)),
					"Columns", navIcon("m6 9 6 6 6-6")),
				shadcn.DropdownMenuContent("dt-columns", "",
					shadcn.DropdownMenuCheckboxItem(true, "Status"),
					shadcn.DropdownMenuCheckboxItem(true, "Email"),
					shadcn.DropdownMenuCheckboxItem(true, "Amount"),
				),
			),
		),
		html.DivClass("rounded-md border",
			shadcn.Table(
				shadcn.TableHeader(
					shadcn.TableRow(
						shadcn.TableHead(html.Class("w-12"),
							shadcn.Checkbox(html.Attrib("aria-label", "Select all"))),
						shadcn.TableHead("Status"),
						shadcn.TableHead(
							shadcn.Button(shadcn.ButtonGhost, shadcn.SizeSM, html.Class("-ml-3 h-8"),
								"Email", navIcon("m7 15 5 5 5-5", "m7 9 5-5 5 5"))),
						shadcn.TableHead(html.Class("text-right"), "Amount"),
						shadcn.TableHead(html.Class("w-12")),
					),
				),
				shadcn.TableBody(
					dtRow(0, "success", "ken99@example.com", "$316.00", true),
					dtRow(1, "success", "abe45@example.com", "$242.00", false),
					dtRow(2, "processing", "monserrat44@example.com", "$837.00", true),
					dtRow(3, "success", "silas22@example.com", "$874.00", false),
					dtRow(4, "failed", "carmella@example.com", "$721.00", false),
				),
			),
		),
		html.DivClass("flex items-center justify-end gap-2 py-4",
			html.DivClass("text-muted-foreground flex-1 text-sm", "2 of 5 row(s) selected."),
			shadcn.Button(shadcn.ButtonOutline, shadcn.SizeSM, "Previous"),
			shadcn.Button(shadcn.ButtonOutline, shadcn.SizeSM, "Next"),
		),
	)
}
