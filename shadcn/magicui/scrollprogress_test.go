package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestScrollProgress(t *testing.T) {
	out := render(t, ScrollProgress())
	mustContain(t, out, `data-slot="scroll-progress"`, "fixed inset-x-0 top-0 z-50 h-px origin-left",
		"from-[#A97CF8] via-[#F38CB8] to-[#FDCC92]", "animate-scroll-progress")

	out = render(t, ScrollProgress(html.Class("top-[65px]")))
	mustContain(t, out, "top-[65px]")
	if strings.Contains(out, "top-0") {
		t.Errorf("caller top should replace top-0: %s", out)
	}
}
