package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// WarpBackgroundDemo is a congratulation card in a tunnel of light beams.
func WarpBackgroundDemo() mx.Component {
	return magicui.WarpBackground(magicui.WarpBackgroundOptions{Seed: 1},
		shadcn.Card(html.Class("w-80"),
			shadcn.CardContent(html.Class("flex flex-col gap-2 p-4"),
				shadcn.CardTitle("Congratulations on Your Promotion!"),
				shadcn.CardDescription("Your hard work and dedication have paid off. We're thrilled to see you take this next step in your career. Keep up the fantastic work!"),
			),
		),
	)
}
