// Package snapshot turns the example servers under cmd/ into static pages for
// the documentation site, which GitHub Pages serves without any server code.
//
// An example registers the flags with [RegisterFlags] and, when
// [Options.Enabled], fetches its pages from its own handler with [Get] or
// [Post] — so a snapshot is the HTML the server would serve, not a copy of it
// — then edits what only works against a live server (an SSE connection, a
// link with a query string) with [Edit], and writes the result with
// [Options.Write], which puts a visible note on every page that it is a
// static snapshot and how to run the live example.
//
// The edits are exact string replacements that fail when their target is not
// found the expected number of times, so a change to an example's markup
// breaks the snapshot build instead of silently publishing a page that still
// tries to reach a server.
package snapshot

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// SourceURL is the GitHub tree URL the note on a snapshot links the example's
// source under: SourceURL + "cmd/example-sse".
const SourceURL = "https://github.com/ungerik/go-mx/tree/main/"

// Options are the snapshot flags of an example command.
type Options struct {
	// Dir is the directory the snapshot pages are written to. Empty means
	// the command runs its server as usual.
	Dir string
	// IndexHref is the URL of the page listing all examples, relative to
	// Dir. Empty leaves the link out of the note.
	IndexHref string
	// Command is the package directory of the example, such as
	// "cmd/example-sse", for the go run command and the source link the
	// note shows.
	Command string
}

// RegisterFlags registers -snapshot and -snapshot-index on fs (use
// flag.CommandLine for the program's flags) and returns the Options they
// fill once fs is parsed. command is the example's package directory, such
// as "cmd/example-sse".
func RegisterFlags(fs *flag.FlagSet, command string) *Options {
	o := &Options{Command: command}
	fs.StringVar(&o.Dir, "snapshot", "",
		"if set, write a static snapshot of the example's pages into this directory and exit, instead of serving")
	fs.StringVar(&o.IndexHref, "snapshot-index", "",
		"URL of the page listing all examples, relative to the -snapshot directory, linked from the snapshot note")
	return o
}

// Enabled reports whether a snapshot was asked for.
func (o *Options) Enabled() bool { return o.Dir != "" }

// Write inserts the note (see [Options.Note]) after the <body> start tag of
// page and writes it to the file name below Dir, creating directories as
// needed.
func (o *Options) Write(name, page string) error {
	note, err := o.Note()
	if err != nil {
		return err
	}
	page, err = InsertAfterStartTags(page, "<body", note, 1)
	if err != nil {
		return fmt.Errorf("snapshot %s: %w", name, err)
	}
	path := filepath.Join(o.Dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(page), 0o644)
}

// Note renders the banner a snapshot page starts with: that the page is a
// static snapshot, the go run command for the live version, a link to the
// source and, with IndexHref set, one to all examples. It is styled inline,
// because the examples bring different stylesheets or none.
func (o *Options) Note() (string, error) {
	if o.Command == "" {
		return "", fmt.Errorf("snapshot: Options.Command is empty")
	}
	codeStyle := html.Style("font-family:ui-monospace,SFMono-Regular,Menlo,monospace;background:#fef3c7;padding:0 .25em;border-radius:.25em")
	linkStyle := html.Style("color:#92400e;text-decoration:underline")
	note := html.Div(
		html.Role("note"),
		mx.ConstAttrib("data-mx-snapshot-note"),
		html.Style("font:14px/1.5 ui-sans-serif,system-ui,sans-serif;background:#fffbeb;color:#78350f;border-bottom:1px solid #fcd34d;padding:.5rem 1rem"),
		html.Strong("Static snapshot. "),
		"This page was rendered at build time by ",
		html.Code(codeStyle, o.Command),
		"; nothing on it reaches a server. For the live version run ",
		html.Code(codeStyle, "go run ./"+o.Command),
		" in a clone of the repository. ",
		html.AHRef(SourceURL+o.Command, linkStyle, "Source"),
		mx.If(o.IndexHref != "",
			mx.Components{
				mx.Text(" · "),
				html.AHRef(o.IndexHref, linkStyle, "All examples"),
			},
		),
	)
	return Render(note)
}

// Get returns the body of a GET request for target served by handler. A
// status other than 200 is an error.
func Get(handler http.Handler, target string) (string, error) {
	return serve(handler, httptest.NewRequest(http.MethodGet, target, nil))
}

// Post returns the body of a POST of form to target served by handler, sent
// as application/x-www-form-urlencoded. A status other than 200 is an error,
// so a form that redirects after a successful save cannot be snapshotted
// this way, which is what a page showing validation errors needs anyway.
func Post(handler http.Handler, target string, form url.Values) (string, error) {
	request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return serve(handler, request)
}

func serve(handler http.Handler, request *http.Request) (string, error) {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	response := recorder.Result()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("snapshot: %s %s: status %d: %s", request.Method, request.URL, response.StatusCode, bytes.TrimSpace(body))
	}
	return string(body), nil
}

// Edit replaces old with new in page, failing unless old occurs exactly n
// times, so a snapshot whose example changed its markup fails rather than
// publishing an unedited page.
func Edit(page, old, new string, n int) (string, error) {
	if count := strings.Count(page, old); count != n {
		return "", fmt.Errorf("snapshot: want %d occurrences of %q, found %d", n, old, count)
	}
	return strings.ReplaceAll(page, old, new), nil
}

// InsertAfterStartTags inserts content right after every start tag of page
// that begins with or contains marker, which may be the beginning of a tag
// such as "<body" or an attribute such as `data-mx-log-lines=""`. It fails
// unless marker occurs exactly n times or a tag it is in has no end.
func InsertAfterStartTags(page, marker, content string, n int) (string, error) {
	if count := strings.Count(page, marker); count != n {
		return "", fmt.Errorf("snapshot: want %d occurrences of %q, found %d", n, marker, count)
	}
	var b strings.Builder
	rest := page
	for range n {
		i := strings.Index(rest, marker)
		end := strings.IndexByte(rest[i:], '>')
		if end < 0 {
			return "", fmt.Errorf("snapshot: the tag of %q has no end", marker)
		}
		cut := i + end + 1
		b.WriteString(rest[:cut])
		b.WriteString(content)
		rest = rest[cut:]
	}
	b.WriteString(rest)
	return b.String(), nil
}

// Render renders comp to a string with a plain CheckedWriter, for content an
// example inserts into a snapshot page.
func Render(comp mx.Component) (string, error) {
	var buf bytes.Buffer
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&buf)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
