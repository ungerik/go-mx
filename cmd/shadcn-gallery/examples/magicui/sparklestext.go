package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// SparklesTextDemo lets sparkles twinkle over a bold text.
func SparklesTextDemo() mx.Component {
	return magicui.SparklesText(magicui.SparklesTextOptions{Seed: 1}, "Magic UI")
}
