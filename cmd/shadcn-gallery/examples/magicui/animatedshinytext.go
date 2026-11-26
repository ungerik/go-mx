package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
	"github.com/ungerik/go-mx/svg"
)

// animatedShinyTextArrow is Radix's arrow-right icon with the given
// classes.
func animatedShinyTextArrow(class string) mx.Component {
	return svg.SVG(
		svg.XMLNS, svg.Width(15), svg.Height(15), svg.ViewBox(0, 0, 15, 15),
		svg.Fill("none"), svg.Class(class),
		svg.Path(
			svg.D("M8.14645 3.14645C8.34171 2.95118 8.65829 2.95118 8.85355 3.14645L12.8536 7.14645C13.0488 7.34171 13.0488 7.65829 12.8536 7.85355L8.85355 11.8536C8.65829 12.0488 8.34171 12.0488 8.14645 11.8536C7.95118 11.6583 7.95118 11.3417 8.14645 11.1464L11.2929 8H2.5C2.22386 8 2 7.77614 2 7.5C2 7.22386 2.22386 7 2.5 7H11.2929L8.14645 3.85355C7.95118 3.65829 7.95118 3.34171 8.14645 3.14645Z"),
			svg.Fill("currentColor"), svg.FillRuleEvenodd, svg.ClipRuleEvenodd,
		),
	)
}

// AnimatedShinyTextDemo is an announcement pill whose text is swept by a
// glare.
func AnimatedShinyTextDemo() mx.Component {
	return html.DivClass("z-10 flex min-h-64 items-center justify-center",
		html.DivClass("group rounded-full border border-black/5 bg-neutral-100 text-base text-white transition-all ease-in hover:cursor-pointer hover:bg-neutral-200 dark:border-white/5 dark:bg-neutral-900 dark:hover:bg-neutral-800",
			magicui.AnimatedShinyText(magicui.AnimatedShinyTextOptions{},
				html.Class("inline-flex items-center justify-center px-4 py-1 transition ease-out hover:text-neutral-600 hover:duration-300 hover:dark:text-neutral-400"),
				html.Span("✨ Introducing Magic UI"),
				animatedShinyTextArrow("ml-1 size-3 transition-transform duration-300 ease-in-out group-hover:translate-x-0.5"),
			),
		),
	)
}
