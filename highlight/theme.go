package highlight

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Style is the appearance of one [TokenClass] in a [Theme].
type Style struct {
	Color  string // any CSS color, "" to inherit
	Bold   bool
	Italic bool
}

// decls renders the Style as the body of a CSS rule, e.g.
// "color: #6a737d; font-style: italic". It returns "" for the empty Style.
func (s Style) decls() string {
	var parts []string
	if s.Color != "" {
		parts = append(parts, "color: "+s.Color)
	}
	if s.Bold {
		parts = append(parts, "font-weight: bold")
	}
	if s.Italic {
		parts = append(parts, "font-style: italic")
	}
	return strings.Join(parts, "; ")
}

// Theme maps token classes to colors and provides the CSS to render them. The
// HTML produced by a [Highlighter] is theme-independent, so the same markup
// can be styled by any theme that uses the same class prefix.
type Theme struct {
	Name       string // human-readable theme name
	Background string // <pre> background-color, "" to inherit
	Foreground string // <pre> base text color, "" to inherit
	Styles     map[TokenClass]Style
}

// CSS renders the theme as a CSS stylesheet for the given class prefix (pass
// [DefaultPrefix] or "" to match the default [Highlighter]). It emits one rule
// for the <pre> block and one rule per styled token class, in deterministic
// order.
func (t Theme) CSS(prefix string) string {
	if prefix == "" {
		prefix = DefaultPrefix
	}
	block := strings.TrimRight(prefix, "-")

	var b strings.Builder
	fmt.Fprintf(&b, "pre.%s {\n", block)
	if t.Background != "" {
		fmt.Fprintf(&b, "\tbackground-color: %s;\n", t.Background)
	}
	if t.Foreground != "" {
		fmt.Fprintf(&b, "\tcolor: %s;\n", t.Foreground)
	}
	b.WriteString("\tpadding: 1rem;\n")
	b.WriteString("\tborder-radius: 6px;\n")
	b.WriteString("\toverflow-x: auto;\n")
	b.WriteString("\ttab-size: 4;\n")
	b.WriteString("\tfont-family: ui-monospace, SFMono-Regular, \"SF Mono\", Menlo, Consolas, monospace;\n")
	b.WriteString("\tfont-size: 0.875rem;\n")
	b.WriteString("\tline-height: 1.5;\n")
	b.WriteString("}\n")

	classes := make([]string, 0, len(t.Styles))
	for class := range t.Styles {
		classes = append(classes, string(class))
	}
	slices.Sort(classes)
	for _, class := range classes {
		decls := t.Styles[TokenClass(class)].decls()
		if decls == "" {
			continue
		}
		fmt.Fprintf(&b, ".%s%s { %s }\n", prefix, class, decls)
	}
	return b.String()
}

// StyleElement returns the theme's CSS wrapped in a <style> element, ready to
// place in a document <head>. It uses the given prefix; pass [DefaultPrefix]
// or "" to match the default [Highlighter].
func (t Theme) StyleElement(prefix string) *mx.Element {
	return html.StyleElem(t.CSS(prefix))
}

// The token colors of [LightTheme] and [DarkTheme] are adapted from GitHub's
// "primer" syntax-highlighting palette (https://github.com/primer/github-vscode-theme,
// MIT licensed). Only the color values are reused; the highlighting engine is
// this package's own.

// LightTheme is a light, GitHub-like color scheme.
var LightTheme = Theme{
	Name:       "light",
	Background: "#f6f8fa",
	Foreground: "#24292e",
	Styles: map[TokenClass]Style{
		ClassKeyword:     {Color: "#d73a49"},
		ClassType:        {Color: "#6f42c1"},
		ClassFunction:    {Color: "#6f42c1"},
		ClassBuiltin:     {Color: "#e36209"},
		ClassConstant:    {Color: "#005cc5"},
		ClassString:      {Color: "#032f62"},
		ClassNumber:      {Color: "#005cc5"},
		ClassComment:     {Color: "#6a737d", Italic: true},
		ClassOperator:    {Color: "#d73a49"},
		ClassPunctuation: {Color: "#24292e"},
		ClassIdent:       {Color: "#24292e"},
	},
}

// DarkTheme is a dark, GitHub-like color scheme.
var DarkTheme = Theme{
	Name:       "dark",
	Background: "#0d1117",
	Foreground: "#c9d1d9",
	Styles: map[TokenClass]Style{
		ClassKeyword:     {Color: "#ff7b72"},
		ClassType:        {Color: "#d2a8ff"},
		ClassFunction:    {Color: "#d2a8ff"},
		ClassBuiltin:     {Color: "#ffa657"},
		ClassConstant:    {Color: "#79c0ff"},
		ClassString:      {Color: "#a5d6ff"},
		ClassNumber:      {Color: "#79c0ff"},
		ClassComment:     {Color: "#8b949e", Italic: true},
		ClassOperator:    {Color: "#ff7b72"},
		ClassPunctuation: {Color: "#c9d1d9"},
		ClassIdent:       {Color: "#c9d1d9"},
	},
}
