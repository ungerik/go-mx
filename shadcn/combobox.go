package shadcn

import (
	"cmp"

	"github.com/domonda/go-errs"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// ComboboxOptions configures a [Combobox].
type ComboboxOptions struct {
	// Name is the form field name the picked value is submitted under.
	// It must not be empty.
	Name string

	// ID is the id of the trigger button, which a <label for> (such as
	// [FieldLabelFor]) points at; the popover gets ID + "-popover". ""
	// uses Name, which then has to be a valid id (letters, digits, '-'
	// and '_'). Distinct names give several comboboxes on one page
	// distinct ids.
	ID string

	// Options are the choices in their order. An option's Value is what
	// the form submits, its Label() is what the list shows, the search
	// matches and the trigger shows once it is picked. An option whose
	// Value is "" stands for no selection, such as a "(none)" choice:
	// the trigger shows its label, muted like the placeholder, and
	// unpicking an option picks it.
	Options []mx.NamedOption

	// Selected is the value picked initially, "" for none. A value none of
	// the Options has is kept and submitted as it is, shown as itself, so
	// an unchanged submit does not clear a stored value the list no
	// longer offers.
	Selected string

	// Placeholder is the trigger's text while nothing is picked and the
	// Options have no option for "". "" renders the context's
	// [Labels].FormChooseOption.
	Placeholder string

	// SearchPlaceholder is the placeholder of the search box. "" renders
	// the context's [Labels].ComboboxSearch.
	SearchPlaceholder string

	// Empty is the text shown when the search matches no option. ""
	// renders the context's [Labels].ComboboxEmpty.
	Empty string

	// Required keeps the form from being submitted while nothing is
	// picked, and keeps the picked option from being unpicked.
	Required bool

	// Disabled disables the trigger, and the value is not submitted, as
	// for a disabled <select>.
	Disabled bool
}

// comboboxScript is the client code of a [Combobox], emitted in every
// Combobox: its first run defines window.comboboxPick, every run wires up
// the Combobox it is in through document.currentScript, so one inserted by
// an htmx swap works too. It runs after the [Command] script inside the
// popover, whose window.commandFilter it uses.
//
// The value lives in the combobox-input <select>: its option with the
// selected attribute is the server-rendered value, which a form reset
// returns to, and a pick adds an option for any other value. show copies
// the selected option to the trigger text and the check marks.
const comboboxScript = /*js*/ `if(!window.comboboxPick){(function(){` +
	`var part=function(r,s){return r.querySelector('[data-slot='+s+']');};` +
	// The trigger shows the selected option's text, which is "" only
	// where nothing is picked: then the placeholder. Either is muted
	// while the value is "", as for the "(none)" option.
	`var show=function(r){var s=part(r,'combobox-input'),t=part(r,'combobox-value'),o=s.options[s.selectedIndex],v=o?o.value:'';` +
	`t.textContent=(o&&o.text)||t.getAttribute('data-placeholder');if(v)t.removeAttribute('data-empty');else t.setAttribute('data-empty','true');` +
	`r.querySelectorAll('[data-slot=command-item]').forEach(function(it){if(o&&it.getAttribute('data-value')===v){it.setAttribute('data-checked','true');it.setAttribute('aria-checked','true');}else{it.removeAttribute('data-checked');it.removeAttribute('aria-checked');}});};` +
	// prune removes every option but the server-rendered one, which it
	// returns: the select then holds no stale pick that a key on the
	// focused select, a reset or a re-inserted copy of the markup (an
	// htmx history restore) could select behind the trigger's back.
	`var prune=function(s){var keep=null;Array.prototype.slice.call(s.options).forEach(function(o){if(o.defaultSelected&&!keep)keep=o;else o.remove();});return keep;};` +
	`var reset=function(r){var t=part(r,'combobox-trigger'),k=prune(part(r,'combobox-input'));if(k)k.selected=true;show(r);if(r.mxInvalid){t.removeAttribute('aria-invalid');r.mxInvalid=false;}};` +
	`window.comboboxPick=function(it){var r=it.closest('[data-slot=combobox]');if(!r||it.getAttribute('data-disabled')==='true')return;` +
	`var s=part(r,'combobox-input'),c=part(r,'combobox-content'),t=part(r,'combobox-trigger'),v=it.getAttribute('data-value')||'',l=it.textContent.trim(),cur=s.options[s.selectedIndex],prev=cur?cur.value:'';` +
	// Like upstream's recipe, picking the picked option again unpicks it,
	// unless a pick is required; the trigger then shows the option that
	// stands for no selection, if the list has one, else the placeholder.
	`if(v&&v===prev&&!s.required){v='';var n=r.querySelector('[data-slot=command-item][data-value=""]');l=n?n.textContent.trim():'';}` +
	// The server-rendered option is kept for a form reset, and reused
	// only for the same value and text: its text can be a caption, such
	// as the one for a stored value the list no longer offers.
	// An empty value goes first, where the browser takes it for the
	// placeholder that a required check rejects.
	`var keep=prune(s);` +
	`var o=keep&&keep.value===v&&keep.text===l?keep:null;if(!o){o=new Option(l,v);s.add(o,v===''?0:null);}o.selected=true;show(r);` +
	`if(v!==prev){s.dispatchEvent(new Event('input',{bubbles:true}));s.dispatchEvent(new Event('change',{bubbles:true}));}` +
	`if(r.mxInvalid&&s.checkValidity()){t.removeAttribute('aria-invalid');r.mxInvalid=false;}` +
	`if(c.matches(':popover-open'))c.hidePopover();t.focus();};` +
	// Init starts from the server-rendered option, so a re-inserted copy
	// of picked markup shows what the select submits.
	`window.comboboxPick.init=function(r){if(!r||r.mxCombobox)return;r.mxCombobox=true;` +
	`var s=part(r,'combobox-input'),c=part(r,'combobox-content'),t=part(r,'combobox-trigger'),q=part(r,'command-input');` +
	`reset(r);` +
	`r.addEventListener('click',function(e){var it=e.target.closest&&e.target.closest('[data-slot=command-item]');if(it&&r.contains(it))window.comboboxPick(it);});` +
	// The search box is no field of the form: its input and change events
	// stay inside, so a change trigger on an enclosing element only sees
	// the select's, fired by a pick that changes the value.
	`var stop=function(e){e.stopPropagation();};q.addEventListener('input',stop);q.addEventListener('change',stop);` +
	// Opening clears the search, as upstream's popover mounts a fresh
	// Command, then selects the picked option and moves the focus into
	// the search box. A list opened by a key is prepared at once, so the
	// next keys already go to the search box; one opened by a click is
	// prepared on its toggle event, unless keys typed before it were.
	// PopoverContent's ontoggle keeps the trigger's aria-expanded.
	`var prepare=function(seed){q.value=seed;window.commandFilter(q);var p=r.querySelector('[data-slot=command-item][data-checked=true]');if(p&&!p.hidden&&!seed)window.commandFilter.select(p);q.focus();r.mxPrepared=true;};` +
	`c.addEventListener('toggle',function(e){var open=e.newState==='open';if(open&&!r.mxPrepared)prepare('');r.mxPrepared=open;});` +
	// ArrowDown, ArrowUp or a typed character open the list from the
	// trigger, and also from the select, which a failed required check
	// focuses: there Enter and Space open it as well. A typed character
	// starts the search, or extends it while the list is open.
	`var key=function(e){if(e.defaultPrevented||e.altKey||e.ctrlKey||e.metaKey||e.isComposing||e.keyCode===229)return;var k=e.key,seed='';` +
	`if(k==='ArrowDown'||k==='ArrowUp'||(e.target===s&&(k==='Enter'||k===' '))){}else if(k.length===1&&k!==' '){seed=k;}else return;` +
	`e.preventDefault();if(t.disabled)return;if(c.matches(':popover-open')){if(seed)prepare(q.value+seed);return;}t.click();if(c.matches(':popover-open'))prepare(seed);};` +
	`t.addEventListener('keydown',key);s.addEventListener('keydown',key);` +
	// A failed required check marks the trigger invalid until a pick
	// makes the value valid; an aria-invalid from the server stays. The
	// browser then focuses the select of the first invalid field, so for
	// as long as it has the focus the select is no longer hidden from
	// assistive technology, and is named and described as the trigger.
	`s.addEventListener('invalid',function(){if(!t.hasAttribute('aria-invalid')){t.setAttribute('aria-invalid','true');r.mxInvalid=true;}` +
	`['aria-label','aria-labelledby','aria-describedby'].forEach(function(a){var v=t.getAttribute(a);if(v)s.setAttribute(a,v);else s.removeAttribute(a);});` +
	`var ids=Array.prototype.map.call(t.labels||[],function(l,i){if(!l.id)l.id=t.id+'-label'+(i||'');return l.id;}).join(' ');if(ids)s.setAttribute('aria-labelledby',ids);` +
	`s.removeAttribute('aria-hidden');setTimeout(function(){if(document.activeElement!==s)s.setAttribute('aria-hidden','true');},0);});` +
	`s.addEventListener('blur',function(){s.setAttribute('aria-hidden','true');});` +
	// One reset listener per form, not per combobox: it finds the
	// comboboxes when it runs, so those swapped out are not kept alive.
	`var f=s.form;if(f&&!f.mxComboboxReset){f.mxComboboxReset=true;f.addEventListener('reset',function(){setTimeout(function(){Array.prototype.forEach.call(f.elements,function(el){var r=el.getAttribute('data-slot')==='combobox-input'&&el.closest('[data-slot=combobox]');if(r&&r.mxCombobox)reset(r);});},0);});}};` +
	`})();}` +
	`window.comboboxPick.init(document.currentScript&&document.currentScript.closest('[data-slot=combobox]'));`

// comboboxTriggerClasses are upstream's outline button with the Combobox
// recipe's justify-between, full width so the caller's class on the root
// sizes it, and the focus ring also while the hidden select is focused,
// which a failed required check does.
const comboboxTriggerClasses = "w-full justify-between font-normal peer-focus:border-ring peer-focus:ring-[3px] peer-focus:ring-ring/50"

// Combobox renders shadcn/ui's Combobox recipe as a form field: a button
// showing the picked option that opens a [Popover] with a filterable
// [Command] of opts.Options. Typing filters the list, ArrowDown/ArrowUp
// and the pointer move the selection, and Enter or a click picks an
// option: the trigger shows its label, the check mark moves to it, the
// popover closes, the focus returns to the trigger, and the form submits
// its value under opts.Name. Every pick that changes the value fires
// bubbling input and change events, so an htmx trigger on an enclosing
// element (hx-trigger="change") reacts to it.
//
// upstream ships Combobox as a copy-paste recipe, which the gallery still
// shows as the Popover and Command it is made of. This function adds what
// a recipe cannot carry: the submitted value, the trigger label, and a
// required check.
//
// The value is held by a <select> with one option, the submitted value,
// laid over the trigger, transparent and not reachable with the pointer
// or Tab, and hidden from assistive technology, which reads the trigger,
// a role="combobox" button. As a real form control the select takes part
// in the browser's constraint validation, which an <input type="hidden">
// never does: with opts.Required a submit with nothing picked is blocked
// and the browser shows its "select an item" message at the trigger, which
// turns aria-invalid. The focus then is on the select, which for that
// time is no longer hidden and takes the trigger's label and
// description, so a screen reader announces the field the message is
// about. ArrowDown, Enter, Space or a typed character open the list from
// there, and Tab moves on to the trigger. A form reset returns to
// opts.Selected.
//
// The search box is no field of the form: its input and change events do
// not leave the combobox, and Enter in it never submits the form, also
// when the search matches nothing.
//
// The trigger also opens the list with ArrowDown, ArrowUp or a typed
// character, which becomes the start of the search. Escape and a click
// outside close it, as for every [PopoverContent].
//
// A class in attribs goes to the root element, which sizes the trigger
// (full width by default); every other attribute goes to the trigger
// button, such as aria-describedby or aria-invalid. htmx attributes do not
// belong on the trigger, a button that would send its request on click:
// put them on an enclosing element and trigger them on change.
//
// The server checks a submitted value against the options as for any
// <select>: the browser submits whatever the select holds.
func Combobox(opts ComboboxOptions, attribs ...mx.Attrib) *mx.Element {
	root, _ := comboboxParts(opts, "", attribs)
	return root
}

// comboboxParts builds a [Combobox] and returns its root and the select
// that holds its value, to which the [FieldDecider]'s widget adds its
// hx-trigger. The select is nil when the root is an error element.
//
// shown is the trigger's text for opts.Selected when it is not "":
// the [FieldDecider] shows a caption there for a stored value the list
// does not offer, where Selected is "" so that nothing is submitted. Unlike
// the placeholder, the caption is gone once the user picked anything, also
// after unpicking, and a form reset brings it back.
func comboboxParts(opts ComboboxOptions, shown string, attribs []mx.Attrib) (root, input *mx.Element) {
	if opts.Name == "" {
		return mx.NewErrElement(errs.New("shadcn: Combobox needs a Name")), nil
	}
	id := cmp.Or(opts.ID, opts.Name)
	if err := validateID(id); err != nil {
		return mx.NewErrElement(err), nil
	}
	popoverID := id + "-popover"

	// The selected option, with the label of the option list. Nothing
	// selected and no option for it leaves the label "", for which the
	// trigger shows the placeholder.
	selectedLabel := opts.Selected
	for _, opt := range opts.Options {
		if opt.Value == opts.Selected {
			selectedLabel = opt.Label()
			break
		}
	}
	if shown != "" {
		selectedLabel = shown
	}
	input = html.Select(
		html.Name(opts.Name),
		html.TabIndex(-1),
		html.Attrib("aria-hidden", "true"),
		html.Attrib("autocomplete", "off"),
		html.Option(html.Value(opts.Selected), html.Selected, selectedLabel),
	)
	if opts.Required {
		input.Attribs = append(input.Attribs, html.Required)
	}
	if opts.Disabled {
		input.Attribs = append(input.Attribs, html.Disabled)
	}
	input = finish(input, "combobox-input", "peer pointer-events-none absolute inset-0 size-full opacity-0")

	// The trigger text: the selected label, or the placeholder, which the
	// script shows again when the pick is cleared. Both are muted while
	// the value is "".
	value := html.Span()
	var placeholder mx.Attrib = html.DataAttr("placeholder", opts.Placeholder)
	if opts.Placeholder == "" {
		placeholder = labelAttrib{name: "data-placeholder", get: func(l Labels) string { return l.FormChooseOption }}
	}
	value.Attribs = append(value.Attribs, placeholder)
	if opts.Selected == "" {
		value.Attribs = append(value.Attribs, html.DataAttr("empty", "true"))
	}
	switch {
	case selectedLabel != "":
		value.Children = mx.Components{mx.Text(selectedLabel)}
	case opts.Placeholder != "":
		value.Children = mx.Components{mx.Text(opts.Placeholder)}
	default:
		value.Children = mx.Components{labelText(func(l Labels) string { return l.FormChooseOption })}
	}
	value = finish(value, "combobox-value", "truncate data-[empty=true]:text-muted-foreground")

	var rootClasses []any
	triggerArgs := []any{html.ID(id), html.Role("combobox"), html.Attrib("aria-controls", popoverID)}
	for _, a := range attribs {
		if mx.IsNilAttrib(a) {
			continue
		}
		if a.AttribName() == "class" {
			rootClasses = append(rootClasses, a)
		} else {
			triggerArgs = append(triggerArgs, a)
		}
	}
	if opts.Required {
		triggerArgs = append(triggerArgs, html.Attrib("aria-required", "true"))
	}
	if opts.Disabled {
		triggerArgs = append(triggerArgs, html.Disabled)
	}
	triggerArgs = append(triggerArgs,
		// The recipe's classes win over the outline button's.
		html.Class(Cn(ButtonClasses(ButtonOutline, SizeDefault), comboboxTriggerClasses)),
		value,
		icon("chevrons-up-down", "opacity-50", svg.Path(svg.D("m7 15 5 5 5-5")), svg.Path(svg.D("m7 9 5-5 5 5"))),
	)
	trigger := finish(PopoverTrigger(popoverID, triggerArgs...), "combobox-trigger", "")

	items := make([]any, 0, len(opts.Options))
	for _, opt := range opts.Options {
		item := []any{html.DataAttr("value", opt.Value), html.Class("group")}
		if opt.Value == opts.Selected {
			item = append(item, html.DataAttr("checked", "true"), html.Attrib("aria-checked", "true"))
		}
		item = append(item,
			mx.Text(opt.Label()),
			icon("check", "ml-auto opacity-0 group-data-[checked=true]:opacity-100", svg.Path(svg.D("M20 6 9 17l-5-5"))),
		)
		items = append(items, CommandItem(item...))
	}
	var searchPlaceholder mx.Attrib = html.Placeholder(opts.SearchPlaceholder)
	if opts.SearchPlaceholder == "" {
		searchPlaceholder = labelAttrib{name: "placeholder", get: func(l Labels) string { return l.ComboboxSearch }}
	}
	var empty mx.Component = mx.Text(opts.Empty)
	if opts.Empty == "" {
		empty = labelText(func(l Labels) string { return l.ComboboxEmpty })
	}
	content := PopoverContent(popoverID, "",
		html.Class("p-0"),
		Command(
			CommandInput(searchPlaceholder),
			CommandList(
				CommandEmpty(empty),
				CommandGroup("", items...),
			),
		),
	)
	// As wide as the trigger where the browser has anchor positioning,
	// as upstream sizes it with Radix's trigger width; elsewhere the
	// declaration is dropped and the popover keeps its w-72.
	mergeStyle(content, "width: anchor-size(width)")
	content = finish(content, "combobox-content", "")

	root = html.Div(append(rootClasses, input, trigger, content, html.ScriptJS(comboboxScript))...)
	return finish(root, "combobox", "relative"), input
}
