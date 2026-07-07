package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// SpinnerDemo renders the default spinner.
func SpinnerDemo() mx.Component {
	return shadcn.Spinner()
}

// SpinnerSizes renders spinners in several sizes via caller classes.
func SpinnerSizes() mx.Component {
	return html.DivClass("flex items-center gap-4",
		shadcn.Spinner(html.Class("size-4")),
		shadcn.Spinner(html.Class("size-6")),
		shadcn.Spinner(html.Class("size-8")),
	)
}

// SpinnerInButton renders a disabled button in its loading state.
func SpinnerInButton() mx.Component {
	return shadcn.Button("", "", html.Disabled,
		shadcn.Spinner(),
		"Loading…",
	)
}
