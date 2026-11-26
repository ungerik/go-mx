package magicui

import (
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestComicText(t *testing.T) {
	out := render(t, ComicText(ComicTextOptions{}, "BOOM!"))
	mustContain(t, out,
		`data-slot="comic-text"`, "text-center select-none animate-comic-text-pop",
		"font-size:5rem;font-family:'Bangers', 'Comic Sans MS', 'Impact', sans-serif;font-weight:900;-webkit-text-stroke:1.75px #000000",
		"transform:skewX(-10deg)", "drop-shadow(3px 3px 0px #EF4444)", "background-color:#FACC15",
		"-webkit-text-fill-color:transparent", ">BOOM!</div>",
	)
	out = render(t, ComicText(ComicTextOptions{FontSize: 2}, html.Class("text-left"), html.Style("color:red"), "Pow"))
	mustContain(t, out, "font-size:2rem", "-webkit-text-stroke:0.7px", "select-none animate-comic-text-pop text-left", "-webkit-text-fill-color:transparent;color:red")
}
