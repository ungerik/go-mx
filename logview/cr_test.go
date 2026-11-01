package logview

import (
	"strings"
	"testing"
)

func TestOverwriteCR(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"no carriage return", "plain line", "plain line"},
		{"progress redraws keep the last", "10%\r20%\r30%", "30%"},
		{"trailing CR ends the line", "10%\r20%\r", "20%\r"},
		{"CRLF is a line ending", "line\r\n", "line\r\n"},
		{"lone CR line ending", "line\r", "line\r"},
		{"per segment", "a\rb\nc\rd", "b\nd"},
		{"progress before a newline", "dl 10%\rdl 100%\r\ndone", "dl 100%\r\ndone"},
		{"leading CR", "\rtext", "text"},
		{"empty after overwrite keeps the segment", "\r\r", "\r\r"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := overwriteCR(c.in); got != c.want {
				t.Errorf("overwriteCR(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestLineOverwriteCROff(t *testing.T) {
	// Off by default: a carriage return is a line break and no text is lost.
	out := render(t, Line("downloading 10%\rdownloading 100%"))
	contains(t, out, "downloading 10%", "downloading 100%")
}

func TestLineOverwriteCRPlainText(t *testing.T) {
	cfg := &Config{OverwriteCR: true}
	out := render(t, cfg.Line("INFO downloading 10%\rINFO downloading 55%\rINFO downloading 100%\r\n"))
	contains(t, out, `downloading 100%`, `log-level-info`)
	excludes(t, out, "10%", "55%", "\r")
}

func TestLineOverwriteCRRecordStrings(t *testing.T) {
	// docker's json-file driver: the progress output is a decoded string.
	cfg := &Config{OverwriteCR: true}
	out := render(t, cfg.Line(`{"log":"pull 10%\rpull 60%\rpull 100%","stream":"stdout"}`))
	contains(t, out, `<span class="log-string">pull 100%</span>`)
	excludes(t, out, "10%", "60%")
}

func TestLineOverwriteCRCountsAgainstMaxLineLen(t *testing.T) {
	// The cap applies to what is shown, so a long run of overwritten redraws
	// does not truncate the final state.
	cfg := &Config{OverwriteCR: true, MaxLineLen: 32}
	raw := strings.Repeat("progress 99%\r", 20) + "finished"
	out := render(t, cfg.Line(raw))
	contains(t, out, `>finished<`)
	excludes(t, out, "…")
}

func TestMatchesOverwriteCR(t *testing.T) {
	// The search sees what the view shows.
	cfg := &Config{OverwriteCR: true}
	if cfg.Matches("step 1\rstep 2", "step 1") {
		t.Error("an overwritten redraw matched")
	}
	if !cfg.Matches("step 1\rstep 2", "step 2") {
		t.Error("the final state did not match")
	}
}
