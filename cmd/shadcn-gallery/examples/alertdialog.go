package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// AlertDialogDemo renders an outline trigger that opens a confirmation alert dialog with cancel and continue actions.
func AlertDialogDemo() mx.Component {
	return shadcn.AlertDialog(
		shadcn.AlertDialogTrigger("ad-demo",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)),
			"Show Dialog"),
		shadcn.AlertDialogContent("ad-demo",
			shadcn.AlertDialogHeader(
				shadcn.AlertDialogTitle("Remove this item?"),
				shadcn.AlertDialogDescription(
					"It will be moved to the archive and can be restored later."),
			),
			shadcn.AlertDialogFooter(
				shadcn.AlertDialogCancel("Cancel"),
				shadcn.AlertDialogAction("Continue"),
			),
		),
	)
}
