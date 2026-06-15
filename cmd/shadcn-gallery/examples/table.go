package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// TableDemo renders a captioned invoices table with header and body rows.
func TableDemo() mx.Component {
	return shadcn.Table(
		shadcn.TableCaption("A list of your recent invoices."),
		shadcn.TableHeader(
			shadcn.TableRow(
				shadcn.TableHead(html.Class("w-[100px]"), "Invoice"),
				shadcn.TableHead("Status"),
				shadcn.TableHead("Method"),
				shadcn.TableHead(html.Class("text-right"), "Amount"),
			),
		),
		shadcn.TableBody(
			shadcn.TableRow(
				shadcn.TableCell(html.Class("font-medium"), "INV001"),
				shadcn.TableCell("Paid"),
				shadcn.TableCell("Credit Card"),
				shadcn.TableCell(html.Class("text-right"), "$250.00"),
			),
			shadcn.TableRow(
				shadcn.TableCell(html.Class("font-medium"), "INV002"),
				shadcn.TableCell("Pending"),
				shadcn.TableCell("PayPal"),
				shadcn.TableCell(html.Class("text-right"), "$150.00"),
			),
			shadcn.TableRow(
				shadcn.TableCell(html.Class("font-medium"), "INV003"),
				shadcn.TableCell("Unpaid"),
				shadcn.TableCell("Bank Transfer"),
				shadcn.TableCell(html.Class("text-right"), "$350.00"),
			),
		),
	)
}
