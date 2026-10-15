// Package schemaorg is a typed subset of the schema.org vocabulary for
// publishing structured data as JSON-LD: the people, organizations, pages,
// events, offers and FAQs that search engines read from a
// <script type="application/ld+json"> to understand what a page is about.
//
// Every type is a plain struct that encoding/json marshals into its JSON-LD
// form, so the package has no rendering code of its own. Embed a value into a
// page with html.JSONLD from github.com/ungerik/go-mx/html:
//
//	org := schemaorg.NewOrganization("Example Inc.")
//	org.URL = "https://example.com/"
//	org.Logo = "https://example.com/logo.png"
//
//	page := &html.Document{
//		Title:      "Example Inc.",
//		HeadCustom: html.JSONLD(org),
//	}
//
// The package describes the vocabulary only: it knows nothing about a
// particular site, and it does not check that a value satisfies the
// requirements a search engine puts on a rich result. Google's structured data
// documentation lists those per feature.
//
// # Types and properties
//
// Each type embeds [Thing], or the struct of its schema.org parent type (an
// [AboutPage] embeds [WebPage]), so the common properties are shared and
// encoding/json flattens them into one object. The properties are a practical
// subset of schema.org, the ones Google's structured data features actually
// read. Every property but @type is omitted while empty, so a value carries
// exactly what was set. A schema.org subtype without properties of its own can
// reuse the struct of its parent by overwriting [Thing.Type], e.g. an
// [ItemList] with Type "BreadcrumbList"; [NewThing] covers a type with no
// struct here at all as far as the common properties describe it.
//
// Where schema.org allows values of several types for a property — an
// organizer that is a [Person] or an [Organization], a location that is a
// [Place], a [VirtualLocation] or a URL — the field is typed any and its doc
// comment lists the types allowed there. Anything else marshals as well, so
// staying within the vocabulary is up to the caller.
//
// # Dates and prices
//
// Dates and times are ISO 8601 strings built with [Date] for a calendar day and
// [DateTime] for an instant with its UTC offset. Both return "" for the zero
// time, so an unset time.Time leaves the property out.
//
// Prices are decimal strings built from integer cents with [Price], never
// floating point: a binary float cannot hold most decimal amounts exactly,
// and a price should not depend on how its nearest float is rounded back to
// decimal.
//
// # The @context
//
// A JSON-LD script needs a @context naming the vocabulary its terms belong to.
// Every constructor sets [Thing.Context] to [Context] along with the @type, so
// whatever it returns is ready to be the root of a script. A value nested in
// another one keeps its @context. That is redundant but valid JSON-LD — an
// embedded context identical to the outer one changes nothing — so crawlers
// read it the same way. Clear Context on nested values for the minimal form.
// The one value a constructor nests itself, the Answer of [NewQuestion], is
// created without it.
package schemaorg
