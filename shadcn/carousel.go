package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Carousel is a Go port of shadcn/ui's Carousel. shadcn wraps embla-carousel;
// this port replaces embla's transform-based sliding with native CSS
// scroll-snap (`snap-x snap-mandatory` on an overflow-x-auto track), so the
// drag/swipe and keyboard scrolling come for free with no client framework.
// CarouselPrevious / CarouselNext scroll the track by one viewport via a tiny
// inline onclick — the only JS, since there is no declarative "scroll by a page"
// control. embla's autoplay, loop and orientation options are not reproduced.
func Carousel(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "carousel", "relative")
}

// CarouselContent is the scroll-snap track. Children are [CarouselItem]s.
func CarouselContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "carousel-content",
		"flex snap-x snap-mandatory overflow-x-auto scroll-smooth [scrollbar-width:none] [&::-webkit-scrollbar]:hidden")
}

// CarouselItem is one snap slide. It defaults to one slide per view
// (basis-full); pass html.Class("basis-1/2") etc. to show several.
func CarouselItem(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "carousel-item",
		"min-w-0 shrink-0 grow-0 basis-full snap-start pl-4 first:pl-0")
}

// carouselScrollBy scrolls the sibling track by ±one viewport width. The
// unquoted [data-slot=…] attribute selectors avoid escaping inside the
// double-quoted onclick.
const carouselScrollPrev = "var c=this.closest('[data-slot=carousel]').querySelector('[data-slot=carousel-content]');c.scrollBy({left:-c.clientWidth,behavior:'smooth'})"
const carouselScrollNext = "var c=this.closest('[data-slot=carousel]').querySelector('[data-slot=carousel-content]');c.scrollBy({left:c.clientWidth,behavior:'smooth'})"

// CarouselPrevious renders the round previous-slide button (outside the track,
// left). Default onclick scrolls the track back one viewport.
func CarouselPrevious(attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick(carouselScrollPrev))
	}
	if e.AttribIndex("aria-label") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-label", "Previous slide"))
	}
	e.Children = append(e.Children, iconChevronLeft())
	return finish(e, "carousel-previous",
		ButtonClasses(ButtonOutline, SizeIcon)+" absolute top-1/2 -left-12 size-8 -translate-y-1/2 rounded-full")
}

// CarouselNext renders the round next-slide button (outside the track, right).
// Default onclick scrolls the track forward one viewport.
func CarouselNext(attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick(carouselScrollNext))
	}
	if e.AttribIndex("aria-label") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-label", "Next slide"))
	}
	e.Children = append(e.Children, iconChevronRight())
	return finish(e, "carousel-next",
		ButtonClasses(ButtonOutline, SizeIcon)+" absolute top-1/2 -right-12 size-8 -translate-y-1/2 rounded-full")
}
