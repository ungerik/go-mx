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
		"shiny-button-label", "shiny-button-shine", `aria-hidden="true"`, "px-4", ">Shiny Button</span>",
	)
	if strings.Contains(out, "style=") {
		t.Errorf("ShinyButton needs no style attribute: %s", out)
	}
	for _, rule := range []string{"@utility shiny-button-label", "@utility shiny-button-shine", "calc(var(--shiny-button-x) + 20%)", "content-box exclude"} {
		if !strings.Contains(CSS, rule) {
			t.Errorf("CSS lacks %q", rule)
		}
	}
	if strings.Contains(out, "px-6") {
		t.Errorf("caller px-4 should replace px-6: %s", out)
	}
	if !strings.Contains(CSS, "@property --shiny-button-x") {
		t.Error("CSS should register --shiny-button-x")
	}
}
