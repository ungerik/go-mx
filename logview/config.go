package logview

import (
	"cmp"

	"github.com/ungerik/go-mx"
)

// Defaults of the zero [Config].
const (
	// DefaultPrefix is the CSS class prefix of the emitted markup.
	DefaultPrefix = "log-"
	// DefaultEvent is the SSE event name [Config.View] subscribes to.
	DefaultEvent = "log"
	// DefaultTimeKey is the JSON field holding a record's timestamp.
	DefaultTimeKey = "time"
	// DefaultLevelKey is the JSON field holding a record's severity.
	DefaultLevelKey = "level"
	// DefaultMaxDepth is how deep nested JSON values are rendered field by
	// field before being dumped as raw JSON.
	DefaultMaxDepth = 8
	// DefaultMaxLineLen is the byte length a raw log line is truncated to.
	DefaultMaxLineLen = 8192
	// DefaultMaxLines is how many lines a view keeps in the DOM.
	DefaultMaxLines = 2000
	// DefaultHeight is the CSS height of a view's scroll area.
	DefaultHeight = "24rem"
)

// DefaultLevels is the level vocabulary [Config.Line] recognizes at the start
// of a plain-text log line. It is deliberately separate from [Theme.Levels]:
// how a line parses must not change when its colors do.
var DefaultLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}

// Labels are the user-visible strings a [Config.View] renders on its own,
// without the caller passing them. An empty field falls back to the English
// default, so a partial translation still renders every control.
//
// Unlike shadcn's context-resolved Labels these live on the Config: a log view
// is built from a Config the application owns, so there is nothing a render
// context would add that setting the field does not.
type Labels struct {
	// Filter is the aria-label of the filter input. Default "Filter log lines".
	Filter string
	// FilterHint is the placeholder of the filter input. Default "Filter…".
	FilterHint string
	// Pause is the caption of the pause toggle. Default "Pause".
	Pause string
	// NewLines is what the script appends to the count of lines that arrived
	// while paused, as in "12 new". Default "new".
	NewLines string
}

// Config renders log lines and the view they stream into. The zero Config is
// usable and equals [Default]; every field selects a default when left empty.
//
// A Config is read-only once it is serving: its maps and slices are shared by
// every handler and every render, without synchronization. Configure it during
// program setup, the way the package-level vars in the mx package are meant to
// be set.
type Config struct {
	// Prefix is the CSS class prefix of the emitted markup and of the
	// stylesheet [Config.StyleElement] produces. "" means [DefaultPrefix].
	Prefix string

	// TimeKey is the JSON field rendered as a record's timestamp. Its value is
	// always styled as [ClassTime]; the "key=" label is omitted only when the
	// field is the first of the record, where it reads as a timestamp on its
	// own. "" means [DefaultTimeKey].
	TimeKey string

	// LevelKey is the JSON field rendered as a record's severity. Its value is
	// always rendered without a label and styled by [Theme.Levels].
	// "" means [DefaultLevelKey].
	LevelKey string

	// MessageKey is the JSON field rendered without a label as the record's
	// message. Empty renders the message field like any other field, which is
	// the default; set it to "msg" for log/slog and zerolog, or "message" for
	// zap's production encoder.
	MessageKey string

	// Levels is the vocabulary of level words recognized at the start of a
	// plain-text log line. nil means [DefaultLevels]; an empty non-nil slice
	// turns the detection off.
	Levels []string

	// QuoteText renders JSON string values in quotes. Off by default: the
	// value's color already says it is a string, and quotes around every
	// message are noise.
	QuoteText bool

	// MaxDepth is how deep nested JSON objects and arrays are rendered field by
	// field. A value deeper than this is rendered as its raw JSON text, which
	// bounds the recursion for a hostile or accidental deeply nested line.
	// 0 means [DefaultMaxDepth].
	MaxDepth int

	// MaxLineLen is the byte length a raw line is truncated to before it is
	// parsed. It bounds both the markup one line can produce and the work the
	// client-side filter does per line. 0 means [DefaultMaxLineLen].
	MaxLineLen int

	// MaxLines is how many lines a [Config.View] keeps in the DOM before it
	// drops the oldest. 0 means [DefaultMaxLines].
	MaxLines int

	// Height is the CSS height of a view's scroll area. A scroll area without
	// one grows instead of scrolling, which leaves nothing for the
	// stick-to-bottom behavior to do. "" means [DefaultHeight].
	Height string

	// Event is the SSE event name a [Config.View] subscribes to and that the
	// handler must pass to [mx.SSEResponse.Send]. "" means [DefaultEvent].
	Event string

	// Theme is the appearance of the view. The zero Theme means [DarkTheme].
	Theme Theme

	// Labels are the strings a view renders on its own. Empty fields fall back
	// to English defaults.
	Labels Labels
}

// Default is the [Config] the package-level [Line], [Lines], [View] and
// [StyleElement] functions use.
var Default = &Config{}

func (c *Config) prefix() string {
	if c.Prefix == "" {
		return DefaultPrefix
	}
	return c.Prefix
}

func (c *Config) timeKey() string {
	if c.TimeKey == "" {
		return DefaultTimeKey
	}
	return c.TimeKey
}

func (c *Config) levelKey() string {
	if c.LevelKey == "" {
		return DefaultLevelKey
	}
	return c.LevelKey
}

func (c *Config) levels() []string {
	if c.Levels == nil {
		return DefaultLevels
	}
	return c.Levels
}

func (c *Config) maxDepth() int {
	if c.MaxDepth <= 0 {
		return DefaultMaxDepth
	}
	return c.MaxDepth
}

func (c *Config) maxLineLen() int {
	if c.MaxLineLen <= 0 {
		return DefaultMaxLineLen
	}
	return c.MaxLineLen
}

func (c *Config) maxLines() int {
	if c.MaxLines <= 0 {
		return DefaultMaxLines
	}
	return c.MaxLines
}

func (c *Config) height() string {
	if c.Height == "" {
		return DefaultHeight
	}
	return c.Height
}

// EventName is the SSE event name this Config subscribes to, with the default
// applied. A handler needs it to name the events it sends, and reading the
// Event field directly would give "" for a Config that never set it — which
// sends an unnamed event the view is not listening for.
func (c *Config) EventName() string {
	if c.Event == "" {
		return DefaultEvent
	}
	return c.Event
}

// isZero reports whether the Theme was left unset, which selects [DarkTheme].
// A Theme that only sets Name is still a caller's theme: it says "no colors",
// which is different from "no theme".
func (t Theme) isZero() bool {
	return t.Name == "" && t.Background == "" && t.Foreground == "" &&
		t.Styles == nil && t.Levels == nil
}

func (c *Config) theme() Theme {
	if c.Theme.isZero() {
		return DarkTheme
	}
	return c.Theme
}

func (c *Config) labels() Labels {
	return Labels{
		Filter:     cmp.Or(c.Labels.Filter, "Filter log lines"),
		FilterHint: cmp.Or(c.Labels.FilterHint, "Filter…"),
		Pause:      cmp.Or(c.Labels.Pause, "Pause"),
		NewLines:   cmp.Or(c.Labels.NewLines, "new"),
	}
}

// class returns the full class name of a [Class] under the Config's prefix.
func (c *Config) class(class Class) string {
	return c.prefix() + string(class)
}

// StyleElement returns the Config's theme as a <style> element for a document
// <head>. It is the only stylesheet the view itself needs.
func (c *Config) StyleElement() *mx.Element {
	return c.theme().StyleElement(c.prefix())
}

// StyleElement returns [Default]'s theme as a <style> element.
func StyleElement() *mx.Element { return Default.StyleElement() }

// Line renders raw as one log line using [Default].
func Line(raw string) mx.Component { return Default.Line(raw) }

// Lines renders raw as a batch of log lines using [Default].
func Lines(raw ...string) mx.Component { return Default.Lines(raw...) }

// View renders a streaming log view using [Default].
func View(attribsChildren ...any) *mx.Element { return Default.View(attribsChildren...) }
