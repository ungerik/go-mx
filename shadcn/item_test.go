package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestItem(t *testing.T) {
	out := render(t, Item("", "",
		ItemMedia(ItemMediaIcon, "i"),
		ItemContent(
			ItemTitle("Title"),
			ItemDescription("Description"),
		),
		ItemActions(Button(ButtonOutline, SizeSM, "Open")),
	))
	for _, want := range []string{
		`data-slot="item"`, `data-variant="default"`, `data-size="default"`,
		`data-slot="item-media"`, `data-slot="item-content"`,
		`data-slot="item-title"`, `data-slot="item-description"`,
		`data-slot="item-actions"`,
		"group/item", "gap-3.5",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// ItemDescription is a <p> like upstream (the one non-div part).
	if !strings.Contains(out, "<p") {
		t.Errorf("ItemDescription should render a <p>: %s", out)
	}
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestItemVariantsAndSizes(t *testing.T) {
	outline := render(t, Item(ItemOutline, ItemSizeSM, "x"))
	for _, want := range []string{"border-border", "px-3", `data-variant="outline"`, `data-size="sm"`} {
		if !strings.Contains(outline, want) {
			t.Errorf("missing %q in %s", want, outline)
		}
	}
	muted := render(t, Item(ItemMuted, ItemSizeXS, "x"))
	if !strings.Contains(muted, "bg-muted/50") || !strings.Contains(muted, "px-2.5") {
		t.Errorf("muted xs item missing variant/size classes: %s", muted)
	}
}

func TestItemGroup(t *testing.T) {
	out := render(t, ItemGroup(Item("", "", "a"), ItemSeparator(), Item("", "", "b")))
	// role="list" is what makes the group announce as a list without <ul>/<li>.
	if !strings.Contains(out, `role="list"`) {
		t.Errorf("missing role list: %s", out)
	}
	// The compact-gap selectors must be scoped to items: this port's Button
	// also emits data-size, so upstream's bare has-data-[size=sm]: would let
	// a small button inside an item collapse the whole group's gap.
	if strings.Contains(out, "has-data-[size=") {
		t.Errorf("group gap selector must be scoped to [data-slot=item]: %s", out)
	}
	if !strings.Contains(out, "has-[[data-slot=item][data-size=sm]]:gap-2.5") {
		t.Errorf("missing item-scoped gap selector: %s", out)
	}
	// ItemSeparator keeps Separator's markup but re-tags the slot so the
	// group's data-slot-based selectors can tell it apart from items.
	if !strings.Contains(out, `data-slot="item-separator"`) {
		t.Errorf("missing item-separator slot: %s", out)
	}
	if strings.Contains(out, `data-slot="separator"`) {
		t.Errorf("separator slot should have been replaced: %s", out)
	}
}

func TestItemMediaImageVariant(t *testing.T) {
	out := render(t, ItemMedia(ItemMediaImage, html.Img(html.Src("/a.png"), html.Alt(""))))
	if !strings.Contains(out, "overflow-hidden") || !strings.Contains(out, `data-variant="image"`) {
		t.Errorf("image variant missing classes: %s", out)
	}
}

func TestItemMediaDefaultVariant(t *testing.T) {
	// An empty (or unknown) variant resolves to the unframed default.
	out := render(t, ItemMedia("", "i"))
	if !strings.Contains(out, `data-variant="default"`) || !strings.Contains(out, "bg-transparent") {
		t.Errorf("default media variant missing classes: %s", out)
	}
}

func TestItemHeaderFooter(t *testing.T) {
	// Header and footer both span the full item width (basis-full) so they
	// wrap above/below the media+content+actions row.
	header := render(t, ItemHeader("Top"))
	if !strings.Contains(header, `data-slot="item-header"`) || !strings.Contains(header, "basis-full") {
		t.Errorf("item-header slot/classes missing: %s", header)
	}
	footer := render(t, ItemFooter("Bottom"))
	if !strings.Contains(footer, `data-slot="item-footer"`) || !strings.Contains(footer, "basis-full") {
		t.Errorf("item-footer slot/classes missing: %s", footer)
	}
}
