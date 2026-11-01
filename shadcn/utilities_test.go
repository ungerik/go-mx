package shadcn

import (
	"strings"
	"testing"
)

// UtilitiesCSS carries upstream's shimmer and scroll-fade utilities, and
// UseUtilities swaps the components' substitutes for them.
func TestUseUtilities(t *testing.T) {
	for _, want := range []string{"@utility shimmer {", "@utility scroll-fade-x {", "@utility scroll-fade-b {", "@keyframes tw-shimmer", "@property --scroll-fade-t"} {
		if !strings.Contains(UtilitiesCSS, want) {
			t.Errorf("UtilitiesCSS lacks %s", want)
		}
	}
	line := func() string { return render(t, TraceLine(TraceRunning, TraceTitle("x"))) }
	if out := line(); !strings.Contains(out, "animate-pulse") || strings.Contains(out, "shimmer") {
		t.Errorf("by default the title pulses: %s", out)
	}
	UseUtilities = true
	defer func() { UseUtilities = false }()
	if out := line(); !strings.Contains(out, "data-[slot=trace-title]:shimmer") || strings.Contains(out, "animate-pulse") {
		t.Errorf("with UseUtilities the title shimmers: %s", out)
	}
	for name, out := range map[string]string{
		"AttachmentGroup":         render(t, AttachmentGroup()),
		"CodeBlockContent":        render(t, CodeBlockContent("x")),
		"MessageScrollerViewport": render(t, MessageScrollerViewport()),
	} {
		if !strings.Contains(out, "scroll-fade-") {
			t.Errorf("%s fades its edges with UseUtilities: %s", name, out)
		}
	}
}
