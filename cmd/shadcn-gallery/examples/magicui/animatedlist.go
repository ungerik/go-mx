package magicui

import (
	"slices"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// animatedListItem is one notification of the animated list example.
type animatedListItem struct {
	Name, Description, Icon, Color, Time string
}

var animatedListItems = []animatedListItem{
	{"Payment received", "Magic UI", "💸", "#00C9A7", "15m ago"},
	{"User signed up", "Magic UI", "👤", "#FFB800", "10m ago"},
	{"New message", "Magic UI", "💬", "#FF3D71", "5m ago"},
	{"New event", "Magic UI", "🗞️", "#1E86FF", "2m ago"},
}

// animatedListNotification renders one notification as a card.
func animatedListNotification(item animatedListItem) mx.Component {
	return html.Figure(
		html.Class("relative mx-auto min-h-fit w-full max-w-[400px] cursor-pointer overflow-hidden rounded-2xl p-4",
			// animation styles
			"transition-all duration-200 ease-in-out hover:scale-[103%]",
			// light styles
			"bg-white [box-shadow:0_0_0_1px_rgba(0,0,0,.03),0_2px_4px_rgba(0,0,0,.05),0_12px_24px_rgba(0,0,0,.05)]",
			// dark styles
			"transform-gpu dark:bg-transparent dark:[box-shadow:0_-20px_80px_-20px_#ffffff1f_inset] dark:backdrop-blur-md dark:[border:1px_solid_rgba(255,255,255,.1)]"),
		html.DivClass("flex flex-row items-center gap-3",
			html.Div(html.Class("flex size-10 items-center justify-center rounded-2xl"), html.Style("background-color:"+item.Color),
				html.SpanClass("text-lg", item.Icon),
			),
			html.DivClass("flex flex-col overflow-hidden",
				html.FigCaption(html.Class("flex flex-row items-center text-lg font-medium whitespace-pre dark:text-white"),
					html.SpanClass("text-sm sm:text-lg", item.Name),
					html.SpanClass("mx-1", "·"),
					html.SpanClass("text-xs text-gray-500", item.Time),
				),
				html.PClass("text-sm font-normal dark:text-white/60", item.Description),
			),
		),
	)
}

// AnimatedListDemo lets forty notifications pop in one per second, newest
// on top.
func AnimatedListDemo() mx.Component {
	var notifications []animatedListItem
	for range 10 {
		notifications = slices.Concat(notifications, animatedListItems)
	}
	return html.DivClass("relative flex h-[500px] w-full flex-col overflow-hidden p-2",
		magicui.AnimatedList(magicui.AnimatedListOptions{}, mx.ForEach(notifications, animatedListNotification)),
		html.DivClass("from-background pointer-events-none absolute inset-x-0 bottom-0 h-1/4 bg-gradient-to-t"),
	)
}
