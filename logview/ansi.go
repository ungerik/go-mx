package logview

import "strings"

const esc = '\x1b'

// stripANSI removes terminal escape sequences from s: SGR color codes and every
// other CSI sequence (ESC [ … final), the string sequences OSC, DCS, SOS, PM and
// APC (ESC ] … up to BEL or ESC \, which is how a terminal hyperlink is
// written), and the short ESC sequences (ESC ( B, ESC 7, ESC =, …). Everything
// else is kept byte for byte.
//
// A sequence that does not complete is dropped as far as it got, the way a
// terminal swallows it: the bytes after a lone ESC are parameters to the
// terminal, not text. A CSI sequence ends early at a byte that cannot belong to
// it, so a stray "ESC [" in front of ordinary text costs only those two bytes.
// A string sequence has no such byte and runs to its terminator or to the end
// of s.
//
// The 8-bit C1 forms of the same introducers are not recognized: in UTF-8 they
// are two-byte runes that no colorizing logger emits.
func stripANSI(s string) string {
	i := strings.IndexByte(s, esc)
	if i < 0 {
		// The common case, a line without a single ESC, costs one scan and no
		// allocation.
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i >= 0 {
		b.WriteString(s[:i])
		s = s[skipEscape(s, i):]
		i = strings.IndexByte(s, esc)
	}
	b.WriteString(s)
	return b.String()
}

// skipEscape returns the index just past the escape sequence that starts with
// the ESC at s[i].
func skipEscape(s string, i int) int {
	i++ // the ESC
	if i >= len(s) {
		return i
	}
	switch c := s[i]; {
	case c == '[':
		// CSI: parameter bytes 0x30–0x3F, then intermediate bytes 0x20–0x2F,
		// then one final byte 0x40–0x7E (ECMA-48 5.4). Parameters and
		// intermediates are accepted in any order, which real terminals do too.
		for i++; i < len(s); i++ {
			switch c := s[i]; {
			case c >= 0x20 && c <= 0x3f:
				continue
			case c >= 0x40 && c <= 0x7e:
				return i + 1
			}
			// Not part of a CSI sequence: it ends here, and the byte is text.
			return i
		}
		return i
	case c == ']' || c == 'P' || c == 'X' || c == '^' || c == '_':
		// OSC, DCS, SOS, PM and APC carry a string terminated by ST (ESC \).
		// OSC is also commonly terminated by BEL.
		for i++; i < len(s); i++ {
			switch s[i] {
			case '\a':
				return i + 1
			case esc:
				if i+1 < len(s) && s[i+1] == '\\' {
					return i + 2
				}
			}
		}
		return i
	default:
		// nF sequences (intermediates 0x20–0x2F, then a final byte 0x30–0x7E)
		// and the single-byte Fp, Fe and Fs ones, which are the same shape with
		// no intermediates.
		for ; i < len(s) && s[i] >= 0x20 && s[i] <= 0x2f; i++ {
		}
		if i < len(s) && s[i] >= 0x30 && s[i] <= 0x7e {
			return i + 1
		}
		return i
	}
}

// stripFieldsANSI removes escape sequences from the decoded string keys and
// values of a parsed record. A JSON line cannot carry a raw ESC — JSON requires
// control characters to be escaped — but it can carry "\u001b[31m", which
// decodes to one: docker's json-file driver wraps every colored output line
// that way. A value past [Config.MaxDepth] is left as written, since its raw
// JSON text shows the escape as "\u001b", not as a control character.
func stripFieldsANSI(fields []field) {
	for i := range fields {
		fields[i].key = stripANSI(fields[i].key)
		stripValueANSI(&fields[i].val)
	}
}

func stripValueANSI(v *value) {
	switch v.kind {
	case kindString:
		v.text = stripANSI(v.text)
	case kindObject:
		stripFieldsANSI(v.fields)
	case kindArray:
		for i := range v.items {
			stripValueANSI(&v.items[i])
		}
	}
}
