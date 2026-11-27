package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestLineShadowText(t *testing.T) {
	out := render(t, LineShadowText(LineShadowTextOptions{}, html.Class("italic"), "Fast"))
	mustContain(t, out,
		`<span `, `data-slot="line-shadow-text"`, `data-text="Fast"`, "[--shadow-color:black]",
		"after:content-[attr(data-text)]", "after:animate-line-shadow", "italic", ">Fast</span>",
	)
	out = render(t, LineShadowText(LineShadowTextOptions{ShadowColor: "white", As: LineShadowTextH2}, "a<", "b"))
	mustContain(t, out, "<h2 ", `data-text="a&lt;b"`, `style="--shadow-color:white"`)
	if !strings.HasSuffix(out, "</h2>") {
		t.Errorf("want an h2: %s", out)
	}
}

func TestLineShadowTextErrors(t *testing.T) {
	if err := renderErr(t, LineShadowText(LineShadowTextOptions{}, html.B("x"))); err == nil {
		t.Error("a non-text child should be a render error")
	}
	if err := renderErr(t, LineShadowText(LineShadowTextOptions{ShadowColor: "red;x:y"}, "x")); err == nil {
		t.Error("an invalid shadow color should be a render error")
	}
}
