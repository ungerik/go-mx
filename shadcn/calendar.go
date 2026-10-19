package shadcn

import (
	"cmp"
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
	// A [CalendarText] shows the picked day's DayLabel.
	DayLabel func(time.Time) string

	// Today marks that day with the accent background and
	// aria-current="date", as react-day-picker marks today. The zero time
	// marks no day. It is an option rather than time.Now() so the caller
	// decides the time zone of "today" and the output stays deterministic.
	Today time.Time

	// StartMonth and EndMonth bound month navigation: the previous-month
	// button is disabled while the displayed month is StartMonth's month or
	// earlier, the next-month button while it is EndMonth's month or later.
	// A zero time leaves that side open. Like react-day-picker's startMonth
	// and endMonth they only limit navigation; Disabled makes days
	// unselectable.
	StartMonth, EndMonth time.Time

	// Name makes the calendar a form control, the date picker's job in
	// react-day-picker's single mode. A hidden input with this name carries
	// the selected day as 2006-01-02 ("" for none), and one shared inline
	// script, emitted once per calendar, lets the user pick a day:
	//
	//   - A click (or Enter or Space on a focused day) selects the day,
	//     writes it to the hidden input and fires bubbling input and change
	//     events there, so htmx triggers and form listeners react.
	//   - Arrow keys move the focus by a day or a week, Home and End to the
	//     start or end of the week row, skipping disabled days and following
	//     the text direction. Only one day is a Tab stop: the selected day,
	//     else Today, else the first enabled day of the month.
	//   - Every [CalendarText] for this name in the same form shows the
	//     picked day's DayLabel.
	//   - A calendar inside an open popover closes it, returning the focus
	//     to the trigger.
	//
	// Month navigation stays a server round-trip, as without Name: the
	// arrow keys stop at the edge of the six-week grid. A request that
	// re-renders the calendar for another month should include the hidden
	// input (hx-include) to keep the picked day. "" renders the static grid.
	//
	// Disabled, StartMonth and EndMonth only shape the grid in the browser;
	// the hidden input submits whatever it holds, so the handler must check
	// a submitted day against the same rules. A selected day that is
	// disabled is not marked selected, but the hidden input keeps its date,
	// so an unchanged submit does not clear a stored value.
	Name string
}

// calendarScript is the once-emitted client code of a calendar with
// [CalendarOptions].Name: calendarSelect picks a day button, calendarKey
// moves the focus between day buttons.
const calendarScript = /*js*/ `if(!window.calendarSelect){window.calendarSelect=function(b){var r=b.closest('[data-slot=calendar]');if(!r||b.disabled)return;r.querySelectorAll('[data-date]').forEach(function(d){d.removeAttribute('data-selected');d.parentElement.removeAttribute('aria-selected');d.tabIndex=-1;});b.setAttribute('data-selected','true');b.parentElement.setAttribute('aria-selected','true');b.tabIndex=0;var i=r.querySelector('[data-slot=calendar-input]');if(!i)return;i.value=b.getAttribute('data-date');i.dispatchEvent(new Event('input',{bubbles:true}));i.dispatchEvent(new Event('change',{bubbles:true}));(i.form||document).querySelectorAll('[data-calendar-text]').forEach(function(t){if(t.getAttribute('data-calendar-text')===i.name){t.textContent=b.getAttribute('aria-label');t.removeAttribute('data-empty');}});var p=r.closest('[popover]');if(p&&p.matches(':popover-open'))p.hidePopover();};window.calendarKey=function(e){var b=e.target.closest('[data-date]');if(!b)return;var r=b.closest('[data-slot=calendar]');var days=Array.prototype.slice.call(r.querySelectorAll('[data-date]'));var i=days.indexOf(b),col=i%7,j,step;var rtl=getComputedStyle(r).direction==='rtl';switch(e.key){case'ArrowLeft':step=rtl?1:-1;j=i+step;break;case'ArrowRight':step=rtl?-1:1;j=i+step;break;case'ArrowUp':step=-7;j=i-7;break;case'ArrowDown':step=7;j=i+7;break;case'Home':step=1;j=i-col;break;case'End':step=-1;j=i-col+6;break;default:return;}e.preventDefault();while(j>=0&&j<days.length&&days[j].disabled&&j!==i){j+=step;}if(j<0||j>=days.length||days[j].disabled)return;days.forEach(function(d){d.tabIndex=-1;});days[j].tabIndex=0;days[j].focus();};}`

// CalendarText renders the text of a control that shows the day picked in
// a [Calendar] whose [CalendarOptions].Name is name, such as the label of a
// date picker's [PopoverTrigger]. Pass the label of the current day as
// children, or a placeholder together with html.DataAttr("empty", "true"),
// which renders muted like upstream's empty date picker. When the user
// picks a day, the calendar script replaces the text with that day's
// DayLabel and removes data-empty. The script only looks in the form that
// holds the calendar, or in the whole document when there is none.
func CalendarText(name string, attribsChildren ...any) *mx.Element {
	e := html.Span(attribsChildren...)
	if e.AttribIndex("data-calendar-text") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("calendar-text", name))
	}
	return finish(e, "calendar-text", "data-[empty=true]:text-muted-foreground")
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
	sameDay := func(a, b time.Time) bool {
		ay, am, ad := a.Date()
		by, bm, bd := b.Date()
		return ay == by && am == bm && ad == bd
	}
	isSelected := func(day time.Time) bool {
		return hasSel && day.Year() == selY && day.Month() == selM && day.Day() == selD
	}
	// The grid's days, with Disabled asked once per day: a booking calendar's
	// predicate may look up availability.
	var days [42]time.Time
	var disabledDays [42]bool
	for i := range days {
		days[i] = gridStart.AddDate(0, 0, i)
		disabledDays[i] = opts.Disabled != nil && opts.Disabled(days[i])
	}
	selectable := opts.Name != ""
	// In select mode one day is the grid's Tab stop: the selected day, else
	// today, else the first enabled day of the month, else the first enabled
	// day of the grid, so a month with every day disabled stays reachable.
	var tabStop time.Time
	if selectable {
		var selectedDay, today, firstInMonth, firstInGrid time.Time
		for i, day := range days {
			if disabledDays[i] {
				continue
			}
			inMonth := day.Month() == first.Month()
			switch {
			case isSelected(day):
				selectedDay = day
			case inMonth && !opts.Today.IsZero() && sameDay(day, opts.Today):
				today = day
			case inMonth && firstInMonth.IsZero():
				firstInMonth = day
			}
			if firstInGrid.IsZero() {
				firstInGrid = day
			}
		}
		tabStop = cmp.Or(selectedDay, today, firstInMonth, firstInGrid)
	}

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
			day := days[week*7+d]
			disabled := disabledDays[week*7+d]
			cls := "inline-flex size-8 items-center justify-center rounded-md text-sm font-normal outline-none hover:bg-accent hover:text-accent-foreground focus-visible:relative focus-visible:z-10 focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none data-[selected=true]:bg-primary data-[selected=true]:text-primary-foreground data-[selected=true]:hover:bg-primary"
			if day.Month() != first.Month() || disabled {
				cls += " text-muted-foreground opacity-50"
			}
			isToday := !opts.Today.IsZero() && sameDay(day, opts.Today)
			if isToday {
				// finish merges the classes, so these replace a muted color.
				cls += " bg-accent text-accent-foreground"
			}
			btn := html.ButtonButton(html.Class(cls),
				html.Attrib("aria-label", dayLabel(day)),
				strconv.Itoa(day.Day()))
			// The grid cell carries aria-selected, which a button may not; the
			// button's data-selected carries the selected look.
			cell := html.Element("td", html.Role("gridcell"), html.Class("p-0 text-center"))
			switch {
			case disabled:
				// A disabled day is never selectable, so it is never marked
				// aria-selected even when it equals selected.
				btn.Attribs = append(btn.Attribs, html.Disabled, html.Attrib("aria-disabled", "true"))
			case isSelected(day):
				btn.Attribs = append(btn.Attribs, html.DataAttr("selected", "true"))
				cell.Attribs = append(cell.Attribs, html.Attrib("aria-selected", "true"))
			}
			if isToday {
				btn.Attribs = append(btn.Attribs, html.Attrib("aria-current", "date"))
			}
			if selectable {
				tabIndex := -1
				if !tabStop.IsZero() && sameDay(day, tabStop) {
					tabIndex = 0
				}
				btn.Attribs = append(btn.Attribs,
					html.DataAttr("date", day.Format(time.DateOnly)),
					html.TabIndex(tabIndex),
					html.OnClick("calendarSelect(this)"),
				)
			}
			cell.Children = append(cell.Children, finish(btn, "calendar-day", ""))
			row.Children = append(row.Children, cell)
		}
		body.Children = append(body.Children, row)
	}

	monthIndex := func(t time.Time) int { return t.Year()*12 + int(t.Month()) }
	prev := html.ButtonButton(
		html.Class(ButtonClasses(ButtonOutline, SizeIcon)+" size-7"),
		labelAriaLabel(func(l Labels) string { return l.CalendarPrevious }),
		iconChevronLeft())
	if !opts.StartMonth.IsZero() && monthIndex(first) <= monthIndex(opts.StartMonth) {
		prev.Attribs = append(prev.Attribs, html.Disabled)
	}
	next := html.ButtonButton(
		html.Class(ButtonClasses(ButtonOutline, SizeIcon)+" size-7"),
		labelAriaLabel(func(l Labels) string { return l.CalendarNext }),
		iconChevronRight())
	if !opts.EndMonth.IsZero() && monthIndex(first) >= monthIndex(opts.EndMonth) {
		next.Attribs = append(next.Attribs, html.Disabled)
	}
	nav := html.DivClass("flex items-center justify-between pb-2",
		finish(prev, "calendar-prev", ""),
		html.DivClass("text-sm font-medium", monthLabel(month)),
		finish(next, "calendar-next", ""),
	)
	// A grid named by its month, like react-day-picker's, so a screen reader
	// announces the picked day and passes the arrow keys to select mode.
	table := html.Element("table", html.Role("grid"), html.Attrib("aria-label", monthLabel(month)),
		html.Class("w-full border-collapse"),
		html.Element("thead", head),
		body,
	)
	structure := []any{nav, table}
	if selectable {
		table.Attribs = append(table.Attribs, html.OnKeyDown("calendarKey(event)"))
		value := ""
		if hasSel {
			value = selected.Format(time.DateOnly)
		}
		structure = append(structure,
			html.Input(html.Type("hidden"), html.Name(opts.Name), html.Value(value),
				mx.NewAttrib("data-slot", "calendar-input")),
			html.ScriptJS(calendarScript),
		)
	}
	// Structural children first, then any caller attribs/children; finish
	// merges a caller class into the base.
	root := html.Div(append(structure, attribsChildren...)...)
	return finish(root, "calendar", "bg-background w-fit rounded-md border p-3")
}
