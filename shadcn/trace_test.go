package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestTrace(t *testing.T) {
	out := render(t, Trace(
		TraceLine(TraceRunning, TraceIcon("i"), TraceTitle("Scanning 52 files"), TraceDescription("2s")),
		TraceLine("", TraceTitle("Done")),
	))
	for _, want := range []string{
		`data-slot="trace"`, `data-slot="trace-line"`, `data-status="running"`,
		`data-slot="trace-icon"`, `aria-hidden="true"`, `data-slot="trace-title"`,
		`data-slot="trace-description"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// A finished step renders no status at all, like upstream's undefined.
	if strings.Count(out, "data-status=") != 1 {
		t.Errorf("only the running line should carry data-status: %s", out)
	}
	// shimmer is a utility of shadcn's own tailwind.css, which this package
	// does not ship; the running title pulses instead.
	if strings.Contains(out, "shimmer") || !strings.Contains(out, "motion-safe:data-[status=running]:*:data-[slot=trace-title]:animate-pulse") {
		t.Errorf("running title should pulse, not shimmer: %s", out)
	}
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestTraceLineUnknownStatus(t *testing.T) {
	if out := render(t, TraceLine("done", "x")); strings.Contains(out, "data-status") {
		t.Errorf("unknown status should render none: %s", out)
	}
}

func TestTraceCollapsible(t *testing.T) {
	out := render(t, TraceCollapsible(html.Open,
		TraceCollapsibleTrigger(TraceLine(TraceError, TraceTitle("Failed"))),
		TraceCollapsibleContent(html.Class("flex gap-1"), html.ID("out"), "command not found"),
	))
	for _, want := range []string{
		`<details data-slot="trace-collapsible"`, " open",
		`<summary data-slot="trace-collapsible-trigger"`,
		`data-slot="trace-collapsible-content"`, `data-slot="trace-collapsible-content-inner"`,
		"lucide-chevron-right", "has-[[data-status=error]]:text-destructive",
		// The browser marker is hidden in favor of the chevron.
		"[&amp;::-webkit-details-marker]:hidden",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The chevron turns with the <details> directly around it, so a nested
	// collapsible cannot turn it; there is no aria-expanded on a <summary>.
	if !strings.Contains(out, "[details[open]>summary>&amp;]:rotate-90") || strings.Contains(out, "aria-expanded") {
		t.Errorf("chevron rotation selector wrong: %s", out)
	}
	// Base UI's data-open/data-*-style transition hooks have no target here.
	for _, gone := range []string{"data-open:", "data-ending-style", "data-starting-style", "--collapsible-panel-height"} {
		if strings.Contains(out, gone) {
			t.Errorf("Base UI selector %q not rewritten: %s", gone, out)
		}
	}
	// The caller class goes to the inner div, the other attributes to the
	// outer content element.
	outer, inner, _ := strings.Cut(out, `data-slot="trace-collapsible-content-inner"`)
	if !strings.Contains(outer, `id="out"`) || strings.Contains(outer[strings.LastIndex(outer, "<div"):], "flex gap-1") {
		t.Errorf("id belongs on the outer content: %s", out)
	}
	if !strings.Contains(inner, "flex gap-1") || !strings.Contains(inner, "command not found") {
		t.Errorf("class and children belong on the inner div: %s", out)
	}
}

func TestTraceLineHoverInSummary(t *testing.T) {
	// Upstream's trigger is a <button>; this port's is a <summary>, which the
	// row's hover selectors must include.
	out := render(t, TraceLine("", "x"))
	if !strings.Contains(out, "in-[a:hover,button:hover,summary:hover]:text-foreground") {
		t.Errorf("summary missing from the hover selector: %s", out)
	}
}
