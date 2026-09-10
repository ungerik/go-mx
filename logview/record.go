package logview

import (
	"encoding/json"
	"io"
	"strings"
)

// kind is the JSON type of a parsed value, which is what its color conveys.
type kind int

const (
	kindString kind = iota
	kindNumber
	kindBool
	kindNull
	kindObject
	kindArray
	// kindRaw is a value past [Config.MaxDepth], kept as its raw JSON text.
	kindRaw
)

// value is one parsed JSON value. Only the field matching kind is populated.
type value struct {
	kind   kind
	text   string  // kindString, kindNumber, kindBool, kindNull, kindRaw
	fields []field // kindObject
	items  []value // kindArray
}

// field is one key/value pair of a JSON object, in source order.
type field struct {
	key string
	val value
}

// parseRecord parses raw as a single JSON object into its fields in source
// order, and reports whether raw was exactly that.
//
// Order is the reason this walks the token stream instead of unmarshalling into
// a map: a log record's first field is meaningful (see [Config.TimeKey]) and a
// map would lose it. json.Number keeps a 19-digit id from being rendered as
// 1.2e+18.
//
// Anything else — an array, a bare value, a truncated object, or an object with
// trailing content — reports false so the caller falls back to plain text
// rather than rendering part of a line and dropping the rest.
func parseRecord(raw string, maxDepth int) ([]field, bool) {
	trimmed := strings.TrimLeft(raw, " \t")
	if !strings.HasPrefix(trimmed, "{") {
		// Cheap reject before allocating a Decoder, because most lines of a
		// plain-text stream reach this and none of them are records.
		return nil, false
	}
	dec := json.NewDecoder(strings.NewReader(trimmed))
	dec.UseNumber()

	if tok, err := dec.Token(); err != nil || tok != json.Token(json.Delim('{')) {
		return nil, false
	}
	fields, ok := parseObject(dec, maxDepth)
	if !ok {
		return nil, false
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, false
	}
	return fields, true
}

// parseObject reads key/value pairs until the closing brace. The opening brace
// has already been consumed.
func parseObject(dec *json.Decoder, depth int) ([]field, bool) {
	var fields []field
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, false
		}
		if tok == json.Token(json.Delim('}')) {
			return fields, true
		}
		key, ok := tok.(string)
		if !ok {
			return nil, false
		}
		val, ok := parseValue(dec, depth)
		if !ok {
			return nil, false
		}
		// Duplicate keys are kept, in order, rather than resolved: a log record
		// with two "err" fields is a bug worth seeing, not one to hide.
		fields = append(fields, field{key: key, val: val})
	}
}

// parseArray reads values until the closing bracket. The opening bracket has
// already been consumed.
func parseArray(dec *json.Decoder, depth int) ([]value, bool) {
	var items []value
	for {
		if !dec.More() {
			if tok, err := dec.Token(); err != nil || tok != json.Token(json.Delim(']')) {
				return nil, false
			}
			return items, true
		}
		item, ok := parseValue(dec, depth)
		if !ok {
			return nil, false
		}
		items = append(items, item)
	}
}

// parseValue reads one value. At depth 0 it stops descending and keeps the
// value's raw JSON instead, which is what bounds the recursion: without it a
// line of a few hundred thousand nested braces would overflow the goroutine
// stack, and a stack overflow is a fatal runtime error no recover can catch.
func parseValue(dec *json.Decoder, depth int) (value, bool) {
	if depth <= 0 {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return value{}, false
		}
		return value{kind: kindRaw, text: string(raw)}, true
	}
	tok, err := dec.Token()
	if err != nil {
		return value{}, false
	}
	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			fields, ok := parseObject(dec, depth-1)
			return value{kind: kindObject, fields: fields}, ok
		case '[':
			items, ok := parseArray(dec, depth-1)
			return value{kind: kindArray, items: items}, ok
		}
		return value{}, false
	case string:
		return value{kind: kindString, text: v}, true
	case json.Number:
		return value{kind: kindNumber, text: v.String()}, true
	case bool:
		if v {
			return value{kind: kindBool, text: "true"}, true
		}
		return value{kind: kindBool, text: "false"}, true
	case nil:
		return value{kind: kindNull, text: "null"}, true
	}
	return value{}, false
}
