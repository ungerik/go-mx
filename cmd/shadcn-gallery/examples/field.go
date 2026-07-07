package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// FieldDemo renders a vertical field with label, control and description.
func FieldDemo() mx.Component {
	return shadcn.Field("", html.Class("max-w-sm"),
		shadcn.FieldLabelFor("field-name", "Name"),
		shadcn.InputID("field-name", html.Placeholder("Evil Rabbit")),
		shadcn.FieldDescription("Choose a name to display on your profile."),
	)
}

// FieldHorizontal renders a horizontal field: a switch with title and
// description beside it.
func FieldHorizontal() mx.Component {
	return shadcn.Field(shadcn.FieldHorizontal, html.Class("max-w-sm"),
		shadcn.FieldContent(
			shadcn.FieldTitle("Notifications"),
			shadcn.FieldDescription("Receive emails about account activity."),
		),
		shadcn.SwitchID("field-notifications"),
	)
}

// FieldSetDemo renders a fieldset with a legend and a grouped set of fields,
// split by a labeled separator.
func FieldSetDemo() mx.Component {
	return shadcn.FieldSet(html.Class("w-full max-w-sm"),
		shadcn.FieldLegend("", "Delivery"),
		shadcn.FieldGroup(
			shadcn.Field("",
				shadcn.FieldLabelFor("field-address", "Address"),
				shadcn.InputID("field-address", html.Placeholder("1 Main St")),
			),
			shadcn.FieldSeparator("Or continue with"),
			shadcn.Field("",
				shadcn.FieldLabelFor("field-pickup", "Pickup location"),
				shadcn.InputID("field-pickup", html.Placeholder("Store #12")),
			),
		),
	)
}

// FieldChoiceCard renders checkbox choices as selectable cards: a FieldLabel
// wrapping a whole Field highlights when its checkbox is checked.
func FieldChoiceCard() mx.Component {
	choice := func(id, title, description string, checked bool) mx.Component {
		attribs := []mx.Attrib{html.ID(id)}
		if checked {
			attribs = append(attribs, html.Checked)
		}
		return shadcn.FieldLabel(html.For(id),
			shadcn.Field(shadcn.FieldHorizontal,
				shadcn.FieldContent(
					shadcn.FieldTitle(title),
					shadcn.FieldDescription(description),
				),
				shadcn.Checkbox(attribs...),
			),
		)
	}
	return shadcn.FieldGroup(html.Class("max-w-sm"),
		choice("field-plan-basic", "Basic", "Up to 3 projects.", true),
		choice("field-plan-pro", "Pro", "Unlimited projects and priority support.", false),
	)
}
