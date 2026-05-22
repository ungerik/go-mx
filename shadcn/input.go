package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// inputClasses is shadcn/ui's input class string (new-york-v4, Tailwind v4),
// transcribed verbatim from the three cn arguments in input.tsx.
const inputClasses = "file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 border-input flex h-9 w-full min-w-0 rounded-md border bg-transparent px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive"

// Input renders a shadcn/ui text input as a styled void <input> element. Pass
// the input type the normal way, e.g. html.Type("email"); with no type the
// browser defaults to text. Children are not valid on a void element and are
// dropped.
func Input(attribsChildren ...any) *mx.Element {
	e := html.Element("input", attribsChildren...)
	e.Children = nil // <input> is a void element
	return finish(e, "input", inputClasses)
}
