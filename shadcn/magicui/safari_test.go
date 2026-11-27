package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

func TestSafari(t *testing.T) {
	out := render(t, Safari(SafariOptions{URL: "magicui.design"}))
	mustContain(t, out,
		`data-slot="safari"`, "aspect-[1203/753]", `viewBox="0 0 1203 753"`,
		`class="absolute inset-0 z-10 size-full [transform:translateZ(0)]"`, `aria-hidden="true"`,
		`font-family="Arial, sans-serif">magicui.design</text>`,
		`d="M265.5 33.8984C265.641`, // the toolbar of the default mode
	)
	if strings.Contains(out, "mask") {
		t.Error("mask without media")
	}

	out = render(t, Safari(SafariOptions{Mode: SafariModeSimple, ImageSrc: "/page.png", ID: "win"}))
	mustContain(t, out,
		`<g mask="url(#win)">`, `<mask id="win" maskUnits="userSpaceOnUse">`,
		"[border-radius:0_0_11px_11px]", `<img src="/page.png" alt=""`, safariScreenClasses,
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
