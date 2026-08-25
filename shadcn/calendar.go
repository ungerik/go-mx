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
// an hx.Get to re-render the adjacent month (see the gallery's Calendar
// example, and its DatePicker example which embeds a Calendar in a [Popover]).
//
// Single-month, single-selection is the ported core. react-day-picker's range
// and multiple-selection modes are not reproduced; its disabled-matcher is
// available as a per-day predicate on [CalendarWith] (see [CalendarOptions]),
// which also localizes the week start, weekday headings and month label.
func Calendar(month, selected time.Time, attribsChildren ...any) *mx.Element {
	return CalendarWith(CalendarOptions{}, month, selected, attribsChildren...)
}

// defaultWeekdayAbbr holds the English two-letter weekday abbreviations indexed
// by time.Weekday (Sunday..Saturday). It is the WeekdayName used when
// [CalendarOptions.WeekdayName] is nil, preserving the original grid headings.
var defaultWeekdayAbbr = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

// CalendarOptions configures [CalendarWith]. Its zero value reproduces
// [Calendar] exactly: a Sunday-first grid, English two-letter weekday headings,
// an English "January 2006" month label, no disabled days, and ISO 8601
// accessible day names.
//
// The hooks carry no locale data on purpose. Go's time package has no locale,
// and this library deliberately keeps golang.org/x/text (and any other locale
// database) out of the core module. A German — or any other localized —
// calendar is produced by supplying the name functions, from the caller or a
// future separate module: the component ships the hooks, not the translations.
type CalendarOptions struct {
	// WeekStart is the weekday shown in the first column of the grid. The
	// zero value is time.Sunday, preserving the current US-convention
	// behavior; set time.Monday for the ISO 8601 / continental-European
	// week. It shifts both the column order and which trailing/leading days
	// of the adjacent months fill the grid.
	WeekStart time.Weekday

	// WeekdayName returns the column heading for a weekday. nil uses the
	// current English two-letter abbreviations ("Su".."Sa").
	WeekdayName func(time.Weekday) string

	// MonthLabel returns the header text for the displayed month. nil uses
	// month.Format("January 2006"), which is always English because Go's
	// time package has no locale.
	MonthLabel func(time.Time) string

	// Disabled reports whether a day cannot be selected. nil disables
	// nothing. A disabled day renders a disabled button, aria-disabled="true"
	// and a muted class, and is never marked aria-selected even when it
	// equals selected. This is what makes the component usable as a booking
	// calendar: grey out dates with no availability, dates in the past, or
	// dates outside a window.
	Disabled func(time.Time) bool

	// DayLabel returns the accessible name (aria-label) for a day button. nil
	// uses the ISO 8601 date (2006-01-02), so a screen reader announces the
	// full date instead of a bare day number with no month or year context.
	DayLabel func(time.Time) string
}

// CalendarWith is the configurable variant of [Calendar]. A zero-value
// [CalendarOptions] renders exactly what [Calendar] renders, byte for byte; the
// options localize the grid (WeekStart, WeekdayName, MonthLabel) and turn it
// into a booking calendar (Disabled). The Xxx / XxxWith split mirrors
// [ReflectFormHandler] / ReflectFormHandlerWith in the root package.
func CalendarWith(opts CalendarOptions, month, selected time.Time, attribsChildren ...any) *mx.Element {
	weekdayName := opts.WeekdayName
	if weekdayName == nil {
		weekdayName = func(wd time.Weekday) string { return defaultWeekdayAbbr[wd] }
	}
	monthLabel := opts.MonthLabel
	if monthLabel == nil {
		monthLabel = func(m time.Time) string { return m.Format("January 2006") }
	}
	dayLabel := opts.DayLabel
	if dayLabel == nil {
		dayLabel = func(d time.Time) string { return d.Format("2006-01-02") }
	}

	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Grid starts on the WeekStart weekday on or before the first of the
	// month, 6 weeks. With the default WeekStart (Sunday) startOffset equals
	// int(first.Weekday()), so this is the original -int(first.Weekday()).
	startOffset := (int(first.Weekday()) - int(opts.WeekStart) + 7) % 7
	gridStart := first.AddDate(0, 0, -startOffset)
	hasSel := !selected.IsZero()
	selY, selM, selD := selected.Date()

	head := html.Element("tr")
	for col := range 7 {
		wd := time.Weekday((int(opts.WeekStart) + col) % 7)
		head.Children = append(head.Children, html.Element("th",
			html.Class("text-muted-foreground w-8 pb-1 text-[0.8rem] font-normal"), weekdayName(wd)))
	}

	body := html.Element("tbody")
	for week := range 6 {
		row := html.Element("tr")
		for d := range 7 {
			day := gridStart.AddDate(0, 0, week*7+d)
			disabled := opts.Disabled != nil && opts.Disabled(day)
			cls := "inline-flex size-8 items-center justify-center rounded-md text-sm font-normal hover:bg-accent hover:text-accent-foreground aria-selected:bg-primary aria-selected:text-primary-foreground aria-selected:hover:bg-primary"
			if day.Month() != first.Month() || disabled {
				cls += " text-muted-foreground opacity-50"
			}
			btn := html.ButtonButton(html.Class(cls),
				html.Attrib("aria-label", dayLabel(day)),
				strconv.Itoa(day.Day()))
			switch {
			case disabled:
				// A disabled day is never selectable, so it is never marked
				// aria-selected even when it equals selected.
				btn.Attribs = append(btn.Attribs, html.Disabled, html.Attrib("aria-disabled", "true"))
			case hasSel && day.Year() == selY && day.Month() == selM && day.Day() == selD:
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
		html.DivClass("text-sm font-medium", monthLabel(month)),
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
