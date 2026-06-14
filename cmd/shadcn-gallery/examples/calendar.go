package examples

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn"
)

func CalendarDemo() mx.Component {
	return shadcn.Calendar(
		time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC),
	)
}
