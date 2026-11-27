package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// NoiseTextureDemo is a grainy surface faded out towards the edges.
func NoiseTextureDemo() mx.Component {
	return html.DivClass("relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border bg-neutral-100/80 dark:bg-neutral-950",
		magicui.NoiseTexture(magicui.NoiseTextureOptions{ID: "noise-texture-demo"}, html.Class(
			"absolute inset-0",
			"mask-[radial-gradient(420px_circle_at_center,white,transparent)]",
		)),
	)
}

// NoiseTextureCard is a newsletter sign-up card with a grainy background.
func NoiseTextureCard() mx.Component {
	return html.DivClass("flex justify-center p-4",
		shadcn.Card(html.Class("bg-card/80 relative w-full max-w-md overflow-hidden border"),
			magicui.NoiseTexture(magicui.NoiseTextureOptions{ID: "noise-texture-card", NoiseOpacity: 0.45}),
			shadcn.CardHeader(html.Class("relative z-10 space-y-1 pb-4"),
				shadcn.CardTitle(html.Class("text-xl"), "The weekly digest"),
				shadcn.CardDescription("One email on Sundays—new components, tips, and changelog highlights. No spam, unsubscribe anytime."),
			),
			shadcn.CardContent(html.Class("relative z-10 space-y-4 pt-0"),
				html.DivClass("space-y-2",
					shadcn.LabelFor("noise-texture-email", "Email"),
					shadcn.InputID("noise-texture-email", html.AutoComplete("email"), html.Placeholder("you@company.com"), html.Type("email")),
				),
				shadcn.Button("", "", html.Class("w-full"), html.Type("button"), "Subscribe"),
			),
		),
	)
}

// NoiseTextureButton is a secondary button with a grainy surface that
// darkens on hover.
func NoiseTextureButton() mx.Component {
	return html.DivClass("flex justify-center p-8",
		shadcn.Button(shadcn.ButtonSecondary, "",
			html.Class("group/button relative cursor-pointer overflow-hidden px-8 active:scale-98"), html.Type("button"),
			html.Span(html.Class("relative z-10 font-medium"), "Subscribe"),
			magicui.NoiseTexture(magicui.NoiseTextureOptions{ID: "noise-texture-button", NoiseOpacity: 0.45},
				html.Class("transition-all group-hover/button:opacity-100")),
		),
	)
}
