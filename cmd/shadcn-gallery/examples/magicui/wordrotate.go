package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// WordRotateDemo alternates between two words.
func WordRotateDemo() mx.Component {
	return magicui.WordRotate(magicui.WordRotateOptions{Words: []string{"Word", "Rotate"}},
		html.Class("text-4xl font-bold text-black dark:text-white"),
	)
}
