package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// AvatarDemo renders an avatar with an image and a text fallback.
func AvatarDemo() mx.Component {
	return shadcn.Avatar(
		shadcn.AvatarImage(html.Src("https://github.com/shadcn.png"), html.Alt("@shadcn")),
		shadcn.AvatarFallback("CN"),
	)
}

// AvatarFallbackDemo renders an avatar whose broken image falls back to initials.
func AvatarFallbackDemo() mx.Component {
	return shadcn.Avatar(
		shadcn.AvatarImage(html.Src("https://broken.example/missing.png"), html.Alt("@error")),
		shadcn.AvatarFallback("JD"),
	)
}
