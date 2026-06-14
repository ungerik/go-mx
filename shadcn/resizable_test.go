package shadcn

import (
	"strings"
	"testing"
)

func TestResizableComposition(t *testing.T) {
	out := render(t, ResizablePanelGroup(ResizeHorizontal,
		ResizablePanel("One"),
		ResizableHandle(),
		ResizablePanel("Two"),
	))
	for _, want := range []string{
		`data-slot="resizable-panel-group"`,
		`data-direction="horizontal"`,
		"group/resizable",
		`data-slot="resizable-panel"`,
		"flex: 1 1 0",
		`data-slot="resizable-handle"`,
		`onpointerdown="resizeStart(event,this)"`,
		`role="separator"`,
		"lucide-grip-vertical",
		"window.resizeStart",
		">One<",
		">Two<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestResizableVertical(t *testing.T) {
	out := render(t, ResizablePanelGroup(ResizeVertical))
	if !strings.Contains(out, `data-direction="vertical"`) {
		t.Errorf("expected vertical direction: %s", out)
	}
}
