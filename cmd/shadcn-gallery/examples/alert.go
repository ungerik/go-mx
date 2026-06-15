package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn"
)

// AlertDefault renders a default alert with a title and description.
func AlertDefault() mx.Component {
	return shadcn.Alert(shadcn.AlertDefault,
		shadcn.AlertTitle("Heads up!"),
		shadcn.AlertDescription("You can add components to your app using the cli."),
	)
}

// AlertDestructive renders a destructive-variant alert with a title and description.
func AlertDestructive() mx.Component {
	return shadcn.Alert(shadcn.AlertDestructive,
		shadcn.AlertTitle("Error"),
		shadcn.AlertDescription("Your session has expired. Please log in again."),
	)
}
