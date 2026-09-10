package logview

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Line renders one raw log line.
//
// A line that is a single JSON object is rendered as its fields: the timestamp
// and level promoted, every other field as key=value with the value colored by
// its JSON type. Anything else is rendered as plain text, with a leading level
// word from [Config.Levels] colored like a JSON level.
//
// The line is rendered through its own non-indenting writer, so the markup is
// identical whether the caller's writer indents or not, and never contains a
// line break. That matters on both ends: an indenting writer would inject
// whitespace between the spans that the view's white-space:pre-wrap would then
// show, and a trailing newline would reach the client as an extra "data:" line
// and become a stray text node between every pair of log lines.
func (c *Config) Line(raw string) mx.Component {
	return compact(c.lineElement(raw))
}

// Lines renders a batch of raw log lines as one component, so a burst can be
// sent as a single SSE event instead of one event, flush and DOM swap per line.
func (c *Config) Lines(raw ...string) mx.Component {
	comps := make(mx.Components, len(raw))
	for i, r := range raw {
		comps[i] = c.lineElement(r)
	}
	return compact(comps)
}

// compact renders comp through a fresh non-indenting [mx.CheckedWriter] that
// writes to w. Nesting a writer this way is the same contract [mx.Raw] uses —
// the bytes go to the destination without the outer writer's element tracking —
// and it is what makes a line's markup independent of how the caller writes.
func compact(comp mx.Component) mx.Component {
	return mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
		return comp.Render(ctx, mx.NewCheckedWriter(w))
	})
}

func (c *Config) lineElement(raw string) *mx.Element {
	raw = c.truncate(raw)
	if fields, ok := parseRecord(raw, c.maxDepth()); ok {
		return c.recordLine(fields)
	}
	return c.textLine(raw)
}

// truncate bounds one line's length. It bounds three things at once: the markup
// a single line can produce, the memory a runaway line takes in the DOM, and
// the work the client-side filter does per keystroke.
func (c *Config) truncate(raw string) string {
	max := c.maxLineLen()
	if len(raw) <= max {
		return raw
	}
	// Cut on a rune boundary: a half-written rune would be invalid UTF-8 in the
	// page, which is not something a log line should be able to cause.
	cut := max
	for cut > 0 && !utf8.RuneStart(raw[cut]) {
		cut--
	}
	return raw[:cut] + "…"
}

// recordLine renders a parsed JSON record.
func (c *Config) recordLine(fields []field) *mx.Element {
	var (
		theme    = c.theme()
		timeKey  = c.timeKey()
		levelKey = c.levelKey()
		msgKey   = c.MessageKey
		parts    = make(mx.Components, 0, len(fields)*2)
		level    string
	)
	for i, f := range fields {
		if len(parts) > 0 {
			parts = append(parts, mx.Text(" "))
		}
		switch {
		case f.key == timeKey && i == 0:
			// A leading timestamp reads as one without a "time=" label. Later
			// in the record it does not, so there it keeps its key.
			parts = append(parts, c.promoted(ClassTime, f.val))
		case f.key == timeKey:
			parts = append(parts, c.pairAs(f.key, f.val, ClassTime))
		case f.key == levelKey && level == "":
			comp, token := c.levelValue(theme, f.val)
			parts = append(parts, comp)
			level = token
		case msgKey != "" && f.key == msgKey:
			parts = append(parts, c.promoted(ClassMessage, f.val))
		default:
			parts = append(parts, c.pair(f.key, f.val))
		}
	}
	return c.line(level, parts)
}

// textLine renders a line that is not a JSON record.
func (c *Config) textLine(raw string) *mx.Element {
	// Trailing line breaks are the line separator of the source, not content.
	// Interior ones are content and survive: the view sets white-space:pre-wrap,
	// so a multi-line chunk keeps its breaks and its indentation without needing
	// a <pre> of its own.
	text := strings.TrimRight(raw, "\r\n")

	var (
		parts mx.Components
		level string
	)
	if word, display, rest := c.splitLevel(text); word != "" {
		comp, token := c.levelText(c.theme(), word, display)
		parts = append(parts, comp)
		level = token
		text = rest
	}
	parts = append(parts, c.span(ClassText, text))
	return c.line(level, parts)
}

// line wraps the rendered parts in the line element, tagging it with the
// normalized level so a theme can style the whole row.
func (c *Config) line(level string, parts mx.Components) *mx.Element {
	attribsChildren := make([]any, 0, 3)
	attribsChildren = append(attribsChildren, html.Class(c.class(ClassLine)))
	if level != "" {
		attribsChildren = append(attribsChildren, html.Attrib(attrLevel, level))
	}
	return html.Div(append(attribsChildren, parts)...)
}

// pair renders "key=value" with the value colored by its JSON type.
func (c *Config) pair(key string, v value) mx.Component {
	return c.pairAs(key, v, "")
}

// pairAs renders "key=value", styling the value as class instead of by its JSON
// type. An empty class uses the type.
func (c *Config) pairAs(key string, v value, class Class) mx.Component {
	val := c.renderValue(v)
	if class != "" {
		if text, ok := c.scalarText(v); ok {
			val = c.span(class, text)
		}
	}
	return mx.Components{
		c.span(ClassKey, key),
		c.span(ClassPunct, "="),
		val,
	}
}

// promoted renders a value without its key, styled as class. A value that is
// not a scalar — an object where a timestamp was expected — falls back to the
// type-colored rendering instead of being flattened into something it is not.
func (c *Config) promoted(class Class, v value) mx.Component {
	if text, ok := c.scalarText(v); ok {
		return c.span(class, text)
	}
	return c.renderValue(v)
}

// levelValue renders a record's level field and returns the modifier token it
// was tagged with.
func (c *Config) levelValue(theme Theme, v value) (mx.Component, string) {
	text, ok := c.scalarText(v)
	if !ok {
		return c.renderValue(v), LevelUndefined
	}
	return c.levelText(theme, text, text)
}

// levelText renders a level as text, or as the image the theme defines for it,
// and returns the modifier token it was tagged with.
//
// value is the level itself and decides the token and the image's alt text;
// display is what is shown, which for a plain-text line includes the brackets
// the line wrote it in. The class is always one the theme defined — see
// [Theme.Levels] for why the raw value can never reach the class attribute.
func (c *Config) levelText(theme Theme, value, display string) (mx.Component, string) {
	token := theme.levelToken(value)
	classes := c.class(ClassLevel) + " " + c.class(ClassLevel) + "-" + token
	if image := theme.Levels[token].Image; image != "" {
		return html.SpanClass(classes, html.Img(
			html.Class(c.class(ClassLevelImage)),
			html.Src(image),
			// The level is the alt text: it is exactly what the image replaced,
			// so it is what a screen reader should read instead.
			html.Alt(value),
		)), token
	}
	return html.SpanClass(classes, mx.Text(display)), token
}

// renderValue renders a value colored by its JSON type.
func (c *Config) renderValue(v value) mx.Component {
	switch v.kind {
	case kindString:
		if strings.ContainsAny(v.text, "\n\r") {
			return c.multiline(v.text)
		}
		text := v.text
		if c.QuoteText {
			text = strconv.Quote(text)
		}
		return c.span(ClassString, text)
	case kindNumber:
		return c.span(ClassNumber, v.text)
	case kindBool:
		return c.span(ClassBool, v.text)
	case kindNull:
		return c.span(ClassNull, v.text)
	case kindObject:
		return c.object(v.fields)
	case kindArray:
		return c.array(v.items)
	}
	// kindRaw: past MaxDepth, so the value's own type is not known here.
	return c.span(ClassText, v.text)
}

// object renders a nested object in braces, its fields formatted like a
// record's. The timestamp, level and message rules are top-level only: a
// "level" inside an attached object is a field of that object, not the
// severity of the line.
func (c *Config) object(fields []field) mx.Component {
	comps := make(mx.Components, 0, len(fields)*2+2)
	comps = append(comps, c.span(ClassPunct, "{"))
	for i, f := range fields {
		if i > 0 {
			comps = append(comps, mx.Text(" "))
		}
		comps = append(comps, c.pair(f.key, f.val))
	}
	return append(comps, c.span(ClassPunct, "}"))
}

// array renders a nested array in brackets, its items colored by their types.
func (c *Config) array(items []value) mx.Component {
	comps := make(mx.Components, 0, len(items)*2+2)
	comps = append(comps, c.span(ClassPunct, "["))
	for i, item := range items {
		if i > 0 {
			comps = append(comps, c.span(ClassPunct, ", "))
		}
		comps = append(comps, c.renderValue(item))
	}
	return append(comps, c.span(ClassPunct, "]"))
}

// multiline renders a string value that spans lines — in practice an error with
// a stack trace — as a <pre> so its indentation survives.
//
// The <pre> has exactly one text child, and that is load-bearing: a
// CheckedWriter with indentation enabled breaks and indents before a child
// element's start tag, which inside a <pre> would be visible whitespace. With
// text as the only child there is no such break, so the block is byte-exact
// under any writer.
func (c *Config) multiline(s string) mx.Component {
	text := normalizeNewlines(s)
	if strings.HasPrefix(text, "\n") {
		// HTML drops a newline directly after the <pre> start tag, so a value
		// that starts with a blank line needs one more to keep it.
		text = "\n" + text
	}
	return html.Pre(html.Class(c.class(ClassMultiline)), mx.Text(text))
}

// normalizeNewlines makes every line break an LF. The SSE transport already
// does this — mx splits event data on CR, LF and CRLF alike and the client
// rejoins with LF — so normalizing here means the rendered markup is what the
// client reconstructs, whether it was streamed or served with the page.
func normalizeNewlines(s string) string {
	if !strings.ContainsRune(s, '\r') {
		return s
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

// span renders text in a span carrying the class's full name.
func (c *Config) span(class Class, text string) *mx.Element {
	return html.SpanClass(c.class(class), mx.Text(text))
}

// scalarText returns a value's display text and whether it has one. A multiline
// string does not: it renders as a block, not as a word in a line.
//
// Promoted values are never quoted, even under [Config.QuoteText] — a quoted
// timestamp or message is noise, and the promotion already says what it is.
func (c *Config) scalarText(v value) (string, bool) {
	switch v.kind {
	case kindString:
		if strings.ContainsAny(v.text, "\n\r") {
			return "", false
		}
		return v.text, true
	case kindNumber, kindBool, kindNull:
		return v.text, true
	}
	return "", false
}

// splitLevel matches a level word at the start of a plain-text line. It returns
// the bare word, the text to display for it (which keeps the brackets the line
// wrote), and the rest of the line. It accepts "WARN …", "WARN: …", "[WARN] …"
// and "<warn> …".
//
// It only matches at the start, only a word from [Config.Levels], and only when
// a delimiter follows, so a line that merely begins with a word like "Error" as
// prose is left alone. The vocabulary comes from the Config rather than from
// [Theme.Levels] on purpose: changing a theme must not change how a line parses.
func (c *Config) splitLevel(line string) (word, display, rest string) {
	levels := c.levels()
	if len(levels) == 0 {
		return "", "", line
	}

	i := 0
	bracketed := len(line) > 0 && (line[0] == '[' || line[0] == '<')
	if bracketed {
		i++
	}
	start := i
	for i < len(line) && isLetter(line[i]) {
		i++
	}
	word = line[start:i]
	if word == "" {
		return "", "", line
	}
	if !slices.ContainsFunc(levels, func(l string) bool { return strings.EqualFold(l, word) }) {
		return "", "", line
	}
	if bracketed {
		if i >= len(line) || (line[i] != ']' && line[i] != '>') {
			return "", "", line
		}
		i++
	}
	// A trailing colon belongs to the level the way the brackets do. Left in the
	// rest it would be stranded after the level column's padding, reading as
	// "INFO   : cache warm-up finished".
	switch {
	case i < len(line) && line[i] == ':':
		i++
	case i < len(line) && line[i] != ' ' && line[i] != '\t':
		return "", "", line
	}
	return word, line[:i], line[i:]
}

func isLetter(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}
