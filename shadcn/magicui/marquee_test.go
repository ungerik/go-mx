package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestMarqueeDefault(t *testing.T) {
	out := render(t, Marquee(MarqueeOptions{}, html.Span("item")))
	mustContain(t, out, `data-slot="marquee"`, "[--duration:40s]", "[--gap:1rem]", "flex-row", "animate-marquee ")
	if n := strings.Count(out, "<span>item</span>"); n != 4 {
		t.Errorf("content repeated %d times, want 4", n)
	}
	if n := strings.Count(out, `aria-hidden="true"`); n != 3 {
		t.Errorf("%d copies aria-hidden, want 3", n)
	}
}

func TestMarqueeOptions(t *testing.T) {
	out := render(t, Marquee(MarqueeOptions{Vertical: true, Reverse: true, PauseOnHover: true, Repeat: 2},
		html.Class("[--duration:20s]"), "x"))
	mustContain(t, out, "flex-col", "animate-marquee-vertical", "[animation-direction:reverse]",
		"group-hover:[animation-play-state:paused]", "[--duration:20s]")
	if strings.Contains(out, "[--duration:40s]") {
		t.Errorf("caller duration should replace the default: %s", out)
	}
	if n := strings.Count(out, "animate-marquee-vertical"); n != 2 {
		t.Errorf("%d copies, want 2", n)
	}
}

func TestMarqueeNegativeRepeat(t *testing.T) {
	if renderErr(t, Marquee(MarqueeOptions{Repeat: -1}, html.Span("item"))) == nil {
		t.Error("negative Repeat accepted")
	}
}
