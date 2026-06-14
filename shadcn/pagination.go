package shadcn

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Pagination renders a shadcn/ui pagination navigation landmark as a <nav>
// with role="navigation" and aria-label="pagination". Compose it with
// [PaginationContent] and the other pagination parts.
func Pagination(attribsChildren ...any) *mx.Element {
	e := html.Nav(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("navigation"))
	}
	if e.AttribIndex("aria-label") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-label", "pagination"))
	}
	return finish(e, "pagination", "mx-auto flex w-full justify-center")
}

// PaginationContent renders the <ul> holding the pagination items.
func PaginationContent(attribsChildren ...any) *mx.Element {
	return finish(html.UL(attribsChildren...), "pagination-content", "flex flex-row items-center gap-1")
}

// PaginationItem renders one <li> in a [PaginationContent].
func PaginationItem(attribsChildren ...any) *mx.Element {
	return finish(html.LI(attribsChildren...), "pagination-item", "")
}

// PaginationLink renders a page link as an <a> styled like a [Button]. active
// marks the current page: it adds aria-current="page" and uses the outline
// button variant, otherwise the ghost variant. size may be "" for the default
// icon size. Pass html.HRef for the target.
func PaginationLink(active bool, size ButtonSize, attribsChildren ...any) *mx.Element {
	if size == "" {
		size = SizeIcon
	}
	e := html.A(attribsChildren...)
	if active && e.AttribIndex("aria-current") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-current", "page"))
	}
	if e.AttribIndex("data-active") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("active", strconv.FormatBool(active)))
	}
	variant := ButtonGhost
	if active {
		variant = ButtonOutline
	}
	return finish(e, "pagination-link", ButtonClasses(variant, size))
}

// PaginationPrevious renders the "previous page" link: a default-size
// [PaginationLink] with a chevron-left icon and a "Previous" label.
func PaginationPrevious(attribsChildren ...any) *mx.Element {
	args := make([]any, 0, len(attribsChildren)+4)
	args = append(args,
		html.Attrib("aria-label", "Go to previous page"),
		html.Class("gap-1 px-2.5 sm:pl-2.5"),
		iconChevronLeft(),
		html.SpanClass("hidden sm:block", "Previous"),
	)
	args = append(args, attribsChildren...)
	return PaginationLink(false, SizeDefault, args...)
}

// PaginationNext renders the "next page" link: a default-size [PaginationLink]
// with a "Next" label and a chevron-right icon.
func PaginationNext(attribsChildren ...any) *mx.Element {
	args := make([]any, 0, len(attribsChildren)+4)
	args = append(args,
		html.Attrib("aria-label", "Go to next page"),
		html.Class("gap-1 px-2.5 sm:pr-2.5"),
		html.SpanClass("hidden sm:block", "Next"),
		iconChevronRight(),
	)
	args = append(args, attribsChildren...)
	return PaginationLink(false, SizeDefault, args...)
}

// PaginationEllipsis renders a collapsed-pages indicator as a <span> with a
// lucide ellipsis icon and screen-reader text.
func PaginationEllipsis(attribsChildren ...any) *mx.Element {
	e := html.Span(attribsChildren...)
	if e.AttribIndex("aria-hidden") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-hidden", "true"))
	}
	if len(e.Children) == 0 {
		e.Children = mx.Components{
			iconEllipsis("size-4"),
			html.SpanClass("sr-only", "More pages"),
		}
	}
	return finish(e, "pagination-ellipsis", "flex size-9 items-center justify-center")
}
