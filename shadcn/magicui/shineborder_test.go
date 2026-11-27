package magicui

import (
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestShineBorderDefault(t *testing.T) {
	out := render(t, ShineBorder(ShineBorderOptions{}))
	mustContain(t, out,
		`data-slot="shine-border"`, `aria-hidden="true"`,
		"--border-width:1px;--duration:14s;",
		"radial-gradient(transparent,transparent, #000000,transparent,transparent)",
		"mask-composite:exclude", "-webkit-mask-composite:xor", "padding:var(--border-width)",
		"motion-safe:animate-shine", "rounded-[inherit]",
	)
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
