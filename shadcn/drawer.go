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
const drawerScript = /*js*/ `if(!window.drawerStart){window.drawerStart=function(e,h){var d=h.closest('dialog');if(!d)return;var y0=e.clientY;d.style.transition='none';function mv(ev){var dy=ev.clientY-y0;if(dy<0)dy=0;d.style.transform='translateY('+dy+'px)';}function up(ev){document.removeEventListener('pointermove',mv);document.removeEventListener('pointerup',up);var dy=ev.clientY-y0;d.style.transition='transform .2s ease';if(dy>d.offsetHeight*0.4){d.style.transform='translateY(100%)';setTimeout(function(){d.close();d.style.transform='';d.style.transition='';},200);}else{d.style.transform='translateY(0)';}}document.addEventListener('pointermove',mv);document.addEventListener('pointerup',up);};}`

// drawerContentClasses pins the drawer to the bottom edge (overriding the UA
// modal centering the same way SheetContent does) and rounds its top corners.
// `open:flex` keeps a closed <dialog> display:none; max-h-[80vh] keeps it below
// full screen.
const drawerContentClasses = "open:flex flex-col fixed inset-x-0 bottom-0 top-auto m-0 h-auto w-full max-h-[80vh] max-w-none rounded-t-lg border-t bg-background shadow-lg backdrop:bg-black/50"

// Drawer wraps a trigger and its drawer content.
func Drawer(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "drawer", "")
}

// DrawerTrigger renders a button that opens the drawer via showModal(). Pass
// content/styling directly (use html.Class([ButtonClasses](...)) for the look).
func DrawerTrigger(drawerID string, attribsChildren ...any) *mx.Element {
	validateID(drawerID)
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs,
			html.OnClick("document.getElementById('"+drawerID+"').showModal()"))
	}
	return finish(e, "drawer-trigger", "")
}

// DrawerContent renders the bottom <dialog id=drawerID> with a grab handle on
// top (drag it down to dismiss) followed by the caller's children. It
// light-dismisses (a backdrop click closes it). side is bottom-only.
func DrawerContent(drawerID string, attribsChildren ...any) *mx.Element {
	validateID(drawerID)
	e := html.Dialog(append([]any{html.ID(drawerID)}, attribsChildren...)...)
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick("if(event.target===this)this.close()"))
	}
	handle := finish(html.Div(html.Attrib("onpointerdown", "drawerStart(event,this)")),
		"drawer-handle", "bg-muted mx-auto mt-4 h-2 w-[100px] shrink-0 cursor-grab touch-none rounded-full")
	// Grab handle first, then the caller's children, then the shared script.
	e.Children = append(mx.Components{handle}, e.Children...)
	e.Children = append(e.Children, html.Script(mx.Raw(drawerScript)))
	return finish(e, "drawer-content", drawerContentClasses)
}

// DrawerHeader groups the drawer's title and description.
func DrawerHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "drawer-header", "flex flex-col gap-1.5 p-4 text-center md:text-left")
}

// DrawerFooter groups the drawer's action buttons at the bottom.
func DrawerFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "drawer-footer", "mt-auto flex flex-col gap-2 p-4")
}

// DrawerTitle renders the drawer's title as an <h2>.
func DrawerTitle(attribsChildren ...any) *mx.Element {
	return finish(html.H2(attribsChildren...), "drawer-title", "text-foreground font-semibold")
}

// DrawerDescription renders the drawer's description as a <p>.
func DrawerDescription(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "drawer-description", "text-muted-foreground text-sm")
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
	return finish(e, "drawer-close", "")
}
