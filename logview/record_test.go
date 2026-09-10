package logview

import (
	"strings"
	"testing"
)

func TestParseRecordKeepsFieldOrder(t *testing.T) {
	// Order is the reason this walks the token stream instead of unmarshalling
	// into a map: whether the timestamp leads the record decides how it renders.
	fields, ok := parseRecord(`{"z":1,"a":2,"m":3}`, DefaultMaxDepth)
	if !ok {
		t.Fatal("a well-formed record was rejected")
	}
	var keys []string
	for _, f := range fields {
		keys = append(keys, f.key)
	}
	if got := strings.Join(keys, ","); got != "z,a,m" {
		t.Errorf("field order is %q, want \"z,a,m\"", got)
	}
}

func TestParseRecordRejectsNonRecords(t *testing.T) {
	// Every rejection here becomes a plain-text line. Rendering half a record
	// and dropping the rest would lose data silently, which is worse than not
	// parsing it at all.
	for _, raw := range []string{
		``,
		`   `,
		`[1,2]`,
		`"a string"`,
		`42`,
		`{`,
		`{"a":1`,
		`{"a"}`,
		`{"a":1} trailing`,
		`{"a":1}{"b":2}`,
		`not json`,
	} {
		if _, ok := parseRecord(raw, DefaultMaxDepth); ok {
			t.Errorf("parseRecord(%q) accepted a non-record", raw)
		}
	}
}

func TestParseRecordAcceptsLeadingWhitespace(t *testing.T) {
	// Log lines arrive indented often enough that rejecting them for it would
	// be surprising.
	if _, ok := parseRecord("  \t{\"a\":1}", DefaultMaxDepth); !ok {
		t.Error("a record with leading whitespace was rejected")
	}
}

func TestParseRecordValueKinds(t *testing.T) {
	fields, ok := parseRecord(`{"s":"x","n":1,"t":true,"f":false,"z":null,"o":{},"a":[]}`, DefaultMaxDepth)
	if !ok {
		t.Fatal("a well-formed record was rejected")
	}
	want := []kind{kindString, kindNumber, kindBool, kindBool, kindNull, kindObject, kindArray}
	if len(fields) != len(want) {
		t.Fatalf("parsed %d fields, want %d", len(fields), len(want))
	}
	for i, k := range want {
		if fields[i].val.kind != k {
			t.Errorf("field %q has kind %v, want %v", fields[i].key, fields[i].val.kind, k)
		}
	}
}

func TestParseRecordDepthCapKeepsRawJSON(t *testing.T) {
	// Past the cap the value is kept verbatim rather than descended into, so
	// nothing is lost and the recursion is bounded. MaxDepth 1 expands the
	// record and one nested object; the object below that is kept raw.
	fields, ok := parseRecord(`{"a":{"b":{"c":1}}}`, 1)
	if !ok {
		t.Fatal("a well-formed record was rejected")
	}
	inner := fields[0].val.fields[0].val
	if inner.kind != kindRaw {
		t.Fatalf("value past the depth cap has kind %v, want kindRaw", inner.kind)
	}
	if inner.text != `{"c":1}` {
		t.Errorf("raw value is %q, want %q", inner.text, `{"c":1}`)
	}
}

func TestParseRecordDeepNestingTerminates(t *testing.T) {
	// The cap exists so that a hostile line cannot overflow the goroutine
	// stack, which would take the whole process down rather than the request.
	// Rejecting such a line is a fine outcome — it renders as plain text — but
	// returning at all is the property under test.
	moderate := strings.Repeat(`{"a":`, 2000) + "1" + strings.Repeat("}", 2000)
	if _, ok := parseRecord(moderate, DefaultMaxDepth); !ok {
		t.Error("a deeply nested but well-formed record was rejected")
	}
	absurd := strings.Repeat(`{"a":`, 200000) + "1" + strings.Repeat("}", 200000)
	parseRecord(absurd, DefaultMaxDepth)
}
