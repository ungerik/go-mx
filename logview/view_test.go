package logview

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
)

func TestViewWiresTheStream(t *testing.T) {
	out := render(t, View(hx.Ext("sse"), hx.SSEConnect("/logs"), hx.SSEClose("done")))
	contains(t, out,
		// Caller attributes land on the root, so one connection feeds the view.
		`hx-ext="sse"`,
		`sse-connect="/logs"`,
		`sse-close="done"`,
		// Appending rather than replacing is what makes this a log.
		`sse-swap="log"`,
		`hx-swap="beforeend"`,
		// An in-band error has nowhere else to land once the 200 is committed.
		`sse-swap="`+mx.SSEEventError+`"`,
		`role="alert"`,
		// The stream is followed by the shadcn scroll area's anchoring.
		`data-stick-to-bottom=""`,
		"window.mxStickToBottom",
	)
}

func TestViewScrollAreaHasAHeight(t *testing.T) {
	// A scroll area without a height grows instead of scrolling, and one without
	// overflow is clipped by the view instead of scrolling. Both leave
	// scrollHeight equal to clientHeight, so stick-to-bottom has nothing to
	// follow — and neither can be left to the Tailwind-only shadcn classes,
	// because a page without that build would then not scroll at all.
	// Anchored on style=" so this cannot pass on the same text appearing in a
	// class name, a data attribute or an inline <style> block: what is being
	// checked is that the declarations reach the element itself.
	contains(t, render(t, View()), `style="flex: 1 1 auto; min-height: 0; height: 24rem; overflow: auto"`)
	contains(t, render(t, (&Config{Height: "40vh"}).View()), `style="flex: 1 1 auto; min-height: 0; height: 40vh; overflow: auto"`)
	// The view is a flex column so a page can give it a height of its own and
	// have the scroll area fill what the toolbar and error sink leave.
	contains(t, DarkTheme.CSS(""), "\tdisplay: flex;\n\tflex-direction: column;\n")
}

func TestViewAccessibility(t *testing.T) {
	// Without a live region a screen reader either says nothing as the log
	// grows or re-reads the whole transcript on every line.
	contains(t, render(t, View()),
		`role="log"`,
		`aria-live="polite"`,
		`aria-relevant="additions"`,
		// The paused state is a status region rather than aria-pressed on the
		// button: the button's caption names the next action, and a pressed
		// button captioned "Resume" would tell a screen reader the opposite of
		// what it tells the eye.
		`role="status"`,
		`data-mx-log-status="Paused"`,
		`aria-label="Filter log lines"`,
		`type="search"`,
	)
	// The two mechanisms cannot be combined, so the absence is the contract:
	// re-adding aria-pressed next to the swapping caption is the regression.
	excludes(t, render(t, View()), `aria-pressed`)
}

func TestViewCarriesTheBehaviorContract(t *testing.T) {
	// The script finds everything it drives through these attributes, so the
	// markup and the script have to agree on every one of them.
	out := render(t, (&Config{MaxLines: 500}).View())
	contains(t, out,
		`data-mx-log-view=""`,
		`data-mx-log-max-lines="500"`,
		`data-mx-log-lines=""`,
		`data-mx-log-filter=""`,
		`data-mx-log-pause="Resume"`,
		`data-mx-log-status="Paused"`,
		`data-mx-log-badge="new"`,
	)
}

func TestViewPauseCarriesBothCaptions(t *testing.T) {
	// The button swaps its caption while paused, so both captions have to be in
	// the markup — the script must not carry a user-visible string of its own,
	// or a translated view would toggle back into English.
	out := render(t, View())
	contains(t, out, `data-mx-log-pause="Resume"`, `>Pause</button>`)
}

func TestViewScriptGuardsItsDefinition(t *testing.T) {
	// Two views on one page must install the behavior once but wire up both,
	// which is why the definition is guarded and the scan is not.
	out := render(t, mx.Components{View(), View()})
	if got := strings.Count(out, "if(!window.mxLogView)"); got != 2 {
		t.Errorf("guard appears %d times, want 2 (once per emitted script)", got)
	}
	if got := strings.Count(out, "window.mxLogViewScan(document);</script>"); got != 2 {
		t.Errorf("the unguarded scan runs %d times, want once per instance", got)
	}
}

func TestViewScriptComesLast(t *testing.T) {
	// The script queries the view when it runs, so it has to be after the
	// elements it looks for.
	out := render(t, View())
	if strings.Index(out, "data-mx-log-lines") > strings.Index(out, "if(!window.mxLogView)") {
		t.Errorf("the view script is emitted before the line container: %s", out)
	}
}

func TestViewBacklogGoesIntoTheLineContainer(t *testing.T) {
	// Caller children are lines already on the page, so they belong where the
	// streamed ones are appended — not next to the toolbar.
	out := render(t, View(Lines(`{"level":"info","msg":"one"}`, `{"msg":"two"}`)))
	lines := strings.Index(out, `data-mx-log-lines`)
	if lines < 0 {
		t.Fatalf("no line container: %s", out)
	}
	first := strings.Index(out, ">one<")
	if first < lines {
		t.Errorf("backlog rendered before the line container: %s", out)
	}
	if got := strings.Count(out, `class="log-line"`); got != 2 {
		t.Errorf("backlog rendered %d lines, want 2: %s", got, out)
	}
}

func TestViewMergesACallerClass(t *testing.T) {
	// Appending a second class attribute would fail the duplicate-attribute
	// check that render() surfaces, so a caller could not style the view at all.
	contains(t, render(t, View(html.Class("mt-4 border"))), `class="log-view mt-4 border"`)
}

func TestViewCustomPrefix(t *testing.T) {
	// The prefix renames the appearance classes only: the behavior attributes
	// stay fixed so one script definition serves every view on the page.
	out := render(t, (&Config{Prefix: "mylog-"}).View())
	contains(t, out, `class="mylog-view"`, `class="mylog-toolbar"`, `data-mx-log-view=""`)
	excludes(t, out, `class="log-view"`)
}

func TestViewCustomLabels(t *testing.T) {
	// The strings a view renders on its own are configurable, and the count
	// suffix travels in the badge attribute so the script holds no English.
	out := render(t, (&Config{Labels: Labels{
		Filter:     "Protokoll filtern",
		FilterHint: "Filtern…",
		Pause:      "Anhalten",
		Resume:     "Fortsetzen",
		Paused:     "Angehalten",
		NewLines:   "neu",
	}}).View())
	contains(t, out,
		`aria-label="Protokoll filtern"`,
		`placeholder="Filtern…"`,
		`>Anhalten<`,
		`data-mx-log-pause="Fortsetzen"`,
		`data-mx-log-status="Angehalten"`,
		`data-mx-log-badge="neu"`,
	)
}

func TestViewCustomEvent(t *testing.T) {
	contains(t, render(t, (&Config{Event: "syslog"}).View()), `sse-swap="syslog"`)
}
