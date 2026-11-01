package shadcn

import (
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

func TestCalendar(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	selected := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	out := render(t, Calendar(month, selected))
	for _, want := range []string{
		`data-slot="calendar"`,
		"January 2026",
		">Su<",
		">Sa<",
		`data-slot="calendar-day"`,
		`data-selected="true"`, // the selected 15th
		`data-slot="calendar-prev"`,
		`data-slot="calendar-next"`,
		"lucide-chevron-left",
		"lucide-chevron-right",
		">15<",
		">31<", // January has 31 days
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestCalendarNoSelection(t *testing.T) {
	for _, out := range []string{
		render(t, Calendar(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), time.Time{})),
		render(t, CalendarWith(CalendarOptions{Name: "due"}, time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), time.Time{})),
	} {
		for _, mark := range []string{`data-selected="true"`, `aria-selected="true"`, `aria-pressed="true"`} {
			if strings.Contains(out, mark) {
				t.Errorf("zero selected time should mark no day with %s: %s", mark, out)
			}
		}
	}
}

// TestCalendarWithZeroMatchesCalendar is the backward-compatibility guard: a
// zero-value CalendarOptions must render byte-for-byte what Calendar renders,
// so no existing caller sees a difference.
func TestCalendarWithZeroMatchesCalendar(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	selected := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	got := render(t, CalendarWith(CalendarOptions{}, month, selected))
	want := render(t, Calendar(month, selected))
	if got != want {
		t.Errorf("zero-value CalendarWith must equal Calendar\nCalendarWith:\n%s\nCalendar:\n%s", got, want)
	}
}

// TestCalendarWeekStartMonday checks that WeekStart moves Monday into the first
// column and shifts which trailing/leading days of the adjacent months appear.
// January 2026 starts on a Thursday, so the Sunday-first grid begins on Sun
// Dec 28 2025 and the Monday-first grid begins one day later on Mon Dec 29,
// running through Sun Feb 8 instead of Sat Feb 7. The full ISO date in each
// day's aria-label makes the shift unambiguous.
func TestCalendarWeekStartMonday(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	out := render(t, CalendarWith(CalendarOptions{WeekStart: time.Monday}, month, time.Time{}))
	if i, j := strings.Index(out, ">Mo<"), strings.Index(out, ">Su<"); i < 0 || j < 0 || i > j {
		t.Errorf("Monday should head the grid (before Sunday): %s", out)
	}
	if strings.Contains(out, `>2025-12-28<`) {
		t.Errorf("Monday-first grid should not include Sun Dec 28: %s", out)
	}
	if !strings.Contains(out, `>2026-02-08<`) {
		t.Errorf("Monday-first grid should include Sun Feb 8: %s", out)
	}
	// The default Sunday-first grid is the mirror image.
	def := render(t, Calendar(month, time.Time{}))
	if !strings.Contains(def, `>2025-12-28<`) || strings.Contains(def, `>2026-02-08<`) {
		t.Errorf("default grid should include Dec 28 and exclude Feb 8: %s", def)
	}
}

// TestCalendarGermanLocale exercises the localization hooks the way a German
// site would: Monday-first, German weekday abbreviations and a German month
// label — none of which Go's locale-less time package can produce on its own.
func TestCalendarGermanLocale(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	deDays := [7]string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"} // indexed by time.Weekday
	deMonths := [12]string{"Januar", "Februar", "März", "April", "Mai", "Juni",
		"Juli", "August", "September", "Oktober", "November", "Dezember"}
	out := render(t, CalendarWith(CalendarOptions{
		WeekStart:   time.Monday,
		WeekdayName: func(wd time.Weekday) string { return deDays[wd] },
		MonthLabel:  func(m time.Time) string { return deMonths[m.Month()-1] + m.Format(" 2006") },
	}, month, time.Time{}))
	if !strings.Contains(out, "Januar 2026") {
		t.Errorf("missing German month label: %s", out)
	}
	for _, wd := range []string{">Mo<", ">Di<", ">Mi<", ">Do<", ">Fr<", ">Sa<", ">So<"} {
		if !strings.Contains(out, wd) {
			t.Errorf("missing German weekday heading %q: %s", wd, out)
		}
	}
	if strings.Contains(out, ">Su<") || strings.Contains(out, ">We<") {
		t.Errorf("English weekday headings leaked into German calendar: %s", out)
	}
	// Without a WeekdayLabel, a localized heading gets no English name.
	if strings.Contains(out, `aria-label="Monday"`) {
		t.Errorf("an English weekday name leaked into the German calendar: %s", out)
	}
	if strings.Index(out, ">Mo<") > strings.Index(out, ">So<") {
		t.Errorf("German week should start Monday and end Sunday: %s", out)
	}
}

// Each weekday heading carries the full name as its accessible name: English
// by default, or WeekdayLabel's.
func TestCalendarWeekdayLabel(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	out := render(t, Calendar(month, time.Time{}))
	if !strings.Contains(out,
		/*html*/ `aria-label="Sunday">Su</th>`) {
		t.Errorf("the default heading lacks the full weekday name: %s", out)
	}
	de := [7]string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
	out = render(t, CalendarWith(CalendarOptions{
		WeekdayName:  func(wd time.Weekday) string { return de[wd][:2] },
		WeekdayLabel: func(wd time.Weekday) string { return de[wd] },
	}, month, time.Time{}))
	if !strings.Contains(out,
		/*html*/ `aria-label="Montag">Mo</th>`) || strings.Contains(out, `aria-label="Monday"`) {
		t.Errorf("WeekdayLabel must name the headings: %s", out)
	}
}

// TestCalendarDisabled verifies the booking-calendar predicate: a disabled day
// is faded, in select mode a disabled button with aria-disabled, and is never
// marked selected even when it equals selected — otherwise an unbookable date
// could still look chosen.
func TestCalendarDisabled(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	selected := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	opts := CalendarOptions{
		Disabled: func(d time.Time) bool {
			return d.Year() == 2026 && d.Month() == time.January && d.Day() == 15
		},
	}
	out := render(t, CalendarWith(opts, month, selected))
	if b := dayButton(t, out, "2026-01-15"); !strings.Contains(b, "opacity-50") {
		t.Errorf("a disabled day is faded: %s", b)
	}
	if strings.Contains(out, `data-selected="true"`) {
		t.Errorf("disabled day matching selected must not be marked selected: %s", out)
	}
	opts.Name = "due"
	picker := render(t, CalendarWith(opts, month, selected))
	if c := strings.Count(picker, `aria-disabled="true"`); c != 1 {
		t.Errorf("want exactly one disabled day, got %d: %s", c, picker)
	}
	if b := dayButton(t, picker, "2026-01-15"); !strings.Contains(b, `disabled="disabled"`) {
		t.Errorf("disabled day should carry the disabled attribute: %s", b)
	}
	// Control: with nothing disabled the same selected day IS marked selected
	// and no day is disabled.
	ok := render(t, CalendarWith(CalendarOptions{}, month, selected))
	if !strings.Contains(ok, `data-selected="true"`) {
		t.Errorf("non-disabled selected day should be marked selected: %s", ok)
	}
	if strings.Contains(ok, `aria-disabled="true"`) {
		t.Errorf("nil Disabled must disable nothing: %s", ok)
	}
}

// TestCalendarDayAccessibleName guards that every day gets an accessible
// name (a screen reader would otherwise announce a bare "15"), defaulting to
// the ISO 8601 date and overridable via DayLabel: a static day's
// screen-reader text, a picker day's aria-label.
func TestCalendarDayAccessibleName(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	out := render(t, Calendar(month, time.Time{}))
	if !strings.Contains(out,
		/*html*/ `<span aria-hidden="true">15</span><span class="sr-only">2026-01-15</span>`) {
		t.Errorf("a static day should carry an ISO 8601 accessible name: %s", out)
	}
	if !strings.Contains(out, `<table class="w-full border-collapse" aria-label="January 2026">`) {
		t.Errorf("the static table is named by its month: %s", out)
	}
	labelled := CalendarOptions{DayLabel: func(d time.Time) string { return d.Format("Jan 2, 2006") }}
	if out := render(t, CalendarWith(labelled, month, time.Time{})); !strings.Contains(out, `class="sr-only">Jan 15, 2026<`) {
		t.Errorf("custom DayLabel should be used verbatim: %s", out)
	}
	labelled.Name = "due"
	if out := render(t, CalendarWith(labelled, month, time.Time{})); !strings.Contains(out, `aria-label="Jan 15, 2026"`) {
		t.Errorf("a picker day button is named by DayLabel: %s", out)
	}
}

// A static calendar's days are text, not buttons that do nothing: no Tab
// stops, and the selected day says so in its name instead of aria-pressed.
func TestCalendarStaticDaysAreText(t *testing.T) {
	june := func(day int) time.Time { return time.Date(2026, time.June, day, 0, 0, 0, 0, time.UTC) }
	out := render(t, CalendarWith(CalendarOptions{Today: june(10)}, june(1), june(14)))
	if strings.Contains(out, `data-slot="calendar-day" type="button"`) || strings.Contains(out, "aria-pressed") {
		t.Errorf("static days must not be buttons: %s", out)
	}
	if b := dayButton(t, out, "2026-06-14"); !strings.Contains(b, `data-selected="true"`) || !strings.Contains(b, ">2026-06-14, selected<") {
		t.Errorf("the selected day: %s", b)
	}
	if b := dayButton(t, out, "2026-06-10"); !strings.Contains(b, `aria-current="date"`) || !strings.Contains(b, ">Today, 2026-06-10<") {
		t.Errorf("today: %s", b)
	}
	if strings.Contains(out, "hover:bg-accent hover:text-accent-foreground focus-visible") {
		t.Errorf("text has no hover or focus look: %s", out)
	}
}

// dayButton returns the start tag of the day button for the ISO date in a
// rendered calendar, or for a static calendar the markup of the day's text.
func dayButton(t *testing.T, out, date string) string {
	t.Helper()
	// The name may be wrapped in the Today and selected labels.
	name := `(Today, )?` + regexp.QuoteMeta(date) + `(, selected)?`
	if loc := regexp.MustCompile(`aria-label="` + name + `"`).FindStringIndex(out); loc != nil {
		start := strings.LastIndex(out[:loc[0]], "<button")
		end := strings.Index(out[loc[0]:], ">")
		return out[start : loc[0]+end+1]
	}
	if loc := regexp.MustCompile(`class="sr-only">` + name + `</span></span>`).FindStringIndex(out); loc != nil {
		start := strings.LastIndex(out[:loc[0]], `<span data-slot="calendar-day"`)
		return out[start:loc[1]]
	}
	t.Fatalf("no day for %s in %s", date, out)
	return ""
}

// TestCalendarSelectMode covers the markup of a calendar with Name: a
// hidden input carrying the selected day, the shared script once, the
// grid's key handler, and on every day button its date, the click handler
// and a tabindex, with the selected day as the only Tab stop.
func TestCalendarSelectMode(t *testing.T) {
	month := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	selected := time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC)
	out := render(t, CalendarWith(CalendarOptions{Name: "due"}, month, selected))
	for _, want := range []string{
		/*html*/ `<input type="hidden" name="due" value="2026-06-14" data-slot="calendar-input" data-label="2026-06-14"/>`,

		/*html*/ `<table class="w-full border-collapse" role="grid" aria-label="June 2026" onkeydown="calendarKey(event)">`,
		// The grid cell carries aria-selected; the button only its look.

		/*html*/ `<td class="p-0 text-center" role="gridcell" aria-selected="true"><button data-slot="calendar-day" type="button" aria-label="2026-06-14, selected" data-selected="true"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if c := strings.Count(out, "<script>if(!window.calendarSelect)"); c != 1 {
		t.Errorf("want the calendar script once, got %d", c)
	}
	if c := strings.Count(out, `onclick="calendarSelect(this)"`); c != 42 {
		t.Errorf("want a click handler on all 42 days, got %d", c)
	}
	if c := strings.Count(out, `tabindex="0"`); c != 1 {
		t.Errorf("want exactly one Tab stop, got %d", c)
	}
	if b := dayButton(t, out, "2026-06-14"); !strings.Contains(b, `data-selected="true" data-date="2026-06-14" data-label="2026-06-14" tabindex="0"`) {
		t.Errorf("the selected day must be the Tab stop: %s", b)
	}
	if b := dayButton(t, out, "2026-05-31"); !strings.Contains(b, `data-date="2026-05-31" data-label="2026-05-31" tabindex="-1"`) {
		t.Errorf("a day of the previous month is selectable too: %s", b)
	}

	empty := render(t, CalendarWith(CalendarOptions{Name: "due"}, month, time.Time{}))
	if !strings.Contains(empty,
		/*html*/ `<input type="hidden" name="due" value="" data-slot="calendar-input"/>`) {
		t.Errorf("no selection submits an empty value: %s", empty)
	}

	// A disabled selected day is not marked selected, but the hidden input
	// keeps its date, so an unchanged submit does not clear a stored value.
	disabled := render(t, CalendarWith(CalendarOptions{Name: "due", Disabled: func(d time.Time) bool { return d.Equal(selected) }}, month, selected))
	if !strings.Contains(disabled,
		/*html*/ `<input type="hidden" name="due" value="2026-06-14" data-slot="calendar-input" data-label="2026-06-14"/>`) {
		t.Errorf("the hidden input must keep a disabled selected date: %s", disabled)
	}
	if strings.Contains(disabled, `aria-selected="true"`) {
		t.Errorf("a disabled selected day must not be marked selected: %s", disabled)
	}
}

// Without a selected day the Tab stop is today, else the first enabled day
// of the displayed month.
func TestCalendarSelectModeTabStop(t *testing.T) {
	month := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	june := func(day int) time.Time { return time.Date(2026, time.June, day, 0, 0, 0, 0, time.UTC) }
	tests := []struct {
		name     string
		opts     CalendarOptions
		selected time.Time
		want     string
	}{
		{"selected", CalendarOptions{Today: june(10)}, june(14), "2026-06-14"},
		{"today", CalendarOptions{Today: june(10)}, time.Time{}, "2026-06-10"},
		{"first of month", CalendarOptions{}, time.Time{}, "2026-06-01"},
		{"today in another month", CalendarOptions{Today: time.Date(2026, time.July, 2, 0, 0, 0, 0, time.UTC)}, time.Time{}, "2026-06-01"},
		{"first enabled", CalendarOptions{Disabled: func(d time.Time) bool { return d.Before(june(3)) }}, time.Time{}, "2026-06-03"},
		{"disabled selection", CalendarOptions{Today: june(10), Disabled: func(d time.Time) bool { return d.Equal(june(14)) }}, june(14), "2026-06-10"},
		// After navigating to another month the selected day is not in the
		// grid, and the grid still needs a Tab stop.
		{"selection in another month", CalendarOptions{Today: june(10)}, time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC), "2026-06-10"},
		// A selected day shown in a spillover cell is marked selected, so it
		// is the Tab stop too.
		{"selection in a spillover cell", CalendarOptions{Today: june(10)}, time.Date(2026, time.July, 2, 0, 0, 0, 0, time.UTC), "2026-07-02"},
		{"disabled today", CalendarOptions{Today: june(10), Disabled: func(d time.Time) bool { return d.Equal(june(10)) }}, time.Time{}, "2026-06-01"},
		// With every day of the month disabled the grid still needs a Tab
		// stop: the first enabled day it shows, May 31 in this grid.
		{"whole month disabled", CalendarOptions{Disabled: func(d time.Time) bool { return d.Month() == time.June }}, time.Time{}, "2026-05-31"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Name = "due"
			out := render(t, CalendarWith(tt.opts, month, tt.selected))
			if c := strings.Count(out, `tabindex="0"`); c != 1 {
				t.Fatalf("want exactly one Tab stop, got %d", c)
			}
			if b := dayButton(t, out, tt.want); !strings.Contains(b, `tabindex="0"`) {
				t.Errorf("want the Tab stop on %s: %s", tt.want, b)
			}
		})
	}
}

// A calendar without Name stays the static grid: no script, no hidden
// input, no handlers.
func TestCalendarStaticHasNoScript(t *testing.T) {
	out := render(t, Calendar(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), time.Time{}))
	for _, unwanted := range []string{"<script", "calendar-input", "data-date", "tabindex", "onclick", "onkeydown"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("static calendar contains %q: %s", unwanted, out)
		}
	}
}

// Today carries aria-current and the accent background, which replaces
// the muted color of a day outside the month.
func TestCalendarToday(t *testing.T) {
	month := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	out := render(t, CalendarWith(CalendarOptions{Today: time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)}, month, time.Time{}))
	if c := strings.Count(out, `aria-current="date"`); c != 1 {
		t.Fatalf("want one day marked as today, got %d", c)
	}
	if b := dayButton(t, out, "2026-06-10"); !strings.Contains(b, `aria-current="date"`) || !strings.Contains(b, "bg-accent text-accent-foreground") {
		t.Errorf("today lacks aria-current or the accent classes: %s", b)
	}

	outside := render(t, CalendarWith(CalendarOptions{Today: time.Date(2026, time.July, 2, 0, 0, 0, 0, time.UTC)}, month, time.Time{}))
	b := dayButton(t, outside, "2026-07-02")
	if strings.Contains(b, "text-muted-foreground") || !strings.Contains(b, "text-accent-foreground") {
		t.Errorf("today outside the month must use the accent color, not the muted one: %s", b)
	}
	if strings.Contains(render(t, Calendar(month, time.Time{})), "aria-current") {
		t.Errorf("a zero Today must mark no day")
	}
}

// StartMonth and EndMonth disable the navigation buttons at the limits.
func TestCalendarMonthLimits(t *testing.T) {
	june := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	opts := CalendarOptions{
		StartMonth: time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC),
		EndMonth:   time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
	}
	navDisabled := func(out, slot string) bool {
		i := strings.Index(out, `data-slot="`+slot+`"`)
		return i >= 0 && strings.Contains(out[i:i+strings.Index(out[i:], ">")], `disabled="disabled"`)
	}
	tests := []struct {
		month            time.Time
		wantPrev, wantNx bool
	}{
		{june, true, false},
		{june.AddDate(0, 1, 0), false, false},
		{june.AddDate(0, 2, 0), false, true},
		{june.AddDate(0, -1, 0), true, false},
		// The limits compare whole months, years included.
		{time.Date(2027, time.June, 1, 0, 0, 0, 0, time.UTC), false, true},
		{time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC), true, false},
	}
	for _, tt := range tests {
		out := render(t, CalendarWith(opts, tt.month, time.Time{}))
		if got := navDisabled(out, "calendar-prev"); got != tt.wantPrev {
			t.Errorf("%s: previous disabled %t, want %t", tt.month.Format("2006-01"), got, tt.wantPrev)
		}
		if got := navDisabled(out, "calendar-next"); got != tt.wantNx {
			t.Errorf("%s: next disabled %t, want %t", tt.month.Format("2006-01"), got, tt.wantNx)
		}
	}
	if out := render(t, Calendar(june, time.Time{})); navDisabled(out, "calendar-prev") || navDisabled(out, "calendar-next") {
		t.Errorf("zero limits must leave navigation open")
	}
}

func TestCalendarText(t *testing.T) {
	out := render(t, CalendarText("due", html.DataAttr("empty", "true"), "Pick a date"))
	want := /*html*/ `<span data-slot="calendar-text" data-empty="true" data-calendar-text="due" class="data-[empty=true]:text-muted-foreground">Pick a date</span>`
	if out != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}

// A calendar in select mode is a grid named by its month, with one
// gridcell per day, and a button never carries aria-selected, which ARIA
// does not allow on a button. The static calendar has no keyboard handling,
// so it is a plain table that marks the selected day only by its look. The
// month buttons are named in both.
// The select-mode grid is named by MonthLabel, so a localized date picker
// announces its month in the page's language.
func TestCalendarSelectModeGridLocalized(t *testing.T) {
	out := render(t, CalendarWith(CalendarOptions{
		Name:       "due",
		MonthLabel: func(time.Time) string { return "Juni 2026" },
	}, time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), time.Time{}))
	if !strings.Contains(out, `role="grid" aria-label="Juni 2026"`) || strings.Contains(out, "June 2026") {
		t.Errorf("the grid must be named by MonthLabel: %s", out)
	}
}

func TestCalendarGridSemantics(t *testing.T) {
	month := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	selected := time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC)
	out := render(t, CalendarWith(CalendarOptions{Name: "due"}, month, selected))
	if !strings.Contains(out, `role="grid" aria-label="June 2026"`) {
		t.Errorf("the select-mode table must be a grid named by its month: %s", out)
	}
	if c := strings.Count(out, `role="gridcell"`); c != 42 {
		t.Errorf("want 42 gridcells, got %d", c)
	}
	button := dayButton(t, out, "2026-06-14")
	if strings.Contains(button, `aria-selected="true"`) {
		t.Errorf("the selected day's button must not carry aria-selected: %s", button)
	}
	// The button keeps the selected look through data-selected.
	if !strings.Contains(button, `data-selected="true"`) || !strings.Contains(button, "data-[selected=true]:bg-primary") {
		t.Errorf("the selected day's button lacks its selected look: %s", button)
	}

	if strings.Contains(out, "aria-pressed") {
		t.Errorf("select mode marks the selection on the gridcell, not with aria-pressed: %s", out)
	}

	static := render(t, Calendar(month, selected))
	for _, unwanted := range []string{`role="grid"`, `role="gridcell"`, `aria-selected`} {
		if strings.Contains(static, unwanted) {
			t.Errorf("the static calendar must not carry %s: %s", unwanted, static)
		}
	}
	// Without the grid, the selected day's name says so.
	if b := dayButton(t, static, "2026-06-14"); !strings.Contains(b, ">2026-06-14, selected<") {
		t.Errorf("the static calendar's selected day must say so: %s", b)
	}
	for _, label := range []string{`aria-label="Go to the previous month"`, `aria-label="Go to the next month"`} {
		if !strings.Contains(static, label) || !strings.Contains(out, label) {
			t.Errorf("missing %s", label)
		}
	}
}

// Days of the adjacent months are muted like upstream's outside days, but
// only disabled days are faded, so a pickable outside day does not look
// disabled.
func TestCalendarOutsideDays(t *testing.T) {
	month := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	out := render(t, CalendarWith(CalendarOptions{
		Disabled: func(d time.Time) bool { return d.Equal(time.Date(2026, time.June, 9, 0, 0, 0, 0, time.UTC)) },
	}, month, time.Time{}))
	if b := dayButton(t, out, "2026-07-01"); !strings.Contains(b, "text-muted-foreground") || strings.Contains(b, "opacity-50") {
		t.Errorf("an outside day is muted, not faded: %s", b)
	}
	if b := dayButton(t, out, "2026-06-09"); !strings.Contains(b, "opacity-50") || !strings.Contains(b, "text-muted-foreground") {
		t.Errorf("a disabled day of the month is muted and faded: %s", b)
	}
}

// Disabled is asked once per day of the grid: a booking calendar's predicate
// may look up availability, and select mode must not double that.
func TestCalendarAsksDisabledOncePerDay(t *testing.T) {
	calls := map[time.Time]int{}
	render(t, CalendarWith(CalendarOptions{
		Name:     "due",
		Disabled: func(d time.Time) bool { calls[d]++; return false },
	}, time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), time.Time{}))
	if len(calls) != 42 {
		t.Errorf("Disabled asked for %d days, want 42", len(calls))
	}
	for day, n := range calls {
		if n != 1 {
			t.Errorf("Disabled asked %d times for %s, want once", n, day.Format(time.DateOnly))
		}
	}
}

// The selection marks follow the selected day wherever it is: a spillover
// cell is marked like any other, a day outside the grid marks nothing, and
// the hidden input keeps the stored day either way. With every day
// disabled no day is a Tab stop.
func TestCalendarSelectModeEdges(t *testing.T) {
	month := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	spill := render(t, CalendarWith(CalendarOptions{Name: "due"}, month, time.Date(2026, time.July, 2, 0, 0, 0, 0, time.UTC)))
	if !strings.Contains(spill,
		/*html*/ `aria-selected="true"><button data-slot="calendar-day" type="button" aria-label="2026-07-02, selected" data-selected="true"`) ||
		!strings.Contains(spill, `name="due" value="2026-07-02"`) {
		t.Errorf("a spillover selection must be marked and submitted: %s", spill)
	}
	outside := render(t, CalendarWith(CalendarOptions{Name: "due"}, month, time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)))
	if strings.Contains(outside, `aria-selected="true"`) || !strings.Contains(outside, `name="due" value="2026-08-20"`) {
		t.Errorf("a selection outside the grid marks nothing and stays in the hidden input: %s", outside)
	}
	blocked := render(t, CalendarWith(CalendarOptions{Name: "due", Disabled: func(time.Time) bool { return true }}, month, time.Time{}))
	if strings.Contains(blocked, `tabindex="0"`) {
		t.Errorf("with every day disabled no day may be a Tab stop: %s", blocked)
	}
}

// MonthURL wires the month buttons to load the adjacent months, and emits
// the script even for a static calendar; a button at a month limit stays
// disabled and gets no URL.
func TestCalendarMonthURL(t *testing.T) {
	june := time.Date(2026, time.June, 17, 0, 0, 0, 0, time.UTC)
	url := func(m time.Time) string { return "/cal/" + m.Format("2006-01") }
	out := render(t, CalendarWith(CalendarOptions{MonthURL: url}, june, time.Time{}))
	for _, want := range []string{
		`data-month-url="/cal/2026-05" data-year-url="/cal/2025-06" onclick="calendarNav(this)"`,
		`data-month-url="/cal/2026-07" data-year-url="/cal/2027-06" onclick="calendarNav(this)"`,
		"<script>if(!window.calendarSelect)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, `data-mode="`) || strings.Contains(out, `data-slot="calendar-input"`) {
		t.Errorf("navigation alone does not make a calendar a form control: %s", out)
	}

	limited := render(t, CalendarWith(CalendarOptions{MonthURL: url, StartMonth: june}, june, time.Time{}))
	if strings.Contains(limited, "/cal/2026-05") {
		t.Errorf("a disabled previous-month button must get no URL: %s", limited)
	}
	// A year on stops at EndMonth.
	capped := render(t, CalendarWith(CalendarOptions{MonthURL: url, EndMonth: june.AddDate(0, 3, 0)}, june, time.Time{}))
	if !strings.Contains(capped, `data-year-url="/cal/2026-09"`) {
		t.Errorf("a year on must stop at EndMonth: %s", capped)
	}
	// MonthURL is asked for the first of the month at midnight UTC.
	var asked []time.Time
	render(t, CalendarWith(CalendarOptions{MonthURL: func(m time.Time) string { asked = append(asked, m); return "x" }}, june, time.Time{}))
	for _, m := range asked {
		if m.Day() != 1 || m.Hour() != 0 || m.Location() != time.UTC {
			t.Errorf("MonthURL got %v, want the first of a month at midnight UTC", m)
		}
	}
}

// The dropdown caption shows native month and year selects bounded by
// StartMonth and EndMonth, each option carrying the URL of its month.
func TestCalendarCaptionDropdown(t *testing.T) {
	june := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	opts := CalendarOptions{
		CaptionLayout: CalendarCaptionDropdown,
		StartMonth:    time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
		EndMonth:      time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
		MonthURL:      func(m time.Time) string { return "/cal/" + m.Format("2006-01") },
	}
	out := render(t, CalendarWith(opts, june, time.Time{}))
	for _, want := range []string{
		`data-slot="calendar-month-select"`,
		`data-slot="calendar-year-select"`,
		`onchange="calendarNav(this)"`,

		/*html*/ `<option value="2026-06" data-month-url="/cal/2026-06" selected="selected">June</option>`,

		/*html*/ `<option value="2026-08" data-month-url="/cal/2026-08">August</option>`,
		// Past EndMonth a month cannot be chosen.

		/*html*/ `<option value="2026-09" disabled="disabled">September</option>`,
		// A year shows the displayed month, moved into the range.

		/*html*/ `<option value="2025" data-month-url="/cal/2025-06">2025</option>`,

		/*html*/ `<option value="2026" data-month-url="/cal/2026-06" selected="selected">2026</option>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out,
		/*html*/ `<option value="2024"`) || strings.Contains(out,
		/*html*/ `<option value="2027"`) {
		t.Errorf("the years must stay within StartMonth and EndMonth: %s", out)
	}
	// A year's month is moved into the range: March 2025 for a displayed
	// January.
	jan := render(t, CalendarWith(opts, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), time.Time{}))
	if !strings.Contains(jan,
		/*html*/ `<option value="2025" data-month-url="/cal/2025-03">2025</option>`) {
		t.Errorf("a year must load its month within the range: %s", jan)
	}

	// Without MonthURL the dropdowns cannot load a month, so are disabled.
	static := render(t, CalendarWith(CalendarOptions{CaptionLayout: CalendarCaptionDropdownMonths}, june, time.Time{}))
	if !strings.Contains(static, `disabled="disabled"`) || strings.Contains(static, "calendar-year-select") || !strings.Contains(static, "<span>2026</span>") {
		t.Errorf("dropdown-months without MonthURL: %s", static)
	}
	// Without bounds the years reach 100 years back from Today's year.
	years := render(t, CalendarWith(CalendarOptions{CaptionLayout: CalendarCaptionDropdownYears, Today: june}, june, time.Time{}))
	if !strings.Contains(years,
		/*html*/ `<option value="1926"`) || strings.Contains(years,
		/*html*/ `<option value="1925"`) || strings.Contains(years,
		/*html*/ `<option value="2027"`) {
		t.Errorf("default year range: %s", years)
	}
}

// A range marks its first, last and middle days for the look and the
// gridcells as selected, and submits its ends in two hidden inputs.
func TestCalendarRange(t *testing.T) {
	june := func(day int) time.Time { return time.Date(2026, time.June, day, 0, 0, 0, 0, time.UTC) }
	// RangeEnd before selected is swapped.
	out := render(t, CalendarWith(CalendarOptions{Name: "stay", Mode: CalendarRange, RangeEnd: june(10)}, june(1), june(12)))
	for _, want := range []string{
		/*html*/ `<input type="hidden" name="stay" value="2026-06-10" data-slot="calendar-input" data-label="2026-06-10"/>`,

		/*html*/ `<input type="hidden" name="stay-end" value="2026-06-12" data-slot="calendar-input" data-label="2026-06-12"/>`,
		`data-mode="range"`,
		`aria-multiselectable="true"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	for day, want := range map[string][]string{
		"2026-06-10": {`data-range-start="true"`, `data-selected="true"`},
		"2026-06-11": {`data-range-middle="true"`},
		"2026-06-12": {`data-range-end="true"`, `data-selected="true"`},
	} {
		b := dayButton(t, out, day)
		for _, w := range want {
			if !strings.Contains(b, w) {
				t.Errorf("%s lacks %s: %s", day, w, b)
			}
		}
	}
	if c := strings.Count(out, `aria-selected="true"`); c != 3 {
		t.Errorf("want the 3 days of the range selected, got %d", c)
	}
	if b := dayButton(t, out, "2026-06-11"); strings.Contains(b, "data-selected") {
		t.Errorf("a middle day has the accent look, not the selected one: %s", b)
	}
	if b := dayButton(t, out, "2026-06-13"); strings.Contains(b, "data-range") || strings.Contains(b, "data-selected") {
		t.Errorf("a day after the range is not marked: %s", b)
	}

	// A range without an end marks its first day as both ends.
	if empty := render(t, CalendarWith(CalendarOptions{Name: "stay", Mode: CalendarRange}, june(1), time.Time{})); strings.Count(empty, `value=""`) != 2 {
		t.Errorf("an empty range submits two empty values: %s", empty)
	}
	open := render(t, CalendarWith(CalendarOptions{Name: "stay", Mode: CalendarRange, EndName: "until"}, june(1), june(12)))
	if b := dayButton(t, open, "2026-06-12"); !strings.Contains(b, `data-range-start="true"`) || !strings.Contains(b, `data-range-end="true"`) {
		t.Errorf("an open range marks its first day as both ends: %s", b)
	}
	if !strings.Contains(open,
		/*html*/ `<input type="hidden" name="until" value="" data-slot="calendar-input"/>`) {
		t.Errorf("EndName names the end input, empty while the range is open: %s", open)
	}
}

// Several days each get a hidden input and a selected gridcell.
func TestCalendarMultiple(t *testing.T) {
	june := func(day int) time.Time { return time.Date(2026, time.June, day, 0, 0, 0, 0, time.UTC) }
	out := render(t, CalendarWith(CalendarOptions{Name: "days", Mode: CalendarMultiple, Selected: []time.Time{june(20), june(3), june(3)}}, june(1), june(9)))
	if !strings.Contains(out,
		/*html*/ `<span data-slot="calendar-inputs" data-name="days" hidden="hidden"><input type="hidden" name="days" value="2026-06-03" data-slot="calendar-input" data-label="2026-06-03"/><input type="hidden" name="days" value="2026-06-09"`) {
		t.Errorf("want one sorted hidden input per day: %s", out)
	}
	if c := strings.Count(out, `name="days" value=`); c != 3 {
		t.Errorf("want 3 hidden inputs, got %d", c)
	}
	if c := strings.Count(out, `aria-selected="true"`); c != 3 {
		t.Errorf("want 3 selected days, got %d", c)
	}
	if b := dayButton(t, out, "2026-06-03"); !strings.Contains(b, `tabindex="0"`) {
		t.Errorf("the first selected day is the Tab stop: %s", b)
	}
	if strings.Contains(out, "data-range") {
		t.Errorf("several days are no range: %s", out)
	}
	empty := render(t, CalendarWith(CalendarOptions{Name: "days", Mode: CalendarMultiple}, june(1), time.Time{}))
	if !strings.Contains(empty,
		/*html*/ `<span data-slot="calendar-inputs" data-name="days" hidden="hidden"></span>`) {
		t.Errorf("no days, no inputs: %s", empty)
	}
}

// TestCalendarScriptInBrowser runs the calendar script in headless Chrome:
// picking a day, the events and the CalendarText, the keys skipping a
// disabled day, PageDown and an arrow key past the grid loading another month
// through MonthURL with the focus and the picked day carried over, the
// popover closing, right-to-left keys, and the range and multiple modes.
func TestCalendarScriptInBrowser(t *testing.T) {
	june := func(day int) time.Time { return time.Date(2026, time.June, day, 0, 0, 0, 0, time.UTC) }
	singleOpts := CalendarOptions{
		Name:     "due",
		Today:    june(10),
		Disabled: func(d time.Time) bool { return d.Equal(june(16)) },
		DayLabel: func(d time.Time) string { return d.Format("January 2, 2006") },
		MonthURL: func(m time.Time) string { return "/month/" + m.Format("2006-01") },
	}
	single := func(month time.Time) *mx.Element {
		return CalendarWith(singleOpts, month, time.Time{}, html.ID("single"))
	}
	body := mx.Components{
		html.Form(html.ID("form"),
			Popover(
				PopoverTrigger("pick", CalendarText("due", html.DataAttr("empty", "true"), "Pick a date")),
				PopoverContent("pick", "", single(june(1))),
			),
		),
		CalendarWith(CalendarOptions{Name: "rtl"}, june(1), time.Time{}, html.ID("rtl"), html.Dir("rtl")),
		html.Form(
			CalendarWith(CalendarOptions{Name: "stay", Mode: CalendarRange}, june(1), time.Time{}, html.ID("range")),
			CalendarText("stay", html.DataAttr("empty", "true"), "Pick a range"),
		),
		html.Form(
			CalendarWith(CalendarOptions{Name: "days", Mode: CalendarMultiple}, june(1), time.Time{}, html.ID("multi")),
			CalendarText("days", html.DataAttr("empty", "true"), "Pick days"),
		),
	}
	routes := map[string]http.HandlerFunc{
		"GET /month/{month}": func(w http.ResponseWriter, r *http.Request) {
			month, err := time.Parse("2006-01", r.PathValue("month"))
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := single(month).Render(r.Context(), mx.NewCheckedWriter(w)); err != nil {
				t.Errorf("render month: %v", err)
			}
		},
	}
	const script = /*js*/ `
var $=function(q,r){return (r||document).querySelector(q);};
var day=function(c,d){return $('#'+c+' [data-date="'+d+'"]');};
var key=function(k){document.activeElement.dispatchEvent(new KeyboardEvent('keydown',{key:k,bubbles:true,cancelable:true}));return document.activeElement.getAttribute('data-date');};
var events={input:0,change:0};
$('#form').addEventListener('input',function(){events.input++;});
$('#form').addEventListener('change',function(){events.change++;});

// Picking a day.
$('#pick').showPopover();
out.openBefore=$('#pick').matches(':popover-open');
day('single','2026-06-14').click();
out.value=$('#single [data-slot=calendar-input]').value;
out.cellSelected=day('single','2026-06-14').parentElement.getAttribute('aria-selected');
out.tabStop=$('#single [data-date][tabindex="0"]').getAttribute('data-date');
out.name=day('single','2026-06-14').getAttribute('aria-label');
out.todayName=day('single','2026-06-10').getAttribute('aria-label');
out.text=$('[data-calendar-text=due]').textContent;
out.textEmpty=$('[data-calendar-text=due]').hasAttribute('data-empty');
out.inputEvents=events.input;out.changeEvents=events.change;
out.openAfter=$('#pick').matches(':popover-open');
day('single','2026-06-15').click();
out.reselectedOld=day('single','2026-06-14').hasAttribute('data-selected');
out.reselectedOldName=day('single','2026-06-14').getAttribute('aria-label');
day('single','2026-06-14').click();

// Keys: a disabled day is skipped, End and Home stay in the row.
$('#pick').showPopover();
day('single','2026-06-14').focus();
out.right=key('ArrowRight');
out.skip=key('ArrowRight');
out.end=key('End');
out.home=key('Home');
key('End');

// PageDown loads July with the focus on the same day.
key('PageDown');
await wait(500);
out.julyLabel=$('#single table').getAttribute('aria-label');
out.julyFocus=document.activeElement.getAttribute('data-date');
out.julyValue=$('#single [data-slot=calendar-input]').value;
out.julyLabelKept=$('#single [data-slot=calendar-input]').getAttribute('data-label');

// The previous-month button loads June, marks the picked day again and
// keeps the focus on the button.
$('#single [data-slot=calendar-prev]').click();
await wait(500);
out.backLabel=$('#single table').getAttribute('aria-label');
out.backSelected=day('single','2026-06-14').getAttribute('data-selected');
out.backSelectedName=day('single','2026-06-14').getAttribute('aria-label');
out.backTabStop=$('#single [data-date][tabindex="0"]').getAttribute('data-date');
out.backFocus=document.activeElement.getAttribute('data-slot');

// An arrow key past the top of the grid loads May.
day('single','2026-05-31').focus();
key('ArrowUp');
await wait(500);
out.mayLabel=$('#single table').getAttribute('aria-label');
out.mayFocus=document.activeElement.getAttribute('data-date');

// Right to left, ArrowRight moves to the previous day.
day('rtl','2026-06-10').focus();
out.rtlRight=key('ArrowRight');

// A range: start, an earlier day swaps the ends, a third click starts anew.
day('range','2026-06-10').click();
out.rangeFirst=[].map.call(document.querySelectorAll('#range [data-slot=calendar-input]'),function(i){return i.value;}).join('|');
day('range','2026-06-08').click();
out.range=[].map.call(document.querySelectorAll('#range [data-slot=calendar-input]'),function(i){return i.value;}).join('|');
out.rangeStart=day('range','2026-06-08').getAttribute('data-range-start');
out.rangeMiddle=day('range','2026-06-09').getAttribute('data-range-middle');
out.rangeMiddleLook=day('range','2026-06-09').hasAttribute('data-selected');
out.rangeEnd=day('range','2026-06-10').getAttribute('data-range-end');
out.rangeSelected=document.querySelectorAll('#range [aria-selected=true]').length;
out.rangeText=$('[data-calendar-text=stay]').textContent;
day('range','2026-06-20').click();
out.rangeRestart=[].map.call(document.querySelectorAll('#range [data-slot=calendar-input]'),function(i){return i.value;}).join('|');

// Several days: each click toggles one.
day('multi','2026-06-05').click();
day('multi','2026-06-03').click();
out.multi=[].map.call(document.querySelectorAll('#multi [data-slot=calendar-input]'),function(i){return i.name+'='+i.value;}).join('|');
out.multiText=$('[data-calendar-text=days]').textContent;
day('multi','2026-06-05').click();
day('multi','2026-06-03').click();
out.multiEmpty=document.querySelectorAll('#multi [data-slot=calendar-input]').length;
out.multiPlaceholder=$('[data-calendar-text=days]').textContent+'/'+$('[data-calendar-text=days]').getAttribute('data-empty');
`
	got := runInChrome(t, body, script, routes)
	checkRecorded(t, got, map[string]any{
		"openBefore":        true,
		"value":             "2026-06-14",
		"cellSelected":      "true",
		"tabStop":           "2026-06-14",
		"name":              "June 14, 2026, selected",
		"todayName":         "Today, June 10, 2026",
		"text":              "June 14, 2026",
		"textEmpty":         false,
		"inputEvents":       float64(1),
		"changeEvents":      float64(1),
		"openAfter":         false,
		"reselectedOld":     false,
		"reselectedOldName": "June 14, 2026",
		"right":             "2026-06-15",
		"skip":              "2026-06-17",
		"end":               "2026-06-20",
		"home":              "2026-06-14",
		"julyLabel":         "July 2026",
		"julyFocus":         "2026-07-20",
		"julyValue":         "2026-06-14",
		"julyLabelKept":     "June 14, 2026",
		"backLabel":         "June 2026",
		"backSelected":      "true",
		"backSelectedName":  "June 14, 2026, selected",
		"backTabStop":       "2026-06-14",
		"backFocus":         "calendar-prev",
		"mayLabel":          "May 2026",
		"mayFocus":          "2026-05-24",
		"rtlRight":          "2026-06-09",
		"rangeFirst":        "2026-06-10|",
		"range":             "2026-06-08|2026-06-10",
		"rangeStart":        "true",
		"rangeMiddle":       "true",
		"rangeMiddleLook":   false,
		"rangeEnd":          "true",
		"rangeSelected":     float64(3),
		"rangeText":         "2026-06-08 – 2026-06-10",
		"rangeRestart":      "2026-06-20|",
		"multi":             "days=2026-06-03|days=2026-06-05",
		"multiText":         "2026-06-03, 2026-06-05",
		"multiEmpty":        float64(0),
		"multiPlaceholder":  "Pick days/true",
	})
}

func BenchmarkCalendarWith(b *testing.B) {
	june := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	opts := CalendarOptions{Name: "due", Today: june.AddDate(0, 0, 9), Disabled: func(d time.Time) bool { return d.Day() == 16 }}
	for b.Loop() {
		CalendarWith(opts, june, june.AddDate(0, 0, 13))
	}
}

// TestCalendarScriptInBrowserMore runs the newer calendar behaviors in
// headless Chrome: Shift+PageDown moving a year, a range refusing to span a
// disabled day with ExcludeDisabled, the arrow keys crossing from one month
// to the next shown beside it, and calendarPick picking from outside.
func TestCalendarScriptInBrowserMore(t *testing.T) {
	june := func(day int) time.Time { return time.Date(2026, time.June, day, 0, 0, 0, 0, time.UTC) }
	yearOpts := CalendarOptions{Name: "year", MonthURL: func(m time.Time) string { return "/year/" + m.Format("2006-01") }}
	yearCal := func(month time.Time) *mx.Element { return CalendarWith(yearOpts, month, time.Time{}, html.ID("year")) }
	body := mx.Components{
		yearCal(june(1)),
		CalendarWith(CalendarOptions{Name: "stay", Mode: CalendarRange, ExcludeDisabled: true,
			Disabled: func(d time.Time) bool { return d.Equal(june(12)) }}, june(1), time.Time{}, html.ID("range")),
		CalendarWith(CalendarOptions{Name: "two", NumberOfMonths: 2}, june(1), time.Time{}, html.ID("two")),
		html.Form(
			CalendarWith(CalendarOptions{Name: "pick"}, june(1), time.Time{}, html.ID("pick")),
			html.Input(html.Type("text"), html.ID("typed"), html.DataAttr("calendar-text", "pick")),
		),
	}
	routes := map[string]http.HandlerFunc{
		"GET /year/{month}": func(w http.ResponseWriter, r *http.Request) {
			month, _ := time.Parse("2006-01", r.PathValue("month"))
			_ = yearCal(month).Render(r.Context(), mx.NewCheckedWriter(w))
		},
	}
	const script = /*js*/ `
var day=function(c,d){return document.querySelector('#'+c+' [data-date="'+d+'"]');};
var key=function(k,shift){document.activeElement.dispatchEvent(new KeyboardEvent('keydown',{key:k,shiftKey:!!shift,bubbles:true,cancelable:true}));return document.activeElement.getAttribute('data-date');};
var vals=function(c){return [].map.call(document.querySelectorAll('#'+c+' [data-slot=calendar-input]'),function(i){return i.value;}).join('|');};

day('year','2026-06-14').focus();
key('PageDown',true);
await wait(500);
out.yearLabel=document.querySelector('#year table').getAttribute('aria-label');
out.yearFocus=document.activeElement.getAttribute('data-date');

day('range','2026-06-10').click();
day('range','2026-06-15').click();
out.spanDisabled=vals('range');
day('range','2026-06-17').click();
out.afterDisabled=vals('range');

out.twoTables=document.querySelectorAll('#two table').length;
out.twoDuplicates=document.querySelectorAll('#two [data-date="2026-07-01"]').length;
day('two','2026-06-30').focus();
out.crossRight=key('ArrowRight');
out.crossUp=key('ArrowUp');

out.picked=calendarPick('pick','2026-06-20','the 20th');
out.pickValue=vals('pick');
out.pickSelected=day('pick','2026-06-20').getAttribute('data-selected');
day('pick','2026-06-21').click();
out.typed=document.getElementById('typed').value;
out.unknown=calendarPick('nope','2026-06-20');
`
	got := runInChrome(t, body, script, routes)
	checkRecorded(t, got, map[string]any{
		"yearLabel":     "June 2027",
		"yearFocus":     "2027-06-14",
		"spanDisabled":  "2026-06-15|",
		"afterDisabled": "2026-06-15|2026-06-17",
		"twoTables":     float64(2),
		"twoDuplicates": float64(1),
		"crossRight":    "2026-07-01",
		"crossUp":       "2026-06-24",
		"picked":        true,
		"pickValue":     "2026-06-20",
		"pickSelected":  "true",
		"typed":         "2026-06-21",
		"unknown":       false,
	})
}
