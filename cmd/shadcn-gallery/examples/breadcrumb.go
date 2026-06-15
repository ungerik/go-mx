package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// BreadcrumbDemo renders a breadcrumb trail of links ending in the current page.
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

// BreadcrumbWithEllipsis renders a breadcrumb trail with an ellipsis collapsing intermediate items.
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
