package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Toaster is a Go port of shadcn/ui's Sonner toaster. shadcn's toast is
// inherently imperative and client-driven — you call toast(message) in response
// to an event — so this port keeps that with one shared script that defines a
// global toast(message, opts) function. Place one [Toaster] on the page, then
// trigger toasts from any onclick/script:
//
//	shadcn.Button("", "", html.OnClick(
//	    "toast('Saved', {description: 'Your changes are saved.'})"), "Save")
//
// Each toast is appended to the fixed Toaster region, styled like Sonner, and
// auto-dismissed after opts.duration ms (default 4000). For server-pushed
// toasts, an HTMX out-of-band swap into the Toaster region is the alternative;
// Sonner's swipe-to-dismiss and stacking offsets are not reproduced.
const toastScript = /*js*/ `if(!window.toast){window.toast=function(msg,opts){opts=opts||{};var r=document.querySelector('[data-slot=toaster]');if(!r)return;var el=document.createElement('div');el.setAttribute('data-slot','toast');el.className='bg-background text-foreground pointer-events-auto flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-1 rounded-lg border p-4 shadow-lg transition-all duration-300';var t=document.createElement('div');t.className='text-sm font-medium';t.textContent=msg;el.appendChild(t);if(opts.description){var d=document.createElement('div');d.className='text-muted-foreground text-sm';d.textContent=opts.description;el.appendChild(d);}r.appendChild(el);setTimeout(function(){el.style.opacity='0';el.style.transform='translateY(8px)';setTimeout(function(){el.remove();},300);},opts.duration||4000);};}`

// Toaster renders the fixed bottom-right region that toasts append to, plus the
// shared toast() script. Render it once per page.
func Toaster(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	e.Children = append(e.Children, html.Script(mx.Raw(toastScript)))
	return finish(e, "toaster",
		"pointer-events-none fixed right-4 bottom-4 z-[100] flex flex-col gap-2")
}
