package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// LightRaysDemo is a headline lit by soft rays of light from above.
func LightRaysDemo() mx.Component {
	return html.DivClass("relative h-[500px] w-full overflow-hidden rounded-lg border",
		html.DivClass("relative z-10 flex h-full flex-col items-center justify-center gap-4 px-6 text-center",
			html.Span(html.Class("text-xs font-semibold tracking-[0.35em] text-slate-800/60 uppercase dark:text-slate-200/60"), "Ambient glow"),
			html.H1(html.Class("text-foreground text-4xl font-bold md:text-5xl"), "Light Rays"),
			html.PClass("max-w-md text-sm text-slate-800/80 md:text-base dark:text-slate-200/80",
				"Drop this component into any container and it will fill the space with softly animated light rays shining from above."),
		),
		magicui.LightRays(magicui.LightRaysOptions{Seed: 1}),
	)
}
