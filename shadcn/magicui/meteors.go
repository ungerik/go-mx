package magicui

import (
	"math"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// MeteorsOptions configures a [Meteors]. The zero value is upstream's
// default: 20 meteors at an angle of 215 degrees, starting after 0.2 to
// 1.2 seconds and falling for 2 to 10 seconds.
type MeteorsOptions struct {
	// Number is the number of meteors; 0 means 20.
	Number int
	// MinDelay is the shortest delay before a meteor starts; 0 means 200ms.
	MinDelay time.Duration
	// MaxDelay is the longest delay before a meteor starts; 0 means 1.2s.
	MaxDelay time.Duration
	// MinDuration is the shortest time a meteor takes to fall; 0 means 2s.
	MinDuration time.Duration
	// MaxDuration is the longest time a meteor takes to fall; 0 means 10s.
	MaxDuration time.Duration
	// Angle is the direction of the meteors' fall in degrees; 0 means 215.
	// Pass 360 for 0 degrees.
	Angle float64
	// Seed seeds the random positions, delays and durations of the meteors;
	// the same seed renders the same markup.
	Seed uint64
}

// Meteors rains meteors with fading tails diagonally across its positioned
// parent, each starting at a random position with a random delay and
// duration. Put it into a parent with overflow-hidden.
//
// Upstream renders the meteors without a wrapper and applies its className
// to every meteor; here they sit in an absolutely positioned, aria-hidden
// <div> that takes the caller's attributes, so restyle the meteors through
// it with a child selector such as html.Class("[&>span]:bg-sky-500").
// Upstream places the meteors at random pixel offsets up to the window
// width, measured in the browser; this port places them at random
// percentages of the parent's width, drawn on the server from opts.Seed.
//
// Needs [CSS] in the app's Tailwind input.
func Meteors(opts MeteorsOptions, attribsChildren ...any) *mx.Element {
	minDelay := meteorsSeconds(opts.MinDelay, 200*time.Millisecond)
	maxDelay := meteorsSeconds(opts.MaxDelay, 1200*time.Millisecond)
	minDuration := meteorsSeconds(opts.MinDuration, 2*time.Second)
	maxDuration := meteorsSeconds(opts.MaxDuration, 10*time.Second)
	angle := num(-or(opts.Angle, 215)) + "deg"
	rnd := newRand(opts.Seed)

	root := html.Div(append([]any{ariaHidden}, attribsChildren...)...)
	for range or(opts.Number, 20) {
		left := math.Floor(rnd.Float64()*1000) / 10
		delay := rnd.Float64()*(maxDelay-minDelay) + minDelay
		duration := math.Floor(rnd.Float64()*(maxDuration-minDuration) + minDuration)
		root.Children = append(root.Children, html.Span(
			html.Class("animate-meteor pointer-events-none absolute size-0.5 rotate-(--angle) rounded-full bg-zinc-500 shadow-[0_0_0_1px_#ffffff10]"),
			html.Style("--angle:"+angle+
				";top:-5%"+
				";left:"+num(left)+"%"+
				";animation-delay:"+secondsRounded(delay)+
				";animation-duration:"+num(duration)+"s"),
			html.DivClass("pointer-events-none absolute top-1/2 -z-10 h-px w-12.5 -translate-y-1/2 bg-linear-to-r from-zinc-500 to-transparent"),
		))
	}
	return finish(root, "meteors", "pointer-events-none absolute inset-0", "")
}

// meteorsSeconds returns d in seconds, or def if d is not positive.
func meteorsSeconds(d, def time.Duration) float64 {
	if d <= 0 {
		d = def
	}
	return d.Seconds()
}
