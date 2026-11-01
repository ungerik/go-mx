package examples

import (
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/highlight"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// codeHighlighter colors the CodeBlock examples' Go tokens under a class
// prefix of their own, so codeTheme's light and dark rules cannot collide
// with the gallery's Code tab, which uses the default prefix.
var codeHighlighter = &highlight.Highlighter{Prefix: "code-"}

// codeTheme returns the colors for codeHighlighter's tokens: the light theme,
// and the dark theme for descendants of the gallery's .dark root.
func codeTheme() mx.Component {
	dark := strings.ReplaceAll(highlight.DarkTheme.CSS("code-"), "\n.code-", "\n.dark .code-")
	return html.StyleElem(highlight.LightTheme.CSS("code-") + dark)
}

const codeBlockExampleSource = `package main

import "fmt"

func main() {
	fmt.Println("Hello, go-mx")
}`

// CodeBlockDemo renders a Go file with a title and a copy button. The
// highlighting comes from the highlight package and is passed in as the
// content's children: CodeBlock does not highlight by itself.
func CodeBlockDemo() mx.Component {
	return html.DivClass("w-full max-w-md",
		codeTheme(),
		shadcn.CodeBlock("", "",
			shadcn.CodeBlockHeader(
				shadcn.CodeBlockTitle(iconFileText(), "main.go"),
				shadcn.CodeBlockActions(shadcn.CodeBlockCopyButton(codeBlockExampleSource)),
			),
			shadcn.CodeBlockContent(codeHighlighter.Components(highlight.TokenizeGo(codeBlockExampleSource))),
		),
	)
}

// CodeBlockDiff renders numbered lines with diff marks, starting at line 12.
func CodeBlockDiff() mx.Component {
	return shadcn.CodeBlock("", "", html.Class("max-w-md"),
		shadcn.CodeBlockHeader(shadcn.CodeBlockTitle("handler.go")),
		shadcn.CodeBlockContent(shadcn.CodeBlockLineNumbers, shadcn.CodeBlockStartLine(12),
			shadcn.CodeBlockLine("", "func handle(w http.ResponseWriter, r *http.Request) {"),
			shadcn.CodeBlockLine(shadcn.CodeBlockLineRemoved, "	log.Println(r.URL)"),
			shadcn.CodeBlockLine(shadcn.CodeBlockLineAdded, "	slog.Info(\"request\", \"url\", r.URL)"),
			shadcn.CodeBlockLine(shadcn.CodeBlockLineHighlighted, "	render(w, r)"),
			shadcn.CodeBlockLine(""),
			shadcn.CodeBlockLine("", "}"),
		),
		shadcn.CodeBlockFooter("1 line removed, 1 added"),
	)
}

// CodeBlockFloating renders a small block whose actions float over the
// corner until hovered, with lines that wrap instead of scrolling.
func CodeBlockFloating() mx.Component {
	return html.DivClass("flex w-full max-w-md flex-col gap-4",
		shadcn.CodeBlock("", shadcn.CodeBlockSizeSM,
			shadcn.CodeBlockActions(shadcn.CodeBlockCopyButton("")),
			shadcn.CodeBlockContent(shadcn.CodeBlockWrap,
				"go run ./cmd/shadcn-gallery -out docs/shadcn/gallery -base /go-mx/shadcn/gallery -static-highlight"),
		),
		shadcn.CodeBlock(shadcn.CodeBlockDestructive, shadcn.CodeBlockSizeSM,
			shadcn.CodeBlockContent(shadcn.CodeBlockWrap, "panic: runtime error: index out of range [3] with length 3"),
		),
	)
}
