package magicui

import (
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// TextReveal reveals its text word by word while the page scrolls: the
// element is 200vh tall, its text sticks to the top of the viewport, and
// each word turns from faint to solid during its share of the scrolling.
// The children must be strings, as upstream's children: string; another
// child makes it render an error. The text is split into words at
// whitespace.
//
// Upstream drives the words' opacity with the motion library's useScroll;
// here a scroll-driven animation does it without JavaScript: the element
// names a view timeline (--text-reveal) and word i of n animates through
// the range contain i/n to contain (i+1)/n of it. Browsers without
// scroll-driven animations, and prefers-reduced-motion, show every word
// solid. Like upstream, it needs to be scrolled by the page (no ancestor
// with overflow other than visible). The faint copy of every word is
// aria-hidden, so assistive technology reads the text once (upstream reads
// it twice).
//
// Needs [CSS] in the app's Tailwind input.
func TextReveal(attribsChildren ...any) *mx.Element {
	root := html.Div(attribsChildren...)
	if root.Err != nil {
		return root
	}
	text, err := childText("TextReveal", root)
	if err != nil {
		return mx.NewErrElement(err)
	}
	words := strings.Fields(text)
	spans := []any{html.Class("flex flex-wrap p-5 text-2xl font-bold text-black/20 md:p-8 md:text-3xl lg:p-10 lg:text-4xl xl:text-5xl dark:text-white/20")}
	for i, word := range words {
		spans = append(spans, html.Span(html.Class("xl:lg-3 relative mx-1 lg:mx-1.5"),
			html.Span(html.Class("absolute opacity-30"), ariaHidden, word),
			html.Span(html.Class("text-black dark:text-white animate-text-reveal-word"), html.Style("--i:"+strconv.Itoa(i)), word),
		))
	}
	root.Children = []mx.Component{
		html.Div(html.Class("sticky top-0 mx-auto flex h-[50%] max-w-4xl items-center bg-transparent px-4 py-20"),
			html.Span(spans...),
		),
	}
	return finish(root, "text-reveal", "relative z-0 h-[200vh] [view-timeline-name:--text-reveal]", "--n:"+strconv.Itoa(len(words)))
}
