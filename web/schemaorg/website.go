package schemaorg

// https://schema.org/WebSite
// https://developers.google.com/search/docs/appearance/site-names

// WebSite is a whole site. Published on the home page it is where search
// engines take the site name they show next to a result from.
//
// https://schema.org/WebSite
type WebSite struct {
	Thing

	// InLanguage is the language of the site's content as a BCP 47 language
	// tag ("en", "de-AT").
	InLanguage string `json:"inLanguage,omitempty"`

	// Publisher is the party publishing the site: an [Organization] or a
	// [Person], or a [Thing] holding just the Type and ID of one published
	// elsewhere.
	Publisher any `json:"publisher,omitempty"`
}

// NewWebSite returns a WebSite with the given name and home page URL, the two
// properties a site name is taken from.
func NewWebSite(name, url string) *WebSite {
	site := &WebSite{Thing: newThing("WebSite")}
	site.Name = name
	site.URL = url
	return site
}
