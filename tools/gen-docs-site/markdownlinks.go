package main

import (
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// rewriteMarkdownLinks returns the Markdown source with the destination of
// every inline link and image, and of every link reference definition,
// replaced by what rewrite returns for it; image reports whether the
// destination is the source of an image. Destinations inside code blocks and
// code spans are left alone.
//
// The Markdown alternates of the pages need it: they are the Markdown source
// itself, read without the site around it, so a link relative to the file in
// the repository has to become an absolute URL. It is a scanner for the link
// syntax the documentation uses rather than a full CommonMark parser; which
// lines are code comes from goldmark's syntax tree, the one the HTML pages
// are rendered from, so fenced and indented code blocks are recognized
// wherever goldmark sees them, in list items and block quotes too.
func rewriteMarkdownLinks(source []byte, rewrite func(dest string, image bool) string) []byte {
	lines := strings.SplitAfter(string(source), "\n")
	code := codeBlockLines(source, lines)
	var b strings.Builder
	for i, line := range lines {
		if code[i] {
			b.WriteString(line)
			continue
		}
		if m := referenceDefinition.FindStringSubmatchIndex(line); m != nil {
			b.WriteString(line[:m[2]])
			b.WriteString(rewriteDestination(line[m[2]:m[3]], false, rewrite))
			b.WriteString(line[m[3]:])
			continue
		}
		b.WriteString(rewriteInlineLinks(line, rewrite))
	}
	return []byte(b.String())
}

// codeBlockLines reports for each of lines, the lines of source, whether
// goldmark parses it as content of a fenced or indented code block.
func codeBlockLines(source []byte, lines []string) []bool {
	starts := make([]int, len(lines))
	offset := 0
	for i, line := range lines {
		starts[i] = offset
		offset += len(line)
	}
	code := make([]bool, len(lines))
	doc := newMarkdown().Parser().Parse(text.NewReader(source))
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || (n.Kind() != ast.KindCodeBlock && n.Kind() != ast.KindFencedCodeBlock) {
			return ast.WalkContinue, nil
		}
		segments := n.Lines()
		for i := range segments.Len() {
			start := segments.At(i).Start
			// The line holding the segment, the last one starting at or
			// before it.
			line := sort.Search(len(starts), func(j int) bool { return starts[j] > start }) - 1
			code[line] = true
		}
		return ast.WalkSkipChildren, nil
	})
	return code
}

// referenceDefinition matches a link reference definition, "[label]: dest",
// its submatch the destination. A footnote definition, "[^label]: text", is
// none; its links are inline links.
var referenceDefinition = regexp.MustCompile(`^ {0,3}\[[^\]^][^\]]*\]:[ \t]*(<[^>\n]*>|\S+)`)

// rewriteInlineLinks rewrites the destinations of the inline links and images
// of one line, the "(dest)" after a "]", skipping code spans. A "]" whose
// "[" is on an earlier line, as in link text wrapped across lines, is a
// link's too.
func rewriteInlineLinks(line string, rewrite func(dest string, image bool) string) string {
	var b strings.Builder
	codeRun := 0 // length of the backtick run that opened the current code span
	// opened holds for each "[" not yet closed whether it opens an image, "![".
	var opened []bool
	for i := 0; i < len(line); {
		switch {
		case line[i] == '`':
			n := len(line[i:]) - len(strings.TrimLeft(line[i:], "`"))
			switch {
			case codeRun == 0:
				codeRun = n
			case codeRun == n:
				codeRun = 0
			}
			b.WriteString(line[i : i+n])
			i += n
			continue
		case codeRun > 0:
		case line[i] == '[':
			opened = append(opened, i > 0 && line[i-1] == '!')
		case line[i] == ']':
			image := false // a "[" on an earlier line is taken as a link's
			if n := len(opened); n > 0 {
				image = opened[n-1]
				opened = opened[:n-1]
			}
			if strings.HasPrefix(line[i:], "](") {
				if end := destinationEnd(line, i+2); end > 0 {
					b.WriteString("](")
					b.WriteString(rewriteDestination(line[i+2:end], image, rewrite))
					i = end
					continue
				}
			}
		}
		b.WriteByte(line[i])
		i++
	}
	return b.String()
}

// destinationEnd returns the index in line just past the link destination
// that starts at start: up to a closing ">" for one in angle brackets, else
// up to the first space or the ")" that balances the parentheses. It returns
// -1 if there is no destination there.
func destinationEnd(line string, start int) int {
	if start < len(line) && line[start] == '<' {
		end := strings.IndexByte(line[start:], '>')
		if end < 0 {
			return -1
		}
		return start + end + 1
	}
	depth := 0
	for i := start; i < len(line); i++ {
		switch line[i] {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				if i == start {
					return -1
				}
				return i
			}
			depth--
		case ' ', '\t', '\n', '\r':
			if i == start {
				return -1
			}
			return i
		}
	}
	return -1
}

// rewriteDestination rewrites one destination, of an image if image is true,
// keeping angle brackets around it.
func rewriteDestination(dest string, image bool, rewrite func(string, bool) string) string {
	if inner, ok := strings.CutPrefix(dest, "<"); ok {
		return "<" + rewrite(strings.TrimSuffix(inner, ">"), image) + ">"
	}
	return rewrite(dest, image)
}
