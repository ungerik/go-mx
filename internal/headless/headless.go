// Package headless runs a page in headless Chrome and returns its DOM after
// the page's scripts ran, for tests of go-mx's inline scripts. It drives the
// browser binary through its command line (--dump-dom), so it adds no Go
// dependency.
//
// A test page runs its own checks and writes the outcome into the DOM, which
// the test reads back with [Result].
package headless

import (
	"bytes"
	"context"
	"errors"
	"html"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"testing"
	"time"
)

// candidates are the Chrome or Chromium binaries looked for, in order, when
// the MX_CHROME environment variable names none.
var candidates = []string{
	"google-chrome",
	"google-chrome-stable",
	"chromium",
	"chromium-browser",
	"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	"/Applications/Chromium.app/Contents/MacOS/Chromium",
}

// Find returns the path of a Chrome or Chromium binary: the MX_CHROME
// environment variable if set, else the first of the usual names found.
func Find() (string, error) {
	if path := os.Getenv("MX_CHROME"); path != "" {
		return path, nil
	}
	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			return path, nil
		}
	}
	return "", errors.New("no Chrome or Chromium found; set MX_CHROME")
}

// Chrome returns the browser binary for t, skipping t when there is none,
// unless the MX_REQUIRE_CHROME environment variable is set, as in CI, which
// fails it instead.
func Chrome(t testing.TB) string {
	t.Helper()
	path, err := Find()
	if err != nil {
		if os.Getenv("MX_REQUIRE_CHROME") != "" {
			t.Fatal(err)
		}
		t.Skip(err)
	}
	return path
}

// DumpDOM loads url in headless Chrome and returns the serialized DOM once
// the page loaded and its timers had virtual time to run: budget is the
// virtual time given to setTimeout and pending fetches, so a test waits for
// a two-second timer without waiting two seconds.
func DumpDOM(ctx context.Context, chrome, url string, budget time.Duration) (string, error) {
	dir, err := os.MkdirTemp("", "mx-headless-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	cmd := exec.CommandContext(ctx, chrome,
		"--headless=new",
		"--disable-gpu",
		// Runners whose kernel restricts user namespaces cannot start the
		// sandbox; the pages are the tests' own.
		"--no-sandbox",
		"--no-first-run",
		"--no-default-browser-check",
		"--user-data-dir="+dir,
		"--virtual-time-budget="+strconv.FormatInt(budget.Milliseconds(), 10),
		"--dump-dom",
		url,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", errors.Join(err, errors.New(stderr.String()))
	}
	return stdout.String(), nil
}

// resultPattern finds the element a test page writes its outcome into.
var resultPattern = regexp.MustCompile(`(?s)<pre id="result">(.*?)</pre>`)

// Result returns the unescaped text of the page's <pre id="result">, where
// a test page writes its outcome, and whether there is one.
func Result(dom string) (string, bool) {
	m := resultPattern.FindStringSubmatch(dom)
	if m == nil {
		return "", false
	}
	return html.UnescapeString(m[1]), true
}
