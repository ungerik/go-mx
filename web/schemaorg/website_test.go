package schemaorg_test

import (
	"testing"

	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestWebSite(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"WebSite","name":"Example","url":"https://example.com/"}`,
		schemaorg.NewWebSite("Example", "https://example.com/"),
	)

	site := schemaorg.NewWebSite("Example", "https://example.com/")
	site.ID = "https://example.com/#website"
	site.Description = "Everything about Example."
	site.InLanguage = "en"
	// The publisher is published on its own, so the site only refers to it.
	site.Publisher = schemaorg.Thing{Type: "Organization", ID: "https://example.com/#organization"}
	requireGolden(t, "website", site)
}
