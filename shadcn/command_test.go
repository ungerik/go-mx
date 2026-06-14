package shadcn

import (
	"strings"
	"testing"

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
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
