package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestInputGroup(t *testing.T) {
	out := render(t, InputGroup(
		InputGroupAddon("", "@"),
		InputGroupInput(html.Placeholder("Username")),
	))
	for _, want := range []string{
		`data-slot="input-group"`, `role="group"`,
		// The group carries border + focus ring, keyed off the control slot.
		"has-[[data-slot=input-group-control]:focus-visible]:border-ring",
		`data-slot="input-group-addon"`, `data-align="inline-start"`,
		`data-slot="input-group-control"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestInputGroupInputBorderless(t *testing.T) {
	out := render(t, InputGroupInput())
	// The control must lose Input's own border/ring — the wrapper draws them;
	// a double border is the visible failure.
	for _, borderful := range []string{"rounded-md", "shadow-xs", "focus-visible:ring-[3px]"} {
		if strings.Contains(out, borderful) {
			t.Errorf("Input class %q should have been overridden: %s", borderful, out)
		}
	}
	for _, want := range []string{"rounded-none", "border-0", "shadow-none"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestInputGroupTextarea(t *testing.T) {
	out := render(t, InputGroupTextarea(html.Placeholder("Message")))
	for _, want := range []string{"<textarea", `data-slot="input-group-control"`, "rounded-none", "resize-none"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestInputGroupAddonFocusClick(t *testing.T) {
	out := render(t, InputGroupAddon("", "@"))
	// Clicking the addon focuses the input (port of upstream's onClick),
	// unless the click hit a button inside the addon.
	if !strings.Contains(out, "onclick=") || !strings.Contains(out, "closest('button')") {
		t.Errorf("missing default focus onclick: %s", out)
	}
	// A caller onclick replaces the default.
	custom := render(t, InputGroupAddon("", html.OnClick("doSomething()"), "@"))
	if strings.Contains(custom, "querySelector") {
		t.Errorf("caller onclick should replace the default: %s", custom)
	}
}

func TestInputGroupAddonAligns(t *testing.T) {
	end := render(t, InputGroupAddon(InputGroupAddonInlineEnd, "x"))
	if !strings.Contains(end, `data-align="inline-end"`) || !strings.Contains(end, "order-last") {
		t.Errorf("inline-end align missing: %s", end)
	}
	block := render(t, InputGroupAddon(InputGroupAddonBlockEnd, "x"))
	if !strings.Contains(block, "w-full") || !strings.Contains(block, "order-last") {
		t.Errorf("block-end align should span full width and order last: %s", block)
	}
	blockStart := render(t, InputGroupAddon(InputGroupAddonBlockStart, "x"))
	// block-start is a full-width row above the control: full width but
	// ordered first, the mirror of block-end.
	if !strings.Contains(blockStart, `data-align="block-start"`) ||
		!strings.Contains(blockStart, "w-full") || !strings.Contains(blockStart, "order-first") {
		t.Errorf("block-start align missing full-width/order-first classes: %s", blockStart)
	}
}

func TestInputGroupButton(t *testing.T) {
	out := render(t, InputGroupButton("", "", "Search"))
	for _, want := range []string{
		`data-slot="button"`, `data-variant="ghost"`, `data-size="xs"`, `type="button"`,
		// xs height must win the merge against ButtonClasses' default h-9,
		// or the button would overflow the group's h-9 frame.
		"h-6",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "h-9") {
		t.Errorf("default button height should have been overridden: %s", out)
	}
}

func TestInputGroupButtonSizesAndVariant(t *testing.T) {
	// The non-default sizes each map to a distinct class string; icon-sm is
	// the square size-8 variant.
	iconSM := render(t, InputGroupButton(ButtonDefault, InputGroupButtonIconSM, "×"))
	for _, want := range []string{`data-size="icon-sm"`, `data-variant="default"`, "size-8"} {
		if !strings.Contains(iconSM, want) {
			t.Errorf("missing %q in %s", want, iconSM)
		}
	}
	iconXS := render(t, InputGroupButton("", InputGroupButtonIconXS, "×"))
	if !strings.Contains(iconXS, `data-size="icon-xs"`) || !strings.Contains(iconXS, "size-6") {
		t.Errorf("icon-xs size classes missing: %s", iconXS)
	}
	// A non-ghost variant flows through to the button unchanged (unlike size,
	// only variant "" is defaulted).
	outline := render(t, InputGroupButton(ButtonOutline, InputGroupButtonSM, "Go"))
	if !strings.Contains(outline, `data-variant="outline"`) || !strings.Contains(outline, `data-size="sm"`) {
		t.Errorf("outline/sm should pass through: %s", outline)
	}
}

func TestInputGroupText(t *testing.T) {
	out := render(t, InputGroupText("USD"))
	if !strings.Contains(out, "<span") || !strings.Contains(out, "text-muted-foreground") {
		t.Errorf("missing span or classes: %s", out)
	}
}
