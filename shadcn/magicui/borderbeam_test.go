package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestBorderBeamDefault(t *testing.T) {
	out := render(t, BorderBeam(BorderBeamOptions{}))
	mustContain(t, out,
		`data-slot="border-beam"`, `aria-hidden="true"`,
		"border-[length:var(--border-beam-width,1px)]", "mask-intersect",
		`data-slot="border-beam-light"`, "animate-border-beam", "from-(--color-from) via-(--color-to)",
		"[--border-beam-size:50px] w-(--border-beam-size) [offset-path:rect(0_auto_auto_0_round_var(--border-beam-size))] [--color-from:#ffaa40] [--color-to:#9c40ff] [offset-distance:0%]",
		"[--border-beam-from:0%] [--border-beam-to:100%] [--border-beam-duration:6s] [--border-beam-delay:0s]",
	)
	if strings.Contains(out, "style=") {
		t.Errorf("the defaults need no style attribute: %s", out)
	}
	// The caller's classes go to the beam, so the outer border sets no
	// --border-beam-width of its own: one set on the BorderBeam's parent
	// reaches it.
	if strings.Contains(out, "[--border-beam-width:") {
		t.Errorf("the border should inherit --border-beam-width, not set it: %s", out)
	}
}

func TestBorderBeamSizeClass(t *testing.T) {
	// Width and path both read --border-beam-size, so a class setting it
	// resizes both together.
	out := render(t, BorderBeam(BorderBeamOptions{}, html.Class("[--border-beam-size:80px]")))
	mustContain(t, out, "w-(--border-beam-size)", "round_var(--border-beam-size)", "[--border-beam-size:80px]")
	if strings.Contains(out, "[--border-beam-size:50px]") {
		t.Errorf("caller size should replace the default: %s", out)
	}
}

func TestBorderBeamOptions(t *testing.T) {
	out := render(t, BorderBeam(BorderBeamOptions{
		Size: 400, Duration: 4 * time.Second, Delay: 3 * time.Second, Reverse: true, InitialOffset: 20,
		BorderWidth: 2, ColorFrom: "red", ColorTo: "blue",
	}, html.Class("from-transparent via-green-500 to-transparent")))
	mustContain(t, out,
		"--border-beam-width:2px", "--border-beam-size:400px", "--color-from:red;--color-to:blue",
		"offset-distance:20%;--border-beam-from:80%;--border-beam-to:-20%;--border-beam-duration:4s;--border-beam-delay:-3s",
		"from-transparent via-green-500 to-transparent",
	)
	for _, replaced := range []string{"from-(--color-from)", "via-(--color-to)"} {
		if strings.Contains(out, replaced) {
			t.Errorf("caller gradient should replace %s", replaced)
		}
	}
	if renderErr(t, BorderBeam(BorderBeamOptions{ColorFrom: "red;x:y"})) == nil {
		t.Error("invalid color accepted")
	}
}
