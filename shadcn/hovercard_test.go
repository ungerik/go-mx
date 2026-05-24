package shadcn

import (
	"strings"
	"testing"
)

func TestHoverCardComposition(t *testing.T) {
	out := render(t, HoverCard(
		HoverCardTrigger("hc1", 0, 0, "@user"),
		HoverCardContent("hc1", "", 0, 0, "user profile card"),
	))
	for _, want := range []string{
		`data-slot="hover-card"`,
		`data-slot="hover-card-trigger"`,
		`<span `,
		`aria-describedby="hc1"`,
		`onmouseover="hoverCardShow(this,'hc1',700)"`, // default open delay
		`onmouseout="hoverCardHide(this,'hc1',300)"`,  // default close delay
		`onfocusin="hoverCardShow(this,'hc1',700)"`,
		`onfocusout="hoverCardHide(this,'hc1',300)"`,
		`data-slot="hover-card-content"`,
		`id="hc1"`,
		`popover="auto"`,
		`role="dialog"`,
		"position-area: bottom",
		"<script>",
		"window.hoverCardShow",
		"window.hoverCardHide",
		">@user<",
		">user profile card<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestHoverCardCustomDelays(t *testing.T) {
	out := render(t, HoverCardTrigger("hc1", 1500, 100, "trigger"))
	for _, want := range []string{
		`hoverCardShow(this,'hc1',1500)`,
		`hoverCardHide(this,'hc1',100)`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestHoverCardValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected HoverCardTrigger panic for id %q", bad)
				}
			}()
			HoverCardTrigger(bad, 0, 0)
		}()
	}
}
