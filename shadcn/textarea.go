package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// textareaClasses is shadcn/ui's textarea class string (new-york-v4,
// Tailwind v4), transcribed verbatim from textarea.tsx. Also the base of
// [InputGroupTextarea].
const textareaClasses = "border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 flex field-sizing-content min-h-16 w-full rounded-md border bg-transparent px-3 py-2 text-base shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"

// Textarea renders a shadcn/ui textarea as a styled <textarea>.
func Textarea(attribsChildren ...any) *mx.Element {
	return finish(html.TextArea(attribsChildren...), "textarea", textareaClasses)
}

// TextareaID renders a [Textarea] with the given id, to link a [Label] or
// [LabelFor], as a shortcut for Textarea(html.ID(id), attribsChildren...).
func TextareaID(id string, attribsChildren ...any) *mx.Element {
	return Textarea(append([]any{html.ID(id)}, attribsChildren...)...)
}
