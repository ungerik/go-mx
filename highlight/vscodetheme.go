package highlight

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

// ParseVSCodeTheme reads a VS Code color theme, the JSON of a
// *-color-theme.json file as Visual Studio Code, Cursor and other forks save
// it (comments and trailing commas allowed), into a [Theme].
//
// The background and foreground come from the "colors" entries
// "editor.background" and "editor.foreground"; a missing one falls back to the
// settings of a "tokenColors" rule without a scope, and the foreground also to
// the workbench "foreground" color, else it stays "" to inherit the page's.
// The token classes take the colors and font styles the "tokenColors" rules
// give the TextMate scopes of VS Code's Go grammar, chosen per class (see
// vscodeScopes): as in TextMate, the rule with the most specific matching
// selector wins, a later rule winning a tie, and the color and the font style
// are resolved separately. Selectors with descendant or exclusion operators
// are not matched. A class no rule matches gets no style and shows in the
// foreground color.
//
// The mapping is an approximation: TextMate scopes are finer than the token
// classes, and "semanticTokenColors" (the colors of a language server's
// tokens) are ignored. A theme that extends another with "include" needs
// [LoadVSCodeTheme], and "tokenColors" given as the path of a TextMate
// .tmTheme file is not supported. Every color must be a #rgb, #rgba, #rrggbb
// or #rrggbbaa hex value, so a theme file cannot inject CSS. The errors are
// plain standard-library errors, as the package has no third-party
// dependencies.
func ParseVSCodeTheme(data []byte) (Theme, error) {
	file, err := parseVSCodeThemeFile(data)
	if err != nil {
		return Theme{}, err
	}
	if file.Include != "" {
		return Theme{}, fmt.Errorf("VS Code theme includes %q: load it with LoadVSCodeTheme", file.Include)
	}
	return file.theme()
}

// LoadVSCodeTheme reads the VS Code color theme file name from fsys like
// [ParseVSCodeTheme], following its "include" chain: an included theme's
// colors are overridden by the including file's, and its token color rules
// come first, so the including file's win a tie. An include path is relative
// to the directory of the file naming it.
func LoadVSCodeTheme(fsys fs.FS, name string) (Theme, error) {
	var merged vscodeThemeFile
	seen := make(map[string]bool)
	for name != "" {
		name = path.Clean(name)
		if seen[name] {
			return Theme{}, fmt.Errorf("VS Code theme %s includes itself", name)
		}
		seen[name] = true
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return Theme{}, err
		}
		file, err := parseVSCodeThemeFile(data)
		if err != nil {
			return Theme{}, fmt.Errorf("%s: %w", name, err)
		}
		// Walking from the including file to the included one, the file
		// read now is the base of everything read before it.
		if merged.Name == "" {
			merged.Name = file.Name
		}
		for key, value := range file.Colors {
			if _, ok := merged.Colors[key]; !ok {
				if merged.Colors == nil {
					merged.Colors = make(map[string]string)
				}
				merged.Colors[key] = value
			}
		}
		merged.TokenColors = append(file.TokenColors, merged.TokenColors...)
		if file.Include != "" {
			name = path.Join(path.Dir(name), file.Include)
		} else {
			name = ""
		}
	}
	return merged.theme()
}

// vscodeThemeFile is the part of a VS Code color theme file the loader reads.
type vscodeThemeFile struct {
	Name        string
	Include     string
	Colors      map[string]string
	TokenColors []vscodeTokenRule
}

// vscodeTokenRule is one "tokenColors" rule: the simple selectors of its
// scope and its settings.
type vscodeTokenRule struct {
	Selectors  []string
	Foreground string
	Background string
	FontStyle  *string // nil if the rule sets none; "" resets it
}

// parseVSCodeThemeFile parses the JSON of one theme file, without following
// its include.
func parseVSCodeThemeFile(data []byte) (vscodeThemeFile, error) {
	var raw struct {
		Name        string          `json:"name"`
		Include     string          `json:"include"`
		Colors      map[string]any  `json:"colors"`
		TokenColors json.RawMessage `json:"tokenColors"`
	}
	if err := json.Unmarshal(stripJSONC(data), &raw); err != nil {
		return vscodeThemeFile{}, fmt.Errorf("VS Code theme: %w", err)
	}
	file := vscodeThemeFile{Name: raw.Name, Include: raw.Include, Colors: make(map[string]string)}
	for key, value := range raw.Colors {
		// A theme may set a color to null to unset it.
		if s, ok := value.(string); ok && s != "" {
			file.Colors[key] = s
		}
	}
	if trimmed := bytes.TrimSpace(raw.TokenColors); len(trimmed) > 0 && trimmed[0] == '"' {
		return vscodeThemeFile{}, errors.New("VS Code theme: tokenColors as the path of a .tmTheme file is not supported")
	}
	var rules []struct {
		Scope    json.RawMessage `json:"scope"`
		Settings struct {
			Foreground string  `json:"foreground"`
			Background string  `json:"background"`
			FontStyle  *string `json:"fontStyle"`
		} `json:"settings"`
	}
	if len(raw.TokenColors) > 0 && string(raw.TokenColors) != "null" {
		if err := json.Unmarshal(raw.TokenColors, &rules); err != nil {
			return vscodeThemeFile{}, fmt.Errorf("VS Code theme tokenColors: %w", err)
		}
	}
	for _, r := range rules {
		selectors, err := scopeSelectors(r.Scope)
		if err != nil {
			return vscodeThemeFile{}, err
		}
		file.TokenColors = append(file.TokenColors, vscodeTokenRule{
			Selectors:  selectors,
			Foreground: r.Settings.Foreground,
			Background: r.Settings.Background,
			FontStyle:  r.Settings.FontStyle,
		})
	}
	return file, nil
}

// scopeSelectors returns the selectors of a rule's scope, a string of
// comma-separated selectors or an array of them; none for a rule without a
// scope, which applies to everything.
func scopeSelectors(scope json.RawMessage) ([]string, error) {
	if len(scope) == 0 || string(scope) == "null" {
		return nil, nil
	}
	var list []string
	var one string
	if err := json.Unmarshal(scope, &one); err == nil {
		list = strings.Split(one, ",")
	} else if err := json.Unmarshal(scope, &list); err != nil {
		return nil, fmt.Errorf("VS Code theme: scope %s is neither a string nor a list of strings", scope)
	}
	selectors := make([]string, 0, len(list))
	for _, s := range list {
		if s = strings.TrimSpace(s); s != "" {
			selectors = append(selectors, s)
		}
	}
	return selectors, nil
}

// vscodeScopes are, per token class, the TextMate scopes of VS Code's Go
// grammar that stand for it, the most telling first: a class takes its style
// from the first scope any rule matches.
var vscodeScopes = map[TokenClass][]string{
	ClassKeyword:     {"keyword.control.go", "keyword.go"},
	ClassType:        {"storage.type.numeric.go", "storage.type.string.go", "support.type.builtin.go", "entity.name.type.go"},
	ClassFunction:    {"entity.name.function.go", "support.function.go"},
	ClassBuiltin:     {"entity.name.function.support.builtin.go", "support.function.builtin.go", "support.function.go"},
	ClassConstant:    {"constant.language.go"},
	ClassString:      {"string.quoted.double.go", "string.go"},
	ClassNumber:      {"constant.numeric.decimal.go", "constant.numeric.go"},
	ClassComment:     {"comment.line.double-slash.go", "comment.go"},
	ClassOperator:    {"keyword.operator.arithmetic.go", "keyword.operator.go"},
	ClassPunctuation: {"punctuation.other.comma.go", "punctuation.separator.go", "punctuation.go"},
	ClassIdent:       {"variable.other.go", "variable.go"},
}

// theme resolves the file's colors and rules into a Theme.
func (f vscodeThemeFile) theme() (Theme, error) {
	t := Theme{Name: f.Name, Styles: make(map[TokenClass]Style)}
	var global vscodeTokenRule
	for _, r := range f.TokenColors {
		if len(r.Selectors) == 0 {
			if r.Foreground != "" {
				global.Foreground = r.Foreground
			}
			if r.Background != "" {
				global.Background = r.Background
			}
		}
	}
	t.Background = firstNonEmpty(f.Colors["editor.background"], global.Background)
	t.Foreground = firstNonEmpty(f.Colors["editor.foreground"], global.Foreground, f.Colors["foreground"])
	for _, c := range []string{t.Background, t.Foreground} {
		if err := checkHexColor(c); err != nil {
			return Theme{}, err
		}
	}
	for class, scopes := range vscodeScopes {
		for _, scope := range scopes {
			style, matched := f.styleOf(scope)
			if !matched {
				continue
			}
			if err := checkHexColor(style.Color); err != nil {
				return Theme{}, err
			}
			if style != (Style{}) {
				t.Styles[class] = style
			}
			break
		}
	}
	return t, nil
}

// styleOf returns the style the rules give scope and whether any rule's
// selector matches it. The color and the font style come from the most
// specific matching rule that sets them, a later rule winning a tie.
func (f vscodeThemeFile) styleOf(scope string) (style Style, matched bool) {
	colorRank, fontRank := 0, 0
	for _, r := range f.TokenColors {
		rank := 0
		for _, sel := range r.Selectors {
			rank = max(rank, selectorRank(sel, scope))
		}
		if rank == 0 {
			continue
		}
		matched = true
		if r.Foreground != "" && rank >= colorRank {
			style.Color, colorRank = r.Foreground, rank
		}
		if r.FontStyle != nil && rank >= fontRank {
			style.Bold = strings.Contains(*r.FontStyle, "bold")
			style.Italic = strings.Contains(*r.FontStyle, "italic")
			fontRank = rank
		}
	}
	return style, matched
}

// selectorRank returns how specifically the simple selector sel matches
// scope, its number of dot-separated parts, or 0 if it does not match:
// "keyword" and "keyword.operator" both match "keyword.operator.go", the
// second more specifically. Selectors with descendant (space) or exclusion
// ("-") operators never match.
func selectorRank(sel, scope string) int {
	if strings.ContainsAny(sel, " -|&()") {
		return 0
	}
	if scope != sel && !strings.HasPrefix(scope, sel+".") {
		return 0
	}
	return strings.Count(sel, ".") + 1
}

// hexColor matches the color values a theme may give: #rgb, #rgba, #rrggbb
// and #rrggbbaa.
var hexColor = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3,4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

// checkHexColor returns an error unless c is empty or a hex color, the only
// values a theme's CSS may carry.
func checkHexColor(c string) error {
	if c != "" && !hexColor.MatchString(c) {
		return fmt.Errorf("VS Code theme: color %q is not a hex color", c)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// stripJSONC returns the JSON of a JSONC document, the JSON with comments and
// trailing commas VS Code reads: comments become spaces and a comma followed
// only by space before a closing bracket or brace is dropped. Strings are
// copied unchanged.
func stripJSONC(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		c := data[i]
		switch {
		case c == '"':
			// Copy the string with its escapes.
			start := i
			for i++; i < len(data) && data[i] != '"'; i++ {
				if data[i] == '\\' {
					i++
				}
			}
			out = append(out, data[start:min(i+1, len(data))]...)
		case c == '/' && i+1 < len(data) && data[i+1] == '/':
			for i < len(data) && data[i] != '\n' {
				i++
			}
			i-- // keep the newline
		case c == '/' && i+1 < len(data) && data[i+1] == '*':
			end := bytes.Index(data[i+2:], []byte("*/"))
			if end < 0 {
				i = len(data)
			} else {
				i += 2 + end + 1
			}
			out = append(out, ' ')
		case c == ',':
			if next := nextNonSpaceJSONC(data[i+1:]); next == ']' || next == '}' {
				continue
			}
			out = append(out, c)
		default:
			out = append(out, c)
		}
	}
	return out
}

// nextNonSpaceJSONC returns the first byte of rest that is neither white
// space nor in a comment, or 0.
func nextNonSpaceJSONC(rest []byte) byte {
	for i := 0; i < len(rest); i++ {
		switch {
		case rest[i] == ' ' || rest[i] == '\t' || rest[i] == '\n' || rest[i] == '\r':
		case rest[i] == '/' && i+1 < len(rest) && rest[i+1] == '/':
			for i < len(rest) && rest[i] != '\n' {
				i++
			}
		case rest[i] == '/' && i+1 < len(rest) && rest[i+1] == '*':
			end := bytes.Index(rest[i+2:], []byte("*/"))
			if end < 0 {
				return 0
			}
			i += 2 + end + 1
		default:
			return rest[i]
		}
	}
	return 0
}
