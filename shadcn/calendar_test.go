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
