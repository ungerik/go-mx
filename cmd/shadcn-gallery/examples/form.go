package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func FormDemo() mx.Component {
	return shadcn.Form(html.Class("w-full max-w-sm space-y-6"),
		shadcn.FormItem(
			shadcn.FormLabel(html.For("form-username"), "Username"),
			shadcn.Input(html.ID("form-username"), html.Placeholder("shadcn")),
			shadcn.FormDescription("This is your public display name."),
		),
		shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, html.Type("submit"), "Submit"),
	)
}

func FormWithError() mx.Component {
	return shadcn.Form(html.Class("w-full max-w-sm space-y-6"),
		shadcn.FormItem(
			shadcn.FormLabel(html.For("form-email"), html.DataAttr("error", "true"), "Email"),
			shadcn.Input(html.ID("form-email"), html.Type("email"),
				html.Value("not-an-email"), html.Attrib("aria-invalid", "true")),
			shadcn.FormMessage("Please enter a valid email address."),
		),
		shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, html.Type("submit"), "Submit"),
	)
}
