package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestGlareHoverDefault(t *testing.T) {
	out := render(t, GlareHover(GlareHoverOptions{}, html.Class("rounded-xl"), "card"))
	mustContain(t, out, `data-slot="glare-hover"`,
		"[--gh-angle:-45deg] [--gh-duration:650ms] [--gh-size:250%] [--gh-rgba:rgba(255,255,255,0.5)] bg-[#000]",
		"relative grid size-fit cursor-pointer place-items-center overflow-hidden",
		"before:[background-image:linear-gradient(var(--gh-angle),transparent_60%,var(--gh-rgba)_70%,transparent,transparent_100%)]",
		"before:transition-[background-position] before:duration-[var(--gh-duration)] before:ease-in-out",
		"hover:before:[background-position:100%_100%,0_0]", "rounded-xl", ">card</div>")
	if strings.Contains(out, "before:transition-none") {
		t.Errorf("transition should run both ways by default: %s", out)
	}
}

func TestGlareHoverOptions(t *testing.T) {
	out := render(t, GlareHover(GlareHoverOptions{
		Width: "100%", Height: "200px", Background: "transparent", Color: "#f60", Opacity: 0.25,
		Angle: 360, Size: 300, Duration: 550 * time.Millisecond, PlayOnce: true,
	}, html.Class("size-auto")))
	mustContain(t, out, "--gh-angle:360deg", "--gh-duration:550ms", "--gh-size:300%",
		"--gh-rgba:rgba(255,102,0,0.25)", "background:transparent;width:100%;height:200px",
		"before:transition-none hover:before:transition-[background-position] hover:before:duration-[var(--gh-duration)]",
		"size-auto")
	if strings.Contains(out, "size-fit") {
		t.Errorf("caller size should replace the default: %s", out)
	}
	out = render(t, GlareHover(GlareHoverOptions{Color: "oklch(0.7 0.1 200)"}))
	mustContain(t, out, "--gh-rgba:oklch(0.7 0.1 200)")
	// Opacity without Color applies to the default white glare.
	out = render(t, GlareHover(GlareHoverOptions{Opacity: 0.25}))
	mustContain(t, out, "--gh-rgba:rgba(255,255,255,0.25)")
}

func TestGlareHoverInvalid(t *testing.T) {
	if renderErr(t, GlareHover(GlareHoverOptions{Background: "red;x:y"})) == nil {
		t.Error("invalid background accepted")
	}
	if renderErr(t, GlareHover(GlareHoverOptions{Color: "red}"})) == nil {
		t.Error("invalid color accepted")
	}
}
