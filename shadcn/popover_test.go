package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestPopoverComposition(t *testing.T) {
	out := render(t, Popover(
		PopoverTrigger("p1", "Open"),
		PopoverContent("p1", "", "Hello"),
	))
	for _, want := range []string{
		`data-slot="popover"`,
		`data-slot="popover-trigger"`,
		`type="button"`,
		`popovertarget="p1"`,
		`popovertargetaction="toggle"`,
		`aria-haspopup="dialog"`,
		`aria-expanded="false"`,
		"anchor-name: --p1",
		`data-slot="popover-content"`,
		`id="p1"`,
		`popover="auto"`,
		`role="dialog"`,
		"position-anchor: --p1",
		"position-area: bottom",
		"margin-top: 4px",
		">Open<",
		">Hello<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=") {
		t.Errorf("Radix data-[state= should have been dropped: %s", out)
	}
}

func TestPopoverSides(t *testing.T) {
	cases := map[PopoverSide]string{
		PopoverTop:    "position-area: top",
		PopoverRight:  "position-area: right",
		PopoverBottom: "position-area: bottom",
		PopoverLeft:   "position-area: left",
	}
	for side, want := range cases {
		out := render(t, PopoverContent("x", side))
		if !strings.Contains(out, want) {
			t.Errorf("side %q: missing %q in %s", side, want, out)
		}
	}
}

func TestPopoverUnknownSideFallsBackToBottom(t *testing.T) {
	out := render(t, PopoverContent("x", PopoverSide("bogus")))
	if !strings.Contains(out, "position-area: bottom") {
		t.Errorf("unknown side should fall back to bottom: %s", out)
	}
}

func TestPopoverCallerStyleMerges(t *testing.T) {
	// A caller-supplied style is preserved; the anchor fragment is appended.
	out := render(t, PopoverTrigger("p1", html.Style("color: red"), "Open"))
	for _, want := range []string{"color: red", "anchor-name: --p1"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if n := strings.Count(out, "style="); n != 1 {
		t.Errorf("expected one style attribute, got %d: %s", n, out)
	}
}

func TestPopoverValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected PopoverTrigger panic for id %q", bad)
				}
			}()
			PopoverTrigger(bad)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected PopoverContent panic for id %q", bad)
				}
			}()
			PopoverContent(bad, "")
		}()
	}
}

// PopoverContent's default ontoggle keeps aria-expanded of its trigger in
// step, so every PopoverTrigger tells assistive technology whether its
// popover is open; a trigger without aria-expanded is left alone.
//
// Value: protects=aria-expanded on PopoverTrigger following the popover; fails_when=PopoverContent loses its ontoggle or the handler matches the wrong triggers; why_new=only Combobox and the menus updated it before; seam=none; no_claim=screen reader output
func TestPopoverExpandedInChrome(t *testing.T) {
	if out := render(t, PopoverContent("p1", "", html.OnToggle("mine()"))); strings.Count(out, "ontoggle") != 1 || !strings.Contains(out, `ontoggle="mine()"`) {
		t.Errorf("a caller's ontoggle replaces the default: %s", out)
	}
	body := html.Div(
		PopoverTrigger("p-1", html.ID("trigger"), "Open"),
		html.Button(html.ID("plain"), html.PopoverTarget("p-1"), "Plain"),
		PopoverContent("p-1", "", "Hello"),
	)
	const script = /*js*/ `
var t=document.getElementById('trigger'),p=document.getElementById('p-1');
t.click();
await wait(0);
out.open=t.getAttribute('aria-expanded');
out.plain=document.getElementById('plain').hasAttribute('aria-expanded');
p.hidePopover();
await wait(0);
out.closed=t.getAttribute('aria-expanded');
`
	got := runInChrome(t, body, script, nil)
	checkRecorded(t, got, map[string]any{
		"open":   "true",
		"plain":  false,
		"closed": "false",
	})
}
