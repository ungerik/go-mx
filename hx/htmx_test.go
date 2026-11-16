package hx

import (
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

// TestScriptFromCDNIntegrity downloads every script the CDN elements load and
// checks that its sha384 digest is the one in the element's integrity
// attribute. A browser refuses to run a script whose bytes do not match that
// hash, so a wrong hash after a version bump would silently disable htmx, and
// with it every hx-* attribute, without anything offline noticing.
//
// The test needs network access, so it runs only when MX_CHECK_CDN is set.
// Run it whenever a version or hash in htmx.go changes:
//
//	MX_CHECK_CDN=1 go test ./hx -run TestScriptFromCDNIntegrity
func TestScriptFromCDNIntegrity(t *testing.T) {
	if os.Getenv("MX_CHECK_CDN") == "" {
		t.Skip("downloads from the CDN; set MX_CHECK_CDN=1 to run")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	for _, tc := range []struct {
		name   string
		script *mx.Element
	}{
		{"ScriptFromCDN", ScriptFromCDN},
		{"ScriptDebugFromCDN", ScriptDebugFromCDN},
		{"ScriptSSEFromCDN", ScriptSSEFromCDN},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := scriptAttrib(t, tc.script, "src")
			integrity := scriptAttrib(t, tc.script, "integrity")
			want, ok := strings.CutPrefix(integrity, "sha384-")
			if !ok || strings.ContainsAny(want, " \t\n") {
				t.Fatalf("integrity %q is not a single sha384 hash", integrity)
			}

			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, src, nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := client.Do(request)
			if err != nil {
				t.Fatalf("GET %s: %v", src, err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				t.Fatalf("GET %s: %s", src, response.Status)
			}
			// htmx is about 150 KB; the limit only keeps a misbehaving server
			// from streaming forever, and a cut-off body fails the comparison.
			hash := sha512.New384()
			_, err = io.Copy(hash, io.LimitReader(response.Body, 10<<20))
			if err != nil {
				t.Fatalf("read %s: %v", src, err)
			}

			got := base64.StdEncoding.EncodeToString(hash.Sum(nil))
			if got != want {
				t.Errorf("%s serves sha384-%s, but the integrity attribute says sha384-%s", src, got, want)
			}
		})
	}
}

// scriptAttrib returns the value of the attribute name of script and fails the
// test if script has no such attribute.
func scriptAttrib(t *testing.T, script *mx.Element, name string) string {
	t.Helper()
	i := script.AttribIndex(name)
	if i < 0 {
		t.Fatalf("script has no %s attribute", name)
	}
	return attribValue(t, script.Attribs[i])
}
