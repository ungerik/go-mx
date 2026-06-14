package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func PaginationDemo() mx.Component {
	return shadcn.Pagination(
		shadcn.PaginationContent(
			shadcn.PaginationItem(shadcn.PaginationPrevious(html.HRef("#"))),
			shadcn.PaginationItem(shadcn.PaginationLink(false, shadcn.SizeIcon, html.HRef("#"), "1")),
			shadcn.PaginationItem(shadcn.PaginationLink(true, shadcn.SizeIcon, html.HRef("#"), "2")),
			shadcn.PaginationItem(shadcn.PaginationLink(false, shadcn.SizeIcon, html.HRef("#"), "3")),
			shadcn.PaginationItem(shadcn.PaginationEllipsis()),
			shadcn.PaginationItem(shadcn.PaginationNext(html.HRef("#"))),
		),
	)
}
