package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestNoiseTextureDefault(t *testing.T) {
	out := render(t, NoiseTexture(NoiseTextureOptions{ID: "noise"}))
	mustContain(t, out, `data-slot="noise-texture"`, `aria-hidden="true"`,
		"pointer-events-none absolute inset-0 z-0 size-full opacity-50 select-none dark:opacity-[0.75]",
		`<filter id="noise">`,
		`<feTurbulence type="fractalNoise" baseFrequency="0.4" numOctaves="6" stitchTiles="stitch">`,
		`<feColorMatrix type="saturate" values="0">`, `<feFuncR type="linear" slope="0.15">`,
		`<feFuncB type="linear" slope="0.15">`, `filter="url(#noise)" opacity="0.6"`)
}

func TestNoiseTextureOptions(t *testing.T) {
	out := render(t, NoiseTexture(NoiseTextureOptions{ID: "n", Frequency: 0.8, Octaves: 3, Slope: 0.3, NoiseOpacity: 0.45},
		html.Class("opacity-100")))
	mustContain(t, out, `baseFrequency="0.8" numOctaves="3"`, `slope="0.3"`, `opacity="0.45"`, "opacity-100")
	if strings.Contains(out, "opacity-50") {
		t.Errorf("caller opacity should replace the default: %s", out)
	}
	if renderErr(t, NoiseTexture(NoiseTextureOptions{ID: "n)"})) == nil {
		t.Error("invalid id accepted")
	}
}
