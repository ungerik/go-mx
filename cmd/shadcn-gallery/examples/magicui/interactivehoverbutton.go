package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// InteractiveHoverButtonDemo is a pill button whose dot floods it on hover
// while the label slides over to make room for an arrow.
func InteractiveHoverButtonDemo() mx.Component {
	return magicui.InteractiveHoverButton("Hover Me")
}
