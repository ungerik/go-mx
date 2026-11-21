package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

func TestCommandComposition(t *testing.T) {
	out := render(t, Command(
		CommandInput(html.Attrib("placeholder", "Type a command...")),
		CommandList(
			CommandEmpty("No results found."),
			CommandGroup("Suggestions",
				CommandItem("Calendar"),
				CommandItem("Search", CommandShortcut("⌘S")),
			),
			CommandSeparator(),
			CommandGroup("Settings",
				CommandItem("Profile"),
			),
		),
	))
	for _, want := range []string{
		`data-slot="command"`,
		"window.commandFilter",
		`data-slot="command-input-wrapper"`,
		"lucide-search",
		`data-slot="command-input"`,
		`oninput="commandFilter(this)"`,
		`placeholder="Type a command..."`,
		`data-slot="command-list"`,
		`data-slot="command-empty"`,
		"hidden",
		"No results found.",
		`data-slot="command-group"`,
		"Suggestions",
		`data-slot="command-item"`,
		"Calendar",
		`data-slot="command-shortcut"`,
		`data-slot="command-separator"`,
		"Profile",
		// cmdk's ARIA: a combobox input controlling a listbox of options in
		// labeled groups, the visible heading hidden as the group's name.
		`role="combobox"`,
		`aria-autocomplete="list"`,
		`aria-expanded="true"`,
		`role="listbox"`,
		`role="group" aria-label="Suggestions"`,
		`aria-hidden="true"`,
		`role="option"`,
		// The listbox's other children get cmdk's roles too.
		`role="presentation"`,
		`role="separator"`,
		// The selected item's look is upstream's data-selected classes, not hover.
		"data-[selected=true]:bg-accent",
		"window.commandFilter.init(document.currentScript",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "hover:bg-accent") {
		t.Errorf("an item is highlighted by its selection, not by hover: %s", out)
	}
}

// TestCommandGroupHeadingFirst checks that the heading stays the group's
// first child when the caller passes attributes and children, and that the
// caller's attributes land on the group.
func TestCommandGroupHeadingFirst(t *testing.T) {
	out := render(t, CommandGroup("Fruit", html.ID("fruit"), CommandItem("Apple")))
	if !strings.Contains(out, `id="fruit"`) {
		t.Errorf("the caller's attribute is missing: %s", out)
	}
	heading, item := strings.Index(out, ">Fruit<"), strings.Index(out, "Apple")
	if heading < 0 || item < 0 || heading > item {
		t.Errorf("the heading should come before the items: %s", out)
	}
}

// TestCommandUntitledGroupAndInputAttributes checks the cmdk details of the
// markup the keyboard script relies on.
//
// Value: protects=an untitled group getting no empty accessible name or heading, and the input keeping the browser's autofill and spelling popups off the list; fails_when=the heading guard is dropped or the input attributes are removed; why_new=no test covered either; seam=none; no_claim=-
func TestCommandUntitledGroupAndInputAttributes(t *testing.T) {
	group := render(t, CommandGroup("", CommandItem("Apple")))
	if strings.Contains(group, "aria-label") || strings.Contains(group, "aria-hidden") {
		t.Errorf("an untitled group should have no label and no heading: %s", group)
	}
	input := render(t, CommandInput())
	for _, want := range []string{`autocomplete="off"`, `autocorrect="off"`, `spellcheck="false"`} {
		if !strings.Contains(input, want) {
			t.Errorf("missing %q in %s", want, input)
		}
	}
}

func TestCommandCallerOverridesRole(t *testing.T) {
	out := render(t, CommandItem(html.Role("menuitem"), "x"))
	if !strings.Contains(out, `role="menuitem"`) || strings.Contains(out, `role="option"`) {
		t.Errorf("the caller's role should replace the default: %s", out)
	}
}

// TestCommandKeyboardInChrome drives the selection script in headless Chrome:
// the first item is selected on load, the arrow keys, Home and End move the
// selection over the visible, enabled items without wrapping, filtering
// selects the first match, the pointer selects the item under it, Enter
// clicks the selected item or the link inside it, IME composition and
// prevented keys are ignored, and items swapped in later are selected.
func TestCommandKeyboardInChrome(t *testing.T) {
	body := mx.Components{
		Command(html.ID("cmd"),
			CommandInput(),
			CommandList(
				CommandEmpty("Nothing."),
				CommandGroup("Fruit",
					CommandItem(html.ID("apple"), html.OnClick("window.clicked='apple'"), "Apple"),
					CommandItem(html.ID("banana"), html.Attrib("data-disabled", "true"), "Banana"),
					CommandItem(html.ID("cherry"), "Cherry"),
				),
				CommandSeparator(),
				CommandGroup("Links",
					CommandItem(html.ID("docs"), html.A(html.HRef("#docs"), html.OnClick("window.clicked='docs';return false"), "Docs")),
				),
			),
		),
		// A second Command on the page is wired up by its own script run.
		Command(html.ID("second"), CommandInput(), CommandList(CommandItem(html.ID("only"), "Only"))),
		// Two Commands whose items have no id, the common markup; the
		// second item of the first wraps a button.
		Command(html.ID("plain"), CommandInput(), CommandList(
			CommandItem("One"),
			CommandItem(html.Button(html.OnClick("window.clicked='button'"), "Two")),
		)),
		Command(html.ID("plain2"), CommandInput(), CommandList(CommandItem("Three"))),
	}
	const script = /*js*/ `
var input=document.querySelector('#cmd [data-slot=command-input]');
var key=function(k){input.dispatchEvent(new KeyboardEvent('keydown',{key:k,bubbles:true,cancelable:true}));};
var selected=function(root){var s=document.querySelectorAll((root||'#cmd')+' [data-selected=true]');return s.length===1?s[0].id:s.length+' selected';};
out.initial=selected();
out.activedescendant=input.getAttribute('aria-activedescendant');
out.controls=input.getAttribute('aria-controls')===document.querySelector('#cmd [data-slot=command-list]').id;
out.secondInitial=selected('#second');
key('ArrowDown');
out.downSkipsDisabled=selected();
key('ArrowDown');
key('ArrowDown');
out.downStopsAtEnd=selected();
key('Home');
out.home=selected();
key('ArrowUp');
out.upStopsAtStart=selected();
key('End');
out.end=selected();
key('Enter');
out.enterClicksLink=window.clicked;
input.value='rr';
input.dispatchEvent(new Event('input',{bubbles:true}));
out.filtered=selected();
input.value='';
input.dispatchEvent(new Event('input',{bubbles:true}));
out.unfiltered=selected();
key('Enter');
out.enterClicksItem=window.clicked;
document.getElementById('cherry').dispatchEvent(new PointerEvent('pointermove',{bubbles:true}));
out.pointer=selected();
document.getElementById('banana').dispatchEvent(new PointerEvent('pointermove',{bubbles:true}));
out.pointerSkipsDisabled=selected();
out.ariaSelected=document.getElementById('cherry').getAttribute('aria-selected');
out.previousCleared=document.getElementById('apple').hasAttribute('aria-selected');
input.value='zzz';
input.dispatchEvent(new Event('input',{bubbles:true}));
out.noMatch=selected();
out.noMatchActivedescendant=input.hasAttribute('aria-activedescendant');
// Value: protects=cmdk's empty state and the hiding of groups without a match; fails_when=the group or CommandEmpty toggling breaks while the selection still works; why_new=noMatch checked only the selection; seam=none; no_claim=no fuzzy ranking
var groups=function(){return Array.prototype.map.call(document.querySelectorAll('#cmd [data-slot=command-group]'),function(g){return g.hidden?'hidden':'shown';}).join(',');};
var empty=function(){return document.querySelector('#cmd [data-slot=command-empty]').hidden?'hidden':'shown';};
out.noMatchGroups=groups();
out.noMatchEmpty=empty();
input.value='app';
input.dispatchEvent(new Event('input',{bubbles:true}));
out.appGroups=groups();
out.appEmpty=empty();
input.value='';
input.dispatchEvent(new Event('input',{bubbles:true}));
// Keys are left alone while an IME composes (isComposing, or Safari's
// keyCode 229) and after an earlier handler called preventDefault.
var ev=new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true});
Object.defineProperty(ev,'keyCode',{value:229});
window.clicked='';
input.dispatchEvent(ev);
out.enterKeyCode229=window.clicked;
input.dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowDown',isComposing:true,bubbles:true,cancelable:true}));
out.arrowComposing=selected();
var stop=function(e){e.preventDefault();};
input.addEventListener('keydown',stop);
key('ArrowDown');
input.removeEventListener('keydown',stop);
out.arrowPrevented=selected();
// Items swapped in after load, as htmx swaps server-side results: the
// removed selection is replaced by the first new item, and a replaced list
// is linked to the input again.
var list=document.querySelector('#cmd [data-slot=command-list]');
list.innerHTML='<div data-slot="command-item" role="option" id="kiwi">Kiwi</div><div data-slot="command-item" role="option" id="lime">Lime</div>';
await wait(0);
out.swapped=selected();
out.swappedActivedescendant=input.getAttribute('aria-activedescendant');
var fresh=document.createElement('div');
fresh.setAttribute('data-slot','command-list');
fresh.innerHTML='<div data-slot="command-item" role="option" id="plum">Plum</div>';
list.replaceWith(fresh);
await wait(0);
out.listReplaced=selected();
out.listRelinked=!!fresh.id&&input.getAttribute('aria-controls')===fresh.id;
// Value: protects=aria-activedescendant on items without an id, with ids unique across Commands; fails_when=the id assignment is dropped or its counter restarts per Command; why_new=every other item has an id; seam=none; no_claim=-
var plainInput=document.querySelector('#plain [data-slot=command-input]');
var first=document.querySelector('#plain [data-selected=true]'),other=document.querySelector('#plain2 [data-selected=true]');
out.plainID=/^mx-command-item-[0-9]+$/.test(first.id)&&plainInput.getAttribute('aria-activedescendant')===first.id;
out.plainIDsDiffer=!!other&&other.id!==first.id;
// Value: protects=typing in the input and Enter not submitting an enclosing form; fails_when=preventDefault moves above the key switch or leaves Enter; why_new=preventDefault ran but was never asserted; seam=none; no_claim=synthetic events trigger no implicit form submission
var prevented=function(k){var e=new KeyboardEvent('keydown',{key:k,bubbles:true,cancelable:true});plainInput.dispatchEvent(e);return e.defaultPrevented;};
out.letterPrevented=prevented('a');
out.arrowPreventedDefault=prevented('ArrowDown');
// Value: protects=Enter on an item that wraps a button clicking that button; fails_when=the selector inside the item is narrowed to links; why_new=only the link and plain-item cases ran; seam=none; no_claim=-
window.clicked='';
out.enterPreventedDefault=prevented('Enter');
out.enterClicksButton=window.clicked;
// Value: protects=the keyboard position while htmx appends items (beforeend, infinite load); fails_when=the observer reselects the first item on every mutation; why_new=only a removed selection was tested; seam=none; no_claim=-
var kept=document.querySelector('#plain [data-selected=true]');
var more=document.createElement('div');
more.setAttribute('data-slot','command-item');
more.textContent='Four';
document.querySelector('#plain [data-slot=command-list]').appendChild(more);
await wait(0);
out.appendKeepsSelection=document.querySelector('#plain [data-selected=true]')===kept;
`
	got := runInChrome(t, body, script, nil)
	checkRecorded(t, got, map[string]any{
		"initial":                 "apple",
		"activedescendant":        "apple",
		"controls":                true,
		"secondInitial":           "only",
		"downSkipsDisabled":       "cherry",
		"downStopsAtEnd":          "docs",
		"home":                    "apple",
		"upStopsAtStart":          "apple",
		"end":                     "docs",
		"enterClicksLink":         "docs",
		"filtered":                "cherry",
		"unfiltered":              "apple",
		"enterClicksItem":         "apple",
		"pointer":                 "cherry",
		"pointerSkipsDisabled":    "cherry",
		"ariaSelected":            "true",
		"previousCleared":         false,
		"noMatch":                 "0 selected",
		"noMatchActivedescendant": false,
		"enterKeyCode229":         "",
		"arrowComposing":          "apple",
		"arrowPrevented":          "apple",
		"swapped":                 "kiwi",
		"swappedActivedescendant": "kiwi",
		"listReplaced":            "plum",
		"listRelinked":            true,
		"noMatchGroups":           "hidden,hidden",
		"noMatchEmpty":            "shown",
		"appGroups":               "shown,hidden",
		"appEmpty":                "hidden",
		"plainID":                 true,
		"plainIDsDiffer":          true,
		"letterPrevented":         false,
		"arrowPreventedDefault":   true,
		"enterPreventedDefault":   true,
		"enterClicksButton":       "button",
		"appendKeepsSelection":    true,
	})
}

// TestCommandLeavesFocusedControlsAndDisabledItemsAloneInChrome checks the
// two cases the selection must not take over: a key pressed on a link or
// button focused inside an item keeps that control's own activation, and an
// item disabled after load stops being selected, so Enter cannot activate it.
//
// Value: protects=Enter on a focused link inside an item running that link, not the selected item, and a disabled item never being activated; fails_when=the keydown handler stops ignoring keys from controls inside items, or the observer stops watching data-disabled; why_new=the keyboard test only sends keys from the input and disables items in the markup; seam=none; no_claim=synthetic events trigger no native link navigation, so the link's own activation is shown by the event not being cancelled
func TestCommandLeavesFocusedControlsAndDisabledItemsAloneInChrome(t *testing.T) {
	body := Command(html.ID("cmd"), CommandInput(), CommandList(
		CommandItem(html.ID("a"), html.OnClick("window.clicked='a'"), html.A(html.HRef("#a"), "A")),
		CommandItem(html.ID("b"), html.A(html.ID("blink"), html.HRef("#b"), "B")),
		CommandItem(html.ID("c"), "C"),
	))
	const script = /*js*/ `
var input=document.querySelector('#cmd [data-slot=command-input]');
var selected=function(){return document.querySelector('#cmd [data-selected=true]').id;};
var link=document.getElementById('blink');
link.focus();
window.clicked='';
var enter=new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true});
link.dispatchEvent(enter);
out.linkEnterPrevented=enter.defaultPrevented;
out.linkEnterClicked=window.clicked;
var down=new KeyboardEvent('keydown',{key:'ArrowDown',bubbles:true,cancelable:true});
link.dispatchEvent(down);
out.linkArrowPrevented=down.defaultPrevented;
out.linkArrowSelected=selected();
// Select the first item from the input, then disable it in place.
input.dispatchEvent(new KeyboardEvent('keydown',{key:'Home',bubbles:true,cancelable:true}));
// Enter in the same task as the disable, before the observer has run, finds
// the disabled item still selected and must leave it alone.
window.clicked='';
document.getElementById('a').setAttribute('data-disabled','true');
var sameTask=new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true});
input.dispatchEvent(sameTask);
out.sameTaskClicked=window.clicked;
out.sameTaskPrevented=sameTask.defaultPrevented;
await wait(0);
out.disabledReselected=selected();
window.clicked='';
input.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));
out.disabledNotClicked=window.clicked;
`
	got := runInChrome(t, body, script, nil)
	checkRecorded(t, got, map[string]any{
		"linkEnterPrevented": false,
		"linkEnterClicked":   "",
		"linkArrowPrevented": false,
		"linkArrowSelected":  "a",
		"disabledReselected": "b",
		"disabledNotClicked": "",
		"sameTaskClicked":    "",
		"sameTaskPrevented":  false,
	})
}
