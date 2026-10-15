package schemaorg_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestDate(t *testing.T) {
	require.Equal(t, "2026-10-15", schemaorg.Date(time.Date(2026, 10, 15, 19, 30, 0, 0, cest)))

	// The day is the one in the time's own location: 22:00 in New York is
	// already the next day in UTC.
	newYork := time.FixedZone("EDT", -4*60*60)
	evening := time.Date(2026, 10, 15, 22, 0, 0, 0, newYork)
	require.Equal(t, "2026-10-15", schemaorg.Date(evening))
	require.Equal(t, "2026-10-16", schemaorg.Date(evening.UTC()))

	require.Empty(t, schemaorg.Date(time.Time{}), "the zero time must be omitted")
}

func TestDateTime(t *testing.T) {
	require.Equal(t,
		"2026-10-15T19:30:00+02:00",
		schemaorg.DateTime(time.Date(2026, 10, 15, 19, 30, 0, 0, cest)),
	)
	require.Equal(t,
		"2026-10-15T17:30:00Z",
		schemaorg.DateTime(time.Date(2026, 10, 15, 17, 30, 0, 0, time.UTC)),
	)
	// Fractions of a second are dropped, not rounded.
	require.Equal(t,
		"2026-10-15T19:30:59+02:00",
		schemaorg.DateTime(time.Date(2026, 10, 15, 19, 30, 59, 999_999_999, cest)),
	)

	require.Empty(t, schemaorg.DateTime(time.Time{}), "the zero time must be omitted")
}
