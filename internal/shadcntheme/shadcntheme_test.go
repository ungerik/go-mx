package shadcntheme

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
)

func TestHead(t *testing.T) {
	var buf bytes.Buffer
	if err := Head(".extra{color:red}").Render(t.Context(), mx.NewCheckedWriter(&buf)); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		InitScript,
		`<style type="text/tailwindcss">`,
		`@import "tailwindcss";`,
		".extra{color:red}",
		`src="` + TailwindBrowserURL + `"`,
		`integrity="` + TailwindBrowserIntegrity + `"`,
		`crossorigin="anonymous"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("head misses %q", want)
		}
	}
	// The init script runs before the style and the CDN script, so the dark
	// class is set before anything paints.
	if strings.Index(got, InitScript) > strings.Index(got, "text/tailwindcss") {
		t.Error("the init script does not come first")
	}
}

func TestTailwindInput(t *testing.T) {
	if got := TailwindInput(); got != CSS {
		t.Error("without extra CSS the input is not the theme alone")
	}
	if got := TailwindInput("a", "b"); got != CSS+"\na\nb" {
		t.Errorf("extra CSS not appended in order: %q", strings.TrimPrefix(got, CSS))
	}
}

// TestTailwindBrowserIntegrity downloads the script at TailwindBrowserURL and
// checks that its sha384 digest is TailwindBrowserIntegrity. A browser refuses
// to run a script whose bytes do not match the hash, so a wrong hash after a
// version bump would leave every demo page unstyled without anything offline
// noticing.
//
// The test needs network access, so it runs only when MX_CHECK_CDN is set.
func TestTailwindBrowserIntegrity(t *testing.T) {
	if os.Getenv("MX_CHECK_CDN") == "" {
		t.Skip("downloads from the CDN; set MX_CHECK_CDN=1 to run")
	}
	want, ok := strings.CutPrefix(TailwindBrowserIntegrity, "sha384-")
	if !ok {
		t.Fatalf("integrity %q is not a sha384 hash", TailwindBrowserIntegrity)
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, TailwindBrowserURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	if err != nil {
		t.Fatalf("GET %s: %v", TailwindBrowserURL, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: %s", TailwindBrowserURL, response.Status)
	}
	// The build is about 300 KB; the limit only keeps a misbehaving server
	// from streaming forever, and a cut-off body fails the comparison.
	hash := sha512.New384()
	if _, err := io.Copy(hash, io.LimitReader(response.Body, 10<<20)); err != nil {
		t.Fatalf("read %s: %v", TailwindBrowserURL, err)
	}
	if got := base64.StdEncoding.EncodeToString(hash.Sum(nil)); got != want {
		t.Errorf("%s serves sha384-%s, but TailwindBrowserIntegrity says sha384-%s", TailwindBrowserURL, got, want)
	}
}

func TestModeToggle(t *testing.T) {
	var buf bytes.Buffer
	if err := ModeToggle("theme-menu").Render(t.Context(), mx.NewCheckedWriter(&buf)); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		`popovertarget="theme-menu"`,
		`id="theme-menu"`,
		`aria-label="Theme"`,
		// The trigger shows the sun in light mode and the moon in dark mode.
		`lucide lucide-sun dark:hidden`,
		`lucide lucide-moon hidden dark:block`,
		"mxSetTheme('light')",
		"mxSetTheme('dark')",
		"mxSetTheme('system')",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("mode toggle misses %q", want)
		}
	}
	// The functions the menu calls are the ones InitScript defines.
	for _, fn := range []string{"window.mxSetTheme=", "window.mxApplyTheme="} {
		if !strings.Contains(InitScript, fn) {
			t.Errorf("InitScript does not define %s", fn)
		}
	}
}
