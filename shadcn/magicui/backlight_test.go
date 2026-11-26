package magicui

import (
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestBacklight(t *testing.T) {
	out := render(t, Backlight(BacklightOptions{ID: "glow"}, html.Class("w-full"), html.Span("content")))
	mustContain(t, out, `<div data-slot="backlight" class="w-full">`,
		`<svg width="0" height="0" aria-hidden="true">`,
		`<filter id="glow" y="-50%" x="-50%" width="200%" height="200%">`,
		`<feGaussianBlur in="SourceGraphic" stdDeviation="20" result="blurred">`,
		`<feColorMatrix type="saturate" in="blurred" values="4">`,
		`<feComposite in="SourceGraphic" operator="over">`,
		`<div style="filter:url(#glow)"><span>content</span></div>`)
	out = render(t, Backlight(BacklightOptions{ID: "b", Blur: 5}))
	mustContain(t, out, `stdDeviation="5"`)
	if renderErr(t, Backlight(BacklightOptions{ID: "b);x:y"})) == nil {
		t.Error("invalid id accepted")
	}
}
