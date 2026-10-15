//go:generate go -C ../tools tool go-enum ../web/$GOFILE

package web

// https://developers.google.com/search/docs/specialty/international/localized-versions
// https://www.rfc-editor.org/rfc/rfc8288.html

import (
	"fmt"
	"mime"
	"net/http"
	"slices"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AlternateRel is the link relation of an [Alternate].
type AlternateRel string //#enum

const (
	// RelAlternate is the link relation of an alternate representation of a
	// page: a translation, or the same page in another format.
	RelAlternate AlternateRel = "alternate"

	// RelDescribedBy is the link relation of a document describing the page,
	// for example its metadata in another format.
	RelDescribedBy AlternateRel = "describedby"
)

// Valid indicates if a is any of the valid values for AlternateRel
func (a AlternateRel) Valid() bool {
	switch a {
	case
		RelAlternate,
		RelDescribedBy:
		return true
	}
	return false
}

// Validate returns an error if a is none of the valid values for AlternateRel
func (a AlternateRel) Validate() error {
	if !a.Valid() {
		return fmt.Errorf("invalid value %#v for type web.AlternateRel", a)
	}
	return nil
}

// ValidOr returns a if it is any of the valid values for AlternateRel, else defaultVal
func (a AlternateRel) ValidOr(defaultVal AlternateRel) AlternateRel {
	if a.Valid() {
		return a
	}
	return defaultVal
}

// Enums returns all valid values for AlternateRel
func (AlternateRel) Enums() []AlternateRel {
	return []AlternateRel{
		RelAlternate,
		RelDescribedBy,
	}
}

// EnumStrings returns all valid values for AlternateRel as strings
func (AlternateRel) EnumStrings() []string {
	return []string{
		"alternate",
		"describedby",
	}
}

// String implements the fmt.Stringer interface for AlternateRel
func (a AlternateRel) String() string {
	return string(a)
}

// HrefLangDefault is the hreflang value "x-default" marking the version of a
// page shown to visitors whose language has no translation of its own.
const HrefLangDefault = "x-default"

// Alternate is another representation of a [Page], linked from the page as
// <link rel="alternate"> in its HTML head and as a Link header of its
// response, from the same list so the two cannot drift apart.
//
// An alternate with a HrefLang and no Type is a translation: the same page in
// another language. Search engines use translations to send each visitor to
// the version in their language, which only works if every version links all
// the others and itself. The link to itself is added by the [Site] from the
// page's language, so Alternates lists only the other versions:
//
//	page := &web.Page{
//	    Path: "/en/faq",
//	    Lang: "en",
//	    Alternates: []web.Alternate{
//	        {HrefLang: "de", Path: "/de/faq"},
//	        {Type: "text/markdown", Path: "/en/faq.md"},
//	    },
//	}
type Alternate struct {
	// Rel is the link relation, [RelAlternate] when empty, or
	// [RelDescribedBy].
	Rel AlternateRel

	// Type is the media type of the alternate, for example "text/markdown".
	// It is empty for a translation, which has the type of the page.
	Type string

	// HrefLang is the BCP 47 language tag of the alternate, or
	// [HrefLangDefault]. [Site.DefaultLang] adds the x-default link for most
	// sites; set it here only to point it somewhere else, like a language
	// chooser.
	HrefLang string

	// Path is the absolute path of the alternate within the site, validated
	// like [Page.Path].
	Path string
}

// RelOrDefault returns Rel, or [RelAlternate] if it is empty.
func (a Alternate) RelOrDefault() AlternateRel {
	if a.Rel == "" {
		return RelAlternate
	}
	return a.Rel
}

// IsTranslation reports whether the alternate is a translation of the page: an
// alternate with a language and the media type of the page itself. An
// x-default link counts too, because it belongs to the same set of links.
func (a Alternate) IsTranslation() bool {
	return a.RelOrDefault() == RelAlternate && a.Type == "" && a.HrefLang != ""
}

// isOwnMarkdown reports whether the alternate is the page's own Markdown
// version: a text/markdown alternate without a language, which would be the
// Markdown version of a translation instead.
func (a Alternate) isOwnMarkdown() bool {
	if a.RelOrDefault() != RelAlternate || a.HrefLang != "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(a.Type)
	return err == nil && mediaType == "text/markdown"
}

// Validate returns an error if Rel is neither [RelAlternate] nor
// [RelDescribedBy], Type is not a media type, HrefLang is not a language tag,
// Path is not a valid path within a site, or if an alternate has neither a
// Type nor a HrefLang, which would leave nothing to tell it apart from the
// page.
func (a Alternate) Validate() error {
	if err := a.RelOrDefault().Validate(); err != nil {
		return fmt.Errorf("alternate rel: %w", err)
	}
	if a.Type != "" {
		// A quote, backslash or control character would end the quoted
		// type parameter of the Link header early, and ParseMediaType
		// accepts a quoted parameter value holding one.
		if strings.ContainsFunc(a.Type, func(r rune) bool { return r < ' ' || r == 0x7F || r == '"' || r == '\\' }) {
			return fmt.Errorf("alternate type %q contains a quote, backslash or control character", a.Type)
		}
		// ParseMediaType also accepts a bare token, the form of a
		// Content-Disposition, so the type/subtype form is checked separately.
		mediaType, _, err := mime.ParseMediaType(a.Type)
		if err != nil {
			return fmt.Errorf("alternate type %q is not a media type: %w", a.Type, err)
		}
		if !strings.Contains(mediaType, "/") {
			return fmt.Errorf("alternate type %q is not a media type: missing subtype", a.Type)
		}
	}
	if a.HrefLang != "" && !html.ValidLangTag(a.HrefLang) {
		return fmt.Errorf("alternate hreflang %q is not a valid BCP 47 language tag", a.HrefLang)
	}
	if a.RelOrDefault() == RelAlternate && a.Type == "" && a.HrefLang == "" {
		return fmt.Errorf("alternate %q has neither a type nor a hreflang", a.Path)
	}
	if err := validateSitePath(a.Path); err != nil {
		return fmt.Errorf("alternate path: %w", err)
	}
	return nil
}

// key identifies what the alternate is an answer to. Two alternates with the
// same key give crawlers two URLs for one question. Media types and language
// tags are both case-insensitive.
func (a Alternate) key() string {
	return string(a.RelOrDefault()) + "\x00" + strings.ToLower(a.Type) + "\x00" + strings.ToLower(a.HrefLang)
}

// pageLink is one link of a page as [Site.AddPageMetadata] renders it into
// the HTML head and [Site.PageHeaders] into a Link header.
type pageLink struct {
	rel      string
	typ      string
	hrefLang string
	url      string
}

// element returns the link as a <link> element.
func (l pageLink) element() *mx.Element {
	attribs := []mx.Attrib{html.Rel(l.rel)}
	if l.typ != "" {
		attribs = append(attribs, html.Type(l.typ))
	}
	if l.hrefLang != "" {
		attribs = append(attribs, html.HRefLang(l.hrefLang))
	}
	return html.Link(append(attribs, html.HRef(l.url))...)
}

// linkHeaderURLEscaper percent-encodes what would end the URL reference of a
// Link header early.
var linkHeaderURLEscaper = strings.NewReplacer("<", "%3C", ">", "%3E", `"`, "%22")

// headerValue returns the link as the value of a Link header (RFC 8288).
// Type and HrefLang have been validated to hold no quote. The URL path is
// escaped by [Site.URL], but the host comes from Site.BaseURL, which the URL
// parser accepts with a ">" or a quote in it, so those are percent-encoded
// here rather than trusted to end the URL reference where it should.
func (l pageLink) headerValue() string {
	var b strings.Builder
	b.WriteString("<" + linkHeaderURLEscaper.Replace(l.url) + `>; rel="` + l.rel + `"`)
	if l.typ != "" {
		b.WriteString(`; type="` + l.typ + `"`)
	}
	if l.hrefLang != "" {
		b.WriteString(`; hreflang="` + l.hrefLang + `"`)
	}
	return b.String()
}

// PageLang returns the language of the page: its own Lang, or the Lang of the
// site. It is the language of the html lang attribute, of the page's own
// hreflang link and of its Markdown alternate, so anything else describing
// the page should report it too.
func (s *Site) PageLang(page *Page) string {
	return firstNonEmpty(page.Lang, s.Lang)
}

// pageLinks returns the links of a page, the one list both
// [Site.AddPageMetadata] and [Site.PageHeaders] render: the canonical link
// first, then the page's own hreflang link if it has translations, then its
// Alternates in order, and finally the x-default link derived from
// Site.DefaultLang. A page without a Path has no URL to link and no links.
func (s *Site) pageLinks(page *Page) ([]pageLink, error) {
	if err := s.validatePageLinks(page); err != nil {
		return nil, err
	}
	if page.Path == "" {
		return nil, nil
	}
	canonical, err := s.PageURL(page)
	if err != nil {
		return nil, err
	}
	links := []pageLink{{rel: "canonical", url: canonical}}

	lang := s.PageLang(page)
	var defaultURL string
	var explicitDefault bool
	if slices.ContainsFunc(page.Alternates, Alternate.IsTranslation) {
		links = append(links, pageLink{rel: string(RelAlternate), hrefLang: lang, url: canonical})
		if strings.EqualFold(lang, s.DefaultLang) {
			defaultURL = canonical
		}
	}
	for _, alternate := range page.Alternates {
		u, err := s.URL(alternate.Path)
		if err != nil {
			return nil, fmt.Errorf("page %q: alternate: %w", page.Path, err)
		}
		if alternate.IsTranslation() {
			switch {
			case strings.EqualFold(alternate.HrefLang, HrefLangDefault):
				explicitDefault = true
			case strings.EqualFold(alternate.HrefLang, s.DefaultLang):
				defaultURL = u
			}
		}
		links = append(links, pageLink{rel: string(alternate.RelOrDefault()), typ: alternate.Type, hrefLang: alternate.HrefLang, url: u})
	}
	if defaultURL != "" && !explicitDefault {
		links = append(links, pageLink{rel: string(RelAlternate), hrefLang: HrefLangDefault, url: defaultURL})
	}
	return links, nil
}

// validatePageLinks returns the error [Site.pageLinks] fails with for the
// page, without building any of its URLs: what [Page.Validate] checks, plus
// what needs the site — a BaseURL to build URLs from and, for a page with
// translations, a valid language of its own that none of them repeats. It is
// what a caller that only needs to know whether the links can be built, like
// the sitemap, calls for every page.
func (s *Site) validatePageLinks(page *Page) error {
	if err := page.Validate(); err != nil {
		return err
	}
	if page.Path == "" {
		return nil
	}
	if _, err := s.baseURL(); err != nil {
		return err
	}
	if !slices.ContainsFunc(page.Alternates, Alternate.IsTranslation) {
		return nil
	}
	// A set of translations is only followed if every version names itself
	// too, so a page that has some has to know its own language.
	lang := s.PageLang(page)
	if lang == "" {
		return fmt.Errorf("page %q has translations but no language, set Page.Lang or Site.Lang", page.Path)
	}
	// Page.Lang has been validated with the page, but Site.Lang only by
	// Site.Validate, which nothing here can rely on having run.
	if err := validateLang(lang); err != nil {
		return fmt.Errorf("page %q: language: %w", page.Path, err)
	}
	for _, alternate := range page.Alternates {
		if alternate.IsTranslation() && strings.EqualFold(alternate.HrefLang, lang) {
			return fmt.Errorf("page %q: alternate %q has the language %q of the page itself", page.Path, alternate.Path, lang)
		}
	}
	return nil
}

// PageHeaders returns the Link headers of the page: its canonical URL and its
// [Alternate]s, built from the same list [Site.AddPageMetadata] renders into
// the HTML head, so a crawler reading either finds the same links. A page
// without a Path has none and gets an empty header.
//
// Every translation is one header line of about a hundred bytes, which adds
// up for a page in dozens of languages: proxies limit the size of response
// headers (nginx to 4 KiB by default). The links in the HTML head alone are
// enough for an HTML page; the headers matter most for responses without a
// head of their own.
//
// The translations are published as the page lists them. Alternates should
// name only public pages, since search engines discard a set that points at
// a page marked noindex, and a draft's URL would be published with it.
//
// Add them to the response of the page before writing it:
//
//	header, err := site.PageHeaders(page)
//	if err != nil { ... }
//	for _, link := range header.Values("Link") {
//	    response.Header().Add("Link", link)
//	}
func (s *Site) PageHeaders(page *Page) (http.Header, error) {
	links, err := s.pageLinks(page)
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	for _, link := range links {
		header.Add("Link", link.headerValue())
	}
	return header, nil
}
