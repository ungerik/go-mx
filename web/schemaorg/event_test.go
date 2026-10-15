package schemaorg_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestEvent(t *testing.T) {
	start := time.Date(2026, 10, 15, 19, 30, 0, 0, cest)
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Event","name":"Autumn Concert","startDate":"2026-10-15T19:30:00+02:00"}`,
		schemaorg.NewEvent("Autumn Concert", start),
	)
	// A zero start leaves the date out rather than rendering year one.
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Event","name":"Autumn Concert"}`,
		schemaorg.NewEvent("Autumn Concert", time.Time{}),
	)

	performer := schemaorg.NewPerson("Clara Schumann")
	performer.Context = ""

	// A mixed event takes place at a venue and online at the same time.
	venue := schemaorg.NewPlace("Konzerthaus")
	venue.Context = ""
	stream := schemaorg.NewVirtualLocation("https://example.com/live")
	stream.Context = ""

	event := schemaorg.NewEvent("Autumn Concert", start)
	event.EndDate = schemaorg.DateTime(start.Add(2 * time.Hour))
	event.PreviousStartDate = schemaorg.DateTime(start.AddDate(0, 0, -7))
	event.EventStatus = schemaorg.EventStatusRescheduled
	event.EventAttendanceMode = schemaorg.EventAttendanceModeMixed
	event.Location = []any{venue, stream}
	event.Organizer = schemaorg.Thing{Type: "Organization", ID: "https://example.com/#organization"}
	event.Performer = performer
	event.InLanguage = "de"
	requireGolden(t, "event", event)
}

func TestEventStatus(t *testing.T) {
	tests := []struct {
		status schemaorg.EventStatus
		want   string
	}{
		{schemaorg.EventStatusScheduled, "https://schema.org/EventScheduled"},
		{schemaorg.EventStatusCancelled, "https://schema.org/EventCancelled"},
		{schemaorg.EventStatusMovedOnline, "https://schema.org/EventMovedOnline"},
		{schemaorg.EventStatusPostponed, "https://schema.org/EventPostponed"},
		{schemaorg.EventStatusRescheduled, "https://schema.org/EventRescheduled"},
	}
	require.Equal(t, len(tests), len(schemaorg.EventStatus("").Enums()))
	for i, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			require.Equal(t, tt.status, tt.status.Enums()[i])
			require.Equal(t, tt.want, tt.status.EnumStrings()[i])
			require.True(t, tt.status.Valid())
			require.NoError(t, tt.status.Validate())
			require.Equal(t, tt.want, tt.status.String())
			requireJSON(t,
				`{"@type":"","eventStatus":"`+tt.want+`"}`,
				schemaorg.Event{EventStatus: tt.status},
			)
		})
	}

	invalid := schemaorg.EventStatus("EventCancelled")
	require.False(t, invalid.Valid())
	require.EqualError(t, invalid.Validate(), `invalid value "EventCancelled" for type schemaorg.EventStatus`)
}

func TestEventAttendanceMode(t *testing.T) {
	tests := []struct {
		mode schemaorg.EventAttendanceMode
		want string
	}{
		{schemaorg.EventAttendanceModeOffline, "https://schema.org/OfflineEventAttendanceMode"},
		{schemaorg.EventAttendanceModeOnline, "https://schema.org/OnlineEventAttendanceMode"},
		{schemaorg.EventAttendanceModeMixed, "https://schema.org/MixedEventAttendanceMode"},
	}
	require.Equal(t, len(tests), len(schemaorg.EventAttendanceMode("").Enums()))
	for i, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			require.Equal(t, tt.mode, tt.mode.Enums()[i])
			require.Equal(t, tt.want, tt.mode.EnumStrings()[i])
			require.True(t, tt.mode.Valid())
			require.NoError(t, tt.mode.Validate())
			require.Equal(t, tt.want, tt.mode.String())
			requireJSON(t,
				`{"@type":"","eventAttendanceMode":"`+tt.want+`"}`,
				schemaorg.Event{EventAttendanceMode: tt.mode},
			)
		})
	}

	invalid := schemaorg.EventAttendanceMode("online")
	require.False(t, invalid.Valid())
	require.EqualError(t, invalid.Validate(), `invalid value "online" for type schemaorg.EventAttendanceMode`)
}

// TestEventJSONLD renders a complete event description the way a page embeds
// it and reads it back the way a crawler would.
func TestEventJSONLD(t *testing.T) {
	const hostile = `Doors open 19:00.</script><script>alert(1)</script><!-- & more`

	address := schemaorg.NewPostalAddress()
	address.StreetAddress = "Lothringerstraße 20"
	address.AddressLocality = "Vienna"
	address.PostalCode = "1030"
	address.AddressCountry = "AT"

	venue := schemaorg.NewPlace("Konzerthaus")
	venue.Address = address

	organizer := schemaorg.NewOrganization("Example Events")
	organizer.URL = "https://example.com/"

	ticket := schemaorg.NewOffer(schemaorg.Price(4500), "EUR")
	ticket.URL = "https://example.com/tickets"
	ticket.Availability = schemaorg.AvailabilityInStock
	ticket.ValidFrom = schemaorg.Date(time.Date(2026, 9, 1, 0, 0, 0, 0, cest))

	start := time.Date(2026, 10, 15, 19, 30, 0, 0, cest)
	event := schemaorg.NewEvent("Autumn Concert", start)
	event.Description = hostile
	event.EndDate = schemaorg.DateTime(start.Add(2 * time.Hour))
	event.EventStatus = schemaorg.EventStatusScheduled
	event.EventAttendanceMode = schemaorg.EventAttendanceModeOffline
	event.Location = venue
	event.Organizer = organizer
	event.Offers = []*schemaorg.Offer{ticket}

	script := html.JSONLD(event)
	require.NoError(t, script.Err)
	var b strings.Builder
	err := script.Render(context.Background(), mx.NewCheckedWriter(&b))
	require.NoError(t, err)
	rendered := b.String()

	// The hostile description cannot end the script or open a comment.
	require.Equal(t, 1, strings.Count(rendered, "</script>"), rendered)
	require.NotContains(t, rendered, "<!--")
	require.NotContains(t, rendered, "alert(1)</")

	body, found := strings.CutPrefix(rendered,
		/*html*/ `<script type="application/ld+json">`,
	)
	require.True(t, found, rendered)
	body, found = strings.CutSuffix(body,
		/*html*/ `</script>`,
	)
	require.True(t, found, rendered)

	var got map[string]any
	err = json.Unmarshal([]byte(body), &got)
	require.NoError(t, err)

	require.Equal(t, "https://schema.org", got["@context"])
	require.Equal(t, "Event", got["@type"])
	require.Equal(t, "Autumn Concert", got["name"])
	require.Equal(t, hostile, got["description"], "escaping must not change the text")
	require.Equal(t, "2026-10-15T19:30:00+02:00", got["startDate"])
	require.Equal(t, "2026-10-15T21:30:00+02:00", got["endDate"])
	require.Equal(t, "https://schema.org/EventScheduled", got["eventStatus"])
	require.Equal(t, "https://schema.org/OfflineEventAttendanceMode", got["eventAttendanceMode"])

	require.Equal(t,
		map[string]any{
			"@context": "https://schema.org",
			"@type":    "Place",
			"name":     "Konzerthaus",
			"address": map[string]any{
				"@context":        "https://schema.org",
				"@type":           "PostalAddress",
				"streetAddress":   "Lothringerstraße 20",
				"addressLocality": "Vienna",
				"postalCode":      "1030",
				"addressCountry":  "AT",
			},
		},
		got["location"],
	)
	require.Equal(t,
		map[string]any{
			"@context": "https://schema.org",
			"@type":    "Organization",
			"name":     "Example Events",
			"url":      "https://example.com/",
		},
		got["organizer"],
	)
	require.Equal(t,
		[]any{map[string]any{
			"@context":      "https://schema.org",
			"@type":         "Offer",
			"url":           "https://example.com/tickets",
			"price":         "45.00",
			"priceCurrency": "EUR",
			"availability":  "https://schema.org/InStock",
			"validFrom":     "2026-09-01",
		}},
		got["offers"],
	)
	require.Len(t, got, 11, "no property beyond the ones set")
}
