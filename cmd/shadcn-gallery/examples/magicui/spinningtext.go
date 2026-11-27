package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// SpinningTextDemo turns a text on a circle.
func SpinningTextDemo() mx.Component {
	return magicui.SpinningText(magicui.SpinningTextOptions{}, "learn more • earn more • grow more •")
}

// SpinningTextReverse turns a larger text counterclockwise, faster and on a
// wider circle.
func SpinningTextReverse() mx.Component {
	return magicui.SpinningText(magicui.SpinningTextOptions{Reverse: true, Duration: 4 * time.Second, Radius: 6},
		html.Class("text-4xl"),
		"learn more • earn more • grow more •",
	)
}
