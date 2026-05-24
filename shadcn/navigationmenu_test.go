package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestNavigationMenuComposition(t *testing.T) {
	out := render(t, NavigationMenu(
		NavigationMenuList(
			NavigationMenuItem(
				NavigationMenuTrigger("products", "Products"),
				NavigationMenuContent("products", "",
					NavigationMenuLink(false, html.HRef("/cards"), "Cards"),
					NavigationMenuLink(true, html.HRef("/forms"), "Forms"),
				),
			),
			NavigationMenuItem(
				NavigationMenuLink(false, html.HRef("/docs"), "Docs"),
			),
		),
	))
	for _, want := range []string{
		`<nav `,
		`data-slot="navigation-menu"`,
		`aria-label="Main"`,
		`<ul `,
		`data-slot="navigation-menu-list"`,
		`<li `,
		`data-slot="navigation-menu-item"`,
		`data-slot="navigation-menu-trigger"`,
		`popovertarget="products"`,
		`aria-haspopup="menu"`,
		`aria-expanded="false"`,
		"anchor-name: --products",
		"lucide-chevron-down",
		`data-slot="navigation-menu-content"`,
		`id="products"`,
		`popover="auto"`,
		`role="menu"`,
		`ontoggle="menuOpen(event)"`,
		"position-area: bottom",
		"window.menuKeyNav",
		`data-slot="navigation-menu-link"`,
		`data-active="false"`,
		`data-active="true"`,
		`aria-current="page"`,
		`href="/cards"`,
		`href="/forms"`,
		`href="/docs"`,
		">Products<",
		">Cards<",
		">Forms<",
		">Docs<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=open]") {
		t.Errorf("Radix data-[state=open] should have been rewritten: %s", out)
	}
}

func TestNavigationMenuTriggerValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected NavigationMenuTrigger panic for id %q", bad)
				}
			}()
			NavigationMenuTrigger(bad)
		}()
	}
}
