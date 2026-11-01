package logview

import "strings"

// overwriteCR applies a terminal's carriage-return semantics to s, for
// [Config.OverwriteCR]: within each line segment — s split at LF — only the
// text after the last carriage return is kept, the way a progress bar that
// redraws itself with "\r" ends up on screen.
//
// A carriage return that nothing follows in its segment is a line ending, not
// an overwrite, and is kept as it is: the CR of a CRLF, or the "\r" a progress
// bar writes after its last redraw before the newline. Dropping the text in
// front of it would erase the line instead of showing its final state.
//
// The overwrite is whole-segment rather than character by character: a
// terminal printing "downloading\rdone" shows "doneloading", which no reader
// wants, and progress bars clear the rest of the line anyway (with ESC [ K,
// which stripANSI removes).
func overwriteCR(s string) string {
	if strings.IndexByte(s, '\r') < 0 {
		// The common case, a line without a carriage return, costs one scan and
		// no allocation.
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for {
		segment, rest, more := strings.Cut(s, "\n")
		content := strings.TrimRight(segment, "\r")
		b.WriteString(content[strings.LastIndexByte(content, '\r')+1:])
		b.WriteString(segment[len(content):])
		if !more {
			return b.String()
		}
		b.WriteByte('\n')
		s = rest
	}
}
