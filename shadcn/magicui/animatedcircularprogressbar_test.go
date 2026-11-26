package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAnimatedCircularProgressBarDefault(t *testing.T) {
	out := render(t, AnimatedCircularProgressBar(AnimatedCircularProgressBarOptions{Value: 42}))
	mustContain(t, out,
		`data-slot="animated-circular-progress-bar"`, `class="relative size-40 text-2xl font-semibold"`,
		"--circle-size:100px;--circumference:282.7433388230814;--percent-to-px:2.827433388230814px;--gap-percent:5",
		"--percent-to-deg:3.6deg;transform:translateZ(0)",
		"stroke:var(--muted);--stroke-percent:48;", "stroke:var(--primary);--stroke-percent:42;",
		`data-current-value="42"`, ">42</span>", "animate-animated-circular-progress-bar-fade",
	)
}

func TestAnimatedCircularProgressBarRange(t *testing.T) {
	out := render(t, AnimatedCircularProgressBar(AnimatedCircularProgressBarOptions{
		Min: 10, Max: new(20.0), Value: 19.5, GaugePrimaryColor: "rgb(79 70 229)", GaugeSecondaryColor: "rgba(0, 0, 0, 0.1)",
	}, html.Class("size-20")))
	mustContain(t, out, `data-current-value="95"`, "stroke:rgb(79 70 229);--stroke-percent:95;", "size-20")
	// Above 90% the remaining arc would overlap the gaps, so upstream drops it.
	if strings.Contains(out, "rgba(0, 0, 0, 0.1)") {
		t.Error("secondary arc drawn above 90%")
	}
	if strings.Contains(out, "size-40") {
		t.Error("caller size should replace size-40")
	}
	// Math.round rounds halves up.
	out = render(t, AnimatedCircularProgressBar(AnimatedCircularProgressBarOptions{Value: 12.5}))
	mustContain(t, out, `data-current-value="13"`)
	if renderErr(t, AnimatedCircularProgressBar(AnimatedCircularProgressBarOptions{GaugePrimaryColor: "red;x:y"})) == nil {
		t.Error("invalid color accepted")
	}
	if renderErr(t, AnimatedCircularProgressBar(AnimatedCircularProgressBarOptions{Min: 100})) == nil {
		t.Error("empty range accepted")
	}
	// A range ending at 0 is possible because Max is a pointer.
	out = render(t, AnimatedCircularProgressBar(AnimatedCircularProgressBarOptions{Min: -100, Max: new(0.0), Value: -50}))
	mustContain(t, out, `data-current-value="50"`)
}
