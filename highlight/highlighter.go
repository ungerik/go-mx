package highlight

import "strings"

// DefaultPrefix is prepended to every token CSS class name. Trimmed of its
// trailing "-" it also names the <pre> block ("hl").
const DefaultPrefix = "hl-"

// DefaultHighlighted is the set of token classes wrapped in a <span> by a
// zero-value [Highlighter]. Operators, punctuation and plain identifiers render
// as text, which keeps the markup small and matches how editors like GitHub's
// highlight Go.
var DefaultHighlighted = map[TokenClass]bool{
	ClassKeyword:  true,
	ClassType:     true,
	ClassFunction: true,
	ClassBuiltin:  true,
	ClassConstant: true,
	ClassString:   true,
	ClassNumber:   true,
	ClassComment:  true,
}

// Highlighter holds the rendering configuration shared by both output
// backends. The zero value is ready to use and is exposed as [Default].
type Highlighter struct {
	// Prefix is prepended to every token CSS class name. It conventionally
	// ends with "-". An empty Prefix means [DefaultPrefix].
	Prefix string
	// Highlighted selects which token classes are wrapped in a <span>; every
	// other class renders as plain text. A nil map means [DefaultHighlighted].
	Highlighted map[TokenClass]bool
}

// Default is the zero-value Highlighter used by the package-level functions.
var Default = &Highlighter{}

func (h *Highlighter) prefix() string {
	if h.Prefix == "" {
		return DefaultPrefix
	}
	return h.Prefix
}

func (h *Highlighter) highlighted() map[TokenClass]bool {
	if h.Highlighted == nil {
		return DefaultHighlighted
	}
	return h.Highlighted
}

// BlockClass is the class name put on the <pre> block, derived from the prefix
// by trimming a trailing "-" (so [DefaultPrefix] "hl-" yields "hl").
func (h *Highlighter) BlockClass() string {
	return strings.TrimRight(h.prefix(), "-")
}

// renderItem is one unit of output: either a highlighted token (Class set) or a
// run of merged plain text (Class == ClassPlain).
type renderItem struct {
	Class TokenClass
	Text  string
}

// items reduces tokens to the units both backends render: highlighted tokens
// are kept individually, while consecutive non-highlighted tokens are merged
// into a single plain-text item to keep the output compact.
func (h *Highlighter) items(tokens []Token) []renderItem {
	highlighted := h.highlighted()
	items := make([]renderItem, 0, len(tokens))
	for _, t := range tokens {
		if t.Text == "" {
			continue
		}
		if highlighted[t.Class] {
			items = append(items, renderItem{Class: t.Class, Text: t.Text})
			continue
		}
		if n := len(items); n > 0 && items[n-1].Class == ClassPlain {
			items[n-1].Text += t.Text
		} else {
			items = append(items, renderItem{Class: ClassPlain, Text: t.Text})
		}
	}
	return items
}
