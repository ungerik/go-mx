// Package highlight turns Go source code into syntax-highlighted output built
// from go-mx components.
//
// It is independent of the shadcn and other higher-level packages: it depends
// only on the root mx package and the html element helpers, so it can be used
// on its own or composed into any go-mx markup, including the shadcn UI.
//
// Highlighting works in two steps, the way an editor does:
//
//  1. [TokenizeGo] splits source into a flat list of [Token]s, each tagged
//     with a semantic [TokenClass] (keyword, string, comment, ...). It uses
//     the standard library go/scanner, so it has no third-party dependencies
//     and is lenient about syntactically invalid input. The concatenation of
//     all token texts reproduces the input byte-for-byte.
//
//  2. A [Highlighter] turns those tokens into output. Two backends share the
//     same tokens and configuration. [Highlighter.Component] and
//     [Highlighter.HTML] emit the highlighted markup, built through mx/html
//     components: every highlighted token becomes a <span class="hl-CLASS">
//     and the rest render as plain escaped text. [Highlighter.GoSource]
//     instead emits Go source code that, using the html package, builds that
//     same markup; it is a generator, not an echo of the input, so feeding it
//     "func main() {}" returns a tree of html.Pre(...) calls.
//
// Colors live in a separate [Theme] that emits a CSS stylesheet, so the same
// HTML works with any theme. Use [LightTheme] or [DarkTheme], or build your
// own.
//
//	import "github.com/ungerik/go-mx/highlight"
//
//	block := highlight.Component(src)             // *mx.Element: <pre><code>…</code></pre>
//	code  := highlight.GoSource(src)              // string: html.Pre(...) Go source
//	style := highlight.LightTheme.StyleElement("") // <style>…</style>
//
// Render the markup with a non-indenting writer (the default
// [mx.NewCheckedWriter]); an indenting writer would inject whitespace between
// the spans and corrupt the code layout inside <pre>.
package highlight

import (
	"go/scanner"
	"go/token"
)

// TokenClass is the semantic category of a token. It is also used, with the
// [Highlighter] prefix, as the CSS class name of a highlighted token's <span>.
type TokenClass string

const (
	// ClassPlain is the zero value: text that is not highlighted (whitespace
	// and, by default, identifiers, operators and punctuation).
	ClassPlain TokenClass = ""

	ClassKeyword     TokenClass = "keyword"     // if, for, func, package, ...
	ClassType        TokenClass = "type"        // predeclared types: int, string, error, ...
	ClassFunction    TokenClass = "function"    // a called or declared function/method name
	ClassBuiltin     TokenClass = "builtin"     // predeclared functions: make, len, append, ...
	ClassConstant    TokenClass = "constant"    // predeclared values: true, false, nil, iota
	ClassString      TokenClass = "string"      // string and rune literals
	ClassNumber      TokenClass = "number"      // int, float and imaginary literals
	ClassComment     TokenClass = "comment"     // line and block comments
	ClassOperator    TokenClass = "operator"    // +, :=, ==, <-, ...
	ClassPunctuation TokenClass = "punctuation" // parentheses, braces, commas, dots, ...
	ClassIdent       TokenClass = "ident"       // any other identifier
)

// Token is a single classified slice of the source. The concatenation of the
// Text of every token returned by [TokenizeGo] equals the original source.
type Token struct {
	Class TokenClass
	Text  string
}

// scanTok is a raw token from go/scanner together with its byte offset in the
// source, kept so classification can look at neighboring tokens.
type scanTok struct {
	off int
	tok token.Token
	lit string
}

// TokenizeGo splits Go source into classified [Token]s. It never returns an
// error: invalid input is scanned as far as possible and any unrecognized
// bytes are emitted as plain text, so the result always reproduces src exactly
// when the token texts are concatenated.
func TokenizeGo(src string) []Token {
	if src == "" {
		return nil
	}

	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s scanner.Scanner
	// A nil error handler keeps the scanner lenient: it counts errors but does
	// not stop, so we still get a token stream for broken code.
	s.Init(file, []byte(src), nil, scanner.ScanComments)

	// First pass: collect the raw scanned tokens so classification can look at
	// neighbors (e.g. an identifier followed by "(" is a function call).
	var scanned []scanTok
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		scanned = append(scanned, scanTok{off: file.Offset(pos), tok: tok, lit: lit})
	}

	// Second pass: emit tokens, reconstructing the gaps (whitespace and any
	// bytes the scanner skipped) from the original source so the output is a
	// byte-faithful copy of src.
	out := make([]Token, 0, len(scanned)*2+1)
	prevEnd := 0
	for i, t := range scanned {
		text := rawText(t.tok, t.lit)
		if text == "" {
			// Auto-inserted semicolon (lit == "\n"): it has no source bytes of
			// its own; the newline is part of the next gap.
			continue
		}
		if t.off > prevEnd {
			out = append(out, Token{Class: ClassPlain, Text: src[prevEnd:t.off]})
		}
		end := t.off + len(text)
		out = append(out, Token{Class: classify(scanned, i), Text: src[t.off:end]})
		prevEnd = end
	}
	if prevEnd < len(src) {
		out = append(out, Token{Class: ClassPlain, Text: src[prevEnd:]})
	}
	return out
}

// rawText returns the source text of a scanned token, or "" for an
// auto-inserted semicolon (which has no bytes in the source).
func rawText(tok token.Token, lit string) string {
	if tok == token.SEMICOLON && lit == "\n" {
		return ""
	}
	if lit != "" {
		// Literals, comments, keywords and real semicolons carry their text.
		return lit
	}
	// Operators and delimiters have a fixed spelling.
	return tok.String()
}

// classify assigns a [TokenClass] to scanned[i] using neighboring tokens where
// needed.
func classify(scanned []scanTok, i int) TokenClass {
	t := scanned[i]
	switch {
	case t.tok == token.COMMENT:
		return ClassComment
	case t.tok == token.STRING || t.tok == token.CHAR:
		return ClassString
	case t.tok == token.INT || t.tok == token.FLOAT || t.tok == token.IMAG:
		return ClassNumber
	case t.tok.IsKeyword():
		return ClassKeyword
	case t.tok == token.IDENT:
		return classifyIdent(scanned, i)
	case isPunctuation(t.tok):
		return ClassPunctuation
	case t.tok.IsOperator():
		return ClassOperator
	default:
		return ClassPlain
	}
}

// classifyIdent refines an identifier into a type, constant, builtin, function
// or plain identifier.
func classifyIdent(scanned []scanTok, i int) TokenClass {
	name := scanned[i].lit
	switch {
	case predeclaredTypes[name]:
		return ClassType
	case predeclaredConsts[name]:
		return ClassConstant
	case predeclaredFuncs[name]:
		return ClassBuiltin
	}
	// An identifier immediately followed by "(" is a function or method call.
	if i+1 < len(scanned) && scanned[i+1].tok == token.LPAREN {
		return ClassFunction
	}
	// The name in a plain function declaration: "func Name". A method's name
	// is already covered by the "(" lookahead above.
	if i > 0 && scanned[i-1].tok == token.FUNC {
		return ClassFunction
	}
	return ClassIdent
}

// isPunctuation reports whether tok is a delimiter rather than an operator.
func isPunctuation(tok token.Token) bool {
	switch tok {
	case token.LPAREN, token.LBRACK, token.LBRACE,
		token.RPAREN, token.RBRACK, token.RBRACE,
		token.COMMA, token.PERIOD, token.SEMICOLON,
		token.COLON, token.ELLIPSIS:
		return true
	}
	return false
}
