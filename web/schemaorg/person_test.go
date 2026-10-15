package schemaorg_test

import (
	"testing"

	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestPerson(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Person","name":"Ada Lovelace"}`,
		schemaorg.NewPerson("Ada Lovelace"),
	)

	employer := schemaorg.NewOrganization("Analytical Engines Ltd")
	employer.Context = "" // nested: the minimal form without the redundant context
	employer.URL = "https://analytical-engines.example/"

	address := schemaorg.NewPostalAddress()
	address.Context = ""
	address.StreetAddress = "12 St James's Square"
	address.AddressLocality = "London"
	address.PostalCode = "SW1Y 4JH"
	address.AddressCountry = "GB"

	person := schemaorg.NewPerson("Ada Lovelace")
	person.ID = "https://example.com/about#ada"
	person.GivenName = "Ada"
	person.FamilyName = "Lovelace"
	person.JobTitle = "Mathematician"
	person.Email = "ada@example.com"
	person.Telephone = "+44 20 7946 0000"
	person.WorksFor = employer
	person.Address = address
	person.KnowsLanguage = []string{"en", "fr"}
	person.SameAs = []string{"https://en.wikipedia.org/wiki/Ada_Lovelace"}
	requireGolden(t, "person", person)
}
