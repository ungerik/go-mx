package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestRainbowButtonDefault(t *testing.T) {
	out := render(t, RainbowButton(RainbowButtonOptions{}, "Get Unlimited Access"))
	mustContain(t, out,
		`data-slot="rainbow-button"`, `type="button"`, "animate-rainbow", "before:animate-rainbow",
		"border-0", "text-primary-foreground", "h-9", "px-4", "rounded-sm", "[border:calc(0.125rem)_solid_transparent]",
	)
}

func TestRainbowButtonVariants(t *testing.T) {
	out := render(t, RainbowButton(RainbowButtonOptions{Variant: RainbowButtonOutline, Size: RainbowButtonSizeLG},
		html.Class("px-2"), "x"))
	mustContain(t, out, "border-b-transparent", "text-accent-foreground", "h-11", "rounded-xl", "px-2")
	if strings.Contains(out, "rounded-sm") || strings.Contains(out, "px-8") {
		t.Errorf("size and caller classes should replace rounded-sm and px-8: %s", out)
	}
	if got, want := RainbowButtonClasses("bogus", ""), RainbowButtonClasses(RainbowButtonDefault, RainbowButtonSizeDefault); got != want {
		t.Errorf("unknown variant should resolve to the default:\n%s\n%s", got, want)
	}
}
