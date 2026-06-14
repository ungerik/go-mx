package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func BreadcrumbDemo() mx.Component {
	return shadcn.Breadcrumb(
		shadcn.BreadcrumbList(
			shadcn.BreadcrumbItem(shadcn.BreadcrumbLink(html.HRef("/"), "Home")),
			shadcn.BreadcrumbSeparator(),
			shadcn.BreadcrumbItem(shadcn.BreadcrumbLink(html.HRef("/components"), "Components")),
			shadcn.BreadcrumbSeparator(),
			shadcn.BreadcrumbItem(shadcn.BreadcrumbPage("Breadcrumb")),
		),
	)
}

func BreadcrumbWithEllipsis() mx.Component {
	return shadcn.Breadcrumb(
		shadcn.BreadcrumbList(
			shadcn.BreadcrumbItem(shadcn.BreadcrumbLink(html.HRef("/"), "Home")),
			shadcn.BreadcrumbSeparator(),
			shadcn.BreadcrumbItem(shadcn.BreadcrumbEllipsis()),
			shadcn.BreadcrumbSeparator(),
			shadcn.BreadcrumbItem(shadcn.BreadcrumbLink(html.HRef("/components"), "Components")),
			shadcn.BreadcrumbSeparator(),
			shadcn.BreadcrumbItem(shadcn.BreadcrumbPage("Breadcrumb")),
		),
	)
}
