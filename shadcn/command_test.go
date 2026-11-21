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
	})
}
