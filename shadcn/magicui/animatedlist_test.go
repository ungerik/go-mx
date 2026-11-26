package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

func TestAnimatedList(t *testing.T) {
	items := mx.ForEach([]string{"b", "c"}, func(s string) mx.Component { return html.P(s) })
	out := render(t, AnimatedList(AnimatedListOptions{}, html.Class("gap-2 [--animated-list-gap:0.5rem]"), html.P("a"), items))
	mustContain(t, out, `data-slot="animated-list"`, "flex flex-col items-center", "gap-2", "[--animated-list-gap:0.5rem]",
		`style="--animated-list-delay:1s"`,
		`<div class="mx-auto grid w-full grid-rows-[1fr] origin-top animate-animated-list-item" style="--i:2"><div class="min-h-0"><p>c</p></div></div>`)
	if strings.Contains(out, "[--animated-list-gap:1rem]") {
		t.Errorf("caller gap variable should replace the default: %s", out)
	}
	// newest on top: c, b, a
	if c, b, a := strings.Index(out, "<p>c</p>"), strings.Index(out, "<p>b</p>"), strings.Index(out, "<p>a</p>"); !(c < b && b < a) {
		t.Errorf("items not in reverse order: %s", out)
	}

	out = render(t, AnimatedList(AnimatedListOptions{Delay: 1500 * time.Millisecond}, "x"))
	mustContain(t, out, "--animated-list-delay:1.5s", `style="--i:0"`)
}
