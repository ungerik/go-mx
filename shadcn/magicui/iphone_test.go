package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestIphoneEmpty(t *testing.T) {
	out := render(t, Iphone(IphoneOptions{}, html.Class("w-64")))
	mustContain(t, out,
		`data-slot="iphone"`, `class="relative inline-block align-middle leading-none w-64"`, `style="aspect-ratio:433/882"`,
		`viewBox="0 0 433 882"`, `aria-hidden="true"`,
		`d="M2 73C2 32.6832 34.6832 0 75 0H357C397.317 0 430 32.6832 430 73V809C430 849.317 397.317 882 357 882H75C34.6832 882 2 849.317 2 809V73Z"`,
		`d="M21.25 75C21.25 44.2101 46.2101 19.25 77 19.25H355`,
	)
	for _, absent := range []string{"<img", "<video", "mask"} {
		if strings.Contains(out, absent) {
			t.Errorf("empty iPhone renders %s", absent)
		}
	}
}

func TestIphoneMedia(t *testing.T) {
	out := render(t, Iphone(IphoneOptions{Src: "/screen.png", ID: "phone"}))
	mustContain(t, out,
		`<img src="/screen.png" alt="" class="block size-full object-cover object-top"/>`,
		"left:4.907621247113164%;top:2.1825396825396823%;width:89.95381062355659%;height:95.63492063492063%;border-radius:14.313222079589217% / 6.609365737996444%",
		`<g mask="url(#phone)">`, `mask="url(#phone)"></path>`, `<mask id="phone" maskUnits="userSpaceOnUse">`,
		`<rect x="21.25" y="19.25" width="389.5" height="843.5" rx="55.75" ry="55.75" fill="black">`,
	)
	out = render(t, Iphone(IphoneOptions{Src: "/screen.png", VideoSrc: "/v.mp4", ID: "phone"}))
	mustContain(t, out, `<video class="block size-full object-cover" src="/v.mp4" autoplay="autoplay" loop="loop" muted="muted" playsinline="playsinline" preload="metadata">`)
	if strings.Contains(out, "<img") {
		t.Error("VideoSrc should take precedence over Src")
	}
}

func TestIphoneDefaultMaskID(t *testing.T) {
	a := render(t, Iphone(IphoneOptions{Src: "/s.png"}))
	b := render(t, Iphone(IphoneOptions{Src: "/s.png"}))
	if a == b {
		t.Error("two iPhones without ID share a mask id")
	}
}
