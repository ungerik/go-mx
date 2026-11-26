package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestBlurFadeDefault(t *testing.T) {
	out := render(t, BlurFade(BlurFadeOptions{}, html.Img(html.Src("/a.png"))))
	mustContain(t, out, `data-slot="blur-fade"`, `class="animate-blur-fade-in"`,
		"--blur-fade-duration:0.4s", "--blur-fade-delay:0.04s", "--blur-fade-blur:6px", "--blur-fade-y:-6px", `<img src="/a.png"/>`)
	if strings.Contains(out, "--blur-fade-x") {
		t.Errorf("direction down should not move horizontally: %s", out)
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
	mustContain(t, out, `class="animate-blur-fade-in-view mb-4"`, "--blur-fade-duration:1s", "--blur-fade-delay:0.29s",
		"--blur-fade-blur:10px", "--blur-fade-x:20px", ";color:red")
	for dir, want := range map[Direction]string{DirectionUp: "--blur-fade-y:6px", DirectionRight: "--blur-fade-x:-6px"} {
		mustContain(t, render(t, BlurFade(BlurFadeOptions{Direction: dir})), want)
	}
}

func TestBlurFadeInvalidBlur(t *testing.T) {
	if err := renderErr(t, BlurFade(BlurFadeOptions{Blur: "red;x:y"})); err == nil {
		t.Error("an unsafe blur value rendered without error")
	}
}
