package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Breadcrumb renders a shadcn/ui breadcrumb navigation landmark as a <nav>
// with aria-label="breadcrumb". Compose it with [BreadcrumbList] and the
// other breadcrumb parts.
func Breadcrumb(attribsChildren ...any) *mx.Element {
	e := html.Nav(attribsChildren...)
	if e.AttribIndex("aria-label") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-label", "breadcrumb"))
	}
	return finish(e, "breadcrumb", "")
}

// BreadcrumbList renders the breadcrumb's <ol> of items.
func BreadcrumbList(attribsChildren ...any) *mx.Element {
	return finish(html.OL(attribsChildren...), "breadcrumb-list",
		"text-muted-foreground flex flex-wrap items-center gap-1.5 text-sm break-words sm:gap-2.5")
}

// BreadcrumbItem renders one <li> in a [BreadcrumbList].
func BreadcrumbItem(attribsChildren ...any) *mx.Element {
	return finish(html.LI(attribsChildren...), "breadcrumb-item", "inline-flex items-center gap-1.5")
}

// BreadcrumbLink renders a breadcrumb hyperlink as an <a>. Pass html.HRef for
// the target.
func BreadcrumbLink(attribsChildren ...any) *mx.Element {
	return finish(html.A(attribsChildren...), "breadcrumb-link", "hover:text-foreground transition-colors")
}

// BreadcrumbPage renders the current page in a breadcrumb: a non-link <span>
// marked with aria-current="page".
func BreadcrumbPage(attribsChildren ...any) *mx.Element {
	e := html.Span(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("link"))
	}
	if e.AttribIndex("aria-disabled") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-disabled", "true"))
	}
	if e.AttribIndex("aria-current") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-current", "page"))
	}
	return finish(e, "breadcrumb-page", "text-foreground font-normal")
}

// BreadcrumbSeparator renders the separator <li> between breadcrumb items; it
// is hidden from assistive technology. With no children it defaults to a
// lucide chevron-right icon — pass children to use a different separator.
func BreadcrumbSeparator(attribsChildren ...any) *mx.Element {
	e := html.LI(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("presentation"))
	}
	if e.AttribIndex("aria-hidden") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-hidden", "true"))
	}
	if len(e.Children) == 0 {
		e.Children = mx.Components{iconChevronRight()}
	}
	return finish(e, "breadcrumb-separator", "[&>svg]:size-3.5")
}

// BreadcrumbEllipsis renders a collapsed-items indicator as a <span>. With no
// children it defaults to a lucide ellipsis icon plus screen-reader text.
func BreadcrumbEllipsis(attribsChildren ...any) *mx.Element {
	e := html.Span(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("presentation"))
	}
	if e.AttribIndex("aria-hidden") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-hidden", "true"))
	}
	if len(e.Children) == 0 {
		e.Children = mx.Components{
			iconEllipsis("size-4"),
			html.SpanClass("sr-only", "More"),
		}
	}
	return finish(e, "breadcrumb-ellipsis", "flex size-9 items-center justify-center")
}
