package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Label renders a shadcn/ui label as a styled <label>. Associate it with a
// form control the normal HTML way, by passing html.For(controlID).
func Label(attribsChildren ...any) *mx.Element {
	return finish(html.Label(attribsChildren...), "label",
		"flex items-center gap-2 text-sm leading-none font-medium select-none group-data-[disabled=true]:pointer-events-none group-data-[disabled=true]:opacity-50 peer-disabled:cursor-not-allowed peer-disabled:opacity-50")
}
