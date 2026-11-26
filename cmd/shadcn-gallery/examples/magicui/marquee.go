package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// marqueeReview is one testimonial of the marquee examples.
type marqueeReview struct {
	Name, Username, Body, Img string
}

var marqueeReviews = []marqueeReview{
	{"Jack", "@jack", "I've never seen anything like this before. It's amazing. I love it.", "https://avatar.vercel.sh/jack"},
	{"Jill", "@jill", "I don't know what to say. I'm speechless. This is amazing.", "https://avatar.vercel.sh/jill"},
	{"John", "@john", "I'm at a loss for words. This is amazing. I love it.", "https://avatar.vercel.sh/john"},
	{"Jane", "@jane", "I'm at a loss for words. This is amazing. I love it.", "https://avatar.vercel.sh/jane"},
	{"Jenny", "@jenny", "I'm at a loss for words. This is amazing. I love it.", "https://avatar.vercel.sh/jenny"},
	{"James", "@james", "I'm at a loss for words. This is amazing. I love it.", "https://avatar.vercel.sh/james"},
}

// marqueeReviewCard renders one testimonial as a card.
func marqueeReviewCard(r marqueeReview) mx.Component {
	return html.Figure(
		html.Class("relative h-full w-64 cursor-pointer overflow-hidden rounded-xl border p-4",
			"border-gray-950/[.1] bg-gray-950/[.01] hover:bg-gray-950/[.05]",
			"dark:border-gray-50/[.1] dark:bg-gray-50/[.10] dark:hover:bg-gray-50/[.15]"),
		html.DivClass("flex flex-row items-center gap-2",
			html.ImgSrc(r.Img, html.Class("rounded-full"), html.Width(32), html.Height(32), html.Alt("")),
			html.DivClass("flex flex-col",
				html.FigCaption(html.Class("text-sm font-medium dark:text-white"), r.Name),
				html.PClass("text-xs font-medium dark:text-white/40", r.Username),
			),
		),
		html.Blockquote(html.Class("mt-2 text-sm"), r.Body),
	)
}

// MarqueeDemo scrolls two rows of testimonials in opposite directions,
// pausing on hover, with the edges faded into the background.
func MarqueeDemo() mx.Component {
	half := len(marqueeReviews) / 2
	return html.DivClass("relative flex w-full flex-col items-center justify-center overflow-hidden",
		magicui.Marquee(magicui.MarqueeOptions{PauseOnHover: true}, html.Class("[--duration:20s]"),
			mx.ForEach(marqueeReviews[:half], marqueeReviewCard),
		),
		magicui.Marquee(magicui.MarqueeOptions{Reverse: true, PauseOnHover: true}, html.Class("[--duration:20s]"),
			mx.ForEach(marqueeReviews[half:], marqueeReviewCard),
		),
		html.DivClass("from-background pointer-events-none absolute inset-y-0 left-0 w-1/4 bg-gradient-to-r"),
		html.DivClass("from-background pointer-events-none absolute inset-y-0 right-0 w-1/4 bg-gradient-to-l"),
	)
}

// MarqueeVertical scrolls two columns of testimonials vertically.
func MarqueeVertical() mx.Component {
	half := len(marqueeReviews) / 2
	return html.DivClass("relative flex h-[500px] w-full flex-row items-center justify-center overflow-hidden",
		magicui.Marquee(magicui.MarqueeOptions{PauseOnHover: true, Vertical: true}, html.Class("[--duration:20s]"),
			mx.ForEach(marqueeReviews[:half], marqueeReviewCard),
		),
		magicui.Marquee(magicui.MarqueeOptions{Reverse: true, PauseOnHover: true, Vertical: true}, html.Class("[--duration:20s]"),
			mx.ForEach(marqueeReviews[half:], marqueeReviewCard),
		),
		html.DivClass("from-background pointer-events-none absolute inset-x-0 top-0 h-1/4 bg-gradient-to-b"),
		html.DivClass("from-background pointer-events-none absolute inset-x-0 bottom-0 h-1/4 bg-gradient-to-t"),
	)
}
