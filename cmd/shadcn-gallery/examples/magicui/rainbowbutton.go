package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// RainbowButtonDemo is the default rainbow button.
func RainbowButtonDemo() mx.Component {
	return magicui.RainbowButton(magicui.RainbowButtonOptions{}, "Get Unlimited Access")
}

// RainbowButtonOutline is the outline variant of the rainbow button.
func RainbowButtonOutline() mx.Component {
	return magicui.RainbowButton(magicui.RainbowButtonOptions{Variant: magicui.RainbowButtonOutline}, "Get Unlimited Access")
}
