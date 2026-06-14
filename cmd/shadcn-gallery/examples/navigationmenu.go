package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func NavigationMenuDemo() mx.Component {
	return shadcn.NavigationMenu(
		shadcn.NavigationMenuList(
			shadcn.NavigationMenuItem(
				shadcn.NavigationMenuTrigger("nm-start", "Getting started"),
				shadcn.NavigationMenuContent("nm-start", "",
					html.DivClass("grid w-[320px] gap-1 p-2",
						shadcn.NavigationMenuLink(false, html.HRef("#"), "Introduction"),
						shadcn.NavigationMenuLink(false, html.HRef("#"), "Installation"),
						shadcn.NavigationMenuLink(true, html.HRef("#"), "Typography"),
					),
				),
			),
			shadcn.NavigationMenuItem(
				shadcn.NavigationMenuTrigger("nm-components", "Components"),
				shadcn.NavigationMenuContent("nm-components", "",
					html.DivClass("grid w-[320px] gap-1 p-2",
						shadcn.NavigationMenuLink(false, html.HRef("#"), "Alert Dialog"),
						shadcn.NavigationMenuLink(false, html.HRef("#"), "Hover Card"),
						shadcn.NavigationMenuLink(false, html.HRef("#"), "Progress"),
					),
				),
			),
			shadcn.NavigationMenuItem(
				shadcn.NavigationMenuLink(false, html.HRef("#"), "Docs"),
			),
		),
	)
}
