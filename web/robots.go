package web

// https://developers.google.com/search/docs/crawling-indexing/robots/intro
// https://www.rfc-editor.org/rfc/rfc9309.html

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
)

// RobotsTxtPath is the only path a robots.txt file can be served from:
// crawlers request it from the root of the host and ignore a robots.txt
// served anywhere else (RFC 9309, section 2.3).
const RobotsTxtPath = "/robots.txt"

// robotsProductTokenRegexp matches the product token of a User-agent line,
// which RFC 9309 limits to letters, digits, underscores and dashes
// ("Googlebot", "Mediapartners-Google", …). The wildcard "*" addressing all
// crawlers is checked separately.
var robotsProductTokenRegexp = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Robots is the content of a robots.txt file: groups of crawl rules addressed
// at user agents, followed by the URLs of the site's sitemaps.
//
// [Robots.MarshalText] and [Robots.WriteTo] validate before writing, so a value
// that would silently change the meaning of the file — an unanchored path, a
// line break smuggled into a user agent — fails instead of being rendered.
//
// Use [AllowAllRobots] and [DisallowAllRobots] for the two common cases; a
// [Site] serves the robots.txt of its Robots field with the URL of its sitemap
// added (see [Site.RobotsTxt]).
type Robots struct {
	// Groups are the crawl rule groups, rendered in order and separated by
	// blank lines. No groups at all means an empty robots.txt, which allows
	// crawling everything.
	Groups []RobotsGroup

	// Sitemaps are absolute sitemap URLs rendered as "Sitemap:" lines after
	// the groups. They are independent of the groups: every crawler reads all
	// of them, which is why a sitemap must be referenced by its full URL.
	Sitemaps []string
}

// AllowAllRobots returns a Robots that allows every crawler to crawl the whole
// site, optionally referencing the given absolute sitemap URLs.
func AllowAllRobots(sitemaps ...string) *Robots {
	return &Robots{
		Groups:   []RobotsGroup{{Disallow: []string{""}}},
		Sitemaps: sitemaps,
	}
}

// DisallowAllRobots returns a Robots that asks every crawler to stay away from
// the whole site, the usual setting for a staging or preview deployment.
//
// It keeps a site out of crawls, not out of search results: a page linked from
// somewhere else can still be indexed without ever being fetched. Only a
// "noindex" robots meta tag on a crawlable page (see [Page.NoIndex]) reliably
// keeps it out of the index.
func DisallowAllRobots() *Robots {
	return &Robots{Groups: []RobotsGroup{{Disallow: []string{"/"}}}}
}

// Validate returns an error if any group or sitemap URL is invalid.
func (r *Robots) Validate() error {
	for i, group := range r.Groups {
		if err := group.Validate(); err != nil {
			return fmt.Errorf("robots.txt group %d: %w", i, err)
		}
	}
	for _, sitemap := range r.Sitemaps {
		if err := validateAbsoluteURL(sitemap); err != nil {
			return fmt.Errorf("robots.txt sitemap URL: %w", err)
		}
	}
	return nil
}

// MarshalText renders the robots.txt file, implementing encoding.TextMarshaler.
// It returns the error from [Robots.Validate] instead of writing a file whose
// meaning would differ from what the value describes.
func (r *Robots) MarshalText() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	for i, group := range r.Groups {
		if i > 0 {
			b.WriteByte('\n')
		}
		group.write(&b)
	}
	if len(r.Sitemaps) > 0 {
		if len(r.Groups) > 0 {
			b.WriteByte('\n')
		}
		for _, sitemap := range r.Sitemaps {
			fmt.Fprintf(&b, "Sitemap: %s\n", sitemap)
		}
	}
	return b.Bytes(), nil
}

// WriteTo writes the robots.txt file to w, implementing io.WriterTo.
func (r *Robots) WriteTo(w io.Writer) (int64, error) {
	text, err := r.MarshalText()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(text)
	return int64(n), err
}

// String renders the robots.txt file, returning a "!ERROR: " line instead of
// the file content if the value does not validate. It is meant for debugging
// and tests; use [Robots.MarshalText] to handle the error.
func (r *Robots) String() string {
	text, err := r.MarshalText()
	if err != nil {
		return "!ERROR: " + err.Error()
	}
	return string(text)
}

// HandleHTTP writes the robots.txt file as a plain text response,
// implementing http.HandlerFunc. On a validation error it responds with a
// generic 500 status via [mx.RespondNonContextError].
func (r *Robots) HandleHTTP(response http.ResponseWriter, request *http.Request) {
	text, err := r.MarshalText()
	if err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	response.Header().Set("Content-Type", mx.ContentTypePlainText)
	response.Write(text)
}

// RobotsGroup is one group of crawl rules in a robots.txt file: the user agents
// it addresses, followed by the rules that apply to them.
//
// A crawler obeys only the one group whose user agent matches it most
// specifically. Groups do not inherit from each other, so rules for a named bot
// have to repeat everything from the "*" group that should still apply to it.
type RobotsGroup struct {
	// UserAgents are the product tokens of the crawlers this group addresses,
	// for example "Googlebot" or "*" for all of them. Matching is
	// case-insensitive. No user agents at all renders as "User-agent: *".
	UserAgents []string

	// Allow are the paths that may be crawled. Because everything that is not
	// disallowed may be crawled anyway, an Allow rule only has an effect as
	// the exception to a Disallow rule of the same group. Each path must start
	// with "/" and may use the wildcards "*" (any sequence of characters) and
	// "$" (end of URL).
	Allow []string

	// Disallow are the paths that must not be crawled, with the same syntax as
	// Allow. The empty path is the idiomatic "Disallow:" line that disallows
	// nothing.
	Disallow []string

	// CrawlDelay is the minimum time a crawler should wait between two
	// requests. It is not part of RFC 9309 and is ignored by Googlebot, but
	// honored by Bing and Yandex. Zero omits the line.
	CrawlDelay time.Duration
}

// Validate returns an error if a user agent is not a valid product token, a
// path is not anchored at "/", or a value contains characters that would break
// the line based robots.txt format.
func (g RobotsGroup) Validate() error {
	for _, userAgent := range g.UserAgents {
		if userAgent != "*" && !robotsProductTokenRegexp.MatchString(userAgent) {
			return fmt.Errorf("invalid robots.txt user agent %q", userAgent)
		}
	}
	for _, path := range g.Allow {
		if err := validateRobotsPath(path); err != nil {
			return fmt.Errorf("Allow: %w", err)
		}
		if path == "" {
			return fmt.Errorf("empty Allow path (an empty path is only meaningful as Disallow)")
		}
	}
	for _, path := range g.Disallow {
		if err := validateRobotsPath(path); err != nil {
			return fmt.Errorf("Disallow: %w", err)
		}
	}
	if g.CrawlDelay < 0 {
		return fmt.Errorf("negative Crawl-delay %s", g.CrawlDelay)
	}
	return nil
}

// write renders the group's lines to b, assuming it has been validated.
func (g RobotsGroup) write(b *bytes.Buffer) {
	userAgents := g.UserAgents
	if len(userAgents) == 0 {
		userAgents = []string{"*"}
	}
	for _, userAgent := range userAgents {
		writeRobotsRule(b, "User-agent", userAgent)
	}
	for _, path := range g.Allow {
		writeRobotsRule(b, "Allow", path)
	}
	for _, path := range g.Disallow {
		writeRobotsRule(b, "Disallow", path)
	}
	if g.CrawlDelay > 0 {
		seconds := strconv.FormatFloat(g.CrawlDelay.Seconds(), 'f', -1, 64)
		writeRobotsRule(b, "Crawl-delay", seconds)
	}
}

// writeRobotsRule writes a "Name: value" line, omitting the space after the
// colon for an empty value so that a "Disallow:" line has no trailing space.
func writeRobotsRule(b *bytes.Buffer, name, value string) {
	if value == "" {
		b.WriteString(name + ":\n")
		return
	}
	b.WriteString(name + ": " + value + "\n")
}

// validateRobotsPath returns an error if path is neither empty nor a path
// anchored at "/", or if it contains a character that would change what the
// rule means once written to the file: whitespace and control characters end
// the line early or add a second rule, and "#" starts a comment.
//
// The "#" case is the dangerous one, because it widens a rule instead of
// breaking it: a parser reads "Disallow: /admin/#internal" as "Disallow:
// /admin/" — or an Allow that was meant for one fragment as an Allow for the
// whole subtree. Percent-encode it as %23 to use it in a path.
func validateRobotsPath(path string) error {
	if path != "" && !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path %q does not start with '/'", path)
	}
	if strings.ContainsFunc(path, func(r rune) bool { return r <= ' ' || r == 0x7F }) {
		return fmt.Errorf("path %q contains whitespace or control characters", path)
	}
	if strings.Contains(path, "#") {
		return fmt.Errorf("path %q contains '#', which starts a robots.txt comment (percent-encode it as %%23)", path)
	}
	return nil
}

// validateAbsoluteURL returns an error if rawURL is not an absolute http or
// https URL with a host, the form required for sitemap references in
// robots.txt and for the <loc> of a sitemap entry.
func validateAbsoluteURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	switch {
	case u.Scheme != "http" && u.Scheme != "https":
		return fmt.Errorf("URL %q is not an http or https URL", rawURL)
	// Hostname rather than Host: an authority like ":443" is a non-empty Host
	// with nothing to resolve, and would silently point crawlers at nothing.
	case u.Hostname() == "":
		return fmt.Errorf("URL %q has no host", rawURL)
	case strings.ContainsFunc(rawURL, func(r rune) bool { return r <= ' ' || r == 0x7F }):
		return fmt.Errorf("URL %q contains whitespace or control characters", rawURL)
	// A "#" would truncate a Sitemap: line at the fragment, and a sitemap
	// <loc> may not carry one either.
	case strings.Contains(rawURL, "#"):
		return fmt.Errorf("URL %q contains a fragment", rawURL)
	}
	return nil
}
