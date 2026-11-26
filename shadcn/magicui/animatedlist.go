package magicui

import (
	"strconv"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AnimatedListOptions configures an [AnimatedList]. The zero value is
// upstream's default: one new item every second.
type AnimatedListOptions struct {
	// Delay is the time between two items appearing; 0 means 1s.
	Delay time.Duration
}

// AnimatedList shows its children one after another, newest on top: every
// opts.Delay the next child springs in at the top of the list from scale 0
// while the items below make room for it. Each child is one item; an
// mx.Components child (as from mx.ForEach) gives one item per component.
// The gap between the items is the CSS variable --animated-list-gap
// (1rem), which a caller class such as html.Class("[--animated-list-gap:0.5rem]")
// overrides; the items that have not appeared yet take no space, which
// needs the variable to match the gap.
//
// Upstream adds the items to the DOM with React state and animates them
// with the motion library's AnimatePresence and layout animations; here
// all items are rendered at once, in reverse order, and each one's CSS
// entrance (growing grid row, scale and fade, eased like upstream's spring
// with stiffness 350 and damping 40) waits Delay times its index. Under
// prefers-reduced-motion the whole list shows at once. Children added
// later are not animated in, since there is no script.
//
// Needs [CSS] in the app's Tailwind input.
func AnimatedList(opts AnimatedListOptions, attribsChildren ...any) *mx.Element {
	root := html.Div(attribsChildren...)
	if root.Err != nil {
		return root
	}
	style, err := styleVars("--animated-list-delay", seconds(opts.Delay, time.Second))
	if err != nil {
		return mx.NewErrElement(err)
	}
	items := animatedListItems(root.Children, nil)
	root.Children = make([]mx.Component, len(items))
	for i, item := range items {
		root.Children[len(items)-1-i] = html.Div(
			html.Class("mx-auto grid w-full grid-rows-[1fr] origin-top animate-animated-list-item"),
			html.Style("--i:"+strconv.Itoa(i)),
			html.Div(html.Class("min-h-0"), item),
		)
	}
	return finish(root, "animated-list", "flex flex-col items-center gap-(--animated-list-gap) [--animated-list-gap:1rem]", style)
}

// animatedListItems appends the items of an [AnimatedList] or
// [OrbitingCircles] to items, flattening mx.Components.
func animatedListItems(children mx.Components, items []mx.Component) []mx.Component {
	for _, child := range children {
		if group, ok := child.(mx.Components); ok {
			items = animatedListItems(group, items)
			continue
		}
		items = append(items, child)
	}
	return items
}
