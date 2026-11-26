package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAuroraTextDefault(t *testing.T) {
	out := render(t, AuroraText(AuroraTextOptions{}, "beautiful"))
	mustContain(t, out,
		`data-slot="aurora-text"`, "relative inline-block", `class="sr-only"`, "animate-aurora",
		"linear-gradient(135deg, #FF0080, #7928CA, #0070F3, #38bdf8, #FF0080)", "animation-duration:10s",
		"-webkit-text-fill-color:transparent",
	)
	if n := strings.Count(out, "beautiful"); n != 2 {
		t.Errorf("text rendered %d times, want 2 (sr-only and gradient)", n)
	}
}

func TestAuroraTextOptions(t *testing.T) {
	out := render(t, AuroraText(AuroraTextOptions{Colors: []string{"red", "blue"}, Speed: 4},
		html.Class("inline"), "x"))
	mustContain(t, out, "linear-gradient(135deg, red, blue, red)", "animation-duration:2.5s", `class="relative inline"`)
	if err := renderErr(t, AuroraText(AuroraTextOptions{Colors: []string{"red;x:y"}}, "x")); err == nil {
		t.Error("an invalid color should be a render error")
	}
}
