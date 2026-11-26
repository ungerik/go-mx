package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestLightRaysDefault(t *testing.T) {
	out := render(t, LightRays(LightRaysOptions{}))
	mustContain(t, out, `data-slot="light-rays"`,
		`style="--light-rays-color:rgba(160, 210, 255, 0.2);--light-rays-blur:36px;--light-rays-length:70vh"`,
		"pointer-events-none absolute inset-0 isolate overflow-hidden rounded-[inherit]",
		`<div aria-hidden="true" class="absolute inset-0 overflow-hidden">`,
		"radial-gradient(circle at 20% 15%, color-mix(in srgb, var(--light-rays-color) 45%, transparent), transparent 70%)",
		"animate-light-rays-ray pointer-events-none absolute -top-[12%] left-[var(--ray-left)]",
		"blur-[var(--light-rays-blur)]", "--ray-rotate:", "--ray-swing:", "--ray-intensity:",
		"animation-duration:", "animation-delay:")
	if n := strings.Count(out, "animate-light-rays-ray"); n != 7 {
		t.Errorf("%d rays, want 7", n)
	}
	// The resting state is visible, so reduced motion keeps the rays.
	if strings.Contains(out, "opacity-0") {
		t.Errorf("rays should rest at their intensity: %s", out)
	}
}

func TestLightRaysOptions(t *testing.T) {
	opts := LightRaysOptions{Count: 3, Color: "#ffcc00", Blur: 10, Speed: 10 * time.Second, Length: "50vh", Seed: 3}
	out := render(t, LightRays(opts, html.Class("rounded-xl")))
	mustContain(t, out, "--light-rays-color:#ffcc00;--light-rays-blur:10px;--light-rays-length:50vh", "rounded-xl")
	if n := strings.Count(out, "animate-light-rays-ray"); n != 3 {
		t.Errorf("%d rays, want 3", n)
	}
	if render(t, LightRays(opts, html.Class("rounded-xl"))) != out {
		t.Error("the same seed rendered different markup")
	}
	opts.Seed = 4
	if render(t, LightRays(opts, html.Class("rounded-xl"))) == out {
		t.Error("a different seed rendered the same markup")
	}
	if n := strings.Count(render(t, LightRays(LightRaysOptions{Count: -1})), "animate-light-rays-ray"); n != 0 {
		t.Errorf("%d rays for a negative count", n)
	}
	if renderErr(t, LightRays(LightRaysOptions{Color: "red;x:y"})) == nil {
		t.Error("invalid color accepted")
	}
}
