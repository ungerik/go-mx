package logview

import (
	"context"
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
	"github.com/ungerik/go-mx/shadcn"
)

// View renders a streaming log surface: a toolbar with a filter and a pause
// toggle, a scroll area that follows the stream, and a sink for an in-band
// stream error.
//
// Caller attributes land on the root element and caller children inside the
// line container as the backlog already on the page, so connecting the view to
// a stream is a matter of passing the htmx attributes:
//
//	logview.View(
//		hx.Ext("sse"),
//		hx.SSEConnect("/logs"),
//		hx.SSEClose("done"),
//		logview.Lines(backlog...),
//	)
//
// The page needs [Config.StyleElement] in its head, plus hx.ScriptFromCDN and
// hx.ScriptSSEFromCDN for the stream itself. The handler on the other end sends
// each line as [Config.Line] under the event name [Config.Event].
func (c *Config) View(attribsChildren ...any) *mx.Element {
	root := html.Div(attribsChildren...)
	backlog := root.Children

	lines := html.Div(
		html.Class(c.class(ClassLines)),
		mx.ConstAttrib(attrLines+"="),
		// A live region that announces additions only: without it a screen
		// reader either says nothing as the log grows or re-reads the whole
		// transcript on every line.
		html.Role("log"),
		html.Attrib("aria-live", "polite"),
		html.Attrib("aria-relevant", "additions"),
		// Appending rather than replacing is what makes this a log instead of
		// one line that keeps being overwritten.
		hx.SSESwap(c.EventName()),
		hx.Swap(hx.SwapBeforeEnd),
	)
	lines.Children = append(lines.Children, backlog...)

	labels := c.labels()
	toolbar := html.DivClass(c.class(ClassToolbar),
		html.Input(
			html.Class(c.class(ClassFilter)),
			html.Type("search"),
			mx.ConstAttrib(attrFilter+"="),
			html.Placeholder(labels.FilterHint),
			html.Attrib("aria-label", labels.Filter),
		),
		html.Button(
			html.Class(c.class(ClassPause)),
			html.Type("button"),
			// The attribute value is the caption the button swaps to while
			// paused, so the script carries no user-visible string of its own.
			//
			// A caption that names the next action rather than a toggle button
			// with aria-pressed: the two cannot be combined, because a pressed
			// button captioned "Resume" tells a screen reader the opposite of
			// what it tells the eye. The state a stylesheet needs is on the
			// root instead, as attrPaused.
			html.Attrib(attrPause, labels.Resume),
			labels.Pause,
		),
		// Pausing has to reach a screen reader from both entry points, and the
		// button's caption only covers the one that has focus: scrolling up
		// pauses too, silently. A status region says the state whichever way it
		// changed. It carries its own text rather than being filled by CSS,
		// because generated content is not reliably announced.
		html.SpanClass(c.class(ClassStatus), html.Attrib(attrStatus, labels.Paused), html.Role("status")),
		// The badge's own attribute value carries the word the script appends
		// to the count, so the script holds no user-visible string either. It
		// is hidden while empty by a stylesheet rule, the way the error sink is
		// — not by the hidden property, whose UA-stylesheet rule any author
		// display declaration beats.
		html.SpanClass(c.class(ClassBadge), html.Attrib(attrBadge, labels.NewLines)),
	)

	// The height and the overflow are both load-bearing, and both are here rather
	// than left to the shadcn classes because those are Tailwind and a page
	// without that build would get neither: with no height the area grows
	// instead of scrolling, and with no overflow it is clipped by the view
	// instead of scrolling. Either way scrollHeight equals clientHeight, so
	// stick-to-bottom has nothing to follow and the log silently stops moving.
	//
	// flex makes the height a preference: in the flex column the view is, a page
	// that gives the view a height of its own has the area fill what is left
	// instead. min-height: 0 is what lets it shrink below its content and
	// actually scroll.
	area := shadcn.ScrollArea(
		html.Style("flex: 1 1 auto; min-height: 0; height: "+c.height()+"; overflow: auto"),
		shadcn.StickToBottom,
		lines,
	)

	// Once the stream's 200 is committed a failure can only travel in band, so
	// mx.SSEEventError needs somewhere to land or the log just stops.
	errorSink := html.Div(
		html.Class(c.class(ClassError)),
		html.Role("alert"),
		hx.SSESwap(mx.SSEEventError),
	)

	mergeClass(root, c.class(ClassView))
	root.Attribs = append(root.Attribs,
		mx.ConstAttrib(attrView+"="),
		html.Attrib(attrMaxLines, strconv.Itoa(c.maxLines())),
	)
	// The script comes last so the view is fully parsed when it wires up.
	root.Children = mx.Components{toolbar, area, errorSink, html.ScriptJS(logViewScript)}
	return root
}

// mergeClass prepends classes to the element's class attribute, merging with
// one the caller supplied. Appending a second class attribute instead would
// make the render fail on the duplicate-attribute check, so a caller could not
// pass a class at all.
func mergeClass(e *mx.Element, classes string) {
	for i, a := range e.Attribs {
		if a.AttribName() != "class" {
			continue
		}
		// Resolving without the render context matches how shadcn's finish
		// merges caller classes: a class value that depends on the request is
		// not something either package supports.
		value, err := a.AttribValue(context.Background())
		if err != nil {
			e.Attribs[i] = mx.ErrAttrib{Name: "class", Err: err}
			return
		}
		e.Attribs[i] = mx.NewAttrib("class", classes+" "+value)
		return
	}
	e.Attribs = append(e.Attribs, mx.NewAttrib("class", classes))
}
