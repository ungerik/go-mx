package magicui

import (
	"math"
	"strconv"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// WarpBackgroundOptions configures a [WarpBackground]. The zero value is
// upstream's default: three beams per side on a 5% grid, each crossing in
// 3 seconds after a delay of up to 3 seconds, in a 100 pixel perspective.
type WarpBackgroundOptions struct {
	// Perspective is the CSS perspective of the tunnel in pixels; 0 means
	// 100.
	Perspective float64
	// BeamsPerSide is the number of beams on each of the four sides; 0
	// means 3.
	BeamsPerSide int
	// BeamSize is the size of a grid cell, and the width of a beam, in
	// percent of a side; 0 means 5.
	BeamSize float64
	// BeamDelayMax is the longest delay before a beam starts; 0 means 3s.
	BeamDelayMax time.Duration
	// BeamDelayMin is the shortest delay before a beam starts.
	BeamDelayMin time.Duration
	// BeamDuration is how long a beam takes to cross a side; 0 means 3s.
	BeamDuration time.Duration
	// GridColor is the CSS color of the grid lines; empty means
	// "var(--border)".
	GridColor string
	// Seed seeds the random delays, hues and lengths of the beams; the same
	// seed renders the same markup.
	Seed uint64
}

// warpBackgroundSides holds the classes of one side of the tunnel, transcribed
// from upstream.
var warpBackgroundSides = []string{
	// top
	"@container absolute z-20 h-[100cqmax] w-[100cqi] origin-[50%_0%] transform-[rotateX(-90deg)] bg-size-[var(--beam-size)_var(--beam-size)] [background:linear-gradient(var(--grid-color)_0_1px,transparent_1px_var(--beam-size))_50%_-0.5px_/var(--beam-size)_var(--beam-size),linear-gradient(90deg,var(--grid-color)_0_1px,transparent_1px_var(--beam-size))_50%_50%_/var(--beam-size)_var(--beam-size)] transform-3d",
	// bottom
	"@container absolute top-full h-[100cqmax] w-[100cqi] origin-[50%_0%] transform-[rotateX(-90deg)] bg-size-[var(--beam-size)_var(--beam-size)] [background:linear-gradient(var(--grid-color)_0_1px,transparent_1px_var(--beam-size))_50%_-0.5px_/var(--beam-size)_var(--beam-size),linear-gradient(90deg,var(--grid-color)_0_1px,transparent_1px_var(--beam-size))_50%_50%_/var(--beam-size)_var(--beam-size)] transform-3d",
	// left
	"@container absolute top-0 left-0 h-[100cqmax] w-[100cqh] origin-[0%_0%] transform-[rotate(90deg)_rotateX(-90deg)] bg-size-[var(--beam-size)_var(--beam-size)] [background:linear-gradient(var(--grid-color)_0_1px,transparent_1px_var(--beam-size))_50%_-0.5px_/var(--beam-size)_var(--beam-size),linear-gradient(90deg,var(--grid-color)_0_1px,transparent_1px_var(--beam-size))_50%_50%_/var(--beam-size)_var(--beam-size)] transform-3d",
	// right
	"@container absolute top-0 right-0 h-[100cqmax] w-[100cqh] origin-[100%_0%] transform-[rotate(-90deg)_rotateX(-90deg)] bg-size-[var(--beam-size)_var(--beam-size)] [background:linear-gradient(var(--grid-color)_0_1px,transparent_1px_var(--beam-size))_50%_-0.5px_/var(--beam-size)_var(--beam-size),linear-gradient(90deg,var(--grid-color)_0_1px,transparent_1px_var(--beam-size))_50%_50%_/var(--beam-size)_var(--beam-size)] transform-3d",
}

// WarpBackground frames its children in a 3D tunnel: the four sides of
// its padded box are grids receding into the distance, along which
// colored beams of light fly outwards. The tunnel is aria-hidden; the
// children are placed in front of it.
//
// The beams' motion-library animation is re-expressed as a CSS animation;
// their random delays, hues and lengths are drawn on the server from
// opts.Seed.
//
// Needs [CSS] in the app's Tailwind input.
func WarpBackground(opts WarpBackgroundOptions, attribsChildren ...any) *mx.Element {
	beamSize := or(opts.BeamSize, 5)
	style, err := styleVars(
		"--perspective", px(or(opts.Perspective, 100)),
		"--grid-color", or(opts.GridColor, "var(--border)"),
		"--beam-size", num(beamSize)+"%",
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	beamsPerSide := or(opts.BeamsPerSide, 3)
	delayMax := or(opts.BeamDelayMax, 3*time.Second).Seconds()
	delayMin := opts.BeamDelayMin.Seconds()
	duration := seconds(opts.BeamDuration, 3*time.Second)
	rnd := newRand(opts.Seed)

	tunnel := html.Div(ariaHidden, html.Style(style), html.Class(
		"@container-[size] pointer-events-none absolute top-0 left-0 size-full overflow-hidden [clipPath:inset(0)] perspective-(--perspective) transform-3d"))
	cellsPerSide := math.Floor(100 / beamSize)
	step := cellsPerSide / float64(beamsPerSide)
	for _, sideClasses := range warpBackgroundSides {
		side := html.DivClass(sideClasses)
		for i := range beamsPerSide {
			x := math.Floor(float64(i) * step)
			delay := rnd.Float64()*(delayMax-delayMin) + delayMin
			hue := math.Floor(rnd.Float64() * 360)
			aspectRatio := math.Floor(rnd.Float64()*10) + 1
			side.Children = append(side.Children, html.Div(
				html.Class("animate-warp-background-beam absolute top-0 left-(--x) aspect-[1/var(--aspect-ratio)] w-(--width) [background:var(--background)]"),
				html.Style("--x:"+num(x*beamSize)+"%"+
					";--width:"+num(beamSize)+"%"+
					";--aspect-ratio:"+strconv.Itoa(int(aspectRatio))+
					";--background:linear-gradient(hsl("+num(hue)+" 80% 60%), transparent)"+
					// The resting position is the animation's end, out of view, so no
					// beam is left standing when the animation is off.
					";translate:-50% -100%"+
					";animation-duration:"+duration+
					";animation-delay:"+secondsRounded(delay)),
			))
		}
		tunnel.Children = append(tunnel.Children, side)
	}

	root := html.Div(attribsChildren...)
	children := root.Children
	root.Children = []mx.Component{tunnel, html.Div(append([]any{html.Class("relative")}, anys(children)...)...)}
	return finish(root, "warp-background", "relative rounded border p-20", "")
}
