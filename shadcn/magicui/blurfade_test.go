package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestBlurFadeDefault(t *testing.T) {
	out := render(t, BlurFade(BlurFadeOptions{}, html.Img(html.Src("/a.png"))))
	mustContain(t, out, `data-slot="blur-fade"`,
		`class="animate-blur-fade-in [--blur-fade-duration:0.4s] [--blur-fade-delay:0.04s] [--blur-fade-blur:6px] [--blur-fade-x:0px] [--blur-fade-y:-6px]"`,
		`<img src="/a.png"/>`)
	if strings.Contains(out, "style=") {
		t.Errorf("the defaults need no style attribute: %s", out)
	}
}

func TestBlurFadeOptions(t *testing.T) {
	out := render(t, BlurFade(BlurFadeOptions{
		Duration:  time.Second,
		Delay:     250 * time.Millisecond,
		Offset:    20,
		Direction: DirectionLeft,
		InView:    true,
		Blur:      "10px",
	}, html.Class("mb-4"), html.Style("color:red"), "x"))
	mustContain(t, out, `class="animate-blur-fade-in-view `, ` mb-4"`,
		`style="--blur-fade-duration:1s;--blur-fade-delay:0.29s;--blur-fade-blur:10px;--blur-fade-x:20px;--blur-fade-y:0px;color:red"`)
	// A direction or an offset alone sets both axes, so the class defaults
	// of a vertical entrance do not leak into a horizontal one.
	for _, tc := range []struct {
		opts BlurFadeOptions
		want string
	}{
		{BlurFadeOptions{Direction: DirectionUp}, `style="--blur-fade-x:0px;--blur-fade-y:6px"`},
		{BlurFadeOptions{Direction: DirectionRight}, `style="--blur-fade-x:-6px;--blur-fade-y:0px"`},
		{BlurFadeOptions{Offset: 20}, `style="--blur-fade-x:0px;--blur-fade-y:-20px"`},
	} {
		mustContain(t, render(t, BlurFade(tc.opts)), tc.want)
	}
	// A caller class replaces a default class instead of competing with it.
	out = render(t, BlurFade(BlurFadeOptions{}, html.Class("[--blur-fade-y:12px]")))
	if strings.Contains(out, "[--blur-fade-y:-6px]") || !strings.Contains(out, "[--blur-fade-y:12px]") {
		t.Errorf("the caller's class should replace the default: %s", out)
	}
}

func TestBlurFadeNegativeDelay(t *testing.T) {
	// A negative Delay shortens the 40ms base delay, as upstream's sum does;
	// one that cancels it falls back to the base.
	mustContain(t, render(t, BlurFade(BlurFadeOptions{Delay: -20 * time.Millisecond})), `style="--blur-fade-delay:0.02s"`)
	mustContain(t, render(t, BlurFade(BlurFadeOptions{Delay: -time.Second})), `style="--blur-fade-delay:0.04s"`)
}

func TestBlurFadeInvalidBlur(t *testing.T) {
	if err := renderErr(t, BlurFade(BlurFadeOptions{Blur: "red;x:y"})); err == nil {
		t.Error("an unsafe blur value rendered without error")
	}
}
