package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// FormDemo renders a form with a labeled input, description text, and a submit button.
func FormDemo() mx.Component {
	return shadcn.Form(html.Class("w-full max-w-sm"),
		shadcn.FieldGroup(
			shadcn.Field("",
				shadcn.FieldLabelFor("form-username", "Username"),
				shadcn.InputID("form-username", html.Placeholder("shadcn")),
				shadcn.FieldDescription("This is your public display name."),
			),
			shadcn.Field("",
				shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, html.Type("submit"), "Submit"),
			),
		),
	)
}

// FormWithError renders a form whose field is in an invalid state with an error message.
func FormWithError() mx.Component {
	return shadcn.Form(html.Class("w-full max-w-sm"),
		shadcn.FieldGroup(
			shadcn.Field("", html.DataAttr("invalid", "true"),
				shadcn.FieldLabelFor("form-email", "Email"),
				shadcn.InputID("form-email", html.Type("email"),
					html.Value("not-an-email"), html.Attrib("aria-invalid", "true")),
				shadcn.FieldError("Please enter a valid email address."),
			),
			shadcn.Field("",
				shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, html.Type("submit"), "Submit"),
			),
		),
	)
}
