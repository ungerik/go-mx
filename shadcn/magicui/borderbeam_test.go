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
		`data-slot="border-beam"`, `aria-hidden="true"`, `style="--border-beam-width:1px"`,
		"border-(length:--border-beam-width)", "mask-intersect",
		`data-slot="border-beam-light"`, "animate-border-beam", "from-(--color-from) via-(--color-to)",
		"width:50px;offset-path:rect(0 auto auto 0 round 50px);--color-from:#ffaa40;--color-to:#9c40ff",
		"offset-distance:0%;--border-beam-from:0%;--border-beam-to:100%;--border-beam-duration:6s;--border-beam-delay:0s",
	)
}

func TestBorderBeamOptions(t *testing.T) {
	out := render(t, BorderBeam(BorderBeamOptions{
		Size: 400, Duration: 4 * time.Second, Delay: 3 * time.Second, Reverse: true, InitialOffset: 20,
		BorderWidth: 2, ColorFrom: "red", ColorTo: "blue",
	}, html.Class("from-transparent via-green-500 to-transparent")))
	mustContain(t, out,
		"--border-beam-width:2px", "width:400px", "round 400px", "--color-from:red;--color-to:blue",
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
