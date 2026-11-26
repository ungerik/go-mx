package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Drawer is a Go port of shadcn/ui's Drawer. shadcn wraps Vaul (a drag-to-
// dismiss bottom sheet). This port keeps the native bottom <dialog> — top
// layer, ::backdrop, Escape-to-close and light-dismiss come for free, the same
// infra as [Dialog] and [Sheet] — and adds the one thing the platform has no
// equivalent for: the drag gesture. A single shared drawerStart pointer-drag
// script (the Slider/Resizable pattern) translates the drawer with a downward
// drag and, on release, closes it past a ~40% threshold or snaps it back.
//
// Dropped vs Vaul: multi snap-points, momentum physics, the background-scale
// effect, and non-bottom directions. Trigger and content are linked by the
// drawer id, not by DOM nesting.
//
// Opening and closing slide the drawer like Vaul's, with
// [modalTransitionClasses]. The drag moves the drawer with an inline
// `translate`, the property those classes transition, and turns transitions
// off while the pointer is down. On release it clears both, so the classes
// take over from wherever the drag left the drawer: they slide it back into
// place, or, when the drag went past the threshold, slide it out as it
// closes. No timer outlives the gesture, so a new drag or an Escape right
// after a release finds nothing left over from the previous one.
const drawerScript = /*js*/ `if(!window.drawerStart){window.drawerStart=function(e,h){var d=h.closest('dialog');if(!d)return;var y0=e.clientY;d.style.transition='none';function mv(ev){var dy=ev.clientY-y0;if(dy<0)dy=0;d.style.translate='0 '+dy+'px';}function up(ev){document.removeEventListener('pointermove',mv);document.removeEventListener('pointerup',up);var dy=ev.clientY-y0;d.style.transition='';d.style.translate='';if(dy>d.offsetHeight*0.4)d.close();}document.addEventListener('pointermove',mv);document.addEventListener('pointerup',up);};}`

// drawerContentClasses pins the drawer to the bottom edge (overriding the UA
// modal centering the same way SheetContent does) and rounds its top corners.
// `open:flex` keeps a closed <dialog> display:none; max-h-[80vh] keeps it below
// full screen. It slides up from below the edge and back down with Vaul's
// 500ms cubic-bezier(0.32,0.72,0,1) transition.
const drawerContentClasses = "open:flex flex-col fixed inset-x-0 bottom-0 top-auto m-0 h-auto w-full max-h-[80vh] max-w-none rounded-t-lg border-t bg-background shadow-lg " +
	"translate-y-full open:translate-y-0 starting:open:translate-y-full duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] " +
	modalTransitionClasses

// Drawer wraps a trigger and its drawer content.
func Drawer(attribsChildren ...any) *mx.Element {
	return StyleSlot(html.Div(attribsChildren...), "drawer", "")
}

// DrawerTrigger renders a button that opens the drawer via showModal(). Pass
// content/styling directly (use html.Class([ButtonClasses](...)) for the look).
func DrawerTrigger(drawerID string, attribsChildren ...any) *mx.Element {
	if err := validateID(drawerID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs,
			html.OnClick("document.getElementById('"+drawerID+"').showModal()"))
	}
	return StyleSlot(e, "drawer-trigger", "")
}

// DrawerContent renders the bottom <dialog id=drawerID> with a grab handle on
// top (drag it down to dismiss) followed by the caller's children. It
// light-dismisses (a backdrop click closes it). side is bottom-only.
func DrawerContent(drawerID string, attribsChildren ...any) *mx.Element {
	if err := validateID(drawerID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Dialog(append([]any{html.ID(drawerID)}, attribsChildren...)...)
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick("if(event.target===this)this.close()"))
	}
	handle := StyleSlot(html.Div(html.Attrib("onpointerdown", "drawerStart(event,this)")),
		"drawer-handle", "bg-muted mx-auto mt-4 h-2 w-[100px] shrink-0 cursor-grab touch-none rounded-full")
	// Grab handle first, then the caller's children, then the shared script.
	e.Children = append(append(mx.Components{handle}, e.Children...), html.ScriptJS(drawerScript))
	return StyleSlot(e, "drawer-content", drawerContentClasses)
}

// DrawerHeader groups the drawer's title and description.
func DrawerHeader(attribsChildren ...any) *mx.Element {
	return StyleSlot(html.Div(attribsChildren...), "drawer-header", "flex flex-col gap-1.5 p-4 text-center md:text-left")
}

// DrawerFooter groups the drawer's action buttons at the bottom.
func DrawerFooter(attribsChildren ...any) *mx.Element {
	return StyleSlot(html.Div(attribsChildren...), "drawer-footer", "mt-auto flex flex-col gap-2 p-4")
}

// DrawerTitle renders the drawer's title as an <h2>.
func DrawerTitle(attribsChildren ...any) *mx.Element {
	return StyleSlot(html.H2(attribsChildren...), "drawer-title", "text-foreground font-semibold")
}

// DrawerDescription renders the drawer's description as a <p>.
func DrawerDescription(attribsChildren ...any) *mx.Element {
	return StyleSlot(html.P(attribsChildren...), "drawer-description", "text-muted-foreground text-sm")
}

// DrawerClose renders a button that closes the enclosing drawer. Like
// DrawerTrigger it is a real <button>; pass content/styling directly.
func DrawerClose(attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick("this.closest('dialog').close()"))
	}
	return StyleSlot(e, "drawer-close", "")
}
