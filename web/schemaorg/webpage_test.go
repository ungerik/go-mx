package schemaorg_test

import (
	"testing"
	"time"

	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestWebPage(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"WebPage","name":"Blog","url":"https://example.com/blog/"}`,
		schemaorg.NewWebPage("Blog", "https://example.com/blog/"),
	)

	author := schemaorg.NewPerson("Ada Lovelace")
	author.Context = ""
	author.URL = "https://example.com/about"

	page := schemaorg.NewWebPage("Notes on the Engine", "https://example.com/blog/notes")
	page.Description = "Notes on the Analytical Engine."
	page.InLanguage = "en"
	page.IsPartOf = schemaorg.Thing{Type: "WebSite", ID: "https://example.com/#website"}
	page.DatePublished = schemaorg.Date(time.Date(2026, 10, 1, 9, 0, 0, 0, cest))
	page.DateModified = schemaorg.DateTime(time.Date(2026, 10, 15, 19, 30, 0, 0, cest))
	page.Author = author
	page.Publisher = schemaorg.Thing{Type: "Organization", ID: "https://example.com/#organization"}
	page.About = schemaorg.Thing{Type: "Thing", Name: "Analytical Engine"}
	requireGolden(t, "webpage", page)
}

func TestAboutPage(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"AboutPage","name":"About","url":"https://example.com/about"}`,
		schemaorg.NewAboutPage("About", "https://example.com/about"),
	)

	person := schemaorg.NewPerson("Ada Lovelace")
	person.Context = ""
	person.ID = "https://example.com/about#ada"

	page := schemaorg.NewAboutPage("About", "https://example.com/about")
	page.InLanguage = "en"
	page.About = person
	requireGolden(t, "aboutpage", page)
}

func TestContactPage(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"ContactPage","name":"Contact","url":"https://example.com/contact"}`,
		schemaorg.NewContactPage("Contact", "https://example.com/contact"),
	)

	organization := schemaorg.NewOrganization("Example")
	organization.Context = ""
	organization.Email = "office@example.com"
	organization.Telephone = "+43 1 234567"

	page := schemaorg.NewContactPage("Contact", "https://example.com/contact")
	page.About = organization
	requireGolden(t, "contactpage", page)
}

func TestCollectionPage(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"CollectionPage","name":"Events","url":"https://example.com/events/"}`,
		schemaorg.NewCollectionPage("Events", "https://example.com/events/"),
	)

	page := schemaorg.NewCollectionPage("Events", "https://example.com/events/")
	page.Description = "Upcoming concerts and talks."
	page.InLanguage = "en"
	page.IsPartOf = "https://example.com/"
	requireGolden(t, "collectionpage", page)
}
