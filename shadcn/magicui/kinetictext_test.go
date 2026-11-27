package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestKineticText(t *testing.T) {
	out := render(t, KineticText(KineticTextOptions{Text: "Hi yo"}, html.Class("font-[200]")))
	mustContain(t, out,
		`<h1 `, `data-slot="kinetic-text"`, "flex flex-wrap [--hover-padding:calc(1em/12)] [--text-stroke-width:calc(1em*125/6000)] font-[200]",
		"hover:font-[900]", "[:hover+span+&amp;]:font-[400]", " ", `<span class="sr-only">Hi yo</span>`,
	)
	if strings.Contains(out, "style=") {
		t.Errorf("KineticText needs no style attribute: %s", out)
	}
	if strings.Contains(out, "font-[300]") {
		t.Errorf("caller font-[200] should replace font-[300]: %s", out)
	}
	if n := strings.Count(out, `aria-hidden="true"`); n != 5 {
		t.Errorf("%d aria-hidden letters, want 5", n)
	}
}

func TestKineticTextOptions(t *testing.T) {
	out := render(t, KineticText(KineticTextOptions{Text: "ä", As: KineticTextP}, html.Style("--hover-padding: 2px")))
	mustContain(t, out, "<p ", ">ä</span>", `style="--hover-padding: 2px"`)
	out = render(t, KineticText(KineticTextOptions{Text: "x", As: "bogus"}))
	mustContain(t, out, "<h1 ")
}
