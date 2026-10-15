package schemaorg_test

import (
	"testing"

	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestPlace(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Place","name":"Konzerthaus"}`,
		schemaorg.NewPlace("Konzerthaus"),
	)

	address := schemaorg.NewPostalAddress()
	address.StreetAddress = "Lothringerstraße 20"
	address.AddressLocality = "Vienna"
	address.PostalCode = "1030"
	address.AddressCountry = "AT"

	place := schemaorg.NewPlace("Konzerthaus")
	place.URL = "https://konzerthaus.example/"
	place.Address = address
	requireGolden(t, "place", place)
}

func TestVirtualLocation(t *testing.T) {
	location := schemaorg.NewVirtualLocation("https://example.com/live")
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"VirtualLocation","url":"https://example.com/live"}`,
		location,
	)

	location.Name = "Live stream"
	requireGolden(t, "virtuallocation", location)
}

func TestPostalAddress(t *testing.T) {
	// No part of an address is required, so the constructor sets none.
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"PostalAddress"}`,
		schemaorg.NewPostalAddress(),
	)

	address := schemaorg.NewPostalAddress()
	address.StreetAddress = "1 Infinite Loop"
	address.AddressLocality = "Cupertino"
	address.AddressRegion = "CA"
	address.PostalCode = "95014"
	address.AddressCountry = "US"
	requireGolden(t, "postaladdress", address)
}
