package schemaorg

// https://schema.org/Date
// https://schema.org/DateTime

import "time"

// Date formats the calendar day of d as an ISO 8601 date ("2026-10-15"), for a
// property that is about a day rather than an instant, such as the start of an
// all-day event. The day is the one in d's location, not in UTC, so a late
// evening in New York does not turn into the next day.
//
// The zero time returns "", so an unset time.Time leaves the property out.
func Date(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format(time.DateOnly)
}

// DateTime formats t as an ISO 8601 date and time with its UTC offset in the
// RFC 3339 profile ("2026-10-15T19:30:00+02:00"), dropping fractions of a
// second. The offset is what makes the value an instant: a time without it
// would be read in whatever zone the reader assumes.
//
// The zero time returns "", so an unset time.Time leaves the property out.
func DateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
