package logview

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Class is the kind of a styled span in a rendered log line. The markup a
// [Config] produces carries only these classes, so the same markup can be
// styled by any [Theme] that uses the same prefix.
type Class string

const (
	// ClassView is the root element of a [Config.View].
	ClassView Class = "view"
	// ClassToolbar is the filter and pause row above the scroll area.
	ClassToolbar Class = "toolbar"
	// ClassFilter is the filter input.
	ClassFilter Class = "filter"
	// ClassPause is the pause button.
	ClassPause Class = "pause"
	// ClassStatus is the live region reporting the paused state.
	ClassStatus Class = "status"
	// ClassBadge is the "N new lines" counter shown while paused.
	ClassBadge Class = "badge"
	// ClassLines is the container the streamed lines are appended to.
	ClassLines Class = "lines"
	// ClassError is the sink an in-band [mx.SSEEventError] is swapped into.
	ClassError Class = "error"

	// ClassLine is one log line, JSON or plain text.
	ClassLine Class = "line"
	// ClassTime is the value of [Config.TimeKey].
	ClassTime Class = "time"
	// ClassLevel is the value of [Config.LevelKey]. A line's level also adds
	// one of the level modifier classes, see [Theme.Levels].
	ClassLevel Class = "level"
	// ClassLevelImage is the <img> rendered in place of a level whose
	// [LevelStyle.Image] is set.
	ClassLevelImage Class = "level-image"
	// ClassMessage is the value of [Config.MessageKey].
	ClassMessage Class = "message"
	// ClassKey is a field name.
	ClassKey Class = "key"
	// ClassString is a JSON string value.
	ClassString Class = "string"
	// ClassNumber is a JSON number value.
	ClassNumber Class = "number"
	// ClassBool is a JSON true or false value.
	ClassBool Class = "bool"
	// ClassNull is a JSON null value.
	ClassNull Class = "null"
	// ClassPunct is structural punctuation: "=", braces, brackets and commas.
	ClassPunct Class = "punct"
	// ClassText is the body of a line that is not a JSON object.
	ClassText Class = "text"
	// ClassMultiline is a string value containing line breaks, rendered as a
	// <pre> so a stack trace keeps its indentation.
	ClassMultiline Class = "multiline"

	// ClassMatch is what the filter matched inside a line. It is the one Class
	// that is not a class: the matches are a CSS custom highlight, so [Theme.CSS]
	// emits it as a ::highlight() rule. That pseudo-element only honors color,
	// background-color and text-decoration, so bold and italic are ignored here.
	ClassMatch Class = "match"
)

// LevelUndefined is the level modifier class used for every level value the
// [Theme] does not define, so an unknown value can never reach a class
// attribute. See [Theme.Levels].
const LevelUndefined = "undefined"

// Style is the appearance of one [Class] or level in a [Theme].
type Style struct {
	Color         string // any CSS color, "" to inherit
	Background    string // any CSS color, "" for none
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
}

// decls renders the Style as the body of a CSS rule, e.g.
// "color: #f85149; font-weight: bold". It returns "" for the empty Style.
func (s Style) decls() string {
	var parts []string
	if s.Color != "" {
		parts = append(parts, "color: "+s.Color)
	}
	if s.Background != "" {
		parts = append(parts, "background-color: "+s.Background)
	}
	if s.Bold {
		parts = append(parts, "font-weight: bold")
	}
	if s.Italic {
		parts = append(parts, "font-style: italic")
	}
	// Underline and strikethrough are one CSS property, so a style asking for
	// both has to combine them instead of emitting a declaration twice.
	var decorations []string
	if s.Underline {
		decorations = append(decorations, "underline")
	}
	if s.Strikethrough {
		decorations = append(decorations, "line-through")
	}
	if len(decorations) > 0 {
		parts = append(parts, "text-decoration: "+strings.Join(decorations, " "))
	}
	return strings.Join(parts, "; ")
}

// LevelStyle is the appearance of one log level value.
type LevelStyle struct {
	Style

	// Image is the URL of a bitmap rendered in place of the level text, with
	// the level value as its alt text. Empty renders the level as text.
	//
	// This is the one piece of a Theme consulted while rendering rather than
	// through CSS, because swapping an element for an image is not something a
	// stylesheet can do accessibly.
	Image string
}

// Theme is the appearance of a log view: the colors of the value types, of the
// known log levels, and the structural CSS of the view itself.
type Theme struct {
	Name       string // human-readable theme name
	Background string // view background-color, "" to inherit
	Foreground string // view base text color, "" to inherit

	// Styles maps a [Class] to its appearance. A class missing from the map is
	// rendered unstyled, which is why the markup never depends on the theme.
	Styles map[Class]Style

	// Levels maps a lowercased log level value to its appearance. It is also
	// the closed set of level modifier classes: a level value is emitted as
	// the class prefix+"level-"+value only when it is a key of this map, and
	// as prefix+"level-"+[LevelUndefined] otherwise.
	//
	// That is what keeps a level value — which comes from whoever wrote the log
	// line — out of the class attribute, where a space would end the class
	// token and let the rest of the value name arbitrary CSS classes. Keys
	// containing characters [mx.ValidIDRune] rejects are ignored for the same
	// reason, so a careless theme cannot open the hole either.
	Levels map[string]LevelStyle
}

// validLevelToken reports whether token is usable as a CSS class suffix. It is
// applied to Theme.Levels keys and to normalized level values alike, so both
// sides of the lookup agree on what a level can be.
func validLevelToken(token string) bool {
	if token == "" || len(token) > 32 {
		return false
	}
	for _, r := range token {
		if !mx.ValidIDRune(r) {
			return false
		}
	}
	return true
}

// MinLevelWidth is the narrowest the level column gets, in characters, for a
// theme that defines no levels of its own.
const MinLevelWidth = 5

// levelWidth is the character width of the level column: the longest level the
// theme defines. [LevelUndefined] is excluded because its key is never
// displayed — an unknown level shows its raw value, which has no bound and so
// cannot set the column width. Such a value simply runs past the column rather
// than being cut off, since a log must not hide what it was told.
func (t Theme) levelWidth() int {
	width := MinLevelWidth
	for level := range t.Levels {
		if level == LevelUndefined || !validLevelToken(level) {
			continue
		}
		// Level tokens are ASCII by validLevelToken, so bytes are characters.
		if len(level) > width {
			width = len(level)
		}
	}
	return width
}

// levelToken returns the modifier class suffix for a raw level value: the
// lowercased value when the theme defines it, and [LevelUndefined] otherwise.
func (t Theme) levelToken(value string) string {
	token := strings.ToLower(strings.TrimSpace(value))
	if !validLevelToken(token) {
		return LevelUndefined
	}
	if _, ok := t.Levels[token]; !ok {
		return LevelUndefined
	}
	return token
}

// CSS renders the theme as a stylesheet for the given class prefix (pass
// [DefaultPrefix] or "" to match the default [Config]). It emits the structural
// rules of the view, one rule per styled [Class] and one per level in
// deterministic order, so the same theme always produces the same bytes.
func (t Theme) CSS(prefix string) string {
	if prefix == "" {
		prefix = DefaultPrefix
	}

	var b strings.Builder
	fmt.Fprintf(&b, ".%sview {\n", prefix)
	if t.Background != "" {
		fmt.Fprintf(&b, "\tbackground-color: %s;\n", t.Background)
	}
	if t.Foreground != "" {
		fmt.Fprintf(&b, "\tcolor: %s;\n", t.Foreground)
	}
	b.WriteString("\tborder-radius: 6px;\n")
	b.WriteString("\toverflow: hidden;\n")
	// A flex column, so that giving the view a height — from a page stylesheet,
	// or by making it a flex item itself — makes the scroll area take whatever
	// the toolbar and the error sink leave. Without it the view can only be as
	// tall as Config.Height plus its chrome.
	b.WriteString("\tdisplay: flex;\n")
	b.WriteString("\tflex-direction: column;\n")
	b.WriteString("\tfont-family: ui-monospace, SFMono-Regular, \"SF Mono\", Menlo, Consolas, monospace;\n")
	b.WriteString("\tfont-size: 0.8125rem;\n")
	b.WriteString("\tline-height: 1.5;\n")
	b.WriteString("}\n")

	fmt.Fprintf(&b, ".%stoolbar {\n", prefix)
	b.WriteString("\tdisplay: flex;\n")
	b.WriteString("\talign-items: center;\n")
	b.WriteString("\tgap: 0.5rem;\n")
	b.WriteString("\tpadding: 0.375rem 0.5rem;\n")
	b.WriteString("\tborder-bottom: 1px solid currentColor;\n")
	b.WriteString("\tborder-bottom-color: color-mix(in srgb, currentColor 20%, transparent);\n")
	b.WriteString("}\n")

	fmt.Fprintf(&b, ".%sfilter {\n", prefix)
	b.WriteString("\tflex: 1;\n")
	b.WriteString("\tmin-width: 0;\n")
	b.WriteString("\tpadding: 0.125rem 0.375rem;\n")
	b.WriteString("\tborder-radius: 4px;\n")
	b.WriteString("\tborder: 1px solid color-mix(in srgb, currentColor 25%, transparent);\n")
	b.WriteString("\tbackground: transparent;\n")
	b.WriteString("\tcolor: inherit;\n")
	b.WriteString("\tfont: inherit;\n")
	b.WriteString("}\n")

	fmt.Fprintf(&b, ".%spause {\n", prefix)
	b.WriteString("\tpadding: 0.125rem 0.625rem;\n")
	b.WriteString("\tborder-radius: 4px;\n")
	b.WriteString("\tborder: 1px solid color-mix(in srgb, currentColor 25%, transparent);\n")
	b.WriteString("\tbackground: transparent;\n")
	b.WriteString("\tcolor: inherit;\n")
	b.WriteString("\tfont: inherit;\n")
	b.WriteString("\tcursor: pointer;\n")
	b.WriteString("}\n")
	// The button is as wide as its widest caption, so the toolbar does not jump
	// when it swaps between them. The width comes from the other caption itself
	// — carried by the same attribute the script reads — rather than from a
	// fixed em value, which would only ever fit the language it was measured in
	// and Labels exists precisely so that is not English.
	fmt.Fprintf(&b, ".%spause::before {\n", prefix)
	fmt.Fprintf(&b, "\tcontent: attr(%s);\n", attrPause)
	b.WriteString("\tdisplay: block;\n")
	b.WriteString("\theight: 0;\n")
	b.WriteString("\toverflow: hidden;\n")
	b.WriteString("\tvisibility: hidden;\n")
	b.WriteString("}\n")
	fmt.Fprintf(&b, ".%sview[%s] .%spause {\n", prefix, attrPaused, prefix)
	b.WriteString("\tbackground: color-mix(in srgb, currentColor 15%, transparent);\n")
	b.WriteString("}\n")

	// The status region keeps its box while empty rather than being hidden the
	// way the badge is: a live region that is display:none at the moment its
	// text is set is not reliably announced. An empty inline box takes no width
	// anyway, so there is nothing to hide.
	fmt.Fprintf(&b, ".%sstatus {\n", prefix)
	b.WriteString("\topacity: 0.7;\n")
	b.WriteString("\tfont-size: 0.75rem;\n")
	b.WriteString("\twhite-space: nowrap;\n")
	b.WriteString("}\n")

	fmt.Fprintf(&b, ".%sbadge {\n", prefix)
	b.WriteString("\topacity: 0.7;\n")
	b.WriteString("\tfont-size: 0.75rem;\n")
	b.WriteString("\twhite-space: nowrap;\n")
	b.WriteString("}\n")
	// Hidden while empty by an author rule, like the error sink below and for
	// the reason the line rules give: the hidden property is a UA-stylesheet
	// rule that any author display declaration beats.
	fmt.Fprintf(&b, ".%sbadge:empty {\n", prefix)
	b.WriteString("\tdisplay: none;\n")
	b.WriteString("}\n")

	fmt.Fprintf(&b, ".%slines {\n", prefix)
	b.WriteString("\tpadding: 0.375rem 0.5rem;\n")
	b.WriteString("}\n")

	// pre-wrap keeps the indentation of a plain-text log line, which HTML would
	// otherwise collapse. Line() renders through a non-indenting writer, so no
	// markup whitespace can leak into what pre-wrap preserves.
	fmt.Fprintf(&b, ".%sline {\n", prefix)
	b.WriteString("\twhite-space: pre-wrap;\n")
	b.WriteString("\tword-break: break-word;\n")
	b.WriteString("}\n")

	// Pausing and filtering are independent reasons to hide the same line, so
	// each gets its own attribute rather than sharing the hidden property —
	// which the UA stylesheet applies at a specificity any author display
	// declaration beats anyway.
	fmt.Fprintf(&b, ".%sline[%s],\n.%sline[%s] {\n", prefix, attrPending, prefix, attrNomatch)
	b.WriteString("\tdisplay: none !important;\n")
	b.WriteString("}\n")

	// A minimum width turns the promoted level into a column: the eye scans one
	// position down the log for the severity instead of following it as it
	// shifts with the length of the word before it. The width is in ch, which in
	// the view's monospace font is exactly one character, and it comes from the
	// theme's own vocabulary — as wide as the longest level it defines and no
	// wider.
	fmt.Fprintf(&b, ".%slevel {\n", prefix)
	b.WriteString("\tdisplay: inline-block;\n")
	fmt.Fprintf(&b, "\tmin-width: %dch;\n", t.levelWidth())
	b.WriteString("}\n")

	fmt.Fprintf(&b, ".%slevel-image {\n", prefix)
	b.WriteString("\theight: 1em;\n")
	b.WriteString("\twidth: auto;\n")
	b.WriteString("\tvertical-align: -0.125em;\n")
	b.WriteString("}\n")

	fmt.Fprintf(&b, ".%smultiline {\n", prefix)
	b.WriteString("\tmargin: 0.25rem 0 0.25rem 1rem;\n")
	b.WriteString("\toverflow-x: auto;\n")
	b.WriteString("\ttab-size: 4;\n")
	b.WriteString("\tfont: inherit;\n")
	b.WriteString("}\n")

	fmt.Fprintf(&b, ".%serror:empty {\n", prefix)
	b.WriteString("\tdisplay: none;\n")
	b.WriteString("}\n")
	fmt.Fprintf(&b, ".%serror {\n", prefix)
	b.WriteString("\tpadding: 0.375rem 0.5rem;\n")
	b.WriteString("\tborder-top: 1px solid currentColor;\n")
	b.WriteString("}\n")

	// The filter's matches are a custom highlight rather than a class, so they
	// need the pseudo-element rule instead of a class rule.
	if decls := t.Styles[ClassMatch].decls(); decls != "" {
		fmt.Fprintf(&b, "::highlight(%s) { %s }\n", matchHighlightName, decls)
	}

	classes := make([]string, 0, len(t.Styles))
	for class := range t.Styles {
		if Class(class) == ClassMatch {
			continue
		}
		classes = append(classes, string(class))
	}
	slices.Sort(classes)
	for _, class := range classes {
		if decls := t.Styles[Class(class)].decls(); decls != "" {
			fmt.Fprintf(&b, ".%s%s { %s }\n", prefix, class, decls)
		}
	}

	levels := make([]string, 0, len(t.Levels))
	for level := range t.Levels {
		if validLevelToken(level) {
			levels = append(levels, level)
		}
	}
	slices.Sort(levels)
	for _, level := range levels {
		if decls := t.Levels[level].decls(); decls != "" {
			fmt.Fprintf(&b, ".%slevel-%s { %s }\n", prefix, level, decls)
		}
	}
	return b.String()
}

// StyleElement returns the theme's CSS wrapped in a <style> element, ready to
// place in a document <head>. Pass [DefaultPrefix] or "" to match the default
// [Config].
func (t Theme) StyleElement(prefix string) *mx.Element {
	return html.StyleElem(t.CSS(prefix))
}

// The value-type colors of [LightTheme] and [DarkTheme] are adapted from
// GitHub's "primer" palette (https://github.com/primer/github-vscode-theme, MIT
// licensed), the same palette the highlight package uses, so a page can show
// code and logs side by side without a color clash.

// DarkTheme is a dark color scheme and the default of a zero [Config], because
// a log view is normally a dark panel even on a light page.
var DarkTheme = Theme{
	Name:       "dark",
	Background: "#0d1117",
	Foreground: "#c9d1d9",
	Styles: map[Class]Style{
		ClassTime:      {Color: "#6e7681"},
		ClassKey:       {Color: "#8b949e"},
		ClassPunct:     {Color: "#6e7681"},
		ClassMessage:   {Color: "#c9d1d9"},
		ClassString:    {Color: "#a5d6ff"},
		ClassNumber:    {Color: "#79c0ff"},
		ClassBool:      {Color: "#d2a8ff"},
		ClassNull:      {Color: "#ffa657", Italic: true},
		ClassText:      {Color: "#c9d1d9"},
		ClassMultiline: {Color: "#ffa198"},
		ClassMatch:     {Background: "rgba(210, 153, 34, 0.45)"},
	},
	Levels: map[string]LevelStyle{
		"trace":        {Style: Style{Color: "#6e7681"}},
		"debug":        {Style: Style{Color: "#8b949e"}},
		"info":         {Style: Style{Color: "#58a6ff"}},
		"warn":         {Style: Style{Color: "#d29922", Bold: true}},
		"warning":      {Style: Style{Color: "#d29922", Bold: true}},
		"error":        {Style: Style{Color: "#f85149", Bold: true}},
		"fatal":        {Style: Style{Color: "#ffffff", Background: "#8e1519", Bold: true}},
		"panic":        {Style: Style{Color: "#ffffff", Background: "#8e1519", Bold: true}},
		LevelUndefined: {Style: Style{Color: "#8b949e", Italic: true}},
	},
}

// LightTheme is a light color scheme.
var LightTheme = Theme{
	Name:       "light",
	Background: "#f6f8fa",
	Foreground: "#24292e",
	Styles: map[Class]Style{
		ClassTime:      {Color: "#6a737d"},
		ClassKey:       {Color: "#6a737d"},
		ClassPunct:     {Color: "#959da5"},
		ClassMessage:   {Color: "#24292e"},
		ClassString:    {Color: "#032f62"},
		ClassNumber:    {Color: "#005cc5"},
		ClassBool:      {Color: "#6f42c1"},
		ClassNull:      {Color: "#e36209", Italic: true},
		ClassText:      {Color: "#24292e"},
		ClassMultiline: {Color: "#86181d"},
		ClassMatch:     {Background: "rgba(255, 212, 0, 0.65)"},
	},
	Levels: map[string]LevelStyle{
		"trace":        {Style: Style{Color: "#959da5"}},
		"debug":        {Style: Style{Color: "#6a737d"}},
		"info":         {Style: Style{Color: "#0366d6"}},
		"warn":         {Style: Style{Color: "#b08800", Bold: true}},
		"warning":      {Style: Style{Color: "#b08800", Bold: true}},
		"error":        {Style: Style{Color: "#d73a49", Bold: true}},
		"fatal":        {Style: Style{Color: "#ffffff", Background: "#d73a49", Bold: true}},
		"panic":        {Style: Style{Color: "#ffffff", Background: "#d73a49", Bold: true}},
		LevelUndefined: {Style: Style{Color: "#6a737d", Italic: true}},
	},
}
