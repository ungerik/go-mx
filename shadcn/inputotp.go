package shadcn

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// inputOTPScript is the once-emitted client functions that drive segmented
// input focus and assemble the hidden field's value:
//
//   - otpAdvance(input): on input, advance focus to the next slot when this
//     slot is filled, then rebuild the hidden field by concatenating slots.
//   - otpKey(input, ev): on Backspace from an empty slot, focus + clear the
//     previous slot. ArrowLeft / ArrowRight move focus across slots.
const inputOTPScript = /*js*/ `if(!window.otpAdvance){window.otpAdvance=function(input){var id=input.dataset.otpId;var r=document.querySelector('[data-input-otp="'+id+'"]');if(!r)return;var s=r.querySelectorAll('[data-slot="input-otp-slot"]');var i=+input.dataset.otpIndex;if(input.value.length===1&&i<s.length-1)s[i+1].focus();var h=r.querySelector('input[type=hidden]');if(h){var v='';s.forEach(function(x){v+=x.value||'';});h.value=v;}};window.otpKey=function(input,ev){var id=input.dataset.otpId;var r=document.querySelector('[data-input-otp="'+id+'"]');if(!r)return;var s=r.querySelectorAll('[data-slot="input-otp-slot"]');var i=+input.dataset.otpIndex;if(ev.key==='Backspace'&&input.value===''&&i>0){s[i-1].focus();s[i-1].value='';}else if(ev.key==='ArrowLeft'&&i>0){s[i-1].focus();}else if(ev.key==='ArrowRight'&&i<s.length-1){s[i+1].focus();}};}`

// inputOTPSlotClasses reproduces shadcn's slot look on a real per-character
// <input>. shadcn/ui's input-otp library renders fake slot <div>s next to one
// hidden input; this port uses N real inputs (simpler styling per slot, real
// per-slot caret) and synthesizes one hidden field for form submission — a
// documented divergence.
const inputOTPSlotClasses = "size-9 rounded-md border border-input bg-transparent text-center text-sm shadow-xs outline-none transition-all focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] focus-visible:z-10 disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-destructive"

// InputOTP renders a shadcn/ui segmented one-time-code input as N real
// <input maxlength="1"> slots plus one hidden field that collects the
// concatenated value for form submission. id is a validated stable identifier
// scoping the shared focus script; name is the form field name (placed on the
// hidden field — the slots themselves are nameless); length is the slot count
// and must be ≥ 1.
//
// Slots default to inputmode="numeric" (override per-slot is not exposed; for
// alphanumeric codes override via CSS / app config). The first slot carries
// autocomplete="one-time-code" so password managers recognize the prompt.
//
// shadcn's input-otp library renders one hidden input plus fake slot <div>s;
// this port uses N real inputs (simpler per-slot styling and a real per-slot
// caret) and synthesizes the hidden value with the shared otpAdvance script —
// a documented divergence.
func InputOTP(id, name string, length int, attribsChildren ...any) *mx.Element {
	validateID(id)
	if length < 1 {
		panic("shadcn: InputOTP length must be >= 1")
	}

	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("group"))
	}
	if e.AttribIndex("data-input-otp") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("input-otp", id))
	}

	slots := make(mx.Components, 0, length+2)
	for i := range length {
		idxStr := strconv.Itoa(i)
		slot := html.Element("input",
			html.Type("text"),
			html.Attrib("maxlength", "1"),
			html.InputModeNumeric,
			html.ID(id+"-slot-"+idxStr),
			html.DataAttr("otp-id", id),
			html.DataAttr("otp-index", idxStr),
			html.OnInput("otpAdvance(this)"),
			html.OnKeyDown("otpKey(this,event)"),
		)
		slot.Children = nil
		if i == 0 {
			slot.Attribs = append(slot.Attribs, html.Attrib("autocomplete", "one-time-code"))
		}
		slots = append(slots, finish(slot, "input-otp-slot", inputOTPSlotClasses))
	}

	hidden := html.Element("input",
		html.Type("hidden"),
		html.Name(name),
		html.ID(id+"-value"),
	)
	hidden.Children = nil

	e.Children = append(e.Children, slots...)
	e.Children = append(e.Children, hidden, html.Script(mx.Raw(inputOTPScript)))

	return finish(e, "input-otp", "flex items-center gap-2")
}

// InputOTPSeparator renders a separator between [InputOTP] slot groups as a
// <div role="separator">. With no children it defaults to a lucide minus icon.
func InputOTPSeparator(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("separator"))
	}
	if len(e.Children) == 0 {
		e.Children = mx.Components{iconMinus()}
	}
	return finish(e, "input-otp-separator", "flex items-center text-muted-foreground")
}
