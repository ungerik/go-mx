package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestSpinningText(t *testing.T) {
	out := render(t, SpinningText(SpinningTextOptions{}, "ab"))
	mustContain(t, out, `data-slot="spinning-text"`, `class="relative animate-spinning-text"`,
		"--spinning-text-duration:10s;--total:3;--spinning-text-radius:5",
		`style="--index:2;transform:translate(-50%, -50%) rotate(calc(360deg / var(--total) * var(--index))) translateY(calc(var(--spinning-text-radius, 5) * -1ch));transform-origin:center"> </span>`,
		`<span class="sr-only">ab</span>`)
	if n := strings.Count(out, `aria-hidden="true"`); n != 3 {
		t.Errorf("%d hidden letters, want 3", n)
	}

	out = render(t, SpinningText(SpinningTextOptions{Duration: 4 * time.Second, Reverse: true, Radius: 6}, html.Class("text-4xl"), "x"))
	mustContain(t, out, "[animation-direction:reverse]", "text-4xl", "--spinning-text-duration:4s", "--spinning-text-radius:6")
}

func TestSpinningTextNonText(t *testing.T) {
	if err := renderErr(t, SpinningText(SpinningTextOptions{}, html.B("x"))); err == nil {
		t.Error("a non-text child rendered without error")
	}
}
