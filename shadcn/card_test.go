package shadcn

import (
	"strings"
	"testing"
)

func TestCard(t *testing.T) {
	out := render(t, Card(
		CardHeader(
			CardTitle("Title"),
			CardDescription("Description"),
			CardAction("Action"),
		),
		CardContent("Body"),
		CardFooter("Footer"),
	))
	for _, want := range []string{
		`data-slot="card"`,
		`data-slot="card-header"`,
		`data-slot="card-title"`,
		`data-slot="card-description"`,
		`data-slot="card-action"`,
		`data-slot="card-content"`,
		`data-slot="card-footer"`,
		">Title<",
		">Body<",
		">Footer<",
		"bg-card",
		"rounded-xl",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
