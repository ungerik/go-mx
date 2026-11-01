package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// NativeSelectDemo renders a native select with a placeholder and a disabled
// option.
func NativeSelectDemo() mx.Component {
	return shadcn.NativeSelect("",
		shadcn.NativeSelectOption("", "Select a fruit"),
		shadcn.NativeSelectOption("apple", "Apple"),
		shadcn.NativeSelectOption("banana", "Banana"),
		shadcn.NativeSelectOption("blueberry", "Blueberry"),
		shadcn.NativeSelectOption("grapes", html.Disabled, "Grapes"),
		shadcn.NativeSelectOption("pineapple", "Pineapple"),
	)
}

// NativeSelectGroups renders options in groups.
func NativeSelectGroups() mx.Component {
	return shadcn.NativeSelect("",
		shadcn.NativeSelectOption("", "Select a food"),
		shadcn.NativeSelectOptGroup("Fruits",
			shadcn.NativeSelectOption("apple", "Apple"),
			shadcn.NativeSelectOption("banana", "Banana"),
		),
		shadcn.NativeSelectOptGroup("Vegetables",
			shadcn.NativeSelectOption("carrot", "Carrot"),
			shadcn.NativeSelectOption("broccoli", "Broccoli"),
		),
	)
}

// NativeSelectSizes renders the small and default sizes.
func NativeSelectSizes() mx.Component {
	return html.DivClass("flex flex-col gap-4",
		shadcn.NativeSelect(shadcn.NativeSelectSizeSM,
			shadcn.NativeSelectOption("", "Small"),
			shadcn.NativeSelectOption("apple", "Apple"),
		),
		shadcn.NativeSelect("",
			shadcn.NativeSelectOption("", "Default"),
			shadcn.NativeSelectOption("apple", "Apple"),
		),
	)
}

// NativeSelectWithField renders a native select in a Field with a label and
// description; the id goes to the <select>, so the label finds it.
func NativeSelectWithField() mx.Component {
	return shadcn.Field("", html.Class("max-w-xs"),
		shadcn.FieldLabelFor("native-select-country", "Country"),
		shadcn.NativeSelect("", html.ID("native-select-country"), html.Class("w-full"),
			shadcn.NativeSelectOption("", "Select a country"),
			shadcn.NativeSelectOption("de", "Germany"),
			shadcn.NativeSelectOption("fr", "France"),
			shadcn.NativeSelectOption("us", "United States"),
		),
		shadcn.FieldDescription("Select your country of residence."),
	)
}

// NativeSelectStates renders a disabled and an invalid native select.
func NativeSelectStates() mx.Component {
	return html.DivClass("flex flex-col gap-4",
		shadcn.NativeSelect("", html.Disabled,
			shadcn.NativeSelectOption("", "Disabled"),
		),
		shadcn.NativeSelect("", html.Attrib("aria-invalid", "true"),
			shadcn.NativeSelectOption("", "Error state"),
			shadcn.NativeSelectOption("apple", "Apple"),
		),
	)
}
