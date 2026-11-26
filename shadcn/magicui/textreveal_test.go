package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestTextReveal(t *testing.T) {
	out := render(t, TextReveal(html.Class("h-[300vh]"), "Magic UI  will", " change"))
	mustContain(t, out, `data-slot="text-reveal"`, "[view-timeline-name:--text-reveal]", `style="--n:4"`,
		"sticky top-0 mx-auto flex h-[50%] max-w-4xl", "text-black/20",
		`<span class="absolute opacity-30" aria-hidden="true">Magic</span>`,
		`<span class="text-black dark:text-white animate-text-reveal-word" style="--i:3">change</span>`)
	if strings.Contains(out, "h-[200vh]") {
		t.Errorf("caller height should replace h-[200vh]: %s", out)
	}
	if n := strings.Count(out, "animate-text-reveal-word"); n != 4 {
		t.Errorf("%d animated words, want 4", n)
	}
}

func TestTextRevealNonText(t *testing.T) {
	if err := renderErr(t, TextReveal(html.Span("x"))); err == nil {
		t.Error("a non-text child rendered without error")
	}
}
