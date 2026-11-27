package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/highlight"
	"github.com/ungerik/go-mx/html"
)

// newMarkdown returns the Markdown converter of the site: CommonMark with the
// GitHub Flavored Markdown extensions (tables, task lists, strikethrough,
// autolinks), heading ids for anchors, raw HTML passed through (the Markdown
// is the repository's own), footnotes as GitHub renders them, and code blocks
// rendered by codeBlockRenderer.
func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Footnote),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(),
			// Below the default 1000, so it replaces the default
			// renderer of code blocks.
			renderer.WithNodeRenderers(util.Prioritized(codeBlockRenderer{}, 100)),
		),
	)
}

// renderMarkdown renders the Markdown body of page as HTML, with the links
// resolved by links and rewritten relative to the page's file.
func renderMarkdown(page *Page, links *resolver) (string, error) {
	md := newMarkdown()
	ctx := parser.NewContext(parser.WithIDs(newGitHubIDs()))
	doc := md.Parser().Parse(text.NewReader(page.Body), parser.WithContext(ctx))
	var errs []error
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		var dest *[]byte
		switch n := n.(type) {
		case *ast.Link:
			dest = &n.Destination
		case *ast.Image:
			dest = &n.Destination
		default:
			return ast.WalkContinue, nil
		}
		t, err := links.resolve(page, string(*dest))
		if err != nil {
			errs = append(errs, err)
			return ast.WalkContinue, nil
		}
		*dest = []byte(t.relativeTo(page.File()))
		return ast.WalkContinue, nil
	})
	if err := errors.Join(errs...); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, page.Body, doc); err != nil {
		return "", fmt.Errorf("%s: %w", page.Source, err)
	}
	return buf.String(), nil
}

// gitHubIDs generates heading ids the way GitHub does for the anchors of a
// rendered Markdown file, so the fragments the READMEs link to work on the
// site too: the heading text lower-cased, letters, digits, "-" and "_" kept,
// each space a "-", everything else dropped, and a "-1", "-2", … appended to
// a repeated id. It implements goldmark's parser.IDs.
type gitHubIDs struct {
	used map[string]int
}

func newGitHubIDs() *gitHubIDs { return &gitHubIDs{used: make(map[string]int)} }

// Generate returns the id of a heading with the Markdown text value. The
// text still holds the inline markup, whose punctuation is dropped like any
// other, so "Build with `html`" becomes "build-with-html".
func (g *gitHubIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	var b strings.Builder
	for _, r := range strings.TrimSpace(string(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) || r == '-' || r == '_':
			b.WriteRune(unicode.ToLower(r))
		case r == ' ':
			b.WriteByte('-')
		}
	}
	id := b.String()
	if id == "" {
		id = "section"
	}
	n, seen := g.used[id]
	g.used[id] = n + 1
	if seen {
		id = fmt.Sprintf("%s-%d", id, n)
		g.used[id]++
	}
	return []byte(id)
}

// Put marks id as used, for an id set explicitly in the Markdown.
func (g *gitHubIDs) Put(value []byte) {
	g.used[string(value)]++
}

// codeBlockRenderer renders fenced and indented code blocks: Go highlighted
// server-side by the highlight package, so the page needs no script for it,
// every other language as plain escaped text with a language-… class.
type codeBlockRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, renderCodeBlock)
	reg.Register(ast.KindCodeBlock, renderCodeBlock)
}

// renderCodeBlock renders a code block node once, on entering it.
func renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	var code bytes.Buffer
	lines := n.Lines()
	for i := range lines.Len() {
		segment := lines.At(i)
		code.Write(segment.Value(source))
	}
	var lang string
	if fenced, ok := n.(*ast.FencedCodeBlock); ok {
		lang = string(fenced.Language(source))
	}
	if err := codeBlock(lang, code.String()).Render(context.Background(), mx.NewCheckedWriter(w)); err != nil {
		return ast.WalkStop, err
	}
	_, _ = w.WriteString("\n")
	return ast.WalkSkipChildren, nil
}

// codeBlock returns the component of a code block in the language lang (the
// info string of a fenced block, empty for none): Go highlighted with the
// highlight package's markup, colored by the dark theme stylesheet in the page
// head; any other language a plain <pre><code> with the same frame.
func codeBlock(lang, code string) mx.Component {
	switch lang {
	case "go", "golang":
		return highlightedGo(code)
	}
	attribs := []any{}
	if lang != "" {
		attribs = append(attribs, html.Class("language-"+lang))
	}
	return html.Pre(html.Class(highlight.Default.BlockClass()+" code-plain"), html.Code(append(attribs, code)...))
}

// highlightedGo renders Go source highlighted by the highlight package, as
// the gallery's static code blocks are.
func highlightedGo(code string) mx.Component {
	return html.Pre(
		html.Class(highlight.Default.BlockClass()),
		html.Code(highlight.Default.Components(highlight.TokenizeGo(code))),
	)
}
