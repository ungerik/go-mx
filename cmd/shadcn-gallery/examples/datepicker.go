package examples

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// DatePickerDemo is shadcn's DatePicker recipe: a Popover whose trigger shows
// the chosen date and whose content is a Calendar. shadcn ships DatePicker as a
// copy-paste composition, not an exported component, so this is a composition
// example rather than a new shadcn primitive.
func DatePickerDemo() mx.Component {
	return shadcn.Popover(
		shadcn.PopoverTrigger("demo-datepicker",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" w-[240px] justify-start text-left font-normal"),
			"June 14, 2026"),
		shadcn.PopoverContent("demo-datepicker", "",
			html.Class("w-auto p-0"),
			shadcn.Calendar(
				time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC),
				html.Class("border-0"),
			),
		),
	)
}
