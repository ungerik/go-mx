package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func AvatarDemo() mx.Component {
	return shadcn.Avatar(
		shadcn.AvatarImage(html.Src("https://github.com/shadcn.png"), html.Alt("@shadcn")),
		shadcn.AvatarFallback("CN"),
	)
}

func AvatarFallbackDemo() mx.Component {
	return shadcn.Avatar(
		shadcn.AvatarImage(html.Src("https://broken.example/missing.png"), html.Alt("@error")),
		shadcn.AvatarFallback("JD"),
	)
}
