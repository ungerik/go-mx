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
// example rather than a new shadcn primitive. The Calendar's Name turns on its
// select mode: picking a day writes it to a hidden "demo-date" input, shows
// its DayLabel in the trigger's CalendarText and closes the popover.
func DatePickerDemo() mx.Component {
	return shadcn.Popover(
		shadcn.PopoverTrigger("demo-datepicker",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" w-[240px] justify-start text-left font-normal"),
			shadcn.CalendarText("demo-date", html.DataAttr("empty", "true"), "Pick a date"),
		),
		shadcn.PopoverContent("demo-datepicker", "",
			html.Class("w-auto p-0"),
			shadcn.CalendarWith(
				shadcn.CalendarOptions{
					Name:     "demo-date",
					DayLabel: func(d time.Time) string { return d.Format("January 2, 2006") },
				},
				time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
				time.Time{},
				html.Class("border-0"),
			),
		),
	)
}
