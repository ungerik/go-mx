package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

func TestOrbitingCirclesDefault(t *testing.T) {
	out := render(t, OrbitingCircles(OrbitingCirclesOptions{}, html.Span("a"), html.Span("b"), html.Span("c")))
	mustContain(t, out,
		`data-slot="orbiting-circles" class="contents"`,
		`<circle class="stroke-black/10 stroke-1 dark:stroke-white/10" cx="50%" cy="50%" r="160" fill="none">`,
		`data-slot="orbiting-circles-item"`, "animate-orbit absolute flex size-(--icon-size)",
		"--duration:20;--orbiting-circles-radius:160;--angle:0;--icon-size:30px",
		"--angle:120;", "--angle:240;",
	)
	if n := strings.Count(out, `data-slot="orbiting-circles-item"`); n != 3 {
		t.Errorf("%d items, want 3", n)
	}
	if strings.Contains(out, "[animation-direction:reverse]") {
		t.Error("reverse without opts.Reverse")
	}
}

func TestOrbitingCirclesNegativeDurationAndSpeed(t *testing.T) {
	// Negative values mean the defaults, like 0, not a negative --duration
	// that makes the animation shorthand invalid.
	out := render(t, OrbitingCircles(OrbitingCirclesOptions{Duration: -time.Second, Speed: -2}, html.Span("a")))
	mustContain(t, out, "--duration:20;")
}

func TestOrbitingCirclesOptions(t *testing.T) {
	out := render(t, OrbitingCircles(
		OrbitingCirclesOptions{Reverse: true, Duration: 30 * time.Second, Speed: 2, Radius: 100, IconSize: 40, HidePath: true},
		html.Class("size-12 border"), html.Style("color:red"), "x", "y",
	))
	mustContain(t, out,
		"[animation-direction:reverse]", "size-12 border",
		`style="--duration:15;--orbiting-circles-radius:100;--angle:0;--icon-size:40px;color:red"`, "--angle:180;",
	)
	if strings.Contains(out, "<svg") {
		t.Error("HidePath still draws the orbit")
	}
	if strings.Contains(out, "size-(--icon-size)") {
		t.Error("caller size class should replace size-(--icon-size)")
	}
}

func TestOrbitingCirclesFlattensComponents(t *testing.T) {
	items := mx.ForEach([]string{"a", "b", "c", "d"}, func(s string) mx.Component { return html.Span(s) })
	out := render(t, OrbitingCircles(OrbitingCirclesOptions{}, items))
	if n := strings.Count(out, `data-slot="orbiting-circles-item"`); n != 4 {
		t.Errorf("got %d orbiting items, want 4", n)
	}
	mustContain(t, out, "--angle:0;", "--angle:90;", "--angle:180;", "--angle:270;")
}
