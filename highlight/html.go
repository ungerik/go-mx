package highlight

import (
	"context"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Components turns tokens into a sequence of mx components: a
// <span class="PREFIX+CLASS"> for every highlighted token and plain escaped
// text for everything else. Use it to place highlighted code inside a custom
// wrapper; [Highlighter.Component] and [Highlighter.Inline] wrap it for you.
func (h *Highlighter) Components(tokens []Token) mx.Components {
	prefix := h.prefix()
	items := h.items(tokens)
	comps := make(mx.Components, 0, len(items))
	for _, it := range items {
		if it.Class == ClassPlain {
			comps = append(comps, mx.Text(it.Text))
		} else {
			comps = append(comps, html.SpanClass(prefix+string(it.Class), mx.Text(it.Text)))
		}
	}
	return comps
}

// Lines turns tokens into one sequence of mx components per source line, the
// markup [Highlighter.Components] renders cut at the line breaks, so each line
// can go into an element of its own, such as a numbered line of a code block.
// A token spanning lines, a block comment or a raw string, becomes one
// <span> per line. The line breaks themselves are left out, along with the
// carriage return of a "\r\n", and a final line break ends the last line
// rather than starting an empty one, so an empty line is an empty sequence.
func (h *Highlighter) Lines(tokens []Token) []mx.Components {
	prefix := h.prefix()
	lines := []mx.Components{{}}
	for _, it := range h.items(tokens) {
		for i, part := range strings.Split(it.Text, "\n") {
			if i > 0 {
				lines = append(lines, mx.Components{})
			}
			if i < strings.Count(it.Text, "\n") {
				part = strings.TrimSuffix(part, "\r")
			}
			if part == "" {
				continue
			}
			line := &lines[len(lines)-1]
			if it.Class == ClassPlain {
				*line = append(*line, mx.Text(part))
			} else {
				*line = append(*line, html.SpanClass(prefix+string(it.Class), mx.Text(part)))
			}
		}
	}
	if len(lines) > 1 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// Component highlights Go source and returns it as a
// <pre class="hl"><code>…</code></pre> block element.
func (h *Highlighter) Component(src string) *mx.Element {
	return html.Pre(
		html.Class(h.BlockClass()),
		html.Code(h.Components(TokenizeGo(src))),
	)
}

// Inline highlights Go source for inline use and returns a <code class="hl">
// element without the <pre> wrapper.
func (h *Highlighter) Inline(src string) *mx.Element {
	return html.Code(
		html.Class(h.BlockClass()),
		h.Components(TokenizeGo(src)),
	)
}

// HTML highlights Go source and renders it directly to an HTML string using a
// non-indenting [mx.CheckedWriter], so the code layout inside <pre> is
// preserved exactly.
func (h *Highlighter) HTML(src string) (string, error) {
	var b strings.Builder
	if err := h.Component(src).Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Component highlights Go source with the [Default] highlighter. See
// [Highlighter.Component].
func Component(src string) *mx.Element { return Default.Component(src) }

// Inline highlights Go source with the [Default] highlighter. See
// [Highlighter.Inline].
func Inline(src string) *mx.Element { return Default.Inline(src) }

// HTML highlights Go source with the [Default] highlighter. See
// [Highlighter.HTML].
func HTML(src string) (string, error) { return Default.HTML(src) }
