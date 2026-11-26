package magicui

import (
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAndroid(t *testing.T) {
	out := render(t, Android(AndroidOptions{ID: "droid"}, html.Class("size-full")))
	mustContain(t, out,
		`<svg data-slot="android" width="433" height="882" viewBox="0 0 433 882" fill="none" xmlns="http://www.w3.org/2000/svg" class="size-full">`,
		`<g clip-path="url(#droid)">`, `<clipPath id="droid">`,
		`<rect width="360" height="800" rx="33" ry="25" class="fill-white dark:fill-[#262626]" transform="translate(9 14)">`,
	)

	out = render(t, Android(AndroidOptions{ID: "droid", Width: 300, Height: 600, Src: "/s.png", VideoSrc: "/v.mp4"}))
	mustContain(t, out,
		`width="300" height="600" viewBox="0 0 300 600"`,
		`<image href="/s.png" width="360" height="800" class="size-full object-cover" preserveAspectRatio="xMidYMid slice" clip-path="url(#droid)">`,
		`<foreignObject width="380" height="820" clip-path="url(#droid)" preserveAspectRatio="xMidYMid slice"><video class="size-full object-cover" src="/v.mp4" autoplay="autoplay" loop="loop" muted="muted" playsinline="playsinline"></video>`,
	)

	// The caller's attributes replace the defaults, as upstream spreads its
	// props last.
	out = render(t, Android(AndroidOptions{ID: "droid"}, html.Attrib("viewBox", "0 0 10 10")))
	mustContain(t, out, `viewBox="0 0 10 10"`)
}
