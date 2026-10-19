package shadcn

import (
	"strings"
	"testing"
	"time"

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
		`aria-selected="true"`, // the selected 15th
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
	out := render(t, Calendar(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), time.Time{}))
	if strings.Contains(out, `aria-selected="true"`) {
		t.Errorf("zero selected time should mark no day: %s", out)
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
	if strings.Contains(out, `aria-label="2025-12-28"`) {
		t.Errorf("Monday-first grid should not include Sun Dec 28: %s", out)
	}
	if !strings.Contains(out, `aria-label="2026-02-08"`) {
		t.Errorf("Monday-first grid should include Sun Feb 8: %s", out)
	}
	// The default Sunday-first grid is the mirror image.
	def := render(t, Calendar(month, time.Time{}))
	if !strings.Contains(def, `aria-label="2025-12-28"`) || strings.Contains(def, `aria-label="2026-02-08"`) {
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
	if strings.Index(out, ">Mo<") > strings.Index(out, ">So<") {
		t.Errorf("German week should start Monday and end Sunday: %s", out)
	}
}

// TestCalendarDisabled verifies the booking-calendar predicate: a disabled day
// is a disabled button with aria-disabled and is never aria-selected even when
// it equals selected — otherwise an unbookable date could still look chosen.
func TestCalendarDisabled(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	selected := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	out := render(t, CalendarWith(CalendarOptions{
		Disabled: func(d time.Time) bool {
			return d.Year() == 2026 && d.Month() == time.January && d.Day() == 15
		},
	}, month, selected))
	if c := strings.Count(out, `aria-disabled="true"`); c != 1 {
		t.Errorf("want exactly one disabled day, got %d: %s", c, out)
	}
	if !strings.Contains(out, `disabled="disabled"`) {
		t.Errorf("disabled day should carry the disabled attribute: %s", out)
	}
	if strings.Contains(out, `aria-selected="true"`) {
		t.Errorf("disabled day matching selected must not be aria-selected: %s", out)
	}
	// Control: with nothing disabled the same selected day IS aria-selected
	// and no day is disabled.
	ok := render(t, CalendarWith(CalendarOptions{}, month, selected))
	if !strings.Contains(ok, `aria-selected="true"`) {
		t.Errorf("non-disabled selected day should be aria-selected: %s", ok)
	}
	if strings.Contains(ok, `aria-disabled="true"`) {
		t.Errorf("nil Disabled must disable nothing: %s", ok)
	}
}

// TestCalendarDayAccessibleName guards that every day button gets an accessible
// name (a screen reader would otherwise announce a bare "15"), defaulting to
// the ISO 8601 date and overridable via DayLabel.
func TestCalendarDayAccessibleName(t *testing.T) {
	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	out := render(t, Calendar(month, time.Time{}))
	if !strings.Contains(out, `aria-label="2026-01-15"`) {
		t.Errorf("day button should carry an ISO 8601 accessible name: %s", out)
	}
	out = render(t, CalendarWith(CalendarOptions{
		DayLabel: func(d time.Time) string { return d.Format("Jan 2, 2006") },
	}, month, time.Time{}))
	if !strings.Contains(out, `aria-label="Jan 15, 2026"`) {
		t.Errorf("custom DayLabel should be used verbatim: %s", out)
	}
}

// dayButton returns the start tag of the day button for the ISO date in a
// rendered calendar.
func dayButton(t *testing.T, out, date string) string {
	t.Helper()
	i := strings.Index(out, `aria-label="`+date+`"`)
	if i < 0 {
		t.Fatalf("no day button for %s in %s", date, out)
	}
	start := strings.LastIndex(out[:i], "<button")
	end := strings.Index(out[i:], ">")
	return out[start : i+end+1]
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
		`<input type="hidden" name="due" value="2026-06-14" data-slot="calendar-input"/>`,
		`<table role="grid" aria-label="June 2026" class="w-full border-collapse" onkeydown="calendarKey(event)">`,
		// The grid cell carries aria-selected; the button only its look.
		`<td role="gridcell" class="p-0 text-center" aria-selected="true"><button data-slot="calendar-day" type="button" aria-label="2026-06-14" data-selected="true"`,
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
	if b := dayButton(t, out, "2026-06-14"); !strings.Contains(b, `data-selected="true" data-date="2026-06-14" tabindex="0"`) {
		t.Errorf("the selected day must be the Tab stop: %s", b)
	}
	if b := dayButton(t, out, "2026-05-31"); !strings.Contains(b, `data-date="2026-05-31" tabindex="-1"`) {
		t.Errorf("a day of the previous month is selectable too: %s", b)
	}

	empty := render(t, CalendarWith(CalendarOptions{Name: "due"}, month, time.Time{}))
	if !strings.Contains(empty, `<input type="hidden" name="due" value="" data-slot="calendar-input"/>`) {
		t.Errorf("no selection submits an empty value: %s", empty)
	}

	// A disabled selected day is not marked selected, but the hidden input
	// keeps its date, so an unchanged submit does not clear a stored value.
	disabled := render(t, CalendarWith(CalendarOptions{Name: "due", Disabled: func(d time.Time) bool { return d.Equal(selected) }}, month, selected))
	if !strings.Contains(disabled, `<input type="hidden" name="due" value="2026-06-14" data-slot="calendar-input"/>`) {
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
	want := `<span data-slot="calendar-text" data-empty="true" data-calendar-text="due" class="data-[empty=true]:text-muted-foreground">Pick a date</span>`
	if out != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}

// Every calendar is a grid named by its month, with one gridcell per day,
// and a button never carries aria-selected, which ARIA does not allow on a
// button. The month buttons are named.
func TestCalendarGridSemantics(t *testing.T) {
	month := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	out := render(t, Calendar(month, time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC)))
	if !strings.Contains(out, `<table role="grid" aria-label="June 2026"`) {
		t.Errorf("the table must be a grid named by its month: %s", out)
	}
	if c := strings.Count(out, `<td role="gridcell"`); c != 42 {
		t.Errorf("want 42 gridcells, got %d", c)
	}
	selected := dayButton(t, out, "2026-06-14")
	if strings.Contains(selected, `aria-selected="true"`) {
		t.Errorf("the selected day's button must not carry aria-selected: %s", selected)
	}
	// The button keeps the selected look through data-selected.
	if !strings.Contains(selected, `data-selected="true"`) || !strings.Contains(selected, "data-[selected=true]:bg-primary") {
		t.Errorf("the selected day's button lacks its selected look: %s", selected)
	}
	for _, label := range []string{`aria-label="Go to the previous month"`, `aria-label="Go to the next month"`} {
		if !strings.Contains(out, label) {
			t.Errorf("missing %s", label)
		}
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
