package shadcn

import (
	"strings"
	"testing"
)

func TestTooltipComposition(t *testing.T) {
	out := render(t, Tooltip(
		TooltipTrigger("tt1", "Hover me"),
		TooltipContent("tt1", "", "Hint!"),
	))
	for _, want := range []string{
		`data-slot="tooltip"`,
		`data-slot="tooltip-trigger"`,
		`<span `,
		`aria-describedby="tt1"`,
		`onmouseover="tooltipShow(this,'tt1')"`,
		`onmouseout="tooltipHide(this,'tt1')"`,
		`onfocusin="tooltipShow(this,'tt1')"`,
		`onfocusout="tooltipHide(this,'tt1')"`,
		`data-slot="tooltip-content"`,
		`id="tt1"`,
		`popover="auto"`,
		`role="tooltip"`,
		"position-area: top", // default side
		"<script>",
		"window.tooltipShow",
		"window.tooltipHide",
		">Hover me<",
		">Hint!<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=") {
		t.Errorf("Radix data-[state= should have been dropped: %s", out)
	}
}

func TestTooltipDefaultSideIsTop(t *testing.T) {
	out := render(t, TooltipContent("tt1", ""))
	if !strings.Contains(out, "position-area: top") {
		t.Errorf("default tooltip side should be top: %s", out)
	}
}

func TestTooltipValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected TooltipTrigger panic for id %q", bad)
				}
			}()
			TooltipTrigger(bad)
		}()
	}
}
