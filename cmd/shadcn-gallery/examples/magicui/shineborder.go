package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// shineBorderLoginCard is the login card of the Shine Border examples.
// idPrefix keeps the input ids unique when both examples share a page.
func shineBorderLoginCard(idPrefix string, cardClass string, opts magicui.ShineBorderOptions) mx.Component {
	return shadcn.Card(html.Class(cardClass),
		magicui.ShineBorder(opts),
		shadcn.CardHeader(
			shadcn.CardTitle("Login"),
			shadcn.CardDescription("Enter your credentials to access your account"),
		),
		shadcn.CardContent(
			html.Form(
				html.DivClass("grid gap-4",
					html.DivClass("grid gap-2",
						shadcn.LabelFor(idPrefix+"-email", "Email"),
						shadcn.InputID(idPrefix+"-email", html.Type("email"), html.Placeholder("name@example.com")),
					),
					html.DivClass("grid gap-2",
						shadcn.LabelFor(idPrefix+"-password", "Password"),
						shadcn.InputID(idPrefix+"-password", html.Type("password")),
					),
				),
			),
		),
		shadcn.CardFooter(
			shadcn.Button("", "", html.Class("w-full"), "Sign In"),
		),
	)
}

// ShineBorderDemo is a login card with a purple, pink and orange shine
// circling its border.
func ShineBorderDemo() mx.Component {
	return shineBorderLoginCard("shine-border-demo", "relative w-full max-w-[350px] overflow-hidden",
		magicui.ShineBorderOptions{ShineColors: []string{"#A07CFE", "#FE8FB5", "#FFBE7B"}})
}

// ShineBorderMonochrome is a login card whose shine has the foreground
// color: upstream picks black or white from the theme in JavaScript, the
// CSS variable --foreground does that without it.
func ShineBorderMonochrome() mx.Component {
	return shineBorderLoginCard("shine-border-monochrome", "relative overflow-hidden",
		magicui.ShineBorderOptions{ShineColors: []string{"var(--foreground)"}})
}
