package magicui

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestTextAnimateDefault(t *testing.T) {
	out := render(t, TextAnimate(TextAnimateOptions{}, "Hello  world"))
	mustContain(t, out, `<p data-slot="text-animate"`, `class="flex flex-wrap whitespace-pre-wrap"`,
		"--text-animate-delay:0s;--text-animate-stagger:0.1s",
		`<span class="sr-only">Hello  world</span>`,
		`<span class="inline-block whitespace-pre animate-text-animate-fade-in" style="--i:0" aria-hidden="true">Hello</span>`,
		`style="--i:1" aria-hidden="true">  </span>`,
		`style="--i:2" aria-hidden="true">world</span>`)
}

func TestTextAnimateOptions(t *testing.T) {
	out := render(t, TextAnimate(TextAnimateOptions{
		Animation:    TextAnimateBlurInUp,
		By:           TextAnimateByCharacter,
		As:           TextAnimateH1,
		Delay:        2 * time.Second,
		Duration:     time.Second,
		SegmentClass: "text-red-500",
	}, html.Class("whitespace-normal"), "äb"))
	mustContain(t, out, `<h1 data-slot="text-animate"`, `class="flex flex-wrap whitespace-normal"`,
		"--text-animate-delay:2s;--text-animate-stagger:0.5s",
		`<span class="inline-block whitespace-pre animate-text-animate-blur-in-up text-red-500" style="--i:0" aria-hidden="true">ä</span>`)

	out = render(t, TextAnimate(TextAnimateOptions{By: TextAnimateByLine, NoScreenReaderText: true, As: "marquee"}, "a\n\nb"))
	mustContain(t, out, `<p data-slot="text-animate"`, `<span class="block animate-text-animate-fade-in" style="--i:1"></span>`)
	if strings.Contains(out, "sr-only") || strings.Contains(out, "aria-hidden") {
		t.Errorf("NoScreenReaderText should leave out the sr-only copy and aria-hidden: %s", out)
	}
}

func TestTextAnimateStartOnView(t *testing.T) {
	out := render(t, TextAnimate(TextAnimateOptions{StartOnView: true, Animation: TextAnimateScaleUp, By: TextAnimateByText}, "x"))
	mustContain(t, out, "[view-timeline-name:--text-animate]", "--n:1", "--text-animate-keyframes:text-animate-scale-up, text-animate-fade",
		`class="inline-block whitespace-pre animate-text-animate-on-view"`)
}

func TestTextAnimatePresets(t *testing.T) {
	for _, a := range TextAnimateAnimation("").Enums() {
		if textAnimateClasses[a] == "" || textAnimateViewKeyframes[a] == "" {
			t.Errorf("preset %s has no class or view keyframes", a)
		}
		if !strings.Contains(CSS, "--"+textAnimateClasses[a]+":") {
			t.Errorf("CSS does not define %s", textAnimateClasses[a])
		}
		for name := range strings.SplitSeq(textAnimateViewKeyframes[a], ", ") {
			if !strings.Contains(CSS, "@keyframes "+name+" {") {
				t.Errorf("CSS has no @keyframes %s", name)
			}
		}
	}
}

func TestTextAnimateSegments(t *testing.T) {
	for _, tt := range []struct {
		text string
		by   TextAnimateBy
		want []string
	}{
		{" a b ", TextAnimateByWord, []string{" ", "a", " ", "b", " "}},
		{"a\nb", TextAnimateByLine, []string{"a", "b"}},
		{"ab", TextAnimateByText, []string{"ab"}},
		{"😀a", TextAnimateByCharacter, []string{"😀", "a"}},
	} {
		if got := textAnimateSegments(tt.text, tt.by); !slices.Equal(got, tt.want) {
			t.Errorf("textAnimateSegments(%q, %s) = %q, want %q", tt.text, tt.by, got, tt.want)
		}
	}
}

func TestTextAnimateNonText(t *testing.T) {
	if err := renderErr(t, TextAnimate(TextAnimateOptions{}, html.Span("x"))); err == nil {
		t.Error("a non-text child rendered without error")
	}
}

// TestTextAnimateCharacterWords checks that, by character, the letters of a
// word share an inline-flex span and the space between words does not, so
// the flex container wraps between words only.
func TestTextAnimateCharacterWords(t *testing.T) {
	out := render(t, TextAnimate(TextAnimateOptions{By: TextAnimateByCharacter}, "ab c"))
	if n := strings.Count(out, `<span class="inline-flex">`); n != 2 {
		t.Errorf("%d word spans, want 2: %s", n, out)
	}
	mustContain(t, out, `style="--i:2" aria-hidden="true"> </span>`)
}
