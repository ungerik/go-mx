package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func SonnerDemo() mx.Component {
	return html.Div(
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault,
			html.OnClick("toast('Event has been created', {description: 'Sunday, June 14, 2026 at 9:00 AM'})"),
			"Show Toast"),
		shadcn.Toaster(),
	)
}
