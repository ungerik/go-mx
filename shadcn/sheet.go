package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// SheetSide selects the edge a [SheetContent] slides in from. "" selects the
// default (right).
type SheetSide string

const (
	// SheetTop slides the sheet in from the top edge.
	SheetTop SheetSide = "top"
	// SheetRight slides the sheet in from the right edge (the default).
	SheetRight SheetSide = "right"
	// SheetBottom slides the sheet in from the bottom edge.
	SheetBottom SheetSide = "bottom"
	// SheetLeft slides the sheet in from the left edge.
	SheetLeft SheetSide = "left"
)

// Sheet is a Go port of shadcn/ui's Sheet — a Dialog variant that slides in
// from a screen edge. Like [Dialog] it is a native modal <dialog> (top layer,
// ::backdrop, Escape-to-close, light-dismiss), but [SheetContent] is pinned to
// an edge and fills it instead of being centered. Trigger and content are
// linked by the sheet id, not by DOM nesting.
//
// shadcn's enter/exit slide animation is not reproduced: a native <dialog>
// snaps open (the same tradeoff as the other ported overlays).
func Sheet(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sheet", "")
}

// SheetTrigger renders a button that opens the sheet via showModal(). Pass
// content/styling directly (use html.Class([ButtonClasses](...)) for the button
// look) rather than a nested [Button].
func SheetTrigger(sheetID string, attribsChildren ...any) *mx.Element {
	if err := validateID(sheetID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs,
			html.OnClick("document.getElementById('"+sheetID+"').showModal()"))
	}
	return finish(e, "sheet-trigger", "")
}

// sheetContentBaseClasses are the side-independent SheetContent classes. As with
// Dialog the Radix animation/z-index classes are dropped; `open:flex` keeps a
// closed <dialog> display:none, `m-0` + `max-h-none max-w-none` clear the UA
// modal centering/clamping so the per-side insets can pin the sheet to its edge.
const sheetContentBaseClasses = "open:flex flex-col fixed m-0 max-h-none max-w-none gap-4 bg-background shadow-lg backdrop:bg-black/50"

// sheetSideClasses pins the sheet to the requested edge. Each side sets the
// opposite inset to auto because the user-agent stylesheet centers a modal
// <dialog> with inset:0 on every side; without the override the sheet would be
// stretched across the viewport instead of pinned.
func sheetSideClasses(side SheetSide) string {
	switch side {
	case SheetTop:
		return "inset-x-0 top-0 bottom-auto h-auto w-full border-b"
	case SheetBottom:
		return "inset-x-0 bottom-0 top-auto h-auto w-full border-t"
	case SheetLeft:
		return "inset-y-0 left-0 right-auto h-full w-3/4 border-r sm:max-w-sm"
	default: // right
		return "inset-y-0 right-0 left-auto h-full w-3/4 border-l sm:max-w-sm"
	}
}

// SheetContent renders the modal <dialog id=sheetID> pinned to side. Children
// are placed inside, followed by the built-in top-right close button. It
// light-dismisses (a click on the backdrop closes it). side may be "" for the
// default (right).
func SheetContent(sheetID string, side SheetSide, attribsChildren ...any) *mx.Element {
	if err := validateID(sheetID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Dialog(append([]any{html.ID(sheetID)}, attribsChildren...)...)
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick("if(event.target===this)this.close()"))
	}
	e.Children = append(e.Children, nativeDialogCloseButton("sheet-close"))
	return finish(e, "sheet-content", sheetContentBaseClasses+" "+sheetSideClasses(side))
}

// SheetHeader groups the sheet's title and description.
func SheetHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sheet-header", "flex flex-col gap-1.5 p-4")
}

// SheetFooter groups the sheet's action buttons at the bottom (mt-auto).
func SheetFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sheet-footer", "mt-auto flex flex-col gap-2 p-4")
}

// SheetTitle renders the sheet's title as an <h2>.
func SheetTitle(attribsChildren ...any) *mx.Element {
	return finish(html.H2(attribsChildren...), "sheet-title", "text-foreground font-semibold")
}

// SheetDescription renders the sheet's description as a <p>.
func SheetDescription(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "sheet-description", "text-muted-foreground text-sm")
}

// SheetClose renders a button that closes the enclosing sheet. Like
// SheetTrigger it is a real <button>; pass content/styling directly.
func SheetClose(attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick("this.closest('dialog').close()"))
	}
	return finish(e, "sheet-close", "")
}
