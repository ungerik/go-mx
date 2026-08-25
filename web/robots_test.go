package web_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/web"
)

func TestAllowAllRobots(t *testing.T) {
	// "Disallow:" with an empty path is the idiomatic way to allow everything,
	// so it has to render without a trailing space that would make the line
	// harder to compare against what crawlers and SEO tools expect.
	require.Equal(t, "User-agent: *\nDisallow:\n", web.AllowAllRobots().String())

	require.Equal(t,
		"User-agent: *\nDisallow:\n\nSitemap: https://example.com/sitemap.xml\n",
		web.AllowAllRobots("https://example.com/sitemap.xml").String(),
	)
}

func TestDisallowAllRobots(t *testing.T) {
	require.Equal(t, "User-agent: *\nDisallow: /\n", web.DisallowAllRobots().String())
}

func TestRobotsMarshalText(t *testing.T) {
	robots := &web.Robots{
		Groups: []web.RobotsGroup{
			{
				Disallow: []string{"/admin/", "/*.json$"},
				Allow:    []string{"/admin/public/"},
			},
			{
				UserAgents: []string{"Googlebot", "Bingbot"},
				Disallow:   []string{"/private/"},
				CrawlDelay: 10 * time.Second,
			},
		},
		Sitemaps: []string{"https://example.com/sitemap.xml"},
	}
	text, err := robots.MarshalText()
	require.NoError(t, err)

	// Groups are separated by a blank line because a crawler reads the lines
	// following a User-agent as one group: without the separator the rules of
	// the second group would be read as part of the first.
	// Allow comes before Disallow so the exception is next to the rule it
	// modifies, and the sitemaps come last because they are not part of any
	// group.
	require.Equal(t,
		"User-agent: *\n"+
			"Allow: /admin/public/\n"+
			"Disallow: /admin/\n"+
			"Disallow: /*.json$\n"+
			"\n"+
			"User-agent: Googlebot\n"+
			"User-agent: Bingbot\n"+
			"Disallow: /private/\n"+
			"Crawl-delay: 10\n"+
			"\n"+
			"Sitemap: https://example.com/sitemap.xml\n",
		string(text),
	)
}

func TestRobotsValidate(t *testing.T) {
	tests := []struct {
		name   string
		robots *web.Robots
		valid  bool
	}{
		{
			name:   "empty",
			robots: &web.Robots{},
			valid:  true,
		},
		{
			name:   "wildcard user agent",
			robots: &web.Robots{Groups: []web.RobotsGroup{{UserAgents: []string{"*"}, Disallow: []string{"/"}}}},
			valid:  true,
		},
		{
			name:   "product token user agent",
			robots: &web.Robots{Groups: []web.RobotsGroup{{UserAgents: []string{"Mediapartners-Google"}}}},
			valid:  true,
		},
		{
			// A user agent with a space would be read as a product token
			// followed by junk, silently addressing a different crawler than
			// the one that was meant.
			name:   "user agent with space",
			robots: &web.Robots{Groups: []web.RobotsGroup{{UserAgents: []string{"Googlebot Image"}}}},
			valid:  false,
		},
		{
			// A newline in a value turns one rule into two, so the file would
			// mean something the Robots value does not say.
			name:   "newline in path",
			robots: &web.Robots{Groups: []web.RobotsGroup{{Disallow: []string{"/a\nDisallow: /"}}}},
			valid:  false,
		},
		{
			// A path that is not anchored at "/" is ignored by crawlers, so
			// the rule would look present but have no effect.
			name:   "unanchored path",
			robots: &web.Robots{Groups: []web.RobotsGroup{{Disallow: []string{"admin/"}}}},
			valid:  false,
		},
		{
			name:   "empty Disallow allows everything",
			robots: &web.Robots{Groups: []web.RobotsGroup{{Disallow: []string{""}}}},
			valid:  true,
		},
		{
			// An empty Allow is not the mirror image of an empty Disallow, it
			// is a rule without a path that crawlers drop.
			name:   "empty Allow",
			robots: &web.Robots{Groups: []web.RobotsGroup{{Allow: []string{""}}}},
			valid:  false,
		},
		{
			name:   "negative crawl delay",
			robots: &web.Robots{Groups: []web.RobotsGroup{{CrawlDelay: -time.Second}}},
			valid:  false,
		},
		{
			// A sitemap is fetched by a crawler that has no page context, so a
			// relative reference cannot be resolved.
			name:   "relative sitemap URL",
			robots: &web.Robots{Sitemaps: []string{"/sitemap.xml"}},
			valid:  false,
		},
		{
			name:   "absolute sitemap URL",
			robots: &web.Robots{Sitemaps: []string{"https://example.com/sitemap.xml"}},
			valid:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.robots.Validate()
			if tt.valid {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			// An invalid value must never reach the client or a file: both
			// output paths report the error instead of writing the file.
			_, err = tt.robots.MarshalText()
			require.Error(t, err)
			_, err = tt.robots.WriteTo(io.Discard)
			require.Error(t, err)
		})
	}
}

func TestRobotsHandleHTTP(t *testing.T) {
	response := httptest.NewRecorder()
	web.AllowAllRobots().HandleHTTP(response, httptest.NewRequest(http.MethodGet, web.RobotsTxtPath, nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, mx.ContentTypePlainText, response.Header().Get("Content-Type"))
	require.Equal(t, "User-agent: *\nDisallow:\n", response.Body.String())
}

func TestRobotsWriteTo(t *testing.T) {
	var b strings.Builder
	n, err := web.AllowAllRobots().WriteTo(&b)
	require.NoError(t, err)
	// WriteTo reports the byte count it wrote, the io.WriterTo contract a
	// caller writing the file to disk relies on.
	require.Equal(t, "User-agent: *\nDisallow:\n", b.String())
	require.Equal(t, int64(b.Len()), n)
}

func TestRobotsStringError(t *testing.T) {
	// String cannot return an error, so an invalid value has to be visible in
	// the string itself rather than passing for a valid robots.txt.
	robots := &web.Robots{Sitemaps: []string{"/sitemap.xml"}}
	require.Contains(t, robots.String(), "!ERROR: ")
}

func TestRobotsHandleHTTPInvalid(t *testing.T) {
	// Serving a broken robots.txt is worse than serving none: crawlers would
	// act on rules the site never described.
	robots := &web.Robots{Groups: []web.RobotsGroup{{Disallow: []string{"admin/"}}}}
	response := httptest.NewRecorder()
	robots.HandleHTTP(response, httptest.NewRequest(http.MethodGet, web.RobotsTxtPath, nil))

	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestRobotsFractionalCrawlDelay(t *testing.T) {
	// A sub-second delay must not round to "0", which would tell a crawler it
	// may hammer the site.
	robots := &web.Robots{Groups: []web.RobotsGroup{{CrawlDelay: 500 * time.Millisecond}}}
	require.Equal(t, "User-agent: *\nCrawl-delay: 0.5\n", robots.String())
}

func TestRobotsRejectsCommentCharacter(t *testing.T) {
	// Regression: a "#" starts a robots.txt comment, so a rule carrying one was
	// silently truncated at it. "Disallow: /admin/#internal" would reach
	// crawlers as "Disallow: /admin/", and the Allow variant would open a whole
	// subtree that was meant to stay closed.
	robots := &web.Robots{Groups: []web.RobotsGroup{{Allow: []string{"/admin/#health"}}}}
	require.ErrorContains(t, robots.Validate(), "#")

	robots = &web.Robots{Groups: []web.RobotsGroup{{Disallow: []string{"/admin/#internal"}}}}
	require.ErrorContains(t, robots.Validate(), "#")

	// A fragment in a Sitemap: URL is truncated the same way.
	robots = &web.Robots{Sitemaps: []string{"https://example.com/sitemap.xml#gzip"}}
	require.Error(t, robots.Validate())

	// The percent-encoded form is a normal path character.
	robots = &web.Robots{Groups: []web.RobotsGroup{{Disallow: []string{"/admin/%23internal"}}}}
	require.NoError(t, robots.Validate())
}

func TestRobotsRejectsHostlessURL(t *testing.T) {
	// Regression: url.URL.Host is non-empty for an authority that is only a
	// port, so ":443" passed the host check and produced a Sitemap: line
	// pointing at nothing.
	robots := &web.Robots{Sitemaps: []string{"https://:443/sitemap.xml"}}
	require.ErrorContains(t, robots.Validate(), "host")
}
