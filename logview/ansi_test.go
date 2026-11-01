package logview

import (
	"strings"
	"testing"
)

func TestStripANSI(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"no escape", "plain text", "plain text"},
		{"empty", "", ""},
		{"sgr color", "\x1b[31mERROR\x1b[0m failed", "ERROR failed"},
		{"sgr reset without params", "\x1b[mtext", "text"},
		{"sgr 256 and truecolor", "\x1b[38;5;196mred\x1b[38;2;255;0;0m too\x1b[39m", "red too"},
		{"cursor and erase", "\x1b[2K\x1b[1Gprogress 40%", "progress 40%"},
		{"private mode", "\x1b[?25lhidden cursor\x1b[?25h", "hidden cursor"},
		{"osc hyperlink with ST", "see \x1b]8;;https://example.com\x1b\\docs\x1b]8;;\x1b\\ now", "see docs now"},
		{"osc title with BEL", "\x1b]0;window title\atext", "text"},
		{"charset designation", "\x1b(Btext", "text"},
		{"single-byte sequences", "\x1b7saved\x1b8 \x1b=keypad", "saved keypad"},
		{"lone trailing ESC", "text\x1b", "text"},
		{"truncated CSI at end", "text\x1b[38;5", "text"},
		{"CSI broken by a non-CSI byte keeps the byte", "\x1b[3\nnext", "\nnext"},
		{"CSI broken by another ESC", "\x1b[3\x1b[31mred", "red"},
		{"unterminated osc runs to the end", "text\x1b]8;;url", "text"},
		{"ESC before a control byte keeps the byte", "a\x1b\tb", "a\tb"},
		{"utf-8 around sequences", "\x1b[1mÄrger\x1b[0m – ok", "Ärger – ok"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := stripANSI(c.in); got != c.want {
				t.Errorf("stripANSI(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestLineStripsANSIFromPlainText(t *testing.T) {
	// The level word is only recognized once the color around it is gone, so
	// stripping has to happen before the level detection, not after rendering.
	out := render(t, Line("\x1b[31mERROR\x1b[0m: disk full"))
	contains(t, out, `<span class="log-level log-level-error">ERROR:</span>`, `disk full`)
	excludes(t, out, "[31m", "[0m", "\x1b")
}

func TestLineStripsANSIBeforeParsing(t *testing.T) {
	// jq -C prints a record with raw escapes around every token, which is
	// invalid JSON until they are gone.
	raw := "\x1b[1;39m{\x1b[0m\x1b[34;1m\"level\"\x1b[0m\x1b[1;39m:\x1b[0;32m\"warn\"\x1b[0m\x1b[1;39m}\x1b[0m"
	out := render(t, Line(raw))
	contains(t, out, `log-level-warn`)
	excludes(t, out, "[0m", "\x1b", "{")
}

func TestLineStripsANSIFromDecodedJSONStrings(t *testing.T) {
	// docker's json-file driver escapes the ESC as \u001b, so it survives the
	// raw-line pass and only appears once the string is decoded.
	out := render(t, Line(`{"log":"\u001b[32mready\u001b[0m","nested":{"k":["\u001b[1mbold\u001b[0m"]}}`))
	contains(t, out, `<span class="log-string">ready</span>`, `<span class="log-string">bold</span>`)
	excludes(t, out, "[32m", "[0m", "[1m", "\x1b")
}

func TestLineStripsANSIBeforeTheLengthCap(t *testing.T) {
	// MaxLineLen counts what is shown: color codes around a line that fits must
	// not push it over the cap.
	cfg := &Config{MaxLineLen: 10}
	out := render(t, cfg.Line("\x1b[31m0123456789\x1b[0m"))
	contains(t, out, "0123456789")
	excludes(t, out, "…")
}

func TestLineKeepANSI(t *testing.T) {
	cfg := &Config{KeepANSI: true}
	out := render(t, cfg.Line("\x1b[31mred\x1b[0m"))
	if !strings.Contains(out, "\x1b[31mred") {
		t.Errorf("KeepANSI dropped the escape sequence: %q", out)
	}
	out = render(t, cfg.Line(`{"msg":"\u001b[31mred"}`))
	if !strings.Contains(out, "\x1b[31mred") {
		t.Errorf("KeepANSI dropped the decoded escape sequence: %q", out)
	}
}
