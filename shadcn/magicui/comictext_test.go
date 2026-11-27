package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestComicText(t *testing.T) {
	out := render(t, ComicText(ComicTextOptions{}, "BOOM!"))
	mustContain(t, out,
		`data-slot="comic-text"`, "text-center select-none animate-comic-text-pop",
		"[--comic-text-font-size:5] text-[length:calc(var(--comic-text-font-size)*1rem)]", "[font-family:Bangers,Comic_Sans_MS,Impact,sans-serif]",
		"[-webkit-text-stroke:calc(var(--comic-text-font-size)*0.35px)_#000000]", "[transform:skewX(-10deg)]",
		"drop-shadow(3px_3px_0px_#EF4444)", "bg-[#FACC15]", "bg-[radial-gradient(", "[-webkit-background-clip:text]", "[-webkit-text-fill-color:transparent]", ">BOOM!</div>",
	)
	if strings.Contains(out, "style=") {
		t.Errorf("the default size needs no style attribute: %s", out)
	}
	out = render(t, ComicText(ComicTextOptions{FontSize: 2}, html.Class("text-left"), html.Style("color:red"), "Pow"))
	mustContain(t, out, `style="--comic-text-font-size:2;color:red"`, "select-none animate-comic-text-pop text-left")
	if strings.Contains(out, "text-center") {
		t.Errorf("caller text-left should replace text-center: %s", out)
	}
	// The font size is a text-* class, so a caller's font-size class
	// replaces it instead of racing it in the stylesheet.
	out = render(t, ComicText(ComicTextOptions{}, html.Class("text-5xl"), "Pow"))
	if strings.Contains(out, "text-[length:") || !strings.Contains(out, "text-5xl") {
		t.Errorf("caller text-5xl should replace the font size class: %s", out)
	}
}
