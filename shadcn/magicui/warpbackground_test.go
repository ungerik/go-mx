package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestWarpBackgroundDefault(t *testing.T) {
	out := render(t, WarpBackground(WarpBackgroundOptions{}, html.Span("card")))
	mustContain(t, out, `data-slot="warp-background"`, `class="relative rounded border p-20"`,
		`style="--perspective:100px;--grid-color:var(--border);--beam-size:5%"`,
		"@container-[size] pointer-events-none absolute top-0 left-0 size-full overflow-hidden [clipPath:inset(0)] perspective-(--perspective) transform-3d",
		"transform-[rotate(90deg)_rotateX(-90deg)]", "transform-[rotate(-90deg)_rotateX(-90deg)]",
		"animate-warp-background-beam absolute top-0 left-(--x) aspect-[1/var(--aspect-ratio)] w-(--width) [background:var(--background)]",
		// Three beams per side on a grid of 20 cells: at cells 0, 6 and 13.
		"--x:0%;--width:5%", "--x:30%;--width:5%", "--x:65%;--width:5%",
		"translate:-50% -100%;animation-duration:3s",
		`<div class="relative"><span>card</span></div>`)
	if n := strings.Count(out, "animate-warp-background-beam"); n != 12 {
		t.Errorf("%d beams, want 12", n)
	}
}

func TestWarpBackgroundOptions(t *testing.T) {
	opts := WarpBackgroundOptions{Perspective: 200, BeamsPerSide: 2, BeamSize: 10, BeamDelayMin: time.Second,
		BeamDelayMax: time.Second, BeamDuration: 5 * time.Second, GridColor: "#333", Seed: 9}
	out := render(t, WarpBackground(opts, html.Class("p-10")))
	mustContain(t, out, "--perspective:200px;--grid-color:#333;--beam-size:10%",
		"--x:50%;--width:10%", "animation-duration:5s;animation-delay:1s", "relative rounded border p-10")
	if n := strings.Count(out, "animate-warp-background-beam"); n != 8 {
		t.Errorf("%d beams, want 8", n)
	}
	if render(t, WarpBackground(opts, html.Class("p-10"))) != out {
		t.Error("the same seed rendered different markup")
	}
	opts.Seed = 10
	if render(t, WarpBackground(opts, html.Class("p-10"))) == out {
		t.Error("a different seed rendered the same markup")
	}
	if renderErr(t, WarpBackground(WarpBackgroundOptions{GridColor: "red;x:y"})) == nil {
		t.Error("invalid grid color accepted")
	}
}
