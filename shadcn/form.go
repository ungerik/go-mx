package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Form is a Go port of shadcn/ui's Form. shadcn's Form is react-hook-form's
// FormProvider (no DOM), and FormField / FormControl are React context / Slot
// plumbing with no server-side equivalent — so this port exposes the renderable
// parts: Form (the native <form>), FormItem, FormLabel, FormDescription and
// FormMessage. The caller wires control ids and aria the normal HTML way
// (html.For, html.ID, html.Attrib("aria-invalid", ...)); the control goes
// directly inside FormItem (there is no FormControl wrapper).
//
// Server-side validation display is the natural fit: render a FormMessage with
// the error text and pass html.DataAttr("error", "true") to the FormLabel when
// the field is invalid.
func Form(attribsChildren ...any) *mx.Element {
	return finish(html.Form(attribsChildren...), "form", "")
}

// FormItem groups one field's label, control, description and message.
func FormItem(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "form-item", "grid gap-2")
}

// FormLabel is a [Label] that turns the destructive color when its field is in
// error. Mark the error state by passing html.DataAttr("error", "true")
// (shadcn's data-error), which the data-[error=true]:text-destructive variant
// keys off.
func FormLabel(attribsChildren ...any) *mx.Element {
	return finish(html.Label(attribsChildren...), "form-label",
		labelClasses+" data-[error=true]:text-destructive")
}

// FormLabelFor is a [FormLabel] bound to the control with the given id
// as a shortcut for FormLabel(html.For(id), attribsChildren...).
func FormLabelFor(id string, attribsChildren ...any) *mx.Element {
	return FormLabel(append([]any{html.For(id)}, attribsChildren...)...)
}

// FormDescription renders helper text below a control.
func FormDescription(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "form-description", "text-muted-foreground text-sm")
}

// FormMessage renders a validation message in the destructive color. Render it
// only when there is an error to show.
func FormMessage(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "form-message", "text-destructive text-sm")
}
