package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func HoverCardDemo() mx.Component {
	return shadcn.HoverCard(
		shadcn.HoverCardTrigger("demo-hovercard", 0, 0,
			shadcn.Button(shadcn.ButtonLink, shadcn.SizeDefault, "@nextjs")),
		shadcn.HoverCardContent("demo-hovercard", "", 0, 0,
			html.Div(html.Class("flex justify-between gap-4"),
				shadcn.Avatar(
					shadcn.AvatarImage(html.Src("https://github.com/vercel.png")),
					shadcn.AvatarFallback("VC"),
				),
				html.Div(html.Class("space-y-1"),
					html.Element("h4", html.Class("text-sm font-semibold"), "@nextjs"),
					html.P(html.Class("text-sm"), "The React Framework – created and maintained by @vercel."),
				),
			),
		),
	)
}
