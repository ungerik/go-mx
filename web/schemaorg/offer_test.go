package schemaorg_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestOffer(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Offer","price":"180.00","priceCurrency":"EUR"}`,
		schemaorg.NewOffer(schemaorg.Price(18000), "EUR"),
	)

	seller := schemaorg.NewOrganization("Example Tickets")
	seller.Context = ""

	offer := schemaorg.NewOffer(schemaorg.Price(4500), "EUR")
	offer.Name = "Standard ticket"
	offer.URL = "https://example.com/tickets"
	offer.PriceValidUntil = schemaorg.Date(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
	offer.Availability = schemaorg.AvailabilityLimited
	offer.ValidFrom = schemaorg.DateTime(time.Date(2026, 9, 1, 10, 0, 0, 0, cest))
	offer.ValidThrough = schemaorg.Date(time.Date(2026, 10, 15, 0, 0, 0, 0, cest))
	offer.Seller = seller
	requireGolden(t, "offer", offer)
}

func TestPrice(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{0, "0.00"},
		{1, "0.01"},
		{5, "0.05"},
		{10, "0.10"},
		{99, "0.99"},
		{100, "1.00"},
		{18000, "180.00"},
		{100005, "1000.05"}, // no thousands separator
		{-1, "-0.01"},
		{-5, "-0.05"},
		{-100, "-1.00"},
		{-18000, "-180.00"},
		{math.MaxInt64, "92233720368547758.07"},
		// The magnitude of the smallest int64 does not fit an int64.
		{math.MinInt64, "-92233720368547758.08"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, schemaorg.Price(tt.cents), "Price(%d)", tt.cents)
	}
}

func TestAvailability(t *testing.T) {
	tests := []struct {
		availability schemaorg.Availability
		want         string
	}{
		{schemaorg.AvailabilityInStock, "https://schema.org/InStock"},
		{schemaorg.AvailabilityLimited, "https://schema.org/LimitedAvailability"},
		{schemaorg.AvailabilitySoldOut, "https://schema.org/SoldOut"},
		{schemaorg.AvailabilityPreOrder, "https://schema.org/PreOrder"},
		{schemaorg.AvailabilityDiscontinued, "https://schema.org/Discontinued"},
		{schemaorg.AvailabilityOutOfStock, "https://schema.org/OutOfStock"},
	}
	require.Equal(t, len(tests), len(schemaorg.Availability("").Enums()))
	for i, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			require.Equal(t, tt.availability, tt.availability.Enums()[i])
			require.Equal(t, tt.want, tt.availability.EnumStrings()[i])
			require.True(t, tt.availability.Valid())
			require.NoError(t, tt.availability.Validate())
			require.Equal(t, tt.want, tt.availability.String())
			requireJSON(t,
				`{"@type":"","availability":"`+tt.want+`"}`,
				schemaorg.Offer{Availability: tt.availability},
			)
		})
	}

	// The bare schema.org name is not the value: Google expects the URL.
	invalid := schemaorg.Availability("InStock")
	require.False(t, invalid.Valid())
	require.EqualError(t, invalid.Validate(), `invalid value "InStock" for type schemaorg.Availability`)
}
