package examples

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn"
)

// CalendarDemo renders a calendar with a selected date within a displayed month.
func CalendarDemo() mx.Component {
	return shadcn.Calendar(
		time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC),
	)
}

// CalendarBookingDemo renders CalendarWith as a German booking calendar:
// Monday-first, German weekday, month and day names, days before a cutoff
// disabled, today marked, no navigation into earlier months, and a day
// picked with a click or the arrow keys into a hidden "booking-date" input.
// The cutoff ("today") is a fixed date rather than time.Now() so the
// pre-rendered gallery stays deterministic.
func CalendarBookingDemo() mx.Component {
	deDays := [7]string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"} // indexed by time.Weekday
	deMonths := [12]string{"Januar", "Februar", "März", "April", "Mai", "Juni",
		"Juli", "August", "September", "Oktober", "November", "Dezember"}
	today := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)
	return shadcn.CalendarWith(
		shadcn.CalendarOptions{
			WeekStart:   time.Monday,
			WeekdayName: func(wd time.Weekday) string { return deDays[wd] },
			MonthLabel:  func(m time.Time) string { return deMonths[m.Month()-1] + m.Format(" 2006") },
			DayLabel: func(d time.Time) string {
				return d.Format("2. ") + deMonths[d.Month()-1] + d.Format(" 2006")
			},
			Disabled:   func(d time.Time) bool { return d.Before(today) },
			Today:      today,
			StartMonth: today,
			Name:       "booking-date",
		},
		time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC),
	)
}
