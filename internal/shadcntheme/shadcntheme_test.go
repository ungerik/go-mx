package shadcntheme

import (
	"bytes"
	"strings"
	"testing"

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
		`<script src="` + TailwindBrowserURL + `"></script>`,
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
