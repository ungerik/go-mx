package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AlertDialog is a Go port of shadcn/ui's Alert Dialog. shadcn's component is
// built on Radix; this port replaces Radix with the native HTML <dialog>
// element, which provides the modal top layer, ::backdrop, focus trap and
// Escape-to-close with no client-side framework.
//
// shadcn's AlertDialogPortal and AlertDialogOverlay are not ported: a modal
// <dialog> already renders in the top layer, and its backdrop is the
// ::backdrop pseudo-element (styled here via the Tailwind v4 `backdrop:`
// variant). See README.md for the full porting rationale.

// AlertDialog wraps a trigger and its dialog content. It is purely structural:
// the trigger and content are linked by the dialog id passed to
// [AlertDialogTrigger] and [AlertDialogContent], not by DOM nesting.
func AlertDialog(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "alert-dialog", "")
}

// AlertDialogTrigger renders a button that opens the <dialog> with the given
// id via dialog.showModal(). Pass the trigger's content (label, icon) and,
// optionally, styling. It renders a <button>, so pass text/class rather than a
// nested [Button]; for the button look use html.Class([ButtonClasses](...)).
func AlertDialogTrigger(dialogID string, attribsChildren ...any) *mx.Element {
	if err := validateID(dialogID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs,
			html.OnClick("document.getElementById('"+dialogID+"').showModal()"))
	}
	return finish(e, "alert-dialog-trigger", "")
}

// alertDialogContentClasses is shadcn's AlertDialogContent class string with
// the Radix-only positioning (fixed/translate/z-50), animation
// (data-[state=*]) and duration classes removed, and `backdrop:bg-black/50`
// added. A native modal <dialog> is centered in the top layer by the user
// agent and has no data-state attribute.
//
// The display utility is `open:grid`, not `grid`: shadcn renders the content
// only while the Radix dialog is open, but a native <dialog> is always in the
// DOM and relies on the user-agent rule `dialog:not([open]){display:none}` to
// stay hidden. An unconditional `grid` is an author-origin style that would
// override that UA rule and leak the closed dialog onto the page; scoping it to
// `open:` (the `[open]` attribute set by showModal()) keeps the closed dialog
// hidden and still lays the open one out as a grid.
//
// `m-auto` restores the user-agent's `margin: auto` that centers a modal
// <dialog> in the top layer. Tailwind v4 Preflight (required by this package)
// resets every element's margin to 0, which otherwise defeats that centering
// and pins the open dialog to the top-left.
const alertDialogContentClasses = "group/alert-dialog-content open:grid m-auto w-full max-w-[calc(100%-2rem)] gap-4 rounded-lg border bg-background p-6 shadow-lg backdrop:bg-black/50 data-[size=sm]:max-w-xs data-[size=default]:sm:max-w-lg"

// AlertDialogContent renders the modal <dialog id=dialogID>. Its children are
// placed inside a <form method="dialog"> so that the submit buttons produced
// by [AlertDialogAction] and [AlertDialogCancel] close the dialog.
//
// A data-size attribute (default "default", or "sm") drives the responsive
// max-width; pass html.DataAttr("size", "sm") to override.
func AlertDialogContent(dialogID string, attribsChildren ...any) *mx.Element {
	if err := validateID(dialogID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Dialog(append([]any{html.ID(dialogID)}, attribsChildren...)...)
	// Move the caller's children inside a <form method="dialog"> so action and
	// cancel buttons close the dialog on submit. Caller attributes stay on the
	// <dialog> element.
	form := html.Form(html.MethodDialog)
	form.Children = e.Children
	e.Children = mx.Components{form}
	if e.AttribIndex("data-size") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("size", "default"))
	}
	return finish(e, "alert-dialog-content", alertDialogContentClasses)
}

// AlertDialogHeader groups the dialog's title and description.
func AlertDialogHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "alert-dialog-header",
		"grid grid-rows-[auto_1fr] place-items-center gap-1.5 text-center has-data-[slot=alert-dialog-media]:grid-rows-[auto_auto_1fr] has-data-[slot=alert-dialog-media]:gap-x-6 sm:group-data-[size=default]/alert-dialog-content:place-items-start sm:group-data-[size=default]/alert-dialog-content:text-left sm:group-data-[size=default]/alert-dialog-content:has-data-[slot=alert-dialog-media]:grid-rows-[auto_1fr]")
}

// AlertDialogFooter groups the dialog's action buttons.
func AlertDialogFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "alert-dialog-footer",
		"flex flex-col-reverse gap-2 group-data-[size=sm]/alert-dialog-content:grid group-data-[size=sm]/alert-dialog-content:grid-cols-2 sm:flex-row sm:justify-end")
}

// AlertDialogTitle renders the dialog's title as an <h2>.
func AlertDialogTitle(attribsChildren ...any) *mx.Element {
	return finish(html.H2(attribsChildren...), "alert-dialog-title",
		"text-lg font-semibold sm:group-data-[size=default]/alert-dialog-content:group-has-data-[slot=alert-dialog-media]/alert-dialog-content:col-start-2")
}

// AlertDialogDescription renders the dialog's description as a <p>.
func AlertDialogDescription(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "alert-dialog-description",
		"text-sm text-muted-foreground")
}

// AlertDialogMedia renders an optional icon/media slot above the title.
func AlertDialogMedia(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "alert-dialog-media",
		"mb-2 inline-flex size-16 items-center justify-center rounded-md bg-muted sm:group-data-[size=default]/alert-dialog-content:row-span-2 *:[svg:not([class*='size-'])]:size-8")
}

// AlertDialogAction renders the confirm button. It is a type="submit" button
// styled with the default [Button] variant; inside the dialog's
// <form method="dialog"> it closes the dialog and sets returnValue to its
// value (default "confirm"). Attach a real side effect with html.OnClick, or
// html.FormAction(...) plus html.FormMethodPOST for a server action.
func AlertDialogAction(attribsChildren ...any) *mx.Element {
	return alertDialogButton(attribsChildren, "alert-dialog-action", ButtonDefault, "confirm")
}

// AlertDialogCancel renders the cancel button: a type="submit" button styled
// with the outline [Button] variant that closes the dialog with
// returnValue "cancel".
func AlertDialogCancel(attribsChildren ...any) *mx.Element {
	return alertDialogButton(attribsChildren, "alert-dialog-cancel", ButtonOutline, "cancel")
}

func alertDialogButton(attribsChildren []any, slot string, variant ButtonVariant, defaultValue string) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("submit"))
	}
	if e.AttribIndex("value") < 0 {
		e.Attribs = append(e.Attribs, html.Value(defaultValue))
	}
	return finish(e, slot, ButtonClasses(variant, SizeDefault))
}
