package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Skeleton renders a shadcn/ui skeleton placeholder: a pulsing <div> that
// stands in for content which has not loaded yet. Give it sizing classes
// (e.g. html.Class("h-4 w-32")) to match the content it replaces.
func Skeleton(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "skeleton",
		"bg-accent animate-pulse rounded-md")
}
