package shadcn

import (
	"strconv"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Calendar is a Go port of shadcn/ui's Calendar. shadcn wraps react-day-picker;
// this port generates one month's grid server-side with Go's time package — no
// client runtime. It renders the month containing `month`, with the day equal
// to `selected` marked (pass the zero time.Time for no selection).
//
// Month navigation is a server round-trip: PrevButton/NextButton are plain
// buttons with no default behavior, so wire them with html.HRef("?month=…") or
// an hx.Get to re-render the adjacent month (see the gallery example, and
// [DatePicker] which embeds a Calendar in a [Popover]).
//
// Single-month, single-selection is the ported core; react-day-picker's range /
// multiple / disabled-matcher features are not reproduced.
func Calendar(month, selected time.Time, attribsChildren ...any) *mx.Element {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Grid starts on the Sunday on or before the first of the month, 6 weeks.
	gridStart := first.AddDate(0, 0, -int(first.Weekday()))
	hasSel := !selected.IsZero()
	selY, selM, selD := selected.Date()

	head := html.Element("tr")
	for _, wd := range []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"} {
		head.Children = append(head.Children, html.Element("th",
			html.Class("text-muted-foreground w-8 pb-1 text-[0.8rem] font-normal"), wd))
	}

	body := html.Element("tbody")
	for week := range 6 {
		row := html.Element("tr")
		for d := range 7 {
			day := gridStart.AddDate(0, 0, week*7+d)
			cls := "inline-flex size-8 items-center justify-center rounded-md text-sm font-normal hover:bg-accent hover:text-accent-foreground aria-selected:bg-primary aria-selected:text-primary-foreground aria-selected:hover:bg-primary"
			if day.Month() != first.Month() {
				cls += " text-muted-foreground opacity-50"
			}
			btn := html.ButtonButton(html.Class(cls), strconv.Itoa(day.Day()))
			if hasSel && day.Year() == selY && day.Month() == selM && day.Day() == selD {
				btn.Attribs = append(btn.Attribs, html.Attrib("aria-selected", "true"))
			}
			row.Children = append(row.Children, html.Element("td",
				html.Class("p-0 text-center"),
				finish(btn, "calendar-day", "")))
		}
		body.Children = append(body.Children, row)
	}

	nav := html.DivClass("flex items-center justify-between pb-2",
		finish(html.ButtonButton(
			html.Class(ButtonClasses(ButtonOutline, SizeIcon)+" size-7"),
			iconChevronLeft()), "calendar-prev", ""),
		html.DivClass("text-sm font-medium", month.Format("January 2006")),
		finish(html.ButtonButton(
			html.Class(ButtonClasses(ButtonOutline, SizeIcon)+" size-7"),
			iconChevronRight()), "calendar-next", ""),
	)
	table := html.Element("table", html.Class("w-full border-collapse"),
		html.Element("thead", head),
		body,
	)
	// Structural children first, then any caller attribs/children; finish
	// merges a caller class into the base.
	root := html.Div(append([]any{nav, table}, attribsChildren...)...)
	return finish(root, "calendar", "bg-background w-fit rounded-md border p-3")
}
