package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/shadcn/magicui"
	"github.com/ungerik/go-mx/svg"
)

// borderBeamLoginCard is the login card of upstream's border-beam demos, with
// beam as its last child; prefix keeps the input ids unique on the page.
func borderBeamLoginCard(prefix string, beam mx.Component) mx.Component {
	return shadcn.Card(html.Class("relative w-[350px] overflow-hidden"),
		shadcn.CardHeader(
			shadcn.CardTitle("Login"),
			shadcn.CardDescription("Enter your credentials to access your account."),
		),
		shadcn.CardContent(
			html.Form(
				html.DivClass("grid w-full items-center gap-4",
					html.DivClass("flex flex-col space-y-1.5",
						shadcn.LabelFor(prefix+"-email", "Email"),
						shadcn.InputID(prefix+"-email", html.Type("email"), html.Placeholder("Enter your email")),
					),
					html.DivClass("flex flex-col space-y-1.5",
						shadcn.LabelFor(prefix+"-password", "Password"),
						shadcn.InputID(prefix+"-password", html.Type("password"), html.Placeholder("Enter your password")),
					),
				),
			),
		),
		shadcn.CardFooter(html.Class("flex justify-between"),
			shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault, "Register"),
			shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Login"),
		),
		beam,
	)
}

// BorderBeamDemo runs the default beam around a login card.
func BorderBeamDemo() mx.Component {
	return borderBeamLoginCard("border-beam-demo",
		magicui.BorderBeam(magicui.BorderBeamOptions{Duration: 8 * time.Second, Size: 100}),
	)
}

// BorderBeamReverse runs a green beam counter-clockwise around a login card.
func BorderBeamReverse() mx.Component {
	return borderBeamLoginCard("border-beam-reverse",
		magicui.BorderBeam(magicui.BorderBeamOptions{Duration: 4 * time.Second, Size: 300, Reverse: true},
			html.Class("from-transparent via-green-500 to-transparent")),
	)
}

// BorderBeamMusicPlayer runs a red and a blue beam, half a circle apart,
// around a music player card.
func BorderBeamMusicPlayer() mx.Component {
	return shadcn.Card(html.Class("relative w-[350px] overflow-hidden"),
		shadcn.CardHeader(
			shadcn.CardTitle("Now Playing"),
			shadcn.CardDescription("Stairway to Heaven - Led Zeppelin"),
		),
		shadcn.CardContent(
			html.DivClass("flex flex-col items-center gap-4",
				html.DivClass("h-48 w-48 rounded-lg bg-linear-to-br from-purple-500 to-pink-500"),
				html.DivClass("bg-secondary h-1 w-full rounded-full",
					html.DivClass("bg-primary h-full w-1/3 rounded-full"),
				),
				html.DivClass("text-muted-foreground flex w-full justify-between text-sm",
					html.Span("2:45"),
					html.Span("8:02"),
				),
			),
		),
		shadcn.CardFooter(html.Class("flex justify-center gap-4"),
			shadcn.Button(shadcn.ButtonOutline, shadcn.SizeIcon, html.Class("rounded-full"), mx.NewAttrib("aria-label", "Previous"),
				playerIcon("M17.971 4.285A2 2 0 0 1 21 6v12a2 2 0 0 1-3.029 1.715l-9.997-5.998a2 2 0 0 1-.003-3.432z", "M3 20V4")),
			shadcn.Button(shadcn.ButtonDefault, shadcn.SizeIcon, html.Class("rounded-full"), mx.NewAttrib("aria-label", "Play"),
				playerIcon("M5 5a2 2 0 0 1 3.008-1.728l11.997 6.998a2 2 0 0 1 .003 3.458l-12 7A2 2 0 0 1 5 19z")),
			shadcn.Button(shadcn.ButtonOutline, shadcn.SizeIcon, html.Class("rounded-full"), mx.NewAttrib("aria-label", "Next"),
				playerIcon("M21 4v16", "M6.029 4.285A2 2 0 0 0 3 6v12a2 2 0 0 0 3.029 1.715l9.997-5.998a2 2 0 0 0 .003-3.432z")),
		),
		magicui.BorderBeam(magicui.BorderBeamOptions{Duration: 6 * time.Second, Size: 400},
			html.Class("from-transparent via-red-500 to-transparent")),
		magicui.BorderBeam(magicui.BorderBeamOptions{Duration: 6 * time.Second, Delay: 3 * time.Second, Size: 400, BorderWidth: 2},
			html.Class("from-transparent via-blue-500 to-transparent")),
	)
}

// playerIcon builds a small lucide-style inline SVG from its path data.
func playerIcon(paths ...string) mx.Component {
	shapes := []any{
		svg.XMLNS, svg.Width(16), svg.Height(16),
		svg.ViewBox(0, 0, 24, 24), svg.Fill("none"), svg.Stroke("currentColor"),
		svg.StrokeWidth(2), svg.StrokeLineCapRound, svg.StrokeLineJoinRound,
	}
	for _, d := range paths {
		shapes = append(shapes, svg.Path(svg.D(d)))
	}
	return svg.SVG(shapes...)
}
