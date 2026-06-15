package xml

import "github.com/ungerik/go-mx"

type (
	// Text is a string rendered as XML character data with the special
	// characters &, <, >, " and ' escaped to their predefined entities.
	// Alias for [mx.Text]. Plain Go strings passed as element children are
	// converted to Text and escaped the same way.
	Text = mx.Text
	// Raw is a string rendered verbatim without XML escaping; the caller is
	// responsible for its safety. Alias for [mx.Raw].
	Raw = mx.Raw
	// RawBytes is a byte slice rendered verbatim without XML escaping; the
	// caller is responsible for its safety. Alias for [mx.RawBytes].
	RawBytes = mx.RawBytes
)
