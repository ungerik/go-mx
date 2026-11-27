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
		"[--aurora-text-gradient:linear-gradient(135deg,#FF0080,#7928CA,#0070F3,#38bdf8,#FF0080)] [--aurora-text-duration:10s]",
		"bg-(image:--aurora-text-gradient)", "[-webkit-text-fill-color:transparent]",
	)
	if strings.Contains(out, "style=") {
		t.Errorf("the defaults need no style attribute: %s", out)
	}
	if n := strings.Count(out, "beautiful"); n != 2 {
		t.Errorf("text rendered %d times, want 2 (sr-only and gradient)", n)
	}
}

func TestAuroraTextOptions(t *testing.T) {
	out := render(t, AuroraText(AuroraTextOptions{Colors: []string{"red", "blue"}, Speed: 4},
		html.Class("inline"), "x"))
	mustContain(t, out,
		`data-slot="aurora-text" style="--aurora-text-gradient:linear-gradient(135deg, red, blue, red);--aurora-text-duration:2.5s"`,
		`class="relative [--aurora-text-gradient:`, `[--aurora-text-duration:10s] inline"`,
	)
	if err := renderErr(t, AuroraText(AuroraTextOptions{Colors: []string{"red;x:y"}}, "x")); err == nil {
		t.Error("an invalid color should be a render error")
	}
}

func TestAuroraTextCallerClassReplacesDefault(t *testing.T) {
	// The defaults sit on the root, where the caller's classes go, and the
	// gradient copy only reads them, so a caller class replaces a default
	// rather than being overridden by the copy's own class.
	out := render(t, AuroraText(AuroraTextOptions{}, html.Class("[--aurora-text-duration:5s]"), "x"))
	mustContain(t, out, `[--aurora-text-gradient:linear-gradient(135deg,#FF0080,#7928CA,#0070F3,#38bdf8,#FF0080)] [--aurora-text-duration:5s]"`)
	if n := strings.Count(out, "--aurora-text-duration:"); n != 1 {
		t.Errorf("--aurora-text-duration set %d times, want once on the root: %s", n, out)
	}
}

func TestAuroraTextChecksEachColor(t *testing.T) {
	// Balanced only together, so the joined gradient alone would pass.
	colors := []string{"red, blue), radial-gradient(green", "yellow)"}
	if renderErr(t, AuroraText(AuroraTextOptions{Colors: colors}, "x")) == nil {
		t.Error("a color that closes the gradient early was accepted")
	}
}
