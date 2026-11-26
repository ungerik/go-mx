package magicui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// WordRotateOptions configures a [WordRotate].
type WordRotateOptions struct {
	// Words are shown one after another, starting over after the last.
	Words []string
	// Duration is how long each word stays; 0 means 2500ms.
	Duration time.Duration
}

// WordRotate shows opts.Words one at a time in an h1: every Duration the
// current word slides down and fades out, then the next one slides in from
// above. The attribsChildren go to the h1, as upstream's className; it
// takes no children, a child makes it render an error.
//
// Upstream swaps the word with React state and the motion library's
// AnimatePresence; here all words are stacked in one grid cell, each
// running the same CSS cycle of len(Words)*Duration with its own delay,
// and the cycle's easing is a linear() function computed from the options
// (browsers without linear() show the first word only). The stack keeps
// the width of the widest word, where upstream's width changes with the
// word. prefers-reduced-motion shows the first word. Assistive technology
// reads the words once, from an sr-only copy. Not ported: the motionProps
// prop.
//
// Needs [CSS] in the app's Tailwind input.
func WordRotate(opts WordRotateOptions, attribsChildren ...any) *mx.Element {
	h1 := html.H1(attribsChildren...)
	if h1.Err != nil {
		return h1
	}
	if len(h1.Children) > 0 {
		return mx.NewErrElement(fmt.Errorf("magicui: WordRotate takes its words from WordRotateOptions.Words, not from children"))
	}
	n := len(opts.Words)
	step := opts.Duration
	if step <= 0 {
		step = 2500 * time.Millisecond
	}
	lead := min(250*time.Millisecond, step/4)
	var style string
	if n > 1 {
		var err error
		style, err = styleVars(
			"--word-rotate-cycle", seconds(time.Duration(n)*step, 0),
			"--word-rotate-step", seconds(step, 0),
			"--word-rotate-lead", seconds(lead, 0),
			"--word-rotate-fade-easing", wordRotateEasing(n, step, lead, false),
			"--word-rotate-slide-easing", wordRotateEasing(n, step, lead, true),
		)
		if err != nil {
			return mx.NewErrElement(err)
		}
	}
	stack := []any{html.Class("grid"), ariaHidden}
	for i, word := range opts.Words {
		class := "col-start-1 row-start-1"
		if n > 1 {
			class += " animate-word-rotate"
		}
		if i > 0 {
			class += " opacity-0"
		}
		stack = append(stack, html.Span(html.Class(class), html.Style("--i:"+strconv.Itoa(i)), word))
	}
	h1.Children = []mx.Component{
		html.Span(html.Class("sr-only"), strings.Join(opts.Words, " ")),
		html.Span(stack...),
	}
	return finish(html.Div(finish(h1, "word-rotate-text", "", "")), "word-rotate", "overflow-hidden py-2", style)
}

// wordRotateEaseOut samples motion's easeOut, cubic-bezier(0, 0, 0.58, 1),
// at 0, 1/4, 1/2, 3/4 and 1.
var wordRotateEaseOut = [5]float64{0, 0.3781, 0.6846, 0.9065, 1}

// wordRotateEasing returns the linear() easing of one word's cycle of
// n*step, which starts lead before the word's turn: hidden for lead while
// the previous word leaves, entering for lead, shown until step, leaving
// for lead, hidden for the rest. The fade easing maps to the opacity 0..1
// of @keyframes word-rotate-fade, the slide easing to the -50px..50px of
// word-rotate-slide, with 0.5 the resting position.
func wordRotateEasing(n int, step, lead time.Duration, slide bool) string {
	cycle := float64(time.Duration(n) * step)
	var points []string
	point := func(value float64, at time.Duration) {
		pct := math.Round(float64(at)/cycle*100*1000) / 1000
		points = append(points, num(math.Round(value*10000)/10000)+" "+num(pct)+"%")
	}
	enter, leave := func(f float64) float64 { return f }, func(f float64) float64 { return 1 - f }
	if slide {
		enter = func(f float64) float64 { return f / 2 }
		leave = func(f float64) float64 { return 0.5 + f/2 }
	}
	point(enter(0), 0)
	for k, f := range wordRotateEaseOut {
		point(enter(f), lead+time.Duration(k)*lead/4)
	}
	for k, f := range wordRotateEaseOut {
		point(leave(f), step+time.Duration(k)*lead/4)
	}
	point(leave(1), time.Duration(cycle))
	return "linear(" + strings.Join(points, ", ") + ")"
}
