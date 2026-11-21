package shadcn

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// accounts is a small chart of accounts, the kind of list a Combobox is
// for: too long to scroll through, searched by number or name.
var accounts = []mx.NamedOption{
	{Value: "1200", Name: "1200 Bank"},
	{Value: "4200", Name: "4200 Rent"},
	{Value: "4930", Name: "4930 Office supplies"},
	{Value: "8400", Name: "8400 Revenue 19%"},
}

// startTag returns the start tag of the first element in out that has
// attr, for checks on one element's attributes. A class can hold ">", as
// the button's has-[>svg], so the tag runs up to the next "<".
func startTag(t *testing.T, out, attr string) string {
	t.Helper()
	tag := regexp.MustCompile(`<[a-z]+ [^<]*` + regexp.QuoteMeta(attr) + `[^<]*`).FindString(out)
	if tag == "" {
		t.Fatalf("no element with %s in %s", attr, out)
	}
	return tag
}

func TestComboboxMarkup(t *testing.T) {
	out := render(t, Combobox(ComboboxOptions{
		Name:              "debit",
		Options:           accounts,
		Selected:          "4200",
		SearchPlaceholder: "Search account…",
		Empty:             "No account found.",
		Required:          true,
	}, html.Class("w-[300px]"), html.Attrib("aria-describedby", "debit-help")))

	root := startTag(t, out, `data-slot="combobox"`)
	if !strings.Contains(root, `class="relative w-[300px]"`) {
		t.Errorf("a caller class must go to the root: %s", root)
	}

	// The value is a real form control, so the browser submits and
	// validates it, and it is kept out of reach of pointer, Tab and
	// assistive technology.
	input := startTag(t, out, `data-slot="combobox-input"`)
	for _, want := range []string{ /*html*/ `<select `, `name="debit"`, `tabindex="-1"`, `aria-hidden="true"`, `required`, `pointer-events-none`, `opacity-0`} {
		if !strings.Contains(input, want) {
			t.Errorf("select: missing %s in %s", want, input)
		}
	}
	if !strings.Contains(out /*html*/, `<option value="4200" selected="selected">4200 Rent</option></select>`) {
		t.Errorf("the select must hold exactly the selected value: %s", out)
	}

	trigger := startTag(t, out, `data-slot="combobox-trigger"`)
	for _, want := range []string{
		`id="debit"`,
		`role="combobox"`,
		`aria-controls="debit-popover"`,
		`popovertarget="debit-popover"`,
		`aria-expanded="false"`,
		`aria-required="true"`,
		`aria-describedby="debit-help"`,
		"justify-between",
		"font-normal",
	} {
		if !strings.Contains(trigger, want) {
			t.Errorf("trigger: missing %s in %s", want, trigger)
		}
	}
	if strings.Contains(trigger, "justify-center") || strings.Contains(trigger, "w-[300px]") {
		t.Errorf("the recipe's classes must win over the button's, and the caller's class stays on the root: %s", trigger)
	}
	if !strings.Contains(out, `>4200 Rent</span>`) {
		t.Errorf("the trigger must show the selected label: %s", out)
	}

	content := startTag(t, out, `data-slot="combobox-content"`)
	for _, want := range []string{`id="debit-popover"`, `popover="auto"`, "width: anchor-size(width)", "p-0"} {
		if !strings.Contains(content, want) {
			t.Errorf("content: missing %s in %s", want, content)
		}
	}
	for _, want := range []string{
		`placeholder="Search account…"`,
		`>No account found.</div>`,
		// The selected option carries the check; the others hide theirs.
		`data-value="4200" data-checked="true" aria-checked="true"`,
		`data-value="1200" class=`,
		`>4930 Office supplies<svg`,
		"window.comboboxPick",
		"window.commandFilter.select",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if n := strings.Count(out, `data-checked="true"`); n != 1 {
		t.Errorf("want one checked option, got %d", n)
	}
}

// Without a selection the trigger shows the placeholder, muted, and the
// select holds an empty first option, which is what a required check
// rejects.
func TestComboboxPlaceholderAndDisabled(t *testing.T) {
	out := render(t, Combobox(ComboboxOptions{Name: "credit", Options: accounts, Placeholder: "Pick an account", Disabled: true}))
	for _, want := range []string{
		`data-placeholder="Pick an account" data-empty="true"`,
		`>Pick an account</span>`,
		/*html*/ `<option value="" selected="selected"></option>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, `data-checked="true"`) {
		t.Errorf("nothing is checked without a selection: %s", out)
	}
	if trigger := startTag(t, out, `data-slot="combobox-trigger"`); !strings.Contains(trigger, "disabled") || strings.Contains(trigger, "aria-required") {
		t.Errorf("a disabled, optional trigger: %s", trigger)
	}
	if input := startTag(t, out, `data-slot="combobox-input"`); !strings.Contains(input, "disabled") || strings.Contains(input, "required") {
		t.Errorf("a disabled select submits nothing: %s", input)
	}
}

// A selected value the options do not have is kept and shown as itself,
// so an unchanged submit does not clear it.
func TestComboboxSelectedNotInOptions(t *testing.T) {
	out := render(t, Combobox(ComboboxOptions{Name: "debit", Options: accounts, Selected: "9999"}))
	if !strings.Contains(out /*html*/, `<option value="9999" selected="selected">9999</option>`) || !strings.Contains(out, `>9999</span>`) {
		t.Errorf("an unknown selection must be kept: %s", out)
	}
}

// An option for "" stands for no selection: while it is picked the
// trigger shows its label, muted, as a select shows its "(none)" option.
func TestComboboxEmptyValueOption(t *testing.T) {
	opts := append([]mx.NamedOption{{Name: "(none)"}}, accounts...)
	out := render(t, Combobox(ComboboxOptions{Name: "credit", Options: opts, Placeholder: "Pick an account"}))
	for _, want := range []string{
		`data-placeholder="Pick an account" data-empty="true" class="truncate data-[empty=true]:text-muted-foreground">(none)</span>`,
		/*html*/ `<option value="" selected="selected">(none)</option>`,
		`data-value="" data-checked="true"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
}

// ID overrides the name as the trigger's id, for a name that is no valid
// id; two comboboxes with distinct names get distinct ids.
func TestComboboxIDs(t *testing.T) {
	out := render(t, mx.Components{
		Combobox(ComboboxOptions{Name: "rows[0].debit", ID: "row0-debit", Options: accounts}),
		Combobox(ComboboxOptions{Name: "credit", Options: accounts}),
		// A nil attribute, as a conditional one can be, is left out.
		Combobox(ComboboxOptions{Name: "cost", Options: accounts}, nil),
	})
	for _, want := range []string{`name="rows[0].debit"`, `id="cost"`, `id="row0-debit"`, `id="row0-debit-popover"`, `id="credit"`, `id="credit-popover"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
}

func TestComboboxErrors(t *testing.T) {
	PanicOnInvalidID = false
	t.Cleanup(func() { PanicOnInvalidID = true })
	for name, opts := range map[string]ComboboxOptions{
		"no name":           {},
		"name is no id":     {Name: "rows[0]"},
		"id is no valid id": {Name: "x", ID: "a b"},
	} {
		if err := Combobox(opts).Render(t.Context(), mx.NewCheckedWriter(&strings.Builder{})); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// A disabled or required attribute reaches the select that submits the
// value, as ComboboxOptions.Disabled and Required do, and a form attribute
// goes to it too; an id attribute is an error, as the trigger's id is
// ComboboxOptions.ID, from which the popover and label ids derive.
//
// Value: protects=a caller's disabled or required attribute applying to the submitted select, and an id attribute not splitting the trigger from its derived ids; fails_when=attributes other than class all go to the trigger again; why_new=review found a disabled trigger over an enabled, required select; seam=none; no_claim=none beyond protects
func TestComboboxAttribs(t *testing.T) {
	out := render(t, Combobox(ComboboxOptions{Name: "acct", Options: accounts}, html.Disabled, html.Required, html.Attrib("form", "booking")))
	input := startTag(t, out, `data-slot="combobox-input"`)
	for _, want := range []string{`disabled="disabled"`, `required="required"`, `form="booking"`} {
		if !strings.Contains(input, want) {
			t.Errorf("select: missing %s in %s", want, input)
		}
	}
	trigger := startTag(t, out, `data-slot="combobox-trigger"`)
	for _, want := range []string{`disabled="disabled"`, `aria-required="true"`} {
		if !strings.Contains(trigger, want) {
			t.Errorf("trigger: missing %s in %s", want, trigger)
		}
	}
	if strings.Contains(trigger, `form=`) {
		t.Errorf("the form attribute belongs to the select: %s", trigger)
	}
	if err := Combobox(ComboboxOptions{Name: "acct", Options: accounts}, html.ID("x")).Render(t.Context(), mx.NewCheckedWriter(&strings.Builder{})); err == nil {
		t.Error("an id attribute must be an error")
	}
}

// A caption replaces the trigger text while nothing is selected, and
// checks no item, not even an option for "" the list has.
//
// Value: protects=the list not contradicting the caption for an unavailable value; fails_when=the option for "" is checked under a caption; why_new=review found "(none)" checked beside "(current value not available)"; seam=none; no_claim=the script's own marking, driven in TestComboboxInChrome
func TestComboboxCaption(t *testing.T) {
	none := append([]mx.NamedOption{{Name: "(none)"}}, accounts...)
	root, _ := comboboxParts(ComboboxOptions{Name: "vendor", Options: none, caption: "(not available)"}, nil)
	out := render(t, root)
	for _, want := range []string{
		/*html*/ `<option value="" selected="selected" data-caption="true">(not available)</option>`,
		`>(not available)</span>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, `data-checked="true"`) {
		t.Errorf("a caption checks no item: %s", out)
	}
}

// TestComboboxInChrome drives the script: picking with the pointer and the
// keyboard, the required check, a form reset, and two comboboxes on one
// page.
//
// Value: protects=the submitted value, the trigger label, the check mark, closing and focus return after a pick, and a required pick blocking submit; fails_when=the script stops writing the select, stops closing the popover, or the select stops taking part in validation; why_new=the Combobox is new; seam=none; no_claim=the browser's validation bubble is not visible to a DOM dump
func TestComboboxInChrome(t *testing.T) {
	// A combobox showing a caption for a value the list does not offer,
	// as the FieldDecider renders one, and an optional one with "(none)".
	unavailable, _ := comboboxParts(ComboboxOptions{Name: "partner", Options: accounts, caption: "(not available)"}, nil)
	none := append([]mx.NamedOption{{Name: "(none)"}}, accounts...)
	// A caption over a list with its own option for "".
	unavailableNone, _ := comboboxParts(ComboboxOptions{Name: "vendor", Options: none, caption: "(not available)"}, nil)
	body := html.Form(html.ID("booking"),
		LabelFor("debit", "Debit account"),
		Combobox(ComboboxOptions{Name: "debit", Options: accounts, Required: true}, html.Attrib("aria-describedby", "debit-help")),
		Combobox(ComboboxOptions{Name: "credit", Options: accounts, Selected: "1200"}),
		unavailable,
		unavailableNone,
		Combobox(ComboboxOptions{Name: "tax", Options: none, Selected: "8400"}),
		// An error the server reported, which no pick or reset clears.
		Combobox(ComboboxOptions{Name: "cost", Options: accounts}, html.Attrib("aria-invalid", "true")),
	)
	const script = /*js*/ `
var form=document.getElementById('booking');
var root=function(n){return document.querySelector('[data-slot=combobox]:has([name='+n+'])');};
var part=function(n,s){return root(n).querySelector('[data-slot='+s+']');};
var text=function(n){return part(n,'combobox-value').textContent;};
var checked=function(n){return Array.prototype.map.call(root(n).querySelectorAll('[data-checked=true]'),function(it){return it.getAttribute('data-value');}).join(',');};
var open=function(n){return part(n,'combobox-content').matches(':popover-open');};
var item=function(n,v){return root(n).querySelector('[data-slot=command-item][data-value="'+v+'"]');};
var key=function(el,k){el.dispatchEvent(new KeyboardEvent('keydown',{key:k,bubbles:true,cancelable:true}));};
var changes=0;form.addEventListener('change',function(){changes++;});
var inputs=0;form.addEventListener('input',function(){inputs++;});
var submitted=0;form.addEventListener('submit',function(e){e.preventDefault();submitted++;});
out.hiddenBefore=part('debit','combobox-input').getAttribute('aria-hidden');

// A required combobox with nothing picked blocks the submit, marks its
// trigger invalid and leaves the focus on its select, from which a key
// opens the list.
form.requestSubmit();
out.blocked=submitted;
out.invalid=part('debit','combobox-trigger').getAttribute('aria-invalid');
out.focusOnSelect=document.activeElement===part('debit','combobox-input');
// While the select has the focus it is not hidden, and is named and
// described as the trigger, by the trigger's label.
var sel=part('debit','combobox-input');
out.hiddenWhileFocused=sel.hasAttribute('aria-hidden');
out.selectLabel=sel.getAttribute('aria-labelledby')&&document.getElementById(sel.getAttribute('aria-labelledby')).textContent;
out.selectDescribedBy=sel.getAttribute('aria-describedby');
key(part('debit','combobox-input'),'Enter');
await wait(0);
out.openFromSelect=open('debit');
out.expanded=part('debit','combobox-trigger').getAttribute('aria-expanded');
out.searchFocused=document.activeElement===part('debit','command-input');

// A click picks: value, label, check, closed, focus back on the trigger.
item('debit','4930').click();
await wait(0);
out.value=new FormData(form).get('debit');
out.text=text('debit');
out.checked=checked('debit');
out.closed=!open('debit');
out.collapsed=part('debit','combobox-trigger').getAttribute('aria-expanded');
out.focusOnTrigger=document.activeElement===part('debit','combobox-trigger');
out.hiddenAfterBlur=sel.getAttribute('aria-hidden');
out.validAgain=part('debit','combobox-trigger').hasAttribute('aria-invalid');
out.changes=changes;

// Reopening clears the search and selects the picked option.
part('debit','combobox-trigger').click();
await wait(0);
out.reopenedSearch=part('debit','command-input').value;
out.reopenedSelected=root('debit').querySelector('[data-selected=true]').getAttribute('data-value');
part('debit','combobox-content').hidePopover();
await wait(0);

// A character typed on the trigger opens the list and starts the search;
// Enter picks the first match.
var trigger=part('debit','combobox-trigger');
trigger.focus();
key(trigger,'8');
// At once, without waiting for the toggle event, so the keys typed next
// reach the search box.
var search=part('debit','command-input');
out.seeded=search.value;
out.seededFocus=document.activeElement===search;
await wait(0);
out.seededKept=search.value;
// The search box's own events stay inside the combobox, so a change
// trigger on the form only sees picks.
var before=changes+'/'+inputs;
search.value='zzz';
search.dispatchEvent(new Event('input',{bubbles:true}));
search.dispatchEvent(new Event('change',{bubbles:true}));
out.searchEventsKept=before===changes+'/'+inputs;
// Enter with nothing matching is not the form's implicit submission.
var enter=new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true});
search.dispatchEvent(enter);
out.noMatchEnterPrevented=enter.defaultPrevented;
out.noMatchStillOpen=open('debit');
search.value='84';
search.dispatchEvent(new Event('input',{bubbles:true}));
key(search,'Enter');
await wait(0);
out.keyboardValue=new FormData(form).get('debit');
out.keyboardText=text('debit');

// Value: protects=browser shortcuts on the trigger and ArrowDown opening the list at the pick; fails_when=the modifier check or the ArrowDown case is removed; why_new=only Enter and characters were driven; seam=none; no_claim=IME composition
var ctrl=new KeyboardEvent('keydown',{key:'c',ctrlKey:true,bubbles:true,cancelable:true});
trigger.dispatchEvent(ctrl);
out.ctrlIgnored=!ctrl.defaultPrevented&&!open('debit');
key(trigger,'ArrowDown');
await wait(0);
out.arrowOpensAtPick=open('debit')&&root('debit').querySelector('[data-selected=true]').getAttribute('data-value');
part('debit','combobox-content').hidePopover();
await wait(0);

// A required pick cannot be unpicked; an optional one can, by picking it
// again, as upstream's recipe does.
// Value: protects=no change event, so no htmx request, when a pick keeps the value; fails_when=the v!==prev guard is dropped; why_new=changes was only counted after the first pick; seam=none; no_claim=htmx itself
var changesBefore=changes;
item('debit','8400').click();
out.requiredKept=new FormData(form).get('debit');
out.repickNoChange=changes===changesBefore;
item('credit','1200').click();
out.optionalCleared=new FormData(form).get('credit');
out.optionalText=text('credit');
out.optionalPlaceholder=part('credit','combobox-value').getAttribute('data-empty');
out.optionalUnchecked=checked('credit');
out.creditDidNotTouchDebit=new FormData(form).get('debit');

// The caption for a value the list does not offer is gone once
// anything is picked, also after unpicking, which shows the placeholder.
item('partner','1200').click();
item('partner','1200').click();
out.unavailableUnpicked=text('partner');
// Value: protects=a caption checking no item, and leaving it for the option for "" being a change; fails_when=show checks the item for "" under a caption, or the change guard compares values only; why_new=review found the list contradicting the caption; seam=none; no_claim=none beyond protects
// Counted, as checked() cannot tell the item for "" from none.
var nChecked=function(n){return root(n).querySelectorAll('[data-checked=true]').length;};
out.captionChecked=text('vendor')+'|'+nChecked('vendor');
var captionChanges=changes;
item('vendor','').click();
out.captionLeft=text('vendor')+'|'+nChecked('vendor')+'|'+(changes-captionChanges)+'|'+new FormData(form).get('vendor');
// Picking "(none)" shows it, as does unpicking.
item('tax','').click();
out.nonePicked=text('tax')+'|'+part('tax','combobox-value').getAttribute('data-empty')+'|'+checked('tax');
item('tax','4200').click();
item('tax','4200').click();
out.noneUnpicked=text('tax')+'|'+new FormData(form).get('tax');

// Value: protects=an aria-invalid the server set surviving a pick and a reset; fails_when=the script clears aria-invalid it did not set; why_new=only the browser's own invalid marking was driven; seam=none; no_claim=what a screen reader announces
item('cost','4200').click();
out.serverInvalidAfterPick=part('cost','combobox-trigger').getAttribute('aria-invalid');

form.requestSubmit();
out.submitted=submitted;

// A reset returns to the server-rendered values.
form.reset();
await wait(10);
out.resetDebit=new FormData(form).get('debit');
out.resetDebitText=text('debit');
out.resetCredit=new FormData(form).get('credit');
out.resetCreditText=text('credit');
out.resetCreditChecked=checked('credit');
out.resetValidity=part('debit','combobox-input').checkValidity();
out.resetUnavailable=text('partner')+'|'+new FormData(form).get('partner');
out.resetTax=text('tax');
out.resetCaption=text('vendor')+'|'+nChecked('vendor');
out.serverInvalidAfterReset=part('cost','combobox-trigger').getAttribute('aria-invalid');
`
	got := runInChrome(t, body, script, nil)
	checkRecorded(t, got, map[string]any{
		"blocked":                 float64(0),
		"invalid":                 "true",
		"focusOnSelect":           true,
		"hiddenBefore":            "true",
		"hiddenWhileFocused":      false,
		"selectLabel":             "Debit account",
		"selectDescribedBy":       "debit-help",
		"hiddenAfterBlur":         "true",
		"searchEventsKept":        true,
		"noMatchEnterPrevented":   true,
		"noMatchStillOpen":        true,
		"unavailableUnpicked":     "Choose…",
		"nonePicked":              "(none)|true|",
		"noneUnpicked":            "(none)|",
		"resetUnavailable":        "(not available)|",
		"resetTax":                "8400 Revenue 19%",
		"captionChecked":          "(not available)|0",
		"captionLeft":             "(none)|1|1|",
		"resetCaption":            "(not available)|0",
		"openFromSelect":          true,
		"expanded":                "true",
		"searchFocused":           true,
		"value":                   "4930",
		"text":                    "4930 Office supplies",
		"checked":                 "4930",
		"closed":                  true,
		"collapsed":               "false",
		"focusOnTrigger":          true,
		"validAgain":              false,
		"changes":                 float64(1),
		"reopenedSearch":          "",
		"reopenedSelected":        "4930",
		"seeded":                  "8",
		"seededFocus":             true,
		"seededKept":              "8",
		"keyboardValue":           "8400",
		"keyboardText":            "8400 Revenue 19%",
		"requiredKept":            "8400",
		"repickNoChange":          true,
		"ctrlIgnored":             true,
		"arrowOpensAtPick":        "8400",
		"serverInvalidAfterPick":  "true",
		"serverInvalidAfterReset": "true",
		"optionalCleared":         "",
		"optionalText":            "Choose…",
		"optionalPlaceholder":     "true",
		"optionalUnchecked":       "",
		"creditDidNotTouchDebit":  "8400",
		"submitted":               float64(1),
		"resetDebit":              "",
		"resetDebitText":          "Choose…",
		"resetCredit":             "1200",
		"resetCreditText":         "1200 Bank",
		"resetCreditChecked":      "1200",
		"resetValidity":           false,
	})
}

// TestComboboxStaysInSyncInChrome checks that the select never holds a
// pick the trigger does not show: a copy of picked markup inserted again,
// as an htmx history restore does, shows the value the select submits, a
// reset leaves no stale option behind, and an empty "(none)" pick in a
// required combobox is the placeholder the browser rejects.
//
// Value: protects=the submitted value matching the trigger after re-inserted markup and a reset, and required rejecting a "(none)" pick; fails_when=init stops showing the select's option, reset keeps picked options, or an empty pick is appended after the server option; why_new=review found all three; seam=none; no_claim=htmx's own history cache
func TestComboboxStaysInSyncInChrome(t *testing.T) {
	none := append([]mx.NamedOption{{Name: "(none)"}}, accounts...)
	body := html.Form(html.ID("f"),
		html.Div(html.ID("host"), Combobox(ComboboxOptions{Name: "debit", Options: accounts, Selected: "1200"})),
		Combobox(ComboboxOptions{Name: "tax", Options: none, Selected: "8400", Required: true}),
	)
	const script = /*js*/ `
var form=document.getElementById('f');
var root=function(n){return document.querySelector('[data-slot=combobox]:has([name='+n+'])');};
var part=function(n,s){return root(n).querySelector('[data-slot='+s+']');};
var item=function(n,v){return root(n).querySelector('[data-slot=command-item][data-value="'+v+'"]');};
item('debit','4200').click();
// Re-insert the picked markup and run its scripts, as htmx restores a
// history snapshot: only the server option keeps its selected attribute.
var host=document.getElementById('host');
host.innerHTML=host.innerHTML;
host.querySelectorAll('script').forEach(function(old){var s=document.createElement('script');s.textContent=old.textContent;old.replaceWith(s);});
await wait(0);
out.restoredValue=new FormData(form).get('debit');
out.restoredText=part('debit','combobox-value').textContent;
out.restoredChecked=root('debit').querySelector('[data-checked=true]').getAttribute('data-value');
out.restoredOptions=part('debit','combobox-input').options.length;

item('debit','4930').click();
form.reset();
await wait(10);
out.resetOptions=part('debit','combobox-input').options.length;
out.resetValue=new FormData(form).get('debit');

item('tax','').click();
out.nonePicked=new FormData(form).get('tax');
out.noneValid=part('tax','combobox-input').validity.valid;

// Value: protects=a pick on an invalid field not exposing its select to assistive technology; fails_when=the pick checks validity with checkValidity(), which fires invalid; why_new=review found the synthetic invalid event; seam=none; no_claim=what a screen reader announces
form.requestSubmit();
var invalids=0;part('tax','combobox-input').addEventListener('invalid',function(){invalids++;});
item('tax','').click();
out.pickFiresNoInvalid=invalids+'|'+part('tax','combobox-input').getAttribute('aria-hidden');

// Value: protects=the trigger following a change the browser makes on the focused select; fails_when=the select's change listener is removed; why_new=Codex and three readers found keys the script leaves to the browser; seam=none; no_claim=which keys a platform maps to a select change (synthetic keys have no default action)
var taxSel=part('tax','combobox-input');
taxSel.selectedIndex=1;
taxSel.dispatchEvent(new Event('change',{bubbles:true}));
out.nativeChangeShown=part('tax','combobox-value').textContent+'|'+new FormData(form).get('tax');

// Value: protects=a reset another listener cancelled keeping the comboboxes; fails_when=the reset listener stops checking defaultPrevented; why_new=Codex found it; seam=none; no_claim=none beyond protects
item('debit','4930').click();
form.addEventListener('reset',function(e){e.preventDefault();});
form.reset();
await wait(10);
out.cancelledReset=new FormData(form).get('debit')+'|'+part('debit','combobox-value').textContent;
`
	got := runInChrome(t, body, script, nil)
	checkRecorded(t, got, map[string]any{
		"restoredValue":      "1200",
		"restoredText":       "1200 Bank",
		"restoredChecked":    "1200",
		"restoredOptions":    float64(1),
		"resetOptions":       float64(1),
		"resetValue":         "1200",
		"nonePicked":         "",
		"noneValid":          false,
		"pickFiresNoInvalid": "0|true",
		"nativeChangeShown":  "8400 Revenue 19%|8400",
		"cancelledReset":     "4930|4930 Office supplies",
	})
}
