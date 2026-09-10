package logview

import (
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

// render renders a component with a plain CheckedWriter and fails the test on
// any render error, which includes the duplicate-attribute check.
func render(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render error: %v\npartial output:\n%s", err, b.String())
	}
	return b.String()
}

// renderIndented renders through an indenting writer, which is what a page
// using mx.NewCheckedWriter(w).WithIndent does to everything it contains.
func renderIndented(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), mx.NewCheckedWriter(&b).WithIndent("", "  ")); err != nil {
		t.Fatalf("render error: %v\npartial output:\n%s", err, b.String())
	}
	return b.String()
}

func contains(t *testing.T, out string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q in %s", w, out)
		}
	}
}

func excludes(t *testing.T, out string, unwanted ...string) {
	t.Helper()
	for _, u := range unwanted {
		if strings.Contains(out, u) {
			t.Errorf("unexpected %q in %s", u, out)
		}
	}
}

func TestLineValueTypesAreColoredByType(t *testing.T) {
	// The type of a value is conveyed by its class alone, which is the reason
	// string values need no quotes to be readable.
	out := render(t, Line(`{"s":"text","n":12,"f":1.5,"b":true,"z":null}`))
	contains(t, out,
		`<span class="log-key">s</span><span class="log-punct">=</span><span class="log-string">text</span>`,
		`<span class="log-number">12</span>`,
		`<span class="log-number">1.5</span>`,
		`<span class="log-bool">true</span>`,
		`<span class="log-null">null</span>`,
	)
}

func TestLineKeepsNumberLiterals(t *testing.T) {
	// A snowflake id is a JSON number, and rendering it through float64 would
	// silently change the value the operator is trying to grep for.
	out := render(t, Line(`{"id":1234567890123456789}`))
	contains(t, out, `<span class="log-number">1234567890123456789</span>`)
	excludes(t, out, "e+18")
}

func TestLineQuoteText(t *testing.T) {
	// Quoting is opt-in because the color already says "string"; a caller who
	// needs to see leading or trailing spaces turns it on.
	unquoted := render(t, Line(`{"msg":"hi there"}`))
	contains(t, unquoted, `<span class="log-string">hi there</span>`)

	cfg := &Config{QuoteText: true}
	contains(t, render(t, cfg.Line(`{"msg":"hi there"}`)),
		`<span class="log-string">&quot;hi there&quot;</span>`)
}

func TestLineTimeLabelOnlyWhenNotFirst(t *testing.T) {
	// A leading timestamp reads as one on its own, so the label is noise. Later
	// in the record the same value is just another field and needs its key.
	first := render(t, Line(`{"time":"10:04:02","msg":"x"}`))
	contains(t, first, `<span class="log-time">10:04:02</span>`)
	excludes(t, first, `<span class="log-key">time</span>`)

	later := render(t, Line(`{"msg":"x","time":"10:04:02"}`))
	contains(t, later,
		`<span class="log-key">time</span><span class="log-punct">=</span><span class="log-time">10:04:02</span>`)
}

func TestLineLevelNeedsNoLabel(t *testing.T) {
	// The level is recognizable by its color, and "level=" in front of every
	// line is pure repetition.
	out := render(t, Line(`{"level":"warn","msg":"x"}`))
	contains(t, out,
		`<span class="log-level log-level-warn">warn</span>`,
		`data-mx-log-level="warn"`,
	)
	excludes(t, out, `<span class="log-key">level</span>`)
}

func TestLineMessageKeyIsOptIn(t *testing.T) {
	// The default matches the specified behavior — only time and level are
	// promoted — while slog and zerolog users opt their message key in.
	contains(t, render(t, Line(`{"msg":"boom"}`)), `<span class="log-key">msg</span>`)

	cfg := &Config{MessageKey: "msg"}
	out := render(t, cfg.Line(`{"msg":"boom"}`))
	contains(t, out, `<span class="log-message">boom</span>`)
	excludes(t, out, `<span class="log-key">msg</span>`)
}

func TestLineHostileLevelCannotReachTheClassAttribute(t *testing.T) {
	// A level comes from whoever wrote the log line. A space would end the
	// class token and let the rest of the value name arbitrary CSS classes —
	// "hidden" alone would make the line disappear from the view.
	for _, level := range []string{
		"foo hidden",
		"fixed inset-0 z-50",
		"info\"><script>",
		"unknown",
		"",
		"a-really-long-level-value-that-goes-well-past-the-limit",
	} {
		out := render(t, Line(`{"level":`+quoteJSON(level)+`,"msg":"x"}`))
		contains(t, out, `class="log-level log-level-undefined"`)
		if got := strings.Count(out, "log-level-"); got != 1 {
			t.Errorf("level %q produced %d level classes, want 1: %s", level, got, out)
		}
		if strings.Contains(out, `data-mx-log-level="`+level+`"`) {
			t.Errorf("level %q reached an attribute verbatim: %s", level, out)
		}
	}
}

func TestLineLevelClassSetIsClosed(t *testing.T) {
	// The guarantee is not "we validate the value" but "we only ever emit a
	// token the theme defined", so the emitted class set is enumerable.
	cfg := &Config{Theme: Theme{Levels: map[string]LevelStyle{
		"info": {Style: Style{Color: "#fff"}},
		// A theme key that is not a legal class token must not open the hole
		// the level values are kept out of.
		"bad key": {Style: Style{Color: "#fff"}},
	}}}
	contains(t, render(t, cfg.Line(`{"level":"info"}`)), `class="log-level log-level-info"`)

	out := render(t, cfg.Line(`{"level":"bad key"}`))
	contains(t, out, `class="log-level log-level-undefined"`)
	excludes(t, out, "log-level-bad")

	// The illegal key is refused by the stylesheet for the same reason, so the
	// two sides cannot disagree about which classes exist.
	excludes(t, cfg.theme().CSS(""), "log-level-bad")
}

func TestLineLevelImageReplacesTheText(t *testing.T) {
	// The image stands in for the level, so the level is what a screen reader
	// should hear in its place.
	cfg := &Config{Theme: Theme{Levels: map[string]LevelStyle{
		"error": {Image: "/icons/error.png"},
	}}}
	out := render(t, cfg.Line(`{"level":"ERROR","msg":"x"}`))
	contains(t, out,
		`<img class="log-level-image" src="/icons/error.png" alt="ERROR"/>`,
		`class="log-level log-level-error"`,
	)
	excludes(t, out, `>ERROR<`)
}

func TestLineNestedObjectAndArray(t *testing.T) {
	// A nested object is formatted like the record itself so the eye reads it
	// the same way; an array keeps its brackets so it is not mistaken for one.
	out := render(t, Line(`{"attrs":{"host":"db1","port":5432},"tags":["a","b"]}`))
	contains(t, out,
		`<span class="log-punct">{</span><span class="log-key">host</span>`,
		`<span class="log-number">5432</span><span class="log-punct">}</span>`,
		`<span class="log-punct">[</span><span class="log-string">a</span><span class="log-punct">, </span>`,
	)
}

func TestLineNestedLevelIsNotThePromotedLevel(t *testing.T) {
	// A "level" inside an attached object is a field of that object, not the
	// severity of the line, so promotion is top-level only.
	out := render(t, Line(`{"attrs":{"level":"error"}}`))
	contains(t, out, `<span class="log-key">level</span>`)
	excludes(t, out, "data-mx-log-level")
}

func TestLineDepthCapStopsDescending(t *testing.T) {
	// Unbounded recursion on attacker-controlled nesting would overflow the
	// goroutine stack, which is a fatal runtime error no recover can catch.
	deep := strings.Repeat(`{"a":`, 5000) + `1` + strings.Repeat(`}`, 5000)
	out := render(t, Line(`{"v":`+deep+`}`))
	if !strings.Contains(out, "log-line") {
		t.Errorf("a deeply nested line did not render: %s", out[:min(len(out), 200)])
	}
}

func TestLineDuplicateAndEmptyKeys(t *testing.T) {
	// Duplicates are kept in order rather than resolved: two "err" fields in
	// one record is a bug worth seeing, not one to hide.
	out := render(t, Line(`{"err":"first","err":"second","":"anon"}`))
	contains(t, out, ">first<", ">second<", ">anon<")
	if got := strings.Count(out, `<span class="log-key">err</span>`); got != 2 {
		t.Errorf("duplicate key rendered %d times, want 2: %s", got, out)
	}
}

func TestLineFallsBackToPlainTextForNonRecords(t *testing.T) {
	// Rendering part of a line and dropping the rest would be worse than not
	// parsing it at all, so anything that is not exactly one object is text.
	for _, raw := range []string{
		`{"a":1} trailing junk`,
		`{"a":1`,
		`[1,2,3]`,
		`not json at all`,
		`{"a":}`,
	} {
		out := render(t, Line(raw))
		contains(t, out, `<span class="log-text">`)
		excludes(t, out, `<span class="log-key">`)
	}
}

func TestLineTruncatesLongLines(t *testing.T) {
	// One runaway line must not be able to fill the DOM, nor make every filter
	// keystroke walk megabytes of text.
	cfg := &Config{MaxLineLen: 32}
	out := render(t, cfg.Line(strings.Repeat("x", 500)))
	contains(t, out, "…")
	if len(out) > 300 {
		t.Errorf("truncation did not bound the output, got %d bytes: %s", len(out), out)
	}
}

func TestLineTruncationKeepsValidUTF8(t *testing.T) {
	// Cutting mid-rune would put invalid UTF-8 in the page, which is not
	// something a log line should be able to cause.
	cfg := &Config{MaxLineLen: 4}
	out := render(t, cfg.Line("äöüß"))
	if !strings.ContainsRune(out, '…') {
		t.Fatalf("expected a truncated line, got %s", out)
	}
	for _, r := range out {
		if r == '�' {
			t.Errorf("truncation produced invalid UTF-8: %q", out)
		}
	}
}

func TestLineMultilineStringKeepsIndentationUnderAnyWriter(t *testing.T) {
	// A <pre> with a text-only child is the whole trick: an indenting writer
	// breaks before a child element's start tag, which inside a <pre> would be
	// visible whitespace. A stack trace is exactly the value that cannot
	// survive that.
	line := Line(`{"err":"panic: boom\n\tmain.go:12\n\tmain.go:44"}`)
	want := "<pre class=\"log-multiline\">panic: boom\n\tmain.go:12\n\tmain.go:44</pre>"
	contains(t, render(t, line), want)
	contains(t, renderIndented(t, line), want)
}

func TestLineMultilineKeepsALeadingBlankLine(t *testing.T) {
	// HTML drops the newline directly after a <pre> start tag, so a value that
	// begins with a blank line needs a second one to keep it.
	contains(t, render(t, Line(`{"err":"\nsecond"}`)),
		"<pre class=\"log-multiline\">\n\nsecond</pre>")
}

func TestLineNormalizesCarriageReturns(t *testing.T) {
	// The SSE transport splits event data on CR, LF and CRLF alike and the
	// client rejoins with LF, so normalizing here makes the markup equal to
	// what the client reconstructs whether it was streamed or served.
	contains(t, render(t, Line(`{"err":"a\r\nb\rc"}`)),
		"<pre class=\"log-multiline\">a\nb\nc</pre>")
}

func TestLineNeverEndsInANewline(t *testing.T) {
	// A trailing newline becomes an extra "data:" line on the wire and a stray
	// text node between every pair of lines in the DOM.
	for _, raw := range []string{`{"a":1}`, "plain text", `{"e":"multi\nline"}`} {
		for _, out := range []string{render(t, Line(raw)), renderIndented(t, Line(raw))} {
			if strings.HasSuffix(out, "\n") {
				t.Errorf("line %q rendered with a trailing newline: %q", raw, out)
			}
		}
	}
}

func TestLinesBatchesWithoutSeparators(t *testing.T) {
	// A batch is one component so a burst can be one SSE event; htmx inserts
	// its element children individually, so nothing may separate them.
	out := render(t, Lines(`{"a":1}`, `{"a":2}`))
	if got := strings.Count(out, `class="log-line"`); got != 2 {
		t.Errorf("batch rendered %d lines, want 2: %s", got, out)
	}
	contains(t, out, "</div><div")
}

func TestPlainTextLevelDetection(t *testing.T) {
	// A mixed stream should read consistently: the level word of a plain line
	// gets the same color as a JSON level, brackets and all.
	for _, tc := range []struct{ line, level, display string }{
		{"ERROR: disk full", "error", "ERROR:"},
		{"[WARN] almost full", "warn", "[WARN]"},
		// The display text is escaped like any other content, brackets included.
		{"<debug> connecting", "debug", "&lt;debug&gt;"},
		{"info starting up", "info", "info"},
	} {
		out := render(t, Line(tc.line))
		contains(t, out,
			`data-mx-log-level="`+tc.level+`"`,
			`<span class="log-level log-level-`+tc.level+`">`+tc.display+`</span>`,
		)
	}
}

func TestPlainTextLevelDetectionLeavesProseAlone(t *testing.T) {
	// Matching anywhere, or without a delimiter, would color ordinary prose and
	// make the level column meaningless.
	for _, line := range []string{
		"Errors happened while starting", // no delimiter after the word
		"could not connect: ERROR",       // not at the start
		"WARNING_SIGN raised",            // the word continues
		"notice: nothing to do",          // not in the vocabulary
	} {
		out := render(t, Line(line))
		excludes(t, out, "data-mx-log-level", "log-level-")
	}
}

func TestPlainTextLevelVocabularyIsIndependentOfTheTheme(t *testing.T) {
	// Changing colors must never change how a line parses, so the vocabulary
	// lives on the Config and the theme only styles what the parser found.
	cfg := &Config{Levels: []string{}}
	excludes(t, render(t, cfg.Line("ERROR: disk full")), "log-level-")

	cfg = &Config{Levels: []string{"notice"}}
	contains(t, render(t, cfg.Line("notice: nothing to do")), `data-mx-log-level="undefined"`)
}

func TestLineEscapesContent(t *testing.T) {
	// Log content is untrusted by definition, and it reaches the page as text
	// nodes for exactly that reason.
	out := render(t, Line(`{"msg":"<script>alert(1)</script>"}`))
	excludes(t, out, "<script>")
	contains(t, out, "&lt;script&gt;")
}

// quoteJSON is a minimal JSON string encoder for building test input.
func quoteJSON(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func TestLineInheritsTheCallersWriterConfigExceptIndentation(t *testing.T) {
	// A line renders through a writer of its own only to escape the caller's
	// indentation. Everything else that writer was configured with — here the
	// quote style, but equally the text escaper and the element allow-list —
	// describes the whole document, so a line may not opt out of it and emit
	// markup in a second style inside a page written in the first.
	var b strings.Builder
	w := mx.NewCheckedWriter(&b).WithSingleQuoteAttribs().WithIndent("", "  ")
	if err := Line(`{"a":1}`).Render(context.Background(), w); err != nil {
		t.Fatalf("render error: %v", err)
	}
	out := b.String()
	contains(t, out, `class='log-line'`)
	excludes(t, out, `class="log-line"`, "\n")
}
