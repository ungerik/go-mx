package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
)

func TestTabsComposition(t *testing.T) {
	out := render(t, Tabs("settings",
		TabsList(
			TabsTrigger("settings", "account", true, "Account"),
			TabsTrigger("settings", "billing", false, "Billing"),
		),
		TabsContent("settings", "account", true, "Account panel"),
		TabsContent("settings", "billing", false, "Billing panel"),
	))
	for _, want := range []string{
		`data-slot="tabs"`,
		`data-tabs="settings"`,
		`role="tablist"`,
		`data-slot="tabs-trigger"`,
		`role="tab"`,
		`type="button"`,
		`aria-selected="true"`,
		`aria-selected="false"`,
		`aria-controls="settings-panel-account"`,
		`id="settings-tab-account"`,
		`data-tabs-value="account"`,
		`tabindex="-1"`,                          // inactive trigger
		`onclick="tabsSelect('settings','account')"`,
		"aria-selected:bg-background",            // rewrite from data-[state=active]
		`role="tabpanel"`,
		`id="settings-panel-account"`,
		`aria-labelledby="settings-tab-account"`,
		`tabindex="0"`,                            // content default
		"hidden",                                  // inactive panel
		">Account panel<",
		">Billing panel<",
		"<script>",
		"window.tabsSelect",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=active]") {
		t.Errorf("Radix data-[state=active] should have been rewritten: %s", out)
	}
	if n := strings.Count(out, "<script>"); n != 1 {
		t.Errorf("tabsSelect script should be emitted exactly once, got %d: %s", n, out)
	}
}

func TestTabsTriggerHXOptOut(t *testing.T) {
	out := render(t, TabsTrigger("s", "account", false,
		hx.Get("/tabs/account"), hx.Target("#panel"), "Account"))
	for _, want := range []string{
		`hx-get="/tabs/account"`,
		`hx-target="#panel"`,
		`aria-selected="false"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "onclick=") {
		t.Errorf("default onclick should be skipped when hx-* is present: %s", out)
	}
}

func TestTabsValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected Tabs panic for id %q", bad)
				}
			}()
			Tabs(bad)
		}()
	}
}

func TestTabsCallerOnClickOverridesDefault(t *testing.T) {
	out := render(t, TabsTrigger("s", "account", true, html.OnClick("custom()"), "Account"))
	if !strings.Contains(out, `onclick="custom()"`) {
		t.Errorf("caller onclick should win: %s", out)
	}
	if strings.Contains(out, "tabsSelect(") {
		t.Errorf("default tabsSelect onclick should be skipped: %s", out)
	}
}
