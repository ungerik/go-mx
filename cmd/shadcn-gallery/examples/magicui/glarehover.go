package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/shadcn/magicui"
	"github.com/ungerik/go-mx/svg"
)

// glareHoverIcon is a lucide icon drawn from its path data, with extra
// classes.
func glareHoverIcon(class string, paths ...string) mx.Component {
	icon := svg.SVG(svg.XMLNS, svg.Width(24), svg.Height(24), svg.ViewBox(0, 0, 24, 24),
		svg.Fill("none"), svg.Stroke("currentColor"), svg.StrokeWidth(2),
		svg.StrokeLineCapRound, svg.StrokeLineJoinRound, glareHoverAriaHidden)
	if class != "" {
		icon.Attribs = append(icon.Attribs, svg.Class(class))
	}
	for _, d := range paths {
		icon.Children = append(icon.Children, svg.Path(svg.D(d)))
	}
	return icon
}

// glareHoverAriaHidden hides the decorative icons of the Glare Hover examples.
var glareHoverAriaHidden = mx.NewAttrib("aria-hidden", "true")

// GlareHoverDemo is a pricing card a glare sweeps across on hover.
func GlareHoverDemo() mx.Component {
	feature := func(text string) mx.Component {
		return html.DivClass("flex items-center gap-2 text-sm",
			svg.SVG(svg.Width(15), svg.Height(15), svg.ViewBox(0, 0, 15, 15), svg.Fill("none"),
				svg.Path(svg.D("M12.5 3.5L6 10L2.5 6.5"), svg.Stroke("currentColor"), svg.StrokeWidth(1.5),
					svg.StrokeLineCapRound, svg.StrokeLineJoinRound),
			),
			text,
		)
	}
	return magicui.GlareHover(magicui.GlareHoverOptions{Duration: 600 * time.Millisecond}, html.Class("rounded-xl"),
		shadcn.Card(html.Class("w-[340px]"),
			shadcn.CardHeader(
				html.DivClass("flex items-center justify-between",
					shadcn.CardTitle("Pro"),
					shadcn.Badge("", "Popular"),
				),
				shadcn.CardDescription("For teams that need more."),
				html.DivClass("flex items-baseline gap-1 pt-2",
					html.Span(html.Class("text-4xl font-semibold tracking-tight"), "$49"),
					html.Span(html.Class("text-muted-foreground text-sm"), "/mo"),
				),
			),
			shadcn.CardContent(html.Class("flex flex-col gap-2.5"),
				feature("Unlimited projects"),
				feature("Team collaboration"),
				feature("Advanced analytics"),
				html.DivClass("text-muted-foreground flex items-center gap-2 text-sm",
					svg.SVG(svg.Width(15), svg.Height(15), svg.ViewBox(0, 0, 15, 15), svg.Fill("none"),
						svg.Circle(svg.CX(7.5), svg.CY(7.5), svg.R(1.5), svg.Fill("currentColor"), svg.Opacity(0.4)),
					),
					"SSO (coming soon)",
				),
			),
			shadcn.CardFooter(
				shadcn.Button("", "", html.Class("w-full"), "Get started"),
			),
		),
	)
}

// GlareHoverCTA is a call-to-action card with a gray glare.
func GlareHoverCTA() mx.Component {
	return magicui.GlareHover(magicui.GlareHoverOptions{Color: "#505050", Duration: 700 * time.Millisecond}, html.Class("rounded-xl"),
		shadcn.Card(html.Class("w-full text-center"),
			shadcn.CardHeader(
				html.PClass("text-muted-foreground text-xs font-medium tracking-widest uppercase", "14-day free trial · No card required"),
				shadcn.CardTitle(html.Class("text-2xl tracking-tight"), "Ready to get started?"),
				shadcn.CardDescription(html.Class("mx-auto max-w-sm"), "Join 4,000+ teams already using our platform to ship faster."),
			),
			shadcn.CardFooter(html.Class("justify-center gap-2"),
				shadcn.Button("", "", "Start free trial"),
				shadcn.Button(shadcn.ButtonOutline, "", "View demo"),
			),
		),
	)
}

// GlareHoverAlerts are three alerts with glares in their own colors.
func GlareHoverAlerts() mx.Component {
	opts := func(color string) magicui.GlareHoverOptions {
		return magicui.GlareHoverOptions{Color: color, Opacity: 0.25, Duration: 550 * time.Millisecond}
	}
	return html.DivClass("flex flex-col gap-3",
		magicui.GlareHover(opts(""), html.Class("rounded-lg"),
			shadcn.Alert("", html.Class("[&>svg~*]:pl-10"),
				glareHoverIcon("", "M2 12a10 10 0 1 0 20 0a10 10 0 1 0-20 0", "M12 16v-4", "M12 8h.01"),
				shadcn.AlertTitle("Action required"),
				shadcn.AlertDescription("Your plan expires in 3 days. Renew to keep access."),
			),
		),
		magicui.GlareHover(opts("#ff6070"), html.Class("w-full rounded-lg"),
			shadcn.Alert(shadcn.AlertDestructive, html.Class("bg-background [&>svg~*]:pl-10"),
				glareHoverIcon("", "m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3", "M12 9v4", "M12 17h.01"),
				shadcn.AlertTitle("Payment declined"),
				shadcn.AlertDescription("Check your card details and try again."),
			),
		),
		magicui.GlareHover(opts("#60ff70"), html.Class("w-full rounded-lg"),
			shadcn.Alert("", html.Class("[&>svg~*]:pl-10"),
				glareHoverIcon("stroke-emerald-500", "M21.801 10A10 10 0 1 1 17 3.335", "m9 11 3 3L22 4"),
				shadcn.AlertTitle(html.Class("text-emerald-500"), "Subscription confirmed"),
				shadcn.AlertDescription(html.Class("text-emerald-500"), "You have full Pro access until April 5, 2027."),
			),
		),
	)
}
