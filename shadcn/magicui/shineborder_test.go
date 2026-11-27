package magicui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestShineBorderDefault(t *testing.T) {
	out := render(t, ShineBorder(ShineBorderOptions{}))
	mustContain(t, out,
		`data-slot="shine-border"`, `aria-hidden="true"`,
		"[--border-width:1px] [--duration:14s]",
		"bg-[radial-gradient(transparent,transparent,#000000,transparent,transparent)]",
		"shine-border-mask",
		"motion-safe:animate-shine", "rounded-[inherit]",
	)
	if strings.Contains(out, "style=") {
		t.Errorf("the defaults need no style attribute: %s", out)
	}
}

func TestShineBorderPaddingInUtility(t *testing.T) {
	// The ring's width is the padding the mask cuts out, so it lives in the
	// shine-border-mask utility where a caller's p-* class cannot drop it.
	if !regexp.MustCompile(`@utility shine-border-mask \{[^}]*padding: var\(--border-width\);`).MatchString(CSS) {
		t.Error("shine-border-mask should set padding: var(--border-width)")
	}
	out := render(t, ShineBorder(ShineBorderOptions{BorderWidth: 3}, html.Class("p-2")))
	mustContain(t, out, "shine-border-mask", "--border-width:3px")
}

func TestShineBorderOptions(t *testing.T) {
	out := render(t, ShineBorder(ShineBorderOptions{
		BorderWidth: 2.5, Duration: 8 * time.Second, ShineColors: []string{"#A07CFE", "#FE8FB5", "#FFBE7B"},
	}, html.Class("rounded-xl"), html.Style("opacity: 0.5")))
	mustContain(t, out, "--border-width:2.5px", "--duration:8s", "#A07CFE,#FE8FB5,#FFBE7B", "rounded-xl", ";opacity: 0.5")
	if err := renderErr(t, ShineBorder(ShineBorderOptions{ShineColors: []string{"red", "blue;x:y"}})); err == nil {
		t.Error("an invalid shine color should be a render error")
	}
}
