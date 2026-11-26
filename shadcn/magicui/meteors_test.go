package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestMeteorsDefault(t *testing.T) {
	out := render(t, Meteors(MeteorsOptions{}))
	mustContain(t, out, `data-slot="meteors"`, `aria-hidden="true"`, "pointer-events-none absolute inset-0",
		"animate-meteor pointer-events-none absolute size-0.5 rotate-(--angle) rounded-full bg-zinc-500 shadow-[0_0_0_1px_#ffffff10]",
		"--angle:-215deg;top:-5%;left:",
		"pointer-events-none absolute top-1/2 -z-10 h-px w-12.5 -translate-y-1/2 bg-linear-to-r from-zinc-500 to-transparent")
	if n := strings.Count(out, "animate-meteor"); n != 20 {
		t.Errorf("%d meteors, want 20", n)
	}
}

func TestMeteorsOptions(t *testing.T) {
	opts := MeteorsOptions{Number: 5, MinDelay: time.Second, MaxDelay: time.Second,
		MinDuration: 3 * time.Second, MaxDuration: 3 * time.Second, Angle: 180, Seed: 7}
	out := render(t, Meteors(opts, html.Class("[&>span]:bg-sky-500")))
	mustContain(t, out, "--angle:-180deg", "animation-delay:1s;animation-duration:3s", "[&amp;>span]:bg-sky-500")
	if n := strings.Count(out, "animate-meteor"); n != 5 {
		t.Errorf("%d meteors, want 5", n)
	}
	if render(t, Meteors(opts, html.Class("[&>span]:bg-sky-500"))) != out {
		t.Error("the same seed rendered different markup")
	}
	opts.Seed = 8
	other := render(t, Meteors(opts, html.Class("[&>span]:bg-sky-500")))
	if other == out {
		t.Error("a different seed rendered the same markup")
	}
}

func TestMeteorsNegativeAngle(t *testing.T) {
	// Upstream rotates by ${-angle}deg, so -30 turns into 30deg.
	out := render(t, Meteors(MeteorsOptions{Angle: -30, Number: 1}))
	mustContain(t, out, "--angle:30deg;")
	if strings.Contains(out, "--30deg") {
		t.Error("negative angle rendered with a double minus")
	}
}
