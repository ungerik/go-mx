package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAnimatedGradientText(t *testing.T) {
	out := render(t, AnimatedGradientText(AnimatedGradientTextOptions{}, "Introducing"))
	mustContain(t, out,
		`data-slot="animated-gradient-text"`, "[--bg-size:300%] [--color-from:#ffaa40] [--color-to:#9c40ff]",
		"animate-gradient", "from-(--color-from)", "via-(--color-to)", "bg-size-[var(--bg-size)_100%]", "bg-clip-text",
	)
	if strings.Contains(out, "style=") {
		t.Errorf("the defaults need no style attribute: %s", out)
	}
	out = render(t, AnimatedGradientText(AnimatedGradientTextOptions{Speed: 2, ColorFrom: "#4ade80", ColorTo: "#06b6d4"},
		html.Class("block"), "Fast"))
	mustContain(t, out, "--bg-size:600%;--color-from:#4ade80;--color-to:#06b6d4", " block")
	if strings.Contains(out, " inline ") {
		t.Errorf("caller block should replace inline: %s", out)
	}
	if err := renderErr(t, AnimatedGradientText(AnimatedGradientTextOptions{ColorTo: "red;x:y"})); err == nil {
		t.Error("an invalid color should be a render error")
	}
}
