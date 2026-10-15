package web

// https://developers.google.com/search/docs/specialty/international/localized-versions
// https://www.rfc-editor.org/rfc/rfc8288.html

import (
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

const (
	// RelAlternate is the link relation of an alternate representation of a
	// page: a translation, or the same page in another format.
	RelAlternate = "alternate"

	// RelDescribedBy is the link relation of a document describing the page,
	// for example its metadata in another format.
	RelDescribedBy = "describedby"

	// HrefLangDefault is the hreflang value "x-default" marking the version of
	// a page shown to visitors whose language has no translation of its own.
	HrefLangDefault = "x-default"
)

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
	Rel string

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
func (a Alternate) RelOrDefault() string {
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

// Validate returns an error if Rel is neither [RelAlternate] nor
// [RelDescribedBy], Type is not a media type, HrefLang is not a language tag,
// Path is not a valid path within a site, or if an alternate has neither a
// Type nor a HrefLang, which would leave nothing to tell it apart from the
// page.
func (a Alternate) Validate() error {
	switch a.RelOrDefault() {
	case RelAlternate, RelDescribedBy:
	default:
		return fmt.Errorf("invalid alternate rel %q, must be %q or %q", a.Rel, RelAlternate, RelDescribedBy)
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
	if a.HrefLang != "" && !langTagRegexp.MatchString(a.HrefLang) {
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
	return a.RelOrDefault() + "\x00" + strings.ToLower(a.Type) + "\x00" + strings.ToLower(a.HrefLang)
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

// headerValue returns the link as the value of a Link header (RFC 8288).
// Type and HrefLang have been validated to hold no quote. The URL path is
// escaped by [Site.URL], but the host comes from Site.BaseURL, which the URL
// parser accepts with a ">" or a quote in it, so those are percent-encoded
// here rather than trusted to end the URL reference where it should.
func (l pageLink) headerValue() string {
	var b strings.Builder
	u := strings.NewReplacer("<", "%3C", ">", "%3E", `"`, "%22").Replace(l.url)
	b.WriteString("<" + u + `>; rel="` + l.rel + `"`)
	if l.typ != "" {
		b.WriteString(`; type="` + l.typ + `"`)
	}
	if l.hrefLang != "" {
		b.WriteString(`; hreflang="` + l.hrefLang + `"`)
	}
	return b.String()
}

// pageLang returns the language of the page: its own Lang, or the Lang of the
// site.
func (s *Site) pageLang(page *Page) string {
	return firstNonEmpty(page.Lang, s.Lang)
}

// pageLinks returns the links of a page, the one list both
// [Site.AddPageMetadata] and [Site.PageHeaders] render: the canonical link
// first, then the page's own hreflang link if it has translations, then its
// Alternates in order, and finally the x-default link derived from
// Site.DefaultLang. A page without a Path has no URL to link and no links.
func (s *Site) pageLinks(page *Page) ([]pageLink, error) {
	if err := page.Validate(); err != nil {
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

	lang := s.pageLang(page)
	var defaultURL string
	var explicitDefault bool
	for _, alternate := range page.Alternates {
		if !alternate.IsTranslation() {
			continue
		}
		// A set of translations is only followed if every version names
		// itself too, so a page that has some has to know its own language.
		if lang == "" {
			return nil, fmt.Errorf("page %q has translations but no language, set Page.Lang or Site.Lang", page.Path)
		}
		// Page.Lang has been validated with the page, but Site.Lang only
		// by Site.Validate, which nothing here can rely on having run.
		if err := validateLang(lang); err != nil {
			return nil, fmt.Errorf("page %q: language: %w", page.Path, err)
		}
		links = append(links, pageLink{rel: RelAlternate, hrefLang: lang, url: canonical})
		if strings.EqualFold(lang, s.DefaultLang) {
			defaultURL = canonical
		}
		break
	}
	for _, alternate := range page.Alternates {
		u, err := s.URL(alternate.Path)
		if err != nil {
			return nil, fmt.Errorf("page %q: alternate: %w", page.Path, err)
		}
		if alternate.IsTranslation() {
			switch {
			case strings.EqualFold(alternate.HrefLang, lang):
				return nil, fmt.Errorf("page %q: alternate %q has the language %q of the page itself", page.Path, alternate.Path, lang)
			case strings.EqualFold(alternate.HrefLang, HrefLangDefault):
				explicitDefault = true
			case strings.EqualFold(alternate.HrefLang, s.DefaultLang):
				defaultURL = u
			}
		}
		links = append(links, pageLink{rel: alternate.RelOrDefault(), typ: alternate.Type, hrefLang: alternate.HrefLang, url: u})
	}
	if defaultURL != "" && !explicitDefault {
		links = append(links, pageLink{rel: RelAlternate, hrefLang: HrefLangDefault, url: defaultURL})
	}
	return links, nil
}

// PageHeaders returns the Link headers of the page: its canonical URL and its
// [Alternate]s, built from the same list [Site.AddPageMetadata] renders into
// the HTML head, so a crawler reading either finds the same links. A page
// without a Path has none and gets an empty header.
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
