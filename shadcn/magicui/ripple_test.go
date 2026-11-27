package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestRippleDefault(t *testing.T) {
	out := render(t, Ripple(RippleOptions{}))
	mustContain(t, out, `data-slot="ripple"`,
		"pointer-events-none absolute inset-0 mask-[linear-gradient(to_bottom,white,transparent)] select-none",
		"animate-ripple bg-foreground/25 absolute rounded-full border shadow-xl",
		"--i:0;width:210px;height:210px;opacity:0.24;animation-delay:0s",
		"--i:1;width:280px;height:280px;opacity:0.21;animation-delay:0.06s",
		"--i:7;width:700px;height:700px;opacity:0.03;animation-delay:0.42s",
		"transform:translate(-50%, -50%) scale(1)")
	if n := strings.Count(out, "animate-ripple"); n != 8 {
		t.Errorf("%d circles, want 8", n)
	}
}

func TestRippleOptions(t *testing.T) {
	out := render(t, Ripple(RippleOptions{MainCircleSize: 100, MainCircleOpacity: 0.5, NumCircles: 2},
		html.Class("[--ripple-duration:3s]"), html.P("Ripple")))
	mustContain(t, out, "width:100px", "opacity:0.5;", "width:170px", "opacity:0.47;", "[--ripple-duration:3s]",
		// Children come before the aria-hidden circles.
		`select-none [--ripple-duration:3s]"><p>Ripple</p><div aria-hidden="true"`)
	if n := strings.Count(out, "animate-ripple"); n != 2 {
		t.Errorf("%d circles, want 2", n)
	}
}
