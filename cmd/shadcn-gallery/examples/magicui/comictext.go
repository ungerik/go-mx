package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// ComicTextDemo shows "BOOM!" in comic lettering.
func ComicTextDemo() mx.Component {
	return html.DivClass("space-y-8 text-center",
		magicui.ComicText(magicui.ComicTextOptions{FontSize: 5}, "BOOM!"),
	)
}
