package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Dialog is a Go port of shadcn/ui's Dialog. Like [AlertDialog], it replaces
// Radix with the native HTML <dialog> element (top layer, ::backdrop, focus
// trap, Escape-to-close — no client framework). The difference from AlertDialog
// is intent: a Dialog also light-dismisses (clicking the backdrop closes it)
// and carries a built-in close button, whereas an AlertDialog requires an
// explicit action or cancel.
//
// shadcn's DialogPortal and DialogOverlay are not ported for the same reason as
// AlertDialog: a modal <dialog> already renders in the top layer and its
// backdrop is the ::backdrop pseudo-element (styled with `backdrop:bg-black/50`).
// Trigger and content are linked by the dialog id, not by DOM nesting.
func Dialog(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "dialog", "")
}

// DialogTrigger renders a button that opens the <dialog> via showModal(). It
// renders a <button>, so pass it content/styling directly (use
// html.Class([ButtonClasses](...)) for the button look) rather than a nested
// [Button].
func DialogTrigger(dialogID string, attribsChildren ...any) *mx.Element {
	validateID(dialogID)
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs,
			html.OnClick("document.getElementById('"+dialogID+"').showModal()"))
	}
	return finish(e, "dialog-trigger", "")
}

// dialogContentClasses is shadcn's DialogContent class string with the
// Radix-only positioning (fixed/translate/z-50) and animation (data-[state=*],
// duration) classes removed, and `m-auto open:grid backdrop:bg-black/50` added —
// the same native-<dialog> adaptation documented for AlertDialog: `open:grid`
// so the closed dialog stays display:none, and `m-auto` so the browser centers
// the open modal despite Tailwind Preflight's margin reset.
const dialogContentClasses = "open:grid m-auto w-full max-w-[calc(100%-2rem)] gap-4 rounded-lg border bg-background p-6 shadow-lg backdrop:bg-black/50 sm:max-w-lg"

// dialogCloseClasses styles the built-in close button in the top-right corner.
// It is shadcn's close-button class set with the Radix data-[state=open]
// selectors dropped; the [&_svg:not([class*='size-'])]:size-4 sizes the X icon.
const dialogCloseClasses = "ring-offset-background focus:ring-ring absolute top-4 right-4 rounded-xs opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 focus:outline-hidden disabled:pointer-events-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4"

// DialogContent renders the modal <dialog id=dialogID>. Its children are placed
// directly inside, followed by a built-in close button (top-right X). The
// dialog light-dismisses: clicking the backdrop (a click whose target is the
// dialog itself) closes it. Pass html.OnClick("") semantics are not
// special-cased — supply your own onclick to override the light-dismiss.
func DialogContent(dialogID string, attribsChildren ...any) *mx.Element {
	validateID(dialogID)
	e := html.Dialog(append([]any{html.ID(dialogID)}, attribsChildren...)...)
	if e.AttribIndex("onclick") < 0 {
		// Light dismiss: close when the click lands on the dialog/backdrop
		// itself rather than on a descendant (the content box).
		e.Attribs = append(e.Attribs, html.OnClick("if(event.target===this)this.close()"))
	}
	e.Children = append(e.Children, nativeDialogCloseButton("dialog-close"))
	return finish(e, "dialog-content", dialogContentClasses)
}

// nativeDialogCloseButton renders the built-in top-right X that closes the
// enclosing <dialog>. Shared by [DialogContent] and [SheetContent]; slot is the
// data-slot value ("dialog-close" or "sheet-close").
func nativeDialogCloseButton(slot string) *mx.Element {
	return finish(html.Button(
		html.Type("button"),
		html.OnClick("this.closest('dialog').close()"),
		iconX(),
		html.Span(html.Class("sr-only"), "Close"),
	), slot, dialogCloseClasses)
}

// DialogHeader groups the dialog's title and description.
func DialogHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "dialog-header",
		"flex flex-col gap-2 text-center sm:text-left")
}

// DialogFooter groups the dialog's action buttons.
func DialogFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "dialog-footer",
		"flex flex-col-reverse gap-2 sm:flex-row sm:justify-end")
}

// DialogTitle renders the dialog's title as an <h2>.
func DialogTitle(attribsChildren ...any) *mx.Element {
	return finish(html.H2(attribsChildren...), "dialog-title",
		"text-lg leading-none font-semibold")
}

// DialogDescription renders the dialog's description as a <p>.
func DialogDescription(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "dialog-description",
		"text-muted-foreground text-sm")
}

// DialogClose renders a button that closes the enclosing <dialog>. It carries
// no styling of its own (unlike the built-in corner X): like DialogTrigger it
// is a real <button>, so pass content/styling directly, e.g.
// html.Class([ButtonClasses](ButtonOutline, SizeDefault)).
func DialogClose(attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick("this.closest('dialog').close()"))
	}
	return finish(e, "dialog-close", "")
}
