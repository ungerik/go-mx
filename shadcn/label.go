package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// labelClasses is shadcn/ui's Label class set, shared by [Label] and the
// error-aware [FormLabel].
const labelClasses = "flex items-center gap-2 text-sm leading-none font-medium select-none group-data-[disabled=true]:pointer-events-none group-data-[disabled=true]:opacity-50 peer-disabled:cursor-not-allowed peer-disabled:opacity-50"

// Label renders a shadcn/ui label as a styled <label>. Associate it with a
// form control the normal HTML way, by passing html.For(controlID).
func Label(attribsChildren ...any) *mx.Element {
	return finish(html.Label(attribsChildren...), "label", labelClasses)
}

// LabelFor renders a shadcn/ui [Label] bound to the control with the given id
// as a shortcut for Label(html.For(id), attribsChildren...).
func LabelFor(id string, attribsChildren ...any) *mx.Element {
	return Label(append([]any{html.For(id)}, attribsChildren...)...)
}
