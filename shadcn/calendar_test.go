package shadcn

import (
	"strings"
	"testing"
	"time"
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
