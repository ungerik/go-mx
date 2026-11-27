package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAnimatedShinyText(t *testing.T) {
	out := render(t, AnimatedShinyText(AnimatedShinyTextOptions{}, "Introducing Magic UI"))
	mustContain(t, out,
		`data-slot="animated-shiny-text"`, "[--shiny-width:100px]", "animate-shiny-text",
		"bg-size-[var(--shiny-width)_100%]", "via-black/80", "via-50%", "dark:via-white/80", "max-w-md",
	)
	if strings.Contains(out, "style=") {
		t.Errorf("the defaults need no style attribute: %s", out)
	}
	out = render(t, AnimatedShinyText(AnimatedShinyTextOptions{ShimmerWidth: 40}, html.Class("max-w-none"), "x"))
	mustContain(t, out, `style="--shiny-width:40px"`, "max-w-none")
	if strings.Contains(out, "max-w-md") {
		t.Errorf("caller max-w-none should replace max-w-md: %s", out)
	}
}
