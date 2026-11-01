package examples

import (
	"context"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func init() {
	addMonthFragments("calendar-booking",
		time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC),
		GermanBookingCalendar)
	addMonthFragments("calendar-dropdown",
		time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2027, time.December, 1, 0, 0, 0, 0, time.UTC),
		DropdownCalendar)
	addMonthFragments("calendar-range",
		time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
		RangeCalendar)
}

// CalendarDemo renders a calendar with a selected date within a displayed month.
func CalendarDemo() mx.Component {
	return shadcn.Calendar(
		time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC),
	)
}

// CalendarBookingDemo renders the German booking calendar for June 2026.
func CalendarBookingDemo() mx.Component {
	return GermanBookingCalendar(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))
}

// GermanBookingCalendar renders CalendarWith as a German booking calendar:
// Monday-first, German weekday, month and day names, days before a cutoff
// disabled, today marked, navigation from June to December 2026, and a day
// picked with a click or the arrow keys into a hidden "booking-date" input.
// The cutoff ("today") is a fixed date rather than time.Now() so the
// pre-rendered gallery stays deterministic. MonthURL loads the other months,
// which the gallery pre-renders as static fragments.
func GermanBookingCalendar(month time.Time) mx.Component {
	deDays := [7]string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"} // indexed by time.Weekday
	deWeekdays := [7]string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
	deMonths := [12]string{"Januar", "Februar", "März", "April", "Mai", "Juni",
		"Juli", "August", "September", "Oktober", "November", "Dezember"}
	today := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)
	// The month buttons and day names take their words from the context's
	// Labels, and lang lets a screen reader read the German names with a
	// German voice.
	de := shadcn.Labels{
		CalendarPrevious: "Zum vorherigen Monat",
		CalendarNext:     "Zum nächsten Monat",
		CalendarToday:    "Heute, %s",
		CalendarSelected: "%s, ausgewählt",
	}
	calendar := shadcn.CalendarWith(
		shadcn.CalendarOptions{
			WeekStart:    time.Monday,
			WeekdayName:  func(wd time.Weekday) string { return deDays[wd] },
			WeekdayLabel: func(wd time.Weekday) string { return deWeekdays[wd] },
			MonthLabel:   func(m time.Time) string { return deMonths[m.Month()-1] + m.Format(" 2006") },
			DayLabel: func(d time.Time) string {
				return d.Format("2. ") + deMonths[d.Month()-1] + d.Format(" 2006")
			},
			Disabled:   func(d time.Time) bool { return d.Before(today) },
			Today:      today,
			StartMonth: today,
			EndMonth:   time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC),
			MonthURL:   monthURL("calendar-booking"),
			Name:       "booking-date",
		},
		month,
		time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC),
		html.Lang("de"),
	)
	return mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
		return calendar.Render(shadcn.ContextWithLabels(ctx, de), w)
	})
}

// CalendarDropdownDemo renders the month and year dropdown caption for June
// 2026.
func CalendarDropdownDemo() mx.Component {
	return DropdownCalendar(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))
}

// DropdownCalendar is upstream's Month and Year Selector: native month and
// year dropdowns in the caption, for 2026 and 2027, each loading the chosen
// month through MonthURL.
func DropdownCalendar(month time.Time) mx.Component {
	return shadcn.CalendarWith(
		shadcn.CalendarOptions{
			CaptionLayout: shadcn.CalendarCaptionDropdown,
			StartMonth:    time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			EndMonth:      time.Date(2027, time.December, 1, 0, 0, 0, 0, time.UTC),
			MonthURL:      monthURL("calendar-dropdown"),
			Name:          "dropdown-date",
		},
		month,
		time.Time{},
	)
}

// CalendarRangeDemo renders a range calendar for June 2026.
func CalendarRangeDemo() mx.Component {
	return RangeCalendar(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))
}

// RangeCalendar is upstream's Range Calendar: the first click starts a
// range, the second ends it, and the hidden inputs "stay" and "stay-end"
// carry its first and last day. It navigates from May to July 2026.
func RangeCalendar(month time.Time) mx.Component {
	return shadcn.CalendarWith(
		shadcn.CalendarOptions{
			Mode:       shadcn.CalendarRange,
			RangeEnd:   time.Date(2026, time.June, 13, 0, 0, 0, 0, time.UTC),
			StartMonth: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
			EndMonth:   time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
			MonthURL:   monthURL("calendar-range"),
			Name:       "stay",
		},
		month,
		time.Date(2026, time.June, 9, 0, 0, 0, 0, time.UTC),
	)
}
