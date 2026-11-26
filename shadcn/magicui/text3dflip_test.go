package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestText3DFlipDefault(t *testing.T) {
	out := render(t, Text3DFlip(Text3DFlipOptions{}, "ab c"))
	mustContain(t, out, `<p data-slot="text-3d-flip"`, `class="group/text-3d-flip relative flex flex-wrap"`,
		"--text-3d-flip-rest:rotateY(90deg) translateX(50%) rotateY(-90deg);--text-3d-flip-turn:rotateY(90deg);--text-3d-flip-duration:0.46s;--text-3d-flip-easing:linear(",
		`<span class="sr-only">ab c</span>`,
		`<span class="inline-flex" aria-hidden="true">`,
		"text-3d-flip-char inline transform-3d", "group-hover/text-3d-flip:[transform:var(--text-3d-flip-turn)]",
		`style="--text-3d-flip-delay:0s"`, `style="--text-3d-flip-delay:0.05s"`, `style="--text-3d-flip-delay:0.1s"`,
		`<span class="relative h-[1lh] backface-hidden" style="transform:rotateY(-90deg) translateX(50%) rotateY(90deg)">a</span>`,
		`<span class="whitespace-pre"> </span>`)
	if n := strings.Count(out, "whitespace-pre"); n != 1 {
		t.Errorf("%d spaces, want 1", n)
	}
}

func TestText3DFlipOptions(t *testing.T) {
	out := render(t, Text3DFlip(Text3DFlipOptions{
		As:              Text3DFlipH2,
		TextClass:       "text-foreground",
		FlipTextClass:   "text-primary",
		StaggerDuration: 30 * time.Millisecond,
		StaggerFrom:     Text3DFlipFromCenter,
		RotateDirection: Text3DFlipTop,
		Duration:        time.Second,
	}, html.Class("flex-nowrap"), "abcde"))
	mustContain(t, out, `<h2 data-slot="text-3d-flip"`, `class="group/text-3d-flip relative flex flex-nowrap"`,
		"--text-3d-flip-rest:translateZ(-0.5lh);--text-3d-flip-turn:rotateX(90deg);--text-3d-flip-duration:1s",
		`backface-hidden text-foreground" style="transform:translateZ(0.5lh)"`,
		`backface-hidden text-primary" style="transform:rotateX(-90deg) translateZ(0.5lh)"`)
	// center of 5 is index 2: delays 2,1,0,1,2 times 30ms
	if got, want := text3DFlipDelays(Text3DFlipOptions{StaggerFrom: Text3DFlipFromCenter, StaggerDuration: 30 * time.Millisecond}, 5),
		[]time.Duration{60, 30, 0, 30, 60}; !equalMillis(got, want) {
		t.Errorf("center delays = %v, want %v ms", got, want)
	}
	if got, want := text3DFlipDelays(Text3DFlipOptions{StaggerFrom: Text3DFlipFromLast}, 3), []time.Duration{100, 50, 0}; !equalMillis(got, want) {
		t.Errorf("last delays = %v, want %v ms", got, want)
	}
	if got, want := text3DFlipDelays(Text3DFlipOptions{StaggerFrom: Text3DFlipFromIndex, StaggerIndex: 1}, 3), []time.Duration{50, 0, 50}; !equalMillis(got, want) {
		t.Errorf("index delays = %v, want %v ms", got, want)
	}
}

// equalMillis compares durations with wants given in milliseconds.
func equalMillis(got, wantMillis []time.Duration) bool {
	if len(got) != len(wantMillis) {
		return false
	}
	for i := range got {
		if got[i] != wantMillis[i]*time.Millisecond {
			return false
		}
	}
	return true
}

func TestText3DFlipRandomSeed(t *testing.T) {
	opts := Text3DFlipOptions{StaggerFrom: Text3DFlipFromRandom, Seed: 7}
	a, b := render(t, Text3DFlip(opts, "random text")), render(t, Text3DFlip(opts, "random text"))
	if a != b {
		t.Error("the same seed rendered different markup")
	}
	opts.Seed = 8
	if render(t, Text3DFlip(opts, "random text")) == a {
		t.Error("another seed rendered the same markup")
	}
}

func TestText3DFlipNonText(t *testing.T) {
	if err := renderErr(t, Text3DFlip(Text3DFlipOptions{}, html.Span("x"))); err == nil {
		t.Error("a non-text child rendered without error")
	}
}
