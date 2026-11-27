package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// TextRevealDemo reveals a sentence word by word while the page scrolls.
func TextRevealDemo() mx.Component {
	return magicui.TextReveal("Magic UI will change the way you design.")
}
