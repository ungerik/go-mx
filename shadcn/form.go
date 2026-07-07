package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Form renders a native <form> tagged data-slot="form". Lay out the fields
// inside it with the [Field] system ([FieldGroup], [Field], [FieldLabel],
// [FieldDescription], [FieldError], …), which replaced the former Form parts
// (FormItem/FormLabel/FormDescription/FormMessage) following upstream's
// July 2026 move from Form to Field.
//
// shadcn's Form itself is react-hook-form's FormProvider (no DOM), and
// FormField/FormControl are React context/Slot plumbing with no server-side
// equivalent — the caller wires control ids and aria the normal HTML way
// (html.For, html.ID, html.Attrib("aria-invalid", …)).
func Form(attribsChildren ...any) *mx.Element {
	return finish(html.Form(attribsChildren...), "form", "")
}
