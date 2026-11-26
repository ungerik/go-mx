package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

func TestSafari(t *testing.T) {
	out := render(t, Safari(SafariOptions{URL: "magicui.design"}))
	mustContain(t, out,
		`data-slot="safari"`, `style="aspect-ratio:1203/753"`, `viewBox="0 0 1203 753"`,
		`class="absolute inset-0 z-10 size-full"`, `aria-hidden="true"`,
		`font-family="Arial, sans-serif">magicui.design</text>`,
		`d="M265.5 33.8984C265.641`, // the toolbar of the default mode
	)
	if strings.Contains(out, "mask") {
		t.Error("mask without media")
	}

	out = render(t, Safari(SafariOptions{Mode: SafariModeSimple, ImageSrc: "/page.png", ID: "win"}))
	mustContain(t, out,
		`<g mask="url(#win)">`, `<mask id="win" maskUnits="userSpaceOnUse">`,
		"border-radius:0 0 11px 11px", `<img src="/page.png" alt=""`,
		"left:0.0831255195344971%;top:6.905710491367861%;width:99.75062344139651%;height:92.96148738379813%",
	)
	if strings.Contains(out, "M265.5 33.8984") {
		t.Error("simple mode draws the toolbar")
	}

	out = render(t, Safari(SafariOptions{VideoSrc: "/v.mp4", ID: "win"}))
	mustContain(t, out, `src="/v.mp4" autoplay="autoplay" loop="loop" muted="muted" playsinline="playsinline" preload="metadata"`)
	if strings.Contains(out, "11px") {
		t.Error("the video has no rounded bottom upstream")
	}
}

func TestDeviceMockupsRejectInvalidID(t *testing.T) {
	const id = "a b)"
	for name, c := range map[string]mx.Component{
		"Safari":  Safari(SafariOptions{ImageSrc: "a.png", ID: id}),
		"Iphone":  Iphone(IphoneOptions{Src: "a.png", ID: id}),
		"Android": Android(AndroidOptions{Src: "a.png", ID: id}),
	} {
		if renderErr(t, c) == nil {
			t.Errorf("%s accepted the invalid id %q", name, id)
		}
	}
}
