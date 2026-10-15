package schemaorg

// https://schema.org/WebPage
// https://schema.org/AboutPage
// https://schema.org/ContactPage
// https://schema.org/CollectionPage

// WebPage is a single page of a site. Its specialized types [AboutPage],
// [ContactPage], [CollectionPage] and [FAQPage] embed it, so it has no property
// that one of them would need with a different type — mainEntity, which
// FAQPage types as its questions, is left to the subtypes for that reason.
//
// https://schema.org/WebPage
type WebPage struct {
	Thing

	// InLanguage is the language of the page as a BCP 47 language tag ("en",
	// "de-AT").
	InLanguage string `json:"inLanguage,omitempty"`

	// IsPartOf is the site the page belongs to: a [WebSite], its URL as
	// text, or a [Thing] holding just the Type and ID of a WebSite published
	// elsewhere.
	IsPartOf any `json:"isPartOf,omitempty"`

	// DatePublished is when the page was first published, built with [Date]
	// or [DateTime].
	DatePublished string `json:"datePublished,omitempty"`

	// DateModified is when the page content last changed, built with [Date]
	// or [DateTime]. Like a sitemap's lastmod it is only useful while it
	// moves with real content changes, not with every deploy.
	DateModified string `json:"dateModified,omitempty"`

	// Author is the author of the page: a [Person] or an [Organization].
	Author any `json:"author,omitempty"`

	// Publisher is the party publishing the page: an [Organization] or a
	// [Person].
	Publisher any `json:"publisher,omitempty"`

	// About is the subject of the page, any schema.org type: the [Person] of
	// a profile, the [Event] of an event page.
	About any `json:"about,omitempty"`
}

// NewWebPage returns a WebPage with the given title and canonical URL.
func NewWebPage(name, url string) *WebPage {
	return newWebPage("WebPage", name, url)
}

// AboutPage is the page about the site or the organization or person behind
// it.
//
// https://schema.org/AboutPage
type AboutPage struct {
	WebPage
}

// NewAboutPage returns an AboutPage with the given title and canonical URL.
func NewAboutPage(name, url string) *AboutPage {
	return &AboutPage{WebPage: *newWebPage("AboutPage", name, url)}
}

// ContactPage is the page with the contact details of the site.
//
// https://schema.org/ContactPage
type ContactPage struct {
	WebPage
}

// NewContactPage returns a ContactPage with the given title and canonical URL.
func NewContactPage(name, url string) *ContactPage {
	return &ContactPage{WebPage: *newWebPage("ContactPage", name, url)}
}

// CollectionPage is a page listing other items, such as a blog index, an
// event calendar or a tag archive. The listed items go in an [ItemList] of
// their own.
//
// https://schema.org/CollectionPage
type CollectionPage struct {
	WebPage
}

// NewCollectionPage returns a CollectionPage with the given title and
// canonical URL.
func NewCollectionPage(name, url string) *CollectionPage {
	return &CollectionPage{WebPage: *newWebPage("CollectionPage", name, url)}
}

// newWebPage returns a WebPage of the given type, the shared part of the
// WebPage family's constructors.
func newWebPage(typ, name, url string) *WebPage {
	page := &WebPage{Thing: newThing(typ)}
	page.Name = name
	page.URL = url
	return page
}
