package schemaorg_test

import (
	"context"
	"os"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestOrganization(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Organization","name":"Example"}`,
		schemaorg.NewOrganization("Example"),
	)

	address := schemaorg.NewPostalAddress()
	address.StreetAddress = "Mariahilfer Straße 1"
	address.AddressLocality = "Vienna"
	address.PostalCode = "1060"
	address.AddressCountry = "AT"

	organization := schemaorg.NewOrganization("Example")
	organization.ID = "https://example.com/#organization"
	organization.URL = "https://example.com/"
	organization.LegalName = "Example GmbH"
	organization.Logo = "https://example.com/logo.png"
	organization.Email = "office@example.com"
	organization.Telephone = "+43 1 234567"
	organization.Address = address
	organization.VATID = "ATU12345678"
	organization.SameAs = []string{"https://www.linkedin.com/company/example"}
	requireGolden(t, "organization", organization)
}

// A schema.org value goes into a page as the JSON-LD script html.JSONLD
// builds, usually as the HeadCustom of an html.Document.
func ExampleNewOrganization() {
	org := schemaorg.NewOrganization("Example Inc.")
	org.URL = "https://example.com/"
	org.Logo = "https://example.com/logo.png"

	err := html.JSONLD(org).Render(context.Background(), mx.NewCheckedWriter(os.Stdout))
	if err != nil {
		panic(err)
	}

	// Output:
	// <script type="application/ld+json">{"@context":"https://schema.org","@type":"Organization","name":"Example Inc.","url":"https://example.com/","logo":"https://example.com/logo.png"}</script>
}
