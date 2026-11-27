package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestShinyButton(t *testing.T) {
	out := render(t, ShinyButton(html.Class("px-4"), "Shiny Button"))
	mustContain(t, out,
		`data-slot="shiny-button"`, `type="button"`, "animate-shiny-button", "active:scale-95", "backdrop-blur-xl",
		"calc(var(--shiny-button-x) + 20%)", "color-mix(in oklab, var(--primary) 50%, transparent)",
		"content-box exclude", `aria-hidden="true"`, "px-4", ">Shiny Button</span>",
	)
	if strings.Contains(out, "px-6") {
		t.Errorf("caller px-4 should replace px-6: %s", out)
	}
	if !strings.Contains(CSS, "@property --shiny-button-x") {
		t.Error("CSS should register --shiny-button-x")
	}
}
