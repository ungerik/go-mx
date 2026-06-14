package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func TextareaDefault() mx.Component {
	return shadcn.Textarea(html.Placeholder("Type your message here."), html.Class("max-w-sm"))
}

func TextareaDisabled() mx.Component {
	return shadcn.Textarea(html.Placeholder("Type your message here."), html.Disabled, html.Class("max-w-sm"))
}

func TextareaWithLabel() mx.Component {
	return html.Div(html.Class("grid w-full max-w-sm gap-1.5"),
		shadcn.Label(html.For("message"), "Your message"),
		shadcn.Textarea(html.ID("message"), html.Placeholder("Type your message here.")),
	)
}
