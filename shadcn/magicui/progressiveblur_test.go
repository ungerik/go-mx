package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestProgressiveBlurDefault(t *testing.T) {
	out := render(t, ProgressiveBlur(ProgressiveBlurOptions{}))
	mustContain(t, out, `data-slot="progressive-blur"`, `aria-hidden="true"`, `style="height:30%"`,
		"gradient-blur pointer-events-none absolute inset-x-0 z-10 bottom-0",
		"z-index:1;backdrop-filter:blur(0.5px);-webkit-backdrop-filter:blur(0.5px)",
		"mask-image:linear-gradient(to bottom, rgba(0,0,0,0) 0%, rgba(0,0,0,1) 12.5%, rgba(0,0,0,1) 25%, rgba(0,0,0,0) 37.5%)",
		"z-index:2;backdrop-filter:blur(1px)",
		"linear-gradient(to bottom, rgba(0,0,0,0) 12.5%, rgba(0,0,0,1) 25%, rgba(0,0,0,1) 37.5%, rgba(0,0,0,0) 50%)",
		"z-index:8;backdrop-filter:blur(64px)",
		"linear-gradient(to bottom, rgba(0,0,0,0) 87.5%, rgba(0,0,0,1) 100%)")
	if n := strings.Count(out, "backdrop-filter:blur(") / 2; n != 8 {
		t.Errorf("%d blur layers, want 8", n)
	}
}

func TestProgressiveBlurOptions(t *testing.T) {
	out := render(t, ProgressiveBlur(ProgressiveBlurOptions{Position: ProgressiveBlurTop, Height: "40%", BlurLevels: []float64{1, 2, 3}},
		html.Style("opacity:0.5")))
	mustContain(t, out, `style="height:40%;opacity:0.5"`, " top-0", "linear-gradient(to top, ",
		"z-index:3;backdrop-filter:blur(3px)")
	if n := strings.Count(out, "backdrop-filter:blur(") / 2; n != 3 {
		t.Errorf("%d blur layers, want 3", n)
	}

	out = render(t, ProgressiveBlur(ProgressiveBlurOptions{Position: ProgressiveBlurBoth, Height: "40%"}))
	mustContain(t, out, `style="height:100%"`, " inset-y-0",
		"mask-image:linear-gradient(rgba(0,0,0,0) 0%, rgba(0,0,0,1) 5%, rgba(0,0,0,1) 95%, rgba(0,0,0,0) 100%)")
}

func TestProgressiveBlurInvalid(t *testing.T) {
	if renderErr(t, ProgressiveBlur(ProgressiveBlurOptions{Height: "red;x:y"})) == nil {
		t.Error("invalid height accepted")
	}
	if renderErr(t, ProgressiveBlur(ProgressiveBlurOptions{BlurLevels: []float64{4}})) == nil {
		t.Error("a single blur level accepted")
	}
}
