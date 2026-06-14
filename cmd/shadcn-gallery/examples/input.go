package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func InputDefault() mx.Component {
	return shadcn.Input(html.Type("email"), html.Attrib("placeholder", "Email"), html.Class("max-w-sm"))
}

func InputDisabled() mx.Component {
	return shadcn.Input(html.Type("email"), html.Attrib("placeholder", "Email"), html.Attrib("disabled", ""), html.Class("max-w-sm"))
}

func InputFile() mx.Component {
	return shadcn.Input(html.Type("file"), html.Class("max-w-sm"))
}

func InputWithLabel() mx.Component {
	return html.Div(html.Class("grid w-full max-w-sm items-center gap-1.5"),
		shadcn.Label(html.For("email"), "Email"),
		shadcn.Input(html.Type("email"), html.ID("email"), html.Attrib("placeholder", "Email")),
	)
}
