package main

import (
	"regexp"
	"strings"
)

// rewriteMarkdownLinks returns the Markdown source with the destination of
// every inline link and image, and of every link reference definition,
// replaced by what rewrite returns for it. Destinations inside fenced code
// blocks and code spans are left alone.
//
// The Markdown alternates of the pages need it: they are the Markdown source
// itself, read without the site around it, so a link relative to the file in
// the repository has to become an absolute URL. It is a scanner for the link
// syntax the documentation uses rather than a full CommonMark parser; the
// HTML pages are rendered from goldmark's syntax tree instead.
func rewriteMarkdownLinks(source []byte, rewrite func(dest string) string) []byte {
	lines := strings.SplitAfter(string(source), "\n")
	var b strings.Builder
	var fence string
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if fence != "" {
			if strings.HasPrefix(trimmed, fence) && strings.TrimSpace(strings.TrimLeft(trimmed, fence[:1])) == "" {
				fence = ""
			}
			b.WriteString(line)
			continue
		}
		if f := openingFence(trimmed); f != "" && len(line)-len(trimmed) < 4 {
			fence = f
			b.WriteString(line)
			continue
		}
		if m := referenceDefinition.FindStringSubmatchIndex(line); m != nil {
			b.WriteString(line[:m[2]])
			b.WriteString(rewriteDestination(line[m[2]:m[3]], rewrite))
			b.WriteString(line[m[3]:])
			continue
		}
		b.WriteString(rewriteInlineLinks(line, rewrite))
	}
	return []byte(b.String())
}

// referenceDefinition matches a link reference definition, "[label]: dest",
// its submatch the destination. A footnote definition, "[^label]: text", is
// none; its links are inline links.
var referenceDefinition = regexp.MustCompile(`^ {0,3}\[[^\]^][^\]]*\]:[ \t]*(<[^>\n]*>|\S+)`)

// openingFence returns the fence a line (without its indentation) opens, a
// run of three or more backticks or tildes, or "".
func openingFence(line string) string {
	for _, c := range []string{"`", "~"} {
		n := len(line) - len(strings.TrimLeft(line, c))
		if n >= 3 {
			return strings.Repeat(c, n)
		}
	}
	return ""
}

// rewriteInlineLinks rewrites the destinations of the inline links and images
// of one line, the "(dest)" after a "]", skipping code spans.
func rewriteInlineLinks(line string, rewrite func(dest string) string) string {
	var b strings.Builder
	codeRun := 0 // length of the backtick run that opened the current code span
	for i := 0; i < len(line); {
		if line[i] == '`' {
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
		}
		if codeRun == 0 && strings.HasPrefix(line[i:], "](") {
			if end := destinationEnd(line, i+2); end > 0 {
				b.WriteString("](")
				b.WriteString(rewriteDestination(line[i+2:end], rewrite))
				i = end
				continue
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

// rewriteDestination rewrites one destination, keeping angle brackets around
// it.
func rewriteDestination(dest string, rewrite func(string) string) string {
	if inner, ok := strings.CutPrefix(dest, "<"); ok {
		return "<" + rewrite(strings.TrimSuffix(inner, ">")) + ">"
	}
	return rewrite(dest)
}
