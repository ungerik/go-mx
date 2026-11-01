package logview

import "testing"

func TestMatchesTheRenderedText(t *testing.T) {
	// The reader types what the view shows, so a record is matched as rendered:
	// "status=200" is in the line although the raw JSON reads "status":200.
	record := `{"time":"12:00:01","level":"warn","msg":"slow request","status":200,"db":{"host":"db2"}}`
	cases := []struct {
		query string
		want  bool
	}{
		{"", true},
		{"   ", true},
		{"status=200", true},
		{"STATUS=200", true},
		{"  slow request  ", true},
		{"{host=db2}", true},
		{"warn", true},
		{`"status":200`, false},
		{"panic", false},
	}
	for _, c := range cases {
		if got := Matches(record, c.query); got != c.want {
			t.Errorf("Matches(record, %q) = %v, want %v", c.query, got, c.want)
		}
	}
}

func TestMatchesPlainTextAndANSI(t *testing.T) {
	// Escape sequences are stripped before rendering, so they neither match
	// nor break a match that spans them.
	line := "\x1b[31mERROR\x1b[0m: disk full"
	if !Matches(line, "error: disk") {
		t.Errorf("Matches did not find text around stripped escape sequences")
	}
	if Matches(line, "[31m") {
		t.Errorf("Matches found a stripped escape sequence")
	}
	if !(&Config{KeepANSI: true}).Matches(line, "[31m") {
		t.Errorf("Matches with KeepANSI did not find the kept escape sequence")
	}
}

func TestMatchesLeavesOutMarkup(t *testing.T) {
	// Only text content counts: a class name or a tag in the rendered markup is
	// not something the reader sees.
	if Matches(`{"level":"info"}`, "log-level") {
		t.Errorf("Matches found a class name")
	}
	if Matches("plain", "span") {
		t.Errorf("Matches found a tag name")
	}
}
